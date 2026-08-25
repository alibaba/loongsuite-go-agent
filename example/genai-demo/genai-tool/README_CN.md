# GenAI 工具 / Function Calling 示例

[English](./README.md) | [中文](./README_CN.md)

一个独立示例，演示如何使用 `util-genai` 模块对 OpenAI 的 **function calling（工具调用）** 完整流程进行埋点。它覆盖的不是单次请求，而是完整回路：模型请求调用工具 → 工具在独立 span 中执行 → 将结果回填给模型得到最终答案。

## 本示例涵盖的内容

- 通过 `FunctionToolDefinition`（`LLMInvocation.ToolDefinitions`，导出为
  `gen_ai.tool.definitions`）向模型声明可用工具
- 将模型的调用请求记录为 `ToolCall` 类型的输出消息 part，finish reason 为
  `tool_calls`
- 通过 `StartExecuteTool`/`StopExecuteTool`（`ExecuteToolInvocation`）在独立的
  `execute_tool` span 中执行工具
- 在第二次 LLM 调用中，将结果作为 `ToolCallResponse` 输入 part 回填，得到最终答案

## 前置条件

- Go 1.24+
- 一个 OpenAI API key（通过环境变量设置）

## 运行方式

```bash
export OPENAI_API_KEY="sk-your-api-key-here"

cd example/genai-demo/genai-tool
go mod tidy
go run main.go
```

## 预期输出

1. 模型请求调用 `get_weather` 工具，并给出 JSON 参数。
2. 工具在本地执行，返回一段固定的天气字符串。
3. 模型基于工具结果生成最终的自然语言答案。

程序会以 JSON 格式向 stdout 打印三个 span：两个 `chat gpt-4o-mini` LLM span
（工具调用前后各一个）以及一个 `execute_tool get_weather` span。第一个 LLM span
带有 `gen_ai.tool.definitions`，而 execute-tool span 带有工具名、call id 以及
输入/输出。

## 项目结构

```
example/genai-demo/genai-tool/
├── go.mod      # 模块定义，包含本地 replace 指令
├── main.go     # Function calling 示例应用
└── README.md   # 英文说明
```
