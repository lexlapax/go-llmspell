// ABOUTME: Integration between promises and async callbacks
// ABOUTME: Enables promise-based async operations with true parallelism

package stdlib

import (
	"time"
	
	lua "github.com/yuin/gopher-lua"
)

// RegisterPromiseAsync adds async-aware promise helpers
func RegisterPromiseAsync(L *lua.LState) {
	// Add helper functions to promise module
	promiseMod := L.GetGlobal("promise").(*lua.LTable)
	
	// Add promise.async for creating async promises
	L.SetField(promiseMod, "async", L.NewFunction(promiseAsync))
	
	// Add promise.await_all for waiting on async operations
	L.SetField(promiseMod, "await_all", L.NewFunction(promiseAwaitAll))
}

// promiseAsync creates a promise that can use async callbacks
// Usage: p = promise.async(function(resolve, reject) ... end)
func promiseAsync(L *lua.LState) int {
	executor := L.CheckFunction(1)
	
	// Create promise
	promise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}
	
	// Create userdata
	ud := L.NewUserData()
	ud.Value = promise
	L.SetMetatable(ud, L.GetTypeMetatable("promise"))
	
	// Create resolve/reject functions
	resolveFunc := L.NewClosure(func(L *lua.LState) int {
		value := L.Get(1)
		promise.resolveWithLua(L, value)
		return 0
	})
	
	rejectFunc := L.NewClosure(func(L *lua.LState) int {
		reason := L.Get(1)
		promise.rejectWithLua(L, reason)
		return 0
	})
	
	// Execute the executor
	L.Push(executor)
	L.Push(resolveFunc)
	L.Push(rejectFunc)
	if err := L.PCall(2, 0, nil); err != nil {
		promise.rejectWithLua(L, lua.LString(err.Error()))
	}
	
	L.Push(ud)
	return 1
}

// promiseAwaitAll waits for all promises while processing async callbacks
// Usage: results = promise.await_all(promises)
func promiseAwaitAll(L *lua.LState) int {
	promises := L.CheckTable(1)
	timeout := L.OptInt(2, 30) // Default 30 second timeout
	
	// Get callback manager for processing
	_ = GetCallbackManager(L)
	
	// Check if all promises are resolved or rejected
	allSettled := func() bool {
		settled := true
		promises.ForEach(func(_, v lua.LValue) {
			if ud, ok := v.(*lua.LUserData); ok {
				if p, ok := ud.Value.(*Promise); ok {
					p.mu.RLock()
					if p.state == PromisePending {
						settled = false
					}
					p.mu.RUnlock()
				}
			}
		})
		return settled
	}
	
	// Event loop with timeout
	iterations := 0
	maxIterations := timeout * 100 // Check every 10ms
	
	// Only run the loop if timeout > 0
	if timeout > 0 {
		for !allSettled() && iterations < maxIterations {
			// Process any pending callbacks
			L.Push(L.GetGlobal("async").(*lua.LTable).RawGetString("process_callbacks"))
			if err := L.PCall(0, 1, nil); err == nil {
				processed := L.ToInt(-1)
				L.Pop(1) // Pop the result
				if processed > 0 {
					// Give promises a chance to update their state
					// after callbacks are processed
					continue
				}
			}
			
			iterations++
			
			// Sleep for 10ms between checks
			time.Sleep(10 * time.Millisecond)
		}
	}
	
	// Collect results
	results := L.NewTable()
	count := 0
	
	// First pass: count total promises
	promises.ForEach(func(_, v lua.LValue) {
		count++
	})
	
	// Second pass: collect results
	idx := 1
	promises.ForEach(func(_, v lua.LValue) {
		if ud, ok := v.(*lua.LUserData); ok {
			if p, ok := ud.Value.(*Promise); ok {
				p.mu.RLock()
				if p.state == PromiseResolved {
					// Convert interface{} back to LValue
					lv := interfaceToLuaValue(L, p.value)
					results.RawSetInt(idx, lv)
				} else {
					results.RawSetInt(idx, lua.LNil)
				}
				p.mu.RUnlock()
			} else {
				// Not a promise
				results.RawSetInt(idx, lua.LNil)
			}
		} else {
			// Not userdata
			results.RawSetInt(idx, lua.LNil)
		}
		idx++
	})
	
	// Set metatable to override length operator
	mt := L.NewTable()
	L.SetField(mt, "__len", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(count))
		return 1
	}))
	L.SetMetatable(results, mt)
	
	L.Push(results)
	return 1
}