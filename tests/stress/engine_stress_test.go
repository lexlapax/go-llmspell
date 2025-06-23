// ABOUTME: Engine stress tests to verify performance and stability under heavy load
// ABOUTME: Tests concurrent execution, memory usage, timeout handling, and resource exhaustion scenarios

package stress

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// TestEngineConcurrentExecution tests concurrent script execution
func TestEngineConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 30 * time.Second,
		MemoryLimit:  100 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	const numGoroutines = 50
	const scriptsPerGoroutine = 10

	var wg sync.WaitGroup
	results := make(chan error, numGoroutines*scriptsPerGoroutine)

	script := `
		local sum = 0
		for i = 1, 1000 do
			sum = sum + i
		end
		return sum
	`

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < scriptsPerGoroutine; j++ {
				ctx := context.Background()
				result, err := luaEngine.Execute(ctx, script, map[string]interface{}{
					"goroutine_id": goroutineID,
					"script_id":    j,
				})
				
				if err != nil {
					results <- err
				} else if result == nil {
					results <- assert.AnError
				} else {
					results <- nil
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Count successes and failures
	successCount := 0
	errorCount := 0
	for result := range results {
		if result == nil {
			successCount++
		} else {
			errorCount++
			t.Logf("Error: %v", result)
		}
	}

	totalTests := numGoroutines * scriptsPerGoroutine
	t.Logf("Concurrent execution results: %d successes, %d errors out of %d total", 
		successCount, errorCount, totalTests)

	// Expect at least 90% success rate
	expectedMinSuccess := int(float64(totalTests) * 0.9)
	assert.GreaterOrEqual(t, successCount, expectedMinSuccess, 
		"Expected at least 90%% success rate in concurrent execution")
}

// TestEngineMemoryStress tests memory usage under stress
func TestEngineMemoryStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	// Set a reasonable memory limit
	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 10 * time.Second,
		MemoryLimit:  50 * 1024 * 1024, // 50MB
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	// Record initial memory stats
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Execute memory-intensive scripts
	const numIterations = 100
	script := `
		local bigTable = {}
		for i = 1, 1000 do
			bigTable[i] = "data_" .. tostring(i)
		end
		-- Return size to verify execution
		return #bigTable
	`

	ctx := context.Background()
	for i := 0; i < numIterations; i++ {
		result, err := luaEngine.Execute(ctx, script, nil)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Force garbage collection periodically
		if i%20 == 0 {
			runtime.GC()
		}
	}

	// Final memory check
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memoryIncrease := m2.Alloc - m1.Alloc
	t.Logf("Memory increase: %d bytes (%.2f MB)", 
		memoryIncrease, float64(memoryIncrease)/(1024*1024))

	// Memory increase should be reasonable (less than 20MB)
	maxReasonableIncrease := uint64(20 * 1024 * 1024)
	assert.LessOrEqual(t, memoryIncrease, maxReasonableIncrease,
		"Memory increase should be reasonable after stress test")
}

// TestEngineTimeoutStress tests timeout handling under stress
func TestEngineTimeoutStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	// Very short timeout to force timeouts
	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 100 * time.Millisecond,
		MemoryLimit:  50 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	const numTimeoutTests = 20
	timeoutScript := `
		local count = 0
		while count < 1000000 do
			count = count + 1
		end
		return count
	`

	timeoutCount := 0
	successCount := 0

	ctx := context.Background()
	for i := 0; i < numTimeoutTests; i++ {
		result, err := luaEngine.Execute(ctx, timeoutScript, nil)
		
		if err != nil {
			timeoutCount++
			t.Logf("Test %d timed out as expected: %v", i, err)
		} else {
			successCount++
			t.Logf("Test %d completed unexpectedly: %v", i, result)
		}
	}

	t.Logf("Timeout stress test results: %d timeouts, %d successes out of %d total", 
		timeoutCount, successCount, numTimeoutTests)

	// Most tests should timeout due to the very short timeout
	assert.Greater(t, timeoutCount, numTimeoutTests/2, 
		"Most tests should timeout with very short timeout limit")
}

// TestEngineResourceExhaustion tests behavior under resource exhaustion
func TestEngineResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	// Very low memory limit to trigger exhaustion
	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 5 * time.Second,
		MemoryLimit:  1024 * 1024, // Only 1MB
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	// Script that tries to allocate too much memory
	memoryExhaustionScript := `
		local bigTable = {}
		for i = 1, 100000 do
			bigTable[i] = string.rep("x", 1000) -- 1KB per entry
		end
		return #bigTable
	`

	ctx := context.Background()
	result, err := luaEngine.Execute(ctx, memoryExhaustionScript, nil)

	// Engine should handle the memory exhaustion gracefully
	// Either by limiting memory usage or returning an error
	if err != nil {
		t.Logf("Memory exhaustion handled with error: %v", err)
		assert.NotNil(t, result) // Should still return an error value
	} else {
		t.Logf("Memory exhaustion test completed: %v", result)
	}

	// Engine should still be functional after exhaustion
	simpleScript := `return "engine still works"`
	result2, err2 := luaEngine.Execute(ctx, simpleScript, nil)
	assert.NoError(t, err2)
	assert.NotNil(t, result2)
	t.Log("Engine remains functional after resource exhaustion test")
}

// TestEngineRapidScriptSwitching tests rapid switching between different scripts
func TestEngineRapidScriptSwitching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	luaEngine := gopherlua.NewLuaEngine()
	require.NotNil(t, luaEngine)

	err := luaEngine.Initialize(engine.EngineConfig{
		TimeoutLimit: 10 * time.Second,
		MemoryLimit:  50 * 1024 * 1024,
	})
	require.NoError(t, err)
	defer luaEngine.Shutdown()

	scripts := []string{
		`return 2 + 2`,
		`return "hello world"`,
		`local t = {a=1, b=2}; return t.a + t.b`,
		`for i=1,10 do end; return "loop done"`,
		`local function add(a,b) return a + b end; return add(5, 3)`,
	}

	const iterations = 200
	ctx := context.Background()

	for i := 0; i < iterations; i++ {
		script := scripts[i%len(scripts)]
		result, err := luaEngine.Execute(ctx, script, nil)
		
		assert.NoError(t, err, "Script %d failed: %s", i, script)
		assert.NotNil(t, result, "Script %d returned nil result", i)

		// Log progress every 50 iterations
		if i%50 == 0 {
			t.Logf("Completed %d rapid script switching iterations", i)
		}
	}

	t.Logf("Successfully completed %d rapid script switching iterations", iterations)
}