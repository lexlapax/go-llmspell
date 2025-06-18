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
	assert.Equal(t, "2.1.0", metadata.Version)
	assert.Contains(t, metadata.Description, "hook system")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestHooksBridge_Methods(t *testing.T) {
	bridge := NewHooksBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"registerHook", "unregisterHook", "executeHook", "listHooks",
		"hasHook", "clearHooks", "getHookPriority", "setHookPriority",
		"enableHook", "disableHook", "isHookEnabled", "getHookMetadata",
		"setHookMetadata", "getHookStats", "resetHookStats", "createHookChain",
		"executeHookChain", "getHookChain", "removeHookChain", "listHookChains",
		"validateHook", "getHookDependencies", "setHookDependencies",
		"resolveHookDependencies", "getHookExecutionOrder", "executeConditionalHook",
		"createHookGroup", "addHookToGroup", "removeHookFromGroup", "executeHookGroup",
		"getHookGroup", "listHookGroups", "removeHookGroup",
	}

	assert.GreaterOrEqual(t, len(methods), len(expectedMethods))

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
			name:        "valid registerHook",
			method:      "registerHook",
			args:        []engine.ScriptValue{engine.NewStringValue("test-hook"), engine.NewStringValue("before"), engine.NewStringValue("handler")},
			expectError: false,
		},
		{
			name:        "invalid registerHook - missing args",
			method:      "registerHook",
			args:        []engine.ScriptValue{engine.NewStringValue("test-hook")},
			expectError: true,
		},
		{
			name:        "valid listHooks",
			method:      "listHooks",
			args:        []engine.ScriptValue{},
			expectError: false,
		},
		{
			name:        "valid executeHook",
			method:      "executeHook",
			args:        []engine.ScriptValue{engine.NewStringValue("test-hook"), engine.NewObjectValue(map[string]engine.ScriptValue{})},
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
	hookName := "test-hook"
	hookType := "before"
	handler := "test-handler-function"

	args := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewStringValue(hookType),
		engine.NewStringValue(handler),
	}

	result, err := bridge.ExecuteMethod(ctx, "registerHook", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (hook ID) from registerHook")
	assert.NotEmpty(t, stringValue.Value(), "Hook ID should not be empty")
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
		engine.NewStringValue("test-hook"),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
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

func TestHooksBridge_ExecuteMethod_HasHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookName := "test-hook"

	// Test hasHook for non-existent hook
	args := []engine.ScriptValue{engine.NewStringValue(hookName)}
	result, err := bridge.ExecuteMethod(ctx, "hasHook", args)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from hasHook")
	assert.False(t, boolValue.Value(), "Should not have hook initially")

	// Register the hook
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// Test hasHook for existing hook
	result, err = bridge.ExecuteMethod(ctx, "hasHook", args)
	assert.NoError(t, err)

	boolValue, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from hasHook")
	assert.True(t, boolValue.Value(), "Should have hook after registration")
}

func TestHooksBridge_ExecuteMethod_ExecuteHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookName := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
	}
	_, err = bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	// Execute the hook
	executeArgs := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"data": engine.NewStringValue("test"),
		}),
	}

	result, err := bridge.ExecuteMethod(ctx, "executeHook", executeArgs)
	assert.NoError(t, err)

	// Should return the hook result
	assert.NotNil(t, result)
}

func TestHooksBridge_ExecuteMethod_UnregisterHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookName := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
	}
	hookID, err := bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	hookIDStr := hookID.(engine.StringValue).Value()

	// Verify hook exists
	hasArgs := []engine.ScriptValue{engine.NewStringValue(hookName)}
	result, err := bridge.ExecuteMethod(ctx, "hasHook", hasArgs)
	require.NoError(t, err)
	assert.True(t, result.(engine.BoolValue).Value())

	// Unregister the hook
	unregisterArgs := []engine.ScriptValue{engine.NewStringValue(hookIDStr)}
	result, err = bridge.ExecuteMethod(ctx, "unregisterHook", unregisterArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from unregisterHook")
	assert.True(t, boolValue.Value(), "Unregister should succeed")

	// Verify hook is removed
	result, err = bridge.ExecuteMethod(ctx, "hasHook", hasArgs)
	require.NoError(t, err)
	assert.False(t, result.(engine.BoolValue).Value(), "Hook should be removed")
}

func TestHooksBridge_ExecuteMethod_SetGetHookPriority(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookName := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
	}
	hookID, err := bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	hookIDStr := hookID.(engine.StringValue).Value()

	// Set hook priority
	setPriorityArgs := []engine.ScriptValue{
		engine.NewStringValue(hookIDStr),
		engine.NewNumberValue(10),
	}
	result, err := bridge.ExecuteMethod(ctx, "setHookPriority", setPriorityArgs)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from setHookPriority")

	// Get hook priority
	getPriorityArgs := []engine.ScriptValue{engine.NewStringValue(hookIDStr)}
	result, err = bridge.ExecuteMethod(ctx, "getHookPriority", getPriorityArgs)
	assert.NoError(t, err)

	numberValue, ok := result.(engine.NumberValue)
	assert.True(t, ok, "Expected NumberValue from getHookPriority")
	assert.Equal(t, float64(10), numberValue.Value(), "Priority should be 10")
}

func TestHooksBridge_ExecuteMethod_EnableDisableHook(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	hookName := "test-hook"

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue(hookName),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
	}
	hookID, err := bridge.ExecuteMethod(ctx, "registerHook", registerArgs)
	require.NoError(t, err)

	hookIDStr := hookID.(engine.StringValue).Value()

	// Check if hook is enabled (should be by default)
	isEnabledArgs := []engine.ScriptValue{engine.NewStringValue(hookIDStr)}
	result, err := bridge.ExecuteMethod(ctx, "isHookEnabled", isEnabledArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from isHookEnabled")
	assert.True(t, boolValue.Value(), "Hook should be enabled by default")

	// Disable the hook
	disableArgs := []engine.ScriptValue{engine.NewStringValue(hookIDStr)}
	result, err = bridge.ExecuteMethod(ctx, "disableHook", disableArgs)
	assert.NoError(t, err)

	_, ok = result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from disableHook")

	// Check if hook is disabled
	result, err = bridge.ExecuteMethod(ctx, "isHookEnabled", isEnabledArgs)
	assert.NoError(t, err)

	boolValue, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from isHookEnabled")
	assert.False(t, boolValue.Value(), "Hook should be disabled")

	// Enable the hook
	enableArgs := []engine.ScriptValue{engine.NewStringValue(hookIDStr)}
	result, err = bridge.ExecuteMethod(ctx, "enableHook", enableArgs)
	assert.NoError(t, err)

	_, ok = result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from enableHook")

	// Check if hook is enabled again
	result, err = bridge.ExecuteMethod(ctx, "isHookEnabled", isEnabledArgs)
	assert.NoError(t, err)

	boolValue, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from isHookEnabled")
	assert.True(t, boolValue.Value(), "Hook should be enabled again")
}

func TestHooksBridge_ExecuteMethod_GetHookStats(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get hook stats
	result, err := bridge.ExecuteMethod(ctx, "getHookStats", []engine.ScriptValue{})
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getHookStats")

	stats := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, stats, "total_hooks")
	assert.Contains(t, stats, "enabled_hooks")
}

func TestHooksBridge_ExecuteMethod_ClearHooks(t *testing.T) {
	bridge := NewHooksBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register a hook first
	registerArgs := []engine.ScriptValue{
		engine.NewStringValue("test-hook"),
		engine.NewStringValue("before"),
		engine.NewStringValue("handler"),
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

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from clearHooks")

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
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for unknown method")
	assert.Contains(t, errorValue.Error().Error(), "unknown method")
}

func TestHooksBridge_RequiredPermissions(t *testing.T) {
	bridge := NewHooksBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasHooksPermission := false
	for _, perm := range permissions {
		if perm.Resource == "hooks" {
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
	expectedTypes := []string{"Hook", "HookChain", "HookGroup"}
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
