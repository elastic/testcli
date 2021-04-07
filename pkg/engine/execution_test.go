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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/elastic/testcli/pkg/engine/teststorage"
)

func TestFindEclPath(t *testing.T) {
	type args struct {
		p      string
		binary string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "finds a different binary",
			args: args{p: ".", binary: "aweirdbinaryname"},
			want: "../apath/aweirdbinaryname",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.want); err != nil {
				dir := filepath.Dir(tt.want)
				if err := os.MkdirAll(dir, 0777); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tt.want, []byte("some"), 0666); err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(dir)
			}

			got, err := FindBinaryPath(tt.args.p, tt.args.binary)
			if got != tt.want {
				t.Errorf("FindEclPath() = %v, want %v", got, tt.want)
			}
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_redactPasswordFlag(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Redact password when `--pass` is found",
			args: args{cmd: "ecl --host http://somehost --user admin --pass MySuperSecretPassword platform info"},
			want: "ecl --host http://somehost --user admin --pass [REDACTED] platform info",
		},
		{
			name: "Redact password when `--pass=pass` is found",
			args: args{cmd: "ecl --host http://somehost --user admin --pass=MySuperSecretPassword platform info"},
			want: "ecl --host http://somehost --user admin --pass [REDACTED] platform info",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redactPasswordFlag(tt.args.cmd); got != tt.want {
				t.Errorf("redactPasswordFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDynamicArguments(t *testing.T) {
	type args struct {
		dynamicArgs []string
		storage     teststorage.Storage
	}
	safemap := teststorage.NewSafeMap()
	safemap.Set("akey", "avalue")
	tests := []struct {
		name string
		args args
		want []string
		err  string
	}{
		{
			name: "Parses the dynamic arguments",
			args: args{
				dynamicArgs: []string{"akey", "strip_stripped_key"},
				storage:     safemap,
			},
			want: []string{"avalue", "stripped_key"},
		},
		{
			name: "Fails parsing unexisting key",
			args: args{
				dynamicArgs: []string{"unexisting key"},
				storage:     safemap,
			},
			err: "failed to obtain value of key unexisting key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDynamicArguments(tt.args.dynamicArgs, tt.args.storage)
			if err != nil && !reflect.DeepEqual(err.Error(), tt.err) {
				t.Errorf("parseDynamicArguments() error = %v, wantErr %v", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDynamicArguments() = %v, want %v", got, tt.want)
			}
		})
	}
}
