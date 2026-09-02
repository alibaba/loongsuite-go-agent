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

package main

import (
	"context"
	"net/http"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	iris "github.com/kataras/iris/v12"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupPatternHttp() {
	app := iris.New()
	app.Get("/users/{id:int}", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"id": ctx.Params().Get("id")})
	})
	app.Run(iris.Addr(":8080"))
}

func main() {
	go setupPatternHttp()
	time.Sleep(3 * time.Second)
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://127.0.0.1:8080/users/42", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8080/users/42", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:8080", 200, 0, 8080)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /users/{id:int}", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8080", "Go-http-client/1.1", "http", "/users/42", "", "/users/{id:int}", 200)
	}, 1)
}
