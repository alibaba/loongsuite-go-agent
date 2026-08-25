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
	"errors"

	redis "github.com/redis/go-redis/v9"
)

// redisSpanEndErr returns the error to pass to Instrumenter.End.
// It mirrors upstream otelc go-redis v9 hook logic:
//
//	if err != nil && !errors.Is(err, redis.Nil) {
//	    span.SetStatus(codes.Error, err.Error())
//	}
//
// See: https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/blob/main/instrumentation/github.com/redis/go-redis/v9/hook.go
// See also: pkg/rules/goredisv8/redis_error.go (same nil-filtering logic for v8).
func redisSpanEndErr(err error) error {
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

func redisPipelineSpanEndErr(cmds []redis.Cmder, err error) error {
	for _, cmd := range cmds {
		if spanErr := redisSpanEndErr(cmd.Err()); spanErr != nil {
			return spanErr
		}
	}
	return redisSpanEndErr(err)
}
