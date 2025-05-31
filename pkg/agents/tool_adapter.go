// ABOUTME: Adapts llmspell tools to work with go-llms agents
// ABOUTME: Provides bidirectional conversion between tool interfaces

package agents

import (
	"context"
	"encoding/json"
	"fmt"

	agentdomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llmspell/pkg/tools"
)

// toolAdapter adapts our Tool interface to go-llms Tool interface
type toolAdapter struct {
	tool tools.Tool
}

// Name returns the tool's name
func (t *toolAdapter) Name() string {
	return t.tool.Name()
}

// Description returns the tool's description
func (t *toolAdapter) Description() string {
	return t.tool.Description()
}

// Execute runs the tool with parameters
func (t *toolAdapter) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	// Convert params to map if needed
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		// Try to convert via JSON
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &paramMap); err != nil {
			return nil, err
		}
	}

	return t.tool.Execute(ctx, paramMap)
}

// ParameterSchema returns the tool's parameter schema
func (t *toolAdapter) ParameterSchema() *schemadomain.Schema {
	// Get the raw JSON schema from our tool
	rawSchema := t.tool.Parameters()

	// Parse it into a map
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(rawSchema, &schemaMap); err != nil {
		// Return empty schema on error
		return &schemadomain.Schema{
			Type: "object",
		}
	}

	// Convert to go-llms schema
	return mapToSchema(schemaMap)
}

// mapToSchema converts a map representation to go-llms Schema
func mapToSchema(m map[string]interface{}) *schemadomain.Schema {
	schema := &schemadomain.Schema{}

	if typ, ok := m["type"].(string); ok {
		schema.Type = typ
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]schemadomain.Property)
		for name, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				schema.Properties[name] = mapToProperty(propMap)
			}
		}
	}

	if required, ok := m["required"].([]interface{}); ok {
		schema.Required = make([]string, len(required))
		for i, r := range required {
			if s, ok := r.(string); ok {
				schema.Required[i] = s
			}
		}
	}

	if addProps, ok := m["additionalProperties"].(bool); ok {
		schema.AdditionalProperties = &addProps
	}

	return schema
}

// mapToProperty converts a map to a Property
func mapToProperty(m map[string]interface{}) schemadomain.Property {
	prop := schemadomain.Property{}

	if typ, ok := m["type"].(string); ok {
		prop.Type = typ
	}

	if desc, ok := m["description"].(string); ok {
		prop.Description = desc
	}

	if format, ok := m["format"].(string); ok {
		prop.Format = format
	}

	if pattern, ok := m["pattern"].(string); ok {
		prop.Pattern = pattern
	}

	if enum, ok := m["enum"].([]interface{}); ok {
		// Convert to []string
		strEnum := make([]string, 0, len(enum))
		for _, e := range enum {
			if s, ok := e.(string); ok {
				strEnum = append(strEnum, s)
			} else {
				// Convert non-string to string
				strEnum = append(strEnum, fmt.Sprintf("%v", e))
			}
		}
		prop.Enum = strEnum
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		prop.Properties = make(map[string]schemadomain.Property)
		for name, propData := range props {
			if propMap, ok := propData.(map[string]interface{}); ok {
				prop.Properties[name] = mapToProperty(propMap)
			}
		}
	}

	if items, ok := m["items"].(map[string]interface{}); ok {
		itemProp := mapToProperty(items)
		prop.Items = &itemProp
	}

	return prop
}

// LLMSAgentAdapter adapts a go-llms agent to our Agent interface
type LLMSAgentAdapter struct {
	agent        agentdomain.Agent
	name         string
	systemPrompt string
	tools        []string
}

// NewLLMSAgentAdapter creates a new adapter for a go-llms agent
func NewLLMSAgentAdapter(name string, agent agentdomain.Agent) Agent {
	return &LLMSAgentAdapter{
		agent: agent,
		name:  name,
	}
}

// Name returns the agent's name
func (a *LLMSAgentAdapter) Name() string {
	return a.name
}

// Initialize prepares the agent for use
func (a *LLMSAgentAdapter) Initialize(ctx context.Context) error {
	// go-llms agents don't require initialization
	return nil
}

// Cleanup releases any resources
func (a *LLMSAgentAdapter) Cleanup() error {
	// go-llms agents don't require cleanup
	return nil
}

// Execute runs the agent with a single input
func (a *LLMSAgentAdapter) Execute(ctx context.Context, input string, opts *ExecutionOptions) (*ExecutionResult, error) {
	// Run the go-llms agent
	response, err := a.agent.Run(ctx, input)
	if err != nil {
		return nil, err
	}

	// Convert response to string
	responseStr, ok := response.(string)
	if !ok {
		// Try JSON conversion
		data, err := json.Marshal(response)
		if err != nil {
			responseStr = "Response could not be converted to string"
		} else {
			responseStr = string(data)
		}
	}

	return &ExecutionResult{
		Response: responseStr,
		Messages: []Message{
			NewUserMessage(input),
			NewAssistantMessage(responseStr),
		},
	}, nil
}

// ExecuteWithHistory runs the agent with conversation history
func (a *LLMSAgentAdapter) ExecuteWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions) (*ExecutionResult, error) {
	// Extract last user message
	var lastInput string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == UserRole {
			lastInput = messages[i].Content
			break
		}
	}

	if lastInput == "" {
		return nil, &Error{
			Code:    "NO_USER_MESSAGE",
			Message: "no user message found in history",
		}
	}

	result, err := a.Execute(ctx, lastInput, opts)
	if err != nil {
		return nil, err
	}

	// Update messages with full history
	result.Messages = append(messages, NewAssistantMessage(result.Response))
	return result, nil
}

// Stream executes with streaming (simulated for go-llms agents)
func (a *LLMSAgentAdapter) Stream(ctx context.Context, input string, opts *ExecutionOptions, callback StreamCallback) error {
	// Execute normally and simulate streaming
	result, err := a.Execute(ctx, input, opts)
	if err != nil {
		return err
	}

	// Send response in chunks
	chunkSize := 20
	for i := 0; i < len(result.Response); i += chunkSize {
		end := i + chunkSize
		if end > len(result.Response) {
			end = len(result.Response)
		}

		if err := callback(result.Response[i:end]); err != nil {
			return err
		}
	}

	return nil
}

// StreamWithHistory streams with history
func (a *LLMSAgentAdapter) StreamWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions, callback StreamCallback) error {
	var lastInput string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == UserRole {
			lastInput = messages[i].Content
			break
		}
	}

	if lastInput == "" {
		return &Error{
			Code:    "NO_USER_MESSAGE",
			Message: "no user message found in history",
		}
	}

	return a.Stream(ctx, lastInput, opts, callback)
}

// SetSystemPrompt updates the system prompt
func (a *LLMSAgentAdapter) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
	a.agent.SetSystemPrompt(prompt)
}

// GetSystemPrompt returns the current system prompt
func (a *LLMSAgentAdapter) GetSystemPrompt() string {
	return a.systemPrompt
}

// AddTool adds a tool to the agent
func (a *LLMSAgentAdapter) AddTool(toolName string) error {
	// Get the tool from registry
	toolRegistry := tools.DefaultRegistry
	tool, err := toolRegistry.Get(toolName)
	if err != nil {
		return err
	}

	// Adapt and add to go-llms agent
	llmsTool := &toolAdapter{tool: tool}
	a.agent.AddTool(llmsTool)

	a.tools = append(a.tools, toolName)
	return nil
}

// GetTools returns the list of tools
func (a *LLMSAgentAdapter) GetTools() []string {
	result := make([]string, len(a.tools))
	copy(result, a.tools)
	return result
}
