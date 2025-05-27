// ABOUTME: Tests for secure execution contexts and resource management
// ABOUTME: Validates resource limits, timeouts, and context propagation

package security

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSecureContext(t *testing.T) {
	t.Run("create secure context", func(t *testing.T) {
		ctx := context.Background()
		config := ContextConfig{
			MaxMemory:        64 * 1024 * 1024, // 64MB
			MaxCPUTime:       5 * time.Second,
			MaxExecutionTime: 10 * time.Second,
			MaxGoroutines:    100,
		}

		secCtx, err := NewSecureContext(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create secure context: %v", err)
		}

		if secCtx == nil {
			t.Fatal("Secure context should not be nil")
		}

		// Should be able to get config back
		retrievedConfig := GetContextConfig(secCtx)
		if retrievedConfig == nil {
			t.Fatal("Should be able to retrieve config from context")
		}

		if retrievedConfig.MaxMemory != config.MaxMemory {
			t.Error("Config values should match")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		ctx := context.Background()

		tests := []struct {
			name   string
			config ContextConfig
		}{
			{
				name: "negative memory limit",
				config: ContextConfig{
					MaxMemory:        -1,
					MaxExecutionTime: 10 * time.Second,
				},
			},
			{
				name: "zero execution time",
				config: ContextConfig{
					MaxMemory:        64 * 1024 * 1024,
					MaxExecutionTime: 0,
				},
			},
			{
				name: "negative goroutine limit",
				config: ContextConfig{
					MaxMemory:        64 * 1024 * 1024,
					MaxExecutionTime: 10 * time.Second,
					MaxGoroutines:    -1,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewSecureContext(ctx, tt.config)
				if err == nil {
					t.Error("Expected error for invalid config")
				}
			})
		}
	})
}

func TestResourceTracker(t *testing.T) {
	t.Run("memory tracking", func(t *testing.T) {
		tracker := NewResourceTracker(ResourceLimits{
			MaxMemory:     100 * 1024, // 100KB
			MaxCPUTime:    time.Minute,
			MaxGoroutines: 10,
		})

		// Allocate some memory
		size := int64(50 * 1024) // 50KB
		if err := tracker.AllocateMemory(size); err != nil {
			t.Fatalf("Failed to allocate memory: %v", err)
		}

		// Check usage
		usage := tracker.GetMemoryUsage()
		if usage != size {
			t.Errorf("Expected memory usage %d, got %d", size, usage)
		}

		// Try to allocate more than limit
		if err := tracker.AllocateMemory(60 * 1024); err == nil {
			t.Error("Expected error when exceeding memory limit")
		}

		// Free some memory
		tracker.FreeMemory(30 * 1024)
		usage = tracker.GetMemoryUsage()
		if usage != 20*1024 {
			t.Errorf("Expected memory usage 20KB after free, got %d", usage)
		}

		// Now allocation should succeed
		if err := tracker.AllocateMemory(60 * 1024); err != nil {
			t.Errorf("Allocation should succeed after freeing memory: %v", err)
		}
	})

	t.Run("goroutine tracking", func(t *testing.T) {
		tracker := NewResourceTracker(ResourceLimits{
			MaxMemory:     100 * 1024 * 1024,
			MaxCPUTime:    time.Minute,
			MaxGoroutines: 5,
		})

		// Start some goroutines
		for i := 0; i < 3; i++ {
			if err := tracker.StartGoroutine(); err != nil {
				t.Fatalf("Failed to start goroutine %d: %v", i, err)
			}
		}

		// Check count
		count := tracker.GetGoroutineCount()
		if count != 3 {
			t.Errorf("Expected 3 goroutines, got %d", count)
		}

		// Try to exceed limit
		for i := 0; i < 3; i++ {
			err := tracker.StartGoroutine()
			if i < 2 && err != nil {
				t.Errorf("Should be able to start goroutine %d: %v", i+3, err)
			}
			if i == 2 && err == nil {
				t.Error("Expected error when exceeding goroutine limit")
			}
		}

		// End a goroutine
		tracker.EndGoroutine()
		count = tracker.GetGoroutineCount()
		if count != 4 {
			t.Errorf("Expected 4 goroutines after ending one, got %d", count)
		}
	})

	t.Run("cpu time tracking", func(t *testing.T) {
		tracker := NewResourceTracker(ResourceLimits{
			MaxMemory:     100 * 1024 * 1024,
			MaxCPUTime:    100 * time.Millisecond,
			MaxGoroutines: 10,
		})

		// Start CPU tracking
		tracker.StartCPUTracking()

		// Simulate some CPU usage
		start := time.Now()
		for time.Since(start) < 50*time.Millisecond {
			// Busy loop
			_ = 1 + 1
		}

		// Update and check CPU time
		tracker.UpdateCPUTime()
		cpuTime := tracker.GetCPUTime()
		if cpuTime < 40*time.Millisecond {
			t.Errorf("CPU time should be at least 40ms, got %v", cpuTime)
		}

		// Should still be within limit
		if err := tracker.CheckCPULimit(); err != nil {
			t.Errorf("Should be within CPU limit: %v", err)
		}

		// Simulate more CPU usage to exceed limit
		start = time.Now()
		for time.Since(start) < 60*time.Millisecond {
			_ = 1 + 1
		}

		tracker.UpdateCPUTime()
		if err := tracker.CheckCPULimit(); err == nil {
			t.Error("Expected error when exceeding CPU limit")
		}
	})
}

func TestTimeoutEnforcement(t *testing.T) {
	t.Run("execution timeout", func(t *testing.T) {
		ctx := context.Background()
		config := ContextConfig{
			MaxMemory:        64 * 1024 * 1024,
			MaxExecutionTime: 100 * time.Millisecond,
		}

		secCtx, err := NewSecureContext(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create secure context: %v", err)
		}

		// Context should timeout
		select {
		case <-time.After(200 * time.Millisecond):
			t.Error("Context should have timed out")
		case <-secCtx.Done():
			// Expected
		}

		// Check error
		if secCtx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", secCtx.Err())
		}
	})

	t.Run("parent context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		config := ContextConfig{
			MaxMemory:        64 * 1024 * 1024,
			MaxExecutionTime: 10 * time.Second,
		}

		secCtx, err := NewSecureContext(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create secure context: %v", err)
		}

		// Cancel parent
		cancel()

		// Secure context should also be cancelled
		select {
		case <-time.After(100 * time.Millisecond):
			t.Error("Secure context should be cancelled when parent is cancelled")
		case <-secCtx.Done():
			// Expected
		}

		if secCtx.Err() != context.Canceled {
			t.Errorf("Expected Canceled, got %v", secCtx.Err())
		}
	})
}

func TestResourceMonitor(t *testing.T) {
	t.Run("periodic monitoring", func(t *testing.T) {
		ctx := context.Background()
		config := ContextConfig{
			MaxMemory:        64 * 1024 * 1024,
			MaxExecutionTime: 5 * time.Second,
			MaxCPUTime:       1 * time.Second,
			MaxGoroutines:    10,
		}

		secCtx, err := NewSecureContext(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create secure context: %v", err)
		}

		// Start monitoring
		monitor := StartResourceMonitor(secCtx, 50*time.Millisecond)
		defer monitor.Stop()

		// Simulate some resource usage
		tracker := GetResourceTracker(secCtx)
		if tracker == nil {
			t.Fatal("Should have resource tracker")
		}

		// Allocate memory to exceed 90% threshold
		if err := tracker.AllocateMemory(60 * 1024 * 1024); err != nil {
			t.Fatalf("Failed to allocate memory: %v", err)
		}

		// Wait for monitor to detect
		time.Sleep(150 * time.Millisecond)

		// Check if monitoring detected high usage
		violations := monitor.GetViolations()
		if len(violations) == 0 {
			t.Error("Monitor should have detected resource violations")
		}

		// Should have memory violation
		foundMemoryViolation := false
		for _, v := range violations {
			if v.Type == "memory" {
				foundMemoryViolation = true
				break
			}
		}
		if !foundMemoryViolation {
			t.Error("Should have detected memory violation")
		}
	})
}

func TestConcurrentResourceAccess(t *testing.T) {
	tracker := NewResourceTracker(ResourceLimits{
		MaxMemory:     100 * 1024 * 1024, // 100MB
		MaxCPUTime:    time.Minute,
		MaxGoroutines: 100,
	})

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Concurrent memory allocations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			size := int64(1024 * 1024) // 1MB
			_ = tracker.AllocateMemory(size)
			time.Sleep(10 * time.Millisecond)
			tracker.FreeMemory(size)
		}()
	}

	// Concurrent goroutine tracking
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = tracker.StartGoroutine()
			time.Sleep(10 * time.Millisecond)
			tracker.EndGoroutine()
		}()
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = tracker.GetMemoryUsage()
			_ = tracker.GetGoroutineCount()
			_ = tracker.GetCPUTime()
		}()
	}

	wg.Wait()

	// Final state should be clean
	if tracker.GetMemoryUsage() != 0 {
		t.Error("Memory usage should be 0 after all operations")
	}
	if tracker.GetGoroutineCount() != 0 {
		t.Error("Goroutine count should be 0 after all operations")
	}
}

func TestSecurityPolicies(t *testing.T) {
	t.Run("apply security policy", func(t *testing.T) {
		ctx := context.Background()
		config := ContextConfig{
			MaxMemory:        64 * 1024 * 1024,
			MaxExecutionTime: 10 * time.Second,
			SecurityPolicy: &SecurityPolicy{
				AllowNetworkAccess: false,
				AllowFileWrite:     false,
				AllowFileRead:      true,
				AllowedPaths:       []string{"/tmp", "/var/tmp"},
				BlockedPaths:       []string{"/etc", "/usr/bin"},
			},
		}

		secCtx, err := NewSecureContext(ctx, config)
		if err != nil {
			t.Fatalf("Failed to create secure context: %v", err)
		}

		policy := GetSecurityPolicy(secCtx)
		if policy == nil {
			t.Fatal("Should have security policy")
		}

		// Test network access
		if policy.AllowNetworkAccess {
			t.Error("Network access should be disabled")
		}

		// Test path validation
		tests := []struct {
			path    string
			allowed bool
		}{
			{"/tmp/test.txt", true},
			{"/var/tmp/data", true},
			{"/etc/passwd", false},
			{"/usr/bin/ls", false},
			{"/home/user/file", false}, // Not in allowed paths
		}

		for _, tt := range tests {
			allowed := policy.IsPathAllowed(tt.path)
			if allowed != tt.allowed {
				t.Errorf("Path %s: expected allowed=%v, got %v", tt.path, tt.allowed, allowed)
			}
		}
	})
}

func BenchmarkResourceTracking(b *testing.B) {
	tracker := NewResourceTracker(ResourceLimits{
		MaxMemory:     1024 * 1024 * 1024, // 1GB
		MaxCPUTime:    time.Hour,
		MaxGoroutines: 1000,
	})

	b.Run("memory allocation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tracker.AllocateMemory(1024)
			tracker.FreeMemory(1024)
		}
	})

	b.Run("goroutine tracking", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tracker.StartGoroutine()
			tracker.EndGoroutine()
		}
	})
}
