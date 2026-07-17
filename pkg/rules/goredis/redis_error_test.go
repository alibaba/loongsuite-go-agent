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

package goredis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisSpanEndErr(t *testing.T) {
	assert.Nil(t, redisSpanEndErr(nil))
	assert.Nil(t, redisSpanEndErr(redis.Nil))
	assert.Nil(t, redisSpanEndErr(errors.Join(redis.Nil)))
	assert.Nil(t, redisSpanEndErr(fmt.Errorf("wrap: %w", redis.Nil)))

	realErr := errors.New("connection refused")
	assert.Equal(t, realErr, redisSpanEndErr(realErr))
}

func TestRedisPipelineSpanEndErr(t *testing.T) {
	nilCmd := redis.NewCmd(context.Background(), "get", "missing")
	nilCmd.SetErr(redis.Nil)
	failCmd := redis.NewCmd(context.Background(), "incr", "string-key")
	realErr := errors.New("ERR value is not an integer or out of range")
	failCmd.SetErr(realErr)

	assert.Nil(t, redisPipelineSpanEndErr([]redis.Cmder{nilCmd}, redis.Nil))
	assert.Equal(t, realErr, redisPipelineSpanEndErr([]redis.Cmder{nilCmd, failCmd}, redis.Nil))
	assert.Equal(t, realErr, redisPipelineSpanEndErr(nil, realErr))
}

func TestGetRedisV9Statement(t *testing.T) {
	cmd := redis.NewCmd(context.Background(), "get", "mykey")
	assert.Equal(t, "get mykey: get mykey", getRedisV9Statement(cmd))

	cmd.SetErr(redis.Nil)
	stmt := getRedisV9Statement(cmd)
	assert.Equal(t, "get mykey: get mykey: redis: nil", stmt)
	assert.Equal(t, 1, strings.Count(stmt, "redis: nil"))

	cmd.SetErr(errors.New("connection refused"))
	assert.Contains(t, getRedisV9Statement(cmd), "connection refused")
}
