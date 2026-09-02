module otel

go 1.25.0

replace github.com/alibaba/loongsuite-go/test/verifier => ../../../loongsuite-go/test/verifier

replace github.com/alibaba/loongsuite-go => ../../../loongsuite-go

replace github.com/alibaba/loongsuite-go/pkg => ../../pkg

require go.opentelemetry.io/otel/trace v1.45.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
)
