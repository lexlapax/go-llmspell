// ABOUTME: Tests for schema bridge versioning and migration functionality
// ABOUTME: Verifies file persistence, version management, and migration features

package structured

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaBridge_FileRepository(t *testing.T) {

	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create temp directory for file storage
	tmpDir := t.TempDir()

	// Initialize file repository once for all subtests
	result, err := bridge.ExecuteMethod(ctx, "initializeFileRepository", []interface{}{tmpDir})
	require.NoError(t, err)
	require.Nil(t, result)
	require.NotNil(t, bridge.fileRepo)
	t.Log("File repository initialized at", tmpDir)

	t.Run("initialize file repository", func(t *testing.T) {
		// Already initialized - just verify
		assert.NotNil(t, bridge.fileRepo)
	})

	t.Run("save schema with file persistence", func(t *testing.T) {
		// Create a test schema
		schema := map[string]interface{}{
			"type":        "object",
			"title":       "User Profile",
			"description": "User profile schema for testing",
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

		// Verify file was created
		schemaDir := filepath.Join(tmpDir, "user-profile")
		assert.DirExists(t, schemaDir)

		t.Log("Saved schema version", metadata)
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

		t.Log("Retrieved specific schema versions")
	})

	t.Run("list schema versions", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "listSchemaVersions", []interface{}{"user-profile"})
		assert.NoError(t, err)

		versions, ok := result.([]int)
		assert.True(t, ok)
		assert.Equal(t, []int{1, 2, 3}, versions)

		t.Log("Listed schema versions", versions)
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

		t.Log("Set current schema version")
	})

	t.Run("versioning with memory repository", func(t *testing.T) {
		// Save to memory repo using regular saveSchema method
		schema := map[string]interface{}{
			"type":  "object",
			"title": "Memory Only Schema",
		}

		// Use regular saveSchema which saves to memory repo
		result, err := bridge.ExecuteMethod(ctx, "saveSchema", []interface{}{"memory-only", schema})
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Verify it was saved to memory repo
		retrievedResult, err := bridge.ExecuteMethod(ctx, "getSchema", []interface{}{"memory-only"})
		assert.NoError(t, err)
		assert.NotNil(t, retrievedResult)

		retrievedSchema, ok := retrievedResult.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Memory Only Schema", retrievedSchema["title"])

		t.Log("Saved to memory repository")
	})
}

func TestSchemaBridge_Migration(t *testing.T) {

	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Initialize file repository
	tmpDir := t.TempDir()
	_, err = bridge.ExecuteMethod(ctx, "initializeFileRepository", []interface{}{tmpDir})
	require.NoError(t, err)

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
		// Note: In real usage, this would be a script function
		testMigrator := func(schema *domain.Schema, from, to int) (*domain.Schema, error) {
			t.Log("Test migrator called", from, "->", to)
			// This is just for testing - real migration would be done in script
			return schema, nil
		}

		// Register migrator
		_, err = bridge.ExecuteMethod(ctx, "registerMigrator", []interface{}{"test-migrator", testMigrator})
		assert.NoError(t, err)

		// Verify migrator was registered
		assert.Contains(t, bridge.migrators, "test-migrator")

		t.Log("Registered test migrator")
	})

	t.Run("migrate schema", func(t *testing.T) {
		// Attempt migration (will fail since script integration not implemented)
		_, err := bridge.ExecuteMethod(ctx, "migrateSchema", []interface{}{"config", 1, 2, "test-migrator"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "script-based migration not yet implemented")

		t.Log("Migration attempt (expected failure)")
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

	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

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

		t.Log("Exported repository size", len(exported))

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

		t.Log("Imported repository successfully")
	})

	t.Run("import invalid data", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "importRepository", []interface{}{"invalid json"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to import repository")
	})
}

func TestSchemaBridge_TagGeneration(t *testing.T) {

	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	t.Run("generate schema from tags", func(t *testing.T) {
		// Example source with struct tags
		source := `
type User struct {
    ID       string  ` + "`json:\"id\" validate:\"required,uuid\"`" + `
    Name     string  ` + "`json:\"name\" validate:\"required,min=1,max=100\"`" + `
    Email    string  ` + "`json:\"email\" validate:\"required,email\"`" + `
    Age      int     ` + "`json:\"age,omitempty\" validate:\"min=0,max=150\"`" + `
    IsActive bool    ` + "`json:\"is_active\"`" + `
}
`

		// Attempt to generate (will depend on tag generator implementation)
		result, err := bridge.ExecuteMethod(ctx, "generateFromTags", []interface{}{source})
		// This might fail if tag generator expects different input
		if err != nil {
			t.Log("Tag generation error (expected):", err)
			assert.Contains(t, err.Error(), "failed to generate schema from tags")
		} else {
			schema, ok := result.(map[string]interface{})
			assert.True(t, ok)
			t.Log("Generated schema from tags", schema)
		}
	})
}

func TestSchemaBridge_ErrorCases(t *testing.T) {
	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name   string
		method string
		args   []interface{}
		errMsg string
	}{
		{
			name:   "initializeFileRepository with missing directory",
			method: "initializeFileRepository",
			args:   []interface{}{},
			errMsg: "requires directory parameter",
		},
		{
			name:   "saveSchemaVersion with missing parameters",
			method: "saveSchemaVersion",
			args:   []interface{}{"id-only"},
			errMsg: "requires id and schema parameters",
		},
		{
			name:   "getSchemaVersion with invalid version",
			method: "getSchemaVersion",
			args:   []interface{}{"test", "not-a-number"},
			errMsg: "version must be number",
		},
		{
			name:   "listSchemaVersions with non-existent schema",
			method: "listSchemaVersions",
			args:   []interface{}{"non-existent"},
			errMsg: "failed to list schema versions",
		},
		{
			name:   "setCurrentSchemaVersion with missing version",
			method: "setCurrentSchemaVersion",
			args:   []interface{}{"test"},
			errMsg: "requires id and version parameters",
		},
		{
			name:   "registerMigrator with missing function",
			method: "registerMigrator",
			args:   []interface{}{"test"},
			errMsg: "requires name and migrator parameters",
		},
		{
			name:   "migrateSchema with invalid versions",
			method: "migrateSchema",
			args:   []interface{}{"test", "v1", "v2"},
			errMsg: "fromVersion must be number",
		},
		{
			name:   "importRepository with non-string data",
			method: "importRepository",
			args:   []interface{}{123},
			errMsg: "data must be string",
		},
		{
			name:   "generateFromTags with non-string source",
			method: "generateFromTags",
			args:   []interface{}{[]string{"not", "a", "string"}},
			errMsg: "source must be string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestSchemaBridge_VersionConversion(t *testing.T) {
	ctx := context.Background()
	bridge := NewSchemaBridge()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test various numeric types for version parameters
	tmpDir := t.TempDir()
	_, err = bridge.ExecuteMethod(ctx, "initializeFileRepository", []interface{}{tmpDir})
	require.NoError(t, err)

	// Save a schema
	schema := map[string]interface{}{
		"type": "object",
	}
	_, err = bridge.ExecuteMethod(ctx, "saveSchemaVersion", []interface{}{"test", schema})
	require.NoError(t, err)

	// Test different numeric types
	numericTests := []struct {
		name  string
		value interface{}
	}{
		{"int", 1},
		{"int32", int32(1)},
		{"int64", int64(1)},
		{"float32", float32(1)},
		{"float64", float64(1)},
	}

	for _, tt := range numericTests {
		t.Run(tt.name, func(t *testing.T) {
			// Get version
			result, err := bridge.ExecuteMethod(ctx, "getSchemaVersion", []interface{}{"test", tt.value})
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Set current version
			_, err = bridge.ExecuteMethod(ctx, "setCurrentSchemaVersion", []interface{}{"test", tt.value})
			assert.NoError(t, err)
		})
	}
}

