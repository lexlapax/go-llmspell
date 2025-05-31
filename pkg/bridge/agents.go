// ABOUTME: Provides a bridge between the agent system and scripting environments
// ABOUTME: Exposes agent creation, execution, and management to scripts

package bridge

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/agents"
)

// AgentBridge provides script access to the agent system
type AgentBridge interface {
	// Create creates a new agent with the given configuration
	Create(config map[string]interface{}) (string, error)

	// Execute runs an agent with a single input
	Execute(agentName, input string, options map[string]interface{}) (string, error)

	// Stream executes an agent with streaming response
	Stream(agentName, input string, options map[string]interface{}, callback func(string) error) error

	// List returns information about all agents
	List() []map[string]interface{}

	// GetInfo returns information about a specific agent
	GetInfo(agentName string) (map[string]interface{}, error)

	// Remove removes an agent
	Remove(agentName string) error

	// UpdateSystemPrompt updates an agent's system prompt
	UpdateSystemPrompt(agentName, prompt string) error

	// AddTool adds a tool to an agent
	AddTool(agentName, toolName string) error
}

// agentBridge is the default implementation of AgentBridge
type agentBridge struct {
	ctx      context.Context
	registry agents.Registry
}

// NewAgentBridge creates a new agent bridge
func NewAgentBridge(ctx context.Context) (AgentBridge, error) {
	return &agentBridge{
		ctx:      ctx,
		registry: agents.DefaultRegistry(),
	}, nil
}

// Create creates a new agent with the given configuration
func (b *agentBridge) Create(config map[string]interface{}) (string, error) {
	// Convert map to agents.Config
	agentConfig := agents.Config{}

	// Required fields
	if name, ok := config["name"].(string); ok {
		agentConfig.Name = name
	} else {
		return "", fmt.Errorf("agent name is required")
	}

	if provider, ok := config["provider"].(string); ok {
		agentConfig.Provider = provider
	} else {
		return "", fmt.Errorf("provider is required")
	}

	if model, ok := config["model"].(string); ok {
		agentConfig.Model = model
	} else {
		return "", fmt.Errorf("model is required")
	}

	// Optional fields
	if systemPrompt, ok := config["systemPrompt"].(string); ok {
		agentConfig.SystemPrompt = systemPrompt
	}

	if maxTokens, ok := config["maxTokens"].(float64); ok {
		agentConfig.MaxTokens = int(maxTokens)
	} else if maxTokens, ok := config["maxTokens"].(int); ok {
		agentConfig.MaxTokens = maxTokens
	}

	if temperature, ok := config["temperature"].(float64); ok {
		agentConfig.Temperature = temperature
	}

	if timeout, ok := config["timeout"].(float64); ok {
		agentConfig.Timeout = time.Duration(timeout) * time.Second
	} else if timeout, ok := config["timeout"].(int); ok {
		agentConfig.Timeout = time.Duration(timeout) * time.Second
	}

	// Handle tools array
	if tools, ok := config["tools"].([]interface{}); ok {
		agentConfig.Tools = make([]string, 0, len(tools))
		for _, t := range tools {
			if toolName, ok := t.(string); ok {
				agentConfig.Tools = append(agentConfig.Tools, toolName)
			}
		}
	} else if tools, ok := config["tools"].([]string); ok {
		agentConfig.Tools = tools
	}

	// Create the agent
	agent, err := b.registry.Create(agentConfig)
	if err != nil {
		return "", err
	}

	return agent.Name(), nil
}

// Execute runs an agent with a single input
func (b *agentBridge) Execute(agentName, input string, options map[string]interface{}) (string, error) {
	agent, err := b.registry.Get(agentName)
	if err != nil {
		return "", err
	}

	// Convert options
	opts := b.convertExecutionOptions(options)

	// Execute
	result, err := agent.Execute(b.ctx, input, opts)
	if err != nil {
		return "", err
	}

	return result.Response, nil
}

// Stream executes an agent with streaming response
func (b *agentBridge) Stream(agentName, input string, options map[string]interface{}, callback func(string) error) error {
	agent, err := b.registry.Get(agentName)
	if err != nil {
		return err
	}

	// Convert options
	opts := b.convertExecutionOptions(options)
	if opts == nil {
		opts = &agents.ExecutionOptions{}
	}
	opts.Stream = true

	// Stream
	return agent.Stream(b.ctx, input, opts, agents.StreamCallback(callback))
}

// List returns information about all agents
func (b *agentBridge) List() []map[string]interface{} {
	// The registry doesn't have a way to list all created agents
	// For now, return empty list
	// TODO: Enhance registry to track all created agents
	return []map[string]interface{}{}
}

// GetInfo returns information about a specific agent
func (b *agentBridge) GetInfo(agentName string) (map[string]interface{}, error) {
	agent, err := b.registry.Get(agentName)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":         agent.Name(),
		"systemPrompt": agent.GetSystemPrompt(),
		"tools":        agent.GetTools(),
	}

	return info, nil
}

// Remove removes an agent
func (b *agentBridge) Remove(agentName string) error {
	return b.registry.Remove(agentName)
}

// UpdateSystemPrompt updates an agent's system prompt
func (b *agentBridge) UpdateSystemPrompt(agentName, prompt string) error {
	agent, err := b.registry.Get(agentName)
	if err != nil {
		return err
	}

	agent.SetSystemPrompt(prompt)
	return nil
}

// AddTool adds a tool to an agent
func (b *agentBridge) AddTool(agentName, toolName string) error {
	agent, err := b.registry.Get(agentName)
	if err != nil {
		return err
	}

	return agent.AddTool(toolName)
}

// convertExecutionOptions converts a map to ExecutionOptions
func (b *agentBridge) convertExecutionOptions(options map[string]interface{}) *agents.ExecutionOptions {
	if options == nil {
		return nil
	}

	opts := &agents.ExecutionOptions{}

	if stream, ok := options["stream"].(bool); ok {
		opts.Stream = stream
	}

	if maxTokens, ok := options["maxTokens"].(float64); ok {
		opts.MaxTokens = int(maxTokens)
	} else if maxTokens, ok := options["maxTokens"].(int); ok {
		opts.MaxTokens = maxTokens
	}

	if temperature, ok := options["temperature"].(float64); ok {
		opts.Temperature = temperature
	}

	if timeout, ok := options["timeout"].(float64); ok {
		opts.Timeout = time.Duration(timeout) * time.Second
	} else if timeout, ok := options["timeout"].(int); ok {
		opts.Timeout = time.Duration(timeout) * time.Second
	}

	return opts
}
