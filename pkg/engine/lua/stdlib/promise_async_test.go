// ABOUTME: Tests for promise-async integration
// ABOUTME: Verifies that promises can work with async callbacks

package stdlib

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestPromiseAsyncCreation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Register all required modules
	RegisterPromise(L)
	RegisterAsyncCallback(L)
	RegisterPromiseAsync(L)

	// Test creating async promise
	if err := L.DoString(`
		local result = nil
		local p = promise.async(function(resolve, reject)
			-- Simulate async operation
			resolve("async result")
		end)
		
		-- Promise should be created
		assert(p ~= nil, "Promise not created")
		
		-- Wait for it
		p:next(function(value)
			result = value
		end)
		
		-- Should resolve immediately in this case
		assert(result == "async result", "Promise did not resolve correctly")
	`); err != nil {
		t.Errorf("Failed to test promise.async: %v", err)
	}
}

func TestPromiseAsyncWithCallbacks(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)
	RegisterAsyncCallback(L)
	RegisterPromiseAsync(L)

	mgr := GetCallbackManager(L)

	// Test promise that uses async callbacks
	if err := L.DoString(`
		local result = nil
		local p = promise.async(function(resolve, reject)
			-- Create async callback
			local id = async.create_callback(function(value)
				resolve(value)
			end)
			-- Callback ID should be returned
			assert(id > 0, "Invalid callback ID")
		end)
		
		p:next(function(value)
			result = value
		end)
		
		-- Result should still be nil (callback not fired yet)
		assert(result == nil, "Promise resolved too early")
	`); err != nil {
		t.Errorf("Failed to setup async promise: %v", err)
	}

	// Get the callback ID and queue result
	mgr.QueueStringResult(1, "callback result")

	// Process callbacks
	if err := L.DoString(`
		async.process_callbacks()
		-- Now the promise should be resolved
		assert(result == "callback result", "Promise did not resolve with callback result")
	`); err != nil {
		t.Errorf("Failed to process async callback: %v", err)
	}
}

func TestPromiseAwaitAll(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)
	RegisterAsyncCallback(L)
	RegisterPromiseAsync(L)

	mgr := GetCallbackManager(L)

	// Create multiple promises
	if err := L.DoString(`
		callback_ids = {}
		promises = {}
		
		for i = 1, 3 do
			local p = promise.async(function(resolve, reject)
				local id = async.create_callback(function(value)
					resolve(value)
				end)
				table.insert(callback_ids, id)
			end)
			table.insert(promises, p)
		end
		
		-- All promises should be pending
		assert(#promises == 3, "Expected 3 promises")
		assert(#callback_ids == 3, "Expected 3 callback IDs")
	`); err != nil {
		t.Errorf("Failed to create promises: %v", err)
	}

	// Queue results for all callbacks
	go func() {
		time.Sleep(50 * time.Millisecond)
		mgr.QueueStringResult(1, "result1")
		mgr.QueueStringResult(2, "result2")
		mgr.QueueStringResult(3, "result3")
	}()

	// Test await_all
	if err := L.DoString(`
		local results = promise.await_all(promises, 5) -- 5 second timeout
		assert(#results == 3, "Expected 3 results")
		assert(results[1] == "result1", "Wrong result 1")
		assert(results[2] == "result2", "Wrong result 2")
		assert(results[3] == "result3", "Wrong result 3")
	`); err != nil {
		t.Errorf("Failed to await all promises: %v", err)
	}
}

func TestPromiseAwaitAllTimeout(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)
	RegisterAsyncCallback(L)
	RegisterPromiseAsync(L)

	// Test timeout behavior
	if err := L.DoString(`
		-- Create a promise that won't resolve
		local p = promise.async(function(resolve, reject)
			local id = async.create_callback(function(value)
				resolve(value)
			end)
			-- Don't queue any result
		end)
		
		-- Try to await with very short timeout
		local start_pending = async.pending_count()
		local results = promise.await_all({p}, 0) -- 0 second timeout
		
		-- Should timeout and return nil for unresolved promise
		assert(#results == 1, "Expected 1 result")
		assert(results[1] == nil, "Expected nil for timed out promise")
		
		-- Callback should still be pending
		assert(async.pending_count() == start_pending + 1, "Callback should still be pending")
	`); err != nil {
		t.Errorf("Failed to test timeout: %v", err)
	}
}

func TestPromiseAsyncRejection(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)
	RegisterAsyncCallback(L)
	RegisterPromiseAsync(L)

	mgr := GetCallbackManager(L)

	// Test rejection handling
	if err := L.DoString(`
		local error_result = nil
		local p = promise.async(function(resolve, reject)
			local id = async.create_callback(
				function(value) resolve(value) end,
				function(err) reject(err) end
			)
		end)
		
		p:catch(function(err)
			error_result = err
		end)
		
		assert(error_result == nil, "Error set too early")
	`); err != nil {
		t.Errorf("Failed to setup rejection test: %v", err)
	}

	// Queue an error
	mgr.QueueError(1, "async error")

	if err := L.DoString(`
		async.process_callbacks()
		assert(error_result == "async error", "Promise did not reject correctly")
	`); err != nil {
		t.Errorf("Failed to test rejection: %v", err)
	}
}