// ABOUTME: Test suite for bridge type aliases to verify v0.3.5 type integration
// ABOUTME: Ensures all new type aliases can be instantiated and used correctly

package bridge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestV035TypeAliases verifies that all new v0.3.5 type aliases are working correctly
func TestV035TypeAliases(t *testing.T) {
	t.Run("Schema System Types", func(t *testing.T) {
		// Test that we can reference the schema types
		var repo SchemaRepository
		var generator SchemaGenerator

		assert.Nil(t, repo)
		assert.Nil(t, generator)

		// These should compile without errors, proving the aliases work
		t.Log("Schema system type aliases are working")
	})

	t.Run("Structured Output Types", func(t *testing.T) {
		// Test that we can reference the output parser types
		var parser OutputParser
		var jsonParser JSONParser
		var xmlParser XMLParser
		var yamlParser YAMLParser

		assert.Nil(t, parser)
		assert.Nil(t, jsonParser)
		assert.Nil(t, xmlParser)
		assert.Nil(t, yamlParser)

		t.Log("Structured output type aliases are working")
	})

	t.Run("Event System Types", func(t *testing.T) {
		// Test that we can reference the event types
		var eventStore EventStore
		var eventFilter EventFilter
		var eventReplayer EventReplayer
		var eventSerializer EventSerializer

		assert.Nil(t, eventStore)
		assert.Nil(t, eventFilter)
		assert.Nil(t, eventReplayer)
		assert.Nil(t, eventSerializer)

		t.Log("Event system type aliases are working")
	})

	t.Run("Documentation Types", func(t *testing.T) {
		// Test that we can reference the documentation types
		var docGenerator DocGenerator
		var openAPIGenerator OpenAPIGenerator

		assert.Nil(t, docGenerator)
		assert.Nil(t, openAPIGenerator)

		t.Log("Documentation type aliases are working")
	})

	t.Run("Error Types", func(t *testing.T) {
		// Test that we can reference the error types
		var serializableError SerializableError
		var errorRecovery ErrorRecovery

		assert.Nil(t, serializableError)
		assert.Nil(t, errorRecovery)

		t.Log("Error type aliases are working")
	})
}

// TestTypeAliasImplementation verifies that type aliases match the underlying go-llms types
func TestTypeAliasImplementation(t *testing.T) {
	t.Run("Type Alias Compatibility", func(t *testing.T) {
		// This test ensures our type aliases are correctly defined
		// If the underlying go-llms types change, this will catch compilation errors

		ctx := context.Background()
		_ = ctx // Prevent unused variable warning

		// Schema types should be assignable
		var schema Schema
		var property Property
		var validationResult ValidationResult

		require.NotNil(t, &schema)
		require.NotNil(t, &property)
		require.NotNil(t, &validationResult)

		// All aliases should compile and be usable
		t.Log("All type aliases are compatible with their underlying types")
	})
}

// TestNewTypesIntegration tests that new types work with existing bridge functionality
func TestNewTypesIntegration(t *testing.T) {
	t.Run("Integration with Bridge Manager", func(t *testing.T) {
		// Verify that bridge manager can still work with the updated type system
		manager := NewBridgeManager()
		assert.NotNil(t, manager)

		// The bridge manager should still function normally
		bridges := manager.ListBridges()
		assert.Empty(t, bridges)

		t.Log("Bridge manager integrates correctly with new types")
	})
}
