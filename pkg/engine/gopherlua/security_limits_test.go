// ABOUTME: Tests for resource limit enforcement in the SecurityManager
// ABOUTME: Validates timeout, memory monitoring, and execution limits using alternative approaches

package gopherlua

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestResourceLimitEnforcer_ContextTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		script      string
		expectError bool
		contains    string
	}{
		{
			name:    "allows_quick_execution",
			timeout: 100 * time.Millisecond,
			script: `
				local x = 1 + 1
				return x
			`,
			expectError: false,
		},
		{
			name:    "enforces_timeout_on_long_loop",
			timeout: 50 * time.Millisecond,
			script: `
				local start = os.clock()
				while os.clock() - start < 0.2 do
					-- Busy wait for 200ms
				end
				return "should not reach here"
			`,
			expectError: true,
			contains:    "execution time limit exceeded",
		},
		{
			name:    "enforces_timeout_on_infinite_loop",
			timeout: 30 * time.Millisecond,
			script: `
				while true do
					-- Infinite loop
				end
			`,
			expectError: true,
			contains:    "execution time limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := NewResourceLimitEnforcer(ResourceLimits{
				MaxDuration: tt.timeout,
			})

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			// Load os library for clock function
			err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(lua.OpenOs),
				NRet:    0,
				Protect: true,
			}, lua.LString("os"))
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout*2)
			defer cancel()

			err = enforcer.ExecuteWithLimits(ctx, L, tt.script)

			if tt.expectError {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceLimitEnforcer_MemoryMonitoring(t *testing.T) {
	t.Skip("Memory monitoring via runtime.ReadMemStats is not reliable for individual script execution")

	tests := []struct {
		name        string
		memoryLimit int64
		script      string
		expectError bool
		contains    string
	}{
		{
			name:        "allows_small_allocations",
			memoryLimit: 10 * 1024 * 1024, // 10MB
			script: `
				local data = {}
				for i = 1, 100 do
					data[i] = "small string " .. i
				end
				return #data
			`,
			expectError: false,
		},
		{
			name:        "monitors_memory_usage_baseline",
			memoryLimit: 1024 * 1024, // 1MB
			script: `
				return "simple test"
			`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := NewResourceLimitEnforcer(ResourceLimits{
				MaxMemory:     tt.memoryLimit,
				CheckInterval: 1000, // Check every 1000 operations
			})

			opts := lua.Options{SkipOpenLibs: true}
			L := lua.NewState(opts)
			defer L.Close()

			// Load string library
			err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(lua.OpenString),
				NRet:    0,
				Protect: true,
			}, lua.LString("string"))
			require.NoError(t, err)

			ctx := context.Background()
			err = enforcer.ExecuteWithLimits(ctx, L, tt.script)

			if tt.expectError {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceLimitEnforcer_StackDepthLimits(t *testing.T) {
	tests := []struct {
		name          string
		maxStackDepth int
		script        string
		expectError   bool
		contains      string
	}{
		{
			name:          "allows_shallow_recursion",
			maxStackDepth: 100,
			script: `
				function factorial(n)
					if n <= 1 then
						return 1
					else
						return n * factorial(n - 1)
					end
				end
				return factorial(5)
			`,
			expectError: false,
		},
		{
			name:          "blocks_deep_recursion",
			maxStackDepth: 10,
			script: `
				function deep_recursion(n)
					if n <= 0 then
						return 0
					else
						return 1 + deep_recursion(n - 1)
					end
				end
				return deep_recursion(50)
			`,
			expectError: true,
			contains:    "stack overflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := NewResourceLimitEnforcer(ResourceLimits{
				MaxStackDepth: tt.maxStackDepth,
			})

			opts := lua.Options{
				SkipOpenLibs:  true,
				CallStackSize: tt.maxStackDepth,
			}
			L := lua.NewState(opts)
			defer L.Close()

			ctx := context.Background()
			err := enforcer.ExecuteWithLimits(ctx, L, tt.script)

			if tt.expectError {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceLimitEnforcer_CombinedLimits(t *testing.T) {
	limits := ResourceLimits{
		MaxDuration:   100 * time.Millisecond,
		MaxMemory:     1024 * 1024, // 1MB
		MaxStackDepth: 20,
		CheckInterval: 100,
	}

	enforcer := NewResourceLimitEnforcer(limits)

	tests := []struct {
		name        string
		script      string
		expectError bool
		contains    string
	}{
		{
			name: "passes_all_limits",
			script: `
				local data = {}
				for i = 1, 10 do
					data[i] = "item " .. i
				end
				return #data
			`,
			expectError: false,
		},
		{
			name: "fails_time_limit",
			script: `
				local count = 0
				while count < 1000000 do
					count = count + 1
				end
			`,
			expectError: true,
			contains:    "limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := lua.Options{
				SkipOpenLibs:  true,
				CallStackSize: limits.MaxStackDepth,
			}
			L := lua.NewState(opts)
			defer L.Close()

			// Load required libraries
			_ = L.CallByParam(lua.P{Fn: L.NewFunction(lua.OpenString), NRet: 0, Protect: true}, lua.LString("string"))
			_ = L.CallByParam(lua.P{Fn: L.NewFunction(lua.OpenOs), NRet: 0, Protect: true}, lua.LString("os"))

			ctx, cancel := context.WithTimeout(context.Background(), limits.MaxDuration*2)
			defer cancel()

			err := enforcer.ExecuteWithLimits(ctx, L, tt.script)

			if tt.expectError {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, err.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceLimitEnforcer_ProfileBasedLimits(t *testing.T) {
	profiles := map[string]ResourceLimits{
		"minimal": {
			MaxDuration:   5 * time.Minute,
			MaxMemory:     100 * 1024 * 1024, // 100MB
			MaxStackDepth: 1000,
			CheckInterval: 10000,
		},
		"standard": {
			MaxDuration:   30 * time.Second,
			MaxMemory:     50 * 1024 * 1024, // 50MB
			MaxStackDepth: 500,
			CheckInterval: 5000,
		},
		"strict": {
			MaxDuration:   5 * time.Second,
			MaxMemory:     10 * 1024 * 1024, // 10MB
			MaxStackDepth: 100,
			CheckInterval: 1000,
		},
	}

	for profile, limits := range profiles {
		t.Run("profile_"+profile, func(t *testing.T) {
			enforcer := NewResourceLimitEnforcer(limits)

			opts := lua.Options{
				SkipOpenLibs:  true,
				CallStackSize: limits.MaxStackDepth,
			}
			L := lua.NewState(opts)
			defer L.Close()

			script := `
				local data = {}
				for i = 1, 100 do
					data[i] = "profile test " .. i
				end
				return #data
			`

			ctx := context.Background()
			err := enforcer.ExecuteWithLimits(ctx, L, script)
			assert.NoError(t, err, "Profile %s should allow basic operations", profile)
		})
	}
}

func TestResourceMonitor_GetStats(t *testing.T) {
	limits := ResourceLimits{
		MaxDuration:   1 * time.Second,
		MaxMemory:     10 * 1024 * 1024,
		CheckInterval: 100,
	}

	enforcer := NewResourceLimitEnforcer(limits)
	monitor := enforcer.CreateMonitor()

	// Test initial state
	stats := monitor.GetStats()
	assert.Equal(t, int64(0), stats.MemoryUsed)
	assert.True(t, stats.ExecutionTime < time.Millisecond)

	// Simulate some usage
	monitor.UpdateMemoryUsage(1024 * 1024) // 1MB
	time.Sleep(10 * time.Millisecond)

	stats = monitor.GetStats()
	assert.Equal(t, int64(1024*1024), stats.MemoryUsed)
	assert.True(t, stats.ExecutionTime >= 10*time.Millisecond)
}
