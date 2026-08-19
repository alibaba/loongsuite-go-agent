![](docs/public/anim-logo.svg)

[![](https://shields.io/badge/-Docs-blue?logo=readthedocs)](https://alibaba.github.io/loongsuite-go/)  &nbsp;
[![](https://shields.io/badge/-商业版-blue?logo=alibabacloud)](https://help.aliyun.com/zh/arms/application-monitoring/getting-started/monitoring-the-golang-applications) &nbsp;

**Loongsuite Go Agent** provides an automatic solution for Golang applications that want to
leverage OpenTelemetry to enable effective observability. No code changes are
required in the target application, the instrumentation is done at compile
time. Simply adding `otel` prefix to `go build` to get started :rocket:

# Installation

### Prebuilt Binaries

- [![Download](https://shields.io/badge/-Linux_AMD64-blue?logo=ubuntu)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-linux-amd64)
- [![Download](https://shields.io/badge/-Linux_ARM64-blue?logo=ubuntu)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-linux-arm64)
- [![Download](https://shields.io/badge/-MacOS_AMD64-blue?logo=apple)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-darwin-amd64)
- [![Download](https://shields.io/badge/-MacOS_ARM64-blue?logo=apple)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-darwin-arm64)
- [![Download](https://shields.io/badge/-Windows_AMD64-blue?logo=wine)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-windows-amd64.exe)

**This is the recommended way to install the tool.**

### Install via Bash
For Linux and MacOS users, the following script will install `otel` in `/usr/local/bin/otel` by default:
```bash
$ sudo curl -fsSL https://cdn.jsdelivr.net/gh/alibaba/loongsuite-go@main/install.sh | sudo bash
```

### Build from Source

```bash
$ make         # build only
$ make install # build and install
```

# Getting Started

Make sure the tool is installed:
```bash
$ # You may use "otel-linux-amd64" instead of "otel"
$ otel version
```

Just adding `otel` prefix to `go build` to build your project:

```bash
$ otel go build
$ otel go build -o app cmd/app
$ otel go build -gcflags="-m" cmd/app
```

That's the whole process! The tool will automatically instrument your code with OpenTelemetry, and you can start to observe your application. :telescope:

The detailed usage of `otel` tool can be found in [**Usage**](./docs/user/config.md).

> [!NOTE]
> If you find any compilation failures while `go build` works, it's likely a bug.
> Please feel free to file a bug
> at [GitHub Issues](https://github.com/alibaba/loongsuite-go/issues)
> to help us enhance this project.

# Examples

- [demo](https://github.com/alibaba/loongsuite-go/tree/main/example/demo) - End-to-end example with OpenTelemetry tracing and metrics
- [zap logging](https://github.com/alibaba/loongsuite-go/tree/main/example/log) - Auto-instrumentation for `github.com/uber-go/zap` logging
- [benchmark](https://github.com/alibaba/loongsuite-go/tree/main/example/benchmark) - Performance testing and overhead measurement
- [sql injection](https://github.com/alibaba/loongsuite-go/tree/main/example/sqlinject) - Custom code injection for SQL injection detection
- [nethttp](https://github.com/alibaba/loongsuite-go/tree/main/example/nethttp) - HTTP monitoring with request/response instrumentation
- [kratos-demo](https://github.com/alibaba/loongsuite-go/tree/main/example/kratos-demo) - Integration with the Kratos framework
- [kafka-demo](https://github.com/alibaba/loongsuite-go/tree/main/example/kafka-demo) - Kafka Consumer Message monitoring

# Supported Libraries
<details>
 <summary>List of Supported Libraries</summary>

| Library            | Repository Url                                  | Min Version | Max Version |
|--------------------|-------------------------------------------------|-------------|-------------|
| adk-go             | https://pkg.go.dev/google.golang.org/adk        | v0.2.0      | -           |
| amqp091            | https://github.com/rabbitmq/amqp091-go          | v1.10.0     | -           |
| ants               | https://github.com/panjf2000/ants               | v1.1.0      | -           |
| anthropic-sdk-go   | https://github.com/anthropics/anthropic-sdk-go  | v1.25.0     | -           |
| asynq              | https://github.com/hibiken/asynq                | v0.23.0     | v0.26.0     |
| aws-sdk-go         | https://github.com/aws/aws-sdk-go               | v1.55.5     | -           |
| clickhouse/v2      | https://github.com/ClickHouse/clickhouse-go/v2  | v2.13.0     | -           |
| cron               | https://github.com/robfig/cron/v3               | v3.0.0      | -           |
| database/sql       | https://pkg.go.dev/database/sql                 | -           | -           |
| deepseek           | https://github.com/cohesion-org/deepseek-go     | v1.3.0      | -           |
| dubbo-go           | https://github.com/apache/dubbo-go              | v3.3.0      | -           |
| echo               | https://github.com/labstack/echo                | v4.0.0      | -           |
| elasticsearch      | https://github.com/elastic/go-elasticsearch     | v8.4.0      | v8.15.1     |
| eino               | https://github.com/cloudwego/eino               | v0.3.51     | -           |
| fasthttp           | https://github.com/valyala/fasthttp             | v1.45.0     | -           |
| fiber              | https://github.com/gofiber/fiber                | v2.43.0     | v2.52.13    |
| fiber/v3           | https://github.com/gofiber/fiber/v3             | v3.0.0      | -           |
| franz-go           | https://github.com/twmb/franz-go                | v1.18.0     | -           |
| gin                | https://github.com/gin-gonic/gin                | v1.7.0      | v1.10.2     |
| go-kit/log         | https://github.com/go-kit/log                   | v0.1.0      | v0.2.2      |
| go-micro           | https://github.com/micro/go-micro               | v5.0.0      | v5.3.1      |
| go-openai          | https://github.com/sashabaranov/go-openai       | v1.30.0     | -           |
| gocql              | https://github.com/gocql/gocql                  | v1.3.0      | v1.7.1      |
| google-genai       | https://pkg.go.dev/google.golang.org/genai      | v1.30.0     | -           |
| gopg               | https://github.com/go-pg/pg                     | v10.10.0    | v10.14.1    |
| gorestful/v3       | https://github.com/emicklei/go-restful/v3       | v3.7.0      | v3.12.2     |
| gorm               | https://github.com/go-gorm/gorm                 | v1.22.0     | v1.25.10    |
| gorilla/mux        | https://github.com/gorilla/mux                  | v1.3.0      | v1.8.2      |
| grpc               | https://google.golang.org/grpc                  | v1.44.0     | v1.63.0     |
| hertz              | https://github.com/cloudwego/hertz              | v0.8.0      | -           |
| ibm-sarama         | https://github.com/IBM/sarama                   | v1.40.0     | -           |
| iris               | https://github.com/kataras/iris                 | v12.2.0     | v12.2.12    |
| k8s client-go      | https://github.com/kubernetes/client-go         | v0.33.3     | -           |
| kitex              | https://github.com/cloudwego/kitex              | v0.5.1      | -           |
| kratos             | https://github.com/go-kratos/kratos             | v2.6.3      | -           |
| kratos/v3          | https://github.com/go-kratos/kratos/v3          | v3.0.0      | -           |
| langchaingo        | https://github.com/tmc/langchaingo              | v0.1.13     | -           |
| log                | https://pkg.go.dev/log                          | -           | -           |
| logrus             | https://github.com/sirupsen/logrus              | v1.5.0      | -           |
| mcp                | https://github.com/mark3labs/mcp-go             | v0.20.0     | v0.20.2     |
| mcp go-sdk         | https://github.com/modelcontextprotocol/go-sdk  | v0.7.0      | -           |
| meguminnnnnnnnn/go-openai | https://github.com/meguminnnnnnnnn/go-openai    | -           | -           |
| mongodb            | https://github.com/mongodb/mongo-go-driver      | v1.11.1     | v1.15.2     |
| mqtt               | https://github.com/mochi-mqtt/server            | v2.6.4      | -           |
| neo4j              | https://github.com/neo4j/neo4j-go-driver        | v6.0.0      | v6.2.0      |
| nacos              | https://github.com/nacos-group/nacos-sdk-go/v2  | v2.0.0      | v2.3.0      |
| net/http           | https://pkg.go.dev/net/http                     | -           | -           |
| new-api            | https://github.com/QuantumNous/new-api          | -           | -           |
| ollama             | https://github.com/ollama/ollama                | v0.3.14     | -           |
| openai-go          | https://github.com/openai/openai-go             | v1.5.0      | -           |
| redigo             | https://github.com/gomodule/redigo              | v1.9.0      | -           |
| redis (go-redis)   | https://github.com/redis/go-redis               | v9.0.5      | -           |
| redis v8           | https://github.com/go-redis/redis/v8            | v8.11.0     | v8.11.6     |
| rocketmq           | https://github.com/apache/rocketmq-client-go/v2 | v2.0.0      | -           |
| rpcx               | https://github.com/smallnest/rpcx               | v1.6.2      | -           |
| rueidis            | https://github.com/redis/rueidis                | v1.0.30     | -           |
| segmentio/kafka-go | https://github.com/segmentio/kafka-go           | v0.4.0      | -           |
| sentinel           | https://github.com/alibaba/sentinel-golang      | v1.0.4      | -           |
| shopify-sarama     | https://github.com/Shopify/sarama               | v1.22.0     | -           |
| slog               | https://pkg.go.dev/log/slog                     | -           | -           |
| sqlx               | https://github.com/jmoiron/sqlx                 | v1.3.0      | v1.4.1      |
| streadway-amqp     | https://github.com/streadway/amqp               | v1.0.0      | -           |
| trpc-agent-go      | https://github.com/trpc-group/trpc-agent-go     | v0.1.0      | -           |
| trpc-go            | https://github.com/trpc-group/trpc-go           | v1.0.0      | -           |
| zap                | https://github.com/uber-go/zap                  | v1.20.0     | v1.27.1     |
| zerolog            | https://github.com/rs/zerolog                   | v1.10.0     | v1.34.1     |

</details>

We are progressively open-sourcing the libraries we have supported, and your contributions are <kbd>Very Welcome</kbd>

> [!IMPORTANT]
> The framework you expected is not in the list? Don't worry, you can easily inject your code into any frameworks/libraries that are not officially supported.
>
> Please refer to [this document](./docs/dev/overview.md) to get started.

# Community

We are looking forward to your feedback and suggestions. You can join
our [DingTalk group](https://qr.dingtalk.com/action/joingroup?code=v1,k1,mexukXI88tZ1uiuLYkKhdaETUx/K59ncyFFFG5Voe9s=&_dt_no_comment=1&origin=11?) or scan the QR code below to engage with us.

| LoongCollector SIG | LoongSuite Python SIG |
|----|----|
| <img src="docs/_assets/img/loongcollector-sig-dingtalk.png" height="150"> | <img src="docs/_assets/img/loongsuite-python-sig-dingtalk.jpg" height="150"> |

| LoongCollector Go SIG | LoongSuite Java SIG |
|----|----|
| <img src="docs/_assets/img/loongsuite-go-sig-dingtalk.png" height="150"> | <img src="docs/_assets/img/loongsuite-java-sig-dingtalk.jpg" height="150"> |

# Star History

<img src="https://star-history.dera.page/svg?repos=alibaba/loongsuite-go&type=Date" height="200" />
