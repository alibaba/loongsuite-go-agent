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

const trpc_agent_go_dependency_name = "trpc.group/trpc-go/trpc-agent-go"
const trpc_agent_go_module_name = "trpc-agent-go"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase(
			"trpc-agent-go-basic-test",
			trpc_agent_go_module_name,
			"v0.1.0", "",
			"1.24", "",
			TestTrpcAgentGoBasic,
		),
		NewMuzzleTestCase(
			"trpc-agent-go-muzzle-test",
			trpc_agent_go_dependency_name,
			trpc_agent_go_module_name,
			"v0.1.0", "",
			"1.24", "",
			[]string{"go", "build", "test_trpc_agent_basic.go"},
		),
		NewLatestDepthTestCase(
			"trpc-agent-go-latest-depth-test",
			trpc_agent_go_dependency_name,
			trpc_agent_go_module_name,
			"v0.1.0", "v0.1.0",
			"1.24", "",
			TestTrpcAgentGoBasic,
		),
	)
}

func TestTrpcAgentGoBasic(t *testing.T, env ...string) {
	UseApp("trpc-agent-go/v1.10.0")
	RunGoBuild(t, "go", "build", "test_trpc_agent_basic.go")
	RunApp(t, "test_trpc_agent_basic", env...)
}
