// ABOUTME: Test suite for the LLM utilities bridge that wraps go-llms LLM utility functions.
// ABOUTME: Tests bridge interface compliance and method definitions.

package util

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

func TestNewUtilLLMBridge(t *testing.T) {
	bridge := NewUtilLLMBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_llm", bridge.GetID())
}

func TestUtilLLMBridgeMetadata(t *testing.T) {
	bridge := NewUtilLLMBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util_llm", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Enhanced LLM utilities")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilLLMBridgeInitialization(t *testing.T) {
	bridge := NewUtilLLMBridge()
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

func TestUtilLLMBridgeMethods(t *testing.T) {
	bridge := NewUtilLLMBridge()
	methods := bridge.Methods()

	// Check that all expected method categories are present
	expectedMethods := map[string]bool{
		// Provider creation
		"createProvider":        false,
		"createProviderFromEnv": false,
		"withProviderOptions":   false,

		// Typed generation
		"generateTyped":            false,
		"validateStructuredOutput": false,

		// Provider pool
		"createProviderPool": false,
		"addProviderToPool":  false,

		// Model inventory
		"createModelInventory": false,
		"fetchModelInfo":       false,
		"cacheModelInfo":       false,

		// Configuration
		"createModelConfig":    false,
		"mergeProviderOptions": false,
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

func TestUtilLLMBridgeMethodDetails(t *testing.T) {
	bridge := NewUtilLLMBridge()
	methods := bridge.Methods()

	// Verify createProvider method details
	var createProviderMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "createProvider" {
			createProviderMethod = &m
			break
		}
	}
	assert.NotNil(t, createProviderMethod)
	assert.Contains(t, createProviderMethod.Description, "provider from configuration")
	assert.Len(t, createProviderMethod.Parameters, 1)
	assert.Equal(t, "Provider", createProviderMethod.ReturnType)

	// Verify generateTyped method details
	var typedMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "generateTyped" {
			typedMethod = &m
			break
		}
	}
	assert.NotNil(t, typedMethod)
	assert.Contains(t, typedMethod.Description, "typed/structured")
	assert.GreaterOrEqual(t, len(typedMethod.Parameters), 3)
	assert.Equal(t, "object", typedMethod.ReturnType)
}

func TestUtilLLMBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilLLMBridge()
	mappings := bridge.TypeMappings()

	// Check that expected type mappings are present
	expectedTypes := []string{"ProviderPool", "ModelInventory", "ModelConfig"}
	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.Equal(t, "object", mapping.ScriptType)
	}
}

func TestUtilLLMBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilLLMBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 2)

	// Check for required permissions
	hasNetworkPerm := false
	hasFileSystemPerm := false

	for _, perm := range permissions {
		if perm.Type == "network" && perm.Resource == "llm" {
			hasNetworkPerm = true
			assert.Contains(t, perm.Actions, "create")
			assert.Contains(t, perm.Actions, "access")
		}
		if perm.Type == "filesystem" && perm.Resource == "cache" {
			hasFileSystemPerm = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasNetworkPerm, "Network permission not found")
	assert.True(t, hasFileSystemPerm, "FileSystem permission not found")
}

func TestUtilLLMBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilLLMBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("createProvider", []engine.ScriptValue{
		engine.NewObjectValue(map[string]engine.ScriptValue{}),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

// Note: Actual LLM utility testing would require real go-llms implementations
// or would be done at integration test level with actual utilities
