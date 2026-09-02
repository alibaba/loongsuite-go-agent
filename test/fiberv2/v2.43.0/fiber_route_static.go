// Copyright (c) 2026 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var port int

func requestStaticServer() {
	client := &fasthttp.Client{}
	reqURL := "http://127.0.0.1:" + strconv.Itoa(port) + "/query"

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)
	if err := client.Do(req, resp); err != nil {
		panic(err)
	}
}

func setupStaticFiberServer() {
	app := fiber.New()
	app.Get("/query", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("ok")
	})

	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	if err := app.Listen("127.0.0.1:" + strconv.Itoa(port)); err != nil {
		panic(err)
	}
}

func main() {
	go setupStaticFiberServer()
	time.Sleep(2 * time.Second)
	requestStaticServer()
	peerAddr := "127.0.0.1:" + strconv.Itoa(port)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://"+peerAddr+"/query", "http", "", "tcp", "ipv4", "", peerAddr, 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /query", "GET", "http", "tcp", "ipv4", "", peerAddr, "fasthttp", "http", "/query", "", "/query", 200)
	}, 1)
}
