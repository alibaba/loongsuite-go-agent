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

// Package rueidisnil holds redis-nil filtering helpers for rueidis instrumentation.
//
// These live in a separate package because pkg/rules/rueidisgo cannot be unit-tested
// directly: it references rueidis.RedisAdders, which is only defined via bytecode
// injection at instrument time.
package rueidisnil

import "github.com/redis/rueidis"

// SpanEndErr returns the error to pass to Instrumenter.End.
// rueidis nil replies must not mark the span as Error.
func SpanEndErr(err error) error {
	if err != nil && !rueidis.IsRedisNil(err) {
		return err
	}
	return nil
}

// IsSpanError reports whether err should mark the span as failed.
func IsSpanError(err error) bool {
	return SpanEndErr(err) != nil
}

// FirstError returns the first non-nil Redis error that is not a redis nil reply.
func FirstError(errs []error) error {
	for _, err := range errs {
		if IsSpanError(err) {
			return err
		}
	}
	return nil
}
