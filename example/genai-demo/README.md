# GenAI Demos

[English](./README.md) | [中文](./README_CN.md)

Runnable examples for the [`util-genai`](../../util-genai) module, showing how to
instrument Generative AI operations with OpenTelemetry.

| Demo | Description | API key |
|------|-------------|---------|
| [genai](./genai) | Chat completion, streaming, and embedding instrumentation | Required |
| [genai-stream](./genai-stream) | Focused streaming demo with time-to-first-chunk and metrics | Required |
| [genai-tool](./genai-tool) | Function-calling (tool) round trip with a dedicated `execute_tool` span | Required |
| [genai-observability](./genai-observability) | Log-based event emission and prompt/response content offloading | Not required (runs offline) |

## Prerequisites

- Go 1.24+
- An OpenAI API key for the demos that call the OpenAI API (see the table above)

Each demo is a standalone Go module. Enter its directory and run it:

```bash
cd example/genai-demo/<demo>
go mod tidy
go run .
```
