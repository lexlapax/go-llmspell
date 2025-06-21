// ABOUTME: This file defines standard error types, categories, and exit codes for the go-llmspell CLI.
// ABOUTME: It provides user-friendly error formatting with context and suggestions for common issues.

package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Exit codes for CLI operations
const (
	ExitSuccess         = 0
	ExitGeneralError    = 1
	ExitUsageError      = 2
	ExitConfigError     = 3
	ExitScriptError     = 4
	ExitEngineError     = 5
	ExitSecurityError   = 6
	ExitNetworkError    = 7
	ExitTimeoutError    = 8
	ExitResourceError   = 9
	ExitValidationError = 10
	ExitDependencyError = 11
	ExitIOError         = 12
	ExitInterrupted     = 130
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	CategoryUnknown     ErrorCategory = "unknown"
	CategoryUsage       ErrorCategory = "usage"
	CategoryConfig      ErrorCategory = "config"
	CategoryScript      ErrorCategory = "script"
	CategoryEngine      ErrorCategory = "engine"
	CategorySecurity    ErrorCategory = "security"
	CategoryNetwork     ErrorCategory = "network"
	CategoryTimeout     ErrorCategory = "timeout"
	CategoryResource    ErrorCategory = "resource"
	CategoryValidation  ErrorCategory = "validation"
	CategoryDependency  ErrorCategory = "dependency"
	CategoryIO          ErrorCategory = "io"
	CategoryInterrupted ErrorCategory = "interrupted"
)

// SpellError is the base error type for go-llmspell
type SpellError struct {
	Category    ErrorCategory
	Code        int
	Message     string
	Cause       error
	Context     map[string]interface{}
	Suggestions []string
	StackTrace  []StackFrame
}

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// Error implements the error interface
func (e *SpellError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *SpellError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is implements errors.Is support
func (e *SpellError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}

	var targetSpellError *SpellError
	if errors.As(target, &targetSpellError) {
		return e.Category == targetSpellError.Category && e.Code == targetSpellError.Code
	}

	return errors.Is(e.Cause, target)
}

// ExitCode returns the appropriate exit code for this error
func (e *SpellError) ExitCode() int {
	if e == nil {
		return 0
	}
	if e.Code != 0 {
		return e.Code
	}

	switch e.Category {
	case CategoryUsage:
		return ExitUsageError
	case CategoryConfig:
		return ExitConfigError
	case CategoryScript:
		return ExitScriptError
	case CategoryEngine:
		return ExitEngineError
	case CategorySecurity:
		return ExitSecurityError
	case CategoryNetwork:
		return ExitNetworkError
	case CategoryTimeout:
		return ExitTimeoutError
	case CategoryResource:
		return ExitResourceError
	case CategoryValidation:
		return ExitValidationError
	case CategoryDependency:
		return ExitDependencyError
	case CategoryIO:
		return ExitIOError
	case CategoryInterrupted:
		return ExitInterrupted
	default:
		return ExitGeneralError
	}
}

// WithContext adds context information to the error
func (e *SpellError) WithContext(key string, value interface{}) *SpellError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion to help resolve the error
func (e *SpellError) WithSuggestion(suggestion string) *SpellError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	// Capture up to 20 stack frames to ensure we get the test function
	pcs := make([]uintptr, 20)
	n := runtime.Callers(skip+2, pcs)

	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pcs[i])
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// New creates a new SpellError
func New(category ErrorCategory, message string) *SpellError {
	return &SpellError{
		Category:   category,
		Message:    message,
		StackTrace: captureStackTrace(1),
	}
}

// Newf creates a new SpellError with formatted message
func Newf(category ErrorCategory, format string, args ...interface{}) *SpellError {
	return &SpellError{
		Category:   category,
		Message:    fmt.Sprintf(format, args...),
		StackTrace: captureStackTrace(1),
	}
}

// Wrap wraps an existing error with SpellError
func Wrap(err error, category ErrorCategory, message string) *SpellError {
	if err == nil {
		return nil
	}

	// If already a SpellError, preserve its properties
	var spellErr *SpellError
	if errors.As(err, &spellErr) {
		return &SpellError{
			Category:    category,
			Code:        spellErr.Code,
			Message:     message,
			Cause:       err,
			Context:     spellErr.Context,
			Suggestions: spellErr.Suggestions,
			StackTrace:  captureStackTrace(1),
		}
	}

	return &SpellError{
		Category:   category,
		Message:    message,
		Cause:      err,
		StackTrace: captureStackTrace(1),
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, category ErrorCategory, format string, args ...interface{}) *SpellError {
	if err == nil {
		return nil
	}

	return Wrap(err, category, fmt.Sprintf(format, args...))
}

// Common error constructors

// UsageError creates a usage error
func UsageError(message string) *SpellError {
	return New(CategoryUsage, message).
		WithSuggestion("Use 'llmspell --help' for usage information")
}

// ConfigError creates a configuration error
func ConfigError(message string) *SpellError {
	return New(CategoryConfig, message).
		WithSuggestion("Check your configuration file syntax").
		WithSuggestion("Use 'llmspell config validate' to check configuration")
}

// ScriptError creates a script execution error
func ScriptError(message string) *SpellError {
	return New(CategoryScript, message).
		WithSuggestion("Check the script syntax").
		WithSuggestion("Use 'llmspell validate <script>' to check for errors")
}

// EngineError creates an engine error
func EngineError(engine, message string) *SpellError {
	return New(CategoryEngine, message).
		WithContext("engine", engine)
}

// SecurityError creates a security error
func SecurityError(message string) *SpellError {
	return New(CategorySecurity, message).
		WithSuggestion("Check security profile settings").
		WithSuggestion("Use '--profile development' for less restrictive mode (development only)")
}

// NetworkError creates a network error
func NetworkError(message string) *SpellError {
	return New(CategoryNetwork, message).
		WithSuggestion("Check your network connection").
		WithSuggestion("Verify proxy settings if behind a firewall")
}

// TimeoutError creates a timeout error
func TimeoutError(message string) *SpellError {
	return New(CategoryTimeout, message).
		WithSuggestion("Try increasing the timeout with --timeout flag").
		WithSuggestion("Check if the script has infinite loops")
}

// ResourceError creates a resource limit error
func ResourceError(message string) *SpellError {
	return New(CategoryResource, message).
		WithSuggestion("Try reducing memory usage in your script").
		WithSuggestion("Increase limits with --memory-limit or --cpu-limit flags")
}

// ValidationError creates a validation error
func ValidationError(message string) *SpellError {
	return New(CategoryValidation, message)
}

// DependencyError creates a dependency error
func DependencyError(message string) *SpellError {
	return New(CategoryDependency, message).
		WithSuggestion("Check if all required dependencies are installed").
		WithSuggestion("Run 'llmspell doctor' to diagnose issues")
}

// IOError creates an I/O error
func IOError(message string) *SpellError {
	return New(CategoryIO, message).
		WithSuggestion("Check file permissions").
		WithSuggestion("Ensure the path exists and is accessible")
}

// InterruptedError creates an interrupted error
func InterruptedError() *SpellError {
	return New(CategoryInterrupted, "Operation interrupted by user")
}

// IsSpellError checks if an error is a SpellError
func IsSpellError(err error) bool {
	var spellErr *SpellError
	return errors.As(err, &spellErr)
}

// GetCategory returns the category of an error
func GetCategory(err error) ErrorCategory {
	var spellErr *SpellError
	if errors.As(err, &spellErr) {
		return spellErr.Category
	}
	return CategoryUnknown
}

// GetExitCode returns the appropriate exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var spellErr *SpellError
	if errors.As(err, &spellErr) {
		return spellErr.ExitCode()
	}

	return ExitGeneralError
}

// GetSuggestions returns suggestions for resolving an error
func GetSuggestions(err error) []string {
	var spellErr *SpellError
	if errors.As(err, &spellErr) {
		return spellErr.Suggestions
	}
	return nil
}

// GetContext returns the context of an error
func GetContext(err error) map[string]interface{} {
	var spellErr *SpellError
	if errors.As(err, &spellErr) {
		return spellErr.Context
	}
	return nil
}

// Chain represents a chain of errors
type Chain struct {
	errors []error
}

// NewChain creates a new error chain
func NewChain() *Chain {
	return &Chain{
		errors: make([]error, 0),
	}
}

// Add adds an error to the chain
func (c *Chain) Add(err error) {
	if err != nil {
		c.errors = append(c.errors, err)
	}
}

// Addf adds a formatted error to the chain
func (c *Chain) Addf(category ErrorCategory, format string, args ...interface{}) {
	c.Add(Newf(category, format, args...))
}

// HasErrors returns true if the chain contains errors
func (c *Chain) HasErrors() bool {
	if c == nil {
		return false
	}
	return len(c.errors) > 0
}

// Errors returns all errors in the chain
func (c *Chain) Errors() []error {
	if c == nil {
		return nil
	}
	return c.errors
}

// First returns the first error in the chain
func (c *Chain) First() error {
	if c == nil || len(c.errors) == 0 {
		return nil
	}
	return c.errors[0]
}

// Error implements the error interface
func (c *Chain) Error() string {
	if c == nil || len(c.errors) == 0 {
		return ""
	}

	var messages []string
	for _, err := range c.errors {
		messages = append(messages, err.Error())
	}

	return strings.Join(messages, "; ")
}

// Merge merges another chain into this one
func (c *Chain) Merge(other *Chain) {
	if other != nil {
		c.errors = append(c.errors, other.errors...)
	}
}
