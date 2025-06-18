package agent

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventsBridge_Initialize(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())
}

func TestEventsBridge_GetID(t *testing.T) {
	bridge := NewEventBridge()
	assert.Equal(t, "events", bridge.GetID())
}

func TestEventsBridge_GetMetadata(t *testing.T) {
	bridge := NewEventBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "Events Bridge", metadata.Name)
	assert.Equal(t, "2.1.0", metadata.Version)
	assert.Contains(t, metadata.Description, "event system")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestEventsBridge_Methods(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"emit", "subscribe", "unsubscribe", "once", "listListeners",
		"removeAllListeners", "hasListeners", "getMaxListeners", "setMaxListeners",
		"getListenerCount", "startRecording", "stopRecording", "getRecordedEvents",
		"clearRecordedEvents", "replayEvents", "pauseRecording", "resumeRecording",
		"isRecording", "createEventFilter", "setEventFilter", "removeEventFilter",
		"listEventFilters", "enableEventPersistence", "disableEventPersistence",
		"saveEvents", "loadEvents", "getEventStats", "clearEventStats",
		"exportEvents", "importEvents", "createEventBatch", "emitEventBatch",
		"scheduleEvent", "cancelScheduledEvent", "listScheduledEvents",
	}

	assert.GreaterOrEqual(t, len(methods), len(expectedMethods))

	// Check that key methods exist
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodNames[expected], "Expected method %s not found", expected)
	}
}

func TestEventsBridge_ValidateMethod(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		args        []engine.ScriptValue
		expectError bool
	}{
		{
			name:        "valid emit",
			method:      "emit",
			args:        []engine.ScriptValue{engine.NewStringValue("test-event"), engine.NewObjectValue(map[string]engine.ScriptValue{})},
			expectError: false,
		},
		{
			name:        "invalid emit - missing args",
			method:      "emit",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
		{
			name:        "valid subscribe",
			method:      "subscribe",
			args:        []engine.ScriptValue{engine.NewStringValue("test-event"), engine.NewStringValue("handler")},
			expectError: false,
		},
		{
			name:        "valid listListeners",
			method:      "listListeners",
			args:        []engine.ScriptValue{},
			expectError: false,
		},
		{
			name:        "unknown method",
			method:      "unknownMethod",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.ValidateMethod(tt.method, tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventsBridge_ExecuteMethod_Emit(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test emit event
	eventType := "test-event"
	eventData := map[string]engine.ScriptValue{
		"message": engine.NewStringValue("Hello World"),
		"count":   engine.NewNumberValue(42),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue(eventType),
		engine.NewObjectValue(eventData),
	}

	result, err := bridge.ExecuteMethod(ctx, "emit", args)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from emit")
}

func TestEventsBridge_ExecuteMethod_Subscribe(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test subscribe to event
	eventType := "test-event"
	handler := "test-handler-function"

	args := []engine.ScriptValue{
		engine.NewStringValue(eventType),
		engine.NewStringValue(handler),
	}

	result, err := bridge.ExecuteMethod(ctx, "subscribe", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (subscription ID) from subscribe")
	assert.NotEmpty(t, stringValue.Value(), "Subscription ID should not be empty")
}

func TestEventsBridge_ExecuteMethod_ListListeners(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test listListeners - should work even with no listeners
	result, err := bridge.ExecuteMethod(ctx, "listListeners", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from listListeners")
	assert.Equal(t, 0, len(arrayValue.ToGo().([]interface{})), "Expected empty array initially")
}

func TestEventsBridge_ExecuteMethod_StartStopRecording(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test startRecording
	result, err := bridge.ExecuteMethod(ctx, "startRecording", []engine.ScriptValue{})
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from startRecording")

	// Test isRecording
	result, err = bridge.ExecuteMethod(ctx, "isRecording", []engine.ScriptValue{})
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from isRecording")
	assert.True(t, boolValue.Value(), "Should be recording after startRecording")

	// Test stopRecording
	result, err = bridge.ExecuteMethod(ctx, "stopRecording", []engine.ScriptValue{})
	assert.NoError(t, err)

	_, ok = result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from stopRecording")

	// Test isRecording again
	result, err = bridge.ExecuteMethod(ctx, "isRecording", []engine.ScriptValue{})
	assert.NoError(t, err)

	boolValue, ok = result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from isRecording")
	assert.False(t, boolValue.Value(), "Should not be recording after stopRecording")
}

func TestEventsBridge_ExecuteMethod_GetRecordedEvents(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Start recording
	_, err = bridge.ExecuteMethod(ctx, "startRecording", []engine.ScriptValue{})
	require.NoError(t, err)

	// Emit an event
	eventArgs := []engine.ScriptValue{
		engine.NewStringValue("test-event"),
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"data": engine.NewStringValue("test"),
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "emit", eventArgs)
	require.NoError(t, err)

	// Get recorded events
	result, err := bridge.ExecuteMethod(ctx, "getRecordedEvents", []engine.ScriptValue{})
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from getRecordedEvents")

	events := arrayValue.ToGo().([]interface{})
	assert.Greater(t, len(events), 0, "Should have recorded events")
}

func TestEventsBridge_ExecuteMethod_GetEventStats(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Get stats
	result, err := bridge.ExecuteMethod(ctx, "getEventStats", []engine.ScriptValue{})
	assert.NoError(t, err)

	objectValue, ok := result.(engine.ObjectValue)
	assert.True(t, ok, "Expected ObjectValue from getEventStats")

	stats := objectValue.ToGo().(map[string]interface{})
	assert.Contains(t, stats, "total_events")
	assert.Contains(t, stats, "total_listeners")
}

func TestEventsBridge_ExecuteMethod_SetMaxListeners(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Set max listeners
	args := []engine.ScriptValue{engine.NewNumberValue(50)}
	result, err := bridge.ExecuteMethod(ctx, "setMaxListeners", args)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from setMaxListeners")

	// Get max listeners
	result, err = bridge.ExecuteMethod(ctx, "getMaxListeners", []engine.ScriptValue{})
	assert.NoError(t, err)

	numberValue, ok := result.(engine.NumberValue)
	assert.True(t, ok, "Expected NumberValue from getMaxListeners")
	assert.Equal(t, float64(50), numberValue.Value(), "Max listeners should be 50")
}

func TestEventsBridge_ExecuteMethod_CreateEventFilter(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create event filter
	filterConfig := map[string]engine.ScriptValue{
		"eventTypes": engine.NewArrayValue([]engine.ScriptValue{
			engine.NewStringValue("test-event"),
		}),
		"priority": engine.NewNumberValue(1),
	}

	args := []engine.ScriptValue{
		engine.NewStringValue("test-filter"),
		engine.NewObjectValue(filterConfig),
	}

	result, err := bridge.ExecuteMethod(ctx, "createEventFilter", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (filter ID) from createEventFilter")
	assert.Equal(t, "test-filter", stringValue.Value(), "Filter ID should match input")
}

func TestEventsBridge_ExecuteMethod_Unsubscribe(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Subscribe first
	subscribeArgs := []engine.ScriptValue{
		engine.NewStringValue("test-event"),
		engine.NewStringValue("handler"),
	}

	subscribeResult, err := bridge.ExecuteMethod(ctx, "subscribe", subscribeArgs)
	require.NoError(t, err)

	subscriptionID := subscribeResult.(engine.StringValue).Value()

	// Now unsubscribe
	unsubscribeArgs := []engine.ScriptValue{engine.NewStringValue(subscriptionID)}
	result, err := bridge.ExecuteMethod(ctx, "unsubscribe", unsubscribeArgs)
	assert.NoError(t, err)

	boolValue, ok := result.(engine.BoolValue)
	assert.True(t, ok, "Expected BoolValue from unsubscribe")
	assert.True(t, boolValue.Value(), "Unsubscribe should succeed")
}

func TestEventsBridge_ExecuteMethod_UnknownMethod(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	result, err := bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue for unknown method")
	assert.Contains(t, errorValue.Error().Error(), "unknown method")
}

func TestEventsBridge_RequiredPermissions(t *testing.T) {
	bridge := NewEventBridge()
	permissions := bridge.RequiredPermissions()

	assert.Greater(t, len(permissions), 0, "Should have required permissions")

	// Check for expected permission types
	hasEventsPermission := false
	for _, perm := range permissions {
		if perm.Resource == "events" {
			hasEventsPermission = true
			break
		}
	}
	assert.True(t, hasEventsPermission, "Should have events permission")
}

func TestEventsBridge_TypeMappings(t *testing.T) {
	bridge := NewEventBridge()
	mappings := bridge.TypeMappings()

	assert.Greater(t, len(mappings), 0, "Should have type mappings")

	// Check for expected mappings
	expectedTypes := []string{"Event", "EventFilter", "EventListener"}
	for _, expectedType := range expectedTypes {
		_, exists := mappings[expectedType]
		assert.True(t, exists, "Expected type mapping for %s", expectedType)
	}
}

func TestEventsBridge_Cleanup(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()

	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestEventsBridge_NotInitialized(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()

	// Should fail when not initialized
	result, err := bridge.ExecuteMethod(ctx, "listListeners", []engine.ScriptValue{})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
