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

package fiberv3

import (
	semconvhttp "github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/valyala/fasthttp"
)

var fiberV3CaptureConfig = semconvhttp.NewCaptureConfigFromEnv()

type fiberV3CaptureAttrsGetter struct{}

func (g fiberV3CaptureAttrsGetter) GetCapturedRequestHeaders(request *fiberv3Request) string {
	return request.requestHeaders
}

func (g fiberV3CaptureAttrsGetter) GetCapturedRequestBody(request *fiberv3Request) string {
	return request.requestBody
}

func (g fiberV3CaptureAttrsGetter) GetCapturedResponseBody(request *fiberv3Request, response *fiberv3Response) string {
	return response.responseBody
}

func captureFiberV3RequestHeaders(header *fasthttp.RequestHeader) string {
	if header == nil {
		return ""
	}
	return fiberV3CaptureConfig.CaptureHeaders(func(add func(name string, value string)) {
		header.VisitAll(func(key []byte, value []byte) {
			add(string(key), string(value))
		})
	})
}

func captureFiberV3RequestBody(req *fasthttp.Request) string {
	if req == nil || !fiberV3CaptureConfig.CaptureBody || fiberV3CaptureConfig.MaxBodyBytes <= 0 {
		return ""
	}
	body := req.Body()
	if len(body) == 0 || int64(len(body)) > fiberV3CaptureConfig.MaxBodyBytes {
		return ""
	}
	return fiberV3CaptureConfig.CaptureBodyContent(
		body,
		string(req.Header.ContentType()),
		string(req.Header.ContentEncoding()),
		false,
	)
}

func captureFiberV3ResponseBody(resp *fasthttp.Response) string {
	if resp == nil {
		return ""
	}
	return fiberV3CaptureConfig.CaptureBodyContent(
		resp.Body(),
		string(resp.Header.ContentType()),
		string(resp.Header.ContentEncoding()),
		true,
	)
}
