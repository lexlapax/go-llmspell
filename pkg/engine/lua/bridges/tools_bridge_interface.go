// ABOUTME: Interface definition for tool bridge operations used by Lua bridge
// ABOUTME: Allows for easier testing by defining the contract needed

package bridges

import (
	"context"
)

// ToolBridgeInterface defines the methods needed by the Lua tools bridge
type ToolBridgeInterface interface {
	// RegisterTool registers a new tool from script
	RegisterTool(name, description string, parameters map[string]interface{}, fn func(map[string]interface{}) (interface{}, error)) error

	// ExecuteTool executes a tool by name with given parameters
	ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error)

	// GetTool returns information about a specific tool
	GetTool(name string) (map[string]interface{}, error)

	// ListTools returns information about all registered tools
	ListTools() []map[string]interface{}

	// RemoveTool removes a tool by name
	RemoveTool(name string) error

	// ValidateParameters validates tool parameters
	ValidateParameters(name string, params map[string]interface{}) error
}
