// ABOUTME: Async extensions for the LLM bridge
// ABOUTME: Provides non-blocking LLM operations using callbacks

package bridges

import (
	"context"

	"github.com/lexlapax/go-llmspell/pkg/engine/lua/stdlib"
	lua "github.com/yuin/gopher-lua"
)

// chatAsync handles async chat requests from Lua
// Usage: id = llm.chat_async(prompt, callback, errback)
func (lb *LLMBridge) chatAsync(L *lua.LState) int {
	prompt := L.CheckString(1)
	callback := L.CheckFunction(2)
	errback := L.OptFunction(3, nil)

	// Get callback manager
	mgr := stdlib.GetCallbackManager(L)

	// Register callback
	id := mgr.RegisterCallback(callback, errback)

	// Start async operation
	go func() {
		result, err := lb.bridge.Chat(context.Background(), prompt)
		if err != nil {
			mgr.QueueError(id, err.Error())
		} else {
			mgr.QueueStringResult(id, result)
		}
	}()

	// Return callback ID
	L.Push(lua.LNumber(id))
	return 1
}

// completeAsync handles async completion requests from Lua
// Usage: id = llm.complete_async(prompt, maxTokens, callback, errback)
func (lb *LLMBridge) completeAsync(L *lua.LState) int {
	prompt := L.CheckString(1)
	maxTokens := L.CheckInt(2)
	callback := L.CheckFunction(3)
	errback := L.OptFunction(4, nil)

	// Get callback manager
	mgr := stdlib.GetCallbackManager(L)

	// Register callback
	id := mgr.RegisterCallback(callback, errback)

	// Start async operation
	go func() {
		result, err := lb.bridge.Complete(context.Background(), prompt, maxTokens)
		if err != nil {
			mgr.QueueError(id, err.Error())
		} else {
			mgr.QueueStringResult(id, result)
		}
	}()

	// Return callback ID
	L.Push(lua.LNumber(id))
	return 1
}
