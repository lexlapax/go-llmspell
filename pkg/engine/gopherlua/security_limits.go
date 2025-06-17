// ABOUTME: ResourceLimitEnforcer implements resource limits for Lua script execution
// ABOUTME: Uses context timeouts, memory monitoring, and stack limits since SetHook is unavailable

package gopherlua

import (
	"context"
	"fmt"
	"runtime"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ResourceLimitEnforcer manages resource limits for Lua execution
type ResourceLimitEnforcer struct {
	limits ResourceLimits
}

// ResourceStats tracks current resource usage
type ResourceStats struct {
	MemoryUsed    int64
	ExecutionTime time.Duration
	StackDepth    int
	StartTime     time.Time
}

// ResourceMonitorLimits extends ResourceMonitor for limit enforcement
// Uses the ResourceMonitor from security.go

// NewResourceLimitEnforcer creates a new resource limit enforcer
func NewResourceLimitEnforcer(limits ResourceLimits) *ResourceLimitEnforcer {
	// Apply defaults if needed
	if limits.CheckInterval == 0 {
		limits.CheckInterval = 1000
	}
	if limits.MaxStackDepth == 0 {
		limits.MaxStackDepth = 1000
	}

	return &ResourceLimitEnforcer{
		limits: limits,
	}
}

// ExecuteWithLimits executes a Lua script with resource limits enforced
func (rle *ResourceLimitEnforcer) ExecuteWithLimits(ctx context.Context, L *lua.LState, script string) error {
	// Create execution context with timeout
	execCtx := ctx
	if rle.limits.MaxDuration > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, rle.limits.MaxDuration)
		defer cancel()
	}

	// Set context on Lua state
	L.SetContext(execCtx)

	// Create monitor
	monitor := rle.CreateMonitor()

	// Pre-execution memory check
	if rle.limits.MaxMemory > 0 {
		runtime.GC() // Force GC to get accurate baseline
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		monitor.UpdateMemoryUsage(int64(m.Alloc))
	}

	// Execute with monitoring
	errChan := make(chan error, 1)
	done := make(chan struct{}) // Signal when goroutine completes

	go func() {
		defer func() {
			close(done) // Signal completion
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					errChan <- err
				} else {
					errChan <- fmt.Errorf("script panic: %v", r)
				}
			}
		}()

		err := L.DoString(script)
		errChan <- err
	}()

	// Monitor execution with periodic checks
	ticker := time.NewTicker(time.Duration(rle.limits.CheckInterval) * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-errChan:
			// Wait for goroutine to fully complete to avoid race
			<-done
			// Execution completed
			if err != nil {
				return fmt.Errorf("script execution failed: %w", err)
			}
			return nil

		case <-execCtx.Done():
			// Timeout or cancellation - wait for goroutine to finish
			// The context will cause the Lua execution to abort
			<-done
			return fmt.Errorf("execution time limit exceeded: %w", execCtx.Err())

		case <-ticker.C:
			// Periodic resource check
			if err := rle.checkResourceUsage(monitor); err != nil {
				// Wait for goroutine completion on limit exceeded
				<-done
				return err
			}
		}
	}
}

// CreateMonitor creates a new resource monitor
func (rle *ResourceLimitEnforcer) CreateMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		limits:           rle.limits,
		startTime:        time.Now(),
		memUsed:          0,
		instructionCount: 0,
	}
}

// checkResourceUsage performs periodic resource usage checks
func (rle *ResourceLimitEnforcer) checkResourceUsage(monitor *ResourceMonitor) error {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	// Check memory usage
	if rle.limits.MaxMemory > 0 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		currentMem := int64(m.Alloc)

		if currentMem > rle.limits.MaxMemory {
			return fmt.Errorf("memory limit exceeded: %d bytes > %d bytes", currentMem, rle.limits.MaxMemory)
		}

		monitor.memUsed = currentMem
	}

	// Check execution time
	if rle.limits.MaxDuration > 0 {
		elapsed := time.Since(monitor.startTime)
		if elapsed > rle.limits.MaxDuration {
			return fmt.Errorf("execution time limit exceeded: %v > %v", elapsed, rle.limits.MaxDuration)
		}
	}

	return nil
}

// UpdateMemoryUsage updates the current memory usage
func (rm *ResourceMonitor) UpdateMemoryUsage(bytes int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.memUsed = bytes
}

// GetStats returns current resource usage statistics
func (rm *ResourceMonitor) GetStats() ResourceStats {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	return ResourceStats{
		MemoryUsed:    rm.memUsed,
		ExecutionTime: time.Since(rm.startTime),
		StartTime:     rm.startTime,
	}
}

// GetMemoryUsage returns current memory usage
func (rm *ResourceMonitor) GetMemoryUsage() int64 {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.memUsed
}

// GetExecutionTime returns elapsed execution time
func (rm *ResourceMonitor) GetExecutionTime() time.Duration {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return time.Since(rm.startTime)
}

// CreateResourceLimitProfile creates predefined resource limit profiles
func CreateResourceLimitProfile(profile string) ResourceLimits {
	profiles := map[string]ResourceLimits{
		"minimal": {
			MaxInstructions: 100_000_000,
			MaxMemory:       100 * 1024 * 1024, // 100MB
			MaxDuration:     5 * time.Minute,
			MaxStackDepth:   1000,
			CheckInterval:   10000,
		},
		"standard": {
			MaxInstructions: 10_000_000,
			MaxMemory:       50 * 1024 * 1024, // 50MB
			MaxDuration:     30 * time.Second,
			MaxStackDepth:   500,
			CheckInterval:   5000,
		},
		"strict": {
			MaxInstructions: 1_000_000,
			MaxMemory:       10 * 1024 * 1024, // 10MB
			MaxDuration:     5 * time.Second,
			MaxStackDepth:   100,
			CheckInterval:   1000,
		},
	}

	if limits, ok := profiles[profile]; ok {
		return limits
	}

	// Return strict as default
	return profiles["strict"]
}

// ApplyResourceLimitsToState configures Lua state with resource limits
func ApplyResourceLimitsToState(L *lua.LState, limits ResourceLimits) {
	// Configure VM options that can be set
	// Note: Some limits must be set during lua.NewState() creation

	// We can't modify CallStackSize after creation, but we can document it
	// The stack depth limit should be set in lua.Options.CallStackSize
}

// ValidateResourceLimits checks if resource limits are reasonable
func ValidateResourceLimits(limits ResourceLimits) error {
	if limits.MaxMemory < 0 {
		return fmt.Errorf("MaxMemory cannot be negative: %d", limits.MaxMemory)
	}

	if limits.MaxDuration < 0 {
		return fmt.Errorf("MaxDuration cannot be negative: %v", limits.MaxDuration)
	}

	if limits.MaxInstructions < 0 {
		return fmt.Errorf("MaxInstructions cannot be negative: %d", limits.MaxInstructions)
	}

	if limits.MaxStackDepth <= 0 {
		return fmt.Errorf("MaxStackDepth must be positive: %d", limits.MaxStackDepth)
	}

	if limits.CheckInterval <= 0 {
		return fmt.Errorf("CheckInterval must be positive: %d", limits.CheckInterval)
	}

	// Warn about unreasonable values
	if limits.MaxMemory > 0 && limits.MaxMemory < 1024*1024 {
		// Less than 1MB might be too restrictive
		return fmt.Errorf("MaxMemory too low (< 1MB): %d bytes", limits.MaxMemory)
	}

	if limits.MaxDuration > 0 && limits.MaxDuration < time.Millisecond {
		return fmt.Errorf("MaxDuration too low (< 1ms): %v", limits.MaxDuration)
	}

	return nil
}
