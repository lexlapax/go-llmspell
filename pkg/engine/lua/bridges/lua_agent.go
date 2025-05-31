// ABOUTME: Wraps Lua functions/tables as agents for the agent system
// ABOUTME: Enables Lua scripts to define custom agent implementations

package bridges

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/agents"
	engLua "github.com/lexlapax/go-llmspell/pkg/engine/lua"
	lua "github.com/yuin/gopher-lua"
)

// LuaAgent wraps a Lua function or table as an agent
type LuaAgent struct {
	name      string
	luaValue  lua.LValue
	L         *lua.LState
	mu        sync.Mutex
	converter *engLua.LuaConverter
}

// NewLuaAgent creates a new agent from a Lua function or table
func NewLuaAgent(name string, luaValue lua.LValue, L *lua.LState) *LuaAgent {
	agent := &LuaAgent{
		name:      name,
		luaValue:  luaValue,
		L:         L,
		converter: engLua.NewLuaConverter(L),
	}

	// If name is empty and luaValue is a table, try to get name from table
	if name == "" && luaValue.Type() == lua.LTTable {
		if nameField := L.GetField(luaValue, "name"); nameField.Type() == lua.LTString {
			agent.name = lua.LVAsString(nameField)
		}
	}

	return agent
}

// Name returns the agent name
func (la *LuaAgent) Name() string {
	return la.name
}

// Initialize prepares the agent for use
func (la *LuaAgent) Initialize(ctx context.Context) error {
	// Lua agents don't need special initialization
	return nil
}

// Cleanup releases any resources held by the agent
func (la *LuaAgent) Cleanup() error {
	// Lua agents don't need special cleanup
	return nil
}

// Execute runs the agent with a single input
func (la *LuaAgent) Execute(ctx context.Context, input string, options *agents.ExecutionOptions) (*agents.ExecutionResult, error) {
	la.mu.Lock()
	defer la.mu.Unlock()

	switch la.luaValue.Type() {
	case lua.LTFunction:
		return la.executeFunction(input, options)
	case lua.LTTable:
		return la.executeTable(input, options)
	default:
		return nil, errors.New("execute not implemented for this Lua type")
	}
}

// executeFunction handles execution when the agent is a Lua function
func (la *LuaAgent) executeFunction(input string, options *agents.ExecutionOptions) (*agents.ExecutionResult, error) {
	// Convert options to Lua
	optionsLua := la.convertOptionsToLua(options)

	// Call the function
	err := la.L.CallByParam(lua.P{
		Fn:      la.luaValue.(*lua.LFunction),
		NRet:    2,
		Protect: true,
	}, lua.LString(input), optionsLua)

	if err != nil {
		return nil, fmt.Errorf("lua execution error: %w", err)
	}

	// Get return values
	ret2 := la.L.Get(-1) // error (if any)
	ret1 := la.L.Get(-2) // result
	la.L.Pop(2)

	// Check for error
	if ret2.Type() == lua.LTString {
		return nil, errors.New(lua.LVAsString(ret2))
	}

	// Convert result to string
	result := la.convertResultToString(ret1)

	return &agents.ExecutionResult{
		Response: result,
	}, nil
}

// executeTable handles execution when the agent is a Lua table with methods
func (la *LuaAgent) executeTable(input string, options *agents.ExecutionOptions) (*agents.ExecutionResult, error) {
	// Get the execute method
	executeMethod := la.L.GetField(la.luaValue, "execute")
	if executeMethod.Type() != lua.LTFunction {
		return nil, errors.New("agent table must have an 'execute' method")
	}

	// Convert options to Lua
	optionsLua := la.convertOptionsToLua(options)

	// Call the method with self
	err := la.L.CallByParam(lua.P{
		Fn:      executeMethod.(*lua.LFunction),
		NRet:    2,
		Protect: true,
	}, la.luaValue, lua.LString(input), optionsLua)

	if err != nil {
		return nil, fmt.Errorf("lua execution error: %w", err)
	}

	// Get return values
	ret2 := la.L.Get(-1) // error (if any)
	ret1 := la.L.Get(-2) // result
	la.L.Pop(2)

	// Check for error
	if ret2.Type() == lua.LTString {
		return nil, errors.New(lua.LVAsString(ret2))
	}

	// Convert result to string
	result := la.convertResultToString(ret1)

	return &agents.ExecutionResult{
		Response: result,
	}, nil
}

// Stream executes the agent with streaming response
func (la *LuaAgent) Stream(ctx context.Context, input string, options *agents.ExecutionOptions, callback agents.StreamCallback) error {
	la.mu.Lock()
	defer la.mu.Unlock()

	if la.luaValue.Type() != lua.LTTable {
		return errors.New("streaming only supported for table-based agents")
	}

	// Get the stream method
	streamMethod := la.L.GetField(la.luaValue, "stream")
	if streamMethod.Type() != lua.LTFunction {
		return errors.New("agent table must have a 'stream' method for streaming")
	}

	// Convert options to Lua
	optionsLua := la.convertOptionsToLua(options)

	// Create Lua callback function
	luaCallback := la.L.NewFunction(func(L *lua.LState) int {
		chunk := L.CheckString(1)
		err := callback(chunk)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		L.Push(lua.LNil)
		return 1
	})

	// Call the stream method
	err := la.L.CallByParam(lua.P{
		Fn:      streamMethod.(*lua.LFunction),
		NRet:    1,
		Protect: true,
	}, la.luaValue, lua.LString(input), optionsLua, luaCallback)

	if err != nil {
		return fmt.Errorf("lua stream error: %w", err)
	}

	// Check return value for error
	ret := la.L.Get(-1)
	la.L.Pop(1)

	if ret.Type() == lua.LTString {
		return errors.New(lua.LVAsString(ret))
	}

	return nil
}

// ExecuteWithHistory runs the agent with conversation history
func (la *LuaAgent) ExecuteWithHistory(ctx context.Context, messages []agents.Message, opts *agents.ExecutionOptions) (*agents.ExecutionResult, error) {
	// For now, convert messages to a single input string
	// In a more advanced implementation, we'd pass the full message history to Lua
	var input string
	if len(messages) > 0 {
		input = messages[len(messages)-1].Content
	}
	return la.Execute(ctx, input, opts)
}

// StreamWithHistory executes with history and streaming response
func (la *LuaAgent) StreamWithHistory(ctx context.Context, messages []agents.Message, opts *agents.ExecutionOptions, callback agents.StreamCallback) error {
	// For now, convert messages to a single input string
	// In a more advanced implementation, we'd pass the full message history to Lua
	var input string
	if len(messages) > 0 {
		input = messages[len(messages)-1].Content
	}
	return la.Stream(ctx, input, opts, callback)
}

// GetSystemPrompt returns the agent's system prompt
func (la *LuaAgent) GetSystemPrompt() string {
	la.mu.Lock()
	defer la.mu.Unlock()

	if la.luaValue.Type() != lua.LTTable {
		return ""
	}

	// Check for get_system_prompt method
	getMethod := la.L.GetField(la.luaValue, "get_system_prompt")
	if getMethod.Type() == lua.LTFunction {
		err := la.L.CallByParam(lua.P{
			Fn:      getMethod.(*lua.LFunction),
			NRet:    1,
			Protect: true,
		}, la.luaValue)

		if err == nil {
			ret := la.L.Get(-1)
			la.L.Pop(1)
			if ret.Type() == lua.LTString {
				return lua.LVAsString(ret)
			}
		}
	}

	// Check for system_prompt field
	field := la.L.GetField(la.luaValue, "system_prompt")
	if field.Type() == lua.LTString {
		return lua.LVAsString(field)
	}

	return ""
}

// SetSystemPrompt updates the agent's system prompt
func (la *LuaAgent) SetSystemPrompt(prompt string) {
	la.mu.Lock()
	defer la.mu.Unlock()

	if la.luaValue.Type() != lua.LTTable {
		return
	}

	// Check for set_system_prompt method
	setMethod := la.L.GetField(la.luaValue, "set_system_prompt")
	if setMethod.Type() == lua.LTFunction {
		_ = la.L.CallByParam(lua.P{
			Fn:      setMethod.(*lua.LFunction),
			NRet:    0,
			Protect: true,
		}, la.luaValue, lua.LString(prompt))
		return
	}

	// Otherwise, set the field directly
	la.L.SetField(la.luaValue, "system_prompt", lua.LString(prompt))
}

// GetTools returns the list of tools available to the agent
func (la *LuaAgent) GetTools() []string {
	la.mu.Lock()
	defer la.mu.Unlock()

	if la.luaValue.Type() != lua.LTTable {
		return []string{}
	}

	// Check for get_tools method
	getMethod := la.L.GetField(la.luaValue, "get_tools")
	if getMethod.Type() == lua.LTFunction {
		err := la.L.CallByParam(lua.P{
			Fn:      getMethod.(*lua.LFunction),
			NRet:    1,
			Protect: true,
		}, la.luaValue)

		if err == nil {
			ret := la.L.Get(-1)
			la.L.Pop(1)
			return la.convertLuaArrayToStringSlice(ret)
		}
	}

	// Check for tools field
	field := la.L.GetField(la.luaValue, "tools")
	return la.convertLuaArrayToStringSlice(field)
}

// AddTool adds a tool to the agent
func (la *LuaAgent) AddTool(toolName string) error {
	la.mu.Lock()
	defer la.mu.Unlock()

	if la.luaValue.Type() != lua.LTTable {
		return errors.New("add_tool only supported for table-based agents")
	}

	// Check for add_tool method
	addMethod := la.L.GetField(la.luaValue, "add_tool")
	if addMethod.Type() == lua.LTFunction {
		err := la.L.CallByParam(lua.P{
			Fn:      addMethod.(*lua.LFunction),
			NRet:    1,
			Protect: true,
		}, la.luaValue, lua.LString(toolName))

		if err != nil {
			return fmt.Errorf("lua add_tool error: %w", err)
		}

		// Check return value for error
		ret := la.L.Get(-1)
		la.L.Pop(1)
		if ret.Type() == lua.LTString {
			return errors.New(lua.LVAsString(ret))
		}

		return nil
	}

	return errors.New("agent table must have an 'add_tool' method")
}

// RemoveTool removes a tool from the agent
func (la *LuaAgent) RemoveTool(toolName string) error {
	la.mu.Lock()
	defer la.mu.Unlock()

	if la.luaValue.Type() != lua.LTTable {
		return errors.New("remove_tool only supported for table-based agents")
	}

	// Check for remove_tool method
	removeMethod := la.L.GetField(la.luaValue, "remove_tool")
	if removeMethod.Type() == lua.LTFunction {
		err := la.L.CallByParam(lua.P{
			Fn:      removeMethod.(*lua.LFunction),
			NRet:    1,
			Protect: true,
		}, la.luaValue, lua.LString(toolName))

		if err != nil {
			return fmt.Errorf("lua remove_tool error: %w", err)
		}

		// Check return value for error
		ret := la.L.Get(-1)
		la.L.Pop(1)
		if ret.Type() == lua.LTString {
			return errors.New(lua.LVAsString(ret))
		}

		return nil
	}

	return errors.New("agent table must have a 'remove_tool' method")
}

// convertOptionsToLua converts ExecutionOptions to a Lua table
func (la *LuaAgent) convertOptionsToLua(options *agents.ExecutionOptions) lua.LValue {
	if options == nil {
		return lua.LNil
	}

	optionsMap := make(map[string]interface{})
	if options.MaxTokens > 0 {
		optionsMap["max_tokens"] = options.MaxTokens
	}
	if options.Temperature > 0 {
		optionsMap["temperature"] = options.Temperature
	}
	optionsMap["stream"] = options.Stream
	if options.Timeout > 0 {
		optionsMap["timeout"] = options.Timeout.Seconds()
	}

	return la.converter.ToLua(optionsMap)
}

// convertResultToString converts a Lua value to a string result
func (la *LuaAgent) convertResultToString(result lua.LValue) string {
	switch result.Type() {
	case lua.LTString:
		return lua.LVAsString(result)
	case lua.LTNumber:
		return lua.LVAsString(result)
	case lua.LTBool:
		if lua.LVAsBool(result) {
			return "true"
		}
		return "false"
	case lua.LTTable:
		// Convert table to JSON
		goValue := la.converter.ToInterface(result)
		jsonBytes, err := json.Marshal(goValue)
		if err != nil {
			return fmt.Sprintf("error converting result: %v", err)
		}
		return string(jsonBytes)
	default:
		return ""
	}
}

// convertLuaArrayToStringSlice converts a Lua table/array to a string slice
func (la *LuaAgent) convertLuaArrayToStringSlice(lv lua.LValue) []string {
	if lv.Type() != lua.LTTable {
		return []string{}
	}

	table := lv.(*lua.LTable)
	var result []string

	table.ForEach(func(_, value lua.LValue) {
		if value.Type() == lua.LTString {
			result = append(result, lua.LVAsString(value))
		}
	})

	return result
}
