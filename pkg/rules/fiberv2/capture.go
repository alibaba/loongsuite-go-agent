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

package fiberv2

import (
	semconvhttp "github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/valyala/fasthttp"
)

var fiberV2CaptureConfig = semconvhttp.NewCaptureConfigFromEnv()

type fiberV2CaptureAttrsGetter struct{}

func (g fiberV2CaptureAttrsGetter) GetCapturedRequestHeaders(request *fiberv2Request) string {
	return request.requestHeaders
}

func (g fiberV2CaptureAttrsGetter) GetCapturedRequestBody(request *fiberv2Request) string {
	return request.requestBody
}

func (g fiberV2CaptureAttrsGetter) GetCapturedResponseBody(request *fiberv2Request, response *fiberv2Response) string {
	return response.responseBody
}

func captureFiberV2RequestHeaders(header *fasthttp.RequestHeader) string {
	if header == nil {
		return ""
	}
	return fiberV2CaptureConfig.CaptureHeaders(func(add func(name string, value string)) {
		header.VisitAll(func(key []byte, value []byte) {
			add(string(key), string(value))
		})
	})
}

func captureFiberV2RequestBody(req *fasthttp.Request) string {
	if req == nil || !fiberV2CaptureConfig.CaptureBody || fiberV2CaptureConfig.MaxBodyBytes <= 0 {
		return ""
	}
	body := req.Body()
	if len(body) == 0 || int64(len(body)) > fiberV2CaptureConfig.MaxBodyBytes {
		return ""
	}
	return fiberV2CaptureConfig.CaptureBodyContent(
		body,
		string(req.Header.ContentType()),
		string(req.Header.ContentEncoding()),
		false,
	)
}

func captureFiberV2ResponseBody(resp *fasthttp.Response) string {
	if resp == nil {
		return ""
	}
	return fiberV2CaptureConfig.CaptureBodyContent(
		resp.Body(),
		string(resp.Header.ContentType()),
		string(resp.Header.ContentEncoding()),
		true,
	)
}
