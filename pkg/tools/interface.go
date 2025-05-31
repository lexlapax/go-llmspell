// ABOUTME: Defines the Tool interface and related types for the tool system
// ABOUTME: Provides a common abstraction for tools that can be used by scripts and agents

package tools

import (
	"context"
	"encoding/json"
)

// Tool defines the interface for all tools in the system
type Tool interface {
	// Name returns the unique name of the tool
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// Parameters returns the JSON schema for the tool's parameters
	Parameters() json.RawMessage

	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Metadata contains additional information about a tool
type Metadata struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Author      string          `json:"author"`
	Tags        []string        `json:"tags"`
	Parameters  json.RawMessage `json:"parameters"`
}

// Result represents the result of a tool execution
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ToolFunc is a function type that implements tool execution
type ToolFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// FunctionTool wraps a function to implement the Tool interface
type FunctionTool struct {
	name        string
	description string
	parameters  json.RawMessage
	fn          ToolFunc
}

// NewFunctionTool creates a new tool from a function
func NewFunctionTool(name, description string, parameters json.RawMessage, fn ToolFunc) *FunctionTool {
	return &FunctionTool{
		name:        name,
		description: description,
		parameters:  parameters,
		fn:          fn,
	}
}

// Name returns the tool's name
func (t *FunctionTool) Name() string {
	return t.name
}

// Description returns the tool's description
func (t *FunctionTool) Description() string {
	return t.description
}

// Parameters returns the tool's parameter schema
func (t *FunctionTool) Parameters() json.RawMessage {
	return t.parameters
}

// Execute runs the tool function
func (t *FunctionTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return t.fn(ctx, params)
}

// Validator defines the interface for parameter validation
type Validator interface {
	Validate(params map[string]interface{}) error
}

// Registry defines the interface for tool registration and lookup
type Registry interface {
	// Register adds a tool to the registry
	Register(tool Tool) error

	// Get retrieves a tool by name
	Get(name string) (Tool, error)

	// List returns all registered tools
	List() []Tool

	// Remove unregisters a tool
	Remove(name string) error
}
