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

package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	captureRequestBody  = `{"request":"ok"}`
	captureResponseBody = `{"response":"ok"}`
	captureRequestID    = "capture-request-id"
)

func setupCaptureHTTP() {
	mux := http.NewServeMux()
	mux.HandleFunc("/capture", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("failed to read server request body: %v", err)
		}
		if string(body) != captureRequestBody {
			log.Fatalf("server request body = %q, want %q", body, captureRequestBody)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(captureResponseBody)))
		_, _ = w.Write([]byte(captureResponseBody))
	})
	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		panic(err)
	}
}

func main() {
	go setupCaptureHTTP()
	time.Sleep(1 * time.Second)

	req, err := http.NewRequest("POST", "http://127.0.0.1:"+strconv.Itoa(port)+"/capture", strings.NewReader(captureRequestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", captureRequestID)
	req.Header.Set("Authorization", "secret")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(body) != captureResponseBody {
		log.Fatalf("client response body = %q, want %q", body, captureResponseBody)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		clientSpan := stubs[0][0]
		serverSpan := stubs[0][1]

		assertHeadersAttr(clientSpan.Attributes)
		assertMissingAttr(clientSpan.Attributes, "http.request.header.authorization")
		assertStringAttr(clientSpan.Attributes, "http.request.body.content", captureRequestBody)
		assertStringAttr(clientSpan.Attributes, "http.response.body.content", captureResponseBody)

		assertHeadersAttr(serverSpan.Attributes)
		assertMissingAttr(serverSpan.Attributes, "http.request.header.authorization")
		assertStringAttr(serverSpan.Attributes, "http.request.body.content", captureRequestBody)
		assertStringAttr(serverSpan.Attributes, "http.response.body.content", captureResponseBody)
	}, 1)
}

func assertStringAttr(attrs []attribute.KeyValue, name string, want string) {
	got, ok := findAttr(attrs, name)
	if !ok {
		log.Fatalf("attribute %q not found", name)
	}
	if got.Value.AsString() != want {
		log.Fatalf("attribute %q = %q, want %q", name, got.Value.AsString(), want)
	}
}

func assertHeadersAttr(attrs []attribute.KeyValue) {
	headers, ok := findAttr(attrs, "http.request.headers")
	if !ok {
		log.Fatalf("attribute %q not found", "http.request.headers")
	}
	value := headers.Value.AsString()
	assertContains(value, `"content-type":["application/json"]`)
	assertContains(value, `"x-request-id":["`+captureRequestID+`"]`)
	assertContains(value, `"authorization":["secret"]`)
}

func assertContains(value string, want string) {
	if !strings.Contains(value, want) {
		log.Fatalf("%q should contain %q", value, want)
	}
}

func assertMissingAttr(attrs []attribute.KeyValue, name string) {
	if _, ok := findAttr(attrs, name); ok {
		log.Fatalf("attribute %q should not be present", name)
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
