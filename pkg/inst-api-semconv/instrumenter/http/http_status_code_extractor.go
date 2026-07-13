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
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type HttpClientSpanStatusExtractor[REQUEST any, RESPONSE any] struct {
	Getter HttpCommonAttrsGetter[REQUEST, RESPONSE]
}

func (h HttpClientSpanStatusExtractor[REQUEST, RESPONSE]) Extract(span trace.Span, request REQUEST, response RESPONSE, err error) {
	// Minor #10: 客户端 Extract 在 err != nil 时提前返回，不调用 span.RecordError(err)。
	// 这是因为 InternalInstrumenter.doEnd 会优先在 Extractor 之前调用 RecordError 并将状态设为 Error，
	// 此处直接 early return 避免重复记录，属于有意设计的隐式依赖关系。
	if err != nil {
		return
	}
	statusCode := h.Getter.GetHttpResponseStatusCode(request, response, err)
	if statusCode >= 400 || statusCode < 100 {
		span.SetStatus(codes.Error, "")
	} else if statusCode >= 200 && statusCode < 300 {
		span.SetStatus(codes.Ok, "success")
	}
}

type HttpServerSpanStatusExtractor[REQUEST any, RESPONSE any] struct {
	Getter HttpCommonAttrsGetter[REQUEST, RESPONSE]
}

func (h HttpServerSpanStatusExtractor[REQUEST, RESPONSE]) Extract(span trace.Span, request REQUEST, response RESPONSE, err error) {
	statusCode := h.Getter.GetHttpResponseStatusCode(request, response, err)
	if statusCode >= 500 || statusCode < 100 {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Error, "")
		}
	} else if statusCode >= 200 && statusCode < 300 {
		span.SetStatus(codes.Ok, "success")
	}
}
