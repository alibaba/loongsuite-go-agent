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

package main

import (
	"context"
	"log"
	"strings"

	"github.com/aoagents/agent-orchestrator/backend/internal/session_manager"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	mgr := sessionmanager.New()

	// Success path: Spawn a real session, then Send it a message.
	rec, err := mgr.Spawn(context.Background(), sessionmanager.SpawnConfig{
		ProjectID: "proj-42",
		IssueID:   "issue-7",
		Kind:      sessionmanager.KindOrchestrator,
		Harness:   "claude-code",
		Branch:    "feat/cool-thing",
		Prompt:    "implement the feature",
	})
	if err != nil {
		log.Fatalf("spawn failed: %v", err)
	}

	if err := mgr.Send(context.Background(), rec.ID, "ping"); err != nil {
		log.Fatalf("send failed: %v", err)
	}

	// Error path: Spawn with empty Harness must fail and mark the span as Error.
	if _, err := mgr.Spawn(context.Background(), sessionmanager.SpawnConfig{
		ProjectID: "proj-42",
		Kind:      sessionmanager.KindOrchestrator,
		Harness:   "",
		Branch:    "feat/cool-thing",
		Prompt:    "implement the feature",
	}); err == nil {
		log.Fatalf("expected spawn error for empty harness")
	}

	// Error path: Send with empty SessionID must fail and mark the span as Error.
	if err := mgr.Send(context.Background(), "", "ping"); err == nil {
		log.Fatalf("expected send error for empty session id")
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		var spawnSuccess, spawnError, sendSuccess, sendError tracetest.SpanStub
		for _, traceStubs := range stubs {
			for _, s := range traceStubs {
				switch s.Name {
				case "create_agent":
					if s.Status.Code == codes.Error {
						spawnError = s
					} else {
						spawnSuccess = s
					}
				case "send_message":
					if s.Status.Code == codes.Error {
						sendError = s
					} else {
						sendSuccess = s
					}
				}
			}
		}

		verifier.Assert(spawnSuccess.Name == "create_agent",
			"expected spawn success span, got %s", spawnSuccess.Name)
		verifier.Assert(spawnSuccess.SpanKind == trace.SpanKindInternal,
			"spawn span kind should be internal, got %d", spawnSuccess.SpanKind)
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.system") == "agent_orchestrator",
			"spawn gen_ai.system mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.operation.name") == "create_agent",
			"spawn gen_ai.operation.name mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.span.kind") == "workflow",
			"spawn gen_ai.span.kind mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.other_input.agent_harness") == "claude-code",
			"spawn agent_harness mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.other_input.session_kind") == "orchestrator",
			"spawn session_kind mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.other_input.project_id") == "proj-42",
			"spawn project_id mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.other_input.issue_id") == "issue-7",
			"spawn issue_id mismatch")
		verifier.Assert(attrVal(spawnSuccess.Attributes, "gen_ai.other_input.branch") == "feat/cool-thing",
			"spawn branch mismatch")
		verifier.Assert(strings.HasPrefix(attrVal(spawnSuccess.Attributes, "gen_ai.other_output.session_id"), "sess-claude-code-"),
			"spawn session_id mismatch: %s", attrVal(spawnSuccess.Attributes, "gen_ai.other_output.session_id"))

		verifier.Assert(spawnError.Name == "create_agent",
			"expected spawn error span, got %s", spawnError.Name)
		verifier.Assert(spawnError.Status.Code == codes.Error,
			"spawn error span should have Error status, got %s", spawnError.Status.Code)

		verifier.Assert(sendSuccess.Name == "send_message",
			"expected send success span, got %s", sendSuccess.Name)
		verifier.Assert(sendSuccess.SpanKind == trace.SpanKindClient,
			"send span kind should be client, got %d", sendSuccess.SpanKind)
		verifier.Assert(attrVal(sendSuccess.Attributes, "gen_ai.system") == "agent_orchestrator",
			"send gen_ai.system mismatch")
		verifier.Assert(attrVal(sendSuccess.Attributes, "gen_ai.operation.name") == "send_message",
			"send gen_ai.operation.name mismatch")
		verifier.Assert(attrVal(sendSuccess.Attributes, "gen_ai.span.kind") == "task",
			"send gen_ai.span.kind mismatch")
		verifier.Assert(strings.HasPrefix(attrVal(sendSuccess.Attributes, "gen_ai.other_input.session_id"), "sess-claude-code-"),
			"send session_id mismatch")
		msgLen := verifier.GetAttribute(sendSuccess.Attributes, "gen_ai.other_input.message_length").AsInt64()
		verifier.Assert(msgLen == 4, "send message_length mismatch: got %d", msgLen)

		verifier.Assert(sendError.Name == "send_message",
			"expected send error span, got %s", sendError.Name)
		verifier.Assert(sendError.Status.Code == codes.Error,
			"send error span should have Error status, got %s", sendError.Status.Code)
	}, 1)
}

func attrVal(attrs []attribute.KeyValue, key string) string {
	for _, a := range attrs {
		if string(a.Key) == key {
			return a.Value.AsString()
		}
	}
	return ""
}
