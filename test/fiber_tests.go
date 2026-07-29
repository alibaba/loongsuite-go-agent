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
		NewGeneralTestCase("fiberv2-capture-test", fiberv2_module_name, "", "", "1.18", "", TestFiberv2Capture),
		NewGeneralTestCase("fiberv2-capture-disabled-test", fiberv2_module_name, "", "", "1.18", "", TestFiberv2CaptureDisabled),
		NewGeneralTestCase("fiberv2-capture-headers-only-test", fiberv2_module_name, "", "", "1.18", "", TestFiberv2CaptureHeadersOnly),
		NewGeneralTestCase("fiberv2-capture-body-only-test", fiberv2_module_name, "", "", "1.18", "", TestFiberv2CaptureBodyOnly),
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

func TestFiberv2Capture(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}

func TestFiberv2CaptureDisabled(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=false",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}

func TestFiberv2CaptureHeadersOnly(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=false",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}

func TestFiberv2CaptureBodyOnly(t *testing.T, env ...string) {
	UseApp("fiberv2/v2.43.0")
	RunGoBuild(t, "go", "build", "fiber_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true",
	}, env...)
	RunApp(t, "fiber_capture", envs...)
}
