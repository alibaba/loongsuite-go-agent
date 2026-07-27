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

package gin

import (
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	otelhttp "github.com/alibaba/loongsuite-go/pkg/rules/http"

	"github.com/gin-gonic/gin"
)

//go:linkname nextOnEnter github.com/gin-gonic/gin.nextOnEnter
func nextOnEnter(call api.CallContext, c *gin.Context) {
	if !ginEnabler.Enable() {
		return
	}
	if c == nil {
		return
	}
	route := c.FullPath()
	if route == "" || c.Request == nil {
		return
	}
	otelhttp.SetServerRouteTemplate(c.Request, route)
	otelhttp.UpdateServerSpanName(c.Request.Method, route)
}
