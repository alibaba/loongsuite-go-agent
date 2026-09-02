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
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func setupRouteUnavailableMetricHttp() {
	var err error
	port, err = verifier.GetFreePort()
	if err != nil {
		panic(err)
	}
	err = http.ListenAndServe(":"+strconv.Itoa(port), http.HandlerFunc(helloHandler))
	if err != nil {
		panic(err)
	}
}

func main() {
	go setupRouteUnavailableMetricHttp()
	time.Sleep(1 * time.Second)
	_, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/users/123")
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"http.server.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No http.server.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("http.server.request.duration metrics count is not positive, actually " + strconv.Itoa(int(point.DataPoints[0].Count)))
			}
			attrs := point.DataPoints[0].Attributes.ToSlice()
			verifier.VerifyHttpServerMetricsAttributes(attrs, "GET", verifier.OmitHttpRoute, "", "http", "1.1", "http", 200)
		},
	})
}
