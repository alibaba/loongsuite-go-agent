package main

import (
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"os"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/gomodule/redigo/redis"
)

func main() {
	c, err := redis.Dial("tcp", "localhost:"+os.Getenv("REDIS_PORT"))
	if err != nil {
		panic(err)
	}
	defer c.Close()
	c.Do("SET", "foo", "bar")
	c.Do("GET", "foo")
	reply, err := c.Do("GET", "missing-key")
	_, err = redis.String(reply, err)
	if !errors.Is(err, redis.ErrNil) {
		panic("expected redis.ErrNil for missing key via redis.String: " + fmt.Sprint(err))
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "SET", "redis", "localhost", "SET foo bar", "SET", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "GET", "redis", "localhost", "GET foo", "GET", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "GET", "redis", "localhost", "GET missing-key", "GET", "", nil)
		if stubs[2][0].Status.Code != codes.Unset {
			panic("redis.ErrNil should not be span error status")
		}
	}, 3)
}
