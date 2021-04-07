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

package teststorage

import (
	"reflect"
	"testing"
)

func TestSafeMap_Set(t *testing.T) {
	type fields struct {
		db map[string]string
	}
	type args struct {
		k string
		v string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name:   "Set a value to an empty db",
			fields: fields{db: make(map[string]string)},
			args:   args{k: "key", v: "value"},
			want:   map[string]string{"key": "value"},
		},
		{
			name: "Set a value to an existing db",
			fields: fields{
				db: map[string]string{"key": "value"},
			},
			args: args{k: "somekey", v: "somevalue"},
			want: map[string]string{
				"key":     "value",
				"somekey": "somevalue",
			},
		},
		{
			name: "Overwites a value to an existing db",
			fields: fields{
				db: map[string]string{"key": "value"},
			},
			args: args{k: "key", v: "somevalue"},
			want: map[string]string{"key": "somevalue"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SafeMap{db: tt.fields.db}
			m.Set(tt.args.k, tt.args.v)
			if !reflect.DeepEqual(m.db, tt.want) {
				t.Errorf("SafeMap.Set() db contents = %v, want %v", m.db, tt.want)
			}
		})
	}
}

func TestSafeMap_Get(t *testing.T) {
	type fields struct {
		db map[string]string
	}
	type args struct {
		k string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		wantOk bool
	}{
		{
			name: "obtain a value",
			fields: fields{
				db: map[string]string{"somekey": "value"},
			},
			args:   args{k: "somekey"},
			want:   "value",
			wantOk: true,
		},
		{
			name: "tries to obtain a value which doesn't exist returns false",
			fields: fields{
				db: map[string]string{"somekey": "value"},
			},
			args:   args{k: "unexisting key"},
			want:   "",
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SafeMap{
				db: tt.fields.db,
			}
			got, got1 := m.Get(tt.args.k)
			if got != tt.want {
				t.Errorf("SafeMap.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantOk {
				t.Errorf("SafeMap.Get() got1 = %v, want %v", got1, tt.wantOk)
			}
		})
	}
}
