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
	"testing"

	"github.com/elastic/testcli/pkg/engine"
)

func TestBasic_ls(t *testing.T) {
	t.Parallel()

	tests := engine.Tests{
		{
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
				CanErrorWithMessage: []string{
					"ls: illegal option --",
				},
			},
		},
	}
	engine.ExecuteTests(t, tests)
}
