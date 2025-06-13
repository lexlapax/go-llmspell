// ABOUTME: Test suite for the utilities bridge that wraps go-llms utility functions.
// ABOUTME: Tests bridge interface compliance and method definitions.

package util

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

func TestNewUtilBridge(t *testing.T) {
	bridge := NewUtilBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util", bridge.GetID())
}

func TestUtilBridgeMetadata(t *testing.T) {
	bridge := NewUtilBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "miscellaneous helper functions")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilBridgeInitialization(t *testing.T) {
	bridge := NewUtilBridge()
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

func TestUtilBridgeMethods(t *testing.T) {
	bridge := NewUtilBridge()
	methods := bridge.Methods()

	// Check that all expected method categories are present
	expectedMethods := map[string]bool{
		// Error handling utilities
		"isRetryableError": false,
		"wrapError":        false,
		"errorToString":    false,

		// String utilities
		"truncateString": false,
		"sanitizeString": false,

		// Time utilities
		"parseHumanDuration": false,
		"formatDuration":     false,

		// Retry utilities
		"retryWithBackoff":  false,
		"createRetryConfig": false,

		// Validation utilities
		"validateURL":   false,
		"validateEmail": false,

		// Misc utilities
		"generateUUID": false,
		"hashString":   false,
		"sleep":        false,
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

func TestUtilBridgeMethodDetails(t *testing.T) {
	bridge := NewUtilBridge()
	methods := bridge.Methods()

	// Verify isRetryableError method details
	var retryMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "isRetryableError" {
			retryMethod = &m
			break
		}
	}
	assert.NotNil(t, retryMethod)
	assert.Contains(t, retryMethod.Description, "retryable")
	assert.Len(t, retryMethod.Parameters, 1)
	assert.Equal(t, "boolean", retryMethod.ReturnType)

	// Verify generateUUID method details
	var uuidMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "generateUUID" {
			uuidMethod = &m
			break
		}
	}
	assert.NotNil(t, uuidMethod)
	assert.Len(t, uuidMethod.Parameters, 0)
	assert.Equal(t, "string", uuidMethod.ReturnType)
}

func TestUtilBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilBridge()
	mappings := bridge.TypeMappings()

	// Check that expected type mappings are present
	expectedTypes := []string{"error", "function"}
	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestUtilBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 2)

	// Check for memory permission
	hasMemoryPerm := false
	hasProcessPerm := false

	for _, perm := range permissions {
		if perm.Type == "memory" && perm.Resource == "util" {
			hasMemoryPerm = true
			assert.Contains(t, perm.Actions, "read")
		}
		if perm.Type == "time" && perm.Resource == "system" {
			hasProcessPerm = true
			assert.Contains(t, perm.Actions, "sleep")
		}
	}

	assert.True(t, hasMemoryPerm, "Memory permission not found")
	assert.True(t, hasProcessPerm, "Time permission not found")
}

func TestUtilBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("isRetryableError", []interface{}{errors.New("test")})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", nil)
	assert.NoError(t, err)
}

// Note: Actual utility function testing would require real go-llms utility implementations
// or would be done at integration test level with actual go-llms functions
