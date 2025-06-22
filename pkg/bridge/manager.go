// ABOUTME: Bridge manager handles lifecycle management of language-agnostic bridges.
// ABOUTME: Provides thread-safe registration, dependency resolution, and hot-reloading functionality.

// Package bridge provides centralized management for script engine bridges.
// The BridgeManager handles registration, initialization, dependency resolution,
// event tracking, metrics collection, and documentation generation for all bridges.
// It ensures proper lifecycle management and thread-safe operations across multiple
// script engines.
package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/docs"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// BridgeManager manages the lifecycle of bridges across all script engines.
// It provides centralized registration, initialization with dependency resolution,
// metrics tracking, event publishing, and documentation generation. The manager
// ensures thread-safe operations and proper cleanup of bridge resources.
type BridgeManager struct {
	mu           sync.RWMutex
	bridges      map[string]engine.Bridge
	initialized  map[string]bool
	dependencies map[string][]string // Bridge ID -> list of dependency IDs
	watchers     map[string][]chan string
	changeNotify chan string

	// Event system fields
	eventBus   *events.EventBus
	eventStore events.EventStorage
	publisher  *events.BridgeEventPublisher
	sessionID  string

	// Metrics
	metrics map[string]*BridgeMetrics
}

// BridgeMetrics tracks performance and usage metrics for a bridge.
// It records initialization times, success/failure counts, and error details
// for monitoring and debugging bridge behavior.
type BridgeMetrics struct {
	InitializationTime  time.Duration
	InitializationCount int64
	FailureCount        int64
	LastError           error
	LastInitialized     time.Time
	LastFailure         time.Time
}

// NewBridgeManager creates a new bridge manager.
// It initializes with default event bus and in-memory event storage.
// For custom event handling, use NewBridgeManagerWithEvents.
func NewBridgeManager() *BridgeManager {
	return NewBridgeManagerWithEvents(nil, nil)
}

// NewBridgeManagerWithEvents creates a new bridge manager with event system support.
// If eventBus or eventStore are nil, defaults are created. The manager generates
// a unique session ID for tracking and publishes bridge lifecycle events.
func NewBridgeManagerWithEvents(eventBus *events.EventBus, eventStore events.EventStorage) *BridgeManager {
	// Create event bus if not provided
	if eventBus == nil {
		eventBus = events.NewEventBus(events.WithBufferSize(1000))
	}

	// Create in-memory event store if not provided
	if eventStore == nil {
		eventStore = events.NewMemoryStorage()
	}

	// Generate session ID for this manager instance
	sessionID := fmt.Sprintf("bridge-manager-%d", time.Now().UnixNano())

	// Create bridge event publisher
	publisher := events.NewBridgeEventPublisher(eventBus, "bridge-manager", sessionID)

	return &BridgeManager{
		bridges:      make(map[string]engine.Bridge),
		initialized:  make(map[string]bool),
		dependencies: make(map[string][]string),
		watchers:     make(map[string][]chan string),
		changeNotify: make(chan string, 100),

		// Event system
		eventBus:   eventBus,
		eventStore: eventStore,
		publisher:  publisher,
		sessionID:  sessionID,

		// Metrics
		metrics: make(map[string]*BridgeMetrics),
	}
}

// RegisterBridge registers a bridge with the manager.
// It validates the bridge, stores its metadata and dependencies, initializes
// metrics tracking, and publishes a registration event. Returns error if the
// bridge is nil, has empty ID, or is already registered.
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

	// Initialize metrics for this bridge
	m.metrics[id] = &BridgeMetrics{}

	// Emit bridge registration event
	if m.publisher != nil {
		eventData := map[string]interface{}{
			"bridgeID":     id,
			"bridgeName":   metadata.Name,
			"version":      metadata.Version,
			"description":  metadata.Description,
			"dependencies": metadata.Dependencies,
		}
		requestID := m.publisher.PublishRequest("bridge.register", eventData)
		m.publisher.PublishResponse(requestID, map[string]interface{}{"status": "registered"}, nil, 0)
	}

	return nil
}

// InitializeBridge initializes a specific bridge.
// It checks dependencies are met, prevents concurrent initialization, tracks
// metrics, and publishes initialization events. Returns immediately if already
// initialized. Thread-safe for concurrent calls.
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

	// Emit initializing event
	if m.publisher != nil {
		eventData := map[string]interface{}{
			"bridgeID": bridgeID,
			"status":   "initializing",
		}
		requestID := m.publisher.PublishRequest("bridge.initialize", eventData)

		// Track initialization time
		startTime := time.Now()

		// Initialize the bridge outside the lock
		err := bridge.Initialize(ctx)
		duration := time.Since(startTime)

		// Update metrics
		m.mu.Lock()
		if metrics, exists := m.metrics[bridgeID]; exists {
			metrics.InitializationCount++
			metrics.InitializationTime = duration
			if err != nil {
				metrics.FailureCount++
				metrics.LastError = err
				metrics.LastFailure = time.Now()
			} else {
				metrics.LastInitialized = time.Now()
			}
		}
		m.mu.Unlock()

		if err != nil {
			// On error, mark as not initialized
			m.mu.Lock()
			m.initialized[bridgeID] = false
			m.mu.Unlock()

			// Emit failure event
			m.publisher.PublishResponse(requestID, nil, err, duration)
			return fmt.Errorf("failed to initialize bridge %s: %w", bridgeID, err)
		}

		// Emit success event
		m.publisher.PublishResponse(requestID, map[string]interface{}{
			"status":   "initialized",
			"duration": duration,
		}, nil, duration)
	} else {
		// Fallback without events
		if err := bridge.Initialize(ctx); err != nil {
			// On error, mark as not initialized
			m.mu.Lock()
			m.initialized[bridgeID] = false
			m.mu.Unlock()
			return fmt.Errorf("failed to initialize bridge %s: %w", bridgeID, err)
		}
	}

	return nil
}

// InitializeAll initializes all registered bridges.
// Bridges are initialized in registration order, not dependency order.
// For dependency-aware initialization, use InitializeWithDependencies.
// Returns on first initialization error.
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
// It calls the bridge's Cleanup method and marks it as not initialized.
// Safe to call multiple times. Returns error if bridge not found.
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
// Attempts to cleanup all bridges even if some fail. Returns the first
// error encountered but continues cleanup for remaining bridges.
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
// It performs topological sort to determine initialization order, detects
// circular dependencies, and initializes bridges from leaves to root.
// Thread-safe and idempotent.
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
// It uses depth-first search to build dependency order and detect cycles.
// Returns error if circular dependency found or dependency not registered.
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
// It also recursively reloads all bridges that depend on this bridge.
// Useful for hot-reloading configuration or recovering from errors.
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
// The callback is invoked when NotifyChange is called for the bridge.
// Watching continues until the context is cancelled. Multiple watchers
// per bridge are supported.
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
// Non-blocking - skips watchers that aren't ready to receive.
// Typically called by bridges when their configuration changes.
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
// Returns the bridge interface and nil error if found,
// or nil and error if not found. Thread-safe for concurrent access.
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
// The returned slice is a copy and safe to modify. Order is not guaranteed.
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
// Returns false if bridge not found or not initialized. Thread-safe.
func (m *BridgeManager) IsBridgeInitialized(bridgeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.initialized[bridgeID]
}

// GetBridgeMetadata retrieves metadata for a bridge.
// Returns a copy of the bridge metadata including name, version,
// description, and dependencies. Returns error if bridge not found.
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
// This is typically called when initializing a new script engine to make
// all available bridges accessible. Returns on first registration error.
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
// Useful when you want to limit which bridges are available to a particular
// engine instance. Returns error if any bridge ID not found.
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

// Event System Methods

// GetEventBus returns the event bus for external subscription.
// External systems can use this to subscribe to bridge lifecycle events.
// The event bus is shared across all bridges managed by this instance.
func (m *BridgeManager) GetEventBus() *events.EventBus {
	return m.eventBus
}

// GetEventStore returns the event store for querying bridge events.
// Provides access to historical bridge events for debugging and analysis.
// The store implementation depends on how the manager was initialized.
func (m *BridgeManager) GetEventStore() events.EventStorage {
	return m.eventStore
}

// SubscribeToBridgeEvents subscribes to bridge events with optional filtering.
// If no patterns provided, subscribes to all bridge events (bridge.*).
// Returns subscription IDs that can be used to unsubscribe later.
func (m *BridgeManager) SubscribeToBridgeEvents(handler events.EventHandlerFunc, patterns ...string) []string {
	if m.eventBus == nil {
		return nil
	}

	var subscriptionIDs []string
	if len(patterns) == 0 {
		patterns = []string{"bridge.*"}
	}

	for _, pattern := range patterns {
		subscriptionID, err := m.eventBus.SubscribePattern(pattern, handler)
		if err == nil {
			subscriptionIDs = append(subscriptionIDs, subscriptionID)
		}
	}

	return subscriptionIDs
}

// UnsubscribeFromBridgeEvents unsubscribes from bridge events.
// Pass the subscription IDs returned from SubscribeToBridgeEvents.
// Safe to call with nil or empty slice.
func (m *BridgeManager) UnsubscribeFromBridgeEvents(subscriptionIDs []string) {
	if m.eventBus == nil {
		return
	}

	for _, subscriptionID := range subscriptionIDs {
		m.eventBus.Unsubscribe(subscriptionID)
	}
}

// Metrics and Monitoring Methods

// GetBridgeMetrics returns metrics for a specific bridge.
// Returns a copy of metrics to prevent race conditions.
// Includes initialization times, counts, and error information.
func (m *BridgeManager) GetBridgeMetrics(bridgeID string) (*BridgeMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics, exists := m.metrics[bridgeID]; exists {
		// Return a copy to prevent race conditions
		metricsCopy := *metrics
		return &metricsCopy, nil
	}

	return nil, fmt.Errorf("bridge %s not found", bridgeID)
}

// GetAllBridgeMetrics returns metrics for all bridges.
// Returns copies of all metrics indexed by bridge ID.
// Useful for monitoring and reporting on bridge health.
func (m *BridgeManager) GetAllBridgeMetrics() map[string]*BridgeMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*BridgeMetrics)
	for bridgeID, metrics := range m.metrics {
		// Return copies to prevent race conditions
		metricsCopy := *metrics
		result[bridgeID] = &metricsCopy
	}

	return result
}

// GenerateBridgeReport generates a comprehensive report of all bridge activity.
// Includes session info, bridge counts, initialization status, and detailed
// metrics for each bridge. Useful for debugging and operational monitoring.
func (m *BridgeManager) GenerateBridgeReport() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := map[string]interface{}{
		"sessionID":     m.sessionID,
		"totalBridges":  len(m.bridges),
		"initialized":   0,
		"failed":        0,
		"bridgeDetails": make(map[string]interface{}),
	}

	initializedCount := 0
	failedCount := 0
	bridgeDetails := make(map[string]interface{})

	for bridgeID, bridge := range m.bridges {
		isInitialized := m.initialized[bridgeID]
		if isInitialized {
			initializedCount++
		}

		metrics := m.metrics[bridgeID]
		if metrics != nil && metrics.FailureCount > 0 {
			failedCount++
		}

		metadata := bridge.GetMetadata()
		bridgeDetails[bridgeID] = map[string]interface{}{
			"name":                metadata.Name,
			"version":             metadata.Version,
			"description":         metadata.Description,
			"dependencies":        metadata.Dependencies,
			"initialized":         isInitialized,
			"initializationCount": metrics.InitializationCount,
			"failureCount":        metrics.FailureCount,
			"lastInitialized":     metrics.LastInitialized,
			"lastFailure":         metrics.LastFailure,
			"initializationTime":  metrics.InitializationTime,
		}
	}

	report["initialized"] = initializedCount
	report["failed"] = failedCount
	report["bridgeDetails"] = bridgeDetails

	return report
}

// Performance Profiling Methods

// StartProfiling enables performance profiling for bridge operations.
// Currently collects basic metrics. Can be extended for more detailed
// profiling such as memory usage and goroutine tracking.
func (m *BridgeManager) StartProfiling() {
	// This method can be extended to add more detailed profiling
	// For now, we're already collecting basic metrics in the existing methods
}

// StopProfiling disables performance profiling.
// Placeholder for future profiling control. Basic metrics collection
// continues regardless of profiling state.
func (m *BridgeManager) StopProfiling() {
	// This method can be extended to stop detailed profiling
}

// Documentation Generation Methods

// BridgeDocumentable implements docs.Documentable for bridges.
// It wraps bridge metadata and methods to provide documentation
// generation capabilities through the go-llms docs system.
type BridgeDocumentable struct {
	ID           string
	Name         string
	Version      string
	Description  string
	Methods      []engine.MethodInfo
	TypeMappings map[string]engine.TypeMapping
	Permissions  []engine.Permission
	Dependencies []string
}

// GetDocumentation returns the documentation for this bridge.
// Implements docs.Documentable interface. Generates documentation
// including methods, examples, dependencies, and permissions.
func (bd *BridgeDocumentable) GetDocumentation() docs.Documentation {
	// Create examples from methods
	examples := make([]docs.Example, 0, len(bd.Methods))
	for _, method := range bd.Methods {
		examples = append(examples, docs.Example{
			Name:        method.Name,
			Description: method.Description,
			Code:        fmt.Sprintf("bridge.%s()", method.Name),
			Language:    "javascript",
		})
	}

	return docs.Documentation{
		Name:        bd.Name,
		Description: bd.Description,
		Category:    "bridge",
		Version:     bd.Version,
		Examples:    examples,
		Metadata: map[string]interface{}{
			"id":           bd.ID,
			"dependencies": bd.Dependencies,
			"permissions":  bd.Permissions,
			"methods":      bd.Methods,
			"typeMappings": bd.TypeMappings,
		},
	}
}

// GenerateDocumentation generates comprehensive documentation for all bridges.
// Supports multiple formats: openapi, markdown, json. Creates appropriate
// generator based on format and returns format-specific documentation.
func (m *BridgeManager) GenerateDocumentation(ctx context.Context, format string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create documentable items for each bridge
	var documentableItems []docs.Documentable

	for bridgeID, bridge := range m.bridges {
		metadata := bridge.GetMetadata()
		methods := bridge.Methods()
		typeMappings := bridge.TypeMappings()
		permissions := bridge.RequiredPermissions()

		bridgeDoc := &BridgeDocumentable{
			ID:           bridgeID,
			Name:         metadata.Name,
			Version:      metadata.Version,
			Description:  metadata.Description,
			Methods:      methods,
			TypeMappings: typeMappings,
			Permissions:  permissions,
			Dependencies: metadata.Dependencies,
		}

		documentableItems = append(documentableItems, bridgeDoc)
	}

	// Create appropriate generator for the format
	config := docs.GeneratorConfig{
		Title:       "Bridge Documentation",
		Version:     "1.0.0",
		Description: "Documentation for all registered bridges",
	}

	// Generate documentation based on format
	switch format {
	case "openapi":
		generator := docs.NewOpenAPIGenerator(config)
		return generator.GenerateOpenAPI(ctx, documentableItems)
	case "markdown":
		generator := docs.NewMarkdownGenerator(config)
		return generator.GenerateMarkdown(ctx, documentableItems)
	case "json":
		generator := docs.NewMarkdownGenerator(config) // Use markdown generator for JSON too
		return generator.GenerateJSON(ctx, documentableItems)
	default:
		return nil, fmt.Errorf("unsupported documentation format: %s", format)
	}
}

// GenerateOpenAPIDocumentation generates OpenAPI specification for all bridges.
// Returns a complete OpenAPI spec that can be used for API documentation,
// client generation, or integration with API gateways.
func (m *BridgeManager) GenerateOpenAPIDocumentation(ctx context.Context) (*docs.OpenAPISpec, error) {
	result, err := m.GenerateDocumentation(ctx, "openapi")
	if err != nil {
		return nil, err
	}

	spec, ok := result.(*docs.OpenAPISpec)
	if !ok {
		return nil, fmt.Errorf("failed to generate OpenAPI specification")
	}

	return spec, nil
}

// GenerateMarkdownDocumentation generates Markdown documentation for all bridges.
// Returns formatted Markdown suitable for documentation sites, README files,
// or integration with documentation systems.
func (m *BridgeManager) GenerateMarkdownDocumentation(ctx context.Context) (string, error) {
	result, err := m.GenerateDocumentation(ctx, "markdown")
	if err != nil {
		return "", err
	}

	markdown, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("failed to generate Markdown documentation")
	}

	return markdown, nil
}

// GenerateJSONDocumentation generates JSON documentation for all bridges.
// Returns structured JSON that can be processed by documentation tools
// or used for programmatic documentation access.
func (m *BridgeManager) GenerateJSONDocumentation(ctx context.Context) ([]byte, error) {
	result, err := m.GenerateDocumentation(ctx, "json")
	if err != nil {
		return nil, err
	}

	jsonData, ok := result.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to generate JSON documentation")
	}

	return jsonData, nil
}

// GenerateBridgeDocumentation generates documentation for a specific bridge.
// Supports same formats as GenerateDocumentation but for a single bridge.
// Useful for bridge-specific documentation pages or debugging.
func (m *BridgeManager) GenerateBridgeDocumentation(ctx context.Context, bridgeID string, format string) (interface{}, error) {
	m.mu.RLock()
	bridge, exists := m.bridges[bridgeID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("bridge %s not found", bridgeID)
	}

	metadata := bridge.GetMetadata()
	methods := bridge.Methods()
	typeMappings := bridge.TypeMappings()
	permissions := bridge.RequiredPermissions()

	bridgeDoc := &BridgeDocumentable{
		ID:           bridgeID,
		Name:         metadata.Name,
		Version:      metadata.Version,
		Description:  metadata.Description,
		Methods:      methods,
		TypeMappings: typeMappings,
		Permissions:  permissions,
		Dependencies: metadata.Dependencies,
	}

	// Create appropriate generator for the format
	config := docs.GeneratorConfig{
		Title:       bridgeDoc.Name + " Documentation",
		Version:     bridgeDoc.Version,
		Description: "Documentation for " + bridgeDoc.Name + " bridge",
	}

	// Generate documentation for single bridge
	switch format {
	case "openapi":
		generator := docs.NewOpenAPIGenerator(config)
		return generator.GenerateOpenAPI(ctx, []docs.Documentable{bridgeDoc})
	case "markdown":
		generator := docs.NewMarkdownGenerator(config)
		return generator.GenerateMarkdown(ctx, []docs.Documentable{bridgeDoc})
	case "json":
		generator := docs.NewMarkdownGenerator(config)
		return generator.GenerateJSON(ctx, []docs.Documentable{bridgeDoc})
	default:
		return nil, fmt.Errorf("unsupported documentation format: %s", format)
	}
}

// ExportAPISchema exports the API schema for all bridges with type mappings.
// Returns a comprehensive schema including methods, types, permissions, and
// current runtime state. Useful for tooling and IDE integration.
func (m *BridgeManager) ExportAPISchema() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schema := map[string]interface{}{
		"version":   "1.0.0",
		"bridges":   make(map[string]interface{}),
		"types":     make(map[string]interface{}),
		"sessionID": m.sessionID,
		"generated": time.Now().UTC(),
	}

	bridges := make(map[string]interface{})
	allTypes := make(map[string]interface{})

	for bridgeID, bridge := range m.bridges {
		metadata := bridge.GetMetadata()
		methods := bridge.Methods()
		typeMappings := bridge.TypeMappings()
		permissions := bridge.RequiredPermissions()

		bridgeSchema := map[string]interface{}{
			"id":           bridgeID,
			"name":         metadata.Name,
			"version":      metadata.Version,
			"description":  metadata.Description,
			"dependencies": metadata.Dependencies,
			"methods":      methods,
			"permissions":  permissions,
			"initialized":  m.initialized[bridgeID],
		}

		// Add metrics if available
		if metrics, exists := m.metrics[bridgeID]; exists {
			bridgeSchema["metrics"] = map[string]interface{}{
				"initializationCount": metrics.InitializationCount,
				"failureCount":        metrics.FailureCount,
				"lastInitialized":     metrics.LastInitialized,
				"initializationTime":  metrics.InitializationTime,
			}
		}

		bridges[bridgeID] = bridgeSchema

		// Collect type mappings
		for typeName, typeMapping := range typeMappings {
			allTypes[fmt.Sprintf("%s.%s", bridgeID, typeName)] = map[string]interface{}{
				"bridge":     bridgeID,
				"name":       typeName,
				"goType":     typeMapping.GoType,
				"scriptType": typeMapping.ScriptType,
				"converter":  typeMapping.Converter != "",
				"metadata":   typeMapping.Metadata,
			}
		}
	}

	schema["bridges"] = bridges
	schema["types"] = allTypes

	return schema
}

// Bridge State Serialization

// SerializableBridgeState represents the state of the bridge manager in serializable format.
// Used for state persistence, backup/restore, and cluster synchronization.
// Does not include bridge instances, only metadata and runtime state.
type SerializableBridgeState struct {
	Version      string                               `json:"version"`
	SessionID    string                               `json:"session_id"`
	Timestamp    time.Time                            `json:"timestamp"`
	Bridges      map[string]SerializableBridgeInfo    `json:"bridges"`
	Initialized  map[string]bool                      `json:"initialized"`
	Dependencies map[string][]string                  `json:"dependencies"`
	Metrics      map[string]SerializableBridgeMetrics `json:"metrics"`
	Metadata     map[string]interface{}               `json:"metadata,omitempty"`
}

// SerializableBridgeInfo represents bridge information in serializable format.
// Contains only metadata that can be safely serialized and restored.
// Bridge instances must be re-registered after state import.
type SerializableBridgeInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SerializableBridgeMetrics represents bridge metrics in serializable format.
// Converts time.Duration to string for JSON compatibility. Preserves all
// metric data for accurate state restoration.
type SerializableBridgeMetrics struct {
	InitializationTime  string    `json:"initialization_time"`
	InitializationCount int64     `json:"initialization_count"`
	FailureCount        int64     `json:"failure_count"`
	LastError           string    `json:"last_error,omitempty"`
	LastInitialized     time.Time `json:"last_initialized"`
	LastFailure         time.Time `json:"last_failure"`
}

// ExportState exports the current state of the bridge manager.
// Returns a serializable representation of all bridge metadata, initialization
// state, dependencies, and metrics. Bridge instances are not included.
func (m *BridgeManager) ExportState() (*SerializableBridgeState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := &SerializableBridgeState{
		Version:      "1.0",
		SessionID:    m.sessionID,
		Timestamp:    time.Now().UTC(),
		Bridges:      make(map[string]SerializableBridgeInfo),
		Initialized:  make(map[string]bool),
		Dependencies: make(map[string][]string),
		Metrics:      make(map[string]SerializableBridgeMetrics),
		Metadata:     make(map[string]interface{}),
	}

	// Export bridge information (metadata only, not the actual bridge instances)
	for bridgeID, bridge := range m.bridges {
		metadata := bridge.GetMetadata()
		state.Bridges[bridgeID] = SerializableBridgeInfo{
			ID:           bridgeID,
			Name:         metadata.Name,
			Version:      metadata.Version,
			Description:  metadata.Description,
			Dependencies: metadata.Dependencies,
			Metadata:     map[string]interface{}{"type": fmt.Sprintf("%T", bridge)},
		}
	}

	// Export initialization state
	for bridgeID, initialized := range m.initialized {
		state.Initialized[bridgeID] = initialized
	}

	// Export dependencies
	for bridgeID, deps := range m.dependencies {
		state.Dependencies[bridgeID] = make([]string, len(deps))
		copy(state.Dependencies[bridgeID], deps)
	}

	// Export metrics
	for bridgeID, metrics := range m.metrics {
		serMetrics := SerializableBridgeMetrics{
			InitializationTime:  metrics.InitializationTime.String(),
			InitializationCount: metrics.InitializationCount,
			FailureCount:        metrics.FailureCount,
			LastInitialized:     metrics.LastInitialized,
			LastFailure:         metrics.LastFailure,
		}
		if metrics.LastError != nil {
			serMetrics.LastError = metrics.LastError.Error()
		}
		state.Metrics[bridgeID] = serMetrics
	}

	// Add additional metadata
	state.Metadata["total_bridges"] = len(m.bridges)
	state.Metadata["initialized_count"] = len(m.initialized)
	state.Metadata["export_time"] = time.Now().UTC()

	return state, nil
}

// ImportState imports bridge manager state from serializable format.
// Validates version compatibility and state integrity before importing.
// Only imports metadata and state; bridges must be re-registered separately.
func (m *BridgeManager) ImportState(state *SerializableBridgeState) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	// Validate state version compatibility
	if err := m.validateStateVersion(state.Version); err != nil {
		return fmt.Errorf("state version validation failed: %w", err)
	}

	// Validate state integrity
	if err := m.validateStateIntegrity(state); err != nil {
		return fmt.Errorf("state integrity validation failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Import initialization state
	m.initialized = make(map[string]bool)
	for bridgeID, initialized := range state.Initialized {
		m.initialized[bridgeID] = initialized
	}

	// Import dependencies
	m.dependencies = make(map[string][]string)
	for bridgeID, deps := range state.Dependencies {
		m.dependencies[bridgeID] = make([]string, len(deps))
		copy(m.dependencies[bridgeID], deps)
	}

	// Import metrics
	m.metrics = make(map[string]*BridgeMetrics)
	for bridgeID, serMetrics := range state.Metrics {
		metrics := &BridgeMetrics{
			InitializationCount: serMetrics.InitializationCount,
			FailureCount:        serMetrics.FailureCount,
			LastInitialized:     serMetrics.LastInitialized,
			LastFailure:         serMetrics.LastFailure,
		}

		// Parse initialization time
		if serMetrics.InitializationTime != "" {
			if duration, err := time.ParseDuration(serMetrics.InitializationTime); err == nil {
				metrics.InitializationTime = duration
			}
		}

		// Parse last error
		if serMetrics.LastError != "" {
			metrics.LastError = fmt.Errorf("%s", serMetrics.LastError)
		}

		m.metrics[bridgeID] = metrics
	}

	// Note: We don't import actual bridge instances as they need to be re-registered
	// This is intentional as bridges contain runtime state and functions that can't be serialized

	return nil
}

// ExportStateToJSON exports state as JSON bytes.
// If pretty is true, formats with indentation for readability.
// Useful for state persistence or debugging.
func (m *BridgeManager) ExportStateToJSON(pretty bool) ([]byte, error) {
	state, err := m.ExportState()
	if err != nil {
		return nil, fmt.Errorf("failed to export state: %w", err)
	}

	if pretty {
		return json.MarshalIndent(state, "", "  ")
	}
	return json.Marshal(state)
}

// ImportStateFromJSON imports state from JSON bytes.
// Unmarshals JSON and validates before importing. Returns error
// if JSON invalid or state validation fails.
func (m *BridgeManager) ImportStateFromJSON(data []byte) error {
	var state SerializableBridgeState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return m.ImportState(&state)
}

// UpdateStateIncremental performs an incremental state update.
// Updates specific fields for a bridge without full state replacement.
// Currently supports metrics and initialization state updates.
func (m *BridgeManager) UpdateStateIncremental(bridgeID string, updates map[string]interface{}) error {
	if bridgeID == "" {
		return fmt.Errorf("bridge ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if bridge exists
	if _, exists := m.bridges[bridgeID]; !exists {
		return fmt.Errorf("bridge %s not found", bridgeID)
	}

	// Apply updates to metrics if provided
	if metricsUpdate, ok := updates["metrics"]; ok {
		if metricsMap, ok := metricsUpdate.(map[string]interface{}); ok {
			metrics, exists := m.metrics[bridgeID]
			if !exists {
				metrics = &BridgeMetrics{}
				m.metrics[bridgeID] = metrics
			}

			// Update individual metrics fields
			if initCount, ok := metricsMap["initialization_count"].(int64); ok {
				metrics.InitializationCount = initCount
			}
			if failCount, ok := metricsMap["failure_count"].(int64); ok {
				metrics.FailureCount = failCount
			}
			if lastInit, ok := metricsMap["last_initialized"].(time.Time); ok {
				metrics.LastInitialized = lastInit
			}
			if lastFail, ok := metricsMap["last_failure"].(time.Time); ok {
				metrics.LastFailure = lastFail
			}
		}
	}

	// Apply initialization state update
	if initialized, ok := updates["initialized"].(bool); ok {
		m.initialized[bridgeID] = initialized
	}

	return nil
}

// validateStateVersion checks if the state version is compatible.
// Ensures imported state can be correctly interpreted. Currently
// only supports version 1.0.
func (m *BridgeManager) validateStateVersion(version string) error {
	switch version {
	case "1.0":
		return nil
	default:
		return fmt.Errorf("unsupported state version: %s", version)
	}
}

// validateStateIntegrity performs basic integrity checks on the state.
// Ensures all references are valid and state is internally consistent.
// Checks dependencies, metrics, and initialization state references.
func (m *BridgeManager) validateStateIntegrity(state *SerializableBridgeState) error {
	if state.SessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if state.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	// Validate that all bridges referenced in initialized map exist in bridges map
	for bridgeID := range state.Initialized {
		if _, exists := state.Bridges[bridgeID]; !exists {
			return fmt.Errorf("bridge %s is in initialized map but not in bridges map", bridgeID)
		}
	}

	// Validate that all bridges referenced in metrics exist in bridges map
	for bridgeID := range state.Metrics {
		if _, exists := state.Bridges[bridgeID]; !exists {
			return fmt.Errorf("bridge %s is in metrics map but not in bridges map", bridgeID)
		}
	}

	// Validate dependencies reference existing bridges
	for bridgeID, deps := range state.Dependencies {
		if _, exists := state.Bridges[bridgeID]; !exists {
			return fmt.Errorf("bridge %s has dependencies but is not in bridges map", bridgeID)
		}
		for _, dep := range deps {
			if _, exists := state.Bridges[dep]; !exists {
				return fmt.Errorf("bridge %s depends on %s but dependency not found in bridges map", bridgeID, dep)
			}
		}
	}

	return nil
}

// GetStateVersion returns the current state format version.
// Used for versioning exported state. Current version is 1.0.
func (m *BridgeManager) GetStateVersion() string {
	return "1.0"
}

// Cleanup method to properly close event system resources.
// Should be called when shutting down the bridge manager to ensure
// proper cleanup of event bus and storage resources.
func (m *BridgeManager) Cleanup() error {
	if m.eventBus != nil {
		m.eventBus.Close()
	}

	if m.eventStore != nil {
		return m.eventStore.Close()
	}

	return nil
}
