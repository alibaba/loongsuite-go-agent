module github.com/alibaba/loongsuite-go/pkg/rules/databasesql

go 1.25.0

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-00010101000000-000000000000
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2
	go.opentelemetry.io/otel/sdk v1.45.0
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)
