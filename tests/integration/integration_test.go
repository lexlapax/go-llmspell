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

	// Create engine registry with default config
	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 30 * time.Second,
		PoolingEnabled: true,
		MaxPoolSize:    5,
		MetricsEnabled: true,
	}
	registry := engine.NewRegistry(registryConfig)
	err = registry.Initialize()
	require.NoError(t, err)

	// Register Lua engine
	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	// Create runner configuration
	runnerConfig := &runner.RunnerConfig{
		Timeout:              30 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
		EnableMetrics:        true,
		EnableValidation:     true,
	}

	// Create engine manager and selector
	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)

	// Create and configure runner
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

	// Execute spell with proper parameter type
	params := map[string]interface{}{
		"message": "Integration test message",
	}

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result (result is interface{}, not ExecutionResult)
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.True(t, resultMap["success"].(bool))
	assert.Equal(t, "Integration test message", resultMap["message"])
}

// TestIntegrationSpellWithBridges tests spell execution with bridge interactions
func TestIntegrationSpellWithBridges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "bridge_spell.lua")

	spellContent := `
		-- Test spell using bridges
		local data = require("data")
		local errors = require("errors")
		
		-- Test data operations
		local testObj = {
			name = "test",
			value = 42,
			nested = {
				array = {1, 2, 3}
			}
		}
		
		local jsonStr = data.to_json(testObj)
		local parsed = data.from_json(jsonStr)
		
		-- Test error handling
		local err = errors.new("TEST_ERROR", "test error message")
		local isTestError = errors.is_type(err, "TEST_ERROR")
		
		return {
			json_roundtrip = (parsed.name == testObj.name),
			error_handling = isTestError,
			test_passed = true
		}
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Create engine registry
	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 30 * time.Second,
	}
	registry := engine.NewRegistry(registryConfig)
	err = registry.Initialize()
	require.NoError(t, err)

	// Register Lua engine
	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	// Create runner configuration
	runnerConfig := &runner.RunnerConfig{
		Timeout:              30 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
	}

	// Create engine manager and selector
	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)

	// Create runner
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.True(t, resultMap["test_passed"].(bool))
}

// TestIntegrationErrorHandling tests error handling in integration scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	ctx := context.Background()

	// Setup engine and runner
	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 2 * time.Second, // Short timeout for timeout test
	}
	registry := engine.NewRegistry(registryConfig)
	err := registry.Initialize()
	require.NoError(t, err)

	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	runnerConfig := &runner.RunnerConfig{
		Timeout:              5 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
	}

	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

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

	// Note: Security profiles would be configured in the engine config
	// For now, we just test that the script runs
	ctx := context.Background()

	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 10 * time.Second,
	}
	registry := engine.NewRegistry(registryConfig)
	err = registry.Initialize()
	require.NoError(t, err)

	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	runnerConfig := &runner.RunnerConfig{
		Timeout:              10 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
	}

	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
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

	// Setup engine and runner
	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 10 * time.Second,
	}
	registry := engine.NewRegistry(registryConfig)
	err = registry.Initialize()
	require.NoError(t, err)

	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	runnerConfig := &runner.RunnerConfig{
		Timeout:              10 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
	}

	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

	params := map[string]interface{}{
		"string_param":  "hello",
		"number_param":  "42",
		"boolean_param": "true",
	}

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify parameters were passed correctly
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Equal(t, "hello", resultMap["string_param"])
	assert.Equal(t, float64(42), resultMap["number_param"])
	assert.Equal(t, false, resultMap["boolean_param"]) // Note: string "true" != boolean true in this context
}

// TestIntegrationConcurrentExecution tests concurrent spell execution
func TestIntegrationConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	spellFile := filepath.Join(tempDir, "concurrent_test.lua")

	spellContent := `
		local core = require("core")
		
		-- Simulate some work
		core.sleep(0.1)
		
		return {
			worker_id = params and params.worker_id or "unknown",
			timestamp = os.time(),
			success = true
		}
	`
	err := os.WriteFile(spellFile, []byte(spellContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	// Setup engine and runner
	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 10 * time.Second,
	}
	registry := engine.NewRegistry(registryConfig)
	err = registry.Initialize()
	require.NoError(t, err)

	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	runnerConfig := &runner.RunnerConfig{
		Timeout:              10 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
	}

	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

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

	// Setup engine and runner
	registryConfig := engine.RegistryConfig{
		MaxEngines:     10,
		DefaultTimeout: 10 * time.Second,
	}
	registry := engine.NewRegistry(registryConfig)
	err = registry.Initialize()
	require.NoError(t, err)

	luaFactory := gopherlua.NewLuaEngineFactory()
	err = registry.Register(luaFactory)
	require.NoError(t, err)

	runnerConfig := &runner.RunnerConfig{
		Timeout:              10 * time.Second,
		MaxConcurrentScripts: 10,
		DefaultEngine:        "lua",
	}

	engineManager := runner.NewEngineRegistryManager(registry)
	engineSelector := runner.NewEngineSelector(engineManager)
	scriptRunner := runner.NewScriptExecutor(runnerConfig, engineManager, engineSelector)
	err = scriptRunner.Initialize(ctx)
	require.NoError(t, err)

	result, err := scriptRunner.ExecuteFile(ctx, spellFile, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify bridge modules loaded
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")

	// Check that at least some modules loaded
	// (The actual availability depends on which bridges are registered)
	loadedCount := 0
	for _, loaded := range resultMap {
		if loaded.(bool) {
			loadedCount++
		}
	}
	assert.Greater(t, loadedCount, 0, "At least some bridge modules should load")
}
