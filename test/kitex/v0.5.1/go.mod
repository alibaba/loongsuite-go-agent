module kitex/v0.5.1

go 1.25.0

replace github.com/alibaba/loongsuite-go => ../../../

replace github.com/alibaba/loongsuite-go/test/verifier => ../../../test/verifier

// Keep thrift pinned for latestdepth runs. Newer thrift releases changed TProtocol
// APIs to require context.Context, but kitex's bthrift implementation in some
// versions still targets the older signatures (see TestLatest3/kitex-latestdepth-test).
replace github.com/apache/thrift => github.com/apache/thrift v0.13.0

replace github.com/alibaba/loongsuite-go/pkg => ../../../pkg

require (
	github.com/alibaba/loongsuite-go/test/verifier v0.0.0-00010101000000-000000000000
	github.com/apache/thrift v0.13.0
	github.com/cloudwego/kitex v0.5.1
	go.opentelemetry.io/otel/sdk v1.45.0
)

require (
	github.com/alibaba/loongsuite-go/pkg v0.0.0-20251031085506-d38edbf99f97 // indirect
	github.com/bytedance/gopkg v0.0.0-20240711085056-a03554c296f8 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/choleraehyq/pid v0.0.21 // indirect
	github.com/cloudwego/fastpb v0.0.4 // indirect
	github.com/cloudwego/frugal v0.2.0 // indirect
	github.com/cloudwego/gopkg v0.1.0 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cloudwego/netpoll v0.3.2 // indirect
	github.com/cloudwego/thriftgo v0.2.8 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jhump/protoreflect v1.8.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tidwall/gjson v1.17.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	golang.org/x/arch v0.2.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto v0.0.0-20210513213006-bf773b8c8384 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
