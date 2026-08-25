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

const fasthttp_dependency_name = "github.com/valyala/fasthttp"
const fasthttp_module_name = "fasthttp"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("basic-fasthttp-test", fasthttp_module_name, "", "", "1.18", "", TestBasicFastHttp),
		NewGeneralTestCase("basic-fasthttps-test", fasthttp_module_name, "", "", "1.18", "", TestBasicFastHttps),
		NewGeneralTestCase("fasthttp-capture-test", fasthttp_module_name, "", "", "1.18", "", TestFastHttpCapture),
		NewGeneralTestCase("fasthttp-capture-disabled-test", fasthttp_module_name, "", "", "1.18", "", TestFastHttpCaptureDisabled),
		NewGeneralTestCase("fasthttp-capture-headers-only-test", fasthttp_module_name, "", "", "1.18", "", TestFastHttpCaptureHeadersOnly),
		NewGeneralTestCase("fasthttp-capture-body-only-test", fasthttp_module_name, "", "", "1.18", "", TestFastHttpCaptureBodyOnly),
		// fasthttp v1.73.0 and above declare `go 1.25.0` in their go.mod, so the
		// latest-depth check requires go 1.25.
		NewLatestDepthTestCase("fasthttp-latestdepth", fasthttp_dependency_name, fasthttp_module_name, "v1.45.0", "", "1.25", "", TestBasicFastHttp),
		NewMuzzleTestCase("fasthttp-muzzle", fasthttp_dependency_name, fasthttp_module_name, "v1.45.0", "", "1.18", "", []string{"go", "build", "test_basic_http.go", "server.go"}))
}

func TestBasicFastHttp(t *testing.T, env ...string) {
	UseApp("fasthttp/v1.45.0")
	RunGoBuild(t, "go", "build", "test_basic_http.go", "server.go")
	RunApp(t, "test_basic_http", env...)
}

func TestBasicFastHttps(t *testing.T, env ...string) {
	UseApp("fasthttp/v1.45.0")
	RunGoBuild(t, "go", "build", "test_basic_https.go", "server.go")
	RunApp(t, "test_basic_https", env...)
}

func TestFastHttpCapture(t *testing.T, env ...string) {
	UseApp("fasthttp/v1.45.0")
	RunGoBuild(t, "go", "build", "test_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true",
	}, env...)
	RunApp(t, "test_capture", envs...)
}

func TestFastHttpCaptureDisabled(t *testing.T, env ...string) {
	UseApp("fasthttp/v1.45.0")
	RunGoBuild(t, "go", "build", "test_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=false",
	}, env...)
	RunApp(t, "test_capture", envs...)
}

func TestFastHttpCaptureHeadersOnly(t *testing.T, env ...string) {
	UseApp("fasthttp/v1.45.0")
	RunGoBuild(t, "go", "build", "test_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=false",
	}, env...)
	RunApp(t, "test_capture", envs...)
}

func TestFastHttpCaptureBodyOnly(t *testing.T, env ...string) {
	UseApp("fasthttp/v1.45.0")
	RunGoBuild(t, "go", "build", "test_capture.go")
	envs := append([]string{
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=",
		"LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=false",
		"OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true",
	}, env...)
	RunApp(t, "test_capture", envs...)
}
