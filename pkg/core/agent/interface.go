// ABOUTME: Engine-agnostic agent interface that defines lifecycle, metadata, and extension points for all agents
// ABOUTME: Provides a common contract for agents to work across Lua, JavaScript, and Tengo script engines

package agent

import (
	"context"
)

// Agent defines the core interface that all agents must implement.
// It provides lifecycle methods, metadata access, and is designed to be
// engine-agnostic, allowing agents to work across different script engines.
type Agent interface {
	// ID returns the unique identifier for this agent
	ID() string

	// Name returns the human-readable name of the agent
	Name() string

	// Description returns a brief description of what the agent does
	Description() string

	// Version returns the version of the agent
	Version() string

	// Capabilities returns a map of agent capabilities and their values
	// Examples: {"streaming": true, "tools": ["file_read"], "maxTokens": 4096}
	Capabilities() map[string]interface{}

	// Metadata returns additional metadata about the agent
	// Examples: author, license, tags, creation date, etc.
	Metadata() map[string]interface{}

	// Init initializes the agent with the given context
	// This is called once before the agent starts processing
	Init(ctx context.Context) error

	// Run executes the agent's main logic with the given input
	// The input and output types are interface{} to allow flexibility
	Run(ctx context.Context, input interface{}) (interface{}, error)

	// Cleanup performs any necessary cleanup when the agent is done
	// This is called once after the agent finishes processing
	Cleanup(ctx context.Context) error
}

// Status represents the current state of an agent
type Status string

const (
	// StatusCreated indicates the agent has been created but not initialized
	StatusCreated Status = "created"

	// StatusInitializing indicates the agent is currently initializing
	StatusInitializing Status = "initializing"

	// StatusReady indicates the agent has been initialized and is ready to run
	StatusReady Status = "ready"

	// StatusRunning indicates the agent is currently running
	StatusRunning Status = "running"

	// StatusStopping indicates the agent is currently stopping
	StatusStopping Status = "stopping"

	// StatusStopped indicates the agent has been stopped
	StatusStopped Status = "stopped"

	// StatusError indicates the agent encountered an error
	StatusError Status = "error"
)

// AgentOption is a function that configures an agent
type AgentOption func(Agent) error

// ExtendedAgent interface for agents that support status tracking
type ExtendedAgent interface {
	Agent

	// Status returns the current status of the agent
	Status() Status

	// SetStatus updates the agent's status
	SetStatus(status Status)
}

// AsyncAgent interface for agents that support asynchronous operations
type AsyncAgent interface {
	Agent

	// RunAsync executes the agent asynchronously and returns a channel for results
	RunAsync(ctx context.Context, input interface{}) (<-chan interface{}, <-chan error)
}

// ConfigurableAgent interface for agents that support runtime configuration
type ConfigurableAgent interface {
	Agent

	// Configure applies configuration options to the agent
	Configure(options ...AgentOption) error

	// GetConfig returns the current configuration
	GetConfig() map[string]interface{}
}
