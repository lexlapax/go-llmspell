// ABOUTME: Tests for async runtime functionality in GopherLua engine
// ABOUTME: Tests coroutine management, promise integration, and async execution contexts

package gopherlua

import (
	"context"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestAsyncRuntime_NewAsyncRuntime(t *testing.T) {
	tests := []struct {
		name     string
		maxCoros int
		wantErr  bool
	}{
		{
			name:     "valid runtime with default max coroutines",
			maxCoros: 0, // Should use default
			wantErr:  false,
		},
		{
			name:     "valid runtime with custom max coroutines",
			maxCoros: 10,
			wantErr:  false,
		},
		{
			name:     "invalid runtime with negative max coroutines",
			maxCoros: -1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime, err := NewAsyncRuntime(tt.maxCoros)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAsyncRuntime() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewAsyncRuntime() unexpected error: %v", err)
				return
			}

			if runtime == nil {
				t.Errorf("NewAsyncRuntime() returned nil runtime")
				return
			}

			// Verify runtime state
			if runtime.maxCoroutines <= 0 {
				t.Errorf("NewAsyncRuntime() maxCoroutines should be positive, got %d", runtime.maxCoroutines)
			}
		})
	}
}

func TestAsyncRuntime_SpawnCoroutine(t *testing.T) {
	runtime, err := NewAsyncRuntime(5)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name    string
		script  string
		args    []lua.LValue
		wantErr bool
	}{
		{
			name:    "simple coroutine",
			script:  `return "hello from coroutine"`,
			args:    []lua.LValue{},
			wantErr: false,
		},
		{
			name:    "coroutine with arguments",
			script:  `local arg = ... return "received: " .. tostring(arg)`,
			args:    []lua.LValue{lua.LString("test")},
			wantErr: false,
		},
		{
			name:    "coroutine with error",
			script:  `error("test error")`,
			args:    []lua.LValue{},
			wantErr: false, // SpawnCoroutine should succeed, error captured in result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			coroID, err := runtime.SpawnCoroutine(ctx, L, tt.script, tt.args...)
			if err != nil {
				t.Errorf("SpawnCoroutine() unexpected error: %v", err)
				return
			}

			if coroID == "" {
				t.Errorf("SpawnCoroutine() returned empty coroutine ID")
				return
			}

			// Wait for coroutine to complete
			result, err := runtime.WaitForCoroutine(ctx, coroID)

			// For the error test case, we expect an error from WaitForCoroutine
			if tt.name == "coroutine with error" {
				if err == nil {
					t.Errorf("WaitForCoroutine() expected error for error script, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("WaitForCoroutine() error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("WaitForCoroutine() returned nil result")
			}
		})
	}
}

func TestAsyncRuntime_CoroutineManagement(t *testing.T) {
	runtime, err := NewAsyncRuntime(3)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	// Test spawning multiple coroutines
	script := `
		local start = os.clock()
		while os.clock() - start < 0.1 do end
		return "done"
	`

	var coroIDs []string
	for i := 0; i < 3; i++ {
		coroID, err := runtime.SpawnCoroutine(ctx, L, script)
		if err != nil {
			t.Errorf("Failed to spawn coroutine %d: %v", i, err)
			continue
		}
		coroIDs = append(coroIDs, coroID)
	}

	// Test that we can't exceed max coroutines
	_, err = runtime.SpawnCoroutine(ctx, L, script)
	if err == nil {
		t.Errorf("Expected error when exceeding max coroutines, got nil")
	}

	// Wait for all coroutines to complete
	for i, coroID := range coroIDs {
		_, err := runtime.WaitForCoroutine(ctx, coroID)
		if err != nil {
			t.Errorf("Failed to wait for coroutine %d: %v", i, err)
		}
	}

	// Should be able to spawn again after coroutines complete
	_, err = runtime.SpawnCoroutine(ctx, L, `return "new coroutine"`)
	if err != nil {
		t.Errorf("Failed to spawn coroutine after others completed: %v", err)
	}
}

func TestAsyncRuntime_CancellationSupport(t *testing.T) {
	runtime, err := NewAsyncRuntime(5)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())

	longScript := `
		local start = os.clock()
		while os.clock() - start < 10 do end  -- Long running
		return "should not complete"
	`

	coroID, err := runtime.SpawnCoroutine(ctx, L, longScript)
	if err != nil {
		t.Fatalf("Failed to spawn coroutine: %v", err)
	}

	// Cancel after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Should return context cancelled error
	_, err = runtime.WaitForCoroutine(ctx, coroID)
	if err == nil {
		t.Errorf("Expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestAsyncRuntime_TimeoutHandling(t *testing.T) {
	runtime, err := NewAsyncRuntime(5)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	longScript := `
		local start = os.clock()
		while os.clock() - start < 1 do end  -- Longer than timeout
		return "should timeout"
	`

	coroID, err := runtime.SpawnCoroutine(ctx, L, longScript)
	if err != nil {
		t.Fatalf("Failed to spawn coroutine: %v", err)
	}

	// Should timeout
	_, err = runtime.WaitForCoroutine(ctx, coroID)
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestAsyncRuntime_PromiseIntegration(t *testing.T) {
	runtime, err := NewAsyncRuntime(5)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	// Test creating a promise from coroutine
	script := `return 42`

	promise, err := runtime.CreatePromise(ctx, L, script)
	if err != nil {
		t.Fatalf("Failed to create promise: %v", err)
	}

	if promise == nil {
		t.Fatalf("CreatePromise returned nil promise")
	}

	// Promise should have methods
	if promise.coroID == "" {
		t.Errorf("Promise should have coroutine ID")
	}
}

func TestAsyncRuntime_AsyncExecutionContext(t *testing.T) {
	runtime, err := NewAsyncRuntime(5)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	// Test execution context creation
	execCtx, err := runtime.CreateExecutionContext(ctx, L)
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}

	if execCtx == nil {
		t.Fatalf("CreateExecutionContext returned nil context")
	}

	// Execution context should have required fields
	if execCtx.ID == "" {
		t.Errorf("Execution context should have ID")
	}

	if execCtx.StartTime.IsZero() {
		t.Errorf("Execution context should have start time")
	}
}

func TestAsyncRuntime_Close(t *testing.T) {
	runtime, err := NewAsyncRuntime(5)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	// Spawn some coroutines
	for i := 0; i < 3; i++ {
		_, err := runtime.SpawnCoroutine(ctx, L, `return "test"`)
		if err != nil {
			t.Errorf("Failed to spawn coroutine %d: %v", i, err)
		}
	}

	// Close should succeed and clean up
	err = runtime.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Should not be able to spawn new coroutines after close
	_, err = runtime.SpawnCoroutine(ctx, L, `return "should fail"`)
	if err == nil {
		t.Errorf("Expected error after Close(), got nil")
	}
}
