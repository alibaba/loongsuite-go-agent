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

package trpcagentgo

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"trpc.group/trpc-go/trpc-agent-go/agent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
)

var trpcAgentInstrumenter = BuildTrpcAgentInstrumenter()

// runnerRunOnEnter hooks (*runner).Run on enter.
// Signature: func (r *runner) Run(ctx context.Context, userID string, sessionID string, message model.Message, runOpts ...agent.RunOption) (<-chan *event.Event, error)
//
//go:linkname runnerRunOnEnter trpc.group/trpc-go/trpc-agent-go/runner.runnerRunOnEnter
func runnerRunOnEnter(call api.CallContext, r interface{}, ctx context.Context, userID string, sessionID string, message model.Message, runOpts ...agent.RunOption) {
	if !trpcAgentGoEnabler.Enable() {
		return
	}

	request := trpcAgentRequest{
		operationName: OperationInvokeAgent,
		spanKind:      ai.GenAISpanKindWorkflow,
		userID:        userID,
		sessionID:     sessionID,
		userMessage:   extractUserMessage(message),
	}

	instrumentedCtx := trpcAgentInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = instrumentedCtx
	data["request"] = request
	call.SetData(data)
	call.SetParam(1, instrumentedCtx)
}

//go:linkname runnerRunOnExit trpc.group/trpc-go/trpc-agent-go/runner.runnerRunOnExit
func runnerRunOnExit(call api.CallContext, out <-chan *event.Event, err error) {
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}
	ctx, _ := data["ctx"].(context.Context)
	request, _ := data["request"].(trpcAgentRequest)
	if ctx == nil {
		return
	}
	trpcAgentInstrumenter.End(ctx, request, trpcAgentResponse{}, err)
}

// extractUserMessage returns the first non-empty text content of the message.
// model.Message may carry text in Content or in ContentParts.
func extractUserMessage(msg model.Message) string {
	if msg.Content != "" {
		return msg.Content
	}
	for _, part := range msg.ContentParts {
		if part.Text != nil && *part.Text != "" {
			return *part.Text
		}
	}
	return ""
}
