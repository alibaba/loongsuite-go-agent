// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fiberv2

import (
	"context"
	"net/url"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

var fiberv2ServerInstrumenter = BuildFiberV2ServerOtelInstrumenter()

//go:linkname fiberHttpOnEnterv2 github.com/gofiber/fiber/v2.fiberHttpOnEnterv2
func fiberHttpOnEnterv2(call api.CallContext, app *fiber.App, ctx *fasthttp.RequestCtx) {
	if !fiberV2Enabler.Enable() {
		return
	}
	u, err := url.Parse(ctx.URI().String())
	if err != nil {
		return
	}
	request := &fiberv2Request{
		method:         string(ctx.Method()),
		url:            u,
		isTls:          ctx.IsTLS(),
		header:         &ctx.Request.Header,
		requestHeaders: captureFiberV2RequestHeaders(&ctx.Request.Header),
		requestBody:    captureFiberV2RequestBody(&ctx.Request),
	}
	ctxSpan := fiberv2ServerInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 2)
	data["ctx"] = ctx
	data["ctxSpan"] = ctxSpan
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname fiberHttpOnExitv2 github.com/gofiber/fiber/v2.fiberHttpOnExitv2
func fiberHttpOnExitv2(call api.CallContext) {
	if call.GetData() == nil {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(*fasthttp.RequestCtx)
	ctxSpan := data["ctxSpan"].(context.Context)
	request, ok := data["request"].(*fiberv2Request)
	if !ok {
		return
	}
	fiberv2ServerInstrumenter.End(ctxSpan, request, &fiberv2Response{
		statusCode:   ctx.Response.StatusCode(),
		header:       &ctx.Response.Header,
		responseBody: captureFiberV2ResponseBody(&ctx.Response),
	}, nil)

}
