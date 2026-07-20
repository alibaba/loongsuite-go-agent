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

package db

import (
	"context"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"testing"
	"time"
)

func TestDbClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	client, err := newDbClientMetric("test", meter)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "db.client.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestDbMetricAttributesShadower(t *testing.T) {
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.KeyValue{
		Key:   semconv.DBSystemNameKey,
		Value: attribute.StringValue("mysql"),
	}, attribute.KeyValue{
		Key:   "unknown",
		Value: attribute.Value{},
	}, attribute.KeyValue{
		Key:   semconv.DBOperationNameKey,
		Value: attribute.StringValue("Db"),
	}, attribute.KeyValue{
		Key:   semconv.ServerAddressKey,
		Value: attribute.StringValue("abc"),
	}, attribute.KeyValue{
		Key:   semconv.ErrorTypeKey,
		Value: attribute.StringValue("*net.OpError"),
	})
	n, attrs := utils.Shadow(attrs, dbMetricsConv)
	if n != 4 {
		panic("wrong shadow array")
	}
	if attrs[4].Key != "unknown" {
		panic("unknown should be the last attribute")
	}
	foundErrorType := false
	for i := 0; i < n; i++ {
		if attrs[i].Key == semconv.ErrorTypeKey {
			foundErrorType = true
		}
	}
	if !foundErrorType {
		panic("error.type should be kept as a metric attribute")
	}
}

func TestDbClientMetricsRecordsErrorType(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	client, err := newDbClientMetric("test", meter)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{
		{Key: semconv.DBSystemNameKey, Value: attribute.StringValue("redis")},
		{Key: semconv.DBOperationNameKey, Value: attribute.StringValue("get")},
	}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{
		{Key: semconv.ErrorTypeKey, Value: attribute.StringValue("*net.OpError")},
		{Key: "db.query.text", Value: attribute.StringValue("GET k")},
	}, time.Now())

	rm := &metricdata.ResourceMetrics{}
	if err := reader.Collect(ctx, rm); err != nil {
		t.Fatal(err)
	}
	hist, ok := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
	if !ok || len(hist.DataPoints) == 0 {
		t.Fatalf("expected histogram datapoint")
	}
	attrs := hist.DataPoints[0].Attributes
	var found bool
	for _, kv := range attrs.ToSlice() {
		if kv.Key == semconv.ErrorTypeKey {
			found = true
			if kv.Value.AsString() != "*net.OpError" {
				t.Fatalf("unexpected error.type %q", kv.Value.AsString())
			}
		}
		if kv.Key == "db.query.text" {
			t.Fatalf("high-cardinality db.query.text must not be metric attribute")
		}
	}
	if !found {
		t.Fatalf("metric datapoint must include error.type")
	}
}

func TestLazyDbClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitDbMetrics(m)
	client := DbClientMetrics("db.client")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "db.client.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestGlobalDbClientMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	m := mp.Meter("test-meter")
	InitDbMetrics(m)
	client := DbClientMetrics("db.client")
	ctx := context.Background()
	start := time.Now()
	ctx = client.OnBeforeStart(ctx, start)
	ctx = client.OnBeforeEnd(ctx, []attribute.KeyValue{}, start)
	client.OnAfterStart(ctx, start)
	client.OnAfterEnd(ctx, []attribute.KeyValue{}, time.Now())
	rm := &metricdata.ResourceMetrics{}
	reader.Collect(ctx, rm)
	if rm.ScopeMetrics[0].Metrics[0].Name != "db.client.request.duration" {
		panic("wrong metrics name, " + rm.ScopeMetrics[0].Metrics[0].Name)
	}
}

func TestNilMeter(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	_ = metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	_, err := newDbClientMetric("test", nil)
	if err == nil {
		panic(err)
	}
}

func TestDbClientMetricsMissingContext(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	client, err := newDbClientMetric("test", meter)
	if err != nil {
		t.Fatal(err)
	}
	// End without matching OnBeforeEnd should not panic.
	client.OnAfterEnd(context.Background(), nil, time.Now())
}

func TestDbClientMetricsWrongTypeContext(t *testing.T) {
	reader := metric.NewManualReader()
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("my-service"),
		semconv.ServiceVersion("v0.1.0"),
	)
	mp := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(reader))
	meter := mp.Meter("test-meter")
	client, err := newDbClientMetric("test", meter)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), attribute.Key("test"), "not-a-dbMetricContext")
	client.OnAfterEnd(ctx, nil, time.Now())
}
