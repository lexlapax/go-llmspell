// ABOUTME: Test suite for State Manager Bridge with flexible type conversion system
// ABOUTME: Tests the ScriptValue to Go type conversion and bridge method execution

package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

func TestStateManagerBridge_Creation(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)
	assert.NotNil(t, bridge)
	assert.Equal(t, "state_manager", bridge.Name())
}

func TestStateManagerBridge_CreateState(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test createState with no arguments
	result, err := bridge.ExecuteMethod(ctx, "createState", []types.ScriptValue{})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestStateManagerBridge_ApplyTransform_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test applyTransform with string name using ToGo() conversion
	transformName := types.ConvertToScriptValue("test_transform")
	stateData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	stateValue := types.ConvertToScriptValue(stateData)

	// This should work with flexible conversion
	_, err = bridge.ExecuteMethod(ctx, "applyTransform", []types.ScriptValue{
		transformName,
		stateValue,
	})
	// We expect this to fail because the transform doesn't exist, but not because of type conversion
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "name must be string")
}

func TestStateManagerBridge_RegisterTransform_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test registerTransform with string name using ToGo() conversion
	transformName := types.ConvertToScriptValue("test_transform")
	transformFunc := types.ConvertToScriptValue(func(ctx context.Context, state interface{}) (interface{}, error) {
		return state, nil
	})

	// This should work with flexible conversion
	_, err = bridge.ExecuteMethod(ctx, "registerTransform", []types.ScriptValue{
		transformName,
		transformFunc,
	})
	// We don't expect type conversion errors
	if err != nil {
		assert.NotContains(t, err.Error(), "name must be string")
	}
}

func TestStateManagerBridge_SetMetadata_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a state first
	stateResult, err := bridge.ExecuteMethod(ctx, "createState", []types.ScriptValue{})
	require.NoError(t, err)

	// Test setMetadata with flexible conversion
	stateValue := types.ConvertToScriptValue(stateResult.ToGo())
	keyValue := types.ConvertToScriptValue("test_key")
	metadataValue := types.ConvertToScriptValue("test_metadata")

	_, err = bridge.ExecuteMethod(ctx, "setMetadata", []types.ScriptValue{
		stateValue,
		keyValue,
		metadataValue,
	})
	// We don't expect type conversion errors
	if err != nil {
		assert.NotContains(t, err.Error(), "key must be string")
		assert.NotContains(t, err.Error(), "state must be object")
	}
}

func TestStateManagerBridge_Messages_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a state first
	stateResult, err := bridge.ExecuteMethod(ctx, "createState", []types.ScriptValue{})
	require.NoError(t, err)

	// Test messages with flexible conversion
	stateValue := types.ConvertToScriptValue(stateResult.ToGo())

	_, err = bridge.ExecuteMethod(ctx, "messages", []types.ScriptValue{
		stateValue,
	})
	// We don't expect type conversion errors
	if err != nil {
		assert.NotContains(t, err.Error(), "state must be object")
	}
}

func TestStateManagerBridge_AddMessage_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a state first
	stateResult, err := bridge.ExecuteMethod(ctx, "createState", []types.ScriptValue{})
	require.NoError(t, err)

	// Test addMessage with flexible conversion
	stateValue := types.ConvertToScriptValue(stateResult.ToGo())
	messageData := map[string]interface{}{
		"role":    "user",
		"content": "test message",
	}
	messageValue := types.ConvertToScriptValue(messageData)

	_, err = bridge.ExecuteMethod(ctx, "addMessage", []types.ScriptValue{
		stateValue,
		messageValue,
	})
	// We don't expect type conversion errors
	if err != nil {
		assert.NotContains(t, err.Error(), "state must be object")
		assert.NotContains(t, err.Error(), "message must be object")
	}
}

func TestStateManagerBridge_RegisterValidator_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test registerValidator with flexible conversion
	validatorName := types.ConvertToScriptValue("test_validator")
	validatorFunc := types.ConvertToScriptValue(func(state interface{}) error {
		return nil
	})

	_, err = bridge.ExecuteMethod(ctx, "registerValidator", []types.ScriptValue{
		validatorName,
		validatorFunc,
	})
	// We don't expect type conversion errors
	if err != nil {
		assert.NotContains(t, err.Error(), "name must be string")
	}
}

func TestStateManagerBridge_ValidateState_FlexibleConversion(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a state first
	stateResult, err := bridge.ExecuteMethod(ctx, "createState", []types.ScriptValue{})
	require.NoError(t, err)

	// Test validateState with flexible conversion
	validatorName := types.ConvertToScriptValue("test_validator")
	stateValue := types.ConvertToScriptValue(stateResult.ToGo())

	_, err = bridge.ExecuteMethod(ctx, "validateState", []types.ScriptValue{
		validatorName,
		stateValue,
	})
	// We expect this to fail because validator doesn't exist, but not due to type conversion
	if err != nil {
		assert.NotContains(t, err.Error(), "name must be string")
		assert.NotContains(t, err.Error(), "state must be object")
	}
}

func TestStateManagerBridge_ErrorHandling(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	ctx := context.Background()
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test unknown method
	_, err = bridge.ExecuteMethod(ctx, "unknownMethod", []types.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method not found")

	// Test method with insufficient arguments
	_, err = bridge.ExecuteMethod(ctx, "applyTransform", []types.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires name and state parameters")
}

func TestStateManagerBridge_InterfaceCompliance(t *testing.T) {
	manager := core.NewStateManager()
	bridge, err := NewStateManagerBridge(manager)
	require.NoError(t, err)

	// Test that bridge implements required interfaces
	assert.Implements(t, (*types.Bridge)(nil), bridge)

	// Test methods list
	methods := bridge.Methods()
	assert.Greater(t, len(methods), 0)

	expectedMethods := []string{
		"createState", "saveState", "loadState", "deleteState",
		"registerTransform", "applyTransform",
		"registerValidator", "validateState",
		"setMetadata", "getMetadata",
		"addMessage", "messages",
	}

	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodNames[expected], "Expected method %s not found", expected)
	}
}
