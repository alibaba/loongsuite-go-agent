# HTTP Header and Body Capture

Status: implemented for `net/http`, `fasthttp`, Fiber v2, and Fiber v3.

This document records two opt-in HTTP instrumentation features:

- capture all request headers as one HTTP span attribute;
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
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS` | `false` | Capture all HTTP request headers into one attribute when set to `true`. |
| `OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED` | `false` | Capture eligible request and response bodies when set to `true`. |

Example:

```bash
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_REQUEST_HEADERS=true
export OTEL_INSTRUMENTATION_HTTP_CAPTURE_BODY_ENABLED=true
```

The header variable applies to client request spans and server request spans for
the supported HTTP instrumentations. Response header capture is out of scope.
When enabled, all request headers are serialized, including sensitive headers
such as `Authorization` and `Cookie`, so keep the default disabled unless this
data is explicitly needed.

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
	captureBody          bool
	maxBodyBytes         int64
}
```

Parse the two environment variables once at package init time. Header capture is
a boolean switch. Values other than `true` keep request header capture disabled.
When enabled, serialize the full request header map as JSON into one attribute.

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
- Capture all request headers from `req.Header` when the boolean switch is
  enabled.
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
- Capture all request headers from `r.Header` when the boolean switch is
  enabled.
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

- parse the boolean header capture config;
- normalize header names and serialize the full header map into one JSON
  attribute;
- classify eligible content types;
- read and restore small request/response bodies;
- skip large, compressed, binary, non-UTF-8, and unknown-length bodies.

HTTP rule tests:

- add tests under `pkg/rules/http` for the capture extractor and body helpers;
- extend `test/nethttp`, `test/fasthttp`, `test/fiberv2`, and `test/fiberv3`
  with focused fixtures that send JSON request and response bodies;
- verify the attributes with capture disabled and enabled;
- verify that large bodies are not captured and that application code can still
  read the original body.

Suggested focused commands:

```bash
(cd pkg/rules/http && go test ./...)
(cd pkg && go test ./inst-api-semconv/instrumenter/http)
TEST_PLUGIN_NAME=nethttp-capture-test go test ./test -run 'TestPlugins4/nethttp-capture-test' -count=1
TEST_PLUGIN_NAME=nethttp-capture-disabled-test go test ./test -run 'TestPlugins4/nethttp-capture-disabled-test' -count=1
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
