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

package redigo

import (
	"errors"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func TestRedigoSpanEndErr(t *testing.T) {
	if got := redigoSpanEndErr(nil); got != nil {
		t.Fatalf("nil should stay nil, got %v", got)
	}
	if got := redigoSpanEndErr(redis.ErrNil); got != nil {
		t.Fatalf("redis.ErrNil must not mark span error, got %v", got)
	}
	realErr := errors.New("connection refused")
	if got := redigoSpanEndErr(realErr); got != realErr {
		t.Fatalf("real error must be preserved, got %v", got)
	}
}
