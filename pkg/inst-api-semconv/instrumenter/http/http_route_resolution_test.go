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

package http

import (
	"context"
	"testing"
	"time"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/net"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
)

type resolveRouteGetter struct {
	route string
	path  string
}

func (g resolveRouteGetter) GetRequestMethod(request testRequest) string {
	return request.Method
}
func (g resolveRouteGetter) GetHttpRequestHeader(testRequest, string) []string {
	return nil
}
func (g resolveRouteGetter) GetHttpResponseStatusCode(testRequest, testResponse, error) int {
	return 200
}
func (g resolveRouteGetter) GetHttpResponseHeader(testRequest, testResponse, string) []string {
	return nil
}
func (g resolveRouteGetter) GetErrorType(testRequest, testResponse, error) string { return "" }
func (g resolveRouteGetter) GetUrlScheme(testRequest) string                      { return "http" }
func (g resolveRouteGetter) GetUrlPath(testRequest) string                        { return g.path }
func (g resolveRouteGetter) GetUrlQuery(testRequest) string                       { return "" }
func (g resolveRouteGetter) GetNetworkType(testRequest, testResponse) string      { return "ipv4" }
func (g resolveRouteGetter) GetNetworkTransport(testRequest, testResponse) string { return "tcp" }
func (g resolveRouteGetter) GetNetworkProtocolName(testRequest, testResponse) string {
	return "http"
}
func (g resolveRouteGetter) GetNetworkProtocolVersion(testRequest, testResponse) string {
	return "1.1"
}
func (g resolveRouteGetter) GetNetworkLocalInetAddress(testRequest, testResponse) string {
	return ""
}
func (g resolveRouteGetter) GetNetworkLocalPort(testRequest, testResponse) int { return 0 }
func (g resolveRouteGetter) GetNetworkPeerInetAddress(testRequest, testResponse) string {
	return ""
}
func (g resolveRouteGetter) GetNetworkPeerPort(testRequest, testResponse) int { return 0 }
func (g resolveRouteGetter) GetHttpRoute(testRequest) string                  { return g.route }

func newRouteAwareServerExtractor(getter resolveRouteGetter) HttpServerAttrsExtractor[testRequest, testResponse, resolveRouteGetter, resolveRouteGetter, resolveRouteGetter] {
	return HttpServerAttrsExtractor[testRequest, testResponse, resolveRouteGetter, resolveRouteGetter, resolveRouteGetter]{
		Base: HttpCommonAttrsExtractor[testRequest, testResponse, resolveRouteGetter, resolveRouteGetter]{
			HttpGetter: getter,
			NetGetter:  resolveRouteGetter{},
		},
		NetworkExtractor: net.NetworkAttrsExtractor[testRequest, testResponse, resolveRouteGetter]{Getter: resolveRouteGetter{}},
		UrlExtractor:     net.UrlAttrsExtractor[testRequest, testResponse, resolveRouteGetter]{Getter: getter},
	}
}

func httpRouteFromAttrs(attrs []attribute.KeyValue) (string, bool) {
	for _, attr := range attrs {
		if attr.Key == semconv.HTTPRouteKey {
			return attr.Value.AsString(), true
		}
	}
	return "", false
}

func TestResolveHttpServerRoute(t *testing.T) {
	tests := []struct {
		name     string
		route    string
		path     string
		expected string
	}{
		{name: "nethttp_restful", route: "/users/{id}", path: "/users/123", expected: "/users/{id}"},
		{name: "gin_restful", route: "/user/:name", path: "/user/abc", expected: "/user/:name"},
		{name: "mux_restful", route: "/{name}/countries/{country}", path: "/1/countries/2", expected: "/{name}/countries/{country}"},
		{name: "echo_restful", route: "/users/:id", path: "/users/1", expected: "/users/:id"},
		{name: "gorestful_restful", route: "/users/{user-id}", path: "/users/123", expected: "/users/{user-id}"},
		{name: "fiber_restful", route: "/users/:id", path: "/users/123", expected: "/users/:id"},
		{name: "hertz_restful", route: "/hertz/:version/*action", path: "/hertz/v2/send", expected: "/hertz/:version/*action"},
		{name: "static_route", route: "/query", path: "/query", expected: "/query"},
		{name: "static_root", route: "/", path: "/", expected: "/"},
		{name: "no_template_restful", route: "", path: "/users/123", expected: ""},
		{name: "no_template_static", route: "", path: "/health", expected: ""},
		{name: "no_template_fasthttp", route: "", path: "/items/7", expected: ""},
		{name: "empty_both", route: "", path: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := resolveRouteGetter{route: tt.route, path: tt.path}
			if got := ResolveHttpServerRoute(getter, testRequest{}); got != tt.expected {
				t.Fatalf("ResolveHttpServerRoute() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHttpServerExtractorEndResolveHttpRouteCases(t *testing.T) {
	tests := []struct {
		name        string
		route       string
		path        string
		recording   bool
		wantRoute   string
		wantPresent bool
	}{
		{name: "nethttp_restful_template", route: "/users/{id}", path: "/users/123", recording: true, wantRoute: "/users/{id}", wantPresent: true},
		{name: "gin_restful_template", route: "/user/:name", path: "/user/abc", recording: true, wantRoute: "/user/:name", wantPresent: true},
		{name: "mux_restful_template", route: "/test/{key}", path: "/test/1", recording: true, wantRoute: "/test/{key}", wantPresent: true},
		{name: "echo_restful_template", route: "/users/:id", path: "/users/1", recording: true, wantRoute: "/users/:id", wantPresent: true},
		{name: "gorestful_template", route: "/users/{user-id}", path: "/users/123", recording: true, wantRoute: "/users/{user-id}", wantPresent: true},
		{name: "fiber_template", route: "/users/:id", path: "/users/123", recording: true, wantRoute: "/users/:id", wantPresent: true},
		{name: "hertz_template", route: "/hertz/:version/*action", path: "/hertz/v1/", recording: true, wantRoute: "/hertz/:version/*action", wantPresent: true},
		{name: "static_route_template", route: "/query", path: "/query", recording: true, wantRoute: "/query", wantPresent: true},
		{name: "static_root", route: "/", path: "/", recording: true, wantRoute: "/", wantPresent: true},
		{name: "bare_handler_omits_route", route: "", path: "/users/123", recording: true, wantPresent: false},
		{name: "bare_handler_static_omits_route", route: "", path: "/a", recording: true, wantPresent: false},
		{name: "fasthttp_omits_route", route: "", path: "/items/7", recording: true, wantPresent: false},
		{name: "empty_path_and_template", route: "", path: "", recording: true, wantPresent: false},
		{name: "non_recording_span", route: "/users/{id}", path: "/users/123", recording: false, wantRoute: "/users/{id}", wantPresent: true},
		{name: "template_preferred_over_path", route: "/users/{id}", path: "/users/999", recording: true, wantRoute: "/users/{id}", wantPresent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := resolveRouteGetter{route: tt.route, path: tt.path}
			extractor := newRouteAwareServerExtractor(getter)
			ctx := trace.ContextWithSpan(context.Background(), &testReadOnlySpan{
				isRecording: tt.recording,
			})
			attrs, _ := extractor.OnEnd(nil, ctx, testRequest{}, testResponse{}, nil)
			route, ok := httpRouteFromAttrs(attrs)
			if ok != tt.wantPresent {
				t.Fatalf("http.route present = %v, want %v", ok, tt.wantPresent)
			}
			if tt.wantPresent && route != tt.wantRoute {
				t.Fatalf("http.route = %q, want %q", route, tt.wantRoute)
			}
		})
	}
}

func TestHttpServerExtractorEndIncludesRouteWhenSpanNotRecording(t *testing.T) {
	getter := resolveRouteGetter{route: "/users/{id}", path: "/users/123"}
	extractor := newRouteAwareServerExtractor(getter)
	ctx := trace.ContextWithSpan(context.Background(), &testReadOnlySpan{isRecording: false})
	attrs, _ := extractor.OnEnd(nil, ctx, testRequest{}, testResponse{}, nil)
	route, ok := httpRouteFromAttrs(attrs)
	if !ok || route != "/users/{id}" {
		t.Fatalf("expected http.route=/users/{id} for metrics attrs, got %q present=%v", route, ok)
	}
}

func TestHttpServerSpanNameResolveRouteCases(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		route    string
		path     string
		expected string
	}{
		{name: "restful_template", method: "GET", route: "/users/{id}", path: "/users/123", expected: "GET /users/{id}"},
		{name: "gin_template", method: "GET", route: "/user/:name", path: "/user/abc", expected: "GET /user/:name"},
		{name: "static_template", method: "GET", route: "/query", path: "/query", expected: "GET /query"},
		{name: "no_template", method: "GET", route: "", path: "/users/123", expected: "GET"},
		{name: "static_no_template", method: "POST", route: "", path: "/a", expected: "POST"},
		{name: "method_only", method: "GET", route: "", path: "", expected: "GET"},
		{name: "empty_method", method: "", route: "/users/{id}", path: "/users/123", expected: "HTTP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := resolveRouteGetter{route: tt.route, path: tt.path}
			extractor := HttpServerSpanNameExtractor[testRequest, testResponse]{Getter: getter}
			if got := extractor.Extract(testRequest{Method: tt.method}); got != tt.expected {
				t.Fatalf("span name = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHttpServerMetricsOmitsHttpRouteWithoutTemplate(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	server, err := newHttpServerMetric("test", meter)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	start := time.Now()
	ctx = server.OnBeforeStart(ctx, start)
	ctx = server.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	server.OnAfterStart(ctx, start)
	server.OnAfterEnd(ctx, []attribute.KeyValue{
		{Key: semconv.HTTPRequestMethodKey, Value: attribute.StringValue("GET")},
		{Key: semconv.HTTPResponseStatusCodeKey, Value: attribute.IntValue(200)},
		{Key: semconv.NetworkProtocolNameKey, Value: attribute.StringValue("http")},
		{Key: semconv.NetworkProtocolVersionKey, Value: attribute.StringValue("1.1")},
		{Key: semconv.URLSchemeKey, Value: attribute.StringValue("http")},
	}, time.Now())

	rm := &metricdata.ResourceMetrics{}
	if err := reader.Collect(ctx, rm); err != nil {
		t.Fatal(err)
	}
	if len(rm.ScopeMetrics) == 0 || len(rm.ScopeMetrics[0].Metrics) == 0 {
		t.Fatal("expected http.server.request.duration metric")
	}
	point := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64]).DataPoints[0]
	attrs := point.Attributes.ToSlice()
	if _, ok := httpRouteFromAttrs(attrs); ok {
		t.Fatalf("expected metrics to omit http.route without template, got attrs=%v", attrs)
	}
}
