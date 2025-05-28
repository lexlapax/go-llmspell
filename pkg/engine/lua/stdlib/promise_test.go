// ABOUTME: Tests for promise implementation
// ABOUTME: Tests promise.new(), then(), catch(), await(), all(), race()

package stdlib

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestPromiseNew(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	// Test immediate resolution
	err := L.DoString(`
		local p = promise.new(function(resolve, reject)
			resolve("hello")
		end)
		
		local value = p:await()
		assert(value == "hello", "Expected 'hello', got " .. tostring(value))
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestPromiseReject(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p = promise.new(function(resolve, reject)
			reject("error occurred")
		end)
		
		local value, err = p:await()
		assert(value == nil, "Expected nil value")
		assert(err == "error occurred", "Expected 'error occurred', got " .. tostring(err))
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestPromiseThen(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p = promise.new(function(resolve, reject)
			resolve(5)
		end)
		
		local p2 = p:next(function(value)
			return value * 2
		end)
		
		local result = p2:await()
		assert(result == 10, "Expected 10, got " .. tostring(result))
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestPromiseCatch(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p = promise.new(function(resolve, reject)
			reject("error")
		end)
		
		local p2 = p:catch(function(err)
			return "handled: " .. err
		end)
		
		local result = p2:await()
		assert(result == "handled: error", "Expected 'handled: error', got " .. tostring(result))
	`)
	if err != nil {
		t.Fatalf("Failed to execute test: %v", err)
	}
}

func TestPromiseAll(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p1 = promise.resolve(1)
		local p2 = promise.resolve(2)
		local p3 = promise.resolve(3)
		
		local p = promise.all({p1, p2, p3})
		local values = p:await()
		
		assert(type(values) == "table", "Expected table result")
		assert(values[1] == 1, "Expected values[1] = 1")
		assert(values[2] == 2, "Expected values[2] = 2")  
		assert(values[3] == 3, "Expected values[3] = 3")
	`)
	if err != nil {
		t.Fatalf("Failed to execute promise.all test: %v", err)
	}
}

func TestPromiseRace(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p1 = promise.resolve("first")
		local p2 = promise.resolve("second")
		
		local p = promise.race({p1, p2})
		local result = p:await()
		
		assert(result == "first" or result == "second", "Expected 'first' or 'second', got " .. tostring(result))
	`)
	if err != nil {
		t.Fatalf("Failed to execute promise.race test: %v", err)
	}
}

func TestPromiseAsync(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	// Create a delayed resolver
	L.SetGlobal("delay_resolve", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckFunction(1) // resolve function
		_ = L.CheckString(2)   // value to resolve with

		go func() {
			time.Sleep(50 * time.Millisecond)
			// Can't safely call Lua from goroutine in this implementation
			// This test just verifies the await timeout works
		}()

		return 0
	}))

	err := L.DoString(`
		local p = promise.new(function(resolve, reject)
			-- Since we can't safely resolve from Go goroutine, test timeout
			delay_resolve(resolve, "delayed")
		end)
		
		local value, err = p:await(0.1) -- 100ms timeout
		assert(err == "timeout", "Expected timeout error")
	`)
	if err != nil {
		t.Fatalf("Failed to execute async test: %v", err)
	}
}

func TestPromiseChaining(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p = promise.resolve(1)
			:next(function(v) return v + 1 end)
			:next(function(v) return v * 2 end)
			:next(function(v) return v + 10 end)
		
		local result = p:await()
		assert(result == 14, "Expected 14, got " .. tostring(result))
	`)
	if err != nil {
		t.Fatalf("Failed to execute chaining test: %v", err)
	}
}

func TestPromiseAllReject(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterPromise(L)

	err := L.DoString(`
		local p1 = promise.resolve(1)
		local p2 = promise.reject("failed")
		local p3 = promise.resolve(3)
		
		local p = promise.all({p1, p2, p3})
		local value, err = p:await()
		
		assert(value == nil, "Expected nil value")
		assert(err == "failed", "Expected 'failed' error")
	`)
	if err != nil {
		t.Fatalf("Failed to execute promise.all reject test: %v", err)
	}
}
