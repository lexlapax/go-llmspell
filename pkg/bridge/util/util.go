// ABOUTME: General utilities bridge provides access to miscellaneous go-llms utility functions.
// ABOUTME: Wraps utilities that don't fit into specific categories like error handling and misc helpers.

// TODO: Consider upstreaming general-purpose utilities to go-llms that aren't specific to bridges/scripts:
// - String manipulation (truncate, sanitize)
// - UUID generation wrapper
// - Hash utilities (consistent hashing interface)
// - Time/duration parsing and formatting
// - Retry/backoff utilities
// - Common validation functions (URL, email)
// These could be useful for go-llms internals and other consumers of the library.

package util

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/go-llmspell/pkg/engine"
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
func (b *UtilBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *UtilBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "generateUUID":
		return engine.NewStringValue(uuid.New().String()), nil

	case "truncateString":
		if len(args) < 2 {
			return nil, fmt.Errorf("truncateString requires text and maxLength parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("text must be string")
		}
		text := args[0].(engine.StringValue).Value()

		if args[1] == nil || args[1].Type() != engine.TypeNumber {
			return nil, fmt.Errorf("maxLength must be number")
		}
		maxLength := int(args[1].(engine.NumberValue).Value())

		suffix := "..."
		if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeString {
			suffix = args[2].(engine.StringValue).Value()
		}

		if len(text) <= maxLength {
			return engine.NewStringValue(text), nil
		}

		if maxLength <= len(suffix) {
			return engine.NewStringValue(suffix), nil
		}

		return engine.NewStringValue(text[:maxLength-len(suffix)] + suffix), nil

	case "hashString":
		if len(args) < 1 {
			return nil, fmt.Errorf("hashString requires text parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("text must be string")
		}
		text := args[0].(engine.StringValue).Value()

		algorithm := "sha256"
		if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeString {
			algorithm = args[1].(engine.StringValue).Value()
		}

		var h hash.Hash
		switch strings.ToLower(algorithm) {
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		case "md5":
			h = md5.New()
		default:
			return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
		}

		h.Write([]byte(text))
		return engine.NewStringValue(hex.EncodeToString(h.Sum(nil))), nil

	case "sleep":
		if len(args) < 1 {
			return nil, fmt.Errorf("sleep requires milliseconds parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeNumber {
			return nil, fmt.Errorf("milliseconds must be number")
		}
		ms := args[0].(engine.NumberValue).Value()

		time.Sleep(time.Duration(ms) * time.Millisecond)
		return engine.NewNilValue(), nil

	case "formatDuration":
		if len(args) < 1 {
			return nil, fmt.Errorf("formatDuration requires milliseconds parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeNumber {
			return nil, fmt.Errorf("milliseconds must be number")
		}
		ms := args[0].(engine.NumberValue).Value()

		d := time.Duration(ms) * time.Millisecond
		return engine.NewStringValue(d.String()), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
