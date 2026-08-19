module test

go 1.25.0

replace github.com/alibaba/loongsuite-go/pkg => ../../pkg

replace github.com/alibaba/loongsuite-go/test/verifier => ../../test/verifier

require (
	go.opentelemetry.io/otel/sdk v1.45.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
