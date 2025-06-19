// ABOUTME: Context creation helpers for testing with various timeout and cancellation scenarios
// ABOUTME: Provides standardized context patterns for bridge and engine testing

package testutils

import (
	"context"
	"testing"
	"time"
)

// TestContext creates a basic test context with background
func TestContext() context.Context {
	return context.Background()
}

// TestContextWithTimeout creates a test context with the specified timeout
func TestContextWithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// Ensure cleanup happens
	t.Cleanup(func() {
		cancel()
	})

	return ctx, cancel
}

// TestContextWithCancel creates a test context with cancellation capability
func TestContextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	// Ensure cleanup happens
	t.Cleanup(func() {
		cancel()
	})

	return ctx, cancel
}

// TestContextWithDeadline creates a test context with the specified deadline
func TestContextWithDeadline(t *testing.T, deadline time.Time) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithDeadline(context.Background(), deadline)

	// Ensure cleanup happens
	t.Cleanup(func() {
		cancel()
	})

	return ctx, cancel
}

// TestContextWithValue creates a test context with a key-value pair
func TestContextWithValue(key, value interface{}) context.Context {
	return context.WithValue(context.Background(), key, value)
}

// ShortTimeout returns a short timeout duration for quick tests
func ShortTimeout() time.Duration {
	return 100 * time.Millisecond
}

// MediumTimeout returns a medium timeout duration for regular tests
func MediumTimeout() time.Duration {
	return 1 * time.Second
}

// LongTimeout returns a long timeout duration for complex tests
func LongTimeout() time.Duration {
	return 10 * time.Second
}

// ExpiredContext creates a context that is already expired for timeout testing
func ExpiredContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()
	return ctx
}

// AlmostExpiredContext creates a context that will expire very soon
func AlmostExpiredContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	return ctx
}
