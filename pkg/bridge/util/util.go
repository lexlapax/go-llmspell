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

// Package util provides a bridge for general-purpose utility functions.
// It offers string manipulation, time/duration handling, UUID generation,
// hashing, validation, retry logic, and error handling utilities for script environments.
// These utilities complement go-llms functionality without reimplementing core features.
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
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// UtilBridge provides script access to general go-llms utilities.
// It bridges various utility functions including string manipulation, time handling,
// UUID generation, hashing, validation, and error handling. The bridge maintains
// thread-safe initialization state and provides a consistent interface for scripts.
type UtilBridge struct {
	mu          sync.RWMutex
	initialized bool
}

// NewUtilBridge creates a new utilities bridge.
// Returns an uninitialized bridge that must be initialized before use.
func NewUtilBridge() *UtilBridge {
	return &UtilBridge{}
}

// GetID returns the bridge identifier.
// Always returns "util" for this bridge.
func (b *UtilBridge) GetID() string {
	return "util"
}

// GetMetadata returns bridge metadata.
// Provides information about the bridge including name, version,
// description, author, and license for documentation and discovery.
func (b *UtilBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "util",
		Version:     "1.0.0",
		Description: "General utilities bridge for miscellaneous helper functions",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
// Thread-safe initialization that can be called multiple times safely.
// Returns nil on success or if already initialized.
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
// Marks the bridge as uninitialized. Thread-safe and idempotent.
func (b *UtilBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
// Thread-safe check of initialization status.
func (b *UtilBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// Delegates to the engine's RegisterBridge method for proper integration.
func (b *UtilBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns the methods exposed by this bridge.
// Provides comprehensive utility functions including error handling, string manipulation,
// time utilities, retry logic, validation, UUID generation, hashing, and sleep functionality.
func (b *UtilBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Error handling utilities
		{
			Name:        "isRetryableError",
			Description: "Check if an error is retryable",
			Parameters: []types.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to check", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "wrapError",
			Description: "Wrap error with additional context",
			Parameters: []types.ParameterInfo{
				{Name: "error", Type: "error", Description: "Original error", Required: true},
				{Name: "message", Type: "string", Description: "Context message", Required: true},
			},
			ReturnType: "error",
		},
		{
			Name:        "errorToString",
			Description: "Convert error to detailed string representation",
			Parameters: []types.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to convert", Required: true},
			},
			ReturnType: "string",
		},

		// String utilities
		{
			Name:        "truncateString",
			Description: "Truncate string to specified length",
			Parameters: []types.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text to truncate", Required: true},
				{Name: "maxLength", Type: "number", Description: "Maximum length", Required: true},
				{Name: "suffix", Type: "string", Description: "Truncation suffix", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "sanitizeString",
			Description: "Sanitize string for safe output",
			Parameters: []types.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text to sanitize", Required: true},
				{Name: "allowedChars", Type: "string", Description: "Allowed character set", Required: false},
			},
			ReturnType: "string",
		},

		// Time utilities
		{
			Name:        "parseHumanDuration",
			Description: "Parse human-readable duration (e.g., '2h30m')",
			Parameters: []types.ParameterInfo{
				{Name: "duration", Type: "string", Description: "Human-readable duration", Required: true},
			},
			ReturnType: "number", // milliseconds
		},
		{
			Name:        "formatDuration",
			Description: "Format duration to human-readable string",
			Parameters: []types.ParameterInfo{
				{Name: "milliseconds", Type: "number", Description: "Duration in milliseconds", Required: true},
			},
			ReturnType: "string",
		},

		// Retry utilities
		{
			Name:        "retryWithBackoff",
			Description: "Execute function with exponential backoff retry",
			Parameters: []types.ParameterInfo{
				{Name: "fn", Type: "function", Description: "Function to retry", Required: true},
				{Name: "maxRetries", Type: "number", Description: "Maximum retry attempts", Required: true},
				{Name: "initialDelay", Type: "number", Description: "Initial delay in ms", Required: false},
			},
			ReturnType: "any",
		},
		{
			Name:        "createRetryConfig",
			Description: "Create retry configuration",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "url", Type: "string", Description: "URL to validate", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "validateEmail",
			Description: "Validate email address format",
			Parameters: []types.ParameterInfo{
				{Name: "email", Type: "string", Description: "Email to validate", Required: true},
			},
			ReturnType: "boolean",
		},

		// Misc utilities
		{
			Name:        "generateUUID",
			Description: "Generate a new UUID",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "string",
		},
		{
			Name:        "hashString",
			Description: "Generate hash of string",
			Parameters: []types.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text to hash", Required: true},
				{Name: "algorithm", Type: "string", Description: "Hash algorithm (sha256/sha512/md5)", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "sleep",
			Description: "Sleep for specified duration",
			Parameters: []types.ParameterInfo{
				{Name: "milliseconds", Type: "number", Description: "Sleep duration in ms", Required: true},
			},
			ReturnType: "void",
		},
	}
}

// TypeMappings returns type conversion mappings.
// Maps Go error and function types to script-compatible object and function types
// for proper data conversion during method execution.
func (b *UtilBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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
// Currently delegates all validation to the engine based on method metadata.
// Returns nil as the engine handles parameter validation.
func (b *UtilBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
// Specifies that scripts need memory access for utility functions and
// time access for sleep operations.
func (b *UtilBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionMemory,
			Resource:    "util",
			Actions:     []string{"read"},
			Description: "Access to utility functions",
		},
		{
			Type:        types.PermissionTime,
			Resource:    "system",
			Actions:     []string{"sleep"},
			Description: "Time-based operations",
		},
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function.
// Routes method calls to appropriate utility implementations including UUID generation,
// string truncation, hashing, sleep, and duration formatting. Returns script-compatible
// values and handles parameter validation for each method.
func (b *UtilBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "generateUUID":
		return types.NewStringValue(uuid.New().String()), nil

	case "truncateString":
		if len(args) < 2 {
			return nil, fmt.Errorf("truncateString requires text and maxLength parameters")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("text must be string")
		}
		text := args[0].(types.StringValue).Value()

		if args[1] == nil || args[1].Type() != types.TypeNumber {
			return nil, fmt.Errorf("maxLength must be number")
		}
		maxLength := int(args[1].(types.NumberValue).Value())

		suffix := "..."
		if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeString {
			suffix = args[2].(types.StringValue).Value()
		}

		if len(text) <= maxLength {
			return types.NewStringValue(text), nil
		}

		if maxLength <= len(suffix) {
			return types.NewStringValue(suffix), nil
		}

		return types.NewStringValue(text[:maxLength-len(suffix)] + suffix), nil

	case "hashString":
		if len(args) < 1 {
			return nil, fmt.Errorf("hashString requires text parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeString {
			return nil, fmt.Errorf("text must be string")
		}
		text := args[0].(types.StringValue).Value()

		algorithm := "sha256"
		if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeString {
			algorithm = args[1].(types.StringValue).Value()
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
		return types.NewStringValue(hex.EncodeToString(h.Sum(nil))), nil

	case "sleep":
		if len(args) < 1 {
			return nil, fmt.Errorf("sleep requires milliseconds parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeNumber {
			return nil, fmt.Errorf("milliseconds must be number")
		}
		ms := args[0].(types.NumberValue).Value()

		time.Sleep(time.Duration(ms) * time.Millisecond)
		return types.NewNilValue(), nil

	case "formatDuration":
		if len(args) < 1 {
			return nil, fmt.Errorf("formatDuration requires milliseconds parameter")
		}
		if args[0] == nil || args[0].Type() != types.TypeNumber {
			return nil, fmt.Errorf("milliseconds must be number")
		}
		ms := args[0].(types.NumberValue).Value()

		d := time.Duration(ms) * time.Millisecond
		return types.NewStringValue(d.String()), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
