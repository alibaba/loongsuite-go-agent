# HTTP 请求头和 Body 采集方案

状态：已在 `net/http`、`fasthttp`、Fiber v2 和 Fiber v3 探针实现。

本文记录两个 opt-in 的 HTTP 探针增强功能：

- 将配置的请求头采集到一个 HTTP span attribute；
- 当请求体/响应体是 text 或 JSON，且长度不超过 1 KiB 时，将 body 内容采集到 HTTP
  span attribute。

该功能必须默认关闭。请求头和 body 经常包含凭证、token、用户标识或业务数据，不能改变
当前默认采集行为。

## 实现位置

共享的采集支撑代码位于
`pkg/inst-api-semconv/instrumenter/http/http_capture.go`。不同框架的 hook 接入位于：

- `pkg/rules/http`：`net/http`；
- `pkg/rules/fasthttp`：`fasthttp`；
- `pkg/rules/fiberv2`：Fiber v2；
- `pkg/rules/fiberv3`：Fiber v3。

基于 `net/http` 的框架，例如 Gin、Echo、chi、Gorilla/mux，会通过已有的
`net/http` 探针路径覆盖。Fiber 通过自身基于 fasthttp 的专用 rule 覆盖。

## 环境变量

| 环境变量 | 默认值 | 含义 |
| --- | --- | --- |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS` | 空 | 逗号分隔的请求头 allow-list，匹配的请求头会采集到一个 attribute。 |
| `LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS` | `false` | LoongSuite 私有 boolean 开关。设置为 `true` 时，采集未显式 allow-list 的非敏感请求头。 |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED` | `false` | 设置为 `true` 时，采集符合条件的请求体和响应体。 |

示例：

```bash
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true
```

请求头变量作用于已支持 HTTP 探针的 client request span 和 server request span。响应头
采集不包含在本方案范围内。OTel 标准请求头变量使用 allow-list 语义，header 名称会在
去除首尾空格后按大小写不敏感方式匹配。

如果设置 `LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=true`，还会采集未出现在 allow-list
中的请求头，但默认跳过常见敏感 header：`Authorization`、`Cookie`、`Set-Cookie`、
`Proxy-Authorization`、`X-Api-Key` 和 `X-Access-Token`。敏感 header 只有在显式写入
`OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS` 时才会被采集。

## Attribute 命名

请求头统一写入一个项目私有 attribute：

```text
http.request.headers
```

attribute value 是 JSON 字符串。Header 名称统一转成小写，value 保留为字符串数组，例如：

```text
http.request.headers = {"content-type":["application/json"],"x-request-id":["abc"]}
```

这里有意不使用 OpenTelemetry 的 `http.request.header.<name>` 约定，避免开启功能后为一次
请求扩展出很多不同的 attribute 名。

序列化后的 header JSON 超过 `4096` bytes 时，会省略该 attribute。

Body 内容目前没有稳定的 OpenTelemetry semantic convention attribute，因此使用项目私有
attribute：

```text
http.request.body.content
http.response.body.content
```

只记录 UTF-8 字符串，不把二进制内容编码后塞进 span attribute。

## Body 采集条件

只有同时满足以下条件时才采集 body：

- `OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true`；
- content type 是 text 或 JSON：
  - `text/*`；
  - `application/json`；
  - `application/*+json`；
- body 长度已知且不超过 `1024` bytes，或者探针能在不阻塞业务的情况下观察到写出的
  body；
- `Content-Encoding` 为空或 `identity`；
- 采集到的字节是合法 UTF-8。

对于 `net/http` request body 和 HTTP client response body，第一版不建议在 hook 路径里
读取未知长度 stream。读取未知长度的 streaming body 可能阻塞业务、延迟 span 结束，或者
改变网络时序。如果 `ContentLength < 0`，建议跳过采集，除非后续实现引入安全的 tee
方案。

对于 HTTP server response body，可以在 `ResponseWriter` wrapper 中采集，因为 wrapper
是在业务写响应时观察字节。最多缓冲 `1025` bytes；只有最终观察到的 body 长度不超过
`1024` 时才写入 attribute。

对于 `fasthttp` 和 Fiber，采集使用 `fasthttp.Request` / `fasthttp.Response` 中已经持有
的 body bytes；超长、压缩、二进制和非 UTF-8 body 都会跳过。

## 实现说明

### 配置模块

`net/http` 保留 HTTP 局部配置模块：

```text
pkg/rules/http/capture_config.go
```

建议的内部接口：

```go
type httpCaptureConfig struct {
	captureRequestHeaders bool
	captureAllHeaders     bool
	captureBody          bool
	maxBodyBytes         int64
	maxHeadersBytes      int
	requestHeaderNames   map[string]struct{}
}
```

环境变量在 package 初始化时解析一次。`OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS`
使用 allow-list 语义。LoongSuite 私有的 capture-all 开关是可选能力，默认关闭。开启
capture-all 时，除非敏感 header 被显式 allow-list，否则默认跳过。最终将选中的 request
header 序列化为 JSON 并写入一个 attribute；序列化结果超过 4096 bytes 时省略该
attribute。

### 数据结构

扩展 `pkg/rules/http/net_http_data_type.go` 中的 HTTP request/response 数据结构：

```go
type netHttpRequest struct {
	// existing fields...
	requestHeaders string
	requestBody    string
}

type netHttpResponse struct {
	// existing fields...
	responseBody string
}
```

不要在结构体里长期保存原始 `[]byte`。转成 string 后保存即可，避免保留大 buffer，也让
attribute extractor 更简单。

### Attribute Extractor

`net/http` 保留 HTTP 专用 extractor：

```text
pkg/rules/http/capture_attrs_extractor.go
```

它实现：

```go
instrumenter.AttributesExtractor[*netHttpRequest, *netHttpResponse]
```

行为：

- `OnStart` 追加 `http.request.headers` 和 `http.request.body.content`；
- `OnEnd` 追加 `http.response.body.content`；
- 未开启配置或没有采集值时保持 no-op。

然后在 `pkg/rules/http/net_http_otel_instrumenter.go` 的 client/server builder 中，把这个
extractor 追加到现有 HTTP semantic convention extractor 后面：

```go
AddAttributesExtractor(existingHTTPExtractor).
AddAttributesExtractor(newHttpCaptureAttrsExtractor(...))
```

fasthttp 体系的 rule 使用 `pkg/inst-api-semconv/instrumenter/http` 中共享的
`CaptureAttrsExtractor`。

### Hook 改动

Client request：

- 修改 `pkg/rules/http/client_setup.go` 的 `clientOnEnter`。
- allow-list 或 capture-all 开启时，从 `req.Header` 采集配置匹配的请求头。
- 仅当 `req.Body != nil`、content type 符合条件、content length 已知且 `<= 1024`、
  content encoding 为空或 `identity` 时采集 request body。
- 读取后必须还原 `req.Body`，保证业务行为不变。
- 如果 `req.GetBody` 可用，优先读取 `GetBody()` 返回的新 reader，避免消费 live body。

Client response：

- 修改 `clientOnExit`。
- 仅当 `res.Body != nil`、content type 符合条件、content length 已知且 `<= 1024`、
  content encoding 为空或 `identity` 时采集 response body。
- 读取后必须还原 `res.Body`。
- 第一版跳过未知长度 response，避免阻塞 streaming response。

Server request：

- 修改 `pkg/rules/http/server_setup.go` 的 `serverOnEnter`。
- allow-list 或 capture-all 开启时，从 `r.Header` 采集配置匹配的请求头。
- request body 的采集条件与 client request 一致。
- 读取后必须还原 `r.Body`。

Server response：

- 扩展 `writerWrapper`。
- 新增 `Write` 方法，只缓冲前 `1025` bytes，同时继续把原始 bytes 写给底层
  `ResponseWriter`。
- 跟踪 response header 和 content type。如果业务没有设置 `Content-Type`，可用
  `http.DetectContentType` 基于首批缓冲字节判断。
- 在 `serverOnExit` 中，只有最终 body 长度 `<= 1024` 且 content type 符合条件时，才把
  response body 放进 `netHttpResponse`。

wrapper 必须保持已有可选接口行为：`Hijacker`、`Flusher`、`Pusher`、`CloseNotifier`。

fasthttp 和 Fiber：

- 从 `fasthttp.Request` 采集请求头和请求体；
- 从 `fasthttp.Response` 采集响应体；
- 将采集值写入 fasthttp client/server span 和 Fiber server span。

## 测试计划

单元测试：

- 请求头 allow-list、capture-all 开关和 body 采集配置解析；
- header 名称规范化，并将选中的 header map 序列化为一个 JSON attribute；
- capture-all 模式默认跳过敏感 header，除非显式 allow-list；
- 超过大小上限的 header JSON attribute 会被跳过；
- content type 判断；
- 小 body 的读取和还原，包括 `req.GetBody` 路径；
- 大 body、压缩 body、二进制 body、非 UTF-8 body、未知长度 body 的跳过逻辑。

HTTP rule 测试：

- 在 `pkg/rules/http` 增加 capture extractor 和 body helper 的测试；
- 扩展 `test/nethttp`、`test/fasthttp`、`test/fiberv2` 和 `test/fiberv3`，发送 JSON
  请求体和响应体；
- 验证关闭/开启环境变量时 attributes 的差异；
- 验证 headers-only 和 body-only 两种环境变量组合；
- 验证大 body 不被采集，并且业务代码仍能读取原始 body。

建议优先跑：

```bash
(cd pkg/rules/http && go test ./...)
(cd pkg && go test ./inst-api-semconv/instrumenter/http)
TEST_PLUGIN_NAME=nethttp-capture-test go test ./test -run 'TestPlugins4/nethttp-capture-test' -count=1
TEST_PLUGIN_NAME=nethttp-capture-disabled-test go test ./test -run 'TestPlugins4/nethttp-capture-disabled-test' -count=1
TEST_PLUGIN_NAME=nethttp-capture-headers-only-test go test ./test -run 'TestPlugins4/nethttp-capture-headers-only-test' -count=1
TEST_PLUGIN_NAME=nethttp-capture-body-only-test go test ./test -run 'TestPlugins4/nethttp-capture-body-only-test' -count=1
TEST_PLUGIN_NAME=fasthttp-capture-test go test ./test -run 'TestPlugins4/fasthttp-capture-test' -count=1
TEST_PLUGIN_NAME=fasthttp-capture-disabled-test go test ./test -run 'TestPlugins4/fasthttp-capture-disabled-test' -count=1
TEST_PLUGIN_NAME=fiberv2-capture-test go test ./test -run 'TestPlugins4/fiberv2-capture-test' -count=1
TEST_PLUGIN_NAME=fiberv2-capture-disabled-test go test ./test -run 'TestPlugins4/fiberv2-capture-disabled-test' -count=1
TEST_PLUGIN_NAME=fiberv3-capture-test go test ./test -run 'TestPlugins4/fiberv3-capture-test' -count=1
TEST_PLUGIN_NAME=fiberv3-capture-disabled-test go test ./test -run 'TestPlugins4/fiberv3-capture-disabled-test' -count=1
```

## 非目标

- 不默认采集所有请求头。
- 本次不采集响应头。
- 不采集二进制、压缩或非 UTF-8 body。
- 不改变现有 HTTP span name、status extraction、propagation 或 metrics。
- 在 OpenTelemetry 有稳定约定前，不把 body content 提升为 OpenTelemetry semantic
  convention attribute。
