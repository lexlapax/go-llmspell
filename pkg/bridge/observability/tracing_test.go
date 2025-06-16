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
				params := []interface{}{"test-tracer"}
				result, err := bridge.createTracer(ctx, params)
				require.NoError(t, err)
				assert.NotNil(t, result)

				tracerInfo, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test-tracer", tracerInfo["name"])
				assert.NotEmpty(t, tracerInfo["id"])
			},
		},
		{
			name: "Start and end span",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create tracer first
				tracerResult, err := bridge.createTracer(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				tracerInfo := tracerResult.(map[string]interface{})
				tracerID := tracerInfo["id"].(string)

				// Start span
				spanResult, err := bridge.startSpan(ctx, []interface{}{tracerID, "test-operation"})
				require.NoError(t, err)
				assert.NotNil(t, spanResult)

				spanInfo, ok := spanResult.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test-operation", spanInfo["name"])
				spanID := spanInfo["id"].(string)

				// End span
				err = bridge.endSpan(ctx, []interface{}{spanID})
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
				tracerResult, err := bridge.createTracer(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				tracerID := tracerResult.(map[string]interface{})["id"].(string)

				spanResult, err := bridge.startSpan(ctx, []interface{}{tracerID, "test-operation"})
				require.NoError(t, err)
				spanID := spanResult.(map[string]interface{})["id"].(string)

				// Set attributes
				attributes := map[string]interface{}{
					"operation.type": "test",
					"user.id":        "123",
					"request.size":   1024,
				}
				err = bridge.setSpanAttributes(ctx, []interface{}{spanID, attributes})
				require.NoError(t, err)

				// End span
				err = bridge.endSpan(ctx, []interface{}{spanID})
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
				tracerResult, err := bridge.createTracer(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				tracerID := tracerResult.(map[string]interface{})["id"].(string)

				spanResult, err := bridge.startSpan(ctx, []interface{}{tracerID, "test-operation"})
				require.NoError(t, err)
				spanID := spanResult.(map[string]interface{})["id"].(string)

				// Record error
				testError := "test error message"
				err = bridge.recordSpanError(ctx, []interface{}{spanID, testError})
				require.NoError(t, err)

				// Set error status
				err = bridge.setSpanStatus(ctx, []interface{}{spanID, "error", "Operation failed"})
				require.NoError(t, err)

				// End span
				err = bridge.endSpan(ctx, []interface{}{spanID})
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
				agentHookResult, err := bridge.createAgentTracingHook(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				assert.NotNil(t, agentHookResult)

				agentHookInfo, ok := agentHookResult.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "agent", agentHookInfo["type"])

				// Create tool call tracing hook
				toolHookResult, err := bridge.createToolCallTracingHook(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				assert.NotNil(t, toolHookResult)

				toolHookInfo, ok := toolHookResult.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "tool_call", toolHookInfo["type"])

				// Create event tracing hook
				eventHookResult, err := bridge.createEventTracingHook(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				assert.NotNil(t, eventHookResult)

				eventHookInfo, ok := eventHookResult.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "event", eventHookInfo["type"])
			},
		},
		{
			name: "Create composite tracing hook",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Create composite hook
				compositeResult, err := bridge.createCompositeTracingHook(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				assert.NotNil(t, compositeResult)

				compositeInfo, ok := compositeResult.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "composite", compositeInfo["type"])
				assert.NotEmpty(t, compositeInfo["id"])
			},
		},
		{
			name: "Span from context",
			test: func(t *testing.T, bridge *TracingBridge) {
				ctx := context.Background()
				err := bridge.Initialize(ctx)
				require.NoError(t, err)

				// Test with no span in context
				result, err := bridge.spanFromContext(ctx, []interface{}{})
				require.NoError(t, err)
				assert.Nil(t, result)

				// Create tracer and span
				tracerResult, err := bridge.createTracer(ctx, []interface{}{"test-tracer"})
				require.NoError(t, err)
				tracerID := tracerResult.(map[string]interface{})["id"].(string)

				spanResult, err := bridge.startSpan(ctx, []interface{}{tracerID, "test-operation"})
				require.NoError(t, err)
				spanID := spanResult.(map[string]interface{})["id"].(string)

				// The span should be available in context through go-llms tracing
				// Note: This test depends on go-llms tracing implementation

				// End span
				err = bridge.endSpan(ctx, []interface{}{spanID})
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
	_, err := bridge.createTracer(ctx, []interface{}{"test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test invalid parameters
	_, err = bridge.createTracer(ctx, []interface{}{})
	assert.Error(t, err)

	_, err = bridge.startSpan(ctx, []interface{}{"invalid-tracer-id", "test"})
	assert.Error(t, err)

	err = bridge.endSpan(ctx, []interface{}{"invalid-span-id"})
	assert.Error(t, err)
}

// Test concurrent operations
func TestTracingBridgeConcurrency(t *testing.T) {
	bridge := NewTracingBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create tracer
	tracerResult, err := bridge.createTracer(ctx, []interface{}{"concurrent-tracer"})
	require.NoError(t, err)
	tracerID := tracerResult.(map[string]interface{})["id"].(string)

	// Create multiple spans concurrently
	numSpans := 10
	done := make(chan bool, numSpans)

	for i := 0; i < numSpans; i++ {
		go func(spanNum int) {
			spanName := fmt.Sprintf("concurrent-span-%d", spanNum)

			// Start span
			spanResult, err := bridge.startSpan(ctx, []interface{}{tracerID, spanName})
			assert.NoError(t, err)
			spanID := spanResult.(map[string]interface{})["id"].(string)

			// Set attributes
			attributes := map[string]interface{}{
				"span.number": spanNum,
				"operation":   "concurrent_test",
			}
			err = bridge.setSpanAttributes(ctx, []interface{}{spanID, attributes})
			assert.NoError(t, err)

			// End span
			err = bridge.endSpan(ctx, []interface{}{spanID})
			assert.NoError(t, err)

			done <- true
		}(i)
	}

	// Wait for all spans to complete
	for i := 0; i < numSpans; i++ {
		<-done
	}
}
