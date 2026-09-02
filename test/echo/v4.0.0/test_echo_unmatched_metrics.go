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
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func setupUnmatchedMetrics() {
	engine := echo.New()
	engine.Use(middleware.Logger())
	engine.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code": 1,
			"msg":  c.Path(),
		})
	})

	engine.Logger.Fatal(engine.Start(":8080"))
}

func main() {
	go setupUnmatchedMetrics()
	time.Sleep(5 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://127.0.0.1:8080/missing")
	if err != nil {
		panic(err)
	}
	_ = resp.Body.Close()

	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"http.server.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No http.server.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("http.server.request.duration metrics count is not positive, actually " + strconv.Itoa(int(point.DataPoints[0].Count)))
			}
			verifier.VerifyHttpServerMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "GET", verifier.OmitHttpRoute, "", "http", "1.1", "http", 404)
		},
	})
}
