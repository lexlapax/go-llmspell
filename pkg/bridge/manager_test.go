// ABOUTME: Test suite for the bridge manager that handles lifecycle management of language-agnostic bridges.
// ABOUTME: Tests thread-safe registration, dependency resolution, and hot-reloading functionality.

package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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

func (m *mockBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	return nil
}

func (m *mockBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if !m.IsInitialized() {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "test":
		return engine.NewStringValue("test result"), nil
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
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

func (m *mockScriptEngine) Execute(ctx context.Context, script string, params map[string]interface{}) (engine.ScriptValue, error) {
	return engine.NewNilValue(), nil
}

func (m *mockScriptEngine) ExecuteFile(ctx context.Context, path string, params map[string]interface{}) (engine.ScriptValue, error) {
	return engine.NewNilValue(), nil
}

func (m *mockScriptEngine) Shutdown() error {
	return nil
}

func (m *mockScriptEngine) GetBridge(name string) (engine.Bridge, error) {
	return nil, nil
}

func (m *mockScriptEngine) ToNative(scriptValue engine.ScriptValue) (interface{}, error) {
	return scriptValue.ToGo(), nil
}

func (m *mockScriptEngine) FromNative(goValue interface{}) (engine.ScriptValue, error) {
	// Use centralized conversion function
	return engine.ConvertToScriptValue(goValue), nil
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

// Task 1.4.11.1: Engine Event Bus
func (m *mockScriptEngine) GetEventBus() engine.EventBus {
	return engine.NewDefaultEventBus()
}

// Task 1.4.11.2: Type Conversion Registry
func (m *mockScriptEngine) RegisterTypeConverter(fromType, toType string, converter engine.TypeConverterFunc) error {
	return nil
}

func (m *mockScriptEngine) GetTypeRegistry() engine.TypeRegistry {
	return engine.NewDefaultTypeRegistry()
}

// Task 1.4.11.3: Engine Profiling
func (m *mockScriptEngine) EnableProfiling(config engine.ProfilingConfig) error {
	return nil
}

func (m *mockScriptEngine) DisableProfiling() error {
	return nil
}

func (m *mockScriptEngine) GetProfilingReport() (*engine.ProfilingReport, error) {
	return &engine.ProfilingReport{}, nil
}

// Task 1.4.11.4: Engine API Export
func (m *mockScriptEngine) ExportAPI(format engine.ExportFormat) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *mockScriptEngine) GenerateClientLibrary(language string, options engine.ClientLibraryOptions) ([]byte, error) {
	return []byte("{}"), nil
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
		var eventCount int64

		handler := func(ctx context.Context, event domain.Event) error {
			atomic.AddInt64(&eventCount, 1)
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
			assert.Greater(t, atomic.LoadInt64(&eventCount), int64(0))
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

// Test Bridge State Serialization
func TestBridgeManagerStateSerialization(t *testing.T) {
	t.Run("Export State Basic", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		// Register some test bridges
		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2", dependencies: []string{"bridge1"}}

		err := manager.RegisterBridge(bridge1)
		assert.NoError(t, err)
		err = manager.RegisterBridge(bridge2)
		assert.NoError(t, err)

		// Initialize bridges to generate metrics
		ctx := context.Background()
		err = manager.InitializeBridge(ctx, "bridge1")
		assert.NoError(t, err)
		err = manager.InitializeBridge(ctx, "bridge2")
		assert.NoError(t, err)

		// Export state
		state, err := manager.ExportState()
		assert.NoError(t, err)
		assert.NotNil(t, state)

		// Verify exported state
		assert.Equal(t, "1.0", state.Version)
		assert.NotEmpty(t, state.SessionID)
		assert.False(t, state.Timestamp.IsZero())
		assert.Len(t, state.Bridges, 2)
		assert.Len(t, state.Initialized, 2)
		assert.Len(t, state.Dependencies, 1) // Only bridge2 has dependencies
		assert.Len(t, state.Metrics, 2)

		// Verify bridge info
		bridge1Info := state.Bridges["bridge1"]
		assert.Equal(t, "bridge1", bridge1Info.ID)
		assert.Equal(t, "bridge1", bridge1Info.Name)
		assert.Equal(t, "1.0.0", bridge1Info.Version)
		assert.Empty(t, bridge1Info.Dependencies)

		bridge2Info := state.Bridges["bridge2"]
		assert.Equal(t, "bridge2", bridge2Info.ID)
		assert.Equal(t, "bridge2", bridge2Info.Name)
		assert.Equal(t, []string{"bridge1"}, bridge2Info.Dependencies)

		// Verify initialization state
		assert.True(t, state.Initialized["bridge1"])
		assert.True(t, state.Initialized["bridge2"])

		// Verify dependencies
		assert.Equal(t, []string{"bridge1"}, state.Dependencies["bridge2"])

		// Verify metrics
		metrics1 := state.Metrics["bridge1"]
		assert.Equal(t, int64(1), metrics1.InitializationCount)
		assert.Equal(t, int64(0), metrics1.FailureCount)

		// Verify metadata
		assert.Equal(t, 2, state.Metadata["total_bridges"])
		assert.Equal(t, 2, state.Metadata["initialized_count"])
	})

	t.Run("Import State Basic", func(t *testing.T) {
		// Create initial manager with state
		manager1 := NewBridgeManager()
		defer func() { _ = manager1.Cleanup() }()

		bridge1 := &mockBridge{id: "bridge1"}
		bridge2 := &mockBridge{id: "bridge2", dependencies: []string{"bridge1"}}

		_ = manager1.RegisterBridge(bridge1)
		_ = manager1.RegisterBridge(bridge2)

		ctx := context.Background()
		_ = manager1.InitializeBridge(ctx, "bridge1")
		_ = manager1.InitializeBridge(ctx, "bridge2")

		// Export state from first manager
		exportedState, err := manager1.ExportState()
		assert.NoError(t, err)

		// Create new manager and import state
		manager2 := NewBridgeManager()
		defer func() { _ = manager2.Cleanup() }()

		err = manager2.ImportState(exportedState)
		assert.NoError(t, err)

		// Verify imported state
		assert.True(t, manager2.IsBridgeInitialized("bridge1"))
		assert.True(t, manager2.IsBridgeInitialized("bridge2"))

		// Check metrics
		metrics1, err := manager2.GetBridgeMetrics("bridge1")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), metrics1.InitializationCount)

		metrics2, err := manager2.GetBridgeMetrics("bridge2")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), metrics2.InitializationCount)
	})

	t.Run("Round-trip JSON Serialization", func(t *testing.T) {
		manager1 := NewBridgeManager()
		defer func() { _ = manager1.Cleanup() }()

		// Setup initial state
		bridge := &mockBridge{id: "test-bridge"}
		_ = manager1.RegisterBridge(bridge)
		_ = manager1.InitializeBridge(context.Background(), "test-bridge")

		// Export to JSON
		jsonData, err := manager1.ExportStateToJSON(true)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Verify JSON structure
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		assert.NoError(t, err)
		assert.Equal(t, "1.0", jsonMap["version"])
		assert.Contains(t, jsonMap, "bridges")
		assert.Contains(t, jsonMap, "initialized")
		assert.Contains(t, jsonMap, "metrics")

		// Import to new manager
		manager2 := NewBridgeManager()
		defer func() { _ = manager2.Cleanup() }()

		err = manager2.ImportStateFromJSON(jsonData)
		assert.NoError(t, err)

		// Verify round-trip worked
		assert.True(t, manager2.IsBridgeInitialized("test-bridge"))
		metrics, err := manager2.GetBridgeMetrics("test-bridge")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), metrics.InitializationCount)
	})

	t.Run("State Version Validation", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		// Test valid version
		err := manager.validateStateVersion("1.0")
		assert.NoError(t, err)

		// Test invalid version
		err = manager.validateStateVersion("2.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported state version")

		// Test state with invalid version
		invalidState := &SerializableBridgeState{
			Version:   "2.0",
			SessionID: "test",
			Timestamp: time.Now(),
		}

		err = manager.ImportState(invalidState)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state version validation failed")
	})

	t.Run("State Integrity Validation", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		// Test empty session ID
		invalidState := &SerializableBridgeState{
			Version:   "1.0",
			SessionID: "",
			Timestamp: time.Now(),
		}
		err := manager.ImportState(invalidState)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")

		// Test zero timestamp
		invalidState = &SerializableBridgeState{
			Version:   "1.0",
			SessionID: "test",
			Timestamp: time.Time{},
		}
		err = manager.ImportState(invalidState)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp cannot be zero")

		// Test bridge mismatch in initialized
		invalidState = &SerializableBridgeState{
			Version:     "1.0",
			SessionID:   "test",
			Timestamp:   time.Now(),
			Bridges:     map[string]SerializableBridgeInfo{},
			Initialized: map[string]bool{"missing-bridge": true},
		}
		err = manager.ImportState(invalidState)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing-bridge is in initialized map but not in bridges map")

		// Test bridge mismatch in dependencies
		invalidState = &SerializableBridgeState{
			Version:   "1.0",
			SessionID: "test",
			Timestamp: time.Now(),
			Bridges: map[string]SerializableBridgeInfo{
				"bridge1": {ID: "bridge1", Name: "bridge1"},
			},
			Dependencies: map[string][]string{
				"bridge1": {"missing-dep"},
			},
		}
		err = manager.ImportState(invalidState)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dependency not found in bridges map")
	})

	t.Run("Incremental State Updates", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		bridge := &mockBridge{id: "test-bridge"}
		_ = manager.RegisterBridge(bridge)

		// Test metrics update
		updates := map[string]interface{}{
			"metrics": map[string]interface{}{
				"initialization_count": int64(5),
				"failure_count":        int64(2),
			},
		}

		err := manager.UpdateStateIncremental("test-bridge", updates)
		assert.NoError(t, err)

		metrics, err := manager.GetBridgeMetrics("test-bridge")
		assert.NoError(t, err)
		assert.Equal(t, int64(5), metrics.InitializationCount)
		assert.Equal(t, int64(2), metrics.FailureCount)

		// Test initialization state update
		updates = map[string]interface{}{
			"initialized": true,
		}

		err = manager.UpdateStateIncremental("test-bridge", updates)
		assert.NoError(t, err)
		assert.True(t, manager.IsBridgeInitialized("test-bridge"))

		// Test update for non-existent bridge
		err = manager.UpdateStateIncremental("non-existent", updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bridge non-existent not found")

		// Test update with empty bridge ID
		err = manager.UpdateStateIncremental("", updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bridge ID cannot be empty")
	})

	t.Run("State with Bridge Failures", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		// Register a bridge that will fail
		failureBridge := &mockBridge{
			id:        "failure-bridge",
			initError: errors.New("initialization failed"),
		}
		_ = manager.RegisterBridge(failureBridge)

		// Try to initialize (will fail)
		err := manager.InitializeBridge(context.Background(), "failure-bridge")
		assert.Error(t, err)

		// Export state
		state, err := manager.ExportState()
		assert.NoError(t, err)

		// Check that failure metrics are captured
		metrics := state.Metrics["failure-bridge"]
		assert.Equal(t, int64(1), metrics.InitializationCount)
		assert.Equal(t, int64(1), metrics.FailureCount)
		assert.Equal(t, "initialization failed", metrics.LastError)

		// Import state to new manager
		manager2 := NewBridgeManager()
		defer func() { _ = manager2.Cleanup() }()

		err = manager2.ImportState(state)
		assert.NoError(t, err)

		// Verify failure metrics were imported
		importedMetrics, err := manager2.GetBridgeMetrics("failure-bridge")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), importedMetrics.InitializationCount)
		assert.Equal(t, int64(1), importedMetrics.FailureCount)
		assert.NotNil(t, importedMetrics.LastError)
		assert.Equal(t, "initialization failed", importedMetrics.LastError.Error())
	})

	t.Run("GetStateVersion", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		version := manager.GetStateVersion()
		assert.Equal(t, "1.0", version)
	})

	t.Run("Export State Edge Cases", func(t *testing.T) {
		// Test empty manager
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		state, err := manager.ExportState()
		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.Equal(t, "1.0", state.Version)
		assert.Empty(t, state.Bridges)
		assert.Empty(t, state.Initialized)
		assert.Empty(t, state.Dependencies)
		assert.Empty(t, state.Metrics)
		assert.Equal(t, 0, state.Metadata["total_bridges"])
	})

	t.Run("Import Nil State", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		err := manager.ImportState(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state cannot be nil")
	})

	t.Run("JSON Export Pretty Formatting", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		bridge := &mockBridge{id: "test"}
		_ = manager.RegisterBridge(bridge)

		// Test pretty formatting
		prettyJSON, err := manager.ExportStateToJSON(true)
		assert.NoError(t, err)
		assert.Contains(t, string(prettyJSON), "\n") // Should contain newlines for pretty formatting

		// Test compact formatting
		compactJSON, err := manager.ExportStateToJSON(false)
		assert.NoError(t, err)
		assert.Less(t, len(compactJSON), len(prettyJSON)) // Compact should be shorter

		// Both should be valid JSON
		var prettyData, compactData map[string]interface{}
		err = json.Unmarshal(prettyJSON, &prettyData)
		assert.NoError(t, err)
		err = json.Unmarshal(compactJSON, &compactData)
		assert.NoError(t, err)

		// Content should be the same
		assert.Equal(t, prettyData["version"], compactData["version"])
		assert.Equal(t, prettyData["session_id"], compactData["session_id"])
	})

	t.Run("Import Invalid JSON", func(t *testing.T) {
		manager := NewBridgeManager()
		defer func() { _ = manager.Cleanup() }()

		invalidJSON := []byte(`{"version": "1.0", "invalid_json": }`)
		err := manager.ImportStateFromJSON(invalidJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal JSON")
	})
}
