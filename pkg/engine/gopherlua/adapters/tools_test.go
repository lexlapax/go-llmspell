// ABOUTME: Tests for Tools bridge adapter that exposes go-llms tool functionality to Lua scripts
// ABOUTME: Validates tool discovery, execution, registration, validation, and metrics capabilities

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

func TestToolsAdapter_Creation(t *testing.T) {
	t.Run("create_tools_adapter", func(t *testing.T) {
		// Create tools bridge mock
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Tools Bridge",
				Version:     "2.0.0",
				Description: "Script access to go-llms tool system with v0.3.5 enhancements",
			}).
			WithMethod("listTools", engine.MethodInfo{
				Name:        "listTools",
				Description: "List all available tools",
				ReturnType:  "array",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock tools
				tools := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("calculator"),
						"description": engine.NewStringValue("Basic math operations"),
						"category":    engine.NewStringValue("math"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("weather"),
						"description": engine.NewStringValue("Get weather information"),
						"category":    engine.NewStringValue("api"),
					}),
				}
				return engine.NewArrayValue(tools), nil
			}).
			WithMethod("searchTools", engine.MethodInfo{
				Name:        "searchTools",
				Description: "Search tools by query",
				ReturnType:  "array",
			}, nil).
			WithMethod("executeTool", engine.MethodInfo{
				Name:        "executeTool",
				Description: "Execute a tool",
				ReturnType:  "object",
			}, nil).
			WithMethod("getToolInfo", engine.MethodInfo{
				Name:        "getToolInfo",
				Description: "Get tool information",
				ReturnType:  "object",
			}, nil).
			WithMethod("getToolSchema", engine.MethodInfo{
				Name:        "getToolSchema",
				Description: "Get tool schema",
				ReturnType:  "object",
			}, nil).
			WithMethod("validateToolInput", engine.MethodInfo{
				Name:        "validateToolInput",
				Description: "Validate tool input",
				ReturnType:  "object",
			}, nil).
			WithMethod("registerCustomTool", engine.MethodInfo{
				Name:        "registerCustomTool",
				Description: "Register a custom tool",
				ReturnType:  "boolean",
			}, nil).
			WithMethod("getToolMetrics", engine.MethodInfo{
				Name:        "getToolMetrics",
				Description: "Get tool metrics",
				ReturnType:  "object",
			}, nil)

		// Create adapter
		adapter := NewToolsAdapter(toolsBridge)
		require.NotNil(t, adapter)

		// Should have tools-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "listTools")
		assert.Contains(t, methods, "searchTools")
		assert.Contains(t, methods, "executeTool")
		assert.Contains(t, methods, "getToolInfo")
		assert.Contains(t, methods, "validateToolInput")
		assert.Contains(t, methods, "registerCustomTool")
	})

	t.Run("tools_module_structure", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "Tools Bridge",
			}).
			WithMethod("listTools", engine.MethodInfo{
				Name: "listTools",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			})

		adapter := NewToolsAdapter(toolsBridge)

		// Create Lua state
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module table
		module := L.Get(-1)
		L.SetGlobal("tools", module)

		// Test module structure
		err = L.DoString(`
			-- Check basic module properties
			assert(tools._adapter == "tools", "should have correct adapter name")
			assert(tools._version == "2.0.0", "should have correct version")
			
			-- Check tool categories
			assert(tools.CATEGORIES.MATH == "math", "should have math category")
			assert(tools.CATEGORIES.API == "api", "should have api category")
			assert(tools.CATEGORIES.TEXT == "text", "should have text category")
			assert(tools.CATEGORIES.FILE == "file", "should have file category")
			assert(tools.CATEGORIES.SYSTEM == "system", "should have system category")
			
			-- Check permission types
			assert(tools.PERMISSIONS.NETWORK == "network", "should have network permission")
			assert(tools.PERMISSIONS.FILE_READ == "file_read", "should have file read permission")
			assert(tools.PERMISSIONS.FILE_WRITE == "file_write", "should have file write permission")
			assert(tools.PERMISSIONS.SYSTEM == "system", "should have system permission")
			
			-- Check resource usage levels
			assert(tools.RESOURCE_USAGE.LOW == "low", "should have low resource usage")
			assert(tools.RESOURCE_USAGE.MEDIUM == "medium", "should have medium resource usage")
			assert(tools.RESOURCE_USAGE.HIGH == "high", "should have high resource usage")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_ToolDiscovery(t *testing.T) {
	t.Run("list_tools", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("listTools", engine.MethodInfo{
				Name: "listTools",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				tools := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("calculator"),
						"description": engine.NewStringValue("Basic math operations"),
						"category":    engine.NewStringValue("math"),
						"tags": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("math"),
							engine.NewStringValue("utility"),
						}),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("weather"),
						"description": engine.NewStringValue("Get weather information"),
						"category":    engine.NewStringValue("api"),
						"tags": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("api"),
							engine.NewStringValue("weather"),
						}),
					}),
				}
				return engine.NewArrayValue(tools), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- List all tools - handles array as multiple return values
			local toolList = {tools.listTools()}
			assert(#toolList == 2, "should have 2 tools")
			
			-- Check first tool
			assert(toolList[1].name == "calculator", "first tool should be calculator")
			assert(toolList[1].category == "math", "calculator should be in math category")
			
			-- Check second tool
			assert(toolList[2].name == "weather", "second tool should be weather")
			assert(toolList[2].category == "api", "weather should be in api category")
		`)
		assert.NoError(t, err)
	})

	t.Run("search_tools", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("searchTools", engine.MethodInfo{
				Name: "searchTools",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				query := args[0].(engine.StringValue).Value()

				// Mock search results based on query
				if query == "math" {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":        engine.NewStringValue("calculator"),
							"description": engine.NewStringValue("Basic math operations"),
						}),
					}), nil
				}

				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Search for math tools
			local mathTools = {tools.searchTools("math")}
			assert(#mathTools == 1, "should find 1 math tool")
			assert(mathTools[1].name == "calculator", "should find calculator")
			
			-- Search for non-existent tools
			local noTools = {tools.searchTools("nonexistent")}
			assert(#noTools == 0, "should find no tools")
		`)
		assert.NoError(t, err)
	})

	t.Run("get_tool_info", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("getToolInfo", engine.MethodInfo{
				Name: "getToolInfo",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				toolName := args[0].(engine.StringValue).Value()

				if toolName == "calculator" {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":        engine.NewStringValue("calculator"),
						"description": engine.NewStringValue("Basic math operations"),
						"category":    engine.NewStringValue("math"),
						"version":     engine.NewStringValue("1.0.0"),
						"usageHint":   engine.NewStringValue("Use for arithmetic calculations"),
						"permissions": engine.NewArrayValue([]engine.ScriptValue{}),
					}), nil
				}

				return engine.NewNilValue(), fmt.Errorf("tool not found: %s", toolName)
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Get existing tool info
			local info, err = tools.getToolInfo("calculator")
			assert(err == nil, "should not error")
			assert(info.name == "calculator", "should have tool name")
			assert(info.version == "1.0.0", "should have version")
			assert(info.usageHint == "Use for arithmetic calculations", "should have usage hint")
			
			-- Get non-existent tool info
			local nilInfo, getErr = tools.getToolInfo("nonexistent")
			assert(getErr ~= nil, "should error for non-existent tool")
			assert(string.find(getErr, "not found"), "error should mention not found")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_ToolExecution(t *testing.T) {
	t.Run("execute_tool", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("executeTool", engine.MethodInfo{
				Name: "executeTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				toolName := args[0].(engine.StringValue).Value()
				params := args[1].(engine.ObjectValue).Fields()

				if toolName == "calculator" {
					// Extract operation and numbers
					op := params["operation"].(engine.StringValue).Value()
					a := params["a"].(engine.NumberValue).Value()
					b := params["b"].(engine.NumberValue).Value()

					var result float64
					switch op {
					case "add":
						result = a + b
					case "multiply":
						result = a * b
					default:
						return engine.NewNilValue(), fmt.Errorf("unknown operation: %s", op)
					}

					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"result": engine.NewNumberValue(result),
						"status": engine.NewStringValue("success"),
					}), nil
				}

				return engine.NewNilValue(), fmt.Errorf("tool not found: %s", toolName)
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Execute calculator tool
			local result, err = tools.executeTool("calculator", {
				operation = "add",
				a = 5,
				b = 3
			})
			assert(err == nil, "should not error")
			assert(result.result == 8, "5 + 3 should equal 8")
			assert(result.status == "success", "should have success status")
			
			-- Execute with multiply
			local multResult, multErr = tools.executeTool("calculator", {
				operation = "multiply",
				a = 4,
				b = 7
			})
			assert(multErr == nil, "multiply should not error")
			assert(multResult.result == 28, "4 * 7 should equal 28")
			
			-- Execute non-existent tool
			local nilResult, execErr = tools.executeTool("nonexistent", {})
			assert(execErr ~= nil, "should error for non-existent tool")
		`)
		assert.NoError(t, err)
	})

	t.Run("execute_tool_async", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("executeToolAsync", engine.MethodInfo{
				Name: "executeToolAsync",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return a promise-like object
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"taskId": engine.NewStringValue("task-123"),
					"status": engine.NewStringValue("pending"),
				}), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Execute tool asynchronously
			local task, err = tools.executeAsync("calculator", {
				operation = "add",
				a = 100,
				b = 200
			})
			assert(err == nil, "should not error")
			assert(task.taskId == "task-123", "should have task ID")
			assert(task.status == "pending", "should be pending")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_CustomTools(t *testing.T) {
	t.Run("register_custom_tool", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("registerCustomTool", engine.MethodInfo{
				Name: "registerCustomTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				toolDef := args[0].(engine.ObjectValue).Fields()

				// Validate required fields
				if _, ok := toolDef["name"]; !ok {
					return engine.NewBoolValue(false), fmt.Errorf("tool name is required")
				}
				if _, ok := toolDef["execute"]; !ok {
					return engine.NewBoolValue(false), fmt.Errorf("execute function is required")
				}

				return engine.NewBoolValue(true), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Register a custom tool
			local success, err = tools.registerCustomTool({
				name = "greeter",
				description = "Greets a person",
				category = "text",
				tags = {"utility", "text"},
				execute = function(params)
					return {
						message = "Hello, " .. params.name .. "!",
						timestamp = os.time()
					}
				end,
				parameterSchema = {
					type = "object",
					properties = {
						name = {type = "string", description = "Name to greet"}
					},
					required = {"name"}
				}
			})
			
			assert(err == nil, "should not error")
			assert(success == true, "should register successfully")
			
			-- Try to register invalid tool (missing execute)
			local failSuccess, failErr = tools.registerCustomTool({
				name = "invalid",
				description = "Invalid tool"
			})
			assert(failErr ~= nil, "should error for invalid tool")
			assert(string.find(failErr, "execute"), "error should mention execute")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_Validation(t *testing.T) {
	t.Run("validate_tool_input", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("getToolSchema", engine.MethodInfo{
				Name: "getToolSchema",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"parameters": engine.NewObjectValue(map[string]engine.ScriptValue{
						"type": engine.NewStringValue("object"),
						"properties": engine.NewObjectValue(map[string]engine.ScriptValue{
							"a": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("number"),
							}),
							"b": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("number"),
							}),
						}),
						"required": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("a"),
							engine.NewStringValue("b"),
						}),
					}),
				}), nil
			}).
			WithMethod("validateToolInput", engine.MethodInfo{
				Name: "validateToolInput",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				params := args[1].(engine.ObjectValue).Fields()

				// Simple validation
				if _, ok := params["a"]; !ok {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"valid": engine.NewBoolValue(false),
						"errors": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("missing required field: a"),
						}),
					}), nil
				}

				return engine.NewObjectValue(map[string]engine.ScriptValue{
					"valid":  engine.NewBoolValue(true),
					"errors": engine.NewArrayValue([]engine.ScriptValue{}),
				}), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Get tool schema
			local schema, schemaErr = tools.getToolSchema("calculator")
			assert(schemaErr == nil, "should not error getting schema")
			assert(schema.parameters.type == "object", "should have object parameters")
			
			-- Validate valid input
			local validResult, validErr = tools.validateToolInput("calculator", {
				a = 5,
				b = 3
			})
			assert(validErr == nil, "should not error")
			assert(validResult.valid == true, "should be valid")
			assert(#validResult.errors == 0, "should have no errors")
			
			-- Validate invalid input
			local invalidResult, invalidErr = tools.validateToolInput("calculator", {
				b = 3  -- missing 'a'
			})
			assert(invalidErr == nil, "validation should not error")
			assert(invalidResult.valid == false, "should be invalid")
			assert(#invalidResult.errors > 0, "should have errors")
			assert(string.find(invalidResult.errors[1], "missing"), "error should mention missing field")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_Metrics(t *testing.T) {
	t.Run("get_tool_metrics", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("getToolMetrics", engine.MethodInfo{
				Name: "getToolMetrics",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				toolName := args[0].(engine.StringValue).Value()

				if toolName == "calculator" {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"totalExecutions": engine.NewNumberValue(100),
						"successCount":    engine.NewNumberValue(95),
						"failureCount":    engine.NewNumberValue(5),
						"averageDuration": engine.NewNumberValue(12.5),
						"lastExecution":   engine.NewStringValue("2024-01-15T10:30:00Z"),
					}), nil
				}

				return engine.NewNilValue(), fmt.Errorf("no metrics for tool: %s", toolName)
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Get metrics for existing tool
			local metrics, err = tools.getToolMetrics("calculator")
			assert(err == nil, "should not error")
			assert(metrics.totalExecutions == 100, "should have execution count")
			assert(metrics.successCount == 95, "should have success count")
			assert(metrics.failureCount == 5, "should have failure count")
			assert(metrics.averageDuration == 12.5, "should have average duration")
			
			-- Get metrics for non-existent tool
			local nilMetrics, metricsErr = tools.getToolMetrics("nonexistent")
			assert(metricsErr ~= nil, "should error for non-existent tool")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_Categories(t *testing.T) {
	t.Run("list_by_category", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("listToolsByCategory", engine.MethodInfo{
				Name: "listToolsByCategory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				category := args[0].(engine.StringValue).Value()

				if category == "math" {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":     engine.NewStringValue("calculator"),
							"category": engine.NewStringValue("math"),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":     engine.NewStringValue("statistics"),
							"category": engine.NewStringValue("math"),
						}),
					}), nil
				}

				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			}).
			WithMethod("getToolCategories", engine.MethodInfo{
				Name: "getToolCategories",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{
					engine.NewStringValue("math"),
					engine.NewStringValue("api"),
					engine.NewStringValue("text"),
					engine.NewStringValue("file"),
				}), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Get all categories
			local categories = {tools.getCategories()}
			assert(#categories == 4, "should have 4 categories")
			assert(categories[1] == "math", "should have math category")
			
			-- List tools by category
			local mathTools = {tools.listByCategory("math")}
			assert(#mathTools == 2, "should have 2 math tools")
			assert(mathTools[1].name == "calculator", "should have calculator")
			assert(mathTools[2].name == "statistics", "should have statistics")
			
			-- List tools for empty category
			local emptyTools = {tools.listByCategory("nonexistent")}
			assert(#emptyTools == 0, "should have no tools")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("executeTool", engine.MethodInfo{
				Name: "executeTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), fmt.Errorf("bridge error: connection failed")
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Try to execute tool with bridge error
			local result, err = tools.executeTool("calculator", {a = 1, b = 2})
			assert(err ~= nil, "should have error")
			assert(string.find(err, "bridge error"), "error should contain bridge error message")
			assert(result == nil, "result should be nil on error")
		`)
		assert.NoError(t, err)
	})

	t.Run("handle_invalid_tool_definition", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("registerCustomTool", engine.MethodInfo{
				Name: "registerCustomTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(false), fmt.Errorf("invalid tool definition: missing required fields")
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Try to register invalid tool
			local success, err = tools.registerCustomTool({
				-- Missing required fields
				description = "Invalid tool"
			})
			assert(err ~= nil, "should have error")
			assert(string.find(err, "invalid tool definition"), "error should mention invalid definition")
			assert(success == false, "should not succeed")
		`)
		assert.NoError(t, err)
	})
}

func TestToolsAdapter_ConvenienceMethods(t *testing.T) {
	t.Run("tool_builder_pattern", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("registerCustomTool", engine.MethodInfo{
				Name: "registerCustomTool",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- Use builder pattern to create tool
			local builder = tools.createBuilder("myTool")
				:withDescription("My custom tool")
				:withCategory("utility")
				:withTags({"custom", "test"})
				:withParameter("input", "string", "Input value", true)
				:withParameter("count", "number", "Repeat count", false)
				:withExecute(function(params)
					local result = ""
					local count = params.count or 1
					for i = 1, count do
						result = result .. params.input
					end
					return {output = result}
				end)
			
			-- Build and register the tool
			local success, err = builder:build()
			assert(err == nil, "should not error")
			assert(success == true, "should register successfully")
		`)
		assert.NoError(t, err)
	})

	t.Run("batch_operations", func(t *testing.T) {
		toolsBridge := testutils.NewMockBridge("tools").
			WithInitialized(true).
			WithMethod("listToolsByTags", engine.MethodInfo{
				Name: "listToolsByTags",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewArrayValue([]engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name": engine.NewStringValue("tool1"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"name": engine.NewStringValue("tool2"),
					}),
				}), nil
			})

		adapter := NewToolsAdapter(toolsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "tools")
		require.NoError(t, err)

		err = ms.LoadModule(L, "tools")
		require.NoError(t, err)

		err = L.DoString(`
			local tools = require("tools")
			
			-- List tools by multiple tags
			local taggedTools = {tools.listByTags({"api", "weather"})}
			assert(#taggedTools == 2, "should find 2 tools")
			assert(taggedTools[1].name == "tool1", "should have tool1")
			assert(taggedTools[2].name == "tool2", "should have tool2")
		`)
		assert.NoError(t, err)
	})
}
