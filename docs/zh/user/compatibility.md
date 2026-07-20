# 兼容性

`loongsuite-go`确保与当前支持的[Go语言](https://golang.org/doc/devel/release#policy)版本兼容：

> 每个主要的 Go 版本都会被支持，直到有两个更新的主要版本发布。
> 例如，Go 1.5 被支持到 Go 1.7 发布，Go 1.6 被支持到 Go 1.8 发布。

对于不再受上游支持的Go版本，`loongsuite-go`将通过以下方式停止确保与这些版本的兼容性：

- 将发布`loongsuite-go`的次要版本，以增加对新支持的Go版本的支持。
- `loongsuite-go`的下一个次要版本将移除对最旧（现已在上游归档）的Go版本的兼容性测试。此版本以及将来的`loongsuite-go`版本可能包含仅受当前支持的Go版本支持的功能。

该项目在以下系统上进行了测试。

| 操作系统 | Go 版本 | 架构  |
| -------- | ------- | ----- |
| Ubuntu   | 1.24    | amd64 |
| Ubuntu   | 1.23    | amd64 |
| Ubuntu   | 1.24    | 386   |
| Ubuntu   | 1.23    | 386   |
| Ubuntu   | 1.24    | arm64 |
| Ubuntu   | 1.23    | arm64 |
| macOS 13 | 1.24    | amd64 |
| macOS 13 | 1.23    | amd64 |
| macOS    | 1.24    | arm64 |
| macOS    | 1.23    | arm64 |
| Windows  | 1.24    | amd64 |
| Windows  | 1.23    | amd64 |
| Windows  | 1.24    | 386   |
| Windows  | 1.23    | 386   |

虽然该项目应该适用于其他系统，但目前不对这些系统提供兼容性保证。

# OpenTelemetry 兼容性

为了解决因缺少上下文而导致的跟踪中断等问题，我们需要使用此`otel`工具来对 OpenTelemetry（OTel）本身进行埋点。这意味着，如果用户明确添加 OTel 依赖项，这些依赖项的版本必须与`otel`的要求相匹配，否则，该工具将无法正常工作。目前，`otel`与支持的 OTel 版本的映射如下：

| 工具版本 | OTel 版本 | OTel Contrib 版本 |
| -------- | --------- | ----------------- |
| 0.1.0-RC | v1.28.0   | -                 |
| v0.2.0   | v1.30.0   | v0.55.0           |
| v0.3.0   | v1.31.0   | v0.56.0           |
| v0.4.0   | v1.32.0   | v0.57.0           |
| v0.4.1   | v1.32.0   | v0.57.0           |
| v0.5.0   | v1.32.0   | v0.57.0           |
| v0.6.0   | v1.33.0   | v0.58.0           |
| v0.7.0   | v1.33.0   | v0.58.0           |
| v0.8.0   | v1.33.0   | v0.58.0           |
| v0.9.0   | v1.35.0   | v0.60.0           |
| v0.9.1   | v1.35.0   | v0.60.0           |
| v0.9.2   | v1.35.0   | v0.60.0           |
| v0.10.0  | v1.35.0   | v0.60.0           |
| v1.8.1   | v1.40.0   | v0.65.0           |
| v1.8.2   | v1.40.0   | v0.65.0           |
| v1.9.0   | v1.40.0   | v0.65.0           |
| v1.10.0  | v1.40.0   | v0.65.0           |

# 遥测迁移说明

以下变更会影响导出的 span 属性与指标。升级后请同步更新相关看板与告警。

## rueidis：`db.system.name` 重命名为 `redis`

此前 rueidis 埋点使用 `db.system.name="rueidis"`。现已改为 `db.system.name="redis"`，
以符合 OpenTelemetry 语义约定，并与 goredis / redigo 保持一致。

- **影响：** 依赖 `db.system.name="rueidis"` 过滤的时间序列与 span 查询在升级后将不再匹配。
- **迁移：** 将过滤/聚合条件改为 `db.system.name="redis"`。如需区分客户端库，请使用
  instrumentation scope name。

## Elasticsearch：`http.server.request.duration` 替换为 `db.client.request.duration`

Elasticsearch 插件此前错误注册了 `HttpServerMetrics("elasticsearch.client")`，会为
客户端操作导出 `http.server.request.duration`。现已替换为
`DbClientMetrics("nosql.elasticsearch")`。

| 之前（不正确） | 之后 |
| ------------- | ---- |
| ES 插件导出的 `http.server.request.duration` | `db.client.request.duration`，且 `db.system.name=elasticsearch` |

- **不变：** net/http 传输层的 `http.client.request.duration` 仍会为同一请求导出。
- **迁移：** ES API 延迟请使用 `db.client.request.duration{db.system.name="elasticsearch"}`；
  HTTP 传输延迟请使用 `http.client.request.duration`。依赖 ES 作用域下
  `http.server.request.duration` 的看板需更新或移除。
- **不做双写：** 原先在客户端 span 上导出的 `http.server.*` 不符合语义约定，不会在
  弃用窗口内继续保留。
