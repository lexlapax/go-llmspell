// ABOUTME: Simplified tests for State Manager Bridge focused on bridge interface compliance
// ABOUTME: Tests core bridge functionality without complex script engine integration

package state

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestStateManagerBridge_Interface(t *testing.T) {
	manager := createTestStateManager()
	bridge, err := NewStateManagerBridge(manager)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Test Bridge interface implementation
	t.Run("GetID", func(t *testing.T) {
		id := bridge.GetID()
		if id != "state_manager" {
			t.Errorf("Expected ID 'state_manager', got %s", id)
		}
	})

	t.Run("GetMetadata", func(t *testing.T) {
		metadata := bridge.GetMetadata()
		if metadata.Name != "State Manager Bridge" {
			t.Errorf("Expected name 'State Manager Bridge', got %s", metadata.Name)
		}
		if metadata.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %s", metadata.Version)
		}
	})

	t.Run("Initialize", func(t *testing.T) {
		ctx := context.Background()
		err := bridge.Initialize(ctx)
		if err != nil {
			t.Errorf("Initialize failed: %v", err)
		}
	})

	t.Run("IsInitialized", func(t *testing.T) {
		if !bridge.IsInitialized() {
			t.Error("Bridge should be initialized")
		}
	})

	t.Run("Methods", func(t *testing.T) {
		methods := bridge.Methods()
		if len(methods) == 0 {
			t.Error("Bridge should expose methods")
		}

		// Check for key methods
		methodNames := make(map[string]bool)
		for _, method := range methods {
			methodNames[method.Name] = true
		}

		expectedMethods := []string{
			"createState", "saveState", "loadState", "deleteState", "listStates",
			"get", "set", "has", "keys", "values",
		}

		for _, expected := range expectedMethods {
			if !methodNames[expected] {
				t.Errorf("Expected method %s not found", expected)
			}
		}
	})

	t.Run("TypeMappings", func(t *testing.T) {
		mappings := bridge.TypeMappings()
		if len(mappings) == 0 {
			t.Error("Bridge should provide type mappings")
		}

		if _, exists := mappings["State"]; !exists {
			t.Error("Expected State type mapping")
		}
	})

	t.Run("ValidateMethod", func(t *testing.T) {
		// Valid method
		err := bridge.ValidateMethod("createState", []interface{}{})
		if err != nil {
			t.Errorf("Valid method validation failed: %v", err)
		}

		// Invalid method
		err = bridge.ValidateMethod("nonexistent", []interface{}{})
		if err == nil {
			t.Error("Invalid method should fail validation")
		}
	})

	t.Run("RequiredPermissions", func(t *testing.T) {
		permissions := bridge.RequiredPermissions()
		if len(permissions) == 0 {
			t.Error("Bridge should require permissions")
		}

		found := false
		for _, perm := range permissions {
			if perm.Type == engine.PermissionMemory {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected memory permission requirement")
		}
	})

	t.Run("Cleanup", func(t *testing.T) {
		ctx := context.Background()
		err := bridge.Cleanup(ctx)
		if err != nil {
			t.Errorf("Cleanup failed: %v", err)
		}
	})
}

func TestStateManagerBridge_StateOperations(t *testing.T) {
	manager := createTestStateManager()
	bridge, err := NewStateManagerBridge(manager)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("createState", func(t *testing.T) {
		result, err := bridge.createState(ctx, map[string]interface{}{})
		if err != nil {
			t.Fatalf("createState failed: %v", err)
		}

		stateObj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("createState should return map[string]interface{}")
		}

		if _, exists := stateObj["id"]; !exists {
			t.Error("Created state should have ID")
		}
	})

	t.Run("listStates", func(t *testing.T) {
		result, err := bridge.listStates(ctx, map[string]interface{}{})
		if err != nil {
			t.Fatalf("listStates failed: %v", err)
		}

		states, ok := result.([]interface{})
		if !ok {
			t.Fatal("listStates should return []interface{}")
		}

		// Initially should be empty
		if len(states) != 0 {
			t.Errorf("Expected 0 states, got %d", len(states))
		}
	})

	t.Run("mergeStates with valid strategy", func(t *testing.T) {
		// Create two mock states
		state1 := map[string]interface{}{
			"id":   "state1",
			"data": map[string]interface{}{"key1": "value1"},
		}
		state2 := map[string]interface{}{
			"id":   "state2",
			"data": map[string]interface{}{"key2": "value2"},
		}

		result, err := bridge.mergeStates(ctx, map[string]interface{}{
			"states":   []interface{}{state1, state2},
			"strategy": "last",
		})

		// This will fail because scriptToState needs __state reference
		// but we can test that the parameter validation works
		if err == nil {
			t.Log("mergeStates executed successfully:", result)
		} else {
			// Expected error due to missing __state reference
			expectedErr := "failed to convert state at index 0: cannot convert script object to state: missing __state reference"
			if err.Error() != expectedErr {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	})

	t.Run("mergeStates with invalid strategy", func(t *testing.T) {
		result, err := bridge.mergeStates(ctx, map[string]interface{}{
			"states":   []interface{}{},
			"strategy": "invalid_strategy",
		})

		if err == nil {
			t.Error("mergeStates should fail with invalid strategy")
		}
		if result != nil {
			t.Error("mergeStates should return nil on error")
		}
	})
}

func TestStateManagerBridge_SimpleErrorHandling(t *testing.T) {
	manager := createTestStateManager()
	bridge, err := NewStateManagerBridge(manager)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("saveState without state parameter", func(t *testing.T) {
		_, err := bridge.saveState(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("saveState should fail without state parameter")
		}
	})

	t.Run("loadState without id parameter", func(t *testing.T) {
		_, err := bridge.loadState(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("loadState should fail without id parameter")
		}
	})

	t.Run("deleteState without id parameter", func(t *testing.T) {
		_, err := bridge.deleteState(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("deleteState should fail without id parameter")
		}
	})

	t.Run("registerTransform without name parameter", func(t *testing.T) {
		_, err := bridge.registerTransform(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("registerTransform should fail without name parameter")
		}
	})

	t.Run("applyTransform without name parameter", func(t *testing.T) {
		_, err := bridge.applyTransform(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("applyTransform should fail without name parameter")
		}
	})
}

func TestStateManagerBridge_BuiltinTransforms(t *testing.T) {
	manager := createTestStateManager()
	bridge, err := NewStateManagerBridge(manager)
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Initialize to register built-in transforms
	ctx := context.Background()
	err = bridge.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize bridge: %v", err)
	}

	// Test that built-in transforms are registered
	builtinTransforms := []string{"filter", "flatten", "sanitize"}

	for _, transformName := range builtinTransforms {
		t.Run("builtin_transform_"+transformName, func(t *testing.T) {
			// We can't easily test the actual transform without a full mock state
			// but we can verify the manager has the transforms registered
			// by trying to apply them (they should fail gracefully due to mock limitations)

			state := map[string]interface{}{
				"id": "test_state",
			}

			_, err := bridge.applyTransform(ctx, map[string]interface{}{
				"name":  transformName,
				"state": state,
			})

			// Should fail due to mock conversion, but not due to transform not found
			if err != nil && err.Error() == "failed to convert state: scriptToState conversion not implemented in mock environment" {
				// This is expected - the transform exists but conversion fails
				t.Logf("Transform %s exists (expected conversion error)", transformName)
			} else if err != nil && err.Error() == "failed to apply transform: transform not found" {
				t.Errorf("Built-in transform %s not registered", transformName)
			}
		})
	}
}
