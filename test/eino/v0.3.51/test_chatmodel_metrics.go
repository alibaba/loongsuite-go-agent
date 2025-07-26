// Copyright (c) 2025 Alibaba Group Holding Ltd.
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
	"fmt"
	"strconv"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/cloudwego/eino/schema"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func main() {
	ctx := context.Background()
	cm, _ := NewMockOpenAIChatModelForInvoke(ctx)
	_, err := cm.Generate(ctx, []*schema.Message{schema.UserMessage("Hello")})
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"gen_ai.client.operation.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No gen_ai.client.operation.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if point.DataPoints[0].Count <= 0 {
				panic("gen_ai.client.operation.duration metrics count is not positive, actually " + strconv.Itoa(int(point.DataPoints[0].Count)))
			}
			verifier.VerifyGenAIOperationDurationMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "chat", "eino", "mock-chat", "mock-chat")
		},
		"gen_ai.client.token.usage": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No gen_ai.client.token.usage metrics received!")
			}
			fmt.Println(mrs.ScopeMetrics[0].Metrics[0])
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[int64])
			if point.DataPoints[0].Count <= 0 || point.DataPoints[1].Count <= 0 {
				panic("gen_ai.client.token.usage metrics count is not positive")
			}
			verifier.VerifyGenAITokenUsageMetricsAttributes(point.DataPoints[0].Attributes.ToSlice(), "chat", "eino", "mock-chat", "mock-chat", "input")
			verifier.VerifyGenAITokenUsageMetricsAttributes(point.DataPoints[1].Attributes.ToSlice(), "chat", "eino", "mock-chat", "mock-chat", "output")
		},
	})
}
