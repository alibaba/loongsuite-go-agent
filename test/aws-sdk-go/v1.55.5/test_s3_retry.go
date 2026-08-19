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
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

const awsScopeName = "github.com/aws/aws-sdk-go"

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

func main() {
	var attempts int32

	// Reproduces the production failure: the first attempt is throttled, and
	// on the retry the server rejects the request if traceparent made it into
	// SignedHeaders — the signed value is the one injected during the previous
	// attempt, while RoundTrip overwrites it with a fresh span before sending.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer mockServer.Close()

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(mockServer.URL),
		Region:           aws.String("cn-north-1"),
		Credentials:      credentials.NewStaticCredentials("test-ak", "test-sk", ""),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
		MaxRetries:       aws.Int(3),
	})
	if err != nil {
		panic(err)
	}

	_, err = s3.New(sess).PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("request_log/test-object"),
		Body:   bytes.NewReader([]byte("hello")),
	})
	if err != nil {
		panic(err)
	}

	verifier.Assert(atomic.LoadInt32(&attempts) == 2, "Expected the request to be retried once, got %d attempts", atomic.LoadInt32(&attempts))

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		var awsSpans []tracetest.SpanStub

		for _, t := range stubs {
			for _, span := range t {
				if span.InstrumentationScope.Name == awsScopeName {
					awsSpans = append(awsSpans, span)
				}
			}
		}

		// One span for the operation, reused across attempts: starting a span
		// per attempt would leak every span but the last.
		verifier.Assert(len(awsSpans) == 1, "Expected exactly 1 aws-sdk-go span, got %d", len(awsSpans))

		span := awsSpans[0]
		verifier.Assert(span.Name == "s3.PutObject", "Expected span name to be s3.PutObject, got %s", span.Name)

		// Internal, so it does not double-count the net/http client span.
		verifier.Assert(span.SpanKind == trace.SpanKindInternal, "Expected span kind to be internal, got %v", span.SpanKind)

		system := verifier.GetAttribute(span.Attributes, "rpc.system").AsString()
		verifier.Assert(system == "aws-api", "Expected rpc.system to be aws-api, got %s", system)

		service := verifier.GetAttribute(span.Attributes, "rpc.service").AsString()
		verifier.Assert(service == "s3", "Expected rpc.service to be s3, got %s", service)

		method := verifier.GetAttribute(span.Attributes, "rpc.method").AsString()
		verifier.Assert(method == "PutObject", "Expected rpc.method to be PutObject, got %s", method)

		verifier.Assert(verifier.GetAttribute(span.Attributes, "server.address").AsString() != "", "Expected server.address to be set")
	}, 1)
}
