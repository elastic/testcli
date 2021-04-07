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

import "sync"

var result = NewSafeMap()

// GetInMemory obtains the singleton instance of a SafeMap to be shared between
// integration tests.
func GetInMemory() *SafeMap { return result }

// NewSafeMap initializes a SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{db: make(map[string]string)}
}

// SafeMap provides concurrent RW access to a map
// to be able to use it across goroutines
type SafeMap struct {
	db map[string]string
	sync.RWMutex
}

// Set sets a a key with a value.
func (m *SafeMap) Set(k, v string) {
	m.Lock()
	defer m.Unlock()
	m.db[k] = v
}

// Get obtains a key and returns it and whether or not it was found.
func (m *SafeMap) Get(k string) (string, bool) {
	m.RLock()
	defer m.RUnlock()
	r, ok := m.db[k]
	return r, ok
}
