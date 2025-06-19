// ABOUTME: Tests for engine integration features including event bus, type registry, profiling, and API export
// ABOUTME: Comprehensive test coverage for enhanced engine capabilities using go-llms infrastructure

package engine

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Task 1.4.11.1: Engine Event Bus
func TestDefaultEventBus(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, eventBus *DefaultEventBus)
	}{
		{
			name: "Subscribe and publish events",
			test: func(t *testing.T, eb *DefaultEventBus) {
				var receivedEvent *EngineEvent
				handler := func(event EngineEvent) error {
					receivedEvent = &event
					return nil
				}

				// Subscribe to events
				subID, err := eb.Subscribe("test.*", handler)
				require.NoError(t, err)
				assert.NotEmpty(t, subID)

				// Publish an event
				event := EngineEvent{
					ID:        "test-1",
					Type:      "test.execution",
					Source:    "engine",
					Timestamp: time.Now(),
					Data:      "test data",
				}

				err = eb.Publish(event)
				require.NoError(t, err)

				// Verify event was received
				assert.NotNil(t, receivedEvent)
				assert.Equal(t, event.ID, receivedEvent.ID)
				assert.Equal(t, event.Type, receivedEvent.Type)
			},
		},
		{
			name: "Event priorities",
			test: func(t *testing.T, eb *DefaultEventBus) {
				var executionOrder []string

				handler1 := func(event EngineEvent) error {
					executionOrder = append(executionOrder, "handler1")
					return nil
				}
				handler2 := func(event EngineEvent) error {
					executionOrder = append(executionOrder, "handler2")
					return nil
				}

				// Subscribe with different priorities
				sub1, err := eb.Subscribe("test.*", handler1)
				require.NoError(t, err)
				sub2, err := eb.Subscribe("test.*", handler2)
				require.NoError(t, err)

				// Set priorities (higher priority executes first)
				err = eb.SetPriority(sub1, 1)
				require.NoError(t, err)
				err = eb.SetPriority(sub2, 2)
				require.NoError(t, err)

				// Publish event
				event := EngineEvent{
					ID:   "test-priority",
					Type: "test.priority",
				}
				err = eb.Publish(event)
				require.NoError(t, err)

				// Verify execution order
				assert.Equal(t, []string{"handler2", "handler1"}, executionOrder)
			},
		},
		{
			name: "Unsubscribe",
			test: func(t *testing.T, eb *DefaultEventBus) {
				var eventCount int
				handler := func(event EngineEvent) error {
					eventCount++
					return nil
				}

				subID, err := eb.Subscribe("test.*", handler)
				require.NoError(t, err)

				// Publish first event
				event1 := EngineEvent{ID: "test-1", Type: "test.event"}
				err = eb.Publish(event1)
				require.NoError(t, err)
				assert.Equal(t, 1, eventCount)

				// Unsubscribe
				err = eb.Unsubscribe(subID)
				require.NoError(t, err)

				// Publish second event (should not be received)
				event2 := EngineEvent{ID: "test-2", Type: "test.event"}
				err = eb.Publish(event2)
				require.NoError(t, err)
				assert.Equal(t, 1, eventCount) // Still 1
			},
		},
		{
			name: "Get subscriptions",
			test: func(t *testing.T, eb *DefaultEventBus) {
				handler := func(event EngineEvent) error { return nil }

				sub1, err := eb.Subscribe("pattern1", handler)
				require.NoError(t, err)
				sub2, err := eb.Subscribe("pattern2", handler)
				require.NoError(t, err)

				subscriptions := eb.GetSubscriptions()
				assert.Len(t, subscriptions, 2)

				// Find our subscriptions
				var found1, found2 bool
				for _, sub := range subscriptions {
					if sub.ID == sub1 && sub.Pattern == "pattern1" {
						found1 = true
					}
					if sub.ID == sub2 && sub.Pattern == "pattern2" {
						found2 = true
					}
				}
				assert.True(t, found1)
				assert.True(t, found2)
			},
		},
		{
			name: "Clear subscriptions",
			test: func(t *testing.T, eb *DefaultEventBus) {
				handler := func(event EngineEvent) error { return nil }

				_, err := eb.Subscribe("test1", handler)
				require.NoError(t, err)
				_, err = eb.Subscribe("test2", handler)
				require.NoError(t, err)

				subscriptions := eb.GetSubscriptions()
				assert.Len(t, subscriptions, 2)

				err = eb.Clear()
				require.NoError(t, err)

				subscriptions = eb.GetSubscriptions()
				assert.Len(t, subscriptions, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := NewDefaultEventBus()
			tt.test(t, eventBus)
		})
	}
}

// Test Task 1.4.11.2: Type Conversion Registry
func TestDefaultTypeRegistry(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, registry *DefaultTypeRegistry)
	}{
		{
			name: "Register and convert types",
			test: func(t *testing.T, tr *DefaultTypeRegistry) {
				// Register string to int converter
				converter := func(value interface{}) (interface{}, error) {
					if str, ok := value.(string); ok {
						if str == "42" {
							return 42, nil
						}
					}
					return nil, assert.AnError
				}

				err := tr.Register("string", "int", converter)
				require.NoError(t, err)

				// Test conversion
				result, err := tr.Convert("42", "string", "int")
				require.NoError(t, err)
				assert.Equal(t, 42, result)

				// Test CanConvert
				assert.True(t, tr.CanConvert("string", "int"))
				assert.False(t, tr.CanConvert("int", "string"))
			},
		},
		{
			name: "Bidirectional conversion",
			test: func(t *testing.T, tr *DefaultTypeRegistry) {
				stringToInt := func(value interface{}) (interface{}, error) {
					if str, ok := value.(string); ok && str == "42" {
						return 42, nil
					}
					return nil, assert.AnError
				}

				intToString := func(value interface{}) (interface{}, error) {
					if i, ok := value.(int); ok && i == 42 {
						return "42", nil
					}
					return nil, assert.AnError
				}

				err := tr.RegisterBidirectional("string", "int", stringToInt, intToString)
				require.NoError(t, err)

				// Test both directions
				result1, err := tr.Convert("42", "string", "int")
				require.NoError(t, err)
				assert.Equal(t, 42, result1)

				result2, err := tr.Convert(42, "int", "string")
				require.NoError(t, err)
				assert.Equal(t, "42", result2)
			},
		},
		{
			name: "Export documentation",
			test: func(t *testing.T, tr *DefaultTypeRegistry) {
				converter := func(value interface{}) (interface{}, error) {
					return value, nil
				}

				err := tr.Register("typeA", "typeB", converter)
				require.NoError(t, err)

				doc, err := tr.ExportDocumentation()
				require.NoError(t, err)

				var parsed map[string]interface{}
				err = json.Unmarshal(doc, &parsed)
				require.NoError(t, err)

				assert.Equal(t, "Type Conversion Registry", parsed["title"])
				assert.Contains(t, parsed, "converters")
				assert.Contains(t, parsed, "generated")
			},
		},
		{
			name: "Cache operations",
			test: func(t *testing.T, tr *DefaultTypeRegistry) {
				err := tr.ClearCache()
				require.NoError(t, err)

				converters := tr.GetConverters()
				assert.NotNil(t, converters)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewDefaultTypeRegistry()
			tt.test(t, registry)
		})
	}
}

// Test Task 1.4.11.3: Engine Profiling
func TestDefaultEngineProfiler(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, profiler *DefaultEngineProfiler)
	}{
		{
			name: "Enable and disable profiling",
			test: func(t *testing.T, ep *DefaultEngineProfiler) {
				config := ProfilingConfig{
					Enabled:        true,
					CPUProfiling:   false, // Disable for test
					MemProfiling:   false, // Disable for test
					TraceProfiling: false,
					SampleRate:     100,
					OutputDir:      "/tmp",
					Duration:       time.Second,
				}

				err := ep.Enable(config)
				require.NoError(t, err)

				// Simulate some work
				time.Sleep(10 * time.Millisecond)

				err = ep.Disable()
				require.NoError(t, err)
			},
		},
		{
			name: "Generate profiling report",
			test: func(t *testing.T, ep *DefaultEngineProfiler) {
				config := ProfilingConfig{
					Enabled:      true,
					CPUProfiling: false, // Disable for test
					MemProfiling: false, // Disable for test
				}

				err := ep.Enable(config)
				require.NoError(t, err)

				// Simulate some work
				time.Sleep(10 * time.Millisecond)

				report, err := ep.GetReport()
				require.NoError(t, err)

				assert.NotZero(t, report.Duration)
				assert.NotNil(t, report.MemoryStats)
				assert.NotNil(t, report.Metrics)
				assert.True(t, report.EndTime.After(report.StartTime))

				err = ep.Disable()
				require.NoError(t, err)
			},
		},
		{
			name: "Optimization hints generation",
			test: func(t *testing.T, ep *DefaultEngineProfiler) {
				config := ProfilingConfig{
					Enabled:      true,
					CPUProfiling: false,
					MemProfiling: false,
				}

				err := ep.Enable(config)
				require.NoError(t, err)

				report, err := ep.GetReport()
				require.NoError(t, err)

				// Check that optimization hints are generated
				assert.NotNil(t, report.Optimizations)
				// Note: The actual hints depend on memory usage patterns

				err = ep.Disable()
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiler := NewDefaultEngineProfiler()
			tt.test(t, profiler)
		})
	}
}

// Test Task 1.4.11.4: API Export
func TestDefaultAPIExporter(t *testing.T) {
	// Create a mock engine for testing
	mockEngine := NewMockScriptEngine()
	err := mockEngine.Initialize(EngineConfig{})
	require.NoError(t, err)

	// Add a mock bridge
	mockBridge := NewMockBridge(
		"test-bridge",
		BridgeMetadata{
			Name:        "Test Bridge",
			Version:     "1.0.0",
			Description: "A test bridge",
		},
		[]MethodInfo{
			{
				Name:        "testMethod",
				Description: "A test method",
				Parameters: []ParameterInfo{
					{Name: "param1", Type: "string", Required: true},
				},
				ReturnType: "string",
				Examples:   []string{"testMethod('hello')"},
			},
		},
	)
	err = mockEngine.RegisterBridge(mockBridge)
	require.NoError(t, err)

	tests := []struct {
		name string
		test func(t *testing.T, exporter *DefaultAPIExporter)
	}{
		{
			name: "Export API as JSON",
			test: func(t *testing.T, ae *DefaultAPIExporter) {
				data, err := ae.ExportAPI(mockEngine, ExportFormatJSON)
				require.NoError(t, err)

				var parsed []interface{}
				err = json.Unmarshal(data, &parsed)
				require.NoError(t, err)

				assert.NotEmpty(t, parsed)
			},
		},
		{
			name: "Export API as Markdown",
			test: func(t *testing.T, ae *DefaultAPIExporter) {
				data, err := ae.ExportAPI(mockEngine, ExportFormatMarkdown)
				require.NoError(t, err)

				content := string(data)
				assert.NotEmpty(t, content)
				assert.True(t, strings.Contains(content, "Test Bridge") ||
					strings.Contains(content, "test-bridge"))
			},
		},
		{
			name: "Generate client library",
			test: func(t *testing.T, ae *DefaultAPIExporter) {
				options := ClientLibraryOptions{
					PackageName:  "test-client",
					Version:      "1.0.0",
					IncludeTypes: true,
					IncludeDocs:  true,
				}

				data, err := ae.GenerateClientLibrary(mockEngine, "javascript", options)
				require.NoError(t, err)

				var parsed map[string]interface{}
				err = json.Unmarshal(data, &parsed)
				require.NoError(t, err)

				assert.Equal(t, "javascript", parsed["language"])
				assert.Equal(t, "test-client", parsed["packageName"])
				assert.Equal(t, "1.0.0", parsed["version"])
			},
		},
		{
			name: "Unsupported format error",
			test: func(t *testing.T, ae *DefaultAPIExporter) {
				_, err := ae.ExportAPI(mockEngine, ExportFormat("unsupported"))
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported export format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := NewDefaultAPIExporter()
			tt.test(t, exporter)
		})
	}
}

// Mock implementations for testing - using test helpers

// MockScriptEngine using the test helper
type MockScriptEngine struct {
	*testMockScriptEngine
}

func NewMockScriptEngine() *MockScriptEngine {
	return &MockScriptEngine{
		testMockScriptEngine: newTestMockScriptEngine("MockEngine"),
	}
}

// MockBridge using the test helper
type MockBridge struct {
	*testMockBridge
}

func NewMockBridge(id string, metadata BridgeMetadata, methods []MethodInfo) *MockBridge {
	mock := newTestMockBridge(id)
	mock.metadata = metadata
	mock.methods = methods
	return &MockBridge{
		testMockBridge: mock,
	}
}
