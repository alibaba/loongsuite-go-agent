# tRPC-Agent-Go Instrumentation Spec

## Goal

Add compile-time instrumentation for [`trpc-group/trpc-agent-go`](https://github.com/trpc-group/trpc-agent-go)
so that agent invocations are bridged into the LoongSuite full-chain tracing system.

The library already ships its own OpenTelemetry spans (chat / execute_tool / workflow)
via the global `otel.Tracer("trpc.agent.go")`. LoongSuite replaces the global tracer
provider, so those spans are exported through LoongSuite automatically. What is
missing is the **agent workflow span** that anchors the library's spans to the
upstream request context (HTTP/RPC) and emits the ARMS-required `gen_ai.*`
attributes for the `invoke_agent` operation. This spec covers that gap.

## Scope

In scope:
- Hook `runner.Run` (the public top-level entry) to start / end a `workflow` span.
- Emit `gen_ai.system`, `gen_ai.operation.name`, `gen_ai.span.kind`, and the
  `gen_ai.other_input.*` attributes required by ARMS for agent-invocation spans.
- Propagate the instrumented context into the callee so the library's own spans
  are parented under the LoongSuite workflow span.

Out of scope (already covered by other plugins / the library itself):
- Per-provider LLM HTTP spans (covered by `go-openai`, `anthropic-sdk-go`,
  `google-genai`, etc.).
- Tool execution spans (the library emits `execute_tool` spans itself).
- Embedding spans (the library emits `embeddings` spans itself).

## Hook target

```text
ImportPath:    trpc.group/trpc-go/trpc-agent-go/runner
Function:      Run
ReceiverType:  *runner
```

Signature (stable from v0.1.0 through v1.10.0):

```go
func (r *runner) Run(
    ctx context.Context,
    userID string,
    sessionID string,
    message model.Message,
    runOpts ...agent.RunOption,
) (out <-chan *event.Event, err error)
```

Version range: `[v0.1.0, )`. Earlier `v0.0.x` releases use `runOpts ...agent.RunOptions`
(a concrete slice type) and are intentionally not supported.

## Span attributes

Per ARMS gen-ai trace semantic conventions (`gen-ai.md`), an `invoke_agent`
span must set:

| Attribute                     | Value                                       |
|-------------------------------|---------------------------------------------|
| `gen_ai.system`               | `trpc_agent_go`                             |
| `gen_ai.operation.name`       | `invoke_agent`                              |
| `gen_ai.span.kind`            | `workflow`                                  |
| `gen_ai.other_input.user_id`  | `<userID>` (when non-empty)                 |
| `gen_ai.other_input.session_id` | `<sessionID>` (when non-empty)            |
| `gen_ai.other_input.user_message` | first non-empty text content of `message` (when present) |

On exit, if `err != nil`, set `error.type` and span status to `Error`.

Span name: `invoke_agent` (matches `gen_ai.operation.name`, following the
adk-go plugin pattern and the `AISpanNameExtractor`).

Span kind: `client` (consistent with `adk-go`, `eino`, `langchain` agent spans —
the workflow span is the entry into the agent system from the caller's perspective).

Instrumentation scope: `loongsuite.instrumentation.trpc-agent-go`.

## Implementation files

- `pkg/rules/trpc-agent-go/go.mod` — independent Go module.
- `pkg/rules/trpc-agent-go/trpc_agent_data_type.go` — request/response structs,
  enabler, system / operation constants.
- `pkg/rules/trpc-agent-go/trpc_agent_otel_instrumenter.go` — attribute getters
  and `BuildTrpcAgentInstrumenter()`.
- `pkg/rules/trpc-agent-go/trpc_agent_setup.go` — `runnerRunOnEnter` /
  `runnerRunOnExit` hook implementations (with `//go:linkname`).

## Rule registration

`tool/data/rules/trpc-agent-go.json`:

```json
[
  {
    "Version": "[v0.1.0,)",
    "ImportPath": "trpc.group/trpc-go/trpc-agent-go/runner",
    "Function": "Run",
    "ReceiverType": "\\*runner",
    "OnEnter": "runnerRunOnEnter",
    "OnExit": "runnerRunOnExit",
    "Path": "github.com/alibaba/loongsuite-go/pkg/rules/trpc-agent-go"
  }
]
```

## Verification

- `go build ./pkg/rules/trpc-agent-go/...` (plugin module compiles)
- `go test ./test/... -run TestTrpcAgentGo` (integration test asserts the
  workflow span is emitted with the expected attributes and parented under
  the LoongSuite tracer)
- `make build` (instrumentation tool builds with the new rule)

## Test app

`test/trpc-agent-go/v1.10.0/test_trpc_agent_basic.go` will:
1. Construct a `runner.Runner` with an in-memory mock model that returns one
   canned assistant message.
2. Call `runner.Run(ctx, userID, sessionID, userMessage)`.
3. Drain the event channel.
4. Use the `verifier` package to assert the workflow span exists with
   `gen_ai.operation.name=invoke_agent`, `gen_ai.span.kind=workflow`,
   `gen_ai.system=trpc_agent_go`, and the expected `gen_ai.other_input.*`
   attributes.

## Enabler

Gated by `OTEL_INSTRUMENTATION_TRPC_AGENT_GO_ENABLED` (default enabled), matching
the convention used by `trpc` and `adk-go`.
