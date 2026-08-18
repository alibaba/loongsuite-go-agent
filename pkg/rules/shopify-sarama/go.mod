module github.com/alibaba/loongsuite-go/pkg/rules/shopify-sarama

go 1.25.0

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg

require (
	github.com/Shopify/sarama v1.22.0
	github.com/alibaba/loongsuite-go/pkg v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/sdk v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
)

require (
	github.com/DataDog/zstd v1.3.5 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pierrec/lz4 v0.0.0-20190327172049-315a67e90e41 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)
