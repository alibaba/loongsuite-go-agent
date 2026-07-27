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
	"strings"
)

// RouteFromPattern extracts the path portion from a Go 1.22+ ServeMux pattern.
// For example, "GET /users/{id}" becomes "/users/{id}".
func RouteFromPattern(pattern string) string {
	if idx := strings.IndexByte(pattern, '/'); idx >= 0 {
		return pattern[idx:]
	}
	return ""
}

// ResolveHttpServerRoute returns the route template when available, otherwise "".
func ResolveHttpServerRoute[REQUEST any, RESPONSE any](getter HttpServerAttrsGetter[REQUEST, RESPONSE], request REQUEST) string {
	return getter.GetHttpRoute(request)
}
