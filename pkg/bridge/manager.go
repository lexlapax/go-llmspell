// ABOUTME: Bridge manager handles lifecycle management of language-agnostic bridges.
// ABOUTME: Provides thread-safe registration, dependency resolution, and hot-reloading functionality.

package bridge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// BridgeManager manages the lifecycle of bridges across all script engines.
type BridgeManager struct {
	mu           sync.RWMutex
	bridges      map[string]engine.Bridge
	initialized  map[string]bool
	dependencies map[string][]string // Bridge ID -> list of dependency IDs
	watchers     map[string][]chan string
	changeNotify chan string
}

// NewBridgeManager creates a new bridge manager.
func NewBridgeManager() *BridgeManager {
	return &BridgeManager{
		bridges:      make(map[string]engine.Bridge),
		initialized:  make(map[string]bool),
		dependencies: make(map[string][]string),
		watchers:     make(map[string][]chan string),
		changeNotify: make(chan string, 100),
	}
}

// RegisterBridge registers a bridge with the manager.
func (m *BridgeManager) RegisterBridge(bridge engine.Bridge) error {
	if bridge == nil {
		return fmt.Errorf("cannot register nil bridge")
	}

	id := bridge.GetID()
	if id == "" {
		return fmt.Errorf("cannot register bridge with empty bridge ID")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.bridges[id]; exists {
		return fmt.Errorf("bridge %s already registered", id)
	}

	m.bridges[id] = bridge
	m.initialized[id] = false

	// Store dependencies
	metadata := bridge.GetMetadata()
	if len(metadata.Dependencies) > 0 {
		m.dependencies[id] = metadata.Dependencies
	}

	return nil
}

// InitializeBridge initializes a specific bridge.
func (m *BridgeManager) InitializeBridge(ctx context.Context, bridgeID string) error {
	m.mu.Lock()
	bridge, exists := m.bridges[bridgeID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("bridge %s not found", bridgeID)
	}

	// Check if already initialized
	if m.initialized[bridgeID] {
		m.mu.Unlock()
		return nil
	}

	// Mark as initializing to prevent concurrent initialization
	m.initialized[bridgeID] = true
	m.mu.Unlock()

	// Initialize the bridge outside the lock
	if err := bridge.Initialize(ctx); err != nil {
		// On error, mark as not initialized
		m.mu.Lock()
		m.initialized[bridgeID] = false
		m.mu.Unlock()
		return fmt.Errorf("failed to initialize bridge %s: %w", bridgeID, err)
	}

	return nil
}

// InitializeAll initializes all registered bridges.
func (m *BridgeManager) InitializeAll(ctx context.Context) error {
	m.mu.RLock()
	bridgeIDs := make([]string, 0, len(m.bridges))
	for id := range m.bridges {
		bridgeIDs = append(bridgeIDs, id)
	}
	m.mu.RUnlock()

	for _, id := range bridgeIDs {
		if err := m.InitializeBridge(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

// CleanupBridge cleans up a specific bridge.
func (m *BridgeManager) CleanupBridge(ctx context.Context, bridgeID string) error {
	m.mu.Lock()
	bridge, exists := m.bridges[bridgeID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("bridge %s not found", bridgeID)
	}
	m.mu.Unlock()

	// Cleanup the bridge
	if err := bridge.Cleanup(ctx); err != nil {
		return fmt.Errorf("failed to cleanup bridge %s: %w", bridgeID, err)
	}

	m.mu.Lock()
	m.initialized[bridgeID] = false
	m.mu.Unlock()

	return nil
}

// CleanupAll cleans up all registered bridges.
func (m *BridgeManager) CleanupAll(ctx context.Context) error {
	m.mu.RLock()
	bridgeIDs := make([]string, 0, len(m.bridges))
	for id := range m.bridges {
		bridgeIDs = append(bridgeIDs, id)
	}
	m.mu.RUnlock()

	var firstErr error
	for _, id := range bridgeIDs {
		if err := m.CleanupBridge(ctx, id); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// InitializeWithDependencies initializes a bridge and all its dependencies.
func (m *BridgeManager) InitializeWithDependencies(ctx context.Context, bridgeID string) error {
	// Build dependency graph and check for cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	order := make([]string, 0)

	if err := m.resolveDependencies(bridgeID, visited, recStack, &order); err != nil {
		return err
	}

	// Initialize in dependency order
	for i := len(order) - 1; i >= 0; i-- {
		if err := m.InitializeBridge(ctx, order[i]); err != nil {
			return err
		}
	}

	return nil
}

// resolveDependencies performs topological sort with cycle detection.
func (m *BridgeManager) resolveDependencies(bridgeID string, visited, recStack map[string]bool, order *[]string) error {
	visited[bridgeID] = true
	recStack[bridgeID] = true

	m.mu.RLock()
	deps := m.dependencies[bridgeID]
	_, exists := m.bridges[bridgeID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("dependency not found: %s", bridgeID)
	}

	for _, dep := range deps {
		if !visited[dep] {
			if err := m.resolveDependencies(dep, visited, recStack, order); err != nil {
				return err
			}
		} else if recStack[dep] {
			return fmt.Errorf("circular dependency detected: %s -> %s", bridgeID, dep)
		}
	}

	recStack[bridgeID] = false
	*order = append(*order, bridgeID)
	return nil
}

// ReloadBridge reloads a bridge by cleaning it up and reinitializing.
func (m *BridgeManager) ReloadBridge(ctx context.Context, bridgeID string) error {
	// Check if bridge exists
	m.mu.RLock()
	_, exists := m.bridges[bridgeID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("bridge %s not found", bridgeID)
	}

	// Cleanup if initialized
	if m.IsBridgeInitialized(bridgeID) {
		if err := m.CleanupBridge(ctx, bridgeID); err != nil {
			return err
		}
	}

	// Reinitialize
	if err := m.InitializeBridge(ctx, bridgeID); err != nil {
		return err
	}

	// Reload dependent bridges
	m.mu.RLock()
	dependentBridges := make([]string, 0)
	for id, deps := range m.dependencies {
		for _, dep := range deps {
			if dep == bridgeID && id != bridgeID {
				dependentBridges = append(dependentBridges, id)
				break
			}
		}
	}
	m.mu.RUnlock()

	// Reload dependents
	for _, dependent := range dependentBridges {
		if err := m.ReloadBridge(ctx, dependent); err != nil {
			return err
		}
	}

	return nil
}

// WatchBridge starts watching a bridge for changes.
func (m *BridgeManager) WatchBridge(ctx context.Context, bridgeID string, interval time.Duration, callback func(string)) error {
	m.mu.RLock()
	_, exists := m.bridges[bridgeID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("bridge %s not found", bridgeID)
	}

	notifyChan := make(chan string, 1)

	m.mu.Lock()
	m.watchers[bridgeID] = append(m.watchers[bridgeID], notifyChan)
	m.mu.Unlock()

	go func() {
		for {
			select {
			case <-ctx.Done():
				// Remove watcher
				m.mu.Lock()
				watchers := m.watchers[bridgeID]
				for i, w := range watchers {
					if w == notifyChan {
						m.watchers[bridgeID] = append(watchers[:i], watchers[i+1:]...)
						break
					}
				}
				m.mu.Unlock()
				close(notifyChan)
				return
			case id := <-notifyChan:
				callback(id)
			}
		}
	}()

	return nil
}

// NotifyChange notifies watchers of a bridge change.
func (m *BridgeManager) NotifyChange(bridgeID string) {
	m.mu.RLock()
	watchers := m.watchers[bridgeID]
	m.mu.RUnlock()

	for _, watcher := range watchers {
		select {
		case watcher <- bridgeID:
		default:
			// Don't block if watcher is not ready
		}
	}
}

// GetBridge retrieves a bridge by ID.
func (m *BridgeManager) GetBridge(bridgeID string) (engine.Bridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bridge, exists := m.bridges[bridgeID]
	if !exists {
		return nil, fmt.Errorf("bridge %s not found", bridgeID)
	}

	return bridge, nil
}

// ListBridges returns a list of all registered bridge IDs.
func (m *BridgeManager) ListBridges() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.bridges))
	for id := range m.bridges {
		ids = append(ids, id)
	}

	return ids
}

// IsBridgeInitialized checks if a bridge is initialized.
func (m *BridgeManager) IsBridgeInitialized(bridgeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.initialized[bridgeID]
}

// GetBridgeMetadata retrieves metadata for a bridge.
func (m *BridgeManager) GetBridgeMetadata(bridgeID string) (engine.BridgeMetadata, error) {
	m.mu.RLock()
	bridge, exists := m.bridges[bridgeID]
	m.mu.RUnlock()

	if !exists {
		return engine.BridgeMetadata{}, fmt.Errorf("bridge %s not found", bridgeID)
	}

	return bridge.GetMetadata(), nil
}

// RegisterBridgesWithEngine registers all bridges with a script engine.
func (m *BridgeManager) RegisterBridgesWithEngine(scriptEngine engine.ScriptEngine) error {
	m.mu.RLock()
	bridges := make([]engine.Bridge, 0, len(m.bridges))
	for _, bridge := range m.bridges {
		bridges = append(bridges, bridge)
	}
	m.mu.RUnlock()

	for _, bridge := range bridges {
		if err := scriptEngine.RegisterBridge(bridge); err != nil {
			return fmt.Errorf("failed to register bridge %s with engine: %w", bridge.GetID(), err)
		}
	}

	return nil
}

// RegisterSpecificBridgesWithEngine registers specific bridges with a script engine.
func (m *BridgeManager) RegisterSpecificBridgesWithEngine(scriptEngine engine.ScriptEngine, bridgeIDs []string) error {
	for _, id := range bridgeIDs {
		m.mu.RLock()
		bridge, exists := m.bridges[id]
		m.mu.RUnlock()

		if !exists {
			return fmt.Errorf("bridge %s not found", id)
		}

		if err := scriptEngine.RegisterBridge(bridge); err != nil {
			return fmt.Errorf("failed to register bridge %s with engine: %w", id, err)
		}
	}

	return nil
}
