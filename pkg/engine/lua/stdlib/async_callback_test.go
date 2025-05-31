// ABOUTME: Tests for the async callback system
// ABOUTME: Verifies callback registration, execution, and cleanup

package stdlib

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestCallbackManager(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Register async callback module
	RegisterAsyncCallback(L)

	// Get manager
	mgr := GetCallbackManager(L)
	if mgr == nil {
		t.Fatal("Failed to get callback manager")
	}

	// Test callback registration
	callback := L.NewFunction(func(L *lua.LState) int {
		L.SetGlobal("callback_result", L.Get(1))
		return 0
	})

	errback := L.NewFunction(func(L *lua.LState) int {
		L.SetGlobal("callback_error", L.Get(1))
		return 0
	})

	id := mgr.RegisterCallback(callback, errback)
	if id <= 0 {
		t.Error("Invalid callback ID")
	}

	// Test queueing result
	mgr.QueueStringResult(id, "test result")

	// Process callbacks
	if err := L.DoString(`
		local processed = async.process_callbacks()
		assert(processed == 1, "Expected 1 callback to be processed")
		assert(callback_result == "test result", "Callback result mismatch")
	`); err != nil {
		t.Errorf("Failed to process callbacks: %v", err)
	}

	// Test error callback
	id2 := mgr.RegisterCallback(callback, errback)
	mgr.QueueError(id2, "test error")

	if err := L.DoString(`
		local processed = async.process_callbacks()
		assert(processed == 1, "Expected 1 callback to be processed")
		assert(callback_error == "test error", "Error callback result mismatch")
	`); err != nil {
		t.Errorf("Failed to process error callback: %v", err)
	}
}

func TestAsyncCallbackConcurrency(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterAsyncCallback(L)
	mgr := GetCallbackManager(L)

	// Create multiple callbacks
	var ids []int64
	for i := 0; i < 10; i++ {
		callback := L.NewFunction(func(L *lua.LState) int {
			count := L.GetGlobal("callback_count")
			if count == lua.LNil {
				L.SetGlobal("callback_count", lua.LNumber(1))
			} else {
				L.SetGlobal("callback_count", lua.LNumber(count.(lua.LNumber)+1))
			}
			return 0
		})
		id := mgr.RegisterCallback(callback, nil)
		ids = append(ids, id)
	}

	// Queue results concurrently
	for _, id := range ids {
		go func(callbackID int64) {
			time.Sleep(10 * time.Millisecond) // Simulate async work
			mgr.QueueStringResult(callbackID, "done")
		}(id)
	}

	// Wait and process
	time.Sleep(100 * time.Millisecond)

	if err := L.DoString(`
		local total = 0
		for i = 1, 10 do
			local processed = async.process_callbacks()
			total = total + processed
			if total >= 10 then break end
		end
		assert(callback_count == 10, "Expected 10 callbacks to be processed, got " .. tostring(callback_count))
	`); err != nil {
		t.Errorf("Failed to process concurrent callbacks: %v", err)
	}
}

func TestAsyncCallbackPendingCount(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterAsyncCallback(L)

	if err := L.DoString(`
		-- Initially no pending callbacks
		assert(async.pending_count() == 0, "Expected 0 pending callbacks")
		
		-- Create a callback
		local id = async.create_callback(function(result)
			-- Do nothing
		end)
		
		-- Now should have 1 pending
		assert(async.pending_count() == 1, "Expected 1 pending callback")
		
		-- Process (no results queued, so nothing happens)
		async.process_callbacks()
		
		-- Still pending
		assert(async.pending_count() == 1, "Expected 1 pending callback after empty process")
	`); err != nil {
		t.Errorf("Failed to test pending count: %v", err)
	}
}

func TestAsyncCallbackCleanup(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterAsyncCallback(L)
	mgr := GetCallbackManager(L)

	// Register callbacks
	callback := L.NewFunction(func(L *lua.LState) int { return 0 })
	id1 := mgr.RegisterCallback(callback, nil)
	_ = mgr.RegisterCallback(callback, nil) // id2 registered but not used in this test

	// Queue and process one
	mgr.QueueStringResult(id1, "done")

	if err := L.DoString(`
		async.process_callbacks()
		assert(async.pending_count() == 1, "Expected 1 pending after processing one")
	`); err != nil {
		t.Errorf("Failed cleanup test: %v", err)
	}

	// Clean up manager
	CleanupCallbackManager(L)

	// Verify cleanup
	managersMu.Lock()
	_, exists := managers[L]
	managersMu.Unlock()
	if exists {
		t.Error("Manager not cleaned up")
	}
}

func TestAsyncCallbackQueueOverflow(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterAsyncCallback(L)
	mgr := GetCallbackManager(L)

	// Register a callback
	callback := L.NewFunction(func(L *lua.LState) int { return 0 })
	id := mgr.RegisterCallback(callback, nil)

	// Try to overflow the queue (capacity is 100)
	for i := 0; i < 200; i++ {
		mgr.QueueStringResult(id, "overflow")
	}

	// Should not panic, just drop excess results
	// Process what made it through
	processed := 0
	for i := 0; i < 10; i++ {
		L.Push(L.GetGlobal("async").(*lua.LTable).RawGetString("process_callbacks"))
		if err := L.PCall(0, 1, nil); err == nil {
			if n, ok := L.Get(-1).(lua.LNumber); ok {
				processed += int(n)
			}
			L.Pop(1)
		}
		if processed > 0 {
			break
		}
	}

	if processed == 0 {
		t.Error("No callbacks were processed despite queue overflow")
	}
}