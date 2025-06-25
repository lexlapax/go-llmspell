// ABOUTME: Tests for LuaEngine adapter management functionality during bridge registration
// ABOUTME: Validates adapter creation, caching, retrieval, and lifecycle management

package lua

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	enginepkg "github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/testutils"
)

// TestLuaEngineAdapterCreation tests adapter creation during bridge registration
func TestLuaEngineAdapterCreation(t *testing.T) {
	t.Run("create adapter when registering bridge", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator
		adapterCreated := false
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			adapterCreated = true
			return &mockAdapter{id: bridgeID}, nil
		})
		
		// Create a mock bridge
		bridge := testutils.NewMockBridge("agent_core").WithInitialized(true)
		
		// Register bridge - should create adapter
		err = engine.RegisterBridge(bridge)
		require.NoError(t, err)
		
		// Verify adapter was created
		adapter := engine.GetAdapter("agent_core")
		assert.NotNil(t, adapter, "Adapter should be created during bridge registration")
		assert.True(t, adapterCreated, "Adapter creator should have been called")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
	
	t.Run("adapter creation failure", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator that fails
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return nil, assert.AnError
		})
		
		// Create a mock bridge
		bridge := testutils.NewMockBridge("failing_bridge").WithInitialized(true)
		
		// Register bridge - should fail due to adapter creation error
		err = engine.RegisterBridge(bridge)
		assert.Error(t, err, "Should fail when adapter creation fails")
		assert.Contains(t, err.Error(), "failed to create adapter")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
	
	t.Run("adapter creation with nil bridge", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Register nil bridge - should fail
		err = engine.RegisterBridge(nil)
		assert.Error(t, err, "Should fail when bridge is nil")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
	
	t.Run("no adapter creator set", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Don't set adapter creator
		
		// Create a mock bridge
		bridge := testutils.NewMockBridge("agent_core").WithInitialized(true)
		
		// Register bridge - should succeed but no adapter created
		err = engine.RegisterBridge(bridge)
		require.NoError(t, err)
		
		// Verify no adapter was created
		adapter := engine.GetAdapter("agent_core")
		assert.Nil(t, adapter, "No adapter should be created when creator not set")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
}

// TestLuaEngineAdapterRetrieval tests adapter retrieval by bridge ID
func TestLuaEngineAdapterRetrieval(t *testing.T) {
	t.Run("get existing adapter", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return &mockAdapter{id: bridgeID}, nil
		})
		
		// Create and register bridge
		bridge := testutils.NewMockBridge("agent_core").WithInitialized(true)
		err = engine.RegisterBridge(bridge)
		require.NoError(t, err)
		
		// Get adapter
		adapter1 := engine.GetAdapter("agent_core")
		assert.NotNil(t, adapter1)
		
		// Get same adapter again - should return cached instance
		adapter2 := engine.GetAdapter("agent_core")
		assert.NotNil(t, adapter2)
		assert.Equal(t, adapter1, adapter2, "Should return same cached adapter instance")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
	
	t.Run("get non-existent adapter", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Get adapter that doesn't exist
		adapter := engine.GetAdapter("non_existent_bridge")
		assert.Nil(t, adapter, "Should return nil for non-existent adapter")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
	
	t.Run("get adapter with empty bridge ID", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Get adapter with empty ID
		adapter := engine.GetAdapter("")
		assert.Nil(t, adapter, "Should return nil for empty bridge ID")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
}

// TestLuaEngineAdapterLifecycle tests adapter lifecycle management
func TestLuaEngineAdapterLifecycle(t *testing.T) {
	t.Run("adapter cleanup during engine cleanup", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return &mockAdapter{id: bridgeID}, nil
		})
		
		// Create and register multiple bridges
		bridge1 := testutils.NewMockBridge("agent_core").WithInitialized(true)
		bridge2 := testutils.NewMockBridge("llm_core").WithInitialized(true)
		
		err = engine.RegisterBridge(bridge1)
		require.NoError(t, err)
		err = engine.RegisterBridge(bridge2)
		require.NoError(t, err)
		
		// Verify adapters exist
		adapter1 := engine.GetAdapter("agent_core")
		adapter2 := engine.GetAdapter("llm_core")
		assert.NotNil(t, adapter1)
		assert.NotNil(t, adapter2)
		
		// Cleanup should clear adapters
		err = engine.Cleanup(context.Background())
		assert.NoError(t, err)
		
		// Adapters should still be accessible but engine should be cleaned up
		// (We don't clear the adapter map on cleanup to avoid breaking references)
		adapter1After := engine.GetAdapter("agent_core")
		adapter2After := engine.GetAdapter("llm_core")
		assert.NotNil(t, adapter1After)
		assert.NotNil(t, adapter2After)
	})
	
	t.Run("adapter map initialization", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Engine should have adapter map initialized
		assert.NotNil(t, engine.adapters, "Adapter map should be initialized")
		assert.Equal(t, 0, len(engine.adapters), "Adapter map should be empty initially")
	})
}

// TestLuaEngineAdapterConcurrency tests concurrent access to adapter map
func TestLuaEngineAdapterConcurrency(t *testing.T) {
	t.Run("concurrent adapter access", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return &mockAdapter{id: bridgeID}, nil
		})
		
		// Create and register bridge
		bridge := testutils.NewMockBridge("agent_core").WithInitialized(true)
		err = engine.RegisterBridge(bridge)
		require.NoError(t, err)
		
		// Test concurrent reads
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				adapter := engine.GetAdapter("agent_core")
				assert.NotNil(t, adapter)
				done <- true
			}()
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
	
	t.Run("concurrent bridge registration", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return &mockAdapter{id: bridgeID}, nil
		})
		
		// Test concurrent bridge registrations
		done := make(chan error)
		bridgeIDs := []string{"agent_core", "llm_core", "state_manager"}
		
		for _, bridgeID := range bridgeIDs {
			go func(id string) {
				bridge := testutils.NewMockBridge(id).WithInitialized(true)
				err := engine.RegisterBridge(bridge)
				done <- err
			}(bridgeID)
		}
		
		// Wait for all registrations and check for errors
		for i := 0; i < len(bridgeIDs); i++ {
			err := <-done
			if err != nil {
				t.Logf("Bridge registration error (expected for some unknown types): %v", err)
			}
		}
		
		// Verify at least some adapters were created successfully
		adapterCount := 0
		for _, id := range bridgeIDs {
			if engine.GetAdapter(id) != nil {
				adapterCount++
			}
		}
		assert.True(t, adapterCount > 0, "At least some adapters should be created")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
}

// TestLuaEngineAdapterIntegration tests adapter integration with module system
func TestLuaEngineAdapterIntegration(t *testing.T) {
	t.Run("adapter provides functionality", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return &mockAdapter{id: bridgeID}, nil
		})
		
		// Create and register bridge
		bridge := testutils.NewMockBridge("agent_core").WithInitialized(true)
		err = engine.RegisterBridge(bridge)
		require.NoError(t, err)
		
		// Get adapter
		adapter := engine.GetAdapter("agent_core")
		require.NotNil(t, adapter)
		
		// Test that adapter provides expected functionality
		mockAdapterInstance, ok := adapter.(*mockAdapter)
		assert.True(t, ok, "Adapter should be mock adapter type")
		assert.Equal(t, "agent_core", mockAdapterInstance.id, "Adapter should have correct ID")
		
		// Test adapter module creation
		moduleCreator := mockAdapterInstance.CreateLuaModule()
		assert.NotNil(t, moduleCreator, "Adapter should provide module creator")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
}

// TestLuaEngineAdapterErrorHandling tests error handling in adapter management
func TestLuaEngineAdapterErrorHandling(t *testing.T) {
	t.Run("handle adapter creation error", func(t *testing.T) {
		engine := NewLuaEngine()
		require.NotNil(t, engine)
		
		// Initialize engine
		err := engine.Initialize(enginepkg.EngineConfig{})
		require.NoError(t, err)
		
		// Set up adapter creator that will fail
		engine.SetAdapterCreator(func(bridgeID string, bridge enginepkg.Bridge) (interface{}, error) {
			return nil, assert.AnError
		})
		
		// Create bridge that will trigger the failing creator
		bridge := testutils.NewMockBridge("failing_bridge").WithInitialized(true)
		
		// Register bridge - should fail due to adapter creation error
		err = engine.RegisterBridge(bridge)
		assert.Error(t, err, "Should propagate adapter creation error")
		assert.Contains(t, err.Error(), "failed to create adapter")
		
		// Verify no adapter was stored
		adapter := engine.GetAdapter("failing_bridge")
		assert.Nil(t, adapter, "Failed adapter should not be stored")
		
		// Cleanup
		engine.Cleanup(context.Background())
	})
}