# Compatibility

`loongsuite-go` ensures compatibility with the current supported
versions of
the [Go language](https://golang.org/doc/devel/release#policy):

> Each major Go release is supported until there are two newer major releases.
> For example, Go 1.5 was supported until the Go 1.7 release, and Go 1.6 was supported until the Go 1.8 release.

For versions of Go that are no longer supported upstream, `loongsuite-go` will
stop ensuring compatibility with these versions in the following manner:

- A minor release of `loongsuite-go` will be made to add support for the new
supported release of Go.
- The following minor release of `loongsuite-go` will remove compatibility
testing for the oldest (now archived upstream) version of Go. This, and
future, releases of `loongsuite-go` may include features only supported by
the currently supported versions of Go.

This project is tested on the following systems.

| OS       | Go Version | Architecture |
| -------- | ---------- | ------------ |
| Ubuntu   | 1.24       | amd64        |
| Ubuntu   | 1.23       | amd64        |
| Ubuntu   | 1.24       | 386          |
| Ubuntu   | 1.23       | 386          |
| Ubuntu   | 1.24       | arm64        |
| Ubuntu   | 1.23       | arm64        |
| macOS 13 | 1.24       | amd64        |
| macOS 13 | 1.23       | amd64        |
| macOS    | 1.24       | arm64        |
| macOS    | 1.23       | arm64        |
| Windows  | 1.24       | amd64        |
| Windows  | 1.23       | amd64        |
| Windows  | 1.24       | 386          |
| Windows  | 1.23       | 386          |

While this project should work for other systems, no compatibility guarantees
are made for those systems currently.

# OpenTelemetry Compatibility

To address issues such as trace interruption caused by missing context, we need to instrument OpenTelemetry (OTel)
itself with this `otel`. This means that if users explicitly add OTel dependencies, the version of those
dependencies must match the `otel`'s requirements, otherwise, the tool will not function properly. Currently, the
mapping of the `otel` to the supported OTel versions is as follows:

| Tool Version | OTel Version | OTel Contrib Version |
| ------------ | ------------ | -------------------- |
| 0.1.0-RC     | v1.28.0      | -                    |
| v0.2.0       | v1.30.0      | v0.55.0              |
| v0.3.0       | v1.31.0      | v0.56.0              |
| v0.4.0       | v1.32.0      | v0.57.0              |
| v0.4.1       | v1.32.0      | v0.57.0              |
| v0.5.0       | v1.32.0      | v0.57.0              |
| v0.6.0       | v1.33.0      | v0.58.0              |
| v0.7.0       | v1.33.0      | v0.58.0              |
| v0.8.0       | v1.33.0      | v0.58.0              |
| v0.9.0       | v1.35.0      | v0.60.0              |
| v0.9.1       | v1.35.0      | v0.60.0              |
| v0.9.2       | v1.35.0      | v0.60.0              |
| v0.10.0      | v1.35.0      | v0.60.0              |
| v1.8.1       | v1.40.0      | v0.65.0              |
| v1.8.2       | v1.40.0      | v0.65.0              |
| v1.9.0       | v1.40.0      | v0.65.0              |
| v1.10.0      | v1.40.0      | v0.65.0              |

# Telemetry Migration Notes

The following changes affect exported span attributes and metrics. Update dashboards
and alerts accordingly when upgrading.

## rueidis: `db.system.name` renamed to `redis`

Previously, rueidis instrumentation set `db.system.name="rueidis"`. It now emits
`db.system.name="redis"` to match OpenTelemetry semantic conventions and the
goredis / redigo instrumentations.

- **Impact:** Time series and span queries that filter on `db.system.name="rueidis"`
  will stop matching after upgrade.
- **Migration:** Change filters/aggregations to `db.system.name="redis"`. Distinguish
  the client library via instrumentation scope name when needed.

## Elasticsearch: `http.server.request.duration` replaced by `db.client.request.duration`

The Elasticsearch plugin previously registered `HttpServerMetrics("elasticsearch.client")`,
which incorrectly emitted `http.server.request.duration` for client operations. That
listener is replaced by `DbClientMetrics("nosql.elasticsearch")`.

| Before (incorrect) | After |
| ------------------ | ----- |
| `http.server.request.duration` from the ES plugin | `db.client.request.duration` with `db.system.name=elasticsearch` |

- **Unchanged:** Transport-level `http.client.request.duration` from net/http
  instrumentation continues to be emitted for the same requests.
- **Migration:** Prefer `db.client.request.duration{db.system.name="elasticsearch"}`
  for ES API latency. Prefer `http.client.request.duration` for HTTP transport
  latency. Remove dashboards that relied on ES-scoped `http.server.request.duration`.
- **No dual-write:** The previous `http.server.*` series on client spans was a
  semconv mismatch and is not preserved during a deprecation window.
