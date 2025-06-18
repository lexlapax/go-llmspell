// ABOUTME: Error utilities bridge provides access to go-llms error serialization system.
// ABOUTME: Wraps SerializableError, recovery strategies, aggregation, and error categorization.

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for error system
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/errors"
)

// UtilErrorsBridge provides script access to go-llms error serialization utilities.
type UtilErrorsBridge struct {
	mu          sync.RWMutex
	initialized bool

	// Enhanced components from go-llms v0.3.5
	eventEmitter    agentDomain.EventEmitter // For error event emission
	eventBus        *events.EventBus         // Event bus for error events
	errorHandlers   map[string]ErrorHandler  // Custom error handlers
	errorCategories map[string]ErrorCategory // Error categorization
	aggregator      errors.ErrorAggregator   // For error aggregation
}

// ErrorHandler defines a custom error handler function
type ErrorHandler func(error) error

// ErrorCategory defines error categorization metadata
type ErrorCategory struct {
	Name        string
	Description string
	Retryable   bool
	Fatal       bool
	Handler     ErrorHandler
}

// NewUtilErrorsBridge creates a new error utilities bridge.
func NewUtilErrorsBridge() *UtilErrorsBridge {
	return &UtilErrorsBridge{
		errorHandlers:   make(map[string]ErrorHandler),
		errorCategories: make(map[string]ErrorCategory),
	}
}

// NewUtilErrorsBridgeWithEventEmitter creates a new error utilities bridge with event emitter.
func NewUtilErrorsBridgeWithEventEmitter(eventEmitter agentDomain.EventEmitter) *UtilErrorsBridge {
	return &UtilErrorsBridge{
		eventEmitter:    eventEmitter,
		errorHandlers:   make(map[string]ErrorHandler),
		errorCategories: make(map[string]ErrorCategory),
	}
}

// GetID returns the bridge identifier.
func (b *UtilErrorsBridge) GetID() string {
	return "util_errors"
}

// GetMetadata returns bridge metadata.
func (b *UtilErrorsBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_errors",
		Version:     "2.0.0",
		Description: "Error serialization utilities with recovery strategies, aggregation, and categorization",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *UtilErrorsBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize event bus if needed
	if b.eventBus == nil {
		b.eventBus = events.NewEventBus()
	}

	// Initialize error aggregator
	if b.aggregator == nil {
		b.aggregator = errors.NewErrorAggregator()
	}

	// Set up default error categories
	b.setupDefaultCategories()

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *UtilErrorsBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *UtilErrorsBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *UtilErrorsBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *UtilErrorsBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Error creation and wrapping
		{
			Name:        "createError",
			Description: "Create a serializable error",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Description: "Error message", Required: true},
				{Name: "metadata", Type: "object", Description: "Error metadata", Required: false},
			},
			ReturnType: "SerializableError",
		},
		{
			Name:        "wrapError",
			Description: "Wrap an error with additional context",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Original error", Required: true},
				{Name: "message", Type: "string", Description: "Wrapper message", Required: true},
				{Name: "metadata", Type: "object", Description: "Additional metadata", Required: false},
			},
			ReturnType: "SerializableError",
		},
		{
			Name:        "createErrorWithCode",
			Description: "Create an error with a specific error code",
			Parameters: []engine.ParameterInfo{
				{Name: "code", Type: "string", Description: "Error code", Required: true},
				{Name: "message", Type: "string", Description: "Error message", Required: true},
				{Name: "metadata", Type: "object", Description: "Error metadata", Required: false},
			},
			ReturnType: "SerializableError",
		},

		// Error serialization
		{
			Name:        "errorToJSON",
			Description: "Serialize an error to JSON",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to serialize", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "errorFromJSON",
			Description: "Deserialize an error from JSON",
			Parameters: []engine.ParameterInfo{
				{Name: "json", Type: "string", Description: "JSON error representation", Required: true},
			},
			ReturnType: "SerializableError",
		},

		// Recovery strategies
		{
			Name:        "createExponentialBackoffStrategy",
			Description: "Create an exponential backoff recovery strategy",
			Parameters: []engine.ParameterInfo{
				{Name: "baseDelay", Type: "number", Description: "Base delay in milliseconds", Required: true},
				{Name: "maxDelay", Type: "number", Description: "Maximum delay in milliseconds", Required: true},
				{Name: "maxRetries", Type: "number", Description: "Maximum retry attempts", Required: true},
			},
			ReturnType: "RecoveryStrategy",
		},
		{
			Name:        "createLinearBackoffStrategy",
			Description: "Create a linear backoff recovery strategy",
			Parameters: []engine.ParameterInfo{
				{Name: "delay", Type: "number", Description: "Delay between retries in milliseconds", Required: true},
				{Name: "maxRetries", Type: "number", Description: "Maximum retry attempts", Required: true},
			},
			ReturnType: "RecoveryStrategy",
		},
		{
			Name:        "applyRecoveryStrategy",
			Description: "Apply a recovery strategy to an error",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to recover from", Required: true},
				{Name: "strategy", Type: "RecoveryStrategy", Description: "Recovery strategy", Required: true},
			},
			ReturnType: "object",
		},

		// Error categorization
		{
			Name:        "categorizeError",
			Description: "Categorize an error",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to categorize", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "registerErrorCategory",
			Description: "Register a custom error category",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Category name", Required: true},
				{Name: "config", Type: "object", Description: "Category configuration", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getErrorCategories",
			Description: "Get all registered error categories",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
		},

		// Error aggregation
		{
			Name:        "createErrorAggregator",
			Description: "Create an error aggregator",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "ErrorAggregator",
		},
		{
			Name:        "addError",
			Description: "Add an error to the aggregator",
			Parameters: []engine.ParameterInfo{
				{Name: "aggregator", Type: "ErrorAggregator", Description: "Error aggregator", Required: true},
				{Name: "error", Type: "error", Description: "Error to add", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "aggregateErrors",
			Description: "Create an aggregated error from multiple errors",
			Parameters: []engine.ParameterInfo{
				{Name: "errors", Type: "array", Description: "Array of errors", Required: true},
				{Name: "message", Type: "string", Description: "Aggregation message", Required: false},
			},
			ReturnType: "SerializableError",
		},
		{
			Name:        "getAggregatedErrors",
			Description: "Get errors from an aggregator",
			Parameters: []engine.ParameterInfo{
				{Name: "aggregator", Type: "ErrorAggregator", Description: "Error aggregator", Required: true},
			},
			ReturnType: "array",
		},

		// Error event emission
		{
			Name:        "emitErrorEvent",
			Description: "Emit an error event",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to emit", Required: true},
				{Name: "context", Type: "object", Description: "Event context", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "subscribeToErrorEvents",
			Description: "Subscribe to error events",
			Parameters: []engine.ParameterInfo{
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
				{Name: "filter", Type: "object", Description: "Event filter", Required: false},
			},
			ReturnType: "string",
		},

		// Custom error handlers
		{
			Name:        "registerErrorHandler",
			Description: "Register a custom error handler",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Handler name", Required: true},
				{Name: "handler", Type: "function", Description: "Handler function", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "applyErrorHandler",
			Description: "Apply a custom error handler",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to handle", Required: true},
				{Name: "handlerName", Type: "string", Description: "Handler name", Required: true},
			},
			ReturnType: "error",
		},

		// Error inspection
		{
			Name:        "isRetryableError",
			Description: "Check if an error is retryable",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to check", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "isFatalError",
			Description: "Check if an error is fatal",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to check", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "getErrorMetadata",
			Description: "Get error metadata",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to inspect", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getErrorStackTrace",
			Description: "Get error stack trace",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to inspect", Required: true},
			},
			ReturnType: "array",
		},

		// Error building
		{
			Name:        "createErrorBuilder",
			Description: "Create an error builder for fluent error construction",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "ErrorBuilder",
		},
		{
			Name:        "buildError",
			Description: "Build the error from the builder",
			Parameters: []engine.ParameterInfo{
				{Name: "builder", Type: "ErrorBuilder", Description: "Error builder", Required: true},
			},
			ReturnType: "SerializableError",
		},

		// Error context
		{
			Name:        "enrichError",
			Description: "Enrich an error with additional context",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to enrich", Required: true},
				{Name: "context", Type: "object", Description: "Additional context", Required: true},
			},
			ReturnType: "SerializableError",
		},
		{
			Name:        "getErrorContext",
			Description: "Get error context",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Description: "Error to inspect", Required: true},
			},
			ReturnType: "object",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *UtilErrorsBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"SerializableError": {
			GoType:     "SerializableError",
			ScriptType: "object",
		},
		"RecoveryStrategy": {
			GoType:     "RecoveryStrategy",
			ScriptType: "object",
		},
		"ErrorAggregator": {
			GoType:     "ErrorAggregator",
			ScriptType: "object",
		},
		"ErrorCategory": {
			GoType:     "ErrorCategory",
			ScriptType: "object",
		},
		"ErrorBuilder": {
			GoType:     "ErrorBuilder",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls.
func (b *UtilErrorsBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *UtilErrorsBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "errors",
			Actions:     []string{"read", "write"},
			Description: "Store and manage error data",
		},
		{
			Type:        engine.PermissionStorage,
			Resource:    "errors",
			Actions:     []string{"emit", "subscribe"},
			Description: "Emit and subscribe to error events",
		},
	}
}

// ExecuteMethod executes a bridge method
func (b *UtilErrorsBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createError":
		return b.createError(ctx, args)
	case "wrapError":
		return b.wrapError(ctx, args)
	case "createErrorWithCode":
		return b.createErrorWithCode(ctx, args)
	case "errorToJSON":
		return b.errorToJSON(ctx, args)
	case "errorFromJSON":
		return b.errorFromJSON(ctx, args)
	case "createExponentialBackoffStrategy":
		return b.createExponentialBackoffStrategy(ctx, args)
	case "createLinearBackoffStrategy":
		return b.createLinearBackoffStrategy(ctx, args)
	case "categorizeError":
		return b.categorizeError(ctx, args)
	case "registerErrorCategory":
		return b.registerErrorCategory(ctx, args)
	case "getErrorCategories":
		return b.getErrorCategories(ctx, args)
	case "createErrorAggregator":
		return b.createErrorAggregator(ctx, args)
	case "addError":
		return b.addError(ctx, args)
	case "aggregateErrors":
		return b.aggregateErrors(ctx, args)
	case "emitErrorEvent":
		return b.emitErrorEvent(ctx, args)
	case "isRetryableError":
		return b.isRetryableError(ctx, args)
	case "isFatalError":
		return b.isFatalError(ctx, args)
	case "enrichError":
		return b.enrichError(ctx, args)
	case "getErrorContext":
		return b.getErrorContext(ctx, args)
	case "createErrorBuilder":
		return b.createErrorBuilder(ctx, args)
	case "buildError":
		return b.buildError(ctx, args)
	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Method implementations

func (b *UtilErrorsBridge) createError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("createError requires message parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be string")
	}
	message := args[0].(engine.StringValue).Value()

	// Create base error
	err := errors.NewError(message)

	// Add metadata if provided
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		objFields := args[1].(engine.ObjectValue).Fields()
		metadata := make(map[string]interface{})
		for k, v := range objFields {
			metadata[k] = v.ToGo()
		}
		err = err.WithContextMap(metadata)
	}

	return engine.NewCustomValue("error", err), nil
}

func (b *UtilErrorsBridge) wrapError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrapError requires error and message parameters")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	originalErr, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be string")
	}
	message := args[1].(engine.StringValue).Value()

	// Wrap error
	wrapped := errors.Wrap(originalErr, message)

	// Add metadata if provided
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
		objFields := args[2].(engine.ObjectValue).Fields()
		metadata := make(map[string]interface{})
		for k, v := range objFields {
			metadata[k] = v.ToGo()
		}
		wrapped = wrapped.WithContextMap(metadata)
	}

	return engine.NewCustomValue("error", wrapped), nil
}

func (b *UtilErrorsBridge) createErrorWithCode(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("createErrorWithCode requires code and message parameters")
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("code must be string")
	}
	code := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("message must be string")
	}
	message := args[1].(engine.StringValue).Value()

	// Create error with code
	err := errors.NewErrorWithCode(code, message)

	// Add metadata if provided
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
		objFields := args[2].(engine.ObjectValue).Fields()
		metadata := make(map[string]interface{})
		for k, v := range objFields {
			metadata[k] = v.ToGo()
		}
		err = err.WithContextMap(metadata)
	}

	return engine.NewCustomValue("error", err), nil
}

func (b *UtilErrorsBridge) errorToJSON(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("errorToJSON requires error parameter")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	// Convert to SerializableError if needed
	var serr errors.SerializableError
	if se, ok := err.(errors.SerializableError); ok {
		serr = se
	} else {
		// Wrap in SerializableError
		serr = errors.NewError(err.Error())
	}

	// Serialize to JSON
	data, jsonErr := json.Marshal(serr)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to serialize error: %w", jsonErr)
	}

	return engine.NewStringValue(string(data)), nil
}

func (b *UtilErrorsBridge) errorFromJSON(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("errorFromJSON requires json parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("json must be string")
	}
	jsonStr := args[0].(engine.StringValue).Value()

	// Deserialize from JSON
	serr, err := errors.ErrorFromJSON([]byte(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize error: %w", err)
	}

	return engine.NewCustomValue("error", serr), nil
}

func (b *UtilErrorsBridge) createExponentialBackoffStrategy(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("createExponentialBackoffStrategy requires baseDelay, maxDelay, and maxRetries")
	}

	if args[0] == nil || args[0].Type() != engine.TypeNumber {
		return nil, fmt.Errorf("baseDelay must be number")
	}
	baseDelay := args[0].(engine.NumberValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeNumber {
		return nil, fmt.Errorf("maxDelay must be number")
	}
	maxDelay := args[1].(engine.NumberValue).Value()

	if args[2] == nil || args[2].Type() != engine.TypeNumber {
		return nil, fmt.Errorf("maxRetries must be number")
	}
	maxRetries := args[2].(engine.NumberValue).Value()

	strategy := errors.NewExponentialBackoffStrategy(
		int(maxRetries),
		time.Duration(baseDelay)*time.Millisecond,
		time.Duration(maxDelay)*time.Millisecond,
	)

	return engine.NewCustomValue("RecoveryStrategy", strategy), nil
}

func (b *UtilErrorsBridge) createLinearBackoffStrategy(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("createLinearBackoffStrategy requires delay and maxRetries")
	}

	if args[0] == nil || args[0].Type() != engine.TypeNumber {
		return nil, fmt.Errorf("delay must be number")
	}
	delay := args[0].(engine.NumberValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeNumber {
		return nil, fmt.Errorf("maxRetries must be number")
	}
	maxRetries := args[1].(engine.NumberValue).Value()

	strategy := errors.NewLinearBackoffStrategy(
		int(maxRetries),
		time.Duration(delay)*time.Millisecond,
	)

	return engine.NewCustomValue("RecoveryStrategy", strategy), nil
}

func (b *UtilErrorsBridge) categorizeError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("categorizeError requires error parameter")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	// Categorize error based on type and metadata
	category := b.categorizeErrorInternal(err)
	return engine.NewStringValue(category), nil
}

func (b *UtilErrorsBridge) registerErrorCategory(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("registerErrorCategory requires name and config")
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("name must be string")
	}
	name := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("config must be object")
	}
	configObj := args[1].(engine.ObjectValue).Fields()
	config := make(map[string]interface{})
	for k, v := range configObj {
		config[k] = v.ToGo()
	}

	// Create category from config
	category := ErrorCategory{
		Name:        name,
		Description: getStringFromMap(config, "description", ""),
		Retryable:   getBoolFromMap(config, "retryable", false),
		Fatal:       getBoolFromMap(config, "fatal", false),
	}

	b.errorCategories[name] = category
	return engine.NewNilValue(), nil
}

func (b *UtilErrorsBridge) getErrorCategories(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Return copy of categories
	categories := make(map[string]engine.ScriptValue)
	for name, cat := range b.errorCategories {
		categories[name] = engine.NewObjectValue(map[string]engine.ScriptValue{
			"name":        engine.NewStringValue(cat.Name),
			"description": engine.NewStringValue(cat.Description),
			"retryable":   engine.NewBoolValue(cat.Retryable),
			"fatal":       engine.NewBoolValue(cat.Fatal),
		})
	}
	return engine.NewObjectValue(categories), nil
}

func (b *UtilErrorsBridge) createErrorAggregator(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Create new aggregator
	aggregator := errors.NewErrorAggregator()
	return engine.NewCustomValue("ErrorAggregator", aggregator), nil
}

func (b *UtilErrorsBridge) addError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("addError requires aggregator and error")
	}

	// Get aggregator from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("aggregator must be ErrorAggregator")
	}
	aggVal := args[0].(engine.CustomValue)
	aggregator, ok := aggVal.Value().(errors.ErrorAggregator)
	if !ok {
		return nil, fmt.Errorf("aggregator must be ErrorAggregator")
	}

	// Get error from custom value
	if args[1] == nil || args[1].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	errVal := args[1].(engine.CustomValue)
	err, ok := errVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	aggregator.Add(err)
	return engine.NewNilValue(), nil
}

func (b *UtilErrorsBridge) aggregateErrors(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("aggregateErrors requires errors array")
	}
	if args[0] == nil || args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("errors must be array")
	}
	errorsArray := args[0].(engine.ArrayValue).Elements()

	// Create aggregator
	aggregator := errors.NewErrorAggregator()

	// Add all errors
	for i, e := range errorsArray {
		if e.Type() != engine.TypeCustom {
			return nil, fmt.Errorf("element at index %d is not an error", i)
		}
		customVal := e.(engine.CustomValue)
		if err, ok := customVal.Value().(error); ok {
			aggregator.Add(err)
		} else {
			return nil, fmt.Errorf("element at index %d is not an error", i)
		}
	}

	// Get message if provided
	message := "Multiple errors occurred"
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeString {
		message = args[1].(engine.StringValue).Value()
	}

	// Create aggregated error with message
	if aggregator.HasErrors() {
		// Get the base error and wrap with message
		aggErr := aggregator.Error()
		wrapped := errors.Wrap(aggErr, message)
		return engine.NewCustomValue("error", wrapped), nil
	}

	return engine.NewNilValue(), nil
}

func (b *UtilErrorsBridge) emitErrorEvent(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("emitErrorEvent requires error parameter")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	// Get context if provided
	eventContext := make(map[string]interface{})
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeObject {
		objFields := args[1].(engine.ObjectValue).Fields()
		for k, v := range objFields {
			eventContext[k] = v.ToGo()
		}
	}

	// Add error information to context
	eventContext["error"] = err.Error()
	eventContext["timestamp"] = time.Now()

	// Check if error is BaseError
	if baseErr, ok := err.(*errors.BaseError); ok {
		eventContext["code"] = baseErr.Code
		eventContext["type"] = baseErr.Type
		eventContext["metadata"] = baseErr.GetContext()
	}

	// Emit event
	if b.eventEmitter != nil {
		b.eventEmitter.EmitCustom("error.occurred", eventContext)
	}

	return engine.NewNilValue(), nil
}

func (b *UtilErrorsBridge) isRetryableError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isRetryableError requires error parameter")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	// Use go-llms helper function
	return engine.NewBoolValue(errors.IsRetryableError(err)), nil
}

func (b *UtilErrorsBridge) isFatalError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("isFatalError requires error parameter")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	// Use go-llms helper function
	return engine.NewBoolValue(errors.IsFatalError(err)), nil
}

func (b *UtilErrorsBridge) enrichError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("enrichError requires error and context")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := args[1].(engine.ObjectValue).Fields()
	context := make(map[string]interface{})
	for k, v := range contextObj {
		context[k] = v.ToGo()
	}

	// Enrich error with context - wrap if needed
	if baseErr, ok := err.(*errors.BaseError); ok {
		return engine.NewCustomValue("error", baseErr.WithContextMap(context)), nil
	}
	// Wrap and add context
	wrapped := errors.Wrap(err, "enriched error")
	return engine.NewCustomValue("error", wrapped.WithContextMap(context)), nil
}

func (b *UtilErrorsBridge) getErrorContext(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getErrorContext requires error parameter")
	}

	// Get error from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("error must be an error type")
	}
	customVal := args[0].(engine.CustomValue)
	err, ok := customVal.Value().(error)
	if !ok {
		return nil, fmt.Errorf("error must be an error type")
	}

	// Extract context from error
	context := make(map[string]engine.ScriptValue)

	// Extract context from error
	if ctx := errors.GetErrorContext(err); ctx != nil {
		for k, v := range ctx {
			// Convert each value to appropriate ScriptValue type
			switch val := v.(type) {
			case string:
				context[k] = engine.NewStringValue(val)
			case int, int32, int64, float32, float64:
				context[k] = engine.NewNumberValue(toFloat64(val))
			case bool:
				context[k] = engine.NewBoolValue(val)
			case nil:
				context[k] = engine.NewNilValue()
			case map[string]interface{}:
				// Recursively convert map
				objMap := make(map[string]engine.ScriptValue)
				for mk, mv := range val {
					objMap[mk] = convertToScriptValue(mv)
				}
				context[k] = engine.NewObjectValue(objMap)
			case []interface{}:
				// Convert array
				arr := make([]engine.ScriptValue, len(val))
				for i, item := range val {
					arr[i] = convertToScriptValue(item)
				}
				context[k] = engine.NewArrayValue(arr)
			default:
				// For unknown types, convert to string
				context[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
			}
		}
	}

	return engine.NewObjectValue(context), nil
}

func (b *UtilErrorsBridge) createErrorBuilder(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Create new error builder
	builder := &ErrorBuilder{
		err: errors.NewError(""),
	}
	return engine.NewCustomValue("ErrorBuilder", builder), nil
}

func (b *UtilErrorsBridge) buildError(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("buildError requires builder parameter")
	}

	// Get builder from custom value
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("builder must be ErrorBuilder")
	}
	customVal := args[0].(engine.CustomValue)
	builder, ok := customVal.Value().(*ErrorBuilder)
	if !ok {
		return nil, fmt.Errorf("builder must be ErrorBuilder")
	}

	return engine.NewCustomValue("error", builder.err), nil
}

// setupDefaultCategories sets up default error categories
func (b *UtilErrorsBridge) setupDefaultCategories() {
	b.errorCategories["network"] = ErrorCategory{
		Name:        "network",
		Description: "Network-related errors",
		Retryable:   true,
		Fatal:       false,
	}

	b.errorCategories["validation"] = ErrorCategory{
		Name:        "validation",
		Description: "Validation errors",
		Retryable:   false,
		Fatal:       false,
	}

	b.errorCategories["authentication"] = ErrorCategory{
		Name:        "authentication",
		Description: "Authentication errors",
		Retryable:   false,
		Fatal:       false,
	}

	b.errorCategories["authorization"] = ErrorCategory{
		Name:        "authorization",
		Description: "Authorization errors",
		Retryable:   false,
		Fatal:       false,
	}

	b.errorCategories["ratelimit"] = ErrorCategory{
		Name:        "ratelimit",
		Description: "Rate limiting errors",
		Retryable:   true,
		Fatal:       false,
	}

	b.errorCategories["system"] = ErrorCategory{
		Name:        "system",
		Description: "System errors",
		Retryable:   false,
		Fatal:       true,
	}
}

// categorizeErrorInternal categorizes an error based on its type and content
func (b *UtilErrorsBridge) categorizeErrorInternal(err error) string {
	if err == nil {
		return "unknown"
	}

	// Check error code if available
	if baseErr, ok := err.(*errors.BaseError); ok {
		code := baseErr.Code

		// Map common error codes to categories
		switch code {
		case "NETWORK_ERROR", "CONNECTION_FAILED", "TIMEOUT":
			return "network"
		case "VALIDATION_ERROR", "INVALID_INPUT":
			return "validation"
		case "UNAUTHORIZED", "AUTH_FAILED":
			return "authentication"
		case "FORBIDDEN", "ACCESS_DENIED":
			return "authorization"
		case "RATE_LIMITED", "TOO_MANY_REQUESTS":
			return "ratelimit"
		case "INTERNAL_ERROR", "SYSTEM_ERROR":
			return "system"
		}
	}

	// Check error message patterns
	errMsg := err.Error()
	if containsAny(errMsg, "network", "connection", "timeout") {
		return "network"
	}
	if containsAny(errMsg, "invalid", "validation", "required") {
		return "validation"
	}
	if containsAny(errMsg, "unauthorized", "authentication") {
		return "authentication"
	}
	if containsAny(errMsg, "forbidden", "permission", "access denied") {
		return "authorization"
	}
	if containsAny(errMsg, "rate limit", "too many requests") {
		return "ratelimit"
	}
	if containsAny(errMsg, "internal", "system", "panic") {
		return "system"
	}

	return "unknown"
}

// Helper functions
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultValue
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ErrorBuilder provides fluent error construction
type ErrorBuilder struct {
	err *errors.BaseError
}

// Helper function to convert interface{} to ScriptValue
func convertToScriptValue(v interface{}) engine.ScriptValue {
	switch val := v.(type) {
	case string:
		return engine.NewStringValue(val)
	case int, int32, int64, float32, float64:
		return engine.NewNumberValue(toFloat64(val))
	case bool:
		return engine.NewBoolValue(val)
	case nil:
		return engine.NewNilValue()
	case map[string]interface{}:
		objMap := make(map[string]engine.ScriptValue)
		for k, v := range val {
			objMap[k] = convertToScriptValue(v)
		}
		return engine.NewObjectValue(objMap)
	case []interface{}:
		arr := make([]engine.ScriptValue, len(val))
		for i, item := range val {
			arr[i] = convertToScriptValue(item)
		}
		return engine.NewArrayValue(arr)
	default:
		return engine.NewStringValue(fmt.Sprintf("%v", v))
	}
}

// Helper function to convert numeric types to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}
