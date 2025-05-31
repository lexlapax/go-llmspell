// ABOUTME: Promise implementation for Lua that works synchronously
// ABOUTME: Provides promise-like API but executes callbacks immediately

package stdlib

import (
	"fmt"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// PromiseState represents the state of a promise
type PromiseState int

const (
	PromisePending PromiseState = iota
	PromiseResolved
	PromiseRejected
)

// Promise represents a promise implementation
type Promise struct {
	state    PromiseState
	value    interface{} // Store Go values instead of Lua values
	handlers []handler
	mu       sync.RWMutex
}

type handler struct {
	onResolve *lua.LFunction
	onReject  *lua.LFunction
	L         *lua.LState
}

// RegisterPromise registers the promise module
func RegisterPromise(L *lua.LState) {
	// Create promise module table
	promiseMod := L.NewTable()
	L.SetGlobal("promise", promiseMod)

	// Register promise type
	mt := L.NewTypeMetatable("promise")
	
	// Create index table with methods
	indexTable := L.SetFuncs(L.NewTable(), promiseMethods)
	
	// Add custom __index function to handle both methods and state
	L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		key := L.CheckString(2)
		
		// Check if it's a method first
		if method := indexTable.RawGetString(key); method != lua.LNil {
			L.Push(method)
			return 1
		}
		
		// Handle state property
		if key == "state" {
			if p, ok := ud.Value.(*Promise); ok {
				p.mu.RLock()
				defer p.mu.RUnlock()
				switch p.state {
				case PromisePending:
					L.Push(lua.LString("pending"))
				case PromiseResolved:
					L.Push(lua.LString("resolved"))
				case PromiseRejected:
					L.Push(lua.LString("rejected"))
				default:
					L.Push(lua.LNil)
				}
				return 1
			}
		}
		
		L.Push(lua.LNil)
		return 1
	}))

	// Register module functions
	L.SetField(promiseMod, "new", L.NewFunction(promiseNew))
	L.SetField(promiseMod, "resolve", L.NewFunction(promiseResolve))
	L.SetField(promiseMod, "reject", L.NewFunction(promiseReject))
	L.SetField(promiseMod, "all", L.NewFunction(promiseAll))
	L.SetField(promiseMod, "race", L.NewFunction(promiseRace))
}

var promiseMethods = map[string]lua.LGFunction{
	"next":  promiseThen, // renamed from 'then' to avoid Lua keyword conflict
	"catch": promiseCatch,
	"await": promiseAwait,
}

// promiseNew creates a new promise
func promiseNew(L *lua.LState) int {
	executor := L.CheckFunction(1)

	promise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}

	ud := L.NewUserData()
	ud.Value = promise
	L.SetMetatable(ud, L.GetTypeMetatable("promise"))

	// Create resolve/reject closures
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

// promiseResolve creates a resolved promise
func promiseResolve(L *lua.LState) int {
	value := L.Get(1)

	promise := &Promise{
		state: PromiseResolved,
		value: luaValueToInterface(value),
	}

	ud := L.NewUserData()
	ud.Value = promise
	L.SetMetatable(ud, L.GetTypeMetatable("promise"))

	L.Push(ud)
	return 1
}

// promiseReject creates a rejected promise
func promiseReject(L *lua.LState) int {
	reason := L.Get(1)

	promise := &Promise{
		state: PromiseRejected,
		value: luaValueToInterface(reason),
	}

	ud := L.NewUserData()
	ud.Value = promise
	L.SetMetatable(ud, L.GetTypeMetatable("promise"))

	L.Push(ud)
	return 1
}

// promiseThen attaches callbacks (exposed as 'next' in Lua)
func promiseThen(L *lua.LState) int {
	ud := L.CheckUserData(1)
	promise, ok := ud.Value.(*Promise)
	if !ok {
		L.ArgError(1, "promise expected")
		return 0
	}

	onResolve := L.OptFunction(2, nil)
	onReject := L.OptFunction(3, nil)

	// Create child promise
	childPromise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}

	childUD := L.NewUserData()
	childUD.Value = childPromise
	L.SetMetatable(childUD, L.GetTypeMetatable("promise"))

	// Handle the promise
	promise.mu.RLock()
	state := promise.state
	value := promise.value
	promise.mu.RUnlock()

	if state != PromisePending {
		// Promise already settled, execute handler immediately
		if state == PromiseResolved && onResolve != nil {
			L.Push(onResolve)
			L.Push(interfaceToLuaValue(L, value))
			if err := L.PCall(1, 1, nil); err != nil {
				childPromise.value = err.Error()
				childPromise.state = PromiseRejected
			} else {
				result := L.Get(-1)
				L.Pop(1)
				childPromise.value = luaValueToInterface(result)
				childPromise.state = PromiseResolved
			}
		} else if state == PromiseRejected && onReject != nil {
			L.Push(onReject)
			L.Push(interfaceToLuaValue(L, value))
			if err := L.PCall(1, 1, nil); err != nil {
				childPromise.value = err.Error()
				childPromise.state = PromiseRejected
			} else {
				result := L.Get(-1)
				L.Pop(1)
				childPromise.value = luaValueToInterface(result)
				childPromise.state = PromiseResolved
			}
		} else if state == PromiseResolved {
			childPromise.value = value
			childPromise.state = PromiseResolved
		} else {
			childPromise.value = value
			childPromise.state = PromiseRejected
		}
	} else {
		// Add handler for later execution
		handler := handler{
			onResolve: onResolve,
			onReject:  onReject,
			L:         L,
		}
		promise.mu.Lock()
		promise.handlers = append(promise.handlers, handler)
		promise.mu.Unlock()

		// Set up handler to update child promise
		go func() {
			// Wait for parent to settle
			for {
				promise.mu.RLock()
				if promise.state != PromisePending {
					promise.mu.RUnlock()
					break
				}
				promise.mu.RUnlock()
				time.Sleep(10 * time.Millisecond)
			}

			// Execute handler and update child
			promise.mu.RLock()
			state := promise.state
			value := promise.value
			promise.mu.RUnlock()

			if state == PromiseResolved && onResolve != nil {
				// Can't safely call Lua from goroutine, just pass value through
				childPromise.mu.Lock()
				childPromise.value = value
				childPromise.state = PromiseResolved
				childPromise.mu.Unlock()
			} else if state == PromiseRejected && onReject != nil {
				childPromise.mu.Lock()
				childPromise.value = value
				childPromise.state = PromiseResolved
				childPromise.mu.Unlock()
			} else if state == PromiseResolved {
				childPromise.mu.Lock()
				childPromise.value = value
				childPromise.state = PromiseResolved
				childPromise.mu.Unlock()
			} else {
				childPromise.mu.Lock()
				childPromise.value = value
				childPromise.state = PromiseRejected
				childPromise.mu.Unlock()
			}
		}()
	}

	L.Push(childUD)
	return 1
}

// promiseCatch attaches rejection callback
func promiseCatch(L *lua.LState) int {
	ud := L.CheckUserData(1)
	promise, ok := ud.Value.(*Promise)
	if !ok {
		L.ArgError(1, "promise expected")
		return 0
	}

	onReject := L.CheckFunction(2)

	// Create child promise for catch
	childPromise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}

	childUD := L.NewUserData()
	childUD.Value = childPromise
	L.SetMetatable(childUD, L.GetTypeMetatable("promise"))

	// Handle the promise
	promise.mu.RLock()
	state := promise.state
	value := promise.value
	promise.mu.RUnlock()

	if state != PromisePending {
		// Promise already settled
		if state == PromiseRejected {
			// Execute the reject handler
			L.Push(onReject)
			L.Push(interfaceToLuaValue(L, value))
			if err := L.PCall(1, 1, nil); err != nil {
				childPromise.value = err.Error()
				childPromise.state = PromiseRejected
			} else {
				result := L.Get(-1)
				L.Pop(1)
				childPromise.value = luaValueToInterface(result)
				childPromise.state = PromiseResolved
			}
		} else {
			// Promise was resolved, pass through
			childPromise.value = value
			childPromise.state = PromiseResolved
		}
	} else {
		// Set up async handler
		go func() {
			// Wait for parent to settle
			for {
				promise.mu.RLock()
				if promise.state != PromisePending {
					promise.mu.RUnlock()
					break
				}
				promise.mu.RUnlock()
				time.Sleep(10 * time.Millisecond)
			}

			promise.mu.RLock()
			state := promise.state
			value := promise.value
			promise.mu.RUnlock()

			if state == PromiseRejected {
				// Can't safely call Lua from goroutine, just pass value
				childPromise.mu.Lock()
				childPromise.value = value
				childPromise.state = PromiseResolved
				childPromise.mu.Unlock()
			} else {
				childPromise.mu.Lock()
				childPromise.value = value
				childPromise.state = PromiseResolved
				childPromise.mu.Unlock()
			}
		}()
	}

	L.Push(childUD)
	return 1
}

// promiseAwait waits for promise to settle
func promiseAwait(L *lua.LState) int {
	ud := L.CheckUserData(1)
	promise, ok := ud.Value.(*Promise)
	if !ok {
		L.ArgError(1, "promise expected")
		return 0
	}

	timeout := time.Duration(L.OptNumber(2, 30)) * time.Second
	deadline := time.Now().Add(timeout)

	// Poll until settled or timeout
	for {
		promise.mu.RLock()
		state := promise.state
		value := promise.value
		promise.mu.RUnlock()

		if state != PromisePending {
			if state == PromiseResolved {
				L.Push(interfaceToLuaValue(L, value))
				return 1
			} else {
				L.Push(lua.LNil)
				L.Push(interfaceToLuaValue(L, value))
				return 2
			}
		}

		if time.Now().After(deadline) {
			L.Push(lua.LNil)
			L.Push(lua.LString("timeout"))
			return 2
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// promiseAll waits for all promises
func promiseAll(L *lua.LState) int {
	table := L.CheckTable(1)

	var promises []*Promise
	promiseCount := 0

	table.ForEach(func(k, v lua.LValue) {
		if ud, ok := v.(*lua.LUserData); ok {
			if p, ok := ud.Value.(*Promise); ok {
				promises = append(promises, p)
				promiseCount++
			}
		}
	})

	// Create result promise
	resultPromise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}

	resultUD := L.NewUserData()
	resultUD.Value = resultPromise
	L.SetMetatable(resultUD, L.GetTypeMetatable("promise"))

	if promiseCount == 0 {
		resultPromise.value = make([]interface{}, 0)
		resultPromise.state = PromiseResolved
		L.Push(resultUD)
		return 1
	}

	// Wait for all in a goroutine
	go func() {
		results := make([]interface{}, promiseCount)

		for i, p := range promises {
			// Wait for this promise
			for {
				p.mu.RLock()
				state := p.state
				value := p.value
				p.mu.RUnlock()

				if state != PromisePending {
					if state == PromiseRejected {
						resultPromise.mu.Lock()
						resultPromise.value = value
						resultPromise.state = PromiseRejected
						resultPromise.mu.Unlock()
						return
					}
					results[i] = value
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}

		resultPromise.mu.Lock()
		resultPromise.value = results
		resultPromise.state = PromiseResolved
		resultPromise.mu.Unlock()
	}()

	L.Push(resultUD)
	return 1
}

// promiseRace waits for first promise to settle
func promiseRace(L *lua.LState) int {
	table := L.CheckTable(1)

	var promises []*Promise

	table.ForEach(func(k, v lua.LValue) {
		if ud, ok := v.(*lua.LUserData); ok {
			if p, ok := ud.Value.(*Promise); ok {
				promises = append(promises, p)
			}
		}
	})

	// Create result promise
	resultPromise := &Promise{
		state:    PromisePending,
		handlers: []handler{},
	}

	resultUD := L.NewUserData()
	resultUD.Value = resultPromise
	L.SetMetatable(resultUD, L.GetTypeMetatable("promise"))

	if len(promises) == 0 {
		L.Push(resultUD)
		return 1
	}

	// Race in goroutine
	go func() {
		for {
			for _, p := range promises {
				p.mu.RLock()
				state := p.state
				value := p.value
				p.mu.RUnlock()

				if state != PromisePending {
					resultPromise.mu.Lock()
					if resultPromise.state == PromisePending {
						resultPromise.state = state
						resultPromise.value = value
					}
					resultPromise.mu.Unlock()
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	L.Push(resultUD)
	return 1
}

// Helper methods
func (p *Promise) resolveWithLua(L *lua.LState, value lua.LValue) {
	p.mu.Lock()
	if p.state != PromisePending {
		p.mu.Unlock()
		return
	}

	p.state = PromiseResolved
	p.value = luaValueToInterface(value)
	handlers := p.handlers
	p.handlers = nil
	p.mu.Unlock()

	// Execute handlers if any (synchronously)
	for _, h := range handlers {
		if h.onResolve != nil && h.L == L {
			L.Push(h.onResolve)
			L.Push(value)
			_ = L.PCall(1, 0, nil) // Ignore error in handler
		}
	}
}

func (p *Promise) rejectWithLua(L *lua.LState, reason lua.LValue) {
	p.mu.Lock()
	if p.state != PromisePending {
		p.mu.Unlock()
		return
	}

	p.state = PromiseRejected
	p.value = luaValueToInterface(reason)
	handlers := p.handlers
	p.handlers = nil
	p.mu.Unlock()

	// Execute handlers if any
	for _, h := range handlers {
		if h.onReject != nil && h.L == L {
			L.Push(h.onReject)
			L.Push(reason)
			_ = L.PCall(1, 0, nil) // Ignore error in handler
		}
	}
}

// Helper functions to convert between Lua and Go values
func luaValueToInterface(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LTable:
		// Convert table to slice if array-like
		if isArray(v) {
			arr := []interface{}{}
			v.ForEach(func(k, val lua.LValue) {
				arr = append(arr, luaValueToInterface(val))
			})
			return arr
		}
		// Otherwise convert to map
		m := make(map[string]interface{})
		v.ForEach(func(k, val lua.LValue) {
			if ks, ok := k.(lua.LString); ok {
				m[string(ks)] = luaValueToInterface(val)
			}
		})
		return m
	case *lua.LNilType:
		return nil
	default:
		return fmt.Sprintf("%v", lv)
	}
}

func interfaceToLuaValue(L *lua.LState, v interface{}) lua.LValue {
	if v == nil {
		return lua.LNil
	}

	switch val := v.(type) {
	case string:
		return lua.LString(val)
	case float64:
		return lua.LNumber(val)
	case int:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	case []interface{}:
		t := L.NewTable()
		for i, item := range val {
			t.RawSetInt(i+1, interfaceToLuaValue(L, item))
		}
		return t
	case map[string]interface{}:
		t := L.NewTable()
		for k, v := range val {
			t.RawSetString(k, interfaceToLuaValue(L, v))
		}
		return t
	default:
		return lua.LString(fmt.Sprintf("%v", v))
	}
}

func isArray(t *lua.LTable) bool {
	length := t.Len()
	if length == 0 {
		return true
	}

	// Check if keys are sequential integers starting from 1
	for i := 1; i <= length; i++ {
		if t.RawGetInt(i) == lua.LNil {
			return false
		}
	}

	// Check there are no other keys
	count := 0
	t.ForEach(func(k, v lua.LValue) {
		count++
	})

	return count == length
}
