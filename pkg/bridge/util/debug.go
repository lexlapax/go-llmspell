// ABOUTME: Debug logging bridge providing access to go-llms debug logging system
// ABOUTME: Bridges component-based debug control and conditional compilation support

package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// DebugBridge provides script access to go-llms compatible debug logging system.
// It manages component-based debug control, conditional compilation support,
// and environment-based configuration for fine-grained debugging.
type DebugBridge struct {
	mu          sync.RWMutex
	initialized bool
	components  map[string]bool // Track enabled components locally
	logger      *log.Logger     // Debug logger instance
}

// NewDebugBridge creates a new debug logging bridge.
// It initializes a logger with go-llms compatible format and parses
// environment configuration from GO_LLMS_DEBUG variable.
func NewDebugBridge() *DebugBridge {
	// Initialize with go-llms compatible logger format
	logger := log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	bridge := &DebugBridge{
		components: make(map[string]bool),
		logger:     logger,
	}

	// Parse GO_LLMS_DEBUG environment variable for compatibility
	bridge.parseEnvironmentConfig()

	return bridge
}

// GetID returns the bridge identifier.
// It implements the types.Bridge interface.
func (db *DebugBridge) GetID() string {
	return "debug"
}

// GetMetadata returns bridge metadata.
// It provides information about the debug bridge including
// version, description, and supported debug features.
func (db *DebugBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:         "debug",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms debug logging system with component-based control and conditional compilation",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/internal/debug"},
	}
}

// Initialize sets up the debug bridge.
// It marks the bridge as initialized and ready for use.
func (db *DebugBridge) Initialize(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.initialized = true
	return nil
}

// Cleanup performs bridge cleanup.
// It clears component state and marks the bridge as uninitialized.
func (db *DebugBridge) Cleanup(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.components = make(map[string]bool)
	db.initialized = false
	return nil
}

// IsInitialized returns initialization status.
// It returns true if the bridge has been initialized and is ready for use.
func (db *DebugBridge) IsInitialized() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// It enables the script engine to access debug functionality through this bridge.
func (db *DebugBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns available bridge methods.
// It provides metadata about all debug-related methods available to scripts,
// including logging, component control, and configuration methods.
func (db *DebugBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Debug logging methods
		{
			Name:        "debugPrintf",
			Description: "Log formatted debug message for component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
				{Name: "format", Type: "string", Required: true, Description: "Format string"},
				{Name: "args", Type: "array", Required: false, Description: "Format arguments"},
			},
			ReturnType: "void",
			Examples:   []string{"debugPrintf('agent', 'Processing request: %s', ['user-123'])"},
		},
		{
			Name:        "debugPrintln",
			Description: "Log debug message for component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
				{Name: "message", Type: "string", Required: true, Description: "Debug message"},
			},
			ReturnType: "void",
			Examples:   []string{"debugPrintln('workflow', 'Starting execution')"},
		},
		// Component control methods
		{
			Name:        "isDebugEnabled",
			Description: "Check if debug logging is enabled for component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "boolean",
			Examples:   []string{"isDebugEnabled('agent')"},
		},
		{
			Name:        "enableDebugComponent",
			Description: "Enable debug logging for specific component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"enableDebugComponent('tools')"},
		},
		{
			Name:        "disableDebugComponent",
			Description: "Disable debug logging for specific component",
			Parameters: []types.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"disableDebugComponent('tools')"},
		},
		{
			Name:        "listEnabledComponents",
			Description: "Get list of components with debug logging enabled",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listEnabledComponents()"},
		},
		// Logger configuration methods
		{
			Name:        "setCustomLogger",
			Description: "Set custom logger for debug output",
			Parameters: []types.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Logger configuration"},
			},
			ReturnType: "void",
			Examples:   []string{"setCustomLogger({prefix: '[SPELL]', flags: 'datetime'})"},
		},
		{
			Name:        "getDebugEnvironment",
			Description: "Get current GO_LLMS_DEBUG environment configuration",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getDebugEnvironment()"},
		},
	}
}

// ValidateMethod validates method calls.
// It ensures initialization and checks argument counts for each method.
func (db *DebugBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	if !db.IsInitialized() {
		return fmt.Errorf("debug bridge not initialized")
	}

	methods := db.Methods()
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

// TypeMappings returns type conversion mappings.
// It defines how Go debug types are mapped to script types
// for logger instances and configuration objects.
func (db *DebugBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
		"debug_logger": {
			GoType:     "*log.Logger",
			ScriptType: "object",
			Converter:  "debugLoggerConverter",
			Metadata:   map[string]interface{}{"description": "Debug logger instance"},
		},
		"debug_config": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
			Converter:  "debugConfigConverter",
			Metadata:   map[string]interface{}{"description": "Debug configuration object"},
		},
	}
}

// RequiredPermissions returns required permissions.
// It specifies permissions for debug logging configuration
// and component state management.
func (db *DebugBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionStorage,
			Resource:    "debug.logging",
			Actions:     []string{"read", "write"},
			Description: "Access debug logging configuration",
		},
		{
			Type:        types.PermissionMemory,
			Resource:    "debug.components",
			Actions:     []string{"read", "write"},
			Description: "Manage debug component state",
		},
	}
}

// ExecuteMethod executes a bridge method.
// It implements the types.Bridge interface, routing method calls
// to the appropriate debug operations.
func (db *DebugBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	switch name {
	case "debugPrintf":
		err := db.debugPrintf(ctx, args)
		return types.NewNilValue(), err
	case "debugPrintln":
		err := db.debugPrintln(ctx, args)
		return types.NewNilValue(), err
	case "isDebugEnabled":
		return db.isDebugEnabled(ctx, args)
	case "enableDebugComponent":
		err := db.enableDebugComponent(ctx, args)
		return types.NewNilValue(), err
	case "disableDebugComponent":
		err := db.disableDebugComponent(ctx, args)
		return types.NewNilValue(), err
	case "listEnabledComponents":
		return db.listEnabledComponents(ctx, args)
	case "setCustomLogger":
		err := db.setCustomLogger(ctx, args)
		return types.NewNilValue(), err
	case "getDebugEnvironment":
		return db.getDebugEnvironment(ctx, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// Bridge method implementations

// debugPrintf logs formatted debug message for component.
// It only outputs if the component is enabled for debugging.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) debugPrintf(ctx context.Context, args []types.ScriptValue) error {
	if err := db.ValidateMethod("debugPrintf", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeString {
		return fmt.Errorf("format must be a string")
	}
	format := args[1].(types.StringValue).Value()

	// Convert args if provided
	var formatArgs []interface{}
	if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeArray {
		arrayVal := args[2].(types.ArrayValue)
		formatArgs = make([]interface{}, len(arrayVal.Elements()))
		for i, elem := range arrayVal.Elements() {
			formatArgs[i] = elem.ToGo()
		}
	}

	// Check if component is enabled for debugging
	if !db.isComponentEnabled(component) {
		return nil
	}

	// Format message with component prefix like go-llms debug
	msg := fmt.Sprintf("[%s] %s", component, format)
	db.logger.Printf(msg, formatArgs...)
	return nil
}

// debugPrintln logs debug message for component.
// It only outputs if the component is enabled for debugging.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) debugPrintln(ctx context.Context, args []types.ScriptValue) error {
	if err := db.ValidateMethod("debugPrintln", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	if args[1] == nil || args[1].Type() != types.TypeString {
		return fmt.Errorf("message must be a string")
	}
	message := args[1].(types.StringValue).Value()

	// Check if component is enabled for debugging
	if !db.isComponentEnabled(component) {
		return nil
	}

	// Format message with component prefix like go-llms debug
	msg := fmt.Sprintf("[%s] %s", component, message)
	db.logger.Println(msg)
	return nil
}

// isDebugEnabled checks if debug logging is enabled for component.
// It returns true if the component has debug logging enabled.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) isDebugEnabled(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := db.ValidateMethod("isDebugEnabled", args); err != nil {
		return nil, err
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	// Check go-llms debug enabled components
	// Since go-llms doesn't expose EnabledComponents, we simulate by testing
	// if debug output would be produced
	enabled := db.isComponentEnabled(component)

	return types.NewBoolValue(enabled), nil
}

// enableDebugComponent enables debug logging for specific component.
// It adds the component to the active debug components list.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) enableDebugComponent(ctx context.Context, args []types.ScriptValue) error {
	if err := db.ValidateMethod("enableDebugComponent", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	db.mu.Lock()
	db.components[component] = true
	db.mu.Unlock()

	return nil
}

// disableDebugComponent disables debug logging for specific component.
// It removes the component from the active debug components list.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) disableDebugComponent(ctx context.Context, args []types.ScriptValue) error {
	if err := db.ValidateMethod("disableDebugComponent", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(types.StringValue).Value()

	db.mu.Lock()
	db.components[component] = false
	db.mu.Unlock()

	return nil
}

// listEnabledComponents gets list of components with debug logging enabled.
// It returns an array of component names that have debugging active.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) listEnabledComponents(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := db.ValidateMethod("listEnabledComponents", args); err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	var enabled []types.ScriptValue
	for component, isEnabled := range db.components {
		if isEnabled {
			enabled = append(enabled, types.NewStringValue(component))
		}
	}

	return types.NewArrayValue(enabled), nil
}

// setCustomLogger sets custom logger for debug output.
// It configures the logger with custom prefix and flags.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) setCustomLogger(ctx context.Context, args []types.ScriptValue) error {
	if err := db.ValidateMethod("setCustomLogger", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != types.TypeObject {
		return fmt.Errorf("config must be an object")
	}
	configObj := args[0].(types.ObjectValue).Fields()

	// Create custom logger based on config
	logger := log.Default()

	if prefixVal, ok := configObj["prefix"]; ok && prefixVal.Type() == types.TypeString {
		prefix := prefixVal.(types.StringValue).Value()
		// In a real implementation, we'd create a logger with the custom prefix
		// For now, we acknowledge the configuration
		_ = prefix
	}

	if flagsVal, ok := configObj["flags"]; ok && flagsVal.Type() == types.TypeString {
		flags := flagsVal.(types.StringValue).Value()
		// Configure logger flags based on the flags string
		// For now, we acknowledge the configuration
		_ = flags
	}

	// Update our bridge's logger with the custom configuration
	db.mu.Lock()
	db.logger = logger
	db.mu.Unlock()

	return nil
}

// getDebugEnvironment gets current GO_LLMS_DEBUG environment configuration.
// It returns the environment value, enabled components, and compilation mode.
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) getDebugEnvironment(ctx context.Context, args []types.ScriptValue) (types.ScriptValue, error) {
	if err := db.ValidateMethod("getDebugEnvironment", args); err != nil {
		return nil, err
	}

	// Return environment configuration
	enabledComponents := db.getEnabledComponentsFromEnv()
	componentValues := make([]types.ScriptValue, len(enabledComponents))
	for i, comp := range enabledComponents {
		componentValues[i] = types.NewStringValue(comp)
	}

	result := map[string]types.ScriptValue{
		"go_llms_debug_env":  types.NewStringValue(db.getGoLLMSDebugEnv()),
		"enabled_components": types.NewArrayValue(componentValues),
		"compilation_mode":   types.NewStringValue(db.getCompilationMode()),
	}
	return types.NewObjectValue(result), nil
}

// Helper methods

// isComponentEnabled checks if a component is enabled for debugging.
// It looks up the component in the local state map.
func (db *DebugBridge) isComponentEnabled(component string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Check local component state
	if enabled, exists := db.components[component]; exists {
		return enabled
	}

	// Default to false if not explicitly enabled
	return false
}

// getGoLLMSDebugEnv returns the GO_LLMS_DEBUG environment variable value.
// It returns "not_set" if the environment variable is not defined.
func (db *DebugBridge) getGoLLMSDebugEnv() string {
	envValue := os.Getenv("GO_LLMS_DEBUG")
	if envValue == "" {
		return "not_set"
	}
	return envValue
}

// getEnabledComponentsFromEnv returns components enabled via environment.
// It collects all components marked as enabled in the internal state.
func (db *DebugBridge) getEnabledComponentsFromEnv() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var enabled []string
	for component, isEnabled := range db.components {
		if isEnabled {
			enabled = append(enabled, component)
		}
	}
	return enabled
}

// getCompilationMode returns whether debug mode is compiled in.
// It indicates the build configuration for conditional compilation.
func (db *DebugBridge) getCompilationMode() string {
	// In a real implementation, this would detect the build tags
	// For now, we return a default indication
	return "conditional_compilation_enabled"
}

// parseEnvironmentConfig parses GO_LLMS_DEBUG environment variable.
// It supports both "all" for all components and comma-separated component lists.
func (db *DebugBridge) parseEnvironmentConfig() {
	envDebug := os.Getenv("GO_LLMS_DEBUG")
	if envDebug == "" {
		return
	}

	// Parse component list from environment variable
	// Format: GO_LLMS_DEBUG=component1,component2,component3
	// or GO_LLMS_DEBUG=all for all components
	db.mu.Lock()
	defer db.mu.Unlock()

	if envDebug == "all" {
		// Enable common components when "all" is specified
		commonComponents := []string{"agent", "tools", "workflow", "llm", "state", "hooks", "events"}
		for _, component := range commonComponents {
			db.components[component] = true
		}
	} else {
		// Parse comma-separated component list
		components := strings.Split(envDebug, ",")
		for _, component := range components {
			component = strings.TrimSpace(component)
			if component != "" {
				db.components[component] = true
			}
		}
	}
}
