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

import (
	"context"
	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"os"
	"time"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + os.Getenv("REDIS_PORT"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := rdb.Set(ctx, "a", "b", 5*time.Second).Result()
	if err != nil {
		panic(err)
	}
	// get a key that does not exist
	rdb.Do(ctx, "get", "key").Result()

	// pipeline with cache miss followed by a real Redis error
	_, err = rdb.Set(ctx, "string-key", "not-a-number", 0).Result()
	if err != nil {
		panic(err)
	}
	pipe := rdb.Pipeline()
	pipe.Get(ctx, "missing-key")
	pipe.Incr(ctx, "string-key")
	_, _ = pipe.Exec(ctx)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "set", "redis", "localhost", "set a b ex 5", "set", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "get", "redis", "localhost", "get key", "get", "", nil)
		if stubs[1][0].Status.Code != codes.Unset {
			panic("redis.Nil should not be span error status")
		}
		if stubs[3][0].Status.Code != codes.Error {
			panic("mixed pipeline with redis.Nil and real error should be span error status")
		}
	}, 4)
}
