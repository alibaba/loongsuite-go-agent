# HTTP Header and Body Capture

Status: implemented for `net/http`, `fasthttp`, Fiber v2, and Fiber v3.

This document records two opt-in HTTP instrumentation features:

- capture configured request headers as one HTTP span attribute;
- capture small text or JSON request/response bodies as HTTP span attributes.

The design is intentionally opt-in. Headers and bodies often contain credentials,
tokens, user identifiers, or business payloads. The default behavior must remain
unchanged.

## Implementation

The shared capture support is implemented in
`pkg/inst-api-semconv/instrumenter/http/http_capture.go`. Framework-specific
hook integration is implemented in:

- `pkg/rules/http` for `net/http`;
- `pkg/rules/fasthttp` for `fasthttp`;
- `pkg/rules/fiberv2` for Fiber v2;
- `pkg/rules/fiberv3` for Fiber v3.

Frameworks built on top of `net/http`, such as Gin, Echo, chi, and
Gorilla/mux, are covered through the existing `net/http` instrumentation path.
Fiber is covered through its dedicated fasthttp-based rule.

## Environment Variables

| Variable | Default | Meaning |
| --- | --- | --- |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS` | empty | Comma-separated request header allow-list. Matching headers are captured into one attribute. |
| `LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS` | `false` | LoongSuite-specific boolean switch. When set to `true`, captures non-sensitive request headers that are not explicitly allow-listed. |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED` | `false` | Capture eligible request and response bodies when set to `true`. |

Example:

```bash
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=content-type,x-request-id
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true
```

The header variable applies to client request spans and server request spans for
the supported HTTP instrumentations. Response header capture is out of scope.
The OTel-standard header variable is an allow-list. Header names are matched
case-insensitively after trimming spaces.

If `LOONGSUITE_HTTP_CAPTURE_ALL_REQUEST_HEADERS=true`, request headers not named
by the allow-list are also captured, except for common sensitive headers:
`Authorization`, `Cookie`, `Set-Cookie`, `Proxy-Authorization`, `X-Api-Key`, and
`X-Access-Token`. Sensitive headers are captured only when explicitly named in
`OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS`.

## Attribute Names

Request headers are stored in one project-specific attribute:

```text
http.request.headers
```

The attribute value is a JSON string. Header names are normalized to lowercase
and values are preserved as string arrays, for example:

```text
http.request.headers = {"content-type":["application/json"],"x-request-id":["abc"]}
```

This intentionally uses a single attribute instead of the OpenTelemetry
`http.request.header.<name>` convention so enabling the feature does not expand
one request into many distinct attribute names.

The serialized header JSON is omitted when it exceeds `4096` bytes.

Body content does not currently have a stable OpenTelemetry semantic convention
attribute. Use project-specific attributes:

```text
http.request.body.content
http.response.body.content
```

Only store UTF-8 strings. Do not encode binary data into span attributes.

## Body Eligibility

A body is eligible only when all conditions below are true:

- `OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true`;
- the content type is text or JSON:
  - `text/*`;
  - `application/json`;
  - `application/*+json`;
- the body length is known to be at most `1024` bytes, or the instrumentation can
  observe the body as it is written without blocking;
- `Content-Encoding` is empty or `identity`;
- the captured bytes are valid UTF-8.

For `net/http` request bodies and client response bodies, prefer not to read
unknown-length streams in the hook path. Reading an unknown streaming body can
block the application, delay span completion, or change network timing. If
`ContentLength` is negative, skip capture unless a later implementation
introduces a safe tee-based design.

For server response bodies, capture can be done in the `ResponseWriter` wrapper
because bytes are observed while the application writes them. Buffer at most
`1025` bytes; record the body only when the final observed length is at most
`1024`.

For `fasthttp` and Fiber, capture uses the body bytes already held by
`fasthttp.Request`/`fasthttp.Response`; oversized, compressed, binary, and
non-UTF-8 bodies are skipped.

## Implementation Notes

### Configuration

`net/http` keeps a small HTTP-local configuration module:

```text
pkg/rules/http/capture_config.go
```

Suggested internal interface:

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

Parse the environment variables once at package init time. Header capture uses
allow-list semantics for `OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS`.
The LoongSuite-specific capture-all switch is optional and defaults to false.
When capture-all is enabled, skip common sensitive headers unless they were
explicitly allow-listed. Serialize the selected request headers as JSON into one
attribute and omit the attribute when the serialized JSON exceeds 4096 bytes.

### Data Model

Extend the HTTP request and response data structs in
`pkg/rules/http/net_http_data_type.go`:

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

Do not store raw byte slices after conversion. Keeping only strings avoids
retaining large buffers and makes the attribute extractor straightforward.

### Attribute Extractor

`net/http` keeps an HTTP-specific extractor:

```text
pkg/rules/http/capture_attrs_extractor.go
```

It should implement:

```go
instrumenter.AttributesExtractor[*netHttpRequest, *netHttpResponse]
```

Behavior:

- `OnStart` appends `http.request.headers` and `http.request.body.content`;
- `OnEnd` appends `http.response.body.content`;
- no-op when the relevant config is disabled or no captured values exist.

Then add the extractor to both builders in
`pkg/rules/http/net_http_otel_instrumenter.go` after the existing HTTP semantic
convention extractor:

```go
AddAttributesExtractor(existingHTTPExtractor).
AddAttributesExtractor(newHttpCaptureAttrsExtractor(...))
```

The fasthttp-based rules use the shared `CaptureAttrsExtractor` from
`pkg/inst-api-semconv/instrumenter/http`.

### Hook Changes

Client request:

- Update `clientOnEnter` in `pkg/rules/http/client_setup.go`.
- Capture configured request headers from `req.Header` when the allow-list or
  capture-all switch enables header capture.
- Capture the request body only when `req.Body != nil`, content type is eligible,
  content length is known and `<= 1024`, and content encoding is empty or
  `identity`.
- Restore `req.Body` after reading so application behavior is unchanged.
- If `req.GetBody` is available, prefer reading from `GetBody()` instead of
  consuming the live body.

Client response:

- Update `clientOnExit`.
- Capture the response body only when `res.Body != nil`, content type is
  eligible, content length is known and `<= 1024`, and content encoding is empty
  or `identity`.
- Restore `res.Body` after reading.
- Skip unknown-length responses in the first implementation to avoid blocking
  stream responses.

Server request:

- Update `serverOnEnter` in `pkg/rules/http/server_setup.go`.
- Capture configured request headers from `r.Header` when the allow-list or
  capture-all switch enables header capture.
- Capture the request body under the same safe conditions as client request
  capture.
- Restore `r.Body` after reading.

Server response:

- Extend `writerWrapper`.
- Add a `Write` method that buffers at most `1025` bytes while still writing the
  original bytes to the underlying `ResponseWriter`.
- Track response headers and content type. If the application does not set
  `Content-Type`, use `http.DetectContentType` on the first buffered bytes.
- In `serverOnExit`, pass the captured response body to `netHttpResponse` only
  when the final observed body length is `<= 1024` and the content type is
  eligible.

The wrapper must keep the existing optional interfaces working: `Hijacker`,
`Flusher`, `Pusher`, and `CloseNotifier`.

fasthttp and Fiber:

- Capture request headers and request bodies from `fasthttp.Request`.
- Capture response bodies from `fasthttp.Response`.
- Attach captured values to fasthttp client/server spans and Fiber server spans.

## Tests

Unit tests:

- parse the header allow-list, capture-all switch, and body capture config;
- normalize header names and serialize the selected header map into one JSON
  attribute;
- skip sensitive headers in capture-all mode unless explicitly allow-listed;
- skip oversized header JSON attributes;
- classify eligible content types;
- read and restore small request/response bodies, including the `req.GetBody`
  path;
- skip large, compressed, binary, non-UTF-8, and unknown-length bodies.

HTTP rule tests:

- add tests under `pkg/rules/http` for the capture extractor and body helpers;
- extend `test/nethttp`, `test/fasthttp`, `test/fiberv2`, and `test/fiberv3`
  with focused fixtures that send JSON request and response bodies;
- verify the attributes with capture disabled and enabled;
- verify headers-only and body-only environment variable combinations;
- verify that large bodies are not captured and that application code can still
  read the original body.

Suggested focused commands:

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

## Non-goals

- Do not capture all request headers by default.
- Do not capture response headers in this change.
- Do not capture binary, compressed, or non-UTF-8 bodies.
- Do not change existing HTTP span names, status extraction, propagation, or
  metrics.
- Do not promote body content to OpenTelemetry semantic convention attributes
  until a stable convention exists.
