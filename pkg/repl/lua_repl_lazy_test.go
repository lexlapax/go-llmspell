// ABOUTME: Tests for Lua REPL integration with lazy bridge loading
// ABOUTME: Verifies that bridges are available in REPL with lazy loading enabled

package repl

import (
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// TestLuaREPLWithLazyBridgeLoading tests that the REPL works with lazy bridge loading
func TestLuaREPLWithLazyBridgeLoading(t *testing.T) {
	// Create engine registry
	engineRegistry := engine.NewRegistry(engine.RegistryConfig{})
	
	// Create engine manager with lazy loading
	config := runner.DefaultRunnerConfig()
	engineManager := runner.NewEngineRegistryManager(engineRegistry, config)
	if err := engineManager.Initialize(); err != nil {
		t.Fatalf("Failed to initialize engine manager: %v", err)
	}
	defer func() {
		if err := engineManager.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown engine manager: %v", err)
		}
	}()

	// Register Lua engine factory
	factory := gopherlua.NewLuaEngineFactory()
	if err := engineRegistry.Register(factory); err != nil {
		t.Fatalf("Failed to register Lua engine factory: %v", err)
	}

	// Create test streams to avoid readline race conditions
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	// Create REPL config with engine registry
	replConfig := REPLConfig{
		Engine:         "lua",
		EngineRegistry: engineManager,
		SaveHistory:    false,
		Input:          input,
		Output:         output,
		Error:          output,
	}

	// Create Lua REPL
	repl, err := NewLuaREPL(replConfig)
	if err != nil {
		t.Fatalf("Failed to create Lua REPL: %v", err)
	}
	defer repl.Close()

	// Test 1: Verify basic Lua evaluation works
	result, err := repl.Evaluate(context.Background(), "return 2 + 2")
	if err != nil {
		t.Errorf("Failed to evaluate basic expression: %v", err)
	}
	if result != "4" {
		t.Errorf("Expected '4', got '%s'", result)
	}

	// Test 2: Verify bridges are available (initially they won't be due to lazy loading)
	result, err = repl.Evaluate(context.Background(), "return type(bridges)")
	if err != nil {
		t.Errorf("Failed to check bridges type: %v", err)
	}
	// Due to lazy loading, bridges won't be available unless explicitly loaded
	if result != "nil" && result != "table" {
		t.Errorf("Expected 'nil' or 'table' for bridges type, got '%s'", result)
	}

	// Test 3: Force engine to load bridges and make them available to REPL
	// Get the engine - with lazy loading, bridges are loaded when engine is requested
	eng, err := engineManager.GetEngine("lua", engine.EngineConfig{}, "development")
	if err != nil {
		t.Fatalf("Failed to get engine: %v", err)
	}

	// Load bridges into REPL state
	if luaEngine, ok := eng.(*gopherlua.LuaEngine); ok {
		if err := luaEngine.LoadBridgeModulesIntoState(repl.luaState); err != nil {
			t.Errorf("Failed to load bridges into REPL: %v", err)
		}
	}

	// Test 4: Verify bridges are now available
	result, err = repl.Evaluate(context.Background(), "return type(bridges)")
	if err != nil {
		t.Errorf("Failed to check bridges type after loading: %v", err)
	}
	if result != "table" {
		t.Errorf("Expected 'table' for bridges type after loading, got '%s'", result)
	}

	// Test 5: Verify specific bridge is accessible
	result, err = repl.Evaluate(context.Background(), "return type(bridges.util_core)")
	if err != nil {
		t.Errorf("Failed to check util bridge type: %v", err)
	}
	if result != "table" {
		t.Errorf("Expected 'table' for util bridge type, got '%s'", result)
	}
}

// TestLuaREPLWithoutEngineRegistry tests backward compatibility without engine registry
func TestLuaREPLWithoutEngineRegistry(t *testing.T) {
	// Create test streams to avoid readline race conditions
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	// Create REPL config without engine registry
	replConfig := REPLConfig{
		Engine:      "lua",
		SaveHistory: false,
		Input:       input,
		Output:      output,
		Error:       output,
	}

	// Create Lua REPL
	repl, err := NewLuaREPL(replConfig)
	if err != nil {
		t.Fatalf("Failed to create Lua REPL: %v", err)
	}
	defer repl.Close()

	// Verify basic evaluation works
	result, err := repl.Evaluate(context.Background(), "return 'Hello, World!'")
	if err != nil {
		t.Errorf("Failed to evaluate expression: %v", err)
	}
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}

	// Verify bridges global table exists but individual bridge globals don't
	// (The LoadBridgeModulesIntoState call in NewLuaREPL sets up an empty bridges table)
	result, err = repl.Evaluate(context.Background(), "return type(bridges)")
	if err != nil {
		t.Errorf("Failed to check bridges type: %v", err)
	}
	// The bridges table might be created but should be empty without registry
	if result == "table" {
		// Check if it's empty
		result, err = repl.Evaluate(context.Background(), `
			local count = 0
			for k, v in pairs(bridges) do
				count = count + 1
			end
			return count
		`)
		if err != nil {
			t.Errorf("Failed to count bridges: %v", err)
		}
		if result != "0" {
			t.Errorf("Expected empty bridges table, but found %s entries", result)
		}
	} else if result != "nil" {
		t.Errorf("Expected 'nil' or empty 'table' for bridges type without registry, got '%s'", result)
	}
}

// TestLuaREPLBridgeUsage tests actual bridge usage in REPL
func TestLuaREPLBridgeUsage(t *testing.T) {
	// Create engine registry
	engineRegistry := engine.NewRegistry(engine.RegistryConfig{})
	
	// Create engine manager
	config := runner.DefaultRunnerConfig()
	engineManager := runner.NewEngineRegistryManager(engineRegistry, config)
	if err := engineManager.Initialize(); err != nil {
		t.Fatalf("Failed to initialize engine manager: %v", err)
	}
	defer func() {
		if err := engineManager.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown engine manager: %v", err)
		}
	}()

	// Register Lua engine factory
	factory := gopherlua.NewLuaEngineFactory()
	if err := engineRegistry.Register(factory); err != nil {
		t.Fatalf("Failed to register Lua engine factory: %v", err)
	}

	// Get engine with bridges
	eng, err := engineManager.GetEngine("lua", engine.EngineConfig{}, "development")
	if err != nil {
		t.Fatalf("Failed to get engine: %v", err)
	}

	// Bridges are already loaded when engine is requested with "development" profile

	// Create test streams to avoid readline race conditions
	input := strings.NewReader("")
	output := &strings.Builder{}
	
	// Create REPL
	replConfig := REPLConfig{
		Engine:         "lua",
		EngineRegistry: engineManager,
		SaveHistory:    false,
		Input:          input,
		Output:         output,
		Error:          output,
	}

	repl, err := NewLuaREPL(replConfig)
	if err != nil {
		t.Fatalf("Failed to create Lua REPL: %v", err)
	}
	defer repl.Close()

	// Load bridges into REPL state
	if luaEngine, ok := eng.(*gopherlua.LuaEngine); ok {
		if err := luaEngine.LoadBridgeModulesIntoState(repl.luaState); err != nil {
			t.Errorf("Failed to load bridges into REPL: %v", err)
		}
	}

	// Debug: Check what's available in bridges.util_core
	result, err := repl.Evaluate(context.Background(), `
		if bridges and bridges.util_core then
			local methods = {}
			for k, v in pairs(bridges.util_core) do
				table.insert(methods, k .. "=" .. type(v))
			end
			if #methods > 0 then
				return "bridges.util_core has: " .. table.concat(methods, ", ")
			else
				return "bridges.util_core is empty table"
			end
		else
			return "bridges.util_core not found"
		end
	`)
	if err != nil {
		t.Errorf("Failed to check bridges.util_core: %v", err)
	}
	t.Logf("Debug: bridges.util_core check result: %s", result)

	// Test util bridge usage - use generateUUID which actually exists
	result, err = repl.Evaluate(context.Background(), "return bridges.util_core.generateUUID()")
	if err != nil {
		t.Errorf("Failed to use util bridge: %v", err)
	}
	// Check if it's a valid UUID format
	if len(result) != 36 || result[8] != '-' || result[13] != '-' || result[18] != '-' || result[23] != '-' {
		t.Errorf("Expected UUID format, got '%s'", result)
	}

	// Test multi-line operations with actual methods
	script := `
local util = bridges.util_core
local str = "This is a very long string that should be truncated"
local truncated = util.truncateString(str, 10)
return truncated
`
	result, err = repl.Evaluate(context.Background(), strings.TrimSpace(script))
	if err != nil {
		t.Errorf("Failed to execute multi-line script: %v", err)
	}
	if result != "This is..." {
		t.Errorf("Expected 'This is...', got '%s'", result)
	}
}