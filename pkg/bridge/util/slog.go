// ABOUTME: Structured logging bridge providing access to slog integration from go-llms
// ABOUTME: Bridges LoggingHook functionality with emoji enhancement and structured key-value logging

package util

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for logging hook
	"github.com/lexlapax/go-llms/pkg/agent/core"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// SlogBridge provides script access to go-llms structured logging system
type SlogBridge struct {
	mu          sync.RWMutex
	initialized bool
	logger      *slog.Logger      // Structured logger instance
	hook        *core.LoggingHook // go-llms logging hook
	level       core.LogLevel     // Current log level
}

// NewSlogBridge creates a new structured logging bridge
func NewSlogBridge() *SlogBridge {
	// Initialize with default slog logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	bridge := &SlogBridge{
		logger: logger,
		level:  core.LogLevelBasic, // Default to basic level
	}

	// Create logging hook with the logger
	bridge.hook = core.NewLoggingHook(logger, bridge.level)

	return bridge
}

// GetID returns the bridge identifier
func (sb *SlogBridge) GetID() string {
	return "slog"
}

// GetMetadata returns bridge metadata
func (sb *SlogBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "slog",
		Version:      "v2.0.0",
		Description:  "Bridge for go-llms structured logging with slog integration, emoji enhancement, and key-value logging",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"log/slog", "github.com/lexlapax/go-llms/pkg/agent/core"},
	}
}

// Initialize sets up the structured logging bridge
func (sb *SlogBridge) Initialize(ctx context.Context) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (sb *SlogBridge) Cleanup(ctx context.Context) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.initialized = false
	return nil
}

// IsInitialized returns initialization status
func (sb *SlogBridge) IsInitialized() bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (sb *SlogBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(sb)
}

// Methods returns available bridge methods
func (sb *SlogBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Basic logging methods
		{
			Name:        "info",
			Description: "Log info message with optional key-value pairs",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Log message"},
				{Name: "emoji", Type: "string", Required: false, Description: "Emoji to include"},
				{Name: "attributes", Type: "object", Required: false, Description: "Key-value attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"info('Processing request', '🤔', {user: 'john', request_id: '123'})"},
		},
		{
			Name:        "warn",
			Description: "Log warning message with optional key-value pairs",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Warning message"},
				{Name: "emoji", Type: "string", Required: false, Description: "Emoji to include"},
				{Name: "attributes", Type: "object", Required: false, Description: "Key-value attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"warn('Rate limit approaching', '⚠️', {remaining: 10})"},
		},
		{
			Name:        "error",
			Description: "Log error message with optional key-value pairs",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Error message"},
				{Name: "emoji", Type: "string", Required: false, Description: "Emoji to include"},
				{Name: "attributes", Type: "object", Required: false, Description: "Key-value attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"error('API call failed', '❌', {error: 'timeout', duration: 5000})"},
		},
		{
			Name:        "debug",
			Description: "Log debug message with optional key-value pairs",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Debug message"},
				{Name: "emoji", Type: "string", Required: false, Description: "Emoji to include"},
				{Name: "attributes", Type: "object", Required: false, Description: "Key-value attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"debug('Variable state', '🐛', {var1: 'value', count: 42})"},
		},
		// Logging hook methods
		{
			Name:        "logBeforeGenerate",
			Description: "Log before LLM generation using logging hook",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Required: true, Description: "Array of messages"},
			},
			ReturnType: "void",
			Examples:   []string{"logBeforeGenerate([{role: 'user', content: 'Hello'}])"},
		},
		{
			Name:        "logAfterGenerate",
			Description: "Log after LLM generation using logging hook",
			Parameters: []engine.ParameterInfo{
				{Name: "response", Type: "object", Required: true, Description: "LLM response object"},
				{Name: "error", Type: "string", Required: false, Description: "Error message if any"},
			},
			ReturnType: "void",
			Examples:   []string{"logAfterGenerate({content: 'Generated text'}, null)"},
		},
		{
			Name:        "logBeforeToolCall",
			Description: "Log before tool execution using logging hook",
			Parameters: []engine.ParameterInfo{
				{Name: "tool", Type: "string", Required: true, Description: "Tool name"},
				{Name: "params", Type: "object", Required: true, Description: "Tool parameters"},
			},
			ReturnType: "void",
			Examples:   []string{"logBeforeToolCall('web_search', {query: 'golang tutorial'})"},
		},
		{
			Name:        "logAfterToolCall",
			Description: "Log after tool execution using logging hook",
			Parameters: []engine.ParameterInfo{
				{Name: "tool", Type: "string", Required: true, Description: "Tool name"},
				{Name: "result", Type: "object", Required: false, Description: "Tool result"},
				{Name: "error", Type: "string", Required: false, Description: "Error message if any"},
			},
			ReturnType: "void",
			Examples:   []string{"logAfterToolCall('web_search', {results: []}, null)"},
		},
		// Log level configuration
		{
			Name:        "setLogLevel",
			Description: "Set the structured logging level",
			Parameters: []engine.ParameterInfo{
				{Name: "level", Type: "string", Required: true, Description: "Log level: 'basic', 'detailed', 'debug'"},
			},
			ReturnType: "void",
			Examples:   []string{"setLogLevel('detailed')"},
		},
		{
			Name:        "getLogLevel",
			Description: "Get the current structured logging level",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "string",
			Examples:    []string{"getLogLevel()"},
		},
		// Logger configuration
		{
			Name:        "configureLogger",
			Description: "Configure the structured logger",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Logger configuration"},
			},
			ReturnType: "void",
			Examples:   []string{"configureLogger({format: 'json', level: 'debug'})"},
		},
		{
			Name:        "withAttributes",
			Description: "Create context with structured attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "attributes", Type: "object", Required: true, Description: "Key-value attributes"},
			},
			ReturnType: "object",
			Examples:   []string{"withAttributes({component: 'agent', session_id: 'abc123'})"},
		},
	}
}

// ValidateMethod validates method calls
func (sb *SlogBridge) ValidateMethod(name string, args []interface{}) error {
	if !sb.IsInitialized() {
		return fmt.Errorf("slog bridge not initialized")
	}

	methods := sb.Methods()
	for _, method := range methods {
		if method.Name == name {
			requiredCount := 0
			for _, param := range method.Parameters {
				if param.Required {
					requiredCount++
				}
			}
			if len(args) < requiredCount {
				return fmt.Errorf("method %s requires at least %d arguments, got %d", name, requiredCount, len(args))
			}
			return nil
		}
	}
	return fmt.Errorf("unknown method: %s", name)
}

// TypeMappings returns type conversion mappings
func (sb *SlogBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"slog_logger": {
			GoType:     "*slog.Logger",
			ScriptType: "object",
			Converter:  "slogLoggerConverter",
			Metadata:   map[string]interface{}{"description": "Structured logger instance"},
		},
		"log_level": {
			GoType:     "core.LogLevel",
			ScriptType: "string",
			Converter:  "logLevelConverter",
			Metadata:   map[string]interface{}{"description": "Logging level enumeration"},
		},
		"logging_hook": {
			GoType:     "*core.LoggingHook",
			ScriptType: "object",
			Converter:  "loggingHookConverter",
			Metadata:   map[string]interface{}{"description": "go-llms logging hook instance"},
		},
		"message_array": {
			GoType:     "[]ldomain.Message",
			ScriptType: "array",
			Converter:  "messageArrayConverter",
			Metadata:   map[string]interface{}{"description": "Array of LLM messages"},
		},
		"llm_response": {
			GoType:     "ldomain.Response",
			ScriptType: "object",
			Converter:  "llmResponseConverter",
			Metadata:   map[string]interface{}{"description": "LLM response object"},
		},
	}
}

// RequiredPermissions returns required permissions
func (sb *SlogBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionStorage,
			Resource:    "slog.logging",
			Actions:     []string{"read", "write"},
			Description: "Access structured logging configuration",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "slog.context",
			Actions:     []string{"read", "write"},
			Description: "Manage logging context and attributes",
		},
	}
}

// Bridge method implementations

// info logs an info message with optional emoji and attributes
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) info(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("info", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	return sb.logWithLevel(ctx, slog.LevelInfo, message, args[1:])
}

// warn logs a warning message with optional emoji and attributes
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) warn(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("warn", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	return sb.logWithLevel(ctx, slog.LevelWarn, message, args[1:])
}

// error logs an error message with optional emoji and attributes
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) error(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("error", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	return sb.logWithLevel(ctx, slog.LevelError, message, args[1:])
}

// debug logs a debug message with optional emoji and attributes
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) debug(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("debug", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	return sb.logWithLevel(ctx, slog.LevelDebug, message, args[1:])
}

// logBeforeGenerate calls the logging hook's BeforeGenerate method
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) logBeforeGenerate(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("logBeforeGenerate", args); err != nil {
		return err
	}

	messagesArg, ok := args[0].([]interface{})
	if !ok {
		return fmt.Errorf("messages must be an array")
	}

	// Convert script messages to go-llms Message format
	messages := make([]ldomain.Message, len(messagesArg))
	for i, msgInterface := range messagesArg {
		msgMap, ok := msgInterface.(map[string]interface{})
		if !ok {
			return fmt.Errorf("message at index %d must be an object", i)
		}

		role, ok := msgMap["role"].(string)
		if !ok {
			return fmt.Errorf("message role must be a string")
		}

		content, ok := msgMap["content"].(string)
		if !ok {
			return fmt.Errorf("message content must be a string")
		}

		messages[i] = ldomain.Message{
			Role: ldomain.Role(role),
			Content: []ldomain.ContentPart{
				{
					Type: ldomain.ContentTypeText,
					Text: content,
				},
			},
		}
	}

	sb.mu.RLock()
	hook := sb.hook
	sb.mu.RUnlock()

	hook.BeforeGenerate(ctx, messages)
	return nil
}

// logAfterGenerate calls the logging hook's AfterGenerate method
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) logAfterGenerate(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("logAfterGenerate", args); err != nil {
		return err
	}

	responseArg, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("response must be an object")
	}

	content, ok := responseArg["content"].(string)
	if !ok {
		return fmt.Errorf("response content must be a string")
	}

	response := ldomain.Response{
		Content: content,
	}

	var err error
	if len(args) > 1 && args[1] != nil {
		if errStr, ok := args[1].(string); ok {
			err = fmt.Errorf("%s", errStr)
		}
	}

	sb.mu.RLock()
	hook := sb.hook
	sb.mu.RUnlock()

	hook.AfterGenerate(ctx, response, err)
	return nil
}

// logBeforeToolCall calls the logging hook's BeforeToolCall method
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) logBeforeToolCall(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("logBeforeToolCall", args); err != nil {
		return err
	}

	tool, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("tool must be a string")
	}

	params, ok := args[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("params must be an object")
	}

	sb.mu.RLock()
	hook := sb.hook
	sb.mu.RUnlock()

	hook.BeforeToolCall(ctx, tool, params)
	return nil
}

// logAfterToolCall calls the logging hook's AfterToolCall method
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) logAfterToolCall(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("logAfterToolCall", args); err != nil {
		return err
	}

	tool, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("tool must be a string")
	}

	var result interface{}
	var err error

	if len(args) > 1 && args[1] != nil {
		result = args[1]
	}

	if len(args) > 2 && args[2] != nil {
		if errStr, ok := args[2].(string); ok {
			err = fmt.Errorf("%s", errStr)
		}
	}

	sb.mu.RLock()
	hook := sb.hook
	sb.mu.RUnlock()

	hook.AfterToolCall(ctx, tool, result, err)
	return nil
}

// setLogLevel sets the structured logging level
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) setLogLevel(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("setLogLevel", args); err != nil {
		return err
	}

	levelStr, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("level must be a string")
	}

	var level core.LogLevel
	switch levelStr {
	case "basic":
		level = core.LogLevelBasic
	case "detailed":
		level = core.LogLevelDetailed
	case "debug":
		level = core.LogLevelDebug
	default:
		return fmt.Errorf("invalid log level: %s (must be 'basic', 'detailed', or 'debug')", levelStr)
	}

	sb.mu.Lock()
	sb.level = level
	// Recreate the hook with the new level
	sb.hook = core.NewLoggingHook(sb.logger, level)
	sb.mu.Unlock()

	return nil
}

// getLogLevel gets the current structured logging level
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) getLogLevel(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := sb.ValidateMethod("getLogLevel", args); err != nil {
		return nil, err
	}

	sb.mu.RLock()
	level := sb.level
	sb.mu.RUnlock()

	switch level {
	case core.LogLevelBasic:
		return "basic", nil
	case core.LogLevelDetailed:
		return "detailed", nil
	case core.LogLevelDebug:
		return "debug", nil
	default:
		return "unknown", nil
	}
}

// configureLogger configures the structured logger
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) configureLogger(ctx context.Context, args []interface{}) error {
	if err := sb.ValidateMethod("configureLogger", args); err != nil {
		return err
	}

	config, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("config must be an object")
	}

	var handlerOptions slog.HandlerOptions

	// Configure log level
	if levelStr, ok := config["level"].(string); ok {
		switch levelStr {
		case "debug":
			handlerOptions.Level = slog.LevelDebug
		case "info":
			handlerOptions.Level = slog.LevelInfo
		case "warn":
			handlerOptions.Level = slog.LevelWarn
		case "error":
			handlerOptions.Level = slog.LevelError
		}
	}

	// Create new handler based on format
	var handler slog.Handler
	if format, ok := config["format"].(string); ok && format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, &handlerOptions)
	} else {
		handler = slog.NewTextHandler(os.Stderr, &handlerOptions)
	}

	sb.mu.Lock()
	sb.logger = slog.New(handler)
	// Recreate the hook with the new logger
	sb.hook = core.NewLoggingHook(sb.logger, sb.level)
	sb.mu.Unlock()

	return nil
}

// withAttributes creates a context with structured attributes
//
//nolint:unused // Bridge method called via reflection
func (sb *SlogBridge) withAttributes(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := sb.ValidateMethod("withAttributes", args); err != nil {
		return nil, err
	}

	attributes, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("attributes must be an object")
	}

	// Convert attributes to slog attributes
	slogAttrs := make([]slog.Attr, 0, len(attributes))
	for key, value := range attributes {
		slogAttrs = append(slogAttrs, slog.Any(key, value))
	}

	// Return context representation that can be used in scripts
	return map[string]interface{}{
		"type":       "logging_context",
		"attributes": attributes,
		"slog_attrs": slogAttrs,
	}, nil
}

// Helper methods

// logWithLevel logs a message at the specified level with optional emoji and attributes
func (sb *SlogBridge) logWithLevel(ctx context.Context, level slog.Level, message string, extraArgs []interface{}) error {
	sb.mu.RLock()
	logger := sb.logger
	sb.mu.RUnlock()

	// Start with the message
	args := []interface{}{message}

	// Add emoji if provided
	if len(extraArgs) > 0 && extraArgs[0] != nil {
		if emoji, ok := extraArgs[0].(string); ok {
			args = append(args, "emoji", emoji)
		}
	}

	// Add attributes if provided
	if len(extraArgs) > 1 && extraArgs[1] != nil {
		if attributes, ok := extraArgs[1].(map[string]interface{}); ok {
			for key, value := range attributes {
				args = append(args, key, value)
			}
		}
	}

	logger.Log(ctx, level, args[0].(string), args[1:]...)
	return nil
}
