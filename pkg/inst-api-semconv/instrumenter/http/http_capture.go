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

package http

import (
	"context"
	"encoding/json"
	"mime"
	stdhttp "net/http"
	"os"
	"strings"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
)

const (
	CaptureRequestHeadersEnv = "OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS"
	CaptureAllHeadersEnv     = "LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS"
	CaptureBodyEnabledEnv    = "OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED"
	DefaultMaxCaptureBytes   = int64(1024)
	DefaultMaxHeadersBytes   = 4096

	RequestHeadersAttr      = "http.request.headers"
	RequestBodyContentAttr  = "http.request.body.content"
	ResponseBodyContentAttr = "http.response.body.content"
)

type CaptureConfig struct {
	CaptureRequestHeaders bool
	CaptureAllHeaders     bool
	CaptureBody           bool
	MaxBodyBytes          int64
	MaxHeadersBytes       int
	RequestHeaderNames    map[string]struct{}
}

func NewCaptureConfigFromEnv() CaptureConfig {
	return NewCaptureConfigWithAllHeaders(
		os.Getenv(CaptureRequestHeadersEnv),
		os.Getenv(CaptureAllHeadersEnv),
		os.Getenv(CaptureBodyEnabledEnv),
	)
}

func NewCaptureConfig(headersValue, bodyValue string) CaptureConfig {
	return NewCaptureConfigWithAllHeaders(headersValue, "", bodyValue)
}

func NewCaptureConfigWithAllHeaders(headersValue, allHeadersValue, bodyValue string) CaptureConfig {
	requestHeaderNames := ParseCaptureHeaderNames(headersValue)
	captureAllHeaders := strings.EqualFold(strings.TrimSpace(allHeadersValue), "true")
	return CaptureConfig{
		CaptureRequestHeaders: len(requestHeaderNames) > 0 || captureAllHeaders,
		CaptureAllHeaders:     captureAllHeaders,
		CaptureBody:           strings.EqualFold(strings.TrimSpace(bodyValue), "true"),
		MaxBodyBytes:          DefaultMaxCaptureBytes,
		MaxHeadersBytes:       DefaultMaxHeadersBytes,
		RequestHeaderNames:    requestHeaderNames,
	}
}

func (c CaptureConfig) CaptureHeaders(visit func(func(name string, value string))) string {
	if !c.CaptureRequestHeaders || visit == nil {
		return ""
	}
	captured := map[string][]string{}
	visit(func(name string, value string) {
		normalized := NormalizeCaptureHeaderName(name)
		if normalized == "" || !c.ShouldCaptureHeader(normalized) {
			return
		}
		captured[normalized] = append(captured[normalized], value)
	})
	if len(captured) == 0 {
		return ""
	}
	data, err := json.Marshal(captured)
	if err != nil {
		return ""
	}
	if c.MaxHeadersBytes <= 0 || len(data) > c.MaxHeadersBytes {
		return ""
	}
	return string(data)
}

func ParseCaptureHeaderNames(value string) map[string]struct{} {
	names := map[string]struct{}{}
	for _, name := range strings.Split(value, ",") {
		normalized := NormalizeCaptureHeaderName(name)
		if normalized == "" {
			continue
		}
		names[normalized] = struct{}{}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func (c CaptureConfig) ShouldCaptureHeader(normalized string) bool {
	if _, ok := c.RequestHeaderNames[normalized]; ok {
		return true
	}
	if !c.CaptureAllHeaders {
		return false
	}
	return !IsSensitiveCaptureHeader(normalized)
}

func (c CaptureConfig) CaptureBodyContent(body []byte, contentType string, contentEncoding string, detectContentType bool) string {
	if !c.CaptureBody || c.MaxBodyBytes <= 0 || len(body) == 0 || int64(len(body)) > c.MaxBodyBytes {
		return ""
	}
	if !IsAllowedCaptureContentEncoding(contentEncoding) {
		return ""
	}
	if contentType == "" && detectContentType {
		contentType = stdhttp.DetectContentType(body)
	}
	if !IsTextOrJSONCaptureContentType(contentType) || !utf8.Valid(body) {
		return ""
	}
	return string(body)
}

func NormalizeCaptureHeaderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func IsSensitiveCaptureHeader(normalized string) bool {
	switch normalized {
	case "authorization",
		"cookie",
		"proxy-authorization",
		"set-cookie",
		"x-api-key",
		"x-access-token":
		return true
	default:
		return false
	}
}

func IsTextOrJSONCaptureContentType(contentType string) bool {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.TrimSpace(strings.Split(contentType, ";")[0])
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	return mediaType == "application/json" ||
		(strings.HasPrefix(mediaType, "application/") && strings.HasSuffix(mediaType, "+json"))
}

func IsAllowedCaptureContentEncoding(contentEncoding string) bool {
	contentEncoding = strings.TrimSpace(contentEncoding)
	return contentEncoding == "" || strings.EqualFold(contentEncoding, "identity")
}

type CaptureAttrsGetter[REQUEST any, RESPONSE any] interface {
	GetCapturedRequestHeaders(request REQUEST) string
	GetCapturedRequestBody(request REQUEST) string
	GetCapturedResponseBody(request REQUEST, response RESPONSE) string
}

type CaptureAttrsExtractor[REQUEST any, RESPONSE any, GETTER CaptureAttrsGetter[REQUEST, RESPONSE]] struct {
	Getter GETTER
}

func (h *CaptureAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request REQUEST) ([]attribute.KeyValue, context.Context) {
	if headers := h.Getter.GetCapturedRequestHeaders(request); headers != "" {
		attributes = append(attributes, attribute.String(RequestHeadersAttr, headers))
	}
	if body := h.Getter.GetCapturedRequestBody(request); body != "" {
		attributes = append(attributes, attribute.String(RequestBodyContentAttr, body))
	}
	return attributes, parentContext
}

func (h *CaptureAttrsExtractor[REQUEST, RESPONSE, GETTER]) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request REQUEST, response RESPONSE, err error) ([]attribute.KeyValue, context.Context) {
	if body := h.Getter.GetCapturedResponseBody(request, response); body != "" {
		attributes = append(attributes, attribute.String(ResponseBodyContentAttr, body))
	}
	return attributes, ctx
}
