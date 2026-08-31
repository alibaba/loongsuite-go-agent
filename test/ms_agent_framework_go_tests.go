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

const ms_agent_framework_go_dependency_name = "github.com/microsoft/agent-framework-go"
const ms_agent_framework_go_module_name = "ms-agent-framework-go"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase(
			"ms-agent-framework-go-basic-test",
			ms_agent_framework_go_module_name,
			"v0.0.0", "",
			"1.25", "",
			TestMSAgentFrameworkGoBasic,
		),
		NewMuzzleTestCase(
			"ms-agent-framework-go-muzzle-test",
			ms_agent_framework_go_dependency_name,
			ms_agent_framework_go_module_name,
			"v0.0.0", "",
			"1.25", "",
			[]string{"go", "build", "test_ms_agent_framework.go"},
		),
		NewLatestDepthTestCase(
			"ms-agent-framework-go-latest-depth-test",
			ms_agent_framework_go_dependency_name,
			ms_agent_framework_go_module_name,
			"v0.0.0", "v0.0.0",
			"1.25", "",
			TestMSAgentFrameworkGoBasic,
		),
	)
}

func TestMSAgentFrameworkGoBasic(t *testing.T, env ...string) {
	UseApp("ms-agent-framework-go/v0.0.0")
	RunGoBuild(t, "go", "build", "test_ms_agent_framework.go")
	RunApp(t, "test_ms_agent_framework", env...)
}
