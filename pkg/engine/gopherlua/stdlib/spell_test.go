// ABOUTME: Comprehensive test suite for Spell Framework Library in Lua standard library
// ABOUTME: Tests spell lifecycle, composition, parameter handling, and context management

package stdlib

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// setupSpellLibrary loads the spell library
func setupSpellLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Load the spell library
	libPath := filepath.Join(".", "spell.lua")
	if err := L.DoFile(libPath); err != nil {
		t.Fatalf("Failed to load spell library: %v", err)
	}
	spell := L.Get(-1)
	L.SetGlobal("spell", spell)
}

func TestSpellLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	// Test that the spell library was loaded
	script := `
		return type(spell) == "table" and
		       type(spell.init) == "function" and
		       type(spell.params) == "function" and
		       type(spell.output) == "function" and
		       type(spell.compose) == "function" and
		       type(spell.library) == "function" and
		       type(spell.context) == "function" and
		       type(spell.config) == "function" and
		       type(spell.cache) == "function"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected spell library to be properly loaded")
	}
}

func TestSpellInitialization(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "basic_initialization",
			script: `
				local config = spell.init({
					name = "test-spell",
					version = "1.0.0",
					description = "Test spell",
					author = "Test Author",
					params = {
						input = { type = "string", required = true }
					}
				})
				
				return config.name == "test-spell" and
				       config.version == "1.0.0" and
				       config.description == "Test spell"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected basic initialization to work correctly")
				}
			},
		},
		{
			name: "default_configuration",
			script: `
				local config = spell.init({
					name = "minimal-spell"
				})
				
				return config.name == "minimal-spell" and
				       config.version == "1.0.0" and
				       type(config.timeout) == "number" and
				       type(config.max_retries) == "number"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected default configuration to be applied")
				}
			},
		},
		{
			name: "context_creation",
			script: `
				spell.init({
					name = "context-test",
					version = "2.0.0"
				})
				
				local context = spell.context()
				
				return context.spell_name == "context-test" and
				       type(context.execution_id) == "string" and
				       type(context.start_time) == "number" and
				       context.metadata.version == "2.0.0"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected context to be created correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset spell state before each test
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestParameterHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "parameter_definition",
			script: `
				spell.init({
					name = "param-test",
					params = {
						test_param = { type = "string", default = "default_value" }
					}
				})
				
				local param_config = spell.params("test_param", { type = "string", required = true })
				
				return param_config.type == "string" and param_config.required == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter definition to work correctly")
				}
			},
		},
		{
			name: "parameter_with_global_params",
			script: `
				-- Simulate global params
				params = { test_input = "hello world" }
				
				spell.init({
					name = "param-test",
					params = {
						test_input = { type = "string", required = true }
					}
				})
				
				local value = spell.params("test_input")
				
				return value == "hello world"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter to be read from global params")
				}
			},
		},
		{
			name: "parameter_default_values",
			script: `
				-- No global params
				params = {}
				
				spell.init({
					name = "default-test",
					params = {
						optional_param = { type = "string", default = "default_value" }
					}
				})
				
				local value = spell.params("optional_param")
				
				return value == "default_value"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter default value to be used")
				}
			},
		},
		{
			name: "parameter_type_validation",
			script: `
				params = { number_param = "not_a_number" }
				
				spell.init({
					name = "validation-test",
					params = {
						number_param = { type = "number", required = true }
					}
				})
				
				local success, err = pcall(function()
					return spell.params("number_param")
				end)
				
				return success == false and string.find(tostring(err), "must be of type") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter type validation to work")
				}
			},
		},
		{
			name: "parameter_enum_validation",
			script: `
				params = { choice = "invalid_choice" }
				
				spell.init({
					name = "enum-test",
					params = {
						choice = { type = "string", enum = {"option1", "option2", "option3"} }
					}
				})
				
				local success, err = pcall(function()
					return spell.params("choice")
				end)
				
				return success == false and string.find(tostring(err), "must be one of") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter enum validation to work")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset spell state and globals before each test
			if err := L.DoString("spell.reset(); params = nil"); err != nil {
				t.Fatalf("Failed to reset state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestSpellComposition(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "basic_composition",
			script: `
				spell.init({
					name = "composition-test",
					params = { topic = "test topic" }
				})
				
				local results = spell.compose({
					{
						name = "fetch",
						spell = "web-fetcher",
						params = { url = "https://example.com/$topic" }
					},
					{
						name = "summarize",
						spell = "text-summarizer",
						params = { text = "$fetch.content" }
					}
				})
				
				return type(results) == "table" and
				       results.fetch ~= nil and
				       results.summarize ~= nil and
				       results.fetch.spell == "web-fetcher"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected basic composition to work correctly")
				}
			},
		},
		{
			name: "variable_substitution",
			script: `
				params = { base_url = "https://api.example.com" }
				
				spell.init({
					name = "substitution-test",
					params = { 
						base_url = { type = "string", required = true }
					}
				})
				
				-- Get the parameter to populate context.params
				local base_url = spell.params("base_url")
				
				local results = spell.compose({
					{
						name = "api_call",
						spell = "http-client",
						params = { url = "$base_url/data" }
					}
				})
				
				-- Check that variable was substituted
				return results.api_call.params.url == "https://api.example.com/data"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected variable substitution to work correctly")
				}
			},
		},
		{
			name: "step_chaining",
			script: `
				spell.init({ name = "chaining-test" })
				
				local results = spell.compose({
					{ name = "step1", spell = "processor1", params = { input = "initial" } },
					{ name = "step2", spell = "processor2", params = { input = "$step1.output" } },
					{ name = "step3", spell = "processor3", params = { input = "$step2.output" } }
				})
				
				-- Count results manually since # doesn't work with string keys
				local count = 0
				for _ in pairs(results) do
					count = count + 1
				end
				
				return count == 3 and
				       results.step1 ~= nil and
				       results.step2 ~= nil and
				       results.step3 ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected step chaining to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestLibraryManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "library_creation",
			script: `
				local lib = spell.library("test-utils", {
					helper1 = function(x) return x * 2 end,
					helper2 = function(a, b) return a + b end,
					helper3 = function(str) return string.upper(str) end
				})
				
				return type(lib) == "table" and
				       type(lib.helper1) == "function" and
				       lib.helper1(5) == 10 and
				       lib.helper2(3, 4) == 7 and
				       lib.helper3("hello") == "HELLO"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected library creation to work correctly")
				}
			},
		},
		{
			name: "library_inclusion",
			script: `
				-- First create a library
				spell.library("math-utils", {
					square = function(x) return x * x end,
					cube = function(x) return x * x * x end
				})
				
				-- Then include it
				local math_lib = spell.include("math-utils")
				
				return type(math_lib) == "table" and
				       math_lib.square(4) == 16 and
				       math_lib.cube(3) == 27
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected library inclusion to work correctly")
				}
			},
		},
		{
			name: "get_libraries",
			script: `
				spell.library("lib1", { func1 = function() end })
				spell.library("lib2", { func2 = function() end })
				
				local libs = spell.get_libraries()
				
				return type(libs) == "table" and
				       libs.lib1 ~= nil and
				       libs.lib2 ~= nil and
				       type(libs.lib1.func1) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected get_libraries to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestCacheManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "cache_set_get",
			script: `
				spell.init({ name = "cache-test" })
				
				-- Set cache value
				local set_result = spell.cache("test_key", "test_value", 60)
				
				-- Get cache value
				local get_result = spell.cache("test_key")
				
				return set_result == "test_value" and get_result == "test_value"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected cache set/get to work correctly")
				}
			},
		},
		{
			name: "cache_expiration",
			script: `
				spell.init({ name = "expiry-test" })
				
				-- Set cache value with very short TTL
				spell.cache("expiry_key", "expiry_value", 0) -- Immediate expiration
				
				-- Try to get expired value
				local value = spell.cache("expiry_key")
				
				return value == nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected cache expiration to work correctly")
				}
			},
		},
		{
			name: "cache_statistics",
			script: `
				spell.init({ name = "stats-test" })
				
				-- Add some cache entries
				spell.cache("key1", "value1", 60)
				spell.cache("key2", "value2", 60)
				spell.cache("key3", "value3", 0) -- Expired
				
				local stats = spell.get_cache_stats()
				
				return type(stats) == "table" and
				       stats.total_entries >= 2 and
				       type(stats.expired_entries) == "number"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected cache statistics to work correctly")
				}
			},
		},
		{
			name: "clear_expired_cache",
			script: `
				spell.init({ name = "clear-test" })
				
				-- Add mix of valid and expired entries
				spell.cache("valid", "value", 60)
				spell.cache("expired", "value", 0)
				
				local cleared = spell.clear_expired_cache()
				
				return type(cleared) == "number" and cleared >= 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected clear expired cache to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestLifecycleHooks(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "init_hooks",
			script: `
				local hook_called = false
				
				spell.on_init(function(config, context)
					hook_called = true
					return config.name == "hook-test"
				end)
				
				spell.init({ name = "hook-test" })
				
				return hook_called == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected init hooks to be called")
				}
			},
		},
		{
			name: "complete_hooks",
			script: `
				local hook_called = false
				local hook_data = nil
				
				spell.on_complete(function(output)
					hook_called = true
					hook_data = output.data
				end)
				
				spell.init({ name = "complete-test" })
				spell.output("test result", "text")
				
				return hook_called == true and hook_data == "test result"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected complete hooks to be called")
				}
			},
		},
		{
			name: "cleanup_hooks",
			script: `
				local cleanup_called = false
				
				spell.on_cleanup(function()
					cleanup_called = true
				end)
				
				spell.init({ name = "cleanup-test" })
				spell.cleanup_resources()
				
				return cleanup_called == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected cleanup hooks to be called")
				}
			},
		},
		{
			name: "error_hooks",
			script: `
				local error_handled = false
				
				spell.on_error(function(err)
					error_handled = true
					return { error = err, handled = true }
				end)
				
				spell.init({ name = "error-test" })
				
				-- Just test that the hook can be registered
				return error_handled == false -- Hook not called yet
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error hooks to be registerable")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestEnvironmentManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "env_set_get",
			script: `
				spell.init({ name = "env-test" })
				
				-- Set environment variable
				local set_result = spell.env("TEST_VAR", "test_value")
				
				-- Get environment variable
				local get_result = spell.env("TEST_VAR")
				
				return set_result == "test_value" and get_result == "test_value"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected environment set/get to work correctly")
				}
			},
		},
		{
			name: "sandbox_creation",
			script: `
				spell.init({ name = "sandbox-test" })
				
				local sandbox = spell.sandbox({
					globals = {
						custom_var = "custom_value"
					}
				})
				
				return type(sandbox) == "table" and
				       type(sandbox.spell) == "table" and
				       sandbox.custom_var == "custom_value" and
				       type(sandbox.string) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected sandbox creation to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestResourceManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "resource_registration",
			script: `
				spell.init({ name = "resource-test" })
				
				local resource = spell.resource("test_resource", {
					type = "file",
					data = { filename = "test.txt" },
					cleanup = function(res) 
						-- Cleanup logic here
						return true
					end
				})
				
				return type(resource) == "table" and
				       resource.name == "test_resource" and
				       resource.type == "file" and
				       type(resource.cleanup) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected resource registration to work correctly")
				}
			},
		},
		{
			name: "resource_cleanup",
			script: `
				spell.init({ name = "cleanup-test" })
				
				local cleanup_called = false
				
				spell.resource("test_resource", {
					type = "test",
					cleanup = function(res)
						cleanup_called = true
						return true
					end
				})
				
				spell.cleanup_resources()
				
				return cleanup_called == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected resource cleanup to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestSpellOutput(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "text_output",
			script: `
				spell.init({ name = "output-test" })
				
				local output = spell.output("Hello, World!", "text")
				
				return type(output) == "table" and
				       output.spell == "output-test" and
				       output.data == "Hello, World!" and
				       output.format == "text"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected text output to work correctly")
				}
			},
		},
		{
			name: "json_output",
			script: `
				spell.init({ name = "json-test" })
				
				local data = { result = "success", count = 42 }
				local output = spell.output(data, "json")
				
				return type(output) == "table" and
				       output.format == "json" and
				       output.data == data
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected JSON output to work correctly")
				}
			},
		},
		{
			name: "output_with_metadata",
			script: `
				spell.init({ name = "metadata-test" })
				
				local metadata = { version = "1.0", author = "Test" }
				local output = spell.output("test data", "text", metadata)
				
				return output.metadata.version == "1.0" and
				       output.metadata.author == "Test"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected output with metadata to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestConfigurationAccess(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "config_access",
			script: `
				spell.init({
					name = "config-test",
					timeout = 120,
					max_retries = 5,
					custom_setting = "custom_value"
				})
				
				local timeout = spell.config("timeout")
				local retries = spell.config("max_retries")
				local custom = spell.config("custom_setting")
				local missing = spell.config("missing_key", "default")
				
				return timeout == 120 and
				       retries == 5 and
				       custom == "custom_value" and
				       missing == "default"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected config access to work correctly")
				}
			},
		},
		{
			name: "full_config_access",
			script: `
				spell.init({
					name = "full-config-test",
					version = "2.0.0"
				})
				
				local config = spell.config()
				
				return type(config) == "table" and
				       config.name == "full-config-test" and
				       config.version == "2.0.0"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected full config access to work correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset()"); err != nil {
				t.Fatalf("Failed to reset spell state: %v", err)
			}

			if err := L.DoString(tt.script); err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
			L.Pop(1)
		})
	}
}

func TestSpellValidation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name          string
		script        string
		expectedError string
	}{
		{
			name: "invalid_config_type",
			script: `
				spell.init("not a table")
			`,
			expectedError: "config must be a table",
		},
		{
			name: "missing_required_parameter",
			script: `
				params = {}
				
				spell.init({
					name = "validation-test",
					params = {
						required_param = { type = "string", required = true }
					}
				})
				
				spell.params("required_param")
			`,
			expectedError: "Required parameter 'required_param' is missing",
		},
		{
			name: "invalid_parameter_type",
			script: `
				params = { number_param = "not_a_number" }
				
				spell.init({
					name = "type-validation-test",
					params = {
						number_param = { type = "number" }
					}
				})
				
				spell.params("number_param")
			`,
			expectedError: "must be of type number",
		},
		{
			name: "invalid_library_functions",
			script: `
				spell.library("invalid-lib", {
					not_a_function = "this is a string"
				})
			`,
			expectedError: "must be a function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := L.DoString("spell.reset(); params = nil"); err != nil {
				t.Fatalf("Failed to reset state: %v", err)
			}

			err := L.DoString(tt.script)
			if err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.expectedError)
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestSystemInfo(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	script := `
		spell.init({ name = "system-test" })
		
		-- Add some libraries and resources
		spell.library("test-lib", { func1 = function() end })
		spell.resource("test-resource", { type = "test" })
		spell.cache("test-key", "test-value")
		
		local info = spell.get_system_info()
		
		return type(info) == "table" and
		       type(info.lua_version) == "string" and
		       info.current_spell == "system-test" and
		       type(info.execution_id) == "string" and
		       type(info.hooks_registered) == "table"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("System info test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected system info to be returned correctly")
	}
}

func TestSpellValidateConfig(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "valid_config",
			script: `
				local valid, err = spell.validate_config({
					name = "test-spell",
					version = "1.0.0",
					params = {
						input = { type = "string", required = true }
					}
				})
				
				return valid == true and err == nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected valid config to pass validation")
				}
			},
		},
		{
			name: "invalid_config_type",
			script: `
				local valid, err = spell.validate_config("not a table")
				
				return valid == false and string.find(tostring(err), "must be a table") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected invalid config type to fail validation")
				}
			},
		},
		{
			name: "invalid_param_config",
			script: `
				local valid, err = spell.validate_config({
					name = "test",
					params = {
						bad_param = "not a table"
					}
				})
				
				return valid == false and string.find(tostring(err), "must be a table") ~= nil
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected invalid parameter config to fail validation")
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

func TestSpellIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	// Integration test that uses multiple features together
	script := `
		-- Simulate a realistic spell scenario
		params = {
			input_text = "Hello, World!",
			output_format = "json",
			enable_cache = true
		}
		
		-- Initialize spell with full configuration
		spell.init({
			name = "integration-test-spell",
			version = "1.2.3",
			description = "Integration test spell",
			author = "Test Suite",
			params = {
				input_text = { type = "string", required = true },
				output_format = { type = "string", default = "text", enum = {"text", "json", "xml"} },
				enable_cache = { type = "boolean", default = false }
			},
			timeout = 60,
			cache_ttl = 120
		})
		
		-- Set up lifecycle hooks
		local hook_events = {}
		
		spell.on_init(function(config, context)
			table.insert(hook_events, "init")
		end)
		
		spell.on_complete(function(output)
			table.insert(hook_events, "complete")
		end)
		
		-- Create a utility library
		spell.library("text-utils", {
			reverse = function(str)
				return string.reverse(str)
			end,
			uppercase = function(str)
				return string.upper(str)
			end
		})
		
		-- Get and validate parameters
		local input = spell.params("input_text")
		local format = spell.params("output_format")
		local use_cache = spell.params("enable_cache")
		
		-- Use cache if enabled
		local cache_key = "processed_" .. input
		local result = nil
		
		if use_cache then
			result = spell.cache(cache_key)
		end
		
		if not result then
			-- Process the input using our library
			local utils = spell.include("text-utils")
			result = utils.uppercase(utils.reverse(input))
			
			-- Cache the result if caching is enabled
			if use_cache then
				spell.cache(cache_key, result, 60)
			end
		end
		
		-- Create a managed resource
		spell.resource("temp_data", {
			type = "memory",
			data = { processed_result = result },
			cleanup = function(res)
				res.data = nil
				return true
			end
		})
		
		-- Compose a simple workflow
		local workflow_results = spell.compose({
			{
				name = "validate",
				spell = "input-validator",
				params = { input = "$input_text" }
			},
			{
				name = "transform",
				spell = "text-transformer",
				params = { text = "$validate.output", method = "reverse" }
			}
		})
		
		-- Output the final result
		local output = spell.output({
			original = input,
			processed = result,
			workflow = workflow_results,
			cache_used = use_cache,
			hooks_fired = hook_events
		}, format, {
			processing_time = 0.05,
			spell_version = spell.config("version")
		})
		
		-- Clean up resources
		spell.cleanup_resources()
		
		-- Verify everything worked
		return type(output) == "table" and
		       output.data.original == "Hello, World!" and
		       output.data.processed == "!DLROW ,OLLEH" and
		       output.format == "json" and
		       #hook_events >= 1 and
		       workflow_results.validate ~= nil and
		       spell.context().spell_name == "integration-test-spell"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected integration test to pass")
	}
}

func TestSpellConcurrency(t *testing.T) {
	// Test that the spell framework can be used safely in concurrent scenarios
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			L := lua.NewState()
			defer L.Close()
			setupSpellLibrary(t, L)

			script := fmt.Sprintf(`
				spell.init({
					name = "concurrent-spell-%d",
					params = {
						id = { type = "number", default = %d }
					}
				})
				
				-- Use various features concurrently
				spell.cache("concurrent_key_%d", "value_%d", 60)
				
				spell.library("concurrent_lib_%d", {
					test_func = function() return %d end
				})
				
				local context = spell.context()
				local config = spell.config()
				
				return context.spell_name and config.name and true
			`, id, id, id, id, id, id)

			if err := L.DoString(script); err != nil {
				t.Errorf("Concurrent test %d failed: %v", id, err)
			}

			result := L.Get(-1)
			if result != lua.LTrue {
				t.Errorf("Concurrent test %d returned false", id)
			}

			done <- true
		}(i)
	}

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

func TestSpellPackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupSpellLibrary(t, L)

	script := `
		-- Test that spell can be used as a module
		local spell_module = spell
		
		return type(spell_module) == "table" and
		       type(spell_module.init) == "function" and
		       type(spell_module.params) == "function" and
		       type(spell_module.output) == "function" and
		       type(spell_module.compose) == "function" and
		       type(spell_module.library) == "function" and
		       type(spell_module.context) == "function" and
		       type(spell_module.config) == "function" and
		       type(spell_module.cache) == "function" and
		       type(spell_module.reset) == "function"
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Package require test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected spell module to be properly exported")
	}
}
