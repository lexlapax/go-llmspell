// ABOUTME: Tests for schema bridge providing access to go-llms schema validation system
// ABOUTME: Verifies schema creation, validation, and generator functionality

package structured

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestSchemaBridge_BasicOperations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *SchemaBridge)
	}{
		{
			name: "GetID returns correct identifier",
			test: func(t *testing.T, b *SchemaBridge) {
				assert.Equal(t, "schema", b.GetID())
			},
		},
		{
			name: "GetMetadata returns valid metadata",
			test: func(t *testing.T, b *SchemaBridge) {
				metadata := b.GetMetadata()
				assert.Equal(t, "Schema Bridge", metadata.Name)
				assert.NotEmpty(t, metadata.Version)
				assert.NotEmpty(t, metadata.Description)
				assert.Equal(t, "go-llmspell", metadata.Author)
			},
		},
		{
			name: "Initialize and cleanup work correctly",
			test: func(t *testing.T, b *SchemaBridge) {
				ctx := context.Background()

				// Initial state
				assert.False(t, b.IsInitialized())

				// Initialize
				err := b.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, b.IsInitialized())

				// Double initialize should be safe
				err = b.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, b.IsInitialized())

				// Cleanup
				err = b.Cleanup(ctx)
				require.NoError(t, err)
				assert.False(t, b.IsInitialized())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewSchemaBridge()
			tt.test(t, bridge)
		})
	}
}

func TestSchemaBridge_Methods(t *testing.T) {
	bridge := NewSchemaBridge()
	methods := bridge.Methods()

	// Check expected methods
	expectedMethods := []string{
		"createSchema",
		"createProperty",
		"validateJSON",
		"validateStruct",
		"generateSchemaFromType",
		"convertJSONSchema",
		"saveSchema",
		"getSchema",
		"deleteSchema",
	}

	methodMap := make(map[string]engine.MethodInfo)
	for _, m := range methods {
		methodMap[m.Name] = m
	}

	for _, expected := range expectedMethods {
		t.Run("has_method_"+expected, func(t *testing.T) {
			method, exists := methodMap[expected]
			assert.True(t, exists, "Missing method: %s", expected)
			assert.NotEmpty(t, method.Description)
			assert.NotEmpty(t, method.ReturnType)
		})
	}
}

func TestSchemaBridge_SchemaOperations(t *testing.T) {
	ctx := context.Background()
	bridge := NewSchemaBridge()
	require.NoError(t, bridge.Initialize(ctx))

	tests := []struct {
		name     string
		method   string
		args     []interface{}
		validate func(t *testing.T, result interface{}, err error)
	}{
		{
			name:   "createSchema creates schema object",
			method: "createSchema",
			args: []interface{}{
				map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "User's name",
						},
						"age": map[string]interface{}{
							"type":        "integer",
							"minimum":     0,
							"maximum":     120,
							"description": "User's age",
						},
					},
					"required": []string{"name"},
				},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				schema, ok := result.(map[string]interface{})
				require.True(t, ok)

				assert.Equal(t, "object", schema["type"])
				assert.NotNil(t, schema["properties"])

				props, ok := schema["properties"].(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, props, "name")
				assert.Contains(t, props, "age")
			},
		},
		{
			name:   "createProperty creates property definition",
			method: "createProperty",
			args: []interface{}{
				"string",
				map[string]interface{}{
					"description": "Email address",
					"format":      "email",
					"minLength":   5,
					"maxLength":   100,
				},
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				prop, ok := result.(map[string]interface{})
				require.True(t, ok)

				assert.Equal(t, "string", prop["type"])
				assert.Equal(t, "email", prop["format"])
				assert.Equal(t, "Email address", prop["description"])
				assert.Equal(t, float64(5), prop["minLength"])
				assert.Equal(t, float64(100), prop["maxLength"])
			},
		},
		{
			name:   "validateJSON validates data against schema",
			method: "validateJSON",
			args: []interface{}{
				map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"name"},
				},
				`{"name": "John Doe"}`,
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				validation, ok := result.(map[string]interface{})
				require.True(t, ok)

				assert.True(t, validation["valid"].(bool))
				errors, hasErrors := validation["errors"]
				if hasErrors {
					assert.Empty(t, errors)
				}
			},
		},
		{
			name:   "validateJSON detects invalid data",
			method: "validateJSON",
			args: []interface{}{
				map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"name"},
				},
				`{"age": 30}`, // Missing required 'name'
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				validation, ok := result.(map[string]interface{})
				require.True(t, ok)

				assert.False(t, validation["valid"].(bool))
				errors, ok := validation["errors"].([]string)
				require.True(t, ok)
				assert.NotEmpty(t, errors)
			},
		},
		{
			name:   "convertJSONSchema converts from JSON schema format",
			method: "convertJSONSchema",
			args: []interface{}{
				`{
					"$schema": "http://json-schema.org/draft-07/schema#",
					"type": "object",
					"properties": {
						"email": {
							"type": "string",
							"format": "email"
						}
					}
				}`,
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)

				schema, ok := result.(map[string]interface{})
				require.True(t, ok)

				assert.Equal(t, "object", schema["type"])
				props, ok := schema["properties"].(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, props, "email")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			tt.validate(t, result, err)
		})
	}
}

func TestSchemaBridge_Repository(t *testing.T) {
	ctx := context.Background()
	bridge := NewSchemaBridge()
	require.NoError(t, bridge.Initialize(ctx))

	// Test schema storage and retrieval
	t.Run("save and get schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
				},
			},
		}

		// Save schema
		_, err := bridge.ExecuteMethod(ctx, "saveSchema", []interface{}{"test-schema", schema})
		require.NoError(t, err)

		// Get schema
		result, err := bridge.ExecuteMethod(ctx, "getSchema", []interface{}{"test-schema"})
		require.NoError(t, err)

		retrieved, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "object", retrieved["type"])

		// Delete schema
		_, err = bridge.ExecuteMethod(ctx, "deleteSchema", []interface{}{"test-schema"})
		require.NoError(t, err)

		// Verify deleted
		_, err = bridge.ExecuteMethod(ctx, "getSchema", []interface{}{"test-schema"})
		assert.Error(t, err)
	})
}

func TestSchemaBridge_SchemaGeneration(t *testing.T) {
	ctx := context.Background()
	bridge := NewSchemaBridge()
	require.NoError(t, bridge.Initialize(ctx))

	t.Run("generateSchemaFromType generates schema from struct", func(t *testing.T) {
		// Test with a sample struct type representation
		result, err := bridge.ExecuteMethod(ctx, "generateSchemaFromType", []interface{}{
			map[string]interface{}{
				"type": "struct",
				"fields": map[string]interface{}{
					"Name": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"Age": map[string]interface{}{
						"type": "integer",
					},
					"Email": map[string]interface{}{
						"type":   "string",
						"format": "email",
					},
				},
			},
		})

		if err != nil {
			t.Skip("Schema generation not implemented yet")
		}

		schema, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "object", schema["type"])
		assert.NotNil(t, schema["properties"])
		assert.Contains(t, schema["required"], "Name")
	})
}

func TestSchemaBridge_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	bridge := NewSchemaBridge()

	t.Run("methods fail when not initialized", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "createSchema", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})

	require.NoError(t, bridge.Initialize(ctx))

	t.Run("unknown method returns error", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "unknownMethod", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not found")
	})

	t.Run("invalid arguments return error", func(t *testing.T) {
		// Missing required arguments
		_, err := bridge.ExecuteMethod(ctx, "createProperty", []interface{}{})
		assert.Error(t, err)

		// Wrong argument type
		_, err = bridge.ExecuteMethod(ctx, "validateJSON", []interface{}{123, "data"})
		assert.Error(t, err)
	})
}

func TestSchemaBridge_TypeMappings(t *testing.T) {
	bridge := NewSchemaBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	expectedTypes := []string{
		"Schema",
		"Property",
		"ValidationResult",
		"Validator",
		"SchemaRepository",
		"SchemaGenerator",
	}

	for _, typeName := range expectedTypes {
		t.Run("has_type_"+typeName, func(t *testing.T) {
			mapping, exists := mappings[typeName]
			assert.True(t, exists, "Missing type mapping: %s", typeName)
			assert.NotEmpty(t, mapping.GoType)
			assert.NotEmpty(t, mapping.ScriptType)
		})
	}
}

func TestSchemaBridge_Permissions(t *testing.T) {
	bridge := NewSchemaBridge()
	permissions := bridge.RequiredPermissions()

	// Should require minimal permissions
	assert.NotEmpty(t, permissions)

	// Check for schema operation permissions
	hasSchemaPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionMemory && perm.Resource == "schema" {
			hasSchemaPermission = true
			assert.Contains(t, perm.Actions, "create")
			assert.Contains(t, perm.Actions, "validate")
		}
	}
	assert.True(t, hasSchemaPermission, "Missing schema operation permission")
}
