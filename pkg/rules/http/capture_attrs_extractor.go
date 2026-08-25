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

	"go.opentelemetry.io/otel/attribute"
)

const (
	httpRequestHeadersAttr      = "http.request.headers"
	httpRequestBodyContentAttr  = "http.request.body.content"
	httpResponseBodyContentAttr = "http.response.body.content"
)

type httpCaptureAttrsExtractor struct{}

func (h *httpCaptureAttrsExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request *netHttpRequest) ([]attribute.KeyValue, context.Context) {
	if request == nil {
		return attributes, parentContext
	}
	if request.requestHeaders != "" {
		attributes = append(attributes, attribute.String(httpRequestHeadersAttr, request.requestHeaders))
	}
	if request.requestBody != "" {
		attributes = append(attributes, attribute.String(httpRequestBodyContentAttr, request.requestBody))
	}
	return attributes, parentContext
}

func (h *httpCaptureAttrsExtractor) OnEnd(attributes []attribute.KeyValue, ctx context.Context, request *netHttpRequest, response *netHttpResponse, err error) ([]attribute.KeyValue, context.Context) {
	if response == nil || response.responseBody == "" {
		return attributes, ctx
	}
	attributes = append(attributes, attribute.String(httpResponseBodyContentAttr, response.responseBody))
	return attributes, ctx
}
