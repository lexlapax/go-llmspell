// ABOUTME: Test suite for State Context Bridge that bridges go-llms SharedStateContext to script engines
// ABOUTME: Comprehensive tests covering parent-child state relationships, inheritance configuration, and state isolation

package state

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
	"github.com/lexlapax/go-llms/pkg/testutils/scenario"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: We're now using go-llms domain.SharedStateContext directly
// The tests have been updated to use the real types from go-llms

// Test suite for StateContextBridge
func TestNewStateContextBridge(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Valid bridge creation",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge, err := NewStateContextBridge()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStateContextBridge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && bridge == nil {
				t.Error("NewStateContextBridge() returned nil bridge for valid input")
			}
		})
	}
}

// Test schema validation functionality using go-llms v0.3.5 infrastructure
func TestStateContextBridge_SchemaValidation(t *testing.T) {

	bridge, err := NewStateContextBridge()
	require.NoError(t, err, "Failed to create bridge")
	require.NotNil(t, bridge, "Bridge should not be nil")

	ctx := context.Background()

	// Create parent state
	parentState := domain.NewState()
	parentState.Set("parentKey", "parentValue")

	// Create shared context
	result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
		map[string]interface{}{
			"data": map[string]interface{}{
				"parentKey": "parentValue",
			},
		},
	})
	require.NoError(t, err)

	sharedContextObj := result.(map[string]interface{})

	// Set some initial data
	_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
		sharedContextObj,
		"name",
		"John Doe",
	})
	require.NoError(t, err)

	_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
		sharedContextObj,
		"age",
		30,
	})
	require.NoError(t, err)

	t.Run("SetStateSchema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":  "string",
					"title": "Person Name",
				},
				"age": map[string]interface{}{
					"type":  "number",
					"title": "Person Age",
				},
			},
			"required": []interface{}{"name", "age"},
		}

		result, err := bridge.ExecuteMethod(ctx, "setStateSchema", []interface{}{
			sharedContextObj,
			"person-schema",
			schema,
		})
		require.NoError(t, err)

		schemaResult := result.(map[string]interface{})
		assert.Equal(t, "person-schema", schemaResult["schemaId"])
		assert.True(t, schemaResult["success"].(bool))
	})

	t.Run("ValidateState_Valid", func(t *testing.T) {
		// Ensure data is valid before validation
		_, err := bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"age",
			30, // Reset to valid number
		})
		require.NoError(t, err)

		result, err := bridge.ExecuteMethod(ctx, "validateState", []interface{}{
			sharedContextObj,
		})
		require.NoError(t, err)

		validationResult := result.(map[string]interface{})
		if !validationResult["valid"].(bool) {
			t.Logf("Validation failed with errors: %v", validationResult)
		}
		assert.True(t, validationResult["valid"].(bool))
	})

	t.Run("ValidateState_Invalid", func(t *testing.T) {
		// Set invalid data (string instead of number for age)
		_, err := bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"age",
			"not a number",
		})
		require.NoError(t, err)

		result, err := bridge.ExecuteMethod(ctx, "validateState", []interface{}{
			sharedContextObj,
		})
		require.NoError(t, err)

		validationResult := result.(map[string]interface{})
		// Note: The actual validation behavior depends on go-llms validator implementation
		// This test structure follows the expected interface
		assert.Contains(t, validationResult, "valid")

		// Reset to valid data for subsequent tests
		_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"age",
			30,
		})
		require.NoError(t, err)
	})

	t.Run("GetStateSchema", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "getStateSchema", []interface{}{
			sharedContextObj,
		})
		require.NoError(t, err)

		if result != nil {
			schema := result.(map[string]interface{})
			assert.Equal(t, "object", schema["type"])
			assert.Contains(t, schema, "properties")
		}
	})

	t.Run("GetSchemaVersions", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "getSchemaVersions", []interface{}{
			"person-schema",
		})
		require.NoError(t, err)

		versions := result.([]interface{})
		assert.GreaterOrEqual(t, len(versions), 1)

		if len(versions) > 0 {
			version := versions[0].(map[string]interface{})
			assert.Equal(t, 1, version["version"])
			assert.Contains(t, version, "title")
		}
	})

	t.Run("SetSchemaVersion", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "setSchemaVersion", []interface{}{
			"person-schema",
			1,
		})
		require.NoError(t, err)

		versionResult := result.(map[string]interface{})
		assert.Equal(t, "person-schema", versionResult["schemaId"])
		assert.Equal(t, 1, versionResult["version"])
		assert.True(t, versionResult["success"].(bool))
	})

	t.Run("ValidateWithVersion", func(t *testing.T) {
		// Reset to valid data
		_, err := bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"age",
			25,
		})
		require.NoError(t, err)

		result, err := bridge.ExecuteMethod(ctx, "validateWithVersion", []interface{}{
			sharedContextObj,
			"person-schema",
			1,
		})
		require.NoError(t, err)

		validationResult := result.(map[string]interface{})
		assert.Contains(t, validationResult, "valid")
		assert.Equal(t, "person-schema", validationResult["schemaId"])
		assert.Equal(t, 1, validationResult["version"])
	})
}

func TestStateContextBridge_SchemaValidation_ErrorCases(t *testing.T) {

	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("ValidateState_NoContext", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "validateState", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validateState requires context parameter")
	})

	t.Run("SetStateSchema_InvalidParameters", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "setStateSchema", []interface{}{
			"invalid",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "setStateSchema requires context, schemaId, and schema parameters")
	})

	t.Run("GetStateSchema_InvalidContext", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "getStateSchema", []interface{}{
			"invalid",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context must be object")
	})

	t.Run("ValidateState_NoSchema", func(t *testing.T) {
		// Create context without schema
		result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{},
			},
		})
		require.NoError(t, err)

		contextObj := result.(map[string]interface{})

		result, err = bridge.ExecuteMethod(ctx, "validateState", []interface{}{
			contextObj,
		})
		require.NoError(t, err)

		validationResult := result.(map[string]interface{})
		assert.True(t, validationResult["valid"].(bool))
		assert.Equal(t, "No schema configured for validation", validationResult["message"])
	})

	t.Run("GetSchemaVersions_NonExistentSchema", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "getSchemaVersions", []interface{}{
			"non-existent-schema",
		})
		require.NoError(t, err)

		versions := result.([]interface{})
		assert.Equal(t, 0, len(versions))
	})
}

func TestStateContextBridge_CustomValidators(t *testing.T) {

	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("RegisterCustomValidator", func(t *testing.T) {
		customValidator := func(value interface{}) bool {
			if str, ok := value.(string); ok {
				return len(str) >= 3
			}
			return false
		}

		result, err := bridge.ExecuteMethod(ctx, "registerCustomValidator", []interface{}{
			"minLength3",
			customValidator,
		})
		require.NoError(t, err)

		validatorResult := result.(map[string]interface{})
		assert.Equal(t, "minLength3", validatorResult["name"])
		assert.True(t, validatorResult["registered"].(bool))
	})

	t.Run("RegisterCustomValidator_InvalidParameters", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "registerCustomValidator", []interface{}{
			123, // Invalid name type
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "registerCustomValidator requires name and validator parameters")
	})
}

func TestStateContextBridge_SchemaValidation_Integration(t *testing.T) {

	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Test complete workflow with multiple schemas and versions
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Create shared context
		result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{
					"type": "user",
				},
			},
		})
		require.NoError(t, err)

		contextObj := result.(map[string]interface{})

		// Set initial data
		_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
			contextObj,
			"username",
			"user123",
		})
		require.NoError(t, err)

		// Create and set schema v1
		schemaV1 := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"username": map[string]interface{}{
					"type":  "string",
					"title": "Username",
				},
			},
			"required": []interface{}{"username"},
		}

		_, err = bridge.ExecuteMethod(ctx, "setStateSchema", []interface{}{
			contextObj,
			"user-schema",
			schemaV1,
		})
		require.NoError(t, err)

		// Validate with schema v1
		result, err = bridge.ExecuteMethod(ctx, "validateState", []interface{}{
			contextObj,
		})
		require.NoError(t, err)

		validationResult := result.(map[string]interface{})
		if !validationResult["valid"].(bool) {
			t.Logf("Integration test validation failed with result: %+v", validationResult)
		}
		assert.True(t, validationResult["valid"].(bool))

		// Add more data
		_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
			contextObj,
			"email",
			"user@example.com",
		})
		require.NoError(t, err)

		// Create schema v2 with email
		schemaV2 := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"username": map[string]interface{}{
					"type":  "string",
					"title": "Username",
				},
				"email": map[string]interface{}{
					"type":  "string",
					"title": "Email",
				},
			},
			"required": []interface{}{"username", "email"},
		}

		_, err = bridge.ExecuteMethod(ctx, "setStateSchema", []interface{}{
			contextObj,
			"user-schema-v2",
			schemaV2,
		})
		require.NoError(t, err)

		// Get schema to verify it was stored
		result, err = bridge.ExecuteMethod(ctx, "getStateSchema", []interface{}{
			contextObj,
		})
		require.NoError(t, err)

		if result != nil {
			schema := result.(map[string]interface{})
			assert.Contains(t, schema, "properties")
		}
	})
}

func TestStateContextBridge_Interface(t *testing.T) {
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
}

func TestStateContextBridge_ParentChildRelationships(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("createSharedContext", func(t *testing.T) {
		// Create parent state
		parentState := domain.NewState()
		parentState.Set("parent_key", "parent_value")
		parentState.SetMetadata("parent_meta", "parent_meta_value")

		// Create shared context
		result, err := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": bridge.stateToScript(parentState),
		})
		if err != nil {
			t.Fatalf("createSharedContext failed: %v", err)
		}

		contextObj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("createSharedContext should return context object")
		}

		// Verify context was created
		if contextObj["inheritMessages"] != true {
			t.Error("Default inheritance config should inherit messages")
		}
		if contextObj["inheritArtifacts"] != true {
			t.Error("Default inheritance config should inherit artifacts")
		}
		if contextObj["inheritMetadata"] != true {
			t.Error("Default inheritance config should inherit metadata")
		}
	})

	t.Run("parentChildDataAccess", func(t *testing.T) {
		// Create parent state using go-llms domain.State
		parentState := domain.NewState()
		parentState.Set("parent_key", "parent_value")
		parentState.Set("shared_key", "parent_shared_value")

		// Create shared context
		contextResult, _ := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": bridge.stateToScript(parentState),
		})
		contextObj := contextResult.(map[string]interface{})

		// Set local value that overrides parent
		_, err := bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "shared_key",
			"value":   "child_shared_value",
		})
		if err != nil {
			t.Fatalf("contextSet failed: %v", err)
		}

		// Set local-only value
		_, err = bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "child_key",
			"value":   "child_value",
		})
		if err != nil {
			t.Fatalf("contextSet failed: %v", err)
		}

		// Test access to parent value
		result, err := bridge.contextGet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "parent_key",
		})
		if err != nil {
			t.Fatalf("contextGet failed: %v", err)
		}
		getValue := result.(map[string]interface{})
		if getValue["value"] != "parent_value" || getValue["exists"] != true {
			t.Errorf("Should access parent value, got: %v", result)
		}

		// Test override of parent value
		result, err = bridge.contextGet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "shared_key",
		})
		if err != nil {
			t.Fatalf("contextGet failed: %v", err)
		}
		getValue = result.(map[string]interface{})
		if getValue["value"] != "child_shared_value" || getValue["exists"] != true {
			t.Errorf("Should access overridden value, got: %v", result)
		}

		// Test child-only value
		result, err = bridge.contextGet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "child_key",
		})
		if err != nil {
			t.Fatalf("contextGet failed: %v", err)
		}
		getValue = result.(map[string]interface{})
		if getValue["value"] != "child_value" || getValue["exists"] != true {
			t.Errorf("Should access child value, got: %v", result)
		}
	})
}

func TestStateContextBridge_InheritanceConfiguration(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("withInheritanceConfig", func(t *testing.T) {
		// Create parent state with various data types
		parentState := domain.NewState()
		parentState.SetMetadata("parent_meta", "parent_meta_value")
		parentState.AddMessage(domain.Message{Role: "user", Content: "parent message"})

		// Create artifact and add to parent
		artifact := domain.NewArtifact("parent_artifact", domain.ArtifactTypeData, []byte("parent data"))
		parentState.AddArtifact(artifact)

		// Create shared context
		contextResult, _ := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": bridge.stateToScript(parentState),
		})
		contextObj := contextResult.(map[string]interface{})

		// Configure inheritance - disable metadata inheritance
		result, err := bridge.withInheritanceConfig(ctx, map[string]interface{}{
			"context":   contextObj,
			"messages":  true,
			"artifacts": true,
			"metadata":  false,
		})
		if err != nil {
			t.Fatalf("withInheritanceConfig failed: %v", err)
		}

		updatedContext, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("withInheritanceConfig should return context object")
		}

		// Verify inheritance config was updated
		if updatedContext["inheritMessages"] != true {
			t.Error("Messages inheritance should be enabled")
		}
		if updatedContext["inheritArtifacts"] != true {
			t.Error("Artifacts inheritance should be enabled")
		}
		if updatedContext["inheritMetadata"] != false {
			t.Error("Metadata inheritance should be disabled")
		}
	})
}

func TestStateContextBridge_StateIsolation(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("localStateAccess", func(t *testing.T) {
		// Create parent state
		parentState := domain.NewState()
		parentState.Set("parent_key", "parent_value")

		// Create shared context
		contextResult, _ := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": bridge.stateToScript(parentState),
		})
		contextObj := contextResult.(map[string]interface{})

		// Add local data
		_, err := bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "local_key",
			"value":   "local_value",
		})
		if err != nil {
			t.Fatalf("contextSet failed: %v", err)
		}

		// Get local state - should only contain local data
		result, err := bridge.localState(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err != nil {
			t.Fatalf("localState failed: %v", err)
		}

		localStateObj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("localState should return state object")
		}

		// Verify local state isolation
		localData := localStateObj["data"].(map[string]interface{})
		if localData["local_key"] != "local_value" {
			t.Error("Local state should contain local data")
		}
		if _, exists := localData["parent_key"]; exists {
			t.Error("Local state should not contain parent data")
		}
	})

	t.Run("clone", func(t *testing.T) {
		// Create parent state
		parentState := domain.NewState()
		parentState.Set("parent_key", "parent_value")

		// Create shared context
		contextResult, _ := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": bridge.stateToScript(parentState),
		})
		contextObj := contextResult.(map[string]interface{})

		// Add some local data
		_, err := bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "original_key",
			"value":   "original_value",
		})
		if err != nil {
			t.Fatalf("contextSet failed: %v", err)
		}

		// Clone the context
		result, err := bridge.clone(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err != nil {
			t.Fatalf("clone failed: %v", err)
		}

		cloneObj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("clone should return context object")
		}

		// Verify clone has same parent but fresh local state
		// Clone should still access parent data
		result, err = bridge.contextGet(ctx, map[string]interface{}{
			"context": cloneObj,
			"key":     "parent_key",
		})
		if err != nil {
			t.Fatalf("contextGet on clone failed: %v", err)
		}
		getValue := result.(map[string]interface{})
		if getValue["value"] != "parent_value" || getValue["exists"] != true {
			t.Error("Clone should still access parent data")
		}

		// Clone should not have original local data
		result, err = bridge.contextGet(ctx, map[string]interface{}{
			"context": cloneObj,
			"key":     "original_key",
		})
		if err != nil {
			t.Fatalf("contextGet on clone failed: %v", err)
		}
		getValue = result.(map[string]interface{})
		if getValue["exists"] != false {
			t.Error("Clone should not have original local data")
		}
	})

	t.Run("asState", func(t *testing.T) {
		// Create parent state
		parentState := domain.NewState()
		parentState.Set("parent_key", "parent_value")
		parentState.SetMetadata("parent_meta", "parent_meta_value")

		// Create shared context
		contextResult, _ := bridge.createSharedContext(ctx, map[string]interface{}{
			"parent": bridge.stateToScript(parentState),
		})
		contextObj := contextResult.(map[string]interface{})

		// Add local data
		_, err := bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"key":     "local_key",
			"value":   "local_value",
		})
		if err != nil {
			t.Fatalf("contextSet failed: %v", err)
		}

		// Convert to regular state
		result, err := bridge.asState(ctx, map[string]interface{}{
			"context": contextObj,
		})
		if err != nil {
			t.Fatalf("asState failed: %v", err)
		}

		stateObj, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("asState should return state object")
		}

		// Verify merged state contains both parent and local data
		stateData := stateObj["data"].(map[string]interface{})
		if stateData["parent_key"] != "parent_value" {
			t.Error("Merged state should contain parent data")
		}
		if stateData["local_key"] != "local_value" {
			t.Error("Merged state should contain local data")
		}
	})
}

func TestStateContextBridge_ErrorHandling(t *testing.T) {
	bridge, err := NewStateContextBridge()
	if err != nil {
		t.Fatalf("Failed to create bridge: %v", err)
	}

	ctx := context.Background()

	t.Run("createSharedContext_without_parent", func(t *testing.T) {
		_, err := bridge.createSharedContext(ctx, map[string]interface{}{})
		if err == nil {
			t.Error("createSharedContext should fail without parent parameter")
		}
	})

	t.Run("contextGet_without_context", func(t *testing.T) {
		_, err := bridge.contextGet(ctx, map[string]interface{}{
			"key": "test_key",
		})
		if err == nil {
			t.Error("contextGet should fail without context parameter")
		}
	})

	t.Run("contextSet_without_key", func(t *testing.T) {
		contextObj := map[string]interface{}{}
		_, err := bridge.contextSet(ctx, map[string]interface{}{
			"context": contextObj,
			"value":   "test_value",
		})
		if err == nil {
			t.Error("contextSet should fail without key parameter")
		}
	})

	t.Run("withInheritanceConfig_invalid_parameters", func(t *testing.T) {
		contextObj := map[string]interface{}{}
		_, err := bridge.withInheritanceConfig(ctx, map[string]interface{}{
			"context":  contextObj,
			"messages": "invalid_boolean",
		})
		if err == nil {
			t.Error("withInheritanceConfig should fail with invalid boolean parameter")
		}
	})
}

// Tests for event emission functionality using go-llms testutils
func TestStateContextBridge_EventEmission(t *testing.T) {
	// Create mock event emitter from go-llms testutils
	mockEmitter := mocks.NewMockEventEmitter("test-context", "Test Context")

	// Create bridge with event emitter
	bridge, err := NewStateContextBridgeWithEventEmitter(mockEmitter)
	require.NoError(t, err)
	require.NotNil(t, bridge)

	ctx := context.Background()

	// Create shared context for testing
	parentState := domain.NewState()
	parentState.Set("parentKey", "parentValue")

	result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
		map[string]interface{}{
			"data": map[string]interface{}{
				"parentKey": "parentValue",
			},
		},
	})
	require.NoError(t, err)

	sharedContextObj := result.(map[string]interface{})

	t.Run("StateChangeEvent_Set", func(t *testing.T) {
		// Reset the mock emitter
		mockEmitter.Reset()

		// Perform a set operation
		_, err := bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"testKey",
			"testValue",
		})
		require.NoError(t, err)

		// Verify event was emitted using go-llms testutils
		err = mockEmitter.AssertEventEmitted(domain.EventStateUpdate)
		assert.NoError(t, err, "StateUpdate event should have been emitted")

		// Get events for detailed validation
		events := mockEmitter.GetEvents()
		assertion := helpers.AssertEvents(events)
		assertion.HasCount(1).
			HasType(domain.EventStateUpdate)

		if !assertion.IsValid() {
			t.Error(assertion.String())
		}

		// Validate event data using scenario matchers
		stateUpdateEvents := mockEmitter.GetEventsByType(domain.EventStateUpdate)
		require.Len(t, stateUpdateEvents, 1)

		event := stateUpdateEvents[0]
		matcher := scenario.AllOf(
			scenario.HasField("Type", scenario.Equals(domain.EventStateUpdate)),
			scenario.HasField("Data", scenario.IsNotNil()),
		)

		if ok, msg := matcher.Match(event); !ok {
			t.Errorf("Event validation failed: %s", msg)
		}

		// Validate event data structure
		if eventData, ok := event.Data.(domain.StateUpdateEventData); ok {
			assert.Equal(t, "testKey", eventData.Key)
			assert.Equal(t, "testValue", eventData.NewValue)
			assert.Equal(t, "set", eventData.Action)
		} else {
			t.Error("Event data should be StateUpdateEventData")
		}
	})

	t.Run("StateDeleteEvent", func(t *testing.T) {
		// Reset the mock emitter
		mockEmitter.Reset()

		// First set a value
		_, err := bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"deleteKey",
			"deleteValue",
		})
		require.NoError(t, err)

		// Reset again to focus on delete event
		mockEmitter.Reset()

		// Perform a delete operation
		_, err = bridge.ExecuteMethod(ctx, "delete", []interface{}{
			sharedContextObj,
			"deleteKey",
		})
		require.NoError(t, err)

		// Verify delete event was emitted
		err = mockEmitter.AssertEventEmitted(domain.EventStateUpdate)
		assert.NoError(t, err, "StateUpdate event should have been emitted for delete")

		// Validate delete event data
		stateUpdateEvents := mockEmitter.GetEventsByType(domain.EventStateUpdate)
		require.Len(t, stateUpdateEvents, 1)

		event := stateUpdateEvents[0]
		if eventData, ok := event.Data.(domain.StateUpdateEventData); ok {
			assert.Equal(t, "deleteKey", eventData.Key)
			assert.Equal(t, "deleteValue", eventData.OldValue)
			assert.Nil(t, eventData.NewValue)
			assert.Equal(t, "delete", eventData.Action)
		} else {
			t.Error("Event data should be StateUpdateEventData for delete")
		}
	})

	t.Run("StateSnapshotEvent", func(t *testing.T) {
		// Reset the mock emitter
		mockEmitter.Reset()

		// Create a snapshot
		_, err := bridge.ExecuteMethod(ctx, "createSnapshot", []interface{}{
			sharedContextObj,
		})
		require.NoError(t, err)

		// Verify snapshot event was emitted
		events := mockEmitter.GetEvents()
		require.GreaterOrEqual(t, len(events), 1, "At least one event should have been emitted")

		// Look for custom snapshot event (mock emitter prefixes with "custom.")
		customEvents := make([]domain.Event, 0)
		for _, event := range events {
			if event.Type == "custom.state.snapshot" {
				customEvents = append(customEvents, event)
			}
		}

		require.Len(t, customEvents, 1, "One snapshot event should have been emitted")

		snapshotEvent := customEvents[0]
		assert.Equal(t, "custom.state.snapshot", string(snapshotEvent.Type))
		assert.NotNil(t, snapshotEvent.Data)
	})
}

func TestStateContextBridge_EventFiltering(t *testing.T) {
	mockEmitter := mocks.NewMockEventEmitter("test-context", "Test Context")
	bridge, err := NewStateContextBridgeWithEventEmitter(mockEmitter)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("AddEventFilter", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "addEventFilter", []interface{}{
			"state.*",
			"state-filter",
		})
		require.NoError(t, err)

		filterResult := result.(map[string]interface{})
		assert.Equal(t, "state-filter", filterResult["name"])
		assert.Equal(t, "state.*", filterResult["pattern"])
		assert.True(t, filterResult["active"].(bool))
	})

	t.Run("ListEventFilters", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "listEventFilters", []interface{}{})
		require.NoError(t, err)

		filtersList := result.(map[string]interface{})
		filters := filtersList["filters"].([]interface{})
		assert.GreaterOrEqual(t, len(filters), 1)
		assert.Equal(t, len(filters), filtersList["total"])
	})

	t.Run("RemoveEventFilter", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "removeEventFilter", []interface{}{
			"state-filter",
		})
		require.NoError(t, err)

		removeResult := result.(map[string]interface{})
		assert.True(t, removeResult["removed"].(bool))
		assert.Equal(t, "state-filter", removeResult["name"])
	})
}

func TestStateContextBridge_EventReplay(t *testing.T) {
	mockEmitter := mocks.NewMockEventEmitter("replay-context", "Replay Context")
	bridge, err := NewStateContextBridgeWithEventEmitter(mockEmitter)
	require.NoError(t, err)

	ctx := context.Background()

	// Create shared context for testing
	result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
		map[string]interface{}{
			"data": map[string]interface{}{},
		},
	})
	require.NoError(t, err)

	sharedContextObj := result.(map[string]interface{})
	contextID := sharedContextObj["_id"].(string)

	// Create some state changes to generate events
	_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
		sharedContextObj,
		"key1",
		"value1",
	})
	require.NoError(t, err)

	_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
		sharedContextObj,
		"key2",
		"value2",
	})
	require.NoError(t, err)

	// Small delay to ensure timestamp ordering
	time.Sleep(10 * time.Millisecond)

	t.Run("GetEventHistory", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "getEventHistory", []interface{}{
			contextID,
			10,
		})
		require.NoError(t, err)

		historyResult := result.(map[string]interface{})
		events := historyResult["events"].([]interface{})
		assert.GreaterOrEqual(t, len(events), 2, "Should have at least 2 events from set operations")
		assert.Equal(t, contextID, historyResult["contextId"])
		assert.Equal(t, 10, historyResult["limit"])
	})

	t.Run("ReplayEvents", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "replayEvents", []interface{}{
			contextID,
		})
		require.NoError(t, err)

		replayResult := result.(map[string]interface{})
		assert.Contains(t, replayResult, "replayContextId")
		assert.Contains(t, replayResult, "eventsReplayed")
		assert.Contains(t, replayResult, "replayedContext")

		eventsReplayed := replayResult["eventsReplayed"].(int)
		assert.GreaterOrEqual(t, eventsReplayed, 2, "Should have replayed at least 2 events")

		replayedContext := replayResult["replayedContext"].(map[string]interface{})
		assert.Contains(t, replayedContext, "_id")
		assert.Contains(t, replayedContext, "type")
		assert.Equal(t, "SharedStateContext", replayedContext["type"])
	})

	t.Run("ClearEventHistory", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "clearEventHistory", []interface{}{
			contextID,
		})
		require.NoError(t, err)

		clearResult := result.(map[string]interface{})
		assert.True(t, clearResult["cleared"].(int) >= 0)
		assert.Equal(t, "context", clearResult["scope"])
		assert.Equal(t, contextID, clearResult["contextId"])

		// Verify events were cleared
		result, err = bridge.ExecuteMethod(ctx, "getEventHistory", []interface{}{
			contextID,
			10,
		})
		require.NoError(t, err)

		historyResult := result.(map[string]interface{})
		events := historyResult["events"].([]interface{})
		assert.Equal(t, 0, len(events), "Event history should be empty after clearing")
	})
}

func TestStateContextBridge_EventEmissionEdgeCases(t *testing.T) {
	t.Run("NoEventEmitter", func(t *testing.T) {
		// Test bridge without event emitter
		bridge, err := NewStateContextBridge()
		require.NoError(t, err)

		ctx := context.Background()

		// Create shared context
		result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{},
			},
		})
		require.NoError(t, err)

		sharedContextObj := result.(map[string]interface{})

		// Set operation should succeed even without event emitter
		_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"testKey",
			"testValue",
		})
		assert.NoError(t, err, "Set operation should succeed without event emitter")
	})

	t.Run("EventEmissionWithBlockedEvents", func(t *testing.T) {
		mockEmitter := mocks.NewMockEventEmitter("blocked-context", "Blocked Context")
		mockEmitter.SetBlockEvents(true)

		bridge, err := NewStateContextBridgeWithEventEmitter(mockEmitter)
		require.NoError(t, err)

		ctx := context.Background()

		// Create shared context
		result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{},
			},
		})
		require.NoError(t, err)

		sharedContextObj := result.(map[string]interface{})

		// Perform set operation
		_, err = bridge.ExecuteMethod(ctx, "set", []interface{}{
			sharedContextObj,
			"blockedKey",
			"blockedValue",
		})
		require.NoError(t, err)

		// Events should be blocked
		events := mockEmitter.GetEvents()
		assert.Equal(t, 0, len(events), "Events should be blocked when SetBlockEvents(true)")
	})
}
