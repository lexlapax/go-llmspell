// ABOUTME: Comprehensive test suite for Promise & Async Library in Lua standard library
// ABOUTME: Tests promise constructor, chaining, concurrency, async/await, and coroutine integration

package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// setupPromiseLibrary loads the promise library and sets it as global
func setupPromiseLibrary(t *testing.T, L *lua.LState) {
	t.Helper()
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		t.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)
}

// TestPromiseLibraryLoading tests that the promise library can be loaded
func TestPromiseLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupPromiseLibrary(t, L)

	// Check that promise table exists and has expected structure
	script := `
		if type(promise) ~= "table" then
			error("Promise module should be a table")
		end
		
		if type(promise.Promise) ~= "table" then
			error("Promise class should be available")
		end
		
		if type(promise.async) ~= "function" then
			error("async function should be available")
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Promise library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}
}

// TestPromiseConstructor tests Promise constructor and executor behavior
func TestPromiseConstructor(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "basic_constructor",
			script: `
				local p = promise.Promise.new(function(resolve, reject)
					resolve("test value")
				end)
				return p
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.Type() != lua.LTTable {
					t.Errorf("Expected promise instance to be a table, got %v", result.Type())
				}
			},
		},
		{
			name: "resolve_with_value",
			script: `
				local result = nil
				local p = promise.Promise.new(function(resolve, reject)
					resolve("success")
				end)
				p:andThen(function(value)
					result = value
				end)
				return result
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "success" {
					t.Errorf("Expected 'success', got %v", result.String())
				}
			},
		},
		{
			name: "reject_with_reason",
			script: `
				local error_reason = nil
				local p = promise.Promise.new(function(resolve, reject)
					reject("error occurred")
				end)
				p:onError(function(reason)
					error_reason = reason
				end)
				return error_reason
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "error occurred" {
					t.Errorf("Expected 'error occurred', got %v", result.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupPromiseLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestPromiseChaining tests andThen/onError/onFinally methods
func TestPromiseChaining(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "andThen_chaining",
			script: `
				local final_result = nil
				promise.Promise.new(function(resolve)
					resolve(10)
				end)
				:andThen(function(value)
					return value * 2
				end)
				:andThen(function(value)
					return value + 5
				end)
				:andThen(function(value)
					final_result = value
				end)
				return final_result
			`,
			check: func(t *testing.T, result lua.LValue) {
				if number, ok := result.(lua.LNumber); ok {
					if float64(number) != 25.0 {
						t.Errorf("Expected 25, got %v", float64(number))
					}
				} else {
					t.Errorf("Expected number result, got %v", result.Type())
				}
			},
		},
		{
			name: "onError_error_propagation",
			script: `
				local error_handled = false
				promise.Promise.new(function(resolve)
					resolve(10)
				end)
				:andThen(function(value)
					error("chain error")
				end)
				:andThen(function(value)
					-- This should not be called
					return value
				end)
				:onError(function(reason)
					error_handled = true
				end)
				return error_handled
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error to be handled, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupPromiseLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestPromiseStatics tests Promise.all, Promise.race, Promise.resolve/reject
func TestPromiseStatics(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "promise_resolve",
			script: `
				local result = nil
				promise.Promise.resolve("resolved value")
				:andThen(function(value)
					result = value
				end)
				return result
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "resolved value" {
					t.Errorf("Expected 'resolved value', got %v", result.String())
				}
			},
		},
		{
			name: "promise_reject",
			script: `
				local error_reason = nil
				promise.Promise.reject("rejected reason")
				:onError(function(reason)
					error_reason = reason
				end)
				return error_reason
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.String() != "rejected reason" {
					t.Errorf("Expected 'rejected reason', got %v", result.String())
				}
			},
		},
		{
			name: "promise_all_success",
			script: `
				local results = nil
				local promises = {
					promise.Promise.resolve("first"),
					promise.Promise.resolve("second"),
					promise.Promise.resolve("third")
				}
				promise.Promise.all(promises)
				:andThen(function(values)
					results = values
				end)
				return results
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result.Type() != lua.LTTable {
					t.Errorf("Expected table result, got %v", result.Type())
				}
				table := result.(*lua.LTable)
				if table.Len() != 3 {
					t.Errorf("Expected 3 results, got %d", table.Len())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupPromiseLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestAsyncAwait tests async/await syntax sugar
func TestAsyncAwait(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "async_function_wrapper",
			script: `
				local async_func = promise.async(function(x, y)
					return x + y
				end)
				local p = async_func(10, 20)
				-- Test that it returns a promise-like object
				return type(p) == "table" and type(p.andThen) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected async function to return a promise, got %v", result)
				}
			},
		},
		{
			name: "spawn_coroutine",
			script: `
				local p = promise.spawn(function(value)
					return value * 2
				end, 21)
				-- Test that spawn returns a promise-like object
				return type(p) == "table" and type(p.andThen) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected spawn to return a promise, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupPromiseLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestChannelCommunication tests channel-based communication
func TestChannelCommunication(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupPromiseLibrary(t, L)

	script := `
		-- Create a channel
		local channel_name = promise.create_channel("test_channel", 1)
		
		local sent = false
		local received_value = nil
		
		-- Send a value
		promise.send(channel_name, "test message"):andThen(function()
			sent = true
		end)
		
		-- Receive the value
		promise.receive(channel_name):andThen(function(value)
			received_value = value
		end)
		
		promise.close_channel(channel_name)
		
		return {sent = sent, received = received_value}
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Channel communication test failed: %v", err)
	}

	result := L.Get(-1)
	if result.Type() != lua.LTTable {
		t.Errorf("Expected table result, got %v", result.Type())
	}

	table := result.(*lua.LTable)
	sent := table.RawGetString("sent")
	received := table.RawGetString("received")

	if sent != lua.LTrue {
		t.Errorf("Expected send to complete, got %v", sent)
	}

	if received.String() != "test message" {
		t.Errorf("Expected 'test message', got %v", received.String())
	}
}

// TestErrorHandling tests error propagation and validation
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "invalid_executor",
			script: `
				local success, err = pcall(function()
					promise.Promise.new("not a function")
				end)
				return not success
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid executor, got %v", result)
				}
			},
		},
		{
			name: "promise_all_invalid_input",
			script: `
				local error_caught = false
				promise.Promise.all("not a table"):onError(function(reason)
					error_caught = true
				end)
				return error_caught
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid input, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupPromiseLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// BenchmarkPromiseCreation benchmarks promise creation performance
func BenchmarkPromiseCreation(b *testing.B) {
	L := lua.NewState()
	defer L.Close()

	// Load the promise library
	promisePath := filepath.Join(".", "promise.lua")
	err := L.DoFile(promisePath)
	if err != nil {
		b.Fatalf("Failed to load promise library: %v", err)
	}
	promise := L.Get(-1)
	L.SetGlobal("promise", promise)

	script := `
		local p = promise.Promise.new(function(resolve)
			resolve("test")
		end)
		return p
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := L.DoString(script)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
		L.Pop(1) // Clean stack
	}
}

// TestPackageRequire tests that the module can be required as a package
func TestPackageRequire(t *testing.T) {
	// Change to the stdlib directory for testing
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	err = os.Chdir(filepath.Dir(filepath.Join(wd, "promise.lua")))
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	L := lua.NewState()
	defer L.Close()

	script := `
		local promise = require('promise')
		
		-- Test that the module loads correctly
		if type(promise) ~= "table" then
			error("Promise module should return a table")
		end
		
		if type(promise.Promise) ~= "table" then
			error("Promise class should be available")
		end
		
		if type(promise.async) ~= "function" then
			error("async function should be available")
		end
		
		return true
	`

	err = L.DoString(script)
	if err != nil {
		t.Fatalf("Package require test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected require test to pass, got %v", result)
	}
}
