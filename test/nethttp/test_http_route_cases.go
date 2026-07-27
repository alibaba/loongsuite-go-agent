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
	"strings"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupRouteCasesHttp() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /query", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/fallback") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		mux.ServeHTTP(w, r)
	})
	err = http.ListenAndServe(":"+strconv.Itoa(port), handler)
	if err != nil {
		panic(err)
	}
}

func main() {
	go setupRouteCasesHttp()
	time.Sleep(1 * time.Second)
	base := "http://127.0.0.1:" + strconv.Itoa(port)
	if _, err := http.Get(base + "/users/123"); err != nil {
		panic(err)
	}
	if _, err := http.Get(base + "/query"); err != nil {
		panic(err)
	}
	if _, err := http.Get(base + "/fallback/foo"); err != nil {
		panic(err)
	}
	peer := "127.0.0.1:" + strconv.Itoa(port)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", base+"/users/123", "http", "1.1", "tcp", "ipv4", "", peer, 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /users/{id}", "GET", "http", "tcp", "ipv4", "", peer, "Go-http-client/1.1", "http", "/users/123", "", "/users/{id}", 200)

		verifier.VerifyHttpClientAttributes(stubs[1][0], "GET", "GET", base+"/query", "http", "1.1", "tcp", "ipv4", "", peer, 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[1][1], "GET /query", "GET", "http", "tcp", "ipv4", "", peer, "Go-http-client/1.1", "http", "/query", "", "/query", 200)

		verifier.VerifyHttpClientAttributes(stubs[2][0], "GET", "GET", base+"/fallback/foo", "http", "1.1", "tcp", "ipv4", "", peer, 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[2][1], "GET", "GET", "http", "tcp", "ipv4", "", peer, "Go-http-client/1.1", "http", "/fallback/foo", "", verifier.OmitHttpRoute, 200)
	}, 3)
}
