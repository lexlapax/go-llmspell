// ABOUTME: Comprehensive test suite for Core Utilities Library in Lua standard library
// ABOUTME: Tests string manipulation, table utilities, crypto functions, and time handling

package stdlib

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// setupCoreLibrary loads the core library
func setupCoreLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Load the core library
	libPath := filepath.Join(".", "core.lua")
	if err := L.DoFile(libPath); err != nil {
		t.Fatalf("Failed to load core library: %v", err)
	}
	core := L.Get(-1)
	L.SetGlobal("core", core)
}

func TestCoreLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	// Test that the core library was loaded
	script := `
		return type(core) == "table" and
		       type(core.crypto) == "table" and
		       type(core.is_callable) == "function" and
		       type(string.template) == "function" and
		       type(table.keys) == "function" and
		       type(os.now) == "function"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected core library to be properly loaded")
	}
}

func TestStringUtilities(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "string_template",
			script: `
				local result1 = string.template("Hello {{name}}!", {name = "World"})
				local result2 = string.template("{{greeting}} {{name}}!", {greeting = "Hi", name = "Lua"})
				local result3 = string.template("Missing {{var}}", {})
				
				return result1 == "Hello World!" and
				       result2 == "Hi Lua!" and
				       result3 == "Missing {{var}}"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string.template to work correctly")
				}
			},
		},
		{
			name: "string_slugify",
			script: `
				local result1 = string.slugify("Hello World!")
				local result2 = string.slugify("This & That")
				local result3 = string.slugify("  Multiple   Spaces  ")
				local result4 = string.slugify("CamelCase-Text_Here")
				
				return result1 == "hello-world" and
				       result2 == "this-that" and
				       result3 == "multiple-spaces" and
				       result4 == "camelcase-text_here"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string.slugify to work correctly")
				}
			},
		},
		{
			name: "string_truncate",
			script: `
				local result1 = string.truncate("Short text", 20)
				local result2 = string.truncate("This is a long text", 10)
				local result3 = string.truncate("Custom suffix", 10, "...")
				local result4 = string.truncate("No truncation needed", 100, "...")
				
				return result1 == "Short text" and
				       result2 == "This is..." and
				       result3 == "Custom ..." and
				       result4 == "No truncation needed"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string.truncate to work correctly")
				}
			},
		},
		{
			name: "string_split",
			script: `
				local result1 = string.split("a,b,c", ",")
				local result2 = string.split("hello world", " ")
				local result3 = string.split("no delimiter here", ",")
				
				return #result1 == 3 and result1[2] == "b" and
				       #result2 == 2 and result2[1] == "hello" and
				       #result3 == 1 and result3[1] == "no delimiter here"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string.split to work correctly")
				}
			},
		},
		{
			name: "string_trim",
			script: `
				local result1 = string.trim("  hello  ")
				local result2 = string.trim("\t\nworld\r\n")
				local result3 = string.trim("no trim")
				
				return result1 == "hello" and
				       result2 == "world" and
				       result3 == "no trim"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string.trim to work correctly")
				}
			},
		},
		{
			name: "string_case_conversions",
			script: `
				local result1 = string.capitalize("hello")
				local result2 = string.camelcase("hello-world")
				local result3 = string.camelcase("snake_case_text")
				local result4 = string.snakecase("camelCaseText")
				local result5 = string.snakecase("kebab-case-text")
				
				return result1 == "Hello" and
				       result2 == "helloWorld" and
				       result3 == "snakeCaseText" and
				       result4 == "camel_case_text" and
				       result5 == "kebab_case_text"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected string case conversions to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestTableUtilities(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "table_keys_values",
			script: `
				local tbl = {a = 1, b = 2, c = 3}
				local keys = table.keys(tbl)
				local values = table.values(tbl)
				
				-- Sort for consistent testing
				table.sort(keys)
				table.sort(values)
				
				return #keys == 3 and keys[1] == "a" and
				       #values == 3 and values[1] == 1
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected table.keys and table.values to work correctly")
				}
			},
		},
		{
			name: "table_merge",
			script: `
				local t1 = {a = 1, b = 2}
				local t2 = {b = 3, c = 4}
				local t3 = {d = 5}
				local result = table.merge(t1, t2, t3)
				
				return result.a == 1 and result.b == 3 and
				       result.c == 4 and result.d == 5
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected table.merge to work correctly")
				}
			},
		},
		{
			name: "table_deep_copy",
			script: `
				local original = {
					a = 1,
					b = {c = 2, d = {e = 3}},
					f = {4, 5, 6}
				}
				local copy = table.deep_copy(original)
				
				-- Modify copy
				copy.a = 10
				copy.b.c = 20
				copy.f[1] = 40
				
				return original.a == 1 and original.b.c == 2 and
				       original.f[1] == 4 and
				       copy.a == 10 and copy.b.c == 20 and copy.f[1] == 40
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected table.deep_copy to work correctly")
				}
			},
		},
		{
			name: "table_array_operations",
			script: `
				local arr = {1, 2, 3, 4, 5}
				
				-- Test slice
				local slice = table.slice(arr, 2, 4)
				
				-- Test reverse (modifies in place)
				local arr2 = {1, 2, 3}
				table.reverse(arr2)
				
				-- Test contains
				local has3 = table.contains(arr, 3)
				local has6 = table.contains(arr, 6)
				
				-- Test is_empty
				local empty1 = table.is_empty({})
				local empty2 = table.is_empty({1})
				
				return #slice == 3 and slice[1] == 2 and slice[3] == 4 and
				       arr2[1] == 3 and arr2[3] == 1 and
				       has3 == true and has6 == false and
				       empty1 == true and empty2 == false
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected table array operations to work correctly")
				}
			},
		},
		{
			name: "table_shuffle",
			script: `
				local arr = {1, 2, 3, 4, 5}
				local original = table.concat(arr, ",")
				
				-- Shuffle should modify in place
				table.shuffle(arr)
				local shuffled = table.concat(arr, ",")
				
				-- Check that all elements are still present
				table.sort(arr)
				local sorted = table.concat(arr, ",")
				
				return sorted == "1,2,3,4,5" and #arr == 5
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected table.shuffle to preserve all elements")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestCryptoUtilities(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue, err error)
	}{
		{
			name: "crypto_uuid",
			script: `
				local uuid1 = core.crypto.uuid()
				local uuid2 = core.crypto.uuid()
				
				-- Check format (simplified)
				local pattern = "^%x+%-%x+%-%x+%-%x+%-%x+$"
				
				return type(uuid1) == "string" and
				       type(uuid2) == "string" and
				       uuid1 ~= uuid2 and
				       string.match(uuid1, pattern) ~= nil
			`,
			check: func(t *testing.T, result lua.LValue, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != lua.LTrue {
					t.Errorf("Expected crypto.uuid to generate valid UUIDs")
				}
			},
		},
		{
			name: "crypto_random_string",
			script: `
				local str1 = core.crypto.random_string(16)
				local str2 = core.crypto.random_string(32)
				local str3 = core.crypto.random_string(8, "ABC")
				
				return type(str1) == "string" and #str1 == 16 and
				       type(str2) == "string" and #str2 == 32 and
				       type(str3) == "string" and #str3 == 8
			`,
			check: func(t *testing.T, result lua.LValue, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != lua.LTrue {
					t.Errorf("Expected crypto.random_string to generate strings of correct length")
				}
			},
		},
		{
			name: "crypto_hash_not_implemented",
			script: `
				local success, err = pcall(function()
					return core.crypto.hash("test", "sha256")
				end)
				
				return success == false and string.find(tostring(err), "bridge implementation") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != lua.LTrue {
					t.Errorf("Expected crypto.hash to require bridge implementation")
				}
			},
		},
		{
			name: "crypto_base64_not_implemented",
			script: `
				local success1, err1 = pcall(function()
					return core.crypto.base64_encode("test")
				end)
				
				local success2, err2 = pcall(function()
					return core.crypto.base64_decode("dGVzdA==")
				end)
				
				return success1 == false and string.find(tostring(err1), "bridge implementation") ~= nil and
				       success2 == false and string.find(tostring(err2), "bridge implementation") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != lua.LTrue {
					t.Errorf("Expected base64 functions to require bridge implementation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := L.DoString(tt.script)
			var result lua.LValue
			if err == nil {
				result = L.Get(-1)
				L.Pop(1)
			}
			tt.check(t, result, err)
		})
	}
}

func TestTimeUtilities(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "os_now",
			script: `
				local now1 = os.now()
				local now2 = os.now()
				
				return type(now1) == "number" and
				       type(now2) == "number" and
				       now2 >= now1
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected os.now to return increasing timestamps")
				}
			},
		},
		{
			name: "os_format",
			script: `
				local timestamp = 1634567890
				local formatted = os.format(timestamp, "%Y-%m-%d")
				
				-- Just check it returns a string with expected format
				return type(formatted) == "string" and
				       string.match(formatted, "%d%d%d%d%-%d%d%-%d%d") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected os.format to format timestamps correctly")
				}
			},
		},
		{
			name: "os_duration",
			script: `
				local start_time = 1000
				local end_time = 3661 -- 1 hour, 1 minute, 1 second later
				
				local duration = os.duration(start_time, end_time)
				
				return duration.seconds == 2661 and
				       math.abs(duration.minutes - 44.35) < 0.01 and
				       math.abs(duration.hours - 0.739) < 0.001
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected os.duration to calculate durations correctly")
				}
			},
		},
		{
			name: "os_add_time",
			script: `
				local base_time = 1000
				local result = os.add_time(base_time, {
					days = 1,
					hours = 2,
					minutes = 30,
					seconds = 45
				})
				
				-- 1 day = 86400, 2 hours = 7200, 30 min = 1800, 45 sec = 45
				local expected = base_time + 86400 + 7200 + 1800 + 45
				
				return result == expected
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected os.add_time to add time correctly")
				}
			},
		},
		{
			name: "os_humanize_duration",
			script: `
				local result1 = os.humanize_duration(3661)
				local result2 = os.humanize_duration(86461)
				local result3 = os.humanize_duration(45)
				local result4 = os.humanize_duration(0)
				
				return result1 == "1 hour 1 minute 1 second" and
				       result2 == "1 day 1 minute 1 second" and
				       result3 == "45 seconds" and
				       result4 == "0 seconds"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected os.humanize_duration to format durations correctly")
				}
			},
		},
		{
			name: "os_parse_time_basic",
			script: `
				local success, result = pcall(function()
					return os.parse_time("2023-10-17", "%Y-%m-%d")
				end)
				
				return success and type(result) == "number"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected os.parse_time to parse basic dates")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestMiscellaneousUtilities(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "type_checking",
			script: `
				local function test_func() end
				local callable_table = setmetatable({}, {__call = function() end})
				local regular_table = {}
				
				local array = {1, 2, 3}
				local object = {a = 1, b = 2}
				local mixed = {1, 2, x = 3}
				
				return core.is_callable(test_func) == true and
				       core.is_callable(callable_table) == true and
				       core.is_callable(regular_table) == false and
				       core.is_array(array) == true and
				       core.is_array(object) == false and
				       core.is_array(mixed) == false and
				       core.is_object(object) == true and
				       core.is_object(array) == false
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected type checking utilities to work correctly")
				}
			},
		},
		{
			name: "throttle_function",
			script: `
				local call_count = 0
				local throttled = core.throttle(function()
					call_count = call_count + 1
				end, 0.01) -- 10ms throttle
				
				-- Call multiple times quickly
				for i = 1, 5 do
					throttled()
				end
				
				-- Should only execute once due to throttle
				return call_count == 1
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected throttle to limit function calls")
				}
			},
		},
		{
			name: "memoize_function",
			script: `
				local call_count = 0
				local expensive_func = core.memoize(function(x, y)
					call_count = call_count + 1
					return x + y
				end)
				
				local r1 = expensive_func(1, 2)
				local r2 = expensive_func(1, 2) -- Should use cache
				local r3 = expensive_func(2, 3) -- Different args
				local r4 = expensive_func(1, 2) -- Should use cache
				
				return r1 == 3 and r2 == 3 and r3 == 5 and
				       r4 == 3 and call_count == 2
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected memoize to cache function results")
				}
			},
		},
		{
			name: "error_handling",
			script: `
				-- Test try-catch
				local result1 = core.try(function()
					return "success"
				end)
				
				local result2 = core.try(function()
					error("test error")
				end, function(err)
					return "caught: " .. tostring(err)
				end)
				
				-- Test safe_call
				local ok1, res1 = core.safe_call(function() return 42 end)
				local ok2, res2 = core.safe_call(function() error("fail") end)
				
				return result1 == "success" and
				       string.find(result2, "caught:") and
				       ok1 == true and res1 == 42 and
				       ok2 == false
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error handling utilities to work correctly")
				}
			},
		},
		{
			name: "debounce_simplified",
			script: `
				local call_count = 0
				local last_value = nil
				
				local debounced = core.debounce(function(val)
					call_count = call_count + 1
					last_value = val
				end, 0.001)
				
				-- Due to simplified implementation, it just calls immediately
				debounced("first")
				debounced("second")
				debounced("third")
				
				-- Each call executes immediately in simplified version
				return call_count == 3 and last_value == "third"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected simplified debounce to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestStringUtilityEdgeCases(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name          string
		script        string
		expectedError string
	}{
		{
			name: "template_with_invalid_input",
			script: `
				string.template(123, {})
			`,
			expectedError: "template must be a string",
		},
		{
			name: "slugify_with_invalid_input",
			script: `
				string.slugify(nil)
			`,
			expectedError: "text must be a string",
		},
		{
			name: "truncate_with_negative_length",
			script: `
				string.truncate("test", -1)
			`,
			expectedError: "length must be a non-negative number",
		},
		{
			name: "split_with_non_string",
			script: `
				string.split(123, ",")
			`,
			expectedError: "str must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := L.DoString(tt.script)
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestTableUtilityEdgeCases(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "deep_copy_with_circular_reference",
			script: `
				local t1 = {a = 1}
				local t2 = {b = t1}
				t1.c = t2 -- Create circular reference
				
				local copy = table.deep_copy(t1)
				
				-- Check structure is preserved
				return copy.a == 1 and copy.c.b.a == 1
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected deep_copy to handle circular references")
				}
			},
		},
		{
			name: "deep_copy_with_metatable",
			script: `
				local mt = {__index = {default = "value"}}
				local t = setmetatable({a = 1}, mt)
				
				local copy = table.deep_copy(t)
				
				-- Check metatable is preserved
				return copy.a == 1 and copy.default == "value"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected deep_copy to preserve metatables")
				}
			},
		},
		{
			name: "merge_with_nil_values",
			script: `
				local t1 = {a = 1, b = nil}
				local t2 = {b = 2, c = 3}
				local result = table.merge(t1, t2)
				
				return result.a == 1 and result.b == 2 and result.c == 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected merge to handle nil values correctly")
				}
			},
		},
		{
			name: "slice_with_out_of_bounds",
			script: `
				local arr = {1, 2, 3}
				local slice1 = table.slice(arr, 2, 10)
				local slice2 = table.slice(arr, 5, 10)
				local slice3 = table.slice(arr)
				
				return #slice1 == 2 and slice1[1] == 2 and
				       #slice2 == 0 and
				       #slice3 == 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected slice to handle out of bounds gracefully")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestErrorHandlingEdgeCases(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	tests := []struct {
		name          string
		script        string
		expectedError string
	}{
		{
			name: "keys_with_non_table",
			script: `
				table.keys("not a table")
			`,
			expectedError: "tbl must be a table",
		},
		{
			name: "is_callable_with_invalid_metatable",
			script: `
				local t = setmetatable({}, {__call = "not a function"})
				return core.is_callable(t)
			`,
			expectedError: "", // Should return false, not error
		},
		{
			name: "throttle_with_invalid_delay",
			script: `
				core.throttle(function() end, -1)
			`,
			expectedError: "delay must be a non-negative number",
		},
		{
			name: "memoize_with_non_function",
			script: `
				core.memoize("not a function")
			`,
			expectedError: "func must be a function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := L.DoString(tt.script)
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConcurrentUtilityUsage(t *testing.T) {
	// Test that utilities can be used concurrently
	done := make(chan bool, 3)

	go func() {
		L := lua.NewState()
		defer L.Close()
		setupCoreLibrary(t, L)

		script := `
			for i = 1, 10 do
				local uuid = core.crypto.uuid()
				local slug = string.slugify("Test " .. i)
			end
			return true
		`
		if err := L.DoString(script); err != nil {
			t.Errorf("Concurrent test 1 failed: %v", err)
		}
		done <- true
	}()

	go func() {
		L := lua.NewState()
		defer L.Close()
		setupCoreLibrary(t, L)

		script := `
			for i = 1, 10 do
				local t = {a = i, b = i * 2}
				local copy = table.deep_copy(t)
				local keys = table.keys(t)
			end
			return true
		`
		if err := L.DoString(script); err != nil {
			t.Errorf("Concurrent test 2 failed: %v", err)
		}
		done <- true
	}()

	go func() {
		L := lua.NewState()
		defer L.Close()
		setupCoreLibrary(t, L)

		script := `
			for i = 1, 10 do
				local now = os.now()
				local formatted = os.format(now, "%Y-%m-%d")
			end
			return true
		`
		if err := L.DoString(script); err != nil {
			t.Errorf("Concurrent test 3 failed: %v", err)
		}
		done <- true
	}()

	// Wait for all goroutines with timeout
	timeout := time.After(5 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Concurrent test timed out")
		}
	}
}

func TestCorePackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	script := `
		-- Test that core can be used as a module
		local core_module = core
		
		return type(core_module) == "table" and
		       type(core_module.crypto) == "table" and
		       type(core_module.is_callable) == "function" and
		       type(core_module.debounce) == "function" and
		       type(core_module.memoize) == "function" and
		       type(core_module.try) == "function"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Package require test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected core module to be properly exported")
	}
}

func TestCoreIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupCoreLibrary(t, L)

	// Integration test that uses multiple utilities together
	script := `
		-- Simulate a real-world use case
		local function process_user_data(users)
			local results = {}
			
			for _, user in ipairs(users) do
				-- Generate ID if missing
				if not user.id then
					user.id = core.crypto.uuid()
				end
				
				-- Create URL-safe username
				user.slug = string.slugify(user.name)
				
				-- Format registration time
				if user.registered then
					user.registered_formatted = os.format(user.registered, "%Y-%m-%d")
				end
				
				-- Deep copy for safety
				local processed = table.deep_copy(user)
				
				-- Add some metadata
				processed.processed_at = os.now()
				processed.metadata = {
					version = "1.0",
					processor = "core_utils"
				}
				
				table.insert(results, processed)
			end
			
			return results
		end
		
		-- Test data
		local users = {
			{name = "John Doe", email = "john@example.com", registered = 1634567890},
			{name = "Jane Smith", email = "jane@example.com"},
			{name = "Bob O'Brien", email = "bob@example.com", registered = 1634567890}
		}
		
		-- Process with error handling
		local success, processed = core.safe_call(process_user_data, users)
		
		return success and #processed == 3 and
		       processed[1].slug == "john-doe" and
		       processed[2].id ~= nil and
		       processed[3].slug == "bob-o-brien" and
		       processed[1].metadata.version == "1.0"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected integration test to pass")
	}
}
