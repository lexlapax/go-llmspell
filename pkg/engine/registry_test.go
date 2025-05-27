// ABOUTME: Tests for the engine registry and factory pattern
// ABOUTME: Validates thread-safe registration and engine creation

package engine

import (
	"fmt"
	"sync"
	"testing"
)

func TestRegistry(t *testing.T) {
	t.Run("register and get engine factory", func(t *testing.T) {
		registry := NewRegistry()

		// Register a mock engine factory
		factory := func(config Config) (Engine, error) {
			return newMockEngine("mock"), nil
		}

		err := registry.Register("mock", factory)
		if err != nil {
			t.Fatalf("Failed to register engine: %v", err)
		}

		// Get the factory back
		retrievedFactory, err := registry.GetFactory("mock")
		if err != nil {
			t.Fatalf("Failed to get factory: %v", err)
		}

		if retrievedFactory == nil {
			t.Error("Retrieved factory is nil")
		}

		// Create engine using factory
		engine, err := retrievedFactory(Config{})
		if err != nil {
			t.Fatalf("Failed to create engine: %v", err)
		}

		if engine.Name() != "mock" {
			t.Errorf("Expected engine name 'mock', got %s", engine.Name())
		}
	})

	t.Run("register duplicate engine", func(t *testing.T) {
		registry := NewRegistry()

		factory := func(config Config) (Engine, error) {
			return newMockEngine("test"), nil
		}

		// First registration should succeed
		err := registry.Register("mock", factory)
		if err != nil {
			t.Fatalf("First registration failed: %v", err)
		}

		// Second registration should fail
		err = registry.Register("mock", factory)
		if err == nil {
			t.Error("Expected error when registering duplicate engine")
		}
	})

	t.Run("get non-existent factory", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.GetFactory("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent factory")
		}
	})

	t.Run("list registered engines", func(t *testing.T) {
		registry := NewRegistry()

		// Register multiple engines
		engines := []string{"lua", "javascript", "tengo"}
		for _, name := range engines {
			factory := func(config Config) (Engine, error) {
				return newMockEngine(name), nil
			}
			if err := registry.Register(name, factory); err != nil {
				t.Fatalf("Failed to register %s: %v", name, err)
			}
		}

		// List should contain all registered engines
		registered := registry.List()
		if len(registered) != len(engines) {
			t.Errorf("Expected %d engines, got %d", len(engines), len(registered))
		}

		// Check each engine is in the list
		for _, expected := range engines {
			found := false
			for _, actual := range registered {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Engine %s not found in list", expected)
			}
		}
	})

	t.Run("unregister engine", func(t *testing.T) {
		registry := NewRegistry()

		factory := func(config Config) (Engine, error) {
			return newMockEngine("test"), nil
		}

		// Register engine
		if err := registry.Register("mock", factory); err != nil {
			t.Fatalf("Failed to register engine: %v", err)
		}

		// Verify it exists
		if _, err := registry.GetFactory("mock"); err != nil {
			t.Error("Engine should exist after registration")
		}

		// Unregister
		if err := registry.Unregister("mock"); err != nil {
			t.Fatalf("Failed to unregister engine: %v", err)
		}

		// Verify it's gone
		if _, err := registry.GetFactory("mock"); err == nil {
			t.Error("Engine should not exist after unregistration")
		}
	})
}

func TestRegistryConcurrency(t *testing.T) {
	registry := NewRegistry()

	// Number of concurrent operations
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numOps * 3) // 3 types of operations

	// Concurrent registrations
	for i := 0; i < numOps; i++ {
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("engine%d", n)
			factory := func(config Config) (Engine, error) {
				return newMockEngine(name), nil
			}
			_ = registry.Register(name, factory)
		}(i)
	}

	// Concurrent gets
	for i := 0; i < numOps; i++ {
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("engine%d", n%10) // Try to get some that exist
			_, _ = registry.GetFactory(name)
		}(i)
	}

	// Concurrent lists
	for i := 0; i < numOps; i++ {
		go func() {
			defer wg.Done()
			registry.List()
		}()
	}

	wg.Wait()

	// Verify some engines were registered
	list := registry.List()
	if len(list) == 0 {
		t.Error("No engines were registered during concurrent test")
	}
}

func TestEngineFactory(t *testing.T) {
	t.Run("create engine with config", func(t *testing.T) {
		factory := func(config Config) (Engine, error) {
			engine := &mockEngineWithConfig{
				mockEngine: mockEngine{
					name: "test",
				},
				config:      config,
				initialized: true,
			}
			return engine, nil
		}

		config := Config{
			MaxMemory:        64 * 1024 * 1024,
			MaxExecutionTime: 30,
			EnableDebug:      true,
		}

		engine, err := factory(config)
		if err != nil {
			t.Fatalf("Failed to create engine: %v", err)
		}

		mockEngine, ok := engine.(*mockEngineWithConfig)
		if !ok {
			t.Fatal("Engine is not of expected type")
		}

		if !mockEngine.initialized {
			t.Error("Engine was not initialized")
		}

		if mockEngine.config.MaxMemory != config.MaxMemory {
			t.Error("Config was not properly passed to engine")
		}
	})

	t.Run("factory with initialization error", func(t *testing.T) {
		factory := func(config Config) (Engine, error) {
			return nil, ErrInvalidConfiguration
		}

		_, err := factory(Config{})
		if err != ErrInvalidConfiguration {
			t.Errorf("Expected ErrInvalidConfiguration, got %v", err)
		}
	})
}

// mockEngineWithConfig extends mockEngine to store configuration
type mockEngineWithConfig struct {
	mockEngine
	config      Config
	initialized bool
}

func (m *mockEngineWithConfig) Initialize(config Config) error {
	m.config = config
	return nil
}

func TestGlobalRegistry(t *testing.T) {
	// Reset global registry for testing
	globalRegistry = NewRegistry()

	t.Run("register to global registry", func(t *testing.T) {
		factory := func(config Config) (Engine, error) {
			return newMockEngine("global"), nil
		}

		err := RegisterEngine("global", factory)
		if err != nil {
			t.Fatalf("Failed to register to global registry: %v", err)
		}

		// Should be able to get from global registry
		engine, err := CreateEngine("global", Config{})
		if err != nil {
			t.Fatalf("Failed to create engine from global registry: %v", err)
		}

		if engine.Name() != "global" {
			t.Errorf("Expected engine name 'global', got %s", engine.Name())
		}
	})

	t.Run("list global engines", func(t *testing.T) {
		engines := ListEngines()
		found := false
		for _, name := range engines {
			if name == "global" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Global engine not found in list")
		}
	})
}

func TestAutoDiscovery(t *testing.T) {
	t.Run("discover engines from registered factories", func(t *testing.T) {
		registry := NewRegistry()

		// Register some engines with metadata
		luaFactory := func(config Config) (Engine, error) {
			return &mockEngineWithConfig{
				mockEngine: mockEngine{name: "lua"},
				config:     config,
			}, nil
		}

		jsFactory := func(config Config) (Engine, error) {
			return &mockEngineWithConfig{
				mockEngine: mockEngine{name: "javascript"},
				config:     config,
			}, nil
		}

		if err := registry.RegisterWithMetadata("lua", luaFactory, EngineMetadata{
			Description:    "Lua scripting engine",
			FileExtensions: []string{".lua"},
			MimeTypes:      []string{"text/x-lua", "application/x-lua"},
		}); err != nil {
			t.Fatalf("Failed to register lua engine: %v", err)
		}
		if err := registry.RegisterWithMetadata("javascript", jsFactory, EngineMetadata{
			Description:    "JavaScript engine with ES2020 support",
			FileExtensions: []string{".js", ".mjs"},
			MimeTypes:      []string{"text/javascript", "application/javascript"},
		}); err != nil {
			t.Fatalf("Failed to register javascript engine: %v", err)
		}

		// Discover by file extension
		engine, err := registry.DiscoverByExtension(".lua")
		if err != nil {
			t.Fatalf("Failed to discover engine by extension: %v", err)
		}
		if engine != "lua" {
			t.Errorf("Expected lua engine, got %s", engine)
		}

		// Discover by mime type
		engine, err = registry.DiscoverByMimeType("application/javascript")
		if err != nil {
			t.Fatalf("Failed to discover engine by mime type: %v", err)
		}
		if engine != "javascript" {
			t.Errorf("Expected javascript engine, got %s", engine)
		}
	})
}
