// ABOUTME: Tests for the Lua agent bridge implementation
// ABOUTME: Verifies agent creation, execution, and management through Lua

package bridges

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"
)

// mockAgentBridge implements bridge.AgentBridge for testing
type mockAgentBridge struct {
	agents        map[string]*mockAgent
	createCalled  bool
	createErr     error
	executeCalled bool
	executeErr    error
	streamCalled  bool
	streamErr     error
}

// mockAgent represents a mock agent
type mockAgent struct {
	name         string
	systemPrompt string
	tools        []string
}

func newMockAgentBridge() *mockAgentBridge {
	return &mockAgentBridge{
		agents: make(map[string]*mockAgent),
	}
}

func (m *mockAgentBridge) Create(config map[string]interface{}) (string, error) {
	m.createCalled = true
	if m.createErr != nil {
		return "", m.createErr
	}

	name, ok := config["name"].(string)
	if !ok {
		return "", errors.New("name is required")
	}

	agent := &mockAgent{
		name:         name,
		systemPrompt: "",
		tools:        []string{},
	}

	if sp, ok := config["systemPrompt"].(string); ok {
		agent.systemPrompt = sp
	}

	if tools, ok := config["tools"].([]string); ok {
		agent.tools = tools
	}

	m.agents[name] = agent
	return name, nil
}

func (m *mockAgentBridge) Execute(agentName, input string, options map[string]interface{}) (string, error) {
	m.executeCalled = true
	if m.executeErr != nil {
		return "", m.executeErr
	}

	if _, exists := m.agents[agentName]; !exists {
		return "", errors.New("agent not found")
	}

	return "Response to: " + input, nil
}

func (m *mockAgentBridge) Stream(agentName, input string, options map[string]interface{}, callback func(string) error) error {
	m.streamCalled = true
	if m.streamErr != nil {
		return m.streamErr
	}

	if _, exists := m.agents[agentName]; !exists {
		return errors.New("agent not found")
	}

	// Simulate streaming
	chunks := []string{"Chunk 1: ", "Processing ", input}
	for _, chunk := range chunks {
		if err := callback(chunk); err != nil {
			return err
		}
	}

	return nil
}

func (m *mockAgentBridge) List() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(m.agents))
	for name, agent := range m.agents {
		result = append(result, map[string]interface{}{
			"name":         name,
			"systemPrompt": agent.systemPrompt,
			"tools":        agent.tools,
		})
	}
	return result
}

func (m *mockAgentBridge) GetInfo(agentName string) (map[string]interface{}, error) {
	agent, exists := m.agents[agentName]
	if !exists {
		return nil, errors.New("agent not found")
	}

	return map[string]interface{}{
		"name":         agent.name,
		"systemPrompt": agent.systemPrompt,
		"tools":        agent.tools,
	}, nil
}

func (m *mockAgentBridge) Remove(agentName string) error {
	if _, exists := m.agents[agentName]; !exists {
		return errors.New("agent not found")
	}
	delete(m.agents, agentName)
	return nil
}

func (m *mockAgentBridge) UpdateSystemPrompt(agentName, prompt string) error {
	agent, exists := m.agents[agentName]
	if !exists {
		return errors.New("agent not found")
	}
	agent.systemPrompt = prompt
	return nil
}

func (m *mockAgentBridge) AddTool(agentName, toolName string) error {
	agent, exists := m.agents[agentName]
	if !exists {
		return errors.New("agent not found")
	}
	agent.tools = append(agent.tools, toolName)
	return nil
}

func TestRegisterAgentsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	err := RegisterAgentsModule(L, mockBridge)
	require.NoError(t, err)

	// Check if agents module is registered
	agents := L.GetGlobal("agents")
	require.NotEqual(t, lua.LNil, agents)
	require.Equal(t, lua.LTTable, agents.Type())

	// Check if all functions are registered
	agentsTable := agents.(*lua.LTable)
	functions := []string{
		"create", "execute", "stream", "list",
		"get", "remove", "update_system_prompt", "add_tool",
	}

	for _, fn := range functions {
		f := agentsTable.RawGetString(fn)
		assert.NotEqual(t, lua.LNil, f, "Function %s should be registered", fn)
		assert.Equal(t, lua.LTFunction, f.Type(), "agents.%s should be a function", fn)
	}
}

func TestAgentsCreate(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Test successful creation
	err := L.DoString(`
		local name, err = agents.create({
			name = "test-agent",
			provider = "openai",
			model = "gpt-4",
			systemPrompt = "You are a helpful assistant"
		})
		
		assert(name == "test-agent", "Agent name should match")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.createCalled)

	// Test creation with error
	mockBridge.createErr = errors.New("creation failed")
	err = L.DoString(`
		local name, err = agents.create({
			name = "fail-agent",
			provider = "openai",
			model = "gpt-4"
		})
		
		assert(name == nil, "Name should be nil on error")
		assert(err == "creation failed", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestAgentsExecute(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Create an agent first
	mockBridge.agents["test-agent"] = &mockAgent{name: "test-agent"}

	// Test successful execution
	err := L.DoString(`
		local response, err = agents.execute("test-agent", "Hello")
		
		assert(response == "Response to: Hello", "Response should match")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.executeCalled)

	// Test execution with options
	err = L.DoString(`
		local response, err = agents.execute("test-agent", "Hello", {
			temperature = 0.7,
			maxTokens = 100
		})
		
		assert(response == "Response to: Hello", "Response should match")
		assert(err == nil, "Error should be nil")
	`)
	require.NoError(t, err)

	// Test execution with non-existent agent
	err = L.DoString(`
		local response, err = agents.execute("non-existent", "Hello")
		
		assert(response == nil, "Response should be nil on error")
		assert(err == "agent not found", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestAgentsStream(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Create an agent first
	mockBridge.agents["test-agent"] = &mockAgent{name: "test-agent"}

	// Test successful streaming
	err := L.DoString(`
		local chunks = {}
		local success, err = agents.stream("test-agent", "Hello", function(chunk)
			table.insert(chunks, chunk)
			return true  -- Continue streaming
		end)
		
		assert(success == true, "Stream should succeed")
		assert(err == nil, "Error should be nil")
		assert(#chunks == 3, "Should receive 3 chunks")
		assert(chunks[1] == "Chunk 1: ", "First chunk should match")
		assert(chunks[2] == "Processing ", "Second chunk should match")
		assert(chunks[3] == "Hello", "Third chunk should match")
	`)
	require.NoError(t, err)
	assert.True(t, mockBridge.streamCalled)

	// Test streaming with callback that stops
	err = L.DoString(`
		local chunks = {}
		local success, err = agents.stream("test-agent", "Hello", function(chunk)
			table.insert(chunks, chunk)
			return false  -- Stop streaming
		end)
		
		assert(success == false, "Stream should fail when stopped")
		assert(err == "streaming stopped by callback", "Error message should match")
		assert(#chunks == 1, "Should receive only 1 chunk before stopping")
	`)
	require.NoError(t, err)
}

func TestAgentsList(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Add some agents
	mockBridge.agents["agent1"] = &mockAgent{
		name:         "agent1",
		systemPrompt: "System 1",
		tools:        []string{"tool1"},
	}
	mockBridge.agents["agent2"] = &mockAgent{
		name:         "agent2",
		systemPrompt: "System 2",
		tools:        []string{"tool2", "tool3"},
	}

	// Test listing agents
	err := L.DoString(`
		local agents_list = agents.list()
		
		assert(#agents_list >= 2, "Should have at least 2 agents")
		
		-- Check if our agents are in the list
		local found_agent1 = false
		local found_agent2 = false
		
		for _, agent in ipairs(agents_list) do
			if agent.name == "agent1" then
				found_agent1 = true
				assert(agent.systemPrompt == "System 1", "System prompt should match")
				assert(#agent.tools == 1, "Should have 1 tool")
			elseif agent.name == "agent2" then
				found_agent2 = true
				assert(agent.systemPrompt == "System 2", "System prompt should match")
				assert(#agent.tools == 2, "Should have 2 tools")
			end
		end
		
		assert(found_agent1, "Agent1 should be in the list")
		assert(found_agent2, "Agent2 should be in the list")
	`)
	require.NoError(t, err)
}

func TestAgentsGetInfo(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Add an agent
	mockBridge.agents["test-agent"] = &mockAgent{
		name:         "test-agent",
		systemPrompt: "Test system prompt",
		tools:        []string{"tool1", "tool2"},
	}

	// Test getting agent info
	err := L.DoString(`
		local info, err = agents.get("test-agent")
		
		assert(info ~= nil, "Info should not be nil")
		assert(err == nil, "Error should be nil")
		assert(info.name == "test-agent", "Name should match")
		assert(info.systemPrompt == "Test system prompt", "System prompt should match")
		assert(#info.tools == 2, "Should have 2 tools")
		assert(info.tools[1] == "tool1", "First tool should match")
		assert(info.tools[2] == "tool2", "Second tool should match")
	`)
	require.NoError(t, err)

	// Test getting non-existent agent info
	err = L.DoString(`
		local info, err = agents.get("non-existent")
		
		assert(info == nil, "Info should be nil for non-existent agent")
		assert(err == "agent not found", "Error message should match")
	`)
	require.NoError(t, err)
}

func TestAgentsRemove(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Add an agent
	mockBridge.agents["test-agent"] = &mockAgent{name: "test-agent"}

	// Test removing agent
	err := L.DoString(`
		local success, err = agents.remove("test-agent")
		
		assert(success == true, "Remove should succeed")
		assert(err == nil, "Error should be nil")
		
		-- Try to get the removed agent
		local info, err2 = agents.get("test-agent")
		assert(info == nil, "Removed agent should not exist")
		assert(err2 == "agent not found", "Should get not found error")
	`)
	require.NoError(t, err)

	// Verify agent was removed
	_, exists := mockBridge.agents["test-agent"]
	assert.False(t, exists)
}

func TestAgentsUpdateSystemPrompt(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Add an agent
	mockBridge.agents["test-agent"] = &mockAgent{
		name:         "test-agent",
		systemPrompt: "Original prompt",
	}

	// Test updating system prompt
	err := L.DoString(`
		local success, err = agents.update_system_prompt("test-agent", "New prompt")
		
		assert(success == true, "Update should succeed")
		assert(err == nil, "Error should be nil")
		
		-- Check the updated prompt
		local info, _ = agents.get("test-agent")
		assert(info.systemPrompt == "New prompt", "System prompt should be updated")
	`)
	require.NoError(t, err)

	// Verify prompt was updated
	assert.Equal(t, "New prompt", mockBridge.agents["test-agent"].systemPrompt)
}

func TestAgentsAddTool(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	mockBridge := newMockAgentBridge()
	require.NoError(t, RegisterAgentsModule(L, mockBridge))

	// Add an agent
	mockBridge.agents["test-agent"] = &mockAgent{
		name:  "test-agent",
		tools: []string{"existing-tool"},
	}

	// Test adding a tool
	err := L.DoString(`
		local success, err = agents.add_tool("test-agent", "new-tool")
		
		assert(success == true, "Add tool should succeed")
		assert(err == nil, "Error should be nil")
		
		-- Check the tools
		local info, _ = agents.get("test-agent")
		assert(#info.tools == 2, "Should have 2 tools now")
		assert(info.tools[1] == "existing-tool", "First tool should match")
		assert(info.tools[2] == "new-tool", "Second tool should match")
	`)
	require.NoError(t, err)

	// Verify tool was added
	assert.Contains(t, mockBridge.agents["test-agent"].tools, "new-tool")
}

func TestAgentsIntegration(t *testing.T) {
	// Skip if not integration test
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	L := lua.NewState()
	defer L.Close()

	// Create a real agent bridge
	ctx := context.Background()
	agentBridge, err := bridge.NewAgentBridge(ctx)
	require.NoError(t, err)

	// Register the module
	err = RegisterAgentsModule(L, agentBridge)
	require.NoError(t, err)

	// Test creating and using an agent with the default registry
	err = L.DoString(`
		-- Register a default agent factory first
		local agent_created = false
		
		-- Try to create an agent (this will fail without a registered factory)
		local name, err = agents.create({
			name = "test-lua-agent",
			provider = "mock",
			model = "test-model",
			systemPrompt = "You are a test assistant"
		})
		
		-- We expect this to fail since we don't have a real provider
		assert(name == nil or err ~= nil, "Should fail without real provider")
	`)
	require.NoError(t, err)
}
