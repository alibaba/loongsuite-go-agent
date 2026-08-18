module test/ollama

go 1.25.0

require (
	github.com/alibaba/loongsuite-go/test/verifier v0.0.0
	github.com/ollama/ollama v0.3.14
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/sdk v1.45.0
	go.opentelemetry.io/otel/sdk/metric v1.45.0
)

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-20251031085506-d38edbf99f97 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/alibaba/loongsuite-go => ../../..

replace github.com/alibaba/loongsuite-go/test/verifier => ../../../test/verifier

replace go.opentelemetry.io/otel/exporters/prometheus => go.opentelemetry.io/otel/exporters/prometheus v0.67.0

replace google.golang.org/protobuf => google.golang.org/protobuf v1.35.2

replace go.opentelemetry.io/otel/exporters/stdout/stdoutmetric => go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.45.0

replace go.opentelemetry.io/otel/exporters/zipkin => go.opentelemetry.io/otel/exporters/zipkin v1.45.0

replace go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.45.0

replace go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.45.0

replace go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.45.0

replace go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.45.0

replace go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp => go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.45.0

replace go.opentelemetry.io/otel/exporters/stdout/stdouttrace => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.45.0

replace go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.45.0

replace go.opentelemetry.io/contrib/instrumentation/runtime => go.opentelemetry.io/contrib/instrumentation/runtime v0.70.0

replace go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.45.0

replace go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.45.0

replace go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.45.0

replace go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc => go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.45.0

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg
