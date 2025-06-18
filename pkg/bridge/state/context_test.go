// ABOUTME: Tests for State Context Bridge with ScriptValue-based API
// ABOUTME: Validates parent-child state sharing and inheritance configuration

package state

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateContextBridgeInitialization(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)
	assert.NotNil(t, bridge)
	assert.Equal(t, "state_context", bridge.GetID())
	assert.True(t, bridge.IsInitialized())
}

func TestStateContextBridgeWithEventEmitter(t *testing.T) {
	bridge, err := NewStateContextBridgeWithEventEmitter(nil)
	require.NoError(t, err)
	assert.NotNil(t, bridge)
}

func TestStateContextBridgeWithOptions(t *testing.T) {
	bridge, err := NewStateContextBridgeWithOptions(nil, "/tmp/test_persist", true)
	require.NoError(t, err)
	assert.NotNil(t, bridge)
	assert.Equal(t, "/tmp/test_persist", bridge.persistDir)
	assert.True(t, bridge.enableCompress)
}

func TestStateContextBridgeMetadata(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	metadata := bridge.GetMetadata()
	assert.Equal(t, "State Context Bridge", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "SharedStateContext")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestStateContextBridgeMethods(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := []string{
		"createSharedContext",
		"withInheritanceConfig",
		"get",
		"set",
		"delete",
		"has",
		"keys",
		"values",
		"getArtifact",
		"artifacts",
		"messages",
		"getMetadata",
		"localState",
		"clone",
		"asState",
		"parentState",
		"generateContextID",
		"validateWithSchema",
		"saveState",
		"loadState",
		"deleteState",
		"getAllStateVersions",
		"loadStateVersion",
		"registerSchema",
		"getSchemaForContext",
		"enableEventEmission",
		"disableEventEmission",
		"emitEvent",
		"subscribeToEvents",
		"unsubscribeFromEvents",
		"getEventHistory",
		"replayEvents",
		"setPersistenceDirectory",
		"enableCompression",
		"disableCompression",
		"registerTransformPipeline",
		"applyTransform",
		"getTransformMetrics",
		"clearTransformCache",
		"importState",
		"exportState",
		"mergeStates",
		"diffStates",
		"lockState",
		"unlockState",
		"isStateLocked",
		"getContextStats",
		"clearContext",
		"getAllContexts",
		"setEventFilter",
		"removeEventFilter",
		"getActiveFilters",
		"validateState",
		"repairState",
		"optimizeState",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestStateContextBridgeCreateSharedContext(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Test creating shared context without parent
	result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, result.Type())

	contextObj := result.(engine.ObjectValue).Fields()
	assert.Contains(t, contextObj, "_id")
	assert.Contains(t, contextObj, "_type")
	assert.Equal(t, "SharedStateContext", contextObj["_type"].(engine.StringValue).Value())

	// Store context ID for further tests
	contextID := contextObj["_id"].(engine.StringValue).Value()

	// Test creating shared context with parent
	childResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		result,
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, childResult.Type())

	childContextObj := childResult.(engine.ObjectValue).Fields()
	assert.Contains(t, childContextObj, "_id")
	assert.Contains(t, childContextObj, "_parent")
	assert.Equal(t, contextID, childContextObj["_parent"].(engine.StringValue).Value())
}

func TestStateContextBridgeInheritanceConfig(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create a context
	contextResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Configure inheritance
	result, err := bridge.ExecuteMethod(ctx, "withInheritanceConfig", []engine.ScriptValue{
		contextResult,
		engine.NewBoolValue(true),  // messages
		engine.NewBoolValue(false), // artifacts
		engine.NewBoolValue(true),  // metadata
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, result.Type())
}

func TestStateContextBridgeGetSet(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create parent context
	parentResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Set value in parent
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		parentResult,
		engine.NewStringValue("parent_key"),
		engine.NewStringValue("parent_value"),
	})
	require.NoError(t, err)

	// Get value from parent
	value, err := bridge.ExecuteMethod(ctx, "get", []engine.ScriptValue{
		parentResult,
		engine.NewStringValue("parent_key"),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, value.Type())
	assert.Equal(t, "parent_value", value.(engine.StringValue).Value())

	// Create child context
	childResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		parentResult,
	})
	require.NoError(t, err)

	// Get parent value from child (should inherit)
	childValue, err := bridge.ExecuteMethod(ctx, "get", []engine.ScriptValue{
		childResult,
		engine.NewStringValue("parent_key"),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, childValue.Type())
	assert.Equal(t, "parent_value", childValue.(engine.StringValue).Value())

	// Set override in child
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		childResult,
		engine.NewStringValue("parent_key"),
		engine.NewStringValue("child_override"),
	})
	require.NoError(t, err)

	// Get overridden value from child
	overriddenValue, err := bridge.ExecuteMethod(ctx, "get", []engine.ScriptValue{
		childResult,
		engine.NewStringValue("parent_key"),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, overriddenValue.Type())
	assert.Equal(t, "child_override", overriddenValue.(engine.StringValue).Value())

	// Parent value should be unchanged
	parentValue, err := bridge.ExecuteMethod(ctx, "get", []engine.ScriptValue{
		parentResult,
		engine.NewStringValue("parent_key"),
	})
	require.NoError(t, err)
	assert.Equal(t, "parent_value", parentValue.(engine.StringValue).Value())
}

func TestStateContextBridgeDelete(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create context
	contextResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Set value
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("test_key"),
		engine.NewStringValue("test_value"),
	})
	require.NoError(t, err)

	// Verify it exists
	hasResult, err := bridge.ExecuteMethod(ctx, "has", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("test_key"),
	})
	require.NoError(t, err)
	assert.True(t, hasResult.(engine.BoolValue).Value())

	// Delete the key
	_, err = bridge.ExecuteMethod(ctx, "delete", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("test_key"),
	})
	require.NoError(t, err)

	// Verify it's gone
	hasResult, err = bridge.ExecuteMethod(ctx, "has", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("test_key"),
	})
	require.NoError(t, err)
	assert.False(t, hasResult.(engine.BoolValue).Value())
}

func TestStateContextBridgeKeysValues(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create context
	contextResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Set multiple values
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range testData {
		_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
			contextResult,
			engine.NewStringValue(k),
			engine.NewStringValue(v),
		})
		require.NoError(t, err)
	}

	// Get keys
	keysResult, err := bridge.ExecuteMethod(ctx, "keys", []engine.ScriptValue{contextResult})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeArray, keysResult.Type())

	keys := keysResult.(engine.ArrayValue).Elements()
	assert.Len(t, keys, 3)

	// Get values
	valuesResult, err := bridge.ExecuteMethod(ctx, "values", []engine.ScriptValue{contextResult})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeArray, valuesResult.Type())

	values := valuesResult.(engine.ArrayValue).Elements()
	assert.Len(t, values, 3)
}

func TestStateContextBridgeSchemaValidation(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Register a schema
	schema := engine.NewObjectValue(map[string]engine.ScriptValue{
		"type": engine.NewStringValue("object"),
		"properties": engine.NewObjectValue(map[string]engine.ScriptValue{
			"name": engine.NewObjectValue(map[string]engine.ScriptValue{
				"type": engine.NewStringValue("string"),
			}),
			"age": engine.NewObjectValue(map[string]engine.ScriptValue{
				"type": engine.NewStringValue("number"),
			}),
		}),
		"required": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("name"),
		}),
	})

	schemaID, err := bridge.ExecuteMethod(ctx, "registerSchema", []engine.ScriptValue{
		engine.NewStringValue("person"),
		schema,
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, schemaID.Type())

	// Create context
	contextResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Validate valid state
	validState := engine.NewObjectValue(map[string]engine.ScriptValue{
		"name": engine.NewStringValue("John"),
		"age":  engine.NewNumberValue(30),
	})

	isValid, err := bridge.ExecuteMethod(ctx, "validateWithSchema", []engine.ScriptValue{
		contextResult,
		schemaID,
		validState,
	})
	require.NoError(t, err)
	assert.True(t, isValid.(engine.BoolValue).Value())

	// Validate invalid state (missing required field)
	invalidState := engine.NewObjectValue(map[string]engine.ScriptValue{
		"age": engine.NewNumberValue(30),
	})

	isValid, err = bridge.ExecuteMethod(ctx, "validateWithSchema", []engine.ScriptValue{
		contextResult,
		schemaID,
		invalidState,
	})
	require.NoError(t, err)
	assert.False(t, isValid.(engine.BoolValue).Value())
}

func TestStateContextBridgeClone(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create context with data
	contextResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Add some data
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("original_key"),
		engine.NewStringValue("original_value"),
	})
	require.NoError(t, err)

	// Clone the context
	cloneResult, err := bridge.ExecuteMethod(ctx, "clone", []engine.ScriptValue{contextResult})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, cloneResult.Type())

	// Verify clone has fresh local state (no data)
	// Clone has the same parent as the original (nil in this case)
	// So it should not see the original's data
	hasKey, err := bridge.ExecuteMethod(ctx, "has", []engine.ScriptValue{
		cloneResult,
		engine.NewStringValue("original_key"),
	})
	require.NoError(t, err)
	assert.False(t, hasKey.(engine.BoolValue).Value())

	// Set value in clone
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		cloneResult,
		engine.NewStringValue("clone_key"),
		engine.NewStringValue("clone_value"),
	})
	require.NoError(t, err)

	// Original should not have clone's data
	hasCloneKey, err := bridge.ExecuteMethod(ctx, "has", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("clone_key"),
	})
	require.NoError(t, err)
	assert.False(t, hasCloneKey.(engine.BoolValue).Value())
}

func TestStateContextBridgeAsState(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create parent context
	parentResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	// Set parent data
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		parentResult,
		engine.NewStringValue("parent_key"),
		engine.NewStringValue("parent_value"),
	})
	require.NoError(t, err)

	// Create child context
	childResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		parentResult,
	})
	require.NoError(t, err)

	// Set child data
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		childResult,
		engine.NewStringValue("child_key"),
		engine.NewStringValue("child_value"),
	})
	require.NoError(t, err)

	// Convert to state (merges parent and child data)
	stateResult, err := bridge.ExecuteMethod(ctx, "asState", []engine.ScriptValue{childResult})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, stateResult.Type())

	stateObj := stateResult.(engine.ObjectValue).Fields()
	assert.Contains(t, stateObj, "type")
	assert.Equal(t, "State", stateObj["type"].(engine.StringValue).Value())

	// Check merged data
	data := stateObj["data"].(engine.ObjectValue).Fields()
	assert.Contains(t, data, "parent_key")
	assert.Contains(t, data, "child_key")
}

func TestStateContextBridgeGetAllContexts(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Initially no contexts
	contexts, err := bridge.ExecuteMethod(ctx, "getAllContexts", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeArray, contexts.Type())
	assert.Len(t, contexts.(engine.ArrayValue).Elements(), 0)

	// Create multiple contexts
	var contextIDs []string
	for i := 0; i < 3; i++ {
		result, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
			engine.NewNilValue(),
		})
		require.NoError(t, err)

		contextObj := result.(engine.ObjectValue).Fields()
		_ = append(contextIDs, contextObj["_id"].(engine.StringValue).Value())
	}

	// Get all contexts
	contexts, err = bridge.ExecuteMethod(ctx, "getAllContexts", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeArray, contexts.Type())
	assert.Len(t, contexts.(engine.ArrayValue).Elements(), 3)
}

func TestStateContextBridgeClearContext(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Create context
	contextResult, err := bridge.ExecuteMethod(ctx, "createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	require.NoError(t, err)

	contextID := contextResult.(engine.ObjectValue).Fields()["_id"].(engine.StringValue).Value()

	// Add data
	_, err = bridge.ExecuteMethod(ctx, "set", []engine.ScriptValue{
		contextResult,
		engine.NewStringValue("test_key"),
		engine.NewStringValue("test_value"),
	})
	require.NoError(t, err)

	// Clear context
	result, err := bridge.ExecuteMethod(ctx, "clearContext", []engine.ScriptValue{
		engine.NewStringValue(contextID),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Verify context is gone
	contexts, err := bridge.ExecuteMethod(ctx, "getAllContexts", []engine.ScriptValue{})
	require.NoError(t, err)
	assert.Len(t, contexts.(engine.ArrayValue).Elements(), 0)
}

func TestStateContextBridgeValidateMethod(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	// ValidateMethod should always return nil as validation is handled by engine
	err = bridge.ValidateMethod("createSharedContext", []engine.ScriptValue{
		engine.NewNilValue(),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

func TestStateContextBridgeRequiredPermissions(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	permissions := bridge.RequiredPermissions()

	// Check for expected permissions
	hasMemory := false
	hasStorage := false

	for _, perm := range permissions {
		switch perm.Type {
		case engine.PermissionMemory:
			if perm.Resource == "state" {
				hasMemory = true
				assert.Contains(t, perm.Actions, "read")
				assert.Contains(t, perm.Actions, "write")
			}
		case engine.PermissionStorage:
			if perm.Resource == "state_persistence" {
				hasStorage = true
				assert.Contains(t, perm.Actions, "read")
				assert.Contains(t, perm.Actions, "write")
			}
		}
	}

	assert.True(t, hasMemory, "Memory permission not found")
	assert.True(t, hasStorage, "Storage permission not found")
}

func TestStateContextBridgeTypeMappings(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	mappings := bridge.TypeMappings()

	// Check expected type mappings
	expectedTypes := []string{
		"SharedStateContext",
		"State",
		"Message",
		"Artifact",
		"Event",
		"Schema",
		"ValidationResult",
		"TransformPipeline",
		"TransformMetrics",
	}

	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestStateContextBridgeErrorHandling(t *testing.T) {
	bridge, err := NewStateContextBridge()
	require.NoError(t, err)

	ctx := context.Background()

	// Test invalid arguments
	_, err = bridge.ExecuteMethod(ctx, "get", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires context and key parameters")

	// Test invalid context type
	_, err = bridge.ExecuteMethod(ctx, "get", []engine.ScriptValue{
		engine.NewStringValue("not a context"),
		engine.NewStringValue("key"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context must be object")

	// Test unknown method
	_, err = bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method not found")
}
