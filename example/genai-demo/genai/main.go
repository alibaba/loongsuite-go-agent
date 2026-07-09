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

// Package main demonstrates how to use the util-genai module to instrument
// OpenAI API calls with OpenTelemetry. It covers chat completion, streaming
// chat completion, and embedding operations.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	utilgenai "github.com/alibaba/loongsuite-go/util-genai"
	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// initTracer sets up an OpenTelemetry TracerProvider with a stdout exporter.
// All spans will be printed to stdout as JSON for demonstration purposes.
func initTracer() (*sdktrace.TracerProvider, error) {
	// Create a stdout exporter that prints spans as JSON
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
	}

	// Create a resource describing this application
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("genai-demo"),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create a TracerProvider with the exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set as the global TracerProvider
	otel.SetTracerProvider(tp)

	return tp, nil
}

func main() {
	// Step 1: Initialize OpenTelemetry
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Step 2: Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	// Step 3: Create the TelemetryHandler
	// The handler uses the global TracerProvider we set above
	handler := utilgenai.NewTelemetryHandler(
		utilgenai.WithTracerProvider(tp),
	)

	ctx := context.Background()

	// Demo 1: Chat Completion with instrumentation
	fmt.Println("=== Demo 1: Chat Completion ===")
	chatCompletion(ctx, client, handler)

	// Demo 2: Streaming Chat Completion with instrumentation
	fmt.Println("\n=== Demo 2: Streaming Chat Completion ===")
	streamingChatCompletion(ctx, client, handler)

	// Demo 3: Embedding with instrumentation
	fmt.Println("\n=== Demo 3: Embedding ===")
	embedding(ctx, client, handler)

	fmt.Println("\n=== All demos completed. Traces printed above. ===")
}

// chatCompletion demonstrates instrumenting a standard OpenAI chat completion request.
func chatCompletion(ctx context.Context, client *openai.Client, handler *utilgenai.TelemetryHandler) {
	// Create an LLM invocation to track this call
	invocation := utilgenai.NewLLMInvocation("gpt-4o-mini")
	invocation.Provider = "openai"
	invocation.OperationName = utilgenai.OperationChat
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "system",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "You are a helpful assistant that explains things concisely."},
			},
		},
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "What is OpenTelemetry in one sentence?"},
			},
		},
	}

	// Start the instrumentation span before making the API call
	ctx = handler.StartLLM(ctx, invocation)

	// Make the actual OpenAI API call
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant that explains things concisely."},
			{Role: openai.ChatMessageRoleUser, Content: "What is OpenTelemetry in one sentence?"},
		},
	})

	if err != nil {
		// On error, record the failure in the span
		handler.FailLLM(invocation, &utilgenai.Error{
			Message: err.Error(),
			Type:    "APIError",
		})
		fmt.Printf("Chat completion failed: %v\n", err)
		return
	}

	// On success, populate response data and finalize the span
	if len(resp.Choices) > 0 {
		invocation.OutputMessages = []utilgenai.OutputMessage{
			{
				Role: "assistant",
				Parts: []utilgenai.MessagePart{
					utilgenai.Text{Content: resp.Choices[0].Message.Content},
				},
				FinishReason: utilgenai.FinishReason(resp.Choices[0].FinishReason),
			},
		}
	}
	inputTokens := resp.Usage.PromptTokens
	outputTokens := resp.Usage.CompletionTokens
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens
	invocation.ResponseModelName = resp.Model
	invocation.ResponseID = resp.ID

	// Finalize the span with success
	handler.StopLLM(invocation)

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
}

// streamingChatCompletion demonstrates instrumenting a streaming OpenAI chat completion.
func streamingChatCompletion(ctx context.Context, client *openai.Client, handler *utilgenai.TelemetryHandler) {
	// Create an LLM invocation for streaming
	invocation := utilgenai.NewLLMInvocation("gpt-4o-mini")
	invocation.Provider = "openai"
	invocation.OperationName = utilgenai.OperationChat
	// Per spec: set gen_ai.request.stream when the request is streaming
	isStreaming := true
	invocation.Stream = &isStreaming
	// Demonstrate conversation tracking with gen_ai.conversation.id
	invocation.ConversationID = "conv_demo_streaming_session"
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role: "user",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: "Count from 1 to 5 with a brief description of each number."},
			},
		},
	}

	// Start the instrumentation span
	ctx = handler.StartLLM(ctx, invocation)

	// Create the streaming request
	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: "Count from 1 to 5 with a brief description of each number."},
		},
	})
	if err != nil {
		handler.FailLLM(invocation, &utilgenai.Error{
			Message: err.Error(),
			Type:    "APIError",
		})
		fmt.Printf("Stream creation failed: %v\n", err)
		return
	}
	defer stream.Close()

	// Collect the streamed response, measuring time to first chunk
	var fullContent string
	var finishReason openai.FinishReason
	firstChunk := true
	streamStart := time.Now()

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			handler.FailLLM(invocation, &utilgenai.Error{
				Message: err.Error(),
				Type:    "StreamError",
			})
			fmt.Printf("Stream receive failed: %v\n", err)
			return
		}

		// Record time_to_first_chunk per spec
		if firstChunk {
			ttfc := time.Since(streamStart).Seconds()
			invocation.TimeToFirstChunk = &ttfc
			firstChunk = false
		}

		if len(response.Choices) > 0 {
			fullContent += response.Choices[0].Delta.Content
			if response.Choices[0].FinishReason != "" {
				finishReason = response.Choices[0].FinishReason
			}
		}
	}

	// Populate the invocation with the complete response
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role: "assistant",
			Parts: []utilgenai.MessagePart{
				utilgenai.Text{Content: fullContent},
			},
			FinishReason: utilgenai.FinishReason(finishReason),
		},
	}
	invocation.ResponseModelName = "gpt-4o-mini"

	// Finalize the span
	handler.StopLLM(invocation)

	fmt.Printf("Streamed response: %s\n", fullContent)
}

// embedding demonstrates instrumenting an OpenAI embedding request.
func embedding(ctx context.Context, client *openai.Client, handler *utilgenai.TelemetryHandler) {
	// Create an embedding invocation
	invocation := utilgenai.NewEmbeddingInvocation("text-embedding-3-small")
	invocation.Provider = "openai"
	inputCount := 2
	invocation.InputCount = &inputCount

	// Start the embedding instrumentation span
	ctx = handler.StartEmbedding(ctx, invocation)

	// Make the actual embedding API call
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3,
		Input: []string{
			"OpenTelemetry is an observability framework for cloud-native software.",
			"Go is a statically typed, compiled programming language.",
		},
	})

	if err != nil {
		// On error, record the failure
		handler.FailEmbedding(invocation, &utilgenai.Error{
			Message: err.Error(),
			Type:    "APIError",
		})
		fmt.Printf("Embedding failed: %v\n", err)
		return
	}

	// Populate response metadata
	inputTokens := resp.Usage.PromptTokens
	invocation.InputTokens = &inputTokens
	if len(resp.Data) > 0 {
		dims := len(resp.Data[0].Embedding)
		invocation.DimensionCount = &dims
	}

	// Finalize the span
	handler.StopEmbedding(invocation)

	fmt.Printf("Generated %d embeddings, dimensions: %d, tokens used: %d\n",
		len(resp.Data),
		len(resp.Data[0].Embedding),
		resp.Usage.PromptTokens,
	)
}
