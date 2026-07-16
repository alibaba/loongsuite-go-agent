// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package msagentframeworkgo

import (
	"os"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
)

type msAgentFrameworkGoInnerEnabler struct {
	enabled bool
}

func (e msAgentFrameworkGoInnerEnabler) Enable() bool {
	return e.enabled
}

var msAgentFrameworkGoEnabler = msAgentFrameworkGoInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_MS_AGENT_FRAMEWORK_GO_ENABLED") != "false"}

const (
	OperationInvokeAgent = "invoke_agent"
	OperationRunWorkflow = "run_workflow"
	OperationExecuteTool = "execute_tool"
	SystemMSAFLKGo      = "microsoft_agent_framework_go"
)

// msAgentRequest holds the data extracted from an (*Agent).Run call.
type msAgentRequest struct {
	operationName string
	spanKind      ai.GenAISpanKind
	userMessage   string
	sessionID     string
}

// msAgentResponse is reserved for future use; the (*Agent).Run return value
// is a streaming iterator that cannot be synchronously inspected.
type msAgentResponse struct{}

// msWorkflowRequest holds the data extracted from an
// (*ExecutionEnvironment).Run call.
type msWorkflowRequest struct {
	operationName string
	spanKind      ai.GenAISpanKind
}

// msWorkflowResponse is reserved for future use; the (*ExecutionEnvironment).Run
// return value is a *Run that streams events asynchronously.
type msWorkflowResponse struct{}

// msToolRequest holds the data extracted from a (*funcTool).Call call.
type msToolRequest struct {
	operationName string
	spanKind      ai.GenAISpanKind
	toolName      string
	toolInput     string
}

// msToolResponse holds the result of a (*funcTool).Call call.
type msToolResponse struct {
	toolOutput string
}
