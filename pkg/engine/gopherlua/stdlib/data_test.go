// ABOUTME: Comprehensive test suite for Structured Data Library in Lua standard library
// ABOUTME: Tests JSON processing, schema validation, data transformation, and utility functions

package stdlib

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge/util"
	lua "github.com/yuin/gopher-lua"
)

// setupDataLibrary loads the data library and sets up required bridges
func setupDataLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Set up mock utils bridge
	utilBridge := util.NewUtilJSONBridge()
	err := utilBridge.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Failed to initialize util bridge: %v", err)
	}

	// Create mock global utils for testing
	utilTable := L.NewTable()

	// Mock JSON functions
	utilTable.RawSetString("jsonDecode", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		switch text {
		case `{"name":"John","age":30}`:
			result := L.NewTable()
			result.RawSetString("name", lua.LString("John"))
			result.RawSetString("age", lua.LNumber(30))
			L.Push(result)
		case `[1,2,3]`:
			result := L.NewTable()
			result.RawSetInt(1, lua.LNumber(1))
			result.RawSetInt(2, lua.LNumber(2))
			result.RawSetInt(3, lua.LNumber(3))
			L.Push(result)
		default:
			L.Push(lua.LNil)
		}
		return 1
	}))

	utilTable.RawSetString("jsonEncode", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1) // Ignore unused parameter for mock
		// Simple mock - just return a string representation
		L.Push(lua.LString(`{"encoded":"mock"}`))
		return 1
	}))

	utilTable.RawSetString("jsonPrettify", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1) // Ignore unused parameter for mock
		L.Push(lua.LString("{\n  \"pretty\": \"formatted\"\n}"))
		return 1
	}))

	utilTable.RawSetString("jsonValidateJSONSchema", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1) // Ignore unused parameter for mock
		_ = L.CheckTable(2)  // Ignore unused parameter for mock

		result := L.NewTable()
		result.RawSetString("valid", lua.LTrue)
		errors := L.NewTable()
		result.RawSetString("errors", errors)
		L.Push(result)
		return 1
	}))

	utilTable.RawSetString("jsonExtractStructuredData", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1) // Ignore unused parameter for mock
		_ = L.CheckTable(2)  // Ignore unused parameter for mock

		result := L.NewTable()
		result.RawSetString("extracted", lua.LString("mock data"))
		L.Push(result)
		return 1
	}))

	utilTable.RawSetString("jsonToJSON", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckAny(1)   // Ignore unused parameter for mock
		_ = L.CheckTable(2) // Ignore unused parameter for mock

		// Return formatted JSON string
		L.Push(lua.LString(`{"formatted":"json"}`))
		return 1
	}))

	L.SetGlobal("util", utilTable)

	// Load the data library
	dataPath := filepath.Join(".", "data.lua")
	err = L.DoFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to load data library: %v", err)
	}
	data := L.Get(-1)
	L.SetGlobal("data", data)
}

// TestDataLibraryLoading tests that the data library can be loaded
func TestDataLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupDataLibrary(t, L)

	// Check that data table exists and has expected functions
	script := `
		if type(data) ~= "table" then
			error("Data module should be a table")
		end
		
		local required_functions = {
			"parse_json", "to_json", "extract_structured", "convert_format",
			"validate", "infer_schema", "migrate_schema",
			"map", "filter", "reduce",
			"clone", "merge", "get_path", "set_path"
		}
		
		for _, func_name in ipairs(required_functions) do
			if type(data[func_name]) ~= "function" then
				error("Function " .. func_name .. " should be available")
			end
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Data library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}
}

// TestJSONProcessing tests JSON parsing, encoding, and formatting
func TestJSONProcessing(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "parse_json_basic",
			script: `
				local result = data.parse_json('{"name":"John","age":30}')
				return result.name
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "John" {
					t.Errorf("Expected 'John', got %v", result.String())
				}
			},
		},
		{
			name: "to_json_basic",
			script: `
				local obj = {name = "John", age = 30}
				local json_str = data.to_json(obj)
				return type(json_str) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected to_json to return string, got %v", result)
				}
			},
		},
		{
			name: "to_json_pretty_format",
			script: `
				local obj = {name = "John"}
				local json_str = data.to_json(obj, "pretty")
				return type(json_str) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected pretty format to return string, got %v", result)
				}
			},
		},
		{
			name: "extract_structured_data",
			script: `
				local text = "Some text with data"
				local schema = {type = "object", properties = {name = {type = "string"}}}
				local result = data.extract_structured(text, schema)
				return result ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected extract_structured to return data, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestSchemaValidation tests schema validation functions
func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "validate_success",
			script: `
				local obj = {name = "John", age = 30}
				local schema = {
					type = "object",
					properties = {
						name = {type = "string"},
						age = {type = "number"}
					}
				}
				local result = data.validate(obj, schema)
				return result.valid
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected validation to succeed, got %v", result)
				}
			},
		},
		{
			name: "infer_schema_object",
			script: `
				local obj = {name = "John", age = 30, active = true}
				local schema = data.infer_schema(obj)
				return schema.type == "object" and schema.properties ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected object schema inference, got %v", result)
				}
			},
		},
		{
			name: "infer_schema_array",
			script: `
				local arr = {"first", "second", "third"}
				local schema = data.infer_schema(arr)
				return schema.type == "array" and schema.items ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected array schema inference, got %v", result)
				}
			},
		},
		{
			name: "migrate_schema_basic",
			script: `
				local obj = {name = "John", age = 30}
				local old_schema = {type = "object", properties = {name = {type = "string"}}}
				local new_schema = {type = "object", properties = {name = {type = "string"}, age = {type = "number"}}}
				local migrated = data.migrate_schema(obj, old_schema, new_schema)
				return migrated.name == "John" and migrated.age == 30
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected successful schema migration, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestDataTransformation tests map, filter, reduce functions
func TestDataTransformation(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "map_array",
			script: `
				local arr = {1, 2, 3, 4, 5}
				local doubled = data.map(arr, function(x) return x * 2 end)
				return doubled[1] == 2 and doubled[3] == 6 and #doubled == 5
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected array mapping to work, got %v", result)
				}
			},
		},
		{
			name: "map_object",
			script: `
				local obj = {a = 1, b = 2, c = 3}
				local doubled = data.map(obj, function(x) return x * 2 end)
				return doubled.a == 2 and doubled.b == 4 and doubled.c == 6
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected object mapping to work, got %v", result)
				}
			},
		},
		{
			name: "filter_array",
			script: `
				local arr = {1, 2, 3, 4, 5, 6}
				local evens = data.filter(arr, function(x) return x % 2 == 0 end)
				return #evens == 3 and evens[1] == 2 and evens[2] == 4 and evens[3] == 6
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected array filtering to work, got %v", result)
				}
			},
		},
		{
			name: "reduce_array_sum",
			script: `
				local arr = {1, 2, 3, 4, 5}
				local sum = data.reduce(arr, function(acc, x) return acc + x end, 0)
				return sum == 15
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected reduce sum to work, got %v", result)
				}
			},
		},
		{
			name: "reduce_without_initial",
			script: `
				local arr = {1, 2, 3, 4, 5}
				local sum = data.reduce(arr, function(acc, x) return acc + x end)
				return sum == 15
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected reduce without initial to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestUtilityFunctions tests clone, merge, get_path, set_path
func TestUtilityFunctions(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "clone_deep",
			script: `
				local original = {a = 1, b = {c = 2, d = {e = 3}}}
				local cloned = data.clone(original)
				cloned.b.c = 999
				-- Original should be unchanged
				return original.b.c == 2 and cloned.b.c == 999
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected deep clone to work, got %v", result)
				}
			},
		},
		{
			name: "merge_objects",
			script: `
				local obj1 = {a = 1, b = {c = 2}}
				local obj2 = {b = {d = 3}, e = 4}
				local merged = data.merge(obj1, obj2)
				return merged.a == 1 and merged.b.c == 2 and merged.b.d == 3 and merged.e == 4
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected object merge to work, got %v", result)
				}
			},
		},
		{
			name: "get_path_nested",
			script: `
				local obj = {user = {profile = {name = "John", age = 30}}}
				local name = data.get_path(obj, "user.profile.name")
				local age = data.get_path(obj, "user.profile.age")
				return name == "John" and age == 30
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected get_path to work, got %v", result)
				}
			},
		},
		{
			name: "set_path_nested",
			script: `
				local obj = {}
				data.set_path(obj, "user.profile.name", "Jane")
				data.set_path(obj, "user.profile.age", 25)
				return obj.user.profile.name == "Jane" and obj.user.profile.age == 25
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected set_path to work, got %v", result)
				}
			},
		},
		{
			name: "get_path_missing",
			script: `
				local obj = {a = {b = 1}}
				local result = data.get_path(obj, "a.missing.path")
				return result == nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected get_path to return nil for missing path, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestFormatConversion tests convert_format function
func TestFormatConversion(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "json_to_json",
			script: `
				local obj = {name = "John", age = 30}
				local result = data.convert_format(obj, "json", "json")
				return type(result) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected JSON to JSON conversion to work, got %v", result)
				}
			},
		},
		{
			name: "string_to_json",
			script: `
				local json_str = '{"name":"John","age":30}'
				local result = data.convert_format(json_str, "string", "json")
				return result.name == "John"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string to JSON conversion to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestDataErrorHandling tests error cases and validation
func TestDataErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "parse_json_missing_text",
			script: `
				local success, err = pcall(function()
					data.parse_json(nil)
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing text parameter, got %v", result)
				}
			},
		},
		{
			name: "map_invalid_collection",
			script: `
				local success, err = pcall(function()
					data.map("not a table", function(x) return x end)
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid collection, got %v", result)
				}
			},
		},
		{
			name: "map_invalid_mapper",
			script: `
				local success, err = pcall(function()
					data.map({1, 2, 3}, "not a function")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid mapper, got %v", result)
				}
			},
		},
		{
			name: "merge_invalid_objects",
			script: `
				local success, err = pcall(function()
					data.merge("not a table", {})
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid merge objects, got %v", result)
				}
			},
		},
		{
			name: "set_path_invalid_path",
			script: `
				local success, err = pcall(function()
					data.set_path({}, 123, "value")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid path type, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestComplexOperations tests combinations of operations
func TestComplexOperations(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupDataLibrary(t, L)

	script := `
		-- Complex data processing pipeline
		local raw_data = {
			users = {
				{name = "John", age = 30, active = true},
				{name = "Jane", age = 25, active = false},
				{name = "Bob", age = 35, active = true}
			}
		}
		
		-- Clone the data to avoid mutation
		local data_copy = data.clone(raw_data)
		
		-- Filter active users
		local active_users = data.filter(data_copy.users, function(user)
			return user.active
		end)
		
		-- Map to get names only
		local active_names = data.map(active_users, function(user)
			return user.name
		end)
		
		-- Reduce to count
		local count = data.reduce(active_names, function(acc, name)
			return acc + 1
		end, 0)
		
		-- Create result object
		local result = {
			total_active = count,
			names = active_names,
			first_name = active_names[1] -- Direct array access instead of get_path
		}
		
		-- Merge with additional info
		local final_result = data.merge(result, {processed = true})
		
		return final_result.total_active == 2 and 
		       final_result.names[1] == "John" and 
		       final_result.names[2] == "Bob" and
		       final_result.first_name == "John" and
		       final_result.processed == true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Complex operations test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected complex operations to work correctly, got %v", result)
	}
}

// BenchmarkDataOperations benchmarks key data operations
func BenchmarkDataOperations(b *testing.B) {
	benchmarks := []struct {
		name   string
		script string
	}{
		{
			name: "map_operation",
			script: `
				local arr = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
				data.map(arr, function(x) return x * 2 end)
			`,
		},
		{
			name: "filter_operation",
			script: `
				local arr = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
				data.filter(arr, function(x) return x % 2 == 0 end)
			`,
		},
		{
			name: "clone_operation",
			script: `
				local obj = {a = 1, b = {c = 2, d = {e = 3, f = {g = 4}}}}
				data.clone(obj)
			`,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()

			setupDataLibrary(nil, L) // Skip t.Helper in benchmark

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := L.DoString(bm.script)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
				L.Pop(1) // Clean stack
			}
		})
	}
}

// TestDataPackageRequire tests that the module can be required as a package
func TestDataPackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Set up mock utils bridge
	setupDataLibrary(t, L)

	script := `
		-- Test that data is available globally
		if type(data) ~= "table" then
			error("Data module should be available globally")
		end
		
		-- Test basic functionality
		local obj = {name = "test"}
		local cloned = data.clone(obj)
		
		if cloned.name ~= "test" then
			error("Basic functionality should work")
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Package availability test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected package test to pass, got %v", result)
	}
}
