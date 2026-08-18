module github.com/alibaba/loongsuite-go/pkg/rules/asynq

go 1.25.0

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-00010101000000-000000000000
	github.com/hibiken/asynq v0.23.0
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-redis/redis/v8 v8.11.2 // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
