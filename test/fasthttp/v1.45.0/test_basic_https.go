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
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	go func() {
		fasthttp.ListenAndServeTLS(":8080", "server.crt", "server.key", hello)
	}()
	time.Sleep(5 * time.Second)
	client := &fasthttp.Client{}
	reqURL := "https://localhost:8080"
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

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "https://localhost:8080/", "https", "", "tcp", "ipv4", "", "localhost:8080", 200, 0, 8080)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET", "GET", "http", "tcp", "ipv4", "", "localhost:8080", "fasthttp", "http", "/", "", verifier.OmitHttpRoute, 200)
	}, 1)
}
