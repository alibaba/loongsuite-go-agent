// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

// Package agentorchestrator instruments the AgentWrapper/agent-orchestrator
// daemon (github.com/aoagents/agent-orchestrator/backend), an agentic
// orchestrator that plans tasks, spawns parallel coding-agent sessions,
// routes CI/review feedback, and observes pull requests. The Session Manager
// is the single entry point for the spawn/send lifecycle, so each of those
// methods becomes a gen_ai workflow/task span whose attributes follow the
// ARMS semantic conventions for agentic systems.
//
// The instrumented types live under the daemon's internal/ tree, which is not
// importable from this plugin module. We therefore declare the receiver and
// config parameters as interface{} and read the fields we need via reflection
// (the trampoline passes the dereferenced values, which are assignable to
// interface{}).
package agentorchestrator

import (
	"context"
	"fmt"
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// tracerName mirrors the daemon's module path so the tracer name is stable
	// across builds and identifiable in ARMS.
	tracerName = "github.com/aoagents/agent-orchestrator/backend"

	// genAISystem is the ARMS gen_ai.system value for the agent-orchestrator
	// daemon. Per the gen-ai semantic conventions, gen_ai.system identifies the
	// agentic system that owns the span.
	genAISystem = "agent_orchestrator"

	spanNameSpawn = "create_agent"
	spanNameSend  = "send_message"
)

// stringField reads a string-typed (or named-string-typed) field from a value
// passed in as interface{}. The daemon's SpawnConfig and SessionRecord use
// distinct string-typed fields (ProjectID, IssueID, AgentHarness, SessionKind,
// SessionID), all of which have reflect.Kind == String, so reflect.Value.String
// returns the underlying value without a type assertion to a name we cannot
// import.
func stringField(v interface{}, name string) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	f := rv.FieldByName(name)
	if !f.IsValid() {
		return ""
	}
	if f.Kind() == reflect.String {
		return f.String()
	}
	return fmt.Sprintf("%v", f.Interface())
}

// stringOrZero converts a value passed as interface{} into its string form when
// the underlying kind is String (covering both plain strings and named string
// types such as domain.SessionID). It avoids a type assertion to a named type
// the plugin cannot import.
func stringOrZero(v interface{}) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.String {
		return rv.String()
	}
	return ""
}

//go:linkname agentOrchestratorSpawnOnEnter github.com/aoagents/agent-orchestrator/backend/internal/session_manager.agentOrchestratorSpawnOnEnter
func agentOrchestratorSpawnOnEnter(call api.CallContext, m interface{}, ctx context.Context, cfg interface{}) {
	_ = m
	harness := stringField(cfg, "Harness")
	kind := stringField(cfg, "Kind")
	projectID := stringField(cfg, "ProjectID")
	issueID := stringField(cfg, "IssueID")
	branch := stringField(cfg, "Branch")

	opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindInternal)}
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, spanNameSpawn, opts...)
	span.SetAttributes(
		attribute.String("gen_ai.system", genAISystem),
		attribute.String("gen_ai.operation.name", "create_agent"),
		attribute.String("gen_ai.span.kind", "workflow"),
		attribute.String("gen_ai.other_input.agent_harness", harness),
		attribute.String("gen_ai.other_input.session_kind", kind),
		attribute.String("gen_ai.other_input.project_id", projectID),
		attribute.String("gen_ai.other_input.issue_id", issueID),
		attribute.String("gen_ai.other_input.branch", branch),
	)

	data := make(map[string]interface{}, 1)
	data["span"] = span
	call.SetData(data)
	call.SetParam(1, spanCtx)
}

//go:linkname agentOrchestratorSpawnOnExit github.com/aoagents/agent-orchestrator/backend/internal/session_manager.agentOrchestratorSpawnOnExit
func agentOrchestratorSpawnOnExit(call api.CallContext, rec interface{}, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	span, _ := data["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}
	span.SetStatus(codes.Ok, "")
	if sessionID := stringField(rec, "ID"); sessionID != "" {
		span.SetAttributes(attribute.String("gen_ai.other_output.session_id", sessionID))
	}
}

//go:linkname agentOrchestratorSendOnEnter github.com/aoagents/agent-orchestrator/backend/internal/session_manager.agentOrchestratorSendOnEnter
func agentOrchestratorSendOnEnter(call api.CallContext, m interface{}, ctx context.Context, id interface{}, message string) {
	_ = m
	sessionID := stringOrZero(id)

	opts := []oteltrace.SpanStartOption{oteltrace.WithSpanKind(oteltrace.SpanKindClient)}
	spanCtx, span := otel.Tracer(tracerName).Start(ctx, spanNameSend, opts...)
	span.SetAttributes(
		attribute.String("gen_ai.system", genAISystem),
		attribute.String("gen_ai.operation.name", "send_message"),
		attribute.String("gen_ai.span.kind", "task"),
		attribute.String("gen_ai.other_input.session_id", sessionID),
		attribute.Int("gen_ai.other_input.message_length", len(message)),
	)

	data := make(map[string]interface{}, 1)
	data["span"] = span
	call.SetData(data)
	call.SetParam(1, spanCtx)
}

//go:linkname agentOrchestratorSendOnExit github.com/aoagents/agent-orchestrator/backend/internal/session_manager.agentOrchestratorSendOnExit
func agentOrchestratorSendOnExit(call api.CallContext, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	span, _ := data["span"].(oteltrace.Span)
	if span == nil {
		return
	}
	defer span.End()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return
	}
	span.SetStatus(codes.Ok, "")
}
