# GenAI 流式埋点示例

[English](./README.md) | [中文](./README_CN.md)

一个专注、独立的示例，演示如何使用 `util-genai` 模块对**流式**的 OpenAI 对话补全进行埋点。与通用的 `example/genai-demo/genai` 示例不同，本示例聚焦于流式路径以及流式特有的遥测数据。

## 本示例涵盖的内容

- 通过 `invocation.Stream` 将请求标记为流式（`gen_ai.request.stream`）
- 测量并记录**首个 chunk 的耗时**（`gen_ai.response.time_to_first_chunk`）
- 逐 token 读取流，并在 token 到达时即时打印
- 从最后的 usage chunk 中获取 token 使用量（`StreamOptions.IncludeUsage`）
- 将 **spans** 和 **metrics** 同时导出到 stdout

## 前置条件

- Go 1.24+
- 一个 OpenAI API key（通过环境变量设置）

## 运行方式

```bash
export OPENAI_API_KEY="sk-your-api-key-here"

cd example/genai-demo/genai-stream
go mod tidy
go run main.go
```

## 预期输出

1. 模型的回答会随着流式返回逐 token 打印。
2. 打印测得的首个 chunk 耗时（time-to-first-chunk）。
3. 一个 GenAI span（`chat gpt-4o-mini`）以及 metrics（`gen_ai.client.operation.duration`、`gen_ai.client.token.usage`）会以 JSON 格式打印到 stdout。

该 span 包含流式相关属性，例如：

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

## 项目结构

```
example/genai-demo/genai-stream/
├── go.mod      # 模块定义，包含本地 replace 指令
├── main.go     # 流式示例应用
└── README.md   # 英文说明
```
