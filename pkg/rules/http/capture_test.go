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
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestHTTPCaptureConfigCaptureHeaders(t *testing.T) {
	config := newHTTPCaptureConfig("content-type,x-request-id,authorization", "", "true")
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
	want := `{"authorization":["secret"],"content-type":["application/json"],"x-request-id":["request-id"]}`
	if captured != want {
		t.Fatalf("captured headers = %q, want %q", captured, want)
	}
}

func TestHTTPCaptureConfigCaptureAllHeadersSkipsSensitiveByDefault(t *testing.T) {
	config := newHTTPCaptureConfig("", "true", "false")
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	headers.Add("X-Request-Id", "request-id")
	headers.Add("Authorization", "secret")
	headers.Add("Cookie", "session")
	headers.Add("X-Api-Key", "secret")

	captured := config.captureHeaders(headers)

	want := `{"content-type":["application/json"],"x-request-id":["request-id"]}`
	if captured != want {
		t.Fatalf("captured headers = %q, want %q", captured, want)
	}
}

func TestHTTPCaptureConfigSkipsHeadersByDefault(t *testing.T) {
	config := newHTTPCaptureConfig("", "", "false")
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	if config.captureRequestHeaders {
		t.Fatal("expected request header capture to be disabled")
	}
	if captured := config.captureHeaders(headers); captured != "" {
		t.Fatalf("captured headers = %q, want empty", captured)
	}
}

func TestHTTPCaptureConfigSkipsOversizedHeaders(t *testing.T) {
	config := newHTTPCaptureConfig("content-type,x-request-id", "", "false")
	config.maxHeadersBytes = 10
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	headers.Add("X-Request-Id", "request-id")

	if captured := config.captureHeaders(headers); captured != "" {
		t.Fatalf("captured headers = %q, want empty", captured)
	}
}

func TestNewHTTPCaptureConfigFromEnv(t *testing.T) {
	t.Setenv(httpCaptureRequestHeadersEnv, " Content-Type, X-Request-Id ")
	t.Setenv(httpCaptureAllHeadersEnv, " TRUE ")
	t.Setenv(httpCaptureBodyEnabledEnv, "yes")

	config := newHTTPCaptureConfigFromEnv()

	if !config.captureRequestHeaders {
		t.Fatal("expected request header capture to be enabled")
	}
	if !config.captureAllHeaders {
		t.Fatal("expected capture-all headers to be enabled")
	}
	if config.captureBody {
		t.Fatal("expected body capture to be disabled")
	}
	if _, ok := config.requestHeaderNames["content-type"]; !ok {
		t.Fatal("expected content-type to be allow-listed")
	}
	if _, ok := config.requestHeaderNames["x-request-id"]; !ok {
		t.Fatal("expected x-request-id to be allow-listed")
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

func TestCaptureHTTPRequestBodyUsesGetBody(t *testing.T) {
	liveBody := `{"source":"live"}`
	getBody := `{"source":"get-body"}`
	getBodyCalls := 0
	req := &http.Request{
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader(liveBody)),
		ContentLength: int64(len(liveBody)),
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		getBodyCalls++
		return io.NopCloser(strings.NewReader(getBody)), nil
	}
	config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

	got := captureHTTPRequestBody(req, config)
	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	if getBodyCalls != 1 {
		t.Fatalf("GetBody calls = %d, want 1", getBodyCalls)
	}
	if got != getBody {
		t.Fatalf("captured body = %q, want %q", got, getBody)
	}
	if string(restored) != liveBody {
		t.Fatalf("live body should remain readable, got %q", restored)
	}
}

func TestCaptureHTTPRequestBodySkipsGetBodyError(t *testing.T) {
	body := `{"key":"value"}`
	getBodyCalls := 0
	req := &http.Request{
		Header:        http.Header{},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
		getBodyCalls++
		return nil, errors.New("get body failed")
	}
	config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

	got := captureHTTPRequestBody(req, config)
	restored, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	if got != "" {
		t.Fatalf("captured body = %q, want empty", got)
	}
	if getBodyCalls != 1 {
		t.Fatalf("GetBody calls = %d, want 1", getBodyCalls)
	}
	if string(restored) != body {
		t.Fatalf("live body should remain readable, got %q", restored)
	}
}

func TestReadAndRestoreBodyPreservesCloser(t *testing.T) {
	body := `{"key":"value"}`
	original := newTrackingReadCloser(body)
	readCloser := io.ReadCloser(original)

	got, ok := readAndRestoreBody(&readCloser, defaultMaxHTTPBodyBytes)
	if !ok {
		t.Fatal("expected body to be captured")
	}
	if got != body {
		t.Fatalf("captured body = %q, want %q", got, body)
	}
	if original.closed {
		t.Fatal("original body should not be closed during capture")
	}

	restored, err := io.ReadAll(readCloser)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != body {
		t.Fatalf("restored body = %q, want %q", restored, body)
	}
	if err := readCloser.Close(); err != nil {
		t.Fatal(err)
	}
	if !original.closed {
		t.Fatal("restored body should close the original body")
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

func TestCaptureHTTPResponseBodySkipsIneligibleResponses(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		contentType string
		encoding    string
		length      int64
	}{
		{name: "gzip", body: []byte("hello"), contentType: "text/plain", encoding: "gzip", length: 5},
		{name: "invalid utf8", body: []byte{0xff}, contentType: "text/plain", length: 1},
		{name: "unknown length", body: []byte(`{"ok":true}`), contentType: "application/json", length: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &http.Response{
				Header:        http.Header{},
				Body:          io.NopCloser(bytes.NewReader(tt.body)),
				ContentLength: tt.length,
			}
			res.Header.Set("Content-Type", tt.contentType)
			res.Header.Set("Content-Encoding", tt.encoding)
			config := httpCaptureConfig{captureBody: true, maxBodyBytes: defaultMaxHTTPBodyBytes}

			if got := captureHTTPResponseBody(res, config); got != "" {
				t.Fatalf("captured response body = %q, want empty", got)
			}
		})
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

func TestWriterWrapperCapturesDetectedTextResponseBody(t *testing.T) {
	body := "hello"
	recorder := httptest.NewRecorder()
	w := &writerWrapper{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		captureBody:    true,
		maxBodyBytes:   defaultMaxHTTPBodyBytes,
	}

	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}

	if got := w.capturedResponseBody(); got != body {
		t.Fatalf("captured response body = %q, want %q", got, body)
	}
}

func TestWriterWrapperSkipsEncodedAndInvalidUTF8ResponseBody(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		encoding string
	}{
		{name: "gzip", body: []byte("hello"), encoding: "gzip"},
		{name: "invalid utf8", body: []byte{0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			w := &writerWrapper{
				ResponseWriter: recorder,
				statusCode:     http.StatusOK,
				captureBody:    true,
				maxBodyBytes:   defaultMaxHTTPBodyBytes,
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Encoding", tt.encoding)

			if _, err := w.Write(tt.body); err != nil {
				t.Fatal(err)
			}

			if got := w.capturedResponseBody(); got != "" {
				t.Fatalf("captured response body = %q, want empty", got)
			}
			if got := recorder.Body.Bytes(); string(got) != string(tt.body) {
				t.Fatalf("underlying response body = %q, want %q", got, tt.body)
			}
		})
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

type trackingReadCloser struct {
	*strings.Reader
	closed bool
}

func newTrackingReadCloser(body string) *trackingReadCloser {
	return &trackingReadCloser{Reader: strings.NewReader(body)}
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}
