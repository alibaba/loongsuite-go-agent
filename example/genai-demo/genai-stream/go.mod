module github.com/alibaba/loongsuite-go/example/genai-demo/genai-stream

go 1.25.0

require (
	github.com/alibaba/loongsuite-go/util-genai v0.0.0-00010101000000-000000000000
	github.com/sashabaranov/go-openai v1.36.1
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.45.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.45.0
	go.opentelemetry.io/otel/sdk v1.45.0
	go.opentelemetry.io/otel/sdk/metric v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/log v0.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/alibaba/loongsuite-go/util-genai => ../../../util-genai
