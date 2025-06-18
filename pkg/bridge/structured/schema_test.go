// ABOUTME: Tests for schema bridge providing access to go-llms schema validation system
// ABOUTME: Verifies schema creation, validation, and generator functionality with ScriptValue support

package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// Test helper functions using ScriptValue patterns

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

	args := []engine.ScriptValue{engine.NewStringValue(tmpDir)}
	_, err := bridge.ExecuteMethod(ctx, "initializeFileRepository", args)
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
				"type":   "string",
				"format": "uuid",
			},
			"name": map[string]interface{}{
				"type":      "string",
				"minLength": 1,
				"maxLength": 100,
			},
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
		},
		"required": []interface{}{"id", "name"},
	}
}

// createTestData creates valid test data matching the test schema
func createTestData() map[string]interface{} {
	return map[string]interface{}{
		"id":    "123e4567-e89b-12d3-a456-426614174000",
		"name":  "John Doe",
		"email": "john@example.com",
	}
}

// TestSchemaBridge_BasicOperations tests basic bridge lifecycle
func TestSchemaBridge_BasicOperations(t *testing.T) {
	bridge := NewSchemaBridge()

	// Test GetID
	assert.Equal(t, "schema", bridge.GetID())

	// Test GetMetadata
	metadata := bridge.GetMetadata()
	assert.Equal(t, "Schema Bridge", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "schema validation")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
	assert.Len(t, metadata.Dependencies, 4)

	// Test IsInitialized before initialization
	assert.False(t, bridge.IsInitialized())

	// Test Initialize
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test double initialization (should not error)
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test Cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

// TestSchemaBridge_Methods tests the method registry
func TestSchemaBridge_Methods(t *testing.T) {
	bridge := NewSchemaBridge()
	methods := bridge.Methods()

	// Should have all 41 methods
	assert.Len(t, methods, 41)

	// Check core methods are present
	coreMethodNames := []string{
		"createSchema", "createProperty", "validateJSON", "validateStruct",
		"generateSchemaFromType", "convertJSONSchema",
		"saveSchema", "getSchema", "deleteSchema",
		"initializeFileRepository", "saveSchemaVersion", "getSchemaVersion",
		"listSchemaVersions", "setCurrentSchemaVersion", "registerMigrator",
		"migrateSchema", "exportRepository", "importRepository",
		"generateFromTags", "setTagPriority", "registerTagParser",
		"extractValidationRules", "generateWithDocumentation",
		"exportToJSONSchema", "exportToOpenAPI", "importFromFile",
		"importFromString", "convertFormat", "mergeSchemas",
		"generateDiff", "exportCollection", "importCollection",
		"registerCustomValidator", "unregisterCustomValidator",
		"listCustomValidators", "validateWithCustom", "validateAsync",
		"getValidationMetrics", "clearValidationCache",
		"registerConditionalValidator", "validateConditional",
	}

	foundMethods := make(map[string]bool)
	for _, method := range methods {
		foundMethods[method.Name] = true
		assert.NotEmpty(t, method.Description)
		assert.NotEmpty(t, method.ReturnType)
	}

	for _, methodName := range coreMethodNames {
		assert.True(t, foundMethods[methodName], "Method %s should be present", methodName)
	}
}

// TestSchemaBridge_ValidateMethod tests parameter validation
func TestSchemaBridge_ValidateMethod(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	tests := []struct {
		name        string
		methodName  string
		args        []engine.ScriptValue
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid createSchema call",
			methodName:  "createSchema",
			args:        []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(createTestSchema()))},
			shouldError: false,
		},
		{
			name:        "missing required parameter",
			methodName:  "createSchema",
			args:        []engine.ScriptValue{},
			shouldError: true,
			errorMsg:    "requires at least 1 arguments",
		},
		{
			name:        "unknown method",
			methodName:  "unknownMethod",
			args:        []engine.ScriptValue{},
			shouldError: true,
			errorMsg:    "unknown method",
		},
		{
			name:        "uninitialized bridge",
			methodName:  "createSchema",
			args:        []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(createTestSchema()))},
			shouldError: true,
			errorMsg:    "not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with uninitialized bridge for the uninitialized test case
			testBridge := bridge
			if tt.name == "uninitialized bridge" {
				testBridge = NewSchemaBridge()
			}

			err := testBridge.ValidateMethod(tt.methodName, tt.args)
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSchemaBridge_SchemaOperations tests core schema functionality
func TestSchemaBridge_SchemaOperations(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("createSchema", func(t *testing.T) {
		schemaData := createTestSchema()
		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData))}

		result, err := bridge.ExecuteMethod(ctx, "createSchema", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.True(t, resultObj["created"].(bool))
		assert.NotNil(t, resultObj["schema"])
		assert.NotNil(t, resultObj["timestamp"])
	})

	t.Run("createProperty", func(t *testing.T) {
		constraints := map[string]interface{}{
			"minLength": 1,
			"maxLength": 100,
		}
		args := []engine.ScriptValue{
			engine.NewStringValue("string"),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(constraints)),
		}

		result, err := bridge.ExecuteMethod(ctx, "createProperty", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.Equal(t, "string", resultObj["type"])

		// Check constraints - numbers may be converted to float64
		constraintsResult := resultObj["constraints"].(map[string]interface{})
		assert.Equal(t, float64(1), constraintsResult["minLength"])
		assert.Equal(t, float64(100), constraintsResult["maxLength"])
	})

	t.Run("validateJSON", func(t *testing.T) {
		schemaData := createTestSchema()
		testData := createTestData()

		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(testData)),
		}

		result, err := bridge.ExecuteMethod(ctx, "validateJSON", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.True(t, resultObj["valid"].(bool))
		assert.NotNil(t, resultObj["errors"])
		assert.NotNil(t, resultObj["schema"])
		assert.Equal(t, testData, resultObj["data"])
	})

	t.Run("validateStruct", func(t *testing.T) {
		schemaData := createTestSchema()
		testData := createTestData()

		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(testData)),
		}

		result, err := bridge.ExecuteMethod(ctx, "validateStruct", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.True(t, resultObj["valid"].(bool))
	})

	t.Run("convertJSONSchema", func(t *testing.T) {
		schemaData := createTestSchema()
		jsonSchema, err := json.Marshal(schemaData)
		require.NoError(t, err)

		args := []engine.ScriptValue{engine.NewStringValue(string(jsonSchema))}

		result, err := bridge.ExecuteMethod(ctx, "convertJSONSchema", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.True(t, resultObj["converted"].(bool))
		assert.Equal(t, "json", resultObj["source"])
		assert.NotNil(t, resultObj["schema"])
	})
}

// TestSchemaBridge_Repository tests schema storage operations
func TestSchemaBridge_Repository(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	schemaData := createTestSchema()
	schemaName := "test-schema"

	t.Run("saveSchema", func(t *testing.T) {
		args := []engine.ScriptValue{
			engine.NewStringValue(schemaName),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
		}

		result, err := bridge.ExecuteMethod(ctx, "saveSchema", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("getSchema", func(t *testing.T) {
		args := []engine.ScriptValue{engine.NewStringValue(schemaName)}

		result, err := bridge.ExecuteMethod(ctx, "getSchema", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.Equal(t, schemaName, resultObj["name"])
		assert.True(t, resultObj["found"].(bool))
		assert.NotNil(t, resultObj["schema"])
	})

	t.Run("deleteSchema", func(t *testing.T) {
		args := []engine.ScriptValue{engine.NewStringValue(schemaName)}

		result, err := bridge.ExecuteMethod(ctx, "deleteSchema", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())

		// Verify schema is deleted
		result, err = bridge.ExecuteMethod(ctx, "getSchema", args)
		require.NoError(t, err)
		// Should return error value for missing schema
		assert.Equal(t, engine.TypeError, result.Type())
	})
}

// TestSchemaBridge_GenerationMethods tests schema generation functionality
func TestSchemaBridge_GenerationMethods(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("generateSchemaFromType", func(t *testing.T) {
		typeInfo := map[string]interface{}{
			"type": "object",
			"name": "User",
			"fields": map[string]interface{}{
				"id":   "string",
				"name": "string",
			},
		}

		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(typeInfo))}

		result, err := bridge.ExecuteMethod(ctx, "generateSchemaFromType", args)
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.True(t, resultObj["generated"].(bool))
		assert.Equal(t, "type", resultObj["source"])
		assert.NotNil(t, resultObj["schema"])
	})
}

// TestSchemaBridge_VersioningMethods tests versioning functionality
func TestSchemaBridge_VersioningMethods(t *testing.T) {
	bridge, ctx, _ := setupTestBridgeWithFileRepo(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	schemaData := createTestSchema()
	schemaName := "versioned-schema"

	t.Run("saveSchemaVersion", func(t *testing.T) {
		args := []engine.ScriptValue{
			engine.NewStringValue(schemaName),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
			engine.NewNumberValue(1),
		}

		result, err := bridge.ExecuteMethod(ctx, "saveSchemaVersion", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("getSchemaVersion", func(t *testing.T) {
		args := []engine.ScriptValue{
			engine.NewStringValue(schemaName),
			engine.NewNumberValue(1),
		}

		result, err := bridge.ExecuteMethod(ctx, "getSchemaVersion", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type()) // Stub implementation
	})

	t.Run("listSchemaVersions", func(t *testing.T) {
		args := []engine.ScriptValue{engine.NewStringValue(schemaName)}

		result, err := bridge.ExecuteMethod(ctx, "listSchemaVersions", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeArray, result.Type())
	})

	t.Run("setCurrentSchemaVersion", func(t *testing.T) {
		args := []engine.ScriptValue{
			engine.NewStringValue(schemaName),
			engine.NewNumberValue(1),
		}

		result, err := bridge.ExecuteMethod(ctx, "setCurrentSchemaVersion", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})
}

// TestSchemaBridge_MigrationMethods tests migration functionality
func TestSchemaBridge_MigrationMethods(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("registerMigrator", func(t *testing.T) {
		migrator := map[string]interface{}{
			"name":        "test-migrator",
			"fromVersion": 1,
			"toVersion":   2,
		}

		args := []engine.ScriptValue{
			engine.NewStringValue("test-migrator"),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(migrator)),
		}

		result, err := bridge.ExecuteMethod(ctx, "registerMigrator", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("migrateSchema", func(t *testing.T) {
		args := []engine.ScriptValue{
			engine.NewStringValue("test-schema"),
			engine.NewNumberValue(1),
			engine.NewNumberValue(2),
		}

		result, err := bridge.ExecuteMethod(ctx, "migrateSchema", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})
}

// TestSchemaBridge_ImportExportMethods tests import/export functionality
func TestSchemaBridge_ImportExportMethods(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("exportRepository", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "exportRepository", []engine.ScriptValue{})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("importRepository", func(t *testing.T) {
		data := map[string]interface{}{
			"schemas": map[string]interface{}{},
			"version": "1.0",
		}

		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(data))}

		result, err := bridge.ExecuteMethod(ctx, "importRepository", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("exportToJSONSchema", func(t *testing.T) {
		schemaData := createTestSchema()
		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData))}

		result, err := bridge.ExecuteMethod(ctx, "exportToJSONSchema", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("exportToOpenAPI", func(t *testing.T) {
		schemaData := createTestSchema()
		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData))}

		result, err := bridge.ExecuteMethod(ctx, "exportToOpenAPI", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("importFromString", func(t *testing.T) {
		schemaData := createTestSchema()
		jsonSchema, err := json.Marshal(schemaData)
		require.NoError(t, err)

		args := []engine.ScriptValue{
			engine.NewStringValue(string(jsonSchema)),
			engine.NewStringValue("jsonschema"),
		}

		result, err := bridge.ExecuteMethod(ctx, "importFromString", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("convertFormat", func(t *testing.T) {
		schemaData := createTestSchema()
		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
			engine.NewStringValue("internal"),
			engine.NewStringValue("jsonschema"),
		}

		result, err := bridge.ExecuteMethod(ctx, "convertFormat", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("mergeSchemas", func(t *testing.T) {
		schemas := []interface{}{
			createTestSchema(),
			createTestSchema(),
		}
		args := []engine.ScriptValue{
			engine.NewArrayValue(engine.ConvertSliceToScriptValue(schemas)),
			engine.NewStringValue("union"),
		}

		result, err := bridge.ExecuteMethod(ctx, "mergeSchemas", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("generateDiff", func(t *testing.T) {
		schema1 := createTestSchema()
		schema2 := createTestSchema()
		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schema1)),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schema2)),
		}

		result, err := bridge.ExecuteMethod(ctx, "generateDiff", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("exportCollection", func(t *testing.T) {
		schemaIds := []interface{}{"schema1", "schema2"}
		args := []engine.ScriptValue{
			engine.NewArrayValue(engine.ConvertSliceToScriptValue(schemaIds)),
			engine.NewStringValue("bundle"),
		}

		result, err := bridge.ExecuteMethod(ctx, "exportCollection", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("importCollection", func(t *testing.T) {
		collection := map[string]interface{}{
			"schemas": map[string]interface{}{},
			"format":  "bundle",
		}
		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(collection)),
			engine.NewBoolValue(false),
		}

		result, err := bridge.ExecuteMethod(ctx, "importCollection", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})
}

// TestSchemaBridge_TagMethods tests tag-based generation
func TestSchemaBridge_TagMethods(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("generateFromTags", func(t *testing.T) {
		structData := map[string]interface{}{
			"type": "struct",
			"tags": map[string]interface{}{
				"json":     "name,omitempty",
				"validate": "required,min=1",
			},
		}

		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(structData))}

		result, err := bridge.ExecuteMethod(ctx, "generateFromTags", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("setTagPriority", func(t *testing.T) {
		tags := []interface{}{"json", "validate", "schema"}
		args := []engine.ScriptValue{engine.NewArrayValue(engine.ConvertSliceToScriptValue(tags))}

		result, err := bridge.ExecuteMethod(ctx, "setTagPriority", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("registerTagParser", func(t *testing.T) {
		parser := map[string]interface{}{
			"name":    "custom",
			"pattern": "^custom:",
		}
		args := []engine.ScriptValue{
			engine.NewStringValue("custom"),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(parser)),
		}

		result, err := bridge.ExecuteMethod(ctx, "registerTagParser", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("extractValidationRules", func(t *testing.T) {
		structData := map[string]interface{}{
			"field": "name",
			"tags":  "required,min=1,max=100",
		}
		args := []engine.ScriptValue{engine.NewObjectValue(engine.ConvertMapToScriptValue(structData))}

		result, err := bridge.ExecuteMethod(ctx, "extractValidationRules", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("generateWithDocumentation", func(t *testing.T) {
		structData := map[string]interface{}{
			"type": "struct",
			"docs": true,
		}
		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(structData)),
			engine.NewBoolValue(true),
		}

		result, err := bridge.ExecuteMethod(ctx, "generateWithDocumentation", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})
}

// TestSchemaBridge_CustomValidationMethods tests custom validation functionality
func TestSchemaBridge_CustomValidationMethods(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("registerCustomValidator", func(t *testing.T) {
		validator := map[string]interface{}{
			"name":        "email-validator",
			"description": "Custom email validation",
		}
		args := []engine.ScriptValue{
			engine.NewStringValue("email-validator"),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(validator)),
		}

		result, err := bridge.ExecuteMethod(ctx, "registerCustomValidator", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("listCustomValidators", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "listCustomValidators", []engine.ScriptValue{})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeArray, result.Type())
	})

	t.Run("validateWithCustom", func(t *testing.T) {
		data := map[string]interface{}{
			"email": "test@example.com",
		}
		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(data)),
			engine.NewStringValue("email-validator"),
		}

		result, err := bridge.ExecuteMethod(ctx, "validateWithCustom", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("unregisterCustomValidator", func(t *testing.T) {
		args := []engine.ScriptValue{engine.NewStringValue("email-validator")}

		result, err := bridge.ExecuteMethod(ctx, "unregisterCustomValidator", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("validateAsync", func(t *testing.T) {
		schemaData := createTestSchema()
		testData := createTestData()
		callback := map[string]interface{}{
			"name": "validation-callback",
		}

		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(testData)),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(callback)),
		}

		result, err := bridge.ExecuteMethod(ctx, "validateAsync", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("registerConditionalValidator", func(t *testing.T) {
		validator := map[string]interface{}{
			"name":      "age-validator",
			"condition": "age > 18",
		}
		args := []engine.ScriptValue{
			engine.NewStringValue("age-validator"),
			engine.NewObjectValue(engine.ConvertMapToScriptValue(validator)),
		}

		result, err := bridge.ExecuteMethod(ctx, "registerConditionalValidator", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})

	t.Run("validateConditional", func(t *testing.T) {
		data := map[string]interface{}{
			"age": 25,
		}
		args := []engine.ScriptValue{
			engine.NewObjectValue(engine.ConvertMapToScriptValue(data)),
			engine.NewStringValue("age-validator"),
		}

		result, err := bridge.ExecuteMethod(ctx, "validateConditional", args)
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})
}

// TestSchemaBridge_MetricsAndCache tests metrics and caching functionality
func TestSchemaBridge_MetricsAndCache(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("getValidationMetrics", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "getValidationMetrics", []engine.ScriptValue{})
		require.NoError(t, err)
		require.Equal(t, engine.TypeObject, result.Type())

		resultObj := result.(engine.ObjectValue).ToGo().(map[string]interface{})
		assert.Contains(t, resultObj, "totalValidations")
		assert.Contains(t, resultObj, "successfulValidations")
		assert.Contains(t, resultObj, "failedValidations")
		assert.Contains(t, resultObj, "averageLatency")
		assert.Contains(t, resultObj, "cacheHits")
		assert.Contains(t, resultObj, "cacheMisses")
		assert.Contains(t, resultObj, "asyncValidations")
	})

	t.Run("clearValidationCache", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "clearValidationCache", []engine.ScriptValue{})
		require.NoError(t, err)
		assert.Equal(t, engine.TypeNil, result.Type())
	})
}

// TestSchemaBridge_ErrorHandling tests error scenarios
func TestSchemaBridge_ErrorHandling(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	tests := []struct {
		name          string
		method        string
		args          []engine.ScriptValue
		expectError   bool
		errorContains string
	}{
		{
			name:          "createSchema with wrong type",
			method:        "createSchema",
			args:          []engine.ScriptValue{engine.NewStringValue("not-an-object")},
			expectError:   true,
			errorContains: "expected object",
		},
		{
			name:          "saveSchema missing name",
			method:        "saveSchema",
			args:          []engine.ScriptValue{},
			expectError:   true,
			errorContains: "requires at least 2 arguments",
		},
		{
			name:          "getSchema with wrong type",
			method:        "getSchema",
			args:          []engine.ScriptValue{engine.NewNumberValue(123)},
			expectError:   true,
			errorContains: "expected string",
		},
		{
			name:          "validateJSON with missing schema",
			method:        "validateJSON",
			args:          []engine.ScriptValue{engine.NewStringValue("not-schema")},
			expectError:   true,
			errorContains: "requires at least 2 arguments",
		},
		{
			name:          "unknown method",
			method:        "unknownMethod",
			args:          []engine.ScriptValue{},
			expectError:   true,
			errorContains: "unknown method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, tt.method, tt.args)

			if tt.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tt.errorContains)
				} else {
					// Check if result is an error value
					assert.Equal(t, engine.TypeError, result.Type())
					errorValue := result.(engine.ErrorValue)
					assert.Contains(t, errorValue.Error().Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, engine.TypeError, result.Type())
			}
		})
	}
}

// TestSchemaBridge_TypeMappings tests type mapping configuration
func TestSchemaBridge_TypeMappings(t *testing.T) {
	bridge := NewSchemaBridge()
	mappings := bridge.TypeMappings()

	expectedMappings := []string{"schema", "validationResult"}
	assert.Len(t, mappings, len(expectedMappings))

	for _, expectedType := range expectedMappings {
		mapping, exists := mappings[expectedType]
		assert.True(t, exists, "Type mapping for %s should exist", expectedType)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
		assert.NotEmpty(t, mapping.Converter)
	}
}

// TestSchemaBridge_Permissions tests permission requirements
func TestSchemaBridge_Permissions(t *testing.T) {
	bridge := NewSchemaBridge()
	permissions := bridge.RequiredPermissions()

	assert.Len(t, permissions, 2)

	// Check for file system permission
	hasFileSystem := false
	hasMemory := false

	for _, perm := range permissions {
		switch perm.Type {
		case engine.PermissionFileSystem:
			hasFileSystem = true
			assert.Equal(t, "schema.files", perm.Resource)
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		case engine.PermissionMemory:
			hasMemory = true
			assert.Equal(t, "schema.cache", perm.Resource)
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasFileSystem, "Should require file system permission")
	assert.True(t, hasMemory, "Should require memory permission")
}

// TestSchemaBridge_HelperFunctions tests utility functions
func TestSchemaBridge_HelperFunctions(t *testing.T) {
	t.Run("schemaToScript", func(t *testing.T) {
		// Test with nil schema
		result := schemaToScript(nil)
		assert.Empty(t, result)

		// Test with valid schema would require constructing a go-llms schema
		// This is tested indirectly through other methods
	})

	t.Run("validationErrorsToScript", func(t *testing.T) {
		errors := []string{"error1", "error2"}
		result := validationErrorsToScript(errors)

		assert.Len(t, result, 2)
		for i, err := range result {
			errorMap := err.(map[string]interface{})
			assert.Equal(t, fmt.Sprintf("error%d", i+1), errorMap["message"])
			assert.Equal(t, "validation_error", errorMap["type"])
		}
	})
}

// TestSchemaBridge_ConcurrentAccess tests thread safety
func TestSchemaBridge_ConcurrentAccess(t *testing.T) {
	bridge, ctx := setupTestBridge(t)
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	// Test concurrent schema operations
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Save schema
			schemaName := fmt.Sprintf("concurrent-schema-%d", id)
			schemaData := createTestSchema()
			args := []engine.ScriptValue{
				engine.NewStringValue(schemaName),
				engine.NewObjectValue(engine.ConvertMapToScriptValue(schemaData)),
			}

			_, err := bridge.ExecuteMethod(ctx, "saveSchema", args)
			assert.NoError(t, err)

			// Get schema
			getArgs := []engine.ScriptValue{engine.NewStringValue(schemaName)}
			_, err = bridge.ExecuteMethod(ctx, "getSchema", getArgs)
			assert.NoError(t, err)

			// Delete schema
			_, err = bridge.ExecuteMethod(ctx, "deleteSchema", getArgs)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
