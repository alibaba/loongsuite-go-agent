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

package test

import "testing"

const echo_dependency_name = "github.com/labstack/echo/v4"
const echo_module_name = "echo"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("echo-basic-test", echo_module_name, "v4.0.0", "", "1.18", "", TestBasicEcho),
		NewGeneralTestCase("echo-metrics-test", echo_module_name, "v4.0.0", "", "1.18", "", TestMetricsEcho),
		NewGeneralTestCase("echo-unmatched-metrics-test", echo_module_name, "v4.0.0", "", "1.18", "", TestMetricsEchoUnmatched),
		NewGeneralTestCase("echo-middleware-test", echo_module_name, "v4.0.0", "", "1.18", "", TestEchoMiddleware),
		NewGeneralTestCase("echo-pattern-test", echo_module_name, "v4.0.0", "", "1.18", "", TestEchoPattern),
		NewMuzzleTestCase("echo-muzzle-test", echo_dependency_name, echo_module_name, "v4.0.0", "v4.9.1", "1.18", "", []string{"go", "build", "test_echo_basic.go"}),
		NewMuzzleTestCase("echo-muzzle-test", echo_dependency_name, echo_module_name, "v4.10.0", "", "1.18", "", []string{"go", "build", "test_echo_middleware.go"}),
		NewLatestDepthTestCase("echo-latestdepth-test", echo_dependency_name, echo_module_name, "v4.10.0", "", "1.18", "", TestBasicEcho),
	)
}

func TestBasicEcho(t *testing.T, env ...string) {
	UseApp("echo/v4.0.0")
	RunGoBuild(t, "go", "build", "test_echo_basic.go")
	RunApp(t, "test_echo_basic", env...)
}

func TestEchoPattern(t *testing.T, env ...string) {
	UseApp("echo/v4.0.0")
	RunGoBuild(t, "go", "build", "test_echo_pattern.go")
	RunApp(t, "test_echo_pattern", env...)
}

func TestEchoMiddleware(t *testing.T, env ...string) {
	UseApp("echo/v4.10.0")
	RunGoBuild(t, "go", "build", "test_echo_middleware.go")
	RunApp(t, "test_echo_middleware", env...)
}

func TestMetricsEcho(t *testing.T, env ...string) {
	UseApp("echo/v4.0.0")
	RunGoBuild(t, "go", "build", "test_echo_metrics.go")
	RunApp(t, "test_echo_metrics", env...)
}

func TestMetricsEchoUnmatched(t *testing.T, env ...string) {
	UseApp("echo/v4.0.0")
	RunGoBuild(t, "go", "build", "test_echo_unmatched_metrics.go")
	RunApp(t, "test_echo_unmatched_metrics", env...)
}
