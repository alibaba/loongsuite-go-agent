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
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
)

const EnvIntervals = "OTEL_METRIC_EXPORT_INTERVALS"
const ResourceIntervalMSAttr = "telemetry.metric.export.interval.ms"
const MaxIntervalMS = int64(1<<63-1) / int64(time.Millisecond)

type WarnFunc func(format string, args ...interface{})

type Exporter struct {
	metric.Exporter
	resource *resource.Resource
}

func ParseIntervals(envValue string, warn WarnFunc) []time.Duration {
	envValue = strings.TrimSpace(envValue)
	if envValue == "" {
		return nil
	}

	seen := make(map[int64]struct{})
	var intervals []time.Duration
	for _, part := range strings.Split(envValue, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		interval, err := ParseInterval(part)
		if err != nil {
			warnf(warn, "Warning: Invalid metric export interval value '%s': %v", part, err)
			continue
		}

		millis := interval.Milliseconds()
		if _, ok := seen[millis]; ok {
			warnf(warn, "Warning: Duplicate metric export interval value '%s' ignored", part)
			continue
		}
		seen[millis] = struct{}{}
		intervals = append(intervals, interval)
	}

	if len(intervals) == 0 {
		warnf(warn, "Warning: No valid %s values found, using SDK metric export interval configuration", EnvIntervals)
	}
	return intervals
}

func ParseInterval(value string) (time.Duration, error) {
	millis, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if millis <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	if millis > MaxIntervalMS {
		return 0, fmt.Errorf("must be less than or equal to %d", MaxIntervalMS)
	}
	return time.Duration(millis) * time.Millisecond, nil
}

func NewExporter(exporter metric.Exporter, interval time.Duration) metric.Exporter {
	return &Exporter{
		Exporter: exporter,
		resource: resource.NewWithAttributes("",
			attribute.Int64(ResourceIntervalMSAttr, interval.Milliseconds()),
		),
	}
}

func (e *Exporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	original := rm.Resource
	merged, err := resource.Merge(original, e.resource)
	if err != nil {
		return fmt.Errorf("merge export interval resource: %w", err)
	}

	rm.Resource = merged
	defer func() {
		rm.Resource = original
	}()
	return e.Exporter.Export(ctx, rm)
}

func warnf(warn WarnFunc, format string, args ...interface{}) {
	if warn != nil {
		warn(format, args...)
	}
}
