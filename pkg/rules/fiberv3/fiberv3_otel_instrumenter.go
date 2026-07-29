// Copyright (c) 2026 Alibaba Group Holding Ltd.
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
package fiberv3

import (
	"os"
	"strconv"

	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var emptyfiberv3Response = fiberv3Response{}

type fiberV3InnerEnabler struct {
	enabled bool
}

func (g fiberV3InnerEnabler) Enable() bool {
	return g.enabled
}

var fiberV3Enabler = fiberV3InnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_FIBERV3_ENABLED") != "false"}

type fiberv3ServerAttrsGetter struct {
}

func (n fiberv3ServerAttrsGetter) GetRequestMethod(request *fiberv3Request) string {
	return request.method
}
func (n fiberv3ServerAttrsGetter) GetHttpRequestHeader(request *fiberv3Request, name string) []string {
	all := make([]string, 0)
	for _, header := range request.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fiberv3ServerAttrsGetter) GetHttpResponseStatusCode(request *fiberv3Request, response *fiberv3Response, err error) int {
	return response.statusCode
}
func (n fiberv3ServerAttrsGetter) GetHttpResponseHeader(request *fiberv3Request, response *fiberv3Response, name string) []string {
	all := make([]string, 0)
	for _, header := range response.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fiberv3ServerAttrsGetter) GetErrorType(request *fiberv3Request, response *fiberv3Response, err error) string {
	return ""
}
func (n fiberv3ServerAttrsGetter) GetUrlScheme(request *fiberv3Request) string {
	return request.url.Scheme
}
func (n fiberv3ServerAttrsGetter) GetUrlPath(request *fiberv3Request) string {
	return request.url.Path
}
func (n fiberv3ServerAttrsGetter) GetUrlQuery(request *fiberv3Request) string {
	return request.url.RawQuery
}
func (n fiberv3ServerAttrsGetter) GetNetworkType(request *fiberv3Request, response *fiberv3Response) string {
	return "ipv4"
}
func (n fiberv3ServerAttrsGetter) GetNetworkTransport(request *fiberv3Request, response *fiberv3Response) string {
	return "tcp"
}
func (n fiberv3ServerAttrsGetter) GetNetworkProtocolName(request *fiberv3Request, response *fiberv3Response) string {
	if !request.isTls {
		return "http"
	}
	return "https"
}
func (n fiberv3ServerAttrsGetter) GetNetworkProtocolVersion(request *fiberv3Request, response *fiberv3Response) string {
	return ""
}
func (n fiberv3ServerAttrsGetter) GetNetworkLocalInetAddress(request *fiberv3Request, response *fiberv3Response) string {
	return ""
}
func (n fiberv3ServerAttrsGetter) GetNetworkLocalPort(request *fiberv3Request, response *fiberv3Response) int {
	return 0
}
func (n fiberv3ServerAttrsGetter) GetNetworkPeerInetAddress(request *fiberv3Request, response *fiberv3Response) string {
	return request.url.Host
}
func (n fiberv3ServerAttrsGetter) GetNetworkPeerPort(request *fiberv3Request, response *fiberv3Response) int {
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}
func (n fiberv3ServerAttrsGetter) GetHttpRoute(request *fiberv3Request) string {
	return request.url.Path
}

type fiberv3RequestCarrier struct {
	req *fasthttp.RequestHeader
}

func (f fiberv3RequestCarrier) Get(key string) string {
	return string(f.req.Peek(key))
}
func (f fiberv3RequestCarrier) Set(key string, value string) {
	f.req.Set(key, value)
}
func (f fiberv3RequestCarrier) Keys() []string {
	keyStrs := make([]string, 0)
	peekKeys := f.req.PeekKeys()
	for _, peekKey := range peekKeys {
		keyStrs = append(keyStrs, string(peekKey))
	}
	return keyStrs
}

func BuildFiberV3ServerOtelInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[*fiberv3Request, *fiberv3Response] {
	builder := instrumenter.Builder[*fiberv3Request, *fiberv3Response]{}
	serverGetter := fiberv3ServerAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[*fiberv3Request, *fiberv3Response, fiberv3ServerAttrsGetter, fiberv3ServerAttrsGetter]{HttpGetter: serverGetter, NetGetter: serverGetter}
	networkExtractor := net.NetworkAttrsExtractor[*fiberv3Request, *fiberv3Response, fiberv3ServerAttrsGetter]{Getter: serverGetter}
	urlExtractor := net.UrlAttrsExtractor[*fiberv3Request, *fiberv3Response, fiberv3ServerAttrsGetter]{Getter: serverGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpServerSpanStatusExtractor[*fiberv3Request, *fiberv3Response]{Getter: serverGetter}).SetSpanNameExtractor(&http.HttpServerSpanNameExtractor[*fiberv3Request, *fiberv3Response]{Getter: serverGetter}).
		AddOperationListeners(http.HttpServerMetrics("fiberv3.server")).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[*fiberv3Request]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.FIBER_V3_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(
			&http.HttpServerAttrsExtractor[*fiberv3Request, *fiberv3Response, fiberv3ServerAttrsGetter, fiberv3ServerAttrsGetter, fiberv3ServerAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor, UrlExtractor: urlExtractor},
			&http.CaptureAttrsExtractor[*fiberv3Request, *fiberv3Response, fiberV3CaptureAttrsGetter]{Getter: fiberV3CaptureAttrsGetter{}},
		).BuildPropagatingFromUpstreamInstrumenter(func(n *fiberv3Request) propagation.TextMapCarrier {
		return fiberv3RequestCarrier{req: n.header}
	}, otel.GetTextMapPropagator())
}
