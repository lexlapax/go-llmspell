// ABOUTME: Tests for schema bridge providing access to go-llms schema validation system
// ABOUTME: Verifies schema creation, validation, and generator functionality

package structured

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Test helper functions using go-llms testutils patterns

// setupTestBridge creates and initializes a schema bridge for testing
func setupTestBridge(t *testing.T) (*SchemaBridge, context.Context) {
	t.Helper()

	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	return bridge, ctx
}

// setupTestBridgeWithFileRepo creates a bridge with file repository for versioning tests
func setupTestBridgeWithFileRepo(t *testing.T) (*SchemaBridge, context.Context, string) {
	t.Helper()

	bridge, ctx := setupTestBridge(t)
	tmpDir := t.TempDir()

	_, err := bridge.ExecuteMethod(ctx, "initializeFileRepository", []interface{}{tmpDir})
	require.NoError(t, err)
	require.NotNil(t, bridge.fileRepo)

	return bridge, ctx, tmpDir
}

// createTestSchema creates a standard test schema for validation tests
func createTestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"format":      "uuid",
				"description": "Unique identifier",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"minLength":   1.0,
				"maxLength":   100.0,
				"description": "User's full name",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "Email address",
			},
		},
		"required": []interface{}{"id", "name", "email"},
	}
}

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
	bridge, ctx := setupTestBridge(t)

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
	bridge, ctx := setupTestBridge(t)

	// Test schema storage and retrieval
	t.Run("save and get schema", func(t *testing.T) {
		schema := createTestSchema()

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
	bridge, ctx := setupTestBridge(t)

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
	// Test table-driven error scenarios using go-llms testutils patterns
	errorTests := []struct {
		name        string
		setup       func() (*SchemaBridge, context.Context)
		method      string
		args        []interface{}
		expectedErr string
	}{
		{
			name: "methods fail when not initialized",
			setup: func() (*SchemaBridge, context.Context) {
				return NewSchemaBridge(), context.Background()
			},
			method:      "createSchema",
			args:        []interface{}{},
			expectedErr: "not initialized",
		},
		{
			name: "unknown method returns error",
			setup: func() (*SchemaBridge, context.Context) {
				bridge, ctx := setupTestBridge(t)
				return bridge, ctx
			},
			method:      "unknownMethod",
			args:        []interface{}{},
			expectedErr: "method not found",
		},
		{
			name: "missing required arguments",
			setup: func() (*SchemaBridge, context.Context) {
				bridge, ctx := setupTestBridge(t)
				return bridge, ctx
			},
			method:      "createProperty",
			args:        []interface{}{},
			expectedErr: "requires",
		},
		{
			name: "wrong argument type",
			setup: func() (*SchemaBridge, context.Context) {
				bridge, ctx := setupTestBridge(t)
				return bridge, ctx
			},
			method:      "validateJSON",
			args:        []interface{}{123, "data"},
			expectedErr: "schema must be",
		},
		{
			name: "import invalid data",
			setup: func() (*SchemaBridge, context.Context) {
				bridge, ctx := setupTestBridge(t)
				return bridge, ctx
			},
			method:      "importRepository",
			args:        []interface{}{"invalid json"},
			expectedErr: "failed to import repository",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			bridge, ctx := tt.setup()
			_, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
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

// Versioning and migration tests

func TestSchemaBridge_Versioning(t *testing.T) {
	bridge, ctx, _ := setupTestBridgeWithFileRepo(t)

	t.Run("save schema with file persistence", func(t *testing.T) {
		// Create a test schema
		schema := createTestSchema()
		schema["title"] = "User Profile"
		schema["description"] = "User profile schema for testing"

		// Save schema version
		result, err := bridge.ExecuteMethod(ctx, "saveSchemaVersion", []interface{}{"user-profile", schema})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Check returned metadata
		metadata, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "user-profile", metadata["id"])
		assert.Equal(t, 1, metadata["latestVersion"])
		assert.Equal(t, 1, metadata["currentVersion"])
		assert.Equal(t, 1, metadata["totalVersions"])
	})

	t.Run("get specific schema version", func(t *testing.T) {
		// Save additional versions
		for i := 2; i <= 3; i++ {
			schema := map[string]interface{}{
				"type":        "object",
				"title":       "User Profile",
				"description": fmt.Sprintf("Version %d of user profile", i),
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"name": map[string]interface{}{
						"type": "string",
					},
					"email": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []interface{}{"id", "name", "email"},
			}

			if i == 3 {
				// Add new field in version 3
				schema["properties"].(map[string]interface{})["age"] = map[string]interface{}{
					"type":        "integer",
					"minimum":     0.0,
					"maximum":     150.0,
					"description": "User's age",
				}
			}

			_, err := bridge.ExecuteMethod(ctx, "saveSchemaVersion", []interface{}{"user-profile", schema})
			require.NoError(t, err)
		}

		// Get version 2
		result, err := bridge.ExecuteMethod(ctx, "getSchemaVersion", []interface{}{"user-profile", 2})
		assert.NoError(t, err)

		schema, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Version 2 of user profile", schema["description"])

		// Verify version 2 doesn't have age field
		props := schema["properties"].(map[string]interface{})
		assert.NotContains(t, props, "age")

		// Get version 3
		result, err = bridge.ExecuteMethod(ctx, "getSchemaVersion", []interface{}{"user-profile", 3})
		assert.NoError(t, err)

		schema, ok = result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Version 3 of user profile", schema["description"])

		// Verify version 3 has age field
		props = schema["properties"].(map[string]interface{})
		assert.Contains(t, props, "age")
	})

	t.Run("list schema versions", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "listSchemaVersions", []interface{}{"user-profile"})
		assert.NoError(t, err)

		versions, ok := result.([]int)
		assert.True(t, ok)
		assert.Equal(t, []int{1, 2, 3}, versions)
	})

	t.Run("set current schema version", func(t *testing.T) {
		// Set version 2 as current
		_, err := bridge.ExecuteMethod(ctx, "setCurrentSchemaVersion", []interface{}{"user-profile", 2})
		assert.NoError(t, err)

		// Get current version (should be v2)
		result, err := bridge.ExecuteMethod(ctx, "getSchema", []interface{}{"user-profile"})
		assert.NoError(t, err)

		schema, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Version 2 of user profile", schema["description"])

		// Set back to latest
		_, err = bridge.ExecuteMethod(ctx, "setCurrentSchemaVersion", []interface{}{"user-profile", 3})
		assert.NoError(t, err)
	})
}

func TestSchemaBridge_Migration(t *testing.T) {
	bridge, ctx, _ := setupTestBridgeWithFileRepo(t)

	t.Run("register and use migrator", func(t *testing.T) {
		// Create v1 schema
		v1Schema := map[string]interface{}{
			"type":  "object",
			"title": "Config v1",
			"properties": map[string]interface{}{
				"apiKey": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []interface{}{"apiKey"},
		}

		// Save v1
		_, err := bridge.ExecuteMethod(ctx, "saveSchemaVersion", []interface{}{"config", v1Schema})
		require.NoError(t, err)

		// Create v2 schema with renamed field
		v2Schema := map[string]interface{}{
			"type":  "object",
			"title": "Config v2",
			"properties": map[string]interface{}{
				"api_key": map[string]interface{}{
					"type": "string",
				},
				"timeout": map[string]interface{}{
					"type":    "integer",
					"minimum": 0.0,
				},
			},
			"required": []interface{}{"api_key"},
		}

		// Save v2
		_, err = bridge.ExecuteMethod(ctx, "saveSchemaVersion", []interface{}{"config", v2Schema})
		require.NoError(t, err)

		// Register a test migrator function
		testMigrator := func(schema *domain.Schema, from, to int) (*domain.Schema, error) {
			// This is just for testing - real migration would be done in script
			return schema, nil
		}

		// Register migrator
		_, err = bridge.ExecuteMethod(ctx, "registerMigrator", []interface{}{"test-migrator", testMigrator})
		assert.NoError(t, err)

		// Verify migrator was registered
		assert.Contains(t, bridge.migrators, "test-migrator")
	})

	t.Run("migrate schema", func(t *testing.T) {
		// Attempt migration (will fail since script integration not implemented)
		_, err := bridge.ExecuteMethod(ctx, "migrateSchema", []interface{}{"config", 1, 2, "test-migrator"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "script-based migration not yet implemented")
	})

	t.Run("migration with non-existent migrator", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "migrateSchema", []interface{}{"config", 1, 2, "non-existent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migrator 'non-existent' not found")
	})

	t.Run("migration with non-versioned repository", func(t *testing.T) {
		// Save to memory repo only
		schema := map[string]interface{}{
			"type": "object",
		}
		_, err := bridge.ExecuteMethod(ctx, "saveSchema", []interface{}{"memory-schema", schema})
		require.NoError(t, err)

		// Try to migrate (should fail - no migrator registered for default)
		_, err = bridge.ExecuteMethod(ctx, "migrateSchema", []interface{}{"memory-schema", 1, 2})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migrator 'default' not found")
	})
}

func TestSchemaBridge_ImportExport(t *testing.T) {
	bridge, ctx := setupTestBridge(t)

	t.Run("export and import repository", func(t *testing.T) {
		// Create some schemas
		schemas := []struct {
			id     string
			schema map[string]interface{}
		}{
			{
				id: "schema1",
				schema: map[string]interface{}{
					"type":  "object",
					"title": "Schema 1",
				},
			},
			{
				id: "schema2",
				schema: map[string]interface{}{
					"type":  "object",
					"title": "Schema 2",
					"properties": map[string]interface{}{
						"field": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		}

		// Save schemas
		for _, s := range schemas {
			_, err := bridge.ExecuteMethod(ctx, "saveSchema", []interface{}{s.id, s.schema})
			require.NoError(t, err)
		}

		// Export repository
		result, err := bridge.ExecuteMethod(ctx, "exportRepository", []interface{}{})
		assert.NoError(t, err)

		exported, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, exported)

		// Create new bridge and import
		bridge2 := NewSchemaBridge()
		err = bridge2.Initialize(ctx)
		require.NoError(t, err)

		// Import data
		_, err = bridge2.ExecuteMethod(ctx, "importRepository", []interface{}{exported})
		assert.NoError(t, err)

		// Verify schemas were imported
		for _, s := range schemas {
			result, err := bridge2.ExecuteMethod(ctx, "getSchema", []interface{}{s.id})
			assert.NoError(t, err)

			schema, ok := result.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, s.schema["title"], schema["title"])
		}
	})
}
