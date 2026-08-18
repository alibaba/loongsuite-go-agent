module github.com/alibaba/loongsuite-go/pkg/rules/deepseek

go 1.25.0

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-20260107074919-08c36b668c42
	github.com/cohesion-org/deepseek-go v1.3.0
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/ollama/ollama v0.6.5 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
)

replace github.com/alibaba/loongsuite-go => ../../../

replace github.com/alibaba/loongsuite-go/pkg => ../..
