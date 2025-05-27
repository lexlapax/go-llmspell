// ABOUTME: Engine-specific error types and definitions
// ABOUTME: Provides common error handling for all script engines

package engine

import (
	"errors"
	"fmt"
)

// Common engine errors
var (
	// ErrEngineNotInitialized is returned when engine methods are called before initialization
	ErrEngineNotInitialized = errors.New("engine not initialized")

	// ErrScriptNotLoaded is returned when execution is attempted without a loaded script
	ErrScriptNotLoaded = errors.New("no script loaded")

	// ErrScriptExecutionFailed is returned when script execution fails
	ErrScriptExecutionFailed = errors.New("script execution failed")

	// ErrInvalidConfiguration is returned when engine configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid engine configuration")

	// ErrMemoryLimitExceeded is returned when script exceeds memory limit
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")

	// ErrExecutionTimeout is returned when script execution times out
	ErrExecutionTimeout = errors.New("execution timeout")

	// ErrStackOverflow is returned when script exceeds stack depth limit
	ErrStackOverflow = errors.New("stack overflow")

	// ErrFunctionNotFound is returned when a script calls an undefined function
	ErrFunctionNotFound = errors.New("function not found")

	// ErrInvalidFunctionSignature is returned when registering a function with invalid signature
	ErrInvalidFunctionSignature = errors.New("invalid function signature")

	// ErrVariableNotFound is returned when accessing an undefined variable
	ErrVariableNotFound = errors.New("variable not found")

	// ErrTypeConversion is returned when type conversion fails
	ErrTypeConversion = errors.New("type conversion failed")

	// ErrSandboxViolation is returned when script attempts forbidden operations
	ErrSandboxViolation = errors.New("sandbox security violation")
)

// ScriptError represents an error that occurred during script execution
type ScriptError struct {
	// ScriptName is the name of the script that failed
	ScriptName string

	// Line is the line number where the error occurred (0 if unknown)
	Line int

	// Column is the column number where the error occurred (0 if unknown)
	Column int

	// Message is the error message
	Message string

	// Cause is the underlying error
	Cause error
}

// Error implements the error interface
func (e ScriptError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s", e.ScriptName, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.ScriptName, e.Message)
}

// Unwrap returns the underlying error
func (e ScriptError) Unwrap() error {
	return e.Cause
}

// LoadError represents an error that occurred while loading a script
type LoadError struct {
	// Path is the path to the script file (if applicable)
	Path string

	// Cause is the underlying error
	Cause error
}

// Error implements the error interface
func (e LoadError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("failed to load script %s: %v", e.Path, e.Cause)
	}
	return fmt.Sprintf("failed to load script: %v", e.Cause)
}

// Unwrap returns the underlying error
func (e LoadError) Unwrap() error {
	return e.Cause
}

// ConfigError represents an error in engine configuration
type ConfigError struct {
	// Field is the configuration field that has an error
	Field string

	// Value is the invalid value
	Value interface{}

	// Message describes the error
	Message string
}

// Error implements the error interface
func (e ConfigError) Error() string {
	return fmt.Sprintf("invalid configuration for %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// SecurityError represents a security violation
type SecurityError struct {
	// Operation is the forbidden operation that was attempted
	Operation string

	// Resource is the resource that was accessed (if applicable)
	Resource string

	// Message provides additional context
	Message string
}

// Error implements the error interface
func (e SecurityError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("security violation: %s attempted on %s: %s", e.Operation, e.Resource, e.Message)
	}
	return fmt.Sprintf("security violation: %s: %s", e.Operation, e.Message)
}

// IsScriptError checks if an error is a ScriptError
func IsScriptError(err error) bool {
	var scriptErr ScriptError
	return errors.As(err, &scriptErr)
}

// IsLoadError checks if an error is a LoadError
func IsLoadError(err error) bool {
	var loadErr LoadError
	return errors.As(err, &loadErr)
}

// IsConfigError checks if an error is a ConfigError
func IsConfigError(err error) bool {
	var configErr ConfigError
	return errors.As(err, &configErr)
}

// IsSecurityError checks if an error is a SecurityError
func IsSecurityError(err error) bool {
	var secErr SecurityError
	return errors.As(err, &secErr)
}
