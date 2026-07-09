# GenAI Events & Content Upload Demo

[English](./README.md) | [中文](./README_CN.md)

This demo shows two `util-genai` capabilities that require **no external API key**:

1. **Log-based event emission** — a `gen_ai.client.inference.operation.details`
   event is emitted through the OpenTelemetry Logs API and exported to stdout.
2. **Prompt/response content offloading** — message content is written to local
   files, and only a content-addressed `*_ref` attribute is stamped on the span
   instead of the full payload.

The demo fabricates an LLM invocation (no real model call), so it runs offline.

## Prerequisites

- Go 1.24+

## How to Run

```bash
cd example/genai-demo/genai-observability
go mod tidy
go run .
```

## What Happens

The program sets the required environment variables in code:

| Variable | Value | Effect |
|----------|-------|--------|
| `OTEL_SEMCONV_STABILITY_OPT_IN` | `gen_ai_latest_experimental` | Enables experimental GenAI semconv |
| `OTEL_INSTRUMENTATION_GENAI_EMIT_EVENT` | `true` | Emits log-based GenAI events |
| `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT` | `SPAN_AND_EVENT` | Captures message content |
| `OTEL_INSTRUMENTATION_GENAI_UPLOAD_BASE_PATH` | temp dir | Enables local content upload |

It then:

1. Prints the emitted **log event** (with GenAI attributes and message content).
2. Prints the **span** — note the `gen_ai.input.messages_ref`,
   `gen_ai.output.messages_ref`, `gen_ai.system_instructions_ref`, and
   `gen_ai.tool.definitions_ref` attributes pointing to uploaded files.
3. Lists the **uploaded content files** in the temp directory.

## Project Structure

```
example/genai-demo/genai-observability/
├── go.mod      # Module definition with local replace directive
├── main.go     # Demo application
└── README.md   # This file
```
