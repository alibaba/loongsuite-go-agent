// Copyright (c) 2026 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"testing"
)

func TestRouteFromPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{"GET /users/{id}", "/users/{id}"},
		{"/users/{id}", "/users/{id}"},
		{"POST /query", "/query"},
		{"/health", "/health"},
		{"GET example.com/users/{id}", "/users/{id}"},
		{"GET example.com", ""},
		{"", ""},
		{"GET", ""},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			if got := RouteFromPattern(tt.pattern); got != tt.expected {
				t.Fatalf("RouteFromPattern(%q) = %q, want %q", tt.pattern, got, tt.expected)
			}
		})
	}
}
