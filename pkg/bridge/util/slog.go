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
func (sb *SlogBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
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

// ExecuteMethod executes a bridge method
func (sb *SlogBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Check initialization first
	sb.mu.RLock()
	if !sb.initialized {
		sb.mu.RUnlock()
		return nil, fmt.Errorf("slog bridge not initialized")
	}
	sb.mu.RUnlock()

	// Methods that need write locks handle their own locking
	switch name {
	case "info":
		return sb.info(ctx, args)
	case "warn":
		return sb.warn(ctx, args)
	case "error":
		return sb.error(ctx, args)
	case "debug":
		return sb.debug(ctx, args)
	case "logBeforeGenerate":
		return sb.logBeforeGenerate(ctx, args)
	case "logAfterGenerate":
		return sb.logAfterGenerate(ctx, args)
	case "logBeforeToolCall":
		return sb.logBeforeToolCall(ctx, args)
	case "logAfterToolCall":
		return sb.logAfterToolCall(ctx, args)
	case "setLogLevel":
		return sb.setLogLevel(ctx, args)
	case "getLogLevel":
		return sb.getLogLevel(ctx, args)
	case "configureLogger":
		return sb.configureLogger(ctx, args)
	case "withAttributes":
		return sb.withAttributes(ctx, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// Bridge method implementations

// info logs an info message with optional emoji and attributes
func (sb *SlogBridge) info(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("info requires at least a message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	return sb.logWithLevel(ctx, slog.LevelInfo, message, args[1:])
}

// warn logs a warning message with optional emoji and attributes
func (sb *SlogBridge) warn(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("warn requires at least a message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	return sb.logWithLevel(ctx, slog.LevelWarn, message, args[1:])
}

// error logs an error message with optional emoji and attributes
func (sb *SlogBridge) error(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("error requires at least a message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	return sb.logWithLevel(ctx, slog.LevelError, message, args[1:])
}

// debug logs a debug message with optional emoji and attributes
func (sb *SlogBridge) debug(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("debug requires at least a message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	return sb.logWithLevel(ctx, slog.LevelDebug, message, args[1:])
}

// logWithLevel handles the actual logging with emoji and attributes
func (sb *SlogBridge) logWithLevel(ctx context.Context, level slog.Level, message string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	var emoji string
	var attrs []slog.Attr

	// Check for optional emoji (second argument)
	if len(args) > 0 && args[0] != nil && args[0].Type() == engine.TypeString {
		emoji = args[0].(engine.StringValue).Value()
		if emoji != "" {
			message = emoji + " " + message
		}
	}

	// Check for optional attributes (third argument)
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		objFields := args[1].(engine.ObjectValue).Fields()
		for k, v := range objFields {
			attrs = append(attrs, slog.Any(k, v.ToGo()))
		}
	}

	// Log with structured attributes
	sb.logger.LogAttrs(ctx, level, message, attrs...)
	return engine.NewNilValue(), nil
}

// logBeforeGenerate calls the logging hook's BeforeGenerate method
func (sb *SlogBridge) logBeforeGenerate(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("logBeforeGenerate requires messages array")
	}

	if args[0] == nil || args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("messages must be an array")
	}
	messagesArray := args[0].(engine.ArrayValue).Elements()

	// Convert script messages to go-llms Message format
	messages := make([]ldomain.Message, len(messagesArray))
	for i, msgVal := range messagesArray {
		if msgVal.Type() != engine.TypeObject {
			return nil, fmt.Errorf("message at index %d must be an object", i)
		}
		msgMap := msgVal.(engine.ObjectValue).Fields()

		roleVal, ok := msgMap["role"]
		if !ok || roleVal.Type() != engine.TypeString {
			return nil, fmt.Errorf("message role must be a string")
		}
		role := roleVal.(engine.StringValue).Value()

		contentVal, ok := msgMap["content"]
		if !ok || contentVal.Type() != engine.TypeString {
			return nil, fmt.Errorf("message content must be a string")
		}
		content := contentVal.(engine.StringValue).Value()

		messages[i] = ldomain.Message{
			Role:    ldomain.Role(role),
			Content: []ldomain.ContentPart{{Type: "text", Text: content}},
		}
	}

	// Call the logging hook
	sb.hook.BeforeGenerate(ctx, messages)
	return engine.NewNilValue(), nil
}

// logAfterGenerate calls the logging hook's AfterGenerate method
func (sb *SlogBridge) logAfterGenerate(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("logAfterGenerate requires response object")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("response must be an object")
	}
	responseMap := args[0].(engine.ObjectValue).Fields()

	// Convert response object to go-llms Response
	var response ldomain.Response
	if contentVal, ok := responseMap["content"]; ok && contentVal.Type() == engine.TypeString {
		response.Content = contentVal.(engine.StringValue).Value()
	}

	// Check for optional error
	var err error
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeString {
		errMsg := args[1].(engine.StringValue).Value()
		if errMsg != "" {
			err = fmt.Errorf("%s", errMsg)
		}
	}

	// Call the logging hook
	sb.hook.AfterGenerate(ctx, response, err)
	return engine.NewNilValue(), nil
}

// logBeforeToolCall calls the logging hook's BeforeToolCall method
func (sb *SlogBridge) logBeforeToolCall(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("logBeforeToolCall requires tool name and params")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("tool name must be a string")
	}
	toolName := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("params must be an object")
	}
	paramsMap := args[1].(engine.ObjectValue).Fields()

	// Convert params to native map
	params := make(map[string]interface{})
	for k, v := range paramsMap {
		params[k] = v.ToGo()
	}

	// Call the logging hook with tool name and params
	sb.hook.BeforeToolCall(ctx, toolName, params)
	return engine.NewNilValue(), nil
}

// logAfterToolCall calls the logging hook's AfterToolCall method
func (sb *SlogBridge) logAfterToolCall(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("logAfterToolCall requires at least tool name")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("tool name must be a string")
	}
	toolName := args[0].(engine.StringValue).Value()

	// Convert optional result
	var result interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		resultMap := args[1].(engine.ObjectValue).Fields()
		resultNative := make(map[string]interface{})
		for k, v := range resultMap {
			resultNative[k] = v.ToGo()
		}
		result = resultNative
	}

	// Check for optional error
	var err error
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeString {
		errMsg := args[2].(engine.StringValue).Value()
		if errMsg != "" {
			err = fmt.Errorf("%s", errMsg)
		}
	}

	// Call the logging hook with tool name, result and error
	sb.hook.AfterToolCall(ctx, toolName, result, err)
	return engine.NewNilValue(), nil
}

// setLogLevel sets the structured logging level
func (sb *SlogBridge) setLogLevel(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("setLogLevel requires level string")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("level must be a string")
	}
	levelStr := args[0].(engine.StringValue).Value()

	// Map string to core.LogLevel
	var level core.LogLevel
	switch levelStr {
	case "basic":
		level = core.LogLevelBasic
	case "detailed":
		level = core.LogLevelDetailed
	case "debug":
		level = core.LogLevelDebug
	default:
		return nil, fmt.Errorf("invalid log level: %s (must be 'basic', 'detailed', or 'debug')", levelStr)
	}

	sb.mu.Lock()
	sb.level = level
	// Update the logging hook with new level
	sb.hook = core.NewLoggingHook(sb.logger, level)
	sb.mu.Unlock()

	return engine.NewNilValue(), nil
}

// getLogLevel gets the current structured logging level
func (sb *SlogBridge) getLogLevel(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	sb.mu.RLock()
	level := sb.level
	sb.mu.RUnlock()

	// Convert level to string
	var levelStr string
	switch level {
	case core.LogLevelBasic:
		levelStr = "basic"
	case core.LogLevelDetailed:
		levelStr = "detailed"
	case core.LogLevelDebug:
		levelStr = "debug"
	default:
		levelStr = "unknown"
	}

	return engine.NewStringValue(levelStr), nil
}

// configureLogger configures the structured logger
func (sb *SlogBridge) configureLogger(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("configureLogger requires config object")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("config must be an object")
	}
	configMap := args[0].(engine.ObjectValue).Fields()

	// Extract configuration options
	var handlerOptions slog.HandlerOptions

	// Check for format
	var handler slog.Handler
	format := "text" // default
	if formatVal, ok := configMap["format"]; ok && formatVal.Type() == engine.TypeString {
		format = formatVal.(engine.StringValue).Value()
	}

	// Check for level
	if levelVal, ok := configMap["level"]; ok && levelVal.Type() == engine.TypeString {
		levelStr := levelVal.(engine.StringValue).Value()
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

	// Create handler based on format
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &handlerOptions)
	default:
		handler = slog.NewTextHandler(os.Stderr, &handlerOptions)
	}

	// Update logger
	sb.mu.Lock()
	sb.logger = slog.New(handler)
	// Recreate hook with new logger
	sb.hook = core.NewLoggingHook(sb.logger, sb.level)
	sb.mu.Unlock()

	return engine.NewNilValue(), nil
}

// withAttributes creates context with structured attributes
func (sb *SlogBridge) withAttributes(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("withAttributes requires attributes object")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}
	attrsMap := args[0].(engine.ObjectValue).Fields()

	// Convert attributes to slog.Attr slice
	var attrs []slog.Attr
	for k, v := range attrsMap {
		attrs = append(attrs, slog.Any(k, v.ToGo()))
	}

	// Create new logger with attributes - convert to any slice
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	contextLogger := sb.logger.With(anyAttrs...)

	// Return a context object that scripts can use
	return engine.NewObjectValue(map[string]engine.ScriptValue{
		"logger":     engine.NewCustomValue("logger", contextLogger),
		"attributes": args[0],
	}), nil
}
