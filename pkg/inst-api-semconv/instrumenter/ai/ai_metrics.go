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
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// GenAI metrics instrumentation (Stability: development).
// Spec: https://opentelemetry.io/docs/specs/semconv/gen-ai/gen-ai-metrics/
const (
	gen_ai_client_token_usage        = "gen_ai.client.token.usage"
	gen_ai_client_operation_duration = "gen_ai.client.operation.duration"
)

type AIClientMetric struct {
	key                     attribute.Key
	clientOperationDuration metric.Float64Histogram
	clientTokenUsage        metric.Int64Histogram
}

var _ instrumenter.OperationListener = (*AIClientMetric)(nil)

var mu sync.Mutex

var aiMetricsConv = map[attribute.Key]bool{
	semconv.GenAIOperationNameKey: true,
	semconv.GenAISystemKey:        true,
	semconv.ErrorTypeKey:          true,
	semconv.GenAIRequestModelKey:  true,
	semconv.ServerAddressKey:      true,
	semconv.ServerPortKey:         true,
	semconv.GenAIResponseModelKey: true,
}

var globalMeter metric.Meter

func InitAIMetrics(m metric.Meter) {
	mu.Lock()
	defer mu.Unlock()
	globalMeter = m
}

func AIClientMetrics(key string) *AIClientMetric {
	mu.Lock()
	defer mu.Unlock()
	return &AIClientMetric{key: attribute.Key(key)}
}

// for test only
func newAIClientMatric(key string, meter metric.Meter) (*AIClientMetric, error) {
	m := &AIClientMetric{
		key: attribute.Key(key),
	}
	clientOperationDuration, err := newAIClientOperationDurationMeasures(meter)
	if err != nil {
		return nil, err
	}
	clientTokenUsage, err := newAIClientTokenUsageMeasures(meter)
	m.clientOperationDuration = clientOperationDuration
	m.clientTokenUsage = clientTokenUsage
	return m, nil
}

func newAIClientOperationDurationMeasures(meter metric.Meter) (metric.Float64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Float64Histogram(gen_ai_client_operation_duration,
		metric.WithUnit("s"),
		metric.WithDescription("Duration of chat completion operation."),
		metric.WithExplicitBucketBoundaries(0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.64, 1.28, 2.56, 5.12, 10.24, 20.48, 40.96, 81.92),
	)
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create gen_ai.client.operation.duration histogram, %v", err))
	}
}

func newAIClientTokenUsageMeasures(meter metric.Meter) (metric.Int64Histogram, error) {
	mu.Lock()
	defer mu.Unlock()
	if meter == nil {
		return nil, errors.New("nil meter")
	}
	d, err := meter.Int64Histogram(gen_ai_client_token_usage,
		metric.WithUnit("token"),
		metric.WithDescription("Number of tokens used in prompt and completions."),
		metric.WithExplicitBucketBoundaries(1, 4, 16, 64, 256, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864),
	)
	if err == nil {
		return d, nil
	} else {
		return d, errors.New(fmt.Sprintf("failed to create gen_ai.client.token.usage histogram, %v", err))
	}
}

type aiMatricContext struct {
	startTime       time.Time
	startAttributes []attribute.KeyValue
}

func (a AIClientMetric) OnBeforeStart(parentContext context.Context, startTimestamp time.Time) context.Context {
	return parentContext
}

func (a AIClientMetric) OnBeforeEnd(ctx context.Context, startAttributes []attribute.KeyValue, startTimestamp time.Time) context.Context {
	return context.WithValue(ctx, a.key, aiMatricContext{
		startTime:       startTimestamp,
		startAttributes: startAttributes,
	})
}

func (a AIClientMetric) OnAfterStart(ctx context.Context, endTimestamp time.Time) {
	return
}

func (a AIClientMetric) OnAfterEnd(ctx context.Context, endAttributes []attribute.KeyValue, endTime time.Time) {
	mc := ctx.Value(a.key).(aiMatricContext)
	startTime, startAttributes := mc.startTime, mc.startAttributes
	// end attributes should be shadowed by AttrsShadower
	if a.clientOperationDuration == nil {
		var err error
		// second change to init the metric
		a.clientOperationDuration, err = newAIClientOperationDurationMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create clientOperationDuration, err is %v\n", err)
		}
	}
	if a.clientTokenUsage == nil {
		var err error
		// second change to init the metric
		a.clientTokenUsage, err = newAIClientTokenUsageMeasures(globalMeter)
		if err != nil {
			log.Printf("failed to create clientTokenUsage, err is %v\n", err)
		}
	}
	endAttributes = append(endAttributes, startAttributes...)
	n, metricsAttrs := utils.Shadow(endAttributes, aiMetricsConv)

	// record the client operation duration
	if a.clientOperationDuration != nil {
		a.clientOperationDuration.Record(ctx, endTime.Sub(startTime).Seconds(), metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)))
	}

	var inputTokens, outputTokens attribute.Value
	var hasInputTokens, hasOutputTokens bool
	for _, kv := range endAttributes {
		switch kv.Key {
		case semconv.GenAIUsageInputTokensKey:
			inputTokens = kv.Value
			hasInputTokens = true
		case semconv.GenAIUsageOutputTokensKey:
			outputTokens = kv.Value
			hasOutputTokens = true
		}
		if hasInputTokens && hasOutputTokens {
			break
		}
	}

	// record the client token usage
	if hasInputTokens {
		a.clientTokenUsage.Record(ctx, inputTokens.AsInt64(),
			metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)),
			metric.WithAttributes(semconv.GenAITokenTypeInput))
	}
	if hasOutputTokens {
		a.clientTokenUsage.Record(ctx, outputTokens.AsInt64(),
			metric.WithAttributeSet(attribute.NewSet(metricsAttrs[0:n]...)),
			metric.WithAttributes(semconv.GenAITokenTypeCompletion))
	}
}
