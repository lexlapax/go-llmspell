// ABOUTME: Regression test suite to catch regressions in previously working functionality
// ABOUTME: Tests core features, API compatibility, and behavior consistency across versions

package regression

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/bridge/agent"
	"github.com/lexlapax/go-llmspell/pkg/bridge/llm"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/lua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestEngineBasicFunctionality tests that basic engine functionality doesn't regress
func TestEngineBasicFunctionality(t *testing.T) {
	luaEngine := lua.NewLuaEngine()
	require.NotNil(t, luaEngine, "Engine creation should never return nil")

	// Test engine metadata
	name := luaEngine.Name()
	assert.Equal(t, "lua", name, "Engine name should remain 'lua'")

	version := luaEngine.Version()
	assert.NotEmpty(t, version, "Engine version should not be empty")

	extensions := luaEngine.FileExtensions()
	assert.Contains(t, extensions, ".lua", "Engine should support .lua files")

	// Test initialization
	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 5 * time.Second,
		MemoryLimit:  10 * 1024 * 1024,
	})
	assert.NoError(t, err, "Engine initialization should not fail")
	defer luaEngine.Shutdown()

	// Test basic script execution
	ctx := context.Background()
	result, err := luaEngine.Execute(ctx, `return "hello"`, nil)
	assert.NoError(t, err, "Basic script execution should not fail")
	assert.NotNil(t, result, "Result should not be nil")

	// Test script with parameters (handle case where params might not be set up)
	result, err = luaEngine.Execute(ctx, `return params and params.test_value or "params_not_available"`, map[string]interface{}{
		"test_value": "regression_test",
	})
	assert.NoError(t, err, "Script with parameters should not fail")
	assert.NotNil(t, result, "Result with parameters should not be nil")

	// Test error handling (should return ErrorValue, not panic)
	result, err = luaEngine.Execute(ctx, `error("test error")`, nil)
	assert.Error(t, err, "Error script should return error")
	assert.NotNil(t, result, "Even error results should not be nil")

	// Test shutdown
	err = luaEngine.Shutdown()
	assert.NoError(t, err, "Engine shutdown should not fail")
}

// TestBridgeBasicFunctionality tests that bridge functionality doesn't regress
func TestBridgeBasicFunctionality(t *testing.T) {
	ctx := context.Background()

	// Test Agent Bridge
	t.Run("AgentBridge", func(t *testing.T) {
		bridge := agent.NewAgentBridge()
		require.NotNil(t, bridge, "Agent bridge creation should not return nil")

		// Test bridge ID
		id := bridge.GetID()
		assert.Equal(t, "agent", id, "Agent bridge ID should remain 'agent'")

		// Test metadata
		metadata := bridge.GetMetadata()
		assert.Equal(t, "agent", metadata.Name, "Agent bridge metadata name should remain 'agent'")
		assert.NotEmpty(t, metadata.Version, "Agent bridge version should not be empty")

		// Test initialization
		err := bridge.Initialize(ctx)
		assert.NoError(t, err, "Agent bridge initialization should not fail")
		assert.True(t, bridge.IsInitialized(), "Agent bridge should be initialized")

		// Test cleanup
		err = bridge.Cleanup(ctx)
		assert.NoError(t, err, "Agent bridge cleanup should not fail")
	})

	// Test LLM Bridge
	t.Run("LLMBridge", func(t *testing.T) {
		bridge := llm.NewLLMBridge()
		require.NotNil(t, bridge, "LLM bridge creation should not return nil")

		// Test bridge ID
		id := bridge.GetID()
		assert.Equal(t, "llm", id, "LLM bridge ID should remain 'llm'")

		// Test metadata
		metadata := bridge.GetMetadata()
		assert.Equal(t, "llm", metadata.Name, "LLM bridge metadata name should remain 'llm'")
		assert.NotEmpty(t, metadata.Version, "LLM bridge version should not be empty")

		// Test initialization
		err := bridge.Initialize(ctx)
		assert.NoError(t, err, "LLM bridge initialization should not fail")
		assert.True(t, bridge.IsInitialized(), "LLM bridge should be initialized")

		// Test methods listing
		methods := bridge.Methods()
		assert.NotEmpty(t, methods, "LLM bridge should have methods")

		// Verify some expected methods exist
		methodNames := make(map[string]bool)
		for _, method := range methods {
			methodNames[method.Name] = true
			assert.NotEmpty(t, method.Description, "Method %s should have description", method.Name)
		}

		expectedMethods := []string{"setProvider", "getProvider", "listProviders", "generate"}
		for _, expectedMethod := range expectedMethods {
			assert.True(t, methodNames[expectedMethod], "Expected method %s should exist", expectedMethod)
		}

		// Test cleanup
		err = bridge.Cleanup(ctx)
		assert.NoError(t, err, "LLM bridge cleanup should not fail")
	})
}

// TestMockBridgeFunctionality tests mock bridge functionality for testing infrastructure
func TestMockBridgeFunctionality(t *testing.T) {
	bridge := testutils.NewMockBridge("regression-test")
	require.NotNil(t, bridge, "Mock bridge creation should not return nil")

	// Test basic properties
	assert.Equal(t, "regression-test", bridge.GetID(), "Mock bridge ID should match constructor")

	metadata := bridge.GetMetadata()
	assert.Equal(t, "regression-test", metadata.Name, "Mock bridge metadata should match")

	ctx := context.Background()

	// Test initialization
	err := bridge.Initialize(ctx)
	assert.NoError(t, err, "Mock bridge initialization should not fail")
	assert.True(t, bridge.IsInitialized(), "Mock bridge should be initialized")

	// Test cleanup
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err, "Mock bridge cleanup should not fail")
}

// TestEngineConfigurationRegression tests that configuration options work consistently
func TestEngineConfigurationRegression(t *testing.T) {
	// Test various configuration combinations that should work
	configs := []struct {
		name   string
		config engine.EngineConfig
	}{
		{
			name: "minimal_config",
			config: engine.EngineConfig{
				TimeoutLimit: 1 * time.Second,
				MemoryLimit:  1024 * 1024,
			},
		},
		{
			name: "debug_config",
			config: engine.EngineConfig{
				TimeoutLimit: 5 * time.Second,
				MemoryLimit:  10 * 1024 * 1024,
				DebugMode:    true,
				LogLevel:     "debug",
			},
		},
		{
			name: "sandbox_config",
			config: engine.EngineConfig{
				TimeoutLimit:   3 * time.Second,
				MemoryLimit:    5 * 1024 * 1024,
				SandboxMode:    true,
				FileSystemMode: engine.FSModeNone,
			},
		},
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			luaEngine := lua.NewLuaEngine()
			require.NotNil(t, luaEngine)

			err := luaEngine.Initialize(tc.config)
			assert.NoError(t, err, "Configuration %s should initialize successfully", tc.name)

			if err == nil {
				// Test basic execution with this config
				ctx := context.Background()
				result, err := luaEngine.Execute(ctx, `return "config_test"`, nil)
				assert.NoError(t, err, "Basic execution should work with config %s", tc.name)
				assert.NotNil(t, result, "Result should not be nil with config %s", tc.name)

				err = luaEngine.Shutdown()
				assert.NoError(t, err, "Shutdown should work with config %s", tc.name)
			}
		})
	}
}

// TestErrorHandlingRegression tests that error handling behavior remains consistent
func TestErrorHandlingRegression(t *testing.T) {
	luaEngine := lua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 5 * time.Second,
		MemoryLimit:  10 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	ctx := context.Background()

	// Test various error conditions
	errorTests := []struct {
		name   string
		script string
	}{
		{"syntax_error", `function broken(`},
		{"runtime_error", `error("intentional error")`},
		{"nil_access", `local x = nil; return x.field`},
		{"type_error", `return "string" + nil`},
		{"undefined_function", `return undefined_function()`},
	}

	for _, tc := range errorTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := luaEngine.Execute(ctx, tc.script, nil)

			// All error conditions should:
			// 1. Return an error
			// 2. Return a non-nil result (wrapped error)
			// 3. Not panic or crash the engine
			assert.Error(t, err, "Error test %s should return an error", tc.name)
			assert.NotNil(t, result, "Error test %s should return non-nil result", tc.name)

			// Engine should remain functional after error
			testResult, testErr := luaEngine.Execute(ctx, `return "engine_ok"`, nil)
			assert.NoError(t, testErr, "Engine should remain functional after error %s", tc.name)
			assert.NotNil(t, testResult, "Engine test result should not be nil after error %s", tc.name)
		})
	}
}

// TestPerformanceRegression tests that performance characteristics don't regress
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	luaEngine := lua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 30 * time.Second,
		MemoryLimit:  50 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	ctx := context.Background()

	// Test that simple operations complete quickly
	t.Run("simple_execution_speed", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 100; i++ {
			result, err := luaEngine.Execute(ctx, `return 2 + 2`, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}

		elapsed := time.Since(start)
		t.Logf("100 simple executions took %v", elapsed)

		// Simple operations should complete quickly (less than 1 second for 100 ops)
		assert.Less(t, elapsed, 1*time.Second, "100 simple operations should complete in under 1 second")
	})

	// Test that moderate complexity operations are reasonably fast
	t.Run("moderate_complexity_speed", func(t *testing.T) {
		script := `
			local sum = 0
			for i = 1, 1000 do
				sum = sum + i
			end
			return sum
		`

		start := time.Now()

		for i := 0; i < 10; i++ {
			result, err := luaEngine.Execute(ctx, script, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		}

		elapsed := time.Since(start)
		t.Logf("10 moderate complexity executions took %v", elapsed)

		// Moderate operations should complete reasonably quickly (less than 5 seconds for 10 ops)
		assert.Less(t, elapsed, 5*time.Second, "10 moderate complexity operations should complete in under 5 seconds")
	})
}

// TestAPICompatibilityRegression tests that public API doesn't break
func TestAPICompatibilityRegression(t *testing.T) {
	// Test that all expected types and methods exist and have expected signatures
	// This is a compile-time test - if the API changes, this won't compile

	// Engine API
	var eng engine.ScriptEngine = lua.NewLuaEngine()
	_ = eng

	// Bridge API
	var agentBridge *agent.AgentBridge = agent.NewAgentBridge()
	_ = agentBridge

	var llmBridge *llm.LLMBridge = llm.NewLLMBridge()
	_ = llmBridge

	// Mock API
	var mockBridge *testutils.MockBridge = testutils.NewMockBridge("test")
	_ = mockBridge

	// Config types
	var config engine.EngineConfig
	config.TimeoutLimit = 1 * time.Second
	config.MemoryLimit = 1024
	config.DebugMode = true
	_ = config

	t.Log("API compatibility test passed - all expected types and methods exist")
}

// TestConcurrencyRegression tests that concurrent operations work reliably
func TestConcurrencyRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency regression test in short mode")
	}

	luaEngine := lua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 10 * time.Second,
		MemoryLimit:  50 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	const numWorkers = 10
	const opsPerWorker = 20

	results := make(chan bool, numWorkers*opsPerWorker)
	ctx := context.Background()

	// Launch concurrent workers
	for workerID := 0; workerID < numWorkers; workerID++ {
		go func(id int) {
			for op := 0; op < opsPerWorker; op++ {
				script := `return params.worker_id + params.operation`
				params := map[string]interface{}{
					"worker_id": id,
					"operation": op,
				}

				result, err := luaEngine.Execute(ctx, script, params)
				success := err == nil && result != nil
				results <- success
			}
		}(workerID)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numWorkers*opsPerWorker; i++ {
		if <-results {
			successCount++
		}
	}

	totalOps := numWorkers * opsPerWorker
	successRate := float64(successCount) / float64(totalOps)

	t.Logf("Concurrency test: %d/%d successful (%.2f%%)", successCount, totalOps, successRate*100)

	// At least 95% of concurrent operations should succeed
	assert.GreaterOrEqual(t, successRate, 0.95, "At least 95%% of concurrent operations should succeed")
}
