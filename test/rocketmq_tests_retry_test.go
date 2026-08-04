package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

func TestStartContainerWithRetryFn_RetryableErrorThenSuccess(t *testing.T) {
	const maxAttempts = 3
	const retryDelay = 5 * time.Millisecond

	callCount := 0
	sleepCount := 0
	_, err := startContainerWithRetryFnWithSleep(
		context.Background(),
		testcontainers.GenericContainerRequest{},
		func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
			callCount++
			if callCount < maxAttempts {
				return nil, errors.New("create container: unauthorized: authentication required")
			}
			return nil, nil
		},
		maxAttempts,
		retryDelay,
		func(time.Duration) {
			sleepCount++
		},
	)

	if err != nil {
		t.Fatalf("expected retry to eventually succeed, got error: %v", err)
	}
	if callCount != maxAttempts {
		t.Fatalf("expected %d attempts, got %d", maxAttempts, callCount)
	}
	if sleepCount != maxAttempts-1 {
		t.Fatalf("expected %d retry delays, got %d", maxAttempts-1, sleepCount)
	}
}

func TestStartContainerWithRetryFn_NonRetryableError(t *testing.T) {
	const maxAttempts = 3
	const retryDelay = 5 * time.Millisecond

	callCount := 0
	expectedErr := errors.New("create container: invalid reference format")
	sleepCount := 0
	_, err := startContainerWithRetryFnWithSleep(
		context.Background(),
		testcontainers.GenericContainerRequest{},
		func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
			callCount++
			return nil, expectedErr
		},
		maxAttempts,
		retryDelay,
		func(time.Duration) {
			sleepCount++
		},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 attempt for non-retryable error, got %d", callCount)
	}
	if sleepCount != 0 {
		t.Fatalf("expected no retry delays for non-retryable error, got %d", sleepCount)
	}
}
