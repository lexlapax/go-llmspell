// ABOUTME: Tests for engine error types and error handling utilities
// ABOUTME: Validates error creation, formatting, and type checking

package engine

import (
	"errors"
	"fmt"
	"testing"
)

func TestScriptError(t *testing.T) {
	tests := []struct {
		name     string
		err      ScriptError
		expected string
	}{
		{
			name: "error with line and column",
			err: ScriptError{
				ScriptName: "test.lua",
				Line:       42,
				Column:     10,
				Message:    "undefined variable 'foo'",
				Cause:      errors.New("variable not found"),
			},
			expected: "test.lua:42:10: undefined variable 'foo'",
		},
		{
			name: "error without line number",
			err: ScriptError{
				ScriptName: "test.js",
				Message:    "syntax error",
			},
			expected: "test.js: syntax error",
		},
		{
			name: "error with line but no column",
			err: ScriptError{
				ScriptName: "test.tengo",
				Line:       10,
				Message:    "type mismatch",
			},
			expected: "test.tengo:10:0: type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}

			// Test Unwrap
			if tt.err.Cause != nil {
				unwrapped := tt.err.Unwrap()
				if unwrapped != tt.err.Cause {
					t.Errorf("Unwrap() = %v, want %v", unwrapped, tt.err.Cause)
				}
			}
		})
	}
}

func TestLoadError(t *testing.T) {
	tests := []struct {
		name     string
		err      LoadError
		expected string
	}{
		{
			name: "error with path",
			err: LoadError{
				Path:  "/path/to/script.lua",
				Cause: errors.New("file not found"),
			},
			expected: "failed to load script /path/to/script.lua: file not found",
		},
		{
			name: "error without path",
			err: LoadError{
				Cause: errors.New("invalid syntax"),
			},
			expected: "failed to load script: invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}

			// Test Unwrap
			unwrapped := tt.err.Unwrap()
			if unwrapped != tt.err.Cause {
				t.Errorf("Unwrap() = %v, want %v", unwrapped, tt.err.Cause)
			}
		})
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      ConfigError
		expected string
	}{
		{
			name: "memory limit error",
			err: ConfigError{
				Field:   "MaxMemory",
				Value:   -1,
				Message: "must be non-negative",
			},
			expected: "invalid configuration for MaxMemory: must be non-negative (value: -1)",
		},
		{
			name: "timeout error",
			err: ConfigError{
				Field:   "MaxExecutionTime",
				Value:   0,
				Message: "must be greater than zero",
			},
			expected: "invalid configuration for MaxExecutionTime: must be greater than zero (value: 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSecurityError(t *testing.T) {
	tests := []struct {
		name     string
		err      SecurityError
		expected string
	}{
		{
			name: "error with resource",
			err: SecurityError{
				Operation: "file access",
				Resource:  "/etc/passwd",
				Message:   "access denied",
			},
			expected: "security violation: file access attempted on /etc/passwd: access denied",
		},
		{
			name: "error without resource",
			err: SecurityError{
				Operation: "network connection",
				Message:   "outbound connections disabled",
			},
			expected: "security violation: network connection: outbound connections disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	// Test IsScriptError
	scriptErr := ScriptError{ScriptName: "test.lua", Message: "error"}
	wrappedScriptErr := fmt.Errorf("wrapped: %w", scriptErr)
	notScriptErr := errors.New("not a script error")

	if !IsScriptError(scriptErr) {
		t.Error("IsScriptError should return true for ScriptError")
	}
	if !IsScriptError(wrappedScriptErr) {
		t.Error("IsScriptError should return true for wrapped ScriptError")
	}
	if IsScriptError(notScriptErr) {
		t.Error("IsScriptError should return false for non-ScriptError")
	}

	// Test IsLoadError
	loadErr := LoadError{Path: "test.lua", Cause: errors.New("not found")}
	wrappedLoadErr := fmt.Errorf("wrapped: %w", loadErr)
	notLoadErr := errors.New("not a load error")

	if !IsLoadError(loadErr) {
		t.Error("IsLoadError should return true for LoadError")
	}
	if !IsLoadError(wrappedLoadErr) {
		t.Error("IsLoadError should return true for wrapped LoadError")
	}
	if IsLoadError(notLoadErr) {
		t.Error("IsLoadError should return false for non-LoadError")
	}

	// Test IsConfigError
	configErr := ConfigError{Field: "MaxMemory", Value: -1, Message: "invalid"}
	wrappedConfigErr := fmt.Errorf("wrapped: %w", configErr)
	notConfigErr := errors.New("not a config error")

	if !IsConfigError(configErr) {
		t.Error("IsConfigError should return true for ConfigError")
	}
	if !IsConfigError(wrappedConfigErr) {
		t.Error("IsConfigError should return true for wrapped ConfigError")
	}
	if IsConfigError(notConfigErr) {
		t.Error("IsConfigError should return false for non-ConfigError")
	}

	// Test IsSecurityError
	secErr := SecurityError{Operation: "file access", Message: "denied"}
	wrappedSecErr := fmt.Errorf("wrapped: %w", secErr)
	notSecErr := errors.New("not a security error")

	if !IsSecurityError(secErr) {
		t.Error("IsSecurityError should return true for SecurityError")
	}
	if !IsSecurityError(wrappedSecErr) {
		t.Error("IsSecurityError should return true for wrapped SecurityError")
	}
	if IsSecurityError(notSecErr) {
		t.Error("IsSecurityError should return false for non-SecurityError")
	}
}

func TestCommonErrors(t *testing.T) {
	// Verify that common errors are defined and have appropriate messages
	commonErrors := []struct {
		err      error
		contains string
	}{
		{ErrEngineNotInitialized, "not initialized"},
		{ErrScriptNotLoaded, "no script loaded"},
		{ErrScriptExecutionFailed, "execution failed"},
		{ErrInvalidConfiguration, "invalid"},
		{ErrMemoryLimitExceeded, "memory limit"},
		{ErrExecutionTimeout, "timeout"},
		{ErrStackOverflow, "stack overflow"},
		{ErrFunctionNotFound, "function not found"},
		{ErrInvalidFunctionSignature, "invalid function signature"},
		{ErrVariableNotFound, "variable not found"},
		{ErrTypeConversion, "type conversion"},
		{ErrSandboxViolation, "sandbox"},
	}

	for _, tc := range commonErrors {
		t.Run(tc.err.Error(), func(t *testing.T) {
			if tc.err == nil {
				t.Error("Error should not be nil")
			}
			errStr := tc.err.Error()
			if errStr == "" {
				t.Error("Error message should not be empty")
			}
			// Check that error message contains expected substring
			if len(tc.contains) > 0 && !contains(errStr, tc.contains) {
				t.Errorf("Error message %q should contain %q", errStr, tc.contains)
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
