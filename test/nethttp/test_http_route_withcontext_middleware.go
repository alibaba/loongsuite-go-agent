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
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Mirrors collector confighttp: outer middleware does r = r.WithContext(...),
// then ServeMux sets Pattern on the copy. http.route must still be recorded.
func setupWithContextMiddlewareHttp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/traces", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	type mwKey struct{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), mwKey{}, "middleware"))
		mux.ServeHTTP(w, r)
	})

	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":"+strconv.Itoa(port), handler)
	if err != nil {
		panic(err)
	}
}

func main() {
	go setupWithContextMiddlewareHttp()
	time.Sleep(1 * time.Second)
	base := "http://127.0.0.1:" + strconv.Itoa(port)
	if _, err := http.Post(base+"/v1/traces", "application/json", nil); err != nil {
		panic(err)
	}
	if _, err := http.Post(base+"/v1/metrics", "application/json", nil); err != nil {
		panic(err)
	}

	peer := "127.0.0.1:" + strconv.Itoa(port)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpServerAttributes(stubs[0][1], "POST /v1/traces", "POST", "http", "tcp", "ipv4", "", peer, "Go-http-client/1.1", "http", "/v1/traces", "", "/v1/traces", 200)
		verifier.VerifyHttpServerAttributes(stubs[1][1], "POST /v1/metrics", "POST", "http", "tcp", "ipv4", "", peer, "Go-http-client/1.1", "http", "/v1/metrics", "", "/v1/metrics", 200)
	}, 2)

	seen := map[string]bool{}
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"http.server.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No http.server.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			for _, dp := range point.DataPoints {
				if dp.Count <= 0 {
					continue
				}
				route := verifier.GetAttribute(dp.Attributes.ToSlice(), "http.route").AsString()
				seen[route] = true
				verifier.VerifyHttpServerMetricsAttributes(dp.Attributes.ToSlice(), "POST", route, "", "http", "1.1", "http", 200)
			}
			if !seen["/v1/traces"] || !seen["/v1/metrics"] {
				panic("expected http.route for /v1/traces and /v1/metrics, got " + strconv.FormatBool(seen["/v1/traces"]) + "/" + strconv.FormatBool(seen["/v1/metrics"]))
			}
		},
	})
}
