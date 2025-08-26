![](images/anim-logo.svg)

[![](https://shields.io/badge/-中文文档-black?logo=dolphin)](./docs/README_CN.md)  &nbsp;
[![](https://shields.io/badge/-阿里云商业版-black?logo=alibabacloud)](https://help.aliyun.com/zh/arms/application-monitoring/getting-started/monitoring-the-golang-applications) &nbsp;
[![](https://shields.io/badge/-Loongsuite_Java-black?logo=alibabacloud)](https://github.com/alibaba/loongsuite-java-agent)  &nbsp;
[![](https://shields.io/badge/-Loongsuite_Python-black?logo=alibabacloud)](https://github.com/alibaba/loongsuite-python-agent)  &nbsp;

**Loongsuite Go Agent** 为希望利用 OpenTelemetry 实现有效可观察性的 Golang 应用程序提供了一个自动解决方案。目标应用程序无需更改代码，插装是在编译时完成的。只需在 `go build` 前添加 `otel` 前缀即可开始 :rocket:

# 安装

### 预编译二进制文件

| Linux AMD64 | Linux ARM64 | MacOS AMD64 | MacOS ARM64 | Windows AMD64 |
| -- | -- | -- | -- | -- |
| [![Download](https://shields.io/badge/-Download-black?logo=winamp)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-amd64) | [![Download](https://shields.io/badge/-Download-black?logo=winamp)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-arm64) | [![Download](https://shields.io/badge/-Download-black?logo=winamp)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-darwin-amd64) | [![Download](https://shields.io/badge/-Download-black?logo=winamp)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-darwin-amd64) | [![Download](https://shields.io/badge/-Download-black?logo=winamp)](https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-windows-amd64.exe) | 

**这是推荐的安装方式。**

### 通过 Bash 安装
对于 Linux 和 MacOS 用户，以下脚本将默认在 `/usr/local/bin/otel` 中安装 `otel`：
```console
$ sudo curl -fsSL https://cdn.jsdelivr.net/gh/alibaba/loongsuite-go-agent@main/install.sh | sudo bash
```

### 从源代码编译

```console
$ make         # 仅构建
$ make install # 构建并安装
```

# 开始使用

确保工具已安装：
```console
$ # 您可以使用 "otel-linux-amd64" 而不是 "otel"
$ otel version
```

只需在 `go build` 前添加 `otel` 前缀来构建您的项目：

```console
$ otel go build
$ otel go build -o app cmd/app
$ otel go build -gcflags="-m" cmd/app
```

这就是整个过程！该工具将自动使用 OpenTelemetry 对您的代码进行插装，您就可以开始观察您的应用程序了。:telescope:

`otel` 工具的详细用法可以在 [**使用指南**](./usage.md) 中找到。

> [!NOTE] 
> 如果您发现 `go build` 能正常工作但出现编译失败，这很可能是一个 bug。
> 请随时在
> [GitHub Issues](https://github.com/alibaba/loongsuite-go-agent/issues)
> 提交问题报告以帮助我们改进本项目。

# 示例

- [demo](../example/demo) - 完整的端到端示例，演示使用 OpenTelemetry 追踪和指标收集的自动插装。
- [zap logging](../example/log) - 集成示例，展示如何自动插装结构化日志记录与 `github.com/uber-go/zap` 包。
- [benchmark](../example/benchmark) - 性能测试套件，用于测量自动插装工具的开销和效率。
- [sql injection](../example/sqlinject) - 安全重点示例，演示用于 SQL 注入检测和预防的自定义代码注入。
- [nethttp](../example/nethttp) - HTTP 监控示例，展示请求/响应头和网络流量分析的自动插装。

# 支持的库

| 库名称         | 存储库网址                                      | 最低支持版本           | 最高支持版本     |
|---------------| ---------------------------------------------- |----------------------|-----------------------|
| database/sql  | https://pkg.go.dev/database/sql                | -                    | -                     |
| dubbo-go      | https://github.com/apache/dubbo-go             | v3.3.0               | -                     |
| echo          | https://github.com/labstack/echo               | v4.0.0               | v4.12.0               |
| eino          | https://github.com/cloudwego/eino              | v0.3.51              | -                     |
| elasticsearch | https://github.com/elastic/go-elasticsearch    | v8.4.0               | v8.15.0               |
| fasthttp      | https://github.com/valyala/fasthttp            | v1.45.0              | v1.63.0               |
| fiber         | https://github.com/gofiber/fiber               | v2.43.0              | v2.52.8               |
| gin           | https://github.com/gin-gonic/gin               | v1.7.0               | v1.10.0               |
| go-redis      | https://github.com/redis/go-redis              | v9.0.5               | v9.5.1                |
| go-redis v8   | https://github.com/redis/go-redis              | v8.11.0              | v8.11.5               |
| gomicro       | https://github.com/micro/go-micro              | v5.0.0               | v5.3.0                |
| gorestful     | https://github.com/emicklei/go-restful         | v3.7.0               | v3.12.1               |
| gorm          | https://github.com/go-gorm/gorm                | v1.22.0              | v1.25.9               |
| grpc          | https://google.golang.org/grpc                 | v1.44.0              | -                     |
| hertz         | https://github.com/cloudwego/hertz             | v0.8.0               | -                     |
| iris          | https://github.com/kataras/iris                | v12.2.0              | v12.2.11              |
| client-go     | https://github.com/kubernetes/client-go        | v0.33.3              | -                     |
| kitex         | https://github.com/cloudwego/kitex             | v0.5.1               | v0.11.3               |
| kratos        | https://github.com/go-kratos/kratos            | v2.6.3               | v2.8.4                |
| langchaingo   | https://github.com/tmc/langchaingo             | v0.1.13              | v0.1.13               |
| log           | https://pkg.go.dev/log                         | -                    | -                     |
| logrus        | https://github.com/sirupsen/logrus             | v1.5.0               | v1.9.3                |
| mongodb       | https://github.com/mongodb/mongo-go-driver     | v1.11.1              | v1.15.1               |
| mux           | https://github.com/gorilla/mux                 | v1.3.0               | v1.8.1                |
| nacos         | https://github.com/nacos-group/nacos-sdk-go/v2 | v2.0.0               | v2.2.7                |
| net/http      | https://pkg.go.dev/net/http                    | -                    | -                     |
| ollama        | https://github.com/ollama/ollama               | v0.3.14              | -                     |
| redigo        | https://github.com/gomodule/redigo             | v1.9.0               | v1.9.2                |
| sentinel      | https://github.com/alibaba/sentinel-golang     | v1.0.4               | -                     |
| slog          | https://pkg.go.dev/log/slog                    | -                    | -                     |
| trpc-go       | https://github.com/trpc-group/trpc-go          | v1.0.0               | v1.0.3                |
| zap           | https://github.com/uber-go/zap                 | v1.20.0              | v1.27.0               |
| zerolog       | https://github.com/rs/zerolog                  | v1.10.0              | v1.33.0               |
| go-kit/log    | https://github.com/go-kit/log                  | v0.1.0               | v0.2.1                |
| pg            | https://github.com/go-pg/pg                    | v1.10.0              | v1.14.0               |
| gocql         | https://github.com/gocql/gocql                 | v1.3.0                | v1.7.0                |

我们正在逐步开源我们支持的库，非常欢迎您的贡献💖！

> [!IMPORTANT]
> 您期望的框架不在列表中？别担心，您可以轻松地将代码注入到任何官方不支持的框架/库中。
>
> 请参考 [这个文档](./how-to-add-a-new-rule.md) 开始使用。

# 社区

我们期待您的反馈和建议。您可以加入我们的 [钉钉群组](https://qr.dingtalk.com/action/joingroup?code=v1,k1,PBuICMTDvdh0En8MrVbHBYTGUcPXJ/NdT6JKCZ8CG+w=&_dt_no_comment=1&origin=11)
与我们交流。

| 钉钉群组 | Star 历史 |
| :---: | :---: |
| <img src="images/dingtalk.png" height="200"> | <img src="https://api.star-history.com/svg?repos=alibaba/loongsuite-go-agent&type=Date" height="200"> |
