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
	"net/http"
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupStaticPatternHttp() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /query", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		panic(err)
	}
}

func main() {
	go setupStaticPatternHttp()
	time.Sleep(1 * time.Second)
	_, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/query")
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:"+strconv.Itoa(port)+"/query", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /query", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:"+strconv.Itoa(port), "Go-http-client/1.1", "http", "/query", "", "/query", 200)
	}, 1)
}
