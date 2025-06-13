// ABOUTME: Event system bridge provides access to go-llms event functionality for script engines
// ABOUTME: Wraps event emission, subscription, filtering, and history without reimplementation

package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	// go-llms imports for event functionality
)

// EventBridge provides script access to go-llms event functionality
type EventBridge struct {
	mu            sync.RWMutex
	initialized   bool
	subscriptions map[string]*EventSubscription // subscription ID -> subscription
	filters       map[string]*EventFilter       // filter ID -> filter
	eventHistory  []bridge.Event                // event history buffer
	maxHistory    int                           // max events to keep in history
}

// EventSubscription represents an event subscription
type EventSubscription struct {
	ID        string
	EventType bridge.EventType
	Handler   interface{} // Script function
	Filter    *EventFilter
	Active    bool
}

// EventFilter represents an event filter
type EventFilter struct {
	ID          string
	EventTypes  []bridge.EventType
	AgentIDs    []string
	ToolNames   []string
	WorkflowIDs []string
	Custom      func(bridge.Event) bool
}

// NewEventBridge creates a new event bridge
func NewEventBridge() *EventBridge {
	return &EventBridge{
		subscriptions: make(map[string]*EventSubscription),
		filters:       make(map[string]*EventFilter),
		eventHistory:  make([]bridge.Event, 0),
		maxHistory:    1000, // Default history size
	}
}

// GetID returns the bridge identifier
func (b *EventBridge) GetID() string {
	return "events"
}

// GetMetadata returns bridge metadata
func (b *EventBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "events",
		Version:     "1.0.0",
		Description: "Event system bridge wrapping go-llms event functionality",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *EventBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *EventBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Clear all subscriptions
	for id := range b.subscriptions {
		delete(b.subscriptions, id)
	}

	// Clear all filters
	for id := range b.filters {
		delete(b.filters, id)
	}

	// Clear event history
	b.eventHistory = b.eventHistory[:0]

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *EventBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *EventBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *EventBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Event emission
		{
			Name:        "emitEvent",
			Description: "Emit a generic event",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "Event", Description: "Event to emit", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "emitAgentEvent",
			Description: "Emit an agent-specific event",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "eventType", Type: "EventType", Description: "Event type", Required: true},
				{Name: "data", Type: "any", Description: "Event data", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "emitToolEvent",
			Description: "Emit a tool-specific event",
			Parameters: []engine.ParameterInfo{
				{Name: "toolName", Type: "string", Description: "Tool name", Required: true},
				{Name: "eventType", Type: "EventType", Description: "Event type", Required: true},
				{Name: "data", Type: "any", Description: "Event data", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "emitWorkflowEvent",
			Description: "Emit a workflow-specific event",
			Parameters: []engine.ParameterInfo{
				{Name: "workflowID", Type: "string", Description: "Workflow ID", Required: true},
				{Name: "eventType", Type: "EventType", Description: "Event type", Required: true},
				{Name: "data", Type: "any", Description: "Event data", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "emitStateEvent",
			Description: "Emit a state change event",
			Parameters: []engine.ParameterInfo{
				{Name: "stateID", Type: "string", Description: "State ID", Required: true},
				{Name: "changeType", Type: "string", Description: "Change type", Required: true},
				{Name: "data", Type: "any", Description: "Change data", Required: false},
			},
			ReturnType: "void",
		},
		// Event subscription
		{
			Name:        "subscribeToEvents",
			Description: "Subscribe to all events",
			Parameters: []engine.ParameterInfo{
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
				{Name: "filter", Type: "EventFilter", Description: "Optional event filter", Required: false},
			},
			ReturnType: "string", // subscription ID
		},
		{
			Name:        "subscribeToEventType",
			Description: "Subscribe to a specific event type",
			Parameters: []engine.ParameterInfo{
				{Name: "eventType", Type: "EventType", Description: "Event type to subscribe to", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
			},
			ReturnType: "string", // subscription ID
		},
		{
			Name:        "subscribeToAgentEvents",
			Description: "Subscribe to events from a specific agent",
			Parameters: []engine.ParameterInfo{
				{Name: "agentID", Type: "string", Description: "Agent ID", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
				{Name: "eventTypes", Type: "array", Description: "Optional event types to filter", Required: false},
			},
			ReturnType: "string", // subscription ID
		},
		{
			Name:        "unsubscribe",
			Description: "Unsubscribe from events",
			Parameters: []engine.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID to remove", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "unsubscribeAll",
			Description: "Remove all event subscriptions",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "listSubscriptions",
			Description: "List all active subscriptions",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "pauseSubscription",
			Description: "Pause a subscription without removing it",
			Parameters: []engine.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "resumeSubscription",
			Description: "Resume a paused subscription",
			Parameters: []engine.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID", Required: true},
			},
			ReturnType: "boolean",
		},
		// Event filtering
		{
			Name:        "createEventFilter",
			Description: "Create a new event filter",
			Parameters: []engine.ParameterInfo{
				{Name: "filterConfig", Type: "object", Description: "Filter configuration", Required: true},
			},
			ReturnType: "EventFilter",
		},
		{
			Name:        "addEventFilter",
			Description: "Add a filter to the system",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "EventFilter", Description: "Filter to add", Required: true},
			},
			ReturnType: "string", // filter ID
		},
		{
			Name:        "removeEventFilter",
			Description: "Remove an event filter",
			Parameters: []engine.ParameterInfo{
				{Name: "filterID", Type: "string", Description: "Filter ID to remove", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "listEventFilters",
			Description: "List all active filters",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "testEventFilter",
			Description: "Test if an event matches a filter",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "Event", Description: "Event to test", Required: true},
				{Name: "filter", Type: "EventFilter", Description: "Filter to test against", Required: true},
			},
			ReturnType: "boolean",
		},
		// Event history
		{
			Name:        "getEventHistory",
			Description: "Get event history",
			Parameters: []engine.ParameterInfo{
				{Name: "limit", Type: "number", Description: "Maximum events to return", Required: false},
				{Name: "filter", Type: "EventFilter", Description: "Optional filter", Required: false},
			},
			ReturnType: "array",
		},
		{
			Name:        "clearEventHistory",
			Description: "Clear event history",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "getEventCount",
			Description: "Get total number of events in history",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "EventFilter", Description: "Optional filter", Required: false},
			},
			ReturnType: "number",
		},
		{
			Name:        "queryEvents",
			Description: "Query events with complex criteria",
			Parameters: []engine.ParameterInfo{
				{Name: "query", Type: "object", Description: "Query criteria", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "setMaxHistory",
			Description: "Set maximum events to keep in history",
			Parameters: []engine.ParameterInfo{
				{Name: "max", Type: "number", Description: "Maximum events", Required: true},
			},
			ReturnType: "void",
		},
		// Event types
		{
			Name:        "getEventTypes",
			Description: "Get all available event types",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "isValidEventType",
			Description: "Check if an event type is valid",
			Parameters: []engine.ParameterInfo{
				{Name: "eventType", Type: "string", Description: "Event type to validate", Required: true},
			},
			ReturnType: "boolean",
		},
		// Event utilities
		{
			Name:        "createEvent",
			Description: "Create a new event object",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "EventType", Description: "Event type", Required: true},
				{Name: "source", Type: "string", Description: "Event source", Required: true},
				{Name: "data", Type: "any", Description: "Event data", Required: false},
			},
			ReturnType: "Event",
		},
		{
			Name:        "formatEvent",
			Description: "Format an event for display",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "Event", Description: "Event to format", Required: true},
				{Name: "format", Type: "string", Description: "Format type (json, text, etc.)", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "parseEvent",
			Description: "Parse an event from string representation",
			Parameters: []engine.ParameterInfo{
				{Name: "eventString", Type: "string", Description: "Event string", Required: true},
				{Name: "format", Type: "string", Description: "Format type", Required: false},
			},
			ReturnType: "Event",
		},
		// Streaming support
		{
			Name:        "startEventStream",
			Description: "Start streaming events to a handler",
			Parameters: []engine.ParameterInfo{
				{Name: "handler", Type: "function", Description: "Stream handler", Required: true},
				{Name: "options", Type: "object", Description: "Stream options", Required: false},
			},
			ReturnType: "string", // stream ID
		},
		{
			Name:        "stopEventStream",
			Description: "Stop an event stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "isStreamActive",
			Description: "Check if a stream is active",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "getStreamStats",
			Description: "Get statistics for an event stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
			},
			ReturnType: "object",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *EventBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Event": {
			GoType:     "Event",
			ScriptType: "object",
		},
		"EventType": {
			GoType:     "EventType",
			ScriptType: "string",
		},
		"EventFilter": {
			GoType:     "*EventFilter",
			ScriptType: "object",
		},
		"EventSubscription": {
			GoType:     "*EventSubscription",
			ScriptType: "object",
		},
		"EventHistory": {
			GoType:     "[]Event",
			ScriptType: "array",
		},
		"EventHandler": {
			GoType:     "func(Event)",
			ScriptType: "function",
		},
		"EventMetadata": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
		},
		"AgentEvent": {
			GoType:     "Event",
			ScriptType: "object",
		},
		"ToolEvent": {
			GoType:     "Event",
			ScriptType: "object",
		},
		"WorkflowEvent": {
			GoType:     "Event",
			ScriptType: "object",
		},
		"StateEvent": {
			GoType:     "Event",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *EventBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions
func (b *EventBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionProcess,
			Resource:    "events",
			Actions:     []string{"emit", "subscribe", "filter"},
			Description: "Access to event system",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "history",
			Actions:     []string{"allocate", "read"},
			Description: "Memory for event history",
		},
	}
}

// Helper methods for event management

// generateSubscriptionID generates a unique subscription ID
//
//nolint:unused // will be used when implementing subscription methods
func (b *EventBridge) generateSubscriptionID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return fmt.Sprintf("sub_%d", len(b.subscriptions)+1)
}

// generateFilterID generates a unique filter ID
//
//nolint:unused // will be used when implementing filter methods
func (b *EventBridge) generateFilterID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return fmt.Sprintf("filter_%d", len(b.filters)+1)
}

// addToHistory adds an event to the history buffer
//
//nolint:unused // will be used when implementing event methods
func (b *EventBridge) addToHistory(event bridge.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.eventHistory = append(b.eventHistory, event)

	// Trim history if it exceeds max size
	if len(b.eventHistory) > b.maxHistory {
		b.eventHistory = b.eventHistory[len(b.eventHistory)-b.maxHistory:]
	}
}

// matchesFilter checks if an event matches a filter
//
//nolint:unused // will be used when implementing filter methods
func (b *EventBridge) matchesFilter(event bridge.Event, filter *EventFilter) bool {
	// Check event type filter
	if len(filter.EventTypes) > 0 {
		matched := false
		for _, eventType := range filter.EventTypes {
			if event.Type == eventType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check agent ID filter
	if len(filter.AgentIDs) > 0 && event.AgentID != "" {
		matched := false
		for _, agentID := range filter.AgentIDs {
			if event.AgentID == agentID {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check custom filter
	if filter.Custom != nil {
		return filter.Custom(event)
	}

	return true
}

// notifySubscribers notifies all matching subscribers of an event
//
//nolint:unused // will be used when implementing event emission
func (b *EventBridge) notifySubscribers(event bridge.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscriptions {
		if !sub.Active {
			continue
		}

		// Check if subscription matches the event
		if sub.EventType != "" && sub.EventType != event.Type {
			continue
		}

		// Check filter if present
		if sub.Filter != nil && !b.matchesFilter(event, sub.Filter) {
			continue
		}

		// Notify subscriber (would invoke script handler in real implementation)
		// This is where we'd bridge to the script engine's function call mechanism
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *EventBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "subscribeToEventType":
		if len(args) < 2 {
			return nil, fmt.Errorf("subscribeToEventType requires eventType and handler parameters")
		}
		eventType, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("eventType must be string")
		}

		// Create subscription
		subID := b.generateSubscriptionID()
		sub := &EventSubscription{
			ID:        subID,
			EventType: bridge.EventType(eventType),
			Handler:   args[1], // Store the handler function
			Active:    true,
		}

		b.subscriptions[subID] = sub
		return subID, nil

	case "unsubscribe":
		if len(args) < 1 {
			return nil, fmt.Errorf("unsubscribe requires subscriptionID parameter")
		}
		subID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("subscriptionID must be string")
		}

		if sub, exists := b.subscriptions[subID]; exists {
			sub.Active = false
			delete(b.subscriptions, subID)
		}
		return nil, nil

	case "getEventHistory":
		// Return copy of event history
		history := make([]map[string]interface{}, len(b.eventHistory))
		for i, event := range b.eventHistory {
			history[i] = map[string]interface{}{
				"id":        event.ID,
				"type":      string(event.Type),
				"timestamp": event.Timestamp,
				"agentID":   event.AgentID,
				"agentName": event.AgentName,
				"data":      event.Data,
				"metadata":  event.Metadata,
			}
		}
		return history, nil

	case "clearEventHistory":
		b.eventHistory = b.eventHistory[:0]
		return nil, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
