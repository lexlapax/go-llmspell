// ABOUTME: Tests for the agent execution context that provides resource limits, cancellation, and tracing
// ABOUTME: Ensures proper context propagation, timeout handling, and multi-engine execution support

package agent

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentContext_Creation tests context creation and initialization
func TestAgentContext_Creation(t *testing.T) {
	t.Run("create with defaults", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		// Check defaults
		assert.NotNil(t, ctx.Context())
		assert.Equal(t, DefaultTimeout, ctx.Timeout())
		assert.Equal(t, DefaultMaxMemory, ctx.MaxMemory())
		assert.Equal(t, DefaultMaxCPU, ctx.MaxCPU())
		assert.Empty(t, ctx.TraceID())
		assert.Empty(t, ctx.SpanID())
	})

	t.Run("create with options", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTimeout(5*time.Second),
			WithContextMaxMemory(1024*1024*100), // 100MB
			WithContextMaxCPU(2),
			WithContextTracing("trace-123", "span-456"),
		)
		require.NotNil(t, ctx)

		// Check custom values
		assert.Equal(t, 5*time.Second, ctx.Timeout())
		assert.Equal(t, int64(1024*1024*100), ctx.MaxMemory())
		assert.Equal(t, 2, ctx.MaxCPU())
		assert.Equal(t, "trace-123", ctx.TraceID())
		assert.Equal(t, "span-456", ctx.SpanID())
	})

	t.Run("create with metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"user_id":    "user-123",
			"request_id": "req-456",
			"engine":     "lua",
		}
		ctx := NewAgentContext(
			context.Background(),
			WithContextMetadata(metadata),
		)
		require.NotNil(t, ctx)

		// Check metadata
		assert.Equal(t, metadata, ctx.Metadata())
		assert.Equal(t, "user-123", ctx.GetMetadata("user_id"))
		assert.Equal(t, "req-456", ctx.GetMetadata("request_id"))
		assert.Equal(t, "lua", ctx.GetMetadata("engine"))
		assert.Nil(t, ctx.GetMetadata("non_existent"))
	})

	t.Run("inherit from parent context", func(t *testing.T) {
		// Create parent with value
		type parentKeyType string
		const parentKey parentKeyType = "parent_key"
		parent := context.WithValue(context.Background(), parentKey, "parent_value")

		ctx := NewAgentContext(parent)
		require.NotNil(t, ctx)

		// Check parent value is accessible
		assert.Equal(t, "parent_value", ctx.Context().Value(parentKey))
	})
}

// TestAgentContext_Cancellation tests context cancellation and timeout
func TestAgentContext_Cancellation(t *testing.T) {
	t.Run("manual cancellation", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		// Check not cancelled initially
		assert.False(t, ctx.IsCancelled())
		assert.NoError(t, ctx.Context().Err())

		// Cancel
		ctx.Cancel()

		// Check cancelled
		assert.True(t, ctx.IsCancelled())
		assert.Equal(t, context.Canceled, ctx.Context().Err())
	})

	t.Run("timeout cancellation", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTimeout(50*time.Millisecond),
		)
		require.NotNil(t, ctx)

		// Wait for timeout
		time.Sleep(100 * time.Millisecond)

		// Check timed out
		assert.True(t, ctx.IsCancelled())
		assert.Equal(t, context.DeadlineExceeded, ctx.Context().Err())
	})

	t.Run("parent cancellation", func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		ctx := NewAgentContext(parent)
		require.NotNil(t, ctx)

		// Cancel parent
		cancel()

		// Check child is also cancelled
		assert.True(t, ctx.IsCancelled())
		assert.Equal(t, context.Canceled, ctx.Context().Err())
	})

	t.Run("done channel", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		done := ctx.Done()
		select {
		case <-done:
			t.Fatal("context should not be done initially")
		default:
			// Expected
		}

		// Cancel
		ctx.Cancel()

		// Check done channel is closed
		select {
		case <-done:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Fatal("done channel should be closed")
		}
	})
}

// TestAgentContext_ResourceLimits tests resource limit enforcement
func TestAgentContext_ResourceLimits(t *testing.T) {
	t.Run("memory limit checking", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextMaxMemory(1024*1024), // 1MB
		)
		require.NotNil(t, ctx)

		// Check under limit
		assert.True(t, ctx.CheckMemoryLimit(512*1024)) // 512KB

		// Record usage
		ctx.RecordMemoryUsage(512 * 1024)
		assert.Equal(t, int64(512*1024), ctx.CurrentMemory())

		// Check would exceed limit
		assert.False(t, ctx.CheckMemoryLimit(600*1024)) // Would be 1.1MB total

		// Can still use up to limit
		assert.True(t, ctx.CheckMemoryLimit(512*1024)) // Exactly 1MB total
	})

	t.Run("CPU limit checking", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextMaxCPU(2),
		)
		require.NotNil(t, ctx)

		// Check under limit
		assert.True(t, ctx.CheckCPULimit(1))

		// Record usage
		ctx.RecordCPUUsage(1)
		assert.Equal(t, 1, ctx.CurrentCPU())

		// Check would exceed limit
		assert.False(t, ctx.CheckCPULimit(2)) // Would be 3 total

		// Can still use up to limit
		assert.True(t, ctx.CheckCPULimit(1)) // Exactly 2 total
	})

	t.Run("resource cleanup", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextMaxMemory(1024*1024),
			WithContextMaxCPU(2),
		)
		require.NotNil(t, ctx)

		// Record usage
		ctx.RecordMemoryUsage(512 * 1024)
		ctx.RecordCPUUsage(1)

		// Release resources
		ctx.ReleaseMemory(256 * 1024)
		ctx.ReleaseCPU(1)

		// Check updated usage
		assert.Equal(t, int64(256*1024), ctx.CurrentMemory())
		assert.Equal(t, 0, ctx.CurrentCPU())
	})

	t.Run("concurrent resource tracking", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextMaxMemory(1024*1024*10), // 10MB
		)
		require.NotNil(t, ctx)

		var wg sync.WaitGroup
		numGoroutines := 100
		memPerGoroutine := int64(1024) // 1KB each

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx.RecordMemoryUsage(memPerGoroutine)
			}()
		}

		wg.Wait()

		// Check total
		assert.Equal(t, memPerGoroutine*int64(numGoroutines), ctx.CurrentMemory())
	})
}

// TestAgentContext_Tracing tests distributed tracing support
func TestAgentContext_Tracing(t *testing.T) {
	t.Run("create span", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTracing("trace-123", "parent-span"),
		)
		require.NotNil(t, ctx)

		// Create child span
		span := ctx.StartSpan("operation", map[string]interface{}{
			"agent_id": "test-agent",
			"action":   "process",
		})
		require.NotNil(t, span)

		// Check span details
		assert.Equal(t, "operation", span.Name())
		assert.Equal(t, "trace-123", span.TraceID())
		assert.NotEmpty(t, span.SpanID())
		assert.Equal(t, "parent-span", span.ParentID())
		assert.Equal(t, "test-agent", span.Attributes()["agent_id"])
		assert.Equal(t, "process", span.Attributes()["action"])
	})

	t.Run("span lifecycle", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTracing("trace-123", "parent-span"),
		)
		require.NotNil(t, ctx)

		// Start span
		span := ctx.StartSpan("test-op", nil)
		require.NotNil(t, span)
		assert.False(t, span.IsEnded())

		// Add attributes
		span.SetAttribute("key1", "value1")
		span.SetAttribute("key2", 42)
		assert.Equal(t, "value1", span.Attributes()["key1"])
		assert.Equal(t, 42, span.Attributes()["key2"])

		// Add event
		span.AddEvent("checkpoint", map[string]interface{}{
			"progress": 50,
		})

		events := span.Events()
		assert.Len(t, events, 1)
		assert.Equal(t, "checkpoint", events[0].Name)
		assert.Equal(t, 50, events[0].Attributes["progress"])

		// End span
		span.End()
		assert.True(t, span.IsEnded())
		assert.NotZero(t, span.Duration())
	})

	t.Run("span error recording", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTracing("trace-123", "parent-span"),
		)
		require.NotNil(t, ctx)

		span := ctx.StartSpan("error-op", nil)
		require.NotNil(t, span)

		// Record error
		testErr := errors.New("test error")
		span.RecordError(testErr)

		// Check error recorded
		assert.True(t, span.HasError())
		assert.Equal(t, "error", span.Status())
		assert.Contains(t, span.Attributes()["error.message"], "test error")
	})

	t.Run("trace propagation", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTracing("trace-123", "root-span"),
		)
		require.NotNil(t, ctx)

		// Create nested spans
		span1 := ctx.StartSpan("span1", nil)
		ctx.SetCurrentSpan(span1)

		span2 := ctx.StartSpan("span2", nil)
		assert.Equal(t, span1.SpanID(), span2.ParentID())

		ctx.SetCurrentSpan(span2)

		span3 := ctx.StartSpan("span3", nil)
		assert.Equal(t, span2.SpanID(), span3.ParentID())
	})
}

// TestAgentContext_Metadata tests metadata management
func TestAgentContext_Metadata(t *testing.T) {
	t.Run("set and get metadata", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		// Set metadata
		ctx.SetMetadata("key1", "value1")
		ctx.SetMetadata("key2", 123)
		ctx.SetMetadata("key3", true)

		// Get metadata
		assert.Equal(t, "value1", ctx.GetMetadata("key1"))
		assert.Equal(t, 123, ctx.GetMetadata("key2"))
		assert.Equal(t, true, ctx.GetMetadata("key3"))
		assert.Nil(t, ctx.GetMetadata("non_existent"))
	})

	t.Run("update metadata", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextMetadata(map[string]interface{}{
				"key1": "initial",
				"key2": 100,
			}),
		)
		require.NotNil(t, ctx)

		// Update existing
		ctx.SetMetadata("key1", "updated")
		ctx.SetMetadata("key2", 200)

		// Add new
		ctx.SetMetadata("key3", "new")

		// Check all
		metadata := ctx.Metadata()
		assert.Equal(t, "updated", metadata["key1"])
		assert.Equal(t, 200, metadata["key2"])
		assert.Equal(t, "new", metadata["key3"])
	})

	t.Run("concurrent metadata access", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		var wg sync.WaitGroup
		numGoroutines := 100

		// Concurrent writes
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				key := "key_" + string(rune('0'+n))
				ctx.SetMetadata(key, n)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = ctx.Metadata()
			}()
		}

		wg.Wait()

		// Verify all writes succeeded
		metadata := ctx.Metadata()
		for i := 0; i < numGoroutines; i++ {
			key := "key_" + string(rune('0'+i))
			assert.Contains(t, metadata, key)
		}
	})
}

// TestAgentContext_MultiEngine tests multi-engine execution support
func TestAgentContext_MultiEngine(t *testing.T) {
	t.Run("engine context isolation", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		// Create engine-specific contexts
		luaCtx := ctx.ForEngine("lua")
		jsCtx := ctx.ForEngine("javascript")
		tengoCtx := ctx.ForEngine("tengo")

		// Set engine-specific metadata
		luaCtx.SetMetadata("engine_data", "lua_specific")
		jsCtx.SetMetadata("engine_data", "js_specific")
		tengoCtx.SetMetadata("engine_data", "tengo_specific")

		// Check isolation
		assert.Equal(t, "lua_specific", luaCtx.GetMetadata("engine_data"))
		assert.Equal(t, "js_specific", jsCtx.GetMetadata("engine_data"))
		assert.Equal(t, "tengo_specific", tengoCtx.GetMetadata("engine_data"))
	})

	t.Run("engine resource tracking", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextMaxMemory(1024*1024*10), // 10MB total
		)
		require.NotNil(t, ctx)

		// Create engine contexts
		luaCtx := ctx.ForEngine("lua")
		jsCtx := ctx.ForEngine("javascript")

		// Record engine-specific usage
		luaCtx.RecordMemoryUsage(1024 * 1024 * 3) // 3MB
		jsCtx.RecordMemoryUsage(1024 * 1024 * 2)  // 2MB

		// Check individual usage
		assert.Equal(t, int64(1024*1024*3), luaCtx.CurrentMemory())
		assert.Equal(t, int64(1024*1024*2), jsCtx.CurrentMemory())

		// Check total usage
		assert.Equal(t, int64(1024*1024*5), ctx.CurrentMemory())
	})

	t.Run("engine cancellation propagation", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		// Create engine contexts
		luaCtx := ctx.ForEngine("lua")
		jsCtx := ctx.ForEngine("javascript")

		// Cancel parent
		ctx.Cancel()

		// Check all engine contexts are cancelled
		assert.True(t, luaCtx.IsCancelled())
		assert.True(t, jsCtx.IsCancelled())
	})
}

// TestAgentContext_Hooks tests execution hooks and middleware
func TestAgentContext_Hooks(t *testing.T) {
	t.Run("before execute hook", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		var called bool
		var hookCtx AgentContext
		var hookInput interface{}

		ctx.AddBeforeExecuteHook(func(actx AgentContext, input interface{}) error {
			called = true
			hookCtx = actx
			hookInput = input
			return nil
		})

		// Trigger hook
		err := ctx.BeforeExecute("test input")
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, ctx, hookCtx)
		assert.Equal(t, "test input", hookInput)
	})

	t.Run("after execute hook", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		var called bool
		var hookOutput interface{}
		var hookErr error

		ctx.AddAfterExecuteHook(func(actx AgentContext, output interface{}, err error) {
			called = true
			hookOutput = output
			hookErr = err
		})

		// Trigger hook
		ctx.AfterExecute("test output", errors.New("test error"))
		assert.True(t, called)
		assert.Equal(t, "test output", hookOutput)
		assert.EqualError(t, hookErr, "test error")
	})

	t.Run("multiple hooks execution order", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		var order []int

		ctx.AddBeforeExecuteHook(func(actx AgentContext, input interface{}) error {
			order = append(order, 1)
			return nil
		})

		ctx.AddBeforeExecuteHook(func(actx AgentContext, input interface{}) error {
			order = append(order, 2)
			return nil
		})

		ctx.AddBeforeExecuteHook(func(actx AgentContext, input interface{}) error {
			order = append(order, 3)
			return nil
		})

		// Execute hooks
		err := ctx.BeforeExecute("input")
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, order)
	})

	t.Run("hook error stops execution", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		var firstCalled, secondCalled bool

		ctx.AddBeforeExecuteHook(func(actx AgentContext, input interface{}) error {
			firstCalled = true
			return errors.New("hook error")
		})

		ctx.AddBeforeExecuteHook(func(actx AgentContext, input interface{}) error {
			secondCalled = true
			return nil
		})

		// Execute hooks
		err := ctx.BeforeExecute("input")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hook error")
		assert.True(t, firstCalled)
		assert.False(t, secondCalled)
	})
}

// TestAgentContext_Values tests context value propagation
func TestAgentContext_Values(t *testing.T) {
	t.Run("set and get values", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		require.NotNil(t, ctx)

		// Set values
		ctx = ctx.WithValue("key1", "value1")
		ctx = ctx.WithValue("key2", 123)
		ctx = ctx.WithValue("key3", true)

		// Get values
		assert.Equal(t, "value1", ctx.Value("key1"))
		assert.Equal(t, 123, ctx.Value("key2"))
		assert.Equal(t, true, ctx.Value("key3"))
		assert.Nil(t, ctx.Value("non_existent"))
	})

	t.Run("value inheritance", func(t *testing.T) {
		ctx := NewAgentContext(context.Background())
		ctx = ctx.WithValue("parent_key", "parent_value")

		// Create child context
		childCtx := ctx.ForEngine("lua")
		childCtx = childCtx.WithValue("child_key", "child_value")

		// Parent value accessible in child
		assert.Equal(t, "parent_value", childCtx.Value("parent_key"))
		assert.Equal(t, "child_value", childCtx.Value("child_key"))

		// Child value not accessible in parent
		assert.Nil(t, ctx.Value("child_key"))
	})
}

// TestAgentContext_Deadline tests deadline management
func TestAgentContext_Deadline(t *testing.T) {
	t.Run("deadline from timeout", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTimeout(5*time.Second),
		)
		require.NotNil(t, ctx)

		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 100*time.Millisecond)
	})

	t.Run("no deadline", func(t *testing.T) {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTimeout(0), // No timeout
		)
		require.NotNil(t, ctx)

		_, ok := ctx.Deadline()
		assert.False(t, ok)
	})

	t.Run("earliest deadline wins", func(t *testing.T) {
		// Parent with 10s deadline
		parent, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Child with 5s timeout
		ctx := NewAgentContext(
			parent,
			WithContextTimeout(5*time.Second),
		)
		require.NotNil(t, ctx)

		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 100*time.Millisecond)
	})
}

// Benchmark tests
func BenchmarkAgentContext_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := NewAgentContext(
			context.Background(),
			WithContextTimeout(5*time.Second),
			WithContextMaxMemory(1024*1024*100),
			WithContextMaxCPU(4),
			WithContextTracing("trace-id", "span-id"),
		)
		_ = ctx
	}
}

func BenchmarkAgentContext_MetadataAccess(b *testing.B) {
	ctx := NewAgentContext(context.Background())
	for i := 0; i < 100; i++ {
		ctx.SetMetadata("key_"+string(rune('0'+i)), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.GetMetadata("key_50")
	}
}

func BenchmarkAgentContext_ResourceTracking(b *testing.B) {
	ctx := NewAgentContext(
		context.Background(),
		WithContextMaxMemory(1024*1024*1024), // 1GB
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.RecordMemoryUsage(1024)
		ctx.ReleaseMemory(1024)
	}
}

// Helper constants are defined in context.go
