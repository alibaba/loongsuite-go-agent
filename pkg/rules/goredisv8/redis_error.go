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

	redis "github.com/go-redis/redis/v8"
)

// redisSpanEndErr returns the error to pass to Instrumenter.End.
// It mirrors upstream otelc go-redis v9 hook logic:
//
//	if err != nil && !errors.Is(err, redis.Nil) {
//	    span.SetStatus(codes.Error, err.Error())
//	}
//
// go-redis v8 uses the same redis.Nil sentinel for protocol nil replies.
// See also: pkg/rules/goredis/redis_error.go (same nil-filtering logic for v9).
func redisSpanEndErr(err error) error {
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

// isRedisSpanError reports whether err should mark the span as failed.
func isRedisSpanError(err error) bool {
	return redisSpanEndErr(err) != nil
}
