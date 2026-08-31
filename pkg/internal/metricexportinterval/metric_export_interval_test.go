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

package metricexportinterval

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestParseInterval(t *testing.T) {
	got, err := ParseInterval("1000")
	if err != nil {
		t.Fatalf("ParseInterval returned error: %v", err)
	}
	if got != time.Second {
		t.Fatalf("ParseInterval = %v, want %v", got, time.Second)
	}
}

func TestParseIntervalInvalid(t *testing.T) {
	for _, value := range []string{"0", "-1", "abc", fmt.Sprintf("%d", MaxIntervalMS+1)} {
		t.Run(value, func(t *testing.T) {
			if _, err := ParseInterval(value); err == nil {
				t.Fatalf("ParseInterval(%q) succeeded, want error", value)
			}
		})
	}
}

func TestParseIntervalsSkipsEmpty(t *testing.T) {
	got := ParseIntervals("1000,,60000", nil)
	want := []time.Duration{time.Second, time.Minute}
	if len(got) != len(want) {
		t.Fatalf("len(ParseIntervals) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseIntervals[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestParseIntervalsInvalidFallback(t *testing.T) {
	var warnings []string
	got := ParseIntervals("bad,also-bad", func(format string, args ...interface{}) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	})
	if len(got) != 0 {
		t.Fatalf("ParseIntervals returned %v, want no valid intervals", got)
	}
	if len(warnings) != 3 {
		t.Fatalf("warning count = %d, want 3: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[2], "No valid OTEL_METRIC_EXPORT_INTERVALS values found") {
		t.Fatalf("fallback warning = %q", warnings[2])
	}
}

func TestParseIntervalsDeduplicates(t *testing.T) {
	var warnings []string
	got := ParseIntervals("1000,1000,60000", func(format string, args ...interface{}) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	})
	want := []time.Duration{time.Second, time.Minute}
	if len(got) != len(want) {
		t.Fatalf("len(ParseIntervals) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseIntervals[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "Duplicate metric export interval") {
		t.Fatalf("warnings = %v, want duplicate warning", warnings)
	}
}

func TestExporterAddsResourceAttribute(t *testing.T) {
	inner := &recordingExporter{}
	exporter := NewExporter(inner, time.Second)
	rm := &metricdata.ResourceMetrics{}

	if err := exporter.Export(context.Background(), rm); err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	if inner.resource == nil {
		t.Fatal("inner exporter did not receive resource")
	}
	value, ok := inner.resource.Set().Value(ResourceIntervalMSAttr)
	if !ok {
		t.Fatalf("resource missing %s", ResourceIntervalMSAttr)
	}
	if got := value.AsInt64(); got != 1000 {
		t.Fatalf("%s = %d, want 1000", ResourceIntervalMSAttr, got)
	}
	if rm.Resource != nil {
		t.Fatalf("ResourceMetrics resource was not restored")
	}
}

func TestExporterReturnsMergeError(t *testing.T) {
	exporter := &Exporter{
		Exporter: &recordingExporter{},
		resource: resource.NewWithAttributes("https://example.com/schema-b",
			attribute.Int64(ResourceIntervalMSAttr, 1000),
		),
	}
	rm := &metricdata.ResourceMetrics{
		Resource: resource.NewWithAttributes("https://example.com/schema-a",
			attribute.String("service.name", "test"),
		),
	}

	err := exporter.Export(context.Background(), rm)
	if err == nil {
		t.Fatal("Export succeeded, want resource merge error")
	}
	if !strings.Contains(err.Error(), "merge export interval resource") {
		t.Fatalf("Export error = %v", err)
	}
}

type recordingExporter struct {
	resource *resource.Resource
}

func (e *recordingExporter) Temporality(metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (e *recordingExporter) Aggregation(metric.InstrumentKind) metric.Aggregation {
	return metric.DefaultAggregationSelector(metric.InstrumentKindCounter)
}

func (e *recordingExporter) Export(_ context.Context, rm *metricdata.ResourceMetrics) error {
	if rm.Resource != nil {
		e.resource = rm.Resource
	}
	return nil
}

func (e *recordingExporter) ForceFlush(context.Context) error {
	return nil
}

func (e *recordingExporter) Shutdown(context.Context) error {
	return nil
}
