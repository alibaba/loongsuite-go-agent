# GenAI 埋点示例

[English](./README.md) | [中文](./README_CN.md)

本示例演示如何使用 `util-genai` 模块，通过 OpenTelemetry 对 OpenAI API 调用进行埋点，展示了为 GenAI 应用添加可观测性的真实用法。

## 本示例涵盖的内容

1. **对话补全（Chat Completion）** —— 使用 `StartLLM`/`StopLLM`/`FailLLM` 对标准的 OpenAI 对话补全请求进行埋点
2. **流式对话补全（Streaming Chat Completion）** —— 对流式响应进行埋点，收集各个 chunk 并记录完整结果
3. **Embedding** —— 使用 `StartEmbedding`/`StopEmbedding`/`FailEmbedding` 对 embedding 请求进行埋点

所有遥测数据（spans）都会以 JSON 格式导出到 stdout，便于查看。

## 前置条件

- Go 1.24+
- 一个 OpenAI API key（通过环境变量设置）

## 运行方式

```bash
# 设置你的 OpenAI API key
export OPENAI_API_KEY="sk-your-api-key-here"

# 进入本目录
cd example/genai-demo/genai

# 下载依赖
go mod tidy

# 运行示例
go run main.go
```

## 预期输出

程序将会：

1. 发起一次对话补全请求并打印响应
2. 发起一次流式对话补全请求并打印流式响应
3. 发起一次 embedding 请求并打印 embedding 维度

在应用输出之间，你会看到以 JSON 格式打印到 stdout 的 OpenTelemetry spans，其中包含：

- Span 名称（例如 `chat gpt-4o-mini`、`embeddings text-embedding-3-small`）
- GenAI 语义约定属性（模型、提供方、token 使用量等）
- 时间信息（起止时间戳、耗时）

Span 输出示例（已简化）：

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

## 项目结构

```
example/genai-demo/genai/
├── go.mod      # 模块定义，包含本地 replace 指令
├── main.go     # 示例应用
└── README.md   # 英文说明
```
