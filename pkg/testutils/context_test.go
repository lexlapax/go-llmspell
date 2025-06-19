// ABOUTME: Tests for context creation helpers ensuring proper timeout and cancellation behavior
// ABOUTME: Validates context utilities for test scenarios including expired and almost-expired contexts

package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestContext(t *testing.T) {
	ctx := TestContext()
	assert.NotNil(t, ctx)
	assert.Equal(t, context.Background(), ctx)
}

func TestTestContextWithTimeout(t *testing.T) {
	timeout := 100 * time.Millisecond

	ctx, cancel := TestContextWithTimeout(t, timeout)
	defer cancel()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)

	// Check that context has deadline
	deadline, ok := ctx.Deadline()
	assert.True(t, ok, "Context should have deadline")
	assert.True(t, deadline.After(time.Now()), "Deadline should be in the future")
	assert.True(t, deadline.Before(time.Now().Add(timeout+10*time.Millisecond)), "Deadline should be approximately timeout duration from now")
}

func TestTestContextWithCancel(t *testing.T) {
	ctx, cancel := TestContextWithCancel(t)
	defer cancel()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)

	// Initially not cancelled
	assert.NoError(t, ctx.Err())

	// Cancel and check
	cancel()
	assert.Error(t, ctx.Err())
	assert.Equal(t, context.Canceled, ctx.Err())
}

func TestTestContextWithDeadline(t *testing.T) {
	deadline := time.Now().Add(100 * time.Millisecond)

	ctx, cancel := TestContextWithDeadline(t, deadline)
	defer cancel()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)

	// Check deadline
	contextDeadline, ok := ctx.Deadline()
	assert.True(t, ok, "Context should have deadline")
	assert.True(t, contextDeadline.Equal(deadline), "Context deadline should match provided deadline")
}

func TestTestContextWithValue(t *testing.T) {
	key := "test_key"
	value := "test_value"

	ctx := TestContextWithValue(key, value)
	assert.NotNil(t, ctx)

	// Check value
	retrievedValue := ctx.Value(key)
	assert.Equal(t, value, retrievedValue)

	// Check missing value
	missingValue := ctx.Value("missing_key")
	assert.Nil(t, missingValue)
}

func TestTimeoutDurations(t *testing.T) {
	short := ShortTimeout()
	medium := MediumTimeout()
	long := LongTimeout()

	assert.Equal(t, 100*time.Millisecond, short)
	assert.Equal(t, 1*time.Second, medium)
	assert.Equal(t, 10*time.Second, long)

	// Verify ordering
	assert.True(t, short < medium)
	assert.True(t, medium < long)
}

func TestExpiredContext(t *testing.T) {
	ctx := ExpiredContext()
	assert.NotNil(t, ctx)

	// Context should already be expired
	assert.Error(t, ctx.Err())
	assert.Equal(t, context.DeadlineExceeded, ctx.Err())

	// Deadline should be in the past
	deadline, ok := ctx.Deadline()
	assert.True(t, ok, "Expired context should have deadline")
	assert.True(t, deadline.Before(time.Now()), "Deadline should be in the past")
}

func TestAlmostExpiredContext(t *testing.T) {
	ctx := AlmostExpiredContext()
	assert.NotNil(t, ctx)

	// Context might or might not be expired yet due to timing
	// But it should have a very recent deadline
	deadline, ok := ctx.Deadline()
	assert.True(t, ok, "Almost expired context should have deadline")

	// Deadline should be very close to now (within a few milliseconds)
	timeDiff := time.Until(deadline)
	assert.True(t, timeDiff < 10*time.Millisecond && timeDiff > -10*time.Millisecond,
		"Deadline should be very close to current time")
}

func TestContextCleanup(t *testing.T) {
	// Test that cleanup functions are properly registered

	// Create contexts that should register cleanup
	ctx1, cancel1 := TestContextWithTimeout(t, 1*time.Second)
	ctx2, cancel2 := TestContextWithCancel(t)
	ctx3, cancel3 := TestContextWithDeadline(t, time.Now().Add(1*time.Second))

	// Contexts should be valid
	assert.NotNil(t, ctx1)
	assert.NotNil(t, ctx2)
	assert.NotNil(t, ctx3)
	assert.NotNil(t, cancel1)
	assert.NotNil(t, cancel2)
	assert.NotNil(t, cancel3)

	// Manual cleanup for testing (normally handled by t.Cleanup)
	cancel1()
	cancel2()
	cancel3()
}

// Test-specific key types to avoid collisions
type contextKey string

func TestContextWithValueChaining(t *testing.T) {
	// Test that we can chain context values
	ctx1 := TestContextWithValue("key1", "value1")
	ctx2 := context.WithValue(ctx1, contextKey("key2"), "value2")

	assert.Equal(t, "value1", ctx2.Value("key1"))
	assert.Equal(t, "value2", ctx2.Value(contextKey("key2")))
	assert.Nil(t, ctx2.Value("nonexistent"))
}

func TestContextTimeoutActualExpiration(t *testing.T) {
	// Test that timeout actually works
	ctx, cancel := TestContextWithTimeout(t, 10*time.Millisecond)
	defer cancel()

	// Wait for timeout
	<-ctx.Done()

	assert.Error(t, ctx.Err())
	assert.Equal(t, context.DeadlineExceeded, ctx.Err())
}
