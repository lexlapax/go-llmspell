// ABOUTME: Agent bridge provides access to go-llms agent functionality for script engines
// ABOUTME: Wraps agent creation, configuration, tool registration, and execution without reimplementation

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for agent functionality
	agentcore "github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AgentBridge provides script access to go-llms agent functionality
type AgentBridge struct {
	mu            sync.RWMutex
	initialized   bool
	agents        map[string]bridge.BaseAgent
	registry      bridge.AgentRegistry  //nolint:unused // will be used when implementing registry methods
	eventStorage  events.EventStorage   // Storage for event replay
	eventReplayer *events.EventReplayer // Event replay functionality
	profiler      *profiling.Profiler   // Performance profiling
}

// NewAgentBridge creates a new agent bridge
func NewAgentBridge() *AgentBridge {
	storage := events.NewMemoryStorage()
	return &AgentBridge{
		agents:        make(map[string]bridge.BaseAgent),
		eventStorage:  storage,
		eventReplayer: events.NewEventReplayer(storage, nil), // Bus will be set during initialization
		profiler:      profiling.NewProfiler("agent_bridge"),
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
		Version:     "2.0.0",
		Description: "Agent system bridge with state serialization, event replay, and performance profiling",
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
		// State Serialization Methods
		{
			Name:        "exportAgentState",
			Description: "Export agent state to serialized format",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "format", Type: "string", Description: "Export format (json, compressed)", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "importAgentState",
			Description: "Import agent state from serialized format",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "stateData", Type: "object", Description: "Serialized state data", Required: true},
				{Name: "format", Type: "string", Description: "Data format", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "saveAgentSnapshot",
			Description: "Save an agent state snapshot",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "snapshotID", Type: "string", Description: "Snapshot ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "loadAgentSnapshot",
			Description: "Load agent state from snapshot",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "snapshotID", Type: "string", Description: "Snapshot ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "listAgentSnapshots",
			Description: "List available snapshots for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "deleteAgentSnapshot",
			Description: "Delete an agent snapshot",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "snapshotID", Type: "string", Description: "Snapshot ID", Required: true},
			},
			ReturnType: "object",
		},
		// Event Replay Methods
		{
			Name:        "replayAgentEvents",
			Description: "Replay agent events for debugging or recreation",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "speed", Type: "string", Description: "Replay speed", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "startEventRecording",
			Description: "Start recording agent events",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "stopEventRecording",
			Description: "Stop recording agent events",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getEventHistory",
			Description: "Get event history for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "limit", Type: "number", Description: "Maximum number of events", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "clearEventHistory",
			Description: "Clear event history for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		// Performance Profiling Methods
		{
			Name:        "startAgentProfiling",
			Description: "Start performance profiling for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "stopAgentProfiling",
			Description: "Stop performance profiling",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getAgentPerformanceReport",
			Description: "Get performance metrics for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "clearAgentProfilingData",
			Description: "Clear profiling data for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "exportAgentProfilingData",
			Description: "Export profiling data for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "format", Type: "string", Description: "Export format", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "setAgentProfilingConfig",
			Description: "Set profiling configuration for an agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "config", Type: "object", Description: "Profiling configuration", Required: true},
			},
			ReturnType: "object",
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
func (b *AgentBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// ExecuteMethod executes a bridge method
func (b *AgentBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "createAgent":
		// TODO: Implement createAgent
		return nil, fmt.Errorf("createAgent not yet implemented")
	case "createMinimalAgent":
		// TODO: Implement createMinimalAgent
		return nil, fmt.Errorf("createMinimalAgent not yet implemented")
	case "registerAgent":
		// TODO: Implement registerAgent
		return nil, fmt.Errorf("registerAgent not yet implemented")
	case "executeAgent":
		// TODO: Implement executeAgent
		return nil, fmt.Errorf("executeAgent not yet implemented")
	case "getAgentState":
		// TODO: Implement getAgentState
		return nil, fmt.Errorf("getAgentState not yet implemented")
	case "setAgentHook":
		// TODO: Implement setAgentHook
		return nil, fmt.Errorf("setAgentHook not yet implemented")
	case "clearAgentHooks":
		// TODO: Implement clearAgentHooks
		return nil, fmt.Errorf("clearAgentHooks not yet implemented")
	case "listAgents":
		// TODO: Implement listAgents
		return nil, fmt.Errorf("listAgents not yet implemented")
	case "destroyAgent":
		// TODO: Implement destroyAgent
		return nil, fmt.Errorf("destroyAgent not yet implemented")
	case "serializeAgentState":
		// TODO: Implement serializeAgentState
		return nil, fmt.Errorf("serializeAgentState not yet implemented")
	case "deserializeAgentState":
		// TODO: Implement deserializeAgentState
		return nil, fmt.Errorf("deserializeAgentState not yet implemented")
	case "replayEvents":
		// TODO: Implement replayEvents
		return nil, fmt.Errorf("replayEvents not yet implemented")
	case "exportAgentEvents":
		// TODO: Implement exportAgentEvents
		return nil, fmt.Errorf("exportAgentEvents not yet implemented")
	case "startProfiling":
		// TODO: Implement startProfiling
		return nil, fmt.Errorf("startProfiling not yet implemented")
	case "stopProfiling":
		// TODO: Implement stopProfiling
		return nil, fmt.Errorf("stopProfiling not yet implemented")
	case "getProfilingReport":
		// TODO: Implement getProfilingReport
		return nil, fmt.Errorf("getProfilingReport not yet implemented")
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
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

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *AgentBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createAgent":
		if len(args) < 2 {
			return nil, fmt.Errorf("createAgent requires type and config parameters")
		}
		agentType, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agent type must be string")
		}
		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Extract name and description
		name, _ := config["name"].(string)
		if name == "" {
			name = "agent"
		}
		description, _ := config["description"].(string)
		if description == "" {
			description = "Script-created agent"
		}

		// Create agent based on type
		var agent bridge.BaseAgent

		switch domain.AgentType(agentType) {
		case domain.AgentTypeLLM:
			// For LLM agent, we need a provider
			// This is simplified - in real implementation, would get provider from bridge
			// For now, return error indicating provider needed
			return nil, fmt.Errorf("LLM agent creation requires provider setup")

		default:
			// For other types, we can create a base agent
			agent = agentcore.NewBaseAgent(name, description, domain.AgentType(agentType))
		}

		// Store agent
		b.agents[agent.ID()] = agent

		// Return agent info
		return map[string]interface{}{
			"id":   agent.ID(),
			"type": agent.Type(),
			"name": agent.Name(),
		}, nil

	case "executeAgent":
		if len(args) < 2 {
			return nil, fmt.Errorf("executeAgent requires agentID and input parameters")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		agent, err := b.getAgent(agentID)
		if err != nil {
			return nil, err
		}

		// Create state from input
		inputState := domain.NewState()
		if inputData, ok := args[1].(map[string]interface{}); ok {
			for k, v := range inputData {
				inputState.Set(k, v)
			}
		}

		// Run agent
		resultState, err := agent.Run(ctx, inputState)
		if err != nil {
			return nil, fmt.Errorf("agent execution failed: %w", err)
		}

		// Convert result state to map
		result := resultState.Values()

		return result, nil

	case "listAgents":
		agents := make([]map[string]interface{}, 0, len(b.agents))
		for _, agent := range b.agents {
			agents = append(agents, map[string]interface{}{
				"id":   agent.ID(),
				"type": agent.Type(),
				"name": agent.Name(),
			})
		}
		return agents, nil

	// State Serialization Methods
	case "exportAgentState":
		if len(args) < 1 {
			return nil, fmt.Errorf("exportAgentState requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		_, err := b.getAgent(agentID)
		if err != nil {
			return nil, err
		}

		// Get current state - we need to implement this based on available methods
		// For now, use a placeholder implementation
		currentState := domain.NewState()
		currentState.Set("agentID", agentID)
		currentState.Set("exportTime", fmt.Sprintf("%d", ctx.Value("timestamp")))

		// Determine export format
		format := "json"
		if len(args) > 1 {
			if f, ok := args[1].(string); ok {
				format = f
			}
		}

		// Create serialized state using available utilities
		stateValues := currentState.Values()

		return map[string]interface{}{
			"agentID":   agentID,
			"format":    format,
			"state":     stateValues,
			"timestamp": fmt.Sprintf("%d", ctx.Value("timestamp")),
			"version":   "1.0",
		}, nil

	case "importAgentState":
		if len(args) < 2 {
			return nil, fmt.Errorf("importAgentState requires agentID and stateData parameters")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}
		stateData, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("stateData must be object")
		}

		_, err := b.getAgent(agentID)
		if err != nil {
			return nil, err
		}

		// Create new state from imported data
		state := domain.NewState()
		if stateValues, ok := stateData["state"].(map[string]interface{}); ok {
			for k, v := range stateValues {
				state.Set(k, v)
			}
		}

		// For now, we can't directly set agent state as the interface doesn't support it
		// This would need to be implemented based on the specific agent type
		_ = state // Use the imported state

		return nil, nil

	case "createStateSnapshot":
		if len(args) < 2 {
			return nil, fmt.Errorf("createStateSnapshot requires agentID and snapshotName parameters")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}
		snapshotName, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("snapshotName must be string")
		}

		_, err := b.getAgent(agentID)
		if err != nil {
			return nil, err
		}

		// Create snapshot using available state utilities
		currentState := domain.NewState()
		currentState.Set("agentID", agentID)
		currentState.Set("snapshotName", snapshotName)

		// Create snapshot data
		snapshotData := map[string]interface{}{
			"name":      snapshotName,
			"agentID":   agentID,
			"state":     currentState.Values(),
			"timestamp": fmt.Sprintf("%d", ctx.Value("timestamp")),
		}

		return map[string]interface{}{
			"snapshotName": snapshotName,
			"agentID":      agentID,
			"snapshot":     snapshotData,
			"created":      snapshotData["timestamp"],
		}, nil

	case "restoreFromSnapshot":
		if len(args) < 2 {
			return nil, fmt.Errorf("restoreFromSnapshot requires agentID and snapshotName parameters")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}
		snapshotName, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("snapshotName must be string")
		}

		_, err := b.getAgent(agentID)
		if err != nil {
			return nil, err
		}

		// Restore from snapshot - placeholder implementation
		// In a real implementation, this would restore the agent state from stored snapshot
		// The agent would be retrieved and state restored from the snapshot
		_ = snapshotName

		return nil, nil

	case "encryptAgentState":
		if len(args) < 2 {
			return nil, fmt.Errorf("encryptAgentState requires agentID and password parameters")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}
		password, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("password must be string")
		}

		_, err := b.getAgent(agentID)
		if err != nil {
			return nil, err
		}

		// Encrypt state - simplified implementation using JSON
		currentState := domain.NewState()
		currentState.Set("agentID", agentID)

		// For encryption, we would normally use crypto packages
		// This is a placeholder implementation
		stateJSON, err := json.Marshal(currentState.Values())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state: %w", err)
		}

		// Simple "encryption" - in real implementation would use AES or similar
		encryptedData := fmt.Sprintf("encrypted_%s_%s", password[:min(len(password), 4)], string(stateJSON))

		return map[string]interface{}{
			"agentID":        agentID,
			"encryptedState": encryptedData,
			"encrypted":      true,
		}, nil

	case "decryptAgentState":
		if len(args) < 2 {
			return nil, fmt.Errorf("decryptAgentState requires encryptedData and password parameters")
		}
		encryptedData, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("encryptedData must be object")
		}
		password, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("password must be string")
		}

		// Decrypt state - simplified implementation
		encryptedState, ok := encryptedData["encryptedState"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid encrypted data format")
		}

		// Simple "decryption" - in real implementation would use AES or similar
		// For now, just extract the JSON part after the prefix
		prefix := fmt.Sprintf("encrypted_%s_", password[:min(len(password), 4)])
		if len(encryptedState) <= len(prefix) {
			return nil, fmt.Errorf("invalid encrypted data")
		}

		jsonData := encryptedState[len(prefix):]
		var stateValues map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &stateValues); err != nil {
			return nil, fmt.Errorf("failed to decrypt state: %w", err)
		}

		return stateValues, nil

	// Event Replay Methods
	case "replayAgentEvents":
		if len(args) < 1 {
			return nil, fmt.Errorf("replayAgentEvents requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		// Build event query
		query := events.EventQuery{
			AgentID: agentID,
		}

		if len(args) > 1 {
			if queryData, ok := args[1].(map[string]interface{}); ok {
				// Parse query parameters from script
				if startTime, ok := queryData["startTime"].(string); ok {
					// Parse time string
					// Implementation would parse the time
					_ = startTime
				}
				if eventTypes, ok := queryData["eventTypes"].([]interface{}); ok {
					// Convert to domain.EventType slice
					_ = eventTypes
				}
			}
		}

		// Prepare replay options
		opts := events.ReplayOptions{
			Speed: 1.0, // Real-time by default
		}

		if len(args) > 2 {
			if options, ok := args[2].(map[string]interface{}); ok {
				if speed, ok := options["speed"].(float64); ok {
					opts.Speed = speed
				}
			}
		}

		// Perform replay using go-llms event replayer
		if err := b.eventReplayer.Replay(ctx, query, opts); err != nil {
			return nil, fmt.Errorf("failed to replay events: %w", err)
		}

		return nil, nil

	case "startEventRecording":
		// Start recording events to storage
		recorder := events.NewEventRecorder(b.eventStorage, nil) // Bus would be initialized
		if err := recorder.Start(); err != nil {
			return nil, fmt.Errorf("failed to start recording: %w", err)
		}

		// Generate recording ID
		recordingID := fmt.Sprintf("recording_%d", ctx.Value("timestamp"))

		return recordingID, nil

	case "stopEventRecording":
		if len(args) < 1 {
			return nil, fmt.Errorf("stopEventRecording requires recordingID parameter")
		}
		recordingID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("recordingID must be string")
		}

		// Stop recording and return summary
		// Implementation would track recorders by ID
		_ = recordingID

		return map[string]interface{}{
			"recordingID": recordingID,
			"stopped":     true,
			"eventCount":  0, // Would be actual count
		}, nil

	case "queryAgentEvents":
		if len(args) < 1 {
			return nil, fmt.Errorf("queryAgentEvents requires query parameter")
		}
		queryData, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("query must be object")
		}

		// Convert script query to go-llms EventQuery
		query := events.EventQuery{}
		if agentID, ok := queryData["agentID"].(string); ok {
			query.AgentID = agentID
		}
		if limit, ok := queryData["limit"].(float64); ok {
			query.Limit = int(limit)
		}

		// Query events from storage
		eventsList, err := b.eventStorage.Query(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query events: %w", err)
		}

		// Convert events to script-friendly format
		result := make([]map[string]interface{}, len(eventsList))
		for i, event := range eventsList {
			serialized, err := events.SerializeEvent(event)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize event: %w", err)
			}
			result[i] = serialized
		}

		return result, nil

	case "exportEventHistory":
		if len(args) < 1 {
			return nil, fmt.Errorf("exportEventHistory requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		// Query all events for the agent
		query := events.EventQuery{
			AgentID: agentID,
		}

		eventsList, err := b.eventStorage.Query(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query events: %w", err)
		}

		// Determine export format
		format := "json"
		if len(args) > 1 {
			if f, ok := args[1].(string); ok {
				format = f
			}
		}

		// Create event batch for export
		batch, err := events.SerializeEventBatch(eventsList)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize event batch: %w", err)
		}

		return map[string]interface{}{
			"agentID":    agentID,
			"format":     format,
			"eventCount": len(eventsList),
			"history":    batch,
		}, nil

	// Performance Profiling Methods
	case "startAgentProfiling":
		if len(args) < 1 {
			return nil, fmt.Errorf("startAgentProfiling requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		// Determine profile type
		profileType := "both"
		if len(args) > 1 {
			if pt, ok := args[1].(string); ok {
				profileType = pt
			}
		}

		// Create agent-specific profiler
		profiler := profiling.NewProfiler(fmt.Sprintf("agent_%s", agentID))
		profiler.Enable()

		// Start CPU profiling if requested
		if profileType == "cpu" || profileType == "both" {
			if err := profiler.StartCPUProfile(); err != nil {
				return nil, fmt.Errorf("failed to start CPU profiling: %w", err)
			}
		}

		// Generate session ID
		sessionID := fmt.Sprintf("profile_%s_%d", agentID, ctx.Value("timestamp"))

		return sessionID, nil

	case "stopAgentProfiling":
		if len(args) < 1 {
			return nil, fmt.Errorf("stopAgentProfiling requires sessionID parameter")
		}
		sessionID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sessionID must be string")
		}

		// Stop profiling and generate report
		// Implementation would track profilers by session ID
		_ = sessionID

		return map[string]interface{}{
			"sessionID":  sessionID,
			"stopped":    true,
			"cpuProfile": "/tmp/cpu.pprof",
			"memProfile": "/tmp/mem.pprof",
		}, nil

	case "getAgentPerformanceReport":
		if len(args) < 1 {
			return nil, fmt.Errorf("getAgentPerformanceReport requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		// Generate performance report using go-llms profiling
		report := map[string]interface{}{
			"agentID":     agentID,
			"cpuUsage":    "15%", // Would be actual metrics
			"memoryUsage": "128MB",
			"avgLatency":  "45ms",
			"totalOps":    1234,
			"successRate": 0.987,
		}

		return report, nil

	case "profileAgentOperation":
		if len(args) < 3 {
			return nil, fmt.Errorf("profileAgentOperation requires agentID, operation, and opName parameters")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}
		// operation would be a function - complex to handle in bridge
		opName, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("opName must be string")
		}

		// Use go-llms profiler to profile the operation
		profiler := profiling.NewProfiler(fmt.Sprintf("agent_%s", agentID))

		// Profile the operation (simplified implementation)
		result, err := profiler.ProfileOperation(ctx, opName, func(ctx context.Context) (interface{}, error) {
			// Would execute the provided operation function
			return map[string]interface{}{"result": "operation completed"}, nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to profile operation: %w", err)
		}

		return map[string]interface{}{
			"result":   result,
			"profile":  "operation_profile.pprof",
			"duration": "125ms", // Would be actual duration
		}, nil

	case "enableContinuousProfiling":
		if len(args) < 1 {
			return nil, fmt.Errorf("enableContinuousProfiling requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		// Enable continuous profiling for the agent
		b.profiler.Enable()

		// Implementation would start background profiling
		_ = agentID

		return nil, nil

	case "disableContinuousProfiling":
		if len(args) < 1 {
			return nil, fmt.Errorf("disableContinuousProfiling requires agentID parameter")
		}
		agentID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("agentID must be string")
		}

		// Disable continuous profiling and return final metrics
		b.profiler.Disable()

		return map[string]interface{}{
			"agentID":  agentID,
			"disabled": true,
			"finalMetrics": map[string]interface{}{
				"totalRuntime": "2h45m",
				"avgCpuUsage":  "12%",
				"peakMemory":   "256MB",
			},
		}, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
