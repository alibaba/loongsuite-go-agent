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

package goredis

import (
	"context"
	"net"
	"os"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"go.opentelemetry.io/otel/trace"

	redis "github.com/redis/go-redis/v9"
)

var goRedisInstrumenter = BuildGoRedisOtelInstrumenter()

type redisV9InnerEnabler struct {
	enabled bool
}

func (r redisV9InnerEnabler) Enable() bool {
	return r.enabled
}

var rv9Enabler = redisV9InnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_REDISV9_ENABLED") != "false"}

var redisV9StartOptions = []trace.SpanStartOption{}

//go:linkname afterNewRedisClient github.com/redis/go-redis/v9.afterNewRedisClient
func afterNewRedisClient(call api.CallContext, client *redis.Client) {
	if !rv9Enabler.Enable() {
		return
	}
	client.AddHook(newOtRedisHook(client.Options().Addr))
}

//go:linkname afterNewFailOverRedisClient github.com/redis/go-redis/v9.afterNewFailOverRedisClient
func afterNewFailOverRedisClient(call api.CallContext, client *redis.Client) {
	if !rv9Enabler.Enable() {
		return
	}
	client.AddHook(newOtRedisHook(client.Options().Addr))
}

//go:linkname afterNewClusterClient github.com/redis/go-redis/v9.afterNewClusterClient
func afterNewClusterClient(call api.CallContext, client *redis.ClusterClient) {
	if !rv9Enabler.Enable() {
		return
	}
	client.OnNewNode(func(rdb *redis.Client) {
		rdb.AddHook(newOtRedisHook(rdb.Options().Addr))
	})
}

//go:linkname afterNewRingClient github.com/redis/go-redis/v9.afterNewRingClient
func afterNewRingClient(call api.CallContext, client *redis.Ring) {
	if !rv9Enabler.Enable() {
		return
	}
	client.OnNewNode(func(rdb *redis.Client) {
		rdb.AddHook(newOtRedisHook(rdb.Options().Addr))
	})
}

//go:linkname afterNewSentinelClient github.com/redis/go-redis/v9.afterNewSentinelClient
func afterNewSentinelClient(call api.CallContext, client *redis.SentinelClient) {
	if !rv9Enabler.Enable() {
		return
	}
	client.AddHook(newOtRedisHook(client.String()))
}

//go:linkname afterClientConn github.com/redis/go-redis/v9.afterClientConn
func afterClientConn(call api.CallContext, client *redis.Conn) {
	if !rv9Enabler.Enable() {
		return
	}
	client.AddHook(newOtRedisHook(client.String()))
}

type otRedisHook struct {
	Addr string
}

func newOtRedisHook(addr string) *otRedisHook {
	return &otRedisHook{
		Addr: addr,
	}
}

func (o *otRedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := next(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return conn, err
	}
}

func (o *otRedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if strings.Contains(cmd.FullName(), "ping") || strings.Contains(cmd.FullName(), "PING") {
			return next(ctx, cmd)
		}
		request := goRedisRequest{
			cmd:      cmd,
			endpoint: o.Addr,
		}
		ctx = goRedisInstrumenter.Start(ctx, request)
		err := next(ctx, cmd)
		// Mirrors otelc ProcessHook: always return err to caller; only non-nil
		// errors excluding redis.Nil are passed to Instrumenter.End.
		goRedisInstrumenter.End(ctx, request, nil, redisSpanEndErr(err))
		return err
	}
}

func (o *otRedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		summary := ""
		summaryCmds := cmds
		if len(summaryCmds) > 10 {
			summaryCmds = summaryCmds[:10]
		}
		for i := range summaryCmds {
			summary += summaryCmds[i].FullName() + "/"
		}
		if len(cmds) > 10 {
			summary += "..."
		}
		cmd := redis.NewCmd(ctx, "pipeline", summary)
		request := goRedisRequest{
			cmd:      cmd,
			endpoint: o.Addr,
		}
		ctx = goRedisInstrumenter.Start(ctx, request)
		err := next(ctx, cmds)
		goRedisInstrumenter.End(ctx, request, nil, redisPipelineSpanEndErr(cmds, err))
		return err
	}
}
