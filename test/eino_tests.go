// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import "testing"

const eino_dependency_name = "github.com/cloudwego/eino"
const eino_module_name = "eino"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("eino-react-agent-test", eino_module_name, "v0.3.51", "", "1.18", "", TestReactAgentEino),
		NewGeneralTestCase("eino-openai-invoke-test", eino_module_name, "v0.3.51", "", "1.18", "", TestOpenAIInvokeEino),
		NewGeneralTestCase("eino-openai-stream-test", eino_module_name, "v0.3.51", "", "1.18", "", TestOpenAIStreamEino),
		NewGeneralTestCase("eino-ollama-invoke-test", eino_module_name, "v0.3.51", "", "1.18", "", TestOllamaInvokeEino),
		NewGeneralTestCase("eino-ollama-stream-test", eino_module_name, "v0.3.51", "", "1.18", "", TestOllamaStreamEino),
		NewGeneralTestCase("eino-ark-invoke-test", eino_module_name, "v0.3.51", "", "1.18", "", TestArkInvokeEino),
		NewGeneralTestCase("eino-ark-stream-test", eino_module_name, "v0.3.51", "", "1.18", "", TestArkStreamEino),
		NewGeneralTestCase("eino-claude-invoke-test", eino_module_name, "v0.3.51", "", "1.18", "", TestClaudeInvokeEino),
		NewGeneralTestCase("eino-claude-stream-test", eino_module_name, "v0.3.51", "", "1.18", "", TestClaudeStreamEino),
		NewGeneralTestCase("eino-qwen-invoke-test", eino_module_name, "v0.3.51", "", "1.18", "", TestQwenInvokeEino),
		NewGeneralTestCase("eino-qwen-stream-test", eino_module_name, "v0.3.51", "", "1.18", "", TestQwenStreamEino),
		NewGeneralTestCase("eino-document-test", eino_module_name, "v0.3.51", "", "1.18", "", TestDocumentEino),
		NewGeneralTestCase("test-chatmodel-metrics", eino_module_name, "v0.3.51", "", "1.18", "", TestChatModelMetrics),
		NewLatestDepthTestCase("eino-latest-depth-test", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", TestOpenAIInvokeEino),
		NewMuzzleTestCase("eino-muzzle-test-react-agent", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", []string{"go", "build", "test_react_agent.go", "eino_common.go"}),
		NewMuzzleTestCase("eino-muzzle-test-openai-invoke", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", []string{"go", "build", "test_openai_invoke_chatmodel.go", "eino_common.go"}),
		NewMuzzleTestCase("eino-muzzle-test-openai-stream", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", []string{"go", "build", "test_openai_stream_chatmodel.go", "eino_common.go"}),
		NewMuzzleTestCase("eino-muzzle-test-ollama-invoke", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", []string{"go", "build", "test_ollama_invoke_chatmodel.go", "eino_common.go"}),
		NewMuzzleTestCase("eino-muzzle-test-ollama-stream", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", []string{"go", "build", "test_ollama_stream_chatmodel.go", "eino_common.go"}),
		NewMuzzleTestCase("eino-muzzle-test-document-test", eino_dependency_name, eino_module_name, "v0.3.51", "", "1.18", "", []string{"go", "build", "test_document_graph.go", "eino_common.go"}),
	)
}

func TestReactAgentEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_react_agent.go", "eino_common.go")
	RunApp(t, "test_react_agent", env...)
}

func TestDocumentEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_document_graph.go", "eino_common.go")
	RunApp(t, "test_document_graph", env...)
}

func TestOpenAIInvokeEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_openai_invoke_chatmodel.go", "eino_common.go")
	RunApp(t, "test_openai_invoke_chatmodel", env...)
}

func TestOpenAIStreamEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_openai_stream_chatmodel.go", "eino_common.go")
	RunApp(t, "test_openai_stream_chatmodel", env...)
}

func TestOllamaInvokeEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_ollama_invoke_chatmodel.go", "eino_common.go")
	RunApp(t, "test_ollama_invoke_chatmodel", env...)
}

func TestOllamaStreamEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_ollama_stream_chatmodel.go", "eino_common.go")
	RunApp(t, "test_ollama_stream_chatmodel", env...)
}

func TestArkInvokeEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_ark_invoke_chatmodel.go", "eino_common.go")
	RunApp(t, "test_ark_invoke_chatmodel", env...)
}

func TestArkStreamEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_ark_stream_chatmodel.go", "eino_common.go")
	RunApp(t, "test_ark_stream_chatmodel", env...)
}

func TestClaudeInvokeEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_claude_invoke_chatmodel.go", "eino_common.go")
	RunApp(t, "test_claude_invoke_chatmodel", env...)
}

func TestClaudeStreamEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_claude_stream_chatmodel.go", "eino_common.go")
	RunApp(t, "test_claude_stream_chatmodel", env...)
}

func TestQwenInvokeEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_qwen_invoke_chatmodel.go", "eino_common.go")
	RunApp(t, "test_qwen_invoke_chatmodel", env...)
}

func TestQwenStreamEino(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_qwen_stream_chatmodel.go", "eino_common.go")
	RunApp(t, "test_qwen_stream_chatmodel", env...)
}

func TestChatModelMetrics(t *testing.T, env ...string) {
	UseApp("eino/v0.3.51")
	RunGoBuild(t, "go", "build", "test_chatmodel_metrics.go", "eino_common.go")
	RunApp(t, "test_chatmodel_metrics", env...)
}
