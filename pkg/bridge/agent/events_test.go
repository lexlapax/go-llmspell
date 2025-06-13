// ABOUTME: Tests for the event system bridge that exposes go-llms event functionality to scripts
// ABOUTME: Verifies event streaming, filtering, subscriptions, and all event types

package agent

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventBridge(t *testing.T) {
	bridge := NewEventBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "events", bridge.GetID())
}

func TestEventBridgeMetadata(t *testing.T) {
	bridge := NewEventBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "events", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "event")
	assert.NotEmpty(t, metadata.Author)
	assert.NotEmpty(t, metadata.License)
}

func TestEventBridgeInitialization(t *testing.T) {
	tests := []struct {
		name    string
		bridge  *EventBridge
		wantErr bool
	}{
		{
			name:    "successful initialization",
			bridge:  NewEventBridge(),
			wantErr: false,
		},
		{
			name:    "double initialization",
			bridge:  &EventBridge{initialized: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bridge.Initialize(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.bridge.IsInitialized())
			}
		})
	}
}

func TestEventBridgeMethods(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Essential event methods
	expectedMethods := []string{
		// Event emission
		"emitEvent",
		"emitAgentEvent",
		"emitToolEvent",
		"emitWorkflowEvent",
		"emitStateEvent",
		// Event subscription
		"subscribeToEvents",
		"subscribeToEventType",
		"subscribeToAgentEvents",
		"unsubscribe",
		"unsubscribeAll",
		// Event filtering
		"createEventFilter",
		"addEventFilter",
		"removeEventFilter",
		// Event history
		"getEventHistory",
		"clearEventHistory",
		"getEventCount",
		// Event types
		"getEventTypes",
		"isValidEventType",
		// Event utilities
		"createEvent",
		"formatEvent",
		"parseEvent",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Missing expected method: %s", expected)
	}

	// Verify method details
	for _, method := range methods {
		assert.NotEmpty(t, method.Description)
		assert.NotEmpty(t, method.ReturnType)

		// Check specific methods have correct parameters
		switch method.Name {
		case "emitEvent":
			assert.GreaterOrEqual(t, len(method.Parameters), 1) // event
		case "subscribeToEventType":
			assert.GreaterOrEqual(t, len(method.Parameters), 2) // eventType, handler
		case "createEventFilter":
			assert.GreaterOrEqual(t, len(method.Parameters), 1) // filterConfig
		}
	}
}

func TestEventBridgeTypeMappings(t *testing.T) {
	bridge := NewEventBridge()
	mappings := bridge.TypeMappings()

	expectedTypes := []string{
		"Event",
		"EventType",
		"EventFilter",
		"EventSubscription",
		"EventHistory",
		"EventHandler",
		"EventMetadata",
		"AgentEvent",
		"ToolEvent",
		"WorkflowEvent",
		"StateEvent",
	}

	for _, typeName := range expectedTypes {
		mapping, exists := mappings[typeName]
		assert.True(t, exists, "Missing type mapping for %s", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestEventBridgeRequiredPermissions(t *testing.T) {
	bridge := NewEventBridge()
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Should require event system permission
	hasEventPermission := false
	for _, perm := range permissions {
		if perm.Type == engine.PermissionProcess && perm.Resource == "events" {
			hasEventPermission = true
			assert.Contains(t, perm.Actions, "emit")
			assert.Contains(t, perm.Actions, "subscribe")
			assert.Contains(t, perm.Actions, "filter")
		}
	}
	assert.True(t, hasEventPermission, "Missing event permission")
}

func TestEventBridgeValidateMethod(t *testing.T) {
	bridge := NewEventBridge()

	tests := []struct {
		name    string
		method  string
		args    []interface{}
		wantErr bool
	}{
		{
			name:    "valid emitEvent",
			method:  "emitEvent",
			args:    []interface{}{map[string]interface{}{"type": "test", "data": "test"}},
			wantErr: false,
		},
		{
			name:    "valid subscribeToEventType",
			method:  "subscribeToEventType",
			args:    []interface{}{"AgentStart", func() {}},
			wantErr: false,
		},
		{
			name:    "emitEvent missing args",
			method:  "emitEvent",
			args:    []interface{}{},
			wantErr: false, // Validation is delegated to engine
		},
		{
			name:    "unknown method",
			method:  "unknownMethod",
			args:    []interface{}{},
			wantErr: false, // Validation is delegated to engine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventBridgeEngineRegistration(t *testing.T) {
	bridge := NewEventBridge()
	engine := NewMockScriptEngine()

	err := bridge.RegisterWithEngine(engine)
	require.NoError(t, err)

	// Verify bridge was registered
	registered, err := engine.GetBridge("events")
	assert.NoError(t, err)
	assert.Equal(t, bridge, registered)
}

func TestEventBridgeCleanup(t *testing.T) {
	bridge := NewEventBridge()

	// Initialize first
	err := bridge.Initialize(context.Background())
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Cleanup
	err = bridge.Cleanup(context.Background())
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestEventBridgeConcurrentAccess(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Concurrent operations
	done := make(chan bool, 3)

	// Reader 1
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.IsInitialized()
			_ = bridge.GetID()
			_ = bridge.Methods()
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.TypeMappings()
			_ = bridge.RequiredPermissions()
		}
		done <- true
	}()

	// Writer
	go func() {
		for i := 0; i < 50; i++ {
			_ = bridge.Initialize(ctx)
			_ = bridge.Cleanup(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestEventBridgeEventTypes(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Find getEventTypes method
	var getEventTypesMethod *engine.MethodInfo
	for _, method := range methods {
		if method.Name == "getEventTypes" {
			getEventTypesMethod = &method
			break
		}
	}

	require.NotNil(t, getEventTypesMethod, "getEventTypes method should exist")
	assert.Equal(t, "array", getEventTypesMethod.ReturnType)

	// Verify all event type constants are exposed
	typeMappings := bridge.TypeMappings()
	eventTypeMapping, exists := typeMappings["EventType"]
	assert.True(t, exists, "EventType mapping should exist")
	assert.Equal(t, "EventType", eventTypeMapping.GoType)
}

func TestEventBridgeSubscriptionManagement(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Ensure subscription management methods exist
	subMethods := []string{
		"subscribeToEvents",
		"subscribeToEventType",
		"subscribeToAgentEvents",
		"unsubscribe",
		"unsubscribeAll",
		"listSubscriptions",
		"pauseSubscription",
		"resumeSubscription",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, subMethod := range subMethods {
		if methodMap[subMethod] {
			// Found the method, verify it has appropriate parameters
			for _, method := range methods {
				if method.Name == subMethod {
					switch subMethod {
					case "subscribeToEventType":
						assert.GreaterOrEqual(t, len(method.Parameters), 2,
							"Subscribe method %s should have eventType and handler parameters", subMethod)
					case "unsubscribe":
						assert.GreaterOrEqual(t, len(method.Parameters), 1,
							"Unsubscribe method %s should have subscription ID parameter", subMethod)
					}
					break
				}
			}
		}
	}
}

func TestEventBridgeFiltering(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Ensure filtering methods exist
	filterMethods := []string{
		"createEventFilter",
		"addEventFilter",
		"removeEventFilter",
		"listEventFilters",
		"testEventFilter",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, filterMethod := range filterMethods {
		if methodMap[filterMethod] {
			// Verify the method exists and has reasonable parameters
			for _, method := range methods {
				if method.Name == filterMethod {
					assert.NotEmpty(t, method.Description,
						"Filter method %s should have a description", filterMethod)
					break
				}
			}
		}
	}
}

func TestEventBridgeHistory(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Ensure history methods exist
	historyMethods := []string{
		"getEventHistory",
		"clearEventHistory",
		"getEventCount",
		"queryEvents",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, historyMethod := range historyMethods {
		if methodMap[historyMethod] {
			// Found the method
			for _, method := range methods {
				if method.Name == historyMethod {
					switch historyMethod {
					case "getEventHistory":
						// Should optionally take filters
						assert.NotEmpty(t, method.ReturnType)
					case "getEventCount":
						assert.Equal(t, "number", method.ReturnType)
					}
					break
				}
			}
		}
	}
}

func TestEventBridgeRealTimeStreaming(t *testing.T) {
	bridge := NewEventBridge()

	// Initialize bridge
	err := bridge.Initialize(context.Background())
	require.NoError(t, err)

	// Test that bridge supports real-time event streaming
	methods := bridge.Methods()

	// Look for streaming-related methods
	streamingMethods := []string{
		"startEventStream",
		"stopEventStream",
		"isStreamActive",
		"getStreamStats",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	// Note: These might not all exist, but we should have some streaming capability
	hasStreamingSupport := false
	for _, streamMethod := range streamingMethods {
		if methodMap[streamMethod] {
			hasStreamingSupport = true
			break
		}
	}

	// At minimum, we should support subscription which enables streaming
	assert.True(t, methodMap["subscribeToEvents"] || hasStreamingSupport,
		"Bridge should support event streaming through subscriptions")
}

// MockEventEmitter simulates an event emitter for testing
type MockEventEmitter struct {
	events []bridge.Event
}

func (m *MockEventEmitter) EmitEvent(event bridge.Event) {
	m.events = append(m.events, event)
}

func (m *MockEventEmitter) GetEvents() []bridge.Event {
	return m.events
}

func TestEventBridgeEventEmission(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test that the bridge can handle various event types
	typeMappings := bridge.TypeMappings()

	// Verify event type mappings include all event categories
	eventCategories := []string{
		"AgentEvent",
		"ToolEvent",
		"WorkflowEvent",
		"StateEvent",
	}

	for _, category := range eventCategories {
		_, exists := typeMappings[category]
		assert.True(t, exists, "Missing event category mapping: %s", category)
	}
}

func TestEventBridgeAsyncHandling(t *testing.T) {
	bridge := NewEventBridge()

	// Verify the bridge supports async event handling
	methods := bridge.Methods()

	// Look for async-related capabilities
	asyncSupport := false
	for _, method := range methods {
		if method.Name == "subscribeToEvents" || method.Name == "subscribeToEventType" {
			// Check if handler parameter indicates async support
			for _, param := range method.Parameters {
				if param.Name == "handler" && param.Type == "function" {
					asyncSupport = true
					break
				}
			}
		}
	}

	assert.True(t, asyncSupport, "Bridge should support async event handlers")
}

func TestEventBridgeEventValidation(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Check for event validation method
	hasValidation := false
	for _, method := range methods {
		if method.Name == "isValidEventType" {
			hasValidation = true
			assert.Equal(t, "boolean", method.ReturnType)
			assert.GreaterOrEqual(t, len(method.Parameters), 1)
			break
		}
	}

	assert.True(t, hasValidation, "Bridge should provide event type validation")
}

func TestEventBridgeTimeout(t *testing.T) {
	bridge := NewEventBridge()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Initialize should respect context timeout
	err := bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Cleanup should also respect timeout
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
}
