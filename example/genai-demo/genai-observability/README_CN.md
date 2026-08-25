# GenAI 事件与内容上传示例

[English](./README.md) | [中文](./README_CN.md)

本示例演示 `util-genai` 的两项**无需外部 API key** 的能力：

1. **基于日志的事件发送** —— 通过 OpenTelemetry Logs API 发送
   `gen_ai.client.inference.operation.details` 事件，并导出到 stdout。
2. **prompt/response 内容卸载（offloading）** —— 将消息内容写入本地文件，
   并在 span 上仅标记一个基于内容寻址的 `*_ref` 属性，而非内联完整内容。

本示例会伪造一次 LLM 调用（不进行真实的模型调用），因此可离线运行。

## 前置条件

- Go 1.24+

## 运行方式

```bash
cd example/genai-demo/genai-observability
go mod tidy
go run .
```

## 运行过程

程序在代码中设置了所需的环境变量：

| 变量 | 取值 | 作用 |
|------|------|------|
| `OTEL_SEMCONV_STABILITY_OPT_IN` | `gen_ai_latest_experimental` | 启用实验性 GenAI 语义约定 |
| `OTEL_INSTRUMENTATION_GENAI_EMIT_EVENT` | `true` | 发送基于日志的 GenAI 事件 |
| `OTEL_INSTRUMENTATION_GENAI_CAPTURE_MESSAGE_CONTENT` | `SPAN_AND_EVENT` | 采集消息内容 |
| `OTEL_INSTRUMENTATION_GENAI_UPLOAD_BASE_PATH` | 临时目录 | 启用本地内容上传 |

随后程序会：

1. 打印发送的**日志事件**（包含 GenAI 属性与消息内容）。
2. 打印 **span** —— 注意其中的 `gen_ai.input.messages_ref`、
   `gen_ai.output.messages_ref`、`gen_ai.system_instructions_ref` 和
   `gen_ai.tool.definitions_ref` 属性，它们指向已上传的文件。
3. 列出临时目录中**已上传的内容文件**。

## 项目结构

```
example/genai-demo/genai-observability/
├── go.mod      # 模块定义，包含本地 replace 指令
├── main.go     # 示例应用
└── README.md   # 英文说明
```
