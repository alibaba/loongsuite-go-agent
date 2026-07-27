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

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("nethttp-basic-test", "nethttp", "", "", "1.18", "", TestBasicNetHttp),
		NewGeneralTestCase("nethttp-http-2-test", "nethttp", "", "", "1.18", "", TestHttp2),
		NewGeneralTestCase("nethttp-https-test", "nethttp", "", "", "1.18", "", TestHttps),
		NewGeneralTestCase("nethttp-metric-test", "nethttp", "", "", "1.18", "", TestHttpMetric),
		NewGeneralTestCase("nethttp-route-pattern-test", "nethttp", "", "", "1.23", "", TestHttpRoutePattern),
		NewGeneralTestCase("nethttp-route-static-pattern-test", "nethttp", "", "", "1.23", "", TestHttpRouteStaticPattern),
		NewGeneralTestCase("nethttp-route-cases-test", "nethttp", "", "", "1.23", "", TestHttpRouteCases),
		NewGeneralTestCase("nethttp-route-metrics-test", "nethttp", "", "", "1.23", "", TestHttpRouteMetrics),
		NewGeneralTestCase("nethttp-route-static-metrics-test", "nethttp", "", "", "1.23", "", TestHttpRouteStaticMetrics),
		NewGeneralTestCase("nethttp-route-unavailable-test", "nethttp", "", "", "1.18", "", TestHttpRouteUnavailable),
		NewGeneralTestCase("nethttp-route-unavailable-metrics-test", "nethttp", "", "", "1.18", "", TestHttpRouteUnavailableMetrics),
	)
}

func TestBasicNetHttp(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http.go", "http_server.go")
	RunApp(t, "test_http", env...)
}

func TestHttp2(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_2.go", "http_server.go")
	RunApp(t, "test_http_2", env...)
}

func TestHttps(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_https.go", "http_server.go")
	RunApp(t, "test_https", env...)
}

func TestHttpMetric(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_metrics.go", "http_server.go")
	RunApp(t, "test_http_metrics", env...)
}

func TestHttpRoutePattern(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_pattern.go", "http_server.go")
	RunApp(t, "test_http_route_pattern", env...)
}

func TestHttpRouteStaticPattern(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_static_pattern.go", "http_server.go")
	RunApp(t, "test_http_route_static_pattern", env...)
}

func TestHttpRouteCases(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_cases.go", "http_server.go")
	RunApp(t, "test_http_route_cases", env...)
}

func TestHttpRouteMetrics(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_metrics.go", "http_server.go")
	RunApp(t, "test_http_route_metrics", env...)
}

func TestHttpRouteStaticMetrics(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_static_metrics.go", "http_server.go")
	RunApp(t, "test_http_route_static_metrics", env...)
}

func TestHttpRouteUnavailable(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_unavailable.go", "http_server.go")
	RunApp(t, "test_http_route_unavailable", env...)
}

func TestHttpRouteUnavailableMetrics(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunGoBuild(t, "go", "build", "test_http_route_unavailable_metrics.go", "http_server.go")
	RunApp(t, "test_http_route_unavailable_metrics", env...)
}
