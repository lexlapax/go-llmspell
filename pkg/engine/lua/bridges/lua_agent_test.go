// ABOUTME: Tests for Lua function wrapping as agents
// ABOUTME: Validates execution, streaming, tool integration, and error handling

package bridges

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestLuaAgentBasicExecution(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a Lua function that acts as an agent
	err := L.DoString(`
		function simple_agent(input, options)
			return "Echo: " .. input
		end
	`)
	require.NoError(t, err)

	// Get the function
	fn := L.GetGlobal("simple_agent")
	require.Equal(t, lua.LTFunction, fn.Type())

	// Create the Lua agent
	agent := NewLuaAgent("test-agent", fn.(*lua.LFunction), L)

	// Test basic properties
	assert.Equal(t, "test-agent", agent.Name())
	assert.Equal(t, "", agent.GetSystemPrompt())
	assert.Empty(t, agent.GetTools())

	// Test execution
	result, err := agent.Execute(context.Background(), "Hello, World!", nil)
	require.NoError(t, err)
	assert.Equal(t, "Echo: Hello, World!", result.Response)
}

func TestLuaAgentWithTable(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a Lua table that implements agent methods
	err := L.DoString(`
		test_agent = {
			name = "custom-agent",
			system_prompt = "You are a helpful assistant",
			tools = {"calculator", "search"},
			
			execute = function(self, input, options)
				local prefix = options and options.prefix or "Response"
				return prefix .. ": " .. input
			end,
			
			stream = function(self, input, options, callback)
				local words = {}
				for word in string.gmatch(input, "%S+") do
					table.insert(words, word)
				end
				
				for i, word in ipairs(words) do
					local err = callback(word .. " ")
					if err then
						return err
					end
				end
				return nil
			end
		}
	`)
	require.NoError(t, err)

	// Get the table
	agentTable := L.GetGlobal("test_agent")
	require.Equal(t, lua.LTTable, agentTable.Type())

	// Create the Lua agent
	agent := NewLuaAgent("", agentTable, L)

	// Test properties from table
	assert.Equal(t, "custom-agent", agent.Name())
	assert.Equal(t, "You are a helpful assistant", agent.GetSystemPrompt())
	assert.Equal(t, []string{"calculator", "search"}, agent.GetTools())

	// Test execution with options
	opts := &agents.ExecutionOptions{
		MaxTokens: 100,
	}
	result, err := agent.Execute(context.Background(), "Hello", opts)
	require.NoError(t, err)
	assert.Equal(t, "Response: Hello", result.Response)

	// Test streaming
	var chunks []string
	err = agent.Stream(context.Background(), "Hello World Test", nil, func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"Hello ", "World ", "Test "}, chunks)
}

func TestLuaAgentSystemPromptMethods(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a Lua table with get/set system prompt methods
	err := L.DoString(`
		test_agent = {
			_system_prompt = "Initial prompt",
			
			get_system_prompt = function(self)
				return self._system_prompt
			end,
			
			set_system_prompt = function(self, prompt)
				self._system_prompt = prompt
			end,
			
			execute = function(self, input, options)
				return "Using prompt: " .. self._system_prompt
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("test_agent")
	agent := NewLuaAgent("test", agentTable, L)

	// Test initial prompt
	assert.Equal(t, "Initial prompt", agent.GetSystemPrompt())

	// Test setting prompt
	agent.SetSystemPrompt("New prompt")
	assert.Equal(t, "New prompt", agent.GetSystemPrompt())

	// Verify it affects execution
	result, err := agent.Execute(context.Background(), "test", nil)
	require.NoError(t, err)
	assert.Equal(t, "Using prompt: New prompt", result.Response)
}

func TestLuaAgentToolManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a Lua table with tool management
	err := L.DoString(`
		test_agent = {
			_tools = {},
			
			get_tools = function(self)
				return self._tools
			end,
			
			add_tool = function(self, tool_name)
				table.insert(self._tools, tool_name)
				return nil
			end,
			
			remove_tool = function(self, tool_name)
				for i, tool in ipairs(self._tools) do
					if tool == tool_name then
						table.remove(self._tools, i)
						return nil
					end
				end
				return "tool not found"
			end,
			
			execute = function(self, input, options)
				return "Tools: " .. table.concat(self._tools, ", ")
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("test_agent")
	agent := NewLuaAgent("test", agentTable, L)

	// Test initial state
	assert.Empty(t, agent.GetTools())

	// Add tools
	err = agent.AddTool("calculator")
	require.NoError(t, err)
	err = agent.AddTool("search")
	require.NoError(t, err)
	assert.Equal(t, []string{"calculator", "search"}, agent.GetTools())

	// Test RemoveTool (not in the Agent interface, but available on LuaAgent)
	// Remove tool
	err = agent.RemoveTool("calculator")
	require.NoError(t, err)
	assert.Equal(t, []string{"search"}, agent.GetTools())

	// Try to remove non-existent tool
	err = agent.RemoveTool("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool not found")
}

func TestLuaAgentErrorHandling(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create an agent that returns errors
	err := L.DoString(`
		error_agent = {
			execute = function(self, input, options)
				if input == "error" then
					return nil, "execution failed"
				end
				return "Success: " .. input
			end,
			
			stream = function(self, input, options, callback)
				if input == "error" then
					return "stream failed"
				end
				return callback("chunk")
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("error_agent")
	agent := NewLuaAgent("error-test", agentTable, L)

	// Test execution error
	_, err = agent.Execute(context.Background(), "error", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execution failed")

	// Test successful execution
	result, err := agent.Execute(context.Background(), "test", nil)
	require.NoError(t, err)
	assert.Equal(t, "Success: test", result.Response)

	// Test stream error
	err = agent.Stream(context.Background(), "error", nil, func(chunk string) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream failed")
}

func TestLuaAgentConcurrency(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a stateful agent
	err := L.DoString(`
		concurrent_agent = {
			execute = function(self, input, options)
				-- Simulate some work
				local sum = 0
				for i = 1, 1000 do
					sum = sum + i
				end
				return "Processed: " .. input .. " (sum=" .. sum .. ")"
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("concurrent_agent")
	agent := NewLuaAgent("concurrent", agentTable, L)

	// Run multiple concurrent executions
	var wg sync.WaitGroup
	errors := make(chan error, 10)
	results := make(chan string, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err := agent.Execute(context.Background(), string(rune('A'+id)), nil)
			if err != nil {
				errors <- err
				return
			}
			results <- result.Response
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent execution error: %v", err)
	}

	// Verify all results
	resultCount := 0
	for result := range results {
		resultCount++
		assert.Contains(t, result, "Processed: ")
		assert.Contains(t, result, "(sum=500500)")
	}
	assert.Equal(t, 10, resultCount)
}

func TestLuaAgentWithOptionsConversion(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create an agent that uses options
	err := L.DoString(`
		options_agent = {
			execute = function(self, input, options)
				if not options then
					return "No options provided"
				end
				
				local parts = {"Input: " .. input}
				
				if options.max_tokens then
					table.insert(parts, "MaxTokens: " .. options.max_tokens)
				end
				
				if options.temperature then
					table.insert(parts, "Temperature: " .. options.temperature)
				end
				
				if options.stream ~= nil then
					table.insert(parts, "Stream: " .. tostring(options.stream))
				end
				
				return table.concat(parts, ", ")
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("options_agent")
	agent := NewLuaAgent("options-test", agentTable, L)

	// Test with various options
	opts := &agents.ExecutionOptions{
		MaxTokens:   100,
		Temperature: 0.7,
		Stream:      false,
		Timeout:     5 * time.Second,
	}

	result, err := agent.Execute(context.Background(), "test", opts)
	require.NoError(t, err)
	assert.Contains(t, result.Response, "Input: test")
	assert.Contains(t, result.Response, "MaxTokens: 100")
	assert.Contains(t, result.Response, "Temperature: 0.7")
	assert.Contains(t, result.Response, "Stream: false")
}

func TestLuaAgentWithActualTools(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Note: In a real scenario, tools would be registered globally
	// For this test, we're just verifying the agent can list tools

	// Create an agent that can use tools
	err := L.DoString(`
		tool_agent = {
			_tools = {"echo"},
			
			get_tools = function(self)
				return self._tools
			end,
			
			execute = function(self, input, options)
				-- In a real implementation, this would call the tool
				-- For testing, we just acknowledge the tools
				return "Agent with tools: " .. table.concat(self._tools, ", ")
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("tool_agent")
	agent := NewLuaAgent("tool-agent", agentTable, L)

	// Verify tool is listed
	assert.Equal(t, []string{"echo"}, agent.GetTools())

	// Execute
	result, err := agent.Execute(context.Background(), "test", nil)
	require.NoError(t, err)
	assert.Equal(t, "Agent with tools: echo", result.Response)
}

func TestLuaAgentStreamCallbackError(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a streaming agent
	err := L.DoString(`
		stream_agent = {
			stream = function(self, input, options, callback)
				local err = callback("First chunk")
				if err then return err end
				
				err = callback("Second chunk")
				if err then return err end
				
				err = callback("Third chunk")
				if err then return err end
				
				return nil
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("stream_agent")
	agent := NewLuaAgent("stream-test", agentTable, L)

	// Test callback that returns error on second chunk
	chunkCount := 0
	err = agent.Stream(context.Background(), "test", nil, func(chunk string) error {
		chunkCount++
		if chunkCount == 2 {
			return errors.New("stop streaming")
		}
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stop streaming")
	assert.Equal(t, 2, chunkCount)
}

func TestLuaAgentInvalidTypes(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Test with nil
	agent := NewLuaAgent("test", lua.LNil, L)
	_, err := agent.Execute(context.Background(), "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	// Test with number
	agent = NewLuaAgent("test", lua.LNumber(42), L)
	_, err = agent.Execute(context.Background(), "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")

	// Test with string
	agent = NewLuaAgent("test", lua.LString("not a function"), L)
	_, err = agent.Execute(context.Background(), "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestLuaAgentComplexReturnValues(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create an agent that returns complex values
	err := L.DoString(`
		complex_agent = {
			execute = function(self, input, options)
				-- Return a table (should be converted to JSON)
				return {
					message = "Processed: " .. input,
					metadata = {
						timestamp = os.time(),
						length = string.len(input)
					},
					tags = {"test", "complex"}
				}
			end
		}
	`)
	require.NoError(t, err)

	agentTable := L.GetGlobal("complex_agent")
	agent := NewLuaAgent("complex", agentTable, L)

	result, err := agent.Execute(context.Background(), "Hello", nil)
	require.NoError(t, err)

	// The result should be a JSON string
	assert.Contains(t, result.Response, "Processed: Hello")
	assert.Contains(t, result.Response, "metadata")
	assert.Contains(t, result.Response, "tags")
	assert.True(t, strings.HasPrefix(result.Response, "{") && strings.HasSuffix(result.Response, "}"))
}
