# GenAI Tool / Function Calling Demo

[English](./README.md) | [中文](./README_CN.md)

A standalone demo showing how to instrument an OpenAI **function-calling
(tool-calling)** round trip with the `util-genai` module. It covers the full
loop rather than a single request: the model requests a tool, the tool is
executed under its own span, and the result is fed back for a final answer.

## What This Demo Covers

- Advertising a tool to the model via `FunctionToolDefinition`
  (`LLMInvocation.ToolDefinitions`, exported as `gen_ai.tool.definitions`)
- Recording the model's request as a `ToolCall` output message part with the
  `tool_calls` finish reason
- Executing the tool inside a dedicated `execute_tool` span via
  `StartExecuteTool`/`StopExecuteTool` (`ExecuteToolInvocation`)
- Feeding the result back as a `ToolCallResponse` input part on the second LLM
  call, producing the final answer

## Prerequisites

- Go 1.24+
- An OpenAI API key (set as an environment variable)

## How to Run

```bash
export OPENAI_API_KEY="sk-your-api-key-here"

cd example/genai-demo/genai-tool
go mod tidy
go run main.go
```

## Expected Output

1. The model requests the `get_weather` tool with JSON arguments.
2. The tool runs locally and returns a canned weather string.
3. The model produces a final natural-language answer using the tool result.

Three spans are printed to stdout as JSON: two `chat gpt-4o-mini` LLM spans
(before and after the tool call) and one `execute_tool get_weather` span. The
first LLM span carries `gen_ai.tool.definitions`, and the execute-tool span
carries the tool name, call id, and input/output.

## Project Structure

```
example/genai-demo/genai-tool/
├── go.mod      # Module definition with local replace directive
├── main.go     # Function-calling demo application
└── README.md   # This file
```
