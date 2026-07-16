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
	"encoding/json"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// --- Common getters (shared by agent / workflow / tool spans) ---

type msAgentCommonGetter struct{}

func (msAgentCommonGetter) GetAIOperationName(request msAgentRequest) string {
	return request.operationName
}

func (msAgentCommonGetter) GetAISystem(request msAgentRequest) string {
	return SystemMSAFLKGo
}

func (msAgentCommonGetter) GetGenAISpanKind(request msAgentRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindWorkflow
	}
	return request.spanKind
}

type msWorkflowCommonGetter struct{}

func (msWorkflowCommonGetter) GetAIOperationName(request msWorkflowRequest) string {
	return request.operationName
}

func (msWorkflowCommonGetter) GetAISystem(request msWorkflowRequest) string {
	return SystemMSAFLKGo
}

func (msWorkflowCommonGetter) GetGenAISpanKind(request msWorkflowRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindWorkflow
	}
	return request.spanKind
}

type msToolCommonGetter struct{}

func (msToolCommonGetter) GetAIOperationName(request msToolRequest) string {
	return request.operationName
}

func (msToolCommonGetter) GetAISystem(request msToolRequest) string {
	return SystemMSAFLKGo
}

func (msToolCommonGetter) GetGenAISpanKind(request msToolRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindTool
	}
	return request.spanKind
}

// --- Agent span ---

type msAgentAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[msAgentRequest, msAgentResponse, msAgentCommonGetter]
}

func (e msAgentAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request msAgentRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, msAgentCommonGetter{}.GetGenAISpanKind(request).Attribute())
	if request.sessionID != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.session_id").String(request.sessionID))
	}
	if request.userMessage != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.user_message").String(request.userMessage))
	}
	return attributes, parentContext
}

func (e msAgentAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request msAgentRequest, response msAgentResponse, err error) ([]attribute.KeyValue, context.Context) {
	return e.Base.OnEnd(attributes, ctx, request, response, err)
}

func BuildMSAgentInstrumenter() instrumenter.Instrumenter[msAgentRequest, msAgentResponse] {
	builder := instrumenter.Builder[msAgentRequest, msAgentResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[msAgentRequest, msAgentResponse]{
			Getter: msAgentCommonGetter{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[msAgentRequest]{}).
		AddAttributesExtractor(&msAgentAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[msAgentRequest, msAgentResponse, msAgentCommonGetter]{
				CommonGetter: msAgentCommonGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MS_AGENT_FRAMEWORK_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// --- Workflow span ---

type msWorkflowAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[msWorkflowRequest, msWorkflowResponse, msWorkflowCommonGetter]
}

func (e msWorkflowAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request msWorkflowRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, msWorkflowCommonGetter{}.GetGenAISpanKind(request).Attribute())
	return attributes, parentContext
}

func (e msWorkflowAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request msWorkflowRequest, response msWorkflowResponse, err error) ([]attribute.KeyValue, context.Context) {
	return e.Base.OnEnd(attributes, ctx, request, response, err)
}

func BuildMSWorkflowInstrumenter() instrumenter.Instrumenter[msWorkflowRequest, msWorkflowResponse] {
	builder := instrumenter.Builder[msWorkflowRequest, msWorkflowResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[msWorkflowRequest, msWorkflowResponse]{
			Getter: msWorkflowCommonGetter{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[msWorkflowRequest]{}).
		AddAttributesExtractor(&msWorkflowAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[msWorkflowRequest, msWorkflowResponse, msWorkflowCommonGetter]{
				CommonGetter: msWorkflowCommonGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MS_AGENT_FRAMEWORK_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// --- Tool span ---

type msToolAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[msToolRequest, msToolResponse, msToolCommonGetter]
}

func (e msToolAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request msToolRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, msToolCommonGetter{}.GetGenAISpanKind(request).Attribute())
	if request.toolName != "" {
		attributes = append(attributes, attribute.Key("gen_ai.tool.name").String(request.toolName))
	}
	if request.toolInput != "" {
		attributes = append(attributes, attribute.Key("gen_ai.tool.input").String(request.toolInput))
	}
	return attributes, parentContext
}

func (e msToolAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request msToolRequest, response msToolResponse, err error) ([]attribute.KeyValue, context.Context) {
	attributes, ctx = e.Base.OnEnd(attributes, ctx, request, response, err)
	if response.toolOutput != "" {
		attributes = append(attributes, attribute.Key("gen_ai.tool.output").String(response.toolOutput))
	}
	return attributes, ctx
}

func BuildMSToolInstrumenter() instrumenter.Instrumenter[msToolRequest, msToolResponse] {
	builder := instrumenter.Builder[msToolRequest, msToolResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[msToolRequest, msToolResponse]{
			Getter: msToolCommonGetter{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[msToolRequest]{}).
		AddAttributesExtractor(&msToolAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[msToolRequest, msToolResponse, msToolCommonGetter]{
				CommonGetter: msToolCommonGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.MS_AGENT_FRAMEWORK_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// marshalToolOutput best-effort JSON-encodes the tool result for the
// gen_ai.tool.output attribute. Falls back to fmt.Sprintf when the value
// cannot be JSON-encoded.
func marshalToolOutput(result any) string {
	if result == nil {
		return ""
	}
	if s, ok := result.(string); ok {
		return s
	}
	b, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(b)
}
