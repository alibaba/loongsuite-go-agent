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

package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	http2 "net/http"
	"os"
	"runtime"
	"strings"

	"github.com/alibaba/loongsuite-go/pkg/core/meter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/ai"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/experimental"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/http"
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/rpc"
	"github.com/alibaba/loongsuite-go/pkg/sdkconfig"
	testaccess "github.com/alibaba/loongsuite-go/pkg/testaccess"
	prometheus_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"

	// The version of the following packages/modules must be fixed
	"go.opentelemetry.io/otel"
	_ "go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
)

// set the following environment variables based on https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables
// your service name: OTEL_SERVICE_NAME
// your otlp endpoint: OTEL_EXPORTER_OTLP_ENDPOINT OTEL_EXPORTER_OTLP_TRACES_ENDPOINT OTEL_EXPORTER_OTLP_METRICS_ENDPOINT OTEL_EXPORTER_OTLP_LOGS_ENDPOINT
// your otlp header: OTEL_EXPORTER_OTLP_HEADERS
const exec_name = "otel"
const report_protocol = "OTEL_EXPORTER_OTLP_PROTOCOL"
const trace_report_protocol = "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"
const metrics_exporter = "OTEL_METRICS_EXPORTER"
const trace_exporter = "OTEL_TRACES_EXPORTER"
const prometheus_exporter_port = "OTEL_EXPORTER_PROMETHEUS_PORT"
const default_prometheus_exporter_port = "9464"
const metrics_temporality_preference = "OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE"

var (
	metricExporters    []metric.Exporter
	spanExporters      []trace.SpanExporter
	traceProvider      *trace.TracerProvider
	metricsProvider    otelmetric.MeterProvider
	spanProcessors     []trace.SpanProcessor
	spanSampler        trace.Sampler
)

func init() {
	if testaccess.IsInTest() {
		trace.GetTestSpans = testaccess.GetTestSpans
		metric.GetTestMetrics = testaccess.GetTestMetrics
		trace.ResetTestSpans = testaccess.ResetTestSpans
	}
	ctx := context.Background()
	// graceful shutdown
	runtime.ExitHook = func() {
		gracefullyShutdown(ctx)
	}
	path, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	// skip when the executable is otel itself
	if strings.HasSuffix(path, exec_name) {
		return
	}
	if err = initOpenTelemetry(ctx); err != nil {
		log.Fatalf("%s: %v", "Failed to initialize opentelemetry resource", err)
	}
}

func newSpanProcessors(ctx context.Context) []trace.SpanProcessor {
	if testaccess.IsInTest() {
		traceExporter := testaccess.GetSpanExporter()
		simpleProcessor := trace.NewSimpleSpanProcessor(traceExporter)
		return []trace.SpanProcessor{simpleProcessor}
	}

	exporterNames := parseExporterNames(os.Getenv(trace_exporter), "otlp")
	var processors []trace.SpanProcessor

	for _, name := range exporterNames {
		if name == "none" {
			continue
		}

		exporter, err := createTraceExporter(ctx, name)
		if err != nil {
			log.Printf("Failed to create trace exporter %s: %v", name, err)
			continue
		}

		spanExporters = append(spanExporters, exporter)

		if name == "console" {
			processors = append(processors, trace.NewSimpleSpanProcessor(exporter))
		} else {
			processor := trace.NewBatchSpanProcessor(exporter)
			processors = append(processors, processor)
		}
	}

	if len(processors) == 0 {
		log.Fatalf("No valid trace exporter configured")
	}

	spanProcessors = processors
	return processors
}

func parseExporterNames(envValue, defaultValue string) []string {
	envValue = strings.TrimSpace(envValue)
	if envValue == "" {
		return []string{defaultValue}
	}

	parts := strings.Split(envValue, ",")
	var names []string
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		return []string{defaultValue}
	}
	return names
}

func createTraceExporter(ctx context.Context, name string) (trace.SpanExporter, error) {
	switch name {
	case "console":
		return stdouttrace.New()
	case "zipkin":
		return zipkin.New("")
	case "otlp":
		if os.Getenv(report_protocol) == "grpc" || os.Getenv(trace_report_protocol) == "grpc" {
			return otlptrace.New(ctx, otlptracegrpc.NewClient())
		}
		return otlptrace.New(ctx, otlptracehttp.NewClient())
	default:
		return nil, fmt.Errorf("unknown trace exporter: %s", name)
	}
}

func getTemporalitySelector() metric.TemporalitySelector {
	pref := strings.ToLower(strings.TrimSpace(os.Getenv(metrics_temporality_preference)))
	
	switch pref {
	case "cumulative":
		return cumulativeTemporalitySelector
	case "delta":
		return deltaTemporalitySelector
	case "lowmemory":
		return lowMemoryTemporalitySelector
	default:
		// Default to cumulative if not set or invalid value
		if pref != "" {
			log.Printf("Warning: Invalid OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE value '%s', using default 'cumulative'", pref)
		}
		return cumulativeTemporalitySelector
	}
}

// cumulativeTemporalitySelector returns Cumulative temporality for all instrument kinds
func cumulativeTemporalitySelector(metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

// deltaTemporalitySelector implements the "delta" preference:
// - Counter, Async Counter, Histogram: Delta
// - UpDownCounter, Async UpDownCounter: Cumulative
// - Gauge: Cumulative
func deltaTemporalitySelector(ik metric.InstrumentKind) metricdata.Temporality {
	switch ik {
	case metric.InstrumentKindCounter,
		metric.InstrumentKindObservableCounter,
		metric.InstrumentKindHistogram:
		return metricdata.DeltaTemporality
	default:
		// UpDownCounter, ObservableUpDownCounter, ObservableGauge
		return metricdata.CumulativeTemporality
	}
}

// lowMemoryTemporalitySelector implements the "lowmemory" preference:
// - Sync Counter, Histogram: Delta
// - Sync UpDownCounter, Async Counter, Async UpDownCounter: Cumulative
// - Gauge: Cumulative
func lowMemoryTemporalitySelector(ik metric.InstrumentKind) metricdata.Temporality {
	switch ik {
	case metric.InstrumentKindCounter,
		metric.InstrumentKindHistogram:
		return metricdata.DeltaTemporality
	default:
		// UpDownCounter, ObservableCounter, ObservableUpDownCounter, ObservableGauge
		return metricdata.CumulativeTemporality
	}
}

func initOpenTelemetry(ctx context.Context) error {
	processors := newSpanProcessors(ctx)
	spanSampler = sdkconfig.NewSpanSampler()

	var options []trace.TracerProviderOption
	if len(processors) > 0 {
		for _, processor := range processors {
			options = append(options, trace.WithSpanProcessor(processor))
		}
	}
	options = append(options, trace.WithSampler(spanSampler))

	traceProvider = trace.NewTracerProvider(options...)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return initMetrics()
}

func initMetrics() error {
	ctx := context.Background()

	if testaccess.IsInTest() {
		metricsProvider = metric.NewMeterProvider(
			metric.WithReader(testaccess.ManualReader),
		)
	} else {
		exporterNames := parseExporterNames(os.Getenv(metrics_exporter), "otlp")
		var readers []metric.Reader

		for _, name := range exporterNames {
			if name == "none" {
				continue
			}

			reader, exporter, err := createMetricReader(ctx, name)
			if err != nil {
				log.Printf("Failed to create metric exporter %s: %v", name, err)
				continue
			}

			if exporter != nil {
				metricExporters = append(metricExporters, exporter)
			}
			readers = append(readers, reader)

			if name == "prometheus" {
				go serveMetrics()
			}
		}

		if len(readers) == 0 {
			metricsProvider = noop.NewMeterProvider()
		} else {
			var options []metric.Option
			for _, reader := range readers {
				options = append(options, metric.WithReader(reader))
			}
			metricsProvider = metric.NewMeterProvider(options...)
		}
	}

	if metricsProvider == nil {
		return errors.New("No MeterProvider is provided")
	}

	otel.SetMeterProvider(metricsProvider)
	m := metricsProvider.Meter("opentelemetry-global-meter")
	meter.SetMeter(m)
	http.InitHttpMetrics(m)
	rpc.InitRpcMetrics(m)
	db.InitDbMetrics(m)
	ai.InitAIMetrics(m)
	experimental.InitNacosExperimentalMetrics(m)
	experimental.InitSentinelExperimentalMetrics(m)
	return otelruntime.Start(otelruntime.WithMeterProvider(metricsProvider))
}

func createMetricReader(ctx context.Context, name string) (metric.Reader, metric.Exporter, error) {
	// Get temporality selector for exporters that support it
	temporalitySelector := getTemporalitySelector()
	
	switch name {
	case "console":
		exporter, err := stdoutmetric.New(stdoutmetric.WithTemporalitySelector(temporalitySelector))
		if err != nil {
			return nil, nil, err
		}
		return metric.NewPeriodicReader(exporter), exporter, nil
	case "prometheus":
		reader, err := prometheus.New()
		if err != nil {
			return nil, nil, err
		}
		return reader, nil, nil
	case "otlp":
		var exporter metric.Exporter
		var err error
		if os.Getenv(report_protocol) == "grpc" || os.Getenv(trace_report_protocol) == "grpc" {
			exporter, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithTemporalitySelector(temporalitySelector))
		} else {
			exporter, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithTemporalitySelector(temporalitySelector))
		}
		if err != nil {
			return nil, nil, err
		}
		return metric.NewPeriodicReader(exporter), exporter, nil
	default:
		return nil, nil, fmt.Errorf("unknown metric exporter: %s", name)
	}
}

func serveMetrics() {
	http2.Handle("/metrics", promhttp.HandlerFor(
		prometheus_client.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	port := os.Getenv(prometheus_exporter_port)
	if port == "" {
		port = default_prometheus_exporter_port
	}
	log.Printf("serving serveMetrics at localhost:%s/metrics", port)
	err := http2.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		fmt.Printf("error serving serveMetrics: %v", err)
		return
	}
}

func gracefullyShutdown(ctx context.Context) {
	if metricsProvider != nil {
		mp, ok := metricsProvider.(*metric.MeterProvider)
		if ok {
			_ = mp.Shutdown(ctx)
		}
	}
	if traceProvider != nil {
		_ = traceProvider.Shutdown(ctx)
	}
	for _, exporter := range spanExporters {
		if exporter != nil {
			_ = exporter.Shutdown(ctx)
		}
	}
	for _, exporter := range metricExporters {
		if exporter != nil {
			_ = exporter.Shutdown(ctx)
		}
	}
	for _, processor := range spanProcessors {
		if processor != nil {
			_ = processor.Shutdown(ctx)
		}
	}
}
