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

package aws_sdk_go

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/aws/aws-sdk-go"

// newSessionOnExit hooks session.NewSession's return value so we can attach
// tracing handlers to every client created from the session.
//
//go:linkname newSessionOnExit github.com/aws/aws-sdk-go/aws/session.newSessionOnExit
func newSessionOnExit(call api.CallContext, sess *session.Session, err error) {
	if err != nil || sess == nil {
		return
	}
	installTraceHandlers(&sess.Handlers)
}

// newSessionWithOptionsOnExit covers session.NewSessionWithOptions.
//
//go:linkname newSessionWithOptionsOnExit github.com/aws/aws-sdk-go/aws/session.newSessionWithOptionsOnExit
func newSessionWithOptionsOnExit(call api.CallContext, sess *session.Session, err error) {
	if err != nil || sess == nil {
		return
	}
	installTraceHandlers(&sess.Handlers)
}

func installTraceHandlers(h *request.Handlers) {
	// (1) Correctness fix: strip W3C trace context BEFORE SigV4 signing.
	//
	// The generic net/http instrumentation injects `traceparent` at
	// Transport.RoundTrip time. On a retry (e.g. server 503 SlowDown),
	// aws-sdk-go reuses the same *http.Request whose header still carries the
	// previously injected traceparent; the re-sign then folds it into
	// SignedHeaders, while RoundTrip overwrites the value again -> the signed
	// value no longer matches the sent value -> 403 SignatureDoesNotMatch /
	// AccessDenied. Removing it before signing keeps it out of SignedHeaders,
	// so it can never corrupt the signature regardless of later injection.
	h.Sign.PushFront(func(r *request.Request) {
		if r.HTTPRequest == nil {
			return
		}
		r.HTTPRequest.Header.Del("traceparent")
		r.HTTPRequest.Header.Del("tracestate")
	})

	// (2) Observability: emit a client span per SDK operation via the SDK's
	// own handler chain, so trace context is carried in-process (not through
	// signed HTTP headers) and never affects signing.
	h.Send.PushFront(func(r *request.Request) {
		spanName := r.Operation.Name
		if r.ClientInfo.ServiceName != "" {
			spanName = r.ClientInfo.ServiceName + "." + r.Operation.Name
		}
		ctx, span := otel.Tracer(tracerName).Start(
			r.Context(), spanName, oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		)
		span.SetAttributes(
			attribute.String("rpc.system", "aws-api"),
			attribute.String("rpc.service", r.ClientInfo.ServiceName),
			attribute.String("rpc.method", r.Operation.Name),
		)
		r.SetContext(ctx)
	})
	h.Complete.PushBack(func(r *request.Request) {
		span := oteltrace.SpanFromContext(r.Context())
		if !span.IsRecording() {
			return
		}
		defer span.End()
		if r.HTTPResponse != nil {
			span.SetAttributes(attribute.Int("http.response.status_code", r.HTTPResponse.StatusCode))
		}
		if r.Error != nil {
			span.SetStatus(codes.Error, r.Error.Error())
		}
	})
}
