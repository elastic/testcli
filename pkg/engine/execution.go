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
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/elastic/cloud-sdk-go/pkg/multierror"

	"github.com/elastic/testcli/pkg/engine/teststorage"
)

const (
	// Red fail text
	failRed = "\x1b[31;1mFAIL\x1b[0m"

	defaultCooldownPeriod = 100 * time.Millisecond
)

// ExecuteTests takes in the testing.T and a list of integration tests to run.
func ExecuteTests(t *testing.T, tests Tests) {
	var storage = teststorage.GetInMemory()

	for testN, tt := range tests {
		t.Run(tt.Name, func(subTest *testing.T) {
			executeTestCase(subTest, testN, tt, storage)

			// Always delay each test case 100ms*0-10 so that the tests don't choke
			// the client machine where the tests are running.
			<-time.After(defaultCooldownPeriod*time.Duration(rand.Intn(9)+1) + tt.WaitBeforeRun)
		})
	}
}

func executeTestCase(t *testing.T, testN int, tt Test, storage teststorage.Storage) {
	// The first part of the command's arguments, having the config slice
	// first and then appending the positional command's arguments or flags.
	//
	// The second part of the arguments, are the dynamic args.
	// uses the list of strings to load the value from result[key] into the argument
	// list which will be passed to the cmd to be executed.
	// Any keys that aren't present in the result map and are prefixed with a "-"
	// will be appended to the arguments as the key that was passed as they're meant
	// to. Useful to add extra arguments after the dynamically loaded ones.
	if tt.Parallel {
		t.Parallel()
	}

	dynamicArgs, err := parseDynamicArguments(tt.Args.DynamicArgs, storage)
	if err != nil {
		t.Fatalf("[Test %d][%s]: %s", testN, failRed, err)
	}

	var args = append(
		append(tt.Args.Config, tt.Args.Args...), dynamicArgs...,
	)

	if tt.Binary == "" {
		t.Fatalf("[Test %d][%s]: binary not set, please set a binary name", testN, failRed)
	}
	binary := tt.Binary

	if tt.FindBinary {
		found, err := FindBinaryPath(".", binary)
		if err != nil {
			t.Fatal(err)
		}
		binary = found
	}

	stdout, stderr, err := runCommand(
		t, binary, testN, args, tt.Args.Interactive, tt.Assert.WantErr,
	)

	// Ensures the assertions.
	var merr = multierror.NewPrefixed(fmt.Sprintf("[Test %d][%s]", testN, failRed))
	if err := tt.Assert.Ensure(stdout, stderr, err, storage,
		redactPasswordFlag(strings.Join(append([]string{binary}, args...), " ")),
	); err != nil {
		merr = merr.Append(err)
	}

	// The callbacks are used to populate the storage on runtime.
	// Decoding happens inside a tailored function which parses the []byte output
	// to a specific data structure, which populates result[key].
	if err := tt.Callbacks.Run(stdout.Bytes(), storage); err != nil {
		merr = merr.Append(err)
	}

	// Make the test fail.
	if merr.ErrorOrNil() != nil {
		t.Error(merr)
	}
}

func parseDynamicArguments(dynamicArgs []string, storage teststorage.Storage) ([]string, error) {
	var result []string
	for _, key := range dynamicArgs {
		// If the key is a flag (Prefixed by "-") or if the
		// key is prefixed by "strip_" the key name is appended as the dynamic
		// argument vs the key value since it hasn't been stored as such.
		// In case "strip_" is used, the "strip_" prefix is removed.
		prefixedKey := strings.HasPrefix(key, "-") || strings.HasPrefix(key, "strip_")
		if prefixedKey {
			if strings.HasPrefix(key, "strip_") {
				key = strings.Replace(key, "strip_", "", 1)
			}
			result = append(result, key)
			continue
		}

		value, ok := storage.Get(key)
		if !ok {
			return nil, fmt.Errorf("failed to obtain value of key %s", key)
		}
		result = append(result, value)
	}
	return result, nil
}

func runCommand(t *testing.T, bin string, testN int, args, interactive []string, wantErr bool) (*bytes.Buffer, *bytes.Buffer, error) {
	// NTH?: CommandContext might be interesting here
	var cmd = exec.Command(bin, args...)
	var stderr, stdout = bytes.Buffer{}, bytes.Buffer{}
	cmd.Stderr, cmd.Stdout = &stderr, &stdout
	cmd.Env = append(cmd.Env, os.Environ()...)

	if len(interactive) == 0 {
		return &stdout, &stderr, cmd.Run()
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("[Test %d][%s]: Command %s: failed to create stdin pipe", testN, failRed, bin)
	}
	defer stdin.Close()

	if err := cmd.Start(); (err != nil) != wantErr {
		printableArgs := redactPasswordFlag(strings.Join(args, " "))
		t.Errorf("[Test %d][%s]: Command %s %v error = %v, wantErr = %v, stderr = %v", testN, failRed, bin, printableArgs, err, wantErr, stderr.String())
		return &stdout, &stderr, err
	}

	for _, line := range interactive {
		_, _ = io.WriteString(stdin, fmt.Sprintln(line))
	}

	return &stdout, &stderr, cmd.Wait()
}

// FindBinaryPath executes a reverse walk to find the ecl binary on the parent path.
func FindBinaryPath(p, binary string) (string, error) {
	var binaryPath string
	p, _ = filepath.Abs(p)
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if info.Name() == binary && !info.IsDir() {
			relPath, err := filepath.Rel(p, path)
			if err != nil {
				return err
			}
			binaryPath = filepath.Join(relPath)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if binaryPath == "" {
		binaryPath, err = FindBinaryPath(filepath.Dir(p), binary)
		return filepath.Join("..", binaryPath), err
	}

	return binaryPath, nil
}

func redactPasswordFlag(cmd string) string {
	var re = regexp.MustCompile(`(?m)\-\-pass?[ =]([^ ]+)`)
	return re.ReplaceAllString(cmd, "--pass [REDACTED]")
}
