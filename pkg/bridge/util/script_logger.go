// ABOUTME: Script-friendly unified logging interface combining debug and structured logging capabilities
// ABOUTME: Provides context propagation, bridge integration, and customizable output formatting for scripts

package util

import (
	"context"
	"fmt"
	"strings"
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
func (sl *ScriptLoggerBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
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

// ExecuteMethod executes a bridge method
func (sl *ScriptLoggerBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	if !sl.initialized {
		return nil, fmt.Errorf("script logger bridge not initialized")
	}

	switch name {
	case "log":
		return sl.log(ctx, args)
	case "logWithContext":
		return sl.logWithContext(ctx, args)
	case "withContext":
		return sl.withContext(ctx, args)
	case "setGlobalContext":
		return sl.setGlobalContext(ctx, args)
	case "clearGlobalContext":
		return sl.clearGlobalContext(ctx, args)
	case "configure":
		return sl.configure(ctx, args)
	case "getConfiguration":
		return sl.getConfiguration(ctx, args)
	case "enableComponent":
		return sl.enableComponent(ctx, args)
	case "disableComponent":
		return sl.disableComponent(ctx, args)
	case "listEnabledComponents":
		return sl.listEnabledComponents(ctx, args)
	case "debug":
		return sl.debug(ctx, args)
	case "info":
		return sl.info(ctx, args)
	case "warn":
		return sl.warn(ctx, args)
	case "error":
		return sl.error(ctx, args)
	case "logBridgeError":
		return sl.logBridgeError(ctx, args)
	case "formatMessage":
		return sl.formatMessage(ctx, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// Bridge method implementations

// log is the unified logging method
func (sl *ScriptLoggerBridge) log(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("log requires at least level and message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("level must be a string")
	}
	level := args[0].(engine.StringValue).Value()

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
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("message must be a string")
		}
		message = args[1].(engine.StringValue).Value()
	} else if len(args) == 3 {
		// Could be log(level, message, attributes) or log(level, component, message)
		// Try to detect: if args[2] is a string, then args[1] is component and args[2] is message
		// If args[2] is not a string, then args[1] is message and args[2] is attributes
		if args[2] != nil && args[2].Type() == engine.TypeString {
			// log(level, component, message)
			if args[1] == nil || args[1].Type() != engine.TypeString {
				return nil, fmt.Errorf("component must be a string")
			}
			component = args[1].(engine.StringValue).Value()
			message = args[2].(engine.StringValue).Value()
		} else {
			// log(level, message, attributes)
			if args[1] == nil || args[1].Type() != engine.TypeString {
				return nil, fmt.Errorf("message must be a string")
			}
			message = args[1].(engine.StringValue).Value()
			if args[2] != nil && args[2].Type() == engine.TypeObject {
				objFields := args[2].(engine.ObjectValue).Fields()
				attributes = make(map[string]interface{})
				for k, v := range objFields {
					attributes[k] = v.ToGo()
				}
			}
		}
	} else if len(args) >= 4 {
		// log(level, component, message, attributes)
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("component must be a string")
		}
		component = args[1].(engine.StringValue).Value()

		if args[2] == nil || args[2].Type() != engine.TypeString {
			return nil, fmt.Errorf("message must be a string")
		}
		message = args[2].(engine.StringValue).Value()

		if args[3] != nil && args[3].Type() == engine.TypeObject {
			objFields := args[3].(engine.ObjectValue).Fields()
			attributes = make(map[string]interface{})
			for k, v := range objFields {
				attributes[k] = v.ToGo()
			}
		}
	}

	return sl.logWithLevelAndContext(ctx, level, component, message, attributes)
}

// logWithContext logs with additional context propagation
func (sl *ScriptLoggerBridge) logWithContext(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("logWithContext requires level and message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("level must be a string")
	}
	level := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[1].(engine.StringValue).Value()

	var context map[string]interface{}
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
		objFields := args[2].(engine.ObjectValue).Fields()
		context = make(map[string]interface{})
		for k, v := range objFields {
			context[k] = v.ToGo()
		}
	}

	// Merge context with global attributes
	mergedAttrs := sl.mergeAttributes(context)
	return sl.logWithLevelAndContext(ctx, level, "", message, mergedAttrs)
}

// withContext creates logging context with attributes
func (sl *ScriptLoggerBridge) withContext(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("withContext requires attributes")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}

	// Create context object
	contextObj := map[string]engine.ScriptValue{
		"attributes": args[0],
		"logger":     engine.NewCustomValue("logger", sl),
	}

	// If using slog, also create a slog context
	if sl.config.EnableStructure {
		result, err := sl.slogBridge.ExecuteMethod(ctx, "withAttributes", args)
		if err == nil && result != nil {
			contextObj["slogContext"] = result
		}
	}

	return engine.NewObjectValue(contextObj), nil
}

// setGlobalContext sets global context attributes
func (sl *ScriptLoggerBridge) setGlobalContext(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("setGlobalContext requires attributes")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Convert to native map
	objFields := args[0].(engine.ObjectValue).Fields()
	sl.contextAttrs = make(map[string]interface{})
	for k, v := range objFields {
		sl.contextAttrs[k] = v.ToGo()
	}

	return engine.NewNilValue(), nil
}

// clearGlobalContext clears global context attributes
func (sl *ScriptLoggerBridge) clearGlobalContext(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.contextAttrs = make(map[string]interface{})
	return engine.NewNilValue(), nil
}

// configure configures the script logger
func (sl *ScriptLoggerBridge) configure(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("configure requires config object")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("config must be an object")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	configObj := args[0].(engine.ObjectValue).Fields()

	// Update configuration
	if levelVal, ok := configObj["level"]; ok && levelVal.Type() == engine.TypeString {
		sl.config.DefaultLevel = levelVal.(engine.StringValue).Value()
	}

	if formatVal, ok := configObj["format"]; ok && formatVal.Type() == engine.TypeString {
		sl.config.Format = formatVal.(engine.StringValue).Value()
		// Configure slog with new format
		_, _ = sl.slogBridge.ExecuteMethod(ctx, "configureLogger", []engine.ScriptValue{
			engine.NewObjectValue(map[string]engine.ScriptValue{
				"format": formatVal,
			}),
		})
	}

	if debugVal, ok := configObj["enable_debug"]; ok && debugVal.Type() == engine.TypeBool {
		sl.config.EnableDebug = debugVal.(engine.BoolValue).Value()
	}

	if structureVal, ok := configObj["enable_structure"]; ok && structureVal.Type() == engine.TypeBool {
		sl.config.EnableStructure = structureVal.(engine.BoolValue).Value()
	}

	if componentsVal, ok := configObj["components"]; ok && componentsVal.Type() == engine.TypeArray {
		components := componentsVal.(engine.ArrayValue).Elements()
		sl.config.Components = make(map[string]bool)
		for _, comp := range components {
			if comp.Type() == engine.TypeString {
				compName := comp.(engine.StringValue).Value()
				sl.config.Components[compName] = true
				// Enable in debug bridge
				_, _ = sl.debugBridge.ExecuteMethod(ctx, "enableDebugComponent", []engine.ScriptValue{
					engine.NewStringValue(compName),
				})
			}
		}
	}

	return engine.NewNilValue(), nil
}

// getConfiguration gets current logger configuration
func (sl *ScriptLoggerBridge) getConfiguration(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// Convert components map to array
	components := make([]engine.ScriptValue, 0, len(sl.config.Components))
	for comp, enabled := range sl.config.Components {
		if enabled {
			components = append(components, engine.NewStringValue(comp))
		}
	}

	// Convert attributes to ScriptValue
	attrs := make(map[string]engine.ScriptValue)
	for k, v := range sl.config.Attributes {
		attrs[k] = engine.NewCustomValue("any", v)
	}

	return engine.NewObjectValue(map[string]engine.ScriptValue{
		"default_level":    engine.NewStringValue(sl.config.DefaultLevel),
		"enable_debug":     engine.NewBoolValue(sl.config.EnableDebug),
		"enable_structure": engine.NewBoolValue(sl.config.EnableStructure),
		"format":           engine.NewStringValue(sl.config.Format),
		"components":       engine.NewArrayValue(components),
		"attributes":       engine.NewObjectValue(attrs),
		"output_target":    engine.NewStringValue(sl.config.OutputTarget),
	}), nil
}

// enableComponent enables debug logging for a component
func (sl *ScriptLoggerBridge) enableComponent(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("enableComponent requires component name")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	sl.mu.Lock()
	sl.config.Components[component] = true
	sl.mu.Unlock()

	// Forward to debug bridge
	return sl.debugBridge.ExecuteMethod(ctx, "enableDebugComponent", args)
}

// disableComponent disables debug logging for a component
func (sl *ScriptLoggerBridge) disableComponent(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("disableComponent requires component name")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	sl.mu.Lock()
	sl.config.Components[component] = false
	sl.mu.Unlock()

	// Forward to debug bridge
	return sl.debugBridge.ExecuteMethod(ctx, "disableDebugComponent", args)
}

// listEnabledComponents lists components with debug logging enabled
func (sl *ScriptLoggerBridge) listEnabledComponents(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Forward to debug bridge
	return sl.debugBridge.ExecuteMethod(ctx, "listEnabledComponents", args)
}

// Convenience logging methods

// debug logs a debug message
func (sl *ScriptLoggerBridge) debug(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("debug requires message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	var component string
	var attributes map[string]interface{}

	// Parse optional component and attributes
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeString {
		component = args[1].(engine.StringValue).Value()
	}

	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
		objFields := args[2].(engine.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	} else if len(args) == 2 && args[1] != nil && args[1].Type() == engine.TypeObject {
		// If only 2 args and second is object, it's attributes
		objFields := args[1].(engine.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "debug", component, message, attributes)
}

// info logs an info message
func (sl *ScriptLoggerBridge) info(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("info requires message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		objFields := args[1].(engine.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "info", "", message, attributes)
}

// warn logs a warning message
func (sl *ScriptLoggerBridge) warn(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("warn requires message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		objFields := args[1].(engine.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "warn", "", message, attributes)
}

// error logs an error message
func (sl *ScriptLoggerBridge) error(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("error requires message")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(engine.StringValue).Value()

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		objFields := args[1].(engine.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "error", "", message, attributes)
}

// logBridgeError logs bridge-related errors
func (sl *ScriptLoggerBridge) logBridgeError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("logBridgeError requires bridgeId, operation, and error")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("bridgeId must be a string")
	}
	bridgeID := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("operation must be a string")
	}
	operation := args[1].(engine.StringValue).Value()

	if args[2] == nil || args[2].Type() != engine.TypeString {
		return nil, fmt.Errorf("error must be a string")
	}
	errorMsg := args[2].(engine.StringValue).Value()

	// Build error attributes
	errorAttrs := map[string]interface{}{
		"bridge_id": bridgeID,
		"operation": operation,
		"error":     errorMsg,
	}

	// Add optional context
	if len(args) > 3 && args[3] != nil && args[3].Type() == engine.TypeObject {
		objFields := args[3].(engine.ObjectValue).Fields()
		for k, v := range objFields {
			errorAttrs[k] = v.ToGo()
		}
	}

	message := fmt.Sprintf("[Bridge Error] %s.%s: %s", bridgeID, operation, errorMsg)
	return sl.logWithLevelAndContext(ctx, "error", "bridge", message, errorAttrs)
}

// formatMessage formats log message with template
func (sl *ScriptLoggerBridge) formatMessage(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("formatMessage requires template and attributes")
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("template must be a string")
	}
	template := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}
	objFields := args[1].(engine.ObjectValue).Fields()

	// Simple template replacement
	result := template
	for k, v := range objFields {
		placeholder := fmt.Sprintf("{%s}", k)
		value := fmt.Sprintf("%v", v.ToGo())
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return engine.NewStringValue(result), nil
}

// Helper methods

// logWithLevelAndContext performs the actual logging with unified handling
func (sl *ScriptLoggerBridge) logWithLevelAndContext(ctx context.Context, level, component, message string, attributes map[string]interface{}) (engine.ScriptValue, error) {
	// Merge with global context
	mergedAttrs := sl.mergeAttributes(attributes)

	// Debug logging for components
	if sl.config.EnableDebug && component != "" {
		// Check if component is enabled
		if enabled, ok := sl.config.Components[component]; ok && enabled {
			// Use debug bridge for component logging
			debugArgs := []engine.ScriptValue{
				engine.NewStringValue(component),
				engine.NewStringValue(fmt.Sprintf("[%s] %s", strings.ToUpper(level), message)),
			}
			_, _ = sl.debugBridge.ExecuteMethod(ctx, "debugPrintln", debugArgs)
		}
	}

	// Structured logging
	if sl.config.EnableStructure {
		// Convert attributes to ScriptValue
		var attrValue engine.ScriptValue
		if len(mergedAttrs) > 0 {
			attrMap := make(map[string]engine.ScriptValue)
			for k, v := range mergedAttrs {
				attrMap[k] = engine.NewCustomValue("any", v)
			}
			attrValue = engine.NewObjectValue(attrMap)
		}

		// Use slog bridge for structured logging
		slogArgs := []engine.ScriptValue{
			engine.NewStringValue(message),
		}
		if attrValue != nil {
			// Add emoji based on level
			emoji := ""
			switch level {
			case "debug":
				emoji = "🐛"
			case "info":
				emoji = "ℹ️"
			case "warn":
				emoji = "⚠️"
			case "error":
				emoji = "❌"
			}
			if emoji != "" {
				slogArgs = append(slogArgs, engine.NewStringValue(emoji))
			}
			slogArgs = append(slogArgs, attrValue)
		}

		// Call appropriate slog method
		_, _ = sl.slogBridge.ExecuteMethod(ctx, level, slogArgs)
	}

	return engine.NewNilValue(), nil
}

// mergeAttributes merges provided attributes with global context
func (sl *ScriptLoggerBridge) mergeAttributes(attrs map[string]interface{}) map[string]interface{} {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// Start with global context
	merged := make(map[string]interface{})
	for k, v := range sl.contextAttrs {
		merged[k] = v
	}

	// Override with provided attributes
	for k, v := range attrs {
		merged[k] = v
	}

	return merged
}
