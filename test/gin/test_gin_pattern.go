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
	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"time"
)

func setupPattern() {
	router := gin.Default()
	router.GET("/user/:name", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	router.Run(":8080")
}

func main() {
	go setupPattern()
	time.Sleep(3 * time.Second)
	client := http.Client{}
	client.Get("http://127.0.0.1:8080/user/abc")
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8080/user/abc", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:8080", 200, 0, 8080)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /user/:name", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8080", "Go-http-client/1.1", "http", "/user/abc", "", "/user/:name", 200)
	}, 1)
}
