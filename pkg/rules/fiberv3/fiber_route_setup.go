// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package fiberv3

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	fiber "github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

//go:linkname fiberRouteOnEnterv3 github.com/gofiber/fiber/v3.fiberRouteOnEnterv3
func fiberRouteOnEnterv3(call api.CallContext, _ *fiber.App, c fiber.CustomCtx) {
	applyFiberRoute(c)
}

//go:linkname fiberRouteOnEnterv3DefaultCtx github.com/gofiber/fiber/v3.fiberRouteOnEnterv3DefaultCtx
func fiberRouteOnEnterv3DefaultCtx(call api.CallContext, _ *fiber.App, c *fiber.DefaultCtx) {
	applyFiberRoute(c)
}


func applyFiberRoute(c fiber.Ctx) {
	if !fiberV3Enabler.Enable() || c == nil {
		return
	}
	r := c.Route()
	if r == nil || len(r.Handlers) == 0 {
		return
	}
	route := r.Path
	if route == "" {
		return
	}
	setFiberRouteTemplate(c.RequestCtx(), route)
}

func setFiberRouteTemplate(ctx *fasthttp.RequestCtx, route string) {
	if ctx != nil && route != "" {
		ctx.SetUserValue(fiberRouteTemplateUserValueKey, route)
	}
}

func takeFiberRouteTemplate(ctx *fasthttp.RequestCtx) string {
	if ctx == nil {
		return ""
	}
	v := ctx.UserValue(fiberRouteTemplateUserValueKey)
	ctx.SetUserValue(fiberRouteTemplateUserValueKey, nil)
	if route, ok := v.(string); ok {
		return route
	}
	return ""
}
