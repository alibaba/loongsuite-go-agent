// Copyright (c) 2024 Alibaba Group Holding Ltd.
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

package utilgenai

import "go.opentelemetry.io/otel/attribute"

// ============================================================================
// Official OpenTelemetry GenAI Semantic Conventions
//
// These attribute keys follow the OpenTelemetry semantic conventions for GenAI.
// See: https://github.com/open-telemetry/semantic-conventions-genai/tree/main/model/gen-ai
// ============================================================================

const (
	// GenAI operation attributes
	AttrGenAIOperationName = "gen_ai.operation.name"
	AttrGenAIProviderName  = "gen_ai.provider.name"

	// GenAI request attributes
	AttrGenAIRequestModel            = "gen_ai.request.model"
	AttrGenAIRequestTemperature      = "gen_ai.request.temperature"
	AttrGenAIRequestTopP             = "gen_ai.request.top_p"
	AttrGenAIRequestTopK             = "gen_ai.request.top_k"
	AttrGenAIRequestFrequencyPenalty = "gen_ai.request.frequency_penalty"
	AttrGenAIRequestPresencePenalty  = "gen_ai.request.presence_penalty"
	AttrGenAIRequestMaxTokens        = "gen_ai.request.max_tokens"
	AttrGenAIRequestStopSequences    = "gen_ai.request.stop_sequences"
	AttrGenAIRequestSeed             = "gen_ai.request.seed"
	AttrGenAIRequestStream           = "gen_ai.request.stream"
	AttrGenAIRequestChoiceCount      = "gen_ai.request.choice.count"
	AttrGenAIRequestEncodingFormats  = "gen_ai.request.encoding_formats"
	AttrGenAIRequestReasoningLevel   = "gen_ai.request.reasoning.level"

	// GenAI output attributes
	AttrGenAIOutputType = "gen_ai.output.type"

	// GenAI response attributes
	AttrGenAIResponseModel            = "gen_ai.response.model"
	AttrGenAIResponseID               = "gen_ai.response.id"
	AttrGenAIResponseFinishReasons    = "gen_ai.response.finish_reasons"
	AttrGenAIResponseTimeToFirstChunk = "gen_ai.response.time_to_first_chunk"

	// GenAI usage attributes
	AttrGenAIUsageInputTokens              = "gen_ai.usage.input_tokens"
	AttrGenAIUsageOutputTokens             = "gen_ai.usage.output_tokens"
	AttrGenAIUsageReasoningOutputTokens    = "gen_ai.usage.reasoning.output_tokens"
	AttrGenAIUsageCacheReadInputTokens     = "gen_ai.usage.cache_read.input_tokens"
	AttrGenAIUsageCacheCreationInputTokens = "gen_ai.usage.cache_creation.input_tokens"

	// GenAI conversation attributes
	AttrGenAIConversationID        = "gen_ai.conversation.id"
	AttrGenAIConversationCompacted = "gen_ai.conversation.compacted"

	// GenAI prompt attributes
	AttrGenAIPromptName    = "gen_ai.prompt.name"
	AttrGenAIPromptVersion = "gen_ai.prompt.version"

	// GenAI message content attributes - experimental
	AttrGenAIInputMessages      = "gen_ai.input.messages"
	AttrGenAIOutputMessages     = "gen_ai.output.messages"
	AttrGenAISystemInstructions = "gen_ai.system_instructions"
	AttrGenAIToolDefinitions    = "gen_ai.tool.definitions"

	// GenAI token type attribute for metrics
	AttrGenAITokenType = "gen_ai.token.type"

	// GenAI agent attributes
	AttrGenAIAgentName        = "gen_ai.agent.name"
	AttrGenAIAgentID          = "gen_ai.agent.id"
	AttrGenAIAgentDescription = "gen_ai.agent.description"
	AttrGenAIAgentVersion     = "gen_ai.agent.version"

	// GenAI tool attributes
	AttrGenAIToolName          = "gen_ai.tool.name"
	AttrGenAIToolCallID        = "gen_ai.tool.call.id"
	AttrGenAIToolDescription   = "gen_ai.tool.description"
	AttrGenAIToolType          = "gen_ai.tool.type"
	AttrGenAIToolCallArguments = "gen_ai.tool.call.arguments"
	AttrGenAIToolCallResult    = "gen_ai.tool.call.result"

	// GenAI embeddings attributes
	AttrGenAIEmbeddingsDimensionCount = "gen_ai.embeddings.dimension.count"

	// GenAI retrieval attributes
	AttrGenAIRetrievalQueryText = "gen_ai.retrieval.query.text"
	AttrGenAIRetrievalTopK      = "gen_ai.retrieval.top_k"
	AttrGenAIRetrievalDocuments = "gen_ai.retrieval.documents"
	AttrGenAIDataSourceID       = "gen_ai.data_source.id"

	// GenAI workflow attributes
	AttrGenAIWorkflowName = "gen_ai.workflow.name"

	// Error attributes
	AttrErrorType = "error.type"

	// ========================================================================
	// LoongSuite Extension Attributes
	//
	// These attribute keys are LoongSuite-specific extensions not part of the
	// official OpenTelemetry GenAI semantic conventions.
	// ========================================================================

	// Span kind attribute (LoongSuite Extension - not in official spec)
	AttrGenAISpanKind = "gen_ai.span.kind"

	// Embedding input count (LoongSuite Extension - not in official spec)
	AttrGenAIEmbeddingInputCount = "gen_ai.embedding.input_count"

	// Rerank attributes (LoongSuite Extension - not in official spec)
	AttrGenAIRerankQuery       = "gen_ai.rerank.query"
	AttrGenAIRerankModel       = "gen_ai.rerank.model"
	AttrGenAIRerankTopN        = "gen_ai.rerank.top_n"
	AttrGenAIRerankInputCount  = "gen_ai.rerank.input_count"
	AttrGenAIRerankOutputCount = "gen_ai.rerank.output_count"
)

// SpanKind values for GenAI operations (LoongSuite Extension - not in official spec)
type SpanKindValue string

const (
	SpanKindLLM       SpanKindValue = "LLM"
	SpanKindEmbedding SpanKindValue = "EMBEDDING"
	SpanKindAgent     SpanKindValue = "AGENT"
	SpanKindTool      SpanKindValue = "TOOL"
	SpanKindRetriever SpanKindValue = "RETRIEVER"
	SpanKindReranker  SpanKindValue = "RERANKER"
	// Chain represents a chain/call unit that groups sub-operations.
	SpanKindChain SpanKindValue = "CHAIN"
	// Task represents a task invocation.
	SpanKindTask SpanKindValue = "TASK"
	// Entry represents an entry-point call marker.
	SpanKindEntry SpanKindValue = "ENTRY"
	// Step represents a ReAct round/step marker.
	SpanKindStep SpanKindValue = "STEP"
)

// TokenType values for metrics
type TokenType string

const (
	TokenTypeInput  TokenType = "input"
	TokenTypeOutput TokenType = "output"
)

// OutputType values for gen_ai.output.type
type OutputType string

const (
	OutputTypeText   OutputType = "text"
	OutputTypeJSON   OutputType = "json"
	OutputTypeImage  OutputType = "image"
	OutputTypeSpeech OutputType = "speech"
)

// Metric names for GenAI
const (
	MetricGenAIClientOperationDuration           = "gen_ai.client.operation.duration"
	MetricGenAIClientTokenUsage                  = "gen_ai.client.token.usage"
	MetricGenAIClientOperationTimeToFirstChunk   = "gen_ai.client.operation.time_to_first_chunk"
	MetricGenAIClientOperationTimePerOutputChunk = "gen_ai.client.operation.time_per_output_chunk"
	MetricGenAIInvokeAgentDuration               = "gen_ai.invoke_agent.duration"
	MetricGenAIExecuteToolDuration               = "gen_ai.execute_tool.duration"
	MetricGenAIWorkflowDuration                  = "gen_ai.workflow.duration"
)

// ============================================================================
// Attribute Helper Functions
// ============================================================================

// GenAIOperationName creates an attribute for the operation name.
func GenAIOperationName(name OperationName) attribute.KeyValue {
	return attribute.String(AttrGenAIOperationName, string(name))
}

// GenAIProviderName creates an attribute for the provider name.
func GenAIProviderName(name string) attribute.KeyValue {
	return attribute.String(AttrGenAIProviderName, name)
}

// GenAIRequestModel creates an attribute for the request model.
func GenAIRequestModel(model string) attribute.KeyValue {
	return attribute.String(AttrGenAIRequestModel, model)
}

// GenAIRequestStream creates an attribute indicating streaming mode.
func GenAIRequestStream(stream bool) attribute.KeyValue {
	return attribute.Bool(AttrGenAIRequestStream, stream)
}

// GenAIResponseModel creates an attribute for the response model.
func GenAIResponseModel(model string) attribute.KeyValue {
	return attribute.String(AttrGenAIResponseModel, model)
}

// GenAIResponseID creates an attribute for the response ID.
func GenAIResponseID(id string) attribute.KeyValue {
	return attribute.String(AttrGenAIResponseID, id)
}

// GenAIResponseFinishReasons creates an attribute for finish reasons.
func GenAIResponseFinishReasons(reasons []string) attribute.KeyValue {
	return attribute.StringSlice(AttrGenAIResponseFinishReasons, reasons)
}

// GenAIResponseTimeToFirstChunk creates an attribute for time to first chunk in seconds.
func GenAIResponseTimeToFirstChunk(seconds float64) attribute.KeyValue {
	return attribute.Float64(AttrGenAIResponseTimeToFirstChunk, seconds)
}

// GenAIUsageInputTokens creates an attribute for input token count.
func GenAIUsageInputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageInputTokens, count)
}

// GenAIUsageOutputTokens creates an attribute for output token count.
func GenAIUsageOutputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageOutputTokens, count)
}

// GenAIUsageReasoningOutputTokens creates an attribute for reasoning output token count.
func GenAIUsageReasoningOutputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageReasoningOutputTokens, count)
}

// GenAIUsageCacheReadInputTokens creates an attribute for cache read input token count.
func GenAIUsageCacheReadInputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageCacheReadInputTokens, count)
}

// GenAIUsageCacheCreationInputTokens creates an attribute for cache creation input token count.
func GenAIUsageCacheCreationInputTokens(count int) attribute.KeyValue {
	return attribute.Int(AttrGenAIUsageCacheCreationInputTokens, count)
}

// GenAIConversationID creates an attribute for the conversation/session ID.
func GenAIConversationID(id string) attribute.KeyValue {
	return attribute.String(AttrGenAIConversationID, id)
}

// GenAIConversationCompacted creates an attribute indicating context compaction.
func GenAIConversationCompacted(compacted bool) attribute.KeyValue {
	return attribute.Bool(AttrGenAIConversationCompacted, compacted)
}

// GenAIPromptName creates an attribute for the prompt template name.
func GenAIPromptName(name string) attribute.KeyValue {
	return attribute.String(AttrGenAIPromptName, name)
}

// GenAIPromptVersion creates an attribute for the prompt template version.
func GenAIPromptVersion(version string) attribute.KeyValue {
	return attribute.String(AttrGenAIPromptVersion, version)
}

// ============================================================================
// LoongSuite Extension Helper Functions
// ============================================================================

// GenAISpanKind creates an attribute for the span kind.
// (LoongSuite Extension - not in official spec)
func GenAISpanKind(kind SpanKindValue) attribute.KeyValue {
	return attribute.String(AttrGenAISpanKind, string(kind))
}
