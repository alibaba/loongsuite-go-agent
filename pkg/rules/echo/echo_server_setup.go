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

package echo

import (
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	otelhttp "github.com/alibaba/loongsuite-go/pkg/rules/http"
	echo "github.com/labstack/echo/v4"
)

type echoInnerEnabler struct {
	enabled bool
}

func (e echoInnerEnabler) Enable() bool {
	return e.enabled
}

var echoEnabler = echoInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_ECHO_ENABLED") != "false"}

func otelTraceMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if err = next(c); err != nil {
				c.Error(err)
			}
			route := c.Path()
			if route != "" && c.Request() != nil {
				otelhttp.SetServerRouteTemplate(c.Request(), route)
				otelhttp.UpdateServerSpanName(c.Request().Method, route)
			}
			return
		}
	}
}

//go:linkname afterNewEcho github.com/labstack/echo/v4.afterNewEcho
func afterNewEcho(call api.CallContext, e *echo.Echo) {
	if !echoEnabler.Enable() {
		return
	}
	if e == nil {
		return
	}

	e.Use(otelTraceMiddleware())
}
