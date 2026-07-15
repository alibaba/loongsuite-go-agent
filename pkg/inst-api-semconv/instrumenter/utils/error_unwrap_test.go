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

package utils

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
)

func TestUnwrapFmtWrappedNil(t *testing.T) {
	if got := UnwrapFmtWrapped(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestUnwrapFmtWrappedNoOp(t *testing.T) {
	base := errors.New("boom")
	if got := UnwrapFmtWrapped(base); got != base {
		t.Fatalf("expected same error, got %v", got)
	}
}

func TestUnwrapFmtWrappedSingleWrap(t *testing.T) {
	base := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	wrapped := fmt.Errorf("middleware failed: %w", base)
	if got := UnwrapFmtWrapped(wrapped); got != base {
		t.Fatalf("expected %T, got %T", base, got)
	}
}

func TestUnwrapFmtWrappedDoubleWrap(t *testing.T) {
	base := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	wrapped := fmt.Errorf("app: %w", fmt.Errorf("db: %w", base))
	if got := UnwrapFmtWrapped(wrapped); got != base {
		t.Fatalf("expected %T, got %T", base, got)
	}
}

func TestUnwrapFmtWrappedJoin(t *testing.T) {
	base := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	joined := errors.Join(base, errors.New("other"))
	if got := UnwrapFmtWrapped(joined); got != base {
		t.Fatalf("expected %T, got %T", base, got)
	}
}

// TestStdlibWrapperTypeNames pins the unexported stdlib type names that
// UnwrapFmtWrapped relies on. Failures here typically mean a Go version
// renamed *fmt.wrapError or *errors.joinError.
func TestStdlibWrapperTypeNames(t *testing.T) {
	wrapped := fmt.Errorf("wrap: %w", errors.New("inner"))
	if got := reflect.TypeOf(wrapped).String(); got != "*fmt.wrapError" {
		t.Fatalf("unexpected fmt wrap type name %q", got)
	}
	joined := errors.Join(errors.New("a"), errors.New("b"))
	if got := reflect.TypeOf(joined).String(); got != "*errors.joinError" {
		t.Fatalf("unexpected errors.Join type name %q", got)
	}
}
