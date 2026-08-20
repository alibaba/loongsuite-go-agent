# OpenTelemetry SDK 配置

`otel`工具除了自动埋点外，还会注入配置代码，在应用启动时会初始化 OpenTelemetry SDK，使用以下环境变量可以改变 OpenTelemetry SDK 的行为。

- `OTEL_SERVICE_NAME`: 为您的应用指定服务名称。
- `OTEL_TRACES_EXPORTER`: 指定链路导出器。支持的值: `none`, `console`, `zipkin`, `otlp`。支持使用逗号分隔指定多个导出器（例如 `console,otlp`）。默认为 `otlp`。
- `OTEL_METRICS_EXPORTER`: 指定指标导出器。支持的值: `none`, `console`, `prometheus`, `otlp`。支持使用逗号分隔指定多个导出器（例如 `console,otlp`）。默认为 `otlp`。
- `OTEL_EXPORTER_OTLP_PROTOCOL`: 指定 OTLP 协议，用于链路和指标。支持的值: `http/protobuf` (默认), `grpc`。
- `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`: 指定用于链路的 OTLP 协议，会覆盖 `OTEL_EXPORTER_OTLP_PROTOCOL` 的设置。支持的值: `http/protobuf` (默认), `grpc`。
- `OTEL_EXPORTER_OTLP_ENDPOINT`: 指定 OTLP 导出器的通用端点。
- `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`: 指定 OTLP 链路导出器的端点。
- `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`: 指定 OTLP 指标导出器的端点。
- `OTEL_EXPORTER_OTLP_HEADERS`: 为所有 OTLP 导出器指定请求头 (例如, `key1=value1,key2=value2`)。
- `OTEL_EXPORTER_PROMETHEUS_PORT`: 当 `OTEL_METRICS_EXPORTER` 设置为 `prometheus` 时，指定 Prometheus 导出器的端口。默认为 `9464`。
- `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE`: 指定指标的聚合时间性偏好（不区分大小写）。支持的值:
  - `cumulative` (默认): 所有指标类型都使用累积时间性
  - `delta`: Counter、Asynchronous Counter 和 Histogram 使用增量时间性；UpDownCounter 和 Asynchronous UpDownCounter 使用累积时间性
  - `lowmemory`: Synchronous Counter 和 Histogram 使用增量时间性；其他类型使用累积时间性（低内存模式）
- `OTEL_TRACES_SAMPLER`: 使用 OpenTelemetry SDK 标准取值指定链路采样器（大小写不敏感）。支持的值：
  - `always_on`、`always_off`
  - `traceidratio`、`parentbased_traceidratio`：采样比率从 `OTEL_TRACES_SAMPLER_ARG` 读取
  - `parentbased_always_on`（规范默认值）、`parentbased_always_off`
  - `jaeger_remote`、`parentbased_jaeger_remote` 和 `xray` 暂不支持，会回退到默认采样器。
- `OTEL_TRACES_SAMPLER_ARG`: 所选采样器的参数。它仅被 `traceidratio` 和 `parentbased_traceidratio` 采样器读取，其余采样器会忽略它。取值为 0.0 到 1.0 之间的浮点数；未设置、无法解析或超出范围时会记录日志并回退为 `1.0`。
- `OTEL_TRACE_SAMPLER`: 早于标准变量引入的 Agent 私有配置，设置后**优先级高于标准变量**。0.0 到 1.0 之间的浮点数会设置一个基于比率的采样器。小于等于 0 的值将永不采样，大于等于 1 的值将始终采样。

  当 `OTEL_TRACE_SAMPLER` 和 `OTEL_TRACES_SAMPLER` 都未设置时，默认是基于父级的采样器，并且始终采样。

  注意两个变量并不等价：`OTEL_TRACE_SAMPLER=0.5` 构建的是**基于父级的**比率采样器，而 `OTEL_TRACES_SAMPLER=traceidratio` 配合 `OTEL_TRACES_SAMPLER_ARG=0.5` 构建的是**非**基于父级的采样器。若要保持原有行为，请使用 `parentbased_traceidratio`。

  **升级提示**：支持标准采样器变量之前的版本会完全忽略 `OTEL_TRACES_SAMPLER`。如果你的平台已经注入了该变量，升级后采样行为会发生变化。显式设置 `OTEL_TRACE_SAMPLER` 可保持当前的采样器不变。
- `OTEL_INSTRUMENTATION_HTTP_EXCLUDE_PATHS`: 指定要从 HTTP 自动埋点中排除的 URL 路径正则表达式（例如 `^/(ping|health|metrics)$`）。路径匹配该正则表达式的请求将不会生成 span。默认不排除任何路径。
