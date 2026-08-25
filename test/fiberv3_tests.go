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

const fiberv3_dependency_name = "github.com/gofiber/fiber/v3"
const fiberv3_module_name = "fiberv3"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("basic-fiberv3-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestBasicFiberv3),
		NewGeneralTestCase("basic-fiberv3s-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestBasicFiberv3Https),
		NewGeneralTestCase("basic-fiberv3-metrics-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestBasicFiberv3Metrics),
		NewGeneralTestCase("fiberv3-capture-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestFiberv3Capture),
		NewGeneralTestCase("fiberv3-capture-disabled-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestFiberv3CaptureDisabled),
		NewGeneralTestCase("fiberv3-capture-headers-only-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestFiberv3CaptureHeadersOnly),
		NewGeneralTestCase("fiberv3-capture-body-only-test", fiberv3_module_name, "v3.0.0", "", "1.25", "", TestFiberv3CaptureBodyOnly),
		NewLatestDepthTestCase("fiberv3-latestdepth", fiberv3_dependency_name, fiberv3_module_name, "v3.0.0", "", "1.25", "", TestBasicFiberv3),
		// Custom-ctx apps route through (*App).customRequestHandler on fiber
		// >= v3.3.0; run it as a latest-depth test so it exercises that path.
		NewLatestDepthTestCase("fiberv3-custom-ctx-latestdepth", fiberv3_dependency_name, fiberv3_module_name, "v3.3.0", "", "1.25", "", TestBasicFiberv3CustomCtx),
		NewMuzzleTestCase("fiberv3-muzzle", fiberv3_dependency_name, fiberv3_module_name, "v3.0.0", "", "1.25", "", []string{"go", "build", "fiber_http.go"}))
}

func TestBasicFiberv3(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_http.go")
	RunApp(t, "fiber_http", env...)
}

func TestBasicFiberv3Https(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_https.go")
	RunApp(t, "fiber_https", env...)
}

func TestBasicFiberv3Metrics(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_http_metrics.go")
	RunApp(t, "fiber_http_metrics", env...)
}

func TestBasicFiberv3CustomCtx(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_custom_ctx.go")
	RunApp(t, "fiber_custom_ctx", env...)
}

func TestFiberv3Capture(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}

func TestFiberv3CaptureDisabled(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=false",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}

func TestFiberv3CaptureHeadersOnly(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=false",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}

func TestFiberv3CaptureBodyOnly(t *testing.T, env ...string) {
	UseApp("fiberv3/v3.0.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}
