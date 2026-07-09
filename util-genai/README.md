# OpenTelemetry Util for GenAI (Go)

[English](./README.md) | [中文](./README_CN.md)

This package provides OpenTelemetry utilities for GenAI instrumentation in Go. It is a port of the Python [opentelemetry-util-genai](https://github.com/open-telemetry/opentelemetry-python-contrib/tree/main/util/opentelemetry-util-genai) package.

## Overview

The GenAI Utils package includes boilerplate and helpers to standardize instrumentation for Generative AI. This package provides APIs and types to minimize the work needed to instrument GenAI libraries, while providing standardization for generating both types of OpenTelemetry data: "spans and metrics" and "spans, metrics and events".

## Installation

```bash
go get github.com/alibaba/loongsuite-go/util-genai
```

## Environment Variables

This package relies on environment variables to configure capturing of message content. By default, message content will not be captured.

| Variable | Description | Values |
|----------|-------------|--------|
| `OTEL_SEMCONV_STABILITY_OPT_IN` | Enable experimental features | `gen_ai_latest_experimental` |
| `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT` | Control message content capture | `NO_CONTENT`, `SPAN_ONLY`, `EVENT_ONLY`, `SPAN_AND_EVENT` |
| `OTEL_INSTRUMENTATION_GENAI_EMIT_EVENT` | Control event emission | `true`, `false` |

## Span Attributes

This package provides these span attributes following the [OpenTelemetry GenAI Semantic Conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/gen-ai/README.md):

- `gen_ai.provider.name`: Provider name (e.g., "openai")
- `gen_ai.operation.name`: Operation name (e.g., "chat")
- `gen_ai.request.model`: Request model name
- `gen_ai.response.finish_reasons`: List of finish reasons
- `gen_ai.response.model`: Response model name
- `gen_ai.response.id`: Response ID
- `gen_ai.usage.input_tokens`: Input token count
- `gen_ai.usage.output_tokens`: Output token count
- `gen_ai.input.messages`: Input messages (when content capturing is enabled)
- `gen_ai.output.messages`: Output messages (when content capturing is enabled)
- `gen_ai.system_instructions`: System instructions (when provided)

## Usage

### Basic LLM Invocation

```go
package main

import (
    "context"
    
    utilgenai "github.com/alibaba/loongsuite-go/util-genai"
)

func main() {
    handler := utilgenai.GetTelemetryHandler()
    ctx := context.Background()

    // Create an invocation object with your request data
    invocation := utilgenai.NewLLMInvocation("gpt-4")
    invocation.Provider = "openai"
    invocation.InputMessages = []utilgenai.InputMessage{
        {
            Role: "user",
            Parts: []utilgenai.MessagePart{
                utilgenai.Text{Content: "Hello, world!"},
            },
        },
    }

    // Start the invocation (opens a span)
    ctx = handler.StartLLM(ctx, invocation)

    // Make the actual LLM call
    // response, err := client.Chat(ctx, request)

    // Populate outputs
    invocation.OutputMessages = []utilgenai.OutputMessage{
        {
            Role: "assistant",
            Parts: []utilgenai.MessagePart{
                utilgenai.Text{Content: "Hello! How can I help you today?"},
            },
            FinishReason: utilgenai.FinishReasonStop,
        },
    }
    inputTokens := 10
    outputTokens := 8
    invocation.InputTokens = &inputTokens
    invocation.OutputTokens = &outputTokens

    // Stop the invocation (closes the span)
    handler.StopLLM(invocation)
}
```

### Handling Errors

```go
func callLLM(ctx context.Context) error {
    handler := utilgenai.GetTelemetryHandler()

    invocation := utilgenai.NewLLMInvocation("gpt-4")
    invocation.Provider = "openai"

    ctx = handler.StartLLM(ctx, invocation)

    response, err := client.Chat(ctx, request)
    if err != nil {
        handler.FailLLM(invocation, &utilgenai.Error{
            Message: err.Error(),
            Type:    "APIError",
        })
        return err
    }

    // Populate outputs and stop
    // ...
    handler.StopLLM(invocation)
    return nil
}
```

### Streaming Mode

For streaming LLM responses, set `invocation.Stream = true` to mark the request
as streaming (`gen_ai.request.stream`), and record the time to receive the first
chunk in `invocation.TimeToFirstChunk` (`gen_ai.response.time_to_first_chunk`).
Start the span *before* the network call so latency is measured accurately, drain
the stream while accumulating the output, then call `StopLLM` once the stream is
fully consumed.

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewLLMInvocation("gpt-4o-mini")
invocation.Provider = "openai"
streaming := true
invocation.Stream = &streaming
invocation.InputMessages = []utilgenai.InputMessage{
    {
        Role:  "user",
        Parts: []utilgenai.MessagePart{utilgenai.Text{Content: "Count from 1 to 5."}},
    },
}

// Open the span before the network call so latency is captured accurately.
ctx = handler.StartLLM(ctx, invocation)

stream, err := client.CreateChatCompletionStream(ctx, request)
if err != nil {
    handler.FailLLM(invocation, &utilgenai.Error{Message: err.Error(), Type: "APIError"})
    return err
}
defer stream.Close()

var (
    fullContent string
    firstChunk  = true
    streamStart = time.Now()
)

for {
    resp, recvErr := stream.Recv()
    if errors.Is(recvErr, io.EOF) {
        break
    }
    if recvErr != nil {
        handler.FailLLM(invocation, &utilgenai.Error{Message: recvErr.Error(), Type: "StreamError"})
        return recvErr
    }

    // Record time-to-first-chunk once, on the first chunk received.
    if firstChunk {
        ttfc := time.Since(streamStart).Seconds()
        invocation.TimeToFirstChunk = &ttfc
        firstChunk = false
    }

    // The usage-only final chunk carries token counts (StreamOptions.IncludeUsage).
    if resp.Usage != nil {
        inTok := resp.Usage.PromptTokens
        outTok := resp.Usage.CompletionTokens
        invocation.InputTokens = &inTok
        invocation.OutputTokens = &outTok
    }
    if len(resp.Choices) > 0 {
        fullContent += resp.Choices[0].Delta.Content
    }
}

// Populate the aggregated response and close the span successfully.
invocation.OutputMessages = []utilgenai.OutputMessage{
    {
        Role:         "assistant",
        Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: fullContent}},
        FinishReason: utilgenai.FinishReasonStop,
    },
}
handler.StopLLM(invocation)
```

When streaming is enabled, this package emits the following additional telemetry:

- Span attribute `gen_ai.request.stream`: `true`
- Span attribute `gen_ai.response.time_to_first_chunk`: seconds until the first chunk
- Metric `gen_ai.client.operation.time_to_first_chunk`: time-to-first-chunk histogram

A complete runnable example is available under
[`example/genai-stream`](../example/genai-stream).

### Embedding Invocation

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewEmbeddingInvocation("text-embedding-3-small")
invocation.Provider = "openai"
inputCount := 5
invocation.InputCount = &inputCount

ctx = handler.StartEmbedding(ctx, invocation)

// Make the embedding call
// ...

inputTokens := 100
invocation.InputTokens = &inputTokens
handler.StopEmbedding(invocation)
```

### Tool Execution

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewExecuteToolInvocation("get_weather")
invocation.ToolCallID = "call_123"
invocation.Input = map[string]any{"location": "San Francisco"}

ctx = handler.StartExecuteTool(ctx, invocation)

// Execute the tool
result := getWeather("San Francisco")
invocation.Output = result

handler.StopExecuteTool(invocation)
```

### Agent Invocation

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewInvokeAgentInvocation()
invocation.AgentName = "assistant"
invocation.Provider = "openai"

ctx = handler.StartInvokeAgent(ctx, invocation)

// Run the agent
// ...

handler.StopInvokeAgent(invocation)
```

### Custom Provider Configuration

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/trace"
)

// Create custom tracer provider
tp := trace.NewTracerProvider(...)
otel.SetTracerProvider(tp)

// Create handler with custom providers
handler := utilgenai.NewTelemetryHandler(
    utilgenai.WithTracerProvider(tp),
    utilgenai.WithMeterProvider(mp),
)
```

## Supported Operations

| Operation | Handler Methods |
|-----------|-----------------|
| LLM/Chat | `StartLLM`, `StopLLM`, `FailLLM` |
| Embeddings | `StartEmbedding`, `StopEmbedding`, `FailEmbedding` |
| Tool Execution | `StartExecuteTool`, `StopExecuteTool`, `FailExecuteTool` |
| Agent Invocation | `StartInvokeAgent`, `StopInvokeAgent`, `FailInvokeAgent` |
| Agent Creation | `StartCreateAgent`, `StopCreateAgent`, `FailCreateAgent` |
| Document Retrieval | `StartRetrieve`, `StopRetrieve`, `FailRetrieve` |
| Document Reranking | `StartRerank`, `StopRerank`, `FailRerank` |

## Metrics

This package automatically records the following metrics:

- `gen_ai.client.operation.duration`: Duration of GenAI client operations (histogram)
- `gen_ai.client.token.usage`: Token usage for input and output (histogram)

## Examples

Complete runnable examples are available:

- [`example/genai`](../example/genai): a comprehensive demo covering chat completion, streaming chat completion, and embedding instrumentation.
- [`example/genai-stream`](../example/genai-stream): a streaming-focused demo highlighting streaming-specific telemetry (`gen_ai.request.stream`, `gen_ai.response.time_to_first_chunk`), exporting both spans and metrics to stdout.

## References

- [OpenTelemetry GenAI Semantic Conventions](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/gen-ai/README.md)
- [Python opentelemetry-util-genai](https://github.com/open-telemetry/opentelemetry-python-contrib/tree/main/util/opentelemetry-util-genai)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)

## License

Apache License 2.0
