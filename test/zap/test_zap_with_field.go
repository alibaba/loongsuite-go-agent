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
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"time"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger.Sync()
	ts_a := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("test info", zap.String("exInfo", "one"))
		_, _ = w.Write([]byte("success"))
	}))
	defer ts_a.Close()

	_, err = http.Get(ts_a.URL)
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)
	Url, err := url.Parse(ts_a.URL)
	if err != nil {
		panic(err)
	}
	port, err := strconv.Atoi(Url.Port())
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", ts_a.URL, "http", "1.1", "tcp", "ipv4", "", Url.Host, 200, 0, int64(port))
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET", "GET", "http", "tcp", "ipv4", "", Url.Host, "Go-http-client/1.1", "http", "/", "", verifier.OmitHttpRoute, 200)
		if stubs[0][1].Parent.TraceID().String() != stubs[0][0].SpanContext.TraceID().String() {
			log.Fatal("span 1 should be child of span 0")
		}
	}, 1)
}
