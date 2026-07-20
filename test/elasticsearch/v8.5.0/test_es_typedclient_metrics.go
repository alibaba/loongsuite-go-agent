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
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/elastic/go-elasticsearch/v8"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var (
	client *elasticsearch.TypedClient
	url    = "http://127.0.0.1:" + os.Getenv("OTEL_ES_PORT")
)

func main() {
	var err error
	client, err = elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{url},
		Password:  "123456",
		Username:  "elastic",
	})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	// creating an index
	_, err = client.Indices.Create("my_index").Do(ctx)
	if err != nil {
		log.Printf("failed to create index %v\n", err)
	}
	time.Sleep(3 * time.Second)
	port, err := strconv.Atoi(os.Getenv("OTEL_ES_PORT"))
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"db.client.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No db.client.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if len(point.DataPoints) == 0 {
				panic("db.client.request.duration has no datapoints")
			}
			found := false
			for _, dp := range point.DataPoints {
				if dp.Count <= 0 {
					continue
				}
				attrs := dp.Attributes.ToSlice()
				if verifier.GetAttribute(attrs, "db.system.name").AsString() == "elasticsearch" &&
					verifier.GetAttribute(attrs, "db.operation.name").AsString() == "put" &&
					strings.Contains(verifier.GetAttribute(attrs, "server.address").AsString(), "127.0.0.1") {
					found = true
					break
				}
			}
			if !found {
				panic("db.client.request.duration missing elasticsearch put datapoint")
			}
		},
		// net/http client instrumentation still records transport-level latency.
		"http.client.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No http.client.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if len(point.DataPoints) == 0 {
				panic("http.client.request.duration has no datapoints")
			}
			found := false
			for _, dp := range point.DataPoints {
				if dp.Count <= 0 {
					continue
				}
				attrs := dp.Attributes.ToSlice()
				if verifier.GetAttribute(attrs, "http.request.method").AsString() == "PUT" {
					verifier.VerifyHttpClientMetricsAttributes(attrs, "PUT", "", "", "http", "1.1", port, 200)
					found = true
					break
				}
			}
			if !found {
				panic("http.client.request.duration missing PUT datapoint")
			}
		},
	})
}
