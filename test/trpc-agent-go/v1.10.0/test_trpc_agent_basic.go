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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"

	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/runner"
)

// mockModel returns one canned assistant response and then a final done
// response, satisfying trpc-agent-go's flow contract without any external
// HTTP call.
type mockModel struct{}

func (m *mockModel) GenerateContent(ctx context.Context, _ *model.Request) (<-chan *model.Response, error) {
	ch := make(chan *model.Response, 2)
	finishReason := "stop"
	ch <- &model.Response{
		ID:     "resp_mock",
		Object: model.ObjectTypeChatCompletion,
		Model:  "mock-model",
		Choices: []model.Choice{
			{
				Index: 0,
				Message: model.Message{
					Role:    model.RoleAssistant,
					Content: "Hello from mock model",
				},
				FinishReason: &finishReason,
			},
		},
		Done: false,
	}
	ch <- &model.Response{
		ID:     "resp_mock_done",
		Object: model.ObjectTypeChatCompletion,
		Model:  "mock-model",
		Choices: []model.Choice{
			{
				Index:        0,
				FinishReason: &finishReason,
			},
		},
		Done: true,
	}
	close(ch)
	return ch, nil
}

func (m *mockModel) Info() model.Info {
	return model.Info{Name: "mock-model"}
}

func main() {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(100*time.Millisecond),
			trace.WithMaxExportBatchSize(10),
		),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	ctx := context.Background()

	ag := llmagent.New(
		"test-agent",
		llmagent.WithModel(&mockModel{}),
		llmagent.WithDescription("A test agent for trpc-agent-go instrumentation"),
		llmagent.WithInstruction("Reply with a greeting."),
	)

	r := runner.NewRunner("trpc-agent-go-test", ag)

	userID := "test-user"
	sessionID := "test-session"
	msg := model.NewUserMessage("Hello, say hi!")

	eventChan, err := r.Run(ctx, userID, sessionID, msg)
	if err != nil {
		log.Fatalf("runner.Run failed: %v", err)
	}
	for evt := range eventChan {
		if evt != nil && evt.Error != nil {
			log.Fatalf("agent error: %v", evt.Error)
		}
	}

	// Give the batch exporter a moment to flush.
	time.Sleep(500 * time.Millisecond)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		spanStr, _ := json.Marshal(stubs)
		fmt.Println(string(spanStr))

		foundWorkflow := false
		for _, spans := range stubs {
			for _, span := range spans {
				system := verifier.GetAttribute(span.Attributes, "gen_ai.system").AsString()
				if system != "trpc_agent_go" {
					continue
				}
				opName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
				if opName != "invoke_agent" {
					continue
				}
				foundWorkflow = true
				verifier.Assert(span.Name == "invoke_agent",
					"Expected span name invoke_agent, got %s", span.Name)
				spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
				verifier.Assert(spanKind == "workflow",
					"Expected gen_ai.span.kind=workflow, got %s", spanKind)
				uid := verifier.GetAttribute(span.Attributes, "gen_ai.other_input.user_id").AsString()
				verifier.Assert(uid == "test-user",
					"Expected gen_ai.other_input.user_id=test-user, got %s", uid)
				sid := verifier.GetAttribute(span.Attributes, "gen_ai.other_input.session_id").AsString()
				verifier.Assert(sid == "test-session",
					"Expected gen_ai.other_input.session_id=test-session, got %s", sid)
				userMsg := verifier.GetAttribute(span.Attributes, "gen_ai.other_input.user_message").AsString()
				verifier.Assert(userMsg == "Hello, say hi!",
					"Expected gen_ai.other_input.user_message=Hello, say hi!, got %s", userMsg)
				verifier.Assert(span.SpanKind == oteltrace.SpanKindClient,
					"Expected client span kind, got %d", span.SpanKind)
			}
		}
		verifier.Assert(foundWorkflow,
			"Expected to find trpc_agent_go invoke_agent workflow span")
	}, 1)
}
