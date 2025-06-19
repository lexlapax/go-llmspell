// ABOUTME: Tests for MockScriptEngine to ensure it provides consistent mock behavior
// ABOUTME: Validates all builder patterns, state tracking, and interface compliance

package testutils

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockScriptEngine_BasicOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("default behavior", func(t *testing.T) {
		mock := NewMockScriptEngine()

		// Test initialization
		config := engine.EngineConfig{
			MemoryLimit:  1024 * 1024,
			TimeoutLimit: 30 * time.Second,
		}
		err := mock.Initialize(config)
		require.NoError(t, err)
		assert.True(t, mock.IsInitialized())

		// Test execution
		result, err := mock.Execute(ctx, "test script", nil)
		require.NoError(t, err)
		assert.Equal(t, "executed: test script", result.String())

		// Test execute calls tracking
		calls := mock.GetExecuteCalls()
		assert.Len(t, calls, 1)
		assert.Equal(t, "test script", calls[0].Script)

		// Test shutdown
		err = mock.Shutdown()
		require.NoError(t, err)
		assert.True(t, mock.IsShutdown())
	})

	t.Run("double initialization error", func(t *testing.T) {
		mock := NewMockScriptEngine()
		config := engine.EngineConfig{}

		err := mock.Initialize(config)
		require.NoError(t, err)

		err = mock.Initialize(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")
	})

	t.Run("execute without initialization", func(t *testing.T) {
		mock := NewMockScriptEngine()

		_, err := mock.Execute(ctx, "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestMockScriptEngine_BuilderPattern(t *testing.T) {
	ctx := context.Background()

	t.Run("custom name and version", func(t *testing.T) {
		mock := NewMockScriptEngine().
			WithName("custom-engine").
			WithVersion("2.0.0")

		info := mock.GetEngineInfo()
		assert.Equal(t, "custom-engine", info.Name)
		assert.Equal(t, "2.0.0", info.Version)
	})

	t.Run("custom execute function", func(t *testing.T) {
		executeCount := 0
		mock := NewMockScriptEngine().
			WithExecuteFunc(func(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error) {
				executeCount++
				return engine.NewNumberValue(float64(executeCount)), nil
			})

		err := mock.Initialize(engine.EngineConfig{})
		assert.NoError(t, err)

		result1, err := mock.Execute(ctx, "test1", nil)
		require.NoError(t, err)
		assert.Equal(t, float64(1), result1.(engine.NumberValue).Value())

		result2, err := mock.Execute(ctx, "test2", nil)
		require.NoError(t, err)
		assert.Equal(t, float64(2), result2.(engine.NumberValue).Value())

		// Verify calls were tracked
		calls := mock.GetExecuteCalls()
		assert.Len(t, calls, 2)
	})

	t.Run("error injection", func(t *testing.T) {
		expectedErr := errors.New("injected error")
		mock := NewMockScriptEngine().
			WithInitError(expectedErr)

		err := mock.Initialize(engine.EngineConfig{})
		assert.Equal(t, expectedErr, err)
		assert.False(t, mock.IsInitialized())
	})
}

func TestMockScriptEngine_BridgeManagement(t *testing.T) {
	mock := NewMockScriptEngine()
	err := mock.Initialize(engine.EngineConfig{})
	assert.NoError(t, err)

	// Create test bridges
	bridge1 := NewMockBridge("bridge1")
	bridge2 := NewMockBridge("bridge2")

	t.Run("register bridges", func(t *testing.T) {
		err := mock.RegisterBridge(bridge1)
		require.NoError(t, err)

		err = mock.RegisterBridge(bridge2)
		require.NoError(t, err)

		// Verify registration
		bridges := mock.ListBridges()
		assert.Len(t, bridges, 2)
		assert.Contains(t, bridges, "bridge1")
		assert.Contains(t, bridges, "bridge2")

		// Check register calls tracking
		calls := mock.GetRegisterCalls()
		assert.Equal(t, []string{"bridge1", "bridge2"}, calls)
	})

	t.Run("get bridge", func(t *testing.T) {
		bridge, err := mock.GetBridge("bridge1")
		require.NoError(t, err)
		assert.Equal(t, "bridge1", bridge.GetID())

		// Non-existent bridge
		_, err = mock.GetBridge("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		err := mock.RegisterBridge(bridge1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("unregister bridge", func(t *testing.T) {
		err := mock.UnregisterBridge("bridge1")
		require.NoError(t, err)

		bridges := mock.ListBridges()
		assert.Len(t, bridges, 1)
		assert.NotContains(t, bridges, "bridge1")

		// Try to unregister again
		err = mock.UnregisterBridge("bridge1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestMockScriptEngine_TypeConversion(t *testing.T) {
	mock := NewMockScriptEngine()

	t.Run("ToNative conversions", func(t *testing.T) {
		tests := []struct {
			name     string
			input    engine.ScriptValue
			expected interface{}
		}{
			{
				name:     "string value",
				input:    engine.NewStringValue("hello"),
				expected: "hello",
			},
			{
				name:     "number value",
				input:    engine.NewNumberValue(42.5),
				expected: 42.5,
			},
			{
				name:     "bool value",
				input:    engine.NewBoolValue(true),
				expected: true,
			},
			{
				name:     "nil value",
				input:    engine.NewNilValue(),
				expected: nil,
			},
			{
				name: "object value",
				input: engine.NewObjectValue(map[string]engine.ScriptValue{
					"key": engine.NewStringValue("value"),
				}),
				expected: map[string]interface{}{
					"key": "value",
				},
			},
			{
				name: "array value",
				input: engine.NewArrayValue([]engine.ScriptValue{
					engine.NewNumberValue(1),
					engine.NewNumberValue(2),
				}),
				expected: []interface{}{1.0, 2.0},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mock.ToNative(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("ToScriptValue conversions", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected engine.ScriptValue
		}{
			{
				name:     "string",
				input:    "hello",
				expected: engine.NewStringValue("hello"),
			},
			{
				name:     "float64",
				input:    42.5,
				expected: engine.NewNumberValue(42.5),
			},
			{
				name:     "int",
				input:    42,
				expected: engine.NewNumberValue(42),
			},
			{
				name:     "bool",
				input:    true,
				expected: engine.NewBoolValue(true),
			},
			{
				name:     "nil",
				input:    nil,
				expected: engine.NewNilValue(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := mock.ToScriptValue(tt.input)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestMockScriptEngine_Context(t *testing.T) {
	mock := NewMockScriptEngine()
	err := mock.Initialize(engine.EngineConfig{})
	assert.NoError(t, err)

	t.Run("create and get context", func(t *testing.T) {
		contextOptions := engine.ContextOptions{
			ID: "test-context",
		}
		ctx, err := mock.CreateContext(contextOptions)
		require.NoError(t, err)
		assert.Equal(t, "test-context", ctx.ID())

		// Get existing context
		ctx2, err := mock.GetContext("test-context")
		require.NoError(t, err)
		assert.Equal(t, ctx.ID(), ctx2.ID())

		// Get non-existent context
		_, err = mock.GetContext("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("duplicate context creation", func(t *testing.T) {
		contextOptions := engine.ContextOptions{
			ID: "test-context",
		}
		_, err := mock.CreateContext(contextOptions)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("context operations", func(t *testing.T) {
		contextOptions := engine.ContextOptions{
			ID: "ops-context",
		}
		ctx, err := mock.CreateContext(contextOptions)
		require.NoError(t, err)

		// Set and get variables
		err = ctx.SetVariable("key", "value")
		require.NoError(t, err)

		val, err := ctx.GetVariable("key")
		require.NoError(t, err)
		assert.Equal(t, "value", val)

		// Get non-existent variable
		_, err = ctx.GetVariable("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Execute script
		result, err := ctx.Execute("test script")
		require.NoError(t, err)
		assert.Equal(t, "executed: test script", result)

		// Destroy context
		err = ctx.Destroy()
		require.NoError(t, err)
	})
}

func TestMockScriptEngine_Concurrency(t *testing.T) {
	mock := NewMockScriptEngine()
	err := mock.Initialize(engine.EngineConfig{})
	assert.NoError(t, err)

	ctx := context.Background()

	// Run multiple concurrent executions
	const goroutines = 10
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			_, err := mock.Execute(ctx, fmt.Sprintf("script-%d", id), nil)
			errors <- err
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < goroutines; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	// Verify all calls were tracked
	calls := mock.GetExecuteCalls()
	assert.Len(t, calls, goroutines)
}
