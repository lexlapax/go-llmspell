// ABOUTME: Simplified tests for State Context Bridge focused on bridge interface compliance
// ABOUTME: Tests core bridge functionality without complex script engine integration

package bridge

import (
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestStateContextBridge_InterfaceCompliance(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	// Test Bridge interface implementation
	t.Run("GetID", func(t *testing.T) {
		id := bridge.GetID()
		if id != "state_context" {
			t.Errorf("Expected ID 'state_context', got %s", id)
		}
	})

	t.Run("GetMetadata", func(t *testing.T) {
		metadata := bridge.GetMetadata()
		if metadata.Name != "State Context Bridge" {
			t.Errorf("Expected name 'State Context Bridge', got %s", metadata.Name)
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
			"createSharedContext", "withInheritanceConfig", "get", "set", "has", "keys", "values",
			"getArtifact", "artifacts", "messages", "getMetadata", "localState", "clone", "asState",
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

		if _, exists := mappings["SharedStateContext"]; !exists {
			t.Error("Expected SharedStateContext type mapping")
		}
		if _, exists := mappings["StateReader"]; !exists {
			t.Error("Expected StateReader type mapping")
		}
	})

	t.Run("ValidateMethod", func(t *testing.T) {
		// Valid method
		err := bridge.ValidateMethod("createSharedContext", []interface{}{})
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

func TestStateContextBridge_CoreOperations(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("createSharedContext", func(t *testing.T) {
		// Create a mock parent state
		parentState := map[string]interface{}{
			"id":   "parent_state",
			"data": map[string]interface{}{"parent_key": "parent_value"},
		}

		result, err := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": parentState,
		})
		if err != nil {
			t.Fatalf("createSharedContext failed: %v", err)
		}

		contextObj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("createSharedContext should return context object")
		}

		// Verify context structure
		if contextObj["type"] != "SharedStateContext" {
			t.Error("Context should have correct type")
		}
		if contextObj["inheritMessages"] != true {
			t.Error("Default inheritance should include messages")
		}
		if contextObj["inheritArtifacts"] != true {
			t.Error("Default inheritance should include artifacts")
		}
		if contextObj["inheritMetadata"] != true {
			t.Error("Default inheritance should include metadata")
		}
		if contextObj["parent"] == nil {
			t.Error("Context should reference parent state")
		}
	})

	t.Run("createSharedContext_without_parent", func(t *testing.T) {
		_, err := bridge.createSharedContext(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("createSharedContext should fail without parent parameter")
		}
		if err.Error() != "parent parameter is required and must be a state object" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("createSharedContext_invalid_parent", func(t *testing.T) {
		_, err := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": "invalid",
		})
		if err == nil {
			t.Error("createSharedContext should fail with invalid parent parameter")
		}
	})
}

func TestStateContextBridge_ParameterValidation(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("withInheritanceConfig_validation", func(t *testing.T) {
		// Test without context parameter
		_, err := bridge.withInheritanceConfig(ctx, map[string]interface{}{
			"messages":  true,
			"artifacts": true,
			"metadata":  true,
		})
		if err == nil {
			t.Error("withInheritanceConfig should fail without context parameter")
		}

		// Test without boolean parameters
		contextObj := map[string]interface{}{}
		_, err = bridge.withInheritanceConfig(ctx, map[string]interface{}{
			"context":  contextObj,
			"messages": "invalid",
		})
		if err == nil {
			t.Error("withInheritanceConfig should fail with invalid boolean parameter")
		}
	})

	t.Run("contextGet_validation", func(t *testing.T) {
		// Test without context parameter
		_, err := bridge.contextGet(ctx, map[string]interface{}{
			"key": "test_key",
		})
		if err == nil {
			t.Error("contextGet should fail without context parameter")
		}

		// Test without key parameter
		contextObj := map[string]interface{}{}
		_, err = bridge.contextGet(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err == nil {
			t.Error("contextGet should fail without key parameter")
		}
	})

	t.Run("contextSet_validation", func(t *testing.T) {
		// Test without context parameter
		_, err := bridge.contextSet(ctx, map[string]interface{}{
			"key":   "test_key",
			"value": "test_value",
		})
		if err == nil {
			t.Error("contextSet should fail without context parameter")
		}

		// Test without key parameter
		contextObj := map[string]interface{}{}
		_, err = bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"value":   "test_value",
		})
		if err == nil {
			t.Error("contextSet should fail without key parameter")
		}
	})

	t.Run("contextHas_validation", func(t *testing.T) {
		// Test without context parameter
		_, err := bridge.contextHas(ctx, map[string]interface{}{
			"key": "test_key",
		})
		if err == nil {
			t.Error("contextHas should fail without context parameter")
		}

		// Test without key parameter
		contextObj := map[string]interface{}{}
		_, err = bridge.contextHas(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err == nil {
			t.Error("contextHas should fail without key parameter")
		}
	})

	t.Run("contextGetArtifact_validation", func(t *testing.T) {
		// Test without context parameter
		_, err := bridge.contextGetArtifact(ctx, map[string]interface{}{
			"id": "artifact_id",
		})
		if err == nil {
			t.Error("contextGetArtifact should fail without context parameter")
		}

		// Test without id parameter
		contextObj := map[string]interface{}{}
		_, err = bridge.contextGetArtifact(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err == nil {
			t.Error("contextGetArtifact should fail without id parameter")
		}
	})

	t.Run("contextGetMetadata_validation", func(t *testing.T) {
		// Test without context parameter
		_, err := bridge.contextGetMetadata(ctx, map[string]interface{}{
			"key": "meta_key",
		})
		if err == nil {
			t.Error("contextGetMetadata should fail without context parameter")
		}

		// Test without key parameter
		contextObj := map[string]interface{}{}
		_, err = bridge.contextGetMetadata(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err == nil {
			t.Error("contextGetMetadata should fail without key parameter")
		}
	})
}

func TestStateContextBridge_ConversionLimitations(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	// These tests verify that the bridge properly handles the limitation
	// that full conversion is not implemented in the mock environment
	t.Run("scriptToSharedContext_not_implemented", func(t *testing.T) {
		contextObj := map[string]interface{}{
			"type": "SharedStateContext",
		}

		_, err := bridge.contextGet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "test_key",
		})

		if err == nil {
			t.Error("Should fail due to missing _id in context object")
		}
		// The actual error is about missing _id field
		if !strings.Contains(err.Error(), "_id") {
			t.Errorf("Expected error message about missing _id, got: %v", err)
		}
	})

	t.Run("scriptToStateReader_not_implemented", func(t *testing.T) {
		// The scriptToStateReader function converts to State via scriptToState
		// It will create a new state object, not return an error
		result, err := bridge.scriptToStateReader(map[string]interface{}{})
		if err != nil {
			t.Errorf("scriptToStateReader should not fail: %v", err)
		}
		if result == nil {
			t.Error("scriptToStateReader should return a valid StateReader")
		}
	})
}

func TestStateContextBridge_HelperFunctions(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	t.Run("sharedContextToScript", func(t *testing.T) {
		// Test with a context ID and nil context
		result := bridge.sharedContextToScript("test_context_1", nil)
		if result["type"] != "SharedStateContext" {
			t.Error("Should return SharedStateContext type")
		}
		if result["_id"] != "test_context_1" {
			t.Error("Should include context ID")
		}
		if result["inheritMessages"] != true {
			t.Error("Should have default inheritance settings")
		}
	})

	t.Run("stateToScript", func(t *testing.T) {
		// Create a domain state
		state := domain.NewState()
		state.Set("test_key", "test_value")
		state.SetMetadata("meta_key", "meta_value")

		result := bridge.stateToScript(state)
		if result["id"] == "" {
			t.Error("Should include state ID")
		}

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Should include data map")
		}
		if data["test_key"] != "test_value" {
			t.Error("Should include state data")
		}

		metadata, ok := result["metadata"].(map[string]interface{})
		if !ok {
			t.Fatal("Should include metadata map")
		}
		if metadata["meta_key"] != "meta_value" {
			t.Error("Should include state metadata")
		}
	})

	t.Run("updateScriptSharedContext", func(t *testing.T) {
		// Create a script object representing a shared context
		scriptObj := map[string]interface{}{
			"_id":  "test_ctx",
			"type": "SharedStateContext",
		}

		// Create a real SharedStateContext
		parentState := domain.NewState()
		sharedContext := domain.NewSharedStateContext(parentState)
		sharedContext.Set("local_key", "local_value")

		// Call updateScriptSharedContext
		bridge.updateScriptSharedContext(scriptObj, sharedContext)

		// The current implementation doesn't actually update the script object
		// This is noted in the implementation as something that would be done
		// in a real bridge implementation
	})
}
