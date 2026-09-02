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

package server

import (
	"context"
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
)

type hertzServerInnerEnabler struct {
	enabled bool
}

func (h hertzServerInnerEnabler) Enable() bool {
	return h.enabled
}

var hertzServerEnabler = hertzServerInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_HERTZ_ENABLED") != "false"}

var hertzInstrumenter = BuildHertzServerInstrumenter()

type hertzOpentelemetryTracer struct{}

func (m *hertzOpentelemetryTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	return ctx
}

func (m *hertzOpentelemetryTracer) Finish(ctx context.Context, c *app.RequestContext) {
	if c.GetTraceInfo().Stats().GetEvent(stats.HTTPStart) != nil && c.GetTraceInfo().Stats().GetEvent(stats.HTTPFinish) != nil {
		start := c.GetTraceInfo().Stats().GetEvent(stats.HTTPStart)
		end := c.GetTraceInfo().Stats().GetEvent(stats.HTTPFinish)
		if ctx == nil {
			ctx = context.Background()
		}
		s := start.Time()
		e := end.Time()
		req, resp := &c.Request, &c.Response
		wrapped := &hertzServerRequest{
			req:           req,
			routeTemplate: c.FullPath(),
		}
		insctx := hertzInstrumenter.StartWithTime(ctx, wrapped, s)
		hertzInstrumenter.EndWithTime(insctx, wrapped, resp, c.GetTraceInfo().Stats().Error(), e)
	}
}

//go:linkname beforeHertzServerBuild github.com/cloudwego/hertz/pkg/app/server.beforeHertzServerBuild
func beforeHertzServerBuild(call api.CallContext, opts ...config.Option) {
	if !hertzServerEnabler.Enable() {
		return
	}
	opts = append(opts, server.WithTracer(&hertzOpentelemetryTracer{}))
	call.SetParam(0, opts)
}
