package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

func TestStartContainerWithRetryFn_RetryableErrorThenSuccess(t *testing.T) {
	callCount := 0
	sleepCount := 0
	_, err := startContainerWithRetryFnWithSleep(
		context.Background(),
		testcontainers.GenericContainerRequest{},
		func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
			callCount++
			if callCount < containerStartMaxRetries {
				return nil, errors.New("create container: unauthorized: authentication required")
			}
			return nil, nil
		},
		func(time.Duration) {
			sleepCount++
		},
	)

	if err != nil {
		t.Fatalf("expected retry to eventually succeed, got error: %v", err)
	}
	if callCount != containerStartMaxRetries {
		t.Fatalf("expected %d attempts, got %d", containerStartMaxRetries, callCount)
	}
	if sleepCount != containerStartMaxRetries-1 {
		t.Fatalf("expected %d retry delays, got %d", containerStartMaxRetries-1, sleepCount)
	}
}

func TestStartContainerWithRetryFn_NonRetryableError(t *testing.T) {
	callCount := 0
	expectedErr := errors.New("create container: invalid reference format")
	_, err := startContainerWithRetryFn(context.Background(), testcontainers.GenericContainerRequest{}, func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		callCount++
		return nil, expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 attempt for non-retryable error, got %d", callCount)
	}
}
