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

package goredisv8

import (
	"errors"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type goRedisV8AttrsGetter struct {
}

func (d goRedisV8AttrsGetter) GetSystem(request redisv8Data) string {
	return "redis"
}

func (d goRedisV8AttrsGetter) GetServerAddress(request redisv8Data) string {
	return request.Host
}

func (d goRedisV8AttrsGetter) GetDbNamespace(request redisv8Data) string {
	return ""
}

func (d goRedisV8AttrsGetter) GetBatchSize(request redisv8Data) int {
	return 0
}

func (d goRedisV8AttrsGetter) GetStatement(request redisv8Data) string {
	return getRedisV8Statement(request.cmd)
}

func (d goRedisV8AttrsGetter) GetCollection(request redisv8Data) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (d goRedisV8AttrsGetter) GetOperation(request redisv8Data) string {
	return request.cmd.FullName()
}

func (d goRedisV8AttrsGetter) GetParameters(request redisv8Data) []any {
	return nil
}

func BuildRedisv8Instrumenter() instrumenter.Instrumenter[redisv8Data, any] {
	builder := instrumenter.Builder[redisv8Data, any]{}
	getter := goRedisV8AttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[redisv8Data]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[redisv8Data]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[redisv8Data, any, db.DbClientAttrsGetter[redisv8Data]]{Base: db.DbClientCommonAttrsExtractor[redisv8Data, any, db.DbClientAttrsGetter[redisv8Data]]{Getter: getter}}).
		AddOperationListeners(db.DbClientMetrics("nosql.goredisv8")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GO_REDIS_V8_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}

// getRedisV8Statement builds db.query.text.
//
// Keeps loongsuite's historical format by appending *redis.Cmd via its Stringer
// (not cmd.Name()). Explicit err.Error() text is skipped for redis.Nil to avoid
// duplicating the sentinel already present in cmd.String(); the statement may
// still contain "redis: nil" via that Stringer.
func getRedisV8Statement(cmd redis.Cmder) string {
	b := make([]byte, 0, 64)

	for i, arg := range cmd.Args() {
		if i > 0 {
			b = append(b, ' ')
		}
		b = redisV8AppendArg(b, arg)
	}

	if err := cmd.Err(); err != nil && !errors.Is(err, redis.Nil) {
		b = append(b, ": "...)
		b = append(b, err.Error()...)
	}

	if cmd, ok := cmd.(*redis.Cmd); ok {
		b = append(b, ": "...)
		b = redisV8AppendArg(b, cmd)
	}
	return redisV8String(b)
}
