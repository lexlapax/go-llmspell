// ABOUTME: Tests for Lua agent registration functionality
// ABOUTME: Validates that Lua functions/tables can be registered as agents

package bridges

import (
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/agents"
	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

func TestAgentsRegister(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a mock agent bridge
	mockBridge := &mockAgentBridgeForRegister{}

	// Register the agents module
	err := RegisterAgentsModule(L, mockBridge)
	require.NoError(t, err)

	// Register a simple Lua function as an agent
	err = L.DoString(`
		function simple_agent(input, options)
			return "Echo from Lua: " .. input
		end

		success, err = agents.register("lua-simple", simple_agent)
	`)
	require.NoError(t, err)

	// Check registration result
	success := L.GetGlobal("success")
	assert.Equal(t, lua.LTBool, success.Type())
	assert.True(t, lua.LVAsBool(success))

	// Get the agent from registry
	registry := agents.DefaultRegistry()
	agent, err := registry.Get("lua-simple")
	require.NoError(t, err)
	assert.NotNil(t, agent)

	// Execute the agent
	result, err := agent.Execute(context.Background(), "Hello", nil)
	require.NoError(t, err)
	assert.Equal(t, "Echo from Lua: Hello", result.Response)

	// Clean up
	err = registry.Remove("lua-simple")
	require.NoError(t, err)
}

func TestAgentsRegisterTable(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a mock agent bridge
	mockBridge := &mockAgentBridgeForRegister{}

	// Register the agents module
	err := RegisterAgentsModule(L, mockBridge)
	require.NoError(t, err)

	// Register a Lua table as an agent
	err = L.DoString(`
		lua_table_agent = {
			name = "lua-table-agent",
			system_prompt = "I am a Lua-based agent",
			
			execute = function(self, input, options)
				return "Table agent says: " .. input
			end,
			
			get_system_prompt = function(self)
				return self.system_prompt
			end
		}

		success, err = agents.register("lua-table", lua_table_agent)
	`)
	require.NoError(t, err)

	// Check registration result
	success := L.GetGlobal("success")
	assert.True(t, lua.LVAsBool(success))

	// Get the agent from registry
	registry := agents.DefaultRegistry()
	agent, err := registry.Get("lua-table")
	require.NoError(t, err)
	assert.NotNil(t, agent)

	// Test system prompt
	assert.Equal(t, "I am a Lua-based agent", agent.GetSystemPrompt())

	// Execute the agent
	result, err := agent.Execute(context.Background(), "Hello from table", nil)
	require.NoError(t, err)
	assert.Equal(t, "Table agent says: Hello from table", result.Response)

	// Clean up
	err = registry.Remove("lua-table")
	require.NoError(t, err)
}

func TestAgentsRegisterWithStreaming(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a mock agent bridge
	mockBridge := &mockAgentBridgeForRegister{}

	// Register the agents module
	err := RegisterAgentsModule(L, mockBridge)
	require.NoError(t, err)

	// Register a streaming agent
	err = L.DoString(`
		streaming_agent = {
			execute = function(self, input, options)
				return "Non-streaming: " .. input
			end,
			
			stream = function(self, input, options, callback)
				-- Stream word by word
				for word in string.gmatch(input, "%S+") do
					local err = callback(word .. " ")
					if err then
						return err
					end
				end
				return nil
			end
		}

		success, err = agents.register("lua-streaming", streaming_agent)
	`)
	require.NoError(t, err)

	// Get the agent
	registry := agents.DefaultRegistry()
	agent, err := registry.Get("lua-streaming")
	require.NoError(t, err)

	// Test streaming
	var chunks []string
	err = agent.Stream(context.Background(), "Hello streaming world", nil, func(chunk string) error {
		chunks = append(chunks, chunk)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"Hello ", "streaming ", "world "}, chunks)

	// Clean up
	err = registry.Remove("lua-streaming")
	require.NoError(t, err)
}

func TestAgentsRegisterInvalidType(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a mock agent bridge
	mockBridge := &mockAgentBridgeForRegister{}

	// Register the agents module
	err := RegisterAgentsModule(L, mockBridge)
	require.NoError(t, err)

	// Try to register invalid type (number)
	err = L.DoString(`
		success, err = agents.register("invalid", 42)
	`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent implementation must be a function or table")
}

func TestAgentsRegisterDuplicate(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Create a mock agent bridge
	mockBridge := &mockAgentBridgeForRegister{}

	// Register the agents module
	err := RegisterAgentsModule(L, mockBridge)
	require.NoError(t, err)

	// Register an agent
	err = L.DoString(`
		function agent1(input)
			return "Agent 1"
		end
		
		success1, err1 = agents.register("duplicate-test", agent1)
	`)
	require.NoError(t, err)

	success1 := L.GetGlobal("success1")
	assert.True(t, lua.LVAsBool(success1))

	// Try to register with same name
	err = L.DoString(`
		function agent2(input)
			return "Agent 2"
		end
		
		success2, err2 = agents.register("duplicate-test", agent2)
	`)
	require.NoError(t, err)

	success2 := L.GetGlobal("success2")
	assert.False(t, lua.LVAsBool(success2))

	err2 := L.GetGlobal("err2")
	assert.Contains(t, lua.LVAsString(err2), "already registered")

	// Clean up
	registry := agents.DefaultRegistry()
	err = registry.Remove("duplicate-test")
	require.NoError(t, err)
}

// mockAgentBridgeForRegister for testing registration
type mockAgentBridgeForRegister struct {
	bridge.AgentBridge
	agents map[string]*mockAgentDataForRegister
}

type mockAgentDataForRegister struct {
	name         string
	systemPrompt string
}

func (m *mockAgentBridgeForRegister) Create(config map[string]interface{}) (string, error) {
	if m.agents == nil {
		m.agents = make(map[string]*mockAgentDataForRegister)
	}

	name, _ := config["name"].(string)
	if name == "" {
		return "", fmt.Errorf("agent name is required")
	}

	m.agents[name] = &mockAgentDataForRegister{
		name:         name,
		systemPrompt: config["systemPrompt"].(string),
	}

	return name, nil
}

func (m *mockAgentBridgeForRegister) List() []map[string]interface{} {
	var result []map[string]interface{}
	for _, agent := range m.agents {
		result = append(result, map[string]interface{}{
			"name":         agent.name,
			"systemPrompt": agent.systemPrompt,
		})
	}
	return result
}
