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

package http

import (
	"context"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel/sdk/trace"
)

type routeTemplateContainer struct {
	template string
}

type routeContainerKey struct{}

var serverRouteTemplates sync.Map

// SetServerRouteTemplate stores a low-cardinality route template for the request.
// Framework hooks call this with the underlying *http.Request before the server
// instrumenter ends the span.
func SetServerRouteTemplate(r *http.Request, route string) {
	if r == nil || route == "" {
		return
	}
	if container, ok := r.Context().Value(routeContainerKey{}).(*routeTemplateContainer); ok && container != nil {
		container.template = route
		return
	}
	// Fallback when the net/http server hook has not injected a container
	// (or the request context was replaced without preserving values).
	serverRouteTemplates.Store(r, route)
}

func takeServerRouteTemplate(ctx context.Context, r *http.Request) string {
	if ctx != nil {
		if container, ok := ctx.Value(routeContainerKey{}).(*routeTemplateContainer); ok && container != nil && container.template != "" {
			return container.template
		}
	}
	if r != nil {
		if container, ok := r.Context().Value(routeContainerKey{}).(*routeTemplateContainer); ok && container != nil && container.template != "" {
			return container.template
		}
		if v, ok := serverRouteTemplates.LoadAndDelete(r); ok {
			return v.(string)
		}
	}
	return ""
}

// UpdateServerSpanName sets the local-root HTTP server span name to
// "{method} {route}" per OTel HTTP semconv when a route template is available.
func UpdateServerSpanName(method, route string) {
	if route == "" {
		return
	}
	lcs := trace.LocalRootSpanFromGLS()
	if lcs == nil {
		return
	}
	if method == "" {
		lcs.SetName(route)
		return
	}
	lcs.SetName(method + " " + route)
}
