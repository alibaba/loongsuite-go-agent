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

const fiberv2_dependency_name = "github.com/gofiber/fiber/v2"
const fiberv2_module_name = "fiberv2"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("basic-fiberv2-test", fiberv2_module_name, "", "", "1.18", "", TestBasicFiberv2),
		NewGeneralTestCase("basic-fiberv2s-test", fiberv2_module_name, "", "", "1.18", "", TestBasicFiberv2Https),
		NewGeneralTestCase("basic-fiberv2-metrics-test", fiberv2_module_name, "", "", "1.18", "", TestBasicFiberv2Metrics),
		NewGeneralTestCase("basic-fiberv2-route-pattern-test", fiberv2_module_name, "", "", "1.18", "", TestBasicFiberv2RoutePattern),
		NewGeneralTestCase("basic-fiberv2-route-static-test", fiberv2_module_name, "", "", "1.18", "", TestBasicFiberv2RouteStatic),
		NewGeneralTestCase("basic-fiberv2-route-metrics-test", fiberv2_module_name, "", "", "1.18", "", TestBasicFiberv2RouteMetrics),
		NewLatestDepthTestCase("fiberv2-latestdepth", fiberv2_dependency_name, fiberv2_module_name, "v2.43.0", "", "1.18", "", TestBasicFiberv2),
		NewMuzzleTestCase("fiberv2-muzzle", fiberv2_dependency_name, fiberv2_module_name, "v2.43.0", "", "1.18", "", []string{"go", "build", "fiber_http.go"}))
}

func TestBasicFiberv2(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_http.go")
	RunApp(t, "fiber_http", env...)
}

func TestBasicFiberv2Https(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_https.go")
	RunApp(t, "fiber_https", env...)
}

func TestBasicFiberv2Metrics(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_http_metrics.go")
	RunApp(t, "fiber_http_metrics", env...)
}

func TestBasicFiberv2RoutePattern(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_route_pattern.go")
	RunApp(t, "fiber_route_pattern", env...)
}

func TestBasicFiberv2RouteStatic(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_route_static.go")
	RunApp(t, "fiber_route_static", env...)
}

func TestBasicFiberv2RouteMetrics(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_route_metrics.go")
	RunApp(t, "fiber_route_metrics", env...)
}
