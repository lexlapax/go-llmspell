// ABOUTME: Test suite for the LLM bridge that wraps go-llms Provider functionality.
// ABOUTME: Tests bridge interface compliance and basic operations without mocking go-llms types.

package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Use go-llms testutils for consistency
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
)

func TestNewLLMBridge(t *testing.T) {
	bridge := NewLLMBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "llm", bridge.GetID())
}

func TestLLMBridgeMetadata(t *testing.T) {
	bridge := NewLLMBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "llm", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "schema validation support")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestLLMBridgeInitialization(t *testing.T) {
	bridge := NewLLMBridge()
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

func TestLLMBridgeMethods(t *testing.T) {
	bridge := NewLLMBridge()
	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := map[string]bool{
		"registerProvider":  false,
		"setActiveProvider": false,
		"generate":          false,
		"generateMessage":   false,
		"stream":            false,
		"streamMessage":     false,
		"listProviders":     false,
		"getActiveProvider": false,
		// Schema validation methods (v0.3.5)
		"generateWithSchema":        false,
		"registerSchema":            false,
		"getSchema":                 false,
		"listSchemas":               false,
		"validateWithSchema":        false,
		"generateSchemaFromExample": false,
		"clearSchemaCache":          false,
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

func TestLLMBridgeTypeMappings(t *testing.T) {
	bridge := NewLLMBridge()
	mappings := bridge.TypeMappings()

	// Check that all expected type mappings are present
	expectedTypes := []string{"Provider", "Message", "Response", "ProviderOptions", "Schema", "ValidationResult", "SchemaInfo"}
	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.Equal(t, typeName, mapping.GoType)
		assert.Equal(t, "object", mapping.ScriptType)
	}
}

func TestLLMBridgeRequiredPermissions(t *testing.T) {
	bridge := NewLLMBridge()
	permissions := bridge.RequiredPermissions()

	assert.Len(t, permissions, 1)
	assert.Equal(t, "network", string(permissions[0].Type))
	assert.Equal(t, "llm", permissions[0].Resource)
	assert.Contains(t, permissions[0].Actions, "access")
}

func TestLLMBridgeValidateMethod(t *testing.T) {
	bridge := NewLLMBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("generate", []interface{}{"prompt"})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", nil)
	assert.NoError(t, err)
}

func TestLLMBridgeProviderManagement(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Test listing providers when empty
	providers := bridge.ListProviders()
	assert.Empty(t, providers)

	// Test getting active provider when none set
	activeProvider := bridge.GetActiveProvider()
	assert.Nil(t, activeProvider)

	// Test setting active provider when no providers exist
	err = bridge.SetActiveProvider("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLLMBridgeSchemaValidation(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Use testutils fixture for test schema
	testState := fixtures.BasicTestState()
	testSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "number",
			},
		},
		"required": []string{"name"},
	}
	testState.Set("schema", testSchema)

	// Test registerSchema
	_, err = bridge.ExecuteMethod(ctx, "registerSchema", []interface{}{"person", testSchema})
	assert.NoError(t, err)

	// Test getSchema
	schema, err := bridge.ExecuteMethod(ctx, "getSchema", []interface{}{"person"})
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	// Test listSchemas
	schemas, err := bridge.ExecuteMethod(ctx, "listSchemas", []interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, schemas)

	// Test validateWithSchema
	testData := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	result, err := bridge.ExecuteMethod(ctx, "validateWithSchema", []interface{}{testData, testSchema})
	assert.NoError(t, err)
	validationResult, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.True(t, validationResult["valid"].(bool))

	// Test generateSchemaFromExample
	example := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		Name:  "Test User",
		Email: "test@example.com",
	}
	generatedSchema, err := bridge.ExecuteMethod(ctx, "generateSchemaFromExample", []interface{}{example})
	assert.NoError(t, err)
	assert.NotNil(t, generatedSchema)

	// Test clearSchemaCache
	_, err = bridge.ExecuteMethod(ctx, "clearSchemaCache", []interface{}{})
	assert.NoError(t, err)
}

func TestLLMBridgeWithTestutils(t *testing.T) {
	bridge := NewLLMBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Use testutils helpers for creating test contexts
	testContext := helpers.CreateTestToolContext()
	assert.NotNil(t, testContext)

	// Use fixtures for test data
	testMessages := fixtures.CreateSimpleConversation()
	assert.NotEmpty(t, testMessages)

	// These would be used in actual provider testing
	// For now, just verify the test infrastructure is available
}

// Note: Actual provider testing would require real go-llms Provider implementations
// or would be done at integration test level with actual LLM providers
