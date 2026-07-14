module github.com/alibaba/loongsuite-go/pkg/rules/rueidisgo

go 1.24.0

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-00010101000000-000000000000
	github.com/redis/rueidis v1.0.30
	go.opentelemetry.io/otel/trace v1.40.0

)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk v1.40.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)
