// ABOUTME: Lua-based tool implementation that properly handles function calls
// ABOUTME: Stores Lua functions and executes them in the correct context

package bridges

import (
	"context"
	"fmt"
	"sync"

	engLua "github.com/lexlapax/go-llmspell/pkg/engine/lua"
	lua "github.com/yuin/gopher-lua"
)

// LuaTool wraps a Lua function as a tool
type LuaTool struct {
	name        string
	description string
	parameters  map[string]interface{}
	fn          *lua.LFunction
	L           *lua.LState
	converter   *engLua.LuaConverter
	mu          sync.Mutex
}

// NewLuaTool creates a new Lua-based tool
func NewLuaTool(name, description string, parameters map[string]interface{}, fn *lua.LFunction, L *lua.LState, converter *engLua.LuaConverter) *LuaTool {
	return &LuaTool{
		name:        name,
		description: description,
		parameters:  parameters,
		fn:          fn,
		L:           L,
		converter:   converter,
	}
}

// Execute runs the Lua tool function
func (lt *LuaTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	// Save current stack state
	oldTop := lt.L.GetTop()
	defer func() {
		// Restore stack state
		lt.L.SetTop(oldTop)
	}()

	// Push the function and parameters
	lt.L.Push(lt.fn)
	lt.L.Push(lt.converter.ToLua(params))

	// Call the function
	if err := lt.L.PCall(1, lua.MultRet, nil); err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Get the results
	nResults := lt.L.GetTop() - oldTop
	if nResults == 0 {
		return nil, nil
	}

	// Get the first result (from top of stack)
	result := lt.L.Get(-nResults)

	// Check if there's an error as second return value
	if nResults >= 2 {
		if errVal := lt.L.Get(-nResults + 1); errVal.Type() == lua.LTString {
			return nil, fmt.Errorf("%s", errVal.String())
		}
	}

	// Convert result to Go value
	return lt.converter.ToInterface(result), nil
}
