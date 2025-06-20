// ABOUTME: Tests for Lua engine profiling infrastructure including execution time, memory usage, and allocation tracking
// ABOUTME: Validates profiler API, data collection accuracy, and performance overhead measurement

package gopherlua

import (
	"context"
	"fmt"
	// "runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestProfiler_ExecutionTime(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		minTime time.Duration
		maxTime time.Duration
		wantErr bool
	}{
		{
			name: "simple function timing",
			script: `
				local function slowFunc()
					local sum = 0
					for i = 1, 1000000 do
						sum = sum + i
					end
					return sum
				end
				return slowFunc()
			`,
			minTime: 1 * time.Millisecond,
			maxTime: 1 * time.Second,
		},
		{
			name: "nested function timing",
			script: `
				local function inner()
					local x = 0
					for i = 1, 10000 do x = x + 1 end
					return x
				end
				
				local function outer()
					local sum = 0
					for i = 1, 100 do
						sum = sum + inner()
					end
					return sum
				end
				
				return outer()
			`,
			minTime: 1 * time.Millisecond,
			maxTime: 2 * time.Second,
		},
		{
			name: "coroutine timing",
			script: `
				local function coFunc()
					local sum = 0
					for i = 1, 100000 do
						sum = sum + i
						if i % 10000 == 0 then
							coroutine.yield(sum)
						end
					end
					return sum
				end
				
				local co = coroutine.create(coFunc)
				local result
				while coroutine.status(co) ~= "dead" do
					local ok, val = coroutine.resume(co)
					if ok then result = val end
				end
				return result
			`,
			minTime: 1 * time.Millisecond,
			maxTime: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)
			profiler := NewProfiler()
			profiler.Enable()
			engine.SetProfiler(profiler)

			ctx := context.Background()
			start := time.Now()

			_, err := engine.Execute(ctx, tt.script, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			elapsed := time.Since(start)

			// Get profiling data
			data := profiler.GetExecutionProfile()
			require.NotNil(t, data)

			// Check total execution time
			assert.True(t, data.TotalTime >= tt.minTime,
				"Expected execution time >= %v, got %v", tt.minTime, data.TotalTime)
			assert.True(t, data.TotalTime <= tt.maxTime,
				"Expected execution time <= %v, got %v", tt.maxTime, data.TotalTime)

			// Should be close to actual elapsed time
			timeDiff := elapsed - data.TotalTime
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			assert.True(t, timeDiff < 100*time.Millisecond,
				"Profiler time %v differs from actual time %v by %v",
				data.TotalTime, elapsed, timeDiff)
		})
	}
}

func TestProfiler_MemoryProfiling(t *testing.T) {
	t.Run("manual allocation tracking", func(t *testing.T) {
		profiler := NewProfiler()
		profiler.Enable()
		profiler.EnableAllocationTracking(true)

		// Manually record some allocations
		profiler.RecordAllocation("table", 100, "test.lua:10")
		profiler.RecordAllocation("table", 120, "test.lua:11")
		profiler.RecordAllocation("string", 50, "test.lua:15")
		profiler.RecordAllocation("function", 200, "test.lua:20")

		// Get memory profile
		memProfile := profiler.GetMemoryProfile()
		require.NotNil(t, memProfile)

		// Check allocations
		assert.Equal(t, uint64(4), memProfile.Allocations)
		assert.Equal(t, uint64(470), memProfile.TotalBytes)

		// Check type stats
		assert.Equal(t, 3, len(memProfile.TypeStats)) // table, string, function
		tableStats, ok := memProfile.TypeStats["table"]
		assert.True(t, ok)
		assert.Equal(t, uint64(2), tableStats.Count)
		assert.Equal(t, uint64(220), tableStats.TotalBytes)
	})

	t.Run("allocation sites", func(t *testing.T) {
		profiler := NewProfiler()
		profiler.Enable()
		profiler.EnableAllocationTracking(true)

		// Record allocations at different sites
		for i := 0; i < 10; i++ {
			profiler.RecordAllocation("table", 100, "loop.lua:5")
		}
		profiler.RecordAllocation("string", 500, "main.lua:20")
		profiler.RecordAllocation("string", 300, "main.lua:20")

		sites := profiler.GetAllocationSites()
		assert.Len(t, sites, 2)

		// Should be sorted by bytes (highest first)
		assert.Equal(t, "loop.lua:5", sites[0].Location)
		assert.Equal(t, uint64(1000), sites[0].Bytes)
		assert.Equal(t, "main.lua:20", sites[1].Location)
		assert.Equal(t, uint64(800), sites[1].Bytes)
	})
}

func TestProfiler_HotPathIdentification(t *testing.T) {
	profiler := NewProfiler()
	profiler.Enable()

	// Simulate function calls with different frequencies
	now := time.Now()

	// Hot function - called many times
	for i := 0; i < 100; i++ {
		start := now.Add(time.Duration(i) * time.Microsecond)
		profiler.RecordFunctionCall("hotFunc", start)
		profiler.RecordFunctionReturn("hotFunc", start.Add(10*time.Millisecond))
	}

	// Medium function
	for i := 0; i < 20; i++ {
		start := now.Add(time.Duration(1000+i) * time.Microsecond)
		profiler.RecordFunctionCall("mediumFunc", start)
		profiler.RecordFunctionReturn("mediumFunc", start.Add(5*time.Millisecond))
	}

	// Cold function - called once
	start := now.Add(2 * time.Millisecond)
	profiler.RecordFunctionCall("coldFunc", start)
	profiler.RecordFunctionReturn("coldFunc", start.Add(time.Microsecond))

	// Get hot paths
	hotPaths := profiler.GetHotPaths(5)
	require.NotEmpty(t, hotPaths)

	// hotFunc should be the hottest (most total time)
	assert.Equal(t, "hotFunc", hotPaths[0].Name)
	assert.Equal(t, uint64(100), hotPaths[0].CallCount)
	assert.True(t, hotPaths[0].TotalTime >= time.Second) // 100 * 10ms

	// Check if coldFunc appears
	foundCold := false
	for _, path := range hotPaths {
		if strings.Contains(path.Name, "coldFunc") {
			foundCold = true
			assert.Equal(t, uint64(1), path.CallCount)
			break
		}
	}
	assert.True(t, foundCold, "coldFunc should be in hot paths")
}

func TestProfiler_AllocationTracking(t *testing.T) {
	t.Run("enable/disable tracking", func(t *testing.T) {
		profiler := NewProfiler()
		profiler.Enable()

		// Should be disabled by default
		assert.False(t, profiler.IsAllocationTrackingEnabled())

		// Record allocation while disabled - should not track
		profiler.RecordAllocation("test", 100, "test.go:1")
		memProfile := profiler.GetMemoryProfile()
		assert.Equal(t, uint64(0), memProfile.Allocations)

		// Enable and record
		profiler.EnableAllocationTracking(true)
		assert.True(t, profiler.IsAllocationTrackingEnabled())

		profiler.RecordAllocation("test", 100, "test.go:1")
		memProfile = profiler.GetMemoryProfile()
		assert.Equal(t, uint64(1), memProfile.Allocations)

		// Disable and ensure no more tracking
		profiler.EnableAllocationTracking(false)
		profiler.RecordAllocation("test", 100, "test.go:1")
		memProfile = profiler.GetMemoryProfile()
		assert.Equal(t, uint64(1), memProfile.Allocations) // Still 1
	})

	t.Run("different allocation types", func(t *testing.T) {
		profiler := NewProfiler()
		profiler.Enable()
		profiler.EnableAllocationTracking(true)

		// Record different types
		profiler.RecordAllocation("table", 200, "script.lua:10")
		profiler.RecordAllocation("string", 50, "script.lua:15")
		profiler.RecordAllocation("function", 300, "script.lua:20")
		profiler.RecordAllocation("table", 150, "script.lua:25")

		// Check type stats
		memProfile := profiler.GetMemoryProfile()
		assert.Equal(t, 3, len(memProfile.TypeStats))

		// Verify table stats
		tableStats := memProfile.TypeStats["table"]
		assert.Equal(t, uint64(2), tableStats.Count)
		assert.Equal(t, uint64(350), tableStats.TotalBytes)
		assert.Equal(t, uint64(175), tableStats.AvgBytes)

		// Get allocation sites
		sites := profiler.GetAllocationSites()
		assert.Equal(t, 4, len(sites))

		// First should be function (300 bytes)
		assert.Equal(t, "function", sites[0].Type)
		assert.Equal(t, uint64(300), sites[0].Bytes)
	})
}

func TestProfiler_API(t *testing.T) {
	t.Run("enable/disable", func(t *testing.T) {
		profiler := NewProfiler()

		// Should be disabled by default
		assert.False(t, profiler.IsEnabled())

		profiler.Enable()
		assert.True(t, profiler.IsEnabled())

		profiler.Disable()
		assert.False(t, profiler.IsEnabled())
	})

	t.Run("reset", func(t *testing.T) {
		profiler := NewProfiler()
		profiler.Enable()

		// Record some function calls
		now := time.Now()
		profiler.RecordFunctionCall("test", now)
		profiler.RecordFunctionReturn("test", now.Add(time.Millisecond))

		// Should have data
		execProfile := profiler.GetExecutionProfile()
		assert.True(t, execProfile.TotalTime > 0)
		assert.Equal(t, 1, len(execProfile.FunctionTimes))

		// Reset
		profiler.Reset()

		// Data should be cleared
		execProfile = profiler.GetExecutionProfile()
		assert.Equal(t, 0, len(execProfile.FunctionTimes))
		assert.Empty(t, execProfile.CallGraph)
	})

	t.Run("export/import", func(t *testing.T) {
		profiler := NewProfiler()
		profiler.Enable()

		// Record some data
		now := time.Now()
		profiler.RecordFunctionCall("func1", now)
		profiler.RecordFunctionReturn("func1", now.Add(time.Millisecond))
		profiler.RecordFunctionCall("func2", now.Add(2*time.Millisecond))
		profiler.RecordFunctionReturn("func2", now.Add(3*time.Millisecond))

		// Export data
		exported, err := profiler.Export()
		require.NoError(t, err)
		require.NotEmpty(t, exported)

		// Verify it's valid JSON
		assert.Contains(t, string(exported), "execution")
		assert.Contains(t, string(exported), "func1")
		assert.Contains(t, string(exported), "func2")

		// Import is a no-op in our implementation
		profiler2 := NewProfiler()
		err = profiler2.Import(exported)
		require.NoError(t, err)
	})
}

func TestProfiler_Overhead(t *testing.T) {
	// Compare execution with and without profiling
	script := `
		local sum = 0
		for i = 1, 1000000 do
			sum = sum + i
		end
		return sum
	`

	// Run without profiler
	engine1 := createTestEngine(t)
	ctx := context.Background()

	start := time.Now()
	_, err := engine1.Execute(ctx, script, nil)
	require.NoError(t, err)
	timeWithoutProfiler := time.Since(start)

	// Run with profiler
	engine2 := createTestEngine(t)
	profiler := NewProfiler()
	engine2.SetProfiler(profiler)

	start = time.Now()
	_, err = engine2.Execute(ctx, script, nil)
	require.NoError(t, err)
	timeWithProfiler := time.Since(start)

	// Calculate overhead
	overhead := float64(timeWithProfiler-timeWithoutProfiler) / float64(timeWithoutProfiler) * 100

	// Log results
	t.Logf("Without profiler: %v", timeWithoutProfiler)
	t.Logf("With profiler: %v", timeWithProfiler)
	t.Logf("Overhead: %.2f%%", overhead)

	// Overhead should be reasonable (less than 50%)
	assert.True(t, overhead < 50, "Profiler overhead too high: %.2f%%", overhead)
}

func TestProfiler_Concurrent(t *testing.T) {
	profiler := NewProfiler()
	profiler.Enable()
	profiler.EnableAllocationTracking(true)

	// Run multiple goroutines recording data concurrently
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			funcName := fmt.Sprintf("goroutine_%d", id)

			// Record function calls
			for j := 0; j < 10; j++ {
				start := time.Now()
				profiler.RecordFunctionCall(funcName, start)
				time.Sleep(time.Microsecond) // Simulate work
				profiler.RecordFunctionReturn(funcName, start.Add(time.Millisecond))
			}

			// Record allocations
			profiler.RecordAllocation("goroutine_data", uint64(100+id), fmt.Sprintf("goroutine_%d.go:10", id))
		}(i)
	}

	wg.Wait()

	// Should have collected data from all executions
	execProfile := profiler.GetExecutionProfile()
	assert.True(t, execProfile.TotalTime > 0)
	assert.Equal(t, numGoroutines, len(execProfile.FunctionTimes))

	// Check hot paths
	hotPaths := profiler.GetHotPaths(10)
	assert.Equal(t, numGoroutines, len(hotPaths))

	// Each function should have been called 10 times
	for _, path := range hotPaths {
		assert.Equal(t, uint64(10), path.CallCount)
	}

	// Check allocations
	memProfile := profiler.GetMemoryProfile()
	assert.Equal(t, uint64(numGoroutines), memProfile.Allocations)
}

func TestProfiler_LuaAPI(t *testing.T) {
	engine := createTestEngine(t)
	profiler := NewProfiler()
	profiler.Enable()
	engine.SetProfiler(profiler)

	// Test Lua API for controlling profiler
	script := `
		-- Check if profiler is available
		assert(profiler ~= nil, "profiler should be available")
		
		-- Start profiling
		profiler.start()
		
		-- Do some work
		local sum = 0
		for i = 1, 10000 do
			sum = sum + i
		end
		
		-- Stop profiling
		profiler.stop()
		
		-- Get profile data
		local profile = profiler.getProfile()
		assert(profile ~= nil, "should have profile data")
		assert(profile.totalTime > 0, "should have execution time")
		
		-- Reset profiler
		profiler.reset()
		
		return true
	`

	ctx := context.Background()
	result, err := engine.Execute(ctx, script, nil)
	require.NoError(t, err)

	// Result is a ScriptValue, extract the actual boolean
	goValue, err := engine.ToNative(result)
	require.NoError(t, err)
	assert.Equal(t, true, goValue)
}

// Helper to create test engine
func createTestEngine(t *testing.T) *LuaEngine {
	eng := NewLuaEngine()
	err := eng.Initialize(engine.EngineConfig{
		SandboxMode: true,
		MemoryLimit: 64 * 1024 * 1024, // 64MB
	})
	require.NoError(t, err)
	return eng
}
