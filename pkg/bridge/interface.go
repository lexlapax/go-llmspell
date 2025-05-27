// ABOUTME: Core bridge interface for exposing Go functionality to scripts
// ABOUTME: Defines contracts for script-accessible bridges and management

package bridge

import (
	"context"
	"fmt"
	"sync"
)

// Bridge defines the interface that all bridges must implement
type Bridge interface {
	// Name returns the name of the bridge (e.g., "llm", "tools", "workflow")
	Name() string

	// Methods returns information about all methods exposed by this bridge
	Methods() []MethodInfo

	// Initialize prepares the bridge for use
	Initialize(ctx context.Context) error

	// Cleanup releases any resources held by the bridge
	Cleanup(ctx context.Context) error
}

// MethodInfo describes a method exposed by a bridge
type MethodInfo struct {
	// Name is the method name as it will be called from scripts
	Name string

	// Description provides documentation for the method
	Description string

	// Parameters describes the method's parameters
	Parameters []ParameterInfo

	// ReturnType describes what the method returns
	ReturnType string

	// IsAsync indicates if this method returns a promise/future
	IsAsync bool
}

// ParameterInfo describes a method parameter
type ParameterInfo struct {
	// Name is the parameter name
	Name string

	// Type is the parameter type (e.g., "string", "number", "object")
	Type string

	// Description provides documentation for the parameter
	Description string

	// Required indicates if the parameter must be provided
	Required bool

	// Default is the default value if not provided (only for optional parameters)
	Default interface{}
}

// BridgeSet manages a collection of bridges
type BridgeSet struct {
	mu      sync.RWMutex
	bridges map[string]Bridge
}

// NewBridgeSet creates a new bridge set
func NewBridgeSet() *BridgeSet {
	return &BridgeSet{
		bridges: make(map[string]Bridge),
	}
}

// Register adds a bridge to the set
func (bs *BridgeSet) Register(name string, bridge Bridge) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if _, exists := bs.bridges[name]; exists {
		return fmt.Errorf("bridge %q already registered", name)
	}

	bs.bridges[name] = bridge
	return nil
}

// Get retrieves a bridge by name
func (bs *BridgeSet) Get(name string) (Bridge, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	bridge, exists := bs.bridges[name]
	if !exists {
		return nil, fmt.Errorf("bridge %q not found", name)
	}

	return bridge, nil
}

// List returns the names of all registered bridges
func (bs *BridgeSet) List() []string {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	names := make([]string, 0, len(bs.bridges))
	for name := range bs.bridges {
		names = append(names, name)
	}

	return names
}

// Unregister removes a bridge from the set
func (bs *BridgeSet) Unregister(name string) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if _, exists := bs.bridges[name]; !exists {
		return fmt.Errorf("bridge %q not found", name)
	}

	delete(bs.bridges, name)
	return nil
}

// InitializeAll initializes all bridges in the set
func (bs *BridgeSet) InitializeAll(ctx context.Context) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	for name, bridge := range bs.bridges {
		if err := bridge.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize bridge %q: %w", name, err)
		}
	}

	return nil
}

// CleanupAll cleans up all bridges in the set
func (bs *BridgeSet) CleanupAll(ctx context.Context) error {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	var firstErr error
	for name, bridge := range bs.bridges {
		if err := bridge.Cleanup(ctx); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to cleanup bridge %q: %w", name, err)
			}
			// Continue cleaning up other bridges even if one fails
		}
	}

	return firstErr
}

// GetBridgeSet retrieves a bridge set with all bridges ready to use
func (bs *BridgeSet) GetBridgeSet() map[string]Bridge {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]Bridge, len(bs.bridges))
	for name, bridge := range bs.bridges {
		result[name] = bridge
	}

	return result
}

// Global bridge set instance
var globalBridgeSet = NewBridgeSet()

// RegisterBridge registers a bridge in the global set
func RegisterBridge(name string, bridge Bridge) error {
	return globalBridgeSet.Register(name, bridge)
}

// GetBridge retrieves a bridge from the global set
func GetBridge(name string) (Bridge, error) {
	return globalBridgeSet.Get(name)
}

// ListBridges returns the names of all bridges in the global set
func ListBridges() []string {
	return globalBridgeSet.List()
}

// UnregisterBridge removes a bridge from the global set
func UnregisterBridge(name string) error {
	return globalBridgeSet.Unregister(name)
}

// InitializeAllBridges initializes all bridges in the global set
func InitializeAllBridges(ctx context.Context) error {
	return globalBridgeSet.InitializeAll(ctx)
}

// CleanupAllBridges cleans up all bridges in the global set
func CleanupAllBridges(ctx context.Context) error {
	return globalBridgeSet.CleanupAll(ctx)
}

// GetGlobalBridgeSet retrieves the global bridge set
func GetGlobalBridgeSet() map[string]Bridge {
	return globalBridgeSet.GetBridgeSet()
}
