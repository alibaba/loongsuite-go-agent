// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package iris

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	otelhttp "github.com/alibaba/loongsuite-go/pkg/rules/http"
	iContext "github.com/kataras/iris/v12/context"
)

//go:linkname irisHttpOnEnter github.com/kataras/iris/v12/core/router.irisHttpOnEnter
func irisHttpOnEnter(call api.CallContext, _ interface{}, iCtx *iContext.Context) {
	if !irisEnabler.Enable() || iCtx == nil {
		return
	}
	r := iCtx.Request()
	if r == nil || r.URL == nil {
		return
	}
	route := iCtx.Path()
	if !shouldUseIrisFallbackRoute(route, r.URL.Path) {
		return
	}
	otelhttp.SetServerRouteTemplate(r, route)
	otelhttp.UpdateServerSpanName(r.Method, route)
}

func shouldUseIrisFallbackRoute(route, requestPath string) bool {
	return route != "" && route != requestPath
}

//go:linkname irisSetCurrentRouteOnEnter github.com/kataras/iris/v12/context.irisSetCurrentRouteOnEnter
func irisSetCurrentRouteOnEnter(call api.CallContext, iCtx *iContext.Context, curr iContext.RouteReadOnly) {
	if !irisEnabler.Enable() || iCtx == nil || curr == nil {
		return
	}
	r := iCtx.Request()
	if r == nil {
		return
	}
	var route string
	tmpl := curr.Tmpl()
	if tmpl.Src != "" {
		route = tmpl.Src
	} else if curr.Path() != "" {
		route = curr.Path()
	}
	if route == "" {
		return
	}
	otelhttp.SetServerRouteTemplate(r, route)
	otelhttp.UpdateServerSpanName(r.Method, route)
}
