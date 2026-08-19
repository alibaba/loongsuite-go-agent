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

package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const awsScopeName = "github.com/aws/aws-sdk-go"

const (
	retryRequestID  = "REQUESTID0000001"
	deniedRequestID = "REQUESTID0000002"
)

// signedHeaders pulls the SignedHeaders list out of a SigV4 Authorization
// header, i.e. the headers whose values the signature commits to.
func signedHeaders(authorization string) string {
	for _, part := range strings.Split(authorization, ",") {
		part = strings.TrimSpace(part)
		if after, ok := strings.CutPrefix(part, "SignedHeaders="); ok {
			return after
		}
	}

	return ""
}

func newSession(endpoint string) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("cn-north-1"),
		Credentials:      credentials.NewStaticCredentials("test-ak", "test-sk", ""),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
		MaxRetries:       aws.Int(3),
	})
	if err != nil {
		panic(err)
	}

	return sess
}

// spansByName groups the spans this rule emitted, across every trace, by span
// name. Each operation is its own root, so they do not share a trace.
func spansByName(stubs []tracetest.SpanStubs) map[string]tracetest.SpanStub {
	spans := make(map[string]tracetest.SpanStub)
	for _, t := range stubs {
		for _, span := range t {
			if span.InstrumentationScope.Name != awsScopeName {
				continue
			}

			_, seen := spans[span.Name]
			verifier.Assert(!seen, "Expected exactly 1 %s span", span.Name)
			spans[span.Name] = span
		}
	}

	return spans
}

func main() {
	var attempts int32

	// Reproduces the production failure: the first attempt is throttled, and
	// on the retry the server rejects the request if traceparent made it into
	// SignedHeaders - the signed value is the one injected during the previous
	// attempt, while RoundTrip overwrites it with a fresh span before sending.
	retryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Amz-Request-Id", retryRequestID)

		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>SlowDown</Code><Message>Please reduce your request rate.</Message></Error>`))

			return
		}

		if strings.Contains(signedHeaders(r.Header.Get("Authorization")), "traceparent") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code><Message>Access Denied</Message></Error>`))

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer retryServer.Close()

	// A second, always-failing endpoint, so the error attributes are covered
	// too. AccessDenied is not retryable, so this is a single attempt.
	deniedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Amz-Request-Id", deniedRequestID)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code><Message>Access Denied</Message></Error>`))
	}))
	defer deniedServer.Close()

	_, err := s3.New(newSession(retryServer.URL)).PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("request_log/test-object"),
		Body:   bytes.NewReader([]byte("hello")),
	})
	if err != nil {
		panic(err)
	}

	verifier.Assert(atomic.LoadInt32(&attempts) == 2, "Expected the request to be retried once, got %d attempts", atomic.LoadInt32(&attempts))

	_, err = s3.New(newSession(deniedServer.URL)).GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("request_log/test-object"),
	})
	verifier.Assert(err != nil, "Expected GetObject to fail with AccessDenied")

	// A third endpoint that is closed before use, so the request never gets
	// an HTTP response at all and the span has no status code to report.
	unreachable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	unreachableURL := unreachable.URL
	unreachable.Close()

	_, err = s3.New(newSession(unreachableURL)).HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String("test-bucket"),
	})
	verifier.Assert(err != nil, "Expected HeadBucket to fail against a closed endpoint")

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		spans := spansByName(stubs)

		// One span per operation, reused across attempts: starting a span per
		// attempt would leak every span but the last.
		verifier.Assert(len(spans) == 3, "Expected exactly 3 aws-sdk-go spans, got %d", len(spans))

		// rpc.service is ServiceID ("S3"), the identifier the OTel AWS
		// semantic conventions use, not ServiceName ("s3").
		put, ok := spans["S3.PutObject"]
		verifier.Assert(ok, "Expected a S3.PutObject span")

		// Internal, so it does not double-count the net/http client span.
		verifier.Assert(put.SpanKind == trace.SpanKindInternal, "Expected span kind to be internal, got %v", put.SpanKind)

		system := verifier.GetAttribute(put.Attributes, "rpc.system").AsString()
		verifier.Assert(system == "aws-api", "Expected rpc.system to be aws-api, got %s", system)

		service := verifier.GetAttribute(put.Attributes, "rpc.service").AsString()
		verifier.Assert(service == "S3", "Expected rpc.service to be S3, got %s", service)

		method := verifier.GetAttribute(put.Attributes, "rpc.method").AsString()
		verifier.Assert(method == "PutObject", "Expected rpc.method to be PutObject, got %s", method)

		// server.address carries the host alone; the port of the test
		// endpoint belongs in server.port.
		address := verifier.GetAttribute(put.Attributes, "server.address").AsString()
		verifier.Assert(address == "127.0.0.1", "Expected server.address to be 127.0.0.1, got %s", address)

		port := verifier.GetAttribute(put.Attributes, "server.port").AsString()
		verifier.Assert(port != "", "Expected server.port to be set")

		status := verifier.GetAttribute(put.Attributes, "http.response.status_code").AsInt64()
		verifier.Assert(status == 200, "Expected http.response.status_code to be 200, got %d", status)

		requestID := verifier.GetAttribute(put.Attributes, "aws.request_id").AsString()
		verifier.Assert(requestID == retryRequestID, "Expected aws.request_id to be %s, got %s", retryRequestID, requestID)

		get, ok := spans["S3.GetObject"]
		verifier.Assert(ok, "Expected a S3.GetObject span")

		errorType := verifier.GetAttribute(get.Attributes, "error.type").AsString()
		verifier.Assert(errorType == "AccessDenied", "Expected error.type to be AccessDenied, got %s", errorType)

		verifier.Assert(get.Status.Code == codes.Error, "Expected the failed span status to be Error, got %v", get.Status.Code)

		deniedID := verifier.GetAttribute(get.Attributes, "aws.request_id").AsString()
		verifier.Assert(deniedID == deniedRequestID, "Expected aws.request_id to be %s, got %s", deniedRequestID, deniedID)

		head, ok := spans["S3.HeadBucket"]
		verifier.Assert(ok, "Expected a S3.HeadBucket span")

		// No response ever arrived, so the attribute must be absent rather
		// than reported as 0.
		verifier.Assert(verifier.GetAttribute(head.Attributes, "http.response.status_code").Type() == attribute.INVALID,
			"Expected http.response.status_code to be absent, got %v", verifier.GetAttribute(head.Attributes, "http.response.status_code"))

		headErrorType := verifier.GetAttribute(head.Attributes, "error.type").AsString()
		verifier.Assert(headErrorType == "RequestError", "Expected error.type to be RequestError, got %s", headErrorType)
	}, 3)
}
