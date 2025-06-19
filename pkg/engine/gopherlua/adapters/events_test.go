// ABOUTME: Tests for Events bridge adapter that exposes go-llms event system functionality to Lua scripts
// ABOUTME: Validates event bus, subscription, emission, filtering, aggregation, recording, and replay operations

package adapters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

func TestEventsAdapter_Creation(t *testing.T) {
	t.Run("create_events_adapter", func(t *testing.T) {
		// Create events bridge mock
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name:        "Events Bridge",
				Version:     "2.0.0",
				Description: "Event system bridge with go-llms v0.3.5 integration",
			}).
			WithMethod("publishEvent", engine.MethodInfo{
				Name:        "publishEvent",
				Description: "Publish an event to the event bus",
				ReturnType:  "void",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful event publication
				return engine.NewNilValue(), nil
			}).
			WithMethod("subscribe", engine.MethodInfo{
				Name:        "subscribe",
				Description: "Subscribe to events with pattern matching",
				ReturnType:  "string",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock subscription ID
				return engine.NewStringValue("sub-123"), nil
			}).
			WithMethod("queryEvents", engine.MethodInfo{
				Name:        "queryEvents",
				Description: "Query stored events",
				ReturnType:  "array",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return empty array by default
				return engine.NewArrayValue([]engine.ScriptValue{}), nil
			})

		// Create adapter
		adapter := NewEventsAdapter(eventsBridge)
		require.NotNil(t, adapter)

		// Should have event-specific methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "publishEvent")
		assert.Contains(t, methods, "subscribe")
		assert.Contains(t, methods, "unsubscribe")
		assert.Contains(t, methods, "queryEvents")
		assert.Contains(t, methods, "createFilter")
		assert.Contains(t, methods, "replayEvents")
		assert.Contains(t, methods, "startRecording")
		assert.Contains(t, methods, "stopRecording")
	})

	t.Run("events_module_structure", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMetadata(engine.BridgeMetadata{
				Name: "Events Bridge",
			}).
			WithMethod("publishEvent", engine.MethodInfo{
				Name: "publishEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("subscribe", engine.MethodInfo{
				Name: "subscribe",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("sub-123"), nil
			}).
			WithMethod("createFilter", engine.MethodInfo{
				Name: "createFilter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("filter-123"), nil
			}).
			WithMethod("startRecording", engine.MethodInfo{
				Name: "startRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("replayEvents", engine.MethodInfo{
				Name: "replayEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("createAggregator", engine.MethodInfo{
				Name: "createAggregator",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("agg-123"), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Create module
		err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(adapter.CreateLuaModule()),
			NRet:    1,
			Protect: true,
		})
		require.NoError(t, err)

		// Get module
		module := L.Get(-1).(*lua.LTable)
		L.Pop(1)

		// Check standard methods exist
		assert.NotEqual(t, lua.LNil, module.RawGetString("publishEvent"))
		assert.NotEqual(t, lua.LNil, module.RawGetString("subscribe"))

		// Check namespaces exist
		bus := module.RawGetString("bus")
		assert.NotEqual(t, lua.LNil, bus, "bus namespace should exist")

		filters := module.RawGetString("filters")
		assert.NotEqual(t, lua.LNil, filters, "filters namespace should exist")

		recording := module.RawGetString("recording")
		assert.NotEqual(t, lua.LNil, recording, "recording namespace should exist")

		replay := module.RawGetString("replay")
		assert.NotEqual(t, lua.LNil, replay, "replay namespace should exist")

		aggregation := module.RawGetString("aggregation")
		assert.NotEqual(t, lua.LNil, aggregation, "aggregation namespace should exist")
	})
}

func TestEventsAdapter_EventPublication(t *testing.T) {
	t.Run("publish_simple_event", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("publishEvent", engine.MethodInfo{
				Name: "publishEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Validate event data
				if len(args) >= 1 && args[0].Type() == engine.TypeObject {
					eventData := args[0].(engine.ObjectValue).Fields()
					if eventType, ok := eventData["type"]; ok && eventType.(engine.StringValue).Value() == "test_event" {
						return engine.NewNilValue(), nil
					}
				}
				return nil, fmt.Errorf("invalid event data")
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Publish event from Lua
		err = L.DoString(`
			local events = require("events")
			local result, err = events.publishEvent({
				type = "test_event",
				data = { message = "Hello, World!" },
				timestamp = "2024-01-01T00:00:00Z"
			})
			assert(err == nil, "publish should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})

	t.Run("publish_event_via_bus_namespace", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("publishEvent", engine.MethodInfo{
				Name: "publishEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test bus namespace methods
		err = L.DoString(`
			local events = require("events")
			
			-- Publish through bus namespace
			local result, err = events.bus.publish({
				type = "bus_event",
				data = { source = "lua_script" }
			})
			assert(err == nil, "bus publish should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventSubscription(t *testing.T) {
	t.Run("subscribe_to_pattern", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("subscribe", engine.MethodInfo{
				Name: "subscribe",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[0].Type() == engine.TypeString {
					pattern := args[0].(engine.StringValue).Value()
					return engine.NewStringValue("sub-" + pattern), nil
				}
				return nil, fmt.Errorf("invalid subscription arguments")
			}).
			WithMethod("unsubscribe", engine.MethodInfo{
				Name: "unsubscribe",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test subscription
		err = L.DoString(`
			local events = require("events")
			
			-- Subscribe to pattern
			local handler = function(event)
				print("Received event:", event.type)
			end
			
			local subId, err = events.subscribe("test.*", handler)
			assert(err == nil, "subscribe should not error: " .. tostring(err))
			assert(subId == "sub-test.*", "should return correct subscription ID")
			
			-- Unsubscribe
			local result, unsubErr = events.unsubscribe(subId)
			assert(unsubErr == nil, "unsubscribe should not error: " .. tostring(unsubErr))
		`)
		assert.NoError(t, err)
	})

	t.Run("subscribe_with_filter", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("subscribeWithFilter", engine.MethodInfo{
				Name: "subscribeWithFilter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[0].Type() == engine.TypeObject {
					return engine.NewStringValue("sub-filter-123"), nil
				}
				return nil, fmt.Errorf("invalid filter subscription arguments")
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test filter subscription
		err = L.DoString(`
			local events = require("events")
			
			-- Subscribe with custom filter
			local filter = {
				type = "type",
				eventType = "user_action"
			}
			
			local handler = function(event)
				print("Filtered event:", event.type)
			end
			
			local subId, err = events.subscribeWithFilter(filter, handler)
			assert(err == nil, "subscribeWithFilter should not error: " .. tostring(err))
			assert(subId == "sub-filter-123", "should return filter subscription ID")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventFiltering(t *testing.T) {
	t.Run("create_pattern_filter", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("createFilter", engine.MethodInfo{
				Name: "createFilter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 1 && args[0].Type() == engine.TypeObject {
					filterData := args[0].(engine.ObjectValue).Fields()
					if filterType, ok := filterData["type"]; ok {
						return engine.NewStringValue("filter-" + filterType.(engine.StringValue).Value()), nil
					}
				}
				return nil, fmt.Errorf("invalid filter configuration")
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test filter creation
		err = L.DoString(`
			local events = require("events")
			
			-- Create pattern filter
			local filterId, err = events.filters.create({
				type = "pattern",
				pattern = "user.*"
			})
			assert(err == nil, "filter creation should not error: " .. tostring(err))
			assert(filterId == "filter-pattern", "should return filter ID")
		`)
		assert.NoError(t, err)
	})

	t.Run("create_composite_filter", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("createFilter", engine.MethodInfo{
				Name: "createFilter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewStringValue("filter-" + time.Now().Format("150405")), nil
			}).
			WithMethod("createCompositeFilter", engine.MethodInfo{
				Name: "createCompositeFilter",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[0].Type() == engine.TypeArray {
					return engine.NewStringValue("composite-filter-123"), nil
				}
				return nil, fmt.Errorf("invalid composite filter arguments")
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test composite filter creation
		err = L.DoString(`
			local events = require("events")
			
			-- Create individual filters
			local filter1, err1 = events.filters.create({
				type = "pattern",
				pattern = "user.*"
			})
			assert(err1 == nil, "first filter creation should not error")
			
			local filter2, err2 = events.filters.create({
				type = "type",
				eventType = "action"
			})
			assert(err2 == nil, "second filter creation should not error")
			
			-- Create composite filter
			local compositeId, err = events.filters.createComposite({filter1, filter2}, "AND")
			assert(err == nil, "composite filter creation should not error: " .. tostring(err))
			assert(compositeId == "composite-filter-123", "should return composite filter ID")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventQuery(t *testing.T) {
	t.Run("query_events", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("queryEvents", engine.MethodInfo{
				Name: "queryEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return mock events
				events := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":        engine.NewStringValue("event-1"),
						"type":      engine.NewStringValue("user_login"),
						"timestamp": engine.NewStringValue("2024-01-01T10:00:00Z"),
					}),
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":        engine.NewStringValue("event-2"),
						"type":      engine.NewStringValue("user_logout"),
						"timestamp": engine.NewStringValue("2024-01-01T11:00:00Z"),
					}),
				}
				return engine.NewArrayValue(events), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test event query
		err = L.DoString(`
			local events = require("events")
			
			-- Query events (arrays return as multiple values)
			local event1, event2 = events.queryEvents({
				agentID = "test-agent",
				limit = 10
			})
			assert(event1 ~= nil, "first event should exist")
			assert(event2 ~= nil, "second event should exist")
			assert(event1.type == "user_login", "first event should be login")
			assert(event2.type == "user_logout", "second event should be logout")
		`)
		assert.NoError(t, err)
	})

	t.Run("get_event_history", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("getEventHistory", engine.MethodInfo{
				Name: "getEventHistory",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return time-filtered events
				events := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"id":        engine.NewStringValue("hist-1"),
						"type":      engine.NewStringValue("historical_event"),
						"timestamp": engine.NewStringValue("2024-01-01T12:00:00Z"),
					}),
				}
				return engine.NewArrayValue(events), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test event history
		err = L.DoString(`
			local events = require("events")
			
			-- Get event history (arrays return as multiple values)
			local histEvent = events.getEventHistory(
				"2024-01-01T00:00:00Z",
				"2024-01-01T23:59:59Z"
			)
			assert(histEvent ~= nil, "history event should exist")
			assert(histEvent.type == "historical_event", "should be historical event")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventRecording(t *testing.T) {
	t.Run("start_stop_recording", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("startRecording", engine.MethodInfo{
				Name: "startRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("stopRecording", engine.MethodInfo{
				Name: "stopRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("isRecording", engine.MethodInfo{
				Name: "isRecording",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewBoolValue(true), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test recording operations
		err = L.DoString(`
			local events = require("events")
			
			-- Start recording
			local result, err = events.recording.start()
			assert(err == nil, "start recording should not error: " .. tostring(err))
			
			-- Check recording status
			local recording, err = events.recording.isRecording()
			assert(err == nil, "isRecording should not error: " .. tostring(err))
			assert(recording == true, "should be recording")
			
			-- Stop recording
			local result, err = events.recording.stop()
			assert(err == nil, "stop recording should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventReplay(t *testing.T) {
	t.Run("replay_events", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("replayEvents", engine.MethodInfo{
				Name: "replayEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Mock successful replay
				return engine.NewNilValue(), nil
			}).
			WithMethod("pauseReplay", engine.MethodInfo{
				Name: "pauseReplay",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("resumeReplay", engine.MethodInfo{
				Name: "resumeReplay",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			}).
			WithMethod("stopReplay", engine.MethodInfo{
				Name: "stopReplay",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNilValue(), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test replay operations
		err = L.DoString(`
			local events = require("events")
			
			-- Start replay
			local result, err = events.replay.start({
				agentID = "test-agent",
				startTime = "2024-01-01T00:00:00Z"
			}, {
				speed = 2.0
			})
			assert(err == nil, "replay should not error: " .. tostring(err))
			
			-- Pause replay
			local result, err = events.replay.pause()
			assert(err == nil, "pause should not error: " .. tostring(err))
			
			-- Resume replay
			local result, err = events.replay.resume()
			assert(err == nil, "resume should not error: " .. tostring(err))
			
			-- Stop replay
			local result, err = events.replay.stop()
			assert(err == nil, "stop should not error: " .. tostring(err))
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventAggregation(t *testing.T) {
	t.Run("create_aggregator", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("createAggregator", engine.MethodInfo{
				Name: "createAggregator",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 2 && args[0].Type() == engine.TypeString {
					aggType := args[0].(engine.StringValue).Value()
					return engine.NewStringValue("agg-" + aggType + "-123"), nil
				}
				return nil, fmt.Errorf("invalid aggregator arguments")
			}).
			WithMethod("getAggregatedData", engine.MethodInfo{
				Name: "getAggregatedData",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return aggregated data
				result := map[string]engine.ScriptValue{
					"id":         engine.NewStringValue("agg-count-123"),
					"type":       engine.NewStringValue("count"),
					"eventCount": engine.NewNumberValue(42),
					"lastUpdate": engine.NewStringValue("2024-01-01T12:00:00Z"),
				}
				return engine.NewObjectValue(result), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test aggregation
		err = L.DoString(`
			local events = require("events")
			
			-- Create aggregator
			local aggId, err = events.aggregation.create("count", {
				windowSize = 300
			})
			assert(err == nil, "create aggregator should not error: " .. tostring(err))
			assert(aggId == "agg-count-123", "should return aggregator ID")
			
			-- Get aggregated data
			local data, err = events.aggregation.getData(aggId)
			assert(err == nil, "get data should not error: " .. tostring(err))
			assert(data.type == "count", "should be count aggregator")
			assert(data.eventCount == 42, "should have 42 events")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventSerialization(t *testing.T) {
	t.Run("serialize_deserialize_event", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("serializeEvent", engine.MethodInfo{
				Name: "serializeEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return serialized event
				serialized := `{"id":"event-123","type":"test_event","timestamp":"2024-01-01T12:00:00Z"}`
				return engine.NewStringValue(serialized), nil
			}).
			WithMethod("deserializeEvent", engine.MethodInfo{
				Name: "deserializeEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return deserialized event
				event := map[string]engine.ScriptValue{
					"id":        engine.NewStringValue("event-123"),
					"type":      engine.NewStringValue("test_event"),
					"timestamp": engine.NewStringValue("2024-01-01T12:00:00Z"),
				}
				return engine.NewObjectValue(event), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test serialization
		err = L.DoString(`
			local events = require("events")
			
			-- Serialize event
			local event = {
				id = "event-123",
				type = "test_event",
				timestamp = "2024-01-01T12:00:00Z"
			}
			
			local serialized, err = events.serializeEvent(event)
			assert(err == nil, "serialize should not error: " .. tostring(err))
			assert(type(serialized) == "string", "should return string")
			
			-- Deserialize event
			local deserialized, err = events.deserializeEvent(serialized)
			assert(err == nil, "deserialize should not error: " .. tostring(err))
			assert(deserialized.id == "event-123", "should deserialize correctly")
			assert(deserialized.type == "test_event", "should preserve event type")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_SubscriptionInfo(t *testing.T) {
	t.Run("subscription_count_and_info", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("getSubscriptionCount", engine.MethodInfo{
				Name: "getSubscriptionCount",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return engine.NewNumberValue(3), nil
			}).
			WithMethod("getSubscriptionInfo", engine.MethodInfo{
				Name: "getSubscriptionInfo",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				if len(args) >= 1 && args[0].Type() == engine.TypeString {
					subID := args[0].(engine.StringValue).Value()
					info := map[string]engine.ScriptValue{
						"subscriptionID": engine.NewStringValue(subID),
						"pattern":        engine.NewStringValue("user.*"),
						"filterCount":    engine.NewNumberValue(1),
					}
					return engine.NewObjectValue(info), nil
				}
				return engine.NewNilValue(), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test subscription info
		err = L.DoString(`
			local events = require("events")
			
			-- Get subscription count
			local count, err = events.getSubscriptionCount()
			assert(err == nil, "getSubscriptionCount should not error: " .. tostring(err))
			assert(count == 3, "should have 3 subscriptions")
			
			-- Get subscription info
			local info, err = events.getSubscriptionInfo("sub-123")
			assert(err == nil, "getSubscriptionInfo should not error: " .. tostring(err))
			assert(info.subscriptionID == "sub-123", "should return subscription ID")
			assert(info.pattern == "user.*", "should return pattern")
			assert(info.filterCount == 1, "should return filter count")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_ErrorHandling(t *testing.T) {
	t.Run("handle_bridge_errors", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("publishEvent", engine.MethodInfo{
				Name: "publishEvent",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				return nil, fmt.Errorf("event bus unavailable")
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test error handling
		err = L.DoString(`
			local events = require("events")
			
			local result, err = events.publishEvent({
				type = "test_event"
			})
			assert(result == nil, "result should be nil on error")
			assert(string.find(err, "event bus unavailable"), "should contain error message")
		`)
		assert.NoError(t, err)
	})
}

func TestEventsAdapter_EventCorrelation(t *testing.T) {
	t.Run("correlate_events", func(t *testing.T) {
		eventsBridge := testutils.NewMockBridge("events").
			WithInitialized(true).
			WithMethod("correlateEvents", engine.MethodInfo{
				Name: "correlateEvents",
			}, func(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
				// Return correlated event groups
				correlatedGroups := []engine.ScriptValue{
					engine.NewObjectValue(map[string]engine.ScriptValue{
						"correlationId": engine.NewStringValue("corr-123"),
						"events": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("event-1"),
							engine.NewStringValue("event-2"),
						}),
						"confidence": engine.NewNumberValue(0.95),
					}),
				}
				return engine.NewArrayValue(correlatedGroups), nil
			})

		adapter := NewEventsAdapter(eventsBridge)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "events")
		require.NoError(t, err)

		err = ms.LoadModule(L, "events")
		require.NoError(t, err)

		// Test event correlation
		err = L.DoString(`
			local events = require("events")
			
			-- Correlate events (arrays return as multiple values)
			local correlation = events.correlateEvents({
				timeWindow = 300,
				algorithm = "temporal"
			})
			assert(correlation ~= nil, "correlation should exist")
			assert(correlation.correlationId == "corr-123", "should have correlation ID")
			assert(correlation.confidence == 0.95, "should have confidence score")
		`)
		assert.NoError(t, err)
	})
}
