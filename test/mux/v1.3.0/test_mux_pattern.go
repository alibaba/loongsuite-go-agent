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
	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"io/ioutil"
	"net/http"
	"time"
)

func setupPattern() {
	r := mux.NewRouter()
	s := r.PathPrefix("/test").Subrouter()
	s.HandleFunc("/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}

func main() {
	go setupPattern()
	time.Sleep(5 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://127.0.0.1:8080/test/1")
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8080/test/1", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:8080", 200, 0, 8080)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /test/{key}", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8080", "Go-http-client/1.1", "http", "/test/1", "", "/test/{key}", 200)
	}, 1)
}
