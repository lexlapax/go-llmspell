// ABOUTME: Defines error types for the agent system
// ABOUTME: Provides structured error handling for agent operations

package agents

import "fmt"

// Error represents an agent-specific error
type Error struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrCodeNotInitialized  = "AGENT_NOT_INITIALIZED"
	ErrCodeInvalidConfig   = "INVALID_CONFIG"
	ErrCodeToolNotFound    = "TOOL_NOT_FOUND"
	ErrCodeExecutionFailed = "EXECUTION_FAILED"
	ErrCodeNoUserMessage   = "NO_USER_MESSAGE"
)
