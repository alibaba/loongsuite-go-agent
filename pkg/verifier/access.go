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

package verifier

import (
	"context"
	"github.com/mohae/deepcopy"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// In memory span exporter
var spanExporter = tracetest.NewInMemoryExporter()

var ManualReader = metric.NewManualReader()

func GetSpanExporter() trace.SpanExporter {
	return spanExporter
}

func GetTestSpans() *tracetest.SpanStubs {
	spans := spanExporter.GetSpans()
	return &spans
}

func ResetTestSpans() {
	spanExporter.Reset()
}

func GetTestMetrics() metricdata.ResourceMetrics {
	var tmp, result metricdata.ResourceMetrics
	_ = ManualReader.Collect(context.Background(), &tmp)
	result = deepcopy.Copy(tmp).(metricdata.ResourceMetrics)
	// The deepcopy can not copy the attributes
	// so we just copy the data again to retain the attributes
	for i, s := range tmp.ScopeMetrics {
		for j, m := range s.Metrics {
			result.ScopeMetrics[i].Metrics[j].Data = m.Data
		}
	}
	return result
}
