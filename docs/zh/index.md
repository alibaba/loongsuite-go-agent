![](../public/anim-logo.svg)

**龙蜥Go-Agent** 为希望利用OpenTelemetry实现有效可观察性的Golang应用程序提供了一个自动化的解决方案。目标应用程序无需更改代码，埋点在编译时完成。只需在`go build`前加上`otel`前缀即可开始使用 :rocket:

# 安装

### 预编译二进制文件

- [![下载](https://shields.io/badge/-Linux_AMD64-blue?logo=ubuntu)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-linux-amd64)
- [![下载](https://shields.io/badge/-Linux_ARM64-blue?logo=ubuntu)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-linux-arm64)
- [![下载](https://shields.io/badge/-MacOS_AMD64-blue?logo=apple)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-darwin-amd64)
- [![下载](https://shields.io/badge/-MacOS_ARM64-blue?logo=apple)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-darwin-arm64)
- [![下载](https://shields.io/badge/-Windows_AMD64-blue?logo=wine)](https://github.com/alibaba/loongsuite-go/releases/latest/download/otel-windows-amd64.exe)

**这是安装该工具的推荐方法。**

### 通过Bash安装
对于Linux和MacOS用户，以下脚本默认会将`otel`安装在`/usr/local/bin/otel`：
```bash
$ sudo curl -fsSL https://cdn.jsdelivr.net/gh/alibaba/loongsuite-go@main/install.sh | sudo bash
```

### 从源码构建

```bash
$ make         # 仅构建
$ make install # 构建并安装
```

# 快速开始

确保工具已安装：
```bash
$ # 你也可以使用 "otel-linux-amd64" 代替 "otel"
$ otel version
```

只需在`go build`前加上`otel`前缀来构建你的项目：

```bash
$ otel go build
$ otel go build -o app cmd/app
$ otel go build -gcflags="-m" cmd/app
```

整个过程就是这样！该工具将自动为你的代码注入OpenTelemetry，你就可以开始观察你的应用程序了。 :telescope:

`otel`工具的详细用法可以在[**用法**](./user/config.md)中找到。

> [!NOTE]
> 如果你在`go build`能正常工作的情况下发现任何编译失败，这很可能是一个bug。
> 请随时在[GitHub Issues](https://github.com/alibaba/loongsuite-go/issues)上提交一个bug
> 来帮助我们改进这个项目。

# 示例

- [demo](https://github.com/alibaba/loongsuite-go/tree/main/example/demo) - 带有OpenTelemetry追踪和指标的端到端示例
- [zap logging](https://github.com/alibaba/loongsuite-go/tree/main/example/log) - `github.com/uber-go/zap`日志记录的自动埋点
- [benchmark](https://github.com/alibaba/loongsuite-go/tree/main/example/benchmark) - 性能测试和开销测量
- [sql injection](https://github.com/alibaba/loongsuite-go/tree/main/example/sqlinject) - 用于SQL注入检测的自定义代码注入
- [nethttp](https://github.com/alibaba/loongsuite-go/tree/main/example/nethttp) - 带有请求/响应埋点的HTTP监控
- [kratos-demo](https://github.com/alibaba/loongsuite-go/tree/main/example/kratos-demo) - 与Kratos框架的集成
- [kafka-demo](https://github.com/alibaba/loongsuite-go/tree/main/example/kafka-demo) - Kafka消息消费流程自定义代码注入


# 社区

我们期待您的反馈和建议。您可以加入我们的[钉钉群](https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11)与我们交流。

| 钉钉 | Star历史 |
| :---: | :---: |
| <img src="../public/dingtalk.png" height="200" /> | <img src="https://star-history.dera.page/svg?repos=alibaba/loongsuite-go&type=Date" height="200" /> |
