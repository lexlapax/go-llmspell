package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHooksBridge_Initialize(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())
}

func TestHooksBridge_GetID(t *testing.T) {
	bridge := NewHooksBridge()
	assert.Equal(t, "hooks", bridge.GetID())
}

func TestHooksBridge_GetMetadata(t *testing.T) {
	bridge := NewHooksBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "Hooks Bridge", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "go-llms agent hook system")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "", metadata.License) // License is empty in the implementation
}

func TestHooksBridge_Methods(t *testing.T) {
	bridge := NewHooksBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"registerHook", "unregisterHook", "listHooks",
		"enableHook", "disableHook", "getHookInfo",
		"executeHooks", "clearHooks",
	}

	assert.Equal(t, len(expectedMethods), len(methods))

	// Check that key methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodNames[expected], "Expected method %s not found", expected)
	}
}

func TestHooksBridge_ValidateMethod(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []engine.ScriptValue
		expectError bool
	}{
		{
			name:   "valid registerHook",
			method: "registerHook",
			args: []engine.ScriptValue{
				sv("test-hook"),
				svMap(map[string]interface{}{
					"priority": 10,
				}),
			},
			expectError: false,
		},
		{
			name:        "invalid registerHook - missing args",
			method:      "registerHook",
			args:        []engine.ScriptValue{sv("test-hook")},
			expectError: true,
		},
		{
			name:        "valid listHooks",
			method:      "listHooks",
			args:        []engine.ScriptValue{},
			expectError: false,
		},
		{
			name:   "valid executeHooks",
			method: "executeHooks",
			args: []engine.ScriptValue{
				sv("beforeGenerate"),
				svMap(map[string]interface{}{}),
			},
			expectError: false,
		},
		{
			name:        "unknown method",
			method:      "unknownMethod",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHooksBridge_ExecuteMethod_RegisterHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test registerHook
	hookID := "test-hook"
	hookDef := map[string]interface{}{
		"priority":       10,
		"beforeGenerate": engine.NewFunctionValue("beforeGenerate", func(ctx interface{}, messages interface{}) {}),
	}

	args := []engine.ScriptValue{
		sv(hookID),
		svMap(hookDef),
	}

	result, err := bridge.ExecuteMethod(ctx, "registerHook", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (hook ID) from registerHook")
	assert.Equal(t, hookID, stringValue.Value(), "Hook ID should match input")
}

func TestHooksBridge_ExecuteMethod_ListHooks(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listHooks - should work even with no hooks
	result, err := bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listHooks")
	assert.Equal(t, 0, len(arrayValue.ToGo().([]interface{})), "Expected empty array initially")

	// Register a hook
	registerArgs := []engine.ScriptValue{
		sv("test-hook"),
		svMap(map[string]interface{}{
			"priority": 5,
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// List hooks again
	result, err = bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok = result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listHooks")
	assert.Equal(t, 1, len(arrayValue.ToGo().([]interface{})), "Expected one hook")
}

func TestHooksBridge_ExecuteMethod_EnableDisableHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookID := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		sv(hookID),
		svMap(map[string]interface{}{
			"priority": 5,
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// Test disableHook
	args := []engine.ScriptValue{sv(hookID)}
	result, err := bridge.ExecuteMethod(ctx, "disableHook", args)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from disableHook")
	assert.True(t, boolValue.Value(), "Should successfully disable hook")

	// Test enableHook
	result, err = bridge.ExecuteMethod(ctx, "enableHook", args)
	assert.NoError(t, err)

	boolValue, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from enableHook")
	assert.True(t, boolValue.Value(), "Should successfully enable hook")
}

func TestHooksBridge_ExecuteMethod_ExecuteHooks(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookID := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		sv(hookID),
		svMap(map[string]interface{}{
			"priority":       5,
			"beforeGenerate": engine.NewFunctionValue("beforeGenerate", func(ctx interface{}, messages interface{}) {}),
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// Execute hooks of a specific type
	executeArgs := []engine.ScriptValue{
		sv("beforeGenerate"),
		svMap(map[string]interface{}{
			"messages": svArray(),
		}),
	}

	result, err := bridge.ExecuteMethod(ctx, "executeHooks", executeArgs)
	assert.NoError(t, err)

	// Should return success
	assert.NotNil(t, result)
}

func TestHooksBridge_ExecuteMethod_UnregisterHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookID := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		sv(hookID),
		svMap(map[string]interface{}{
			"priority": 5,
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// List hooks to verify it exists
	result, err := bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	require.NoError(t, err)
	hooks := result.(engine.ArrayValue).ToGo().([]interface{})
	assert.Equal(t, 1, len(hooks), "Should have one hook")

	// Unregister the hook
	unregisterArgs := []engine.ScriptValue{sv(hookID)}
	result, err = bridge.ExecuteMethod(ctx, "unregisterHook", unregisterArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from unregisterHook")
	assert.True(t, boolValue.Value(), "Unregister should succeed")

	// List hooks to verify it's removed
	result, err = bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	require.NoError(t, err)
	hooks = result.(engine.ArrayValue).ToGo().([]interface{})
	assert.Equal(t, 0, len(hooks), "Should have no hooks after unregister")
}

func TestHooksBridge_ExecuteMethod_ClearHooks(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		sv("test-hook"),
		svMap(map[string]interface{}{
			"priority": 5,
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// Verify hook exists
	listResult, err := bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	require.NoError(t, err)
	arrayValue := listResult.(engine.ArrayValue)
	assert.Equal(t, 1, len(arrayValue.ToGo().([]interface{})))

	// Clear all hooks
	result, err := bridge.ExecuteMethod(ctx, "clearHooks", []engine.ScriptValue{})
	assert.NoError(t, err)

	// clearHooks returns the number of hooks cleared
	numberValue, ok := result.(engine.NumberValue)
	assert.True(t, ok, "Expected NumberValue from clearHooks")
	assert.Equal(t, float64(1), numberValue.Value(), "Should have cleared 1 hook")

	// Verify hooks are cleared
	listResult, err = bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	require.NoError(t, err)
	arrayValue = listResult.(engine.ArrayValue)
	assert.Equal(t, 0, len(arrayValue.ToGo().([]interface{})))
}

func TestHooksBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.Error(t, err) // ExecuteMethod returns Go error for unknown method
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "method not found")
}

func TestHooksBridge_RequiredPermissions(t *testing.T) {
	bridge := NewHooksBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasHooksPermission := false
	for _, perm := range permissions {
		if perm.Resource == "hook" {
			hasHooksPermission = true
			break
		}
	}
	assert.True(t, hasHooksPermission, "Should have hooks permission")
}

func TestHooksBridge_TypeMappings(t *testing.T) {
	bridge := NewHooksBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"Hook", "HookInfo", "HookType", "HookContext"}
	for _, expectedType := range expectedTypes {
		_, exists := mappings[expectedType]
		assert.True(t, exists, "Expected type mapping for %s", expectedType)
	}
}

func TestHooksBridge_Cleanup(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestHooksBridge_NotInitialized(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()

	// Should fail when not initialized
	result, err := bridge.ExecuteMethod(ctx, "listHooks", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
