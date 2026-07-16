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
	"fmt"
	"iter"
	"log"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/microsoft/agent-framework-go/agent"
	"github.com/microsoft/agent-framework-go/message"
	"github.com/microsoft/agent-framework-go/tool"
	"github.com/microsoft/agent-framework-go/tool/functool"
	"github.com/microsoft/agent-framework-go/workflow"
	"github.com/microsoft/agent-framework-go/workflow/inproc"
)

// mockProvider is a minimal agent.RunFunc that yields one canned assistant
// text update with finish_reason=stop, exercising the (*Agent).Run hook
// without any external HTTP call.
func mockProvider(_ context.Context, _ []*message.Message, _ ...agent.Option) iter.Seq2[*agent.ResponseUpdate, error] {
	return func(yield func(*agent.ResponseUpdate, error) bool) {
		yield(&agent.ResponseUpdate{
			Role: message.RoleAssistant,
			Contents: message.Contents{
				&message.TextContent{Text: "Hello from mock provider"},
			},
			FinishReason: "stop",
		}, nil)
	}
}

func newAgent() *agent.Agent {
	return agent.New(
		agent.ProviderConfig{
			ProviderName: "mock",
			Run:          mockProvider,
		},
		agent.Config{
			ID:          "test-agent-id",
			Name:        "test-agent",
			Description: "A test agent for ms-agent-framework-go instrumentation",
		},
	)
}

type echoArgs struct {
	Text string
}

func newEchoTool() tool.FuncTool {
	t, err := functool.New[echoArgs, string](
		functool.Config{
			Name:        "echo",
			Description: "Echoes the provided argument back to the caller.",
		},
		functool.HandlerFor[echoArgs, string](func(_ context.Context, args echoArgs) (string, error) {
			return args.Text, nil
		}),
	)
	if err != nil {
		log.Fatalf("functool.New: %v", err)
	}
	return t
}

// noOpExecutorBinding mirrors the workflow_test scaffolding so we can build
// a minimal workflow without dragging in the framework's test helpers.
type noOpExecutor struct {
	id string
}

func (n *noOpExecutor) NewExecutor(_ string) (*workflow.Executor, error) {
	return &workflow.Executor{
		ID: n.id,
		ConfigureProtocol: func(rb *workflow.ProtocolBuilder) (*workflow.ProtocolBuilder, error) {
			rb.RouteBuilder.AddCatchAll(func(_ *workflow.Context, msg workflow.PortableValue) (any, error) {
				return nil, nil
			})
			return rb, nil
		},
	}, nil
}

func newNoOpBinding(id string) workflow.ExecutorBinding {
	n := &noOpExecutor{id: id}
	return workflow.ExecutorBinding{
		ID:               id,
		ImplementationID: "*noOpExecutor",
		NewExecutorFunc:  n.NewExecutor,
	}
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

	// --- Scenario 1: agent.RunText drives the agent workflow span. ---
	a := newAgent()
	for upd, err := range a.RunText(ctx, "Hello, say hi!") {
		if err != nil {
			log.Fatalf("agent.RunText: %v", err)
		}
		_ = upd
	}

	// --- Scenario 2: (*funcTool).Call drives the tool span. ---
	echo := newEchoTool()
	if _, err := echo.Call(ctx, `{"Text":"hi"}`); err != nil {
		log.Fatalf("funcTool.Call: %v", err)
	}

	// --- Scenario 3: (*ExecutionEnvironment).Run drives the workflow span. ---
	wf, err := workflow.NewBuilder(newNoOpBinding("start")).Build()
	if err != nil {
		log.Fatalf("workflow.Build: %v", err)
	}
	if _, err := inproc.Default.Run(ctx, wf, "input"); err != nil {
		log.Fatalf("inproc.Default.Run: %v", err)
	}

	// Give the batch exporter a moment to flush.
	time.Sleep(500 * time.Millisecond)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		foundAgent := false
		foundTool := false
		foundWorkflow := false

		for _, spans := range stubs {
			for _, span := range spans {
				system := verifier.GetAttribute(span.Attributes, "gen_ai.system").AsString()
				if system != "microsoft_agent_framework_go" {
					continue
				}
				opName := verifier.GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()

				switch opName {
				case "invoke_agent":
					foundAgent = true
					verifier.Assert(span.Name == "invoke_agent",
						"Expected span name invoke_agent, got %s", span.Name)
					spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
					verifier.Assert(spanKind == "workflow",
						"Expected gen_ai.span.kind=workflow, got %s", spanKind)
					userMsg := verifier.GetAttribute(span.Attributes, "gen_ai.other_input.user_message").AsString()
					verifier.Assert(userMsg == "Hello, say hi!",
						"Expected gen_ai.other_input.user_message=Hello, say hi!, got %s", userMsg)
					verifier.Assert(span.SpanKind == oteltrace.SpanKindClient,
						"Expected client span kind, got %d", span.SpanKind)

				case "execute_tool":
					foundTool = true
					verifier.Assert(span.Name == "execute_tool",
						"Expected span name execute_tool, got %s", span.Name)
					spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
					verifier.Assert(spanKind == "tool",
						"Expected gen_ai.span.kind=tool, got %s", spanKind)
					toolName := verifier.GetAttribute(span.Attributes, "gen_ai.tool.name").AsString()
					verifier.Assert(toolName == "echo",
						"Expected gen_ai.tool.name=echo, got %s", toolName)
					toolInput := verifier.GetAttribute(span.Attributes, "gen_ai.tool.input").AsString()
					verifier.Assert(toolInput == `{"Text":"hi"}`,
						"Expected gen_ai.tool.input={\"Text\":\"hi\"}, got %s", toolInput)
					toolOutput := verifier.GetAttribute(span.Attributes, "gen_ai.tool.output").AsString()
					verifier.Assert(toolOutput == "hi",
						"Expected gen_ai.tool.output=hi, got %s", toolOutput)
					verifier.Assert(span.SpanKind == oteltrace.SpanKindClient,
						"Expected client span kind, got %d", span.SpanKind)

				case "run_workflow":
					foundWorkflow = true
					verifier.Assert(span.Name == "run_workflow",
						"Expected span name run_workflow, got %s", span.Name)
					spanKind := verifier.GetAttribute(span.Attributes, "gen_ai.span.kind").AsString()
					verifier.Assert(spanKind == "workflow",
						"Expected gen_ai.span.kind=workflow, got %s", spanKind)
					verifier.Assert(span.SpanKind == oteltrace.SpanKindClient,
						"Expected client span kind, got %d", span.SpanKind)
				}
			}
		}

		verifier.Assert(foundAgent, "Expected invoke_agent workflow span")
		verifier.Assert(foundTool, "Expected execute_tool span")
		verifier.Assert(foundWorkflow, "Expected run_workflow span")
	}, 1)

	fmt.Println("ms-agent-framework-go instrumentation test passed")
}
