// ABOUTME: Tests for async runtime functionality in GopherLua engine
// ABOUTME: Tests coroutine management, promise integration, and async execution contexts

package gopherlua

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
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

// Comprehensive Coroutine Lifecycle Tests

func TestAsyncRuntime_CoroutineLifecycle(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	t.Run("coroutine states", func(t *testing.T) {
		// Test coroutine state tracking
		script := `
			local count = 0
			for i = 1, 10 do
				count = count + i
			end
			return count
		`

		coroID, err := runtime.SpawnCoroutine(ctx, L, script)
		if err != nil {
			t.Fatalf("Failed to spawn coroutine: %v", err)
		}

		// Check if coroutine is active
		if !runtime.IsCoroutineActive(coroID) {
			t.Errorf("Coroutine should be active immediately after spawn")
		}

		// Wait for completion
		result, err := runtime.WaitForCoroutine(ctx, coroID)
		if err != nil {
			t.Errorf("WaitForCoroutine failed: %v", err)
		}

		// Check result
		if num, ok := result.(lua.LNumber); ok {
			if float64(num) != 55 {
				t.Errorf("Expected result 55, got %v", num)
			}
		} else {
			t.Errorf("Expected LNumber result, got %T", result)
		}

		// Coroutine should no longer be active
		if runtime.IsCoroutineActive(coroID) {
			t.Errorf("Coroutine should not be active after completion")
		}

		// Waiting again should return cached result
		result2, err := runtime.WaitForCoroutine(ctx, coroID)
		if err != nil {
			t.Errorf("Second WaitForCoroutine failed: %v", err)
		}
		if result != result2 {
			t.Errorf("Cached result mismatch")
		}
	})

	t.Run("coroutine error propagation", func(t *testing.T) {
		errorScripts := []struct {
			name   string
			script string
			errMsg string
		}{
			{
				name:   "runtime error",
				script: `error("custom error message")`,
				errMsg: "custom error message",
			},
			{
				name:   "syntax error",
				script: `invalid syntax here`,
				errMsg: "syntax",
			},
			{
				name:   "nil operation error",
				script: `local x = nil; return x.field`,
				errMsg: "attempt to",
			},
		}

		for _, tt := range errorScripts {
			t.Run(tt.name, func(t *testing.T) {
				coroID, err := runtime.SpawnCoroutine(ctx, L, tt.script)
				if err != nil {
					t.Fatalf("Failed to spawn coroutine: %v", err)
				}

				_, err = runtime.WaitForCoroutine(ctx, coroID)
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			})
		}
	})

	t.Run("multiple return values", func(t *testing.T) {
		script := `
			local function multi()
				return 1, "two", true
			end
			local a, b, c = multi()
			return a  -- Explicitly return first value
		`

		coroID, err := runtime.SpawnCoroutine(ctx, L, script)
		if err != nil {
			t.Fatalf("Failed to spawn coroutine: %v", err)
		}

		result, err := runtime.WaitForCoroutine(ctx, coroID)
		if err != nil {
			t.Errorf("WaitForCoroutine failed: %v", err)
		}

		// Note: Lua coroutines only return the first value
		if num, ok := result.(lua.LNumber); ok {
			if float64(num) != 1 {
				t.Errorf("Expected first return value 1, got %v", num)
			}
		} else {
			t.Errorf("Expected LNumber as first return value, got %T", result)
		}
	})
}

// Comprehensive Promise Integration Tests

func TestAsyncRuntime_PromiseIntegrationComplete(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	t.Run("promise await", func(t *testing.T) {
		script := `
			local start = os.clock()
			while os.clock() - start < 0.05 do end
			return "promise result"
		`

		promise, err := runtime.CreatePromise(ctx, L, script)
		if err != nil {
			t.Fatalf("Failed to create promise: %v", err)
		}

		// Check initial state
		if promise.IsResolved() {
			t.Errorf("Promise should not be resolved immediately")
		}

		// Await result
		result, err := promise.Await(ctx)
		if err != nil {
			t.Errorf("Promise.Await failed: %v", err)
		}

		if str, ok := result.(lua.LString); ok {
			if string(str) != "promise result" {
				t.Errorf("Expected 'promise result', got %s", str)
			}
		} else {
			t.Errorf("Expected LString result, got %T", result)
		}

		// Should be resolved now
		if !promise.IsResolved() {
			t.Errorf("Promise should be resolved after await")
		}
	})

	t.Run("promise cancellation", func(t *testing.T) {
		longScript := `
			local start = os.clock()
			while os.clock() - start < 10 do end
			return "should not complete"
		`

		promise, err := runtime.CreatePromise(ctx, L, longScript)
		if err != nil {
			t.Fatalf("Failed to create promise: %v", err)
		}

		// Cancel after short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			promise.Cancel()
		}()

		// Create timeout context for await
		awaitCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		// Await should fail with cancellation
		_, err = promise.Await(awaitCtx)
		if err == nil {
			t.Errorf("Expected cancellation error, got nil")
		}
		// Either context cancelled or deadline exceeded is acceptable
		if err != context.Canceled && err != context.DeadlineExceeded {
			t.Errorf("Expected context.Canceled or DeadlineExceeded, got %v", err)
		}
	})

	t.Run("empty promise resolution", func(t *testing.T) {
		promise, err := runtime.CreateEmptyPromise(ctx)
		if err != nil {
			t.Fatalf("Failed to create empty promise: %v", err)
		}

		// Set result manually
		testValue := lua.LString("manual result")
		runtime.SetCoroutineResult(promise.GetCoroID(), testValue, nil)

		// Await should return our value
		result, err := promise.Await(ctx)
		if err != nil {
			t.Errorf("Promise.Await failed: %v", err)
		}

		if result != testValue {
			t.Errorf("Expected manual result, got %v", result)
		}
	})

	t.Run("promise error resolution", func(t *testing.T) {
		promise, err := runtime.CreateEmptyPromise(ctx)
		if err != nil {
			t.Fatalf("Failed to create empty promise: %v", err)
		}

		// Set error result
		testErr := fmt.Errorf("test error")
		runtime.SetCoroutineResult(promise.GetCoroID(), lua.LNil, testErr)

		// Await should return error
		_, err = promise.Await(ctx)
		if err == nil {
			t.Errorf("Expected error from promise, got nil")
		}
		if err.Error() != testErr.Error() {
			t.Errorf("Expected test error, got %v", err)
		}
	})
}

// Channel Operations with Async

func TestAsyncRuntime_ChannelAsyncOperations(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	t.Run("async channel send/receive", func(t *testing.T) {
		// Create channel
		channelID, err := channelMgr.CreateChannel(L, 5)
		if err != nil {
			t.Fatalf("Failed to create channel: %v", err)
		}

		// Spawn coroutine to send values
		sendScript := `return "sent"`
		sendCoro, err := runtime.SpawnCoroutine(ctx, L, sendScript)
		if err != nil {
			t.Fatalf("Failed to spawn send coroutine: %v", err)
		}

		// Send values in goroutine
		go func() {
			for i := 0; i < 3; i++ {
				err := channelMgr.Send(ctx, channelID, lua.LNumber(i))
				if err != nil {
					t.Errorf("Failed to send value %d: %v", i, err)
				}
			}
			_, _ = runtime.WaitForCoroutine(ctx, sendCoro)
		}()

		// Receive values
		var received []float64
		for i := 0; i < 3; i++ {
			value, err := channelMgr.Receive(ctx, channelID)
			if err != nil {
				t.Errorf("Failed to receive value: %v", err)
			}
			if num, ok := value.(lua.LNumber); ok {
				received = append(received, float64(num))
			}
		}

		// Verify received values
		for i, v := range received {
			if v != float64(i) {
				t.Errorf("Expected value %d, got %f", i, v)
			}
		}
	})

	t.Run("async channel select with timeout", func(t *testing.T) {
		// Create multiple channels
		var channelIDs []string
		for i := 0; i < 3; i++ {
			id, err := channelMgr.CreateChannel(L, 1)
			if err != nil {
				t.Fatalf("Failed to create channel %d: %v", i, err)
			}
			channelIDs = append(channelIDs, id)
		}

		// Spawn coroutine to send on one channel after delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = channelMgr.Send(ctx, channelIDs[1], lua.LString("delayed value"))
		}()

		// Create select cases
		var selectCases []SelectCase
		for _, id := range channelIDs {
			selectCases = append(selectCases, SelectCase{
				ChannelID: id,
				Operation: SelectReceive,
			})
		}

		// Select with timeout context
		selectCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		selectedIdx, value, err := channelMgr.Select(selectCtx, selectCases)
		if err != nil {
			t.Errorf("Select failed: %v", err)
		}

		if selectedIdx != 1 {
			t.Errorf("Expected channel 1 to be selected, got %d", selectedIdx)
		}

		if str, ok := value.(lua.LString); ok {
			if string(str) != "delayed value" {
				t.Errorf("Expected 'delayed value', got %s", str)
			}
		}
	})
}

// Complex Cancellation Scenarios

func TestAsyncRuntime_ComplexCancellation(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	L := lua.NewState()
	defer L.Close()

	t.Run("cascading cancellation", func(t *testing.T) {
		parentCtx, parentCancel := context.WithCancel(context.Background())
		defer parentCancel()

		// Create nested contexts
		childCtx1, childCancel1 := context.WithCancel(parentCtx)
		defer childCancel1()
		childCtx2, childCancel2 := context.WithCancel(parentCtx)
		defer childCancel2()

		// Spawn coroutines with different contexts
		script := `
			local start = os.clock()
			while os.clock() - start < 10 do end
			return "should not complete"
		`

		coro1, _ := runtime.SpawnCoroutine(childCtx1, L, script)
		coro2, _ := runtime.SpawnCoroutine(childCtx2, L, script)

		// Cancel parent context
		go func() {
			time.Sleep(50 * time.Millisecond)
			parentCancel()
		}()

		// Both should be cancelled
		_, err1 := runtime.WaitForCoroutine(childCtx1, coro1)
		_, err2 := runtime.WaitForCoroutine(childCtx2, coro2)

		if err1 == nil || err2 == nil {
			t.Errorf("Expected both coroutines to be cancelled")
		}
	})

	t.Run("selective cancellation", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple coroutines
		var coroIDs []string
		var cancels []context.CancelFunc

		for i := 0; i < 5; i++ {
			coroCtx, cancel := context.WithCancel(ctx)
			cancels = append(cancels, cancel)

			script := `
				local start = os.clock()
				while os.clock() - start < 10 do end
				return "should not complete"
			`

			coroID, err := runtime.SpawnCoroutine(coroCtx, L, script)
			if err != nil {
				t.Fatalf("Failed to spawn coroutine %d: %v", i, err)
			}
			coroIDs = append(coroIDs, coroID)
		}

		// Cancel even-indexed coroutines
		for i := 0; i < len(cancels); i += 2 {
			cancels[i]()
		}

		// Check results
		for i, coroID := range coroIDs {
			coroCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			_, err := runtime.WaitForCoroutine(coroCtx, coroID)
			cancel()

			if i%2 == 0 {
				// Even indices should be cancelled
				if err == nil {
					t.Errorf("Expected coroutine %d to be cancelled", i)
				}
			} else {
				// Odd indices should timeout
				if err != context.DeadlineExceeded {
					t.Errorf("Expected coroutine %d to timeout, got %v", i, err)
				}
			}
		}
	})
}

// Concurrent Async Operations

func TestAsyncRuntime_ConcurrentOperations(t *testing.T) {
	runtime, err := NewAsyncRuntime(50)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	ctx := context.Background()

	t.Run("concurrent coroutine spawn", func(t *testing.T) {
		const numGoroutines = 10
		const corosPerGoroutine = 5

		var wg sync.WaitGroup
		errors := make([]error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// Each goroutine gets its own LState
				L := lua.NewState()
				defer L.Close()

				for j := 0; j < corosPerGoroutine; j++ {
					script := fmt.Sprintf(`return "goroutine %d coro %d"`, idx, j)
					coroID, err := runtime.SpawnCoroutine(ctx, L, script)
					if err != nil {
						errors[idx] = err
						return
					}

					result, err := runtime.WaitForCoroutine(ctx, coroID)
					if err != nil {
						errors[idx] = err
						return
					}

					expected := fmt.Sprintf("goroutine %d coro %d", idx, j)
					if str, ok := result.(lua.LString); ok {
						if string(str) != expected {
							errors[idx] = fmt.Errorf("expected %s, got %s", expected, str)
							return
						}
					}
				}
			}(i)
		}

		wg.Wait()

		// Check for errors
		for i, err := range errors {
			if err != nil {
				t.Errorf("Goroutine %d error: %v", i, err)
			}
		}
	})

	t.Run("concurrent promise operations", func(t *testing.T) {
		// Create multiple promises concurrently
		const numPromises = 20
		promises := make([]*Promise, numPromises)
		errors := make([]error, numPromises)

		var wg sync.WaitGroup
		for i := 0; i < numPromises; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				L := lua.NewState()
				defer L.Close()

				script := fmt.Sprintf(`return %d * %d`, idx, idx)
				promise, err := runtime.CreatePromise(ctx, L, script)
				if err != nil {
					errors[idx] = err
					return
				}
				promises[idx] = promise
			}(i)
		}
		wg.Wait()

		// Check for creation errors
		for i, err := range errors {
			if err != nil {
				t.Errorf("Promise creation %d error: %v", i, err)
			}
		}

		// Await all promises concurrently
		wg = sync.WaitGroup{}
		results := make([]lua.LValue, numPromises)

		for i, promise := range promises {
			if promise == nil {
				continue
			}

			wg.Add(1)
			go func(idx int, p *Promise) {
				defer wg.Done()
				result, err := p.Await(ctx)
				if err != nil {
					errors[idx] = err
					return
				}
				results[idx] = result
			}(i, promise)
		}
		wg.Wait()

		// Verify results
		for i, result := range results {
			if errors[i] != nil {
				t.Errorf("Promise await %d error: %v", i, errors[i])
				continue
			}

			if num, ok := result.(lua.LNumber); ok {
				expected := float64(i * i)
				if float64(num) != expected {
					t.Errorf("Promise %d: expected %f, got %f", i, expected, num)
				}
			}
		}
	})

	t.Run("stress test - max coroutines", func(t *testing.T) {
		// Try to exceed max coroutines limit
		maxCoros := runtime.maxCoroutines
		script := `
			local start = os.clock()
			while os.clock() - start < 0.01 do end
			return true
		`

		L := lua.NewState()
		defer L.Close()

		// Spawn max coroutines
		var coroIDs []string
		for i := 0; i < maxCoros; i++ {
			coroID, err := runtime.SpawnCoroutine(ctx, L, script)
			if err != nil {
				t.Errorf("Failed to spawn coroutine %d: %v", i, err)
				break
			}
			coroIDs = append(coroIDs, coroID)
		}

		// Should not be able to spawn more
		_, err := runtime.SpawnCoroutine(ctx, L, script)
		if err == nil {
			t.Errorf("Expected error when exceeding max coroutines")
		}

		// Active count should be at max
		activeCount := runtime.GetActiveCoroutineCount()
		if activeCount != maxCoros {
			t.Errorf("Expected %d active coroutines, got %d", maxCoros, activeCount)
		}

		// Wait for all to complete
		for _, coroID := range coroIDs {
			_, _ = runtime.WaitForCoroutine(ctx, coroID)
		}

		// Active count should be 0
		activeCount = runtime.GetActiveCoroutineCount()
		if activeCount != 0 {
			t.Errorf("Expected 0 active coroutines after completion, got %d", activeCount)
		}
	})
}

// Integration with Bridge Async

func TestAsyncRuntime_BridgeIntegration(t *testing.T) {
	runtime, err := NewAsyncRuntime(10)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	channelMgr, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = channelMgr.Close() }()

	L := lua.NewState()
	defer L.Close()

	ctx := context.Background()

	// Create mock bridge and wrapper
	mockBridge := newMockAsyncBridge()
	wrapper, err := NewAsyncBridgeWrapper(mockBridge, runtime, channelMgr)
	if err != nil {
		t.Fatalf("Failed to create async bridge wrapper: %v", err)
	}

	t.Run("bridge method async execution", func(t *testing.T) {
		// Execute multiple bridge methods concurrently
		const numCalls = 10
		var promises []*Promise

		for i := 0; i < numCalls; i++ {
			args := []engine.ScriptValue{engine.NewStringValue(fmt.Sprintf("call-%d", i))}
			promise, err := wrapper.ExecuteMethodAsync(ctx, L, "fastMethod", args)
			if err != nil {
				t.Errorf("Failed to execute async method %d: %v", i, err)
				continue
			}
			promises = append(promises, promise)
		}

		// Wait for all promises
		for i, promise := range promises {
			result, err := promise.Await(ctx)
			if err != nil {
				t.Errorf("Promise %d await failed: %v", i, err)
				continue
			}

			expected := fmt.Sprintf("result: call-%d", i)
			if str, ok := result.(lua.LString); ok {
				if string(str) != expected {
					t.Errorf("Promise %d: expected %s, got %s", i, expected, str)
				}
			}
		}
	})

	t.Run("streaming with async", func(t *testing.T) {
		// Test streaming method
		args := []engine.ScriptValue{engine.NewNumberValue(5)}
		stream, err := wrapper.ExecuteMethodStream(ctx, L, "streamMethod", args)
		if err != nil {
			t.Fatalf("Failed to execute stream method: %v", err)
		}

		// Collect stream values asynchronously
		values := make(chan float64, 5)
		go func() {
			defer close(values)
			for {
				value, err := stream.Next(ctx)
				if err != nil {
					break
				}
				if num, ok := value.(lua.LNumber); ok {
					values <- float64(num)
				}
			}
		}()

		// Verify received values
		var received []float64
		for v := range values {
			received = append(received, v)
		}

		if len(received) != 5 {
			t.Errorf("Expected 5 values, got %d", len(received))
		}

		for i, v := range received {
			if v != float64(i) {
				t.Errorf("Expected value %d, got %f", i, v)
			}
		}
	})
}

// Race Condition Testing

func TestAsyncRuntime_RaceConditions(t *testing.T) {
	runtime, err := NewAsyncRuntime(100)
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}
	defer func() { _ = runtime.Close() }()

	ctx := context.Background()

	t.Run("concurrent state modifications", func(t *testing.T) {
		const numGoroutines = 50
		const opsPerGoroutine = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				defer wg.Done()

				L := lua.NewState()
				defer L.Close()

				for j := 0; j < opsPerGoroutine; j++ {
					// Mix of operations
					switch j % 3 {
					case 0:
						// Spawn and wait
						script := fmt.Sprintf(`return %d`, idx*100+j)
						coroID, err := runtime.SpawnCoroutine(ctx, L, script)
						if err == nil {
							_, _ = runtime.WaitForCoroutine(ctx, coroID)
						}
					case 1:
						// Check active count
						_ = runtime.GetActiveCoroutineCount()
					case 2:
						// Create and cancel promise
						promise, err := runtime.CreatePromise(ctx, L, `return "test"`)
						if err == nil {
							promise.Cancel()
						}
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify runtime is still healthy
		finalCount := runtime.GetActiveCoroutineCount()
		if finalCount < 0 {
			t.Errorf("Invalid active coroutine count: %d", finalCount)
		}
	})
}
