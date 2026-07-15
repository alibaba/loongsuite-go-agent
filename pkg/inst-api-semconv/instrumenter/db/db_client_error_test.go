// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package db

import (
	"errors"
	"fmt"
	"net"
	"testing"
)

type customDBError struct{}

func (customDBError) Error() string     { return "custom" }
func (customDBError) ErrorType() string { return "db.custom" }

type typedNilDBError struct {
	kind string
}

func (e *typedNilDBError) Error() string     { return e.kind }
func (e *typedNilDBError) ErrorType() string { return e.kind }

func TestNormalizeDBClientErrorType(t *testing.T) {
	if got := NormalizeDBClientErrorType(nil); got != "" {
		t.Fatalf("nil err should return empty, got %q", got)
	}
	var typedNil error = (*typedNilDBError)(nil)
	if got := NormalizeDBClientErrorType(typedNil); got != "" {
		t.Fatalf("typed nil should return empty, got %q", got)
	}
	if got := NormalizeDBClientErrorType(errors.New("x")); got != "*errors.errorString" {
		t.Fatalf("unexpected error.type %q", got)
	}
	if got := NormalizeDBClientErrorType(&net.OpError{}); got != "*net.OpError" {
		t.Fatalf("unexpected error.type %q", got)
	}
	if got := NormalizeDBClientErrorType(fmt.Errorf("dial: %w", &net.OpError{})); got != "*net.OpError" {
		t.Fatalf("wrapped error should unwrap to *net.OpError, got %q", got)
	}
	if got := NormalizeDBClientErrorType(errors.Join(&net.OpError{}, errors.New("a"))); got != "*net.OpError" {
		t.Fatalf("joined error should use first cause *net.OpError, got %q", got)
	}
	if got := NormalizeDBClientErrorType(customDBError{}); got != "db.custom" {
		t.Fatalf("ErrorType() interface should take precedence, got %q", got)
	}
}

func TestNormalizeDBClientErrorTypeDoubleWrapped(t *testing.T) {
	base := customDBError{}
	wrapped := fmt.Errorf("app: %w", fmt.Errorf("db: %w", base))
	if got := NormalizeDBClientErrorType(wrapped); got != "db.custom" {
		t.Fatalf("expected db.custom after unwrap, got %q", got)
	}
}
