# OpenTelemetry SDK Configuration

In addition to automatic instrumentation, the `otel` tool injects configuration code to initialize the OpenTelemetry SDK when the application starts. The following environment variables can be used to change the behavior of the OpenTelemetry SDK.

- `OTEL_SERVICE_NAME`: Specifies the service name for your application.
- `OTEL_TRACES_EXPORTER`: Specifies the trace exporter. Supported values: `none`, `console`, `zipkin`, `otlp`. Multiple exporters can be specified using comma-separated values (e.g., `console,otlp`). The default is `otlp`.
- `OTEL_METRICS_EXPORTER`: Specifies the metrics exporter. Supported values: `none`, `console`, `prometheus`, `otlp`. Multiple exporters can be specified using comma-separated values (e.g., `console,otlp`). The default is `otlp`.
- `OTEL_EXPORTER_OTLP_PROTOCOL`: Specifies the OTLP protocol for both traces and metrics. Supported values: `http/protobuf` (default), `grpc`.
- `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`: Specifies the OTLP protocol for traces, overriding `OTEL_EXPORTER_OTLP_PROTOCOL`. Supported values: `http/protobuf` (default), `grpc`.
- `OTEL_EXPORTER_OTLP_ENDPOINT`: Specifies the common endpoint for OTLP exporters.
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: Specifies the endpoint for OTLP trace exporter.
- `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`: Specifies the endpoint for OTLP metrics exporter.
- `OTEL_EXPORTER_OTLP_HEADERS`: Specifies headers for all OTLP exporters (e.g., `key1=value1,key2=value2`).
- `OTEL_EXPORTER_PROMETHEUS_PORT`: Specifies the port for the Prometheus exporter when `OTEL_METRICS_EXPORTER` is set to `prometheus`. Defaults to `9464`.
- `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE`: Specifies the aggregation temporality preference for metrics (case-insensitive). Supported values:
  - `cumulative` (default): All instrument kinds use Cumulative temporality
  - `delta`: Counter, Asynchronous Counter, and Histogram use Delta temporality; UpDownCounter and Asynchronous UpDownCounter use Cumulative temporality
  - `lowmemory`: Synchronous Counter and Histogram use Delta temporality; other types use Cumulative temporality (low memory mode)
- `OTEL_TRACES_SAMPLER`: Specifies the trace sampler using the standard OpenTelemetry SDK values (case-insensitive). Supported values:
  - `always_on`, `always_off`
  - `traceidratio`, `parentbased_traceidratio`: the ratio is read from `OTEL_TRACES_SAMPLER_ARG`
  - `parentbased_always_on` (specification default), `parentbased_always_off`
  - `jaeger_remote`, `parentbased_jaeger_remote` and `xray` are not supported and fall back to the default sampler.
- `OTEL_TRACES_SAMPLER_ARG`: Argument for the selected sampler. It is only read by the `traceidratio` and `parentbased_traceidratio` samplers and ignored by the others. It is a floating-point number between 0.0 and 1.0; a missing, unparsable or out-of-range value is logged and falls back to `1.0`.
- `OTEL_TRACE_SAMPLER`: Agent-specific shorthand that predates the standard variables and **takes precedence over them** when set. A floating-point number between 0.0 and 1.0 sets a ratio-based sampler. Values <= 0 will never sample, and values >= 1 will always sample.

  When neither `OTEL_TRACE_SAMPLER` nor `OTEL_TRACES_SAMPLER` is set, the default is a parent-based sampler that always samples.

  Note that the two variables are not interchangeable: `OTEL_TRACE_SAMPLER=0.5` builds a **parent-based** ratio sampler, whereas `OTEL_TRACES_SAMPLER=traceidratio` with `OTEL_TRACES_SAMPLER_ARG=0.5` builds a **non** parent-based one. Use `parentbased_traceidratio` to keep the previous behaviour.

  **Upgrading**: releases before standard sampler support ignored `OTEL_TRACES_SAMPLER` entirely. If your platform already injects it, sampling will change once you upgrade. Set `OTEL_TRACE_SAMPLER` explicitly to keep the sampler you have today.
- `OTEL_INSTRUMENTATION_HTTP_EXCLUDE_PATHS`: Specifies a regular expression pattern to exclude URL paths from HTTP auto-instrumentation (e.g., `^/(ping|health|metrics)$`). Requests whose paths match the pattern will not generate spans. By default, no paths are excluded.
