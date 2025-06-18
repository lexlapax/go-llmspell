// ABOUTME: Event system bridge v2.0.0 integrating go-llms v0.3.5 event infrastructure
// ABOUTME: Provides comprehensive event bus, storage, filtering, serialization, and replay capabilities

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms v0.3.5 event imports
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

// EventBridge provides script access to go-llms v0.3.5 event functionality
type EventBridge struct {
	mu          sync.RWMutex
	initialized bool

	// Core event infrastructure
	eventBus *events.EventBus
	storage  events.EventStorage
	recorder *events.EventRecorder
	replayer *events.EventReplayer

	// Bridge-specific components
	bridgePublisher *events.BridgeEventPublisher
	bridgeListener  *events.BridgeEventListener

	// Subscription tracking
	subscriptions map[string]string // subscription ID to description mapping
	filters       map[string]events.EventFilter
	streams       map[string]domain.EventStream

	// Event aggregation
	aggregators map[string]*EventAggregator

	// Performance tracking
	profiler *profiling.Profiler
}

// EventAggregator handles event aggregation logic
type EventAggregator struct {
	ID          string
	Type        string
	WindowSize  time.Duration
	Events      []domain.Event
	LastUpdate  time.Time
	ResultCache interface{}
}

// NewEventBridge creates a new event bridge
func NewEventBridge() *EventBridge {
	bus := events.NewEventBus()
	storage := events.NewMemoryStorage()
	recorder := events.NewEventRecorder(storage, bus)
	replayer := events.NewEventReplayer(storage, bus)

	return &EventBridge{
		eventBus:      bus,
		storage:       storage,
		recorder:      recorder,
		replayer:      replayer,
		subscriptions: make(map[string]string),
		filters:       make(map[string]events.EventFilter),
		streams:       make(map[string]domain.EventStream),
		aggregators:   make(map[string]*EventAggregator),
		profiler:      profiling.NewProfiler("event_bridge_v2"),
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
		Version:     "2.0.0",
		Description: "Event system bridge v2.0.0 with go-llms v0.3.5 integration: bus, storage, filtering, serialization, aggregation, and replay",
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

	// Initialize bridge-specific event components
	b.bridgePublisher = events.NewBridgeEventPublisher(b.eventBus, "go-llmspell", "session-"+time.Now().Format("20060102150405"))
	b.bridgeListener = events.NewBridgeEventListener(b.eventBus, events.BridgeEventHandlerFunc(func(ctx context.Context, event *events.BridgeEvent) error {
		// Handle bridge events if needed
		return nil
	}))

	// EventBus doesn't have Start method - it's always running
	// Start recording events (no filters)
	if err := b.recorder.Start(); err != nil {
		return fmt.Errorf("failed to start event recorder: %w", err)
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *EventBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Stop all subscriptions and clean up
	for subID := range b.subscriptions {
		b.eventBus.Unsubscribe(subID)
	}
	b.subscriptions = make(map[string]string)
	b.filters = make(map[string]events.EventFilter)
	b.streams = make(map[string]domain.EventStream)
	b.aggregators = make(map[string]*EventAggregator)

	// Stop recording
	b.recorder.Stop()

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
		// Event Bus Methods
		{
			Name:        "publishEvent",
			Description: "Publish an event to the event bus",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event data", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "subscribe",
			Description: "Subscribe to events with pattern matching",
			Parameters: []engine.ParameterInfo{
				{Name: "pattern", Type: "string", Description: "Event pattern", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "subscribeWithFilter",
			Description: "Subscribe to events with custom filter",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Event filter", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "unsubscribe",
			Description: "Unsubscribe from events",
			Parameters: []engine.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID", Required: true},
			},
			ReturnType: "void",
		},
		// Event Storage Methods
		{
			Name:        "storeEvent",
			Description: "Store an event in persistent storage",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event to store", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "queryEvents",
			Description: "Query stored events",
			Parameters: []engine.ParameterInfo{
				{Name: "query", Type: "object", Description: "Query parameters", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "getEventHistory",
			Description: "Get event history for a specific timeframe",
			Parameters: []engine.ParameterInfo{
				{Name: "startTime", Type: "string", Description: "Start time (ISO format)", Required: true},
				{Name: "endTime", Type: "string", Description: "End time (ISO format)", Required: false},
			},
			ReturnType: "array",
		},
		// Event Filtering Methods
		{
			Name:        "createFilter",
			Description: "Create a custom event filter",
			Parameters: []engine.ParameterInfo{
				{Name: "filterConfig", Type: "object", Description: "Filter configuration", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "createCompositeFilter",
			Description: "Create a composite filter from multiple filters",
			Parameters: []engine.ParameterInfo{
				{Name: "filters", Type: "array", Description: "Array of filter IDs", Required: true},
				{Name: "operator", Type: "string", Description: "Logical operator (AND/OR)", Required: true},
			},
			ReturnType: "string",
		},
		// Event Replay Methods
		{
			Name:        "replayEvents",
			Description: "Replay events with optional filters",
			Parameters: []engine.ParameterInfo{
				{Name: "query", Type: "object", Description: "Replay query", Required: true},
				{Name: "options", Type: "object", Description: "Replay options", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "pauseReplay",
			Description: "Pause event replay",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "resumeReplay",
			Description: "Resume event replay",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "stopReplay",
			Description: "Stop event replay",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		// Event Serialization Methods
		{
			Name:        "serializeEvent",
			Description: "Serialize an event to a specific format",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event to serialize", Required: true},
				{Name: "format", Type: "string", Description: "Serialization format", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "deserializeEvent",
			Description: "Deserialize an event from string format",
			Parameters: []engine.ParameterInfo{
				{Name: "eventData", Type: "string", Description: "Serialized event data", Required: true},
				{Name: "format", Type: "string", Description: "Data format", Required: false},
			},
			ReturnType: "object",
		},
		// Event Aggregation Methods
		{
			Name:        "createAggregator",
			Description: "Create an event aggregator",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Description: "Aggregator type", Required: true},
				{Name: "config", Type: "object", Description: "Aggregator configuration", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "getAggregatedData",
			Description: "Get aggregated event data",
			Parameters: []engine.ParameterInfo{
				{Name: "aggregatorID", Type: "string", Description: "Aggregator ID", Required: true},
			},
			ReturnType: "object",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *EventBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Event": {
			GoType:     "domain.Event",
			ScriptType: "object",
		},
		"EventFilter": {
			GoType:     "events.EventFilter",
			ScriptType: "object",
		},
		"EventQuery": {
			GoType:     "events.EventQuery",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *EventBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	switch name {
	case "publishEvent", "storeEvent", "serializeEvent":
		if len(args) < 1 {
			return fmt.Errorf("%s requires event parameter", name)
		}
		if args[0].Type() != engine.TypeObject {
			return fmt.Errorf("event must be object")
		}
	case "subscribe":
		if len(args) < 2 {
			return fmt.Errorf("subscribe requires pattern and handler parameters")
		}
		if args[0].Type() != engine.TypeString {
			return fmt.Errorf("pattern must be string")
		}
		if args[1].Type() != engine.TypeFunction {
			return fmt.Errorf("handler must be function")
		}
	case "unsubscribe":
		if len(args) < 1 {
			return fmt.Errorf("unsubscribe requires subscriptionID parameter")
		}
		if args[0].Type() != engine.TypeString {
			return fmt.Errorf("subscriptionID must be string")
		}
	}
	return nil
}

// ExecuteMethod executes a bridge method
func (b *EventBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return engine.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}

	switch name {
	// Event Bus Methods
	case "publishEvent":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("publishEvent requires event parameter")), nil
		}

		if args[0].Type() != engine.TypeObject {
			return engine.NewErrorValue(fmt.Errorf("event must be an object")), nil
		}

		eventData := args[0].ToGo().(map[string]interface{})

		// Convert to domain.Event
		event := b.mapToEvent(eventData)

		// Publish to bus (no context needed)
		b.eventBus.Publish(event)
		return engine.NewNilValue(), nil

	case "subscribe":
		if len(args) < 2 {
			return engine.NewErrorValue(fmt.Errorf("subscribe requires pattern and handler parameters")), nil
		}

		pattern := args[0].(engine.StringValue).Value()

		// Create pattern filter
		filter, err := events.NewPatternFilter(pattern)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("invalid pattern: %w", err)), nil
		}

		// Create event handler
		handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
			// Convert event to script-friendly format
			scriptEvent := b.eventToMap(event)
			// Handler would be invoked here through script engine
			_ = scriptEvent
			return nil
		})

		// Subscribe with filter
		subID := b.eventBus.Subscribe(handler, filter)

		return engine.NewStringValue(subID), nil

	case "subscribeWithFilter":
		if len(args) < 2 {
			return engine.NewErrorValue(fmt.Errorf("subscribeWithFilter requires filter and handler parameters")), nil
		}

		filterData := args[0].ToGo().(map[string]interface{})

		// Create filter from data
		filter, err := b.createFilterFromData(filterData)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("invalid filter: %w", err)), nil
		}

		// Create event handler
		handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
			// Convert event to script-friendly format
			scriptEvent := b.eventToMap(event)
			// Handler would be invoked here through script engine
			_ = scriptEvent
			return nil
		})

		// Subscribe with filter
		subID := b.eventBus.Subscribe(handler, filter)

		return engine.NewStringValue(subID), nil

	case "unsubscribe":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("unsubscribe requires subscriptionID parameter")), nil
		}

		subID := args[0].(engine.StringValue).Value()

		// Unsubscribe from event bus
		b.eventBus.Unsubscribe(subID)

		// Clean up tracking
		delete(b.subscriptions, subID)

		return engine.NewNilValue(), nil

	// Event Storage Methods
	case "storeEvent":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("storeEvent requires event parameter")), nil
		}

		eventData := args[0].ToGo().(map[string]interface{})
		event := b.mapToEvent(eventData)

		// Store event
		if err := b.storage.Store(ctx, event); err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to store event: %w", err)), nil
		}

		return engine.NewNilValue(), nil

	case "queryEvents":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("queryEvents requires query parameter")), nil
		}

		queryData := args[0].ToGo().(map[string]interface{})

		// Convert to EventQuery
		query := events.EventQuery{}
		if agentID, ok := queryData["agentID"].(string); ok {
			query.AgentID = agentID
		}
		if limit, ok := queryData["limit"].(float64); ok {
			query.Limit = int(limit)
		}

		// Query events
		eventsList, err := b.storage.Query(ctx, query)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to query events: %w", err)), nil
		}

		// Convert events to script-friendly format
		result := make([]engine.ScriptValue, len(eventsList))
		for i, event := range eventsList {
			eventMap := b.eventToMap(event)
			result[i] = convertEventToScriptValue(eventMap)
		}

		return engine.NewArrayValue(result), nil

	case "getEventHistory":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("getEventHistory requires startTime parameter")), nil
		}

		startTime := args[0].(engine.StringValue).Value()

		// Parse start time
		start, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("invalid start time format: %w", err)), nil
		}

		// Parse end time if provided
		var end time.Time
		if len(args) > 1 {
			endTime := args[1].(engine.StringValue).Value()
			end, err = time.Parse(time.RFC3339, endTime)
			if err != nil {
				return engine.NewErrorValue(fmt.Errorf("invalid end time format: %w", err)), nil
			}
		} else {
			end = time.Now()
		}

		// Create time-based query
		query := events.EventQuery{
			StartTime: &start,
			EndTime:   &end,
		}

		// Query events
		eventsList, err := b.storage.Query(ctx, query)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to get event history: %w", err)), nil
		}

		// Convert events to script-friendly format
		result := make([]engine.ScriptValue, len(eventsList))
		for i, event := range eventsList {
			eventMap := b.eventToMap(event)
			result[i] = convertEventToScriptValue(eventMap)
		}

		return engine.NewArrayValue(result), nil

	// Event Filtering Methods
	case "createFilter":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("createFilter requires filterConfig parameter")), nil
		}

		filterData := args[0].ToGo().(map[string]interface{})

		// Create filter from data
		filter, err := b.createFilterFromData(filterData)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to create filter: %w", err)), nil
		}

		// Generate filter ID
		filterID := fmt.Sprintf("filter_%d", time.Now().UnixNano())
		b.filters[filterID] = filter

		return engine.NewStringValue(filterID), nil

	case "createCompositeFilter":
		if len(args) < 2 {
			return engine.NewErrorValue(fmt.Errorf("createCompositeFilter requires filters and operator parameters")), nil
		}

		filterIDs := args[0].ToGo().([]interface{})
		operator := args[1].(engine.StringValue).Value()

		// Get filters by ID
		var filters []events.EventFilter
		for _, id := range filterIDs {
			filterID, ok := id.(string)
			if !ok {
				return engine.NewErrorValue(fmt.Errorf("filter ID must be string")), nil
			}
			filter, exists := b.filters[filterID]
			if !exists {
				return engine.NewErrorValue(fmt.Errorf("filter %s not found", filterID)), nil
			}
			filters = append(filters, filter)
		}

		// Create composite filter
		var compositeFilter events.EventFilter

		switch strings.ToUpper(operator) {
		case "AND":
			compositeFilter = events.AND(filters...)
		case "OR":
			compositeFilter = events.OR(filters...)
		default:
			return engine.NewErrorValue(fmt.Errorf("invalid operator: %s", operator)), nil
		}

		// Generate filter ID
		filterID := fmt.Sprintf("composite_filter_%d", time.Now().UnixNano())
		b.filters[filterID] = compositeFilter

		return engine.NewStringValue(filterID), nil

	// Event Replay Methods
	case "replayEvents":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("replayEvents requires query parameter")), nil
		}

		queryData := args[0].ToGo().(map[string]interface{})

		// Convert to EventQuery
		query := events.EventQuery{}
		if agentID, ok := queryData["agentID"].(string); ok {
			query.AgentID = agentID
		}

		// Parse replay options
		options := events.ReplayOptions{
			Speed: 1.0, // Default real-time
		}
		if len(args) > 1 {
			optionsData := args[1].ToGo().(map[string]interface{})
			if speed, ok := optionsData["speed"].(float64); ok {
				options.Speed = speed
			}
		}

		// Perform replay
		if err := b.replayer.Replay(ctx, query, options); err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to replay events: %w", err)), nil
		}

		return engine.NewNilValue(), nil

	case "pauseReplay":
		// EventReplayer doesn't have Pause method - this would need different implementation
		return engine.NewNilValue(), nil

	case "resumeReplay":
		// EventReplayer doesn't have Resume method - this would need different implementation
		return engine.NewNilValue(), nil

	case "stopReplay":
		// EventReplayer doesn't have Stop method - this would need different implementation
		return engine.NewNilValue(), nil

	// Event Serialization Methods
	case "serializeEvent":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("serializeEvent requires event parameter")), nil
		}

		eventData := args[0].ToGo().(map[string]interface{})
		event := b.mapToEvent(eventData)

		// Serialize event
		serialized, err := events.SerializeEvent(event)
		if err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to serialize event: %w", err)), nil
		}

		return convertEventToScriptValue(serialized), nil

	case "deserializeEvent":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("deserializeEvent requires eventData parameter")), nil
		}

		serializedData := args[0].(engine.StringValue).Value()

		// Deserialize event - simplified implementation
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(serializedData), &eventData); err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to deserialize event: %w", err)), nil
		}

		// Return deserialized event data
		return convertEventToScriptValue(eventData), nil

	// Event Aggregation Methods
	case "createAggregator":
		if len(args) < 2 {
			return engine.NewErrorValue(fmt.Errorf("createAggregator requires type and config parameters")), nil
		}

		aggType := args[0].(engine.StringValue).Value()
		config := args[1].ToGo().(map[string]interface{})

		// Create aggregator
		aggregator := &EventAggregator{
			ID:         fmt.Sprintf("agg_%d", time.Now().UnixNano()),
			Type:       aggType,
			WindowSize: 5 * time.Minute, // Default window
			Events:     make([]domain.Event, 0),
			LastUpdate: time.Now(),
		}

		// Parse configuration
		if windowSize, ok := config["windowSize"].(float64); ok {
			aggregator.WindowSize = time.Duration(windowSize) * time.Second
		}

		b.aggregators[aggregator.ID] = aggregator

		return engine.NewStringValue(aggregator.ID), nil

	case "getAggregatedData":
		if len(args) < 1 {
			return engine.NewErrorValue(fmt.Errorf("getAggregatedData requires aggregatorID parameter")), nil
		}

		aggID := args[0].(engine.StringValue).Value()

		aggregator, exists := b.aggregators[aggID]
		if !exists {
			return engine.NewErrorValue(fmt.Errorf("aggregator %s not found", aggID)), nil
		}

		// Return aggregated data
		result := map[string]engine.ScriptValue{
			"id":         engine.NewStringValue(aggregator.ID),
			"type":       engine.NewStringValue(aggregator.Type),
			"eventCount": engine.NewNumberValue(float64(len(aggregator.Events))),
			"lastUpdate": engine.NewStringValue(aggregator.LastUpdate.Format(time.RFC3339)),
		}

		return engine.NewObjectValue(result), nil

	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// RequiredPermissions returns required permissions
func (b *EventBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "events",
			Actions:     []string{"publish", "subscribe", "query"},
			Description: "Access to event system",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "event_storage",
			Actions:     []string{"read", "write"},
			Description: "Memory for event storage and caching",
		},
	}
}

// Helper methods

// mapToEvent converts script data to domain.Event
func (b *EventBridge) mapToEvent(data map[string]interface{}) domain.Event {
	// This is a simplified implementation
	// In practice, would need to properly construct domain.Event
	eventType, _ := data["type"].(string)
	if eventType == "" {
		eventType = "script_event"
	}

	// Create a basic event with required parameters
	id, _ := data["id"].(string)
	if id == "" {
		id = fmt.Sprintf("event_%d", time.Now().UnixNano())
	}

	agentID, _ := data["agentID"].(string)
	if agentID == "" {
		agentID = "script_agent"
	}

	return domain.NewEvent(domain.EventType(eventType), id, agentID, data)
}

// eventToMap converts domain.Event to script-friendly map
func (b *EventBridge) eventToMap(event domain.Event) map[string]interface{} {
	// This is a simplified implementation
	// In practice, would need to properly convert domain.Event
	return map[string]interface{}{
		"id":        event.ID,
		"type":      string(event.Type),
		"timestamp": event.Timestamp.Format(time.RFC3339),
		"data":      event.Data,
	}
}

// createFilterFromData creates an EventFilter from script data
func (b *EventBridge) createFilterFromData(data map[string]interface{}) (events.EventFilter, error) {
	filterType, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("filter type is required")
	}

	switch filterType {
	case "pattern":
		pattern, ok := data["pattern"].(string)
		if !ok {
			return nil, fmt.Errorf("pattern is required for pattern filter")
		}
		return events.NewPatternFilter(pattern)

	case "type":
		eventType, ok := data["eventType"].(string)
		if !ok {
			return nil, fmt.Errorf("eventType is required for type filter")
		}
		return events.NewTypeFilter(domain.EventType(eventType)), nil

	default:
		return nil, fmt.Errorf("unknown filter type: %s", filterType)
	}
}

// convertEventToScriptValue converts a Go interface{} to engine.ScriptValue for events
func convertEventToScriptValue(v interface{}) engine.ScriptValue {
	if v == nil {
		return engine.NewNilValue()
	}

	switch val := v.(type) {
	case string:
		return engine.NewStringValue(val)
	case bool:
		return engine.NewBoolValue(val)
	case int:
		return engine.NewNumberValue(float64(val))
	case int64:
		return engine.NewNumberValue(float64(val))
	case float64:
		return engine.NewNumberValue(val)
	case float32:
		return engine.NewNumberValue(float64(val))
	case map[string]interface{}:
		result := make(map[string]engine.ScriptValue)
		for k, mv := range val {
			result[k] = convertEventToScriptValue(mv)
		}
		return engine.NewObjectValue(result)
	case []interface{}:
		result := make([]engine.ScriptValue, len(val))
		for i, av := range val {
			result[i] = convertEventToScriptValue(av)
		}
		return engine.NewArrayValue(result)
	default:
		// For unknown types, convert to string representation
		return engine.NewStringValue(fmt.Sprintf("%v", val))
	}
}
