# GenAI 示例

[English](./README.md) | [中文](./README_CN.md)

[`util-genai`](../../util-genai) 模块的可运行示例，演示如何使用 OpenTelemetry 对生成式 AI 操作进行埋点。

| 示例 | 说明 | 是否需要 API key |
|------|------|------------------|
| [genai](./genai) | 对话补全、流式、embedding 埋点 | 需要 |
| [genai-stream](./genai-stream) | 聚焦流式的示例，含 time-to-first-chunk 与 metrics | 需要 |
| [genai-tool](./genai-tool) | Function calling（工具调用）完整回路，含独立的 `execute_tool` span | 需要 |
| [genai-observability](./genai-observability) | 基于日志的事件发送，以及 prompt/response 内容卸载 | 不需要（可离线运行） |

## 前置条件

- Go 1.24+
- 对于会调用 OpenAI API 的示例，需要一个 OpenAI API key（见上表）

每个示例都是独立的 Go 模块。进入对应目录后运行：

```bash
cd example/genai-demo/<示例>
go mod tidy
go run .
```
