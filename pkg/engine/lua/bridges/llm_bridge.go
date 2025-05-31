// ABOUTME: Implements the Lua-specific bridge adapter for LLM functionality
// ABOUTME: Exposes LLM operations as Lua functions with proper type conversions

package bridges

import (
	"context"
	"fmt"

	llmspellua "github.com/lexlapax/go-llmspell/pkg/engine/lua"
	lua "github.com/yuin/gopher-lua"
)

// LLMBridge wraps the core LLM bridge for Lua integration
type LLMBridge struct {
	bridge    LLMBridgeInterface
	converter *llmspellua.LuaConverter
}

// NewLLMBridge creates a new Lua LLM bridge
func NewLLMBridge(b LLMBridgeInterface) *LLMBridge {
	return &LLMBridge{
		bridge: b,
	}
}

// Register registers all LLM functions to the Lua state
func (lb *LLMBridge) Register(L *lua.LState) error {
	// Set the converter after L is available
	lb.converter = llmspellua.NewLuaConverter(L)

	// Create llm module table
	llmModule := L.NewTable()

	// Register functions
	L.SetField(llmModule, "chat", L.NewFunction(lb.chat))
	L.SetField(llmModule, "complete", L.NewFunction(lb.complete))
	L.SetField(llmModule, "stream_chat", L.NewFunction(lb.streamChat))
	L.SetField(llmModule, "list_models", L.NewFunction(lb.listModels))
	L.SetField(llmModule, "list_providers", L.NewFunction(lb.listProviders))
	L.SetField(llmModule, "get_provider", L.NewFunction(lb.getProvider))
	L.SetField(llmModule, "set_provider", L.NewFunction(lb.setProvider))

	// Register async functions
	L.SetField(llmModule, "chat_async", L.NewFunction(lb.chatAsync))
	L.SetField(llmModule, "complete_async", L.NewFunction(lb.completeAsync))

	// Register the module
	L.SetGlobal("llm", llmModule)

	return nil
}

// chat handles chat requests from Lua
// Usage: result, err = llm.chat(prompt)
func (lb *LLMBridge) chat(L *lua.LState) int {
	prompt := L.CheckString(1)

	// Call the bridge
	result, err := lb.bridge.Chat(context.Background(), prompt)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Return result
	L.Push(lua.LString(result))
	return 1
}

// complete handles text completion requests from Lua
// Usage: result, err = llm.complete(prompt, maxTokens)
func (lb *LLMBridge) complete(L *lua.LState) int {
	prompt := L.CheckString(1)
	maxTokens := L.OptInt(2, 0) // Optional maxTokens parameter

	// Call the bridge
	result, err := lb.bridge.Complete(context.Background(), prompt, maxTokens)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert result to Lua
	L.Push(lua.LString(result))
	return 1
}

// streamChat handles streaming chat requests from Lua
// Usage: err = llm.stream_chat(prompt, callback)
func (lb *LLMBridge) streamChat(L *lua.LState) int {
	prompt := L.CheckString(1)
	callback := L.CheckFunction(2)

	// Create a Go callback that calls the Lua callback
	goCallback := func(chunk string) error {
		// Push the callback and argument
		L.Push(callback)
		L.Push(lua.LString(chunk))

		// Call the Lua function
		if err := L.PCall(1, 1, nil); err != nil {
			return fmt.Errorf("lua callback error: %w", err)
		}

		// Check if callback returned an error
		if L.Get(-1).Type() != lua.LTNil {
			errStr := L.ToString(-1)
			L.Pop(1)
			if errStr != "" {
				return fmt.Errorf("%s", errStr)
			}
		}
		L.Pop(1)

		return nil
	}

	// Call the bridge
	err := lb.bridge.StreamChat(context.Background(), prompt, goCallback)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

// listModels returns available models
// Usage: models, err = llm.list_models()
func (lb *LLMBridge) listModels(L *lua.LState) int {
	models, err := lb.bridge.ListModels(context.Background())
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert to Lua table
	L.Push(lb.converter.ToLua(models))
	return 1
}

// listProviders returns available providers
// Usage: providers = llm.list_providers()
func (lb *LLMBridge) listProviders(L *lua.LState) int {
	providers := lb.bridge.ListProviders()
	L.Push(lb.converter.ToLua(providers))
	return 1
}

// getProvider returns the current provider name
// Usage: provider = llm.get_provider()
func (lb *LLMBridge) getProvider(L *lua.LState) int {
	provider := lb.bridge.GetCurrentProvider()
	L.Push(lua.LString(provider))
	return 1
}

// setProvider sets the current provider
// Usage: err = llm.set_provider(name)
func (lb *LLMBridge) setProvider(L *lua.LState) int {
	name := L.CheckString(1)

	err := lb.bridge.SetProvider(name)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}
