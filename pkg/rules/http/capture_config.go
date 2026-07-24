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
	httpCaptureBodyEnabledEnv    = "OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED"
	defaultMaxHTTPBodyBytes      = int64(1024)
)

var netHttpCaptureConfig = newHTTPCaptureConfigFromEnv()

type httpCaptureConfig struct {
	captureRequestHeaders bool
	captureBody           bool
	maxBodyBytes          int64
}

func newHTTPCaptureConfigFromEnv() httpCaptureConfig {
	return newHTTPCaptureConfig(
		os.Getenv(httpCaptureRequestHeadersEnv),
		os.Getenv(httpCaptureBodyEnabledEnv),
	)
}

func newHTTPCaptureConfig(headersValue, bodyValue string) httpCaptureConfig {
	return httpCaptureConfig{
		captureRequestHeaders: strings.EqualFold(strings.TrimSpace(headersValue), "true"),
		captureBody:           strings.EqualFold(strings.TrimSpace(bodyValue), "true"),
		maxBodyBytes:          defaultMaxHTTPBodyBytes,
	}
}

func (c httpCaptureConfig) captureHeaders(header http.Header) string {
	if !c.captureRequestHeaders || len(header) == 0 {
		return ""
	}
	captured := make(map[string][]string, len(header))
	for name, values := range header {
		normalized := normalizeHeaderName(name)
		if normalized == "" {
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
	return string(data)
}

func normalizeHeaderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
