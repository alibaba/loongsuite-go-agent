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

// Package integration contains end-to-end tests for the util-genai telemetry
// handler. Unlike the unit tests that live alongside the library, these tests
// wire the handler to in-memory OpenTelemetry SDK exporters and assert on the
// spans, metrics, and log events that are actually produced across a full
// invocation lifecycle.
//
// The tests live in a separate module so that the util-genai library itself can
// keep an API-only dependency surface (no OTel SDK), while the tests are free to
// pull in the SDK exporters they need.
package integration

import (
	"context"
	"sync"
	"testing"

	utilgenai "github.com/alibaba/loongsuite-go/util-genai"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// memLogExporter is a minimal in-memory implementation of sdklog.Exporter that
// records every log record it receives so tests can assert on emitted events.
type memLogExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *memLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.records = append(e.records, records...)
	return nil
}

func (e *memLogExporter) Shutdown(context.Context) error { return nil }

func (e *memLogExporter) ForceFlush(context.Context) error { return nil }

func (e *memLogExporter) eventNames() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	names := make([]string, 0, len(e.records))
	for i := range e.records {
		names = append(names, e.records[i].EventName())
	}
	return names
}

// harness bundles the in-memory providers and the handler under test.
type harness struct {
	handler   *utilgenai.TelemetryHandler
	spans     *tracetest.SpanRecorder
	reader    *sdkmetric.ManualReader
	logExport *memLogExporter
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	spans := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	logExport := &memLogExporter{}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(logExport)),
	)

	handler := utilgenai.NewTelemetryHandler(
		utilgenai.WithTracerProvider(tp),
		utilgenai.WithMeterProvider(mp),
		utilgenai.WithLoggerProvider(lp),
	)

	return &harness{handler: handler, spans: spans, reader: reader, logExport: logExport}
}

// collectMetric returns the metric with the given name from the manual reader,
// failing the test if it is absent.
func (h *harness) collectMetric(t *testing.T, name string) metricdata.Metrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := h.reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m
			}
		}
	}
	t.Fatalf("metric %q not found in collected metrics", name)
	return metricdata.Metrics{}
}

func attrMap(kvs []attribute.KeyValue) map[string]attribute.Value {
	m := make(map[string]attribute.Value, len(kvs))
	for _, kv := range kvs {
		m[string(kv.Key)] = kv.Value
	}
	return m
}

// TestLLMLifecycleSuccess drives a full successful chat completion through the
// handler and asserts the resulting span, metrics, and event.
func TestLLMLifecycleSuccess(t *testing.T) {
	// Enable experimental semconv + event emission so the inference event fires.
	t.Setenv(utilgenai.EnvSemconvStabilityOptIn, "gen_ai_latest_experimental")
	t.Setenv(utilgenai.EnvEmitEvent, "true")

	h := newHarness(t)
	ctx := context.Background()

	inv := utilgenai.NewLLMInvocation("gpt-4")
	inv.Provider = "openai"
	inv.InputMessages = []utilgenai.InputMessage{
		{Role: "user", Parts: []utilgenai.MessagePart{utilgenai.Text{Content: "Hello!"}}},
	}

	ctx = h.handler.StartLLM(ctx, inv)
	_ = ctx

	inv.OutputMessages = []utilgenai.OutputMessage{
		{
			Role:         "assistant",
			Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: "Hi there!"}},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}
	inTok, outTok := 12, 8
	inv.InputTokens = &inTok
	inv.OutputTokens = &outTok
	inv.ResponseModelName = "gpt-4-0613"
	inv.ResponseID = "chatcmpl-xyz"

	h.handler.StopLLM(inv)

	// --- span assertions ---
	ended := h.spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("expected 1 span, got %d", len(ended))
	}
	span := ended[0]
	if got, want := span.Name(), "chat gpt-4"; got != want {
		t.Errorf("span name = %q, want %q", got, want)
	}
	if span.Status().Code == codes.Error {
		t.Errorf("span status = Error, want non-error; description=%q", span.Status().Description)
	}
	sa := attrMap(span.Attributes())
	assertStr(t, sa, utilgenai.AttrGenAIOperationName, "chat")
	assertStr(t, sa, utilgenai.AttrGenAIProviderName, "openai")
	assertStr(t, sa, utilgenai.AttrGenAIRequestModel, "gpt-4")
	assertStr(t, sa, utilgenai.AttrGenAIResponseModel, "gpt-4-0613")
	assertInt(t, sa, utilgenai.AttrGenAIUsageInputTokens, 12)
	assertInt(t, sa, utilgenai.AttrGenAIUsageOutputTokens, 8)

	// --- metric assertions ---
	dur := h.collectMetric(t, utilgenai.MetricGenAIClientOperationDuration)
	if hist, ok := dur.Data.(metricdata.Histogram[float64]); !ok {
		t.Errorf("%s is not a float64 histogram: %T", dur.Name, dur.Data)
	} else if len(hist.DataPoints) == 0 || hist.DataPoints[0].Count == 0 {
		t.Errorf("%s has no recorded data points", dur.Name)
	}

	tok := h.collectMetric(t, utilgenai.MetricGenAIClientTokenUsage)
	hist, ok := tok.Data.(metricdata.Histogram[int64])
	if !ok {
		t.Fatalf("%s is not an int64 histogram: %T", tok.Name, tok.Data)
	}
	// Expect one data point for input tokens and one for output tokens.
	if len(hist.DataPoints) < 2 {
		t.Errorf("%s expected >=2 data points (input+output), got %d",
			tok.Name, len(hist.DataPoints))
	}
	sawInput, sawOutput := false, false
	for _, dp := range hist.DataPoints {
		if v, ok := dp.Attributes.Value(attribute.Key(utilgenai.AttrGenAITokenType)); ok {
			switch v.AsString() {
			case "input":
				sawInput = true
			case "output":
				sawOutput = true
			}
		}
	}
	if !sawInput || !sawOutput {
		t.Errorf("%s missing token.type data points: input=%v output=%v",
			tok.Name, sawInput, sawOutput)
	}

	// --- event assertions ---
	names := h.logExport.eventNames()
	if !contains(names, utilgenai.EventGenAIInferenceOperationDetails) {
		t.Errorf("expected event %q to be emitted, got events: %v",
			utilgenai.EventGenAIInferenceOperationDetails, names)
	}
}

// TestLLMLifecycleFailure drives a failed invocation and asserts the span is
// marked with an error status while the duration metric is still recorded.
func TestLLMLifecycleFailure(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	inv := utilgenai.NewLLMInvocation("gpt-4")
	inv.Provider = "openai"

	_ = h.handler.StartLLM(ctx, inv)

	h.handler.FailLLM(inv, &utilgenai.Error{
		Message: "Rate limit exceeded",
		Type:    "RateLimitError",
	})

	ended := h.spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("expected 1 span, got %d", len(ended))
	}
	if code := ended[0].Status().Code; code != codes.Error {
		t.Errorf("span status = %v, want Error", code)
	}

	// The duration metric should be recorded even on failure.
	dur := h.collectMetric(t, utilgenai.MetricGenAIClientOperationDuration)
	if hist, ok := dur.Data.(metricdata.Histogram[float64]); !ok || len(hist.DataPoints) == 0 {
		t.Errorf("%s not recorded on failure", dur.Name)
	}
}

// TestEventSuppressedWithoutOptIn verifies that no inference event is emitted
// when experimental mode / event emission is not enabled.
func TestEventSuppressedWithoutOptIn(t *testing.T) {
	// Explicitly clear the opt-in envs for this test.
	t.Setenv(utilgenai.EnvSemconvStabilityOptIn, "")
	t.Setenv(utilgenai.EnvEmitEvent, "")

	h := newHarness(t)
	inv := utilgenai.NewLLMInvocation("gpt-4")
	inv.Provider = "openai"
	_ = h.handler.StartLLM(context.Background(), inv)
	inTok, outTok := 1, 1
	inv.InputTokens = &inTok
	inv.OutputTokens = &outTok
	h.handler.StopLLM(inv)

	if names := h.logExport.eventNames(); len(names) != 0 {
		t.Errorf("expected no events without opt-in, got: %v", names)
	}
}

func assertStr(t *testing.T, m map[string]attribute.Value, key, want string) {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Errorf("attribute %q missing", key)
		return
	}
	if got := v.AsString(); got != want {
		t.Errorf("attribute %q = %q, want %q", key, got, want)
	}
}

func assertInt(t *testing.T, m map[string]attribute.Value, key string, want int64) {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Errorf("attribute %q missing", key)
		return
	}
	if got := v.AsInt64(); got != want {
		t.Errorf("attribute %q = %d, want %d", key, got, want)
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
