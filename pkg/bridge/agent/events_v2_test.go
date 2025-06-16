// ABOUTME: Tests for event system bridge v2.0.0 with go-llms v0.3.5 integration
// ABOUTME: Verifies event bus, storage, filtering, serialization, aggregation, and replay functionality

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

func TestNewEventBridgeV2(t *testing.T) {
	bridge := NewEventBridgeV2()
	assert.NotNil(t, bridge)
	assert.Equal(t, "events", bridge.GetID())
	assert.NotNil(t, bridge.eventBus)
	assert.NotNil(t, bridge.storage)
	assert.NotNil(t, bridge.recorder)
	assert.NotNil(t, bridge.replayer)
}

func TestEventBridgeV2Metadata(t *testing.T) {
	bridge := NewEventBridgeV2()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "events", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "v2.0.0")
	assert.Contains(t, metadata.Description, "v0.3.5")
	assert.Contains(t, metadata.Description, "aggregation")
	assert.Contains(t, metadata.Description, "replay")
	assert.NotEmpty(t, metadata.Author)
	assert.NotEmpty(t, metadata.License)
}

func TestEventBridgeV2Initialization(t *testing.T) {
	tests := []struct {
		name    string
		bridge  *EventBridgeV2
		wantErr bool
	}{
		{
			name:    "successful initialization",
			bridge:  NewEventBridgeV2(),
			wantErr: false,
		},
		{
			name:    "double initialization",
			bridge:  &EventBridgeV2{initialized: true},
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

func TestEventBridgeV2Methods(t *testing.T) {
	bridge := NewEventBridgeV2()
	methods := bridge.Methods()

	// Essential v2.0.0 methods
	expectedMethods := []string{
		// Event Bus
		"publishEvent",
		"subscribe",
		"subscribeWithFilter",
		"unsubscribe",
		// Event Storage
		"storeEvent",
		"queryEvents",
		"streamEvents",
		"stopStream",
		// Event Filtering
		"createPatternFilter",
		"createTypeFilter",
		"createCompositeFilter",
		"createFieldFilter",
		// Event Serialization
		"serializeEvent",
		"deserializeEvent",
		"serializeEventBatch",
		// Event Replay
		"startEventRecording",
		"stopEventRecording",
		"replayEvents",
		"pauseReplay",
		"resumeReplay",
		// Event Aggregation
		"createAggregator",
		"getAggregatedData",
		"resetAggregator",
		"removeAggregator",
		// Bridge Events
		"publishBridgeEvent",
		"onBridgeEvent",
		// Event Streams
		"createEventStream",
		"filterStream",
		"mapStream",
		"reduceStream",
		// Utilities
		"getEventStats",
		"exportEventStore",
		"importEventStore",
	}

	methodMap := make(map[string]bool)
	for _, method := range methods {
		methodMap[method.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Missing expected method: %s", expected)
	}

	// Verify method count increased from v1.0.0
	assert.GreaterOrEqual(t, len(methods), 30, "Should have at least 30 methods in v2.0.0")
}

func TestEventBridgeV2TypeMappings(t *testing.T) {
	bridge := NewEventBridgeV2()
	mappings := bridge.TypeMappings()

	expectedTypes := []string{
		"Event",
		"EventType",
		"EventFilter",
		"EventQuery",
		"EventStream",
		"BridgeEvent",
		"ReplayOptions",
		"EventSerializer",
		"EventAggregator",
	}

	for _, typeName := range expectedTypes {
		mapping, exists := mappings[typeName]
		assert.True(t, exists, "Missing type mapping for %s", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestEventBridgeV2RequiredPermissions(t *testing.T) {
	bridge := NewEventBridgeV2()
	permissions := bridge.RequiredPermissions()

	assert.NotEmpty(t, permissions)

	// Should require event system permissions
	hasEventPermission := false
	hasStoragePermission := false
	hasFilePermission := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionProcess && perm.Resource == "events" {
			hasEventPermission = true
			assert.Contains(t, perm.Actions, "publish")
			assert.Contains(t, perm.Actions, "subscribe")
			assert.Contains(t, perm.Actions, "filter")
			assert.Contains(t, perm.Actions, "aggregate")
			assert.Contains(t, perm.Actions, "replay")
		}
		if perm.Type == engine.PermissionMemory && perm.Resource == "event_storage" {
			hasStoragePermission = true
		}
		if perm.Type == engine.PermissionFileSystem && perm.Resource == "event_files" {
			hasFilePermission = true
		}
	}

	assert.True(t, hasEventPermission, "Missing event permission")
	assert.True(t, hasStoragePermission, "Missing storage permission")
	assert.True(t, hasFilePermission, "Missing file permission")
}

func TestEventBridgeV2_EventBusOperations(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("publish and subscribe", func(t *testing.T) {
		// Subscribe to events
		result, err := bridge.ExecuteMethod(ctx, "subscribe", []interface{}{
			"test.*",
			func(event domain.Event) {},
		})
		require.NoError(t, err)
		subID, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, subID)

		// Publish event
		_, err = bridge.ExecuteMethod(ctx, "publishEvent", []interface{}{
			map[string]interface{}{
				"type":    "test.event",
				"agentID": "agent1",
				"data":    map[string]interface{}{"key": "value"},
			},
		})
		assert.NoError(t, err)

		// Unsubscribe
		_, err = bridge.ExecuteMethod(ctx, "unsubscribe", []interface{}{subID})
		assert.NoError(t, err)
	})

	t.Run("subscribe with filter", func(t *testing.T) {
		// Create filter config
		filterConfig := map[string]interface{}{
			"type":    "pattern",
			"pattern": "agent.*",
		}

		result, err := bridge.ExecuteMethod(ctx, "subscribeWithFilter", []interface{}{
			filterConfig,
			func(event domain.Event) {},
		})
		require.NoError(t, err)
		subID, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, subID)
	})
}

func TestEventBridgeV2_EventStorage(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("store and query events", func(t *testing.T) {
		// Store event
		_, err := bridge.ExecuteMethod(ctx, "storeEvent", []interface{}{
			map[string]interface{}{
				"type":    "test.stored",
				"agentID": "agent1",
				"data":    map[string]interface{}{"value": 42},
			},
		})
		assert.NoError(t, err)

		// Query events
		result, err := bridge.ExecuteMethod(ctx, "queryEvents", []interface{}{
			map[string]interface{}{
				"agentID": "agent1",
				"limit":   10,
			},
		})
		assert.NoError(t, err)

		events, ok := result.([]map[string]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(events), 1)
	})
}

func TestEventBridgeV2_EventFiltering(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("create pattern filter", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "createPatternFilter", []interface{}{
			"agent.*",
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("create type filter", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "createTypeFilter", []interface{}{
			[]interface{}{"agent.started", "agent.stopped"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("create composite filter", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "createCompositeFilter", []interface{}{
			"AND",
			[]interface{}{
				map[string]interface{}{
					"type":    "pattern",
					"pattern": "agent.*",
				},
				map[string]interface{}{
					"type":    "agent",
					"agentID": "agent1",
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestEventBridgeV2_EventSerialization(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	testEvent := map[string]interface{}{
		"type":    "test.serialize",
		"agentID": "agent1",
		"data":    map[string]interface{}{"key": "value"},
	}

	t.Run("serialize and deserialize event", func(t *testing.T) {
		// Serialize
		result, err := bridge.ExecuteMethod(ctx, "serializeEvent", []interface{}{
			testEvent,
			"json",
		})
		require.NoError(t, err)
		serialized, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, serialized)

		// Verify it's valid JSON
		var jsonData map[string]interface{}
		err = json.Unmarshal([]byte(serialized), &jsonData)
		assert.NoError(t, err)

		// Deserialize
		result, err = bridge.ExecuteMethod(ctx, "deserializeEvent", []interface{}{
			serialized,
			"json",
		})
		assert.NoError(t, err)

		deserialized, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test.serialize", deserialized["type"])
		assert.Equal(t, "agent1", deserialized["agentID"])
	})

	t.Run("serialize event batch", func(t *testing.T) {
		events := []interface{}{
			testEvent,
			map[string]interface{}{
				"type":    "test.batch",
				"agentID": "agent2",
			},
		}

		result, err := bridge.ExecuteMethod(ctx, "serializeEventBatch", []interface{}{events})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestEventBridgeV2_EventReplay(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("recording lifecycle", func(t *testing.T) {
		// Stop any existing recording first
		_, _ = bridge.ExecuteMethod(ctx, "stopEventRecording", []interface{}{})

		// Start recording
		_, err := bridge.ExecuteMethod(ctx, "startEventRecording", []interface{}{})
		assert.NoError(t, err)

		// Store some events
		for i := 0; i < 3; i++ {
			_, err = bridge.ExecuteMethod(ctx, "storeEvent", []interface{}{
				map[string]interface{}{
					"type":    fmt.Sprintf("test.replay.%d", i),
					"agentID": "replay-agent",
					"data":    map[string]interface{}{"index": i},
				},
			})
			assert.NoError(t, err)
		}

		// Stop recording
		_, err = bridge.ExecuteMethod(ctx, "stopEventRecording", []interface{}{})
		assert.NoError(t, err)

		// Replay events
		result, err := bridge.ExecuteMethod(ctx, "replayEvents", []interface{}{
			map[string]interface{}{
				"speed": 2.0,
			},
		})
		assert.NoError(t, err)
		sessionID, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, sessionID)
	})
}

func TestEventBridgeV2_EventAggregation(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("aggregator lifecycle", func(t *testing.T) {
		// Create aggregator
		result, err := bridge.ExecuteMethod(ctx, "createAggregator", []interface{}{
			"test-aggregator",
			map[string]interface{}{
				"window": 5.0, // 5 seconds
				"filter": map[string]interface{}{
					"type":    "pattern",
					"pattern": "metrics.*",
				},
			},
		})
		require.NoError(t, err)
		aggID, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, aggID)

		// Get aggregated data
		result, err = bridge.ExecuteMethod(ctx, "getAggregatedData", []interface{}{aggID})
		assert.NoError(t, err)
		data, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, aggID, data["id"])
		assert.Equal(t, "test-aggregator", data["name"])
		assert.Equal(t, 5.0, data["window"])

		// Reset aggregator
		_, err = bridge.ExecuteMethod(ctx, "resetAggregator", []interface{}{aggID})
		assert.NoError(t, err)

		// Remove aggregator
		_, err = bridge.ExecuteMethod(ctx, "removeAggregator", []interface{}{aggID})
		assert.NoError(t, err)
	})
}

func TestEventBridgeV2_BridgeEvents(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("publish bridge event", func(t *testing.T) {
		_, err := bridge.ExecuteMethod(ctx, "publishBridgeEvent", []interface{}{
			"script_executed",
			map[string]interface{}{
				"script": "test.lua",
				"result": "success",
			},
		})
		assert.NoError(t, err)
	})
}

func TestEventBridgeV2_EventStreams(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("create event stream", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "createEventStream", []interface{}{
			"stream.*",
		})
		require.NoError(t, err)
		streamID, ok := result.(string)
		assert.True(t, ok)
		assert.NotEmpty(t, streamID)
	})
}

func TestEventBridgeV2_Utilities(t *testing.T) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(t, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	t.Run("get event stats", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "getEventStats", []interface{}{})
		assert.NoError(t, err)

		stats, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, stats, "storage")
		assert.Contains(t, stats, "bus")
		assert.Contains(t, stats, "aggregators")
		assert.Contains(t, stats, "streams")
	})
}

func TestEventBridgeV2_ConcurrentAccess(t *testing.T) {
	bridge := NewEventBridgeV2()
	ctx := context.Background()

	// Initialize bridge
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Publisher goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := bridge.ExecuteMethod(ctx, "publishEvent", []interface{}{
					map[string]interface{}{
						"type":    fmt.Sprintf("concurrent.test.%d", id),
						"agentID": fmt.Sprintf("agent-%d", id),
						"data":    map[string]interface{}{"iteration": j},
					},
				})
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// Subscriber goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err := bridge.ExecuteMethod(ctx, "subscribe", []interface{}{
				fmt.Sprintf("concurrent.test.%d", id),
				func(event domain.Event) {},
			})
			if err != nil {
				errors <- err
				return
			}

			// Unsubscribe after short delay
			time.Sleep(50 * time.Millisecond)
			if subID, ok := result.(string); ok {
				_, err = bridge.ExecuteMethod(ctx, "unsubscribe", []interface{}{subID})
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}

	// Cleanup
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
}

func TestEventBridgeV2_ErrorHandling(t *testing.T) {
	bridge := NewEventBridgeV2()
	ctx := context.Background()

	t.Run("methods fail when not initialized", func(t *testing.T) {
		methods := []string{
			"publishEvent",
			"subscribe",
			"storeEvent",
			"queryEvents",
		}

		for _, method := range methods {
			_, err := bridge.ExecuteMethod(ctx, method, []interface{}{})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not initialized")
		}
	})

	t.Run("invalid method parameters", func(t *testing.T) {
		require.NoError(t, bridge.Initialize(ctx))
		defer func() {
			_ = bridge.Cleanup(ctx)
		}()

		// publishEvent with invalid event
		_, err := bridge.ExecuteMethod(ctx, "publishEvent", []interface{}{
			"not an object",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be an object")

		// subscribe without handler
		_, err = bridge.ExecuteMethod(ctx, "subscribe", []interface{}{
			"test.*",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires pattern and handler")

		// Invalid filter config
		_, err = bridge.ExecuteMethod(ctx, "createCompositeFilter", []interface{}{
			"INVALID_OP",
			[]interface{}{},
		})
		assert.Error(t, err)
	})

	t.Run("unknown method", func(t *testing.T) {
		require.NoError(t, bridge.Initialize(ctx))
		defer func() {
			_ = bridge.Cleanup(ctx)
		}()

		_, err := bridge.ExecuteMethod(ctx, "unknownMethod", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not found")
	})
}

// Benchmark tests
func BenchmarkEventBridgeV2_PublishEvent(b *testing.B) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(b, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	event := map[string]interface{}{
		"type":    "benchmark.event",
		"agentID": "bench-agent",
		"data":    map[string]interface{}{"value": 42},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bridge.ExecuteMethod(ctx, "publishEvent", []interface{}{event})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventBridgeV2_EventSerialization(b *testing.B) {
	ctx := context.Background()
	bridge := NewEventBridgeV2()
	require.NoError(b, bridge.Initialize(ctx))
	defer func() {
		_ = bridge.Cleanup(ctx)
	}()

	event := map[string]interface{}{
		"type":    "benchmark.serialize",
		"agentID": "bench-agent",
		"data":    map[string]interface{}{"nested": map[string]interface{}{"value": 42}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := bridge.ExecuteMethod(ctx, "serializeEvent", []interface{}{event, "json"})
		if err != nil {
			b.Fatal(err)
		}

		serialized := result.(string)
		_, err = bridge.ExecuteMethod(ctx, "deserializeEvent", []interface{}{serialized, "json"})
		if err != nil {
			b.Fatal(err)
		}
	}
}
