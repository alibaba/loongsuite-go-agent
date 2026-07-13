// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package gomicro

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/net"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go-micro.dev/v5/metadata"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type goMicroHttpClientAttrsGetter struct {
}

func (n goMicroHttpClientAttrsGetter) GetRequestMethod(request goMicroRequest) string {
	switch request.reqType {
	case CallRequest:
		return request.request.Method()
	case StreamRequest:
		return request.request.Method()
	case MessageRequest:
		return "pub"
	}
	return ""
}
func (n goMicroHttpClientAttrsGetter) GetHttpRequestHeader(request goMicroRequest, name string) []string {
	all := make([]string, 0)
	md, ok := metadata.FromContext(request.ctx)
	if ok {
		value, ok := md.Get(name)
		if ok {
			all = append(all, string(value))
		}
	}
	return all
}
func (n goMicroHttpClientAttrsGetter) GetHttpResponseStatusCode(request goMicroRequest, response goMicroResponse, err error) int {
	// Minor #7: 恢复原有的 GetHttpResponseStatusCode 逻辑以满足接口契约
	if response.err != nil {
		return 500
	}
	return 200
}

func (n goMicroHttpClientAttrsGetter) HasHttpResponse(request goMicroRequest, response goMicroResponse, err error) bool {
	return err == nil
}

func (n goMicroHttpClientAttrsGetter) GetHttpResponseHeader(request goMicroRequest, response goMicroResponse, name string) []string {
	all := make([]string, 0)
	md, ok := metadata.FromContext(response.ctx)
	if ok {
		value, ok := md.Get(name)
		if ok {
			all = append(all, string(value))
		}
	}
	return all
}
func (n goMicroHttpClientAttrsGetter) GetErrorType(request goMicroRequest, response goMicroResponse, err error) string {
	return ""
}
func (n goMicroHttpClientAttrsGetter) GetNetworkType(request goMicroRequest, response goMicroResponse) string {
	return "ipv4"
}
func (n goMicroHttpClientAttrsGetter) GetNetworkTransport(request goMicroRequest, response goMicroResponse) string {
	return "tcp"
}
func (n goMicroHttpClientAttrsGetter) GetNetworkProtocolName(request goMicroRequest, response goMicroResponse) string {
	return "http"
}
func (n goMicroHttpClientAttrsGetter) GetNetworkProtocolVersion(request goMicroRequest, response goMicroResponse) string {
	return ""
}
func (n goMicroHttpClientAttrsGetter) GetNetworkLocalInetAddress(request goMicroRequest, response goMicroResponse) string {
	return ""
}
func (n goMicroHttpClientAttrsGetter) GetNetworkLocalPort(request goMicroRequest, response goMicroResponse) int {
	return 0
}
func (n goMicroHttpClientAttrsGetter) GetNetworkPeerInetAddress(request goMicroRequest, response goMicroResponse) string {
	if request.reqType == MessageRequest {
		return request.msg.Topic()
	}
	return request.request.Service()
}
func (n goMicroHttpClientAttrsGetter) GetNetworkPeerPort(request goMicroRequest, response goMicroResponse) int {
	return 0
}
func (n goMicroHttpClientAttrsGetter) GetUrlFull(request goMicroRequest) string {
	if request.reqType == MessageRequest {
		return request.msg.Topic()
	}
	return request.request.Endpoint()
}
func (n goMicroHttpClientAttrsGetter) GetServerPort(request goMicroRequest) int {
	return 0
}

func (h goMicroHttpClientAttrsGetter) GetServerAddress(request goMicroRequest) string {
	switch request.reqType {
	case CallRequest:
		return request.request.Service()
	case StreamRequest:
		return request.request.Service()
	case MessageRequest:
		return request.msg.Topic()
	}
	return ""
}

func BuildGoMicroClientInstrumenter() *instrumenter.PropagatingToDownstreamInstrumenter[goMicroRequest, goMicroResponse] {
	builder := instrumenter.Builder[goMicroRequest, goMicroResponse]{}
	clientGetter := goMicroHttpClientAttrsGetter{}
	commonExtractor := http.HttpCommonAttrsExtractor[goMicroRequest, goMicroResponse, goMicroHttpClientAttrsGetter, goMicroHttpClientAttrsGetter]{HttpGetter: clientGetter, NetGetter: clientGetter}
	networkExtractor := net.NetworkAttrsExtractor[goMicroRequest, goMicroResponse, goMicroHttpClientAttrsGetter]{Getter: clientGetter}
	return builder.Init().SetSpanStatusExtractor(http.HttpClientSpanStatusExtractor[goMicroRequest, goMicroResponse]{Getter: clientGetter}).SetSpanNameExtractor(&http.HttpClientSpanNameExtractor[goMicroRequest, goMicroResponse]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[goMicroRequest]{}).
		AddOperationListeners(http.HttpClientMetrics("gomicro.client")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GOMICRO_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(&http.HttpClientAttrsExtractor[goMicroRequest, goMicroResponse, goMicroHttpClientAttrsGetter, goMicroHttpClientAttrsGetter]{Base: commonExtractor, NetworkExtractor: networkExtractor}).BuildPropagatingToDownstreamInstrumenter(nil, nil)
}
