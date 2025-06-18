// ABOUTME: Tests for tracing bridge functionality including distributed tracing and span management
// ABOUTME: Comprehensive test coverage for tracing hooks and OpenTelemetry-compatible tracing interfaces

package observability

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for tracing functionality
	"github.com/lexlapax/go-llms/pkg/agent/core"
)

// Test TracingBridge core functionality
func TestTracingBridge(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, bridge *TracingBridge)
	}{
		{
			name: "Bridge initialization",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)
				assert.True(t, bridge.IsInitialized())

				metadata := bridge.GetMetadata()
				assert.Equal(t, "tracing", metadata.Name)
				assert.Equal(t, "v1.0.0", metadata.Version)
				assert.Contains(t, metadata.Description, "distributed tracing")
			},
		},
		{
			name: "Create tracer",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test createTracer method
				params := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				result, err := bridge.ExecuteMethod(ctx, "createTracer", params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				tracerInfo, ok := result.(engine.ObjectValue)
				require.True(t, ok)
				tracerMap := tracerInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test-tracer", tracerMap["name"])
				assert.NotEmpty(t, tracerMap["id"])
			},
		},
		{
			name: "Start and end span",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create tracer first
				tracerParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				tracerResult, err := bridge.ExecuteMethod(ctx, "createTracer", tracerParams)
				require.NoError(t, err)
				tracerInfo, ok := tracerResult.(engine.ObjectValue)
				require.True(t, ok)
				tracerMap := tracerInfo.ToGo().(map[string]interface{})
				tracerID := tracerMap["id"].(string)

				// Start span
				spanParams := []engine.ScriptValue{
					engine.NewStringValue(tracerID),
					engine.NewStringValue("test-operation"),
				}
				spanResult, err := bridge.ExecuteMethod(ctx, "startSpan", spanParams)
				require.NoError(t, err)
				assert.NotNil(t, spanResult)

				spanInfo, ok := spanResult.(engine.ObjectValue)
				require.True(t, ok)
				spanMap := spanInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "test-operation", spanMap["name"])
				spanID := spanMap["id"].(string)

				// End span
				endParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
				}
				_, err = bridge.ExecuteMethod(ctx, "endSpan", endParams)
				require.NoError(t, err)
			},
		},
		{
			name: "Set span attributes",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create tracer and span
				tracerParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				tracerResult, err := bridge.ExecuteMethod(ctx, "createTracer", tracerParams)
				require.NoError(t, err)
				tracerInfo := tracerResult.(engine.ObjectValue)
				tracerID := tracerInfo.ToGo().(map[string]interface{})["id"].(string)

				spanParams := []engine.ScriptValue{
					engine.NewStringValue(tracerID),
					engine.NewStringValue("test-operation"),
				}
				spanResult, err := bridge.ExecuteMethod(ctx, "startSpan", spanParams)
				require.NoError(t, err)
				spanInfo := spanResult.(engine.ObjectValue)
				spanID := spanInfo.ToGo().(map[string]interface{})["id"].(string)

				// Set attributes
				attributes := map[string]engine.ScriptValue{
					"operation.type": engine.NewStringValue("test"),
					"user.id":        engine.NewStringValue("123"),
					"request.size":   engine.NewNumberValue(1024),
				}
				attrParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
					engine.NewObjectValue(attributes),
				}
				_, err = bridge.ExecuteMethod(ctx, "setSpanAttributes", attrParams)
				require.NoError(t, err)

				// End span
				endParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
				}
				_, err = bridge.ExecuteMethod(ctx, "endSpan", endParams)
				require.NoError(t, err)
			},
		},
		{
			name: "Record span error",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create tracer and span
				tracerParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				tracerResult, err := bridge.ExecuteMethod(ctx, "createTracer", tracerParams)
				require.NoError(t, err)
				tracerInfo := tracerResult.(engine.ObjectValue)
				tracerID := tracerInfo.ToGo().(map[string]interface{})["id"].(string)

				spanParams := []engine.ScriptValue{
					engine.NewStringValue(tracerID),
					engine.NewStringValue("test-operation"),
				}
				spanResult, err := bridge.ExecuteMethod(ctx, "startSpan", spanParams)
				require.NoError(t, err)
				spanInfo := spanResult.(engine.ObjectValue)
				spanID := spanInfo.ToGo().(map[string]interface{})["id"].(string)

				// Record error
				errorParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
					engine.NewStringValue("test error message"),
				}
				_, err = bridge.ExecuteMethod(ctx, "recordSpanError", errorParams)
				require.NoError(t, err)

				// Set error status
				statusParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
					engine.NewStringValue("error"),
					engine.NewStringValue("Operation failed"),
				}
				_, err = bridge.ExecuteMethod(ctx, "setSpanStatus", statusParams)
				require.NoError(t, err)

				// End span
				endParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
				}
				_, err = bridge.ExecuteMethod(ctx, "endSpan", endParams)
				require.NoError(t, err)
			},
		},
		{
			name: "Create tracing hooks",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create agent tracing hook
				agentParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				agentHookResult, err := bridge.ExecuteMethod(ctx, "createAgentTracingHook", agentParams)
				require.NoError(t, err)
				assert.NotNil(t, agentHookResult)

				agentHookInfo, ok := agentHookResult.(engine.ObjectValue)
				require.True(t, ok)
				agentHookMap := agentHookInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "agent", agentHookMap["type"])

				// Create tool call tracing hook
				toolParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				toolHookResult, err := bridge.ExecuteMethod(ctx, "createToolCallTracingHook", toolParams)
				require.NoError(t, err)
				assert.NotNil(t, toolHookResult)

				toolHookInfo, ok := toolHookResult.(engine.ObjectValue)
				require.True(t, ok)
				toolHookMap := toolHookInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "tool_call", toolHookMap["type"])

				// Create event tracing hook
				eventParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				eventHookResult, err := bridge.ExecuteMethod(ctx, "createEventTracingHook", eventParams)
				require.NoError(t, err)
				assert.NotNil(t, eventHookResult)

				eventHookInfo, ok := eventHookResult.(engine.ObjectValue)
				require.True(t, ok)
				eventHookMap := eventHookInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "event", eventHookMap["type"])
			},
		},
		{
			name: "Create composite tracing hook",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create composite hook
				compositeParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				compositeResult, err := bridge.ExecuteMethod(ctx, "createCompositeTracingHook", compositeParams)
				require.NoError(t, err)
				assert.NotNil(t, compositeResult)

				compositeInfo, ok := compositeResult.(engine.ObjectValue)
				require.True(t, ok)
				compositeMap := compositeInfo.ToGo().(map[string]interface{})
				assert.Equal(t, "composite", compositeMap["type"])
				assert.NotEmpty(t, compositeMap["id"])
			},
		},
		{
			name: "Span from context",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test with no span in context
				contextParams := []engine.ScriptValue{}
				result, err := bridge.ExecuteMethod(ctx, "spanFromContext", contextParams)
				require.NoError(t, err)
				assert.True(t, result.IsNil())

				// Create tracer and span
				tracerParams := []engine.ScriptValue{
					engine.NewStringValue("test-tracer"),
				}
				tracerResult, err := bridge.ExecuteMethod(ctx, "createTracer", tracerParams)
				require.NoError(t, err)
				tracerInfo := tracerResult.(engine.ObjectValue)
				tracerID := tracerInfo.ToGo().(map[string]interface{})["id"].(string)

				spanParams := []engine.ScriptValue{
					engine.NewStringValue(tracerID),
					engine.NewStringValue("test-operation"),
				}
				spanResult, err := bridge.ExecuteMethod(ctx, "startSpan", spanParams)
				require.NoError(t, err)
				spanInfo := spanResult.(engine.ObjectValue)
				spanID := spanInfo.ToGo().(map[string]interface{})["id"].(string)

				// The span should be available in context through go-llms tracing
				// Note: This test depends on go-llms tracing implementation

				// End span
				endParams := []engine.ScriptValue{
					engine.NewStringValue(spanID),
				}
				_, err = bridge.ExecuteMethod(ctx, "endSpan", endParams)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewTracingBridge()
			tt.test(t, bridge)
		})
	}
}

// Test integration with go-llms tracing hooks
func TestTracingBridgeIntegration(t *testing.T) {
	bridge := NewTracingBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create a mock tracer for testing
	mockTracer := &MockTracer{spans: make(map[string]*MockSpan)}

	// Test with go-llms tracing hooks
	tracingHook := core.NewTracingHook("test-tracer", mockTracer)
	assert.NotNil(t, tracingHook)

	// Test with no span context initially
	span := core.SpanFromContext(ctx)
	assert.Nil(t, span)

	// Test creating a span through the tracer
	newCtx, span := mockTracer.Start(ctx, "test-operation")
	assert.NotNil(t, span)

	// Test span from context
	retrievedSpan := core.SpanFromContext(newCtx)
	assert.Equal(t, span, retrievedSpan)

	// Verify span was created
	assert.Len(t, mockTracer.spans, 1)
}

// Mock implementations for testing

type MockTracer struct {
	spans map[string]*MockSpan
}

func (t *MockTracer) Start(ctx context.Context, name string, opts ...core.SpanOption) (context.Context, core.Span) {
	span := &MockSpan{
		name:       name,
		attributes: make(map[string]interface{}),
		started:    time.Now(),
		recording:  true,
	}

	spanID := fmt.Sprintf("span-%d", len(t.spans))
	t.spans[spanID] = span

	// Add span to context using go-llms helper
	ctx = core.ContextWithSpan(ctx, span)

	return ctx, span
}

type MockSpan struct {
	name       string
	attributes map[string]interface{}
	errors     []error
	status     core.StatusCode
	statusDesc string
	started    time.Time
	ended      time.Time
	recording  bool
}

func (s *MockSpan) End() {
	s.ended = time.Now()
	s.recording = false
}

func (s *MockSpan) SetAttributes(attributes ...core.Attribute) {
	for _, attr := range attributes {
		s.attributes[attr.Key] = attr.Value
	}
}

func (s *MockSpan) RecordError(err error) {
	s.errors = append(s.errors, err)
}

func (s *MockSpan) SetStatus(code core.StatusCode, description string) {
	s.status = code
	s.statusDesc = description
}

func (s *MockSpan) IsRecording() bool {
	return s.recording
}

// Test error scenarios
func TestTracingBridgeErrors(t *testing.T) {
	bridge := NewTracingBridge()
	ctx := context.Background()

	// Test methods without initialization
	params := []engine.ScriptValue{
		engine.NewStringValue("test"),
	}
	_, err := bridge.ExecuteMethod(ctx, "createTracer", params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.ExecuteMethod(ctx, "createTracer", []engine.ScriptValue{})
	assert.Error(t, err)

	spanParams := []engine.ScriptValue{
		engine.NewStringValue("invalid-tracer-id"),
		engine.NewStringValue("test"),
	}
	_, err = bridge.ExecuteMethod(ctx, "startSpan", spanParams)
	assert.Error(t, err)

	endParams := []engine.ScriptValue{
		engine.NewStringValue("invalid-span-id"),
	}
	_, err = bridge.ExecuteMethod(ctx, "endSpan", endParams)
	assert.Error(t, err)
}

// Test concurrent operations
func TestTracingBridgeConcurrency(t *testing.T) {
	bridge := NewTracingBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create tracer
	tracerParams := []engine.ScriptValue{
		engine.NewStringValue("concurrent-tracer"),
	}
	tracerResult, err := bridge.ExecuteMethod(ctx, "createTracer", tracerParams)
	require.NoError(t, err)
	tracerInfo := tracerResult.(engine.ObjectValue)
	tracerID := tracerInfo.ToGo().(map[string]interface{})["id"].(string)

	// Create multiple spans concurrently
	numSpans := 10
	done := make(chan bool, numSpans)

	for i := 0; i < numSpans; i++ {
		go func(spanNum int) {
			spanName := fmt.Sprintf("concurrent-span-%d", spanNum)

			// Start span
			spanParams := []engine.ScriptValue{
				engine.NewStringValue(tracerID),
				engine.NewStringValue(spanName),
			}
			spanResult, err := bridge.ExecuteMethod(ctx, "startSpan", spanParams)
			assert.NoError(t, err)
			spanInfo := spanResult.(engine.ObjectValue)
			spanID := spanInfo.ToGo().(map[string]interface{})["id"].(string)

			// Set attributes
			attributes := map[string]engine.ScriptValue{
				"span.number": engine.NewNumberValue(float64(spanNum)),
				"operation":   engine.NewStringValue("concurrent_test"),
			}
			attrParams := []engine.ScriptValue{
				engine.NewStringValue(spanID),
				engine.NewObjectValue(attributes),
			}
			_, err = bridge.ExecuteMethod(ctx, "setSpanAttributes", attrParams)
			assert.NoError(t, err)

			// End span
			endParams := []engine.ScriptValue{
				engine.NewStringValue(spanID),
			}
			_, err = bridge.ExecuteMethod(ctx, "endSpan", endParams)
			assert.NoError(t, err)

			done <- true
		}(i)
	}

	// Wait for all spans to complete
	for i := 0; i < numSpans; i++ {
		<-done
	}
}
