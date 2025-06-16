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
func (b *UtilErrorsBridge) ValidateMethod(name string, args []interface{}) error {
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
func (b *UtilErrorsBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createError":
		if len(args) < 1 {
			return nil, fmt.Errorf("createError requires message parameter")
		}
		message, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("message must be string")
		}

		// Create base error
		err := errors.NewError(message)

		// Add metadata if provided
		if len(args) > 1 && args[1] != nil {
			if metadata, ok := args[1].(map[string]interface{}); ok {
				err = err.WithContextMap(metadata)
			}
		}

		return err, nil

	case "wrapError":
		if len(args) < 2 {
			return nil, fmt.Errorf("wrapError requires error and message parameters")
		}
		originalErr, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}
		message, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("message must be string")
		}

		// Wrap error
		wrapped := errors.Wrap(originalErr, message)

		// Add metadata if provided
		if len(args) > 2 && args[2] != nil {
			if metadata, ok := args[2].(map[string]interface{}); ok {
				wrapped = wrapped.WithContextMap(metadata)
			}
		}

		return wrapped, nil

	case "createErrorWithCode":
		if len(args) < 2 {
			return nil, fmt.Errorf("createErrorWithCode requires code and message parameters")
		}
		code, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("code must be string")
		}
		message, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("message must be string")
		}

		// Create error with code
		err := errors.NewErrorWithCode(code, message)

		// Add metadata if provided
		if len(args) > 2 && args[2] != nil {
			if metadata, ok := args[2].(map[string]interface{}); ok {
				err = err.WithContextMap(metadata)
			}
		}

		return err, nil

	case "errorToJSON":
		if len(args) < 1 {
			return nil, fmt.Errorf("errorToJSON requires error parameter")
		}
		err, ok := args[0].(error)
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

		return string(data), nil

	case "errorFromJSON":
		if len(args) < 1 {
			return nil, fmt.Errorf("errorFromJSON requires json parameter")
		}
		jsonStr, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("json must be string")
		}

		// Deserialize from JSON
		serr, err := errors.ErrorFromJSON([]byte(jsonStr))
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize error: %w", err)
		}

		return serr, nil

	case "createExponentialBackoffStrategy":
		if len(args) < 3 {
			return nil, fmt.Errorf("createExponentialBackoffStrategy requires baseDelay, maxDelay, and maxRetries")
		}

		baseDelay, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("baseDelay must be number")
		}
		maxDelay, ok := args[1].(float64)
		if !ok {
			return nil, fmt.Errorf("maxDelay must be number")
		}
		maxRetries, ok := args[2].(float64)
		if !ok {
			return nil, fmt.Errorf("maxRetries must be number")
		}

		strategy := errors.NewExponentialBackoffStrategy(
			int(maxRetries),
			time.Duration(baseDelay)*time.Millisecond,
			time.Duration(maxDelay)*time.Millisecond,
		)

		return strategy, nil

	case "createLinearBackoffStrategy":
		if len(args) < 2 {
			return nil, fmt.Errorf("createLinearBackoffStrategy requires delay and maxRetries")
		}

		delay, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("delay must be number")
		}
		maxRetries, ok := args[1].(float64)
		if !ok {
			return nil, fmt.Errorf("maxRetries must be number")
		}

		strategy := errors.NewLinearBackoffStrategy(
			int(maxRetries),
			time.Duration(delay)*time.Millisecond,
		)

		return strategy, nil

	case "categorizeError":
		if len(args) < 1 {
			return nil, fmt.Errorf("categorizeError requires error parameter")
		}
		err, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}

		// Categorize error based on type and metadata
		category := b.categorizeError(err)
		return category, nil

	case "registerErrorCategory":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerErrorCategory requires name and config")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}
		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Create category from config
		category := ErrorCategory{
			Name:        name,
			Description: getStringFromMap(config, "description", ""),
			Retryable:   getBoolFromMap(config, "retryable", false),
			Fatal:       getBoolFromMap(config, "fatal", false),
		}

		b.errorCategories[name] = category
		return nil, nil

	case "getErrorCategories":
		// Return copy of categories
		categories := make(map[string]interface{})
		for name, cat := range b.errorCategories {
			categories[name] = map[string]interface{}{
				"name":        cat.Name,
				"description": cat.Description,
				"retryable":   cat.Retryable,
				"fatal":       cat.Fatal,
			}
		}
		return categories, nil

	case "createErrorAggregator":
		// Create new aggregator
		aggregator := errors.NewErrorAggregator()
		return aggregator, nil

	case "addError":
		if len(args) < 2 {
			return nil, fmt.Errorf("addError requires aggregator and error")
		}
		aggregator, ok := args[0].(errors.ErrorAggregator)
		if !ok {
			return nil, fmt.Errorf("aggregator must be ErrorAggregator")
		}
		err, ok := args[1].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}

		aggregator.Add(err)
		return nil, nil

	case "aggregateErrors":
		if len(args) < 1 {
			return nil, fmt.Errorf("aggregateErrors requires errors array")
		}
		errorsArray, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("errors must be array")
		}

		// Create aggregator
		aggregator := errors.NewErrorAggregator()

		// Add all errors
		for i, e := range errorsArray {
			if err, ok := e.(error); ok {
				aggregator.Add(err)
			} else {
				return nil, fmt.Errorf("element at index %d is not an error", i)
			}
		}

		// Get message if provided
		message := "Multiple errors occurred"
		if len(args) > 1 && args[1] != nil {
			if msg, ok := args[1].(string); ok {
				message = msg
			}
		}

		// Create aggregated error with message
		if aggregator.HasErrors() {
			// Get the base error and wrap with message
			aggErr := aggregator.Error()
			wrapped := errors.Wrap(aggErr, message)
			return wrapped, nil
		}

		return nil, nil

	case "emitErrorEvent":
		if len(args) < 1 {
			return nil, fmt.Errorf("emitErrorEvent requires error parameter")
		}
		err, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}

		// Get context if provided
		eventContext := make(map[string]interface{})
		if len(args) > 1 && args[1] != nil {
			if ctx, ok := args[1].(map[string]interface{}); ok {
				eventContext = ctx
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

		return nil, nil

	case "isRetryableError":
		if len(args) < 1 {
			return nil, fmt.Errorf("isRetryableError requires error parameter")
		}
		err, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}

		// Use go-llms helper function
		return errors.IsRetryableError(err), nil

	case "isFatalError":
		if len(args) < 1 {
			return nil, fmt.Errorf("isFatalError requires error parameter")
		}
		err, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}

		// Use go-llms helper function
		return errors.IsFatalError(err), nil

	case "enrichError":
		if len(args) < 2 {
			return nil, fmt.Errorf("enrichError requires error and context")
		}
		err, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}
		context, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}

		// Enrich error with context - wrap if needed
		if baseErr, ok := err.(*errors.BaseError); ok {
			return baseErr.WithContextMap(context), nil
		}
		// Wrap and add context
		wrapped := errors.Wrap(err, "enriched error")
		return wrapped.WithContextMap(context), nil

	case "getErrorContext":
		if len(args) < 1 {
			return nil, fmt.Errorf("getErrorContext requires error parameter")
		}
		err, ok := args[0].(error)
		if !ok {
			return nil, fmt.Errorf("error must be an error type")
		}

		// Extract context from error
		context := make(map[string]interface{})

		// Extract context from error
		if ctx := errors.GetErrorContext(err); ctx != nil {
			for k, v := range ctx {
				context[k] = v
			}
		}

		return context, nil

	case "createErrorBuilder":
		// Create new error builder
		return &ErrorBuilder{
			err: errors.NewError(""),
		}, nil

	case "buildError":
		if len(args) < 1 {
			return nil, fmt.Errorf("buildError requires builder parameter")
		}
		builder, ok := args[0].(*ErrorBuilder)
		if !ok {
			return nil, fmt.Errorf("builder must be ErrorBuilder")
		}
		return builder.err, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
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

// categorizeError categorizes an error based on its type and content
func (b *UtilErrorsBridge) categorizeError(err error) string {
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
