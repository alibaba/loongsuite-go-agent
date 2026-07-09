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

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// CompletionParams carries the prompt and response data of a finished
// invocation to a CompletionHook. The Span is provided so a hook may stamp
// reference attributes (e.g. gen_ai.input.messages_ref) before it is ended.
type CompletionParams struct {
	InputMessages     []InputMessage
	OutputMessages    []OutputMessage
	SystemInstruction []MessagePart
	ToolDefinitions   []FunctionToolDefinition
	Span              trace.Span
}

// CompletionHook is invoked when an invocation finishes, allowing prompt and
// response content to be offloaded to external storage. Implementations must be
// safe for concurrent use. OnCompletion is expected to return quickly; any
// slow work (such as uploading) should be performed asynchronously.
type CompletionHook interface {
	// OnCompletion is called synchronously when an invocation finishes, before
	// its span is ended. Implementations may set reference attributes on
	// params.Span and enqueue the actual content for asynchronous processing.
	OnCompletion(ctx context.Context, params CompletionParams)
	// Shutdown flushes any pending work and releases resources. It should honor
	// ctx cancellation/deadline.
	Shutdown(ctx context.Context) error
}

// noopCompletionHook is the default hook that does nothing.
type noopCompletionHook struct{}

func (noopCompletionHook) OnCompletion(context.Context, CompletionParams) {}

func (noopCompletionHook) Shutdown(context.Context) error { return nil }

// NewNoopCompletionHook returns a CompletionHook that performs no work.
func NewNoopCompletionHook() CompletionHook { return noopCompletionHook{} }
