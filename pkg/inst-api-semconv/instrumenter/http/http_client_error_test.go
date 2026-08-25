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

package http

import (
	"errors"
	"fmt"
	"net"
	"testing"
)

func TestNormalizeHTTPClientErrorTypeConnect(t *testing.T) {
	err := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	if got := NormalizeHTTPClientErrorType(err); got != "*net.OpError" {
		t.Fatalf("expected %q, got %q", "*net.OpError", got)
	}
}

func TestNormalizeHTTPClientErrorTypeDNS(t *testing.T) {
	err := &net.DNSError{Err: "no such host", Name: "example.invalid"}
	if got := NormalizeHTTPClientErrorType(err); got != "*net.DNSError" {
		t.Fatalf("expected %q, got %q", "*net.DNSError", got)
	}
}

func TestNormalizeHTTPClientErrorTypeTimeout(t *testing.T) {
	if got := NormalizeHTTPClientErrorType(errors.New("timeout")); got != "*errors.errorString" {
		t.Fatalf("expected %q, got %q", "*errors.errorString", got)
	}
}

func TestNormalizeHTTPClientErrorTypeOther(t *testing.T) {
	if got := NormalizeHTTPClientErrorType(errors.New("boom")); got != "*errors.errorString" {
		t.Fatalf("expected %q, got %q", "*errors.errorString", got)
	}
}

func TestNormalizeHTTPClientErrorTypeNil(t *testing.T) {
	if got := NormalizeHTTPClientErrorType(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

type customErrorType struct {
	msg string
}

func (c customErrorType) Error() string {
	return c.msg
}

func (c customErrorType) ErrorType() string {
	return "my-custom-error-type"
}

func TestNormalizeHTTPClientErrorTypeCustom(t *testing.T) {
	err := customErrorType{msg: "error"}
	if got := NormalizeHTTPClientErrorType(err); got != "my-custom-error-type" {
		t.Fatalf("expected %q, got %q", "my-custom-error-type", got)
	}
}

func TestNormalizeHTTPClientErrorTypeWrapped(t *testing.T) {
	baseErr := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	wrappedErr := fmt.Errorf("middleware failed: %w", baseErr)
	if got := NormalizeHTTPClientErrorType(wrappedErr); got != "*net.OpError" {
		t.Fatalf("expected %q, got %q", "*net.OpError", got)
	}
}

func TestNormalizeHTTPClientErrorTypeDoubleWrapped(t *testing.T) {
	baseErr := customErrorType{msg: "db timeout"}
	wrappedErr1 := fmt.Errorf("db failure: %w", baseErr)
	wrappedErr2 := fmt.Errorf("app execution failed: %w", wrappedErr1)
	if got := NormalizeHTTPClientErrorType(wrappedErr2); got != "my-custom-error-type" {
		t.Fatalf("expected %q, got %q", "my-custom-error-type", got)
	}
}
