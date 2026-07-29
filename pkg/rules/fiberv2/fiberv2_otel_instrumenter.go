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
package fiberv2

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

var emptyFiberv2Response = fiberv2Response{}

type fiberV2InnerEnabler struct {
	enabled bool
}

func (g fiberV2InnerEnabler) Enable() bool {
	return g.enabled
}

var fiberV2Enabler = fiberV2InnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_FIBERV2_ENABLED") != "false"}

type fiberv2ServerAttrsGetter struct {
}

func (n fiberv2ServerAttrsGetter) GetRequestMethod(request *fiberv2Request) string {
	return request.method
}
func (n fiberv2ServerAttrsGetter) GetHttpRequestHeader(request *fiberv2Request, name string) []string {
	all := make([]string, 0)
	for _, header := range request.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fiberv2ServerAttrsGetter) GetHttpResponseStatusCode(request *fiberv2Request, response *fiberv2Response, err error) int {
	return response.statusCode
}
func (n fiberv2ServerAttrsGetter) GetHttpResponseHeader(request *fiberv2Request, response *fiberv2Response, name string) []string {
	all := make([]string, 0)
	for _, header := range response.header.PeekAll(name) {
		all = append(all, string(header))
	}
	return all
}
func (n fiberv2ServerAttrsGetter) GetErrorType(request *fiberv2Request, response *fiberv2Response, err error) string {
	return ""
}
func (n fiberv2ServerAttrsGetter) GetUrlScheme(request *fiberv2Request) string {
	return request.url.Scheme
}
func (n fiberv2ServerAttrsGetter) GetUrlPath(request *fiberv2Request) string {
	return request.url.Path
}
func (n fiberv2ServerAttrsGetter) GetUrlQuery(request *fiberv2Request) string {
	return request.url.RawQuery
}
func (n fiberv2ServerAttrsGetter) GetNetworkType(request *fiberv2Request, response *fiberv2Response) string {
	return "ipv4"
}
func (n fiberv2ServerAttrsGetter) GetNetworkTransport(request *fiberv2Request, response *fiberv2Response) string {
	return "tcp"
}
func (n fiberv2ServerAttrsGetter) GetNetworkProtocolName(request *fiberv2Request, response *fiberv2Response) string {
	if !request.isTls {
		return "http"
	}
	return "https"
}
func (n fiberv2ServerAttrsGetter) GetNetworkProtocolVersion(request *fiberv2Request, response *fiberv2Response) string {
	return ""
}
func (n fiberv2ServerAttrsGetter) GetNetworkLocalInetAddress(request *fiberv2Request, response *fiberv2Response) string {
	return ""
}
func (n fiberv2ServerAttrsGetter) GetNetworkLocalPort(request *fiberv2Request, response *fiberv2Response) int {
	return 0
}
func (n fiberv2ServerAttrsGetter) GetNetworkPeerInetAddress(request *fiberv2Request, response *fiberv2Response) string {
	return request.url.Host
}
func (n fiberv2ServerAttrsGetter) GetNetworkPeerPort(request *fiberv2Request, response *fiberv2Response) int {
	port, err := strconv.Atoi(request.url.Port())
	if err != nil {
		return 0
	}
	return port
}
func (n fiberv2ServerAttrsGetter) GetHttpRoute(request *fiberv2Request) string {
	return request.url.Path
}

type fiberv2RequestCarrier struct {
	req *fasthttp.RequestHeader
}

func (f fiberv2RequestCarrier) Get(key string) string {
	return string(f.req.Peek(key))
}
func (f fiberv2RequestCarrier) Set(key string, value string) {
	f.req.Set(key, value)
}
func (f fiberv2RequestCarrier) Keys() []string {
	keyStrs := make([]string, 0)
	peekKeys := f.req.PeekKeys()
	for _, peekKey := range peekKeys {
		keyStrs = append(keyStrs, string(peekKey))
	}
	return keyStrs
}

func BuildFiberV2ServerOtelInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[*fiberv2Request, *fiberv2Response] {
	builder := instrumenter.Builder[*fiberv2Request, *fiberv2Response]{}
	serverGetter := fiberv2ServerAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[*fiberv2Request, *fiberv2Response, fiberv2ServerAttrsGetter, fiberv2ServerAttrsGetter]{HttpGetter: serverGetter, NetGetter: serverGetter}
	networkExtractor := net.NetworkAttrsExtractor[*fiberv2Request, *fiberv2Response, fiberv2ServerAttrsGetter]{Getter: serverGetter}
	urlExtractor := net.UrlAttrsExtractor[*fiberv2Request, *fiberv2Response, fiberv2ServerAttrsGetter]{Getter: serverGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpServerSpanStatusExtractor[*fiberv2Request, *fiberv2Response]{Getter: serverGetter}).SetSpanNameExtractor(&http.HttpServerSpanNameExtractor[*fiberv2Request, *fiberv2Response]{Getter: serverGetter}).
		AddOperationListeners(http.HttpServerMetrics("fiberv2.server")).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[*fiberv2Request]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.FIBER_V2_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(
			&http.HttpServerAttrsExtractor[*fiberv2Request, *fiberv2Response, fiberv2ServerAttrsGetter, fiberv2ServerAttrsGetter, fiberv2ServerAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor, UrlExtractor: urlExtractor},
			&http.CaptureAttrsExtractor[*fiberv2Request, *fiberv2Response, fiberV2CaptureAttrsGetter]{Getter: fiberV2CaptureAttrsGetter{}},
		).BuildPropagatingFromUpstreamInstrumenter(func(n *fiberv2Request) propagation.TextMapCarrier {
		return fiberv2RequestCarrier{req: n.header}
	}, otel.GetTextMapPropagator())
}
