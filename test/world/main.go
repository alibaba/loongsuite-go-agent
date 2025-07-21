// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// This test matches rules as much as possible and check if compilation works
import (
	_ "database/sql"
	_ "log"
	_ "log/slog"
	_ "net/http"
	_ "runtime"

	_ "dubbo.apache.org/dubbo-go/v3/filter/graceful_shutdown"
	_ "github.com/apache/rocketmq-client-go/v2/consumer"
	_ "github.com/apache/rocketmq-client-go/v2/producer"
	_ "github.com/cloudwego/hertz/pkg/app/server"
	_ "github.com/gin-gonic/gin"
	_ "github.com/go-kratos/kratos/v2/transport/grpc"
	_ "github.com/go-kratos/kratos/v2/transport/http"
	_ "github.com/go-redis/redis/v8"
	_ "github.com/gofiber/fiber/v2"
	_ "github.com/gorilla/mux"
	_ "github.com/labstack/echo/v4"
	_ "github.com/mark3labs/mcp-go/mcp"
	_ "github.com/segmentio/kafka-go"
	_ "github.com/sirupsen/logrus"
	_ "github.com/valyala/fasthttp"
	_ "go.mongodb.org/mongo-driver/mongo"
	_ "go.opentelemetry.io/otel"
	_ "go.opentelemetry.io/otel/baggage"
	_ "go.opentelemetry.io/otel/trace"
	_ "go.uber.org/zap/zapcore"
	_ "google.golang.org/grpc"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/gorm"
)

func main() {

}
