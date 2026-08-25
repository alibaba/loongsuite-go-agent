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
	"os"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
)

type trpcAgentGoInnerEnabler struct {
	enabled bool
}

func (e trpcAgentGoInnerEnabler) Enable() bool {
	return e.enabled
}

var trpcAgentGoEnabler = trpcAgentGoInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_TRPC_AGENT_GO_ENABLED") != "false"}

const (
	OperationInvokeAgent = "invoke_agent"
	SystemTrpcAgentGo    = "trpc_agent_go"
)

// trpcAgentRequest holds the data extracted from a trpc-agent-go runner.Run call.
type trpcAgentRequest struct {
	operationName string
	spanKind      ai.GenAISpanKind
	userID        string
	sessionID     string
	userMessage   string
}

// trpcAgentResponse is reserved for future use; the runner.Run return value
// is a channel that cannot be synchronously inspected.
type trpcAgentResponse struct{}
