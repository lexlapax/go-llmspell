// ABOUTME: Comprehensive test suite for Tools & Registry Library in Lua standard library
// ABOUTME: Tests tool registration, execution, validation, composition, and registry management

package stdlib

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge/agent"
	lua "github.com/yuin/gopher-lua"
)

// setupToolsLibrary loads the tools library and sets up required bridges
func setupToolsLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Set up mock tools bridge
	toolsBridge := agent.NewToolsBridge()
	err := toolsBridge.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Failed to initialize tools bridge: %v", err)
	}

	// Create mock global tools bridge for testing
	toolsTable := L.NewTable()

	// Mock tool discovery methods
	toolsTable.RawSetString("listTools", L.NewFunction(func(L *lua.LState) int {
		result := L.NewTable()
		// Mock tool 1
		tool1 := L.NewTable()
		tool1.RawSetString("name", lua.LString("mock_tool_1"))
		tool1.RawSetString("description", lua.LString("Mock tool for testing"))
		tool1.RawSetString("category", lua.LString("test"))
		tool1.RawSetString("version", lua.LString("1.0.0"))
		result.RawSetInt(1, tool1)

		// Mock tool 2
		tool2 := L.NewTable()
		tool2.RawSetString("name", lua.LString("mock_tool_2"))
		tool2.RawSetString("description", lua.LString("Another mock tool"))
		tool2.RawSetString("category", lua.LString("utility"))
		tool2.RawSetString("version", lua.LString("1.1.0"))
		result.RawSetInt(2, tool2)

		L.Push(result)
		return 1
	}))

	toolsTable.RawSetString("searchTools", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1) // Ignore query for mock
		result := L.NewTable()
		tool := L.NewTable()
		tool.RawSetString("name", lua.LString("search_result"))
		tool.RawSetString("description", lua.LString("Found tool"))
		result.RawSetInt(1, tool)
		L.Push(result)
		return 1
	}))

	toolsTable.RawSetString("getToolInfo", L.NewFunction(func(L *lua.LState) int {
		toolName := L.CheckString(1)
		if toolName == "existing_tool" {
			tool := L.NewTable()
			tool.RawSetString("name", lua.LString("existing_tool"))
			tool.RawSetString("description", lua.LString("An existing tool"))
			tool.RawSetString("category", lua.LString("test"))

			// Add parameters schema
			params := L.NewTable()
			required := L.NewTable()
			required.RawSetInt(1, lua.LString("input"))
			params.RawSetString("required", required)

			properties := L.NewTable()
			inputProp := L.NewTable()
			inputProp.RawSetString("type", lua.LString("string"))
			properties.RawSetString("input", inputProp)
			params.RawSetString("properties", properties)

			tool.RawSetString("parameters", params)

			// Add execute function
			tool.RawSetString("execute", L.NewFunction(func(L *lua.LState) int {
				input := L.CheckTable(1)
				result := L.NewTable()
				result.RawSetString("processed", lua.LString("success"))
				result.RawSetString("input_received", input.RawGetString("input"))
				L.Push(result)
				return 1
			}))

			L.Push(tool)
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))

	toolsTable.RawSetString("registerCustomTool", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1) // Ignore tool definition for mock
		L.Push(lua.LTrue)   // Always succeed in mock
		return 1
	}))

	L.SetGlobal("tools", toolsTable)

	// Load the tools library
	toolsPath := filepath.Join(".", "tools.lua")
	err = L.DoFile(toolsPath)
	if err != nil {
		t.Fatalf("Failed to load tools library: %v", err)
	}
	tools := L.Get(-1)
	L.SetGlobal("tools", tools)
}

// TestToolsLibraryLoading tests that the tools library can be loaded
func TestToolsLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupToolsLibrary(t, L)

	// Check that tools table exists and has expected functions
	script := `
		if type(tools) ~= "table" then
			error("Tools module should be a table")
		end
		
		local required_functions = {
			"define", "register_library", "compose",
			"execute_safe", "pipeline", "parallel_execute",
			"validate_params", "test_tool", "benchmark_tool",
			"list", "search", "get_info", "get_metrics", "get_history"
		}
		
		for _, func_name in ipairs(required_functions) do
			if type(tools[func_name]) ~= "function" then
				error("Function " .. func_name .. " should be available")
			end
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Tools library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}
}

// TestToolRegistration tests tool registration and discovery
func TestToolRegistration(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "define_basic_tool",
			script: `
				local tool = tools.define("test_tool", "A test tool", {}, function(params)
					return "processed: " .. (params.input or "no input")
				end)
				return tool.name == "test_tool" and tool.description == "A test tool"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected basic tool definition to work, got %v", result)
				}
			},
		},
		{
			name: "define_complex_tool",
			script: `
				local schema = {
					version = "2.0.0",
					category = "data",
					tags = {"processing", "test"},
					parameters = {
						required = {"input", "mode"},
						properties = {
							input = {type = "string"},
							mode = {type = "string"}
						}
					}
				}
				local tool = tools.define("complex_tool", "Complex test tool", schema, function(params)
					return {result = params.input, mode = params.mode}
				end)
				return tool.version == "2.0.0" and tool.category == "data"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected complex tool definition to work, got %v", result)
				}
			},
		},
		{
			name: "register_library",
			script: `
				local library = {
					lib_tool_1 = function(params) return "lib1: " .. params.input end,
					lib_tool_2 = {
						description = "Library tool 2",
						execute = function(params) return "lib2: " .. params.input end
					}
				}
				local result = tools.register_library(library)
				return result.registered == 2 and result.failed == 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected library registration to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestToolExecution tests tool execution utilities
func TestToolExecution(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "execute_safe_basic",
			script: `
				-- Define a simple tool
				tools.define("math_tool", "Math operations", {}, function(params)
					return params.a + params.b
				end)
				
				local result = tools.execute_safe("math_tool", {a = 5, b = 3})
				return result == 8
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected safe execution to work, got %v", result)
				}
			},
		},
		{
			name: "pipeline_execution",
			script: `
				-- Define tools for pipeline
				tools.define("add_one", "Add one", {}, function(params)
					return params + 1
				end)
				tools.define("multiply_two", "Multiply by two", {}, function(params)
					return params * 2
				end)
				
				local result = tools.pipeline({"add_one", "multiply_two"}, 5)
				return result.final_output == 12 and result.success == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected pipeline execution to work, got %v", result)
				}
			},
		},
		{
			name: "parallel_execute",
			script: `
				-- Define tools for parallel execution
				tools.define("tool_a", "Tool A", {}, function(params)
					return "A: " .. params.input
				end)
				tools.define("tool_b", "Tool B", {}, function(params)
					return "B: " .. params.input
				end)
				
				local result = tools.parallel_execute({"tool_a", "tool_b"}, {input = "test"})
				return result.success_count == 2 and result.error_count == 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parallel execution to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestToolComposition tests tool composition functionality
func TestToolComposition(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "compose_pipeline",
			script: `
				-- Define tools
				tools.define("step1", "Step 1", {}, function(params)
					return params .. "_step1"
				end)
				tools.define("step2", "Step 2", {}, function(params)
					return params .. "_step2"
				end)
				
				-- Compose as pipeline
				local composite = tools.compose({"step1", "step2"}, {mode = "pipeline"})
				local result = tools.execute_safe(composite, "input")
				return result == "input_step1_step2"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected pipeline composition to work, got %v", result)
				}
			},
		},
		{
			name: "compose_parallel",
			script: `
				-- Define tools
				tools.define("parallel1", "Parallel 1", {}, function(params)
					return "P1: " .. params.data
				end)
				tools.define("parallel2", "Parallel 2", {}, function(params)
					return "P2: " .. params.data
				end)
				
				-- Compose as parallel
				local composite = tools.compose({"parallel1", "parallel2"}, {mode = "parallel"})
				local result = tools.execute_safe(composite, {data = "test"})
				return result.success_count == 2
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parallel composition to work, got %v", result)
				}
			},
		},
		{
			name: "compose_conditional",
			script: `
				-- Define conditional tools
				local tools_list = {
					{
						condition = function(input) return input.type == "A" end,
						execute = function(input) return "Processed A: " .. input.value end
					},
					{
						condition = function(input) return input.type == "B" end,
						execute = function(input) return "Processed B: " .. input.value end
					}
				}
				
				local composite = tools.compose(tools_list, {mode = "conditional"})
				local result = tools.execute_safe(composite, {type = "A", value = "test"})
				return result == "Processed A: test"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected conditional composition to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestToolValidation tests parameter validation
func TestToolValidation(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "validate_params_success",
			script: `
				local tool = {
					parameters = {
						required = {"name", "age"},
						properties = {
							name = {type = "string"},
							age = {type = "number"}
						}
					}
				}
				local result = tools.validate_params(tool, {name = "John", age = 30})
				return result.valid == true and #result.errors == 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter validation to succeed, got %v", result)
				}
			},
		},
		{
			name: "validate_params_missing_required",
			script: `
				local tool = {
					parameters = {
						required = {"name", "age"},
						properties = {
							name = {type = "string"},
							age = {type = "number"}
						}
					}
				}
				local result = tools.validate_params(tool, {name = "John"})
				return result.valid == false and #result.errors > 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter validation to fail for missing required, got %v", result)
				}
			},
		},
		{
			name: "validate_params_wrong_type",
			script: `
				local tool = {
					parameters = {
						required = {"age"},
						properties = {
							age = {type = "number"}
						}
					}
				}
				local result = tools.validate_params(tool, {age = "thirty"})
				return result.valid == false and #result.errors > 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected parameter validation to fail for wrong type, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestToolTesting tests the tool testing functionality
func TestToolTesting(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupToolsLibrary(t, L)

	script := `
		-- Define a tool to test
		tools.define("string_processor", "Process strings", {}, function(params)
			return string.upper(params.input)
		end)
		
		-- Define test cases
		local test_cases = {
			{
				name = "uppercase_test",
				input = {input = "hello"},
				expected = "HELLO"
			},
			{
				name = "empty_string_test", 
				input = {input = ""},
				expected = ""
			},
			{
				name = "validator_test",
				input = {input = "world"},
				validator = function(result) return result == "WORLD" end
			}
		}
		
		local test_result = tools.test_tool("string_processor", test_cases)
		return test_result.summary.total == 3 and 
		       test_result.summary.passed == 3 and 
		       test_result.summary.failed == 0 and
		       test_result.summary.pass_rate == 1.0
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Tool testing test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected tool testing to work correctly, got %v", result)
	}
}

// TestToolBenchmarking tests performance benchmarking
func TestToolBenchmarking(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupToolsLibrary(t, L)

	script := `
		-- Define a tool to benchmark
		tools.define("simple_math", "Simple math", {}, function(params)
			local result = 0
			for i = 1, params.iterations do
				result = result + i
			end
			return result
		end)
		
		local benchmark_result = tools.benchmark_tool("simple_math", {iterations = 100}, {iterations = 5})
		
		return benchmark_result.iterations == 5 and
		       benchmark_result.successful > 0 and
		       benchmark_result.avg_time > 0 and
		       benchmark_result.throughput > 0
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Tool benchmarking test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected tool benchmarking to work correctly, got %v", result)
	}
}

// TestToolDiscovery tests tool discovery and information retrieval
func TestToolDiscovery(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "list_tools",
			script: `
				local tool_list = tools.list()
				return type(tool_list) == "table" and #tool_list >= 2
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected tool listing to work, got %v", result)
				}
			},
		},
		{
			name: "search_tools",
			script: `
				local search_results = tools.search("test")
				return type(search_results) == "table" and #search_results > 0
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected tool search to work, got %v", result)
				}
			},
		},
		{
			name: "get_tool_info",
			script: `
				local tool_info = tools.get_info("existing_tool")
				return tool_info ~= nil and tool_info.name == "existing_tool"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected tool info retrieval to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestToolErrorHandling tests error handling in tool operations
func TestToolErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_tool_name_error",
			script: `
				local success, err = pcall(function()
					tools.define(nil, "description", {}, function() end)
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing tool name, got %v", result)
				}
			},
		},
		{
			name: "invalid_tool_function_error",
			script: `
				local success, err = pcall(function()
					tools.define("test", "description", {}, "not a function")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid tool function, got %v", result)
				}
			},
		},
		{
			name: "execute_nonexistent_tool_error",
			script: `
				local success, err = pcall(function()
					tools.execute_safe("nonexistent_tool", {})
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for nonexistent tool, got %v", result)
				}
			},
		},
		{
			name: "pipeline_with_empty_tools_error",
			script: `
				local success, err = pcall(function()
					tools.pipeline({}, "input")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for empty pipeline, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestToolMetrics tests metrics collection and retrieval
func TestToolMetrics(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupToolsLibrary(t, L)

	script := `
		-- Define and execute a tool to generate metrics
		tools.define("metrics_tool", "Tool for metrics", {}, function(params)
			return "result: " .. params.input
		end)
		
		-- Execute the tool a few times
		tools.execute_safe("metrics_tool", {input = "test1"})
		tools.execute_safe("metrics_tool", {input = "test2"})
		
		-- Get metrics
		local metrics = tools.get_metrics("metrics_tool")
		
		return metrics ~= nil and 
		       metrics.total_calls == 2 and
		       metrics.successful_calls == 2 and
		       metrics.failed_calls == 0 and
		       metrics.avg_duration >= 0
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Tool metrics test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected tool metrics to work correctly, got %v", result)
	}
}

// TestToolHistory tests execution history tracking
func TestToolHistory(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupToolsLibrary(t, L)

	script := `
		-- Define and execute tools to generate history
		tools.define("history_tool", "Tool for history", {}, function(params)
			return "processed: " .. params.data
		end)
		
		-- Execute multiple times
		for i = 1, 3 do
			tools.execute_safe("history_tool", {data = "test" .. i})
		end
		
		-- Get history
		local history = tools.get_history(5)
		
		-- Check history
		return type(history) == "table" and 
		       #history == 3 and
		       history[1].tool == "history_tool" and
		       history[1].success == true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Tool history test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected tool history tracking to work correctly, got %v", result)
	}
}

// BenchmarkToolOperations benchmarks key tool operations
func BenchmarkToolOperations(b *testing.B) {
	benchmarks := []struct {
		name   string
		script string
	}{
		{
			name: "tool_definition",
			script: `
				tools.define("bench_tool_" .. math.random(10000), "Benchmark tool", {}, function(params)
					return params.input * 2
				end)
			`,
		},
		{
			name: "tool_execution",
			script: `
				tools.execute_safe("bench_execution_tool", {input = 42})
			`,
		},
		{
			name: "parameter_validation",
			script: `
				local tool = {
					parameters = {
						required = {"input"},
						properties = {input = {type = "number"}}
					}
				}
				tools.validate_params(tool, {input = 123})
			`,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()

			setupToolsLibrary(nil, L) // Skip t.Helper in benchmark

			// Pre-setup for execution benchmark
			if bm.name == "tool_execution" {
				setupScript := `
					tools.define("bench_execution_tool", "Execution benchmark", {}, function(params)
						return params.input * 2
					end)
				`
				err := L.DoString(setupScript)
				if err != nil {
					b.Fatalf("Setup failed: %v", err)
				}
			}

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

// TestToolsPackageRequire tests that the module can be required as a package
func TestToolsPackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Set up mock tools bridge
	setupToolsLibrary(t, L)

	script := `
		-- Test that tools is available globally
		if type(tools) ~= "table" then
			error("Tools module should be available globally")
		end
		
		-- Test basic functionality
		tools.define("require_test_tool", "Test tool", {}, function(params)
			return "success"
		end)
		
		local result = tools.execute_safe("require_test_tool", {})
		
		if result ~= "success" then
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
