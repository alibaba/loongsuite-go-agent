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

package http

import (
	"os"
	"strconv"

	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type netHttpInnerEnabler struct {
	enabled bool
}

func (n netHttpInnerEnabler) Enable() bool {
	return n.enabled
}

var netHttpEnabler = netHttpInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_NETHTTP_ENABLED") != "false"}

var emptyHttpResponse = netHttpResponse{}

type netHttpClientAttrsGetter struct {
}

func (n netHttpClientAttrsGetter) GetRequestMethod(request *netHttpRequest) string {
	return request.method
}

func (n netHttpClientAttrsGetter) GetHttpRequestHeader(request *netHttpRequest, name string) []string {
	return request.header.Values(name)
}

func (n netHttpClientAttrsGetter) GetHttpResponseStatusCode(request *netHttpRequest, response *netHttpResponse, err error) int {
	return response.statusCode
}

func (n netHttpClientAttrsGetter) HasHttpResponse(request *netHttpRequest, response *netHttpResponse, err error) bool {
	return response != nil && response.hasResponse
}

func (n netHttpClientAttrsGetter) GetHttpResponseHeader(request *netHttpRequest, response *netHttpResponse, name string) []string {
	return response.header.Values(name)
}

func (n netHttpClientAttrsGetter) GetErrorType(request *netHttpRequest, response *netHttpResponse, err error) string {
	// TODO return status code as error type
	return ""
}

func (n netHttpClientAttrsGetter) GetNetworkType(request *netHttpRequest, response *netHttpResponse) string {
	return "ipv4"
}

func (n netHttpClientAttrsGetter) GetNetworkTransport(request *netHttpRequest, response *netHttpResponse) string {
	return "tcp"
}

func (n netHttpClientAttrsGetter) GetNetworkProtocolName(request *netHttpRequest, response *netHttpResponse) string {
	if request.isTls == false {
		return "http"
	} else {
		return "https"
	}
}

func (n netHttpClientAttrsGetter) GetNetworkProtocolVersion(request *netHttpRequest, response *netHttpResponse) string {
	return request.version
}

func (n netHttpClientAttrsGetter) GetNetworkLocalInetAddress(request *netHttpRequest, response *netHttpResponse) string {
	return ""
}

func (n netHttpClientAttrsGetter) GetNetworkLocalPort(request *netHttpRequest, response *netHttpResponse) int {
	return 0
}

func (n netHttpClientAttrsGetter) GetNetworkPeerInetAddress(request *netHttpRequest, response *netHttpResponse) string {
	return request.host
}

func (n netHttpClientAttrsGetter) GetNetworkPeerPort(request *netHttpRequest, response *netHttpResponse) int {
	if request.url == nil {
		return 0
	}
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}

func (n netHttpClientAttrsGetter) GetUrlFull(request *netHttpRequest) string {
	return request.url.String()
}

func (n netHttpClientAttrsGetter) GetServerAddress(request *netHttpRequest) string {
	return request.host
}

func (n netHttpClientAttrsGetter) GetServerPort(request *netHttpRequest) int {
	if request.url == nil {
		return 0
	}
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}

type netHttpServerAttrsGetter struct {
}

func (n netHttpServerAttrsGetter) GetRequestMethod(request *netHttpRequest) string {
	return request.method
}

func (n netHttpServerAttrsGetter) GetHttpRequestHeader(request *netHttpRequest, name string) []string {
	return request.header.Values(name)
}

func (n netHttpServerAttrsGetter) GetHttpResponseStatusCode(request *netHttpRequest, response *netHttpResponse, err error) int {
	return response.statusCode
}

func (n netHttpServerAttrsGetter) GetHttpResponseHeader(request *netHttpRequest, response *netHttpResponse, name string) []string {
	return response.header.Values(name)
}

func (n netHttpServerAttrsGetter) GetErrorType(request *netHttpRequest, response *netHttpResponse, err error) string {
	// TODO return status code as error type
	return ""
}

func (n netHttpServerAttrsGetter) GetUrlScheme(request *netHttpRequest) string {
	if request.url.Scheme != "" {
		return request.url.Scheme
	}
	return n.GetNetworkProtocolName(request, &emptyHttpResponse)
}

func (n netHttpServerAttrsGetter) GetUrlPath(request *netHttpRequest) string {
	return request.url.Path
}

func (n netHttpServerAttrsGetter) GetUrlQuery(request *netHttpRequest) string {
	return request.url.RawQuery
}

func (n netHttpServerAttrsGetter) GetNetworkType(request *netHttpRequest, response *netHttpResponse) string {
	return "ipv4"
}

func (n netHttpServerAttrsGetter) GetNetworkTransport(request *netHttpRequest, response *netHttpResponse) string {
	return "tcp"
}

func (n netHttpServerAttrsGetter) GetNetworkProtocolName(request *netHttpRequest, response *netHttpResponse) string {
	if !request.isTls {
		return "http"
	}
	return "https"
}

func (n netHttpServerAttrsGetter) GetNetworkProtocolVersion(request *netHttpRequest, response *netHttpResponse) string {
	return request.version
}

func (n netHttpServerAttrsGetter) GetNetworkLocalInetAddress(request *netHttpRequest, response *netHttpResponse) string {
	return ""
}

func (n netHttpServerAttrsGetter) GetNetworkLocalPort(request *netHttpRequest, response *netHttpResponse) int {
	return 0
}

func (n netHttpServerAttrsGetter) GetNetworkPeerInetAddress(request *netHttpRequest, response *netHttpResponse) string {
	return request.host
}

func (n netHttpServerAttrsGetter) GetNetworkPeerPort(request *netHttpRequest, response *netHttpResponse) int {
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}

func (n netHttpServerAttrsGetter) GetHttpRoute(request *netHttpRequest) string {
	return request.url.Path
}

func BuildNetHttpClientOtelInstrumenter() *instrumenter.PropagatingToDownstreamInstrumenter[*netHttpRequest, *netHttpResponse] {
	builder := &instrumenter.Builder[*netHttpRequest, *netHttpResponse]{}
	clientGetter := netHttpClientAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[*netHttpRequest, *netHttpResponse, http.HttpClientAttrsGetter[*netHttpRequest, *netHttpResponse], net.NetworkAttrsGetter[*netHttpRequest, *netHttpResponse]]{HttpGetter: clientGetter, NetGetter: clientGetter}
	networkExtractor := net.NetworkAttrsExtractor[*netHttpRequest, *netHttpResponse, net.NetworkAttrsGetter[*netHttpRequest, *netHttpResponse]]{Getter: clientGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpClientSpanStatusExtractor[*netHttpRequest, *netHttpResponse]{Getter: clientGetter}).SetSpanNameExtractor(&http.HttpClientSpanNameExtractor[*netHttpRequest, *netHttpResponse]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[*netHttpRequest]{}).
		AddOperationListeners(http.HttpClientMetrics("net.http.client")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.NET_HTTP_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(
			&http.HttpClientAttrsExtractor[*netHttpRequest, *netHttpResponse, http.HttpClientAttrsGetter[*netHttpRequest, *netHttpResponse], net.NetworkAttrsGetter[*netHttpRequest, *netHttpResponse]]{Base: commonExtractor, NetworkExtractor: networkExtractor},
			&httpCaptureAttrsExtractor{},
		).BuildPropagatingToDownstreamInstrumenter(func(n *netHttpRequest) propagation.TextMapCarrier {
		if n.header == nil {
			return nil
		}
		return propagation.HeaderCarrier(n.header)
	}, otel.GetTextMapPropagator())
}

func BuildNetHttpServerOtelInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[*netHttpRequest, *netHttpResponse] {
	builder := &instrumenter.Builder[*netHttpRequest, *netHttpResponse]{}
	serverGetter := netHttpServerAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[*netHttpRequest, *netHttpResponse, http.HttpServerAttrsGetter[*netHttpRequest, *netHttpResponse], net.NetworkAttrsGetter[*netHttpRequest, *netHttpResponse]]{HttpGetter: serverGetter, NetGetter: serverGetter}
	networkExtractor := net.NetworkAttrsExtractor[*netHttpRequest, *netHttpResponse, net.NetworkAttrsGetter[*netHttpRequest, *netHttpResponse]]{Getter: serverGetter}
	urlExtractor := net.UrlAttrsExtractor[*netHttpRequest, *netHttpResponse, net.UrlAttrsGetter[*netHttpRequest]]{Getter: serverGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpServerSpanStatusExtractor[*netHttpRequest, *netHttpResponse]{Getter: serverGetter}).SetSpanNameExtractor(&http.HttpServerSpanNameExtractor[*netHttpRequest, *netHttpResponse]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[*netHttpRequest]{}).
		AddOperationListeners(http.HttpServerMetrics("net.http.server")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.NET_HTTP_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(
			&http.HttpServerAttrsExtractor[*netHttpRequest, *netHttpResponse, http.HttpServerAttrsGetter[*netHttpRequest, *netHttpResponse], net.NetworkAttrsGetter[*netHttpRequest, *netHttpResponse], net.UrlAttrsGetter[*netHttpRequest]]{Base: commonExtractor, NetworkExtractor: networkExtractor, UrlExtractor: urlExtractor},
			&httpCaptureAttrsExtractor{},
		).BuildPropagatingFromUpstreamInstrumenter(func(n *netHttpRequest) propagation.TextMapCarrier {
		if n.header == nil {
			return nil
		}
		return propagation.HeaderCarrier(n.header)
	}, otel.GetTextMapPropagator())
}
