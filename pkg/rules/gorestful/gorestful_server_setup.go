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

package gorestful

import (
	"net/http"
	"os"
	"sync"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	otelhttp "github.com/alibaba/loongsuite-go/pkg/rules/http"
	restful "github.com/emicklei/go-restful/v3"
)

type goRestfulInnerEnabler struct {
	enabled bool
}

func (g goRestfulInnerEnabler) Enable() bool {
	return g.enabled
}

var goRestfulEnabler = goRestfulInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_GORESTFUL_ENABLED") != "false"}

var registeredContainers sync.Map

func ensureFilterRegistered(c *restful.Container) {
	if c == nil {
		return
	}
	if _, loaded := registeredContainers.LoadOrStore(c, struct{}{}); !loaded {
		c.Filter(filterRest)
	}
}

//go:linkname restContainerAddOnEnter github.com/emicklei/go-restful/v3.restContainerAddOnEnter
func restContainerAddOnEnter(call api.CallContext, c *restful.Container, service *restful.WebService) {
	ensureFilterRegistered(c)
	call.SetParam(0, c)
}

//go:linkname restContainerAddOnExit github.com/emicklei/go-restful/v3.restContainerAddOnExit
func restContainerAddOnExit(call api.CallContext, c *restful.Container) {
	return
}

//go:linkname restContainerDispatchOnEnter github.com/emicklei/go-restful/v3.restContainerDispatchOnEnter
func restContainerDispatchOnEnter(call api.CallContext, c *restful.Container, httpWriter http.ResponseWriter, httpRequest *http.Request) {
	ensureFilterRegistered(c)
	call.SetParam(0, c)
}

//go:linkname restContainerDispatchOnExit github.com/emicklei/go-restful/v3.restContainerDispatchOnExit
func restContainerDispatchOnExit(call api.CallContext) {
	return
}

//go:linkname restContainerHandleOnEnter github.com/emicklei/go-restful/v3.restContainerHandleOnEnter
func restContainerHandleOnEnter(call api.CallContext, c *restful.Container, pattern string, handler http.Handler) {
	ensureFilterRegistered(c)
	call.SetParam(0, c)
}

//go:linkname restContainerHandleOnExit github.com/emicklei/go-restful/v3.restContainerHandleOnExit
func restContainerHandleOnExit(call api.CallContext) {
	return
}

var filterRest = func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if goRestfulEnabler.Enable() && req != nil {
		route := req.SelectedRoutePath()
		if route != "" && req.Request != nil {
			otelhttp.SetServerRouteTemplate(req.Request, route)
			otelhttp.UpdateServerSpanName(req.Request.Method, route)
		}
	}
	chain.ProcessFilter(req, resp)
}
