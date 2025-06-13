// ABOUTME: Test suite for the Model Info bridge that wraps go-llms ModelRegistry.
// ABOUTME: Tests bridge interface compliance without mocking go-llms types.

package bridge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModelInfoBridge(t *testing.T) {
	bridge := NewModelInfoBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "modelinfo", bridge.GetID())
}

func TestModelInfoBridgeMetadata(t *testing.T) {
	bridge := NewModelInfoBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "Model Info Bridge", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "ModelRegistry")
	assert.Equal(t, "go-llmspell", metadata.Author)
}

func TestModelInfoBridgeInitialization(t *testing.T) {
	bridge := NewModelInfoBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test double initialization
	err = bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test cleanup
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestModelInfoBridgeMethods(t *testing.T) {
	bridge := NewModelInfoBridge()
	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := map[string]bool{
		"registerModelRegistry": false,
		"listModels":            false,
		"listModelsByRegistry":  false,
		"getModel":              false,
		"listRegistries":        false,
	}

	for _, method := range methods {
		if _, ok := expectedMethods[method.Name]; ok {
			expectedMethods[method.Name] = true
		}
	}

	for method, found := range expectedMethods {
		assert.True(t, found, "Method %s not found", method)
	}
}

func TestModelInfoBridgeTypeMappings(t *testing.T) {
	bridge := NewModelInfoBridge()
	mappings := bridge.TypeMappings()

	// Check that expected type mappings are present
	expectedTypes := []string{"ModelRegistry", "Model"}
	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.Equal(t, typeName, mapping.GoType)
		assert.Equal(t, "object", mapping.ScriptType)
	}
}

func TestModelInfoBridgeRequiredPermissions(t *testing.T) {
	bridge := NewModelInfoBridge()
	permissions := bridge.RequiredPermissions()

	assert.Len(t, permissions, 1)
	assert.Equal(t, "memory", string(permissions[0].Type))
	assert.Equal(t, "modelinfo", permissions[0].Resource)
	assert.Contains(t, permissions[0].Actions, "read")
}

func TestModelInfoBridgeValidateMethod(t *testing.T) {
	bridge := NewModelInfoBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("listModels", nil)
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", nil)
	assert.NoError(t, err)
}

func TestModelInfoBridgeRegistryManagement(t *testing.T) {
	bridge := NewModelInfoBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Test listing registries when empty
	registries := bridge.ListRegistries()
	assert.Empty(t, registries)

	// Test getting registry when none exist
	registry := bridge.GetRegistry("nonexistent")
	assert.Nil(t, registry)
}

// Note: Actual ModelRegistry testing would require real go-llms ModelRegistry implementations
// or would be done at integration test level
