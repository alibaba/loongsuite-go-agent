package main

import (
	"flag"
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/file"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	server "kratos/v3.0.0/pkg"
	"kratos/v3.0.0/pkg/biz"
	"kratos/v3.0.0/pkg/conf"
	"kratos/v3.0.0/pkg/service"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "opentelemetry-kratos-server"
	// Version is the version of the compiled software.
	Version = "v1"
	// flagconf is the config flag.
	flagconf string

	id = "opentelemetry-id"
)

func init() {
	flag.StringVar(&flagconf, "conf", "pkg/configs", "config path, eg: -conf config.yaml")
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{
			"agent": "opentelemetry-go",
		}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func wireApp(confServer *conf.Server, confData *conf.Data, logger *slog.Logger) (*kratos.App, func(), error) {
	greeterUsecase := biz.NewGreeterUsecase()
	greeterService := service.NewGreeterService(greeterUsecase)
	grpcServer := server.NewGRPCServer(confServer, greeterService)
	httpServer := server.NewHTTPServer(confServer, greeterService)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {}, nil
}

func startup() {
	flag.Parse()
	handler := log.NewHandler()
	logger := log.NewLogger(handler)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
