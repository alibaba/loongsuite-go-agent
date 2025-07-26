// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ai

import (
	"context"
	"testing"
	"time"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func TestAIClientMetric(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("genai-service"),
		semconv.ServiceVersion("v1.0.0"),
	)
	mp := metric.NewMeterProvider(metric.WithReader(reader), metric.WithResource(res))
	m := mp.Meter("test-meter")
	client, err := newAIClientMatric("test", m)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, time.Now())
	client.OnAfterEnd(ctx, []attribute.KeyValue{
		semconv.GenAISystemKey.String("openai"),
		semconv.GenAIOperationNameKey.String("chat"),
		semconv.GenAIUsageInputTokens(123),
		semconv.GenAIUsageOutputTokens(456),
	}, time.Now())

	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)

	if len(rm.ScopeMetrics) <= 0 || len(rm.ScopeMetrics[0].Metrics) <= 0 {
		panic("no metrics collected")
	}

	names := map[string]bool{}
	for _, m := range rm.ScopeMetrics[0].Metrics {
		names[m.Name] = true
	}
	if !names["gen_ai.client.operation.duration"] || !names["gen_ai.client.token.usage"] {
		panic("required metrics not found")
	}
}

func TestLazyAIClientMetric(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("genai-service"),
		semconv.ServiceVersion("v1.0.0"),
	)
	mp := metric.NewMeterProvider(metric.WithReader(reader), metric.WithResource(res))
	m := mp.Meter("test-meter")
	InitAIMetrics(m)
	client := AIClientMetrics("genai.metric")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{
		semconv.GenAIUsageInputTokens(111),
		semconv.GenAIUsageOutputTokens(222),
	}, start)
	client.OnAfterStart(ctx, time.Now())
	client.OnAfterEnd(ctx, []attribute.KeyValue{
		semconv.GenAISystemKey.String("openai"),
		semconv.GenAIOperationNameKey.String("chat"),
	}, time.Now())

	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)

	if len(rm.ScopeMetrics) <= 0 || len(rm.ScopeMetrics[0].Metrics) <= 0 {
		panic("no metrics collected")
	}

	names := map[string]bool{}
	for _, m := range rm.ScopeMetrics[0].Metrics {
		names[m.Name] = true
	}
	if !names["gen_ai.client.operation.duration"] || !names["gen_ai.client.token.usage"] {
		panic("required metrics not found")
	}
}

func TestAINilMeter(t *testing.T) {
	_, err := newAIClientMatric("test", nil)
	if err == nil {
		panic("expected error on nil meter")
	}
}

func TestAIAttrShadower(t *testing.T) {
	attrs := []attribute.KeyValue{
		semconv.GenAISystemKey.String("openai"),
		semconv.GenAIOperationNameKey.String("chat"),
		semconv.GenAIRequestModel("gpt-4"),
		attribute.String("other", "value"),
	}
	n, shadowed := utils.Shadow(attrs, aiMetricsConv)
	if n != 3 {
		panic("expected 3 valid metric attributes")
	}
	if shadowed[n].Key != "other" {
		panic("unexpected attribute shadowing order")
	}
}
