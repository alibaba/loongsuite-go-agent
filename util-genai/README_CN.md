# OpenTelemetry GenAI 工具库（Go）

[English](./README.md) | [中文](./README_CN.md)

本包为 Go 提供了用于 GenAI 埋点的 OpenTelemetry 工具库，是 Python [opentelemetry-util-genai](https://github.com/open-telemetry/opentelemetry-python-contrib/tree/main/util/opentelemetry-util-genai) 包的 Go 移植版本。

## 概述

GenAI 工具库封装了用于标准化生成式 AI 埋点的样板代码和辅助函数。本包提供了一组 API 和类型，以尽量减少对 GenAI 库进行埋点所需的工作量，同时为生成两类 OpenTelemetry 数据提供标准化支持：“spans 与 metrics”以及“spans、metrics 与 events”。

## 安装

```bash
go get github.com/alibaba/loongsuite-go/util-genai
```

## 环境变量

本包依赖环境变量来配置是否采集消息内容。默认情况下，不会采集消息内容。

| 变量 | 说明 | 取值 |
|------|------|------|
| `OTEL_SEMCONV_STABILITY_OPT_IN` | 启用实验性特性 | `gen_ai_latest_experimental` |
| `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT` | 控制消息内容采集 | `NO_CONTENT`、`SPAN_ONLY`、`EVENT_ONLY`、`SPAN_AND_EVENT` |
| `OTEL_INSTRUMENTATION_GENAI_EMIT_EVENT` | 控制事件发送 | `true`、`false` |

## Span 属性

本包按照 [OpenTelemetry GenAI 语义约定](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/gen-ai/README.md) 提供以下 span 属性：

- `gen_ai.provider.name`：提供方名称（例如 "openai"）
- `gen_ai.operation.name`：操作名称（例如 "chat"）
- `gen_ai.request.model`：请求的模型名称
- `gen_ai.response.finish_reasons`：结束原因列表
- `gen_ai.response.model`：响应的模型名称
- `gen_ai.response.id`：响应 ID
- `gen_ai.usage.input_tokens`：输入 token 数量
- `gen_ai.usage.output_tokens`：输出 token 数量
- `gen_ai.input.messages`：输入消息（在启用内容采集时）
- `gen_ai.output.messages`：输出消息（在启用内容采集时）
- `gen_ai.system_instructions`：系统指令（在提供时）

## 使用方式

### 基础 LLM 调用

```go
package main

import (
    "context"

    utilgenai "github.com/alibaba/loongsuite-go/util-genai"
)

func main() {
    handler := utilgenai.GetTelemetryHandler()
    ctx := context.Background()

    // 使用请求数据创建一个调用对象
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

    // 开始调用（打开一个 span）
    ctx = handler.StartLLM(ctx, invocation)

    // 执行真正的 LLM 调用
    // response, err := client.Chat(ctx, request)

    // 填充输出
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

    // 结束调用（关闭 span）
    handler.StopLLM(invocation)
}
```

### 错误处理

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

    // 填充输出并结束
    // ...
    handler.StopLLM(invocation)
    return nil
}
```

### 流式（Stream）模式

对于流式 LLM 响应，将 `invocation.Stream` 设置为 `true`，以标记该请求为流式请求（`gen_ai.request.stream`），并将接收到首个 chunk 的耗时记录到 `invocation.TimeToFirstChunk`（`gen_ai.response.time_to_first_chunk`）。请在发起网络调用**之前**开启 span，以便准确测量时延；随后在读取流的过程中持续累积输出，待整个流读取完毕后再调用 `StopLLM`。

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

// 在发起网络调用之前开启 span，以准确捕获时延。
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

    // 仅在收到首个 chunk 时记录一次 time-to-first-chunk。
    if firstChunk {
        ttfc := time.Since(streamStart).Seconds()
        invocation.TimeToFirstChunk = &ttfc
        firstChunk = false
    }

    // 仅包含 usage 的最后一个 chunk 携带 token 计数（StreamOptions.IncludeUsage）。
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

// 填充聚合后的响应并成功关闭 span。
invocation.OutputMessages = []utilgenai.OutputMessage{
    {
        Role:         "assistant",
        Parts:        []utilgenai.MessagePart{utilgenai.Text{Content: fullContent}},
        FinishReason: utilgenai.FinishReasonStop,
    },
}
handler.StopLLM(invocation)
```

启用流式模式后，本包会额外产生以下遥测数据：

- Span 属性 `gen_ai.request.stream`：`true`
- Span 属性 `gen_ai.response.time_to_first_chunk`：接收到首个 chunk 的耗时（秒）
- Metric `gen_ai.client.operation.time_to_first_chunk`：time-to-first-chunk 直方图

完整的可运行示例见 [`example/genai-stream`](../example/genai-stream)。

### Embedding 调用

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewEmbeddingInvocation("text-embedding-3-small")
invocation.Provider = "openai"
inputCount := 5
invocation.InputCount = &inputCount

ctx = handler.StartEmbedding(ctx, invocation)

// 执行 embedding 调用
// ...

inputTokens := 100
invocation.InputTokens = &inputTokens
handler.StopEmbedding(invocation)
```

### 工具执行

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewExecuteToolInvocation("get_weather")
invocation.ToolCallID = "call_123"
invocation.Input = map[string]any{"location": "San Francisco"}

ctx = handler.StartExecuteTool(ctx, invocation)

// 执行工具
result := getWeather("San Francisco")
invocation.Output = result

handler.StopExecuteTool(invocation)
```

### Agent 调用

```go
handler := utilgenai.GetTelemetryHandler()

invocation := utilgenai.NewInvokeAgentInvocation()
invocation.AgentName = "assistant"
invocation.Provider = "openai"

ctx = handler.StartInvokeAgent(ctx, invocation)

// 运行 agent
// ...

handler.StopInvokeAgent(invocation)
```

### 自定义 Provider 配置

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/trace"
)

// 创建自定义 tracer provider
tp := trace.NewTracerProvider(...)
otel.SetTracerProvider(tp)

// 使用自定义 provider 创建 handler
handler := utilgenai.NewTelemetryHandler(
    utilgenai.WithTracerProvider(tp),
    utilgenai.WithMeterProvider(mp),
)
```

## 支持的操作

| 操作 | Handler 方法 |
|------|-------------|
| LLM/Chat | `StartLLM`、`StopLLM`、`FailLLM` |
| Embeddings | `StartEmbedding`、`StopEmbedding`、`FailEmbedding` |
| 工具执行 | `StartExecuteTool`、`StopExecuteTool`、`FailExecuteTool` |
| Agent 调用 | `StartInvokeAgent`、`StopInvokeAgent`、`FailInvokeAgent` |
| Agent 创建 | `StartCreateAgent`、`StopCreateAgent`、`FailCreateAgent` |
| 文档检索 | `StartRetrieve`、`StopRetrieve`、`FailRetrieve` |
| 文档重排 | `StartRerank`、`StopRerank`、`FailRerank` |

## Metrics

本包会自动记录以下指标：

- `gen_ai.client.operation.duration`：GenAI 客户端操作的耗时（直方图）
- `gen_ai.client.token.usage`：输入和输出的 token 使用量（直方图）
- `gen_ai.client.operation.time_to_first_chunk`：流式响应中接收到首个 chunk 的耗时（直方图）

## 参考资料

- [OpenTelemetry GenAI 语义约定](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/gen-ai/README.md)
- [Python opentelemetry-util-genai](https://github.com/open-telemetry/opentelemetry-python-contrib/tree/main/util/opentelemetry-util-genai)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)

## 许可证

Apache License 2.0
