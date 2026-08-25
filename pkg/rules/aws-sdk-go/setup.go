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
	"errors"
	"net"
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/aws/aws-sdk-go"

// Span name for a request that carries no Operation to name it after.
const defaultSpanName = "aws-sdk-go"

// Handler names, so the handlers are installed by name (replace-if-present)
// rather than appended. NewSession delegates to NewSessionWithOptions, so both
// hooks fire for a single NewSession call and every session is installed onto
// twice; naming the handlers keeps that idempotent instead of stacking
// duplicates onto the same list.
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
	// What makes a header dangerous here is that its value changes between
	// being signed and being sent. traceparent does: the net/http rule starts
	// a fresh client span per attempt, so the span id differs every time.
	// `baggage`, also injected by the default propagator, does not: it is
	// derived from the context rather than from the per-attempt span, so the
	// signed and sent values still match and it is left alone. Any future
	// header that varies per attempt has to be added here.
	//
	// Two residual windows, both narrower than the one this closes:
	//   - A caller that sets `traceparent` on the request itself while the
	//     net/http instrumentation is enabled still gets the signed value
	//     overwritten on the very first attempt.
	//   - Conversely, with the net/http instrumentation disabled, a caller
	//     that sets `traceparent` itself loses it from the retry onwards,
	//     because nothing re-injects after the strip.
	// Stripping unconditionally would close the first at the cost of widening
	// the second for everyone; the retry path is the one observed in practice.
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

			service := serviceName(r)

			spanName := operation
			if spanName == "" {
				spanName = defaultSpanName
			} else if service != "" {
				spanName = service + "." + operation
			}

			ctx, span := otel.Tracer(tracerName).Start(
				r.Context(), spanName, oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
			)
			span.SetAttributes(
				attribute.String("rpc.system", "aws-api"),
				attribute.String("rpc.service", service),
				attribute.String("rpc.method", operation),
			)

			if r.HTTPRequest != nil && r.HTTPRequest.URL != nil {
				host, port := splitHostPort(r.HTTPRequest.URL.Host)
				if host != "" {
					span.SetAttributes(attribute.String("server.address", host))
				}
				if port != "" {
					span.SetAttributes(attribute.String("server.port", port))
				}
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

			// Send() substitutes an empty *http.Response before running
			// Complete, so a request that never reached a server arrives here
			// with a non-nil response whose StatusCode is 0. Reporting that as
			// a status code is worse than omitting the attribute.
			if r.HTTPResponse != nil && r.HTTPResponse.StatusCode != 0 {
				span.SetAttributes(attribute.Int("http.response.status_code", r.HTTPResponse.StatusCode))
			}

			// Populated by the protocol's UnmarshalMeta handler, which runs
			// before Complete. It is the identifier AWS support asks for, so
			// it is worth carrying even on success.
			if r.RequestID != "" {
				span.SetAttributes(attribute.String("aws.request_id", r.RequestID))
			}

			if r.Error != nil {
				span.SetAttributes(attribute.String("error.type", errorType(r.Error)))
				span.RecordError(r.Error)
				span.SetStatus(codes.Error, r.Error.Error())
			}
		},
	})
}

// serviceName prefers ServiceID ("S3"), the identifier the OTel AWS semantic
// conventions use for rpc.service, and falls back to ServiceName ("s3") for
// clients built without one.
func serviceName(r *request.Request) string {
	if r.ClientInfo.ServiceID != "" {
		return r.ClientInfo.ServiceID
	}

	return r.ClientInfo.ServiceName
}

// splitHostPort separates server.address from server.port, which the semantic
// conventions keep apart. URL.Host carries the port only for non-default
// endpoints (custom endpoints, MinIO, localstack), so the port is usually
// absent and reported as such.
func splitHostPort(hostport string) (string, string) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport, ""
	}

	return host, port
}

// errorType reports the class of a failed request. AWS error codes
// ("SlowDown", "AccessDenied") are the meaningful classification here; other
// failures fall back to the concrete Go type.
func errorType(err error) string {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code()
	}

	return reflect.TypeOf(err).String()
}
