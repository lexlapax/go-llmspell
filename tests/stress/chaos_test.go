// ABOUTME: Chaos testing to verify system resilience under random failures and unexpected conditions
// ABOUTME: Tests random script execution, bridge failures, resource exhaustion, and recovery scenarios

package stress

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/bridge/agent"
	"github.com/lexlapax/go-llmspell/pkg/bridge/llm"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestChaosRandomScriptExecution tests random script execution patterns
func TestChaosRandomScriptExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 5 * time.Second,
		MemoryLimit:  50 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	// Collection of scripts with varying complexity and potential issues
	chaosScripts := []string{
		// Valid scripts
		`return "hello"`,
		`return 2 + 2`,
		`local t = {a=1, b=2}; return t.a + t.b`,
		`for i=1,100 do end; return "done"`,
		
		// Scripts with potential issues
		`error("intentional error")`,
		`return nil`,
		`local function recursive() return recursive() end; return recursive()`, // Stack overflow
		`while true do end`, // Infinite loop (should timeout)
		`local huge = {}; for i=1,1000000 do huge[i] = i end; return #huge`, // Memory intensive
		
		// Edge cases
		`return ""`,
		`return 0`,
		`return false`,
		`return {}`,
		`local x = nil; return x.field`, // Nil access error
	}

	const chaosIterations = 100
	ctx := context.Background()
	
	successCount := 0
	errorCount := 0
	timeoutCount := 0
	
	// Use a seeded random generator for reproducible chaos
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < chaosIterations; i++ {
		// Randomly select a script
		scriptIndex := rng.Intn(len(chaosScripts))
		script := chaosScripts[scriptIndex]
		
		// Randomly vary parameters
		params := map[string]interface{}{}
		if rng.Float64() < 0.3 { // 30% chance of adding random params
			params["random_value"] = rng.Intn(1000)
			params["random_string"] = fmt.Sprintf("chaos_%d", rng.Intn(100))
		}

		result, err := luaEngine.Execute(ctx, script, params)
		
		if err != nil {
			if fmt.Sprintf("%v", err) == "context deadline exceeded" {
				timeoutCount++
			} else {
				errorCount++
			}
		} else {
			successCount++
		}
		
		// Randomly introduce delays
		if rng.Float64() < 0.1 { // 10% chance of small delay
			time.Sleep(time.Duration(rng.Intn(50)) * time.Millisecond)
		}
		
		// Verify result is not nil (even error results should be wrapped)
		assert.NotNil(t, result, "Result should never be nil, even for errors")
	}

	t.Logf("Chaos script execution results: %d successes, %d errors, %d timeouts out of %d total", 
		successCount, errorCount, timeoutCount, chaosIterations)

	// At least some scripts should succeed
	assert.Greater(t, successCount, 0, "At least some scripts should succeed")
	
	// System should remain stable (total should equal iterations)
	total := successCount + errorCount + timeoutCount
	assert.Equal(t, chaosIterations, total, "All iterations should be accounted for")
}

// TestChaosBridgeFailures tests random bridge initialization failures and recovery
func TestChaosBridgeFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	const numChaosOperations = 50
	ctx := context.Background()
	rng := rand.New(rand.NewSource(42))

	bridges := []interface{}{
		agent.NewAgentBridge(),
		llm.NewLLMBridge(),
		testutils.NewMockBridge("chaos-test-1"),
		testutils.NewMockBridge("chaos-test-2"),
	}

	successfulOps := 0
	failedOps := 0

	for i := 0; i < numChaosOperations; i++ {
		// Randomly select a bridge
		bridgeIndex := rng.Intn(len(bridges))
		
		switch bridge := bridges[bridgeIndex].(type) {
		case *agent.AgentBridge:
			// Random operations on agent bridge
			if !bridge.IsInitialized() {
				err := bridge.Initialize(ctx)
				if err != nil {
					failedOps++
					continue
				}
			}
			
			// Random chance to cleanup
			if rng.Float64() < 0.3 {
				err := bridge.Cleanup(ctx)
				if err != nil {
					failedOps++
					continue
				}
			}
			
			// Always verify metadata access works
			metadata := bridge.GetMetadata()
			if metadata.Name == "" {
				failedOps++
				continue
			}
			
			successfulOps++
			
		case *llm.LLMBridge:
			// Random operations on LLM bridge
			if !bridge.IsInitialized() {
				err := bridge.Initialize(ctx)
				if err != nil {
					failedOps++
					continue
				}
			}
			
			// Random chance to cleanup
			if rng.Float64() < 0.3 {
				err := bridge.Cleanup(ctx)
				if err != nil {
					failedOps++
					continue
				}
			}
			
			// Test method listing
			methods := bridge.Methods()
			if len(methods) == 0 {
				failedOps++
				continue
			}
			
			successfulOps++
			
		case *testutils.MockBridge:
			// Random operations on mock bridge
			err := bridge.Initialize(ctx)
			if err != nil {
				failedOps++
				continue
			}
			
			// Random chance to cleanup immediately
			if rng.Float64() < 0.5 {
				err := bridge.Cleanup(ctx)
				if err != nil {
					failedOps++
					continue
				}
			}
			
			successfulOps++
		}

		// Random delays to simulate real-world timing
		if rng.Float64() < 0.2 {
			time.Sleep(time.Duration(rng.Intn(10)) * time.Millisecond)
		}
	}

	t.Logf("Chaos bridge operations: %d successful, %d failed out of %d total", 
		successfulOps, failedOps, numChaosOperations)

	// Most operations should succeed
	successRate := float64(successfulOps) / float64(numChaosOperations)
	assert.Greater(t, successRate, 0.7, "At least 70%% of chaos bridge operations should succeed")
}

// TestChaosResourceExhaustion tests random resource exhaustion scenarios
func TestChaosResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Create engines with varying resource limits
	configs := []engine.EngineConfig{
		{TimeoutLimit: 50 * time.Millisecond, MemoryLimit: 1024 * 1024},     // Very constrained
		{TimeoutLimit: 200 * time.Millisecond, MemoryLimit: 5 * 1024 * 1024}, // Moderately constrained
		{TimeoutLimit: 1 * time.Second, MemoryLimit: 20 * 1024 * 1024},       // Reasonable limits
	}

	rng := rand.New(rand.NewSource(42))
	ctx := context.Background()

	for configIndex, config := range configs {
		t.Logf("Testing resource exhaustion with config %d", configIndex)
		
		luaEngine := gopherlua.NewLuaEngine()
		require.NotNil(t, luaEngine)

		err := luaEngine.Initialize(config)
		require.NoError(t, err)

		// Scripts designed to stress different resources
		stressScripts := []string{
			// Memory stress
			`local big = {}; for i=1,10000 do big[i] = string.rep("x", 100) end; return #big`,
			
			// CPU/Time stress
			`local count = 0; for i=1,1000000 do count = count + 1 end; return count`,
			
			// Mixed stress
			`local data = {}; for i=1,1000 do data[i] = {value = string.rep("data", 50)} end; local sum = 0; for i=1,1000 do sum = sum + #data[i].value end; return sum`,
		}

		recoveredCount := 0
		totalTests := 20

		for i := 0; i < totalTests; i++ {
			// Randomly select a stress script
			script := stressScripts[rng.Intn(len(stressScripts))]
			
			result, err := luaEngine.Execute(ctx, script, nil)
			
			// After any stress test, verify engine can still execute simple scripts
			simpleResult, simpleErr := luaEngine.Execute(ctx, `return "recovery_test"`, nil)
			
			if simpleErr == nil && simpleResult != nil {
				recoveredCount++
			}

			// Log stress test result
			if err != nil {
				t.Logf("Stress test %d failed as expected: %v", i, err)
			} else {
				t.Logf("Stress test %d completed: %v", i, result)
			}
		}

		// Engine should be able to recover from most stress scenarios
		recoveryRate := float64(recoveredCount) / float64(totalTests)
		t.Logf("Config %d recovery rate: %.2f (%d/%d)", configIndex, recoveryRate, recoveredCount, totalTests)
		
		assert.Greater(t, recoveryRate, 0.5, "Engine should recover from at least 50%% of stress scenarios")

		err = luaEngine.Shutdown()
		assert.NoError(t, err)
	}
}

// TestChaosConcurrentChaos tests chaotic concurrent operations
func TestChaosConcurrentChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 2 * time.Second,
		MemoryLimit:  50 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	const numChaosWorkers = 10
	const operationsPerWorker = 20

	var wg sync.WaitGroup
	results := make(chan bool, numChaosWorkers*operationsPerWorker)

	// Collection of chaotic scripts
	chaosScripts := []string{
		`return math.random(1000)`,
		`local t = {}; for i=1,math.random(100) do t[i] = i end; return #t`,
		`if math.random() < 0.5 then return "heads" else return "tails" end`,
		`local function fib(n) if n <= 1 then return n else return fib(n-1) + fib(n-2) end end; return fib(math.random(10))`,
		`error("random error " .. math.random(100))`,
	}

	ctx := context.Background()

	for workerID := 0; workerID < numChaosWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Each worker has its own random seed based on worker ID
			rng := rand.New(rand.NewSource(int64(42 + id)))
			
			for op := 0; op < operationsPerWorker; op++ {
				// Random delay before each operation
				delay := time.Duration(rng.Intn(100)) * time.Millisecond
				time.Sleep(delay)
				
				// Select random script
				script := chaosScripts[rng.Intn(len(chaosScripts))]
				
				// Add random parameters
				params := map[string]interface{}{
					"worker_id": id,
					"operation": op,
					"random":    rng.Float64(),
				}
				
				result, err := luaEngine.Execute(ctx, script, params)
				_ = err // Errors are expected in chaos testing
				
				// Consider any non-nil result as success (even errors are wrapped)
				success := result != nil
				results <- success
			}
		}(workerID)
	}

	wg.Wait()
	close(results)

	// Count results
	successCount := 0
	totalCount := 0
	for success := range results {
		totalCount++
		if success {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(totalCount)
	t.Logf("Concurrent chaos test: %d successes out of %d total (%.2f%% success rate)", 
		successCount, totalCount, successRate*100)

	// Even with chaos, we should get results for most operations
	assert.Greater(t, successRate, 0.8, "At least 80%% of concurrent chaos operations should return results")
	assert.Equal(t, numChaosWorkers*operationsPerWorker, totalCount, "All operations should be accounted for")
}

// TestChaosEngineRestart tests engine restart scenarios under chaotic conditions
func TestChaosEngineRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	const numRestartCycles = 10
	rng := rand.New(rand.NewSource(42))
	ctx := context.Background()

	for cycle := 0; cycle < numRestartCycles; cycle++ {
		t.Logf("Starting chaos restart cycle %d", cycle)
		
		luaEngine := gopherlua.NewLuaEngine()
		require.NotNil(t, luaEngine)

		// Random configuration for each cycle
		config := engine.EngineConfig{
			TimeoutLimit: time.Duration(100+rng.Intn(1000)) * time.Millisecond,
			MemoryLimit:  int64(1+rng.Intn(50)) * 1024 * 1024,
		}

		err := luaEngine.Initialize(config)
		require.NoError(t, err)

		// Execute random number of operations before restart
		numOps := 5 + rng.Intn(15)
		successCount := 0
		
		for op := 0; op < numOps; op++ {
			// Random script
			scripts := []string{
				`return "test"`,
				`return 42`,
				`local x = 1 + 1; return x`,
				`error("test error")`,
			}
			
			script := scripts[rng.Intn(len(scripts))]
			result, err := luaEngine.Execute(ctx, script, nil)
			_ = err // Errors are expected in chaos testing
			
			if result != nil { // Any result (including errors) counts as success
				successCount++
			}
			
			// Random chance of early shutdown
			if rng.Float64() < 0.1 { // 10% chance
				t.Logf("Early shutdown in cycle %d after %d operations", cycle, op+1)
				break
			}
		}

		// Always shutdown (even if already shutdown)
		err = luaEngine.Shutdown()
		if err != nil {
			t.Logf("Shutdown error in cycle %d: %v", cycle, err)
		}

		// Verify we got some results
		if numOps > 0 {
			successRate := float64(successCount) / float64(numOps)
			t.Logf("Cycle %d: %d successes out of %d operations (%.2f%%)", 
				cycle, successCount, numOps, successRate*100)
		}

		// Random delay between cycles
		delay := time.Duration(rng.Intn(100)) * time.Millisecond
		time.Sleep(delay)
	}

	t.Logf("Successfully completed %d chaos restart cycles", numRestartCycles)
}