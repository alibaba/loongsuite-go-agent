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
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

const (
	httpCaptureRequestHeadersEnv = "OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS"
	httpCaptureAllHeadersEnv     = "LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS"
	httpCaptureBodyEnabledEnv    = "OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED"
	defaultMaxHTTPBodyBytes      = int64(1024)
	defaultMaxHTTPHeadersBytes   = 4096
)

var netHttpCaptureConfig = newHTTPCaptureConfigFromEnv()

type httpCaptureConfig struct {
	captureRequestHeaders bool
	captureAllHeaders     bool
	captureBody           bool
	maxBodyBytes          int64
	maxHeadersBytes       int
	requestHeaderNames    map[string]struct{}
}

func newHTTPCaptureConfigFromEnv() httpCaptureConfig {
	return newHTTPCaptureConfig(
		os.Getenv(httpCaptureRequestHeadersEnv),
		os.Getenv(httpCaptureAllHeadersEnv),
		os.Getenv(httpCaptureBodyEnabledEnv),
	)
}

func newHTTPCaptureConfig(headersValue, allHeadersValue, bodyValue string) httpCaptureConfig {
	requestHeaderNames := parseCaptureHeaderNames(headersValue)
	captureAllHeaders := strings.EqualFold(strings.TrimSpace(allHeadersValue), "true")
	return httpCaptureConfig{
		captureRequestHeaders: len(requestHeaderNames) > 0 || captureAllHeaders,
		captureAllHeaders:     captureAllHeaders,
		captureBody:           strings.EqualFold(strings.TrimSpace(bodyValue), "true"),
		maxBodyBytes:          defaultMaxHTTPBodyBytes,
		maxHeadersBytes:       defaultMaxHTTPHeadersBytes,
		requestHeaderNames:    requestHeaderNames,
	}
}

func (c httpCaptureConfig) captureHeaders(header http.Header) string {
	if !c.captureRequestHeaders || len(header) == 0 {
		return ""
	}
	captured := make(map[string][]string, len(header))
	for name, values := range header {
		normalized := normalizeHeaderName(name)
		if normalized == "" || !c.shouldCaptureHeader(normalized) {
			continue
		}
		captured[normalized] = append([]string(nil), values...)
	}
	if len(captured) == 0 {
		return ""
	}
	data, err := json.Marshal(captured)
	if err != nil {
		return ""
	}
	if c.maxHeadersBytes <= 0 || len(data) > c.maxHeadersBytes {
		return ""
	}
	return string(data)
}

func parseCaptureHeaderNames(value string) map[string]struct{} {
	names := map[string]struct{}{}
	for _, name := range strings.Split(value, ",") {
		normalized := normalizeHeaderName(name)
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

func (c httpCaptureConfig) shouldCaptureHeader(normalized string) bool {
	if _, ok := c.requestHeaderNames[normalized]; ok {
		return true
	}
	if !c.captureAllHeaders {
		return false
	}
	return !isSensitiveCaptureHeader(normalized)
}

func normalizeHeaderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func isSensitiveCaptureHeader(normalized string) bool {
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
