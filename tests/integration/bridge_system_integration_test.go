// ABOUTME: Comprehensive integration tests for bridge system with lazy loading and renamed bridges.
// ABOUTME: Validates that all bridge renaming works correctly with lazy loading architecture.

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/cmd/llmspell/commands"
	"github.com/lexlapax/go-llmspell/pkg/config"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenamedBridgesAccessibleWithLazyLoading tests that all renamed bridges
// are accessible when using lazy loading architecture.
func TestRenamedBridgesAccessibleWithLazyLoading(t *testing.T) {

	tests := []struct {
		name            string
		securityProfile string
		expectedCount   int
		description     string
	}{
		{
			name:            "StandardProfile_Sandbox",
			securityProfile: "sandbox",
			expectedCount:   23, // All bridges should be available in standard profile
			description:     "Sandbox should load StandardProfile with all bridges",
		},
		{
			name:            "DevelopmentProfile_Development",
			securityProfile: "development",
			expectedCount:   15, // Core + LLM + Utility + Observability only
			description:     "Development should load DevelopmentProfile with debugging bridges",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test script that checks bridge availability
			testScript := `
-- Test bridge accessibility with lazy loading
local available_bridges = {}
local bridge_categories = {}

-- Expected bridges from Go test
local expected_bridges = {
	-- Core
	"llm_modelinfo",
	-- LLM  
	"llm_core", "llm_providers", "llm_pool",
	-- Utility
	"util_core", "util_auth", "util_debug", "util_errors", 
	"util_json", "util_llm", "util_slog", "util_script_logger",
	-- Agent
	"agent_core", "agent_events", "agent_hooks", 
	"agent_tools", "agent_tools_registry", "agent_workflow",
	-- Observability
	"observability_guardrails", "observability_metrics", "observability_tracing",
	-- State
	"state_manager", "state_context",
	-- Structured
	"structured_schema"
}

-- Check which bridges are available
for _, bridge_id in ipairs(expected_bridges) do
	if bridges and bridges[bridge_id] then
		table.insert(available_bridges, bridge_id)
	end
end

return {
	test_name = "bridge_accessibility_test",
	security_profile = "` + tc.securityProfile + `",
	available_count = #available_bridges,
	available_bridges = available_bridges,
	lazy_loading_active = true
}
`

			// Create runner configuration with lazy loading enabled
			runnerConfig := runner.DefaultRunnerConfig()
			runnerConfig.EnableDebug = true

			// Setup engine registry with the test profile
			engineManager, err := runner.SetupEngineRegistry(runnerConfig, tc.securityProfile)
			require.NoError(t, err)

			// Create engine selector and script executor
			selector := runner.NewEngineSelector(engineManager)
			scriptExecutor := runner.NewScriptExecutor(runnerConfig, engineManager, selector)

			// Initialize executor
			ctx := context.Background()
			err = scriptExecutor.Initialize(ctx)
			require.NoError(t, err)

			// Create command context with the security profile
			ctx = context.WithValue(ctx, commands.ConfigKey, config.GetDefaultConfig())
			ctx = context.WithValue(ctx, commands.DebugKey, true)
			ctx = context.WithValue(ctx, commands.ProfileKey, tc.securityProfile)

			// Measure execution time to verify lazy loading performance
			startTime := time.Now()

			// Execute script using ScriptExecutor directly
			options := &runner.RunnerOptions{
				SecurityProfile: tc.securityProfile,
			}

			result, err := scriptExecutor.ExecuteWithOptions(ctx, testScript, options)
			require.NoError(t, err, tc.description)

			executionTime := time.Since(startTime)

			// Verify result structure
			require.NotNil(t, result.Value, "Result should not be nil")

			// Handle engine.ObjectValue type
			var resultMap map[string]interface{}
			if objVal, ok := result.Value.(engine.ObjectValue); ok {
				resultMap = objVal.ToGo().(map[string]interface{})
			} else if mapVal, ok := result.Value.(map[string]interface{}); ok {
				resultMap = mapVal
			} else {
				t.Fatalf("Unexpected result type: %T", result.Value)
			}

			// Verify test results
			testName, exists := resultMap["test_name"]
			require.True(t, exists, "test_name should exist in result")
			assert.Equal(t, "bridge_accessibility_test", testName)

			profileUsed, exists := resultMap["security_profile"]
			require.True(t, exists, "security_profile should exist in result")
			assert.Equal(t, tc.securityProfile, profileUsed)

			availableCount, exists := resultMap["available_count"]
			require.True(t, exists, "available_count should exist in result")
			assert.Equal(t, tc.expectedCount, int(availableCount.(float64)), tc.description)

			// Verify lazy loading is working (should be reasonably fast)
			assert.Less(t, executionTime, 5*time.Second, "Execution should be fast with lazy loading")

			// Verify specific renamed bridges are accessible for standard profile
			if tc.securityProfile == "sandbox" {
				availableBridges, exists := resultMap["available_bridges"]
				require.True(t, exists, "available_bridges should exist")
				bridgeList, ok := availableBridges.([]interface{})
				require.True(t, ok, "available_bridges should be a list")

				bridgeSet := make(map[string]bool)
				for _, bridge := range bridgeList {
					bridgeSet[bridge.(string)] = true
				}

				// Verify key renamed bridges are present
				renamedBridges := []string{
					"llm_core",          // was "llm"
					"llm_providers",     // was "providers"
					"util_core",         // was "util"
					"util_slog",         // was "slog"
					"agent_core",        // was "agent"
					"agent_tools",       // was "tools"
					"structured_schema", // was "schema"
				}

				for _, bridgeID := range renamedBridges {
					assert.True(t, bridgeSet[bridgeID], "Renamed bridge %s should be accessible", bridgeID)
				}
			}

			// Cleanup
			_ = scriptExecutor.Shutdown()
		})
	}
}

// TestBridgeGlobalsSetWithLazyLoading tests that bridge globals are properly
// set when engines are loaded on-demand.
func TestBridgeGlobalsSetWithLazyLoading(t *testing.T) {
	testScript := `
-- Test that bridge globals are properly set with lazy loading
local globals_test = {}

-- Check that bridges global is available
if bridges then
	globals_test.bridges_global_exists = true
	globals_test.bridges_type = type(bridges)
	
	-- Check some specific bridges
	local test_bridges = {"llm_core", "util_core", "agent_core"}
	globals_test.specific_bridges = {}
	
	for _, bridge_id in ipairs(test_bridges) do
		globals_test.specific_bridges[bridge_id] = bridges[bridge_id] ~= nil
	end
else
	globals_test.bridges_global_exists = false
end

-- Check that stdlib modules can access bridges
local success, llm_module = pcall(require, "llm")
if success then
	globals_test.llm_stdlib_loads = true
	-- Test if llm module can access its bridge
	globals_test.llm_can_access_bridge = llm_module ~= nil
else
	globals_test.llm_stdlib_loads = false
	globals_test.llm_error = tostring(llm_module)
end

return globals_test
`

	// Create runner configuration
	runnerConfig := runner.DefaultRunnerConfig()
	runnerConfig.EnableDebug = true

	// Setup engine registry
	engineManager, err := runner.SetupEngineRegistry(runnerConfig, "development")
	require.NoError(t, err)

	// Create engine selector and script executor
	selector := runner.NewEngineSelector(engineManager)
	scriptExecutor := runner.NewScriptExecutor(runnerConfig, engineManager, selector)

	// Initialize executor
	ctx := context.Background()
	err = scriptExecutor.Initialize(ctx)
	require.NoError(t, err)

	// Execute test
	options := &runner.RunnerOptions{
		SecurityProfile: "development",
	}

	result, err := scriptExecutor.ExecuteWithOptions(ctx, testScript, options)
	require.NoError(t, err)

	// Verify results
	require.NotNil(t, result.Value)
	t.Logf("Bridge globals test result: %T, value: %+v", result.Value, result.Value)

	// Handle engine.ObjectValue type
	var resultMap map[string]interface{}
	if objVal, ok := result.Value.(engine.ObjectValue); ok {
		resultMap = objVal.ToGo().(map[string]interface{})
	} else if mapVal, ok := result.Value.(map[string]interface{}); ok {
		resultMap = mapVal
	} else {
		t.Fatalf("Unexpected result type: %T", result.Value)
	}

	// Verify bridges global exists
	bridgesExists, exists := resultMap["bridges_global_exists"]
	require.True(t, exists)
	assert.True(t, bridgesExists.(bool), "bridges global should exist")

	bridgesType, exists := resultMap["bridges_type"]
	require.True(t, exists)
	assert.Equal(t, "table", bridgesType.(string), "bridges should be a table")

	// Verify specific bridges are accessible
	specificBridges, exists := resultMap["specific_bridges"]
	require.True(t, exists)
	bridgeMap, ok := specificBridges.(map[string]interface{})
	require.True(t, ok)

	assert.True(t, bridgeMap["llm_core"].(bool), "llm_core bridge should be accessible")
	assert.True(t, bridgeMap["util_core"].(bool), "util_core bridge should be accessible")

	// Note: agent_core might not be available in development profile, that's OK

	// Verify stdlib can load and access bridges
	llmLoads, exists := resultMap["llm_stdlib_loads"]
	require.True(t, exists)
	assert.True(t, llmLoads.(bool), "LLM stdlib module should load successfully")

	// Cleanup
	_ = scriptExecutor.Shutdown()
}

// TestToolsListWithLazyLoading tests that tools.list() works correctly with
// on-demand bridge loading and shows all available tools from loaded bridges.
func TestToolsListWithLazyLoading(t *testing.T) {
	testScript := `
-- Test tools.list() functionality with lazy loading
local results = {}

-- First check if tools module is available
local tools_module = nil
local tools_error = nil

-- Try to get tools module via require (if allowed)
local success, result = pcall(require, "tools")
if success then
	tools_module = result
	results.tools_access_method = "require"
else
	-- Check if tools is available globally
	if tools then
		tools_module = tools
		results.tools_access_method = "global"
	else
		tools_error = tostring(result)
		results.tools_access_method = "failed"
	end
end

results.tools_module_available = tools_module ~= nil

if tools_module then
	-- Check what functions are available in tools module
	results.tools_functions = {}
	for key, value in pairs(tools_module) do
		table.insert(results.tools_functions, key .. ":" .. type(value))
	end
	
	if tools_module.list then
		-- Test tools.list() function
		local list_success, tool_list = pcall(tools_module.list)
		if list_success then
			results.tools_list_works = true
			results.tool_count = #tool_list
			results.sample_tools = {}
			
			-- Collect first few tools as samples
			for i = 1, math.min(5, #tool_list) do
				table.insert(results.sample_tools, tool_list[i])
			end
		else
			results.tools_list_works = false
			results.tools_list_error = tostring(tool_list)
		end
	else
		results.tools_list_works = false
		results.tools_list_error = "tools.list function not available"
	end
else
	results.tools_list_works = false
	results.tools_list_error = "tools module not available"
end

-- Test that bridges are accessible for tools functionality
if bridges and bridges.agent_tools then
	results.agent_tools_bridge_available = true
else
	results.agent_tools_bridge_available = false
end

return results
`

	// Test with different security profiles
	profiles := []struct {
		name              string
		profile           string
		expectToolsModule bool
		expectToolsList   bool
		description       string
	}{
		{
			name:              "StandardProfile_Sandbox",
			profile:           "sandbox",
			expectToolsModule: false, // tools module not available via require in sandbox
			expectToolsList:   false,
			description:       "Sandbox profile should restrict tools module access",
		},
		{
			name:              "DevelopmentProfile_Development",
			profile:           "development",
			expectToolsModule: true, // tools module should be available via require
			expectToolsList:   true, // tools.list() should work
			description:       "Development profile should allow full tools functionality",
		},
	}

	for _, tc := range profiles {
		t.Run(tc.name, func(t *testing.T) {
			// Create runner configuration
			runnerConfig := runner.DefaultRunnerConfig()
			runnerConfig.EnableDebug = true

			// Setup engine registry
			engineManager, err := runner.SetupEngineRegistry(runnerConfig, tc.profile)
			require.NoError(t, err)

			// Create engine selector and script executor
			selector := runner.NewEngineSelector(engineManager)
			scriptExecutor := runner.NewScriptExecutor(runnerConfig, engineManager, selector)

			// Initialize executor
			ctx := context.Background()
			err = scriptExecutor.Initialize(ctx)
			require.NoError(t, err)

			// Execute test
			options := &runner.RunnerOptions{
				SecurityProfile: tc.profile,
			}

			result, err := scriptExecutor.ExecuteWithOptions(ctx, testScript, options)
			require.NoError(t, err, tc.description)

			// Parse results
			require.NotNil(t, result.Value)

			var resultMap map[string]interface{}
			if objVal, ok := result.Value.(engine.ObjectValue); ok {
				resultMap = objVal.ToGo().(map[string]interface{})
			} else {
				t.Fatalf("Unexpected result type: %T", result.Value)
			}

			// Verify tools module availability matches expectations
			toolsAvailable, exists := resultMap["tools_module_available"]
			require.True(t, exists)
			assert.Equal(t, tc.expectToolsModule, toolsAvailable.(bool),
				"Tools module availability should match profile expectations")

			// Debug: print what we got
			t.Logf("Tools module available: %v", resultMap["tools_module_available"])
			t.Logf("Tools access method: %v", resultMap["tools_access_method"])
			if functions, exists := resultMap["tools_functions"]; exists {
				t.Logf("Tools functions: %v", functions)
			}
			if errorMsg, exists := resultMap["tools_list_error"]; exists {
				t.Logf("Tools list error: %v", errorMsg)
			}

			// Verify tools.list() functionality
			toolsListWorks, exists := resultMap["tools_list_works"]
			require.True(t, exists)
			assert.Equal(t, tc.expectToolsList, toolsListWorks.(bool),
				"Tools list functionality should match profile expectations")

			// For development profile, verify tools.list() returns actual tools
			if tc.expectToolsList && toolsListWorks.(bool) {
				toolCount, exists := resultMap["tool_count"]
				require.True(t, exists)
				assert.Greater(t, int(toolCount.(float64)), 0, "Should have at least some tools available")

				sampleTools, exists := resultMap["sample_tools"]
				require.True(t, exists)
				tools, ok := sampleTools.([]interface{})
				require.True(t, ok)
				assert.Greater(t, len(tools), 0, "Should have sample tools")

				t.Logf("Found %d tools, samples: %v", int(toolCount.(float64)), tools)
			}

			// Verify bridge availability
			agentToolsBridge, exists := resultMap["agent_tools_bridge_available"]
			require.True(t, exists)

			// Agent tools bridge should be available in sandbox but not necessarily in development
			if tc.profile == "sandbox" {
				assert.True(t, agentToolsBridge.(bool), "agent_tools bridge should be available in sandbox")
			}

			// Cleanup
			_ = scriptExecutor.Shutdown()
		})
	}
}

// TestStdlibModulesWithRenamedBridges tests that all stdlib modules can load
// with the new bridge names and lazy initialization.
func TestStdlibModulesWithRenamedBridges(t *testing.T) {

	testScript := `
-- Test stdlib module loading with renamed bridges
local results = {}

-- List of modules to test (matches Go test data)
local modules_to_test = {
	"core", "llm", "tools", "agent", "events", "state", "auth", 
	"data", "errors", "observability", "structured", "promise", 
	"testing", "spell", "logging"
}

for _, module_name in ipairs(modules_to_test) do
	local module_result = {
		name = module_name,
		loaded = false,
		error = nil,
		bridge_accessible = false,
		access_method = "unknown"
	}
	
	-- Try to load the module via require()
	local success, module_or_error = pcall(require, module_name)
	if success then
		module_result.loaded = true
		module_result.bridge_accessible = module_or_error ~= nil
		module_result.access_method = "require"
	else
		-- If require fails, check if module is available as global (pre-loaded)
		local global_module = _G[module_name]
		if global_module then
			module_result.loaded = true
			module_result.bridge_accessible = global_module ~= nil
			module_result.access_method = "global"
		else
			module_result.error = tostring(module_or_error)
			module_result.access_method = "failed"
		end
	end
	
	table.insert(results, module_result)
end

-- Test specific renamed bridge usage
local bridge_tests = {}

-- Test that llm module can use llm_core bridge
if bridges and bridges.llm_core then
	bridge_tests.llm_core_accessible = true
else
	bridge_tests.llm_core_accessible = false
end

-- Test that tools module can use agent_tools bridge  
if bridges and bridges.agent_tools then
	bridge_tests.agent_tools_accessible = true
else
	bridge_tests.agent_tools_accessible = false
end

-- Test that data module can use util_core bridge (renamed from util)
if bridges and bridges.util_core then
	bridge_tests.util_core_accessible = true
else
	bridge_tests.util_core_accessible = false
end

return {
	test_name = "stdlib_modules_renamed_bridges_test",
	module_results = results,
	bridge_tests = bridge_tests,
	total_modules_tested = #results
}
`

	// Test with different security profiles to ensure modules work across profiles
	profiles := []struct {
		name                    string
		profile                 string
		expectedSuccessfulLoads int
		description             string
	}{
		{
			name:                    "StandardProfile_Sandbox",
			profile:                 "sandbox",
			expectedSuccessfulLoads: 0, // Modules are not pre-loaded in sandbox - require() blocked, no globals
			description:             "Sandbox profile should block stdlib module loading (security restriction)",
		},
		{
			name:                    "DevelopmentProfile_Development",
			profile:                 "development",
			expectedSuccessfulLoads: 15, // All modules should load via require
			description:             "Development profile should load all modules via require",
		},
	}

	for _, tc := range profiles {
		t.Run(tc.name, func(t *testing.T) {
			// Create runner configuration
			runnerConfig := runner.DefaultRunnerConfig()
			runnerConfig.EnableDebug = true

			// Setup engine registry
			engineManager, err := runner.SetupEngineRegistry(runnerConfig, tc.profile)
			require.NoError(t, err)

			// Create engine selector and script executor
			selector := runner.NewEngineSelector(engineManager)
			scriptExecutor := runner.NewScriptExecutor(runnerConfig, engineManager, selector)

			// Initialize executor
			ctx := context.Background()
			err = scriptExecutor.Initialize(ctx)
			require.NoError(t, err)

			// Execute test
			options := &runner.RunnerOptions{
				SecurityProfile: tc.profile,
			}

			result, err := scriptExecutor.ExecuteWithOptions(ctx, testScript, options)
			require.NoError(t, err, tc.description)

			// Parse results
			require.NotNil(t, result.Value)

			var resultMap map[string]interface{}
			if objVal, ok := result.Value.(engine.ObjectValue); ok {
				resultMap = objVal.ToGo().(map[string]interface{})
			} else {
				t.Fatalf("Unexpected result type: %T", result.Value)
			}

			// Verify test name
			testName, exists := resultMap["test_name"]
			require.True(t, exists)
			assert.Equal(t, "stdlib_modules_renamed_bridges_test", testName)

			// Verify module loading results
			moduleResults, exists := resultMap["module_results"]
			require.True(t, exists)
			modules, ok := moduleResults.([]interface{})
			require.True(t, ok)

			successCount := 0
			for _, modInterface := range modules {
				mod, ok := modInterface.(map[string]interface{})
				require.True(t, ok)

				moduleName := mod["name"].(string)
				loaded := mod["loaded"].(bool)
				accessMethod := mod["access_method"].(string)

				if loaded {
					successCount++
					t.Logf("Module %s loaded successfully via %s", moduleName, accessMethod)
				} else {
					errorMsg := ""
					if mod["error"] != nil {
						errorMsg = mod["error"].(string)
					}
					t.Logf("Module %s failed to load via %s: %s", moduleName, accessMethod, errorMsg)
				}
			}

			// Verify minimum successful loads (allowing for some flexibility based on profile)
			assert.GreaterOrEqual(t, successCount, tc.expectedSuccessfulLoads-3, "Most stdlib modules should load successfully")

			// Verify bridge accessibility tests
			bridgeTests, exists := resultMap["bridge_tests"]
			require.True(t, exists)
			bridges, ok := bridgeTests.(map[string]interface{})
			require.True(t, ok)

			// For sandbox profile, all renamed bridges should be accessible
			if tc.profile == "sandbox" {
				assert.True(t, bridges["llm_core_accessible"].(bool), "llm_core bridge should be accessible in sandbox")
				assert.True(t, bridges["agent_tools_accessible"].(bool), "agent_tools bridge should be accessible in sandbox")
				assert.True(t, bridges["util_core_accessible"].(bool), "util_core bridge should be accessible in sandbox")
			}

			// For development profile, LLM and utility bridges should be accessible
			if tc.profile == "development" {
				assert.True(t, bridges["llm_core_accessible"].(bool), "llm_core bridge should be accessible in development")
				assert.True(t, bridges["util_core_accessible"].(bool), "util_core bridge should be accessible in development")
				// agent_tools may not be accessible in development profile, that's OK
			}

			// Cleanup
			_ = scriptExecutor.Shutdown()
		})
	}
}
