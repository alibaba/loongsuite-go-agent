// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      Db://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"log"
	"sync"
	"time"
)

const db_client_request_duration = "db.client.request.duration"

type DbClientMetric struct {
	key                   attribute.Key
	clientRequestDuration metric.Float64Histogram
}

var mu sync.Mutex

var dbMetricsConv = map[attribute.Key]bool{
	semconv.DBSystemNameKey:    true,
	semconv.DBOperationNameKey: true,
	semconv.ServerAddressKey:   true,
	semconv.DBNamespaceKey:     true,
	semconv.ErrorTypeKey:       true,
}

var globalMeter metric.Meter

// InitDbMetrics so we need to make sure the otel_setup is executed before all the init() function
// related to issue Dbs://github.com/alibaba/loongsuite-go/issues/48
func InitDbMetrics(m metric.Meter) {
	mu.Lock()
	defer mu.Unlock()
	globalMeter = m
}

func DbClientMetrics(key string) *DbClientMetric {
	mu.Lock()
	defer mu.Unlock()
	return &DbClientMetric{key: attribute.Key(key)}
}

// for test only
func newDbClientMetric(key string, meter metric.Meter) (*DbClientMetric, error) {
	m := &DbClientMetric{
		key: attribute.Key(key),
	}
	d, err := newDbClientRequestDurationMeasures(meter)
	if err != nil {
		return nil, err
	}
	m.clientRequestDuration = d
	return m, nil
}

func newDbClientRequestDurationMeasures(meter metric.Meter) (metric.Float64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Float64Histogram(db_client_request_duration,
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of Db client requests."))
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create Db.client.request.duration histogram, %v", err))
	}
}

type dbMetricContext struct {
	startTime       time.Time
	startAttributes []attribute.KeyValue
}

func (h DbClientMetric) OnBeforeStart(parentContext context.Context, startTime time.Time) context.Context {
	return parentContext
}

func (h DbClientMetric) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTime time.Time) context.Context {
	return context.WithValue(ctx, h.key, dbMetricContext{
		startTime:       startTime,
		startAttributes: startAttributes,
	})
}

func (h DbClientMetric) OnAfterStart(context context.Context, endTime time.Time) {
	return
}

func (h DbClientMetric) OnAfterEnd(context context.Context, endAttributes []attribute.KeyValue, endTime time.Time) {
	mc := context.Value(h.key).(dbMetricContext)
	startTime, startAttributes := mc.startTime, mc.startAttributes
	// end attributes should be shadowed by AttrsShadower
	if h.clientRequestDuration == nil {
		var err error
		// second change to init the metric
		h.clientRequestDuration, err = newDbClientRequestDurationMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create clientRequestDuration, err is %v\n", err)
		}
	}
	endAttributes = append(endAttributes, startAttributes...)
	n, metricsAttrs := utils.Shadow(endAttributes, dbMetricsConv)
	if h.clientRequestDuration != nil {
		h.clientRequestDuration.Record(context, float64(endTime.Sub(startTime).Milliseconds()), metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)))
	}
}
