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

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type trpcAgentCommonGetter struct{}

func (trpcAgentCommonGetter) GetAIOperationName(request trpcAgentRequest) string {
	return request.operationName
}

func (trpcAgentCommonGetter) GetAISystem(request trpcAgentRequest) string {
	return SystemTrpcAgentGo
}

func (trpcAgentCommonGetter) GetGenAISpanKind(request trpcAgentRequest) ai.GenAISpanKind {
	if request.spanKind == "" {
		return ai.GenAISpanKindWorkflow
	}
	return request.spanKind
}

// trpcAgentAttrsExtractor adds gen_ai.span.kind and gen_ai.other_input.* attributes
// for the agent invocation span.
type trpcAgentAttrsExtractor struct {
	Base ai.AICommonAttrsExtractor[trpcAgentRequest, trpcAgentResponse, trpcAgentCommonGetter]
}

func (e trpcAgentAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request trpcAgentRequest) ([]attribute.KeyValue, context.Context) {
	attributes, parentContext = e.Base.OnStart(attributes, parentContext, request)
	attributes = append(attributes, trpcAgentCommonGetter{}.GetGenAISpanKind(request).Attribute())
	if request.userID != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.user_id").String(request.userID))
	}
	if request.sessionID != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.session_id").String(request.sessionID))
	}
	if request.userMessage != "" {
		attributes = append(attributes, attribute.Key("gen_ai.other_input.user_message").String(request.userMessage))
	}
	return attributes, parentContext
}

func (e trpcAgentAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request trpcAgentRequest, response trpcAgentResponse, err error) ([]attribute.KeyValue, context.Context) {
	return e.Base.OnEnd(attributes, ctx, request, response, err)
}

// BuildTrpcAgentInstrumenter builds the instrumenter for trpc-agent-go agent invocation spans.
func BuildTrpcAgentInstrumenter() instrumenter.Instrumenter[trpcAgentRequest, trpcAgentResponse] {
	builder := instrumenter.Builder[trpcAgentRequest, trpcAgentResponse]{}
	return builder.Init().
		SetSpanNameExtractor(&ai.AISpanNameExtractor[trpcAgentRequest, trpcAgentResponse]{
			Getter: trpcAgentCommonGetter{},
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[trpcAgentRequest]{}).
		AddAttributesExtractor(&trpcAgentAttrsExtractor{
			Base: ai.AICommonAttrsExtractor[trpcAgentRequest, trpcAgentResponse, trpcAgentCommonGetter]{
				CommonGetter: trpcAgentCommonGetter{},
			},
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.TRPC_AGENT_GO_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
