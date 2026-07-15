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

package rueidisnil

import (
	"errors"
	"testing"

	"github.com/redis/rueidis"
)

func TestSpanEndErr(t *testing.T) {
	if got := SpanEndErr(nil); got != nil {
		t.Fatalf("nil should stay nil, got %v", got)
	}
	if got := SpanEndErr(rueidis.Nil); got != nil {
		t.Fatalf("rueidis.Nil must not mark span error, got %v", got)
	}
	realErr := errors.New("connection refused")
	if got := SpanEndErr(realErr); got != realErr {
		t.Fatalf("real error must be preserved, got %v", got)
	}
}

func TestIsSpanError(t *testing.T) {
	if IsSpanError(nil) || IsSpanError(rueidis.Nil) {
		t.Fatalf("nil sentinels must not be span errors")
	}
	if !IsSpanError(errors.New("connection refused")) {
		t.Fatalf("real error must be a span error")
	}
}

func TestFirstError(t *testing.T) {
	if got := FirstError(nil); got != nil {
		t.Fatalf("nil slice should return nil, got %v", got)
	}
	if got := FirstError([]error{}); got != nil {
		t.Fatalf("empty slice should return nil, got %v", got)
	}
	if got := FirstError([]error{nil, rueidis.Nil}); got != nil {
		t.Fatalf("nil sentinels only should return nil, got %v", got)
	}

	realErr := errors.New("boom")
	got := FirstError([]error{rueidis.Nil, realErr})
	if got != realErr {
		t.Fatalf("expected first real error after Nil, got %v", got)
	}
	got = FirstError([]error{realErr, rueidis.Nil})
	if got != realErr {
		t.Fatalf("expected first real error before Nil, got %v", got)
	}
}
