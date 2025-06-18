// ABOUTME: Test suite for the JSON utilities bridge that wraps go-llms JSON functions.
// ABOUTME: Tests bridge interface compliance and method definitions.

package util

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"

	// Use go-llms testutils for consistency
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
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
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Structured output parser")
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

	// Check that all expected method categories are present for v0.3.5 structured output
	expectedMethods := map[string]bool{
		// Optimized marshaling (go-llms JSON utils)
		"marshal":        false,
		"marshalIndent":  false,
		"marshalToBytes": false,

		// Optimized unmarshaling (go-llms JSON utils)
		"unmarshal":          false,
		"unmarshalFromBytes": false,
		"unmarshalStrict":    false,

		// Streaming operations (go-llms v0.3.5)
		"createEncoder": false,
		"createDecoder": false,
		"encodeStream":  false,
		"decodeStream":  false,

		// Schema operations (go-llms v0.3.5)
		"validateWithSchema": false,
		"generateFromSchema": false,
		"inferSchema":        false,

		// NEW: Structured output operations (go-llms v0.3.5)
		"parseStructured":   false,
		"parseWithRecovery": false,
		"enhancePrompt":     false,

		// NEW: Format conversion operations (go-llms v0.3.5)
		"convertFormat": false,
		"streamConvert": false,

		// JSON utilities (go-llms optimized)
		"prettyPrint": false,
		"minify":      false,
		"merge":       false,
		"diff":        false,

		// Performance utilities (go-llms optimized)
		"marshalWithBuffer": false,
		"marshalConcurrent": false,
	}

	for _, method := range methods {
		if _, ok := expectedMethods[method.Name]; ok {
			expectedMethods[method.Name] = true
		}
	}

	for method, found := range expectedMethods {
		assert.True(t, found, "Method %s not found in v0.3.5 implementation", method)
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

	// Use testutils fixture for realistic test data
	testState := fixtures.BasicTestState()
	testData := make(map[string]interface{})
	if data, exists := testState.Get("data"); exists {
		testData = data.(map[string]interface{})
	} else {
		testData["test"] = "value"
	}

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("marshal", []interface{}{testData})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", nil)
	assert.NoError(t, err)
}

// Tests for NEW v0.3.5 Parser Capabilities

func TestUtilJSONBridge_StructuredOutputCapabilities(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	// Test that new structured output methods are exposed
	structuredMethods := map[string]bool{
		"parseStructured":   false,
		"parseWithRecovery": false,
		"enhancePrompt":     false,
	}

	for _, method := range methods {
		if _, exists := structuredMethods[method.Name]; exists {
			structuredMethods[method.Name] = true
		}
	}

	for methodName, found := range structuredMethods {
		assert.True(t, found, "New v0.3.5 structured output method %s not found", methodName)
	}
}

func TestUtilJSONBridge_FormatConversionCapabilities(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	// Test that new format conversion methods are exposed
	conversionMethods := map[string]bool{
		"convertFormat": false,
		"streamConvert": false,
	}

	for _, method := range methods {
		if _, exists := conversionMethods[method.Name]; exists {
			conversionMethods[method.Name] = true
		}
	}

	for methodName, found := range conversionMethods {
		assert.True(t, found, "New v0.3.5 format conversion method %s not found", methodName)
	}
}

func TestUtilJSONBridge_ParseStructuredMethodSignature(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	var parseStructuredMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "parseStructured" {
			parseStructuredMethod = &m
			break
		}
	}

	assert.NotNil(t, parseStructuredMethod, "parseStructured method not found")
	assert.Contains(t, parseStructuredMethod.Description, "Parse and validate structured output")
	assert.Len(t, parseStructuredMethod.Parameters, 2)

	// Check parameters
	assert.Equal(t, "output", parseStructuredMethod.Parameters[0].Name)
	assert.Equal(t, "string", parseStructuredMethod.Parameters[0].Type)
	assert.True(t, parseStructuredMethod.Parameters[0].Required)

	assert.Equal(t, "schema", parseStructuredMethod.Parameters[1].Name)
	assert.Equal(t, "object", parseStructuredMethod.Parameters[1].Type)
	assert.True(t, parseStructuredMethod.Parameters[1].Required)

	assert.Equal(t, "any", parseStructuredMethod.ReturnType)
}

func TestUtilJSONBridge_ParseWithRecoveryMethodSignature(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	var parseWithRecoveryMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "parseWithRecovery" {
			parseWithRecoveryMethod = &m
			break
		}
	}

	assert.NotNil(t, parseWithRecoveryMethod, "parseWithRecovery method not found")
	assert.Contains(t, parseWithRecoveryMethod.Description, "Extract JSON from malformed")
	assert.Len(t, parseWithRecoveryMethod.Parameters, 1)

	// Check parameter
	assert.Equal(t, "output", parseWithRecoveryMethod.Parameters[0].Name)
	assert.Equal(t, "string", parseWithRecoveryMethod.Parameters[0].Type)
	assert.True(t, parseWithRecoveryMethod.Parameters[0].Required)

	assert.Equal(t, "string", parseWithRecoveryMethod.ReturnType)
}

func TestUtilJSONBridge_EnhancePromptMethodSignature(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	var enhancePromptMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "enhancePrompt" {
			enhancePromptMethod = &m
			break
		}
	}

	assert.NotNil(t, enhancePromptMethod, "enhancePrompt method not found")
	assert.Contains(t, enhancePromptMethod.Description, "Add schema information to prompt")
	assert.Len(t, enhancePromptMethod.Parameters, 3)

	// Check parameters
	assert.Equal(t, "prompt", enhancePromptMethod.Parameters[0].Name)
	assert.Equal(t, "string", enhancePromptMethod.Parameters[0].Type)
	assert.True(t, enhancePromptMethod.Parameters[0].Required)

	assert.Equal(t, "schema", enhancePromptMethod.Parameters[1].Name)
	assert.Equal(t, "object", enhancePromptMethod.Parameters[1].Type)
	assert.True(t, enhancePromptMethod.Parameters[1].Required)

	assert.Equal(t, "options", enhancePromptMethod.Parameters[2].Name)
	assert.Equal(t, "object", enhancePromptMethod.Parameters[2].Type)
	assert.False(t, enhancePromptMethod.Parameters[2].Required)

	assert.Equal(t, "string", enhancePromptMethod.ReturnType)
}

func TestUtilJSONBridge_ConvertFormatMethodSignature(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	var convertFormatMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "convertFormat" {
			convertFormatMethod = &m
			break
		}
	}

	assert.NotNil(t, convertFormatMethod, "convertFormat method not found")
	assert.Contains(t, convertFormatMethod.Description, "Convert between JSON, YAML, and XML")
	assert.Len(t, convertFormatMethod.Parameters, 4)

	// Check parameters
	assert.Equal(t, "data", convertFormatMethod.Parameters[0].Name)
	assert.Equal(t, "string", convertFormatMethod.Parameters[0].Type)
	assert.True(t, convertFormatMethod.Parameters[0].Required)

	assert.Equal(t, "fromFormat", convertFormatMethod.Parameters[1].Name)
	assert.Equal(t, "string", convertFormatMethod.Parameters[1].Type)
	assert.True(t, convertFormatMethod.Parameters[1].Required)

	assert.Equal(t, "toFormat", convertFormatMethod.Parameters[2].Name)
	assert.Equal(t, "string", convertFormatMethod.Parameters[2].Type)
	assert.True(t, convertFormatMethod.Parameters[2].Required)

	assert.Equal(t, "options", convertFormatMethod.Parameters[3].Name)
	assert.Equal(t, "object", convertFormatMethod.Parameters[3].Type)
	assert.False(t, convertFormatMethod.Parameters[3].Required)

	assert.Equal(t, "string", convertFormatMethod.ReturnType)
}

func TestUtilJSONBridge_V0_3_5_ComponentsInitialized(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()

	// Test that v0.3.5 components are properly initialized
	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// The bridge should have initialized these v0.3.5 components:
	// - processor (domain.Processor)
	// - promptEnhancer (domain.PromptEnhancer)
	// - validator (schemaDomain.Validator)
	// - converter (*outputs.Converter)
	// These are internal so we can't test them directly, but initialization success
	// indicates they were created properly
}

func TestUtilJSONBridge_StreamingCapabilitiesExposed(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	// Test that enhanced streaming capabilities are exposed
	streamingMethods := map[string]bool{
		"createEncoder": false,
		"createDecoder": false,
		"encodeStream":  false,
		"decodeStream":  false,
		"streamConvert": false, // New in v0.3.5
	}

	for _, method := range methods {
		if _, exists := streamingMethods[method.Name]; exists {
			streamingMethods[method.Name] = true
		}
	}

	for methodName, found := range streamingMethods {
		assert.True(t, found, "Streaming method %s not found in v0.3.5 implementation", methodName)
	}
}

// Note: Actual execution testing of v0.3.5 parser capabilities would require
// fully initialized go-llms components and real JSON/schema data
// or would be done at integration test level with actual utilities
