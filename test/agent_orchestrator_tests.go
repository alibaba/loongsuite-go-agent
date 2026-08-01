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

package test

import "testing"

const agentOrchestratorModuleName = "agent-orchestrator"

func init() {
	tc1 := NewGeneralTestCase("agent-orchestrator-v0.10.1-test", agentOrchestratorModuleName, "v0.10.1", "", "1.24", "", TestAgentOrchestratorBasic)
	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
}

func TestAgentOrchestratorBasic(t *testing.T, env ...string) {
	UseApp("agent-orchestrator/v0.10.1")
	RunGoBuild(t, "go", "build", "test_agent_orchestrator.go")
	RunApp(t, "test_agent_orchestrator", env...)
}
