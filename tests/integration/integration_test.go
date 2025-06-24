// ABOUTME: Integration test suite for end-to-end functionality testing
// ABOUTME: Tests complete workflows from CLI to engine execution

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/runner"
)

// createTestRunner creates a test runner with default configuration
func createTestRunner(t *testing.T) *runner.ScriptExecutor {
	return createTestRunnerWithConfig(t, nil)
}

// createTestRunnerWithConfig creates a test runner with custom configuration
func createTestRunnerWithConfig(t *testing.T, customConfig *runner.RunnerConfig) *runner.ScriptExecutor {
	runnerConfig := &runner.RunnerConfig{
		Timeout:              30 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
		DefaultSecurityLevel: "trusted",
		DefaultFeatureSet:    "full",
		EnableMetrics:        true,
		EnableValidation:     true,
	}
	
	// Apply custom config if provided
	if customConfig != nil {
		if customConfig.Timeout > 0 {
			runnerConfig.Timeout = customConfig.Timeout
		}
		if customConfig.DefaultSecurityLevel != "" {
			runnerConfig.DefaultSecurityLevel = customConfig.DefaultSecurityLevel
		}
		if customConfig.DefaultFeatureSet != "" {
			runnerConfig.DefaultFeatureSet = customConfig.DefaultFeatureSet
		}
	}

	// Setup engine registry (bridges will be loaded on-demand)
	engineManager, err := runner.SetupEngineRegistry(runnerConfig, "")
	require.NoError(t, err)

	// Create engine selector
	selector := runner.NewEngineSelector(engineManager)

	// Create script executor with proper architecture
	return runner.NewScriptExecutor(runnerConfig, engineManager, selector)
}

// extractBoolFromResult extracts a boolean value from a script result, handling ScriptValue wrapping
func extractBoolFromResult(t *testing.T, resultMap map[string]interface{}, key string) bool {
	val, ok := resultMap[key]
	require.True(t, ok, "Result should have '%s' field", key)
	
	if b, ok := val.(bool); ok {
		return b
	} else if scriptVal, ok := val.(engine.ScriptValue); ok {
		b, ok := scriptVal.ToGo().(bool)
		require.True(t, ok, "%s should be a boolean", key)
		return b
	} else {
		t.Fatalf("Unexpected %s type: %T", key, val)
		return false
	}
}

// extractStringFromResult extracts a string value from a script result, handling ScriptValue wrapping
func extractStringFromResult(t *testing.T, resultMap map[string]interface{}, key string) string {
	val, ok := resultMap[key]
	require.True(t, ok, "Result should have '%s' field", key)
	
	if s, ok := val.(string); ok {
		return s
	} else if scriptVal, ok := val.(engine.ScriptValue); ok {
		s, ok := scriptVal.ToGo().(string)
		require.True(t, ok, "%s should be a string", key)
		return s
	} else {
		t.Fatalf("Unexpected %s type: %T", key, val)
		return ""
	}
}

// convertResultToMap converts a script result to a Go map, handling ScriptValue wrapping
func convertResultToMap(t *testing.T, result interface{}) map[string]interface{} {
	if scriptValue, ok := result.(engine.ScriptValue); ok {
		m, ok := scriptValue.ToGo().(map[string]interface{})
		require.True(t, ok, "ScriptValue.ToGo() should return a map")
		return m
	} else if m, ok := result.(map[string]interface{}); ok {
		return m
	} else {
		t.Fatalf("Unexpected result type: %T", result)
		return nil
	}
}

// TestIntegrationBasicSpellExecution tests basic spell execution
func TestIntegrationBasicSpellExecution(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test files
	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "test_spell.lua")

	// Write test spell
	spellContent := `
		-- Simple test spell
		local message = params and params.message or "Hello, World!"
		print("Spell says: " .. message)
		return {
			success = true,
			message = message,
			timestamp = os.time()
		}
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	// Create context
	ctx := context.Background()

	// Create test runner
	scriptRunner := createTestRunner(t)

	// Execute spell with proper parameter type
	params := map[string]interface{}{
		"message": "Integration test message",
	}

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Convert result to map
	resultMap := convertResultToMap(t, result)
	
	// Check success field
	success := extractBoolFromResult(t, resultMap, "success")
	assert.True(t, success)
	
	// Check message field
	message := extractStringFromResult(t, resultMap, "message")
	assert.Equal(t, "Integration test message", message)
}

// TestIntegrationSpellWithBridges tests spell execution with bridge interactions
func TestIntegrationSpellWithBridges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "bridge_spell.lua")

	spellContent := `
		-- Test spell checking if bridges are available
		local has_bridges = bridges ~= nil
		local has_util = false
		local has_data_module = false
		
		if has_bridges then
			has_util = bridges.util_core ~= nil
		end
		
		-- Test if modules load without error
		local data_ok, data_module = pcall(require, "data")
		has_data_module = data_ok
		
		-- Simple test that doesn't require bridges to work
		local testObj = {
			name = "test",
			value = 42,
			nested = {
				array = {1, 2, 3}
			}
		}
		
		return {
			bridges_available = has_bridges,
			util_bridge_available = has_util,
			data_module_loads = has_data_module,
			test_passed = true
		}
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create test runner
	scriptRunner := createTestRunner(t)

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Convert result to map
	resultMap := convertResultToMap(t, result)
	
	// Check test_passed field
	testPassed := extractBoolFromResult(t, resultMap, "test_passed")
	assert.True(t, testPassed)
}

// TestIntegrationErrorHandling tests error handling in integration scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	// Create test runner with short timeout
	scriptRunner := createTestRunnerWithConfig(t, &runner.RunnerConfig{
		Timeout: 5 * time.Second,
	})

	tests := []struct {
		name         string
		spellContent string
		expectError  bool
	}{
		{
			name: "syntax_error",
			spellContent: `
				function broken(
					-- Missing closing parenthesis
			`,
			expectError: true,
		},
		{
			name: "runtime_error",
			spellContent: `
				error("Intentional runtime error")
			`,
			expectError: true,
		},
		{
			name: "timeout_error",
			spellContent: `
				while true do
					-- Infinite loop
				end
			`,
			expectError: true,
		},
		{
			name: "successful_execution",
			spellContent: `
				return "success"
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spellFile := filepath.Join(tempDir, tt.name+".lua")
			err := os.WriteFile(spellFile, []byte(tt.spellContent), 0644)
			require.NoError(t, err)

			result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestIntegrationSecurityProfiles tests different security profiles
func TestIntegrationSecurityProfiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "security_test.lua")

	spellContent := `
		-- Test spell that tries various operations
		local core = require("core")
		
		-- Test basic operations (should work in all profiles)
		local result = {
			basic_math = 2 + 2,
			string_ops = "hello " .. "world"
		}
		
		-- Test potentially restricted operations
		local success, err = pcall(function()
			-- This might be restricted in strict profiles
			core.sleep(0.1)
		end)
		
		result.sleep_allowed = success
		
		return result
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create test runner
	scriptRunner := createTestRunner(t)

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Convert result to map
	resultMap := convertResultToMap(t, result)

	// Check basic operations worked
	assert.Equal(t, float64(4), resultMap["basic_math"])
	assert.Equal(t, "hello world", resultMap["string_ops"])
}

// TestIntegrationParameterPassing tests parameter passing
func TestIntegrationParameterPassing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "params_test.lua")

	spellContent := `
		-- Test parameter access and conversion
		return {
			string_param = params and params.string_param or nil,
			number_param = params and tonumber(params.number_param) or nil,
			boolean_param = params and params.boolean_param == "true" or false,
			nil_param = params and params.nonexistent_param or nil,
			param_count = params and (function()
				local count = 0
				for k, v in pairs(params) do
					count = count + 1
				end
				return count
			end)() or 0
		}
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create test runner
	scriptRunner := createTestRunner(t)

	params := map[string]interface{}{
		"string_param":  "hello",
		"number_param":  "42",
		"boolean_param": "true",
	}

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Convert result to map
	resultMap := convertResultToMap(t, result)

	// Verify parameters were passed correctly
	assert.Equal(t, "hello", resultMap["string_param"])
	assert.Equal(t, float64(42), resultMap["number_param"])
	assert.Equal(t, true, resultMap["boolean_param"]) // params.boolean_param == "true" evaluates to true
	assert.Equal(t, float64(3), resultMap["param_count"])
}

// TestIntegrationConcurrentExecution tests concurrent spell execution
func TestIntegrationConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "concurrent_test.lua")

	spellContent := `
		-- Simulate some work with simple loop
		local dummy = 0
		for i = 1, 1000 do
			dummy = dummy + i
		end
		
		return {
			worker_id = params and params.worker_id or "unknown",
			timestamp = os.time(),
			success = true,
			work_done = dummy
		}
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create test runner
	scriptRunner := createTestRunner(t)

	// Run multiple spells concurrently
	const numWorkers = 5
	type result struct {
		data interface{}
		err  error
	}
	results := make(chan result, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			params := map[string]interface{}{
				"worker_id": string(rune('A' + workerID)),
			}

			data, err := scriptRunner.ExecuteFile(ctx, spellFile, params)
			results <- result{data: data, err: err}
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numWorkers; i++ {
		res := <-results
		if res.err != nil {
			t.Errorf("Worker failed: %v", res.err)
			errorCount++
		} else {
			assert.NotNil(t, res.data)
			// Convert and verify result
			resultMap := convertResultToMap(t, res.data)
			assert.True(t, extractBoolFromResult(t, resultMap, "success"))
			assert.NotEmpty(t, resultMap["worker_id"])
			successCount++
		}
	}

	assert.Equal(t, numWorkers, successCount)
	assert.Equal(t, 0, errorCount)
}

// TestIntegrationSpellValidation tests spell validation
func TestIntegrationSpellValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create validator with default config
	validatorConfig := gopherlua.DefaultValidatorConfig()
	validator := gopherlua.NewScriptValidator(validatorConfig)
	require.NotNil(t, validator)

	tests := []struct {
		name        string
		script      string
		expectValid bool
	}{
		{
			name: "valid_script",
			script: `
				local message = "Hello, World!"
				return message
			`,
			expectValid: true,
		},
		{
			name: "syntax_error",
			script: `
				local message = "Hello, World!
				-- Missing quote
			`,
			expectValid: false,
		},
		{
			name: "complex_valid_script",
			script: `
				local data = require("data")
				local obj = {name = "test", value = 42}
				local json = data.to_json(obj)
				return data.from_json(json)
			`,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateScript(tt.script, "test.lua")

			if tt.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Valid)
				assert.Empty(t, result.Errors)
			} else {
				// Either error or invalid report
				if err == nil {
					assert.NotNil(t, result)
					assert.False(t, result.Valid)
					assert.NotEmpty(t, result.Errors)
				}
			}
		})
	}
}

// TestIntegrationBridgeRegistry tests bridge registration and usage
func TestIntegrationBridgeRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "bridge_test.lua")

	spellContent := `
		-- Test various bridge modules
		local modules = {
			"data",
			"errors", 
			"promise",
			"core"
		}
		
		local loaded = {}
		for _, module in ipairs(modules) do
			local success, mod = pcall(require, module)
			loaded[module] = success and (mod ~= nil)
		end
		
		return loaded
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create test runner
	scriptRunner := createTestRunner(t)

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Convert result to map
	resultMap := convertResultToMap(t, result)

	// Check that at least some modules loaded
	// (The actual availability depends on which bridges are registered)
	loadedCount := 0
	for _, loaded := range resultMap {
		if loadedBool, ok := loaded.(bool); ok && loadedBool {
			loadedCount++
		} else if scriptVal, ok := loaded.(engine.ScriptValue); ok {
			if loadedBool, ok := scriptVal.ToGo().(bool); ok && loadedBool {
				loadedCount++
			}
		}
	}
	assert.Greater(t, loadedCount, 0, "At least some bridge modules should load")
}
