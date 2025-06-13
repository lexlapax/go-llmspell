// ABOUTME: Agent bridge provides access to go-llms agent functionality for script engines
// ABOUTME: Wraps agent creation, configuration, tool registration, and execution without reimplementation

package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for agent functionality
	_ "github.com/lexlapax/go-llms/pkg/agent/core"   // TODO: Will be used for agent creation
	_ "github.com/lexlapax/go-llms/pkg/agent/domain" // TODO: Will be used for agent types
)

// AgentBridge provides script access to go-llms agent functionality
type AgentBridge struct {
	mu          sync.RWMutex
	initialized bool
	agents      map[string]bridge.BaseAgent
	registry    bridge.AgentRegistry //nolint:unused // will be used when implementing registry methods
}

// NewAgentBridge creates a new agent bridge
func NewAgentBridge() *AgentBridge {
	return &AgentBridge{
		agents: make(map[string]bridge.BaseAgent),
	}
}

// GetID returns the bridge identifier
func (b *AgentBridge) GetID() string {
	return "agent"
}

// GetMetadata returns bridge metadata
func (b *AgentBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "agent",
		Version:     "1.0.0",
		Description: "Agent system bridge wrapping go-llms agent functionality",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *AgentBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *AgentBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clean up any registered agents
	for id, agent := range b.agents {
		if err := agent.Cleanup(ctx); err != nil {
			// Log error but continue cleanup
			_ = err
		}
		delete(b.agents, id)
	}

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *AgentBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *AgentBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *AgentBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "createAgent",
			Description: "Create a new agent with configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Agent ID", Required: true},
				{Name: "config", Type: "object", Description: "Agent configuration", Required: true},
			},
			ReturnType: "Agent",
		},
		{
			Name:        "createLLMAgent",
			Description: "Create a new LLM-powered agent",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Agent name", Required: true},
				{Name: "provider", Type: "Provider", Description: "LLM provider", Required: true},
				{Name: "options", Type: "object", Description: "Additional options", Required: false},
			},
			ReturnType: "Agent",
		},
		{
			Name:        "registerTool",
			Description: "Register a tool with an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "tool", Type: "Tool", Description: "Tool to register", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "runAgent",
			Description: "Run an agent with input",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "input", Type: "any", Description: "Input for the agent", Required: true},
				{Name: "options", Type: "object", Description: "Run options", Required: false},
			},
			ReturnType: "any",
		},
		{
			Name:        "runAgentAsync",
			Description: "Run an agent asynchronously",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "input", Type: "any", Description: "Input for the agent", Required: true},
				{Name: "options", Type: "object", Description: "Run options", Required: false},
			},
			ReturnType: "channel",
		},
		{
			Name:        "addSubAgent",
			Description: "Add a sub-agent to an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "parentID", Type: "string", Description: "Parent agent ID", Required: true},
				{Name: "subAgentID", Type: "string", Description: "Sub-agent ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getAgentState",
			Description: "Get the current state of an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "State",
		},
		{
			Name:        "setAgentState",
			Description: "Set the state of an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "state", Type: "State", Description: "New state", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "listAgents",
			Description: "List all registered agents",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "getAgent",
			Description: "Get an agent by ID",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "Agent",
		},
		{
			Name:        "removeAgent",
			Description: "Remove an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "setAgentHook",
			Description: "Set a lifecycle hook for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "hookType", Type: "string", Description: "Hook type (beforeRun, afterRun, etc.)", Required: true},
				{Name: "handler", Type: "function", Description: "Hook handler function", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "emitAgentEvent",
			Description: "Emit a custom agent event",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "eventType", Type: "string", Description: "Event type", Required: true},
				{Name: "data", Type: "object", Description: "Event data", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "subscribeToEvents",
			Description: "Subscribe to agent events",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Event filter", Required: false},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string", // subscription ID
		},
		{
			Name:        "unsubscribeFromEvents",
			Description: "Unsubscribe from agent events",
			Parameters: []engine.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getAgentMetrics",
			Description: "Get metrics for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "createWorkflow",
			Description: "Create a workflow agent",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Description: "Workflow type (sequential, parallel, conditional, loop)", Required: true},
				{Name: "config", Type: "object", Description: "Workflow configuration", Required: true},
			},
			ReturnType: "Agent",
		},
		{
			Name:        "addWorkflowStep",
			Description: "Add a step to a workflow",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow agent ID", Required: true},
				{Name: "step", Type: "object", Description: "Step configuration", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getAgentTools",
			Description: "Get tools registered with an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "configureAgent",
			Description: "Update agent configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "config", Type: "object", Description: "New configuration", Required: true},
			},
			ReturnType: "void",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *AgentBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Agent": {
			GoType:     "BaseAgent",
			ScriptType: "object",
		},
		"Tool": {
			GoType:     "Tool",
			ScriptType: "object",
		},
		"State": {
			GoType:     "*State",
			ScriptType: "object",
		},
		"AgentConfig": {
			GoType:     "AgentConfig",
			ScriptType: "object",
		},
		"LLMConfig": {
			GoType:     "LLMConfig",
			ScriptType: "object",
		},
		"AgentType": {
			GoType:     "AgentType",
			ScriptType: "string",
		},
		"AgentEvent": {
			GoType:     "Event",
			ScriptType: "object",
		},
		"Message": {
			GoType:     "Message",
			ScriptType: "object",
		},
		"Artifact": {
			GoType:     "*Artifact",
			ScriptType: "object",
		},
		"Provider": {
			GoType:     "Provider",
			ScriptType: "object",
		},
		"Hook": {
			GoType:     "Hook",
			ScriptType: "function",
		},
		"ToolContext": {
			GoType:     "ToolContext",
			ScriptType: "object",
		},
		"AgentRegistry": {
			GoType:     "*AgentRegistry",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *AgentBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions
func (b *AgentBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "agent",
			Actions:     []string{"create", "execute", "manage"},
			Description: "Access to agent system",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "state",
			Actions:     []string{"allocate", "manage"},
			Description: "Memory for agent state and execution",
		},
	}
}

// Helper methods for type conversion and agent management would go here

// getAgent retrieves an agent by ID
//
//nolint:unused // will be used when implementing agent methods
func (b *AgentBridge) getAgent(id string) (bridge.BaseAgent, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	agent, exists := b.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", id)
	}
	return agent, nil
}

// registerAgent registers an agent in the bridge
//
//nolint:unused // will be used when implementing agent creation methods
func (b *AgentBridge) registerAgent(agent bridge.BaseAgent) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.agents[agent.ID()]; exists {
		return fmt.Errorf("agent %s already registered", agent.ID())
	}

	b.agents[agent.ID()] = agent
	return nil
}

// removeAgentInternal removes an agent from the bridge
//
//nolint:unused // will be used when implementing removeAgent method
func (b *AgentBridge) removeAgentInternal(id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	agent, exists := b.agents[id]
	if !exists {
		return fmt.Errorf("agent %s not found", id)
	}

	// Cleanup the agent
	if err := agent.Cleanup(context.Background()); err != nil {
		return fmt.Errorf("failed to cleanup agent %s: %w", id, err)
	}

	delete(b.agents, id)
	return nil
}
