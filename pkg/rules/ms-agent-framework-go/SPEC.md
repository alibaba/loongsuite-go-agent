# Microsoft Agent Framework for Go — Instrumentation Spec

## Goal

Add compile-time instrumentation for
[`github.com/microsoft/agent-framework-go`](https://github.com/microsoft/agent-framework-go)
so that agent / workflow / tool invocations are bridged into the LoongSuite
full-chain tracing system, emitting spans that follow the OpenTelemetry GenAI
semantic conventions and the ARMS internal `gen-ai.md` attribute requirements.

The framework itself ships its own OTel spans through
`workflow/observability/opentelemetry`, but those spans do not carry the
ARMS-required `gen_ai.*` attributes nor the `gen_ai.span.kind` /
`gen_ai.operation.name` pair that downstream consumers expect. This spec
covers that gap by anchoring an ARMS-conformant workflow / tool span onto
the framework's public entry points.

## Target library status

The upstream repository (`microsoft/agent-framework-go`) currently
**publishes no releases or tags** (verified 2026-07-16) — the API surface is
still fast-iterating. To keep the instrumentation robust against future
renames / refactors, the hook points chosen below are all **exported types
and methods with a stable signature** (`(*Agent).Run`,
`(*ExecutionEnvironment).Run`, `(*funcTool).Call`). Hooking unexported
methods on unexported provider types (e.g. `openaiprovider.(*chatClient).run`)
is intentionally avoided: those are the most likely to change, and the
actual LLM HTTP spans are already emitted by existing provider plugins
(`go-openai`, `openai-go`, `anthropic-sdk-go`, `google-genai`,
`go-openai`/`openai-go-v2`/`-v3`, …) once the framework calls into the
underlying SDK.

## Scope

In scope:

- Hook `(*Agent).Run` — start / end a `workflow` span for
  `gen_ai.operation.name = invoke_agent`.
- Hook `(*inproc.ExecutionEnvironment).Run` — start / end a `workflow`
  span for `gen_ai.operation.name = run_workflow`.
- Hook `(*functool.funcTool).Call` — start / end a `tool` span for
  `gen_ai.operation.name = execute_tool`.

Out of scope (covered elsewhere or deferred):

- Per-provider LLM HTTP spans — already emitted by `go-openai`,
  `anthropic-sdk-go`, `google-genai`, `openai-go(-v2/-v3)`, etc.
- `mcpWrapper.Call`, `Local.Call` (shelltool), `agenttool.functool.Call`
  — same shape but on different receiver types; deferred to a follow-up
  because the framework's tool package is still iterating and the
  function-tool path is by far the dominant case.
- Embedding spans — the framework does not expose a stable embedding entry
  point; defer.

## Hook targets

### 1. Agent workflow span

```text
ImportPath:    github.com/microsoft/agent-framework-go/agent
Function:      Run
ReceiverType:  *Agent
```

Signature (from `agent/agent.go`):

```go
func (a *Agent) Run(ctx context.Context, messages []*message.Message, options ...Option) ResponseStream
```

`ResponseStream` is a type alias for `iter.Seq2[*ResponseUpdate, error]`.
OnExit receives the returned `ResponseStream` value (an iterator function);
we do not drain it — the span is closed as soon as `Run` returns control to
the caller, mirroring how the `trpc-agent-go` plugin handles its `<-chan`
return value.

### 2. Workflow execution span

```text
ImportPath:    github.com/microsoft/agent-framework-go/workflow/inproc
Function:      Run
ReceiverType:  *ExecutionEnvironment
```

Signature (from `workflow/inproc/environment.go`):

```go
func (e *ExecutionEnvironment) Run(ctx context.Context, wf *workflow.Workflow, msg any, opts ...ExecutionOption) (*Run, error)
```

### 3. Tool execution span

```text
ImportPath:    github.com/microsoft/agent-framework-go/tool/functool
Function:      Call
ReceiverType:  *funcTool
```

Signature (from `tool/functool/func.go`):

```go
func (t *funcTool) Call(ctx context.Context, args string) (any, error)
```

## Span attributes

### Agent / Workflow spans (operation = `invoke_agent` / `run_workflow`)

Per ARMS `gen-ai.md`, an agent-invocation span must set:

| Attribute                        | Value                                              |
|----------------------------------|----------------------------------------------------|
| `gen_ai.system`                  | `microsoft_agent_framework_go`                      |
| `gen_ai.operation.name`          | `invoke_agent` (Agent) / `run_workflow` (Workflow) |
| `gen_ai.span.kind`               | `workflow`                                          |
| `gen_ai.other_input.user_message`| first non-empty text content of the input message(s) (when present) |
| `gen_ai.other_input.session_id`  | `<session id>` when discoverable from `agent.Option` (best-effort) |

On exit, if `err != nil`, set `error.type` and span status to `Error`.

Span name: `invoke_agent` / `run_workflow` (matches `gen_ai.operation.name`,
following the `AISpanNameExtractor` pattern used by `adk-go` /
`trpc-agent-go`).

Span kind: `client` (consistent with `adk-go`, `eino`, `trpc-agent-go`
agent/workflow spans — the workflow span is the entry into the agent system
from the caller's perspective).

### Tool span (operation = `execute_tool`)

| Attribute                       | Value                                            |
|---------------------------------|--------------------------------------------------|
| `gen_ai.system`                 | `microsoft_agent_framework_go`                   |
| `gen_ai.operation.name`         | `execute_tool`                                   |
| `gen_ai.span.kind`              | `tool`                                           |
| `gen_ai.tool.name`              | `t.cfg.Name` (the function-tool's configured name)|
| `gen_ai.tool.input`             | `<args string>` (when non-empty)                 |
| `gen_ai.tool.output`            | JSON-encoded result, when err == nil and result != nil |

On exit, if `err != nil`, set `error.type` and span status to `Error`.

Instrumentation scope: `loongsuite.instrumentation.ms-agent-framework-go`.

## Implementation files

- `pkg/rules/ms-agent-framework-go/go.mod` — independent Go module.
- `pkg/rules/ms-agent-framework-go/ms_agent_framework_data_type.go` —
  request/response structs, enabler, system / operation constants.
- `pkg/rules/ms-agent-framework-go/ms_agent_framework_otel_instrumenter.go`
  — attribute getters, `Build*Instrumenter()` for agent / tool spans.
- `pkg/rules/ms-agent-framework-go/ms_agent_framework_setup.go` —
  `agentRunOnEnter` / `agentRunOnExit`,
  `executionEnvRunOnEnter` / `executionEnvRunOnExit`,
  `funcToolCallOnEnter` / `funcToolCallOnExit` hook implementations
  (with `//go:linkname`).

## Rule registration

`tool/data/rules/ms-agent-framework-go.json`:

```json
[
  {
    "ImportPath": "github.com/microsoft/agent-framework-go/agent",
    "Function": "Run",
    "ReceiverType": "\\*Agent",
    "OnEnter": "agentRunOnEnter",
    "OnExit": "agentRunOnExit",
    "Path": "github.com/alibaba/loongsuite-go/pkg/rules/ms-agent-framework-go"
  },
  {
    "ImportPath": "github.com/microsoft/agent-framework-go/workflow/inproc",
    "Function": "Run",
    "ReceiverType": "\\*ExecutionEnvironment",
    "OnEnter": "executionEnvRunOnEnter",
    "OnExit": "executionEnvRunOnExit",
    "Path": "github.com/alibaba/loongsuite-go/pkg/rules/ms-agent-framework-go"
  },
  {
    "ImportPath": "github.com/microsoft/agent-framework-go/tool/functool",
    "Function": "Call",
    "ReceiverType": "\\*funcTool",
    "OnEnter": "funcToolCallOnEnter",
    "OnExit": "funcToolCallOnExit",
    "Path": "github.com/alibaba/loongsuite-go/pkg/rules/ms-agent-framework-go"
  }
]
```

The `Version` field is intentionally omitted. The upstream module has not
published any tags yet — its `go.mod` resolves to the pseudo-version
`v0.0.0-<date>-<commit>`. Per Go's `golang.org/x/mod/semver`, any pre-release
of `v0.0.0` compares *less than* `v0.0.0`, so a constraint like `[0.0.0,)`
would fail to match the pseudo-version. Omitting `Version` makes the rule
match unconditionally, which is the desired behavior here: the hook surface
is the framework's exported API, and we want the rule to keep matching
through any future `v1.x` release without churn.

## Verification

- `go build ./pkg/rules/ms-agent-framework-go/...` — plugin module compiles
  against an in-tree replace of `loongsuite-go/pkg`.
- `go test ./test/... -run TestMSAgentFrameworkGo` — integration test asserts
  the agent / workflow / tool spans are emitted with the expected
  attributes and parented under the LoongSuite tracer.
- `make build` — instrumentation tool builds with the new rule.
- Single-otel-version check: the otel tool's preprocessor adds
  `replace go.opentelemetry.io/otel* => ... v1.40.0` directives to the
  woven module's `go.mod` (see `tool/preprocess/update.go::otelDeps`),
  so the final binary resolves every otel sub-module to exactly one
  version — `v1.40.0` — regardless of the plugin's `v1.44.0` pin. After
  weaving, `go list -m go.opentelemetry.io/otel` from the woven module
  prints `go.opentelemetry.io/otel v1.40.0`, and
  `go list -m all | grep '^go.opentelemetry.io/otel '` returns exactly
  one line. The plugin only uses stable otel v1.x symbols
  (`attribute.KeyValue`, `instrumentation.Scope`), so v1.40 satisfies
  both the plugin and `pkg` — no link-time type mismatch.

## Test app

`test/ms-agent-framework-go/v0.0.0/test_ms_agent_framework.go` will:

1. Construct an `agent.Agent` backed by an in-process mock `RunFunc`
   provider that yields one canned assistant text update and a final
   done update (no network).
2. Register a `functool` whose handler returns a fixed string, attach it
   to the agent through `agent.WithTools`, and let the framework's
   `toolautocall` middleware invoke it.
3. Drive `agent.Run(ctx, userMessage)` and drain the response stream.
4. Use the `verifier` package to assert:
   - the `invoke_agent` workflow span exists with
     `gen_ai.system=microsoft_agent_framework_go`,
     `gen_ai.span.kind=workflow`, `gen_ai.operation.name=invoke_agent`,
     non-empty `gen_ai.other_input.user_message`, and `SpanKind == client`;
   - the `execute_tool` span exists with `gen_ai.span.kind=tool`,
     `gen_ai.operation.name=execute_tool`,
     `gen_ai.tool.name=<configured name>`, and `SpanKind == client`.

The Workflow `(*ExecutionEnvironment).Run` path is exercised by a second
test scenario that builds a tiny single-node workflow through the
`workflow.Builder` and calls `inproc.Default.Run`; it asserts a
`run_workflow` workflow span with the expected attributes.

## Enabler

Gated by `OTEL_INSTRUMENTATION_MS_AGENT_FRAMEWORK_GO_ENABLED` (default
enabled), matching the convention used by `trpc-agent-go` and `adk-go`.

## Open risks

- **Version baseline divergence.** The upstream module
  (`microsoft/agent-framework-go`, verified 2026-07-16) declares
  `go 1.25.0` and `go.opentelemetry.io/otel v1.44.0` in its `go.mod`, and
  its API surface (e.g. `iter.Seq2`-returning `Agent.Run`) requires the
  Go 1.23+ iterator support, so this plugin module must declare
  `go 1.25.0` and pin `otel v1.44.0` to compile against upstream at all.
  This diverges from the rest of `pkg/rules/*` (e.g. `trpc-agent-go`,
  `adk-go`), which sit on `go 1.24 / otel v1.40`.
- **Single-otel-version guarantee.** Although the plugin module's
  `go.mod` requires `otel v1.44`, the otel tool's preprocessor
  (`tool/preprocess/update.go::otelDeps`) rewrites every
  `go.opentelemetry.io/otel*` dependency — across the user module *and*
  every imported rule module — to `v1.40.0` via `replace` directives
  before the woven build runs. MVS therefore resolves a single otel
  version (`v1.40.0`) in the final binary, regardless of the
  version pins in any plugin's `go.mod`. The plugin only references
  stable otel v1.x symbols (`attribute.KeyValue`,
  `instrumentation.Scope`) that are present in both `v1.40` and `v1.44`,
  so the rewrite is source-compatible and there is no link-time type
  mismatch. Empirical verification is recorded in the Verification
  section below.
- **`funcTool` rename risk.** `*functool.funcTool` is an *unexported*
  type. If upstream renames it, the tool rule silently no-ops. The
  muzzle test only verifies the instrumentation compiles; it does
  **not** assert that the hook actually fires, and `LatestDepth` is
  pinned to `v0.0.0`, so neither test catches an upstream rename. This
  is an accepted limitation — the agent and workflow hooks anchor on
  exported types and are far less likely to move.
- **Streaming duration.** `(*Agent).Run` returns an `iter.Seq2`
  iterator; the `invoke_agent` span is closed in `agentRunOnExit`,
  which fires as soon as `Run` returns the iterator to the caller —
  *before* the caller drains the stream. The span duration therefore
  measures the setup phase, not the full streaming consumption. This
  mirrors the `trpc-agent-go` `<-chan` precedent and is an accepted
  limitation.
