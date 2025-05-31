// ABOUTME: Provides a mock agent implementation for testing purposes
// ABOUTME: Allows controlled testing of agent-dependent code

package agents

import (
	"context"
	"time"
)

// MockAgent implements the Agent interface for testing
type MockAgent struct {
	name         string
	systemPrompt string
	tools        []string
	initialized  bool
	responses    []string
	responseIdx  int
	executeErr   error
}

// NewMockAgent creates a new mock agent for testing
func NewMockAgent(name string) *MockAgent {
	return &MockAgent{
		name:      name,
		responses: []string{},
	}
}

// Name returns the agent's unique identifier
func (m *MockAgent) Name() string {
	return m.name
}

// Initialize prepares the agent for use
func (m *MockAgent) Initialize(ctx context.Context) error {
	m.initialized = true
	return nil
}

// Cleanup releases any resources held by the agent
func (m *MockAgent) Cleanup() error {
	m.initialized = false
	return nil
}

// Execute runs the agent with a single input
func (m *MockAgent) Execute(ctx context.Context, input string, opts *ExecutionOptions) (*ExecutionResult, error) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}

	response := "Mock response"
	if m.responseIdx < len(m.responses) {
		response = m.responses[m.responseIdx]
		m.responseIdx++
	}

	return &ExecutionResult{
		Response: response,
		Messages: []Message{
			NewUserMessage(input),
			NewAssistantMessage(response),
		},
		TokensUsed: len(input) + len(response),
		Duration:   10 * time.Millisecond,
	}, nil
}

// ExecuteWithHistory runs the agent with conversation history
func (m *MockAgent) ExecuteWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions) (*ExecutionResult, error) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}

	// Get last user message
	var lastInput string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == UserRole {
			lastInput = messages[i].Content
			break
		}
	}

	response := "Mock response with history"
	if m.responseIdx < len(m.responses) {
		response = m.responses[m.responseIdx]
		m.responseIdx++
	}

	allMessages := append(messages, NewAssistantMessage(response))

	return &ExecutionResult{
		Response:   response,
		Messages:   allMessages,
		TokensUsed: len(lastInput) + len(response),
		Duration:   15 * time.Millisecond,
	}, nil
}

// Stream executes the agent with streaming response
func (m *MockAgent) Stream(ctx context.Context, input string, opts *ExecutionOptions, callback StreamCallback) error {
	if m.executeErr != nil {
		return m.executeErr
	}

	response := "Mock streaming response"
	if m.responseIdx < len(m.responses) {
		response = m.responses[m.responseIdx]
		m.responseIdx++
	}

	// Simulate streaming by sending response in chunks
	words := []string{"Mock", " ", "streaming", " ", "response"}
	if len(m.responses) > 0 {
		words = splitIntoChunks(response, 5)
	}

	for _, chunk := range words {
		if err := callback(chunk); err != nil {
			return err
		}
		// Simulate delay
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}

	return nil
}

// StreamWithHistory executes with history and streaming response
func (m *MockAgent) StreamWithHistory(ctx context.Context, messages []Message, opts *ExecutionOptions, callback StreamCallback) error {
	if m.executeErr != nil {
		return m.executeErr
	}

	response := "Mock streaming response with history"
	if m.responseIdx < len(m.responses) {
		response = m.responses[m.responseIdx]
		m.responseIdx++
	}

	words := splitIntoChunks(response, 5)
	for _, chunk := range words {
		if err := callback(chunk); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}

	return nil
}

// SetSystemPrompt updates the agent's system prompt
func (m *MockAgent) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}

// GetSystemPrompt returns the current system prompt
func (m *MockAgent) GetSystemPrompt() string {
	return m.systemPrompt
}

// AddTool adds a tool to the agent
func (m *MockAgent) AddTool(toolName string) error {
	m.tools = append(m.tools, toolName)
	return nil
}

// GetTools returns the list of available tools
func (m *MockAgent) GetTools() []string {
	return m.tools
}

// SetResponse sets a response for the mock agent
func (m *MockAgent) SetResponse(response string) {
	m.responses = append(m.responses, response)
}

// SetError sets an error for the mock agent to return
func (m *MockAgent) SetError(err error) {
	m.executeErr = err
}

// Helper function to split string into chunks
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	runes := []rune(s)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}
