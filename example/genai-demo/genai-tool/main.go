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

// Package main demonstrates how to instrument an OpenAI function-calling
// (tool-calling) round trip with the util-genai module. It shows the full
// loop: the model requests a tool call, the tool is executed under its own
// span, and the result is fed back to the model for a final answer.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	utilgenai "github.com/alibaba/loongsuite-go/util-genai"
	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const weatherToolName = "get_weather"

// weatherParametersSchema is the JSON Schema describing the tool's arguments.
var weatherParametersSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"location": map[string]any{
			"type":        "string",
			"description": "The city name, e.g. San Francisco",
		},
	},
	"required": []string{"location"},
}

// initTracer sets up an OpenTelemetry TracerProvider with a stdout exporter.
func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("genai-tool-demo"),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

// getWeather is the local implementation of the tool the model can call.
func getWeather(location string) string {
	return fmt.Sprintf("The weather in %s is sunny, 22°C.", location)
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(apiKey)
	handler := utilgenai.NewTelemetryHandler(utilgenai.WithTracerProvider(tp))

	ctx := context.Background()

	fmt.Println("=== GenAI Tool / Function Calling Demo ===")
	if err := runToolCalling(ctx, client, handler); err != nil {
		log.Fatalf("tool calling demo failed: %v", err)
	}
	fmt.Println("\n=== Demo completed. Traces printed above. ===")
}

func runToolCalling(ctx context.Context, client *openai.Client, handler *utilgenai.TelemetryHandler) error {
	// The tool advertised to the model, declared once and reused across calls.
	toolDefinition := utilgenai.FunctionToolDefinition{
		Name:        weatherToolName,
		Description: "Get the current weather for a given location.",
		Parameters:  weatherParametersSchema,
	}
	openAITools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        toolDefinition.Name,
				Description: toolDefinition.Description,
				Parameters:  toolDefinition.Parameters,
			},
		},
	}

	// Running message history shared by both model calls.
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "What's the weather like in San Francisco?"},
	}

	// --- Step 1: first LLM call, model decides to call the tool. ---
	toolCall, err := requestToolCall(ctx, client, handler, &messages, toolDefinition, openAITools)
	if err != nil {
		return err
	}
	if toolCall == nil {
		fmt.Println("Model answered directly without calling a tool.")
		return nil
	}

	// --- Step 2: execute the requested tool under its own span. ---
	toolResult, err := executeToolCall(ctx, handler, *toolCall)
	if err != nil {
		return err
	}

	// Append the tool result to the conversation so the model can use it.
	messages = append(messages, openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		ToolCallID: toolCall.ID,
		Content:    toolResult,
	})

	// --- Step 3: second LLM call, model produces the final answer. ---
	return finalAnswer(ctx, client, handler, messages, *toolCall, toolResult)
}

// requestToolCall performs the first LLM call and returns the tool call the
// model requested (or nil if it answered directly).
func requestToolCall(
	ctx context.Context,
	client *openai.Client,
	handler *utilgenai.TelemetryHandler,
	messages *[]openai.ChatCompletionMessage,
	toolDefinition utilgenai.FunctionToolDefinition,
	openAITools []openai.Tool,
) (*openai.ToolCall, error) {
	invocation := utilgenai.NewLLMInvocation("gpt-4o-mini")
	invocation.Provider = "openai"
	invocation.OperationName = utilgenai.OperationChat
	invocation.ToolDefinitions = []utilgenai.FunctionToolDefinition{toolDefinition}
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role:  "user",
			Parts: []utilgenai.MessagePart{utilgenai.Text{Content: (*messages)[0].Content}},
		},
	}

	ctx = handler.StartLLM(ctx, invocation)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: *messages,
		Tools:    openAITools,
	})
	if err != nil {
		handler.FailLLM(invocation, &utilgenai.Error{Message: err.Error(), Type: "APIError"})
		return nil, err
	}

	choice := resp.Choices[0]
	// Keep the assistant message (with its tool_calls) in the history.
	*messages = append(*messages, choice.Message)

	inputTokens := resp.Usage.PromptTokens
	outputTokens := resp.Usage.CompletionTokens
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens
	invocation.ResponseModelName = resp.Model
	invocation.ResponseID = resp.ID

	// Record the model's tool call as an output message part.
	if len(choice.Message.ToolCalls) > 0 {
		tc := choice.Message.ToolCalls[0]
		var args any
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		invocation.OutputMessages = []utilgenai.OutputMessage{
			{
				Role: "assistant",
				Parts: []utilgenai.MessagePart{
					utilgenai.ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: args},
				},
				FinishReason: utilgenai.FinishReasonToolCalls,
			},
		}
		handler.StopLLM(invocation)
		fmt.Printf("Model requested tool %q with arguments: %s\n", tc.Function.Name, tc.Function.Arguments)
		return &tc, nil
	}

	// No tool call: the model answered directly.
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role:         "assistant",
			Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: choice.Message.Content}},
			FinishReason: utilgenai.FinishReason(choice.FinishReason),
		},
	}
	handler.StopLLM(invocation)
	return nil, nil
}

// executeToolCall runs the requested tool inside a dedicated execute_tool span.
func executeToolCall(ctx context.Context, handler *utilgenai.TelemetryHandler, tc openai.ToolCall) (string, error) {
	var args struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	invocation := utilgenai.NewExecuteToolInvocation(tc.Function.Name)
	invocation.ToolCallID = tc.ID
	invocation.ToolType = "function"
	invocation.Input = map[string]any{"location": args.Location}

	ctx = handler.StartExecuteTool(ctx, invocation)
	_ = ctx

	result := getWeather(args.Location)

	invocation.Output = result
	handler.StopExecuteTool(invocation)

	fmt.Printf("Executed tool %q -> %s\n", tc.Function.Name, result)
	return result, nil
}

// finalAnswer performs the second LLM call, feeding the tool result back in.
func finalAnswer(
	ctx context.Context,
	client *openai.Client,
	handler *utilgenai.TelemetryHandler,
	messages []openai.ChatCompletionMessage,
	tc openai.ToolCall,
	toolResult string,
) error {
	invocation := utilgenai.NewLLMInvocation("gpt-4o-mini")
	invocation.Provider = "openai"
	invocation.OperationName = utilgenai.OperationChat
	// The tool result is the input for this turn.
	invocation.InputMessages = []utilgenai.InputMessage{
		{
			Role:  "tool",
			Parts: []utilgenai.MessagePart{utilgenai.ToolCallResponse{ID: tc.ID, Response: toolResult}},
		},
	}

	ctx = handler.StartLLM(ctx, invocation)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: messages,
	})
	if err != nil {
		handler.FailLLM(invocation, &utilgenai.Error{Message: err.Error(), Type: "APIError"})
		return err
	}

	choice := resp.Choices[0]
	inputTokens := resp.Usage.PromptTokens
	outputTokens := resp.Usage.CompletionTokens
	invocation.InputTokens = &inputTokens
	invocation.OutputTokens = &outputTokens
	invocation.ResponseModelName = resp.Model
	invocation.ResponseID = resp.ID
	invocation.OutputMessages = []utilgenai.OutputMessage{
		{
			Role:         "assistant",
			Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: choice.Message.Content}},
			FinishReason: utilgenai.FinishReason(choice.FinishReason),
		},
	}
	handler.StopLLM(invocation)

	fmt.Printf("Final answer: %s\n", choice.Message.Content)
	return nil
}
