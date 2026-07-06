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

// Package main is a focused demo of instrumenting a *streaming* OpenAI chat
// completion with the util-genai module. It highlights the streaming-specific
// telemetry: gen_ai.request.stream and gen_ai.response.time_to_first_chunk,
// exported to stdout as both spans and metrics.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	utilgenai "github.com/alibaba/loongsuite-go/util-genai"
	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// initTelemetry wires up an OpenTelemetry TracerProvider and MeterProvider,
// both exporting to stdout so the streaming spans and metrics are visible.
func initTelemetry() (*sdktrace.TracerProvider, *metric.MeterProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("genai-stream-demo"),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return tp, mp, nil
}

func main() {
	tp, mp, err := initTelemetry()
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}
	client := openai.NewClient(apiKey)

	// Build a handler that emits both spans and metrics for this invocation.
	handler := utilgenai.NewTelemetryHandler(
		utilgenai.WithTracerProvider(tp),
		utilgenai.WithMeterProvider(mp),
	)

	fmt.Println("=== Streaming Chat Completion Demo ===")
	if err := streamChat(context.Background(), client, handler); err != nil {
		log.Fatalf("streaming demo failed: %v", err)
	}
	fmt.Println("\n=== Demo completed. Spans and metrics printed above. ===")
}

// streamChat instruments a streaming OpenAI chat completion. It records the
// streaming flag and the time-to-first-chunk, prints tokens as they arrive,
// and finalizes the span once the stream is fully drained.
func streamChat(ctx context.Context, client *openai.Client, handler *utilgenai.TelemetryHandler) error {
	const model = openai.GPT4oMini
	const prompt = "Count from 1 to 5 with a brief description of each number."

	// Describe the invocation. Setting Stream marks gen_ai.request.stream=true.
	invocation := utilgenai.NewLLMInvocation(model)
	invocation.Provider = "openai"
	invocation.OperationName = utilgenai.OperationChat
	streaming := true
	invocation.Stream = &streaming
	invocation.ConversationID = "conv_stream_demo"
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role:  "user",
			Parts: []utilgenai.MessagePart{utilgenai.Text{Content: prompt}},
		},
	}

	// Open the span before the network call so latency is captured accurately.
	ctx = handler.StartLLM(ctx, invocation)

	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
		// Ask the API to include usage stats in the final streamed chunk.
		StreamOptions: &openai.StreamOptions{IncludeUsage: true},
	})
	if err != nil {
		handler.FailLLM(invocation, &utilgenai.Error{Message: err.Error(), Type: "APIError"})
		return fmt.Errorf("create stream: %w", err)
	}
	defer stream.Close()

	var (
		fullContent  string
		finishReason openai.FinishReason
		firstChunk   = true
		streamStart  = time.Now()
	)

	fmt.Print("Streamed response: ")
	for {
		resp, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			handler.FailLLM(invocation, &utilgenai.Error{Message: recvErr.Error(), Type: "StreamError"})
			return fmt.Errorf("stream recv: %w", recvErr)
		}

		// Record time-to-first-chunk once, on the first token we see.
		if firstChunk {
			ttfc := time.Since(streamStart).Seconds()
			invocation.TimeToFirstChunk = &ttfc
			firstChunk = false
		}

		// The usage-only final chunk carries no choices.
		if resp.Usage != nil {
			inTok := resp.Usage.PromptTokens
			outTok := resp.Usage.CompletionTokens
			invocation.InputTokens = &inTok
			invocation.OutputTokens = &outTok
		}
		if len(resp.Choices) > 0 {
			delta := resp.Choices[0].Delta.Content
			fullContent += delta
			fmt.Print(delta) // print token-by-token to show real streaming
			if resp.Choices[0].FinishReason != "" {
				finishReason = resp.Choices[0].FinishReason
			}
		}
		if resp.ID != "" {
			invocation.ResponseID = resp.ID
		}
		if resp.Model != "" {
			invocation.ResponseModelName = resp.Model
		}
	}
	fmt.Println()

	// Populate the aggregated response and close the span successfully.
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role:         "assistant",
			Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: fullContent}},
			FinishReason: utilgenai.FinishReason(finishReason),
		},
	}
	handler.StopLLM(invocation)

	if invocation.TimeToFirstChunk != nil {
		fmt.Printf("Time to first chunk: %.3fs\n", *invocation.TimeToFirstChunk)
	}
	return nil
}
