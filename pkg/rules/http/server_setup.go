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
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	semconvhttp "github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
)

var netHttpServerInstrumenter = BuildNetHttpServerOtelInstrumenter()

//go:linkname serverOnEnter net/http.serverOnEnter
func serverOnEnter(call api.CallContext, _ interface{}, w http.ResponseWriter, r *http.Request) {
	if !netHttpEnabler.Enable() {
		return
	}
	if netHttpFilter.FilterUrl(r.URL) {
		return
	}
	request := &netHttpRequest{
		method:     r.Method,
		url:        r.URL,
		header:     r.Header,
		version:    getProtocolVersion(r.ProtoMajor, r.ProtoMinor),
		host:       r.Host,
		isTls:      r.TLS != nil,
		rawRequest: r,
	}
	container := &routeTemplateContainer{}
	rCtx := context.WithValue(r.Context(), routeContainerKey{}, container)
	*r = *r.WithContext(rCtx)
	ctx := netHttpServerInstrumenter.Start(rCtx, request)
	if x, ok := call.GetParam(1).(http.ResponseWriter); ok {
		x1 := &writerWrapper{ResponseWriter: x, statusCode: http.StatusOK}
		call.SetParam(1, x1)
	}
	data := make(map[string]interface{}, 2)
	data["ctx"] = ctx
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname serverOnExit net/http.serverOnExit
func serverOnExit(call api.CallContext) {
	if !netHttpEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request, ok := data["request"].(*netHttpRequest)
	if !ok {
		return
	}
	if request.rawRequest != nil {
		r := request.rawRequest
		if route := takeServerRouteTemplate(ctx, r); route != "" {
			request.routeTemplate = route
		} else if pattern := r.Pattern; pattern != "" {
			// ServeMux patterns (including static ones like "GET /query" or "/a")
			// are low-cardinality templates. Bare handlers leave Pattern empty.
			request.routeTemplate = semconvhttp.RouteFromPattern(pattern)
		}
	}
	if request.routeTemplate != "" {
		UpdateServerSpanName(request.method, request.routeTemplate)
	}

	if p, ok := call.GetParam(1).(http.ResponseWriter); ok {
		if w1, ok := p.(*writerWrapper); ok {
			netHttpServerInstrumenter.End(ctx, request, &netHttpResponse{
				statusCode: w1.statusCode,
			}, nil)
		}
	}

	return
}

type writerWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *writerWrapper) WriteHeader(statusCode int) {
	// cache the status code
	if w.statusCode == statusCode {
		return // Prevent multiple header writes
	}
	w.statusCode = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *writerWrapper) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWriter does not implement http.Hijacker")
}

func (w *writerWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *writerWrapper) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

func (w *writerWrapper) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
