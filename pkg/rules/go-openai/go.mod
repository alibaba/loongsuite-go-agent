module github.com/alibaba/loongsuite-go/pkg/rules/go-openai

go 1.25.0

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-20260107074919-08c36b668c42
	github.com/sashabaranov/go-openai v1.30.0
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/sdk v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

replace github.com/alibaba/loongsuite-go => ../../../

replace github.com/alibaba/loongsuite-go/pkg => ../..
