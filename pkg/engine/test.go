// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package engine

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/elastic/testcli/pkg/engine/teststorage"
)

// Tests is a collection of tests.
type Tests []Test

// Test defines a test
type Test struct {
	// The test name
	Name string

	// The relative or full path of the binary to use for the test.
	Binary string

	// When set, the Binary name will be used to reverse find the binary on the
	// path. It will traverse the current path backwards until it finds a file
	// This is meant to be used with binary artifacts that have been built and
	// can be found within the project directory boundaries.
	FindBinary bool

	// Arguments to pass to the binary.
	Args Args

	// the following list of strings must be found in the specified
	// channels, stdout, stderr.
	Assert Assertions

	// callbacks to be run after the test is finished, the stdout output
	// is passed as the first argument and the key is used, see decode...
	// functions for callback examples
	Callbacks TestCallback

	// optionally set how much time the test should wait before run
	WaitBeforeRun time.Duration

	// If set, the test will be run in parallel instead of sequentially.
	Parallel bool
}

// Args represent the test arguments.
type Args struct {
	// Arguments to pass to the test binary
	Args []string

	// Args that represent the configuration. This field is appended to Args.
	// This needn't be used, it is a nice way to keep your command configuration
	// sepparate from the command arguments.
	Config []string

	// Uses the strings as keys to load the stored value from `teststorage.Storage`
	// the parameter is ignored if not found in the result map, and passed as the key
	DynamicArgs []string

	// list of commands to be run when an interactive session is open
	Interactive []string
}

// Assertions defines a series of Must and MustNot assertions after a test is
// run.
type Assertions struct {
	// WantErr ensures that a exit code > 0 is returned after the command's
	// execution.
	WantErr bool

	// CanError causes the test not to fail in case the command returns an error.
	// This is useful for commands which can return an error depending on external
	// factors, useful when the output is still asserted but an error might be returned.
	CanError bool

	// CanErrorWithMessage contains a slice of known failure states where the test
	// will not fail, if there's a partial match of any of the messages.
	CanErrorWithMessage []string

	// Must ensures that the defined assertions are found.
	Must Assertion

	// Not ensures that the defined assertions are not found.
	Not Assertion
}

// Assertion represent the test assertions after the test has run.
type Assertion struct {
	// Asserts the Output
	Output []string

	// Asserts the errors
	Errors []string

	// Asserts dynamically stored values (Key-based).
	Dynamic []string

	// When set to true, it ensures that all the items in Output and Errors are
	// are found.
	Strict bool

	// Regex Patterns to match.
	Pattern []string
}

// Callback is a function which receives the output in the form of []byte and
// a string which is a storage key for the value. The function will normally
// decode the output to cleanly store it in the key. In case any errors occur,
// an error will be returned.
type Callback func(output []byte, storageKey string, storage teststorage.Storage) error

// TestCallback defines a relationship between a test function and its key.
type TestCallback map[string]Callback

// Run calls each stored callback and stores the output of the command on the
// passed storage via the prefixed key in the callback map.
func (tc TestCallback) Run(out []byte, storage teststorage.Storage) error {
	var errs []error
	for key, callback := range tc {
		if err := callback(out, key, storage); err != nil {
			errs = append(errs, err)
		}
	}
	return NewPrefixedError("callback", errors.Join(errs...))
}

// NewTestCallback creates a new callback
func NewTestCallback(s string, t Callback) TestCallback {
	tc := make(TestCallback, 1)
	tc[s] = t
	return tc
}

// Ensure verifies that the assertions match, otherwise it throws an error via t.Error
func (a Assertions) Ensure(stdout, stderr *bytes.Buffer, err error, storage teststorage.Storage, args string) error {
	// Checks standard for unexpected errors when running the command
	// if err is true when WantErr is false, it will error out
	// The same applies when WantErr is true, but err is false.
	var stderrString = stderr.String()
	if (err != nil) != a.WantErr && !a.CanError && len(a.CanErrorWithMessage) == 0 {
		return fmt.Errorf(
			"command: \"%s\"\nerror = %v, wantErr = %v, stderr = %v", args, err, a.WantErr, stderrString,
		)
	}

	// If an error is returned and partially matches CanErrorWithMessage,
	// returning nil, and skipping any further assertions.
	for _, knownFailure := range a.CanErrorWithMessage {
		if strings.Contains(stderrString, knownFailure) {
			return nil
		}
	}

	// Performs all the assertions necessary to validate the output and result
	// of a test case.
	var out = stdout.String()
	var errs []error
	if err := assertWanted(out, a.Must); err != nil {
		errs = append(errs, err)
	}

	if err := assertPattern(out, a.Must.Pattern); err != nil {
		errs = append(errs, err)
	}

	if err := assertErrors(stderr, a.Must.Errors); err != nil {
		errs = append(errs, err)
	}

	if err := assertDynamic(out, a.Must.Dynamic, storage); err != nil {
		errs = append(errs, err)
	}

	// Ensures that the mustNot Output or Error is not found
	// in the respective outputs
	if err := assertMustNot(out, stderrString, a.Not); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return NewPrefixedError("assertion", errors.Join(errs...))
	}
	return nil
}

// Assertions

func assertWanted(out string, w Assertion) error {
	var errs []error
	for _, want := range w.Output {
		if w.Strict && out != want {
			errs = append(errs, fmt.Errorf("strict match got \"%s\" want \"%s\"", out, want))
		}

		if !w.Strict && !strings.Contains(out, want) {
			errs = append(errs, fmt.Errorf("didn't find \"%s\" in standard output: \"%s\"", want, out))
		}
	}

	if len(errs) > 0 {
		return NewPrefixedError("must find", errors.Join(errs...))
	}
	return nil
}

func assertPattern(out string, patterns []string) error {
	var errs []error
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			errs = append(errs,
				fmt.Errorf("match pattern \"%s\" did not compile", pattern),
			)
		}
		if re.FindStringIndex(out) == nil {
			errs = append(errs,
				fmt.Errorf("couldn't match pattern \"%s\" to standard output: \"%s\"", pattern, out),
			)
		}
	}

	if len(errs) > 0 {
		return NewPrefixedError("must find pattern", errors.Join(errs...))
	}
	return nil
}

func assertErrors(stderr *bytes.Buffer, errrs []string) error {
	var errs []error
	for _, want := range errrs {
		if !strings.Contains(stderr.String(), want) {
			errs = append(errs,
				fmt.Errorf("didn't find \"%s\" in standard error: \"%s\"", want, stderr.String()),
			)
		}
	}

	if len(errs) > 0 {
		return NewPrefixedError("must find errors", errors.Join(errs...))
	}
	return nil
}

func assertDynamic(out string, dynamic []string, storage teststorage.Storage) error {
	var errs []error
	for _, key := range dynamic {
		value := key
		if v, ok := storage.Get(key); ok {
			value = v
		}
		if !strings.Contains(out, value) {
			errs = append(errs,
				fmt.Errorf("didn't find dynamic key \"%s\" with value \"%s\" in standard output: \"%s\"", key, value, out),
			)
		}
	}

	if len(errs) > 0 {
		return NewPrefixedError("must find values from dynamic storage",
			errors.Join(errs...),
		)
	}
	return nil
}

func assertMustNot(out, stderr string, not Assertion) error {
	var errs []error
	for _, mustNot := range not.Output {
		if not.Strict && out == mustNot {
			errs = append(errs, fmt.Errorf("strict match got \"%s\" must not: \"%s\"", out, mustNot))
		}

		if !not.Strict && strings.Contains(out, mustNot) {
			errs = append(errs, fmt.Errorf("found \"%s\" in standard output: \"%s\"", mustNot, out))
		}
	}

	for _, mustNot := range not.Errors {
		if strings.Contains(stderr, mustNot) {
			errs = append(errs, fmt.Errorf("found \"%s\" in standard error:\"%s\"", mustNot, stderr))
		}
	}

	if len(errs) > 0 {
		return NewPrefixedError("must not find values", errors.Join(errs...))
	}
	return nil
}
