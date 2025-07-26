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

package eino

import (
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type einoLLMAttrsGetter struct{}

var _ ai.LLMAttrsGetter[einoLLMRequest, einoLLMResponse] = einoLLMAttrsGetter{}
var _ ai.CommonAttrsGetter[einoLLMRequest, any] = einoLLMAttrsGetter{}

func (e einoLLMAttrsGetter) GetAIOperationName(request einoLLMRequest) string {
	return request.operationName
}

func (e einoLLMAttrsGetter) GetAISystem(request einoLLMRequest) string {
	return "eino"
}

func (e einoLLMAttrsGetter) GetAIRequestModel(request einoLLMRequest) string {
	return request.modelName
}

func (e einoLLMAttrsGetter) GetAIRequestEncodingFormats(request einoLLMRequest) []string {
	return request.encodingFormats
}

func (e einoLLMAttrsGetter) GetAIRequestFrequencyPenalty(request einoLLMRequest) float64 {
	return request.frequencyPenalty
}

func (e einoLLMAttrsGetter) GetAIRequestPresencePenalty(request einoLLMRequest) float64 {
	return request.frequencyPenalty
}

func (e einoLLMAttrsGetter) GetAIResponseFinishReasons(request einoLLMRequest, response einoLLMResponse) []string {
	return response.responseFinishReasons
}

func (e einoLLMAttrsGetter) GetAIResponseModel(request einoLLMRequest, response einoLLMResponse) string {
	return response.responseModel
}

func (e einoLLMAttrsGetter) GetAIRequestMaxTokens(request einoLLMRequest) int64 {
	return request.maxTokens
}

func (e einoLLMAttrsGetter) GetAIUsageInputTokens(request einoLLMRequest) int64 {
	return request.usageInputTokens
}

func (e einoLLMAttrsGetter) GetAIUsageOutputTokens(request einoLLMRequest, response einoLLMResponse) int64 {
	return response.usageOutputTokens
}

func (e einoLLMAttrsGetter) GetAIRequestStopSequences(request einoLLMRequest) []string {
	return request.stopSequences
}

func (e einoLLMAttrsGetter) GetAIRequestTemperature(request einoLLMRequest) float64 {
	return request.temperature
}

func (e einoLLMAttrsGetter) GetAIRequestTopK(request einoLLMRequest) float64 {
	return request.topK
}

func (e einoLLMAttrsGetter) GetAIRequestTopP(request einoLLMRequest) float64 {
	return request.topP
}

func (e einoLLMAttrsGetter) GetAIResponseID(request einoLLMRequest, response einoLLMResponse) string {
	return response.responseID
}

func (e einoLLMAttrsGetter) GetAIServerAddress(request einoLLMRequest) string {
	return request.serverAddress
}

func (e einoLLMAttrsGetter) GetAIRequestSeed(request einoLLMRequest) int64 {
	return request.seed
}

func BuildEinoLLMInstrumenter() instrumenter.Instrumenter[einoLLMRequest, einoLLMResponse] {
	builder := instrumenter.Builder[einoLLMRequest, einoLLMResponse]{}
	return builder.Init().SetSpanNameExtractor(&ai.AISpanNameExtractor[einoLLMRequest, einoLLMResponse]{Getter: einoLLMAttrsGetter{}}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[einoLLMRequest]{}).
		AddAttributesExtractor(&ai.AILLMAttrsExtractor[einoLLMRequest, einoLLMResponse, einoLLMAttrsGetter, einoLLMAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.EINO_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(ai.AIClientMetrics("eino-llm")).
		BuildInstrumenter()
}
