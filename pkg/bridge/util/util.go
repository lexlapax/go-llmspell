// ABOUTME: General utilities bridge provides access to miscellaneous go-llms utility functions.
// ABOUTME: Wraps utilities that don't fit into specific categories like error handling and misc helpers.

package util

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	// go-llms imports for general utilities
	// TODO: Add specific utility imports as needed
)

// UtilBridge provides script access to general go-llms utilities.
type UtilBridge struct {
	mu          sync.RWMutex
	initialized bool
}

// NewUtilBridge creates a new utilities bridge.
func NewUtilBridge() *UtilBridge {
	return &UtilBridge{}
}

// GetID returns the bridge identifier.
func (b *UtilBridge) GetID() string {
	return "util"
}

// GetMetadata returns bridge metadata.
func (b *UtilBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util",
		Version:     "1.0.0",
		Description: "General utilities bridge for miscellaneous helper functions",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *UtilBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *UtilBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *UtilBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *UtilBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *UtilBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Error handling utilities
		{
			Name:        "isRetryableError",
			Description: "Check if an error is retryable",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to check", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "wrapError",
			Description: "Wrap error with additional context",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Original error", Required: true},
				{Name: "message", Type: "string", Description: "Context message", Required: true},
			},
			ReturnType: "error",
		},
		{
			Name:        "errorToString",
			Description: "Convert error to detailed string representation",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to convert", Required: true},
			},
			ReturnType: "string",
		},

		// String utilities
		{
			Name:        "truncateString",
			Description: "Truncate string to specified length",
			Parameters: []engine.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text to truncate", Required: true},
				{Name: "maxLength", Type: "number", Description: "Maximum length", Required: true},
				{Name: "suffix", Type: "string", Description: "Truncation suffix", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "sanitizeString",
			Description: "Sanitize string for safe output",
			Parameters: []engine.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text to sanitize", Required: true},
				{Name: "allowedChars", Type: "string", Description: "Allowed character set", Required: false},
			},
			ReturnType: "string",
		},

		// Time utilities
		{
			Name:        "parseHumanDuration",
			Description: "Parse human-readable duration (e.g., '2h30m')",
			Parameters: []engine.ParameterInfo{
				{Name: "duration", Type: "string", Description: "Human-readable duration", Required: true},
			},
			ReturnType: "number", // milliseconds
		},
		{
			Name:        "formatDuration",
			Description: "Format duration to human-readable string",
			Parameters: []engine.ParameterInfo{
				{Name: "milliseconds", Type: "number", Description: "Duration in milliseconds", Required: true},
			},
			ReturnType: "string",
		},

		// Retry utilities
		{
			Name:        "retryWithBackoff",
			Description: "Execute function with exponential backoff retry",
			Parameters: []engine.ParameterInfo{
				{Name: "fn", Type: "function", Description: "Function to retry", Required: true},
				{Name: "maxRetries", Type: "number", Description: "Maximum retry attempts", Required: true},
				{Name: "initialDelay", Type: "number", Description: "Initial delay in ms", Required: false},
			},
			ReturnType: "any",
		},
		{
			Name:        "createRetryConfig",
			Description: "Create retry configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "maxRetries", Type: "number", Description: "Maximum retries", Required: true},
				{Name: "backoffMultiplier", Type: "number", Description: "Backoff multiplier", Required: false},
				{Name: "maxDelay", Type: "number", Description: "Maximum delay in ms", Required: false},
			},
			ReturnType: "object",
		},

		// Validation utilities
		{
			Name:        "validateURL",
			Description: "Validate URL format",
			Parameters: []engine.ParameterInfo{
				{Name: "url", Type: "string", Description: "URL to validate", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "validateEmail",
			Description: "Validate email address format",
			Parameters: []engine.ParameterInfo{
				{Name: "email", Type: "string", Description: "Email to validate", Required: true},
			},
			ReturnType: "boolean",
		},

		// Misc utilities
		{
			Name:        "generateUUID",
			Description: "Generate a new UUID",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "string",
		},
		{
			Name:        "hashString",
			Description: "Generate hash of string",
			Parameters: []engine.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text to hash", Required: true},
				{Name: "algorithm", Type: "string", Description: "Hash algorithm (sha256/sha512/md5)", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "sleep",
			Description: "Sleep for specified duration",
			Parameters: []engine.ParameterInfo{
				{Name: "milliseconds", Type: "number", Description: "Sleep duration in ms", Required: true},
			},
			ReturnType: "void",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *UtilBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"error": {
			GoType:     "error",
			ScriptType: "object",
		},
		"function": {
			GoType:     "func() (interface{}, error)",
			ScriptType: "function",
		},
	}
}

// ValidateMethod validates method calls.
func (b *UtilBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *UtilBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "util",
			Actions:     []string{"read"},
			Description: "Access to utility functions",
		},
		{
			Type:        engine.PermissionTime,
			Resource:    "system",
			Actions:     []string{"sleep"},
			Description: "Time-based operations",
		},
	}
}

// The actual method implementations would be provided by the script engine
// which would call the appropriate go-llms utility functions or standard library functions.
