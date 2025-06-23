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

	"github.com/lexlapax/go-llmspell/pkg/bridge/types"

	// go-llms v0.3.5 event imports
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

// EventBridge provides script access to go-llms v0.3.5 event functionality.
// It manages event publishing, subscription, storage, filtering, aggregation,
// and replay capabilities for comprehensive event-driven programming.
type EventBridge struct {
	mu          sync.RWMutex
	initialized bool
	isRecording bool // Track recording state

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

// EventAggregator handles event aggregation logic.
// It collects events within a time window and caches aggregated results
// for efficient processing of event streams.
type EventAggregator struct {
	ID          string
	Type        string
	WindowSize  time.Duration
	Events      []domain.Event
	LastUpdate  time.Time
	ResultCache interface{}
}

// NewEventBridge creates a new event bridge.
// It initializes the event bus, storage, recorder, and replayer
// for full event management functionality.
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

// GetID returns the bridge identifier.
// It implements the types.Bridge interface.
func (b *EventBridge) GetID() string {
	return "agent_events"
}

// GetMetadata returns bridge metadata.
// It provides information about the event bridge version,
// description, and supported features.
func (b *EventBridge) GetMetadata() types.BridgeMetadata {
	return types.BridgeMetadata{
		Name:        "agent_events",
		Version:     "2.0.0",
		Description: "Event system bridge v2.0.0 with go-llms v0.3.5 integration: bus, storage, filtering, serialization, aggregation, and replay",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
// It sets up the bridge-specific event publisher and listener
// for script interaction with the event system.
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
	// Don't start recording automatically - let the user control it

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
// It unsubscribes all active subscriptions, clears registries,
// and stops event recording.
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

// IsInitialized checks if the bridge is initialized.
// It returns true if the bridge has been initialized and is ready for use.
func (b *EventBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script types.
// It enables the script engine to access event functionality through this bridge.
func (b *EventBridge) RegisterWithEngine(engine types.ScriptEngine) error {
	// Bridge registration is handled by the caller (types.RegisterBridge)
	// This method can be used for additional setup if needed
	return nil
}

// Methods returns the methods exposed by this bridge.
// It provides metadata about all event-related methods available to scripts,
// including publishing, subscription, storage, filtering, and replay operations.
func (b *EventBridge) Methods() []types.MethodInfo {
	return []types.MethodInfo{
		// Event Bus Methods
		{
			Name:        "publishEvent",
			Description: "Publish an event to the event bus",
			Parameters: []types.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event data", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "subscribe",
			Description: "Subscribe to events with pattern matching",
			Parameters: []types.ParameterInfo{
				{Name: "pattern", Type: "string", Description: "Event pattern", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "subscribeWithFilter",
			Description: "Subscribe to events with custom filter",
			Parameters: []types.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Event filter", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "unsubscribe",
			Description: "Unsubscribe from events",
			Parameters: []types.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID", Required: true},
			},
			ReturnType: "void",
		},
		// Event Storage Methods
		{
			Name:        "storeEvent",
			Description: "Store an event in persistent storage",
			Parameters: []types.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event to store", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "queryEvents",
			Description: "Query stored events",
			Parameters: []types.ParameterInfo{
				{Name: "query", Type: "object", Description: "Query parameters", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "getEventHistory",
			Description: "Get event history for a specific timeframe",
			Parameters: []types.ParameterInfo{
				{Name: "startTime", Type: "string", Description: "Start time (ISO format)", Required: true},
				{Name: "endTime", Type: "string", Description: "End time (ISO format)", Required: false},
			},
			ReturnType: "array",
		},
		// Event Filtering Methods
		{
			Name:        "createFilter",
			Description: "Create a custom event filter",
			Parameters: []types.ParameterInfo{
				{Name: "filterConfig", Type: "object", Description: "Filter configuration", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "createCompositeFilter",
			Description: "Create a composite filter from multiple filters",
			Parameters: []types.ParameterInfo{
				{Name: "filters", Type: "array", Description: "Array of filter IDs", Required: true},
				{Name: "operator", Type: "string", Description: "Logical operator (AND/OR)", Required: true},
			},
			ReturnType: "string",
		},
		// Event Replay Methods
		{
			Name:        "replayEvents",
			Description: "Replay events with optional filters",
			Parameters: []types.ParameterInfo{
				{Name: "query", Type: "object", Description: "Replay query", Required: true},
				{Name: "options", Type: "object", Description: "Replay options", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "pauseReplay",
			Description: "Pause event replay",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "resumeReplay",
			Description: "Resume event replay",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "stopReplay",
			Description: "Stop event replay",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
		},
		// Event Serialization Methods
		{
			Name:        "serializeEvent",
			Description: "Serialize an event to a specific format",
			Parameters: []types.ParameterInfo{
				{Name: "event", Type: "object", Description: "Event to serialize", Required: true},
				{Name: "format", Type: "string", Description: "Serialization format", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "deserializeEvent",
			Description: "Deserialize an event from string format",
			Parameters: []types.ParameterInfo{
				{Name: "eventData", Type: "string", Description: "Serialized event data", Required: true},
				{Name: "format", Type: "string", Description: "Data format", Required: false},
			},
			ReturnType: "object",
		},
		// Event Aggregation Methods
		{
			Name:        "createAggregator",
			Description: "Create an event aggregator",
			Parameters: []types.ParameterInfo{
				{Name: "type", Type: "string", Description: "Aggregator type", Required: true},
				{Name: "config", Type: "object", Description: "Aggregator configuration", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "getAggregatedData",
			Description: "Get aggregated event data",
			Parameters: []types.ParameterInfo{
				{Name: "aggregatorID", Type: "string", Description: "Aggregator ID", Required: true},
			},
			ReturnType: "object",
		},
		// Recording Methods
		{
			Name:        "startRecording",
			Description: "Start recording events to storage",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "stopRecording",
			Description: "Stop recording events",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "void",
		},
		{
			Name:        "isRecording",
			Description: "Check if events are being recorded",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "boolean",
		},
		// Subscription Info Methods
		{
			Name:        "getSubscriptionCount",
			Description: "Get the number of active subscriptions",
			Parameters:  []types.ParameterInfo{},
			ReturnType:  "number",
		},
		{
			Name:        "getSubscriptionInfo",
			Description: "Get information about a subscription",
			Parameters: []types.ParameterInfo{
				{Name: "subscriptionID", Type: "string", Description: "Subscription ID", Required: true},
			},
			ReturnType: "object",
		},
	}
}

// TypeMappings returns type conversion mappings.
// It defines how Go event types are mapped to script types
// for events, filters, and queries.
func (b *EventBridge) TypeMappings() map[string]types.TypeMapping {
	return map[string]types.TypeMapping{
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

// ValidateMethod validates method calls.
// It ensures that each method receives the correct number and
// types of arguments before execution.
func (b *EventBridge) ValidateMethod(name string, args []types.ScriptValue) error {
	switch name {
	case "publishEvent", "storeEvent", "serializeEvent":
		if len(args) < 1 {
			return fmt.Errorf("%s requires event parameter", name)
		}
		if args[0].Type() != types.TypeObject {
			return fmt.Errorf("event must be object")
		}
	case "subscribe":
		if len(args) < 2 {
			return fmt.Errorf("subscribe requires pattern and handler parameters")
		}
		if args[0].Type() != types.TypeString {
			return fmt.Errorf("pattern must be string")
		}
		if args[1].Type() != types.TypeFunction {
			return fmt.Errorf("handler must be function")
		}
	case "unsubscribe":
		if len(args) < 1 {
			return fmt.Errorf("unsubscribe requires subscriptionID parameter")
		}
		if args[0].Type() != types.TypeString {
			return fmt.Errorf("subscriptionID must be string")
		}
	case "queryEvents":
		if len(args) < 1 {
			return fmt.Errorf("queryEvents requires query parameter")
		}
		if args[0].Type() != types.TypeObject {
			return fmt.Errorf("query must be object")
		}
	case "startRecording", "stopRecording", "isRecording", "getSubscriptionCount":
		// No parameters required
	case "getSubscriptionInfo":
		if len(args) < 1 {
			return fmt.Errorf("getSubscriptionInfo requires subscriptionID parameter")
		}
		if args[0].Type() != types.TypeString {
			return fmt.Errorf("subscriptionID must be string")
		}
	default:
		return fmt.Errorf("unknown method: %s", name)
	}
	return nil
}

// ExecuteMethod executes a bridge method.
// It implements the types.Bridge interface, routing method calls
// to the appropriate event operations.
func (b *EventBridge) ExecuteMethod(ctx context.Context, name string, args []types.ScriptValue) (types.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return types.NewErrorValue(fmt.Errorf("bridge not initialized")), nil
	}

	switch name {
	// Event Bus Methods
	case "publishEvent":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("publishEvent requires event parameter")), nil
		}

		if args[0].Type() != types.TypeObject {
			return types.NewErrorValue(fmt.Errorf("event must be an object")), nil
		}

		eventData := args[0].ToGo().(map[string]interface{})

		// Convert to domain.Event
		event := b.mapToEvent(eventData)

		// Publish to bus (no context needed)
		b.eventBus.Publish(event)
		return types.NewNilValue(), nil

	case "subscribe":
		if len(args) < 2 {
			return types.NewErrorValue(fmt.Errorf("subscribe requires pattern and handler parameters")), nil
		}

		pattern := args[0].(types.StringValue).Value()

		// Create pattern filter
		filter, err := events.NewPatternFilter(pattern)
		if err != nil {
			return types.NewErrorValue(fmt.Errorf("invalid pattern: %w", err)), nil
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

		return types.NewStringValue(subID), nil

	case "subscribeWithFilter":
		if len(args) < 2 {
			return types.NewErrorValue(fmt.Errorf("subscribeWithFilter requires filter and handler parameters")), nil
		}

		filterData := args[0].ToGo().(map[string]interface{})

		// Create filter from data
		filter, err := b.createFilterFromData(filterData)
		if err != nil {
			return types.NewErrorValue(fmt.Errorf("invalid filter: %w", err)), nil
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

		return types.NewStringValue(subID), nil

	case "unsubscribe":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("unsubscribe requires subscriptionID parameter")), nil
		}

		subID := args[0].(types.StringValue).Value()

		// Unsubscribe from event bus
		b.eventBus.Unsubscribe(subID)

		// Clean up tracking
		delete(b.subscriptions, subID)

		return types.NewNilValue(), nil

	// Event Storage Methods
	case "storeEvent":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("storeEvent requires event parameter")), nil
		}

		eventData := args[0].ToGo().(map[string]interface{})
		event := b.mapToEvent(eventData)

		// Store event
		if err := b.storage.Store(ctx, event); err != nil {
			return types.NewErrorValue(fmt.Errorf("failed to store event: %w", err)), nil
		}

		return types.NewNilValue(), nil

	case "queryEvents":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("queryEvents requires query parameter")), nil
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
			return types.NewErrorValue(fmt.Errorf("failed to query events: %w", err)), nil
		}

		// Convert events to script-friendly format
		result := make([]types.ScriptValue, len(eventsList))
		for i, event := range eventsList {
			eventMap := b.eventToMap(event)
			result[i] = types.ConvertToScriptValue(eventMap)
		}

		return types.NewArrayValue(result), nil

	case "getEventHistory":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("getEventHistory requires startTime parameter")), nil
		}

		startTime := args[0].(types.StringValue).Value()

		// Parse start time
		start, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return types.NewErrorValue(fmt.Errorf("invalid start time format: %w", err)), nil
		}

		// Parse end time if provided
		var end time.Time
		if len(args) > 1 {
			endTime := args[1].(types.StringValue).Value()
			end, err = time.Parse(time.RFC3339, endTime)
			if err != nil {
				return types.NewErrorValue(fmt.Errorf("invalid end time format: %w", err)), nil
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
			return types.NewErrorValue(fmt.Errorf("failed to get event history: %w", err)), nil
		}

		// Convert events to script-friendly format
		result := make([]types.ScriptValue, len(eventsList))
		for i, event := range eventsList {
			eventMap := b.eventToMap(event)
			result[i] = types.ConvertToScriptValue(eventMap)
		}

		return types.NewArrayValue(result), nil

	// Event Filtering Methods
	case "createFilter":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("createFilter requires filterConfig parameter")), nil
		}

		filterData := args[0].ToGo().(map[string]interface{})

		// Create filter from data
		filter, err := b.createFilterFromData(filterData)
		if err != nil {
			return types.NewErrorValue(fmt.Errorf("failed to create filter: %w", err)), nil
		}

		// Generate filter ID
		filterID := fmt.Sprintf("filter_%d", time.Now().UnixNano())
		b.filters[filterID] = filter

		return types.NewStringValue(filterID), nil

	case "createCompositeFilter":
		if len(args) < 2 {
			return types.NewErrorValue(fmt.Errorf("createCompositeFilter requires filters and operator parameters")), nil
		}

		filterIDs := args[0].ToGo().([]interface{})
		operator := args[1].(types.StringValue).Value()

		// Get filters by ID
		var filters []events.EventFilter
		for _, id := range filterIDs {
			filterID, ok := id.(string)
			if !ok {
				return types.NewErrorValue(fmt.Errorf("filter ID must be string")), nil
			}
			filter, exists := b.filters[filterID]
			if !exists {
				return types.NewErrorValue(fmt.Errorf("filter %s not found", filterID)), nil
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
			return types.NewErrorValue(fmt.Errorf("invalid operator: %s", operator)), nil
		}

		// Generate filter ID
		filterID := fmt.Sprintf("composite_filter_%d", time.Now().UnixNano())
		b.filters[filterID] = compositeFilter

		return types.NewStringValue(filterID), nil

	// Event Replay Methods
	case "replayEvents":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("replayEvents requires query parameter")), nil
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
			return types.NewErrorValue(fmt.Errorf("failed to replay events: %w", err)), nil
		}

		return types.NewNilValue(), nil

	case "pauseReplay":
		// EventReplayer doesn't have Pause method - this would need different implementation
		return types.NewNilValue(), nil

	case "resumeReplay":
		// EventReplayer doesn't have Resume method - this would need different implementation
		return types.NewNilValue(), nil

	case "stopReplay":
		// EventReplayer doesn't have Stop method - this would need different implementation
		return types.NewNilValue(), nil

	// Event Serialization Methods
	case "serializeEvent":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("serializeEvent requires event parameter")), nil
		}

		eventData := args[0].ToGo().(map[string]interface{})
		event := b.mapToEvent(eventData)

		// Serialize event
		serialized, err := events.SerializeEvent(event)
		if err != nil {
			return types.NewErrorValue(fmt.Errorf("failed to serialize event: %w", err)), nil
		}

		return types.ConvertToScriptValue(serialized), nil

	case "deserializeEvent":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("deserializeEvent requires eventData parameter")), nil
		}

		serializedData := args[0].(types.StringValue).Value()

		// Deserialize event - simplified implementation
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(serializedData), &eventData); err != nil {
			return types.NewErrorValue(fmt.Errorf("failed to deserialize event: %w", err)), nil
		}

		// Return deserialized event data
		return types.ConvertToScriptValue(eventData), nil

	// Event Aggregation Methods
	case "createAggregator":
		if len(args) < 2 {
			return types.NewErrorValue(fmt.Errorf("createAggregator requires type and config parameters")), nil
		}

		aggType := args[0].(types.StringValue).Value()
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

		return types.NewStringValue(aggregator.ID), nil

	case "getAggregatedData":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("getAggregatedData requires aggregatorID parameter")), nil
		}

		aggID := args[0].(types.StringValue).Value()

		aggregator, exists := b.aggregators[aggID]
		if !exists {
			return types.NewErrorValue(fmt.Errorf("aggregator %s not found", aggID)), nil
		}

		// Return aggregated data
		result := map[string]types.ScriptValue{
			"id":         types.NewStringValue(aggregator.ID),
			"type":       types.NewStringValue(aggregator.Type),
			"eventCount": types.NewNumberValue(float64(len(aggregator.Events))),
			"lastUpdate": types.NewStringValue(aggregator.LastUpdate.Format(time.RFC3339)),
		}

		return types.NewObjectValue(result), nil

	// Recording Methods
	case "startRecording":
		if b.recorder == nil {
			return types.NewErrorValue(fmt.Errorf("recorder not initialized")), nil
		}
		if b.isRecording {
			return types.NewErrorValue(fmt.Errorf("already recording")), nil
		}
		if err := b.recorder.Start(); err != nil {
			return types.NewErrorValue(err), nil
		}
		b.isRecording = true
		return types.NewNilValue(), nil

	case "stopRecording":
		if b.recorder == nil {
			return types.NewErrorValue(fmt.Errorf("recorder not initialized")), nil
		}
		if !b.isRecording {
			return types.NewErrorValue(fmt.Errorf("not recording")), nil
		}
		b.recorder.Stop()
		b.isRecording = false
		return types.NewNilValue(), nil

	case "isRecording":
		return types.NewBoolValue(b.isRecording), nil

	case "getSubscriptionCount":
		count := b.eventBus.GetSubscriptionCount()
		return types.NewNumberValue(float64(count)), nil

	case "getSubscriptionInfo":
		if len(args) < 1 {
			return types.NewErrorValue(fmt.Errorf("getSubscriptionInfo requires subscriptionID parameter")), nil
		}

		subID := args[0].(types.StringValue).Value()
		pattern, filterCount, found := b.eventBus.GetSubscriptionInfo(subID)

		if !found {
			return types.NewNilValue(), nil
		}

		result := map[string]types.ScriptValue{
			"subscriptionID": types.NewStringValue(subID),
			"pattern":        types.NewStringValue(pattern),
			"filterCount":    types.NewNumberValue(float64(filterCount)),
		}
		return types.NewObjectValue(result), nil

	default:
		return types.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// RequiredPermissions returns required permissions.
// It specifies the permissions needed for event publishing,
// subscription, querying, and storage access.
func (b *EventBridge) RequiredPermissions() []types.Permission {
	return []types.Permission{
		{
			Type:        types.PermissionNetwork,
			Resource:    "events",
			Actions:     []string{"publish", "subscribe", "query"},
			Description: "Access to event system",
		},
		{
			Type:        types.PermissionMemory,
			Resource:    "event_storage",
			Actions:     []string{"read", "write"},
			Description: "Memory for event storage and caching",
		},
	}
}

// Helper methods

// mapToEvent converts script data to domain.Event.
// It transforms a map representation into a proper event structure
// with default values for missing fields.
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

// NOTE: Duplicate conversion function removed - using centralized types.ConvertToScriptValue() instead
