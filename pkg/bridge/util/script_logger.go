// ABOUTME: Script-friendly unified logging interface combining debug and structured logging capabilities
// ABOUTME: Provides context propagation, bridge integration, and customizable output formatting for scripts

package util

import (
	"context"
	"fmt"
	"strings"
	"sync"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// ScriptLoggerBridge provides a unified script-friendly logging interface.
// It combines debug and structured logging capabilities with context propagation
// and customizable output formatting for comprehensive script logging.
type ScriptLoggerBridge struct {
	mu           sync.RWMutex
	initialized  bool
	debugBridge  *DebugBridge           // Debug logging functionality
	slogBridge   *SlogBridge            // Structured logging functionality
	contextAttrs map[string]interface{} // Global context attributes
	config       *LoggerConfig          // Logger configuration
}

// LoggerConfig holds configuration for the script logger.
// It specifies logging levels, formats, components, and output targets.
type LoggerConfig struct {
	DefaultLevel    string                 `json:"default_level"`    // Default log level
	EnableDebug     bool                   `json:"enable_debug"`     // Enable debug logging
	EnableStructure bool                   `json:"enable_structure"` // Enable structured logging
	Format          string                 `json:"format"`           // Output format: text, json
	Components      map[string]bool        `json:"components"`       // Debug component states
	Attributes      map[string]interface{} `json:"attributes"`       // Global attributes
	OutputTarget    string                 `json:"output_target"`    // Output target: stderr, stdout, file
}

// DefaultLoggerConfig returns default configuration.
// It sets up reasonable defaults for script logging with both debug and structured modes.
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

// NewScriptLoggerBridge creates a new unified script logger bridge.
// It initializes both debug and structured logging bridges for comprehensive logging.
func NewScriptLoggerBridge() *ScriptLoggerBridge {
	return &ScriptLoggerBridge{
		debugBridge:  NewDebugBridge(),
		slogBridge:   NewSlogBridge(),
		contextAttrs: make(map[string]interface{}),
		config:       DefaultLoggerConfig(),
	}
}

// GetID returns the bridge identifier.
// It implements the types.Bridge interface.
func (sl *ScriptLoggerBridge) GetID() string {
	return "util_script_logger"
}

// GetMetadata returns bridge metadata.
// It provides information about the script logger bridge including
// version, description, and dependencies on debug and slog bridges.
func (sl *ScriptLoggerBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "util_script_logger",
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

// Initialize sets up the script logger bridge.
// It initializes both debug and structured logging sub-bridges.
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

// Cleanup performs bridge cleanup.
// It cleans up both sub-bridges and clears context attributes.
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

// IsInitialized returns initialization status.
// It returns true if the bridge has been initialized and is ready for use.
func (sl *ScriptLoggerBridge) IsInitialized() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// It enables the script engine to access unified logging through this bridge.
func (sl *ScriptLoggerBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns available bridge methods.
// It provides metadata about all logging methods available to scripts,
// including unified logging, context management, and convenience methods.
func (sl *ScriptLoggerBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Unified logging methods
		{
			Name:        "log",
			Description: "Unified logging method with level, component, and structured attributes",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "attributes", Type: "object", Required: true, Description: "Context attributes"},
			},
			ReturnType: "object",
			Examples:   []string{"withContext({session_id: 'abc123', user: 'john'})"},
		},
		{
			Name:        "setGlobalContext",
			Description: "Set global context attributes for all logs",
			Parameters: []types.ParameterInfo{
				{Name: "attributes", Type: "object", Required: true, Description: "Global attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"setGlobalContext({app: 'go-llmspell', version: '1.0'})"},
		},
		{
			Name:        "clearGlobalContext",
			Description: "Clear global context attributes",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
			Examples:    []string{"clearGlobalContext()"},
		},
		// Configuration methods
		{
			Name:        "configure",
			Description: "Configure the script logger with comprehensive settings",
			Parameters: []types.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Logger configuration"},
			},
			ReturnType: "void",
			Examples:   []string{"configure({format: 'json', level: 'debug', components: ['agent', 'tools']})"},
		},
		{
			Name:        "getConfiguration",
			Description: "Get current logger configuration",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getConfiguration()"},
		},
		// Component management (for debug logging)
		{
			Name:        "enableComponent",
			Description: "Enable debug logging for a component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"enableComponent('agent')"},
		},
		{
			Name:        "disableComponent",
			Description: "Disable debug logging for a component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"disableComponent('agent')"},
		},
		{
			Name:        "listEnabledComponents",
			Description: "List components with debug logging enabled",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listEnabledComponents()"},
		},
		// Convenience methods
		{
			Name:        "debug",
			Description: "Log debug message with optional component and attributes",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Info message"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"info('Process started', {pid: 1234, user: 'john'})"},
		},
		{
			Name:        "warn",
			Description: "Log warning message with optional attributes",
			Parameters: []types.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Warning message"},
				{Name: "attributes", Type: "object", Required: false, Description: "Structured attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"warn('Rate limit approaching', {remaining: 10, limit: 100})"},
		},
		{
			Name:        "error",
			Description: "Log error message with optional attributes",
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
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
			Parameters: []types.ParameterInfo{
				{Name: "template", Type: "string", Required: true, Description: "Message template"},
				{Name: "attributes", Type: "object", Required: true, Description: "Template attributes"},
			},
			ReturnType: "string",
			Examples:   []string{"formatMessage('User {user} completed task {task}', {user: 'john', task: 'upload'})"},
		},
	}
}

// ValidateMethod validates method calls.
// It delegates validation to the engine based on Methods() metadata.
func (sl *ScriptLoggerBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// TypeMappings returns type conversion mappings.
// It defines how Go logging types are mapped to script types
// for configuration, context, and logger instances.
func (sl *ScriptLoggerBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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

// RequiredPermissions returns required permissions.
// It specifies permissions for context management and log output.
func (sl *ScriptLoggerBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionMemory,
			Resource:    "util_script_logger.context",
			Actions:     []string{"read", "write"},
			Description: "Manage logging context and global attributes",
		},
		{
			Type:        types.PermissionStorage,
			Resource:    "util_script_logger.output",
			Actions:     []string{"write"},
			Description: "Write log output to configured targets",
		},
	}
}

// ExecuteMethod executes a bridge method.
// It implements the types.Bridge interface, routing method calls
// to the appropriate logging operations.
func (sl *ScriptLoggerBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	// Check initialization first
	sl.mu.RLock()
	if !sl.initialized {
		sl.mu.RUnlock()
		return nil, fmt.Errorf("script logger bridge not initialized")
	}
	sl.mu.RUnlock()

	// Methods that need write locks handle their own locking
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

// log is the unified logging method.
// It handles multiple argument patterns for flexible logging with level, component, and attributes.
func (sl *ScriptLoggerBridge) log(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("log requires at least level and message")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("level must be a string")
	}
	level := args[0].(types.StringValue).Value()

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
		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("message must be a string")
		}
		message = args[1].(types.StringValue).Value()
	} else if len(args) == 3 {
		// Could be log(level, message, attributes) or log(level, component, message)
		// Try to detect: if args[2] is a string, then args[1] is component and args[2] is message
		// If args[2] is not a string, then args[1] is message and args[2] is attributes
		if args[2] != nil && args[2].Type() == types.TypeString {
			// log(level, component, message)
			if args[1] == nil || args[1].Type() != types.TypeString {
				return nil, fmt.Errorf("component must be a string")
			}
			component = args[1].(types.StringValue).Value()
			message = args[2].(types.StringValue).Value()
		} else {
			// log(level, message, attributes)
			if args[1] == nil || args[1].Type() != types.TypeString {
				return nil, fmt.Errorf("message must be a string")
			}
			message = args[1].(types.StringValue).Value()
			if args[2] != nil && args[2].Type() == types.TypeObject {
				objFields := args[2].(types.ObjectValue).Fields()
				attributes = make(map[string]interface{})
				for k, v := range objFields {
					attributes[k] = v.ToGo()
				}
			}
		}
	} else if len(args) >= 4 {
		// log(level, component, message, attributes)
		if args[1] == nil || args[1].Type() != types.TypeString {
			return nil, fmt.Errorf("component must be a string")
		}
		component = args[1].(types.StringValue).Value()

		if args[2] == nil || args[2].Type() != types.TypeString {
			return nil, fmt.Errorf("message must be a string")
		}
		message = args[2].(types.StringValue).Value()

		if args[3] != nil && args[3].Type() == types.TypeObject {
			objFields := args[3].(types.ObjectValue).Fields()
			attributes = make(map[string]interface{})
			for k, v := range objFields {
				attributes[k] = v.ToGo()
			}
		}
	}

	return sl.logWithLevelAndContext(ctx, level, component, message, attributes)
}

// logWithContext logs with additional context propagation.
// It merges provided context with global attributes for enriched logging.
func (sl *ScriptLoggerBridge) logWithContext(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("logWithContext requires level and message")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("level must be a string")
	}
	level := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[1].(types.StringValue).Value()

	var context map[string]interface{}
	if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeObject {
		objFields := args[2].(types.ObjectValue).Fields()
		context = make(map[string]interface{})
		for k, v := range objFields {
			context[k] = v.ToGo()
		}
	}

	// Merge context with global attributes
	mergedAttrs := sl.mergeAttributes(context)
	return sl.logWithLevelAndContext(ctx, level, "", message, mergedAttrs)
}

// withContext creates logging context with attributes.
// It returns a context object that can be used for scoped logging.
func (sl *ScriptLoggerBridge) withContext(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("withContext requires attributes")
	}

	if args[0] == nil || args[0].Type() != types.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}

	// Create context object
	contextObj := map[string]types.ScriptValue{
		"attributes": args[0],
		"logger":     types.NewCustomValue("logger", sl),
	}

	// If using slog, also create a slog context
	if sl.config.EnableStructure {
		result, err := sl.slogBridge.ExecuteMethod(ctx, "withAttributes", args)
		if err == nil && result != nil {
			contextObj["slogContext"] = result
		}
	}

	return types.NewObjectValue(contextObj), nil
}

// setGlobalContext sets global context attributes.
// These attributes are included in all subsequent log messages.
func (sl *ScriptLoggerBridge) setGlobalContext(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("setGlobalContext requires attributes")
	}

	if args[0] == nil || args[0].Type() != types.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	// Convert to native map
	objFields := args[0].(types.ObjectValue).Fields()
	sl.contextAttrs = make(map[string]interface{})
	for k, v := range objFields {
		sl.contextAttrs[k] = v.ToGo()
	}

	return types.NewNilValue(), nil
}

// clearGlobalContext clears global context attributes.
// It removes all global attributes from future log messages.
func (sl *ScriptLoggerBridge) clearGlobalContext(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.contextAttrs = make(map[string]interface{})
	return types.NewNilValue(), nil
}

// configure configures the script logger.
// It updates logger settings including level, format, and enabled components.
func (sl *ScriptLoggerBridge) configure(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("configure requires config object")
	}

	if args[0] == nil || args[0].Type() != types.TypeObject {
		return nil, fmt.Errorf("config must be an object")
	}

	sl.mu.Lock()
	defer sl.mu.Unlock()

	configObj := args[0].(types.ObjectValue).Fields()

	// Update configuration
	if levelVal, ok := configObj["level"]; ok && levelVal.Type() == types.TypeString {
		sl.config.DefaultLevel = levelVal.(types.StringValue).Value()
	}

	if formatVal, ok := configObj["format"]; ok && formatVal.Type() == types.TypeString {
		sl.config.Format = formatVal.(types.StringValue).Value()
		// Configure slog with new format
		_, _ = sl.slogBridge.ExecuteMethod(ctx, "configureLogger", []types.ScriptValue{
			types.NewObjectValue(map[string]types.ScriptValue{
				"format": formatVal,
			}),
		})
	}

	if debugVal, ok := configObj["enable_debug"]; ok && debugVal.Type() == types.TypeBool {
		sl.config.EnableDebug = debugVal.(types.BoolValue).Value()
	}

	if structureVal, ok := configObj["enable_structure"]; ok && structureVal.Type() == types.TypeBool {
		sl.config.EnableStructure = structureVal.(types.BoolValue).Value()
	}

	if componentsVal, ok := configObj["components"]; ok && componentsVal.Type() == types.TypeArray {
		components := componentsVal.(types.ArrayValue).Elements()
		sl.config.Components = make(map[string]bool)
		for _, comp := range components {
			if comp.Type() == types.TypeString {
				compName := comp.(types.StringValue).Value()
				sl.config.Components[compName] = true
				// Enable in debug bridge
				_, _ = sl.debugBridge.ExecuteMethod(ctx, "enableDebugComponent", []types.ScriptValue{
					types.NewStringValue(compName),
				})
			}
		}
	}

	return types.NewNilValue(), nil
}

// getConfiguration gets current logger configuration.
// It returns all current logger settings and enabled components.
func (sl *ScriptLoggerBridge) getConfiguration(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	// Convert components map to array
	components := make([]types.ScriptValue, 0, len(sl.config.Components))
	for comp, enabled := range sl.config.Components {
		if enabled {
			components = append(components, types.NewStringValue(comp))
		}
	}

	// Convert attributes to ScriptValue
	attrs := make(map[string]types.ScriptValue)
	for k, v := range sl.config.Attributes {
		attrs[k] = types.NewCustomValue("any", v)
	}

	return types.NewObjectValue(map[string]types.ScriptValue{
		"default_level":    types.NewStringValue(sl.config.DefaultLevel),
		"enable_debug":     types.NewBoolValue(sl.config.EnableDebug),
		"enable_structure": types.NewBoolValue(sl.config.EnableStructure),
		"format":           types.NewStringValue(sl.config.Format),
		"components":       types.NewArrayValue(components),
		"attributes":       types.NewObjectValue(attrs),
		"output_target":    types.NewStringValue(sl.config.OutputTarget),
	}), nil
}

// enableComponent enables debug logging for a component.
// It activates debug output for the specified component name.
func (sl *ScriptLoggerBridge) enableComponent(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("enableComponent requires component name")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	sl.mu.Lock()
	sl.config.Components[component] = true
	sl.mu.Unlock()

	// Forward to debug bridge
	return sl.debugBridge.ExecuteMethod(ctx, "enableDebugComponent", args)
}

// disableComponent disables debug logging for a component.
// It deactivates debug output for the specified component name.
func (sl *ScriptLoggerBridge) disableComponent(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("disableComponent requires component name")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	sl.mu.Lock()
	sl.config.Components[component] = false
	sl.mu.Unlock()

	// Forward to debug bridge
	return sl.debugBridge.ExecuteMethod(ctx, "disableDebugComponent", args)
}

// listEnabledComponents lists components with debug logging enabled.
// It delegates to the debug bridge to get the component list.
func (sl *ScriptLoggerBridge) listEnabledComponents(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	// Forward to debug bridge
	return sl.debugBridge.ExecuteMethod(ctx, "listEnabledComponents", args)
}

// Convenience logging methods

// debug logs a debug message.
// It supports optional component name and structured attributes.
func (sl *ScriptLoggerBridge) debug(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("debug requires message")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(types.StringValue).Value()

	var component string
	var attributes map[string]interface{}

	// Parse optional component and attributes
	if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeString {
		component = args[1].(types.StringValue).Value()
	}

	if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeObject {
		objFields := args[2].(types.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	} else if len(args) == 2 && args[1] != nil && args[1].Type() == types.TypeObject {
		// If only 2 args and second is object, it's attributes
		objFields := args[1].(types.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "debug", component, message, attributes)
}

// info logs an info message.
// It supports optional structured attributes for context.
func (sl *ScriptLoggerBridge) info(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("info requires message")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(types.StringValue).Value()

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeObject {
		objFields := args[1].(types.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "info", "", message, attributes)
}

// warn logs a warning message.
// It supports optional structured attributes for context.
func (sl *ScriptLoggerBridge) warn(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("warn requires message")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(types.StringValue).Value()

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeObject {
		objFields := args[1].(types.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "warn", "", message, attributes)
}

// error logs an error message.
// It supports optional structured attributes for context.
func (sl *ScriptLoggerBridge) error(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("error requires message")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("message must be a string")
	}
	message := args[0].(types.StringValue).Value()

	var attributes map[string]interface{}
	if len(args) > 1 && args[1] != nil && args[1].Type() == types.TypeObject {
		objFields := args[1].(types.ObjectValue).Fields()
		attributes = make(map[string]interface{})
		for k, v := range objFields {
			attributes[k] = v.ToGo()
		}
	}

	return sl.logWithLevelAndContext(ctx, "error", "", message, attributes)
}

// logBridgeError logs bridge-related errors.
// It provides standardized error logging for bridge operations.
func (sl *ScriptLoggerBridge) logBridgeError(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("logBridgeError requires bridgeId, operation, and error")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("bridgeId must be a string")
	}
	bridgeID := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeString {
		return nil, fmt.Errorf("operation must be a string")
	}
	operation := args[1].(types.StringValue).Value()

	if args[2] == nil || args[2].Type() != types.TypeString {
		return nil, fmt.Errorf("error must be a string")
	}
	errorMsg := args[2].(types.StringValue).Value()

	// Build error attributes
	errorAttrs := map[string]interface{}{
		"bridge_id": bridgeID,
		"operation": operation,
		"error":     errorMsg,
	}

	// Add optional context
	if len(args) > 3 && args[3] != nil && args[3].Type() == types.TypeObject {
		objFields := args[3].(types.ObjectValue).Fields()
		for k, v := range objFields {
			errorAttrs[k] = v.ToGo()
		}
	}

	message := fmt.Sprintf("[Bridge Error] %s.%s: %s", bridgeID, operation, errorMsg)
	return sl.logWithLevelAndContext(ctx, "error", "bridge", message, errorAttrs)
}

// formatMessage formats log message with template.
// It performs simple template substitution with provided attributes.
func (sl *ScriptLoggerBridge) formatMessage(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("formatMessage requires template and attributes")
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("template must be a string")
	}
	template := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeObject {
		return nil, fmt.Errorf("attributes must be an object")
	}
	objFields := args[1].(types.ObjectValue).Fields()

	// Simple template replacement
	result := template
	for k, v := range objFields {
		placeholder := fmt.Sprintf("{%s}", k)
		value := fmt.Sprintf("%v", v.ToGo())
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return types.NewStringValue(result), nil
}

// Helper methods

// logWithLevelAndContext performs the actual logging with unified handling.
// It routes to both debug and structured logging based on configuration.
func (sl *ScriptLoggerBridge) logWithLevelAndContext(ctx context.Context, level, component, message string, attributes map[string]interface{}) (types.ScriptValue, error) {
	// Merge with global context
	mergedAttrs := sl.mergeAttributes(attributes)

	// Debug logging for components
	if sl.config.EnableDebug && component != "" {
		// Check if component is enabled
		if enabled, ok := sl.config.Components[component]; ok && enabled {
			// Use debug bridge for component logging
			debugArgs := []types.ScriptValue{
				types.NewStringValue(component),
				types.NewStringValue(fmt.Sprintf("[%s] %s", strings.ToUpper(level), message)),
			}
			_, _ = sl.debugBridge.ExecuteMethod(ctx, "debugPrintln", debugArgs)
		}
	}

	// Structured logging
	if sl.config.EnableStructure {
		// Convert attributes to ScriptValue
		var attrValue types.ScriptValue
		if len(mergedAttrs) > 0 {
			attrMap := make(map[string]types.ScriptValue)
			for k, v := range mergedAttrs {
				attrMap[k] = types.NewCustomValue("any", v)
			}
			attrValue = types.NewObjectValue(attrMap)
		}

		// Use slog bridge for structured logging
		slogArgs := []types.ScriptValue{
			types.NewStringValue(message),
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
				slogArgs = append(slogArgs, types.NewStringValue(emoji))
			}
			slogArgs = append(slogArgs, attrValue)
		}

		// Call appropriate slog method
		_, _ = sl.slogBridge.ExecuteMethod(ctx, level, slogArgs)
	}

	return types.NewNilValue(), nil
}

// mergeAttributes merges provided attributes with global context.
// It combines global attributes with message-specific attributes.
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
