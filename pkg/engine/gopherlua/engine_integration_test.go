// ABOUTME: Integration tests for complete LuaEngine functionality combining all components
// ABOUTME: Tests real-world scenarios with bridges, security, type conversion, and execution pipeline

package gopherlua

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestLuaEngine_FullIntegration(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	// Initialize with comprehensive configuration
	config := engine.EngineConfig{
		MemoryLimit:     64 * 1024 * 1024, // 64MB
		TimeoutLimit:    10 * time.Second,
		SandboxMode:     true,
		AllowedModules:  []string{"string", "math", "table"},
		DisabledModules: []string{"io", "os", "debug"},
		DebugMode:       false,
		EngineOptions: map[string]interface{}{
			"pool_min_size":     2,
			"pool_max_size":     8,
			"pool_idle_timeout": "5m",
			"health_threshold":  0.8,
			"cleanup_interval":  "30s",
			"security_level":    "standard",
		},
	}

	err := eng.Initialize(config)
	require.NoError(t, err)

	// Verify engine metadata
	assert.Equal(t, "lua", eng.Name())
	assert.NotEmpty(t, eng.Version())
	assert.Contains(t, eng.FileExtensions(), ".lua")

	features := eng.Features()
	assert.Contains(t, features, engine.FeatureCoroutines)
	assert.Contains(t, features, engine.FeatureModules)
	assert.Contains(t, features, engine.FeatureCompilation)

	// Test multiple execution scenarios
	ctx := context.Background()

	t.Run("basic_script_execution", func(t *testing.T) {
		script := `
			local result = {
				sum = a + b,
				product = a * b,
				message = "Hello " .. name,
				timestamp = 12345 -- Fixed timestamp since os.time is restricted
			}
			return result
		`

		params := map[string]interface{}{
			"a":    15,
			"b":    25,
			"name": "Integration Test",
		}

		result, err := eng.Execute(ctx, script, params)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 40.0, resultMap["sum"])
		assert.Equal(t, 375.0, resultMap["product"])
		assert.Equal(t, "Hello Integration Test", resultMap["message"])
	})

	t.Run("type_conversion_comprehensive", func(t *testing.T) {
		script := `
			-- Test various data type conversions
			local output = {}
			
			-- Process input data
			output.bool_flag = input.enabled
			output.count = input.items and #input.items or 0
			output.total = 0
			
			if input.items then
				for i = 1, #input.items do
					local item = input.items[i]
					output.total = output.total + (item.value or 0)
				end
			end
			
			-- Create nested structure
			output.nested = {
				config = input.config or {},
				computed = {
					average = output.count > 0 and output.total / output.count or 0,
					status = output.total > 100 and "high" or "low"
				}
			}
			
			return output
		`

		params := map[string]interface{}{
			"input": map[string]interface{}{
				"enabled": true,
				"items": []interface{}{
					map[string]interface{}{"value": 30},
					map[string]interface{}{"value": 45},
					map[string]interface{}{"value": 25},
				},
				"config": map[string]interface{}{
					"debug":   false,
					"version": "1.0",
				},
			},
		}

		result, err := eng.Execute(ctx, script, params)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, true, resultMap["bool_flag"])
		assert.Equal(t, 3.0, resultMap["count"])
		assert.Equal(t, 100.0, resultMap["total"])

		nested, ok := resultMap["nested"].(map[string]interface{})
		require.True(t, ok)

		computed, ok := nested["computed"].(map[string]interface{})
		require.True(t, ok)
		assert.InDelta(t, 33.333, computed["average"].(float64), 0.01)
		assert.Equal(t, "low", computed["status"])
	})

	t.Run("security_sandbox_enforcement", func(t *testing.T) {
		// Test that blocked modules are not accessible
		blockedScripts := []string{
			`return io.open("/tmp/test", "r")`,
			`return os.execute("echo test")`,
			`return require("debug")`,
			`return loadfile("/etc/passwd")`,
		}

		for i, script := range blockedScripts {
			_, err := eng.Execute(ctx, script, nil)
			assert.Error(t, err, "Script %d should be blocked", i)
		}

		// Test that allowed modules work
		allowedScript := `
			return {
				string_len = string.len("test"),
				math_pi = math.pi,
				table_size = table.getn and table.getn({1,2,3}) or 3
			}
		`

		result, err := eng.Execute(ctx, allowedScript, nil)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 4.0, resultMap["string_len"])
		assert.InDelta(t, 3.14159, resultMap["math_pi"].(float64), 0.001)
	})

	t.Run("performance_and_caching", func(t *testing.T) {
		script := `
			-- Fibonacci calculation (somewhat expensive)
			local function fib(n)
				if n <= 1 then
					return n
				else
					return fib(n-1) + fib(n-2)
				end
			end
			
			return {
				input = n,
				result = fib(n),
				cached = "this execution"
			}
		`

		// Execute the same script multiple times to test caching
		for i := 0; i < 5; i++ {
			result, err := eng.Execute(ctx, script, map[string]interface{}{"n": 10})
			require.NoError(t, err)

			resultMap, ok := result.(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, 10.0, resultMap["input"])
			assert.Equal(t, 55.0, resultMap["result"]) // fib(10) = 55
		}

		// Check metrics to verify caching is working
		metrics := eng.GetMetrics()
		assert.GreaterOrEqual(t, metrics.ScriptsExecuted, int64(5))
		assert.GreaterOrEqual(t, metrics.CacheHits, int64(4)) // First miss, others hit
	})

	t.Run("concurrent_execution_stability", func(t *testing.T) {
		script := `
			local sum = 0
			for i = 1, iterations do
				sum = sum + (i * multiplier)
			end
			return {
				sum = sum,
				worker_id = worker_id
			}
		`

		const numWorkers = 8
		const iterations = 100
		results := make(chan testResult, numWorkers)

		// Launch concurrent executions
		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				params := map[string]interface{}{
					"iterations": iterations,
					"multiplier": 2,
					"worker_id":  workerID,
				}

				result, err := eng.Execute(ctx, script, params)
				results <- testResult{
					workerID: workerID,
					result:   result,
					error:    err,
				}
			}(i)
		}

		// Collect and verify results
		for i := 0; i < numWorkers; i++ {
			res := <-results
			require.NoError(t, res.error, "Worker %d failed", res.workerID)

			resultMap, ok := res.result.(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, 10100.0, resultMap["sum"]) // sum of 2*i for i=1 to 100
			assert.Equal(t, float64(res.workerID), resultMap["worker_id"])
		}
	})

	t.Run("error_handling_and_recovery", func(t *testing.T) {
		errorScripts := []struct {
			name     string
			script   string
			errType  engine.ErrorType
			contains string
		}{
			{
				name:    "syntax_error",
				script:  `return 2 +`,
				errType: engine.ErrorTypeSyntax,
			},
			{
				name:    "runtime_error",
				script:  `error("intentional error")`,
				errType: engine.ErrorTypeRuntime,
			},
			{
				name:    "nil_operation",
				script:  `local x = nil; return x + 5`,
				errType: engine.ErrorTypeRuntime,
			},
		}

		for _, tt := range errorScripts {
			t.Run(tt.name, func(t *testing.T) {
				_, err := eng.Execute(ctx, tt.script, nil)
				require.Error(t, err)

				var engineErr *engine.EngineError
				if assert.ErrorAs(t, err, &engineErr) {
					assert.Equal(t, tt.errType, engineErr.Type)
					if tt.contains != "" {
						assert.Contains(t, engineErr.Message, tt.contains)
					}
				}
			})
		}

		// Verify engine can still execute scripts after errors
		result, err := eng.Execute(ctx, `return "recovery successful"`, nil)
		require.NoError(t, err)
		assert.Equal(t, "recovery successful", result)
	})

	t.Run("resource_management", func(t *testing.T) {
		// Test memory limit setting
		err := eng.SetMemoryLimit(128 * 1024 * 1024) // 128MB
		assert.NoError(t, err)

		// Test timeout limit setting
		err = eng.SetTimeout(5 * time.Second)
		assert.NoError(t, err)

		// Test comprehensive resource limits
		limits := engine.ResourceLimits{
			MaxMemory:     256 * 1024 * 1024, // 256MB
			MaxExecTime:   30 * time.Second,
			MaxGoroutines: 50,
		}
		err = eng.SetResourceLimits(limits)
		assert.NoError(t, err)

		// Execute a script and verify it works within limits
		result, err := eng.Execute(ctx, `
			local data = {}
			for i = 1, 1000 do
				data[i] = "item_" .. i
			end
			return #data
		`, nil)
		require.NoError(t, err)
		assert.Equal(t, 1000.0, result)
	})

	// Final metrics check
	finalMetrics := eng.GetMetrics()
	assert.Greater(t, finalMetrics.ScriptsExecuted, int64(0))
	assert.Greater(t, finalMetrics.TotalExecTime, time.Duration(0))
	assert.GreaterOrEqual(t, finalMetrics.ErrorCount, int64(0))
}

func TestLuaEngine_BridgeIntegration(t *testing.T) {
	eng := NewLuaEngine()
	defer func() {
		_ = eng.Shutdown()
	}()

	config := engine.EngineConfig{
		SandboxMode: false,
	}
	err := eng.Initialize(config)
	require.NoError(t, err)

	// Create and register a test bridge
	bridge := &testBridgeForRegistration{
		id: "test_integration_bridge",
		meta: engine.BridgeMetadata{
			Name:        "Integration Test Bridge",
			Version:     "1.0.0",
			Description: "Bridge for integration testing",
		},
		methods: []engine.MethodInfo{
			{
				Name:        "calculate",
				Description: "Performs calculation",
				Parameters: []engine.ParameterInfo{
					{Name: "operation", Type: "string", Required: true},
					{Name: "a", Type: "number", Required: true},
					{Name: "b", Type: "number", Required: true},
				},
				ReturnType: "number",
			},
		},
	}

	err = eng.RegisterBridge(bridge)
	require.NoError(t, err)

	// Test bridge access from Lua
	script := `
		-- Test bridge access
		local bridge = bridges.test_integration_bridge
		if not bridge then
			error("Bridge not found")
		end
		
		-- Check bridge metadata
		local meta = bridge._meta
		local result = {
			bridge_name = meta.name,
			bridge_version = meta.version,
			has_calculate = bridge.calculate ~= nil
		}
		
		return result
	`

	ctx := context.Background()
	result, err := eng.Execute(ctx, script, nil)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Integration Test Bridge", resultMap["bridge_name"])
	assert.Equal(t, "1.0.0", resultMap["bridge_version"])
	assert.Equal(t, true, resultMap["has_calculate"])

	// Verify bridge list
	bridges := eng.ListBridges()
	assert.Contains(t, bridges, "test_integration_bridge")

	// Unregister and verify cleanup
	err = eng.UnregisterBridge("test_integration_bridge")
	require.NoError(t, err)

	bridges = eng.ListBridges()
	assert.NotContains(t, bridges, "test_integration_bridge")
}

// testResult holds results from concurrent test executions
type testResult struct {
	workerID int
	result   interface{}
	error    error
}
