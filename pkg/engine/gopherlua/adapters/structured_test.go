// ABOUTME: Tests for Structured bridge adapter that exposes go-llms schema validation and generation functionality to Lua scripts
// ABOUTME: Validates schema creation, validation, generation, repository operations, tag-based generation, and import/export functionality

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestStructuredAdapter_Creation(t *testing.T) {
	t.Run("create_structured_adapter", func(t *testing.T) {
		// Create structured bridge mock
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Schema Bridge",
				Version:     "2.0.0",
				Description: "Provides access to go-llms schema validation, generation, versioning, and migration system",
			}).
			WithMethod("createSchema", engine.MethodInfo{
				Name:        "createSchema",
				Description: "Create a new schema object",
				ReturnType:  "object",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock schema creation
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type":       engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{}),
					}),
					"created":   engine.NewBoolValue(true),
					"timestamp": engine.NewStringValue("2024-01-01T00:00:00Z"),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("validateJSON", engine.MethodInfo{
				Name:        "validateJSON",
				Description: "Validate JSON data against a schema",
				ReturnType:  "object",
			}, nil).
			WithMethod("generateSchemaFromType", engine.MethodInfo{
				Name:        "generateSchemaFromType",
				Description: "Generate schema from a type definition",
				ReturnType:  "object",
			}, nil)

		// Create adapter
		adapter := NewStructuredAdapter(structuredBridge)
		require.NotNil(t, adapter)

		// Should have flattened structured-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "createSchema")
		assert.Contains(t, methods, "createProperty")
		assert.Contains(t, methods, "validationValidateJSON")
		assert.Contains(t, methods, "validationValidateStruct")
		assert.Contains(t, methods, "generationFromType")
		assert.Contains(t, methods, "generationFromJSONSchema")
		assert.Contains(t, methods, "repositorySave")
		assert.Contains(t, methods, "repositoryGet")
		assert.Contains(t, methods, "repositoryDelete")
	})

	t.Run("structured_module_structure", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "Schema Bridge",
			}).
			WithMethod("createSchema", engine.MethodInfo{
				Name: "createSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("mock-schema"), nil
			}).
			WithMethod("validateJSON", engine.MethodInfo{
				Name: "validateJSON",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"valid": engine.NewBoolValue(true),
				}), nil
			}).
			WithMethod("generateFromTags", engine.MethodInfo{
				Name: "generateFromTags",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("mock-schema"), nil
			}).
			WithMethod("exportToJSONSchema", engine.MethodInfo{
				Name: "exportToJSONSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("{}"), nil
			}).
			WithMethod("registerCustomValidator", engine.MethodInfo{
				Name: "registerCustomValidator",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("saveSchema", engine.MethodInfo{
				Name: "saveSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check standard methods exist
		assert.NotEqual(t, lua.LNil, module.RawGetString("createSchema"))
		assert.NotEqual(t, lua.LNil, module.RawGetString("validateJSON"))

		// Check flattened methods exist (namespaces have been flattened)
		validationMethod := module.RawGetString("validationValidateJSON")
		assert.NotEqual(t, lua.LNil, validationMethod, "validationValidateJSON flattened method should exist")

		generationMethod := module.RawGetString("generationFromType")
		assert.NotEqual(t, lua.LNil, generationMethod, "generationFromType flattened method should exist")

		repositoryMethod := module.RawGetString("repositorySave")
		assert.NotEqual(t, lua.LNil, repositoryMethod, "repositorySave flattened method should exist")

		importExportMethod := module.RawGetString("importExportToJSONSchema")
		assert.NotEqual(t, lua.LNil, importExportMethod, "importExportToJSONSchema flattened method should exist")

		customMethod := module.RawGetString("customRegisterValidator")
		assert.NotEqual(t, lua.LNil, customMethod, "customRegisterValidator flattened method should exist")
	})
}

func TestStructuredAdapter_SchemaCreation(t *testing.T) {
	t.Run("create_simple_schema", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("createSchema", engine.MethodInfo{
				Name: "createSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock schema object
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type": engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("string"),
							}),
							"age": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("number"),
							}),
						}),
					}),
					"created":   engine.NewBoolValue(true),
					"timestamp": engine.NewStringValue("2024-01-01T00:00:00Z"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Create schema from Lua
		err = L.DoString(`
			local structured = require("structured")
			local result = structured.createSchema({
				type = "object",
				properties = {
					name = { type = "string" },
					age = { type = "number" }
				}
			})
			assert(result ~= nil)
			assert(result.schema ~= nil)
			assert(result.created == true)
			assert(result.schema.type == "object")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_property_definition", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("createProperty", engine.MethodInfo{
				Name: "createProperty",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if args[0].Type() != engine.TypeString {
					return nil, fmt.Errorf("expected string for property type")
				}

				propertyType := args[0].(engine.StringValue).Value()

				// Mock property creation with constraints
				property := map[string]engine.ScriptValue{
					"type":        engine.NewStringValue(propertyType),
					"constraints": engine.NewObjectValue(map[string]engine.ScriptValue{}),
					"created":     engine.NewStringValue("2024-01-01T00:00:00Z"),
				}

				if len(args) > 1 && args[1].Type() == engine.TypeObject {
					constraints := args[1].(engine.ObjectValue).Fields()
					property["constraints"] = engine.NewObjectValue(constraints)
				}

				return engine.NewObjectValue(property), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Create property with constraints
		err = L.DoString(`
			local structured = require("structured")
			local property = structured.createProperty("string", {
				minLength = 5,
				maxLength = 50,
				pattern = "^[a-zA-Z]+$"
			})
			assert(property.type == "string")
			assert(property.constraints.minLength == 5)
			assert(property.constraints.maxLength == 50)
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_SchemaValidation(t *testing.T) {
	t.Run("validate_json_data", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("validateJSON", engine.MethodInfo{
				Name: "validateJSON",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock validation result
				result := map[string]engine.ScriptValue{
					"valid":  engine.NewBoolValue(true),
					"errors": engine.NewArrayValue([]engine.ScriptValue{}),
					"schema": args[0], // Echo back the schema
					"data":   args[1], // Echo back the data
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test JSON validation
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = {
				type = "object",
				properties = {
					name = { type = "string" },
					age = { type = "number" }
				},
				required = {"name"}
			}
			
			local data = {
				name = "John Doe",
				age = 30
			}
			
			local result, err = structured.validationValidateJSON(schema, data)
			assert(err == nil, "validation should not error: " .. tostring(err))
			assert(result.valid == true, "data should be valid")
			assert(#result.errors == 0, "should have no errors")
		`)
		assert.NoError(t, err)
	})

	t.Run("validate_with_errors", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("validateJSON", engine.MethodInfo{
				Name: "validateJSON",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock validation with errors
				errors := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"message": engine.NewStringValue("required property 'name' is missing"),
						"type":    engine.NewStringValue("validation_error"),
					}),
				}

				result := map[string]engine.ScriptValue{
					"valid":  engine.NewBoolValue(false),
					"errors": engine.NewArrayValue(errors),
					"schema": args[0],
					"data":   args[1],
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test validation with errors
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = {
				type = "object",
				properties = {
					name = { type = "string" }
				},
				required = {"name"}
			}
			
			local data = {
				age = 30  -- missing required 'name'
			}
			
			local result, err = structured.validationValidateJSON(schema, data)
			assert(err == nil, "validation should not error: " .. tostring(err))
			assert(result.valid == false, "data should be invalid")
			assert(#result.errors == 1, "should have one error")
			assert(string.find(result.errors[1].message, "required"), "error should mention required field")
		`)
		assert.NoError(t, err)
	})

	t.Run("validate_struct_data", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("validateStruct", engine.MethodInfo{
				Name: "validateStruct",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock struct validation (similar to JSON validation)
				result := map[string]engine.ScriptValue{
					"valid":  engine.NewBoolValue(true),
					"errors": engine.NewArrayValue([]engine.ScriptValue{}),
					"schema": args[0],
					"data":   args[1],
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test struct validation
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = {
				type = "object",
				properties = {
					id = { type = "string" },
					count = { type = "number" }
				}
			}
			
			local structData = {
				id = "user-123",
				count = 42
			}
			
			local result, err = structured.validationValidateStruct(schema, structData)
			assert(err == nil, "struct validation should not error: " .. tostring(err))
			assert(result.valid == true, "struct should be valid")
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_SchemaGeneration(t *testing.T) {
	t.Run("generate_from_type", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("generateSchemaFromType", engine.MethodInfo{
				Name: "generateSchemaFromType",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock type-based schema generation
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type": engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("string"),
							}),
							"value": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("number"),
							}),
						}),
					}),
					"generated": engine.NewBoolValue(true),
					"source":    engine.NewStringValue("type"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test schema generation from type
		err = L.DoString(`
			local structured = require("structured")
			
			local typeInfo = {
				properties = {
					name = { type = "string" },
					value = { type = "number" }
				}
			}
			
			local result, err = structured.generationFromType(typeInfo)
			assert(err == nil, "generation should not error: " .. tostring(err))
			assert(result.generated == true, "schema should be generated")
			assert(result.source == "type", "source should be 'type'")
			assert(result.schema.type == "object", "schema should be object type")
		`)
		assert.NoError(t, err)
	})

	t.Run("generate_from_tags", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("generateFromTags", engine.MethodInfo{
				Name: "generateFromTags",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock tag-based schema generation
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type": engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{
							"email": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type":   engine.NewStringValue("string"),
								"format": engine.NewStringValue("email"),
							}),
						}),
					}),
					"generated": engine.NewBoolValue(true),
					"source":    engine.NewStringValue("tags"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test schema generation from tags
		err = L.DoString(`
			local structured = require("structured")
			
			local structData = {
				fields = {
					{
						name = "email",
						tags = {
							json = "email",
							validate = "email"
						}
					}
				}
			}
			
			local result, err = structured.generationFromTags(structData)
			assert(err == nil, "tag generation should not error: " .. tostring(err))
			assert(result.generated == true, "schema should be generated")
			assert(result.source == "tags", "source should be 'tags'")
		`)
		assert.NoError(t, err)
	})

	t.Run("convert_json_schema", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("convertJSONSchema", engine.MethodInfo{
				Name: "convertJSONSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock JSON schema conversion
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type":       engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{}),
					}),
					"converted": engine.NewBoolValue(true),
					"source":    engine.NewStringValue("json"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test JSON schema conversion
		err = L.DoString(`
			local structured = require("structured")
			
			local jsonSchema = [[{
				"type": "object",
				"properties": {
					"name": {"type": "string"}
				}
			}]]
			
			local result, err = structured.generationFromJSONSchema(jsonSchema)
			assert(err == nil, "conversion should not error: " .. tostring(err))
			assert(result.converted == true, "schema should be converted")
			assert(result.source == "json", "source should be 'json'")
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_SchemaRepository(t *testing.T) {
	t.Run("save_and_get_schema", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("saveSchema", engine.MethodInfo{
				Name: "saveSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful save
				return engine.NewNilValue(), nil
			}).
			WithMethod("getSchema", engine.MethodInfo{
				Name: "getSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if args[0].Type() != engine.TypeString {
					return nil, fmt.Errorf("expected string for schema name")
				}

				name := args[0].(engine.StringValue).Value()

				// Mock retrieved schema
				result := map[string]engine.ScriptValue{
					"name": engine.NewStringValue(name),
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type":       engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{}),
					}),
					"found": engine.NewBoolValue(true),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("deleteSchema", engine.MethodInfo{
				Name: "deleteSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test schema repository operations
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = {
				type = "object",
				properties = {
					name = { type = "string" }
				}
			}
			
			-- Save schema
			local saveResult, saveErr = structured.repositorySave("user-schema", schema)
			assert(saveErr == nil, "save should not error: " .. tostring(saveErr))
			
			-- Get schema
			local getResult, getErr = structured.repositoryGet("user-schema")
			assert(getErr == nil, "get should not error: " .. tostring(getErr))
			assert(getResult.found == true, "schema should be found")
			assert(getResult.name == "user-schema", "schema name should match")
			
			-- Delete schema
			local deleteResult, deleteErr = structured.repositoryDelete("user-schema")
			assert(deleteErr == nil, "delete should not error: " .. tostring(deleteErr))
		`)
		assert.NoError(t, err)
	})

	t.Run("initialize_file_repository", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("initializeFileRepository", engine.MethodInfo{
				Name: "initializeFileRepository",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful file repository initialization
				return engine.NewNilValue(), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test file repository initialization
		err = L.DoString(`
			local structured = require("structured")
			
			local result, err = structured.repositoryInitializeFile("/tmp/schemas")
			assert(err == nil, "file repository init should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_ImportExport(t *testing.T) {
	t.Run("export_to_json_schema", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("exportToJSONSchema", engine.MethodInfo{
				Name: "exportToJSONSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock JSON schema export
				jsonSchema := `{
					"type": "object",
					"properties": {
						"name": {"type": "string"}
					}
				}`
				return engine.NewStringValue(jsonSchema), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test JSON schema export
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = {
				type = "object",
				properties = {
					name = { type = "string" }
				}
			}
			
			local jsonSchema, err = structured.importExportToJSONSchema(schema)
			assert(err == nil, "export should not error: " .. tostring(err))
			assert(type(jsonSchema) == "string", "result should be string")
			assert(string.find(jsonSchema, "object"), "should contain 'object'")
		`)
		assert.NoError(t, err)
	})

	t.Run("export_to_openapi", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("exportToOpenAPI", engine.MethodInfo{
				Name: "exportToOpenAPI",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock OpenAPI export
				openApiSchema := `{
					"type": "object",
					"properties": {
						"name": {"type": "string"}
					}
				}`
				return engine.NewStringValue(openApiSchema), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test OpenAPI export
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = {
				type = "object",
				properties = {
					name = { type = "string" }
				}
			}
			
			local openApiSchema, err = structured.importExportToOpenAPI(schema)
			assert(err == nil, "OpenAPI export should not error: " .. tostring(err))
			assert(type(openApiSchema) == "string", "result should be string")
		`)
		assert.NoError(t, err)
	})

	t.Run("import_from_file", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("importFromFile", engine.MethodInfo{
				Name: "importFromFile",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock file import
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type":       engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{}),
					}),
					"imported": engine.NewBoolValue(true),
					"source":   engine.NewStringValue("file"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test file import
		err = L.DoString(`
			local structured = require("structured")
			
			local result, err = structured.importExportFromFile("/path/to/schema.json", "json")
			assert(err == nil, "import should not error: " .. tostring(err))
			assert(result.imported == true, "schema should be imported")
			assert(result.source == "file", "source should be 'file'")
		`)
		assert.NoError(t, err)
	})

	t.Run("merge_schemas", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("mergeSchemas", engine.MethodInfo{
				Name: "mergeSchemas",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock schema merging
				result := map[string]engine.ScriptValue{
					"schema": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type": engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{
							"name": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("string"),
							}),
							"age": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("number"),
							}),
						}),
					}),
					"merged": engine.NewBoolValue(true),
					"count":  engine.NewNumberValue(2),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test schema merging
		err = L.DoString(`
			local structured = require("structured")
			
			local schema1 = {
				type = "object",
				properties = {
					name = { type = "string" }
				}
			}
			
			local schema2 = {
				type = "object", 
				properties = {
					age = { type = "number" }
				}
			}
			
			local result, err = structured.importExportMerge({schema1, schema2}, "deep")
			assert(err == nil, "merge should not error: " .. tostring(err))
			assert(result.merged == true, "schemas should be merged")
			assert(result.count == 2, "should have merged 2 schemas")
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_CustomValidation(t *testing.T) {
	t.Run("register_custom_validator", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("registerCustomValidator", engine.MethodInfo{
				Name: "registerCustomValidator",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful registration
				return engine.NewNilValue(), nil
			}).
			WithMethod("validateWithCustom", engine.MethodInfo{
				Name: "validateWithCustom",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock custom validation
				result := map[string]engine.ScriptValue{
					"valid":         engine.NewBoolValue(true),
					"customResult":  engine.NewBoolValue(true),
					"validatorName": engine.NewStringValue("email"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test custom validator registration and usage
		err = L.DoString(`
			local structured = require("structured")
			
			-- Define custom validator
			local function emailValidator(value)
				return string.find(value, "@") ~= nil
			end
			
			-- Register validator
			local regResult, regErr = structured.customRegisterValidator("email", {
				validator = emailValidator,
				description = "Validates email format"
			})
			assert(regErr == nil, "registration should not error: " .. tostring(regErr))
			
			-- Use custom validator
			local data = { email = "test@example.com" }
			local result, err = structured.customValidate(data, "email")
			assert(err == nil, "custom validation should not error: " .. tostring(err))
			assert(result.valid == true, "validation should pass")
			assert(result.customResult == true, "custom validation should pass")
		`)
		assert.NoError(t, err)
	})

	t.Run("list_custom_validators", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("listCustomValidators", engine.MethodInfo{
				Name: "listCustomValidators",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock list of validators
				validators := []engine.ScriptValue{
					engine.NewStringValue("email"),
					engine.NewStringValue("phone"),
					engine.NewStringValue("url"),
				}
				return engine.NewArrayValue(validators), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test listing custom validators (arrays return as multiple values)
		err = L.DoString(`
			local structured = require("structured")
			
			local validator1, validator2, validator3 = structured.customListValidators()
			assert(validator1 == "email", "first validator should be email")
			assert(validator2 == "phone", "second validator should be phone")
			assert(validator3 == "url", "third validator should be url")
		`)
		assert.NoError(t, err)
	})

	t.Run("async_validation", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("validateAsync", engine.MethodInfo{
				Name: "validateAsync",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock async validation
				result := map[string]engine.ScriptValue{
					"validationId": engine.NewStringValue("async-123"),
					"status":       engine.NewStringValue("queued"),
					"async":        engine.NewBoolValue(true),
				}
				return engine.NewObjectValue(result), nil
			}).
			WithMethod("getValidationMetrics", engine.MethodInfo{
				Name: "getValidationMetrics",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock validation metrics
				metrics := map[string]engine.ScriptValue{
					"totalValidations":      engine.NewNumberValue(100),
					"successfulValidations": engine.NewNumberValue(95),
					"failedValidations":     engine.NewNumberValue(5),
					"averageLatency":        engine.NewNumberValue(50),
					"asyncValidations":      engine.NewNumberValue(10),
				}
				return engine.NewObjectValue(metrics), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test async validation and metrics
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = { type = "object" }
			local data = { name = "test" }
			
			-- Start async validation
			local asyncResult, asyncErr = structured.customValidateAsync(schema, data)
			assert(asyncErr == nil, "async validation should not error: " .. tostring(asyncErr))
			assert(asyncResult.async == true, "should be async validation")
			assert(asyncResult.status == "queued", "should be queued")
			
			-- Get validation metrics
			local metrics, metricsErr = structured.customGetMetrics()
			assert(metricsErr == nil, "metrics should not error: " .. tostring(metricsErr))
			assert(metrics.totalValidations == 100, "should have 100 total validations")
			assert(metrics.asyncValidations == 10, "should have 10 async validations")
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("validateJSON", engine.MethodInfo{
				Name: "validateJSON",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("schema validation service unavailable")
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test error handling
		err = L.DoString(`
			local structured = require("structured")
			
			local schema = { type = "object" }
			local data = { name = "test" }
			
			local result, err = structured.validateJSON(schema, data)
			assert(result == nil, "result should be nil on error")
			assert(string.find(err, "schema validation service unavailable"), "should contain error message")
		`)
		assert.NoError(t, err)
	})

	t.Run("handle_invalid_schema_format", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("convertJSONSchema", engine.MethodInfo{
				Name: "convertJSONSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("invalid JSON schema format")
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test invalid schema handling
		err = L.DoString(`
			local structured = require("structured")
			
			local invalidSchema = "{ invalid json }"
			
			local result, err = structured.generationFromJSONSchema(invalidSchema)
			assert(result == nil, "result should be nil on error")
			assert(string.find(err, "invalid JSON schema format"), "should contain format error")
		`)
		assert.NoError(t, err)
	})
}

func TestStructuredAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("schema_constants", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("createSchema", engine.MethodInfo{
				Name: "createSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("mock-schema"), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test schema constants
		err = L.DoString(`
			local structured = require("structured")
			
			-- Check that constants are available
			assert(structured.TYPES ~= nil, "TYPES constants should exist")
			assert(structured.TYPES.STRING == "string", "string type should be available")
			assert(structured.TYPES.NUMBER == "number", "number type should be available")
			assert(structured.TYPES.OBJECT == "object", "object type should be available")
			assert(structured.TYPES.ARRAY == "array", "array type should be available")
			
			assert(structured.FORMATS ~= nil, "FORMATS constants should exist")
			assert(structured.FORMATS.EMAIL == "email", "email format should be available")
			assert(structured.FORMATS.DATE == "date", "date format should be available")
			assert(structured.FORMATS.URI == "uri", "uri format should be available")
		`)
		assert.NoError(t, err)
	})

	t.Run("utility_methods", func(t *testing.T) {
		structuredBridge := testutils.NewMockBridge("structured").
			WithInitialized(true).
			WithMethod("generateDiff", engine.MethodInfo{
				Name: "generateDiff",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock schema diff generation
				diff := map[string]engine.ScriptValue{
					"added": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewStringValue("newProperty"),
					}),
					"removed": engine.NewArrayValue([]engine.ScriptValue{
						engine.NewStringValue("oldProperty"),
					}),
					"modified":    engine.NewArrayValue([]engine.ScriptValue{}),
					"changeCount": engine.NewNumberValue(2),
				}
				return engine.NewObjectValue(diff), nil
			})

		adapter := NewStructuredAdapter(structuredBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "structured")
		require.NoError(t, err)

		err = ms.LoadModule(L, "structured")
		require.NoError(t, err)

		// Test utility methods
		err = L.DoString(`
			local structured = require("structured")
			
			local oldSchema = {
				type = "object",
				properties = {
					oldProperty = { type = "string" }
				}
			}
			
			local newSchema = {
				type = "object",
				properties = {
					newProperty = { type = "string" }
				}
			}
			
			local diff, err = structured.utilsGenerateDiff(oldSchema, newSchema)
			assert(err == nil, "diff generation should not error: " .. tostring(err))
			assert(diff.changeCount == 2, "should have 2 changes")
			assert(#diff.added == 1, "should have 1 added property")
			assert(#diff.removed == 1, "should have 1 removed property")
		`)
		assert.NoError(t, err)
	})
}
