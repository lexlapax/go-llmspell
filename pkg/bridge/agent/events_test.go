package agent

import (
	"context"
	"testing"
	"time"

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

	assert.Equal(t, "events", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Event system bridge")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestEventsBridge_Methods(t *testing.T) {
	bridge := NewEventBridge()
	methods := bridge.Methods()

	// Should have all expected methods
	expectedMethods := []string{
		"publishEvent", "subscribe", "subscribeWithFilter", "unsubscribe",
		"storeEvent", "queryEvents", "getEventHistory", 
		"createFilter", "createCompositeFilter",
		"serializeEvent", "deserializeEvent",
		"replayEvents", "pauseReplay", "resumeReplay", "stopReplay",
		"createAggregator", "getAggregatedData",
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
			name:        "valid publishEvent",
			method:      "publishEvent",
			args:        []engine.ScriptValue{engine.NewObjectValue(map[string]engine.ScriptValue{"type": engine.NewStringValue("test")})},
			expectError: false,
		},
		{
			name:        "invalid publishEvent - missing args",
			method:      "publishEvent",
			args:        []engine.ScriptValue{},
			expectError: true,
		},
		{
			name:        "valid subscribe",
			method:      "subscribe",
			args:        []engine.ScriptValue{engine.NewStringValue("test-event"), engine.NewFunctionValue("handler", func(args []engine.ScriptValue) (engine.ScriptValue, error) { return engine.NewNilValue(), nil })},
			expectError: false,
		},
		{
			name:        "valid queryEvents",
			method:      "queryEvents",
			args:        []engine.ScriptValue{engine.NewObjectValue(map[string]engine.ScriptValue{})},
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

func TestEventsBridge_ExecuteMethod_PublishEvent(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test publish event
	eventData := map[string]engine.ScriptValue{
		"type":    engine.NewStringValue("test-event"),
		"message": engine.NewStringValue("Hello World"),
		"count":   engine.NewNumberValue(42),
	}

	args := []engine.ScriptValue{
		engine.NewObjectValue(eventData),
	}

	result, err := bridge.ExecuteMethod(ctx, "publishEvent", args)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from publishEvent")
}

func TestEventsBridge_ExecuteMethod_Subscribe(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test subscribe to event
	eventPattern := "test-event"
	handler := engine.NewFunctionValue("handler", func(args []engine.ScriptValue) (engine.ScriptValue, error) {
		return engine.NewNilValue(), nil
	})

	args := []engine.ScriptValue{
		engine.NewStringValue(eventPattern),
		handler,
	}

	result, err := bridge.ExecuteMethod(ctx, "subscribe", args)
	assert.NoError(t, err)

	stringValue, ok := result.(engine.StringValue)
	assert.True(t, ok, "Expected StringValue (subscription ID) from subscribe")
	assert.NotEmpty(t, stringValue.Value(), "Subscription ID should not be empty")
}

func TestEventsBridge_ExecuteMethod_GetSubscriptionCount(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test getSubscriptionCount - should return 0 initially
	result, err := bridge.ExecuteMethod(ctx, "getSubscriptionCount", []engine.ScriptValue{})
	assert.NoError(t, err)

	numberValue, ok := result.(engine.NumberValue)
	assert.True(t, ok, "Expected NumberValue from getSubscriptionCount")
	assert.Equal(t, float64(0), numberValue.Value(), "Expected 0 subscriptions initially")
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

func TestEventsBridge_ExecuteMethod_QueryEvents(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Start recording
	_, err = bridge.ExecuteMethod(ctx, "startRecording", []engine.ScriptValue{})
	require.NoError(t, err)

	// Publish an event
	eventArgs := []engine.ScriptValue{
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"type": engine.NewStringValue("test-event"),
			"data": engine.NewStringValue("test"),
		}),
	}
	_, err = bridge.ExecuteMethod(ctx, "publishEvent", eventArgs)
	require.NoError(t, err)

	// Give a moment for the event to be processed and stored
	time.Sleep(10 * time.Millisecond)

	// Query events
	queryArgs := []engine.ScriptValue{
		engine.NewObjectValue(map[string]engine.ScriptValue{
			"limit": engine.NewNumberValue(10),
		}),
	}
	result, err := bridge.ExecuteMethod(ctx, "queryEvents", queryArgs)
	assert.NoError(t, err)

	arrayValue, ok := result.(engine.ArrayValue)
	assert.True(t, ok, "Expected ArrayValue from queryEvents")

	events := arrayValue.ToGo().([]interface{})
	assert.GreaterOrEqual(t, len(events), 1, "Should have at least one event")
}


func TestEventsBridge_ExecuteMethod_Unsubscribe(t *testing.T) {
	bridge := NewEventBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Subscribe first
	subscribeArgs := []engine.ScriptValue{
		engine.NewStringValue("test-event"),
		engine.NewFunctionValue("handler", func(args []engine.ScriptValue) (engine.ScriptValue, error) {
			return engine.NewNilValue(), nil
		}),
	}

	subscribeResult, err := bridge.ExecuteMethod(ctx, "subscribe", subscribeArgs)
	require.NoError(t, err)

	subscriptionID := subscribeResult.(engine.StringValue).Value()

	// Now unsubscribe
	unsubscribeArgs := []engine.ScriptValue{engine.NewStringValue(subscriptionID)}
	result, err := bridge.ExecuteMethod(ctx, "unsubscribe", unsubscribeArgs)
	assert.NoError(t, err)

	_, ok := result.(engine.NilValue)
	assert.True(t, ok, "Expected NilValue from unsubscribe")
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
	expectedTypes := []string{"Event", "EventFilter", "EventQuery"}
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
	result, err := bridge.ExecuteMethod(ctx, "publishEvent", []engine.ScriptValue{
		engine.NewObjectValue(map[string]engine.ScriptValue{"type": engine.NewStringValue("test")}),
	})
	assert.NoError(t, err) // Should return error value, not Go error

	errorValue, ok := result.(engine.ErrorValue)
	assert.True(t, ok, "Expected ErrorValue when not initialized")
	assert.Contains(t, errorValue.Error().Error(), "not initialized")
}
