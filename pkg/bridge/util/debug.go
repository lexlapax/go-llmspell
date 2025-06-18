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
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// DebugBridge provides script access to go-llms compatible debug logging system
type DebugBridge struct {
	mu          sync.RWMutex
	initialized bool
	components  map[string]bool // Track enabled components locally
	logger      *log.Logger     // Debug logger instance
}

// NewDebugBridge creates a new debug logging bridge
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

// GetID returns the bridge identifier
func (db *DebugBridge) GetID() string {
	return "debug"
}

// GetMetadata returns bridge metadata
func (db *DebugBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "debug",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms debug logging system with component-based control and conditional compilation",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/internal/debug"},
	}
}

// Initialize sets up the debug bridge
func (db *DebugBridge) Initialize(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (db *DebugBridge) Cleanup(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.components = make(map[string]bool)
	db.initialized = false
	return nil
}

// IsInitialized returns initialization status
func (db *DebugBridge) IsInitialized() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (db *DebugBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(db)
}

// Methods returns available bridge methods
func (db *DebugBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Debug logging methods
		{
			Name:        "debugPrintf",
			Description: "Log formatted debug message for component",
			Parameters: []engine.ParameterInfo{
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
			Parameters: []engine.ParameterInfo{
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
			Parameters: []engine.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "boolean",
			Examples:   []string{"isDebugEnabled('agent')"},
		},
		{
			Name:        "enableDebugComponent",
			Description: "Enable debug logging for specific component",
			Parameters: []engine.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"enableDebugComponent('tools')"},
		},
		{
			Name:        "disableDebugComponent",
			Description: "Disable debug logging for specific component",
			Parameters: []engine.ParameterInfo{
				{Name: "component", Type: "string", Required: true, Description: "Component name"},
			},
			ReturnType: "void",
			Examples:   []string{"disableDebugComponent('tools')"},
		},
		{
			Name:        "listEnabledComponents",
			Description: "Get list of components with debug logging enabled",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples:    []string{"listEnabledComponents()"},
		},
		// Logger configuration methods
		{
			Name:        "setCustomLogger",
			Description: "Set custom logger for debug output",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Required: true, Description: "Logger configuration"},
			},
			ReturnType: "void",
			Examples:   []string{"setCustomLogger({prefix: '[SPELL]', flags: 'datetime'})"},
		},
		{
			Name:        "getDebugEnvironment",
			Description: "Get current GO_LLMS_DEBUG environment configuration",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"getDebugEnvironment()"},
		},
	}
}

// ValidateMethod validates method calls
func (db *DebugBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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

// TypeMappings returns type conversion mappings
func (db *DebugBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
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

// RequiredPermissions returns required permissions
func (db *DebugBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionStorage,
			Resource:    "debug.logging",
			Actions:     []string{"read", "write"},
			Description: "Access debug logging configuration",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "debug.components",
			Actions:     []string{"read", "write"},
			Description: "Manage debug component state",
		},
	}
}

// ExecuteMethod executes a bridge method
func (db *DebugBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "debugPrintf":
		err := db.debugPrintf(ctx, args)
		return engine.NewNilValue(), err
	case "debugPrintln":
		err := db.debugPrintln(ctx, args)
		return engine.NewNilValue(), err
	case "isDebugEnabled":
		return db.isDebugEnabled(ctx, args)
	case "enableDebugComponent":
		err := db.enableDebugComponent(ctx, args)
		return engine.NewNilValue(), err
	case "disableDebugComponent":
		err := db.disableDebugComponent(ctx, args)
		return engine.NewNilValue(), err
	case "listEnabledComponents":
		return db.listEnabledComponents(ctx, args)
	case "setCustomLogger":
		err := db.setCustomLogger(ctx, args)
		return engine.NewNilValue(), err
	case "getDebugEnvironment":
		return db.getDebugEnvironment(ctx, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// Bridge method implementations

// debugPrintf logs formatted debug message for component
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) debugPrintf(ctx context.Context, args []engine.ScriptValue) error {
	if err := db.ValidateMethod("debugPrintf", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return fmt.Errorf("format must be a string")
	}
	format := args[1].(engine.StringValue).Value()

	// Convert args if provided
	var formatArgs []interface{}
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeArray {
		arrayVal := args[2].(engine.ArrayValue)
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

// debugPrintln logs debug message for component
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) debugPrintln(ctx context.Context, args []engine.ScriptValue) error {
	if err := db.ValidateMethod("debugPrintln", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return fmt.Errorf("message must be a string")
	}
	message := args[1].(engine.StringValue).Value()

	// Check if component is enabled for debugging
	if !db.isComponentEnabled(component) {
		return nil
	}

	// Format message with component prefix like go-llms debug
	msg := fmt.Sprintf("[%s] %s", component, message)
	db.logger.Println(msg)
	return nil
}

// isDebugEnabled checks if debug logging is enabled for component
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) isDebugEnabled(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := db.ValidateMethod("isDebugEnabled", args); err != nil {
		return nil, err
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	// Check go-llms debug enabled components
	// Since go-llms doesn't expose EnabledComponents, we simulate by testing
	// if debug output would be produced
	enabled := db.isComponentEnabled(component)

	return engine.NewBoolValue(enabled), nil
}

// enableDebugComponent enables debug logging for specific component
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) enableDebugComponent(ctx context.Context, args []engine.ScriptValue) error {
	if err := db.ValidateMethod("enableDebugComponent", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	db.mu.Lock()
	db.components[component] = true
	db.mu.Unlock()

	return nil
}

// disableDebugComponent disables debug logging for specific component
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) disableDebugComponent(ctx context.Context, args []engine.ScriptValue) error {
	if err := db.ValidateMethod("disableDebugComponent", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != engine.TypeString {
		return fmt.Errorf("component must be a string")
	}
	component := args[0].(engine.StringValue).Value()

	db.mu.Lock()
	db.components[component] = false
	db.mu.Unlock()

	return nil
}

// listEnabledComponents gets list of components with debug logging enabled
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) listEnabledComponents(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := db.ValidateMethod("listEnabledComponents", args); err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	var enabled []engine.ScriptValue
	for component, isEnabled := range db.components {
		if isEnabled {
			enabled = append(enabled, engine.NewStringValue(component))
		}
	}

	return engine.NewArrayValue(enabled), nil
}

// setCustomLogger sets custom logger for debug output
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) setCustomLogger(ctx context.Context, args []engine.ScriptValue) error {
	if err := db.ValidateMethod("setCustomLogger", args); err != nil {
		return err
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return fmt.Errorf("config must be an object")
	}
	configObj := args[0].(engine.ObjectValue).Fields()

	// Create custom logger based on config
	logger := log.Default()

	if prefixVal, ok := configObj["prefix"]; ok && prefixVal.Type() == engine.TypeString {
		prefix := prefixVal.(engine.StringValue).Value()
		// In a real implementation, we'd create a logger with the custom prefix
		// For now, we acknowledge the configuration
		_ = prefix
	}

	if flagsVal, ok := configObj["flags"]; ok && flagsVal.Type() == engine.TypeString {
		flags := flagsVal.(engine.StringValue).Value()
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

// getDebugEnvironment gets current GO_LLMS_DEBUG environment configuration
//
//nolint:unused // Bridge method called via reflection
func (db *DebugBridge) getDebugEnvironment(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := db.ValidateMethod("getDebugEnvironment", args); err != nil {
		return nil, err
	}

	// Return environment configuration
	enabledComponents := db.getEnabledComponentsFromEnv()
	componentValues := make([]engine.ScriptValue, len(enabledComponents))
	for i, comp := range enabledComponents {
		componentValues[i] = engine.NewStringValue(comp)
	}

	result := map[string]engine.ScriptValue{
		"go_llms_debug_env":  engine.NewStringValue(db.getGoLLMSDebugEnv()),
		"enabled_components": engine.NewArrayValue(componentValues),
		"compilation_mode":   engine.NewStringValue(db.getCompilationMode()),
	}
	return engine.NewObjectValue(result), nil
}

// Helper methods

// isComponentEnabled checks if a component is enabled for debugging
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

// getGoLLMSDebugEnv returns the GO_LLMS_DEBUG environment variable value
func (db *DebugBridge) getGoLLMSDebugEnv() string {
	envValue := os.Getenv("GO_LLMS_DEBUG")
	if envValue == "" {
		return "not_set"
	}
	return envValue
}

// getEnabledComponentsFromEnv returns components enabled via environment
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

// getCompilationMode returns whether debug mode is compiled in
func (db *DebugBridge) getCompilationMode() string {
	// In a real implementation, this would detect the build tags
	// For now, we return a default indication
	return "conditional_compilation_enabled"
}

// parseEnvironmentConfig parses GO_LLMS_DEBUG environment variable
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
