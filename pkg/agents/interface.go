// ABOUTME: Defines the Agent interface and related types for LLM agent orchestration
// ABOUTME: Provides a consistent API for creating and managing agents across the system

package agents

import (
	"context"
	"errors"
	"time"
)

// Role represents the role of a message in a conversation
type Role string

const (
	// UserRole represents a message from the user
	UserRole Role = "user"
	// AssistantRole represents a message from the assistant
	AssistantRole Role = "assistant"
	// SystemRole represents a system message
	SystemRole Role = "system"
	// ToolRole represents a tool response message
	ToolRole Role = "tool"
)

// Message represents a single message in a conversation
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// NewUserMessage creates a new user message
func NewUserMessage(content string) Message {
	return Message{Role: UserRole, Content: content}
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(content string) Message {
	return Message{Role: AssistantRole, Content: content}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(content string) Message {
	return Message{Role: SystemRole, Content: content}
}

// Config holds configuration for an agent
type Config struct {
	// Name is the unique identifier for the agent
	Name string `json:"name"`
	
	// SystemPrompt defines the agent's behavior and personality
	SystemPrompt string `json:"system_prompt"`
	
	// Provider specifies which LLM provider to use (e.g., "openai", "anthropic", "gemini")
	Provider string `json:"provider"`
	
	// Model specifies which model to use (e.g., "gpt-4", "claude-3", "gemini-pro")
	Model string `json:"model"`
	
	// Tools is a list of tool names available to the agent
	Tools []string `json:"tools,omitempty"`
	
	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`
	
	// Temperature controls randomness (0.0 to 1.0)
	Temperature float64 `json:"temperature,omitempty"`
	
	// Timeout for agent operations
	Timeout time.Duration `json:"timeout,omitempty"`
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.Name == "" {
		return errors.New("agent name is required")
	}
	if c.Provider == "" {
		return errors.New("provider is required")
	}
	if c.Model == "" {
		return errors.New("model is required")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	return nil
}

// ExecutionOptions provides options for agent execution
type ExecutionOptions struct {
	// Stream enables streaming responses
	Stream bool
	
	// MaxTokens overrides the agent's default max tokens
	MaxTokens int
	
	// Temperature overrides the agent's default temperature
	Temperature float64
	
	// Timeout overrides the agent's default timeout
	Timeout time.Duration
}

// ExecutionResult contains the result of an agent execution
type ExecutionResult struct {
	// Response is the final response text
	Response string
	
	// Messages contains the full conversation history
	Messages []Message
	
	// TokensUsed is the total number of tokens consumed
	TokensUsed int
	
	// Duration is how long the execution took
	Duration time.Duration
	
	// Metadata contains provider-specific information
	Metadata map[string]interface{}
}

// StreamCallback is called for each chunk of a streaming response
type StreamCallback func(chunk string) error

// Agent represents an LLM agent that can process inputs and generate responses
type Agent interface {
	// Name returns the agent's unique identifier
	Name() string
	
	// Initialize prepares the agent for use
	Initialize(ctx context.Context) error
	
	// Cleanup releases any resources held by the agent
	Cleanup() error
	
	// Execute runs the agent with a single input
	Execute(ctx context.Context, input string, opts *ExecutionOptions) (*ExecutionResult, error)
	
	// ExecuteWithHistory runs the agent with conversation history
	ExecuteWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions) (*ExecutionResult, error)
	
	// Stream executes the agent with streaming response
	Stream(ctx context.Context, input string, opts *ExecutionOptions, callback StreamCallback) error
	
	// StreamWithHistory executes with history and streaming response
	StreamWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions, callback StreamCallback) error
	
	// SetSystemPrompt updates the agent's system prompt
	SetSystemPrompt(prompt string)
	
	// GetSystemPrompt returns the current system prompt
	GetSystemPrompt() string
	
	// AddTool adds a tool to the agent
	AddTool(toolName string) error
	
	// GetTools returns the list of available tools
	GetTools() []string
}

// Factory is a function that creates a new agent
type Factory func(config Config) (Agent, error)

// Registry manages agent registration and creation
type Registry interface {
	// Register adds a new agent factory
	Register(name string, factory Factory) error
	
	// Create creates a new agent instance
	Create(config Config) (Agent, error)
	
	// Get retrieves an existing agent by name
	Get(name string) (Agent, error)
	
	// List returns all registered agent names
	List() []string
	
	// Remove removes an agent from the registry
	Remove(name string) error
}