![](docs/anim-logo.svg)

[![](https://shields.io/badge/Docs-English-blue?logo=Read%20The%20Docs)](./README.md) &nbsp;
[![](https://shields.io/badge/Readme-中文-blue?logo=Read%20The%20Docs)](./docs/README_CN.md)  &nbsp;
[![codecov](https://codecov.io/gh/alibaba/loongsuite-go-agent/branch/main/graph/badge.svg)](https://codecov.io/gh/alibaba/loongsuite-go-agent)  &nbsp;
[![](https://shields.io/badge/Aliyun-Commercial-orange?logo=alibabacloud)](https://help.aliyun.com/zh/arms/application-monitoring/getting-started/monitoring-the-golang-applications) &nbsp;
[![](https://img.shields.io/badge/New-Adopter-orange?logo=githubsponsors)](https://github.com/alibaba/loongsuite-go-agent/issues/225) &nbsp;

**Loongsuite Go Agent** provides an automatic solution for Golang applications that want to
leverage OpenTelemetry to enable effective observability. No code changes are
required in the target application, the instrumentation is done at compile
time. Simply adding `otel` prefix to `go build` to get started :rocket:

# Installation

### Prebuilt Binaries

| Linux AMD64 | Linux ARM64 | MacOS AMD64 | MacOS ARM64 | Windows AMD64 |
| -- | -- | -- | -- | -- |
| [Download](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-amd64) | [Download](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-arm64) | [Download](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-darwin-amd64) | [Download](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-darwin-arm64) | [Download](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-windows-amd64.exe) | 

**This is the recommended way to install the tool.**

### Install via Bash
For Linux and MacOS users, the following script will install `otel` in `/usr/local/bin/otel` by default:
```console
$ sudo curl -fsSL https://cdn.jsdelivr.net/gh/alibaba/loongsuite-go-agent@main/install.sh | sudo bash
```

### Build from Source

```console
$ make         # build only
$ make install # build and install
```

# Getting Started

Make sure the tool is installed:
```console
$ # You may use "otel-linux-amd64" instead of "otel"
$ otel version
```

Just adding `otel` prefix to `go build` to build your project:

```console
$ otel go build
$ otel go build -o app cmd/app
$ otel go build -gcflags="-m" cmd/app
```

That's the whole process! The tool will automatically instrument your code with OpenTelemetry, and you can start to observe your application. :telescope:

The detailed usage of `otel` tool can be found in [**Usage**](./docs/usage.md).

> [!NOTE]
> If you find any compilation failures while `go build` works, it's likely a bug.
> Please feel free to file a bug
> at [GitHub Issues](https://github.com/alibaba/loongsuite-go-agent/issues)
> to help us enhance this project.

# Examples

- [demo](./example/demo) - Complete end-to-end example demonstrating automatic instrumentation with OpenTelemetry tracing and metrics collection.
- [zap logging](./example/log) - Integration example showing how to automatically instrument structured logging with the `github.com/uber-go/zap` package.
- [benchmark](./example/benchmark) - Performance testing suite to measure the overhead and efficiency of the auto-instrumentation tool.
- [sql injection](./example/sqlinject) - Security-focused example demonstrating custom code injection for SQL injection detection and prevention.
- [nethttp](./example/nethttp) - HTTP monitoring example showcasing automatic instrumentation of request/response headers and network traffic analysis.

# Supported Libraries
| Library       | Repository Url                                 | Min           Version | Max           Version |
|---------------|------------------------------------------------|-----------------------|-----------------------|
| database/sql  | https://pkg.go.dev/database/sql                | -                     | -                     |
| dubbo-go      | https://github.com/apache/dubbo-go             | v3.3.0                | -                     |
| echo          | https://github.com/labstack/echo               | v4.0.0                | v4.12.0               |
| eino          | https://github.com/cloudwego/eino              | v0.3.51               | -                     |
| elasticsearch | https://github.com/elastic/go-elasticsearch    | v8.4.0                | v8.15.0               |
| fasthttp      | https://github.com/valyala/fasthttp            | v1.45.0               | v1.63.0               |
| fiber         | https://github.com/gofiber/fiber               | v2.43.0               | v2.52.8               |
| gin           | https://github.com/gin-gonic/gin               | v1.7.0                | v1.10.0               |
| go-redis      | https://github.com/redis/go-redis              | v9.0.5                | v9.5.1                |
| go-redis v8   | https://github.com/redis/go-redis              | v8.11.0               | v8.11.5               |
| gomicro       | https://github.com/micro/go-micro              | v5.0.0                | v5.3.0                |
| gorestful     | https://github.com/emicklei/go-restful         | v3.7.0                | v3.12.1               |
| gorm          | https://github.com/go-gorm/gorm                | v1.22.0               | v1.25.9               |
| grpc          | https://google.golang.org/grpc                 | v1.44.0               | -                     |
| hertz         | https://github.com/cloudwego/hertz             | v0.8.0                | -                     |
| iris          | https://github.com/kataras/iris                | v12.2.0               | v12.2.11              |
| kitex         | https://github.com/cloudwego/kitex             | v0.5.1                | v0.11.3               |
| kratos        | https://github.com/go-kratos/kratos            | v2.6.3                | v2.8.4                |
| langchaingo   | https://github.com/tmc/langchaingo             | v0.1.13               | v0.1.13               |
| log           | https://pkg.go.dev/log                         | -                     | -                     |
| logrus        | https://github.com/sirupsen/logrus             | v1.5.0                | v1.9.3                |
| mongodb       | https://github.com/mongodb/mongo-go-driver     | v1.11.1               | v1.15.1               |
| mux           | https://github.com/gorilla/mux                 | v1.3.0                | v1.8.1                |
| nacos         | https://github.com/nacos-group/nacos-sdk-go/v2 | v2.0.0                | v2.2.7                |
| net/http      | https://pkg.go.dev/net/http                    | -                     | -                     |
| ollama        | https://github.com/ollama/ollama               | v0.3.14               | -                     |
| redigo        | https://github.com/gomodule/redigo             | v1.9.0                | v1.9.2                |
| slog          | https://pkg.go.dev/log/slog                    | -                     | -                     |
| trpc-go       | https://github.com/trpc-group/trpc-go          | v1.0.0                | v1.0.3                |
| zap           | https://github.com/uber-go/zap                 | v1.20.0               | v1.27.0               |
| zerolog       | https://github.com/rs/zerolog                  | v1.10.0               | v1.33.0               |
| go-kit/log    | https://github.com/go-kit/log                  | v0.1.0                | v0.2.1                |
| pg            | https://github.com/go-pg/pg                    | v1.10.0               | v1.14.0               |

We are progressively open-sourcing the libraries we have supported, and your contributions are [**very welcome** 💖!](https://github.com/alibaba/loongsuite-go-agent/issues?q=is%3Aissue%20state%3Aopen%20label%3A%22contribution%20welcome%22)

> [!IMPORTANT]
> The framework you expected is not in the list? Don't worry, you can easily inject your code into any frameworks/libraries that are not officially supported.
>
> Please refer to [this document](./docs/how-to-add-a-new-rule.md) to get started.

# Community

We are looking forward to your feedback and suggestions. You can join
our [DingTalk group](https://qr.dingtalk.com/action/joingroup?code=v1,k1,GyDX5fUTYnJ0En8MrVbHBYTGUcPXJ/NdsmLODGibd0w=&_dt_no_comment=1&origin=11? )
to engage with us.

<img src="docs/dingtalk.png" height="200">

We would thankfully acknowledge the following contributors for their valuable contributions to this project:

<a href="https://github.com/alibaba/loongsuite-go-agent/graphs/contributors">
  <img alt="contributors" src="https://contrib.rocks/image?repo=alibaba/loongsuite-go-agent" height="100"/>
</a>

The star history of this project is as follows, which shows the growth of this project over time:

<img src="https://api.star-history.com/svg?repos=alibaba/loongsuite-go-agent&type=Date" height="200">
