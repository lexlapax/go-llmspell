// ABOUTME: Provides integration with built-in tools from the go-llms library
// ABOUTME: Allows go-llms tools to be used through the llmspell tool system

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	agentdomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	agenttools "github.com/lexlapax/go-llms/pkg/agent/tools"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// LLMSToolAdapter adapts a go-llms tool to our Tool interface
type LLMSToolAdapter struct {
	tool agentdomain.Tool
}

// NewLLMSToolAdapter creates a new adapter for a go-llms tool
func NewLLMSToolAdapter(tool agentdomain.Tool) Tool {
	return &LLMSToolAdapter{tool: tool}
}

// Name returns the tool's name
func (a *LLMSToolAdapter) Name() string {
	return a.tool.Name()
}

// Description returns the tool's description
func (a *LLMSToolAdapter) Description() string {
	return a.tool.Description()
}

// Parameters returns the tool's parameter schema as JSON
func (a *LLMSToolAdapter) Parameters() json.RawMessage {
	schema := a.tool.ParameterSchema()
	if schema == nil {
		return json.RawMessage("{}")
	}

	// Convert the schema to JSON
	data, err := json.Marshal(schemaToMap(schema))
	if err != nil {
		// Return empty object on error
		return json.RawMessage("{}")
	}

	return data
}

// Execute runs the tool with the given parameters
func (a *LLMSToolAdapter) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return a.tool.Execute(ctx, params)
}

// schemaToMap converts a go-llms schema to a map for JSON serialization
func schemaToMap(schema *schemadomain.Schema) map[string]interface{} {
	result := make(map[string]interface{})

	if schema.Type != "" {
		result["type"] = schema.Type
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for name, prop := range schema.Properties {
			propMap := make(map[string]interface{})
			if prop.Type != "" {
				propMap["type"] = prop.Type
			}
			if prop.Description != "" {
				propMap["description"] = prop.Description
			}
			if prop.Format != "" {
				propMap["format"] = prop.Format
			}
			if prop.Pattern != "" {
				propMap["pattern"] = prop.Pattern
			}
			if prop.Enum != nil {
				propMap["enum"] = prop.Enum
			}
			// Handle nested properties
			if prop.Properties != nil {
				propMap["properties"] = schemaPropertiesToMap(prop.Properties)
			}
			if prop.Items != nil {
				propMap["items"] = propertyToMap(*prop.Items)
			}
			props[name] = propMap
		}
		result["properties"] = props
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	if schema.AdditionalProperties != nil {
		result["additionalProperties"] = *schema.AdditionalProperties
	}

	return result
}

// schemaPropertiesToMap converts schema properties to a map
func schemaPropertiesToMap(props map[string]schemadomain.Property) map[string]interface{} {
	result := make(map[string]interface{})
	for name, prop := range props {
		result[name] = propertyToMap(prop)
	}
	return result
}

// propertyToMap converts a property to a map
func propertyToMap(prop schemadomain.Property) map[string]interface{} {
	propMap := make(map[string]interface{})
	if prop.Type != "" {
		propMap["type"] = prop.Type
	}
	if prop.Description != "" {
		propMap["description"] = prop.Description
	}
	if prop.Format != "" {
		propMap["format"] = prop.Format
	}
	if prop.Pattern != "" {
		propMap["pattern"] = prop.Pattern
	}
	if prop.Enum != nil {
		propMap["enum"] = prop.Enum
	}
	if prop.Properties != nil {
		propMap["properties"] = schemaPropertiesToMap(prop.Properties)
	}
	if prop.Items != nil {
		propMap["items"] = propertyToMap(*prop.Items)
	}
	return propMap
}

// BuiltinToolConfig controls which built-in tools are available
type BuiltinToolConfig struct {
	EnableWebFetch       bool
	EnableSearch         bool
	EnableExecuteCommand bool
	EnableReadFile       bool
	EnableWriteFile      bool
}

// DefaultBuiltinToolConfig returns a safe default configuration
func DefaultBuiltinToolConfig() *BuiltinToolConfig {
	return &BuiltinToolConfig{
		EnableWebFetch:       true,
		EnableSearch:         false, // Disabled by default as it's not implemented in go-llms
		EnableExecuteCommand: false, // Disabled by default for security
		EnableReadFile:       false, // Disabled by default for security
		EnableWriteFile:      false, // Disabled by default for security
	}
}

// RegisterBuiltinTools registers the enabled built-in tools from go-llms
func RegisterBuiltinTools(registry Registry, config *BuiltinToolConfig) error {
	if config == nil {
		config = DefaultBuiltinToolConfig()
	}

	if config.EnableWebFetch {
		tool := agenttools.WebFetch()
		adapter := NewLLMSToolAdapter(tool)
		if err := registry.Register(adapter); err != nil {
			return fmt.Errorf("failed to register WebFetch tool: %w", err)
		}
	}

	if config.EnableExecuteCommand {
		tool := agenttools.ExecuteCommand()
		adapter := NewLLMSToolAdapter(tool)
		if err := registry.Register(adapter); err != nil {
			return fmt.Errorf("failed to register ExecuteCommand tool: %w", err)
		}
	}

	if config.EnableReadFile {
		tool := agenttools.ReadFile()
		adapter := NewLLMSToolAdapter(tool)
		if err := registry.Register(adapter); err != nil {
			return fmt.Errorf("failed to register ReadFile tool: %w", err)
		}
	}

	if config.EnableWriteFile {
		tool := agenttools.WriteFile()
		adapter := NewLLMSToolAdapter(tool)
		if err := registry.Register(adapter); err != nil {
			return fmt.Errorf("failed to register WriteFile tool: %w", err)
		}
	}

	// Note: Search tool is not implemented in go-llms yet
	if config.EnableSearch {
		// TODO: Register Search tool when it's implemented in go-llms
		return fmt.Errorf("Search tool is not yet implemented in go-llms")
	}

	return nil
}
