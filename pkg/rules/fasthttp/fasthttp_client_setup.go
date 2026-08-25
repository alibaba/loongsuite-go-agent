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

package fasthttp

import (
	"context"
	"net/url"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/valyala/fasthttp"
)

var fastHttpClientInstrumenter = BuildFastHttpClientOtelInstrumenter()

//go:linkname clientFastHttpOnEnter github.com/valyala/fasthttp.clientFastHttpOnEnter
func clientFastHttpOnEnter(call api.CallContext, c *fasthttp.HostClient, req *fasthttp.Request, resp *fasthttp.Response) {
	if !fastHttpEnabler.Enable() {
		return
	}
	scheme := req.URI().Scheme()
	isTLS := false
	if string(scheme) == "https" {
		isTLS = true
	}
	u, err := url.Parse(req.URI().String())
	if err != nil {
		return
	}
	request := fastHttpRequest{
		method:         string(req.Header.Method()),
		url:            u,
		isTls:          isTLS,
		header:         &req.Header,
		requestHeaders: captureFastHTTPRequestHeaders(&req.Header),
		requestBody:    captureFastHTTPRequestBody(req),
	}
	ctx := fastHttpClientInstrumenter.Start(context.Background(), request)
	data := make(map[string]interface{}, 3)
	data["ctx"] = ctx
	data["request"] = request
	data["response"] = resp
	call.SetData(data)
}

//go:linkname clientFastHttpOnExit github.com/valyala/fasthttp.clientFastHttpOnExit
func clientFastHttpOnExit(call api.CallContext, err error) {
	if !fastHttpEnabler.Enable() {
		return
	}
	dataRaw := call.GetData()
	if dataRaw == nil {
		return
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return
	}
	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}
	request := data["request"].(fastHttpRequest)
	resp := data["response"].(*fasthttp.Response)
	fastHttpClientInstrumenter.End(ctx, request, fastHttpResponse{
		statusCode:   resp.StatusCode(),
		header:       &resp.Header,
		responseBody: captureFastHTTPResponseBody(resp),
	}, err)
}
