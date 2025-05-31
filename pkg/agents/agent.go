// ABOUTME: Provides the default agent implementation wrapping go-llms agents
// ABOUTME: Integrates with the tool registry and provides a consistent API

package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	agentworkflow "github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
	"github.com/lexlapax/go-llmspell/pkg/tools"
)

// defaultAgent is the standard implementation of the Agent interface
type defaultAgent struct {
	config       Config
	llmsAgent    *agentworkflow.DefaultAgent
	systemPrompt string
	tools        []string
	mu           sync.RWMutex
	initialized  bool
}

// NewDefaultAgent creates a new default agent with the given configuration
func NewDefaultAgent(config Config) Agent {
	return &defaultAgent{
		config:       config,
		systemPrompt: config.SystemPrompt,
		tools:        config.Tools,
	}
}

// Name returns the agent's unique identifier
func (a *defaultAgent) Name() string {
	return a.config.Name
}

// Initialize prepares the agent for use
func (a *defaultAgent) Initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		return nil
	}

	// Create LLM provider
	llm, err := createLLMProvider(a.config)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Create the go-llms agent
	a.llmsAgent = agentworkflow.NewAgent(llm)
	
	// Set system prompt
	if a.systemPrompt != "" {
		a.llmsAgent.SetSystemPrompt(a.systemPrompt)
	}
	
	// Set model
	if a.config.Model != "" {
		a.llmsAgent.WithModel(a.config.Model)
	}

	// Add tools if specified
	toolRegistry := tools.DefaultRegistry
	for _, toolName := range a.tools {
		tool, err := toolRegistry.Get(toolName)
		if err != nil {
			// Log warning but don't fail initialization
			continue
		}

		// Convert our tool to go-llms tool
		llmsTool := &toolAdapter{tool: tool}
		a.llmsAgent.AddTool(llmsTool)
	}

	a.initialized = true
	return nil
}

// Cleanup releases any resources held by the agent
func (a *defaultAgent) Cleanup() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.initialized = false
	a.llmsAgent = nil
	return nil
}

// Execute runs the agent with a single input
func (a *defaultAgent) Execute(ctx context.Context, input string, opts *ExecutionOptions) (*ExecutionResult, error) {
	a.mu.RLock()
	if !a.initialized {
		a.mu.RUnlock()
		return nil, fmt.Errorf("agent not initialized")
	}
	agent := a.llmsAgent
	a.mu.RUnlock()

	// Apply execution options
	if opts != nil && opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Execute the agent
	start := time.Now()
	response, err := agent.Run(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Convert response to string
	responseStr := fmt.Sprintf("%v", response)

	return &ExecutionResult{
		Response: responseStr,
		Messages: []Message{
			NewUserMessage(input),
			NewAssistantMessage(responseStr),
		},
		Duration: time.Since(start),
	}, nil
}

// ExecuteWithHistory runs the agent with conversation history
func (a *defaultAgent) ExecuteWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions) (*ExecutionResult, error) {
	a.mu.RLock()
	if !a.initialized {
		a.mu.RUnlock()
		return nil, fmt.Errorf("agent not initialized")
	}
	a.mu.RUnlock()

	// For now, we'll just use the last user message
	// Full history support would require modifying the go-llms agent
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == UserRole {
			lastUserMessage = messages[i].Content
			break
		}
	}

	if lastUserMessage == "" {
		return nil, fmt.Errorf("no user message found in history")
	}

	result, err := a.Execute(ctx, lastUserMessage, opts)
	if err != nil {
		return nil, err
	}

	// Replace messages with full history
	result.Messages = append(messages, NewAssistantMessage(result.Response))
	return result, nil
}

// Stream executes the agent with streaming response
func (a *defaultAgent) Stream(ctx context.Context, input string, opts *ExecutionOptions, callback StreamCallback) error {
	a.mu.RLock()
	if !a.initialized {
		a.mu.RUnlock()
		return fmt.Errorf("agent not initialized")
	}
	a.mu.RUnlock()

	// For now, execute normally and simulate streaming
	// Full streaming support would require go-llms agent modifications
	result, err := a.Execute(ctx, input, opts)
	if err != nil {
		return err
	}

	// Simulate streaming by sending response in chunks
	chunkSize := 10
	response := result.Response
	for i := 0; i < len(response); i += chunkSize {
		end := i + chunkSize
		if end > len(response) {
			end = len(response)
		}
		
		if err := callback(response[i:end]); err != nil {
			return err
		}
		
		// Small delay to simulate streaming
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}

	return nil
}

// StreamWithHistory executes with history and streaming response
func (a *defaultAgent) StreamWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions, callback StreamCallback) error {
	// Extract last user message
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == UserRole {
			lastUserMessage = messages[i].Content
			break
		}
	}

	if lastUserMessage == "" {
		return fmt.Errorf("no user message found in history")
	}

	return a.Stream(ctx, lastUserMessage, opts, callback)
}

// SetSystemPrompt updates the agent's system prompt
func (a *defaultAgent) SetSystemPrompt(prompt string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.systemPrompt = prompt
	if a.llmsAgent != nil {
		a.llmsAgent.SetSystemPrompt(prompt)
	}
}

// GetSystemPrompt returns the current system prompt
func (a *defaultAgent) GetSystemPrompt() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.systemPrompt
}

// AddTool adds a tool to the agent
func (a *defaultAgent) AddTool(toolName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if tool already exists
	for _, t := range a.tools {
		if t == toolName {
			return nil // Already added
		}
	}

	// Add to tools list
	a.tools = append(a.tools, toolName)

	// If initialized, add to llms agent
	if a.initialized && a.llmsAgent != nil {
		toolRegistry := tools.DefaultRegistry
		tool, err := toolRegistry.Get(toolName)
		if err != nil {
			return fmt.Errorf("tool %s not found: %w", toolName, err)
		}

		llmsTool := &toolAdapter{tool: tool}
		a.llmsAgent.AddTool(llmsTool)
	}

	return nil
}

// GetTools returns the list of available tools
func (a *defaultAgent) GetTools() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]string, len(a.tools))
	copy(result, a.tools)
	return result
}

// createLLMProvider creates an LLM provider based on the config
func createLLMProvider(config Config) (llmdomain.Provider, error) {
	llmConfig := llmutil.ModelConfig{
		Provider:  config.Provider,
		Model:     config.Model,
		MaxTokens: config.MaxTokens,
		// API key will be read from environment
	}

	return llmutil.CreateProvider(llmConfig)
}