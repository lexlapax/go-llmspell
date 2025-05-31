// ABOUTME: Bridge implementation for exposing tool functionality to scripts
// ABOUTME: Provides tool registration, execution, and management capabilities

package bridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lexlapax/go-llmspell/pkg/tools"
)

// ToolBridge provides tool functionality to script environments
type ToolBridge struct {
	registry tools.Registry
}

// NewToolBridge creates a new tool bridge
func NewToolBridge(registry tools.Registry) *ToolBridge {
	if registry == nil {
		registry = tools.DefaultRegistry
	}
	return &ToolBridge{
		registry: registry,
	}
}

// NewToolBridgeWithBuiltins creates a new tool bridge with built-in tools registered
func NewToolBridgeWithBuiltins(registry tools.Registry, config *tools.BuiltinToolConfig) (*ToolBridge, error) {
	if registry == nil {
		registry = tools.DefaultRegistry
	}

	// Register built-in tools
	if err := tools.RegisterBuiltinTools(registry, config); err != nil {
		return nil, fmt.Errorf("failed to register built-in tools: %w", err)
	}

	return &ToolBridge{
		registry: registry,
	}, nil
}

// RegisterTool registers a new tool from script
func (tb *ToolBridge) RegisterTool(name, description string, parameters map[string]interface{}, fn func(map[string]interface{}) (interface{}, error)) error {
	// Convert parameters to JSON
	paramsJSON, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	// Create a function tool
	tool := tools.NewFunctionTool(
		name,
		description,
		paramsJSON,
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Call the script function
			return fn(params)
		},
	)

	// Register the tool
	return tb.registry.Register(tool)
}

// ExecuteTool executes a tool by name
func (tb *ToolBridge) ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	// Get the tool
	tool, err := tb.registry.Get(name)
	if err != nil {
		return nil, err
	}

	// Execute the tool
	return tool.Execute(ctx, params)
}

// GetTool retrieves tool information
func (tb *ToolBridge) GetTool(name string) (map[string]interface{}, error) {
	tool, err := tb.registry.Get(name)
	if err != nil {
		return nil, err
	}

	// Build tool info
	info := map[string]interface{}{
		"name":        tool.Name(),
		"description": tool.Description(),
	}

	// Parse parameters to include as object
	var params interface{}
	if err := json.Unmarshal(tool.Parameters(), &params); err == nil {
		info["parameters"] = params
	} else {
		// If parsing fails, return as string
		info["parameters"] = string(tool.Parameters())
	}

	return info, nil
}

// ListTools returns all available tools
func (tb *ToolBridge) ListTools() []map[string]interface{} {
	tools := tb.registry.List()
	result := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
		}

		// Parse parameters to include as object
		var params interface{}
		if err := json.Unmarshal(tool.Parameters(), &params); err == nil {
			result[i]["parameters"] = params
		} else {
			// If parsing fails, return as string
			result[i]["parameters"] = string(tool.Parameters())
		}
	}

	return result
}

// RemoveTool unregisters a tool
func (tb *ToolBridge) RemoveTool(name string) error {
	return tb.registry.Remove(name)
}

// ValidateParameters validates tool parameters against schema
func (tb *ToolBridge) ValidateParameters(name string, params map[string]interface{}) error {
	tool, err := tb.registry.Get(name)
	if err != nil {
		return err
	}

	// Get parameter schema
	schema := tool.Parameters()
	if len(schema) == 0 {
		return nil // No schema to validate against
	}

	// Parse schema
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("failed to parse parameter schema: %w", err)
	}

	// Basic validation - check required fields
	if properties, ok := schemaMap["properties"].(map[string]interface{}); ok {
		if required, ok := schemaMap["required"].([]interface{}); ok {
			for _, req := range required {
				if reqName, ok := req.(string); ok {
					if _, exists := params[reqName]; !exists {
						return fmt.Errorf("missing required parameter: %s", reqName)
					}
				}
			}
		}

		// Type validation
		for paramName, paramValue := range params {
			if propDef, ok := properties[paramName].(map[string]interface{}); ok {
				if propType, ok := propDef["type"].(string); ok {
					if err := validateType(paramValue, propType); err != nil {
						return fmt.Errorf("parameter %s: %w", paramName, err)
					}
				}
			}
		}
	}

	return nil
}

// validateType checks if a value matches the expected type
func validateType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case float64, float32, int, int64, int32:
			// Valid number types
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	case "array":
		switch value.(type) {
		case []interface{}, []string, []float64, []int:
			// Valid array types
		default:
			return fmt.Errorf("expected array, got %T", value)
		}
	}
	return nil
}
