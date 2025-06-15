// ABOUTME: Bridge manager handles lifecycle management of language-agnostic bridges.
// ABOUTME: Provides thread-safe registration, dependency resolution, and hot-reloading functionality.

package bridge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/events"
	"github.com/lexlapax/go-llms/pkg/docs"
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

	// Event system fields
	eventBus   *events.EventBus
	eventStore events.EventStorage
	publisher  *events.BridgeEventPublisher
	sessionID  string

	// Metrics
	metrics map[string]*BridgeMetrics
}

// BridgeMetrics tracks performance and usage metrics for a bridge
type BridgeMetrics struct {
	InitializationTime  time.Duration
	InitializationCount int64
	FailureCount        int64
	LastError           error
	LastInitialized     time.Time
	LastFailure         time.Time
}

// NewBridgeManager creates a new bridge manager.
func NewBridgeManager() *BridgeManager {
	return NewBridgeManagerWithEvents(nil, nil)
}

// NewBridgeManagerWithEvents creates a new bridge manager with event system support.
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

// Event System Methods

// GetEventBus returns the event bus for external subscription
func (m *BridgeManager) GetEventBus() *events.EventBus {
	return m.eventBus
}

// GetEventStore returns the event store for querying bridge events
func (m *BridgeManager) GetEventStore() events.EventStorage {
	return m.eventStore
}

// SubscribeToBridgeEvents subscribes to bridge events with optional filtering
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

// UnsubscribeFromBridgeEvents unsubscribes from bridge events
func (m *BridgeManager) UnsubscribeFromBridgeEvents(subscriptionIDs []string) {
	if m.eventBus == nil {
		return
	}

	for _, subscriptionID := range subscriptionIDs {
		m.eventBus.Unsubscribe(subscriptionID)
	}
}

// Metrics and Monitoring Methods

// GetBridgeMetrics returns metrics for a specific bridge
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

// GetAllBridgeMetrics returns metrics for all bridges
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

// GenerateBridgeReport generates a comprehensive report of all bridge activity
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

// StartProfiling enables performance profiling for bridge operations
func (m *BridgeManager) StartProfiling() {
	// This method can be extended to add more detailed profiling
	// For now, we're already collecting basic metrics in the existing methods
}

// StopProfiling disables performance profiling
func (m *BridgeManager) StopProfiling() {
	// This method can be extended to stop detailed profiling
}

// Documentation Generation Methods

// BridgeDocumentable implements docs.Documentable for bridges
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

// GetDocumentation returns the documentation for this bridge
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

// GenerateDocumentation generates comprehensive documentation for all bridges
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

// GenerateOpenAPIDocumentation generates OpenAPI specification for all bridges
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

// GenerateMarkdownDocumentation generates Markdown documentation for all bridges
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

// GenerateJSONDocumentation generates JSON documentation for all bridges
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

// GenerateBridgeDocumentation generates documentation for a specific bridge
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

// ExportAPISchema exports the API schema for all bridges with type mappings
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

// Cleanup method to properly close event system resources
func (m *BridgeManager) Cleanup() error {
	if m.eventBus != nil {
		m.eventBus.Close()
	}

	if m.eventStore != nil {
		return m.eventStore.Close()
	}

	return nil
}
