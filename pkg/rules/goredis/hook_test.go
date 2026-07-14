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
	"testing"

	"github.com/redis/go-redis/v9"
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
	// Instrumenter captures tracer at build time; rebuild after test provider is set.
	goRedisInstrumenter = BuildGoRedisOtelInstrumenter()
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return sr
}

func TestRedisSpanEndErr(t *testing.T) {
	assert.Nil(t, redisSpanEndErr(nil))
	assert.Nil(t, redisSpanEndErr(redis.Nil))

	realErr := errors.New("connection refused")
	assert.Equal(t, realErr, redisSpanEndErr(realErr))
}

func TestGetRedisV9Statement(t *testing.T) {
	cmd := redis.NewCmd(context.Background(), "get", "mykey")
	assert.Contains(t, getRedisV9Statement(cmd), "get mykey")
}

func TestProcessHook_RedisNilNotError(t *testing.T) {
	sr := setupTestTracer(t)

	hook := newOtRedisHook("localhost:6379")
	processHook := hook.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
		return redis.Nil
	})

	cmd := redis.NewCmd(context.Background(), "get", "nonexistent")
	err := processHook(context.Background(), cmd)
	assert.ErrorIs(t, err, redis.Nil)

	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Unset, span.Status().Code)
}

func TestProcessHook_RecordsError(t *testing.T) {
	sr := setupTestTracer(t)

	hook := newOtRedisHook("localhost:6379")
	expectedErr := errors.New("connection refused")
	processHook := hook.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
		return expectedErr
	})

	cmd := redis.NewCmd(context.Background(), "get", "mykey")
	err := processHook(context.Background(), cmd)
	assert.Equal(t, expectedErr, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Error, span.Status().Code)
	require.NotEmpty(t, span.Events())
	assert.Equal(t, "exception", span.Events()[0].Name)
}

func TestProcessPipelineHook_RedisNilNotError(t *testing.T) {
	sr := setupTestTracer(t)

	hook := newOtRedisHook("localhost:6379")
	pipelineHook := hook.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
		return redis.Nil
	})

	cmds := []redis.Cmder{
		redis.NewCmd(context.Background(), "get", "key1"),
	}
	err := pipelineHook(context.Background(), cmds)
	assert.ErrorIs(t, err, redis.Nil)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Unset, spans[0].Status().Code)
}
