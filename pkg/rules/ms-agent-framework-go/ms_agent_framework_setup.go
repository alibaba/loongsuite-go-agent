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
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"

	"github.com/microsoft/agent-framework-go/agent"
	"github.com/microsoft/agent-framework-go/message"
)

var msAgentInstrumenter = BuildMSAgentInstrumenter()
var msWorkflowInstrumenter = BuildMSWorkflowInstrumenter()
var msToolInstrumenter = BuildMSToolInstrumenter()

// --- (*Agent).Run hooks: agent workflow span ---
//
// Signature:
//   func (a *Agent) Run(ctx context.Context, messages []*message.Message, options ...agent.Option) ResponseStream

//go:linkname agentRunOnEnter github.com/microsoft/agent-framework-go/agent.agentRunOnEnter
func agentRunOnEnter(call api.CallContext, a interface{}, ctx context.Context, messages []*message.Message, options ...agent.Option) {
	if !msAgentFrameworkGoEnabler.Enable() {
		return
	}

	request := msAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		userMessage:   extractUserMessage(messages),
		sessionID:     extractSessionID(options),
	}

	instrumentedCtx := msAgentInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname agentRunOnExit github.com/microsoft/agent-framework-go/agent.agentRunOnExit
func agentRunOnExit(call api.CallContext, result interface{}) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(msAgentRequest)
	if ctx == nil {
		return
	}
	msAgentInstrumenter.End(ctx, request, msAgentResponse{}, nil)
}

// --- (*inproc.ExecutionEnvironment).Run hooks: workflow execution span ---
//
// Signature:
//   func (e *ExecutionEnvironment) Run(ctx context.Context, wf *workflow.Workflow, msg any, opts ...ExecutionOption) (*Run, error)

//go:linkname executionEnvRunOnEnter github.com/microsoft/agent-framework-go/workflow/inproc.executionEnvRunOnEnter
func executionEnvRunOnEnter(call api.CallContext, e interface{}, ctx context.Context, wf interface{}, msg interface{}, opts ...interface{}) {
	if !msAgentFrameworkGoEnabler.Enable() {
		return
	}

	request := msWorkflowRequest{
		operationName: OperationRunWorkflow,
		spanKind:      ai.GenAISpanKindWorkflow,
	}

	instrumentedCtx := msWorkflowInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname executionEnvRunOnExit github.com/microsoft/agent-framework-go/workflow/inproc.executionEnvRunOnExit
func executionEnvRunOnExit(call api.CallContext, run interface{}, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(msWorkflowRequest)
	if ctx == nil {
		return
	}
	msWorkflowInstrumenter.End(ctx, request, msWorkflowResponse{}, err)
}

// --- (*functool.funcTool).Call hooks: tool execution span ---
//
// Signature:
//   func (t *funcTool) Call(ctx context.Context, args string) (any, error)

//go:linkname funcToolCallOnEnter github.com/microsoft/agent-framework-go/tool/functool.funcToolCallOnEnter
func funcToolCallOnEnter(call api.CallContext, t interface{}, ctx context.Context, args string) {
	if !msAgentFrameworkGoEnabler.Enable() {
		return
	}

	request := msToolRequest{
		operationName: OperationExecuteTool,
		spanKind:      ai.GenAISpanKindTool,
		toolName:      extractToolName(t),
		toolInput:     args,
	}

	instrumentedCtx := msToolInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname funcToolCallOnExit github.com/microsoft/agent-framework-go/tool/functool.funcToolCallOnExit
func funcToolCallOnExit(call api.CallContext, result any, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(msToolRequest)
	if ctx == nil {
		return
	}
	response := msToolResponse{}
	if err == nil {
		response.toolOutput = marshalToolOutput(result)
	}
	msToolInstrumenter.End(ctx, request, response, err)
}

// extractUserMessage returns the concatenated text of all user-role messages,
// or the first non-empty message text when no user message is found.
func extractUserMessage(messages []*message.Message) string {
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if msg.Role == message.RoleUser {
			if t := msg.Contents.Text(); t != "" {
				return t
			}
		}
	}
	for _, msg := range messages {
		if msg == nil {
			continue
		}
		if t := msg.Contents.Text(); t != "" {
			return t
		}
	}
	return ""
}

// extractSessionID best-effort extracts a session / conversation identifier
// from the agent run options. WithSession takes precedence; WithServiceID is
// the fallback for callers that pass a bare service ID.
func extractSessionID(options []agent.Option) string {
	if sess, ok := agent.GetOption(options, agent.WithSession); ok && sess != nil {
		if sid := sess.ServiceID(); sid != "" {
			return sid
		}
	}
	if sid, ok := agent.GetOption(options, agent.WithServiceID); ok && sid != "" {
		return sid
	}
	return ""
}

// extractToolName returns the configured name of a function-tool by duck-typing
// on the Name() string method exported by functool.funcTool.
func extractToolName(t interface{}) string {
	if t == nil {
		return ""
	}
	if n, ok := t.(interface{ Name() string }); ok {
		return n.Name()
	}
	return ""
}
