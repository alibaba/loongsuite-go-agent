package goredisv8

import (
	"errors"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisSpanEndErr(t *testing.T) {
	assert.Nil(t, redisSpanEndErr(nil))
	assert.Nil(t, redisSpanEndErr(redis.Nil))

	realErr := errors.New("connection refused")
	assert.Equal(t, realErr, redisSpanEndErr(realErr))
}

func TestIsRedisSpanError(t *testing.T) {
	assert.False(t, isRedisSpanError(nil))
	assert.False(t, isRedisSpanError(redis.Nil))
	assert.True(t, isRedisSpanError(errors.New("connection refused")))
}
