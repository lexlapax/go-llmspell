// ABOUTME: Script-friendly unified logging interface combining debug and structured logging capabilities
// ABOUTME: Provides context propagation, bridge integration, and customizable output formatting for scripts

package util

import (
	"context"
	"fmt"
	"sync"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// ScriptLoggerBridge provides a unified script-friendly logging interface
type ScriptLoggerBridge struct {
	mu           sync.RWMutex
	initialized  bool
	debugBridge  *DebugBridge           // Debug logging functionality
	slogBridge   *SlogBridge            // Structured logging functionality
	contextAttrs map[string]interface{} // Global context attributes
	config       *LoggerConfig          // Logger configuration
}

// LoggerConfig holds configuration for the script logger
type LoggerConfig struct {
	DefaultLevel    string                 `json:"default_level"`    // Default log level
	EnableDebug     bool                   `json:"enable_debug"`     // Enable debug logging
	EnableStructure bool                   `json:"enable_structure"` // Enable structured logging
	Format          string                 `json:"format"`           // Output format: text, json
	Components      map[string]bool        `json:"components"`       // Debug component states
	Attributes      map[string]interface{} `json:"attributes"`       // Global attributes
	OutputTarget    string                 `json:"output_target"`    // Output target: stderr, stdout, file
}

// DefaultLoggerConfig returns default configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		DefaultLevel:    "info",
		EnableDebug:     true,
		EnableStructure: true,
		Format:          "text",
		Components:      make(map[string]bool),
		Attributes:      make(map[string]interface{}),
		OutputTarget:    "stderr",
	}
}

// NewScriptLoggerBridge creates a new unified script logger bridge
func NewScriptLoggerBridge() *ScriptLoggerBridge {
	return &ScriptLoggerBridge{
		debugBridge:  NewDebugBridge(),
		slogBridge:   NewSlogBridge(),
		contextAttrs: make(map[string]interface{}),
		config:       DefaultLoggerConfig(),
	}
}

// GetID returns the bridge identifier
func (sl *ScriptLoggerBridge) GetID() string {
	return "script_logger"
}

// GetMetadata returns bridge metadata
func (sl *ScriptLoggerBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "script_logger",
		Version:     "v1.0.0",
		Description: "Unified script-friendly logging interface combining debug and structured logging with context propagation",
		Author:      "go-llmspell",
		License:     "MIT",
		Dependencies: []string{
			"github.com/lexlapax/go-llmspell/pkg/bridge/util.DebugBridge",
			"github.com/lexlapax/go-llmspell/pkg/bridge/util.SlogBridge",
		},
	}
}

// Initialize sets up the script logger bridge
func (sl *ScriptLoggerBridge) Initialize(ctx context.Context) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Initialize sub-bridges
	if err := sl.debugBridge.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize debug bridge: %w", err)
	}

	if err := sl.slogBridge.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize slog bridge: %w", err)
	}

	sl.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (sl *ScriptLoggerBridge) Cleanup(ctx context.Context) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	var err error

	// Cleanup sub-bridges
	if debugErr := sl.debugBridge.Cleanup(ctx); debugErr != nil {
		err = fmt.Errorf("debug bridge cleanup failed: %w", debugErr)
	}

	if slogErr := sl.slogBridge.Cleanup(ctx); slogErr != nil {
		if err != nil {
			err = fmt.Errorf("%w; slog bridge cleanup failed: %w", err, slogErr)
		} else {
			err = fmt.Errorf("slog bridge cleanup failed: %w", slogErr)
		}
	}

	sl.contextAttrs = make(map[string]interface{})
	sl.initialized = false
	return err
}

// IsInitialized returns initialization status
func (sl *ScriptLoggerBridge) IsInitialized() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (sl *ScriptLoggerBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(sl)
}

// Methods returns available bridge methods
func (sl *ScriptLoggerBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Unified logging methods
		{
			Name:        "log",
			Description: "Unified logging method with level, component, and structured attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "level", Type: "string", Required: true, Description: "Log level: debug, info, warn, error"},
				{Name: "component", Type: "string", Required: false, Description: "Component name for debug logging"},
				{Name: "message", Type: "string", Required: true, Description: "Log message"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"log('info', 'agent', 'Processing request', {user: 'john', id: 123})"},
		},
		{
			Name:        "logWithContext",
			Description: "Log with additional context propagation",
			Parameters: []engine.ParameterInfo{
				{Name: "level", Type: "string", Required: true, Description: "Log level"},
				{Name: "message", Type: "string", Required: true, Description: "Log message"},
				{Name: "context", Type: "object", Required: false, Description: "Context object with attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"logWithContext('info', 'Task completed', {duration: 1500, success: true})"},
		},
		// Context management
		{
			Name:        "withContext",
			Description: "Create logging context with attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "attributes", Type: "object", Required: true, Description: "Context attributes"},
			},
			ReturnType: "object",
			Examples:   []string{"withContext({session_id: 'abc123', user: 'john'})"},
		},
		{
			Name:        "setGlobalContext",
			Description: "Set global context attributes for all logs",
			Parameters: []engine.ParameterInfo{
				{Name: "attributes", Type: "object", Required: true, Description: "Global attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"setGlobalContext({app: 'go-llmspell', version: '1.0'})"},
		},
		{
			Name:        "clearGlobalContext",
			Description: "Clear global context attributes",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
			Examples:    []string{"clearGlobalContext()"},
		},
		// Configuration methods
		{
			Name:        "configure",
			Description: "Configure the script logger with comprehensive settings",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Logger configuration"},
			},
			ReturnType: "void",
			Examples:   []string{"configure({format: 'json', level: 'debug', components: ['agent', 'tools']})"},
		},
		{
			Name:        "getConfiguration",
			Description: "Get current logger configuration",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getConfiguration()"},
		},
		// Component management (for debug logging)
		{
			Name:        "enableComponent",
			Description: "Enable debug logging for a component",
			Parameters: []engine.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"enableComponent('agent')"},
		},
		{
			Name:        "disableComponent",
			Description: "Disable debug logging for a component",
			Parameters: []engine.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"disableComponent('agent')"},
		},
		{
			Name:        "listEnabledComponents",
			Description: "List components with debug logging enabled",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listEnabledComponents()"},
		},
		// Convenience methods
		{
			Name:        "debug",
			Description: "Log debug message with optional component and attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Debug message"},
				{Name: "component", Type: "string", Required: false, Description: "Component name"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"debug('Variable state', 'workflow', {step: 1, value: 'test'})"},
		},
		{
			Name:        "info",
			Description: "Log info message with optional attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Info message"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"info('Process started', {pid: 1234, user: 'john'})"},
		},
		{
			Name:        "warn",
			Description: "Log warning message with optional attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Warning message"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"warn('Rate limit approaching', {remaining: 10, limit: 100})"},
		},
		{
			Name:        "error",
			Description: "Log error message with optional attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Error message"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"error('Connection failed', {host: 'api.example.com', timeout: 5000})"},
		},
		// Integration with bridge error handling
		{
			Name:        "logBridgeError",
			Description: "Log bridge-related errors with standardized format",
			Parameters: []engine.ParameterInfo{
				{Name: "bridgeId", Type: "string", Required: true, Description: "Bridge identifier"},
				{Name: "operation", Type: "string", Required: true, Description: "Operation that failed"},
				{Name: "error", Type: "string", Required: true, Description: "Error message"},
				{Name: "context", Type: "object", Required: false, Description: "Additional context"},
			},
			ReturnType: "void",
			Examples:   []string{"logBridgeError('llm', 'generate', 'API timeout', {duration: 5000})"},
		},
		// Log formatting and output
		{
			Name:        "formatMessage",
			Description: "Format log message with template and attributes",
			Parameters: []engine.ParameterInfo{
				{Name: "template", Type: "string", Required: true, Description: "Message template"},
				{Name: "attributes", Type: "object", Required: true, Description: "Template attributes"},
			},
			ReturnType: "string",
			Examples:   []string{"formatMessage('User {user} completed task {task}', {user: 'john', task: 'upload'})"},
		},
	}
}

// ValidateMethod validates method calls
func (sl *ScriptLoggerBridge) ValidateMethod(name string, args []interface{}) error {
	if !sl.IsInitialized() {
		return fmt.Errorf("script logger bridge not initialized")
	}

	methods := sl.Methods()
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
func (sl *ScriptLoggerBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"logger_config": {
			GoType:     "*LoggerConfig",
			ScriptType: "object",
			Converter:  "loggerConfigConverter",
			Metadata:   map[string]interface{}{"description": "Logger configuration object"},
		},
		"log_context": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
			Converter:  "logContextConverter",
			Metadata:   map[string]interface{}{"description": "Logging context with attributes"},
		},
		"unified_logger": {
			GoType:     "*ScriptLoggerBridge",
			ScriptType: "object",
			Converter:  "unifiedLoggerConverter",
			Metadata:   map[string]interface{}{"description": "Unified script logger instance"},
		},
	}
}

// RequiredPermissions returns required permissions
func (sl *ScriptLoggerBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionStorage,
			Resource:    "script_logger.config",
			Actions:     []string{"read", "write"},
			Description: "Access script logger configuration",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "script_logger.context",
			Actions:     []string{"read", "write"},
			Description: "Manage logging context and global attributes",
		},
		{
			Type:        engine.PermissionStorage,
			Resource:    "script_logger.output",
			Actions:     []string{"write"},
			Description: "Write log output to configured targets",
		},
	}
}

// Bridge method implementations

// log is the unified logging method
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) log(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("log", args); err != nil {
		return err
	}

	level, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("level must be a string")
	}

	// Parse arguments based on structure:
	// log(level, message) - 2 args
	// log(level, message, attributes) - 3 args
	// log(level, component, message) - 3 args where component is string and message is string
	// log(level, component, message, attributes) - 4 args

	var component string
	var message string
	var attributes map[string]interface{}

	if len(args) == 2 {
		// log(level, message)
		if msg, ok := args[1].(string); ok {
			message = msg
		} else {
			return fmt.Errorf("message must be a string")
		}
	} else if len(args) == 3 {
		// Could be log(level, message, attributes) or log(level, component, message)
		// Try to detect: if args[2] is a string, then args[1] is component and args[2] is message
		// If args[2] is not a string, then args[1] is message and args[2] is attributes
		if msg, ok := args[2].(string); ok {
			// log(level, component, message)
			if comp, ok := args[1].(string); ok {
				component = comp
				message = msg
			} else {
				return fmt.Errorf("component must be a string")
			}
		} else {
			// log(level, message, attributes)
			if msg, ok := args[1].(string); ok {
				message = msg
				if attrs, ok := args[2].(map[string]interface{}); ok {
					attributes = attrs
				}
			} else {
				return fmt.Errorf("message must be a string")
			}
		}
	} else if len(args) >= 4 {
		// log(level, component, message, attributes)
		if comp, ok := args[1].(string); ok {
			component = comp
		} else {
			return fmt.Errorf("component must be a string")
		}

		if msg, ok := args[2].(string); ok {
			message = msg
		} else {
			return fmt.Errorf("message must be a string")
		}

		if len(args) > 3 && args[3] != nil {
			if attrs, ok := args[3].(map[string]interface{}); ok {
				attributes = attrs
			}
		}
	} else {
		return fmt.Errorf("insufficient arguments for log method")
	}

	return sl.logWithLevelAndContext(ctx, level, component, message, attributes)
}

// logWithContext logs with additional context propagation
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) logWithContext(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("logWithContext", args); err != nil {
		return err
	}

	level, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("level must be a string")
	}

	message, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	var contextAttrs map[string]interface{}
	if len(args) > 2 && args[2] != nil {
		if attrs, ok := args[2].(map[string]interface{}); ok {
			contextAttrs = attrs
		}
	}

	// Merge with global context
	mergedAttrs := sl.mergeWithGlobalContext(contextAttrs)

	return sl.logWithLevelAndContext(ctx, level, "", message, mergedAttrs)
}

// withContext creates a logging context with attributes
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) withContext(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := sl.ValidateMethod("withContext", args); err != nil {
		return nil, err
	}

	attributes, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("attributes must be an object")
	}

	// Merge with global context
	mergedAttrs := sl.mergeWithGlobalContext(attributes)

	return map[string]interface{}{
		"type":       "log_context",
		"attributes": mergedAttrs,
		"timestamp":  ctx.Value("timestamp"),
	}, nil
}

// setGlobalContext sets global context attributes
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) setGlobalContext(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("setGlobalContext", args); err != nil {
		return err
	}

	attributes, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("attributes must be an object")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Merge with existing global context
	for key, value := range attributes {
		sl.contextAttrs[key] = value
	}

	return nil
}

// clearGlobalContext clears global context attributes
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) clearGlobalContext(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("clearGlobalContext", args); err != nil {
		return err
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.contextAttrs = make(map[string]interface{})
	return nil
}

// configure configures the script logger
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) configure(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("configure", args); err != nil {
		return err
	}

	configMap, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("config must be an object")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Update configuration
	if level, ok := configMap["level"].(string); ok {
		sl.config.DefaultLevel = level
	}

	if format, ok := configMap["format"].(string); ok {
		sl.config.Format = format
		// Reconfigure slog bridge
		if err := sl.slogBridge.configureLogger(ctx, []interface{}{
			map[string]interface{}{"format": format},
		}); err != nil {
			return fmt.Errorf("failed to configure slog bridge: %w", err)
		}
	}

	if enableDebug, ok := configMap["enable_debug"].(bool); ok {
		sl.config.EnableDebug = enableDebug
	}

	if enableStructure, ok := configMap["enable_structure"].(bool); ok {
		sl.config.EnableStructure = enableStructure
	}

	if components, ok := configMap["components"].([]interface{}); ok {
		// Enable specified components
		for _, comp := range components {
			if compStr, ok := comp.(string); ok {
				if err := sl.debugBridge.enableDebugComponent(ctx, []interface{}{compStr}); err != nil {
					return fmt.Errorf("failed to enable debug component %s: %w", compStr, err)
				}
			}
		}
	}

	if attributes, ok := configMap["attributes"].(map[string]interface{}); ok {
		sl.config.Attributes = attributes
		// Merge with global context
		for key, value := range attributes {
			sl.contextAttrs[key] = value
		}
	}

	return nil
}

// getConfiguration gets current logger configuration
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) getConfiguration(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := sl.ValidateMethod("getConfiguration", args); err != nil {
		return nil, err
	}

	sl.mu.RLock()
	defer sl.mu.RUnlock()

	return map[string]interface{}{
		"default_level":    sl.config.DefaultLevel,
		"enable_debug":     sl.config.EnableDebug,
		"enable_structure": sl.config.EnableStructure,
		"format":           sl.config.Format,
		"components":       sl.config.Components,
		"attributes":       sl.config.Attributes,
		"output_target":    sl.config.OutputTarget,
		"global_context":   sl.contextAttrs,
	}, nil
}

// enableComponent enables debug logging for a component
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) enableComponent(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("enableComponent", args); err != nil {
		return err
	}

	return sl.debugBridge.enableDebugComponent(ctx, args)
}

// disableComponent disables debug logging for a component
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) disableComponent(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("disableComponent", args); err != nil {
		return err
	}

	return sl.debugBridge.disableDebugComponent(ctx, args)
}

// listEnabledComponents lists components with debug logging enabled
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) listEnabledComponents(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := sl.ValidateMethod("listEnabledComponents", args); err != nil {
		return nil, err
	}

	return sl.debugBridge.listEnabledComponents(ctx, args)
}

// Convenience logging methods

// debug logs a debug message
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) debug(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("debug", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	var component string
	var attributes map[string]interface{}

	if len(args) > 1 && args[1] != nil {
		if comp, ok := args[1].(string); ok {
			component = comp
		}
	}

	if len(args) > 2 && args[2] != nil {
		if attrs, ok := args[2].(map[string]interface{}); ok {
			attributes = attrs
		}
	}

	return sl.logWithLevelAndContext(ctx, "debug", component, message, attributes)
}

// info logs an info message
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) info(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("info", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil {
		if attrs, ok := args[1].(map[string]interface{}); ok {
			attributes = attrs
		}
	}

	return sl.logWithLevelAndContext(ctx, "info", "", message, attributes)
}

// warn logs a warning message
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) warn(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("warn", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil {
		if attrs, ok := args[1].(map[string]interface{}); ok {
			attributes = attrs
		}
	}

	return sl.logWithLevelAndContext(ctx, "warn", "", message, attributes)
}

// error logs an error message
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) error(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("error", args); err != nil {
		return err
	}

	message, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("message must be a string")
	}

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil {
		if attrs, ok := args[1].(map[string]interface{}); ok {
			attributes = attrs
		}
	}

	return sl.logWithLevelAndContext(ctx, "error", "", message, attributes)
}

// logBridgeError logs bridge-related errors
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) logBridgeError(ctx context.Context, args []interface{}) error {
	if err := sl.ValidateMethod("logBridgeError", args); err != nil {
		return err
	}

	bridgeId, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("bridgeId must be a string")
	}

	operation, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("operation must be a string")
	}

	errorMsg, ok := args[2].(string)
	if !ok {
		return fmt.Errorf("error must be a string")
	}

	// Standard bridge error attributes
	attrs := map[string]interface{}{
		"bridge_id": bridgeId,
		"operation": operation,
		"error":     errorMsg,
		"category":  "bridge_error",
	}

	// Add additional context if provided
	if len(args) > 3 && args[3] != nil {
		if contextAttrs, ok := args[3].(map[string]interface{}); ok {
			for key, value := range contextAttrs {
				attrs[key] = value
			}
		}
	}

	message := fmt.Sprintf("Bridge error in %s.%s: %s", bridgeId, operation, errorMsg)
	return sl.logWithLevelAndContext(ctx, "error", "bridge", message, attrs)
}

// formatMessage formats a log message with template and attributes
//
//nolint:unused // Bridge method called via reflection
func (sl *ScriptLoggerBridge) formatMessage(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := sl.ValidateMethod("formatMessage", args); err != nil {
		return nil, err
	}

	template, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("template must be a string")
	}

	attributes, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("attributes must be an object")
	}

	// Simple template replacement
	result := template
	for key, value := range attributes {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = fmt.Sprintf(result, placeholder, replacement)
	}

	return result, nil
}

// Helper methods

// logWithLevelAndContext performs unified logging with level, component, and context
func (sl *ScriptLoggerBridge) logWithLevelAndContext(ctx context.Context, level, component, message string, attributes map[string]interface{}) error {
	sl.mu.RLock()
	config := sl.config
	sl.mu.RUnlock()

	// Merge with global context
	mergedAttrs := sl.mergeWithGlobalContext(attributes)

	// Debug logging if enabled and component provided
	if config.EnableDebug && component != "" && (level == "debug" || level == "info") {
		if level == "debug" {
			if err := sl.debugBridge.debugPrintln(ctx, []interface{}{component, message}); err != nil {
				// Log error but don't fail the entire operation
				fmt.Printf("Warning: debug logging failed: %v\n", err)
			}
		} else {
			if err := sl.debugBridge.debugPrintf(ctx, []interface{}{component, "%s", []interface{}{message}}); err != nil {
				// Log error but don't fail the entire operation
				fmt.Printf("Warning: debug logging failed: %v\n", err)
			}
		}
	}

	// Structured logging if enabled
	if config.EnableStructure {
		var slogArgs []interface{}
		switch level {
		case "debug":
			slogArgs = []interface{}{message, "🐛", mergedAttrs}
			return sl.slogBridge.debug(ctx, slogArgs)
		case "info":
			slogArgs = []interface{}{message, "ℹ️", mergedAttrs}
			return sl.slogBridge.info(ctx, slogArgs)
		case "warn":
			slogArgs = []interface{}{message, "⚠️", mergedAttrs}
			return sl.slogBridge.warn(ctx, slogArgs)
		case "error":
			slogArgs = []interface{}{message, "❌", mergedAttrs}
			return sl.slogBridge.error(ctx, slogArgs)
		default:
			// Default to info level
			slogArgs = []interface{}{message, "", mergedAttrs}
			return sl.slogBridge.info(ctx, slogArgs)
		}
	}

	return nil
}

// mergeWithGlobalContext merges attributes with global context
func (sl *ScriptLoggerBridge) mergeWithGlobalContext(attributes map[string]interface{}) map[string]interface{} {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	merged := make(map[string]interface{})

	// Add global context first
	for key, value := range sl.contextAttrs {
		merged[key] = value
	}

	// Override with specific attributes
	for key, value := range attributes {
		merged[key] = value
	}

	return merged
}
