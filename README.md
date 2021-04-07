# Test CLI

[![Go Reference](https://pkg.go.dev/badge/github.com/elastic/testcli.svg)](https://pkg.go.dev/github.com/elastic/testcli)

_Test your compiled Go binaries as easily as if they were unit tests!_

Provides a suite on top of the built-in Go `testing.T` to run a series of tests and assertions as commands. It can be used as the integration or end to end suite to ensure that your CLI is doing what is supposed to.

## Features

* Stdout / Stderr / execution assertions.
* Run any number of tests sequentially or in parallel, a combination can be used when a few tests are expected to run sequentially before running the rest concurrently.
* Decode the output of any command into a variable that can be later used.

## Example usage

### Basic suite testing the "ls" command

```go
// +build integration

package integration

import (
	"testing"

	"github.com/elastic/testcli/pkg/engine"
)

func TestBasic_ls(t *testing.T) {
    // This allows the test to be run in parallel with others.
	t.Parallel()

	tests := engine.Tests{
		{
            // This flag sets this sub-test or test case to be run in parallel
            // with other tests in this suite.
			Parallel: true,
			Name:     "ls with assertion",
			Binary:   "ls",
			Args: engine.Args{
				Args: []string{"-1"},
			},
			Assert: engine.Assertions{
				Must: engine.Assertion{
					Output: []string{
						"ls_test.go",
					},
				},
				Not: engine.Assertion{
					Output: []string{
						"LICENSE",
					},
				},
			},
		},
		{
			Parallel: true,
			Name:     "ls -l with assertion",
			Binary:   "ls",
			Args: engine.Args{
				Args: []string{"-l"},
			},
			Assert: engine.Assertions{
				Must: engine.Assertion{
					Output: []string{
						"ls_test.go",
					},
				},
				Not: engine.Assertion{
					Output: []string{
						"LICENSE",
					},
				},
			},
		},
		{
			Parallel: true,
			Name:     "ls path that doesn't exist returns an error",
			Binary:   "ls",
			Args: engine.Args{
				Args: []string{"unexisting_path"},
			},
			Assert: engine.Assertions{
                // Allows the command to exit with status code > 0
				CanError: true,
				Must: engine.Assertion{
					Errors: []string{
						"No such file or directory",
					},
				},
			},
		},
		{
			Parallel: true,
			Name:     "ls -invalidflag a fails with concrete error",
			Binary:   "ls",
			Args: engine.Args{
				Args: []string{"-invalidflag a"},
			},
			Assert: engine.Assertions{
                // Allows the test to "fail" with a specific error message.
                // Useful when the binary tests some conditions that might
                // cause the test to due to circumstances that are out of
                // the control of the tester.
				CanErrorWithMessage: []string{
					"ls: illegal option --",
				},
			},
		},
	}
	engine.ExecuteTests(t, tests)
}
```

### Simulates a curl call and asserts its output.

The test simulates a curl call (the payload is hardcoded for stability) and performs two test assertions concurrently and also performs a Go assertion on a decoded piece of data from the JSON payload.

```go
// +build integration

package integration

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/elastic/testcli/pkg/engine"
	"github.com/elastic/testcli/pkg/engine/teststorage"
)

// This test case simulates decoding of a returned JSON structure.
// This could happen through an API call or running a command that
// returns JSON.
func TestBasic_decode(t *testing.T) {
	t.Parallel()

	type echoTestData struct {
		Message string                 `json:"message,omitempty"`
		HRefs   map[string]interface{} `json:"hrefs,omitempty"`
	}

	// This callback persist data outside of its boundary by using the pointer
	// address of testOutput to decode the JSON. The length of HRefs is asserted
	// at the end of the tests.
	var testOutput echoTestData
	decodeEchoData := func(out []byte, key string, storage teststorage.Storage) error {
		if err := json.Unmarshal(out, &testOutput); err != nil {
			return err
		}

		storage.Set(key, testOutput.Message)
		return nil
	}

	// This callback is self-contained, does not persist data outside of its
	// function boundary.
	decodeEchoDataHrefKeys := func(out []byte, key string, storage teststorage.Storage) error {
		var d echoTestData
		if err := json.Unmarshal(out, &d); err != nil {
			return err
		}

		var keys []string
		for key := range d.HRefs {
			keys = append(keys, key)
		}

		sort.Strings(keys)
		storage.Set(key, strings.Join(keys, ", "))
		return nil
	}

	// These keys must be unique on the whole integration suite.
	const (
		messageKey = "stored_message"
		hrefKey    = "stored_href"
	)

	tests := engine.Tests{
		{
			Name:   "decode json",
			Binary: "echo",
			Args: engine.Args{
				// This is equivalent to "curl https://api.elastic-cloud.com".
				Args: []string{`{"message":"You Know, for Cloud.","hrefs":{"api/v0":"https://api.elastic-cloud.com/api/v0","api/latest":"https://api.elastic-cloud.com/api/latest","api/v1":"https://api.elastic-cloud.com/api/v1","app":"https://api.elastic-cloud.com/app","api/v0.1":"https://api.elastic-cloud.com/api/v0.1"}}`},
			},
			Callbacks: engine.TestCallback{
				messageKey: decodeEchoData,
				hrefKey:    decodeEchoDataHrefKeys,
			},
			Assert: engine.Assertions{
				Must: engine.Assertion{
					Output: []string{
						"You Know, for Cloud.",
					},
				},
			},
		},
		{
			Parallel: true,
			Name:     "print decoded json field and assert strict output",
			Binary:   "echo",
			Args: engine.Args{
				DynamicArgs: []string{messageKey},
			},
			Assert: engine.Assertions{
				Must: engine.Assertion{
					Strict: true,
					Output: []string{
						"You Know, for Cloud.\n",
					},
				},
			},
		},
		{
			Parallel: true,
			Name:     "print decoded json hrefs and assert them",
			Binary:   "echo",
			Args: engine.Args{
				DynamicArgs: []string{hrefKey},
			},
			Assert: engine.Assertions{
				Must: engine.Assertion{
					Strict: true,
					Output: []string{
						"api/latest, api/v0, api/v0.1, api/v1, app\n",
					},
				},
			},
		},
	}
	engine.ExecuteTests(t, tests)

	// Assert that the output of the hrefs is 4.
	if length := len(testOutput.HRefs); length != 4 {
		t.Errorf("expected hrefs to contain 4 items but got: %d", length)
	}
}
```

## Contributing

See the [CONTRIBUTING](./CONTRIBUTING.md) doc for more information on how to contribute.
