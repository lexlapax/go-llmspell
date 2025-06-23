// ABOUTME: Tracing bridge for go-llms distributed tracing infrastructure
// ABOUTME: Provides script-accessible tracing capabilities with OpenTelemetry-compatible interfaces

package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	// go-llms imports for tracing functionality
	"github.com/lexlapax/go-llms/pkg/agent/core"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/bridge/types"
)

// TracingBridge provides script access to go-llms distributed tracing.
// It manages distributed tracing spans and OpenTelemetry-compatible
// operations. The bridge supports span creation, attribute management,
// error recording, and lifecycle hooks for comprehensive application
// tracing across service boundaries.
type TracingBridge struct {
	initialized bool
	tracers     map[string]core.Tracer
	spans       map[string]core.Span
	hooks       map[string]interface{}
	mu          sync.RWMutex
}

// NewTracingBridge creates a new tracing bridge.
// Returns an initialized bridge with empty tracer and span registries
// ready for distributed tracing operations.
func NewTracingBridge() *TracingBridge {
	return &TracingBridge{
		tracers: make(map[string]core.Tracer),
		spans:   make(map[string]core.Span),
		hooks:   make(map[string]interface{}),
	}
}

// GetID returns the bridge identifier.
// Always returns "tracing" for this bridge.
func (tb *TracingBridge) GetID() string {
	return "tracing"
}

// GetMetadata returns bridge metadata.
// Provides comprehensive information about the tracing bridge
// including version, dependencies, and OpenTelemetry compatibility.
func (tb *TracingBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:         "tracing",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms distributed tracing and OpenTelemetry-compatible span management",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/agent/core"},
	}
}

// Initialize sets up the tracing bridge.
// Prepares the bridge for span creation and management.
// Returns an error if initialization fails.
func (tb *TracingBridge) Initialize(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup.
// Ends all active spans and clears registries.
// Ensures clean shutdown of tracing resources.
func (tb *TracingBridge) Cleanup(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// End all active spans
	for spanID, span := range tb.spans {
		if span.IsRecording() {
			span.End()
		}
		delete(tb.spans, spanID)
	}

	// Clear tracers and hooks
	tb.tracers = make(map[string]core.Tracer)
	tb.hooks = make(map[string]interface{})
	tb.initialized = false

	return nil
}

// IsInitialized returns initialization status.
// Thread-safe check for bridge initialization state.
// Returns true if the bridge has been initialized.
func (tb *TracingBridge) IsInitialized() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// Enables the engine to access distributed tracing functionality.
// Returns an error if registration fails.
func (tb *TracingBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns available bridge methods.
// Provides comprehensive tracing operations including span management,
// attribute setting, error recording, and lifecycle hooks.
func (tb *TracingBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		{
			Name:        "createTracer",
			Description: "Create a new tracer for distributed tracing",
			Parameters: []types.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Tracer name"},
			},
			ReturnType: "object",
			Examples:   []string{"createTracer('my-app-tracer')"},
		},
		{
			Name:        "startSpan",
			Description: "Start a new tracing span",
			Parameters: []types.ParameterInfo{
				{Name: "tracerID", Type: "string", Required: true, Description: "Tracer identifier"},
				{Name: "name", Type: "string", Required: true, Description: "Span operation name"},
				{Name: "attributes", Type: "object", Required: false, Description: "Initial span attributes"},
			},
			ReturnType: "object",
			Examples:   []string{"startSpan(tracerID, 'http-request', {method: 'GET', url: '/api/v1/users'})"},
		},
		{
			Name:        "endSpan",
			Description: "End a tracing span",
			Parameters: []types.ParameterInfo{
				{Name: "spanID", Type: "string", Required: true, Description: "Span identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"endSpan(spanID)"},
		},
		{
			Name:        "setSpanAttributes",
			Description: "Set attributes on a span",
			Parameters: []types.ParameterInfo{
				{Name: "spanID", Type: "string", Required: true, Description: "Span identifier"},
				{Name: "attributes", Type: "object", Required: true, Description: "Key-value attributes"},
			},
			ReturnType: "void",
			Examples:   []string{"setSpanAttributes(spanID, {user_id: '123', operation: 'create'})"},
		},
		{
			Name:        "recordSpanError",
			Description: "Record an error on a span",
			Parameters: []types.ParameterInfo{
				{Name: "spanID", Type: "string", Required: true, Description: "Span identifier"},
				{Name: "error", Type: "string", Required: true, Description: "Error message"},
			},
			ReturnType: "void",
			Examples:   []string{"recordSpanError(spanID, 'Database connection failed')"},
		},
		{
			Name:        "setSpanStatus",
			Description: "Set span status",
			Parameters: []types.ParameterInfo{
				{Name: "spanID", Type: "string", Required: true, Description: "Span identifier"},
				{Name: "status", Type: "string", Required: true, Description: "Status code: 'ok', 'error', 'unset'"},
				{Name: "description", Type: "string", Required: false, Description: "Status description"},
			},
			ReturnType: "void",
			Examples:   []string{"setSpanStatus(spanID, 'error', 'Operation failed')"},
		},
		{
			Name:        "createAgentTracingHook",
			Description: "Create tracing hook for agent lifecycle",
			Parameters: []types.ParameterInfo{
				{Name: "tracerName", Type: "string", Required: true, Description: "Tracer name"},
			},
			ReturnType: "object",
			Examples:   []string{"createAgentTracingHook('agent-tracer')"},
		},
		{
			Name:        "createToolCallTracingHook",
			Description: "Create tracing hook for tool calls",
			Parameters: []types.ParameterInfo{
				{Name: "tracerName", Type: "string", Required: true, Description: "Tracer name"},
			},
			ReturnType: "object",
			Examples:   []string{"createToolCallTracingHook('tool-tracer')"},
		},
		{
			Name:        "createEventTracingHook",
			Description: "Create tracing hook for events",
			Parameters: []types.ParameterInfo{
				{Name: "tracerName", Type: "string", Required: true, Description: "Tracer name"},
			},
			ReturnType: "object",
			Examples:   []string{"createEventTracingHook('event-tracer')"},
		},
		{
			Name:        "createCompositeTracingHook",
			Description: "Create comprehensive tracing hook",
			Parameters: []types.ParameterInfo{
				{Name: "tracerName", Type: "string", Required: true, Description: "Tracer name"},
			},
			ReturnType: "object",
			Examples:   []string{"createCompositeTracingHook('app-tracer')"},
		},
		{
			Name:        "spanFromContext",
			Description: "Get current span from context",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "object",
			Examples:    []string{"spanFromContext()"},
		},
	}
}

// ValidateMethod validates method calls.
// Checks that the method exists and has the required number of arguments.
// Returns an error if the bridge is not initialized or validation fails.
func (tb *TracingBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	if !tb.IsInitialized() {
		return fmt.Errorf("tracing bridge not initialized")
	}

	methods := tb.Methods()
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

// ExecuteMethod executes a bridge method.
// Routes method calls to their implementations and handles return value conversion.
// Returns an error if the method is unknown or execution fails.
func (tb *TracingBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	switch name {
	case "createTracer":
		result, err := tb.createTracer(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	case "startSpan":
		result, err := tb.startSpan(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	case "endSpan":
		err := tb.endSpan(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewNilValue(), nil
	case "setSpanAttributes":
		err := tb.setSpanAttributes(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewNilValue(), nil
	case "recordSpanError":
		err := tb.recordSpanError(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewNilValue(), nil
	case "setSpanStatus":
		err := tb.setSpanStatus(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewNilValue(), nil
	case "createAgentTracingHook":
		result, err := tb.createAgentTracingHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	case "createToolCallTracingHook":
		result, err := tb.createToolCallTracingHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	case "createEventTracingHook":
		result, err := tb.createEventTracingHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	case "createCompositeTracingHook":
		result, err := tb.createCompositeTracingHook(ctx, args)
		if err != nil {
			return nil, err
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	case "spanFromContext":
		result, err := tb.spanFromContext(ctx, args)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return types.NewNilValue(), nil
		}
		return types.NewObjectValue(result.(map[string]types.ScriptValue)), nil
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// TypeMappings returns type conversion mappings.
// Maps go-llms tracing types to script-compatible representations
// for seamless integration with script engines.
func (tb *TracingBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
		"tracer": {
			GoType:     "core.Tracer",
			ScriptType: "object",
			Converter:  "tracerConverter",
			Metadata:   map[string]interface{}{"description": "Distributed tracer instance"},
		},
		"span": {
			GoType:     "core.Span",
			ScriptType: "object",
			Converter:  "spanConverter",
			Metadata:   map[string]interface{}{"description": "Tracing span instance"},
		},
		"attribute": {
			GoType:     "core.Attribute",
			ScriptType: "object",
			Converter:  "attributeConverter",
			Metadata:   map[string]interface{}{"description": "Span attribute key-value pair"},
		},
		"status_code": {
			GoType:     "core.StatusCode",
			ScriptType: "string",
			Converter:  "statusCodeConverter",
			Metadata:   map[string]interface{}{"description": "Span status code"},
		},
	}
}

// RequiredPermissions returns required permissions.
// Specifies the permissions needed for tracing operations including
// span creation, modification, and hook management.
func (tb *TracingBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionMemory,
			Resource:    "tracing.tracers",
			Actions:     []string{"create", "list"},
			Description: "Create and list tracers",
		},
		{
			Type:        types.PermissionMemory,
			Resource:    "tracing.spans",
			Actions:     []string{"create", "modify", "end"},
			Description: "Create, modify, and end spans",
		},
		{
			Type:        types.PermissionProcess,
			Resource:    "tracing.hooks",
			Actions:     []string{"create"},
			Description: "Create tracing hooks",
		},
	}
}

// Bridge method implementations

// createTracer creates a new tracer.
// Tracers are the entry point for creating spans in distributed tracing.
// Returns a tracer object with ID for span creation operations.
func (tb *TracingBridge) createTracer(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if err := tb.ValidateMethod("createTracer", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("tracer name must be a string")
	}
	name := args[0].(types.StringValue).Value()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Create a new tracer (using NoOpTracer for now, can be enhanced with real implementations)
	tracer := &core.NoOpTracer{}
	tracerID := fmt.Sprintf("tracer-%s-%d", name, time.Now().UnixNano())
	tb.tracers[tracerID] = tracer

	return map[string]types.ScriptValue{
		"id":      types.NewStringValue(tracerID),
		"name":    types.NewStringValue(name),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// startSpan starts a new tracing span.
// Creates a span within a tracer for tracking a specific operation.
// Supports optional attributes for span initialization.
func (tb *TracingBridge) startSpan(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if err := tb.ValidateMethod("startSpan", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("tracer ID must be a string")
	}
	tracerID := args[0].(types.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != types.TypeString {
		return nil, fmt.Errorf("span name must be a string")
	}
	name := args[1].(types.StringValue).Value()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	tracer, exists := tb.tracers[tracerID]
	if !exists {
		return nil, fmt.Errorf("tracer not found: %s", tracerID)
	}

	// Build span options
	var opts []core.SpanOption
	if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeObject {
		if objValue, ok := args[2].(types.ObjectValue); ok {
			attributes := objValue.Fields()
			var attrs []core.Attribute
			for key, value := range attributes {
				attrs = append(attrs, core.Attribute{Key: key, Value: value.ToGo()})
			}
			opts = append(opts, core.WithAttributes(attrs...))
		}
	}

	// Start the span
	_, span := tracer.Start(ctx, name, opts...)
	spanID := fmt.Sprintf("span-%s-%d", name, time.Now().UnixNano())
	tb.spans[spanID] = span

	return map[string]types.ScriptValue{
		"id":      types.NewStringValue(spanID),
		"name":    types.NewStringValue(name),
		"tracer":  types.NewStringValue(tracerID),
		"started": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// endSpan ends a tracing span.
// Marks the completion of a traced operation.
// Removes the span from active tracking after ending.
func (tb *TracingBridge) endSpan(ctx context.Context, args []types.ScriptValue) error {
	if err := tb.ValidateMethod("endSpan", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("span ID must be a string")
	}
	spanID := args[0].(types.StringValue).Value()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	span, exists := tb.spans[spanID]
	if !exists {
		return fmt.Errorf("span not found: %s", spanID)
	}

	span.End()
	delete(tb.spans, spanID)

	return nil
}

// setSpanAttributes sets attributes on a span.
// Adds or updates key-value pairs on an active span.
// Useful for adding context during operation execution.
func (tb *TracingBridge) setSpanAttributes(ctx context.Context, args []types.ScriptValue) error {
	if err := tb.ValidateMethod("setSpanAttributes", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("span ID must be a string")
	}
	spanID := args[0].(types.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != types.TypeObject {
		return fmt.Errorf("attributes must be an object")
	}
	attributes := args[1].(types.ObjectValue).Fields()

	tb.mu.RLock()
	span, exists := tb.spans[spanID]
	tb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("span not found: %s", spanID)
	}

	// Convert to go-llms attributes
	var attrs []core.Attribute
	for key, value := range attributes {
		attrs = append(attrs, core.Attribute{Key: key, Value: value.ToGo()})
	}

	span.SetAttributes(attrs...)

	return nil
}

// recordSpanError records an error on a span.
// Marks the span with an error for debugging and monitoring.
// The error is recorded as an event with error details.
func (tb *TracingBridge) recordSpanError(ctx context.Context, args []types.ScriptValue) error {
	if err := tb.ValidateMethod("recordSpanError", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("span ID must be a string")
	}
	spanID := args[0].(types.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != types.TypeString {
		return fmt.Errorf("error message must be a string")
	}
	errorMsg := args[1].(types.StringValue).Value()

	tb.mu.RLock()
	span, exists := tb.spans[spanID]
	tb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("span not found: %s", spanID)
	}

	span.RecordError(fmt.Errorf("%s", errorMsg))

	return nil
}

// setSpanStatus sets span status.
// Updates the span's status code (ok, error, unset) with optional description.
// Status indicates the overall success or failure of the operation.
func (tb *TracingBridge) setSpanStatus(ctx context.Context, args []types.ScriptValue) error {
	if err := tb.ValidateMethod("setSpanStatus", args); err != nil {
		return err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return fmt.Errorf("span ID must be a string")
	}
	spanID := args[0].(types.StringValue).Value()

	if len(args) < 2 || args[1] == nil || args[1].Type() != types.TypeString {
		return fmt.Errorf("status must be a string")
	}
	statusStr := args[1].(types.StringValue).Value()

	// Convert string to status code
	var status core.StatusCode
	switch statusStr {
	case "ok":
		status = core.StatusCodeOk
	case "error":
		status = core.StatusCodeError
	case "unset":
		status = core.StatusCodeUnset
	default:
		return fmt.Errorf("invalid status code: %s", statusStr)
	}

	description := ""
	if len(args) > 2 && args[2] != nil && args[2].Type() == types.TypeString {
		description = args[2].(types.StringValue).Value()
	}

	tb.mu.RLock()
	span, exists := tb.spans[spanID]
	tb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("span not found: %s", spanID)
	}

	span.SetStatus(status, description)

	return nil
}

// createAgentTracingHook creates a tracing hook for agent lifecycle.
// Automatically traces agent initialization, execution, and cleanup.
// Returns a hook object that can be attached to agents.
func (tb *TracingBridge) createAgentTracingHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if err := tb.ValidateMethod("createAgentTracingHook", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("tracer name must be a string")
	}
	tracerName := args[0].(types.StringValue).Value()

	// Create a tracer for the hook
	tracer := &core.NoOpTracer{}
	hook := core.NewTracingHook(tracerName, tracer)

	hookID := fmt.Sprintf("agent-hook-%s-%d", tracerName, time.Now().UnixNano())

	tb.mu.Lock()
	tb.hooks[hookID] = hook
	tb.mu.Unlock()

	return map[string]types.ScriptValue{
		"id":      types.NewStringValue(hookID),
		"type":    types.NewStringValue("agent"),
		"tracer":  types.NewStringValue(tracerName),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// createToolCallTracingHook creates a tracing hook for tool calls.
// Automatically traces tool invocations with input/output details.
// Returns a hook object for comprehensive tool monitoring.
func (tb *TracingBridge) createToolCallTracingHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if err := tb.ValidateMethod("createToolCallTracingHook", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("tracer name must be a string")
	}
	tracerName := args[0].(types.StringValue).Value()

	tracer := &core.NoOpTracer{}
	hook := core.NewToolCallTracingHook(tracerName, tracer)

	hookID := fmt.Sprintf("tool-hook-%s-%d", tracerName, time.Now().UnixNano())

	tb.mu.Lock()
	tb.hooks[hookID] = hook
	tb.mu.Unlock()

	return map[string]types.ScriptValue{
		"id":      types.NewStringValue(hookID),
		"type":    types.NewStringValue("tool_call"),
		"tracer":  types.NewStringValue(tracerName),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// createEventTracingHook creates a tracing hook for events
func (tb *TracingBridge) createEventTracingHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if err := tb.ValidateMethod("createEventTracingHook", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("tracer name must be a string")
	}
	tracerName := args[0].(types.StringValue).Value()

	tracer := &core.NoOpTracer{}
	hook := core.NewEventTracingHook(tracerName, tracer)

	hookID := fmt.Sprintf("event-hook-%s-%d", tracerName, time.Now().UnixNano())

	tb.mu.Lock()
	tb.hooks[hookID] = hook
	tb.mu.Unlock()

	return map[string]types.ScriptValue{
		"id":      types.NewStringValue(hookID),
		"type":    types.NewStringValue("event"),
		"tracer":  types.NewStringValue(tracerName),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// createCompositeTracingHook creates a comprehensive tracing hook
func (tb *TracingBridge) createCompositeTracingHook(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	if err := tb.ValidateMethod("createCompositeTracingHook", args); err != nil {
		return nil, err
	}

	if len(args) < 1 || args[0] == nil || args[0].Type() != types.TypeString {
		return nil, fmt.Errorf("tracer name must be a string")
	}
	tracerName := args[0].(types.StringValue).Value()

	tracer := &core.NoOpTracer{}
	hook := core.NewCompositeTracingHook(tracerName, tracer)

	hookID := fmt.Sprintf("composite-hook-%s-%d", tracerName, time.Now().UnixNano())

	tb.mu.Lock()
	tb.hooks[hookID] = hook
	tb.mu.Unlock()

	return map[string]types.ScriptValue{
		"id":      types.NewStringValue(hookID),
		"type":    types.NewStringValue("composite"),
		"tracer":  types.NewStringValue(tracerName),
		"created": types.NewStringValue(time.Now().Format(time.RFC3339)),
	}, nil
}

// spanFromContext gets the current span from context.
// Retrieves the active span from the current execution context.
// Returns nil if no span is currently active.
func (tb *TracingBridge) spanFromContext(ctx context.Context, args []types.ScriptValue) (interface{}, error) {
	span := core.SpanFromContext(ctx)
	if span == nil {
		return nil, nil
	}

	// Find span ID in our registry
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	for spanID, registeredSpan := range tb.spans {
		if registeredSpan == span {
			return map[string]types.ScriptValue{
				"id":        types.NewStringValue(spanID),
				"recording": types.NewBoolValue(span.IsRecording()),
			}, nil
		}
	}

	// Span exists in context but not in our registry
	return map[string]types.ScriptValue{
		"recording": types.NewBoolValue(span.IsRecording()),
		"external":  types.NewBoolValue(true),
	}, nil
}
