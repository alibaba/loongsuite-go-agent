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
	"log"
	"os"
	"strings"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	fiber "github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	captureRequestBody  = `{"request":"ok"}`
	captureResponseBody = `{"response":"ok"}`
	captureRequestID    = "capture-request-id"
)

func requestServer() {
	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.SetRequestURI("http://localhost:3000/fiber/capture")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("X-Request-Id", captureRequestID)
	req.Header.Set("Authorization", "secret")
	req.SetBodyString(captureRequestBody)

	if err := client.Do(req, resp); err != nil {
		log.Fatal(err)
	}
	if string(resp.Body()) != captureResponseBody {
		log.Fatalf("client response body = %q, want %q", resp.Body(), captureResponseBody)
	}
}

func setupHttp() {
	app := fiber.New()
	app.Post("/fiber/capture", func(c fiber.Ctx) error {
		if string(c.Body()) != captureRequestBody {
			log.Fatalf("server request body = %q, want %q", c.Body(), captureRequestBody)
		}
		c.Set("Content-Type", "application/json")
		return c.SendString(captureResponseBody)
	})
	log.Fatal(app.Listen(":3000"))
}

func main() {
	go setupHttp()
	time.Sleep(3 * time.Second)
	requestServer()

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		assertMissingAttr(stubs[0][0].Attributes, "http.request.header.authorization")
		assertCaptureAttrs(stubs[0][0].Attributes)

		assertMissingAttr(stubs[0][1].Attributes, "http.request.header.authorization")
		assertCaptureAttrs(stubs[0][1].Attributes)
	}, 1)
}

func assertCaptureAttrs(attrs []attribute.KeyValue) {
	if captureRequestHeadersEnabled() {
		assertHeadersAttr(attrs)
	} else {
		assertMissingAttr(attrs, "http.request.headers")
	}

	if captureBodyEnabled() {
		assertStringAttr(attrs, "http.request.body.content", captureRequestBody)
		assertStringAttr(attrs, "http.response.body.content", captureResponseBody)
	} else {
		assertMissingAttr(attrs, "http.request.body.content")
		assertMissingAttr(attrs, "http.response.body.content")
	}
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
	assertNotContains(value, "authorization")
}

func assertContains(value string, want string) {
	if !strings.Contains(value, want) {
		log.Fatalf("%q should contain %q", value, want)
	}
}

func assertNotContains(value string, want string) {
	if strings.Contains(value, want) {
		log.Fatalf("%q should not contain %q", value, want)
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

func captureRequestHeadersEnabled() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS")), "true") {
		return true
	}
	for _, name := range strings.Split(os.Getenv("OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS"), ",") {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "content-type", "x-request-id", "authorization":
			return true
		}
	}
	return false
}

func captureBodyEnabled() bool {
	return strings.EqualFold(os.Getenv("OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED"), "true")
}
