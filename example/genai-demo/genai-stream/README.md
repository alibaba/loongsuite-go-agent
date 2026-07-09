# GenAI Streaming Instrumentation Demo

[English](./README.md) | [中文](./README_CN.md)

A focused, standalone demo showing how to instrument a **streaming** OpenAI chat
completion with the `util-genai` module. Unlike the general `example/genai-demo/genai` demo,
this one concentrates on the streaming path and the streaming-specific telemetry.

## What This Demo Covers

- Marking a request as streaming via `invocation.Stream` (`gen_ai.request.stream`)
- Measuring and recording **time to first chunk**
  (`gen_ai.response.time_to_first_chunk`)
- Draining the stream token-by-token and printing tokens as they arrive
- Capturing token usage from the final usage chunk (`StreamOptions.IncludeUsage`)
- Exporting both **spans** and **metrics** to stdout

## Prerequisites

- Go 1.24+
- An OpenAI API key (set as an environment variable)

## How to Run

```bash
export OPENAI_API_KEY="sk-your-api-key-here"

cd example/genai-demo/genai-stream
go mod tidy
go run main.go
```

## Expected Output

1. The model's answer is printed token-by-token as it streams in.
2. The measured time-to-first-chunk is printed.
3. A GenAI span (`chat gpt-4o-mini`) and metrics
   (`gen_ai.client.operation.duration`, `gen_ai.client.token.usage`) are printed
   to stdout as JSON.

The span includes streaming attributes such as:

```json
{
  "Key": "gen_ai.request.stream",
  "Value": { "Type": "BOOL", "Value": true }
},
{
  "Key": "gen_ai.response.time_to_first_chunk",
  "Value": { "Type": "FLOAT64", "Value": 0.342 }
}
```

## Project Structure

```
example/genai-demo/genai-stream/
├── go.mod      # Module definition with local replace directive
├── main.go     # Streaming demo application
└── README.md   # This file
```
