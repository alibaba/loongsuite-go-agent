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

package fasthttp

import (
	semconvhttp "github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/valyala/fasthttp"
)

var fastHttpCaptureConfig = semconvhttp.NewCaptureConfigFromEnv()

type fastHttpCaptureAttrsGetter struct{}

func (g fastHttpCaptureAttrsGetter) GetCapturedRequestHeaders(request fastHttpRequest) string {
	return request.requestHeaders
}

func (g fastHttpCaptureAttrsGetter) GetCapturedRequestBody(request fastHttpRequest) string {
	return request.requestBody
}

func (g fastHttpCaptureAttrsGetter) GetCapturedResponseBody(request fastHttpRequest, response fastHttpResponse) string {
	return response.responseBody
}

func captureFastHTTPRequestHeaders(header *fasthttp.RequestHeader) string {
	if header == nil {
		return ""
	}
	return fastHttpCaptureConfig.CaptureHeaders(func(add func(name string, value string)) {
		header.VisitAll(func(key []byte, value []byte) {
			add(string(key), string(value))
		})
	})
}

func captureFastHTTPRequestBody(req *fasthttp.Request) string {
	if req == nil || !fastHttpCaptureConfig.CaptureBody || fastHttpCaptureConfig.MaxBodyBytes <= 0 {
		return ""
	}
	body := req.Body()
	if len(body) == 0 || int64(len(body)) > fastHttpCaptureConfig.MaxBodyBytes {
		return ""
	}
	return fastHttpCaptureConfig.CaptureBodyContent(
		body,
		string(req.Header.ContentType()),
		string(req.Header.ContentEncoding()),
		false,
	)
}

func captureFastHTTPResponseBody(resp *fasthttp.Response) string {
	if resp == nil {
		return ""
	}
	return fastHttpCaptureConfig.CaptureBodyContent(
		resp.Body(),
		string(resp.Header.ContentType()),
		string(resp.Header.ContentEncoding()),
		true,
	)
}
