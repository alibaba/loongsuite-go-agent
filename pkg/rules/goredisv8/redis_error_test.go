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

package goredisv8

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisSpanEndErr(t *testing.T) {
	assert.Nil(t, redisSpanEndErr(nil))
	assert.Nil(t, redisSpanEndErr(redis.Nil))
	assert.Nil(t, redisSpanEndErr(fmt.Errorf("wrap: %w", redis.Nil)))

	realErr := errors.New("connection refused")
	assert.Equal(t, realErr, redisSpanEndErr(realErr))
}

func TestIsRedisSpanError(t *testing.T) {
	assert.False(t, isRedisSpanError(nil))
	assert.False(t, isRedisSpanError(redis.Nil))
	assert.False(t, isRedisSpanError(fmt.Errorf("wrap: %w", redis.Nil)))
	assert.True(t, isRedisSpanError(errors.New("connection refused")))
}

func TestGetRedisV8Statement(t *testing.T) {
	cmd := redis.NewCmd(context.Background(), "get", "mykey")
	assert.Equal(t, "get mykey: get mykey", getRedisV8Statement(cmd))

	cmd.SetErr(redis.Nil)
	stmt := getRedisV8Statement(cmd)
	assert.Equal(t, "get mykey: get mykey: redis: nil", stmt)
	assert.Equal(t, 1, strings.Count(stmt, "redis: nil"))

	cmd.SetErr(errors.New("connection refused"))
	assert.Contains(t, getRedisV8Statement(cmd), "connection refused")
}
