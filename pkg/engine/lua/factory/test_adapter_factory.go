// ABOUTME: Test adapter factory that extends production factory with test bridge support
// ABOUTME: Provides mock adapters for test bridges while maintaining production adapter functionality

package factory

import (
	"fmt"
	
	"github.com/lexlapax/go-llmspell/pkg/engine"
	lua "github.com/yuin/gopher-lua"
)

// TestAdapterFactory extends AdapterFactory with support for test bridge IDs.
// This allows tests to register mock bridges without polluting the production factory.
type TestAdapterFactory struct {
	*AdapterFactory
}

// NewTestAdapterFactory creates a new test adapter factory
// that supports both production and test bridge IDs.
func NewTestAdapterFactory() *TestAdapterFactory {
	return &TestAdapterFactory{
		AdapterFactory: NewAdapterFactory(),
	}
}

// CreateAdapter creates an adapter for the given bridge ID.
// It first tries production bridge IDs, then falls back to test bridge IDs.
func (f *TestAdapterFactory) CreateAdapter(bridgeID string, bridgeMap map[string]engine.Bridge) (interface{}, error) {
	// Try production factory first
	adapter, err := f.AdapterFactory.CreateAdapter(bridgeID, bridgeMap)
	if err == nil {
		return adapter, nil
	}
	
	// Check if it's a test bridge ID
	if f.isTestBridgeID(bridgeID) {
		return f.createTestAdapter(bridgeID, bridgeMap)
	}
	
	// Return original error if not a test bridge
	return nil, err
}

// isTestBridgeID checks if the bridge ID is a known test bridge
func (f *TestAdapterFactory) isTestBridgeID(bridgeID string) bool {
	testBridgeIDs := []string{
		"test_bridge",
		"test_integration_bridge", 
		"bridge_1",
		"bridge_2", 
		"bridge_3",
		"bridge1",
		"bridge2",
		"concurrent_bridge",
		// Add more test bridge IDs as needed
	}
	
	for _, testID := range testBridgeIDs {
		if bridgeID == testID {
			return true
		}
	}
	return false
}

// createTestAdapter creates a mock adapter for test bridges
func (f *TestAdapterFactory) createTestAdapter(bridgeID string, bridgeMap map[string]engine.Bridge) (interface{}, error) {
	bridge, ok := bridgeMap[bridgeID]
	if !ok {
		return nil, fmt.Errorf("bridge not found for ID: %s", bridgeID)
	}
	
	return &TestBridgeAdapter{
		bridgeID: bridgeID,
		bridge:   bridge,
	}, nil
}

// TestBridgeAdapter is a simple mock adapter for test bridges
type TestBridgeAdapter struct {
	bridgeID string
	bridge   engine.Bridge
}

// CreateLuaModule creates a simple Lua module for the test bridge
func (a *TestBridgeAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		module := L.NewTable()
		
		// Add bridge metadata from the actual bridge
		bridgeMeta := a.bridge.GetMetadata()
		meta := L.NewTable()
		L.SetField(meta, "bridge_id", lua.LString(a.bridgeID))
		L.SetField(meta, "type", lua.LString("test_adapter"))
		L.SetField(meta, "name", lua.LString(bridgeMeta.Name))
		L.SetField(meta, "version", lua.LString(bridgeMeta.Version))
		L.SetField(meta, "description", lua.LString(bridgeMeta.Description))
		L.SetField(module, "_meta", meta)
		
		// Add test methods based on bridge methods
		bridgeMethods := a.bridge.Methods()
		for _, method := range bridgeMethods {
			methodName := method.Name
			L.SetField(module, methodName, L.NewFunction(func(L *lua.LState) int {
				L.Push(lua.LString("test result from " + a.bridgeID + "." + methodName))
				return 1
			}))
		}
		
		// Add a simple test method if no methods exist
		if len(bridgeMethods) == 0 {
			L.SetField(module, "test_method", L.NewFunction(func(L *lua.LState) int {
				L.Push(lua.LString("test result from " + a.bridgeID))
				return 1
			}))
		}
		
		L.Push(module)
		return 1
	}
}