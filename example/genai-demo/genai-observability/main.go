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

// Package main is a self-contained demo of two util-genai capabilities that do
// not require any external API key:
//
//   - Log-based event emission (gen_ai.client.inference.operation.details),
//     printed to stdout via the OpenTelemetry log SDK.
//   - Prompt/response content offloading, where message content is written to
//     local files and only a content-addressed reference attribute is stamped
//     on the span.
//
// It fabricates an LLM invocation instead of calling a real model so it can be
// run offline.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	utilgenai "github.com/alibaba/loongsuite-go/util-genai"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func main() {
	// Enable experimental GenAI semconv, event emission, and content capture.
	// These are read from the environment by util-genai at finish time.
	os.Setenv(utilgenai.EnvSemconvStabilityOptIn, string(utilgenai.StabilityModeGenAILatestExperimental))
	os.Setenv(utilgenai.EnvEmitEvent, "true")
	os.Setenv(utilgenai.EnvCaptureMessageContent, "SPAN_AND_EVENT")

	// Offload prompt/response content to a temp directory. The handler will
	// auto-create an upload hook because this base path is set.
	uploadDir, err := os.MkdirTemp("", "genai-upload-*")
	if err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}
	os.Setenv(utilgenai.EnvUploadBasePath, uploadDir)
	fmt.Printf("Uploading message content to: %s\n\n", uploadDir)

	ctx := context.Background()

	// Trace exporter -> stdout.
	res, _ := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String("genai-observability-demo"),
	))
	traceExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(traceExp),
		sdktrace.WithResource(res),
	)
	defer func() { _ = tp.Shutdown(ctx) }()

	// Log exporter -> stdout (carries the emitted GenAI events).
	logExp, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to create log exporter: %v", err)
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(logExp)),
		sdklog.WithResource(res),
	)
	defer func() { _ = lp.Shutdown(ctx) }()

	handler := utilgenai.NewTelemetryHandler(
		utilgenai.WithTracerProvider(tp),
		utilgenai.WithLoggerProvider(lp),
	)

	runInvocation(ctx, handler)

	// Flush pending uploads before inspecting the directory.
	if err := handler.Shutdown(ctx); err != nil {
		log.Printf("handler shutdown: %v", err)
	}

	fmt.Println("\n=== Uploaded content files ===")
	entries, _ := os.ReadDir(uploadDir)
	for _, e := range entries {
		fmt.Printf("- %s\n", filepath.Join(uploadDir, e.Name()))
	}
	fmt.Println("\nDone. See the span (with *_ref attributes) and the log event above.")
}

// runInvocation fabricates a complete LLM invocation lifecycle.
func runInvocation(ctx context.Context, handler *utilgenai.TelemetryHandler) {
	inv := utilgenai.NewLLMInvocation("gpt-4o-mini")
	inv.Provider = "openai"
	inv.OperationName = utilgenai.OperationChat
	inv.SystemInstruction = []utilgenai.MessagePart{
		utilgenai.Text{Content: "You are a concise assistant."},
	}
	inv.InputMessages = []utilgenai.InputMessage{
		{
			Role:  "user",
			Parts: []utilgenai.MessagePart{utilgenai.Text{Content: "What is OpenTelemetry?"}},
		},
	}
	inv.ToolDefinitions = []utilgenai.FunctionToolDefinition{
		{Name: "search_docs", Description: "Search the documentation"},
	}

	ctx = handler.StartLLM(ctx, inv)
	_ = ctx

	inv.OutputMessages = []utilgenai.OutputMessage{
		{
			Role:         "assistant",
			Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: "An observability framework."}},
			FinishReason: utilgenai.FinishReasonStop,
		},
	}
	inputTokens, outputTokens := 12, 5
	inv.InputTokens = &inputTokens
	inv.OutputTokens = &outputTokens

	handler.StopLLM(inv)
}
