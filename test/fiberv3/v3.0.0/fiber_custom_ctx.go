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
	"fmt"
	"log"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

// customCtx is a minimal custom Fiber context. Registering it via
// NewWithCustomCtx makes the app serve requests through
// (*App).customRequestHandler instead of (*App).defaultRequestHandler. That
// method is the request entry point the fiberv3 rule hooks on fiber >= v3.3.0,
// so this app exercises the customRequestHandler instrumentation path.
type customCtx struct {
	*fiber.DefaultCtx
}

func requestServer() {
	client := &fasthttp.Client{}

	reqURL := "http://localhost:3000/fiber"

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)

	err := client.Do(req, resp)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Response body is:\n%s", resp.Body())
}

func setupHttp() {
	// Initialize a new Fiber app with a custom context implementation so that
	// requests are routed through customRequestHandler.
	app := fiber.NewWithCustomCtx(func(app *fiber.App) fiber.CustomCtx {
		return &customCtx{DefaultCtx: fiber.NewDefaultCtx(app)}
	})

	// Define a route for the GET method on the path '/fiber'
	app.Get("/fiber", func(c fiber.Ctx) error {
		// Send a string response to the client
		return c.SendString("Hello, World 👋!")
	})

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}

func main() {
	// starter server
	go setupHttp()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	requestServer()
	// verify trace
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://localhost:3000/fiber", "http", "", "tcp", "ipv4", "", "localhost:3000", 200, 0, 3000)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /fiber", "GET", "http", "tcp", "ipv4", "", "localhost:3000", "fasthttp", "http", "/fiber", "", "/fiber", 200)
	}, 1)
}
