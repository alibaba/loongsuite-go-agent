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
	"context"
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

// Handler names, so the handlers are installed by name (replace-if-present)
// and re-running the rule cannot stack duplicates onto the same list.
const (
	stripTraceHeadersHandlerName = "loongsuite-go/aws-sdk-go.StripTraceHeaders"
	startSpanHandlerName         = "loongsuite-go/aws-sdk-go.StartSpan"
	endSpanHandlerName           = "loongsuite-go/aws-sdk-go.EndSpan"
)

// spanContextKey carries the span started by this rule on the request
// context. Looking it up with trace.SpanFromContext instead would return the
// caller's parent span whenever this rule never started one (e.g. the request
// failed during Validate/Build/Sign, so Send never ran) and end it early.
type spanContextKey struct{}

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
	// (1) Correctness fix: strip W3C trace context before re-signing a retry.
	//
	// The generic net/http instrumentation injects `traceparent` at
	// Transport.RoundTrip time, i.e. after signing. The first attempt is
	// therefore always safe: the header does not exist yet when SigV4 runs,
	// so it never enters SignedHeaders. On a retry (e.g. server 503
	// SlowDown), aws-sdk-go reuses the same *http.Request, whose header still
	// carries the traceparent injected during the previous attempt; the
	// re-sign folds it into SignedHeaders, while RoundTrip then overwrites
	// the value with a new span -> the signed value no longer matches the
	// sent value -> 403 SignatureDoesNotMatch / AccessDenied.
	//
	// Only retries are stripped, so trace context injected or set by the
	// caller still propagates on the first attempt, and disabling the
	// net/http instrumentation does not silently lose it.
	//
	// Residual window: a caller that sets `traceparent` on the request itself
	// while the net/http instrumentation is enabled still gets the signed
	// value overwritten on the very first attempt. Stripping unconditionally
	// would close that too, at the cost of severing propagation for everyone
	// else; the retry path is the one observed in practice.
	h.Sign.SetFrontNamed(request.NamedHandler{
		Name: stripTraceHeadersHandlerName,
		Fn: func(r *request.Request) {
			if r.HTTPRequest == nil || r.RetryCount == 0 {
				return
			}

			r.HTTPRequest.Header.Del("traceparent")
			r.HTTPRequest.Header.Del("tracestate")
		},
	})

	// (2) Observability: emit a span per SDK operation via the SDK's own
	// handler chain, so trace context is carried in-process (not through
	// signed HTTP headers) and never affects signing.
	//
	// The span is Internal, not Client: aws-sdk-go issues its requests
	// through net/http, which the generic instrumentation already covers with
	// its own client span. That span is not suppressed here (SpanKeySuppressor
	// matches on scope names registered in InstrumentationRegistry, which this
	// rule is not), so a Client span would double-count every outbound AWS
	// call. Sibling SDK rules that sit on top of net/http (anthropic-sdk-go,
	// deepseek) make the same choice.
	//
	// Send runs once per attempt while Complete runs once per request, so the
	// span is created on the first attempt only and reused across retries;
	// otherwise every retry would start a nested span that is never ended.
	h.Send.SetFrontNamed(request.NamedHandler{
		Name: startSpanHandlerName,
		Fn: func(r *request.Request) {
			if _, ok := r.Context().Value(spanContextKey{}).(oteltrace.Span); ok {
				return
			}

			// Standard SDK clients always set Operation, but these handlers
			// are attached to every request from the session, including
			// hand-built ones.
			operation := ""
			if r.Operation != nil {
				operation = r.Operation.Name
			}

			spanName := operation
			if spanName == "" {
				spanName = tracerName
			} else if r.ClientInfo.ServiceName != "" {
				spanName = r.ClientInfo.ServiceName + "." + operation
			}

			ctx, span := otel.Tracer(tracerName).Start(
				r.Context(), spanName, oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
			)
			span.SetAttributes(
				attribute.String("rpc.system", "aws-api"),
				attribute.String("rpc.service", r.ClientInfo.ServiceName),
				attribute.String("rpc.method", operation),
			)

			if r.HTTPRequest != nil && r.HTTPRequest.URL != nil {
				span.SetAttributes(attribute.String("server.address", r.HTTPRequest.URL.Host))
			}

			r.SetContext(context.WithValue(ctx, spanContextKey{}, span))
		},
	})

	h.Complete.SetBackNamed(request.NamedHandler{
		Name: endSpanHandlerName,
		Fn: func(r *request.Request) {
			// Only end the span this rule started; a request that failed
			// before Send never stored one.
			span, ok := r.Context().Value(spanContextKey{}).(oteltrace.Span)
			if !ok || span == nil {
				return
			}

			defer span.End()

			if r.HTTPResponse != nil {
				span.SetAttributes(attribute.Int("http.response.status_code", r.HTTPResponse.StatusCode))
			}

			if r.Error != nil {
				span.SetStatus(codes.Error, r.Error.Error())
			}
		},
	})
}
