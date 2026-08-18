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

package trpc

import (
	"fmt"
	"os"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/rpc"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/trace"
	"trpc.group/trpc-go/trpc-go/codec"
)

type trpcInnerEnabler struct {
	enabled bool
}

func (t trpcInnerEnabler) Enable() bool {
	return t.enabled
}

var trpcEnabler = trpcInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_TRPC_ENABLED") != "false"}

type trpcClientAttrsGetter struct {
}

func (t trpcClientAttrsGetter) GetSystem(request trpcReq) string {
	return "trpc"
}

func (t trpcClientAttrsGetter) GetService(request trpcReq) string {
	return request.msg.CallerService()
}

func (t trpcClientAttrsGetter) GetMethod(request trpcReq) string {
	return request.msg.CallerMethod()
}

func (t trpcClientAttrsGetter) GetServerAddress(request trpcReq) string {
	return request.addr
}

type trpcServerAttrsGetter struct {
}

func (t trpcServerAttrsGetter) GetSystem(request trpcReq) string {
	return "trpc"
}

func (t trpcServerAttrsGetter) GetService(request trpcReq) string {
	return request.msg.CalleeService()
}

func (t trpcServerAttrsGetter) GetMethod(request trpcReq) string {
	return request.msg.CalleeMethod()
}

func (t trpcServerAttrsGetter) GetServerAddress(request trpcReq) string {
	if request.msg.LocalAddr() != nil {
		return request.msg.LocalAddr().String()
	}
	return ""
}

type trpcStatusCodeExtractor[REQUEST trpcReq, RESPONSE trpcRes] struct {
}

func (t trpcStatusCodeExtractor[REQUEST, RESPONSE]) Extract(span trace.Span, request trpcReq, response trpcRes, err error) {
	statusCode := response.statusCode
	if statusCode != 0 {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("trpc error status code %d", statusCode))
		}
	}
}

type trpcRequestCarrier struct {
	reqHeader codec.Msg
}

func (t trpcRequestCarrier) Get(key string) string {
	return string(t.reqHeader.ServerMetaData()[key])
}

func (t trpcRequestCarrier) Set(key string, value string) {
	md := t.reqHeader.ClientMetaData()
	if md == nil {
		md = codec.MetaData{}
	}
	if _, ok := md[key]; ok {
		return
	}
	md[key] = []byte(value)
	t.reqHeader.WithClientMetaData(md)
}

func (t trpcRequestCarrier) Keys() []string {
	vals := []string{}
	for _, byteV := range t.reqHeader.ClientMetaData() {
		vals = append(vals, string(byteV))
	}
	return vals
}

func BuildTrpcClientInstrumenter() instrumenter.Instrumenter[trpcReq, trpcRes] {
	builder := instrumenter.Builder[trpcReq, trpcRes]{}
	clientGetter := trpcClientAttrsGetter{}
	return builder.Init().SetSpanStatusExtractor(&trpcStatusCodeExtractor[trpcReq, trpcRes]{}).SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[trpcReq]{Getter: clientGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[trpcReq]{}).
		AddAttributesExtractor(&rpc.ClientRpcAttrsExtractor[trpcReq, trpcRes, trpcClientAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.TRPCGO_CLIENT_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(rpc.RpcClientMetrics("trpc.client")).
		BuildPropagatingToDownstreamInstrumenter(
			func(n trpcReq) propagation.TextMapCarrier {
				return trpcRequestCarrier{reqHeader: n.msg}
			},
			otel.GetTextMapPropagator(),
		)
}

func BuildTrpcServerInstrumenter() instrumenter.Instrumenter[trpcReq, trpcRes] {
	builder := instrumenter.Builder[trpcReq, trpcRes]{}
	serverGetter := trpcServerAttrsGetter{}
	return builder.Init().SetSpanStatusExtractor(&trpcStatusCodeExtractor[trpcReq, trpcRes]{}).SetSpanNameExtractor(&rpc.RpcSpanNameExtractor[trpcReq]{Getter: serverGetter}).
		SetSpanKindExtractor(&instrumenter.AlwaysServerExtractor[trpcReq]{}).
		AddAttributesExtractor(&rpc.ServerRpcAttrsExtractor[trpcReq, trpcRes, trpcServerAttrsGetter]{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.TRPCGO_SERVER_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddOperationListeners(rpc.RpcServerMetrics("trpc.server")).
		BuildPropagatingFromUpstreamInstrumenter(
			func(n trpcReq) propagation.TextMapCarrier {
				return trpcRequestCarrier{reqHeader: n.msg}
			},
			otel.GetTextMapPropagator(),
		)
}
