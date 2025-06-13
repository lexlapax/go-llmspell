// ABOUTME: Test suite for the JSON utilities bridge that wraps go-llms JSON functions.
// ABOUTME: Tests bridge interface compliance and method definitions.

package bridge

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

func TestNewUtilJSONBridge(t *testing.T) {
	bridge := NewUtilJSONBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_json", bridge.GetID())
}

func TestUtilJSONBridgeMetadata(t *testing.T) {
	bridge := NewUtilJSONBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util_json", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "JSON utilities")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilJSONBridgeInitialization(t *testing.T) {
	bridge := NewUtilJSONBridge()
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

func TestUtilJSONBridgeMethods(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	// Check that all expected method categories are present
	expectedMethods := map[string]bool{
		// Basic JSON operations
		"marshal":            false,
		"marshalIndent":      false,
		"marshalToBytes":     false,
		"unmarshal":          false,
		"unmarshalFromBytes": false,
		"unmarshalStrict":    false,

		// Streaming JSON
		"createEncoder": false,
		"createDecoder": false,
		"encodeStream":  false,
		"decodeStream":  false,

		// JSON schema
		"validateWithSchema": false,
		"generateFromSchema": false,
		"inferSchema":        false,

		// JSON utilities
		"prettyPrint": false,
		"minify":      false,
		"merge":       false,
		"diff":        false,

		// Performance utilities
		"marshalWithBuffer": false,
		"marshalConcurrent": false,
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

func TestUtilJSONBridgeMethodDetails(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	// Verify marshal method details
	var marshalMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "marshal" {
			marshalMethod = &m
			break
		}
	}
	assert.NotNil(t, marshalMethod)
	assert.Contains(t, marshalMethod.Description, "Marshal")
	assert.Len(t, marshalMethod.Parameters, 1)
	assert.Equal(t, "string", marshalMethod.ReturnType)

	// Verify validateWithSchema method details
	var schemaMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "validateWithSchema" {
			schemaMethod = &m
			break
		}
	}
	assert.NotNil(t, schemaMethod)
	assert.Contains(t, schemaMethod.Description, "JSON against schema")
	assert.Len(t, schemaMethod.Parameters, 2)
	assert.Equal(t, "boolean", schemaMethod.ReturnType)

	// Verify createEncoder method details
	var encoderMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "createEncoder" {
			encoderMethod = &m
			break
		}
	}
	assert.NotNil(t, encoderMethod)
	assert.Contains(t, encoderMethod.Description, "streaming")
	assert.Len(t, encoderMethod.Parameters, 1)
	assert.Equal(t, "JSONEncoder", encoderMethod.ReturnType)
}

func TestUtilJSONBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilJSONBridge()
	mappings := bridge.TypeMappings()

	// Check that expected type mappings are present
	expectedTypes := []string{"JSONEncoder", "JSONDecoder", "io.Writer", "io.Reader", "bytes"}
	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		if typeName == "bytes" {
			assert.Equal(t, "array", mapping.ScriptType)
		} else {
			assert.Equal(t, "object", mapping.ScriptType)
		}
	}
}

func TestUtilJSONBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilJSONBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 1)

	// Check for required permissions
	hasMemoryPerm := false

	for _, perm := range permissions {
		if perm.Type == "memory" && perm.Resource == "json" {
			hasMemoryPerm = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasMemoryPerm, "Memory permission not found")
}

func TestUtilJSONBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilJSONBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("marshal", []interface{}{map[string]interface{}{}})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", nil)
	assert.NoError(t, err)
}

// Note: Actual JSON utility testing would require real go-llms implementations
// or would be done at integration test level with actual utilities
