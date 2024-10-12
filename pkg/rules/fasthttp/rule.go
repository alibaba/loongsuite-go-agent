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

package fasthttp

import "github.com/alibaba/opentelemetry-go-auto-instrumentation/api"

func init() {
	api.NewRule("github.com/valyala/fasthttp", "Do", "*HostClient", "clientFastHttpOnEnter", "clientFastHttpOnExit").
		WithVersion("[1.45.0,1.56.1)").
		WithFileDeps("fasthttp_data_type.go", "fasthttp_otel_instrumenter.go").
		Register()

	api.NewRule("github.com/valyala/fasthttp", "ListenAndServe", "*Server", "listenAndServeFastHttpOnEnter", "").
		WithVersion("[1.45.0,1.56.1)").
		WithFileDeps("fasthttp_data_type.go", "fasthttp_otel_instrumenter.go").
		Register()
}
