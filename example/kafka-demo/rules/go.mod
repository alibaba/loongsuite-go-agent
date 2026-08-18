module rules

go 1.25.0

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-20260105021138-e8a678009ab0
	github.com/segmentio/kafka-go v0.4.49
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
)

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg
