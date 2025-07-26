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
	"context"
	"io"
	"log"
	"runtime/debug"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	callbacksutils "github.com/cloudwego/eino/utils/callbacks"
)

var einoLLMInstrument = BuildEinoLLMInstrumenter()

var einoCommonInstrument = BuildEinoCommonInstrumenter()

func einoPromptCallbackHandler() *callbacksutils.PromptCallbackHandler {
	return &callbacksutils.PromptCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *prompt.CallbackInput) context.Context {
			request := einoRequest{operationName: "prompt"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, promptRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *prompt.CallbackOutput) context.Context {
			request := ctx.Value(promptRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "prompt"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(promptRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "prompt"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoModelCallHandler(config ChatModelConfig) *callbacksutils.ModelCallbackHandler {
	return &callbacksutils.ModelCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *model.CallbackInput) context.Context {
			clientConfig := input.Config
			request := einoLLMRequest{
				operationName:    "chat",
				serverAddress:    config.BaseURL,
				frequencyPenalty: config.FrequencyPenalty,
				presencePenalty:  config.PresencePenalty,
				seed:             config.Seed,
				topK:             config.TopK,
			}
			if clientConfig != nil {
				request.modelName = clientConfig.Model
				request.maxTokens = int64(clientConfig.MaxTokens)
				request.temperature = float64(clientConfig.Temperature)
				request.stopSequences = clientConfig.Stop
				request.topP = float64(clientConfig.TopP)
			}
			ctx = einoLLMInstrument.Start(ctx, request)
			return context.WithValue(ctx, llmRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			request := ctx.Value(llmRequestKey{}).(einoLLMRequest)
			response := einoLLMResponse{}
			if output.TokenUsage != nil {
				response.usageOutputTokens = int64(output.TokenUsage.CompletionTokens)
				request.usageInputTokens = int64(output.TokenUsage.PromptTokens)
			}
			if output.Message != nil && output.Message.ResponseMeta != nil {
				response.responseFinishReasons = []string{output.Message.ResponseMeta.FinishReason}
			}
			if output.Config != nil {
				response.responseModel = output.Config.Model
			}
			einoLLMInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			request := ctx.Value(llmRequestKey{}).(einoLLMRequest)
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						log.Printf("recover update otel span panic: %v, runinfo: %+v, stack: %s", err, runInfo, string(debug.Stack()))
					}
					output.Close()
				}()
				response := einoLLMResponse{}
				var outs []*model.CallbackOutput
				for {
					chunk, err := output.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("read stream output error: %v, runinfo: %+v", err, runInfo)
					}
					outs = append(outs, chunk)
				}

				var usage *model.TokenUsage
				var mas []*schema.Message
				for _, out := range outs {
					if out == nil {
						continue
					}
					if out.TokenUsage != nil {
						usage = out.TokenUsage
					}
					if out.Message != nil {
						mas = append(mas, out.Message)
					}
				}
				if len(mas) != 0 {
					message, err := schema.ConcatMessages(mas)
					if err == nil {
						response.responseFinishReasons = []string{message.ResponseMeta.FinishReason}
					}
				}
				if usage != nil {
					response.usageOutputTokens = int64(usage.CompletionTokens)
					request.usageInputTokens = int64(usage.PromptTokens)
				}
				response.responseModel = request.modelName
				einoLLMInstrument.End(ctx, request, response, nil)
			}()
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(llmRequestKey{}).(einoLLMRequest)
			response := einoLLMResponse{}
			einoLLMInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoEmbeddingCallbackHandler() *callbacksutils.EmbeddingCallbackHandler {
	return &callbacksutils.EmbeddingCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *embedding.CallbackInput) context.Context {
			request := einoRequest{operationName: "embedding"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, embeddingRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *embedding.CallbackOutput) context.Context {
			request := ctx.Value(embeddingRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "embedding"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(embeddingRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "embedding"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoIndexerCallbackHandler() *callbacksutils.IndexerCallbackHandler {
	return &callbacksutils.IndexerCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *indexer.CallbackInput) context.Context {
			request := einoRequest{operationName: "indexer"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, indexerRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *indexer.CallbackOutput) context.Context {
			request := ctx.Value(indexerRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "indexer"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(indexerRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "indexer"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoRetrieverCallbackHandler() *callbacksutils.RetrieverCallbackHandler {
	return &callbacksutils.RetrieverCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *retriever.CallbackInput) context.Context {
			request := einoRequest{operationName: "retriever"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, retrieverRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *retriever.CallbackOutput) context.Context {
			request := ctx.Value(retrieverRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "retriever"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(retrieverRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "retriever"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoLoaderCallbackHandler() *callbacksutils.LoaderCallbackHandler {
	return &callbacksutils.LoaderCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *document.LoaderCallbackInput) context.Context {
			request := einoRequest{operationName: "loader"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, loaderRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *document.LoaderCallbackOutput) context.Context {
			request := ctx.Value(loaderRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "loader"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(loaderRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "loader"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoToolCallbackHandler() *callbacksutils.ToolCallbackHandler {
	return &callbacksutils.ToolCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
			request := einoRequest{operationName: "execute_tool"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, toolRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "execute_tool"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[*tool.CallbackOutput]) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "execute_tool"}
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						log.Printf("recover update otel span panic: %v, runinfo: %+v, stack: %s", err, info, string(debug.Stack()))
					}
					output.Close()
				}()
				var outs []*tool.CallbackOutput
				for {
					chunk, err := output.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("read stream output error: %v, runinfo: %+v", err, info)
					}
					outs = append(outs, chunk)
				}
				response = extractCallbackOutput(outs, response)
				einoCommonInstrument.End(ctx, request, response, nil)
			}()
			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "execute_tool"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoToolsNodeCallbackHandler() *callbacksutils.ToolsNodeCallbackHandlers {
	return &callbacksutils.ToolsNodeCallbackHandlers{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *schema.Message) context.Context {
			request := einoRequest{operationName: "tool_node"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, toolRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, input []*schema.Message) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "tool_node"}
			response = extractCallbackOutput(input, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[[]*schema.Message]) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "tool_node"}
			go func() {
				defer func() {
					err := recover()
					if err != nil {
						log.Printf("recover update otel span panic: %v, runinfo: %+v, stack: %s", err, info, string(debug.Stack()))
					}
					output.Close()
				}()
				var outs []*schema.Message
				for {
					chunk, err := output.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("read stream output error: %v, runinfo: %+v", err, info)
					}
					outs = append(outs, chunk...)
				}
				response = extractCallbackOutput(outs, response)
				einoCommonInstrument.End(ctx, request, response, nil)
			}()
			return ctx
		},
		OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "tool_node"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

func einoTransformCallbackHandler() *callbacksutils.TransformerCallbackHandler {
	return &callbacksutils.TransformerCallbackHandler{
		OnStart: func(ctx context.Context, runInfo *callbacks.RunInfo, input *document.TransformerCallbackInput) context.Context {
			request := einoRequest{operationName: "transform"}
			request = extractCallbackInput(input, request)
			ctx = einoCommonInstrument.Start(ctx, request)
			return context.WithValue(ctx, toolRequestKey{}, request)
		},
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *document.TransformerCallbackOutput) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "transform"}
			response = extractCallbackOutput(output, response)
			einoCommonInstrument.End(ctx, request, response, nil)
			return ctx
		},
		OnError: func(ctx context.Context, runInfo *callbacks.RunInfo, err error) context.Context {
			request := ctx.Value(toolRequestKey{}).(einoRequest)
			response := einoResponse{operationName: "transform"}
			einoCommonInstrument.End(ctx, request, response, err)
			return ctx
		},
	}
}

type ComposeHandler struct {
	operationName string
}

var _ callbacks.Handler = ComposeHandler{}

func NewComposeHandler(operationName string) *ComposeHandler {
	return &ComposeHandler{
		operationName: operationName,
	}
}

func (c ComposeHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	request := einoRequest{operationName: c.operationName}
	return einoCommonInstrument.Start(ctx, request)
}

func (c ComposeHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	request := einoRequest{operationName: c.operationName}
	response := einoResponse{operationName: "transform"}
	einoCommonInstrument.End(ctx, request, response, nil)
	return ctx
}

func (c ComposeHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	request := einoRequest{operationName: c.operationName}
	response := einoResponse{operationName: "transform"}
	einoCommonInstrument.End(ctx, request, response, err)
	return ctx
}

func (c ComposeHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	request := einoRequest{operationName: c.operationName}
	return einoCommonInstrument.Start(ctx, request)
}

func (c ComposeHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	request := einoRequest{operationName: c.operationName}
	response := einoResponse{operationName: "transform"}
	einoCommonInstrument.End(ctx, request, response, nil)
	return ctx
}

func extractCallbackInput(input interface{}, request einoRequest) einoRequest {
	if input == nil {
		return request
	}
	i, err := sonic.MarshalString(input)
	if err != nil {
		return request
	}
	request.input = map[string]string{
		"input": i,
	}
	return request
}

func extractCallbackOutput(output interface{}, response einoResponse) einoResponse {
	if output == nil {
		return response
	}
	i, err := sonic.MarshalString(output)
	if err != nil {
		return response
	}
	response.output = map[string]string{
		"output": i,
	}
	return response
}
