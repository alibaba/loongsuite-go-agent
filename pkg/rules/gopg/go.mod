module github.com/alibaba/loongsuite-go/pkg/rules/gopg

go 1.25.0

replace (
	github.com/alibaba/loongsuite-go/pkg => ../..
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20250218202821-56aae31c358a
)

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-00010101000000-000000000000
	github.com/go-pg/pg/v10 v10.10.0
	go.opentelemetry.io/otel/sdk v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-pg/zerochecker v0.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/bufpool v0.1.11 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.4 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	mellium.im/sasl v0.3.1 // indirect
)
