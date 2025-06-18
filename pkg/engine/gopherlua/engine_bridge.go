// ABOUTME: Bridge registration and management for LuaEngine
// ABOUTME: Handles bridge lifecycle, module creation, method wrapping, and Lua-side access

package gopherlua

import (
	"context"
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// BridgeManager manages bridge registration and Lua module creation
type BridgeManager struct {
	bridges   map[string]engine.Bridge
	modules   map[string]*lua.LTable
	converter *LuaTypeConverter
	mu        sync.RWMutex
}

// NewBridgeManager creates a new bridge manager
func NewBridgeManager(converter *LuaTypeConverter) *BridgeManager {
	return &BridgeManager{
		bridges:   make(map[string]engine.Bridge),
		modules:   make(map[string]*lua.LTable),
		converter: converter,
	}
}

// RegisterBridge registers a bridge with the engine and creates its Lua module
func (bm *BridgeManager) RegisterBridge(bridge engine.Bridge) error {
	if bridge == nil {
		return fmt.Errorf("bridge cannot be nil")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	id := bridge.GetID()
	if id == "" {
		return fmt.Errorf("bridge ID cannot be empty")
	}

	// Check for duplicate registration
	if _, exists := bm.bridges[id]; exists {
		return fmt.Errorf("bridge %s already registered", id)
	}

	// Initialize bridge if needed
	if !bridge.IsInitialized() {
		ctx := context.Background()
		if err := bridge.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize bridge %s: %w", id, err)
		}
	}

	// Register bridge
	bm.bridges[id] = bridge

	return nil
}

// UnregisterBridge unregisters a bridge and cleans up its resources
func (bm *BridgeManager) UnregisterBridge(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bridge, exists := bm.bridges[id]
	if !exists {
		return fmt.Errorf("bridge %s not found", id)
	}

	// Cleanup bridge
	if bridge.IsInitialized() {
		ctx := context.Background()
		if err := bridge.Cleanup(ctx); err != nil {
			return fmt.Errorf("failed to cleanup bridge %s: %w", id, err)
		}
	}

	// Remove from maps
	delete(bm.bridges, id)
	delete(bm.modules, id)

	return nil
}

// GetBridge retrieves a bridge by ID
func (bm *BridgeManager) GetBridge(id string) (engine.Bridge, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bridge, exists := bm.bridges[id]
	if !exists {
		return nil, fmt.Errorf("bridge %s not found", id)
	}

	return bridge, nil
}

// ListBridges returns a list of registered bridge IDs
func (bm *BridgeManager) ListBridges() []string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	ids := make([]string, 0, len(bm.bridges))
	for id := range bm.bridges {
		ids = append(ids, id)
	}

	return ids
}

// CreateLuaModule creates a Lua module for a bridge
func (bm *BridgeManager) CreateLuaModule(L *lua.LState, bridgeID string) (*lua.LTable, error) {
	bm.mu.RLock()
	bridge, exists := bm.bridges[bridgeID]
	bm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("bridge %s not found", bridgeID)
	}

	// Check if module already exists
	bm.mu.Lock()
	if module, exists := bm.modules[bridgeID]; exists {
		bm.mu.Unlock()
		return module, nil
	}
	bm.mu.Unlock()

	// Create new module
	module := L.NewTable()

	// Add metadata
	meta := bridge.GetMetadata()
	metaTable := L.NewTable()
	metaTable.RawSetString("name", lua.LString(meta.Name))
	metaTable.RawSetString("version", lua.LString(meta.Version))
	metaTable.RawSetString("description", lua.LString(meta.Description))
	if meta.Author != "" {
		metaTable.RawSetString("author", lua.LString(meta.Author))
	}
	if meta.License != "" {
		metaTable.RawSetString("license", lua.LString(meta.License))
	}
	module.RawSetString("_meta", metaTable)

	// Add methods
	methods := bridge.Methods()
	for _, methodInfo := range methods {
		wrappedMethod := bm.wrapBridgeMethod(L, bridge, methodInfo)
		module.RawSetString(methodInfo.Name, wrappedMethod)
	}

	// Cache the module
	bm.mu.Lock()
	bm.modules[bridgeID] = module
	bm.mu.Unlock()

	return module, nil
}

// wrapBridgeMethod wraps a bridge method for Lua consumption
func (bm *BridgeManager) wrapBridgeMethod(L *lua.LState, bridge engine.Bridge, methodInfo engine.MethodInfo) *lua.LFunction {
	return L.NewFunction(func(L *lua.LState) int {
		// Get number of arguments
		argc := L.GetTop()

		// Extract arguments and convert to ScriptValue
		args := make([]engine.ScriptValue, argc)
		for i := 1; i <= argc; i++ {
			luaValue := L.Get(i)
			scriptValue, err := bm.converter.ToLuaScriptValue(L, luaValue)
			if err != nil {
				L.ArgError(i, fmt.Sprintf("failed to convert argument to ScriptValue: %v", err))
				return 0
			}
			args[i-1] = scriptValue
		}

		// Validate method call
		if err := bridge.ValidateMethod(methodInfo.Name, args); err != nil {
			L.RaiseError("validation failed: %v", err)
			return 0
		}

		// Execute the method on the bridge
		ctx := context.Background() // TODO: Pass context from caller
		result, err := bridge.ExecuteMethod(ctx, methodInfo.Name, args)
		if err != nil {
			L.RaiseError("execution failed: %v", err)
			return 0
		}

		// Convert result ScriptValue back to Lua value
		if result != nil {
			luaResult, err := bm.converter.FromLuaScriptValue(L, result)
			if err != nil {
				L.RaiseError("failed to convert result: %v", err)
				return 0
			}
			L.Push(luaResult)
			return 1
		} else {
			L.Push(lua.LNil)
			return 1
		}
	})
}

// LoadBridgeModules loads all registered bridge modules into a Lua state
func (bm *BridgeManager) LoadBridgeModules(L *lua.LState) error {
	bm.mu.RLock()
	bridgeIDs := make([]string, 0, len(bm.bridges))
	for id := range bm.bridges {
		bridgeIDs = append(bridgeIDs, id)
	}
	bm.mu.RUnlock()

	// Create a bridges global table
	bridgesTable := L.NewTable()

	for _, bridgeID := range bridgeIDs {
		module, err := bm.CreateLuaModule(L, bridgeID)
		if err != nil {
			return fmt.Errorf("failed to create module for bridge %s: %w", bridgeID, err)
		}

		bridgesTable.RawSetString(bridgeID, module)
	}

	// Set global bridges table
	L.SetGlobal("bridges", bridgesTable)

	return nil
}

// GetBridgeMetadata returns metadata for all registered bridges
func (bm *BridgeManager) GetBridgeMetadata() map[string]engine.BridgeMetadata {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	metadata := make(map[string]engine.BridgeMetadata)
	for id, bridge := range bm.bridges {
		metadata[id] = bridge.GetMetadata()
	}

	return metadata
}

// ValidateBridgeMethod validates a method call against a bridge
func (bm *BridgeManager) ValidateBridgeMethod(bridgeID, methodName string, args []engine.ScriptValue) error {
	bm.mu.RLock()
	bridge, exists := bm.bridges[bridgeID]
	bm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("bridge %s not found", bridgeID)
	}

	return bridge.ValidateMethod(methodName, args)
}

// Cleanup cleans up all registered bridges
func (bm *BridgeManager) Cleanup() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	var errs []error
	ctx := context.Background()

	for id, bridge := range bm.bridges {
		if bridge.IsInitialized() {
			if err := bridge.Cleanup(ctx); err != nil {
				errs = append(errs, fmt.Errorf("failed to cleanup bridge %s: %w", id, err))
			}
		}
	}

	// Clear maps
	bm.bridges = make(map[string]engine.Bridge)
	bm.modules = make(map[string]*lua.LTable)

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}
