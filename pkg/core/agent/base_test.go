// ABOUTME: Tests for the base agent implementation that provides common functionality for all agents
// ABOUTME: Ensures state management, event emission, error handling, and metrics work correctly

package agent

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBaseAgent_Creation tests the creation and initialization of base agents
func TestBaseAgent_Creation(t *testing.T) {
	tests := []struct {
		name        string
		config      BaseAgentConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid configuration",
			config: BaseAgentConfig{
				ID:          "test-agent",
				Name:        "Test Agent",
				Description: "A test agent",
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			config: BaseAgentConfig{
				Name:        "Test Agent",
				Description: "A test agent",
				Version:     "1.0.0",
			},
			wantErr:     true,
			errContains: "agent ID is required",
		},
		{
			name: "missing name",
			config: BaseAgentConfig{
				ID:          "test-agent",
				Description: "A test agent",
				Version:     "1.0.0",
			},
			wantErr:     true,
			errContains: "agent name is required",
		},
		{
			name: "with capabilities",
			config: BaseAgentConfig{
				ID:          "test-agent",
				Name:        "Test Agent",
				Description: "A test agent",
				Version:     "1.0.0",
				Capabilities: map[string]interface{}{
					"streaming": true,
					"tools":     []string{"file_read", "file_write"},
				},
			},
			wantErr: false,
		},
		{
			name: "with metadata",
			config: BaseAgentConfig{
				ID:          "test-agent",
				Name:        "Test Agent",
				Description: "A test agent",
				Version:     "1.0.0",
				Metadata: map[string]interface{}{
					"author": "test",
					"tags":   []string{"test", "example"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewBaseAgent(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, agent)

			// Verify basic properties
			assert.Equal(t, tt.config.ID, agent.ID())
			assert.Equal(t, tt.config.Name, agent.Name())
			assert.Equal(t, tt.config.Description, agent.Description())
			assert.Equal(t, tt.config.Version, agent.Version())

			// Check initial status
			assert.Equal(t, StatusCreated, agent.Status())
		})
	}
}

// TestBaseAgent_StateManagement tests the state management functionality
func TestBaseAgent_StateManagement(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "state-test",
		Name:    "State Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("initial state", func(t *testing.T) {
		state := agent.GetState(ctx)
		assert.NotNil(t, state)
		assert.Empty(t, state)
	})

	t.Run("set and get state", func(t *testing.T) {
		// Set state
		newState := map[string]interface{}{
			"counter": 42,
			"message": "hello",
			"data":    map[string]string{"key": "value"},
		}
		err := agent.SetState(ctx, newState)
		assert.NoError(t, err)

		// Get state
		state := agent.GetState(ctx)
		assert.Equal(t, newState, state)
	})

	t.Run("update state", func(t *testing.T) {
		// Update existing state
		updates := map[string]interface{}{
			"counter": 43,
			"new_key": "new_value",
		}
		err := agent.UpdateState(ctx, updates)
		assert.NoError(t, err)

		// Verify updates
		state := agent.GetState(ctx)
		assert.Equal(t, 43, state["counter"])
		assert.Equal(t, "hello", state["message"]) // unchanged
		assert.Equal(t, "new_value", state["new_key"])
	})

	t.Run("clear state", func(t *testing.T) {
		err := agent.ClearState(ctx)
		assert.NoError(t, err)

		state := agent.GetState(ctx)
		assert.Empty(t, state)
	})

	t.Run("concurrent state access", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		// Concurrent writes
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				err := agent.UpdateState(ctx, map[string]interface{}{
					"key_" + string(rune('0'+n)): n,
				})
				assert.NoError(t, err)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				state := agent.GetState(ctx)
				assert.NotNil(t, state)
			}()
		}

		wg.Wait()

		// Verify all writes succeeded
		state := agent.GetState(ctx)
		for i := 0; i < numGoroutines; i++ {
			key := "key_" + string(rune('0'+i))
			assert.Contains(t, state, key)
			assert.Equal(t, i, state[key])
		}
	})
}

// TestBaseAgent_EventEmission tests the event emission functionality
func TestBaseAgent_EventEmission(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "event-test",
		Name:    "Event Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("emit and receive events", func(t *testing.T) {
		// Subscribe to events
		eventChan := agent.Subscribe(ctx, "test.event")
		require.NotNil(t, eventChan)

		// Emit event
		event := Event{
			Type:      "test.event",
			Source:    agent.ID(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"message": "test event",
			},
		}
		err := agent.EmitEvent(ctx, event)
		assert.NoError(t, err)

		// Receive event
		select {
		case received := <-eventChan:
			assert.Equal(t, event.Type, received.Type)
			assert.Equal(t, event.Source, received.Source)
			assert.Equal(t, event.Data, received.Data)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("event not received")
		}
	})

	t.Run("filtered subscription", func(t *testing.T) {
		// Subscribe to specific event type
		eventChan := agent.Subscribe(ctx, "specific.event")

		// Emit different event
		err := agent.EmitEvent(ctx, Event{
			Type:   "other.event",
			Source: agent.ID(),
		})
		assert.NoError(t, err)

		// Should not receive event
		select {
		case <-eventChan:
			t.Fatal("received unexpected event")
		case <-time.After(50 * time.Millisecond):
			// Expected timeout
		}

		// Emit matching event
		err = agent.EmitEvent(ctx, Event{
			Type:   "specific.event",
			Source: agent.ID(),
		})
		assert.NoError(t, err)

		// Should receive event
		select {
		case received := <-eventChan:
			assert.Equal(t, "specific.event", received.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("event not received")
		}
	})

	t.Run("wildcard subscription", func(t *testing.T) {
		// Subscribe to all events
		eventChan := agent.Subscribe(ctx, "*")

		// Emit various events
		events := []string{"event.one", "event.two", "other.event"}
		for _, eventType := range events {
			err := agent.EmitEvent(ctx, Event{
				Type:   eventType,
				Source: agent.ID(),
			})
			assert.NoError(t, err)
		}

		// Should receive all events
		for i := 0; i < len(events); i++ {
			select {
			case received := <-eventChan:
				assert.Contains(t, events, received.Type)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("event not received")
			}
		}
	})

	t.Run("unsubscribe", func(t *testing.T) {
		eventChan := agent.Subscribe(ctx, "test.event")

		// Unsubscribe
		err := agent.Unsubscribe(ctx, eventChan)
		assert.NoError(t, err)

		// Emit event
		err = agent.EmitEvent(ctx, Event{
			Type:   "test.event",
			Source: agent.ID(),
		})
		assert.NoError(t, err)

		// Should not receive event
		select {
		case <-eventChan:
			t.Fatal("received event after unsubscribe")
		case <-time.After(50 * time.Millisecond):
			// Expected timeout
		}
	})
}

// TestBaseAgent_ErrorHandling tests error handling and recovery
func TestBaseAgent_ErrorHandling(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "error-test",
		Name:    "Error Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("handle error", func(t *testing.T) {
		testErr := errors.New("test error")

		// Handle error
		agent.HandleError(ctx, testErr)

		// Check error was recorded
		assert.Equal(t, StatusError, agent.Status())
		assert.Equal(t, testErr, agent.LastError())
	})

	t.Run("recover from error", func(t *testing.T) {
		// Set error state
		agent.HandleError(ctx, errors.New("recoverable error"))
		assert.Equal(t, StatusError, agent.Status())

		// Recover
		err := agent.Recover(ctx)
		assert.NoError(t, err)
		assert.Equal(t, StatusReady, agent.Status())
		assert.Nil(t, agent.LastError())
	})

	t.Run("error with retry", func(t *testing.T) {
		retryCount := 0
		operation := func() error {
			retryCount++
			if retryCount < 3 {
				return errors.New("temporary error")
			}
			return nil
		}

		// Execute with retry
		err := agent.ExecuteWithRetry(ctx, operation, 3, 10*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 3, retryCount)
	})

	t.Run("error with max retries exceeded", func(t *testing.T) {
		operation := func() error {
			return errors.New("permanent error")
		}

		// Execute with retry
		err := agent.ExecuteWithRetry(ctx, operation, 3, 10*time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permanent error")
		assert.Equal(t, StatusError, agent.Status())
	})
}

// TestBaseAgent_Metrics tests metrics collection functionality
func TestBaseAgent_Metrics(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "metrics-test",
		Name:    "Metrics Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("record metric", func(t *testing.T) {
		// Record various metrics
		agent.RecordMetric(ctx, "requests", 10)
		agent.RecordMetric(ctx, "errors", 2)
		agent.RecordMetric(ctx, "latency_ms", 150.5)

		// Get metrics
		metrics := agent.GetMetrics(ctx)
		assert.Equal(t, float64(10), metrics["requests"])
		assert.Equal(t, float64(2), metrics["errors"])
		assert.Equal(t, 150.5, metrics["latency_ms"])
	})

	t.Run("increment metric", func(t *testing.T) {
		// Increment counter
		agent.IncrementMetric(ctx, "requests", 5)

		metrics := agent.GetMetrics(ctx)
		assert.Equal(t, float64(15), metrics["requests"]) // 10 + 5
	})

	t.Run("reset metrics", func(t *testing.T) {
		// Reset all metrics
		agent.ResetMetrics(ctx)

		metrics := agent.GetMetrics(ctx)
		assert.Empty(t, metrics)
	})

	t.Run("concurrent metric updates", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				agent.IncrementMetric(ctx, "concurrent", 1)
			}()
		}

		wg.Wait()

		metrics := agent.GetMetrics(ctx)
		assert.Equal(t, float64(numGoroutines), metrics["concurrent"])
	})
}

// TestBaseAgent_Lifecycle tests the agent lifecycle management
func TestBaseAgent_Lifecycle(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "lifecycle-test",
		Name:    "Lifecycle Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("initialization", func(t *testing.T) {
		assert.Equal(t, StatusCreated, agent.Status())

		// Initialize
		err := agent.Init(ctx)
		assert.NoError(t, err)
		assert.Equal(t, StatusReady, agent.Status())

		// Initialize again should be idempotent
		err = agent.Init(ctx)
		assert.NoError(t, err)
		assert.Equal(t, StatusReady, agent.Status())
	})

	t.Run("running", func(t *testing.T) {
		// Define run function
		runFunc := func(ctx context.Context, input interface{}) (interface{}, error) {
			// Record that we're running
			agent.RecordMetric(ctx, "runs", 1)
			return "result", nil
		}
		agent.SetRunFunc(runFunc)

		// Run
		result, err := agent.Run(ctx, "input")
		assert.NoError(t, err)
		assert.Equal(t, "result", result)

		// Check metrics
		metrics := agent.GetMetrics(ctx)
		assert.Equal(t, float64(1), metrics["runs"])
	})

	t.Run("cleanup", func(t *testing.T) {
		// Subscribe to cleanup event
		eventChan := agent.Subscribe(ctx, "agent.cleanup")

		// Cleanup
		err := agent.Cleanup(ctx)
		assert.NoError(t, err)
		assert.Equal(t, StatusStopped, agent.Status())

		// Check cleanup event was emitted
		select {
		case event := <-eventChan:
			assert.Equal(t, "agent.cleanup", event.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("cleanup event not received")
		}

		// Cleanup again should be idempotent
		err = agent.Cleanup(ctx)
		assert.NoError(t, err)
		assert.Equal(t, StatusStopped, agent.Status())
	})
}

// TestBaseAgent_Configuration tests runtime configuration
func TestBaseAgent_Configuration(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "config-test",
		Name:    "Config Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	t.Run("configure agent", func(t *testing.T) {
		// Configure with options
		err := agent.Configure(
			WithTimeout(5*time.Second),
			WithMaxRetries(5),
			WithDebug(true),
		)
		assert.NoError(t, err)

		// Get config
		config := agent.GetConfig()
		assert.Equal(t, 5*time.Second, config["timeout"])
		assert.Equal(t, 5, config["maxRetries"])
		assert.Equal(t, true, config["debug"])
	})

	t.Run("invalid configuration", func(t *testing.T) {
		// Try invalid option
		invalidOption := func(a Agent) error {
			return errors.New("invalid configuration")
		}

		err := agent.Configure(invalidOption)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})
}

// TestBaseAgent_AsyncOperations tests asynchronous operations
func TestBaseAgent_AsyncOperations(t *testing.T) {
	agent, err := NewBaseAgent(BaseAgentConfig{
		ID:      "async-test",
		Name:    "Async Test Agent",
		Version: "1.0.0",
	})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("async run", func(t *testing.T) {
		// Set run function
		runFunc := func(ctx context.Context, input interface{}) (interface{}, error) {
			time.Sleep(50 * time.Millisecond) // Simulate work
			return "async result", nil
		}
		agent.SetRunFunc(runFunc)

		// Run async
		resultChan, errChan := agent.RunAsync(ctx, "input")

		// Wait for result
		select {
		case result := <-resultChan:
			assert.Equal(t, "async result", result)
		case err := <-errChan:
			t.Fatalf("unexpected error: %v", err)
		case <-time.After(200 * time.Millisecond):
			t.Fatal("async operation timed out")
		}
	})

	t.Run("async run with error", func(t *testing.T) {
		// Set run function that errors
		runFunc := func(ctx context.Context, input interface{}) (interface{}, error) {
			return nil, errors.New("async error")
		}
		agent.SetRunFunc(runFunc)

		// Run async
		resultChan, errChan := agent.RunAsync(ctx, "input")

		// Wait for error
		select {
		case <-resultChan:
			t.Fatal("expected error, got result")
		case err := <-errChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "async error")
		case <-time.After(200 * time.Millisecond):
			t.Fatal("async operation timed out")
		}
	})

	t.Run("async run with context cancellation", func(t *testing.T) {
		// Set slow run function
		runFunc := func(ctx context.Context, input interface{}) (interface{}, error) {
			select {
			case <-time.After(1 * time.Second):
				return "should not reach", nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		agent.SetRunFunc(runFunc)

		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())

		// Run async
		resultChan, errChan := agent.RunAsync(ctx, "input")

		// Cancel after short delay
		time.Sleep(50 * time.Millisecond)
		cancel()

		// Wait for cancellation error
		select {
		case <-resultChan:
			t.Fatal("expected error, got result")
		case err := <-errChan:
			assert.Error(t, err)
			assert.True(t, errors.Is(err, context.Canceled))
		case <-time.After(200 * time.Millisecond):
			t.Fatal("async operation timed out")
		}
	})
}
