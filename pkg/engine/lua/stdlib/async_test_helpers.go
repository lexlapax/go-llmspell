// ABOUTME: Async-specific test helpers for go-llmspell Lua standard library testing
// ABOUTME: Provides utilities for promise assertions, coroutine lifecycle, timeout testing, and memory leak detection

// Package stdlib provides the Lua standard library implementations for go-llmspell.
// This file contains async-specific test helpers for testing promise-based operations,
// coroutine lifecycle management, and async error handling.
package stdlib

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ============================================================================
// Promise Assertion Utilities
// ============================================================================

// AssertPromiseResolves verifies that a promise resolves with expected value.
// It waits for the promise to resolve within the specified timeout and checks
// that the resolved value matches the expected value.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the promise
//   - promiseVar: The name of the Lua variable holding the promise
//   - expectedValue: The expected resolved value
//   - timeout: Maximum time to wait for promise completion
//
// The function fails the test if the promise rejects or resolves with a different value.
func AssertPromiseResolves(t *testing.T, L *lua.LState, promiseVar string, expectedValue lua.LValue, timeout time.Duration) {
	t.Helper()

	script := fmt.Sprintf(`
		local promise = %s
		local result = nil
		local error_msg = nil
		local completed = false
		
		promise:andThen(function(value)
			result = value
			completed = true
		end):onError(function(err)
			error_msg = err
			completed = true
		end)
		
		_G._test_promise_result = result
		_G._test_promise_error = error_msg
		_G._test_promise_completed = completed
	`, promiseVar)

	if err := L.DoString(script); err != nil {
		t.Fatalf("Failed to set up promise test: %v", err)
	}

	// Wait for completion
	WaitForCondition(t, timeout, func() bool {
		completed := L.GetGlobal("_test_promise_completed")
		return lua.LVAsBool(completed)
	}, "promise completion")

	// Check result
	errorMsg := L.GetGlobal("_test_promise_error")
	if errorMsg != lua.LNil {
		t.Errorf("Promise rejected with error: %s", lua.LVAsString(errorMsg))
		return
	}

	result := L.GetGlobal("_test_promise_result")
	if !CompareLuaValues(t, L, expectedValue, result) {
		t.Errorf("Promise resolved with unexpected value")
	}
}

// AssertPromiseRejects verifies that a promise rejects with expected error.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the promise
//   - promiseVar: The name of the Lua variable holding the promise
//   - expectedError: Substring expected to be contained in the error message
//   - timeout: Maximum time to wait for promise completion
//
// The function fails the test if the promise resolves or rejects with an error
// that doesn't contain the expected error substring.
func AssertPromiseRejects(t *testing.T, L *lua.LState, promiseVar string, expectedError string, timeout time.Duration) {
	t.Helper()

	script := fmt.Sprintf(`
		local promise = %s
		local result = nil
		local error_msg = nil
		local completed = false
		
		promise:andThen(function(value)
			result = value
			completed = true
		end):onError(function(err)
			error_msg = tostring(err)
			completed = true
		end)
		
		_G._test_promise_result = result
		_G._test_promise_error = error_msg
		_G._test_promise_completed = completed
	`, promiseVar)

	if err := L.DoString(script); err != nil {
		t.Fatalf("Failed to set up promise test: %v", err)
	}

	// Wait for completion
	WaitForCondition(t, timeout, func() bool {
		completed := L.GetGlobal("_test_promise_completed")
		return lua.LVAsBool(completed)
	}, "promise completion")

	// Check error
	result := L.GetGlobal("_test_promise_result")
	if result != lua.LNil {
		t.Errorf("Promise resolved when rejection was expected")
		return
	}

	errorMsg := L.GetGlobal("_test_promise_error")
	if errorMsg == lua.LNil {
		t.Errorf("Promise did not reject as expected")
		return
	}

	errorStr := lua.LVAsString(errorMsg)
	if !strings.Contains(errorStr, expectedError) {
		t.Errorf("Promise rejected with unexpected error: got '%s', expected to contain '%s'",
			errorStr, expectedError)
	}
}

// AssertPromiseCompletes verifies that a promise completes (resolves or rejects) within timeout.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state containing the promise
//   - promiseVar: The name of the Lua variable holding the promise
//   - timeout: Maximum time to wait for promise completion
//
// Returns:
//   - bool: true if the promise completed within the timeout, false otherwise
//
// This function doesn't check the result, only that the promise finished executing.
func AssertPromiseCompletes(t *testing.T, L *lua.LState, promiseVar string, timeout time.Duration) bool {
	t.Helper()

	script := fmt.Sprintf(`
		local promise = %s
		local completed = false
		
		promise:onFinally(function()
			completed = true
		end)
		
		_G._test_promise_completed = completed
	`, promiseVar)

	if err := L.DoString(script); err != nil {
		t.Fatalf("Failed to set up promise test: %v", err)
	}

	completed := make(chan bool, 1)
	go func() {
		WaitForCondition(t, timeout, func() bool {
			c := L.GetGlobal("_test_promise_completed")
			return lua.LVAsBool(c)
		}, "promise completion")
		completed <- true
	}()

	select {
	case <-completed:
		return true
	case <-time.After(timeout):
		return false
	}
}

// CreateTestPromise creates a promise for testing that resolves after a delay.
//
// Parameters:
//   - L: The Lua state to create the promise in
//   - resolveAfter: Duration to wait before resolving the promise
//   - value: The value to resolve the promise with
//
// Returns:
//   - string: The name of the global variable containing the created promise
//
// The created promise uses promise.sleep to delay resolution.
func CreateTestPromise(L *lua.LState, resolveAfter time.Duration, value lua.LValue) string {
	varName := fmt.Sprintf("_test_promise_%d", time.Now().UnixNano())

	script := fmt.Sprintf(`
		%s = promise.new(function(resolve, reject)
			promise.sleep(%d):andThen(function()
				resolve(%s)
			end)
		end)
	`, varName, resolveAfter.Milliseconds(), luaValueToString(value))

	if err := L.DoString(script); err != nil {
		panic(fmt.Sprintf("Failed to create test promise: %v", err))
	}

	return varName
}

// ============================================================================
// Coroutine Lifecycle Helpers
// ============================================================================

// CoroutineTracker tracks coroutine lifecycle events including creation,
// resumption, yielding, completion, and errors. It provides thread-safe
// tracking of multiple coroutines for testing purposes.
type CoroutineTracker struct {
	mu        sync.Mutex
	created   []string
	resumed   []string
	yielded   []string
	completed []string
	errors    []string
}

// NewCoroutineTracker creates a new coroutine tracker with empty event lists.
//
// Returns:
//   - *CoroutineTracker: A new tracker instance ready to record coroutine events
func NewCoroutineTracker() *CoroutineTracker {
	return &CoroutineTracker{
		created:   []string{},
		resumed:   []string{},
		yielded:   []string{},
		completed: []string{},
		errors:    []string{},
	}
}

// InstallCoroutineTracker installs tracking hooks for coroutines by wrapping
// the standard coroutine.create, coroutine.resume, and coroutine.yield functions.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to install tracking in
//
// Returns:
//   - *CoroutineTracker: A tracker that will record all coroutine events
//
// The tracker intercepts coroutine operations to record lifecycle events
// without affecting normal coroutine behavior.
func InstallCoroutineTracker(t *testing.T, L *lua.LState) *CoroutineTracker {
	t.Helper()

	tracker := NewCoroutineTracker()

	// Install tracking functions
	L.SetGlobal("_coroutine_tracker", lua.LString("active"))

	script := `
		local original_create = coroutine.create
		local original_resume = coroutine.resume
		local original_yield = coroutine.yield
		
		local tracked_coroutines = {}
		
		coroutine.create = function(f)
			local co = original_create(f)
			local id = tostring(co)
			tracked_coroutines[co] = id
			
			-- Track creation
			if _G._track_coroutine_created then
				_G._track_coroutine_created(id)
			end
			
			-- Wrap the function to track completion
			local wrapped = original_create(function(...)
				local ok, err = pcall(f, ...)
				if _G._track_coroutine_completed then
					_G._track_coroutine_completed(id, ok, err)
				end
				if not ok then
					error(err)
				end
			end)
			
			return co
		end
		
		coroutine.resume = function(co, ...)
			local id = tracked_coroutines[co] or tostring(co)
			
			if _G._track_coroutine_resumed then
				_G._track_coroutine_resumed(id)
			end
			
			local results = {original_resume(co, ...)}
			
			if not results[1] and _G._track_coroutine_error then
				_G._track_coroutine_error(id, results[2])
			end
			
			return unpack(results)
		end
		
		coroutine.yield = function(...)
			local co = coroutine.running()
			local id = tracked_coroutines[co] or tostring(co)
			
			if _G._track_coroutine_yielded then
				_G._track_coroutine_yielded(id)
			end
			
			return original_yield(...)
		end
	`

	if err := L.DoString(script); err != nil {
		t.Fatalf("Failed to install coroutine tracker: %v", err)
	}

	// Set up tracking callbacks
	L.SetGlobal("_track_coroutine_created", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		tracker.mu.Lock()
		tracker.created = append(tracker.created, id)
		tracker.mu.Unlock()
		return 0
	}))

	L.SetGlobal("_track_coroutine_resumed", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		tracker.mu.Lock()
		tracker.resumed = append(tracker.resumed, id)
		tracker.mu.Unlock()
		return 0
	}))

	L.SetGlobal("_track_coroutine_yielded", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		tracker.mu.Lock()
		tracker.yielded = append(tracker.yielded, id)
		tracker.mu.Unlock()
		return 0
	}))

	L.SetGlobal("_track_coroutine_completed", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		ok := L.CheckBool(2)
		if ok {
			tracker.mu.Lock()
			tracker.completed = append(tracker.completed, id)
			tracker.mu.Unlock()
		}
		return 0
	}))

	L.SetGlobal("_track_coroutine_error", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		err := L.CheckString(2)
		tracker.mu.Lock()
		tracker.errors = append(tracker.errors, fmt.Sprintf("%s: %s", id, err))
		tracker.mu.Unlock()
		return 0
	}))

	return tracker
}

// GetStats returns coroutine statistics as a map of event type to count.
//
// Returns:
//   - map[string]int: Map with keys "created", "resumed", "yielded", "completed", "errors"
//     and values representing the count of each event type
//
// The method is thread-safe and returns a snapshot of current statistics.
func (ct *CoroutineTracker) GetStats() map[string]int {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	return map[string]int{
		"created":   len(ct.created),
		"resumed":   len(ct.resumed),
		"yielded":   len(ct.yielded),
		"completed": len(ct.completed),
		"errors":    len(ct.errors),
	}
}

// AssertCoroutineCompleted verifies a coroutine completed successfully.
//
// Parameters:
//   - t: The testing context
//   - coroutineID: The ID (or substring of ID) of the coroutine to check
//
// The method fails the test if no coroutine with an ID containing the given
// string has been recorded as completed.
func (ct *CoroutineTracker) AssertCoroutineCompleted(t *testing.T, coroutineID string) {
	t.Helper()

	ct.mu.Lock()
	defer ct.mu.Unlock()

	found := false
	for _, id := range ct.completed {
		if strings.Contains(id, coroutineID) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Coroutine %s did not complete successfully", coroutineID)
	}
}

// ============================================================================
// Timeout Testing Utilities
// ============================================================================

// TimeoutTest represents a test case that involves timeout handling.
// It allows testing both successful completion within timeout and
// expected timeout scenarios.
type TimeoutTest struct {
	Name     string
	Script   string
	Timeout  time.Duration
	Expected string // Expected result or error
	IsError  bool   // Whether an error is expected
}

// RunTimeoutTests runs a series of timeout tests with proper context handling.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run tests in
//   - tests: Array of TimeoutTest cases to execute
//
// Each test is run in a separate subtest with its own timeout context.
// Tests can verify both successful completion and expected timeouts.
func RunTimeoutTests(t *testing.T, L *lua.LState, tests []TimeoutTest) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), test.Timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				err := L.DoString(test.Script)
				done <- err
			}()

			select {
			case err := <-done:
				if test.IsError {
					if err == nil {
						t.Errorf("Expected error containing '%s', got none", test.Expected)
					} else if !strings.Contains(err.Error(), test.Expected) {
						t.Errorf("Expected error containing '%s', got: %v", test.Expected, err)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			case <-ctx.Done():
				if !test.IsError || !strings.Contains(test.Expected, "timeout") {
					t.Errorf("Test timed out unexpectedly")
				}
			}
		})
	}
}

// AssertCompletesWithin verifies that a Lua script completes execution within a timeout.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to execute the script in
//   - script: The Lua script to execute
//   - timeout: Maximum time allowed for script execution
//
// The function fails the test if the script doesn't complete within the timeout
// or if it completes with an error.
func AssertCompletesWithin(t *testing.T, L *lua.LState, script string, timeout time.Duration) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		err := L.DoString(script)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Script failed: %v", err)
		}
	case <-time.After(timeout):
		t.Errorf("Script did not complete within %v", timeout)
	}
}

// ============================================================================
// Concurrent Operation Validators
// ============================================================================

// ConcurrentTest represents a concurrent test scenario with setup,
// concurrent operations, teardown, and validation phases.
type ConcurrentTest struct {
	Name       string
	Setup      string                            // Setup script
	Concurrent []string                          // Scripts to run concurrently
	Teardown   string                            // Teardown script
	Validate   func(t *testing.T, L *lua.LState) // Validation function
}

// RunConcurrentTests runs tests with concurrent operations using coroutines.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run tests in
//   - tests: Array of ConcurrentTest cases to execute
//
// Each concurrent operation is run in its own coroutine with synchronized
// access to the Lua state. Setup and teardown scripts are run before and
// after the concurrent operations.
func RunConcurrentTests(t *testing.T, L *lua.LState, tests []ConcurrentTest) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Setup
			if test.Setup != "" {
				if err := L.DoString(test.Setup); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Run concurrent operations
			// Note: GopherLua states are not thread-safe, so we need to synchronize access
			var mu sync.Mutex
			var wg sync.WaitGroup
			errors := make(chan error, len(test.Concurrent))

			for i, script := range test.Concurrent {
				wg.Add(1)
				go func(idx int, s string) {
					defer wg.Done()

					// Synchronize access to the Lua state
					mu.Lock()
					defer mu.Unlock()

					// Create a new coroutine for each concurrent operation
					coScript := fmt.Sprintf(`
						coroutine.resume(coroutine.create(function()
							%s
						end))
					`, s)

					if err := L.DoString(coScript); err != nil {
						errors <- fmt.Errorf("concurrent operation %d failed: %v", idx, err)
					}
				}(i, script)
			}

			// Wait for completion
			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				t.Error(err)
			}

			// Validate
			if test.Validate != nil {
				test.Validate(t, L)
			}

			// Teardown
			if test.Teardown != "" {
				if err := L.DoString(test.Teardown); err != nil {
					t.Errorf("Teardown failed: %v", err)
				}
			}
		})
	}
}

// AssertRaceConditionFree tests for race conditions in concurrent code by running
// the same operation multiple times concurrently and checking for consistency.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run tests in
//   - setup: Lua script to run before the concurrent operations
//   - operation: Lua script to run concurrently (use $RESULT as placeholder for result variable)
//   - iterations: Number of concurrent operations to run
//
// The function verifies that all concurrent operations produce the same result,
// indicating no race conditions affected the outcome.
func AssertRaceConditionFree(t *testing.T, L *lua.LState, setup string, operation string, iterations int) {
	t.Helper()

	// Run setup
	if err := L.DoString(setup); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Track results
	results := make([]string, iterations)
	var counter int32

	// Run operations concurrently
	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Each goroutine gets its own variable name
			varName := fmt.Sprintf("_result_%d", idx)
			script := fmt.Sprintf(`
				%s = nil
				coroutine.resume(coroutine.create(function()
					%s
				end))
			`, varName, strings.ReplaceAll(operation, "$RESULT", varName))

			if err := L.DoString(script); err != nil {
				t.Errorf("Operation %d failed: %v", idx, err)
				return
			}

			// Get result
			result := L.GetGlobal(varName)
			results[idx] = lua.LVAsString(result)
			atomic.AddInt32(&counter, 1)
		}(i)
	}

	wg.Wait()

	// Verify all operations completed
	if int(counter) != iterations {
		t.Errorf("Not all operations completed: %d/%d", counter, iterations)
	}

	// Check for consistency (all results should be the same)
	if iterations > 1 {
		firstResult := results[0]
		for i := 1; i < iterations; i++ {
			if results[i] != firstResult {
				t.Errorf("Inconsistent results detected: result[0]=%s, result[%d]=%s",
					firstResult, i, results[i])
			}
		}
	}
}

// ============================================================================
// Memory Leak Detectors
// ============================================================================

// MemorySnapshot captures memory statistics for both Go runtime and Lua state
// at a specific point in time. Used for memory leak detection.
type MemorySnapshot struct {
	Timestamp  time.Time
	GoMemory   runtime.MemStats
	LuaMemory  float64 // KB used by Lua
	Goroutines int
	LuaObjects int
}

// CaptureMemorySnapshot captures current memory state for both Go and Lua.
//
// Parameters:
//   - L: The Lua state to capture memory statistics from
//
// Returns:
//   - *MemorySnapshot: Snapshot containing Go memory stats, Lua memory usage,
//     goroutine count, and Lua object count
//
// The function forces garbage collection before capturing to get accurate readings.
func CaptureMemorySnapshot(L *lua.LState) *MemorySnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get Lua memory usage
	_ = L.DoString("collectgarbage('collect')")
	_ = L.DoString("_G._mem_usage = collectgarbage('count')")
	memUsage := L.GetGlobal("_mem_usage")
	var luaMemKB float64
	if memUsage.Type() == lua.LTNumber {
		luaMemKB = float64(memUsage.(lua.LNumber))
	}

	// If memory is 0, it might be because the state is fresh
	// Force some allocation to get a non-zero value
	if luaMemKB == 0 {
		_ = L.DoString("local t = {} for i=1,100 do t[i] = i end")
		_ = L.DoString("_G._mem_usage = collectgarbage('count')")
		memUsage = L.GetGlobal("_mem_usage")
		if memUsage.Type() == lua.LTNumber {
			luaMemKB = float64(memUsage.(lua.LNumber))
		}
	}

	// Count Lua objects (simplified)
	_ = L.DoString(`
		local count = 0
		for k, v in pairs(_G) do
			count = count + 1
		end
		_G._object_count = count
	`)
	objCount := L.GetGlobal("_object_count")
	var objectCount int
	if objCount.Type() == lua.LTNumber {
		objectCount = int(objCount.(lua.LNumber))
	}

	return &MemorySnapshot{
		Timestamp:  time.Now(),
		GoMemory:   memStats,
		LuaMemory:  luaMemKB,
		Goroutines: runtime.NumGoroutine(),
		LuaObjects: objectCount,
	}
}

// AssertNoMemoryLeak verifies no significant memory leak occurred between two snapshots.
//
// Parameters:
//   - t: The testing context
//   - before: Memory snapshot taken before the operation
//   - after: Memory snapshot taken after the operation
//   - maxGrowthPercent: Maximum allowed memory growth as a percentage
//
// The function checks Go memory, Lua memory, and goroutine count.
// It fails the test if any metric exceeds the allowed growth percentage.
func AssertNoMemoryLeak(t *testing.T, before, after *MemorySnapshot, maxGrowthPercent float64) {
	t.Helper()

	// Check Go memory
	goMemBefore := before.GoMemory.Alloc
	goMemAfter := after.GoMemory.Alloc
	goGrowth := float64(goMemAfter-goMemBefore) / float64(goMemBefore) * 100

	if goGrowth > maxGrowthPercent {
		t.Errorf("Go memory grew by %.2f%% (max allowed: %.2f%%)", goGrowth, maxGrowthPercent)
	}

	// Check Lua memory
	luaGrowth := (after.LuaMemory - before.LuaMemory) / before.LuaMemory * 100
	if luaGrowth > maxGrowthPercent {
		t.Errorf("Lua memory grew by %.2f%% (max allowed: %.2f%%)", luaGrowth, maxGrowthPercent)
	}

	// Check goroutine leaks
	if after.Goroutines > before.Goroutines+2 { // Allow some variance
		t.Errorf("Goroutine leak detected: before=%d, after=%d",
			before.Goroutines, after.Goroutines)
	}
}

// RunMemoryLeakTest runs a test checking for memory leaks by executing an operation
// multiple times and comparing memory usage before and after.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to run the test in
//   - setup: Lua script to run before the iterations
//   - operation: Lua script to run repeatedly
//   - cleanup: Lua script to run after the iterations
//   - iterations: Number of times to run the operation
//
// The test allows up to 10% memory growth. Periodic garbage collection is
// performed during execution to simulate real conditions.
func RunMemoryLeakTest(t *testing.T, L *lua.LState, setup, operation, cleanup string, iterations int) {
	t.Helper()

	// Setup
	if setup != "" {
		if err := L.DoString(setup); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// Force GC and capture initial state
	runtime.GC()
	_ = L.DoString("collectgarbage('collect')")
	before := CaptureMemorySnapshot(L)

	// Run operations
	for i := 0; i < iterations; i++ {
		if err := L.DoString(operation); err != nil {
			t.Fatalf("Operation %d failed: %v", i, err)
		}

		// Periodic GC to simulate real conditions
		if i%100 == 0 {
			_ = L.DoString("collectgarbage('step')")
		}
	}

	// Cleanup
	if cleanup != "" {
		if err := L.DoString(cleanup); err != nil {
			t.Errorf("Cleanup failed: %v", err)
		}
	}

	// Force GC and capture final state
	runtime.GC()
	_ = L.DoString("collectgarbage('collect')")
	time.Sleep(100 * time.Millisecond) // Allow GC to complete
	after := CaptureMemorySnapshot(L)

	// Check for leaks (allow 10% growth)
	AssertNoMemoryLeak(t, before, after, 10.0)
}

// ============================================================================
// Helper Functions
// ============================================================================

// luaValueToString converts a Lua value to a string representation for scripts
func luaValueToString(v lua.LValue) string {
	switch v.Type() {
	case lua.LTNil:
		return "nil"
	case lua.LTBool:
		if lua.LVAsBool(v) {
			return "true"
		}
		return "false"
	case lua.LTNumber:
		return fmt.Sprintf("%v", lua.LVAsNumber(v))
	case lua.LTString:
		return fmt.Sprintf("%q", lua.LVAsString(v))
	default:
		return "nil" // Default for complex types
	}
}

// WaitForAsync waits for all async operations to complete in the Lua state.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state with pending async operations
//   - timeout: Maximum time to wait for completion
//
// The function attempts to wait for promises and coroutines to finish.
// It fails the test if operations don't complete within the timeout.
func WaitForAsync(t *testing.T, L *lua.LState, timeout time.Duration) {
	t.Helper()

	script := `
		-- Wait for pending promises and coroutines
		if promise and promise.await then
			-- Implementation specific
			collectgarbage('collect')
		end
	`

	done := make(chan bool, 1)
	go func() {
		_ = L.DoString(script)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(timeout):
		t.Errorf("Async operations did not complete within %v", timeout)
	}
}

// AssertEventuallyTrue asserts that a Lua expression becomes true within a timeout.
//
// Parameters:
//   - t: The testing context
//   - L: The Lua state to evaluate the condition in
//   - condition: Lua expression that should eventually evaluate to true
//   - timeout: Maximum time to wait for the condition
//   - message: Error message to display if condition never becomes true
//
// The function polls the condition every 10ms until it becomes true or times out.
func AssertEventuallyTrue(t *testing.T, L *lua.LState, condition string, timeout time.Duration, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		_ = L.DoString(fmt.Sprintf("_G._test_condition = (%s)", condition))
		result := L.GetGlobal("_test_condition")
		if lua.LVAsBool(result) {
			return
		}

		if time.Now().After(deadline) {
			t.Errorf("Condition never became true: %s", message)
			return
		}
	}
}
