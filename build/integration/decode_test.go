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
