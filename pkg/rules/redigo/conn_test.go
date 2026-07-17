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
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gomodule/redigo/redis"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type fakeRedigoConn struct {
	doErr        error
	doReply      interface{}
	receiveErr   error
	receiveReply interface{}
}

func (f *fakeRedigoConn) Close() error { return nil }
func (f *fakeRedigoConn) Err() error   { return nil }
func (f *fakeRedigoConn) Do(string, ...interface{}) (interface{}, error) {
	return f.doReply, f.doErr
}
func (f *fakeRedigoConn) Send(string, ...interface{}) error { return nil }
func (f *fakeRedigoConn) Flush() error                      { return nil }
func (f *fakeRedigoConn) Receive() (interface{}, error) {
	return f.receiveReply, f.receiveErr
}

func setupRedigoTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	commandQueue.Init()
	transactionQueue.Init()

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	redigoInstrumenter = BuildRedigoInstrumenter()
	t.Cleanup(func() {
		commandQueue.Init()
		transactionQueue.Init()
		_ = tp.Shutdown(context.Background())
	})
	return sr
}

func TestRedigoSpanEndErr(t *testing.T) {
	if got := redigoSpanEndErr(nil); got != nil {
		t.Fatalf("nil should stay nil, got %v", got)
	}
	if got := redigoSpanEndErr(redis.ErrNil); got != nil {
		t.Fatalf("redis.ErrNil must not mark span error, got %v", got)
	}
	if got := redigoSpanEndErr(fmt.Errorf("wrap: %w", redis.ErrNil)); got != nil {
		t.Fatalf("wrapped ErrNil must not mark span error, got %v", got)
	}
	realErr := errors.New("connection refused")
	if got := redigoSpanEndErr(realErr); got != realErr {
		t.Fatalf("real error must be preserved, got %v", got)
	}
}

func TestDo_ErrNilNotSpanError(t *testing.T) {
	sr := setupRedigoTestTracer(t)
	conn := &armsConn{
		Conn:     &fakeRedigoConn{doErr: redis.ErrNil},
		endpoint: "localhost:6379",
		ctx:      context.Background(),
	}

	_, err := conn.Do("GET", "missing")
	if !errors.Is(err, redis.ErrNil) {
		t.Fatalf("caller must still see ErrNil, got %v", err)
	}

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Unset {
		t.Fatalf("ErrNil must not mark span Error, got %v", spans[0].Status().Code)
	}
}

func TestDo_RealErrorIsSpanError(t *testing.T) {
	sr := setupRedigoTestTracer(t)
	expected := errors.New("connection refused")
	conn := &armsConn{
		Conn:     &fakeRedigoConn{doErr: expected},
		endpoint: "localhost:6379",
		ctx:      context.Background(),
	}

	_, err := conn.Do("GET", "key")
	if err != expected {
		t.Fatalf("caller must see real error, got %v", err)
	}

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Error {
		t.Fatalf("real error must mark span Error, got %v", spans[0].Status().Code)
	}
}

func TestReceive_ErrNilNotSpanError(t *testing.T) {
	sr := setupRedigoTestTracer(t)
	conn := &armsConn{
		Conn:     &fakeRedigoConn{receiveErr: redis.ErrNil},
		endpoint: "localhost:6379",
		ctx:      context.Background(),
	}

	if err := conn.Send("GET", "missing"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	_, err := conn.Receive()
	if !errors.Is(err, redis.ErrNil) {
		t.Fatalf("caller must still see ErrNil, got %v", err)
	}

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Unset {
		t.Fatalf("ErrNil must not mark span Error, got %v", spans[0].Status().Code)
	}
}
