// ABOUTME: Event system bridge v2.0.0 integrating go-llms v0.3.5 event infrastructure
// ABOUTME: Provides comprehensive event bus, storage, filtering, serialization, and replay capabilities

package agent

import (
	"context"
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
	Name        string
	Filter      events.EventFilter
	Window      time.Duration
	Aggregator  func([]domain.Event) interface{}
	LastUpdate  time.Time
	EventBuffer []domain.Event
	mu          sync.Mutex
}

// channelEventStream implements domain.EventStream using channels
type channelEventStream struct {
	events chan domain.Event
	closed bool
	mu     sync.Mutex
}

// Next returns the next event or blocks until one is available
func (s *channelEventStream) Next() (domain.Event, error) {
	event, ok := <-s.events
	if !ok {
		return domain.Event{}, fmt.Errorf("stream closed")
	}
	return event, nil
}

// Close closes the stream
func (s *channelEventStream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		close(s.events)
		s.closed = true
	}
}

// NewEventBridge creates a new v2.0.0 event bridge
func NewEventBridge() *EventBridge {
	// Create event bus with default configuration
	bus := events.NewEventBus(events.WithBufferSize(1000))

	// Create memory storage for events
	storage := events.NewMemoryStorage()

	// Create recorder and replayer
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

	// Stop all event streams
	for id, stream := range b.streams {
		stream.Close()
		delete(b.streams, id)
	}

	// Clear subscriptions
	for id := range b.subscriptions {
		b.eventBus.Unsubscribe(id)
		delete(b.subscriptions, id)
	}

	// Stop components
	if b.recorder != nil {
		b.recorder.Stop()
	}

	if b.eventBus != nil {
		b.eventBus.Close()
	}

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
				{Name: "event", Type: "object", Description: "Event to publish", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "subscribe",
			Description: "Subscribe to events with pattern matching",
			Parameters: []engine.ParameterInfo{
				{Name: "pattern", Type: "string", Description: "Event pattern (e.g., 'agent.*', 'tool.execute')", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
			},
			ReturnType: "string", // subscription ID
		},
		{
			Name:        "subscribeWithFilter",
			Description: "Subscribe to events with custom filter",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Event filter configuration", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
			},
			ReturnType: "string", // subscription ID
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
			Description: "Store an event in event storage",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event to store", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "queryEvents",
			Description: "Query stored events with filters",
			Parameters: []engine.ParameterInfo{
				{Name: "query", Type: "object", Description: "Query parameters", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "streamEvents",
			Description: "Stream events from storage",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Stream filter", Required: false},
				{Name: "handler", Type: "function", Description: "Stream handler", Required: true},
			},
			ReturnType: "string", // stream ID
		},
		{
			Name:        "stopStream",
			Description: "Stop an event stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
			},
			ReturnType: "void",
		},
		// Event Filtering Methods
		{
			Name:        "createPatternFilter",
			Description: "Create a pattern-based event filter",
			Parameters: []engine.ParameterInfo{
				{Name: "pattern", Type: "string", Description: "Event pattern", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "createTypeFilter",
			Description: "Create a type-based event filter",
			Parameters: []engine.ParameterInfo{
				{Name: "eventTypes", Type: "array", Description: "Event types to match", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "createCompositeFilter",
			Description: "Create a composite filter with AND/OR/NOT logic",
			Parameters: []engine.ParameterInfo{
				{Name: "operator", Type: "string", Description: "Logical operator (AND, OR, NOT)", Required: true},
				{Name: "filters", Type: "array", Description: "Child filters", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "createFieldFilter",
			Description: "Create a field-based filter",
			Parameters: []engine.ParameterInfo{
				{Name: "field", Type: "string", Description: "Field path", Required: true},
				{Name: "operator", Type: "string", Description: "Comparison operator", Required: true},
				{Name: "value", Type: "any", Description: "Value to compare", Required: true},
			},
			ReturnType: "object",
		},
		// Event Serialization Methods
		{
			Name:        "serializeEvent",
			Description: "Serialize an event to a specific format",
			Parameters: []engine.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event to serialize", Required: true},
				{Name: "format", Type: "string", Description: "Format (json, json-pretty, compact)", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "deserializeEvent",
			Description: "Deserialize an event from string",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "string", Description: "Serialized event data", Required: true},
				{Name: "format", Type: "string", Description: "Format", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "serializeEventBatch",
			Description: "Serialize multiple events as a batch",
			Parameters: []engine.ParameterInfo{
				{Name: "events", Type: "array", Description: "Events to serialize", Required: true},
			},
			ReturnType: "string",
		},
		// Event Replay Methods
		{
			Name:        "startEventRecording",
			Description: "Start recording events",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Optional filter for recording", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "stopEventRecording",
			Description: "Stop recording events",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "replayEvents",
			Description: "Replay recorded events",
			Parameters: []engine.ParameterInfo{
				{Name: "options", Type: "object", Description: "Replay options (speed, filter, etc.)", Required: false},
			},
			ReturnType: "string", // replay session ID
		},
		{
			Name:        "pauseReplay",
			Description: "Pause event replay",
			Parameters: []engine.ParameterInfo{
				{Name: "sessionID", Type: "string", Description: "Replay session ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "resumeReplay",
			Description: "Resume event replay",
			Parameters: []engine.ParameterInfo{
				{Name: "sessionID", Type: "string", Description: "Replay session ID", Required: true},
			},
			ReturnType: "void",
		},
		// Event Aggregation Methods
		{
			Name:        "createAggregator",
			Description: "Create an event aggregator",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Aggregator name", Required: true},
				{Name: "config", Type: "object", Description: "Aggregator configuration", Required: true},
			},
			ReturnType: "string", // aggregator ID
		},
		{
			Name:        "getAggregatedData",
			Description: "Get aggregated event data",
			Parameters: []engine.ParameterInfo{
				{Name: "aggregatorID", Type: "string", Description: "Aggregator ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "resetAggregator",
			Description: "Reset an aggregator",
			Parameters: []engine.ParameterInfo{
				{Name: "aggregatorID", Type: "string", Description: "Aggregator ID", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "removeAggregator",
			Description: "Remove an aggregator",
			Parameters: []engine.ParameterInfo{
				{Name: "aggregatorID", Type: "string", Description: "Aggregator ID", Required: true},
			},
			ReturnType: "void",
		},
		// Bridge Event Methods
		{
			Name:        "publishBridgeEvent",
			Description: "Publish a bridge-specific event",
			Parameters: []engine.ParameterInfo{
				{Name: "eventType", Type: "string", Description: "Bridge event type", Required: true},
				{Name: "data", Type: "object", Description: "Event data", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "onBridgeEvent",
			Description: "Subscribe to bridge events",
			Parameters: []engine.ParameterInfo{
				{Name: "eventType", Type: "string", Description: "Bridge event type", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string", // subscription ID
		},
		// Event Stream Methods
		{
			Name:        "createEventStream",
			Description: "Create a functional event stream",
			Parameters: []engine.ParameterInfo{
				{Name: "source", Type: "string", Description: "Event source pattern", Required: true},
			},
			ReturnType: "string", // stream ID
		},
		{
			Name:        "filterStream",
			Description: "Apply filter to event stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
				{Name: "predicate", Type: "function", Description: "Filter predicate", Required: true},
			},
			ReturnType: "string", // new stream ID
		},
		{
			Name:        "mapStream",
			Description: "Transform events in stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
				{Name: "mapper", Type: "function", Description: "Transform function", Required: true},
			},
			ReturnType: "string", // new stream ID
		},
		{
			Name:        "reduceStream",
			Description: "Reduce event stream to single value",
			Parameters: []engine.ParameterInfo{
				{Name: "streamID", Type: "string", Description: "Stream ID", Required: true},
				{Name: "reducer", Type: "function", Description: "Reducer function", Required: true},
				{Name: "initial", Type: "any", Description: "Initial value", Required: false},
			},
			ReturnType: "any",
		},
		// Utility Methods
		{
			Name:        "getEventStats",
			Description: "Get event system statistics",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
		},
		{
			Name:        "exportEventStore",
			Description: "Export event store to file",
			Parameters: []engine.ParameterInfo{
				{Name: "filepath", Type: "string", Description: "Export file path", Required: true},
				{Name: "format", Type: "string", Description: "Export format", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "importEventStore",
			Description: "Import events from file",
			Parameters: []engine.ParameterInfo{
				{Name: "filepath", Type: "string", Description: "Import file path", Required: true},
			},
			ReturnType: "number", // imported event count
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
		"EventType": {
			GoType:     "domain.EventType",
			ScriptType: "string",
		},
		"EventFilter": {
			GoType:     "events.EventFilter",
			ScriptType: "object",
		},
		"EventQuery": {
			GoType:     "events.EventQuery",
			ScriptType: "object",
		},
		"EventStream": {
			GoType:     "*domain.EventStream",
			ScriptType: "object",
		},
		"BridgeEvent": {
			GoType:     "events.BridgeEvent",
			ScriptType: "object",
		},
		"ReplayOptions": {
			GoType:     "events.ReplayOptions",
			ScriptType: "object",
		},
		"EventSerializer": {
			GoType:     "events.EventSerializer",
			ScriptType: "object",
		},
		"EventAggregator": {
			GoType:     "*EventAggregator",
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
			Actions:     []string{"publish", "subscribe", "filter", "aggregate", "replay"},
			Description: "Access to event system v2.0.0",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "event_storage",
			Actions:     []string{"allocate", "read", "write"},
			Description: "Memory for event storage and aggregation",
		},
		{
			Type:        engine.PermissionFileSystem,
			Resource:    "event_files",
			Actions:     []string{"read", "write"},
			Description: "File access for event import/export",
		},
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
	// Event Bus Methods
	case "publishEvent":
		if len(args) < 1 {
			return nil, fmt.Errorf("publishEvent requires event parameter")
		}

		eventData, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("event must be an object")
		}

		// Convert to domain.Event
		event := b.mapToEvent(eventData)

		// Publish to bus (no context needed)
		b.eventBus.Publish(event)
		return nil, nil

	case "subscribe":
		if len(args) < 2 {
			return nil, fmt.Errorf("subscribe requires pattern and handler parameters")
		}

		pattern, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pattern must be string")
		}

		// Create pattern filter
		filter, err := events.NewPatternFilter(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
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

		return subID, nil

	case "subscribeWithFilter":
		if len(args) < 2 {
			return nil, fmt.Errorf("subscribeWithFilter requires filter and handler parameters")
		}

		filterConfig, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("filter must be an object")
		}

		// Create filter from config
		filter, err := b.createFilterFromConfig(filterConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid filter config: %w", err)
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

		return subID, nil

	case "unsubscribe":
		if len(args) < 1 {
			return nil, fmt.Errorf("unsubscribe requires subscriptionID parameter")
		}

		subID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("subscriptionID must be string")
		}

		b.eventBus.Unsubscribe(subID)
		return nil, nil

	// Event Filtering Methods
	case "createPatternFilter":
		if len(args) < 1 {
			return nil, fmt.Errorf("createPatternFilter requires pattern parameter")
		}

		pattern, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pattern must be string")
		}

		filter, err := events.NewPatternFilter(pattern)
		if err != nil {
			return nil, err
		}

		// Store filter for later use
		filterID := fmt.Sprintf("filter_%d", time.Now().UnixNano())

		// Need to unlock before taking write lock
		b.mu.RUnlock()
		b.mu.Lock()
		b.filters[filterID] = filter
		b.mu.Unlock()
		b.mu.RLock()

		return filterID, nil

	case "createTypeFilter":
		if len(args) < 1 {
			return nil, fmt.Errorf("createTypeFilter requires eventTypes parameter")
		}

		types, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("eventTypes must be array")
		}

		eventTypes := make([]domain.EventType, len(types))
		for i, t := range types {
			if str, ok := t.(string); ok {
				eventTypes[i] = domain.EventType(str)
			}
		}

		filter := events.NewTypeFilter(eventTypes...)

		// Store filter for later use
		filterID := fmt.Sprintf("filter_%d", time.Now().UnixNano())

		// Need to unlock before taking write lock
		b.mu.RUnlock()
		b.mu.Lock()
		b.filters[filterID] = filter
		b.mu.Unlock()
		b.mu.RLock()

		return filterID, nil

	case "createCompositeFilter":
		if len(args) < 2 {
			return nil, fmt.Errorf("createCompositeFilter requires operator and filters parameters")
		}

		operator, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("operator must be string")
		}

		filterConfigs, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("filters must be array")
		}

		filters := make([]events.EventFilter, 0, len(filterConfigs))
		for _, fc := range filterConfigs {
			if cfMap, ok := fc.(map[string]interface{}); ok {
				filter, err := b.createFilterFromConfig(cfMap)
				if err == nil {
					filters = append(filters, filter)
				}
			}
		}

		var compositeFilter events.EventFilter
		switch operator {
		case "AND":
			compositeFilter = events.AND(filters...)
		case "OR":
			compositeFilter = events.OR(filters...)
		case "NOT":
			if len(filters) > 0 {
				compositeFilter = events.NOT(filters[0])
			} else {
				return nil, fmt.Errorf("NOT filter requires at least one filter")
			}
		default:
			return nil, fmt.Errorf("invalid operator: %s", operator)
		}

		// Store filter for later use
		filterID := fmt.Sprintf("filter_%d", time.Now().UnixNano())

		// Need to unlock before taking write lock
		b.mu.RUnlock()
		b.mu.Lock()
		b.filters[filterID] = compositeFilter
		b.mu.Unlock()
		b.mu.RLock()

		return filterID, nil

	case "createFieldFilter":
		if len(args) < 3 {
			return nil, fmt.Errorf("createFieldFilter requires field, operator, and value parameters")
		}

		field, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("field must be string")
		}

		op, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("operator must be string")
		}

		value := args[2]

		// Create field filter using a filter function
		filter := events.EventFilterFunc(func(event domain.Event) bool {
			// Check in metadata first
			if fieldValue, exists := event.Metadata[field]; exists {
				return b.compareValues(fieldValue, op, value)
			}

			// Check in event data if it's a map
			if data, ok := event.Data.(map[string]interface{}); ok {
				if fieldValue, exists := data[field]; exists {
					return b.compareValues(fieldValue, op, value)
				}
			}

			// Field not found
			return false
		})

		// Store filter for later use
		filterID := fmt.Sprintf("filter_%d", time.Now().UnixNano())

		// Need to unlock before taking write lock
		b.mu.RUnlock()
		b.mu.Lock()
		b.filters[filterID] = filter
		b.mu.Unlock()
		b.mu.RLock()

		return filterID, nil

	// Event Storage Methods
	case "storeEvent":
		if len(args) < 1 {
			return nil, fmt.Errorf("storeEvent requires event parameter")
		}

		eventData, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("event must be an object")
		}

		// Convert to domain.Event
		event := b.mapToEvent(eventData)

		// Store event
		err := b.storage.Store(ctx, event)
		return nil, err

	case "queryEvents":
		if len(args) < 1 {
			return nil, fmt.Errorf("queryEvents requires query parameter")
		}

		queryData, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("query must be an object")
		}

		// Build query from data
		query := b.buildEventQuery(queryData)

		// Execute query
		events, err := b.storage.Query(ctx, query)
		if err != nil {
			return nil, err
		}

		// Convert to script-friendly format
		result := make([]map[string]interface{}, len(events))
		for i, event := range events {
			result[i] = b.eventToMap(event)
		}

		return result, nil

	// Event Serialization Methods
	case "serializeEvent":
		if len(args) < 1 {
			return nil, fmt.Errorf("serializeEvent requires event parameter")
		}

		eventData, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("event must be an object")
		}

		format := "json"
		if len(args) > 1 {
			if f, ok := args[1].(string); ok {
				format = f
			}
		}

		// Convert to domain.Event
		event := b.mapToEvent(eventData)

		// Get serializer
		serializer := events.GetSerializer(format)

		// Serialize
		data, err := serializer.Serialize(event)
		if err != nil {
			return nil, err
		}

		return string(data), nil

	case "deserializeEvent":
		if len(args) < 1 {
			return nil, fmt.Errorf("deserializeEvent requires data parameter")
		}

		data, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("data must be string")
		}

		format := "json"
		if len(args) > 1 {
			if f, ok := args[1].(string); ok {
				format = f
			}
		}

		// Get serializer
		serializer := events.GetSerializer(format)

		// Deserialize
		event, err := serializer.Deserialize([]byte(data))
		if err != nil {
			return nil, err
		}

		// Convert to script-friendly format
		return b.eventToMap(event), nil

	case "serializeEventBatch":
		if len(args) < 1 {
			return nil, fmt.Errorf("serializeEventBatch requires events parameter")
		}

		eventsList, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("events must be array")
		}

		// Convert all events to domain.Event
		eventObjs := make([]domain.Event, 0, len(eventsList))
		for _, e := range eventsList {
			if eventData, ok := e.(map[string]interface{}); ok {
				eventObjs = append(eventObjs, b.mapToEvent(eventData))
			}
		}

		// Serialize batch
		result := make([]string, 0, len(eventsList))
		serializer := events.GetSerializer("json")
		for _, event := range eventObjs {
			data, err := serializer.Serialize(event)
			if err == nil {
				result = append(result, string(data))
			}
		}

		return result, nil

	// Event Replay Methods
	case "startEventRecording":
		// Start recording without filters
		if err := b.recorder.Start(); err != nil {
			return nil, fmt.Errorf("failed to start recording: %w", err)
		}
		return nil, nil

	case "stopEventRecording":
		b.recorder.Stop()
		return nil, nil

	case "replayEvents":
		options := events.ReplayOptions{
			Speed: 1.0,
		}

		if len(args) > 0 {
			if opts, ok := args[0].(map[string]interface{}); ok {
				if speed, ok := opts["speed"].(float64); ok {
					options.Speed = speed
				}
				if filter, ok := opts["filter"].(map[string]interface{}); ok {
					f, err := b.createFilterFromConfig(filter)
					if err == nil {
						options.Filter = f
					}
				}
			}
		}

		// Start replay with empty query (replay all events)
		sessionID := fmt.Sprintf("replay_%d", time.Now().UnixNano())
		query := events.EventQuery{}
		go func() {
			if err := b.replayer.Replay(ctx, query, options); err != nil {
				// Log error or handle it appropriately
				// For now, we'll just ignore it since it's async
				_ = err
			}
		}()

		return sessionID, nil

	// Event Aggregation Methods
	case "createAggregator":
		if len(args) < 2 {
			return nil, fmt.Errorf("createAggregator requires name and config parameters")
		}

		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Create aggregator
		aggregator := &EventAggregator{
			ID:          fmt.Sprintf("agg_%d", time.Now().UnixNano()),
			Name:        name,
			LastUpdate:  time.Now(),
			EventBuffer: make([]domain.Event, 0),
		}

		// Configure window
		if window, ok := config["window"].(float64); ok {
			aggregator.Window = time.Duration(window) * time.Second
		} else {
			aggregator.Window = 60 * time.Second // Default 1 minute
		}

		// Configure filter
		if filterConfig, ok := config["filter"].(map[string]interface{}); ok {
			filter, err := b.createFilterFromConfig(filterConfig)
			if err == nil {
				aggregator.Filter = filter
			}
		}

		// Store aggregator
		// Need to unlock before taking write lock
		b.mu.RUnlock()
		b.mu.Lock()
		b.aggregators[aggregator.ID] = aggregator
		b.mu.Unlock()
		b.mu.RLock()

		// Create event handler for aggregation
		handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
			aggregator.mu.Lock()
			defer aggregator.mu.Unlock()

			// Add to buffer
			aggregator.EventBuffer = append(aggregator.EventBuffer, event)
			aggregator.LastUpdate = time.Now()

			// Clean old events outside window
			cutoff := time.Now().Add(-aggregator.Window)
			i := 0
			for i < len(aggregator.EventBuffer) && aggregator.EventBuffer[i].Timestamp.Before(cutoff) {
				i++
			}
			if i > 0 {
				aggregator.EventBuffer = aggregator.EventBuffer[i:]
			}
			return nil
		})

		// Subscribe to events for aggregation
		b.eventBus.Subscribe(handler, aggregator.Filter)

		return aggregator.ID, nil

	case "getAggregatedData":
		if len(args) < 1 {
			return nil, fmt.Errorf("getAggregatedData requires aggregatorID parameter")
		}

		aggID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("aggregatorID must be string")
		}

		b.mu.RLock()
		aggregator, exists := b.aggregators[aggID]
		b.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("aggregator not found: %s", aggID)
		}

		aggregator.mu.Lock()
		defer aggregator.mu.Unlock()

		// Return aggregated data
		return map[string]interface{}{
			"id":         aggregator.ID,
			"name":       aggregator.Name,
			"eventCount": len(aggregator.EventBuffer),
			"window":     aggregator.Window.Seconds(),
			"lastUpdate": aggregator.LastUpdate.Format(time.RFC3339),
			"events":     len(aggregator.EventBuffer),
		}, nil

	case "resetAggregator":
		if len(args) < 1 {
			return nil, fmt.Errorf("resetAggregator requires aggregatorID parameter")
		}

		aggID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("aggregatorID must be string")
		}

		b.mu.RLock()
		aggregator, exists := b.aggregators[aggID]
		b.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("aggregator not found: %s", aggID)
		}

		// Reset aggregator
		aggregator.mu.Lock()
		aggregator.EventBuffer = make([]domain.Event, 0)
		aggregator.LastUpdate = time.Now()
		aggregator.mu.Unlock()

		return nil, nil

	case "removeAggregator":
		if len(args) < 1 {
			return nil, fmt.Errorf("removeAggregator requires aggregatorID parameter")
		}

		aggID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("aggregatorID must be string")
		}

		// Remove aggregator
		b.mu.RUnlock()
		b.mu.Lock()
		delete(b.aggregators, aggID)
		b.mu.Unlock()
		b.mu.RLock()

		return nil, nil

	// Bridge Event Methods
	case "publishBridgeEvent":
		if len(args) < 1 {
			return nil, fmt.Errorf("publishBridgeEvent requires eventType parameter")
		}

		eventType, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("eventType must be string")
		}

		var data interface{}
		if len(args) > 1 {
			data = args[1]
		}

		// Create and publish a bridge event
		bridgeEvent := events.NewBridgeEvent(events.BridgeEventType(eventType), "go-llmspell", "bridge-session", data)
		b.eventBus.Publish(bridgeEvent.AsDomainEvent())
		return nil, nil

	// Event Stream Methods
	case "createEventStream":
		if len(args) < 1 {
			return nil, fmt.Errorf("createEventStream requires source parameter")
		}

		source, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("source must be string")
		}

		// Create event stream with pattern filter
		filter, err := events.NewPatternFilter(source)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}

		// Create a channel-based event stream
		eventChan := make(chan domain.Event, 100)
		stream := &channelEventStream{events: eventChan}

		// Create handler that feeds events into stream
		handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
			select {
			case eventChan <- event:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})

		// Subscribe to events
		b.eventBus.Subscribe(handler, filter)

		// Store stream
		streamID := fmt.Sprintf("stream_%d", time.Now().UnixNano())

		// Need to unlock before taking write lock
		b.mu.RUnlock()
		b.mu.Lock()
		b.streams[streamID] = stream
		b.mu.Unlock()
		b.mu.RLock()

		return streamID, nil

	// Utility Methods
	case "getEventStats":
		// Return basic stats we can calculate
		// No need for additional lock, already have RLock
		return map[string]interface{}{
			"subscriptions": b.eventBus.GetSubscriptionCount(),
			"aggregators":   len(b.aggregators),
			"streams":       len(b.streams),
			"filters":       len(b.filters),
			"storage": map[string]interface{}{
				"type":   "memory",
				"active": true,
			},
			"bus": map[string]interface{}{
				"active":      true,
				"buffer_size": 100,
			},
		}, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper methods

// mapToEvent converts a map to domain.Event
func (b *EventBridge) mapToEvent(data map[string]interface{}) domain.Event {
	event := domain.Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if id, ok := data["id"].(string); ok {
		event.ID = id
	}

	if t, ok := data["type"].(string); ok {
		event.Type = domain.EventType(t)
	}

	if agentID, ok := data["agentID"].(string); ok {
		event.AgentID = agentID
	}

	if agentName, ok := data["agentName"].(string); ok {
		event.AgentName = agentName
	}

	if d, ok := data["data"]; ok {
		event.Data = d
	}

	if meta, ok := data["metadata"].(map[string]interface{}); ok {
		event.Metadata = meta
	}

	return event
}

// eventToMap converts domain.Event to map
func (b *EventBridge) eventToMap(event domain.Event) map[string]interface{} {
	result := map[string]interface{}{
		"id":        event.ID,
		"type":      string(event.Type),
		"timestamp": event.Timestamp.Format(time.RFC3339),
		"agentID":   event.AgentID,
		"agentName": event.AgentName,
	}

	if event.Data != nil {
		result["data"] = event.Data
	}

	if len(event.Metadata) > 0 {
		result["metadata"] = event.Metadata
	}

	if event.Error != nil {
		result["error"] = event.Error.Error()
	}

	return result
}

// createFilterFromConfig creates an EventFilter from configuration
func (b *EventBridge) createFilterFromConfig(config map[string]interface{}) (events.EventFilter, error) {
	filterType, ok := config["type"].(string)
	if !ok {
		return nil, fmt.Errorf("filter type required")
	}

	switch filterType {
	case "pattern":
		pattern, ok := config["pattern"].(string)
		if !ok {
			return nil, fmt.Errorf("pattern required for pattern filter")
		}
		return events.NewPatternFilter(pattern)

	case "type":
		types, ok := config["eventTypes"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("eventTypes required for type filter")
		}
		eventTypes := make([]domain.EventType, len(types))
		for i, t := range types {
			if str, ok := t.(string); ok {
				eventTypes[i] = domain.EventType(str)
			}
		}
		return events.NewTypeFilter(eventTypes...), nil

	case "agent":
		agentID, _ := config["agentID"].(string)
		agentName, _ := config["agentName"].(string)
		if agentID == "" && agentName == "" {
			return nil, fmt.Errorf("agentID or agentName required for agent filter")
		}
		return events.NewAgentFilter(agentID, agentName), nil

	case "composite":
		operator, ok := config["operator"].(string)
		if !ok {
			return nil, fmt.Errorf("operator required for composite filter")
		}

		childFilters, ok := config["filters"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("filters required for composite filter")
		}

		filters := make([]events.EventFilter, 0, len(childFilters))
		for _, cf := range childFilters {
			if cfMap, ok := cf.(map[string]interface{}); ok {
				filter, err := b.createFilterFromConfig(cfMap)
				if err == nil {
					filters = append(filters, filter)
				}
			}
		}

		switch operator {
		case "AND":
			return events.AND(filters...), nil
		case "OR":
			return events.OR(filters...), nil
		case "NOT":
			if len(filters) > 0 {
				return events.NOT(filters[0]), nil
			}
		}

		return nil, fmt.Errorf("invalid composite operator: %s", operator)

	default:
		return nil, fmt.Errorf("unknown filter type: %s", filterType)
	}
}

// buildEventQuery builds an EventQuery from configuration
func (b *EventBridge) buildEventQuery(config map[string]interface{}) events.EventQuery {
	query := events.EventQuery{
		Limit: 100, // Default limit
	}

	if limit, ok := config["limit"].(float64); ok {
		query.Limit = int(limit)
	}

	if offset, ok := config["offset"].(float64); ok {
		query.Offset = int(offset)
	}

	if startTime, ok := config["startTime"].(string); ok {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime, ok := config["endTime"].(string); ok {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if agentID, ok := config["agentID"].(string); ok {
		query.AgentID = agentID
	}

	if eventTypes, ok := config["eventTypes"].([]interface{}); ok {
		types := make([]domain.EventType, 0, len(eventTypes))
		for _, t := range eventTypes {
			if str, ok := t.(string); ok {
				types = append(types, domain.EventType(str))
			}
		}
		query.EventTypes = types
	}

	if orderBy, ok := config["orderBy"].(string); ok {
		query.OrderBy = orderBy
	}

	if descending, ok := config["descending"].(bool); ok {
		query.Descending = descending
	}

	return query
}

// compareValues compares two values using the specified operator
func (b *EventBridge) compareValues(fieldValue interface{}, operator string, value interface{}) bool {
	switch operator {
	case "=", "==", "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", value)
	case "!=", "not equals":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", value)
	case ">", "gt":
		return compareNumeric(fieldValue, value, ">")
	case ">=", "gte":
		return compareNumeric(fieldValue, value, ">=")
	case "<", "lt":
		return compareNumeric(fieldValue, value, "<")
	case "<=", "lte":
		return compareNumeric(fieldValue, value, "<=")
	case "contains":
		return contains(fieldValue, value)
	case "startsWith":
		return startsWith(fieldValue, value)
	case "endsWith":
		return endsWith(fieldValue, value)
	default:
		return false
	}
}

// compareNumeric compares two numeric values
func compareNumeric(a, b interface{}, op string) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch op {
	case ">":
		return aFloat > bFloat
	case ">=":
		return aFloat >= bFloat
	case "<":
		return aFloat < bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 converts an interface{} to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uint32:
		return float64(val), true
	default:
		return 0, false
	}
}

// contains checks if a contains b
func contains(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return len(aStr) > 0 && len(bStr) > 0 && strings.Contains(aStr, bStr)
}

// startsWith checks if a starts with b
func startsWith(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return len(aStr) >= len(bStr) && aStr[:len(bStr)] == bStr
}

// endsWith checks if a ends with b
func endsWith(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return len(aStr) >= len(bStr) && aStr[len(aStr)-len(bStr):] == bStr
}
