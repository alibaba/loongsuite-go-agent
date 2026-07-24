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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestHTTPCaptureConfigCaptureHeaders(t *testing.T) {
	config := newHTTPCaptureConfig("true", "true")
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	headers.Add("X-Request-Id", "request-id")
	headers.Add("Authorization", "secret")
	headers.Add("X-Api-Key", "secret")
	headers.Add("X-Access-Token", "secret")
	headers.Add("Other", "skip")

	captured := config.captureHeaders(headers)

	if !config.captureBody {
		t.Fatal("expected body capture to be enabled")
	}
	if !config.captureRequestHeaders {
		t.Fatal("expected request header capture to be enabled")
	}
	want := `{"authorization":["secret"],"content-type":["application/json"],"other":["skip"],"x-access-token":["secret"],"x-api-key":["secret"],"x-request-id":["request-id"]}`
	if captured != want {
		t.Fatalf("captured headers = %q, want %q", captured, want)
	}
}

func TestHTTPCaptureConfigSkipsHeadersByDefault(t *testing.T) {
	config := newHTTPCaptureConfig("", "false")
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	if config.captureRequestHeaders {
		t.Fatal("expected request header capture to be disabled")
	}
	if captured := config.captureHeaders(headers); captured != "" {
		t.Fatalf("captured headers = %q, want empty", captured)
	}
}

func TestHTTPCaptureBodyContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{name: "text", contentType: "text/plain; charset=utf-8", want: true},
		{name: "json", contentType: "application/json", want: true},
		{name: "json suffix", contentType: "application/problem+json", want: true},
		{name: "binary", contentType: "application/octet-stream", want: false},
		{name: "empty", contentType: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTextOrJSONContentType(tt.contentType); got != tt.want {
				t.Fatalf("isTextOrJSONContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestCaptureHTTPRequestBodyRestoresBody(t *testing.T) {
	body := `{"key":"value"}`
	req := &http.Request{
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	req.Header.Set("Content-Type", "application/json")
	config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

	got := captureHTTPRequestBody(req, config)
	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	if got != body {
		t.Fatalf("captured body = %q, want %q", got, body)
	}
	if string(restored) != body {
		t.Fatalf("restored body = %q, want %q", restored, body)
	}
}

func TestCaptureHTTPRequestBodySkipsUnknownLength(t *testing.T) {
	body := `{"key":"value"}`
	req := &http.Request{
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: -1,
	}
	req.Header.Set("Content-Type", "application/json")
	config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

	got := captureHTTPRequestBody(req, config)
	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	if got != "" {
		t.Fatalf("captured body = %q, want empty", got)
	}
	if string(restored) != body {
		t.Fatalf("body should not be consumed, got %q", restored)
	}
}

func TestCaptureHTTPResponseBodyRestoresBody(t *testing.T) {
	body := "hello"
	res := &http.Response{
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	res.Header.Set("Content-Type", "text/plain")
	config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

	got := captureHTTPResponseBody(res, config)
	restored, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if got != body {
		t.Fatalf("captured response body = %q, want %q", got, body)
	}
	if string(restored) != body {
		t.Fatalf("restored response body = %q, want %q", restored, body)
	}
}

func TestCaptureBodySkipsCompressedAndInvalidUTF8(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		encoding string
	}{
		{name: "compressed", body: "hello", encoding: "gzip"},
		{name: "invalid utf8", body: string([]byte{0xff}), encoding: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Header:        http.Header{},
				Body:          io.NopCloser(strings.NewReader(tt.body)),
				ContentLength: int64(len(tt.body)),
			}
			req.Header.Set("Content-Type", "text/plain")
			req.Header.Set("Content-Encoding", tt.encoding)
			config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

			if got := captureHTTPRequestBody(req, config); got != "" {
				t.Fatalf("captured body = %q, want empty", got)
			}
		})
	}
}

func TestWriterWrapperCapturesSmallResponseBody(t *testing.T) {
	body := `{"ok":true}`
	recorder := httptest.NewRecorder()
	w := &writerWrapper{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		captureBody:    true,
		maxBodyBytes:   defaultMaxHTTPBodyBytes,
	}
	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}

	if got := w.capturedResponseBody(); got != body {
		t.Fatalf("captured response body = %q, want %q", got, body)
	}
	if got := recorder.Body.String(); got != body {
		t.Fatalf("underlying response body = %q, want %q", got, body)
	}
}

func TestWriterWrapperSkipsLargeResponseBody(t *testing.T) {
	body := strings.Repeat("a", int(defaultMaxHTTPBodyBytes)+1)
	recorder := httptest.NewRecorder()
	w := &writerWrapper{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		captureBody:    true,
		maxBodyBytes:   defaultMaxHTTPBodyBytes,
	}
	w.Header().Set("Content-Type", "text/plain")

	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}

	if got := w.capturedResponseBody(); got != "" {
		t.Fatalf("captured response body = %q, want empty", got)
	}
	if got := recorder.Body.String(); got != body {
		t.Fatalf("underlying response body = %q, want %q", got, body)
	}
}

func TestHTTPCaptureAttrsExtractor(t *testing.T) {
	extractor := &httpCaptureAttrsExtractor{}
	request := &netHttpRequest{
		requestHeaders: `{"content-type":["application/json"],"x-request-id":["request-id"]}`,
		requestBody:    `{"key":"value"}`,
	}
	response := &netHttpResponse{responseBody: `{"ok":true}`}

	attrs, _ := extractor.OnStart(nil, context.Background(), request)
	attrs, _ = extractor.OnEnd(attrs, context.Background(), request, response, nil)

	assertAttrString(t, attrs, "http.request.headers", `{"content-type":["application/json"],"x-request-id":["request-id"]}`)
	assertAttrString(t, attrs, "http.request.body.content", `{"key":"value"}`)
	assertAttrString(t, attrs, "http.response.body.content", `{"ok":true}`)
}

func assertAttrString(t *testing.T, attrs []attribute.KeyValue, name string, want string) {
	t.Helper()
	attr, ok := findAttr(attrs, name)
	if !ok {
		t.Fatalf("attribute %q not found", name)
	}
	if got := attr.Value.AsString(); got != want {
		t.Fatalf("attribute %q = %q, want %q", name, got, want)
	}
}

func findAttr(attrs []attribute.KeyValue, name string) (attribute.KeyValue, bool) {
	for _, attr := range attrs {
		if string(attr.Key) == name {
			return attr, true
		}
	}
	return attribute.KeyValue{}, false
}
