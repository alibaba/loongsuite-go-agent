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
	"context"
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	kt "github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
)

const OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE = "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE"

var kratosInternalInstrument = BuildKratosInternalInstrumenter()

// clientInstrumentedKey guards against the client tracing middleware being
// installed more than once in the same middleware chain (NewClient injects it,
// and WithMiddleware appends it), which would otherwise emit duplicate spans.
type clientInstrumentedKey struct{}

//go:linkname kratosNewHTTPServiceOnEnterV3 github.com/go-kratos/kratos/v3/transport/http.kratosNewHTTPServiceOnEnterV3
func kratosNewHTTPServiceOnEnterV3(call api.CallContext, opts ...http.ServerOption) {
	if os.Getenv(OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE) != "true" {
		return
	}
	opts = append(opts, AddHTTPMiddleware(ServerTracingMiddleWare()))
	call.SetParam(0, opts)
}

// kratosNewHTTPClientOnEnterV3 injects the client tracing middleware for the case
// where the user does not pass any WithMiddleware option. When the user does pass
// one, http.WithMiddleware overwrites o.middleware, so kratosHTTPWithMiddlewareOnEnterV3
// additionally appends the tracing middleware to preserve it.
//
//go:linkname kratosNewHTTPClientOnEnterV3 github.com/go-kratos/kratos/v3/transport/http.kratosNewHTTPClientOnEnterV3
func kratosNewHTTPClientOnEnterV3(call api.CallContext, ctx context.Context, opts ...http.ClientOption) {
	if os.Getenv(OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE) != "true" {
		return
	}
	nopts := []http.ClientOption{
		http.WithMiddleware(ClientTracingMiddleWare()),
	}
	nopts = append(nopts, opts...)
	call.SetParam(1, nopts)
}

//go:linkname kratosHTTPWithMiddlewareOnEnterV3 github.com/go-kratos/kratos/v3/transport/http.kratosHTTPWithMiddlewareOnEnterV3
func kratosHTTPWithMiddlewareOnEnterV3(call api.CallContext, m ...middleware.Middleware) {
	if os.Getenv(OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE) != "true" {
		return
	}
	m = append(m, ClientTracingMiddleWare())
	call.SetParam(0, m)
}

func AddHTTPMiddleware(m middleware.Middleware) http.ServerOption {
	return func(o *http.Server) {
		o.Use("*", m)
	}
}

func AddGRPCMiddleware(m middleware.Middleware) grpc.ServerOption {
	return func(o *grpc.Server) {
		o.Use("*", m)
	}
}

func buildKratosRequest(ctx context.Context, tr transport.Transporter, spanKind string) kratosRequest {
	serviceName, serviceId, serviceVersion := "", "", ""
	serviceEndpoint := make([]string, 0, 0)
	serviceMeta := make(map[string]string)
	app, hasApp := kt.FromContext(ctx)
	if hasApp {
		serviceName, serviceId, serviceVersion, serviceEndpoint = app.Name(), app.ID(), app.Version(), app.Endpoint()
		serviceMeta = app.Metadata()
	}
	request := kratosRequest{
		serviceId:       serviceId,
		serviceName:     serviceName,
		serviceVersion:  serviceVersion,
		serviceEndpoint: serviceEndpoint,
		serviceMeta:     serviceMeta,
		spanKind:        spanKind,
		operation:       tr.Operation(),
	}
	switch tr.Kind() {
	case transport.KindGRPC:
		request.protocolType = "grpc"
	case transport.KindHTTP:
		request.protocolType = "http"
	}
	return request
}

func ServerTracingMiddleWare() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				request := buildKratosRequest(ctx, tr, spanKindServer)
				sCtx := kratosInternalInstrument.Start(ctx, request)
				defer func() {
					kratosInternalInstrument.End(sCtx, request, nil, err)
				}()
			}
			return handler(ctx, req)
		}
	}
}

func ClientTracingMiddleWare() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				if ctx.Value(clientInstrumentedKey{}) != nil {
					return handler(ctx, req)
				}
				ctx = context.WithValue(ctx, clientInstrumentedKey{}, struct{}{})
				request := buildKratosRequest(ctx, tr, spanKindClient)
				sCtx := kratosInternalInstrument.Start(ctx, request)
				defer func() {
					kratosInternalInstrument.End(sCtx, request, nil, err)
				}()
				return handler(sCtx, req)
			}
			return handler(ctx, req)
		}
	}
}
