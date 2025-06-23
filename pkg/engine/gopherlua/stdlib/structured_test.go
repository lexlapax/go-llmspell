// ABOUTME: Tests for the Structured Data standard library module that provides schema validation and generation
// ABOUTME: Validates structured.lua functionality including schema operations, validation, and repository management

package stdlib

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// MockStructuredBridge represents a mock structured bridge for testing
type MockStructuredBridge struct {
	schemas    map[string]interface{}
	validators map[string]interface{}
	callLog    []string
}

// NewMockStructuredBridge creates a new mock structured bridge
func NewMockStructuredBridge() *MockStructuredBridge {
	return &MockStructuredBridge{
		schemas:    make(map[string]interface{}),
		validators: make(map[string]interface{}),
		callLog:    []string{},
	}
}

// SetupMockStructuredBridge sets up mock bridges in the Lua state
func setupMockStructuredBridge(L *lua.LState, mockBridge *MockStructuredBridge) {
	// Create mock structured bridge table
	structuredBridge := L.NewTable()

	// Mock createSchema method
	structuredBridge.RawSetString("createSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self (bridge table)
		schemaData := L.CheckTable(2)
		mockBridge.callLog = append(mockBridge.callLog, "createSchema")

		result := L.NewTable()
		result.RawSetString("type", lua.LString("object"))
		if props := L.GetField(schemaData, "properties"); props != lua.LNil {
			result.RawSetString("properties", props)
		}
		if req := L.GetField(schemaData, "required"); req != lua.LNil {
			result.RawSetString("required", req)
		}
		result.RawSetString("id", lua.LString("test-schema"))

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock validateJSON method
	structuredBridge.RawSetString("validateJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // schema
		_ = L.CheckTable(3) // data
		mockBridge.callLog = append(mockBridge.callLog, "validateJSON")

		result := L.NewTable()
		result.RawSetString("valid", lua.LTrue)
		result.RawSetString("errors", L.NewTable())

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock saveSchema method
	structuredBridge.RawSetString("saveSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)
		schema := L.CheckTable(3)
		mockBridge.callLog = append(mockBridge.callLog, "saveSchema:"+name)
		mockBridge.schemas[name] = schema

		L.Push(lua.LNil) // no return value
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock getSchema method
	structuredBridge.RawSetString("getSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)
		mockBridge.callLog = append(mockBridge.callLog, "getSchema:"+name)

		if name == "test-schema" {
			result := L.NewTable()
			result.RawSetString("type", lua.LString("object"))
			
			props := L.NewTable()
			nameProp := L.NewTable()
			nameProp.RawSetString("type", lua.LString("string"))
			props.RawSetString("name", nameProp)
			result.RawSetString("properties", props)
			
			req := L.NewTable()
			req.RawSetInt(1, lua.LString("name"))
			result.RawSetString("required", req)

			L.Push(result)
			L.Push(lua.LNil) // no error
			return 2
		}

		L.Push(lua.LNil)
		L.Push(lua.LString("schema not found"))
		return 2
	}))

	// Create bridges global table
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)
}

func TestStructuredModule_BasicOperations(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Set up mock bridge
	mockBridge := NewMockStructuredBridge()
	setupMockStructuredBridge(L, mockBridge)

	// Load module
	LoadModule(t, L, "structured")

	// Test 1: Create schema
	err := L.DoString(`
		local structured = require("structured")
		
		-- Create a simple schema
		local schema = structured.create_schema({
			type = "object",
			properties = {
				name = { type = "string" },
				age = { type = "number" }
			},
			required = { "name" }
		})
		
		assert(schema.type == "object", "schema should be object type")
		assert(schema.id == "test-schema", "schema should have ID")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}

	// Test 2: Validate JSON
	err = L.DoString(`
		local structured = require("structured")
		
		local schema = {
			type = "object",
			properties = {
				name = { type = "string" }
			}
		}
		
		local result = structured.validate_json(schema, { name = "John" })
		assert(result.valid == true, "validation should succeed")
		assert(#result.errors == 0, "should have no errors")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}

	// Test 3: Save and retrieve schema
	err = L.DoString(`
		local structured = require("structured")
		
		local schema = structured.create_schema({
			type = "object",
			properties = {
				name = { type = "string" }
			},
			required = { "name" }
		})
		
		-- Save schema
		structured.save_schema("test-schema", schema)
		
		-- Retrieve schema
		local retrieved = structured.get_schema("test-schema")
		assert(retrieved.type == "object", "retrieved schema should be object")
		assert(retrieved.properties.name.type == "string", "should have name property")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_SchemaBuilder(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create mock structured bridge
	structuredBridge := L.NewTable()

	// Mock createProperty method
	structuredBridge.RawSetString("createProperty", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		propertyType := L.CheckString(2)
		constraints := L.OptTable(3, L.NewTable())

		result := L.NewTable()
		result.RawSetString("type", lua.LString(propertyType))
		
		// Add constraints
		constraints.ForEach(func(k, v lua.LValue) {
			if key, ok := k.(lua.LString); ok {
				result.RawSet(key, v)
			}
		})

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock createSchema method
	structuredBridge.RawSetString("createSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		schemaData := L.CheckTable(2)
		
		L.Push(schemaData) // Return schema as-is
		L.Push(lua.LNil)   // no error
		return 2
	}))

	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test schema builder
	err := L.DoString(`
		local structured = require("structured")
		
		-- Use builder to create schema
		local schema = structured.builder()
			:add_string("name", true, 1, 100)
			:add_number("age", true, 0, 150)
			:add_boolean("active", false)
			:add_array("tags", false, { type = "string" }, 0, 10)
			:build()
		
		assert(schema.type == "object", "schema should be object")
		assert(schema.properties.name.type == "string", "name should be string")
		assert(schema.properties.name.minLength == 1, "name should have minLength")
		assert(schema.properties.name.maxLength == 100, "name should have maxLength")
		assert(schema.properties.age.type == "number", "age should be number")
		assert(schema.properties.age.minimum == 0, "age should have minimum")
		assert(schema.properties.age.maximum == 150, "age should have maximum")
		assert(schema.properties.active.type == "boolean", "active should be boolean")
		assert(schema.properties.tags.type == "array", "tags should be array")
		assert(#schema.required == 2, "should have 2 required fields")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_Conversion(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create mock structured bridge
	structuredBridge := L.NewTable()

	// Mock exportToJSONSchema method
	structuredBridge.RawSetString("exportToJSONSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // schema

		L.Push(lua.LString(`{"type":"object","properties":{"name":{"type":"string"}}}`)) 
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock exportToOpenAPI method
	structuredBridge.RawSetString("exportToOpenAPI", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // schema

		L.Push(lua.LString(`{"type":"object","properties":{"name":{"type":"string"}},"x-api-version":"3.0"}`))
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock convertJSONSchema method
	structuredBridge.RawSetString("convertJSONSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckString(2) // json schema string

		result := L.NewTable()
		result.RawSetString("type", lua.LString("object"))
		
		props := L.NewTable()
		nameProp := L.NewTable()
		nameProp.RawSetString("type", lua.LString("string"))
		props.RawSetString("name", nameProp)
		result.RawSetString("properties", props)

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test conversion functions
	err := L.DoString(`
		local structured = require("structured")
		
		local schema = {
			type = "object",
			properties = {
				name = { type = "string" }
			}
		}
		
		-- Export to JSON Schema
		local json_schema = structured.to_json_schema(schema)
		assert(type(json_schema) == "string", "should return JSON string")
		assert(string.find(json_schema, "properties"), "should contain properties")
		
		-- Export to OpenAPI
		local openapi = structured.to_openapi(schema)
		assert(type(openapi) == "string", "should return OpenAPI string")
		assert(openapi:find("api.version") ~= nil, "should contain OpenAPI extensions")
		
		-- Import from JSON Schema
		local imported = structured.from_json_schema('{"type":"object"}')
		assert(imported.type == "object", "should import schema correctly")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_BatchValidation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	validationCount := 0
	structuredBridge := L.NewTable()

	// Mock validateJSON method
	structuredBridge.RawSetString("validateJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // schema
		data := L.CheckTable(3) // data
		validationCount++

		result := L.NewTable()
		errors := L.NewTable()
		
		// Check if data has required "name" field
		if nameField := L.GetField(data, "name"); nameField != lua.LNil {
			result.RawSetString("valid", lua.LTrue)
		} else {
			result.RawSetString("valid", lua.LFalse)
			errors.RawSetInt(1, lua.LString("missing required field: name"))
		}
		result.RawSetString("errors", errors)

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test batch validation
	err := L.DoString(`
		local structured = require("structured")
		
		local schema = {
			type = "object",
			properties = {
				name = { type = "string" }
			},
			required = { "name" }
		}
		
		local data_array = {
			{ name = "John" },
			{ age = 30 },  -- missing name
			{ name = "Jane" }
		}
		
		local results = structured.validate_batch(schema, data_array)
		
		assert(#results == 3, "should have 3 results")
		assert(results[1].valid == true, "first should be valid")
		assert(results[2].valid == false, "second should be invalid")
		assert(results[3].valid == true, "third should be valid")
		assert(results[2].error ~= nil, "invalid result should have error")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
	if validationCount != 3 {
		t.Errorf("Expected 3 validations, got %d", validationCount)
	}
}

func TestStructuredModule_AsyncValidation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create mock structured bridge
	structuredBridge := L.NewTable()

	// Mock validateAsync method
	structuredBridge.RawSetString("validateAsync", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // schema
		_ = L.CheckTable(3) // data
		_ = L.Get(4)        // callback (optional)

		result := L.NewTable()
		result.RawSetString("id", lua.LString("async-validation-123"))
		result.RawSetString("status", lua.LString("pending"))

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Register bridges
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load promise module first
	LoadModule(t, L, "promise")
	
	// Load structured module
	LoadModule(t, L, "structured")

	// Test async validation
	err := L.DoString(`
		local structured = require("structured")
		
		local schema = {
			type = "object",
			properties = {
				name = { type = "string" }
			}
		}
		
		local data = { name = "Test" }
		
		-- Async validation returns a promise
		local promise = structured.validate_async(schema, data)
		
		-- Check promise is created
		assert(promise ~= nil, "should return a promise")
		assert(type(promise.andThen) == "function", "should have promise methods")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_ErrorHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Register empty bridges table
	L.SetGlobal("bridges", L.NewTable())

	// Load module
	LoadModule(t, L, "structured")

	// Test missing bridge error
	err := L.DoString(`
		local structured = require("structured")
		
		local ok, err = pcall(function()
			structured.create_schema({})
		end)
		
		assert(not ok, "should fail without bridge")
		assert(string.find(err, "Structured bridge not available"), "should have correct error message")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}

	// Test required parameter validation
	err = L.DoString(`
		local structured = require("structured")
		
		-- Test missing schema parameter
		local ok, err = pcall(function()
			structured.validate_json(nil, {})
		end)
		
		assert(not ok, "should fail with nil schema")
		assert(string.find(err, "schema is required"), "should have correct error message")
		
		-- Test missing data parameter
		local ok2, err2 = pcall(function()
			structured.validate_json({}, nil)
		end)
		
		assert(not ok2, "should fail with nil data")
		assert(string.find(err2, "data is required"), "should have correct error message")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_CustomValidators(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	validators := make(map[string]lua.LValue)
	structuredBridge := L.NewTable()

	// Mock registerCustomValidator method
	structuredBridge.RawSetString("registerCustomValidator", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)
		validator := L.CheckTable(3)
		validators[name] = validator

		L.Push(lua.LNil) // no return value
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock listCustomValidators method
	structuredBridge.RawSetString("listCustomValidators", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self

		names := L.NewTable()
		i := 1
		for name := range validators {
			names.RawSetInt(i, lua.LString(name))
			i++
		}

		L.Push(names)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock validateWithCustom method
	structuredBridge.RawSetString("validateWithCustom", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // data
		validatorName := L.CheckString(3)

		if _, ok := validators[validatorName]; ok {
			result := L.NewTable()
			result.RawSetString("valid", lua.LTrue)
			result.RawSetString("errors", L.NewTable())
			L.Push(result)
			L.Push(lua.LNil) // no error
		} else {
			L.Push(lua.LNil)
			L.Push(lua.LString("validator not found"))
		}
		return 2
	}))

	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test custom validators
	err := L.DoString(`
		local structured = require("structured")
		
		-- Register custom validator
		structured.register_validator("email", {
			pattern = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			message = "Invalid email format"
		})
		
		-- List validators
		local validators = structured.list_validators()
		assert(#validators == 1, "should have 1 validator")
		assert(validators[1] == "email", "should have email validator")
		
		-- Use custom validator
		local result = structured.validate_custom({email = "test@example.com"}, "email")
		assert(result.valid == true, "should validate successfully")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_RepositoryOperations(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	repository := make(map[string]lua.LValue)
	structuredBridge := L.NewTable()

	// Mock initializeFileRepository method
	structuredBridge.RawSetString("initializeFileRepository", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)  // self
		_ = L.CheckString(2) // directory

		L.Push(lua.LNil) // no return value
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock exportRepository method
	structuredBridge.RawSetString("exportRepository", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self

		L.Push(lua.LString(`{"schemas":{},"version":"1.0"}`))
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock importRepository method
	structuredBridge.RawSetString("importRepository", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		data := L.CheckTable(2)
		repository["imported"] = data

		result := L.NewTable()
		result.RawSetString("schemasImported", lua.LNumber(3))
		result.RawSetString("errors", L.NewTable())

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test repository operations
	err := L.DoString(`
		local structured = require("structured")
		
		-- Initialize repository
		structured.init_repository("/tmp/schemas")
		
		-- Export repository
		local exported = structured.export_repository()
		assert(type(exported) == "string", "should export as string")
		assert(string.find(exported, "schemas"), "should contain schemas")
		
		-- Import repository
		local result = structured.import_repository({
			schemas = {},
			version = "1.0"
		})
		assert(result.schemasImported == 3, "should import schemas")
		assert(#result.errors == 0, "should have no errors")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_ConvenienceFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	schemas := make(map[string]lua.LValue)
	structuredBridge := L.NewTable()

	// Mock createSchema method
	structuredBridge.RawSetString("createSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		schemaData := L.CheckTable(2)
		
		L.Push(schemaData) // Return schema as-is
		L.Push(lua.LNil)   // no error
		return 2
	}))

	// Mock saveSchema method
	structuredBridge.RawSetString("saveSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)
		schema := L.CheckTable(3)
		schemas[name] = schema

		L.Push(lua.LNil) // no return value
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock getSchema method
	structuredBridge.RawSetString("getSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)

		if schema, ok := schemas[name]; ok {
			L.Push(schema)
			L.Push(lua.LNil) // no error
		} else {
			result := L.NewTable()
			result.RawSetString("type", lua.LString("object"))
			L.Push(result)
			L.Push(lua.LNil) // no error
		}
		return 2
	}))

	// Mock validateJSON method
	structuredBridge.RawSetString("validateJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		_ = L.CheckTable(2) // schema
		_ = L.CheckTable(3) // data

		result := L.NewTable()
		result.RawSetString("valid", lua.LTrue)
		result.RawSetString("errors", L.NewTable())

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test convenience functions
	err := L.DoString(`
		local structured = require("structured")
		
		-- Create and save in one operation
		local schema = structured.create_and_save("user-schema", {
			type = "object",
			properties = {
				name = { type = "string" }
			}
		})
		
		assert(schema.type == "object", "should create schema")
		
		-- Quick validate with schema name
		local result = structured.validate("user-schema", { name = "John" })
		assert(result.valid == true, "should validate with schema name")
		
		-- Quick validate with schema object
		local result2 = structured.validate(schema, { name = "Jane" })
		assert(result2.valid == true, "should validate with schema object")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestStructuredModule_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	L := lua.NewState()
	defer L.Close()

	// Create a more complete mock structured bridge
	schemas := make(map[string]lua.LValue)
	structuredBridge := createFullMockStructuredBridge(L, schemas)
	
	// Register bridge
	bridges := L.NewTable()
	bridges.RawSetString("structured_schema", structuredBridge)
	L.SetGlobal("bridges", bridges)

	// Load module
	LoadModule(t, L, "structured")

	// Test complete workflow
	err := L.DoString(`
		local structured = require("structured")
		
		-- Create a complex schema
		local user_schema = structured.builder()
			:add_string("id", true, nil, nil, "^[a-zA-Z0-9-]+$")
			:add_string("email", true, nil, nil, "^[^@]+@[^@]+$")
			:add_string("name", true, 1, 100)
			:add_number("age", false, 0, 150)
			:add_array("roles", false, { type = "string" })
			:add_object("profile", false, {
				bio = { type = "string", maxLength = 500 },
				avatar = { type = "string", format = "uri" }
			})
			:build()
		
		-- Save it
		structured.save_schema("user", user_schema)
		
		-- Validate some data
		local valid_user = {
			id = "user-123",
			email = "user@example.com",
			name = "John Doe",
			age = 30,
			roles = { "admin", "user" },
			profile = {
				bio = "Software developer",
				avatar = "https://example.com/avatar.jpg"
			}
		}
		
		local result = structured.validate("user", valid_user)
		assert(result.valid == true, "valid user should pass validation")
		
		-- Test invalid data
		local invalid_user = {
			id = "user@123",  -- invalid pattern
			email = "not-an-email",
			name = ""  -- too short
		}
		
		local result2 = structured.validate("user", invalid_user)
		assert(result2.valid == false or #result2.errors > 0, "invalid user should fail validation")
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

// Helper to create a more complete mock structured bridge
func createFullMockStructuredBridge(L *lua.LState, schemas map[string]lua.LValue) *lua.LTable {
	structuredBridge := L.NewTable()

	// Mock createProperty method
	structuredBridge.RawSetString("createProperty", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		propertyType := L.CheckString(2)
		constraints := L.OptTable(3, L.NewTable())

		result := L.NewTable()
		result.RawSetString("type", lua.LString(propertyType))
		
		// Add constraints
		constraints.ForEach(func(k, v lua.LValue) {
			if key, ok := k.(lua.LString); ok {
				result.RawSet(key, v)
			}
		})

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock createSchema method
	structuredBridge.RawSetString("createSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		schemaData := L.CheckTable(2)
		
		L.Push(schemaData) // Return schema as-is
		L.Push(lua.LNil)   // no error
		return 2
	}))

	// Mock saveSchema method
	structuredBridge.RawSetString("saveSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)
		schema := L.CheckTable(3)
		schemas[name] = schema

		L.Push(lua.LNil) // no return value
		L.Push(lua.LNil) // no error
		return 2
	}))

	// Mock getSchema method
	structuredBridge.RawSetString("getSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		name := L.CheckString(2)

		if schema, ok := schemas[name]; ok {
			L.Push(schema)
			L.Push(lua.LNil) // no error
		} else {
			L.Push(lua.LNil)
			L.Push(lua.LString("schema not found"))
		}
		return 2
	}))

	// Mock validateJSON method
	structuredBridge.RawSetString("validateJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // self
		schema := L.CheckTable(2)
		data := L.CheckTable(3)

		result := L.NewTable()
		errors := L.NewTable()
		errorCount := 0

		// Check required fields if any
		if requiredField := L.GetField(schema, "required"); requiredField != lua.LNil {
			if requiredTable, ok := requiredField.(*lua.LTable); ok {
				requiredTable.ForEach(func(k, v lua.LValue) {
					if fieldName, ok := v.(lua.LString); ok {
						fieldValue := L.GetField(data, string(fieldName))
						if fieldValue == lua.LNil {
							errorCount++
							errors.RawSetInt(errorCount, lua.LString("missing required field: "+string(fieldName)))
						} else if strValue, ok := fieldValue.(lua.LString); ok && string(strValue) == "" {
							// Check for empty strings on required fields
							errorCount++
							errors.RawSetInt(errorCount, lua.LString("required field is empty: "+string(fieldName)))
						}
					}
				})
			}
		}

		result.RawSetString("valid", lua.LBool(errorCount == 0))
		result.RawSetString("errors", errors)

		L.Push(result)
		L.Push(lua.LNil) // no error
		return 2
	}))

	return structuredBridge
}