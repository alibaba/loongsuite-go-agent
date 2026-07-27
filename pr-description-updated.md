# PR Title

`feat(http)!: resolve http.route from route templates and align span names with OTel semconv`

---

## ⚠️ BREAKING CHANGES SUMMARY

This PR introduces **two breaking changes** to align with OpenTelemetry HTTP Semantic Conventions:

### 1. `http.route` attribute behavior (Option A)
- **Before**: Always present (used raw URL path when no template)
- **After**: Only present when framework provides route template; **omitted** otherwise
- **Affected**: Bare handlers, fasthttp, unmatched routes

### 2. HTTP server span names
- **Before**: Route template only (e.g., `/users/{id}`) or method only (e.g., `GET`)
- **After**: `{method} {route}` format (e.g., `GET /users/{id}`) or `{method}` when no route
- **Affected**: **All frameworks** (gin, mux, echo, iris, gorestful, hertz, net/http, Fiber)

See [Migration Guide](#migration--upgrade-impact) below for details.

---

## English (for GitHub PR body)

## Summary

Align server `http.route` resolution and span naming with OTel semconv and maintainer decision on Issue #729 (Option A):

1. **Use low-cardinality route templates** from framework/stdlib router APIs when available (e.g. `/users/123` → `/users/{id}` or `/users/:id`).
2. **Omit `http.route`** when no template exists — do not fall back to `url.path`.
3. **Standardize span names** to `{method} {route}` format across all frameworks per [OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name).

This reduces metric cardinality on RESTful APIs and ensures compliance with [OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-route), which states that `http.route` must not be populated from URI path when the framework does not provide a template.

Template extraction follows [opentelemetry-go-compile-instrumentation](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation): native router APIs only, no heuristic normalization.

## Motivation

Previously, many stacks (notably `net/http`) always used the raw request path as `http.route`. Under RESTful APIs, each resource ID became a separate metric label on `http.server.request.duration`, causing cardinality blow-ups.

Gin / mux / echo already had partial template support via framework hooks, but resolution was inconsistent and still fell back to raw path in the shared layer. Additionally, span names varied across frameworks and didn't follow OTel conventions.

This change unifies Option A resolution in the shared instrumenter, extends template wiring to `net/http`, Fiber, iris, go-restful, and hertz, and standardizes span naming across all frameworks.

## Resolution policy (Option A)

```text
GetHttpRoute() != ""  →  http.route = template, span name = "{method} {route}"
GetHttpRoute() == ""  →  http.route omitted, span name = "{method}"
```

| Case | `http.route` | Span name | Notes |
|------|--------------|-----------|-------|
| RESTful template | `/users/{id}` | `GET /users/{id}` | ServeMux `GET /users/{id}`, gin `:id`, mux `{key}`, etc. |
| Static template | `/query` | `GET /query` | Template equals path; still from router API |
| No template | _(absent)_ | `GET` | Bare `HandlerFunc`, unmatched route, fasthttp, etc. |
| Template preferred | `/users/{id}` | `GET /users/{id}` | Even when actual path is `/users/999` |

Use `url.path` attribute for per-request debugging when `http.route` is absent.

## What changed

### Shared (`pkg/inst-api-semconv/instrumenter/http`)

- Add `RouteFromPattern()` — strip method prefix from Go 1.22+ ServeMux patterns (`GET /users/{id}` → `/users/{id}`).
- Add `ResolveHttpServerRoute()` — returns `GetHttpRoute()` only; no path fallback.
- `HttpServerAttrsExtractor.OnEnd` emits `http.route` only when the resolved route is non-empty.
- `HttpServerSpanNameExtractor.Extract` uses `{method} {route}` format when route available, `{method}` otherwise.
- Remove span-name fallback logic that previously copied `localRootSpan.Name()` into `http.route`.

### `net/http`

- Add `SetServerRouteTemplate(r, route)` — store template keyed by `*http.Request` pointer so framework hooks can inject templates without GLS.
- On `serverHandler.ServeHTTP` **exit**, read `r.Pattern` into `routeTemplate` when not already set.
- `GetHttpRoute` returns `routeTemplate` only; never `url.Path`.
- Update span name to `{method} {route}` when route available (Go 1.22+).

### Framework template injection and span naming

| Framework | Template source | Span name format |
|-----------|-----------------|------------------|
| gin / gin-html | `c.FullPath()` → `SetServerRouteTemplate` | `{method} {route}` (e.g., `GET /user/:name`) |
| gorilla/mux | `route.GetPathTemplate()` → `SetServerRouteTemplate` | `{method} {route}` (e.g., `GET /users/{id}`) |
| echo | `c.Path()` → `SetServerRouteTemplate` | `{method} {route}` (e.g., `POST /api/orders`) |
| iris | `curr.Tmpl().Src` / `curr.Path()` → `SetServerRouteTemplate` | `{method} {route}` (e.g., `GET /products/{id:int}`) |
| go-restful | `SelectedRoutePath()` → `SetServerRouteTemplate` | `{method} {route}` (e.g., `GET /services/{name}`) |
| Fiber v2 / v3 | `c.Route().Path` via `ReleaseCtx` hook | `{method} {route}` or `{method}` |
| hertz | `c.FullPath()` via custom extractor | `{method} {route}` or `{method}` |
| fasthttp | _(none)_ | `{method}` only; `http.route` omitted |

### Span names (⚠️ Breaking Change)

**All frameworks now follow [OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name):**

- **When `http.route` is available**: span name is `{method} {route}`
- **When no route is available**: span name is `{method}` only

**Framework-specific changes:**

| Framework | Before | After | Impact |
|-----------|--------|-------|--------|
| gin / gin-html | `/user/:name` | `GET /user/:name` | ⚠️ **Breaking** |
| gorilla/mux | `/users/{id}` | `GET /users/{id}` | ⚠️ **Breaking** |
| echo | `/api/v1/posts` | `GET /api/v1/posts` | ⚠️ **Breaking** |
| iris | `/products/{id:int}` | `GET /products/{id:int}` | ⚠️ **Breaking** |
| go-restful | `/services/{name}` | `GET /services/{name}` | ⚠️ **Breaking** |
| hertz | `/user/:id` or `GET` | `GET /user/:id` or `GET` | ⚠️ **Breaking** (when route present) |
| net/http ServeMux (Go 1.22+) | `GET` | `GET /users/{id}` | ⚠️ **Breaking** (when pattern available) |
| net/http bare handler | `GET` | `GET` | ✅ **No change** |
| Fiber v2/v3 | `GET` | `GET /api/users` or `GET` | ⚠️ **Breaking** (when route available) |
| fasthttp | `GET` | `GET` | ✅ **No change** |

## Semconv compliance

### `http.route` attribute

[OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-route) requires that `http.route` **must not** be populated from URI path when the framework does not provide a low-cardinality template.

### Span name

[OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name) specifies:

> HTTP server span names **SHOULD be** `{HTTP method}` if there is no (low-cardinality) `http.route` available. If there is a (low-cardinality) `http.route` available, the span name **SHOULD be** `{HTTP method} {http.route}`.

This PR adopts **Option A** as agreed by maintainers on Issue #729 for `http.route`, and additionally aligns **all frameworks** with OTel span naming conventions.

**Breaking changes:** 

1. **`http.route` omission**: Deployments that relied on always-present `http.route` labels (e.g. bare `HandlerFunc`, fasthttp) will no longer see the attribute. Migrate to router-based handlers (Go 1.22+ ServeMux, framework routers) for low-cardinality `http.route`, or use `url.path` for per-request aggregation.

2. **Span name format**: All dashboards, alerts, or trace queries filtering by span name must be updated to use the new `{method} {route}` format, or switch to filtering by `http.route` attribute.

`net/http` route-template integration uses Go `Request.Pattern` (Go 1.22+). Integration test floor for pattern-based `nethttp` tests is **Go 1.23**. The `otel` tool itself still requires **Go 1.24+** to build.

## Migration / upgrade impact

### `http.route` attribute changes

| Scenario | Before | After | Action Required |
|----------|--------|-------|-----------------|
| `net/http` ServeMux `GET /users/{id}` | `http.route=/users/123` | `http.route=/users/{id}` | ✅ **Improved** - lower cardinality |
| gin `/user/:name` | `http.route=/user/:name` | `http.route=/user/:name` | ✅ **No change** |
| Bare `HandlerFunc` | `http.route=/users/123` | `http.route` **absent** | ⚠️ **Update dashboards** - use `url.path` or migrate to router |
| fasthttp | `http.route=/users/123` | `http.route` **absent** | ⚠️ **Update dashboards** - use `url.path` or migrate to router |
| Metrics grouping on raw paths | Per-ID time series | Templated routes collapse; untemplated have no label | ⚠️ **Review metric queries** |

### Span name changes

| Scenario | Before | After | Action Required |
|----------|--------|-------|-----------------|
| gin `GET /user/:name` | Span name: `/user/:name` | Span name: `GET /user/:name` | ⚠️ **Update trace queries** |
| mux `GET /users/{id}` | Span name: `/users/{id}` | Span name: `GET /users/{id}` | ⚠️ **Update trace queries** |
| net/http ServeMux (Go 1.22+) | Span name: `GET` | Span name: `GET /users/{id}` | ⚠️ **Update trace queries** |
| Bare handler | Span name: `GET` | Span name: `GET` | ✅ **No change** |

### Migration checklist

- [ ] **Dashboards**: Update panels filtering by `http.route` to handle missing values
- [ ] **Dashboards**: Update panels filtering by span name to use `{method} {route}` format
- [ ] **Alerts**: Review alert rules that assume `http.route` is always present
- [ ] **Alerts**: Update alert rules filtering by span name
- [ ] **Trace queries**: Update saved queries to use new span name format
- [ ] **Documentation**: Update internal docs referencing span name patterns
- [ ] **Code**: Migrate bare handlers to router-based handlers for `http.route` visibility (optional but recommended)

## Test plan

### Unit tests (`pkg/inst-api-semconv/instrumenter/http`)

- [x] `TestRouteFromPattern` — RESTful, static, empty patterns
- [x] `TestResolveHttpServerRoute` — template shapes, static routes, no-template returns `""`
- [x] `TestHttpServerExtractorEndResolveHttpRouteCases` — `OnEnd` emits template; omits when empty; non-recording span
- [x] `TestHttpServerSpanNameExtractor` — span name with `{method} {route}` format
- [x] `TestHttpServerMetricsOmitsHttpRouteWithoutTemplate` — metrics omit `http.route` without template

### Integration tests

**net/http (Go 1.22+ for pattern tests)**

- [x] `nethttp-route-pattern-test` — RESTful `GET /users/{id}`, span name `GET /users/{id}`
- [x] `nethttp-route-static-pattern-test` — static `GET /query`, span name `GET /query`
- [x] `nethttp-route-cases-test` — RESTful + static + bare handler
- [x] `nethttp-route-metrics-test` / `nethttp-route-static-metrics-test`
- [x] `nethttp-route-unavailable-test` — bare handler, span name `GET`
- [x] `nethttp-route-unavailable-metrics-test` — bare handler metrics omit `http.route`

**Frameworks**

- [x] gin: span name `GET /user/:name` (pattern), `GET /query` (static), `gin-test-html`
- [x] mux / echo / gorestful: span name with method prefix
- [x] gorestful: `gorestful-test-handle-with-filter` — template injection via `HandleWithFilter`
- [x] iris: span name `GET /products/{id:int}`
- [x] Fiber v2/v3: span name with route or method only (`basic-fiberv2-route-*`, `basic-fiberv3-route-*`)
- [x] Fiber v3: `fiberv3-custom-ctx-route-latestdepth` (custom ctx route via UserValue)
- [x] hertz: span name `GET /user/:id` with custom extractor
- [x] hertz: `hertz-090-basic-test-with-regex`
- [x] fasthttp: span name `GET`, `http.route` omitted

### How to run locally

```bash
export PATH="/path/to/go1.25.0/bin:$PATH"
make package-pkg && go build -o otel ./tool/otel
cd pkg/inst-api-semconv && go test ./instrumenter/http/...
cd test && go test -run 'TestHttpRoute|TestGin|TestMux|TestEcho|TestIris|TestFiber|TestHertz' -v
```

## Related

- Issue #729 maintainer decision: **Option A** — template when available, omit when not; no `url.path` fallback.
- [OTel HTTP Semantic Conventions - Span Name](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name)
- Out of scope: `fasthttp/router` `SaveMatchedRoutePath` integration; configurable path-fallback mode.

---

## 中文（评审 / 内部说明）

## ⚠️ 破坏性变更总结

本 PR 引入 **两项破坏性变更**，以符合 OpenTelemetry HTTP 语义规范：

### 1. `http.route` 属性行为（Option A）
- **变更前**：始终存在（无模板时使用原始 URL path）
- **变更后**：仅当框架提供路由模板时存在；否则**省略**
- **影响范围**：裸 Handler、fasthttp、未匹配路由

### 2. HTTP 服务端 span name 格式
- **变更前**：仅路由模板（如 `/users/{id}`）或仅 method（如 `GET`）
- **变更后**：`{method} {route}` 格式（如 `GET /users/{id}`）或无路由时 `{method}`
- **影响范围**：**所有框架**（gin、mux、echo、iris、gorestful、hertz、net/http、Fiber）

详见下方升级影响。

---

## 摘要

按 Issue #729 维护者决议（Option A）统一各 HTTP 框架服务端的 `http.route` 解析和 span 命名策略：

1. **有路由模板时使用低基数模板**（如 `/users/123` → `/users/{id}` 或 `/users/:id`）
2. **无模板时省略 `http.route`**（unset，不是空字符串）——**不回退到 `url.path`**
3. **统一 span name 为 `{method} {route}` 格式**，符合 [OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name)

降低 RESTful API 的指标基数，确保符合 [OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-route)。

模板提取对齐 [opentelemetry-go-compile-instrumentation](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation)（仅用框架/标准库原生 API，不做启发式归一化）。

## 动机

此前 `net/http` 等栈长期把原始请求 path 当作 `http.route`，RESTful 场景下每个资源 ID 都会产生独立指标序列。gin/mux/echo 已有部分模板能力，但共享层仍可能回退到 path；且各框架 span name 格式不统一，不符合 OTel 规范。

本 PR 统一 Option A 解析策略，扩展模板接入到所有框架，并将所有框架的 span name 对齐到 OTel 规范格式。

## 解析策略（Option A）

```text
GetHttpRoute() 非空  →  http.route = 路由模板, span name = "{method} {route}"
GetHttpRoute() 为空  →  省略 http.route, span name = "{method}"
```

| 场景 | `http.route` | Span name | 说明 |
|------|--------------|-----------|------|
| RESTful 模板 | `/users/{id}` | `GET /users/{id}` | ServeMux、gin `:id`、mux `{key}` 等 |
| 静态模板 | `/query` | `GET /query` | 模板与 path 相同 |
| 无模板 | _(不存在)_ | `GET` | 裸 Handler、fasthttp 等 |
| 模板优先 | `/users/{id}` | `GET /users/{id}` | 实际 path 为 `/users/999` |

无 `http.route` 时可用 `url.path` 做明细排查。

## 改动说明

### 公共层（`pkg/inst-api-semconv/instrumenter/http`）

- 新增 `RouteFromPattern()`：解析 Go 1.22+ ServeMux pattern，去掉 method 前缀
- 新增 `ResolveHttpServerRoute()`：仅返回 `GetHttpRoute()`，无 path 回退
- `HttpServerAttrsExtractor.OnEnd`：仅 route 非空时写入 `http.route`
- `HttpServerSpanNameExtractor.Extract`：使用 `{method} {route}` 格式
- 移除原先从 span name（GLS）回退填充 `http.route` 的逻辑

### `net/http`

- 新增 `SetServerRouteTemplate(r, route)`：基于 `*http.Request` 指针注入模板
- 在 `ServeHTTP` **exit** 读取 `r.Pattern` 写入 `routeTemplate`
- `GetHttpRoute` 只返回 `routeTemplate`，永不返回 `url.Path`
- Span name 更新为 `{method} {route}`（Go 1.22+）

### 各框架模板注入和 span 命名

| 框架 | 模板来源 | Span name 格式 |
|------|----------|---------------|
| gin / gin-html | `c.FullPath()` → `SetServerRouteTemplate` | `{method} {route}`（如 `GET /user/:name`）|
| gorilla/mux | route 模板 → `SetServerRouteTemplate` | `{method} {route}`（如 `GET /users/{id}`）|
| echo | `c.Path()` → `SetServerRouteTemplate` | `{method} {route}`（如 `POST /api/orders`）|
| iris | `curr.Tmpl().Src` / `curr.Path()` → `SetServerRouteTemplate` | `{method} {route}`（如 `GET /products/{id:int}`）|
| go-restful | `SelectedRoutePath()` → `SetServerRouteTemplate` | `{method} {route}`（如 `GET /services/{name}`）|
| Fiber v2/v3 | `ReleaseCtx` hook → `c.Route().Path` | `{method} {route}` 或 `{method}` |
| hertz | `c.FullPath()` 通过自定义 extractor | `{method} {route}` 或 `{method}` |
| fasthttp | 无 | `{method}`；`http.route` 省略 |

### Span Name 变更（⚠️ 破坏性）

**所有框架遵循 [OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name)：**

- **有 `http.route` 时**：span name 为 `{method} {route}`
- **无路由时**：span name 为 `{method}`

**框架具体变更：**

| 框架 | 变更前 | 变更后 | 影响 |
|------|--------|-------|------|
| gin / gin-html | `/user/:name` | `GET /user/:name` | ⚠️ **破坏性** |
| gorilla/mux | `/users/{id}` | `GET /users/{id}` | ⚠️ **破坏性** |
| echo | `/api/v1/posts` | `GET /api/v1/posts` | ⚠️ **破坏性** |
| iris | `/products/{id:int}` | `GET /products/{id:int}` | ⚠️ **破坏性** |
| go-restful | `/services/{name}` | `GET /services/{name}` | ⚠️ **破坏性** |
| hertz | `/user/:id` 或 `GET` | `GET /user/:id` 或 `GET` | ⚠️ **破坏性**（有路由时）|
| net/http ServeMux (Go 1.22+) | `GET` | `GET /users/{id}` | ⚠️ **破坏性**（有 pattern 时）|
| 裸 Handler | `GET` | `GET` | ✅ **无变化** |
| Fiber v2/v3 | `GET` | `GET /api/users` 或 `GET` | ⚠️ **破坏性**（有路由时）|
| fasthttp | `GET` | `GET` | ✅ **无变化** |

## 语义合规（semconv）

### `http.route` 属性

[OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-route) 规定：框架无法提供低基数模板时，**不得**用 URI path 填充 `http.route`。

### Span name

[OTel HTTP semconv](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name) 规定：

> HTTP 服务端 span name **应该**为 `{HTTP method}`（无低基数 `http.route` 时）。若有低基数 `http.route`，span name **应该**为 `{HTTP method} {http.route}`。

本 PR 采用 Issue #729 维护者确认的 **Option A**（针对 `http.route`），同时将**所有框架**的 span name 对齐到 OTel 规范。

**破坏性变更：**

1. **`http.route` 省略**：依赖「`http.route` label 永远存在」的看板/告警（裸 Handler、fasthttp 等）将不再看到该属性。请迁移到带路由器模板的 handler（Go 1.22+ ServeMux 或各 Web 框架），或使用 `url.path` 聚合。

2. **Span name 格式**：所有按 span name 过滤的看板、告警、trace 查询都需更新为新的 `{method} {route}` 格式，或改用 `http.route` 属性过滤。

`net/http` 路由模板依赖 **Go 1.22+** 的 `Request.Pattern`；带 pattern 的 `nethttp` 集成测试最低 **Go 1.23**。`otel` 工具编译仍为 **Go 1.24+**。

## 升级影响

### `http.route` 属性变更

| 场景 | 变更前 | 变更后 | 需要行动 |
|------|--------|-------|---------|
| `net/http` ServeMux `GET /users/{id}` | `http.route=/users/123` | `http.route=/users/{id}` | ✅ **改进** - 基数降低 |
| gin `/user/:name` | `http.route=/user/:name` | `http.route=/user/:name` | ✅ **无变化** |
| 裸 `HandlerFunc` | `http.route=/users/123` | `http.route` **不存在** | ⚠️ **更新看板** - 改用 `url.path` 或迁移到路由器 |
| fasthttp | `http.route=/users/123` | `http.route` **不存在** | ⚠️ **更新看板** - 改用 `url.path` 或迁移到路由器 |
| 按 raw path 聚合的指标 | 每个 id 一条序列 | 有模板的路由合并；无模板则无 `http.route` label | ⚠️ **检查指标查询** |

### Span Name 变更

| 场景 | 变更前 | 变更后 | 需要行动 |
|------|--------|-------|---------|
| gin `GET /user/:name` | Span name: `/user/:name` | Span name: `GET /user/:name` | ⚠️ **更新 trace 查询** |
| mux `GET /users/{id}` | Span name: `/users/{id}` | Span name: `GET /users/{id}` | ⚠️ **更新 trace 查询** |
| net/http ServeMux (Go 1.22+) | Span name: `GET` | Span name: `GET /users/{id}` | ⚠️ **更新 trace 查询** |
| 裸 handler | Span name: `GET` | Span name: `GET` | ✅ **无变化** |

### 迁移清单

- [ ] **看板**：更新过滤 `http.route` 的面板，处理缺失值
- [ ] **看板**：更新过滤 span name 的面板，使用 `{method} {route}` 格式
- [ ] **告警**：检查假设 `http.route` 始终存在的告警规则
- [ ] **告警**：更新过滤 span name 的告警规则
- [ ] **Trace 查询**：更新使用 span name 的查询
- [ ] **文档**：更新引用 span name 模式的内部文档
- [ ] **代码**：将裸 Handler 迁移到路由器（可选，推荐）

## 测试

详见英文版 Test Plan。

## 相关说明

- Issue #729：**Option A** 决议 — 有模板用模板，无模板省略，不做 path 回退
- [OTel HTTP 语义规范 - Span Name](https://opentelemetry.io/docs/specs/semconv/http/http-spans/#name)
- 不在本次范围：`fasthttp/router` 集成、可配置回退模式

---

## Suggested commit message

```
feat(http)!: resolve http.route from route templates and align span names with OTel semconv

BREAKING CHANGE: http.route attribute is now omitted when no route template
is available (bare handlers, fasthttp). Migrate to router-based handlers for
low-cardinality http.route, or use url.path attribute for per-request analysis.

BREAKING CHANGE: HTTP server span names now follow OTel semconv format
"{method} {route}" (e.g., "GET /users/{id}") across all frameworks (gin, mux,
echo, iris, gorestful, hertz, net/http, Fiber). Update dashboards and trace
queries that filter by span name.

Implement Option A from Issue #729: emit http.route only when framework/stdlib
provides route template. Unify resolution in ResolveHttpServerRoute and wire
templates for net/http, Fiber, iris, gorestful, and hertz.
```
