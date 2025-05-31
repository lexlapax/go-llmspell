// ABOUTME: Integration test for tools system with Lua engine
// ABOUTME: Tests the complete tool workflow including registration and execution

package tools_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua/bridges"
	"github.com/lexlapax/go-llmspell/pkg/tools"
)

func TestToolsLuaIntegration(t *testing.T) {
	// Create Lua engine
	eng, err := lua.NewLuaEngine(&engine.Config{
		MaxExecutionTime: 5,
		MaxMemory:        10 * 1024 * 1024,
	})
	if err != nil {
		t.Fatalf("Failed to create Lua engine: %v", err)
	}
	defer eng.Close()

	// Register standard library first
	err = eng.RegisterBridge("stdlib", nil)
	if err != nil {
		t.Fatalf("Failed to register stdlib: %v", err)
	}

	// Create tools infrastructure
	toolRegistry := tools.NewRegistry()
	toolBridge := bridge.NewToolBridge(toolRegistry)
	toolsWrapper := bridges.NewToolsBridgeWrapper(toolBridge)

	// Register tools bridge
	err = eng.RegisterBridge("tools", toolsWrapper)
	if err != nil {
		t.Fatalf("Failed to register tools bridge: %v", err)
	}

	// Load a dummy script to trigger VM initialization
	err = eng.LoadScript(strings.NewReader("-- init"))
	if err != nil {
		t.Fatalf("Failed to load init script: %v", err)
	}

	// Test basic tool functionality
	t.Run("register and execute tool", func(t *testing.T) {
		script := `
		-- Verify tools module exists
		if not tools then
			error("tools module not found")
		end
		
		-- Test that all functions exist
		if not tools.register then error("tools.register not found") end
		if not tools.execute then error("tools.execute not found") end
		if not tools.list then error("tools.list not found") end
		if not tools.get then error("tools.get not found") end
		if not tools.remove then error("tools.remove not found") end
		if not tools.validate then error("tools.validate not found") end
		
		-- Register a calculator tool
		local success, err = tools.register(
			"calculator",
			"Basic arithmetic operations",
			{
				type = "object",
				properties = {
					op = { type = "string", enum = {"add", "sub", "mul", "div"} },
					a = { type = "number" },
					b = { type = "number" }
				},
				required = {"op", "a", "b"}
			},
			function(params)
				if params.op == "add" then
					return params.a + params.b
				elseif params.op == "sub" then
					return params.a - params.b
				elseif params.op == "mul" then
					return params.a * params.b
				elseif params.op == "div" then
					if params.b == 0 then
						return nil, "division by zero"
					end
					return params.a / params.b
				else
					return nil, "unknown operation: " .. params.op
				end
			end
		)
		
		if not success then
			error("Failed to register tool: " .. tostring(err))
		end
		
		-- Test addition
		local result, err = tools.execute("calculator", {op = "add", a = 10, b = 5})
		if err then
			error("Failed to execute add: " .. tostring(err))
		end
		if result ~= 15 then
			error("Expected 15, got " .. tostring(result))
		end
		
		-- Test division by zero
		result, err = tools.execute("calculator", {op = "div", a = 10, b = 0})
		if not err then
			error("Expected error for division by zero")
		end
		
		print("Tool registration and execution tests passed!")
		`

		err := eng.LoadScript(strings.NewReader(script))
		if err != nil {
			t.Fatalf("Failed to load script: %v", err)
		}

		ctx := context.Background()
		err = eng.Execute(ctx)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
	})

	// Test tool management
	t.Run("tool management", func(t *testing.T) {
		script := `
		-- List tools
		local toolList = tools.list()
		local foundCalc = false
		for _, tool in ipairs(toolList) do
			if tool.name == "calculator" then
				foundCalc = true
			end
		end
		if not foundCalc then
			error("Calculator tool not found in list")
		end
		
		-- Get tool info
		local info, err = tools.get("calculator")
		if err then
			error("Failed to get tool info: " .. tostring(err))
		end
		if info.name ~= "calculator" then
			error("Tool name mismatch")
		end
		
		-- Validate parameters
		local valid, err = tools.validate("calculator", {op = "add", a = 1, b = 2})
		if not valid then
			error("Valid parameters failed validation: " .. tostring(err))
		end
		
		-- Invalid parameters
		valid, err = tools.validate("calculator", {op = "add", a = 1})
		if valid then
			error("Invalid parameters passed validation")
		end
		
		-- Remove tool
		local removed, err = tools.remove("calculator")
		if not removed then
			error("Failed to remove tool: " .. tostring(err))
		end
		
		-- Verify removal
		info, err = tools.get("calculator")
		if not err then
			error("Expected error when getting removed tool")
		end
		
		print("Tool management tests passed!")
		`

		err := eng.LoadScript(strings.NewReader(script))
		if err != nil {
			t.Fatalf("Failed to load script: %v", err)
		}

		ctx := context.Background()
		err = eng.Execute(ctx)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}
	})
}
