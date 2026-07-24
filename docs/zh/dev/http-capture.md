# HTTP 请求头和 Body 采集方案

状态：已在 `net/http` 探针实现。

本文记录两个 opt-in 的 HTTP 探针增强功能：

- 将全部请求头采集到一个 HTTP span attribute；
- 当请求体/响应体是 text 或 JSON，且长度不超过 1 KiB 时，将 body 内容采集到 HTTP
  span attribute。

该功能必须默认关闭。请求头和 body 经常包含凭证、token、用户标识或业务数据，不能改变
当前默认采集行为。

## 实现位置

实现范围限定在 `pkg/rules/http`：

- `capture_config.go` 解析环境变量；
- `capture_body.go` 校验、读取、还原和缓冲符合条件的 body；
- `capture_attrs_extractor.go` 将采集值写入 span attribute；
- `client_setup.go` 和 `server_setup.go` 将采集数据接入 net/http span。

## 环境变量

| 环境变量 | 默认值 | 含义 |
| --- | --- | --- |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS` | `false` | 设置为 `true` 时，将全部 HTTP 请求头采集到一个 attribute。 |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED` | `false` | 设置为 `true` 时，采集符合条件的请求体和响应体。 |

示例：

```bash
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=true
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true
```

请求头变量同时作用于 HTTP client span 和 HTTP server span 的 request headers。响应头采集
不包含在本方案范围内。开启后会采集全部请求头，包括 `Authorization`、`Cookie` 等敏感
header，所以除非明确需要这些数据，否则应保持默认关闭。

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

对于 request body 和 HTTP client response body，第一版不建议在 hook 路径里读取未知长度
stream。读取未知长度的 streaming body 可能阻塞业务、延迟 span 结束，或者改变网络时序。
如果 `ContentLength < 0`，建议跳过采集，除非后续实现引入安全的 tee 方案。

对于 HTTP server response body，可以在 `ResponseWriter` wrapper 中采集，因为 wrapper
是在业务写响应时观察字节。最多缓冲 `1025` bytes；只有最终观察到的 body 长度不超过
`1024` 时才写入 attribute。

## 实现计划

### 配置模块

新增 HTTP 局部配置模块：

```text
pkg/rules/http/capture_config.go
```

建议的内部接口：

```go
type httpCaptureConfig struct {
	captureRequestHeaders bool
	captureBody          bool
	maxBodyBytes         int64
}
```

两个环境变量在 package 初始化时解析一次。请求头采集是 boolean 开关，只有值为 `true`
时启用，其他值都保持关闭。开启后将完整 request header map 序列化为 JSON，并写入一个
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

新增 HTTP 专用 extractor：

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

这样新逻辑只影响 `net/http`，不会污染其他复用通用 HTTP semantic convention extractor
的探针。

### Hook 改动

Client request：

- 修改 `pkg/rules/http/client_setup.go` 的 `clientOnEnter`。
- boolean 开启时，从 `req.Header` 采集全部请求头。
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
- boolean 开启时，从 `r.Header` 采集全部请求头。
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

## 测试计划

单元测试：

- boolean 请求头采集配置解析；
- header 名称规范化，并将完整 header map 序列化为一个 JSON attribute；
- content type 判断；
- 小 body 的读取和还原；
- 大 body、压缩 body、二进制 body、非 UTF-8 body、未知长度 body 的跳过逻辑。

HTTP rule 测试：

- 在 `pkg/rules/http` 增加 capture extractor 和 body helper 的测试；
- 扩展 `test/nethttp`，或新增聚焦 fixture，发送 JSON/text 请求体和响应体；
- 验证关闭/开启环境变量时 attributes 的差异；
- 验证大 body 不被采集，并且业务代码仍能读取原始 body。

建议优先跑：

```bash
go test ./pkg/rules/http/...
go test ./test -run NetHttp
```

## 非目标

- 不默认采集所有请求头。
- 本次不采集响应头。
- 不采集二进制、压缩或非 UTF-8 body。
- 不改变现有 HTTP span name、status extraction、propagation 或 metrics。
- 在 OpenTelemetry 有稳定约定前，不把 body content 放进通用 semantic convention 包。
