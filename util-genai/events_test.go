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

package utilgenai

import (
	"context"
	"sync"
	"testing"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
)

// recordingLogger is a log.Logger that captures emitted records for assertions.
type recordingLogger struct {
	embedded.Logger
	mu      sync.Mutex
	records []log.Record
}

func (l *recordingLogger) Emit(_ context.Context, record log.Record) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = append(l.records, record.Clone())
}

func (l *recordingLogger) Enabled(context.Context, log.EnabledParameters) bool { return true }

func (l *recordingLogger) events() []log.Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]log.Record, len(l.records))
	copy(out, l.records)
	return out
}

// recordingLoggerProvider hands out a shared recordingLogger.
type recordingLoggerProvider struct {
	embedded.LoggerProvider
	logger *recordingLogger
}

func (p *recordingLoggerProvider) Logger(string, ...log.LoggerOption) log.Logger {
	return p.logger
}

// attrMap flattens a record's attributes into a key -> string-value map.
func attrMap(r log.Record) map[string]string {
	m := make(map[string]string)
	r.WalkAttributes(func(kv log.KeyValue) bool {
		m[kv.Key] = kv.Value.AsString()
		return true
	})
	return m
}

func newRecordingHandler(t *testing.T) (*TelemetryHandler, *recordingLogger) {
	t.Helper()
	rec := &recordingLogger{}
	provider := &recordingLoggerProvider{logger: rec}
	h := NewTelemetryHandler(WithLoggerProvider(provider))
	return h, rec
}

func TestEmitLLMEventDisabledByDefault(t *testing.T) {
	h, rec := newRecordingHandler(t)

	inv := NewLLMInvocation("gpt-4")
	inv.Provider = "openai"
	ctx := h.StartLLM(context.Background(), inv)
	_ = ctx
	h.StopLLM(inv)

	if got := len(rec.events()); got != 0 {
		t.Fatalf("expected no events without experimental mode, got %d", got)
	}
}

func TestEmitLLMEventWithContent(t *testing.T) {
	t.Setenv(EnvSemconvStabilityOptIn, string(StabilityModeGenAILatestExperimental))
	t.Setenv(EnvEmitEvent, "true")
	t.Setenv(EnvCaptureMessageContent, "SPAN_AND_EVENT")

	h, rec := newRecordingHandler(t)

	inv := NewLLMInvocation("gpt-4")
	inv.Provider = "openai"
	inv.InputMessages = []InputMessage{{
		Role:  "user",
		Parts: []MessagePart{Text{Content: "Hello"}},
	}}
	ctx := h.StartLLM(context.Background(), inv)
	_ = ctx
	inv.OutputMessages = []OutputMessage{{
		Role:         "assistant",
		Parts:        []MessagePart{Text{Content: "Hi there"}},
		FinishReason: FinishReasonStop,
	}}
	h.StopLLM(inv)

	events := rec.events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if name := events[0].EventName(); name != EventGenAIInferenceOperationDetails {
		t.Fatalf("unexpected event name: %q", name)
	}
	attrs := attrMap(events[0])
	if attrs[AttrGenAIInputMessages] == "" {
		t.Errorf("expected input messages captured in event")
	}
	if attrs[AttrGenAIOutputMessages] == "" {
		t.Errorf("expected output messages captured in event")
	}
}

func TestEmitLLMEventWithoutContent(t *testing.T) {
	t.Setenv(EnvSemconvStabilityOptIn, string(StabilityModeGenAILatestExperimental))
	t.Setenv(EnvEmitEvent, "true")
	t.Setenv(EnvCaptureMessageContent, "NO_CONTENT")

	h, rec := newRecordingHandler(t)

	inv := NewLLMInvocation("gpt-4")
	inv.InputMessages = []InputMessage{{
		Role:  "user",
		Parts: []MessagePart{Text{Content: "Hello"}},
	}}
	ctx := h.StartLLM(context.Background(), inv)
	_ = ctx
	h.StopLLM(inv)

	events := rec.events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	attrs := attrMap(events[0])
	if _, ok := attrs[AttrGenAIInputMessages]; ok {
		t.Errorf("input messages should not be captured with NO_CONTENT")
	}
}

func TestEmitInvokeAgentEvent(t *testing.T) {
	t.Setenv(EnvSemconvStabilityOptIn, string(StabilityModeGenAILatestExperimental))
	t.Setenv(EnvEmitEvent, "true")

	h, rec := newRecordingHandler(t)

	inv := NewInvokeAgentInvocation()
	inv.AgentName = "assistant"
	inv.Provider = "openai"
	ctx := h.StartInvokeAgent(context.Background(), inv)
	_ = ctx
	h.StopInvokeAgent(inv)

	events := rec.events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if name := events[0].EventName(); name != EventGenAIAgentInvokeOperationDetails {
		t.Fatalf("unexpected event name: %q", name)
	}
	attrs := attrMap(events[0])
	if attrs[AttrGenAIAgentName] != "assistant" {
		t.Errorf("expected agent name attribute, got %q", attrs[AttrGenAIAgentName])
	}
}
