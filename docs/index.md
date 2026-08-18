![](./public/anim-logo.svg)

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

The detailed usage of `otel` tool can be found in [**Usage**](./user/config.md).

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

# Community

We are looking forward to your feedback and suggestions. You can join
our [DingTalk group](https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11)
to engage with us.

| DingTalk | Star History |
| :---: | :---: |
| <img src="./public/dingtalk.png" height="200" /> | <img src="https://star-history.dera.page/svg?repos=alibaba/loongsuite-go&type=Date" height="200" /> |