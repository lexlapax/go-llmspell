// ABOUTME: Simple tests for async LLM bridge methods
// ABOUTME: Tests the mechanics without requiring actual LLM providers

package bridges

import (
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine/lua/stdlib"
	lua "github.com/yuin/gopher-lua"
)

func TestAsyncCallbackIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Register async callback system
	stdlib.RegisterAsyncCallback(L)

	// Test basic async functionality
	if err := L.DoString(`
		-- Test callback registration and execution
		_G.result = nil
		_G.callback_called = false
		
		-- Create a callback
		local id = async.create_callback(function(value)
			_G.result = value
			_G.callback_called = true
		end)
		
		assert(id > 0, "Invalid callback ID")
		assert(async.pending_count() == 1, "Expected 1 pending callback")
		
		-- Nothing to process yet
		local processed = async.process_callbacks()
		assert(processed == 0, "Should not process anything yet")
		assert(_G.callback_called == false, "Callback called too early")
	`); err != nil {
		t.Errorf("Failed basic async test: %v", err)
	}

	// Now simulate the Go side queuing a result
	mgr := stdlib.GetCallbackManager(L)
	mgr.QueueStringResult(1, "test result")

	// Process the callback
	if err := L.DoString(`
		local processed = async.process_callbacks()
		print("Processed callbacks:", processed)
		print("Callback called:", _G.callback_called)
		print("Result:", _G.result)
		print("Pending count:", async.pending_count())
		
		assert(processed == 1, "Expected to process 1 callback, got " .. processed)
		assert(_G.callback_called == true, "Callback not called")
		assert(_G.result == "test result", "Wrong callback result: " .. tostring(_G.result))
		assert(async.pending_count() == 0, "Callback not cleaned up")
	`); err != nil {
		t.Errorf("Failed callback processing: %v", err)
	}
}

func TestAsyncErrorCallback(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	stdlib.RegisterAsyncCallback(L)

	// Test error callback
	if err := L.DoString(`
		_G.error_result = nil
		_G.success_called = false
		
		-- Create callback with error handler
		local id = async.create_callback(
			function(value)
				_G.success_called = true
			end,
			function(err)
				_G.error_result = err
			end
		)
		
		assert(id > 0, "Invalid callback ID")
	`); err != nil {
		t.Errorf("Failed to create error callback: %v", err)
	}

	// Queue an error
	mgr := stdlib.GetCallbackManager(L)
	mgr.QueueError(1, "test error")

	// Process
	if err := L.DoString(`
		async.process_callbacks()
		assert(_G.success_called == false, "Success callback should not be called")
		assert(_G.error_result == "test error", "Wrong error message")
	`); err != nil {
		t.Errorf("Failed error callback test: %v", err)
	}
}

func TestMultipleCallbacks(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	stdlib.RegisterAsyncCallback(L)

	// Create multiple callbacks
	if err := L.DoString(`
		results = {}
		
		for i = 1, 3 do
			async.create_callback(function(value)
				table.insert(results, value)
			end)
		end
		
		assert(async.pending_count() == 3, "Expected 3 pending callbacks")
	`); err != nil {
		t.Errorf("Failed to create multiple callbacks: %v", err)
	}

	// Queue results
	mgr := stdlib.GetCallbackManager(L)
	mgr.QueueStringResult(1, "first")
	mgr.QueueStringResult(2, "second")
	mgr.QueueStringResult(3, "third")

	// Process all
	if err := L.DoString(`
		local total_processed = 0
		for i = 1, 5 do
			local processed = async.process_callbacks()
			total_processed = total_processed + processed
			if total_processed >= 3 then break end
		end
		
		assert(#results == 3, "Expected 3 results, got " .. #results)
		assert(results[1] == "first", "Wrong first result")
		assert(results[2] == "second", "Wrong second result")
		assert(results[3] == "third", "Wrong third result")
		assert(async.pending_count() == 0, "Callbacks not cleaned up")
	`); err != nil {
		t.Errorf("Failed multiple callbacks test: %v", err)
	}
}
