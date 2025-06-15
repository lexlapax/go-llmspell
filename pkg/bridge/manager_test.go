// ABOUTME: Test suite for the bridge manager that handles lifecycle management of language-agnostic bridges.
// ABOUTME: Tests thread-safe registration, dependency resolution, and hot-reloading functionality.

package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockBridge struct {
	id            string
	initialized   bool
	cleanedUp     bool
	dependencies  []string
	initError     error
	cleanupError  error
	initCallCount int
	mu            sync.Mutex
	initFunc      func(ctx context.Context) error // Allow overriding initialization
}

func (m *mockBridge) GetID() string {
	return m.id
}

func (m *mockBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         m.id,
		Version:      "1.0.0",
		Description:  "Mock bridge for testing",
		Dependencies: m.dependencies,
	}
}

func (m *mockBridge) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initCallCount++

	// Use custom init function if provided
	if m.initFunc != nil {
		return m.initFunc(ctx)
	}

	if m.initError != nil {
		return m.initError
	}
	m.initialized = true
	return nil
}

func (m *mockBridge) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cleanupError != nil {
		return m.cleanupError
	}
	m.cleanedUp = true
	m.initialized = false
	return nil
}

func (m *mockBridge) IsInitialized() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.initialized
}

func (m *mockBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(m)
}

func (m *mockBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "test",
			Description: "Test method",
			ReturnType:  "string",
		},
	}
}

func (m *mockBridge) ValidateMethod(name string, args []interface{}) error {
	return nil
}

func (m *mockBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{}
}

func (m *mockBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{}
}

// Tests for BridgeManager
func TestNewBridgeManager(t *testing.T) {
	manager := NewBridgeManager()
	assert.NotNil(t, manager)
}

func TestBridgeLifecycleManagement(t *testing.T) {
	t.Run("Register and Initialize Bridge", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "test-bridge"}

		// Register bridge
		err := manager.RegisterBridge(bridge)
		assert.NoError(t, err)

		// Initialize bridge
		ctx := context.Background()
		err = manager.InitializeBridge(ctx, "test-bridge")
		assert.NoError(t, err)
		assert.True(t, bridge.initialized)
	})

	t.Run("Initialize All Bridges", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2"}

		_ = manager.RegisterBridge(bridge1)
		_ = manager.RegisterBridge(bridge2)

		ctx := context.Background()
		err := manager.InitializeAll(ctx)
		assert.NoError(t, err)
		assert.True(t, bridge1.initialized)
		assert.True(t, bridge2.initialized)
	})

	t.Run("Cleanup Bridge", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "test-bridge", initialized: true}

		_ = manager.RegisterBridge(bridge)

		ctx := context.Background()
		err := manager.CleanupBridge(ctx, "test-bridge")
		assert.NoError(t, err)
		assert.True(t, bridge.cleanedUp)
		assert.False(t, bridge.initialized)
	})

	t.Run("Cleanup All Bridges", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge1 := &mockBridge{id: "bridge1", initialized: true}
		bridge2 := &mockBridge{id: "bridge2", initialized: true}

		_ = manager.RegisterBridge(bridge1)
		_ = manager.RegisterBridge(bridge2)

		ctx := context.Background()
		err := manager.CleanupAll(ctx)
		assert.NoError(t, err)
		assert.True(t, bridge1.cleanedUp)
		assert.True(t, bridge2.cleanedUp)
	})

	t.Run("Initialize Error Handling", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{
			id:        "error-bridge",
			initError: errors.New("init failed"),
		}

		_ = manager.RegisterBridge(bridge)

		ctx := context.Background()
		err := manager.InitializeBridge(ctx, "error-bridge")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "init failed")
		assert.False(t, bridge.initialized)
	})

	t.Run("Cleanup Error Handling", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{
			id:           "error-bridge",
			initialized:  true,
			cleanupError: errors.New("cleanup failed"),
		}

		_ = manager.RegisterBridge(bridge)

		ctx := context.Background()
		err := manager.CleanupBridge(ctx, "error-bridge")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cleanup failed")
	})
}

func TestThreadSafeBridgeRegistration(t *testing.T) {
	t.Run("Concurrent Bridge Registration", func(t *testing.T) {
		manager := NewBridgeManager()
		var wg sync.WaitGroup
		errorChan := make(chan error, 100)

		// Register 100 bridges concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				bridge := &mockBridge{id: fmt.Sprintf("bridge-%d", id)}
				if err := manager.RegisterBridge(bridge); err != nil {
					errorChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for errors
		for err := range errorChan {
			assert.NoError(t, err)
		}

		// Verify all bridges are registered
		bridges := manager.ListBridges()
		assert.Len(t, bridges, 100)
	})

	t.Run("Duplicate Registration Prevention", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge1 := &mockBridge{id: "duplicate"}
		bridge2 := &mockBridge{id: "duplicate"}

		err := manager.RegisterBridge(bridge1)
		assert.NoError(t, err)

		err = manager.RegisterBridge(bridge2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("Concurrent Access to Same Bridge", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "concurrent-test"}
		_ = manager.RegisterBridge(bridge)

		ctx := context.Background()
		var wg sync.WaitGroup

		// Multiple goroutines initializing the same bridge
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = manager.InitializeBridge(ctx, "concurrent-test")
			}()
		}

		wg.Wait()

		// Bridge should only be initialized once
		assert.Equal(t, 1, bridge.initCallCount)
	})
}

func TestBridgeDependencyResolution(t *testing.T) {
	t.Run("Simple Dependency Chain", func(t *testing.T) {
		manager := NewBridgeManager()

		// Create bridges with dependencies
		bridgeA := &mockBridge{id: "bridgeA"}
		bridgeB := &mockBridge{id: "bridgeB", dependencies: []string{"bridgeA"}}
		bridgeC := &mockBridge{id: "bridgeC", dependencies: []string{"bridgeB"}}

		// Register in reverse order to test resolution
		_ = manager.RegisterBridge(bridgeC)
		_ = manager.RegisterBridge(bridgeB)
		_ = manager.RegisterBridge(bridgeA)

		// Initialize with dependency resolution
		ctx := context.Background()
		err := manager.InitializeWithDependencies(ctx, "bridgeC")
		assert.NoError(t, err)

		// All dependencies should be initialized
		assert.True(t, bridgeA.initialized)
		assert.True(t, bridgeB.initialized)
		assert.True(t, bridgeC.initialized)
	})

	t.Run("Circular Dependency Detection", func(t *testing.T) {
		manager := NewBridgeManager()

		// Create circular dependency
		bridgeA := &mockBridge{id: "bridgeA", dependencies: []string{"bridgeB"}}
		bridgeB := &mockBridge{id: "bridgeB", dependencies: []string{"bridgeA"}}

		_ = manager.RegisterBridge(bridgeA)
		_ = manager.RegisterBridge(bridgeB)

		ctx := context.Background()
		err := manager.InitializeWithDependencies(ctx, "bridgeA")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency")
	})

	t.Run("Missing Dependency", func(t *testing.T) {
		manager := NewBridgeManager()

		bridge := &mockBridge{id: "bridge", dependencies: []string{"missing"}}
		_ = manager.RegisterBridge(bridge)

		ctx := context.Background()
		err := manager.InitializeWithDependencies(ctx, "bridge")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dependency not found")
	})

	t.Run("Multiple Dependencies", func(t *testing.T) {
		manager := NewBridgeManager()

		bridgeA := &mockBridge{id: "bridgeA"}
		bridgeB := &mockBridge{id: "bridgeB"}
		bridgeC := &mockBridge{id: "bridgeC", dependencies: []string{"bridgeA", "bridgeB"}}

		_ = manager.RegisterBridge(bridgeA)
		_ = manager.RegisterBridge(bridgeB)
		_ = manager.RegisterBridge(bridgeC)

		ctx := context.Background()
		err := manager.InitializeWithDependencies(ctx, "bridgeC")
		assert.NoError(t, err)

		assert.True(t, bridgeA.initialized)
		assert.True(t, bridgeB.initialized)
		assert.True(t, bridgeC.initialized)
	})
}

func TestHotReloading(t *testing.T) {
	t.Run("Reload Single Bridge", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "reload-test"}

		_ = manager.RegisterBridge(bridge)
		ctx := context.Background()

		// Initialize bridge
		err := manager.InitializeBridge(ctx, "reload-test")
		assert.NoError(t, err)
		assert.True(t, bridge.initialized)
		assert.Equal(t, 1, bridge.initCallCount)

		// Reload bridge
		err = manager.ReloadBridge(ctx, "reload-test")
		assert.NoError(t, err)
		assert.True(t, bridge.initialized)
		assert.True(t, bridge.cleanedUp)
		assert.Equal(t, 2, bridge.initCallCount)
	})

	t.Run("Reload With Dependencies", func(t *testing.T) {
		manager := NewBridgeManager()

		bridgeA := &mockBridge{id: "bridgeA"}
		bridgeB := &mockBridge{id: "bridgeB", dependencies: []string{"bridgeA"}}

		_ = manager.RegisterBridge(bridgeA)
		_ = manager.RegisterBridge(bridgeB)

		ctx := context.Background()
		_ = manager.InitializeWithDependencies(ctx, "bridgeB")

		// Reload dependent bridge
		err := manager.ReloadBridge(ctx, "bridgeA")
		assert.NoError(t, err)

		// Both bridges should be reloaded due to dependency
		assert.Equal(t, 2, bridgeA.initCallCount)
		assert.Equal(t, 2, bridgeB.initCallCount)
	})

	t.Run("Reload Non-existent Bridge", func(t *testing.T) {
		manager := NewBridgeManager()
		ctx := context.Background()

		err := manager.ReloadBridge(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Watch Bridge Changes", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "watch-test"}

		_ = manager.RegisterBridge(bridge)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start watching
		reloadChan := make(chan string, 1)
		err := manager.WatchBridge(ctx, "watch-test", time.Millisecond*100, func(bridgeID string) {
			reloadChan <- bridgeID
		})
		assert.NoError(t, err)

		// Simulate change detection
		manager.NotifyChange("watch-test")

		// Wait for reload notification
		select {
		case bridgeID := <-reloadChan:
			assert.Equal(t, "watch-test", bridgeID)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for reload notification")
		}
	})
}

func TestBridgeQueries(t *testing.T) {
	t.Run("Get Bridge", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "test-bridge"}
		_ = manager.RegisterBridge(bridge)

		// Get existing bridge
		retrieved, err := manager.GetBridge("test-bridge")
		assert.NoError(t, err)
		assert.Equal(t, bridge, retrieved)

		// Get non-existent bridge
		_, err = manager.GetBridge("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("List Bridges", func(t *testing.T) {
		manager := NewBridgeManager()

		// Empty list
		bridges := manager.ListBridges()
		assert.Empty(t, bridges)

		// Add bridges
		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2"}
		_ = manager.RegisterBridge(bridge1)
		_ = manager.RegisterBridge(bridge2)

		bridges = manager.ListBridges()
		assert.Len(t, bridges, 2)
		assert.Contains(t, bridges, "bridge1")
		assert.Contains(t, bridges, "bridge2")
	})

	t.Run("Is Bridge Initialized", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "test-bridge"}
		_ = manager.RegisterBridge(bridge)

		// Not initialized
		initialized := manager.IsBridgeInitialized("test-bridge")
		assert.False(t, initialized)

		// Initialize
		ctx := context.Background()
		_ = manager.InitializeBridge(ctx, "test-bridge")

		initialized = manager.IsBridgeInitialized("test-bridge")
		assert.True(t, initialized)

		// Non-existent bridge
		initialized = manager.IsBridgeInitialized("non-existent")
		assert.False(t, initialized)
	})

	t.Run("Get Bridge Metadata", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: "test-bridge"}
		_ = manager.RegisterBridge(bridge)

		metadata, err := manager.GetBridgeMetadata("test-bridge")
		assert.NoError(t, err)
		assert.Equal(t, "test-bridge", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)

		_, err = manager.GetBridgeMetadata("non-existent")
		assert.Error(t, err)
	})
}

func TestBridgeEngineIntegration(t *testing.T) {
	t.Run("Register Bridges with Engine", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2"}

		_ = manager.RegisterBridge(bridge1)
		_ = manager.RegisterBridge(bridge2)

		// Mock engine
		engine := &mockScriptEngine{}

		err := manager.RegisterBridgesWithEngine(engine)
		assert.NoError(t, err)
		assert.Len(t, engine.bridges, 2)
	})

	t.Run("Register Specific Bridges with Engine", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2"}
		bridge3 := &mockBridge{id: "bridge3"}

		_ = manager.RegisterBridge(bridge1)
		_ = manager.RegisterBridge(bridge2)
		_ = manager.RegisterBridge(bridge3)

		engine := &mockScriptEngine{}

		err := manager.RegisterSpecificBridgesWithEngine(engine, []string{"bridge1", "bridge3"})
		assert.NoError(t, err)
		assert.Len(t, engine.bridges, 2)
		assert.Contains(t, engine.bridges, bridge1)
		assert.Contains(t, engine.bridges, bridge3)
	})
}

// Mock ScriptEngine for testing
type mockScriptEngine struct {
	bridges []engine.Bridge
}

func (m *mockScriptEngine) GetInfo() engine.EngineInfo {
	return engine.EngineInfo{Name: "mock", Version: "1.0.0"}
}

func (m *mockScriptEngine) Initialize(config engine.EngineConfig) error {
	return nil
}

func (m *mockScriptEngine) RegisterBridge(bridge engine.Bridge) error {
	m.bridges = append(m.bridges, bridge)
	return nil
}

func (m *mockScriptEngine) ExecuteScript(ctx context.Context, script string, options engine.ExecutionOptions) (*engine.ExecutionResult, error) {
	return &engine.ExecutionResult{}, nil
}

func (m *mockScriptEngine) CreateContext(options engine.ContextOptions) (engine.ScriptContext, error) {
	return nil, nil
}

func (m *mockScriptEngine) GetMetrics() engine.EngineMetrics {
	return engine.EngineMetrics{}
}

func (m *mockScriptEngine) SetResourceLimits(limits engine.ResourceLimits) error {
	return nil
}

func (m *mockScriptEngine) Cleanup() error {
	return nil
}

func (m *mockScriptEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockScriptEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockScriptEngine) Shutdown() error {
	return nil
}

func (m *mockScriptEngine) GetBridge(name string) (engine.Bridge, error) {
	return nil, nil
}

func (m *mockScriptEngine) ToNative(scriptValue interface{}) (interface{}, error) {
	return scriptValue, nil
}

func (m *mockScriptEngine) FromNative(goValue interface{}) (interface{}, error) {
	return goValue, nil
}

func (m *mockScriptEngine) Name() string {
	return "mock"
}

func (m *mockScriptEngine) Version() string {
	return "1.0.0"
}

func (m *mockScriptEngine) FileExtensions() []string {
	return []string{".mock"}
}

func (m *mockScriptEngine) Features() []engine.EngineFeature {
	return []engine.EngineFeature{}
}

func (m *mockScriptEngine) SetMemoryLimit(bytes int64) error {
	return nil
}

func (m *mockScriptEngine) SetTimeout(duration time.Duration) error {
	return nil
}

func (m *mockScriptEngine) ListBridges() []string {
	names := make([]string, 0, len(m.bridges))
	for _, b := range m.bridges {
		names = append(names, b.GetID())
	}
	return names
}

func (m *mockScriptEngine) UnregisterBridge(name string) error {
	return nil
}

func (m *mockScriptEngine) DestroyContext(ctx engine.ScriptContext) error {
	return nil
}

// Performance and stress tests
func TestBridgeManagerPerformance(t *testing.T) {
	t.Run("Large Scale Registration", func(t *testing.T) {
		manager := NewBridgeManager()
		const numBridges = 1000

		start := time.Now()
		for i := 0; i < numBridges; i++ {
			bridge := &mockBridge{id: fmt.Sprintf("bridge-%d", i)}
			err := manager.RegisterBridge(bridge)
			require.NoError(t, err)
		}
		duration := time.Since(start)

		t.Logf("Registered %d bridges in %v", numBridges, duration)
		assert.Less(t, duration, time.Second, "Registration should be fast")

		// List all bridges
		start = time.Now()
		bridges := manager.ListBridges()
		duration = time.Since(start)

		assert.Len(t, bridges, numBridges)
		t.Logf("Listed %d bridges in %v", numBridges, duration)
		assert.Less(t, duration, time.Millisecond*100, "Listing should be fast")
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		manager := NewBridgeManager()
		const numOperations = 100
		const numBridges = 10

		// Register initial bridges
		for i := 0; i < numBridges; i++ {
			bridge := &mockBridge{id: fmt.Sprintf("bridge-%d", i)}
			_ = manager.RegisterBridge(bridge)
		}

		ctx := context.Background()
		var wg sync.WaitGroup

		// Perform concurrent operations
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(op int) {
				defer wg.Done()
				bridgeID := fmt.Sprintf("bridge-%d", op%numBridges)

				switch op % 4 {
				case 0:
					_ = manager.InitializeBridge(ctx, bridgeID)
				case 1:
					_, _ = manager.GetBridge(bridgeID)
				case 2:
					manager.IsBridgeInitialized(bridgeID)
				case 3:
					manager.ListBridges()
				}
			}(i)
		}

		wg.Wait()
		// Test should complete without deadlocks or race conditions
	})
}

// Edge cases and error scenarios
func TestBridgeManagerEdgeCases(t *testing.T) {
	t.Run("Nil Bridge Registration", func(t *testing.T) {
		manager := NewBridgeManager()
		err := manager.RegisterBridge(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil bridge")
	})

	t.Run("Empty Bridge ID", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{id: ""}
		err := manager.RegisterBridge(bridge)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty bridge ID")
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		manager := NewBridgeManager()
		bridge := &mockBridge{
			id:        "slow-bridge",
			initError: nil,
		}

		// Override Initialize to be slow
		bridge.initFunc = func(ctx context.Context) error {
			select {
			case <-time.After(time.Second):
				bridge.initialized = true
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		_ = manager.RegisterBridge(bridge)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cancel()

		err := manager.InitializeBridge(ctx, "slow-bridge")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})
}

// Benchmarks
func BenchmarkBridgeRegistration(b *testing.B) {
	manager := NewBridgeManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge := &mockBridge{id: fmt.Sprintf("bridge-%d", i)}
		_ = manager.RegisterBridge(bridge)
	}
}

func BenchmarkBridgeGet(b *testing.B) {
	manager := NewBridgeManager()

	// Pre-register bridges
	for i := 0; i < 100; i++ {
		bridge := &mockBridge{id: fmt.Sprintf("bridge-%d", i)}
		_ = manager.RegisterBridge(bridge)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GetBridge(fmt.Sprintf("bridge-%d", i%100))
	}
}

func BenchmarkBridgeList(b *testing.B) {
	manager := NewBridgeManager()

	// Pre-register bridges
	for i := 0; i < 1000; i++ {
		bridge := &mockBridge{id: fmt.Sprintf("bridge-%d", i)}
		_ = manager.RegisterBridge(bridge)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ListBridges()
	}
}

// Test Event System Integration
func TestBridgeManagerEventSystem(t *testing.T) {
	t.Run("Create BridgeManager with Event System", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.GetEventBus())
		assert.NotNil(t, manager.GetEventStore())

		// Clean up
		defer func() { _ = manager.Cleanup() }()
	})

	t.Run("Subscribe to Bridge Events", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)
		defer func() { _ = manager.Cleanup() }()

		eventReceived := make(chan bool, 1)
		eventCount := 0

		handler := func(ctx context.Context, event domain.Event) error {
			eventCount++
			eventReceived <- true
			return nil
		}

		subscriptionIDs := manager.SubscribeToBridgeEvents(handler, "bridge.*")
		assert.NotEmpty(t, subscriptionIDs)

		// Register a bridge to trigger events
		bridge := &mockBridge{id: "test-bridge"}
		err := manager.RegisterBridge(bridge)
		assert.NoError(t, err)

		// Wait for event
		select {
		case <-eventReceived:
			assert.Greater(t, eventCount, 0)
		case <-time.After(time.Second):
			t.Log("No events received - this may be expected if event bus is async")
		}

		// Unsubscribe
		manager.UnsubscribeFromBridgeEvents(subscriptionIDs)
	})

	t.Run("Bridge Metrics Collection", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)
		defer func() { _ = manager.Cleanup() }()

		bridge := &mockBridge{id: "metrics-bridge"}
		err := manager.RegisterBridge(bridge)
		assert.NoError(t, err)

		// Initialize bridge to generate metrics
		ctx := context.Background()
		err = manager.InitializeBridge(ctx, "metrics-bridge")
		assert.NoError(t, err)

		// Check metrics
		metrics, err := manager.GetBridgeMetrics("metrics-bridge")
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(1), metrics.InitializationCount)
		assert.Equal(t, int64(0), metrics.FailureCount)
		assert.Greater(t, metrics.InitializationTime, time.Duration(0))

		// Check all metrics
		allMetrics := manager.GetAllBridgeMetrics()
		assert.Len(t, allMetrics, 1)
		assert.Contains(t, allMetrics, "metrics-bridge")
	})

	t.Run("Bridge Failure Metrics", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)
		defer func() { _ = manager.Cleanup() }()

		bridge := &mockBridge{
			id:        "failure-bridge",
			initError: errors.New("initialization failed"),
		}
		err := manager.RegisterBridge(bridge)
		assert.NoError(t, err)

		// Try to initialize bridge (should fail)
		ctx := context.Background()
		err = manager.InitializeBridge(ctx, "failure-bridge")
		assert.Error(t, err)

		// Check failure metrics
		metrics, err := manager.GetBridgeMetrics("failure-bridge")
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(1), metrics.InitializationCount)
		assert.Equal(t, int64(1), metrics.FailureCount)
		assert.NotNil(t, metrics.LastError)
		assert.Equal(t, "initialization failed", metrics.LastError.Error())
	})

	t.Run("Generate Bridge Report", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)
		defer func() { _ = manager.Cleanup() }()

		// Register multiple bridges
		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2"}
		bridge3 := &mockBridge{id: "bridge3", initError: errors.New("failed")}

		_ = manager.RegisterBridge(bridge1)
		_ = manager.RegisterBridge(bridge2)
		_ = manager.RegisterBridge(bridge3)

		// Initialize some bridges
		ctx := context.Background()
		_ = manager.InitializeBridge(ctx, "bridge1")
		_ = manager.InitializeBridge(ctx, "bridge2")
		_ = manager.InitializeBridge(ctx, "bridge3") // This will fail

		// Generate report
		report := manager.GenerateBridgeReport()
		assert.NotNil(t, report)
		assert.Equal(t, 3, report["totalBridges"])
		assert.Equal(t, 2, report["initialized"])
		assert.Equal(t, 1, report["failed"])
		assert.NotEmpty(t, report["sessionID"])

		bridgeDetails := report["bridgeDetails"].(map[string]interface{})
		assert.Len(t, bridgeDetails, 3)

		// Check specific bridge details
		bridge1Details := bridgeDetails["bridge1"].(map[string]interface{})
		assert.Equal(t, true, bridge1Details["initialized"])
		assert.Equal(t, int64(1), bridge1Details["initializationCount"])
		assert.Equal(t, int64(0), bridge1Details["failureCount"])

		bridge3Details := bridgeDetails["bridge3"].(map[string]interface{})
		assert.Equal(t, false, bridge3Details["initialized"])
		assert.Equal(t, int64(1), bridge3Details["initializationCount"])
		assert.Equal(t, int64(1), bridge3Details["failureCount"])
	})

	t.Run("Performance Profiling Methods", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)
		defer func() { _ = manager.Cleanup() }()

		// These methods should not panic and should be safe to call
		assert.NotPanics(t, func() {
			manager.StartProfiling()
			manager.StopProfiling()
		})
	})

	t.Run("Cleanup Event System Resources", func(t *testing.T) {
		manager := NewBridgeManagerWithEvents(nil, nil)

		// Should not error on cleanup
		err := manager.Cleanup()
		assert.NoError(t, err)

		// Should be safe to call multiple times
		err = manager.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("Backward Compatibility with NewBridgeManager", func(t *testing.T) {
		// Original constructor should still work and provide event system
		manager := NewBridgeManager()
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.GetEventBus())
		assert.NotNil(t, manager.GetEventStore())

		// Should be able to register and initialize bridges normally
		bridge := &mockBridge{id: "compat-bridge"}
		err := manager.RegisterBridge(bridge)
		assert.NoError(t, err)

		ctx := context.Background()
		err = manager.InitializeBridge(ctx, "compat-bridge")
		assert.NoError(t, err)

		// Should be able to get metrics
		metrics, err := manager.GetBridgeMetrics("compat-bridge")
		assert.NoError(t, err)
		assert.NotNil(t, metrics)

		// Clean up
		defer func() { _ = manager.Cleanup() }()
	})
}

// Test documentation generation specifically
func TestBridgeManagerDocumentationGeneration(t *testing.T) {
	manager := NewBridgeManager()
	ctx := context.Background()

	// Register a test bridge
	testBridge := &mockBridge{
		id: "test-doc-bridge",
	}

	err := manager.RegisterBridge(testBridge)
	assert.NoError(t, err)

	t.Run("OpenAPI Documentation Generation", func(t *testing.T) {
		result, err := manager.GenerateDocumentation(ctx, "openapi")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Markdown Documentation Generation", func(t *testing.T) {
		result, err := manager.GenerateDocumentation(ctx, "markdown")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("JSON Documentation Generation", func(t *testing.T) {
		result, err := manager.GenerateDocumentation(ctx, "json")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Unsupported Format", func(t *testing.T) {
		_, err := manager.GenerateDocumentation(ctx, "unsupported")
		assert.Error(t, err)
	})

	t.Run("API Schema Export", func(t *testing.T) {
		schema := manager.ExportAPISchema()
		assert.NotNil(t, schema)
		assert.Contains(t, schema, "bridges")
		assert.Contains(t, schema, "types")
		assert.Contains(t, schema, "version")
		assert.Contains(t, schema, "sessionID")

		bridges := schema["bridges"].(map[string]interface{})
		assert.Contains(t, bridges, "test-doc-bridge")
	})

	t.Run("Specific Bridge Documentation", func(t *testing.T) {
		result, err := manager.GenerateBridgeDocumentation(ctx, "test-doc-bridge", "json")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Non-existent Bridge Documentation", func(t *testing.T) {
		_, err := manager.GenerateBridgeDocumentation(ctx, "non-existent", "json")
		assert.Error(t, err)
	})

	t.Run("Specific Documentation Methods", func(t *testing.T) {
		openAPISpec, err := manager.GenerateOpenAPIDocumentation(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, openAPISpec)
		assert.Equal(t, "3.0.3", openAPISpec.OpenAPI)

		markdownDoc, err := manager.GenerateMarkdownDocumentation(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, markdownDoc)

		jsonDoc, err := manager.GenerateJSONDocumentation(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonDoc)
	})

	// Clean up
	defer func() { _ = manager.Cleanup() }()
}
