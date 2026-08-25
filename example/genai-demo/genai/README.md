# GenAI Instrumentation Demo

[English](./README.md) | [中文](./README_CN.md)

This demo shows how to use the `util-genai` module to instrument OpenAI API calls with OpenTelemetry. It demonstrates real-world usage patterns for adding observability to GenAI applications.

## What This Demo Covers

1. **Chat Completion** — Instrument a standard OpenAI chat completion request with `StartLLM`/`StopLLM`/`FailLLM`
2. **Streaming Chat Completion** — Instrument a streaming response, collecting chunks and recording the full result
3. **Embedding** — Instrument an embedding request with `StartEmbedding`/`StopEmbedding`/`FailEmbedding`

All telemetry (spans) is exported to stdout as JSON for easy inspection.

## Prerequisites

- Go 1.24+
- An OpenAI API key (set as environment variable)

## How to Run

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="sk-your-api-key-here"

# Navigate to this directory
cd example/genai-demo/genai

# Download dependencies
go mod tidy

# Run the demo
go run main.go
```

## Expected Output

The program will:

1. Make a chat completion request and print the response
2. Make a streaming chat completion request and print the streamed response
3. Make an embedding request and print the embedding dimensions

Between the application output, you'll see JSON-formatted OpenTelemetry spans printed to stdout, containing:

- Span name (e.g., `chat gpt-4o-mini`, `embeddings text-embedding-3-small`)
- GenAI semantic convention attributes (model, provider, token usage, etc.)
- Timing information (start/end timestamps, duration)

Example span output (abbreviated):

```json
{
  "Name": "chat gpt-4o-mini",
  "SpanContext": { ... },
  "Attributes": [
    { "Key": "gen_ai.system", "Value": { "Type": "STRING", "Value": "openai" } },
    { "Key": "gen_ai.request.model", "Value": { "Type": "STRING", "Value": "gpt-4o-mini" } },
    { "Key": "gen_ai.usage.input_tokens", "Value": { "Type": "INT64", "Value": 25 } },
    { "Key": "gen_ai.usage.output_tokens", "Value": { "Type": "INT64", "Value": 42 } }
  ]
}
```

## Project Structure

```
example/genai-demo/genai/
├── go.mod      # Module definition with local replace directive
├── main.go     # Demo application
└── README.md   # This file
```
