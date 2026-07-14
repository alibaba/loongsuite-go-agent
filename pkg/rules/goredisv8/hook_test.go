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
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	redisv8Instrumenter = BuildRedisv8Instrumenter()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return sr
}

func TestRedisSpanEndErr(t *testing.T) {
	assert.Nil(t, redisSpanEndErr(nil))
	assert.Nil(t, redisSpanEndErr(redis.Nil))

	realErr := errors.New("connection refused")
	assert.Equal(t, realErr, redisSpanEndErr(realErr))
}

func TestIsRedisSpanError(t *testing.T) {
	assert.False(t, isRedisSpanError(nil))
	assert.False(t, isRedisSpanError(redis.Nil))
	assert.True(t, isRedisSpanError(errors.New("connection refused")))
}

func TestGetRedisV8Statement(t *testing.T) {
	cmd := redis.NewCmd(context.Background(), "get", "mykey")
	assert.Contains(t, getRedisV8Statement(cmd), "get mykey")

	cmd.SetErr(redis.Nil)
	assert.Contains(t, getRedisV8Statement(cmd), "get mykey")

	cmd.SetErr(errors.New("connection refused"))
	assert.Contains(t, getRedisV8Statement(cmd), "connection refused")
}

func TestAfterProcess_RedisNilNotError(t *testing.T) {
	sr := setupTestTracer(t)
	hook := newOtRedisV8Hook("localhost:6379")

	cmd := redis.NewCmd(context.Background(), "get", "nonexistent")
	cmd.SetErr(redis.Nil)

	ctx, err := hook.BeforeProcess(context.Background(), cmd)
	require.NoError(t, err)
	require.NoError(t, hook.AfterProcess(ctx, cmd))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Unset, spans[0].Status().Code)
}

func TestAfterProcess_RecordsError(t *testing.T) {
	sr := setupTestTracer(t)
	hook := newOtRedisV8Hook("localhost:6379")

	cmd := redis.NewCmd(context.Background(), "get", "mykey")
	expectedErr := errors.New("connection refused")
	cmd.SetErr(expectedErr)

	ctx, err := hook.BeforeProcess(context.Background(), cmd)
	require.NoError(t, err)
	require.NoError(t, hook.AfterProcess(ctx, cmd))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestAfterProcessPipeline_RedisNilNotError(t *testing.T) {
	sr := setupTestTracer(t)
	hook := newOtRedisV8Hook("localhost:6379")

	cmd := redis.NewCmd(context.Background(), "get", "key1")
	cmd.SetErr(redis.Nil)
	cmds := []redis.Cmder{cmd}

	ctx, err := hook.BeforeProcessPipeline(context.Background(), cmds)
	require.NoError(t, err)
	require.NoError(t, hook.AfterProcessPipeline(ctx, cmds))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Unset, spans[0].Status().Code)
}

func TestAfterProcessPipeline_RecordsError(t *testing.T) {
	sr := setupTestTracer(t)
	hook := newOtRedisV8Hook("localhost:6379")

	nilCmd := redis.NewCmd(context.Background(), "get", "missing")
	nilCmd.SetErr(redis.Nil)
	failCmd := redis.NewCmd(context.Background(), "get", "key")
	failCmd.SetErr(errors.New("connection refused"))
	cmds := []redis.Cmder{nilCmd, failCmd}

	ctx, err := hook.BeforeProcessPipeline(context.Background(), cmds)
	require.NoError(t, err)
	require.NoError(t, hook.AfterProcessPipeline(ctx, cmds))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}
