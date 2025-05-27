// ABOUTME: Tests for the bridge interface and bridge management
// ABOUTME: Validates bridge registration, lifecycle, and method exposure

package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestBridgeInterface(t *testing.T) {
	// Ensure Bridge interface can be implemented
	var _ Bridge = (*mockBridge)(nil)
}

func TestBridgeRegistration(t *testing.T) {
	t.Run("register and expose methods", func(t *testing.T) {
		bridge := newMockBridge("test")

		// Get exposed methods
		methods := bridge.Methods()
		if len(methods) == 0 {
			t.Error("Bridge should expose at least one method")
		}

		// Check method exists
		found := false
		for _, method := range methods {
			if method.Name == "echo" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'echo' method to be exposed")
		}
	})

	t.Run("method metadata", func(t *testing.T) {
		bridge := newMockBridge("test")
		methods := bridge.Methods()

		var echoMethod *MethodInfo
		for _, method := range methods {
			if method.Name == "echo" {
				echoMethod = &method
				break
			}
		}

		if echoMethod == nil {
			t.Fatal("Echo method not found")
		}

		if echoMethod.Description == "" {
			t.Error("Method should have a description")
		}

		if len(echoMethod.Parameters) == 0 {
			t.Error("Echo method should have parameters")
		}

		if echoMethod.ReturnType == "" {
			t.Error("Method should have a return type")
		}
	})
}

func TestBridgeSet(t *testing.T) {
	t.Run("register bridge", func(t *testing.T) {
		set := NewBridgeSet()
		bridge := newMockBridge("test")

		err := set.Register("mock", bridge)
		if err != nil {
			t.Fatalf("Failed to register bridge: %v", err)
		}

		// Should be able to get it back
		retrieved, err := set.Get("mock")
		if err != nil {
			t.Fatalf("Failed to get bridge: %v", err)
		}

		if retrieved != bridge {
			t.Error("Retrieved different bridge instance")
		}
	})

	t.Run("register duplicate bridge", func(t *testing.T) {
		set := NewBridgeSet()
		bridge := newMockBridge("test")

		// First registration should succeed
		err := set.Register("mock", bridge)
		if err != nil {
			t.Fatalf("First registration failed: %v", err)
		}

		// Second registration should fail
		err = set.Register("mock", bridge)
		if err == nil {
			t.Error("Expected error when registering duplicate bridge")
		}
	})

	t.Run("get non-existent bridge", func(t *testing.T) {
		set := NewBridgeSet()

		_, err := set.Get("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent bridge")
		}
	})

	t.Run("list bridges", func(t *testing.T) {
		set := NewBridgeSet()

		bridges := []string{"llm", "tools", "workflow"}
		for _, name := range bridges {
			bridge := newMockBridge(name)
			if err := set.Register(name, bridge); err != nil {
				t.Fatalf("Failed to register %s: %v", name, err)
			}
		}

		registered := set.List()
		if len(registered) != len(bridges) {
			t.Errorf("Expected %d bridges, got %d", len(bridges), len(registered))
		}

		// Check each bridge is in the list
		for _, expected := range bridges {
			found := false
			for _, actual := range registered {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Bridge %s not found in list", expected)
			}
		}
	})

	t.Run("unregister bridge", func(t *testing.T) {
		set := NewBridgeSet()
		bridge := newMockBridge("test")

		// Register bridge
		if err := set.Register("mock", bridge); err != nil {
			t.Fatalf("Failed to register bridge: %v", err)
		}

		// Verify it exists
		if _, err := set.Get("mock"); err != nil {
			t.Error("Bridge should exist after registration")
		}

		// Unregister
		if err := set.Unregister("mock"); err != nil {
			t.Fatalf("Failed to unregister bridge: %v", err)
		}

		// Verify it's gone
		if _, err := set.Get("mock"); err == nil {
			t.Error("Bridge should not exist after unregistration")
		}
	})
}

func TestBridgeLifecycle(t *testing.T) {
	t.Run("initialize bridge", func(t *testing.T) {
		bridge := newMockBridge("test")

		ctx := context.Background()
		err := bridge.Initialize(ctx)
		if err != nil {
			t.Fatalf("Failed to initialize bridge: %v", err)
		}

		mockBridge := bridge.(*mockBridge)
		if !mockBridge.initialized {
			t.Error("Bridge should be marked as initialized")
		}
	})

	t.Run("cleanup bridge", func(t *testing.T) {
		bridge := newMockBridge("test")

		ctx := context.Background()
		// Initialize first
		if err := bridge.Initialize(ctx); err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}

		// Then cleanup
		err := bridge.Cleanup(ctx)
		if err != nil {
			t.Fatalf("Failed to cleanup bridge: %v", err)
		}

		mockBridge := bridge.(*mockBridge)
		if mockBridge.initialized {
			t.Error("Bridge should not be initialized after cleanup")
		}
	})

	t.Run("bridge set lifecycle", func(t *testing.T) {
		set := NewBridgeSet()

		// Register multiple bridges
		bridges := []string{"bridge1", "bridge2", "bridge3"}
		for _, name := range bridges {
			bridge := newMockBridge(name)
			if err := set.Register(name, bridge); err != nil {
				t.Fatalf("Failed to register %s: %v", name, err)
			}
		}

		ctx := context.Background()

		// Initialize all
		if err := set.InitializeAll(ctx); err != nil {
			t.Fatalf("Failed to initialize all bridges: %v", err)
		}

		// Verify all initialized
		for _, name := range bridges {
			bridge, _ := set.Get(name)
			mockBridge := bridge.(*mockBridge)
			if !mockBridge.initialized {
				t.Errorf("Bridge %s should be initialized", name)
			}
		}

		// Cleanup all
		if err := set.CleanupAll(ctx); err != nil {
			t.Fatalf("Failed to cleanup all bridges: %v", err)
		}

		// Verify all cleaned up
		for _, name := range bridges {
			bridge, _ := set.Get(name)
			mockBridge := bridge.(*mockBridge)
			if mockBridge.initialized {
				t.Errorf("Bridge %s should be cleaned up", name)
			}
		}
	})
}

func TestBridgeSetConcurrency(t *testing.T) {
	set := NewBridgeSet()

	const numOps = 100
	var wg sync.WaitGroup
	wg.Add(numOps * 3)

	// Concurrent registrations
	for i := 0; i < numOps; i++ {
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("bridge%d", n)
			bridge := newMockBridge(name)
			_ = set.Register(name, bridge)
		}(i)
	}

	// Concurrent gets
	for i := 0; i < numOps; i++ {
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("bridge%d", n%10)
			_, _ = set.Get(name)
		}(i)
	}

	// Concurrent lists
	for i := 0; i < numOps; i++ {
		go func() {
			defer wg.Done()
			_ = set.List()
		}()
	}

	wg.Wait()

	// Verify some bridges were registered
	list := set.List()
	if len(list) == 0 {
		t.Error("No bridges were registered during concurrent test")
	}
}

// mockBridge is a test implementation of the Bridge interface
type mockBridge struct {
	name        string
	initialized bool
	methods     []MethodInfo
	mu          sync.Mutex
}

func newMockBridge(name string) Bridge {
	return &mockBridge{
		name: name,
		methods: []MethodInfo{
			{
				Name:        "echo",
				Description: "Echoes the input back",
				Parameters: []ParameterInfo{
					{
						Name:        "message",
						Type:        "string",
						Description: "The message to echo",
						Required:    true,
					},
				},
				ReturnType: "string",
			},
			{
				Name:        "add",
				Description: "Adds two numbers",
				Parameters: []ParameterInfo{
					{
						Name:        "a",
						Type:        "number",
						Description: "First number",
						Required:    true,
					},
					{
						Name:        "b",
						Type:        "number",
						Description: "Second number",
						Required:    true,
					},
				},
				ReturnType: "number",
			},
		},
	}
}

func (m *mockBridge) Name() string {
	return m.name
}

func (m *mockBridge) Methods() []MethodInfo {
	return m.methods
}

func (m *mockBridge) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return errors.New("already initialized")
	}

	m.initialized = true
	return nil
}

func (m *mockBridge) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.initialized = false
	return nil
}

func TestGlobalBridgeSet(t *testing.T) {
	// Reset global bridge set for testing
	globalBridgeSet = NewBridgeSet()

	t.Run("register to global bridge set", func(t *testing.T) {
		bridge := newMockBridge("global")

		err := RegisterBridge("global", bridge)
		if err != nil {
			t.Fatalf("Failed to register to global bridge set: %v", err)
		}

		// Should be able to get from global set
		retrieved, err := GetBridge("global")
		if err != nil {
			t.Fatalf("Failed to get bridge from global set: %v", err)
		}

		if retrieved.Name() != "global" {
			t.Errorf("Expected bridge name 'global', got %s", retrieved.Name())
		}
	})

	t.Run("list global bridges", func(t *testing.T) {
		bridges := ListBridges()
		found := false
		for _, name := range bridges {
			if name == "global" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Global bridge not found in list")
		}
	})
}
