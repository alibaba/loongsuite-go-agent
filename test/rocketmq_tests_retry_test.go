package test

import (
	"context"
	"errors"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

func TestStartContainerWithRetryFn_RetryableErrorThenSuccess(t *testing.T) {
	originalDelay := containerStartRetryDelay
	containerStartRetryDelay = 0
	defer func() {
		containerStartRetryDelay = originalDelay
	}()

	callCount := 0
	_, err := startContainerWithRetryFn(context.Background(), testcontainers.GenericContainerRequest{}, func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
		callCount++
		if callCount < containerStartMaxRetries {
			return nil, errors.New("create container: unauthorized: authentication required")
		}
		return nil, nil
	})

	if err != nil {
		t.Fatalf("expected retry to eventually succeed, got error: %v", err)
	}
	if callCount != containerStartMaxRetries {
		t.Fatalf("expected %d attempts, got %d", containerStartMaxRetries, callCount)
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
