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
