// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fasthttp

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"os"
	"strconv"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var emptyFastHttpResponse = fastHttpResponse{}

type fastHttpInnerEnabler struct {
	enabled bool
}

func (g fastHttpInnerEnabler) Enable() bool {
	return g.enabled
}

var fastHttpEnabler = fastHttpInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_FASTHTTP_ENABLED") != "false"}

type fastHttpClientAttrsGetter struct {
}

func (n fastHttpClientAttrsGetter) GetRequestMethod(request fastHttpRequest) string {
	return request.method
}
func (n fastHttpClientAttrsGetter) GetHttpRequestHeader(request fastHttpRequest, name string) []string {
	all := make([]string, 0)
	for _, header := range request.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fastHttpClientAttrsGetter) GetHttpResponseStatusCode(request fastHttpRequest, response fastHttpResponse, err error) int {
	return response.statusCode
}

func (n fastHttpClientAttrsGetter) HasHttpResponse(request fastHttpRequest, response fastHttpResponse, err error) bool {
	return response.statusCode >= 100
}

func (n fastHttpClientAttrsGetter) GetHttpResponseHeader(request fastHttpRequest, response fastHttpResponse, name string) []string {
	all := make([]string, 0)
	for _, header := range response.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fastHttpClientAttrsGetter) GetErrorType(request fastHttpRequest, response fastHttpResponse, err error) string {
	return ""
}
func (n fastHttpClientAttrsGetter) GetNetworkType(request fastHttpRequest, response fastHttpResponse) string {
	return "ipv4"
}
func (n fastHttpClientAttrsGetter) GetNetworkTransport(request fastHttpRequest, response fastHttpResponse) string {
	return "tcp"
}
func (n fastHttpClientAttrsGetter) GetNetworkProtocolName(request fastHttpRequest, response fastHttpResponse) string {
	if request.isTls == false {
		return "http"
	} else {
		return "https"
	}
}
func (n fastHttpClientAttrsGetter) GetNetworkProtocolVersion(request fastHttpRequest, response fastHttpResponse) string {
	return ""
}
func (n fastHttpClientAttrsGetter) GetNetworkLocalInetAddress(request fastHttpRequest, response fastHttpResponse) string {
	return ""
}
func (n fastHttpClientAttrsGetter) GetNetworkLocalPort(request fastHttpRequest, response fastHttpResponse) int {
	return 0
}
func (n fastHttpClientAttrsGetter) GetNetworkPeerInetAddress(request fastHttpRequest, response fastHttpResponse) string {
	return request.url.Host
}
func (n fastHttpClientAttrsGetter) GetNetworkPeerPort(request fastHttpRequest, response fastHttpResponse) int {
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}
func (n fastHttpClientAttrsGetter) GetUrlFull(request fastHttpRequest) string {
	return request.url.String()
}
func (n fastHttpClientAttrsGetter) GetServerAddress(request fastHttpRequest) string {
	return request.url.Host
}

func (n fastHttpClientAttrsGetter) GetServerPort(request fastHttpRequest) int {
	return n.GetNetworkPeerPort(request, emptyFastHttpResponse)
}

type fastHttpServerAttrsGetter struct {
}

func (n fastHttpServerAttrsGetter) GetRequestMethod(request fastHttpRequest) string {
	return request.method
}
func (n fastHttpServerAttrsGetter) GetHttpRequestHeader(request fastHttpRequest, name string) []string {
	all := make([]string, 0)
	for _, header := range request.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fastHttpServerAttrsGetter) GetHttpResponseStatusCode(request fastHttpRequest, response fastHttpResponse, err error) int {
	return response.statusCode
}
func (n fastHttpServerAttrsGetter) GetHttpResponseHeader(request fastHttpRequest, response fastHttpResponse, name string) []string {
	all := make([]string, 0)
	for _, header := range response.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fastHttpServerAttrsGetter) GetErrorType(request fastHttpRequest, response fastHttpResponse, err error) string {
	return ""
}
func (n fastHttpServerAttrsGetter) GetUrlScheme(request fastHttpRequest) string {
	return request.url.Scheme
}
func (n fastHttpServerAttrsGetter) GetUrlPath(request fastHttpRequest) string {
	return request.url.Path
}
func (n fastHttpServerAttrsGetter) GetUrlQuery(request fastHttpRequest) string {
	return request.url.RawQuery
}
func (n fastHttpServerAttrsGetter) GetNetworkType(request fastHttpRequest, response fastHttpResponse) string {
	return "ipv4"
}
func (n fastHttpServerAttrsGetter) GetNetworkTransport(request fastHttpRequest, response fastHttpResponse) string {
	return "tcp"
}
func (n fastHttpServerAttrsGetter) GetNetworkProtocolName(request fastHttpRequest, response fastHttpResponse) string {
	if request.isTls == false {
		return "http"
	} else {
		return "https"
	}
}
func (n fastHttpServerAttrsGetter) GetNetworkProtocolVersion(request fastHttpRequest, response fastHttpResponse) string {
	return ""
}
func (n fastHttpServerAttrsGetter) GetNetworkLocalInetAddress(request fastHttpRequest, response fastHttpResponse) string {
	return ""
}
func (n fastHttpServerAttrsGetter) GetNetworkLocalPort(request fastHttpRequest, response fastHttpResponse) int {
	return 0
}
func (n fastHttpServerAttrsGetter) GetNetworkPeerInetAddress(request fastHttpRequest, response fastHttpResponse) string {
	return request.url.Host
}
func (n fastHttpServerAttrsGetter) GetNetworkPeerPort(request fastHttpRequest, response fastHttpResponse) int {
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}
func (n fastHttpServerAttrsGetter) GetHttpRoute(request fastHttpRequest) string {
	return request.url.Path
}

type fastHttpRequestCarrier struct {
	req *fasthttp.RequestHeader
}

func (f fastHttpRequestCarrier) Get(key string) string {
	return string(f.req.Peek(key))
}
func (f fastHttpRequestCarrier) Set(key string, value string) {
	f.req.Set(key, value)
}
func (f fastHttpRequestCarrier) Keys() []string {
	keyStrs := make([]string, 0)
	peekKeys := f.req.PeekKeys()
	for _, peekKey := range peekKeys {
		keyStrs = append(keyStrs, string(peekKey))
	}
	return keyStrs
}
func BuildFastHttpClientOtelInstrumenter() *instrumenter.PropagatingToDownstreamInstrumenter[fastHttpRequest, fastHttpResponse] {
	builder := instrumenter.Builder[fastHttpRequest, fastHttpResponse]{}
	clientGetter := fastHttpClientAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpClientAttrsGetter, fastHttpClientAttrsGetter]{HttpGetter: clientGetter, NetGetter: clientGetter}
	networkExtractor := net.NetworkAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpClientAttrsGetter]{Getter: clientGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpClientSpanStatusExtractor[fastHttpRequest, fastHttpResponse]{Getter: clientGetter}).SetSpanNameExtractor(&http.HttpClientSpanNameExtractor[fastHttpRequest, fastHttpResponse]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[fastHttpRequest]{}).
		AddOperationListeners(http.HttpClientMetrics("fasthttp.client")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.FAST_HTTP_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(
			&http.HttpClientAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpClientAttrsGetter, fastHttpClientAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor},
			&http.CaptureAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpCaptureAttrsGetter]{Getter: fastHttpCaptureAttrsGetter{}},
		).BuildPropagatingToDownstreamInstrumenter(func(n fastHttpRequest) propagation.TextMapCarrier {
		return fastHttpRequestCarrier{req: n.header}
	}, otel.GetTextMapPropagator())
}
func BuildFastHttpServerOtelInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[fastHttpRequest, fastHttpResponse] {
	builder := instrumenter.Builder[fastHttpRequest, fastHttpResponse]{}
	serverGetter := fastHttpServerAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpServerAttrsGetter, fastHttpServerAttrsGetter]{HttpGetter: serverGetter, NetGetter: serverGetter}
	networkExtractor := net.NetworkAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpServerAttrsGetter]{Getter: serverGetter}
	urlExtractor := net.UrlAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpServerAttrsGetter]{Getter: serverGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpServerSpanStatusExtractor[fastHttpRequest, fastHttpResponse]{Getter: serverGetter}).SetSpanNameExtractor(&http.HttpServerSpanNameExtractor[fastHttpRequest, fastHttpResponse]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[fastHttpRequest]{}).
		AddOperationListeners(http.HttpServerMetrics("fasthttp.server")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.FAST_HTTP_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(
			&http.HttpServerAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpServerAttrsGetter, fastHttpServerAttrsGetter, fastHttpServerAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor, UrlExtractor: urlExtractor},
			&http.CaptureAttrsExtractor[fastHttpRequest, fastHttpResponse, fastHttpCaptureAttrsGetter]{Getter: fastHttpCaptureAttrsGetter{}},
		).BuildPropagatingFromUpstreamInstrumenter(func(n fastHttpRequest) propagation.TextMapCarrier {
		return fastHttpRequestCarrier{req: n.header}
	}, otel.GetTextMapPropagator())
}
