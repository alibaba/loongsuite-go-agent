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

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupPattern() {
	engine := echo.New()
	engine.Use(middleware.Logger())
	engine.GET("/users/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": 1,
			"msg":  c.Path(),
		})
	})

	// Start server
	engine.Logger.Fatal(engine.Start(":8080"))
}

func main() {
	go setupPattern()
	time.Sleep(5 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://127.0.0.1:8080/users/1")
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8080/users/1", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:8080", 200, 0, 8080)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /users/:id", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8080", "Go-http-client/1.1", "http", "/users/1", "", "/users/:id", 200)
	}, 1)
}
