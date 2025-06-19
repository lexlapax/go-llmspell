// ABOUTME: Tests for MockBridge to ensure it provides consistent mock behavior
// ABOUTME: Validates method handling, state tracking, and builder pattern functionality

package testutils

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockBridge_BasicOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("default behavior", func(t *testing.T) {
		bridge := NewMockBridge("test-bridge")

		// Check initial state
		assert.Equal(t, "test-bridge", bridge.GetID())
		assert.False(t, bridge.IsInitialized())

		metadata := bridge.GetMetadata()
		assert.Equal(t, "test-bridge", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)

		// Initialize
		err := bridge.Initialize(ctx)
		require.NoError(t, err)
		assert.True(t, bridge.IsInitialized())
		assert.Equal(t, 1, bridge.GetInitCallCount())

		// Cleanup
		err = bridge.Cleanup(ctx)
		require.NoError(t, err)
		assert.False(t, bridge.IsInitialized())
	})

	t.Run("double initialization", func(t *testing.T) {
		bridge := NewMockBridge("test-bridge")

		err := bridge.Initialize(ctx)
		require.NoError(t, err)

		err = bridge.Initialize(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")
		assert.Equal(t, 2, bridge.GetInitCallCount())
	})
}

func TestMockBridge_BuilderPattern(t *testing.T) {
	t.Run("custom metadata", func(t *testing.T) {
		customMeta := engine.BridgeMetadata{
			Name:        "Custom Bridge",
			Version:     "2.0.0",
			Description: "Custom description",
			Author:      "Test Author",
			License:     "Apache-2.0",
		}

		bridge := NewMockBridge("test").WithMetadata(customMeta)

		metadata := bridge.GetMetadata()
		assert.Equal(t, customMeta, metadata)
	})

	t.Run("with methods", func(t *testing.T) {
		bridge := NewMockBridge("test").
			WithMethod("method1", engine.MethodInfo{
				Name:        "method1",
				Description: "Test method 1",
				ReturnType:  "string",
			}, nil).
			WithMethod("method2", engine.MethodInfo{
				Name:        "method2",
				Description: "Test method 2",
				Parameters: []engine.ParameterInfo{
					{Name: "param1", Type: "string", Required: true},
				},
				ReturnType: "number",
			}, nil)

		methods := bridge.Methods()
		assert.Len(t, methods, 2)

		// Find methods (order not guaranteed)
		var method1, method2 *engine.MethodInfo
		for i := range methods {
			switch methods[i].Name {
			case "method1":
				method1 = &methods[i]
			case "method2":
				method2 = &methods[i]
			}
		}

		require.NotNil(t, method1)
		require.NotNil(t, method2)
		assert.Equal(t, "Test method 1", method1.Description)
		assert.Equal(t, "Test method 2", method2.Description)
		assert.Len(t, method2.Parameters, 1)
	})

	t.Run("with dependencies and permissions", func(t *testing.T) {
		bridge := NewMockBridge("test").
			WithDependencies("dep1", "dep2").
			WithPermissions(
				engine.Permission{Type: engine.PermissionFileSystem, Resource: "read"},
				engine.Permission{Type: engine.PermissionNetwork, Resource: "all"},
			)

		metadata := bridge.GetMetadata()
		assert.Equal(t, []string{"dep1", "dep2"}, metadata.Dependencies)

		perms := bridge.RequiredPermissions()
		assert.Len(t, perms, 2)
		assert.Equal(t, engine.PermissionFileSystem, perms[0].Type)
		assert.Equal(t, engine.PermissionNetwork, perms[1].Type)
	})
}

func TestMockBridge_MethodExecution(t *testing.T) {
	ctx := context.Background()

	t.Run("custom method handlers", func(t *testing.T) {
		callCount := 0
		bridge := NewMockBridge("test").
			WithMethod("increment", engine.MethodInfo{
				Name:       "increment",
				ReturnType: "number",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				callCount++
				return engine.NewNumberValue(float64(callCount)), nil
			})

		err := bridge.Initialize(ctx)
		assert.NoError(t, err)

		// Execute method multiple times
		result1, err := bridge.ExecuteMethod(ctx, "increment", nil)
		require.NoError(t, err)
		assert.Equal(t, float64(1), result1.(engine.NumberValue).Value())

		result2, err := bridge.ExecuteMethod(ctx, "increment", nil)
		require.NoError(t, err)
		assert.Equal(t, float64(2), result2.(engine.NumberValue).Value())

		// Verify calls were tracked
		calls := bridge.GetMethodCalls("increment")
		assert.Len(t, calls, 2)
	})

	t.Run("method with arguments", func(t *testing.T) {
		bridge := NewMockBridge("test").
			WithMethod("concat", engine.MethodInfo{
				Name: "concat",
				Parameters: []engine.ParameterInfo{
					{Name: "a", Type: "string", Required: true},
					{Name: "b", Type: "string", Required: true},
				},
				ReturnType: "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) < 2 {
					return nil, errors.New("requires 2 arguments")
				}
				a := args[0].String()
				b := args[1].String()
				return engine.NewStringValue(a + b), nil
			})

		err := bridge.Initialize(ctx)
		assert.NoError(t, err)

		args := []engine.ScriptValue{
			engine.NewStringValue("hello"),
			engine.NewStringValue("world"),
		}
		result, err := bridge.ExecuteMethod(ctx, "concat", args)
		require.NoError(t, err)
		assert.Equal(t, "helloworld", result.String())

		// Verify call was tracked with args
		calls := bridge.GetMethodCalls("concat")
		assert.Len(t, calls, 1)
		assert.Len(t, calls[0].Args, 2)
		assert.Equal(t, "hello", calls[0].Args[0].String())
		assert.Equal(t, "world", calls[0].Args[1].String())
	})

	t.Run("unknown method", func(t *testing.T) {
		bridge := NewMockBridge("test")
		err := bridge.Initialize(ctx)
		assert.NoError(t, err)

		_, err = bridge.ExecuteMethod(ctx, "unknown", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown method")

		// Even failed calls should be tracked
		calls := bridge.GetMethodCalls("unknown")
		assert.Len(t, calls, 1)
		assert.NotNil(t, calls[0].Error)
	})

	t.Run("execute without initialization", func(t *testing.T) {
		bridge := NewMockBridge("test")

		_, err := bridge.ExecuteMethod(ctx, "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestMockBridge_Validation(t *testing.T) {
	t.Run("default validation", func(t *testing.T) {
		bridge := NewMockBridge("test").
			WithMethod("required_params", engine.MethodInfo{
				Name: "required_params",
				Parameters: []engine.ParameterInfo{
					{Name: "p1", Type: "string", Required: true},
					{Name: "p2", Type: "number", Required: true},
					{Name: "p3", Type: "bool", Required: false},
				},
			}, nil)

		// Valid call with all required params
		err := bridge.ValidateMethod("required_params", []engine.ScriptValue{
			engine.NewStringValue("test"),
			engine.NewNumberValue(42),
		})
		assert.NoError(t, err)

		// Invalid - too few params
		err = bridge.ValidateMethod("required_params", []engine.ScriptValue{
			engine.NewStringValue("test"),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires at least 2 arguments")

		// Unknown method
		err = bridge.ValidateMethod("unknown", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown method")

		// Verify validation calls were tracked
		calls := bridge.GetValidateCalls()
		assert.Len(t, calls, 3)
	})

	t.Run("custom validation function", func(t *testing.T) {
		bridge := NewMockBridge("test").
			WithValidateFunc(func(method string, args []engine.ScriptValue) error {
				if method == "special" && len(args) != 3 {
					return errors.New("special method requires exactly 3 arguments")
				}
				return nil
			})

		err := bridge.ValidateMethod("special", []engine.ScriptValue{
			engine.NewStringValue("a"),
			engine.NewStringValue("b"),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 3 arguments")

		err = bridge.ValidateMethod("special", []engine.ScriptValue{
			engine.NewStringValue("a"),
			engine.NewStringValue("b"),
			engine.NewStringValue("c"),
		})
		assert.NoError(t, err)
	})
}

func TestMockBridge_ErrorInjection(t *testing.T) {
	ctx := context.Background()

	t.Run("init error", func(t *testing.T) {
		expectedErr := errors.New("init failed")
		bridge := NewMockBridge("test").WithInitError(expectedErr)

		err := bridge.Initialize(ctx)
		assert.Equal(t, expectedErr, err)
		assert.False(t, bridge.IsInitialized())
	})

	t.Run("custom init function", func(t *testing.T) {
		initCalled := false
		bridge := NewMockBridge("test").
			WithInitFunc(func(ctx context.Context) error {
				initCalled = true
				return errors.New("custom init error")
			})

		err := bridge.Initialize(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom init error")
		assert.True(t, initCalled)
	})
}

func TestMockBridge_EngineRegistration(t *testing.T) {
	bridge := NewMockBridge("test")
	mockEngine := NewMockScriptEngine()
	err := mockEngine.Initialize(engine.EngineConfig{})
	assert.NoError(t, err)

	err = bridge.RegisterWithEngine(mockEngine)
	require.NoError(t, err)

	// Verify bridge was registered
	registeredBridge, err := mockEngine.GetBridge("test")
	require.NoError(t, err)
	assert.Equal(t, bridge.GetID(), registeredBridge.GetID())
}

func TestMockAsyncBridge(t *testing.T) {
	ctx := context.Background()

	t.Run("async method marking", func(t *testing.T) {
		asyncBridge := NewMockAsyncBridge("async-test")
		asyncBridge.WithAsyncMethod("asyncOp", engine.MethodInfo{
			Name:       "asyncOp",
			ReturnType: "promise",
		}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
			return engine.NewStringValue("async result"), nil
		})
		asyncBridge.WithMethod("syncOp", engine.MethodInfo{
			Name:       "syncOp",
			ReturnType: "string",
		}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
			return engine.NewStringValue("sync result"), nil
		})

		assert.True(t, asyncBridge.IsAsyncMethod("asyncOp"))
		assert.False(t, asyncBridge.IsAsyncMethod("syncOp"))
		assert.False(t, asyncBridge.IsAsyncMethod("unknown"))

		// Both methods should work normally
		err := asyncBridge.Initialize(ctx)
		assert.NoError(t, err)

		result, err := asyncBridge.ExecuteMethod(ctx, "asyncOp", nil)
		require.NoError(t, err)
		assert.Equal(t, "async result", result.String())

		result, err = asyncBridge.ExecuteMethod(ctx, "syncOp", nil)
		require.NoError(t, err)
		assert.Equal(t, "sync result", result.String())
	})
}
