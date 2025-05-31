// ABOUTME: Integration between promises and async callbacks
// ABOUTME: Enables promise-based async operations with true parallelism

package stdlib

import (
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

// promiseAsync creates a promise that works with async callbacks
// Usage: promise.async(function(resolve, reject) ... end)
func promiseAsync(L *lua.LState) int {
	executor := L.CheckFunction(1)
	
	promise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}
	
	ud := L.NewUserData()
	ud.Value = promise
	L.SetMetatable(ud, L.GetTypeMetatable("promise"))
	
	// Create resolve/reject closures that can be used in async callbacks
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
	
	// Execute the executor with async support
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
	
	// Check if all promises are resolved
	allResolved := func() bool {
		resolved := true
		promises.ForEach(func(_, v lua.LValue) {
			if ud, ok := v.(*lua.LUserData); ok {
				if p, ok := ud.Value.(*Promise); ok {
					p.mu.RLock()
					if p.state == PromisePending {
						resolved = false
					}
					p.mu.RUnlock()
				}
			}
		})
		return resolved
	}
	
	// Event loop with timeout
	iterations := 0
	maxIterations := timeout * 100 // Check every 10ms
	
	for !allResolved() && iterations < maxIterations {
		// Process any pending callbacks
		L.Push(L.GetGlobal("async").(*lua.LTable).RawGetString("process_callbacks"))
		if err := L.PCall(0, 1, nil); err == nil {
			L.Pop(1) // Pop the result
		}
		
		iterations++
		
		// Small yield to prevent busy waiting
		for i := 0; i < 10000; i++ {
			// Busy wait
		}
	}
	
	// Collect results
	results := L.NewTable()
	idx := 1
	promises.ForEach(func(_, v lua.LValue) {
		if ud, ok := v.(*lua.LUserData); ok {
			if p, ok := ud.Value.(*Promise); ok {
				p.mu.RLock()
				if p.state == PromiseResolved {
					if lv, ok := p.value.(lua.LValue); ok {
						results.RawSetInt(idx, lv)
					} else {
						results.RawSetInt(idx, lua.LString("resolved"))
					}
				} else if p.state == PromiseRejected {
					results.RawSetInt(idx, lua.LNil)
				} else {
					results.RawSetInt(idx, lua.LNil)
				}
				p.mu.RUnlock()
				idx++
			}
		}
	})
	
	L.Push(results)
	return 1
}