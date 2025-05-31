// ABOUTME: Async callback system for non-blocking operations
// ABOUTME: Enables true parallel execution without multiple Lua states

package stdlib

import (
	"sync"
	"sync/atomic"
	
	lua "github.com/yuin/gopher-lua"
)

// CallbackManager manages async callbacks for a Lua state
type CallbackManager struct {
	mu        sync.Mutex
	callbacks map[int64]*pendingCallback
	results   chan *callbackResult
	nextID    int64
	L         *lua.LState
}

type pendingCallback struct {
	id       int64
	callback *lua.LFunction
	errback  *lua.LFunction
}

type callbackResult struct {
	id    int64
	value lua.LValue
	err   string
}

// Global callback manager per Lua state
var (
	managersMu sync.Mutex
	managers   = make(map[*lua.LState]*CallbackManager)
)

// GetCallbackManager returns the callback manager for a Lua state
func GetCallbackManager(L *lua.LState) *CallbackManager {
	managersMu.Lock()
	defer managersMu.Unlock()
	
	if mgr, exists := managers[L]; exists {
		return mgr
	}
	
	// Create new manager
	mgr := &CallbackManager{
		callbacks: make(map[int64]*pendingCallback),
		results:   make(chan *callbackResult, 100),
		L:         L,
	}
	managers[L] = mgr
	return mgr
}

// RegisterAsyncCallback registers the async callback module
func RegisterAsyncCallback(L *lua.LState) {
	mgr := GetCallbackManager(L)
	
	asyncMod := L.NewTable()
	L.SetField(asyncMod, "process_callbacks", L.NewFunction(mgr.processCallbacks))
	L.SetField(asyncMod, "pending_count", L.NewFunction(mgr.pendingCount))
	L.SetField(asyncMod, "create_callback", L.NewFunction(mgr.createCallback))
	L.SetGlobal("async", asyncMod)
}

// RegisterCallback registers a new callback and returns its ID
func (cm *CallbackManager) RegisterCallback(callback, errback *lua.LFunction) int64 {
	id := atomic.AddInt64(&cm.nextID, 1)
	
	cm.mu.Lock()
	cm.callbacks[id] = &pendingCallback{
		id:       id,
		callback: callback,
		errback:  errback,
	}
	cm.mu.Unlock()
	
	return id
}

// QueueResult queues a result for a callback
func (cm *CallbackManager) QueueResult(id int64, value lua.LValue) {
	select {
	case cm.results <- &callbackResult{id: id, value: value}:
	default:
		// Queue is full, drop result
	}
}

// QueueError queues an error for a callback
func (cm *CallbackManager) QueueError(id int64, err string) {
	select {
	case cm.results <- &callbackResult{id: id, err: err}:
	default:
		// Queue is full, drop result
	}
}

// QueueStringResult is a convenience method for string results
func (cm *CallbackManager) QueueStringResult(id int64, result string) {
	cm.QueueResult(id, lua.LString(result))
}

// processCallbacks processes pending callbacks
func (cm *CallbackManager) processCallbacks(L *lua.LState) int {
	processed := 0
	
	// Process all available results
	for {
		select {
		case result := <-cm.results:
			cm.mu.Lock()
			cb, exists := cm.callbacks[result.id]
			if exists {
				delete(cm.callbacks, result.id)
			}
			cm.mu.Unlock()
			
			if !exists {
				continue
			}
			
			// Execute callback
			if result.err != "" && cb.errback != nil {
				L.Push(cb.errback)
				L.Push(lua.LString(result.err))
				if err := L.PCall(1, 0, nil); err != nil {
					// Ignore callback errors for now
				}
			} else if result.err == "" && cb.callback != nil {
				L.Push(cb.callback)
				if result.value != nil {
					L.Push(result.value)
				} else {
					L.Push(lua.LNil)
				}
				if err := L.PCall(1, 0, nil); err != nil {
					// Ignore callback errors for now
				}
			}
			
			processed++
			
		default:
			// No more results
			L.Push(lua.LNumber(processed))
			return 1
		}
	}
}

// pendingCount returns the number of pending callbacks
func (cm *CallbackManager) pendingCount(L *lua.LState) int {
	cm.mu.Lock()
	count := len(cm.callbacks)
	cm.mu.Unlock()
	
	L.Push(lua.LNumber(count))
	return 1
}

// createCallback creates a callback pair for testing
func (cm *CallbackManager) createCallback(L *lua.LState) int {
	callback := L.CheckFunction(1)
	errback := L.OptFunction(2, nil)
	
	id := cm.RegisterCallback(callback, errback)
	
	L.Push(lua.LNumber(id))
	return 1
}

// Cleanup removes the manager when Lua state is closed
func CleanupCallbackManager(L *lua.LState) {
	managersMu.Lock()
	defer managersMu.Unlock()
	
	if mgr, exists := managers[L]; exists {
		close(mgr.results)
		delete(managers, L)
	}
}