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

package client

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"net/url"
	"strconv"

	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func GetRequest(req *protocol.Request) (dst *protocol.Request) {
	dst = &protocol.Request{}
	req.CopyToSkipBody(dst)
	return
}

type hertzHttpClientAttrsGetter struct {
}

func (h hertzHttpClientAttrsGetter) GetRequestMethod(request *protocol.Request) string {
	return string(request.Method())
}

func (h hertzHttpClientAttrsGetter) GetHttpRequestHeader(request *protocol.Request, name string) []string {
	keys := make([]string, 0)
	request.Header.VisitAllCustomHeader(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func (h hertzHttpClientAttrsGetter) GetHttpResponseStatusCode(request *protocol.Request, response *protocol.Response, err error) int {
	return response.StatusCode()
}

func (h hertzHttpClientAttrsGetter) HasHttpResponse(request *protocol.Request, response *protocol.Response, err error) bool {
	return response != nil && response.StatusCode() >= 100
}

func (h hertzHttpClientAttrsGetter) GetHttpResponseHeader(request *protocol.Request, response *protocol.Response, name string) []string {
	keys := make([]string, 0)
	response.Header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func (h hertzHttpClientAttrsGetter) GetErrorType(request *protocol.Request, response *protocol.Response, err error) string {
	return ""
}

func (h hertzHttpClientAttrsGetter) GetNetworkType(request *protocol.Request, response *protocol.Response) string {
	return "ipv4"
}

func (h hertzHttpClientAttrsGetter) GetNetworkTransport(request *protocol.Request, response *protocol.Response) string {
	return "tcp"
}

func (h hertzHttpClientAttrsGetter) GetNetworkProtocolName(request *protocol.Request, response *protocol.Response) string {
	scheme := string(request.Scheme())
	if scheme != "" {
		return scheme
	}
	return "http"
}

func (h hertzHttpClientAttrsGetter) GetNetworkProtocolVersion(request *protocol.Request, response *protocol.Response) string {
	return ""
}

func (h hertzHttpClientAttrsGetter) GetNetworkLocalInetAddress(request *protocol.Request, response *protocol.Response) string {
	return ""
}

func (h hertzHttpClientAttrsGetter) GetNetworkLocalPort(request *protocol.Request, response *protocol.Response) int {
	return 0
}

func (h hertzHttpClientAttrsGetter) GetNetworkPeerInetAddress(request *protocol.Request, response *protocol.Response) string {
	return string(request.Host())
}

func (h hertzHttpClientAttrsGetter) GetNetworkPeerPort(request *protocol.Request, response *protocol.Response) int {
	return getPeerPort(request)
}

func (h hertzHttpClientAttrsGetter) GetUrlFull(request *protocol.Request) string {
	return string(request.RequestURI())
}

func (h hertzHttpClientAttrsGetter) GetServerAddress(request *protocol.Request) string {
	return string(request.Host())
}

func (h hertzHttpClientAttrsGetter) GetServerPort(request *protocol.Request) int {
	return getPeerPort(request)
}

func getPeerPort(request *protocol.Request) int {
	u, err := url.Parse(GetRequest(request).URI().String())
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return 0
	}
	return port
}

type hertzTextMapCarrier struct {
	request *protocol.Request
}

func (h hertzTextMapCarrier) Get(key string) string {
	return h.request.Header.Get(key)
}

func (h hertzTextMapCarrier) Set(key string, value string) {
	h.request.SetHeader(key, value)
}

func (h hertzTextMapCarrier) Keys() []string {
	keys := make([]string, 0)
	h.request.Header.VisitAllCustomHeader(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func BuildHertzClientInstrumenter() *instrumenter.PropagatingToDownstreamInstrumenter[*protocol.Request, *protocol.Response] {
	builder := instrumenter.Builder[*protocol.Request, *protocol.Response]{}
	clientGetter := hertzHttpClientAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpClientAttrsGetter, hertzHttpClientAttrsGetter]{HttpGetter: clientGetter, NetGetter: clientGetter}
	networkExtractor := net.NetworkAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpClientAttrsGetter]{Getter: clientGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpClientSpanStatusExtractor[*protocol.Request, *protocol.Response]{Getter: clientGetter}).SetSpanNameExtractor(&http.HttpClientSpanNameExtractor[*protocol.Request, *protocol.Response]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[*protocol.Request]{}).
		AddOperationListeners(http.HttpClientMetrics("hertz.client")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.HERTZ_HTTP_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(&http.HttpClientAttrsExtractor[*protocol.Request, *protocol.Response, hertzHttpClientAttrsGetter, hertzHttpClientAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor}).BuildPropagatingToDownstreamInstrumenter(func(n *protocol.Request) propagation.TextMapCarrier {
		return hertzTextMapCarrier{n}
	}, otel.GetTextMapPropagator())
}
