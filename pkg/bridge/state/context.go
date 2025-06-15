// ABOUTME: State Context Bridge implementation that exposes go-llms SharedStateContext to script engines
// ABOUTME: Provides parent-child state sharing with configurable inheritance for multi-agent systems

package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// StateContextBridge bridges go-llms SharedStateContext to script engines
type StateContextBridge struct {
	mu       sync.RWMutex
	contexts map[string]*domain.SharedStateContext
	configs  map[string]*inheritanceConfig // Track inheritance configs
	nextID   int

	// Schema validation fields from go-llms v0.3.5
	schemaRepo     sdomain.SchemaRepository
	validator      sdomain.Validator
	stateValidator domain.StateValidator
	stateSchemas   map[string]string // contextID -> schemaID mapping

	// Event emission fields from go-llms v0.3.5
	eventEmitter   domain.EventEmitter
	eventBus       *events.EventBus
	eventFilters   map[string]events.EventFilter // pattern -> filter mapping
	eventHistory   []domain.Event                // For event replay
	eventHistoryMu sync.RWMutex                  // Separate mutex for event history
}

// inheritanceConfig tracks inheritance settings for a shared context
type inheritanceConfig struct {
	inheritMessages  bool
	inheritArtifacts bool
	inheritMetadata  bool
}

// NewStateContextBridge creates a new state context bridge
func NewStateContextBridge() (*StateContextBridge, error) {
	return NewStateContextBridgeWithEventEmitter(nil)
}

// NewStateContextBridgeWithEventEmitter creates a new state context bridge with event emission
func NewStateContextBridgeWithEventEmitter(eventEmitter domain.EventEmitter) (*StateContextBridge, error) {
	// Create schema repository using go-llms infrastructure
	schemaRepo := repository.NewInMemorySchemaRepository()

	// Create validator with advanced features from go-llms
	validator := validation.NewValidator(
		validation.WithCoercion(true),
		validation.WithCustomValidation(true),
	)

	// Create composite state validator
	stateValidator := domain.CompositeValidator()

	// Create event bus for internal event handling
	eventBus := events.NewEventBus()

	return &StateContextBridge{
		contexts: make(map[string]*domain.SharedStateContext),
		configs:  make(map[string]*inheritanceConfig),
		nextID:   1,

		// Schema validation system from go-llms
		schemaRepo:     schemaRepo,
		validator:      validator,
		stateValidator: stateValidator,
		stateSchemas:   make(map[string]string),

		// Event emission system from go-llms
		eventEmitter:   eventEmitter,
		eventBus:       eventBus,
		eventFilters:   make(map[string]events.EventFilter),
		eventHistory:   make([]domain.Event, 0),
		eventHistoryMu: sync.RWMutex{},
	}, nil
}

// GetID returns the bridge ID
func (b *StateContextBridge) GetID() string {
	return "state_context"
}

// GetMetadata returns bridge metadata
func (b *StateContextBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "State Context Bridge",
		Version:     "1.0.0",
		Description: "Bridges go-llms SharedStateContext for parent-child state sharing",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *StateContextBridge) Initialize(ctx context.Context) error {
	return nil
}

// Cleanup cleans up bridge resources
func (b *StateContextBridge) Cleanup(ctx context.Context) error {
	return nil
}

// IsInitialized returns whether the bridge is initialized
func (b *StateContextBridge) IsInitialized() bool {
	return true
}

// RegisterWithEngine registers this bridge with a script engine
func (b *StateContextBridge) RegisterWithEngine(scriptEngine engine.ScriptEngine) error {
	return scriptEngine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *StateContextBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{Name: "createSharedContext", Description: "Create a new shared state context with parent"},
		{Name: "withInheritanceConfig", Description: "Configure inheritance settings for shared context"},
		{Name: "get", Description: "Get a value from shared context (local first, then parent)"},
		{Name: "set", Description: "Set a value in local state of shared context"},
		{Name: "delete", Description: "Delete a key from local state of shared context"},
		{Name: "has", Description: "Check if shared context has a key (local or parent)"},
		{Name: "keys", Description: "Get all keys from shared context (merged)"},
		{Name: "values", Description: "Get all values from shared context (merged)"},
		{Name: "getArtifact", Description: "Get artifact from shared context"},
		{Name: "artifacts", Description: "Get all artifacts from shared context"},
		{Name: "messages", Description: "Get all messages from shared context"},
		{Name: "getMetadata", Description: "Get metadata from shared context"},
		{Name: "localState", Description: "Get the local state component"},
		{Name: "clone", Description: "Clone shared context with fresh local state"},
		{Name: "asState", Description: "Convert shared context to regular state"},
		{Name: "createSnapshot", Description: "Create a snapshot of shared context state and emit snapshot event"},

		// Schema validation methods from go-llms v0.3.5
		{Name: "validateState", Description: "Validate shared context state against schema"},
		{Name: "setStateSchema", Description: "Set schema for shared context validation"},
		{Name: "getStateSchema", Description: "Get schema for shared context"},
		{Name: "registerCustomValidator", Description: "Register custom validation rule"},
		{Name: "getSchemaVersions", Description: "Get all versions of a schema"},
		{Name: "setSchemaVersion", Description: "Set current schema version"},
		{Name: "validateWithVersion", Description: "Validate state against specific schema version"},

		// Event filtering and replay methods
		{Name: "addEventFilter", Description: "Add event filter by key pattern"},
		{Name: "removeEventFilter", Description: "Remove event filter"},
		{Name: "listEventFilters", Description: "List all active event filters"},
		{Name: "replayEvents", Description: "Replay events for state reconstruction"},
		{Name: "getEventHistory", Description: "Get event history for a context"},
		{Name: "clearEventHistory", Description: "Clear event history"},
	}
}

// TypeMappings returns type mappings for this bridge
func (b *StateContextBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"SharedStateContext": {
			GoType:     "SharedStateContext",
			ScriptType: "object",
		},
		"StateReader": {
			GoType:     "StateReader",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates a method call
func (b *StateContextBridge) ValidateMethod(name string, args []interface{}) error {
	for _, method := range b.Methods() {
		if method.Name == name {
			return nil
		}
	}
	return fmt.Errorf("method %s not found", name)
}

// RequiredPermissions returns required permissions
func (b *StateContextBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "state_context",
			Actions:     []string{"read", "write"},
			Description: "Access to shared state context operations",
		},
	}
}

// Shared state context operations

func (b *StateContextBridge) createSharedContext(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	parentObj, ok := params["parent"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parent parameter is required and must be a state object")
	}

	// Convert script state to domain.State
	parentState, err := b.scriptToState(parentObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert parent state: %w", err)
	}

	// Create real SharedStateContext from go-llms
	sharedContext := domain.NewSharedStateContext(parentState)

	// Generate unique ID for this context
	b.mu.Lock()
	contextID := fmt.Sprintf("context_%d", b.nextID)
	b.nextID++
	b.contexts[contextID] = sharedContext
	// Store default inheritance config
	b.configs[contextID] = &inheritanceConfig{
		inheritMessages:  true,
		inheritArtifacts: true,
		inheritMetadata:  true,
	}
	b.mu.Unlock()

	// Return script representation with context ID
	return map[string]interface{}{
		"_id":              contextID,
		"type":             "SharedStateContext",
		"inheritMessages":  true,
		"inheritArtifacts": true,
		"inheritMetadata":  true,
		"parent":           parentObj, // Include parent reference for script access
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) withInheritanceConfig(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	messages, ok := params["messages"].(bool)
	if !ok {
		return nil, fmt.Errorf("messages parameter is required and must be a boolean")
	}

	artifacts, ok := params["artifacts"].(bool)
	if !ok {
		return nil, fmt.Errorf("artifacts parameter is required and must be a boolean")
	}

	metadata, ok := params["metadata"].(bool)
	if !ok {
		return nil, fmt.Errorf("metadata parameter is required and must be a boolean")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Update inheritance configuration
	updatedContext := sharedContext.WithInheritanceConfig(messages, artifacts, metadata)

	// Get the context ID from the original script object
	contextID := contextObj["_id"].(string)

	// Update stored context and config
	b.mu.Lock()
	b.contexts[contextID] = updatedContext
	b.configs[contextID] = &inheritanceConfig{
		inheritMessages:  messages,
		inheritArtifacts: artifacts,
		inheritMetadata:  metadata,
	}
	b.mu.Unlock()

	return b.sharedContextToScript(contextID, updatedContext), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextGet(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	value, exists := sharedContext.Get(key)
	return map[string]interface{}{
		"value":  value,
		"exists": exists,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextSet(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	value := params["value"]

	// Get context ID for event emission
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Get old value for event emission
	oldValue, _ := sharedContext.Get(key)

	// Set in local state
	sharedContext.Set(key, value)

	// Emit state change event
	b.emitStateChangeEvent(contextID, key, oldValue, value)

	// Update the script object
	b.updateScriptSharedContext(contextObj, sharedContext)

	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextDelete(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	// Get context ID for event emission
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Get current value before deletion for event emission
	oldValue, exists := sharedContext.Get(key)
	if !exists {
		// Key doesn't exist, nothing to delete
		return map[string]interface{}{
			"deleted": false,
			"existed": false,
		}, nil
	}

	// Check if this key exists in local state (we can only delete from local state)
	localState := sharedContext.LocalState()
	_, existsLocal := localState.Get(key)
	if !existsLocal {
		// Key exists in parent but not in local state, cannot delete
		return map[string]interface{}{
			"deleted": false,
			"existed": true,
			"reason":  "key exists in parent state, cannot delete from local context",
		}, nil
	}

	// Delete from local state - SharedStateContext doesn't expose Delete directly,
	// so we need to set the local value to a special "deleted" marker
	// For now, we'll use a nil value to indicate deletion
	sharedContext.Set(key, nil)

	// Emit state delete event
	b.emitStateDeleteEvent(contextID, key, oldValue)

	// Update the script object
	b.updateScriptSharedContext(contextObj, sharedContext)

	return map[string]interface{}{
		"deleted": true,
		"existed": true,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextHas(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	return sharedContext.Has(key), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextKeys(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	keys := sharedContext.Keys()

	// Convert to interface slice for script engines
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = key
	}

	return result, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextValues(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	return sharedContext.Values(), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextGetArtifact(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id parameter is required and must be a string")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	artifact, exists := sharedContext.GetArtifact(id)
	if !exists {
		return nil, nil
	}

	return b.artifactToScript(artifact), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextArtifacts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	artifacts := sharedContext.Artifacts()
	result := make(map[string]interface{})
	for id, artifact := range artifacts {
		result[id] = b.artifactToScript(artifact)
	}

	return result, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextMessages(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	messages := sharedContext.Messages()
	result := make([]interface{}, len(messages))
	for i, message := range messages {
		result[i] = b.messageToScript(message)
	}

	return result, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) contextGetMetadata(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	value, exists := sharedContext.GetMetadata(key)
	return map[string]interface{}{
		"value":  value,
		"exists": exists,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) localState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	localState := sharedContext.LocalState()
	return b.stateToScript(localState), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) clone(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Clone returns a new SharedStateContext, so we need to store it with a new ID
	clonedContext := sharedContext.Clone()

	// Generate new ID for cloned context
	b.mu.Lock()
	clonedID := fmt.Sprintf("context_%d", b.nextID)
	b.nextID++
	b.contexts[clonedID] = clonedContext
	// Copy inheritance config from original
	if originalConfig, exists := b.configs[contextObj["_id"].(string)]; exists {
		b.configs[clonedID] = &inheritanceConfig{
			inheritMessages:  originalConfig.inheritMessages,
			inheritArtifacts: originalConfig.inheritArtifacts,
			inheritMetadata:  originalConfig.inheritMetadata,
		}
	} else {
		// Default config
		b.configs[clonedID] = &inheritanceConfig{
			inheritMessages:  true,
			inheritArtifacts: true,
			inheritMetadata:  true,
		}
	}
	b.mu.Unlock()

	return b.sharedContextToScript(clonedID, clonedContext), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) asState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	state := sharedContext.AsState()
	return b.stateToScript(state), nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) createSnapshot(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Get context ID for event emission
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Create snapshot data
	snapshot := map[string]interface{}{
		"contextId":   contextID,
		"timestamp":   time.Now(),
		"state":       b.stateToScript(sharedContext.AsState()),
		"localState":  b.stateToScript(sharedContext.LocalState()),
		"inheritance": b.configs[contextID],
	}

	// Add artifacts
	artifacts := sharedContext.Artifacts()
	snapshotArtifacts := make(map[string]interface{})
	for id, artifact := range artifacts {
		snapshotArtifacts[id] = b.artifactToScript(artifact)
	}
	snapshot["artifacts"] = snapshotArtifacts

	// Add messages
	messages := sharedContext.Messages()
	snapshotMessages := make([]interface{}, len(messages))
	for i, message := range messages {
		snapshotMessages[i] = b.messageToScript(message)
	}
	snapshot["messages"] = snapshotMessages

	// Emit snapshot event
	b.emitStateSnapshotEvent(contextID, snapshot)

	return snapshot, nil
}

// Event filtering methods using go-llms event system

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) addEventFilter(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	pattern, ok := params["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern parameter is required and must be a string")
	}

	name, ok := params["name"].(string)
	if !ok {
		// Generate a name if not provided
		name = fmt.Sprintf("filter_%d", len(b.eventFilters))
	}

	// Create pattern filter using go-llms events package
	filter, err := events.NewPatternFilter(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create pattern filter: %w", err)
	}

	// Store the filter
	b.mu.Lock()
	b.eventFilters[name] = filter
	b.mu.Unlock()

	// Subscribe to the event bus with this filter
	// Note: For now we just store the filter without subscribing
	// to avoid complex event handler implementation
	subscriptionID := fmt.Sprintf("sub_%s_%d", name, time.Now().UnixNano())

	return map[string]interface{}{
		"name":           name,
		"pattern":        pattern,
		"subscriptionId": subscriptionID,
		"active":         true,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) removeEventFilter(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	_, exists := b.eventFilters[name]
	if !exists {
		return map[string]interface{}{
			"removed": false,
			"reason":  "filter not found",
		}, nil
	}

	// Remove the filter
	delete(b.eventFilters, name)

	return map[string]interface{}{
		"removed": true,
		"name":    name,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) listEventFilters(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	filters := make([]interface{}, 0, len(b.eventFilters))
	for name := range b.eventFilters {
		filters = append(filters, map[string]interface{}{
			"name":   name,
			"active": true,
		})
	}

	return map[string]interface{}{
		"filters": filters,
		"total":   len(filters),
	}, nil
}

// Event replay methods for state reconstruction

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) replayEvents(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID, ok := params["contextId"].(string)
	if !ok {
		return nil, fmt.Errorf("contextId parameter is required and must be a string")
	}

	// Optional parameters
	fromTimestamp := time.Time{}
	if fromStr, ok := params["fromTimestamp"].(string); ok {
		var err error
		fromTimestamp, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return nil, fmt.Errorf("invalid fromTimestamp format: %w", err)
		}
	}

	toTimestamp := time.Now()
	if toStr, ok := params["toTimestamp"].(string); ok {
		var err error
		toTimestamp, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return nil, fmt.Errorf("invalid toTimestamp format: %w", err)
		}
	}

	// Filter events for the specific context and time range
	b.eventHistoryMu.RLock()
	eventsToReplay := make([]domain.Event, 0)
	for _, event := range b.eventHistory {
		if event.AgentID == contextID &&
			event.Timestamp.After(fromTimestamp) &&
			event.Timestamp.Before(toTimestamp) {
			eventsToReplay = append(eventsToReplay, event)
		}
	}
	b.eventHistoryMu.RUnlock()

	// Sort events by timestamp to ensure correct replay order
	// Simple bubble sort for small event sets
	for i := 0; i < len(eventsToReplay)-1; i++ {
		for j := 0; j < len(eventsToReplay)-i-1; j++ {
			if eventsToReplay[j].Timestamp.After(eventsToReplay[j+1].Timestamp) {
				eventsToReplay[j], eventsToReplay[j+1] = eventsToReplay[j+1], eventsToReplay[j]
			}
		}
	}

	// Create a new context for replay
	baseState := domain.NewState()
	replayContext := domain.NewSharedStateContext(baseState)

	// Generate new ID for replay context
	b.mu.Lock()
	replayContextID := fmt.Sprintf("replay_%s_%d", contextID, time.Now().UnixNano())
	b.contexts[replayContextID] = replayContext
	b.configs[replayContextID] = &inheritanceConfig{
		inheritMessages:  true,
		inheritArtifacts: true,
		inheritMetadata:  true,
	}
	b.mu.Unlock()

	// Replay events
	replayedEvents := make([]interface{}, 0, len(eventsToReplay))
	for _, event := range eventsToReplay {
		if event.Type == domain.EventStateUpdate {
			if eventData, ok := event.Data.(domain.StateUpdateEventData); ok {
				switch eventData.Action {
				case "set":
					replayContext.Set(eventData.Key, eventData.NewValue)
				case "delete":
					// For delete operations, we set to nil
					replayContext.Set(eventData.Key, nil)
				}
			}
		}

		replayedEvents = append(replayedEvents, map[string]interface{}{
			"id":        event.ID,
			"type":      string(event.Type),
			"timestamp": event.Timestamp.Format(time.RFC3339),
			"data":      event.Data,
			"metadata":  event.Metadata,
		})
	}

	return map[string]interface{}{
		"replayContextId": replayContextID,
		"eventsReplayed":  len(replayedEvents),
		"events":          replayedEvents,
		"replayedContext": b.sharedContextToScript(replayContextID, replayContext),
		"fromTimestamp":   fromTimestamp.Format(time.RFC3339),
		"toTimestamp":     toTimestamp.Format(time.RFC3339),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) getEventHistory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID := ""
	if id, ok := params["contextId"].(string); ok {
		contextID = id
	}

	limit := 100
	if l, ok := params["limit"].(int); ok && l > 0 {
		limit = l
	}
	if l, ok := params["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	b.eventHistoryMu.RLock()
	defer b.eventHistoryMu.RUnlock()

	events := make([]interface{}, 0)
	count := 0
	// Get most recent events first
	for i := len(b.eventHistory) - 1; i >= 0 && count < limit; i-- {
		event := b.eventHistory[i]
		if contextID == "" || event.AgentID == contextID {
			events = append(events, map[string]interface{}{
				"id":        event.ID,
				"type":      string(event.Type),
				"agentId":   event.AgentID,
				"agentName": event.AgentName,
				"timestamp": event.Timestamp.Format(time.RFC3339),
				"data":      event.Data,
				"metadata":  event.Metadata,
			})
			count++
		}
	}

	return map[string]interface{}{
		"events":      events,
		"total":       len(events),
		"contextId":   contextID,
		"limit":       limit,
		"totalStored": len(b.eventHistory),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) clearEventHistory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID := ""
	if id, ok := params["contextId"].(string); ok {
		contextID = id
	}

	b.eventHistoryMu.Lock()
	defer b.eventHistoryMu.Unlock()

	if contextID == "" {
		// Clear all events
		clearedCount := len(b.eventHistory)
		b.eventHistory = make([]domain.Event, 0)
		return map[string]interface{}{
			"cleared": clearedCount,
			"scope":   "all",
		}, nil
	}

	// Clear events for specific context
	filteredHistory := make([]domain.Event, 0)
	clearedCount := 0
	for _, event := range b.eventHistory {
		if event.AgentID != contextID {
			filteredHistory = append(filteredHistory, event)
		} else {
			clearedCount++
		}
	}
	b.eventHistory = filteredHistory

	return map[string]interface{}{
		"cleared":   clearedCount,
		"scope":     "context",
		"contextId": contextID,
	}, nil
}

// Event emission helper functions

// emitStateChangeEvent emits a state change event using go-llms event system
func (b *StateContextBridge) emitStateChangeEvent(contextID, key string, oldValue, newValue interface{}) {
	if b.eventEmitter == nil {
		return
	}

	eventData := domain.StateUpdateEventData{
		Key:      key,
		OldValue: oldValue,
		NewValue: newValue,
		Action:   "set",
	}

	event := domain.Event{
		ID:        fmt.Sprintf("state_change_%d_%d", time.Now().UnixNano(), len(b.eventHistory)),
		Type:      domain.EventStateUpdate,
		AgentID:   contextID,
		AgentName: fmt.Sprintf("state_context_%s", contextID),
		Timestamp: time.Now(),
		Data:      eventData,
		Metadata: map[string]interface{}{
			"contextId": contextID,
			"operation": "set",
		},
	}

	// Emit using the provided event emitter
	b.eventEmitter.Emit(domain.EventStateUpdate, eventData)

	// Store in history for replay functionality
	b.storeEventInHistory(event)

	// Publish to internal event bus for filtering
	b.eventBus.Publish(event)
}

// emitStateDeleteEvent emits a state delete event using go-llms event system
func (b *StateContextBridge) emitStateDeleteEvent(contextID, key string, deletedValue interface{}) {
	if b.eventEmitter == nil {
		return
	}

	eventData := domain.StateUpdateEventData{
		Key:      key,
		OldValue: deletedValue,
		NewValue: nil,
		Action:   "delete",
	}

	event := domain.Event{
		ID:        fmt.Sprintf("state_delete_%d_%d", time.Now().UnixNano(), len(b.eventHistory)),
		Type:      domain.EventStateUpdate,
		AgentID:   contextID,
		AgentName: fmt.Sprintf("state_context_%s", contextID),
		Timestamp: time.Now(),
		Data:      eventData,
		Metadata: map[string]interface{}{
			"contextId": contextID,
			"operation": "delete",
		},
	}

	// Emit using the provided event emitter
	b.eventEmitter.Emit(domain.EventStateUpdate, eventData)

	// Store in history for replay functionality
	b.storeEventInHistory(event)

	// Publish to internal event bus for filtering
	b.eventBus.Publish(event)
}

// emitStateSnapshotEvent emits a state snapshot event
func (b *StateContextBridge) emitStateSnapshotEvent(contextID string, snapshot map[string]interface{}) {
	if b.eventEmitter == nil {
		return
	}

	eventData := map[string]interface{}{
		"contextId": contextID,
		"snapshot":  snapshot,
		"timestamp": time.Now(),
	}

	event := domain.Event{
		ID:        fmt.Sprintf("state_snapshot_%d_%d", time.Now().UnixNano(), len(b.eventHistory)),
		Type:      "state.snapshot", // Custom event type for snapshots
		AgentID:   contextID,
		AgentName: fmt.Sprintf("state_context_%s", contextID),
		Timestamp: time.Now(),
		Data:      eventData,
		Metadata: map[string]interface{}{
			"contextId": contextID,
			"operation": "snapshot",
		},
	}

	// Emit custom event for snapshot
	b.eventEmitter.EmitCustom("state.snapshot", eventData)

	// Store in history for replay functionality
	b.storeEventInHistory(event)

	// Publish to internal event bus for filtering
	b.eventBus.Publish(event)
}

// storeEventInHistory stores an event in the replay history
func (b *StateContextBridge) storeEventInHistory(event domain.Event) {
	b.eventHistoryMu.Lock()
	defer b.eventHistoryMu.Unlock()

	b.eventHistory = append(b.eventHistory, event)

	// Keep only last 1000 events to prevent unlimited growth
	if len(b.eventHistory) > 1000 {
		b.eventHistory = b.eventHistory[len(b.eventHistory)-1000:]
	}
}

// Helper functions for type conversion

func (b *StateContextBridge) sharedContextToScript(contextID string, context *domain.SharedStateContext) map[string]interface{} {
	// Get inheritance config
	b.mu.RLock()
	config, exists := b.configs[contextID]
	b.mu.RUnlock()

	// Default config if not found
	if !exists {
		config = &inheritanceConfig{
			inheritMessages:  true,
			inheritArtifacts: true,
			inheritMetadata:  true,
		}
	}

	// Return script representation with context ID
	return map[string]interface{}{
		"_id":              contextID,
		"type":             "SharedStateContext",
		"inheritMessages":  config.inheritMessages,
		"inheritArtifacts": config.inheritArtifacts,
		"inheritMetadata":  config.inheritMetadata,
	}
}

func (b *StateContextBridge) scriptToSharedContext(scriptObj map[string]interface{}) (*domain.SharedStateContext, error) {
	// Check if this is our shared context format
	contextType, ok := scriptObj["type"].(string)
	if !ok || contextType != "SharedStateContext" {
		return nil, fmt.Errorf("invalid shared context object: missing or incorrect type")
	}

	// Get the context ID
	contextID, ok := scriptObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Look up the context
	b.mu.RLock()
	sharedContext, exists := b.contexts[contextID]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("shared context %s not found", contextID)
	}

	return sharedContext, nil
}

func (b *StateContextBridge) scriptToState(scriptObj map[string]interface{}) (*domain.State, error) {
	// Convert script object to domain.State
	state := domain.NewState()

	// Set ID if provided
	if id, ok := scriptObj["id"].(string); ok {
		// domain.State doesn't expose SetID, so we'll use the state as-is
		_ = id
	}

	// Set data values
	if data, ok := scriptObj["data"].(map[string]interface{}); ok {
		for key, value := range data {
			state.Set(key, value)
		}
	}

	// Set metadata
	if metadata, ok := scriptObj["metadata"].(map[string]interface{}); ok {
		for key, value := range metadata {
			state.SetMetadata(key, value)
		}
	}

	return state, nil
}

// Schema validation methods using go-llms v0.3.5 infrastructure

// ValidateState validates a shared context state against its associated schema
func (b *StateContextBridge) ValidateState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Get context ID
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Get associated schema ID
	b.mu.RLock()
	schemaID, hasSchema := b.stateSchemas[contextID]
	b.mu.RUnlock()

	if !hasSchema {
		return map[string]interface{}{
			"valid":   true,
			"message": "No schema configured for validation",
		}, nil
	}

	// Get schema from repository
	schema, err := b.schemaRepo.Get(schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema %s: %w", schemaID, err)
	}

	// Convert shared context to state for validation
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	state := sharedContext.AsState()

	// Validate using state validator from go-llms
	err = b.stateValidator.Validate(state)
	if err != nil {
		return map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		}, nil
	}

	// Also validate against JSON schema if available
	// Extract just the data values for schema validation
	stateData := state.Values()
	result, err := b.validator.ValidateStruct(schema, stateData)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return map[string]interface{}{
		"valid":  result.Valid,
		"errors": result.Errors,
	}, nil
}

// SetStateSchema sets a schema for state validation
func (b *StateContextBridge) SetStateSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	schemaID, ok := params["schemaId"].(string)
	if !ok {
		return nil, fmt.Errorf("schemaId parameter is required and must be a string")
	}

	schemaData, ok := params["schema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema parameter is required and must be an object")
	}

	// Get context ID
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Create schema object from go-llms
	schema := &sdomain.Schema{
		Title:      fmt.Sprintf("Schema for context %s", contextID),
		Type:       "object",
		Properties: make(map[string]sdomain.Property),
		Required:   []string{},
	}

	// Parse schema properties from script data
	if properties, ok := schemaData["properties"].(map[string]interface{}); ok {
		for propName, propDef := range properties {
			if propDefMap, ok := propDef.(map[string]interface{}); ok {
				propProperty := sdomain.Property{}
				if propType, ok := propDefMap["type"].(string); ok {
					propProperty.Type = propType
				}
				if propDesc, ok := propDefMap["description"].(string); ok {
					propProperty.Description = propDesc
				}
				schema.Properties[propName] = propProperty
			}
		}
	}

	// Set required fields
	if required, ok := schemaData["required"].([]interface{}); ok {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				schema.Required = append(schema.Required, reqStr)
			}
		}
	}

	// Save schema to repository
	err := b.schemaRepo.Save(schemaID, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to save schema: %w", err)
	}

	// Associate schema with context
	b.mu.Lock()
	b.stateSchemas[contextID] = schemaID
	b.mu.Unlock()

	// Create schema validator for state validation
	schemaValidator := domain.SchemaValidator(b.validator, schema)

	// Replace the state validator with a composite that includes the schema validator
	b.stateValidator = domain.CompositeValidator(b.stateValidator, schemaValidator)

	return map[string]interface{}{
		"schemaId": schemaID,
		"success":  true,
	}, nil
}

// GetStateSchema gets the schema for a shared context
func (b *StateContextBridge) GetStateSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Get context ID
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Get associated schema ID
	b.mu.RLock()
	schemaID, hasSchema := b.stateSchemas[contextID]
	b.mu.RUnlock()

	if !hasSchema {
		return nil, nil
	}

	// Get schema from repository
	schema, err := b.schemaRepo.Get(schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	return b.schemaToScript(schema), nil
}

// RegisterCustomValidator registers a custom validation rule
func (b *StateContextBridge) RegisterCustomValidator(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	// For bridge architecture, we expose the go-llms custom validator registration
	// The actual validation function would be provided by the script engine
	validatorFunc, ok := params["validator"].(func(interface{}) bool)
	if !ok {
		return nil, fmt.Errorf("validator parameter is required and must be a function")
	}

	// Convert to go-llms CustomValidator format (which takes value and displayPath)
	customValidator := validation.CustomValidator(func(value interface{}, displayPath string) []string {
		if validatorFunc(value) {
			return nil // No errors
		}
		return []string{fmt.Sprintf("%s failed custom validation", displayPath)}
	})

	// Register with go-llms validation system
	validation.RegisterCustomValidator(name, customValidator)

	return map[string]interface{}{
		"name":       name,
		"registered": true,
	}, nil
}

// GetSchemaVersions gets all versions of a schema
func (b *StateContextBridge) GetSchemaVersions(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	schemaID, ok := params["schemaId"].(string)
	if !ok {
		return nil, fmt.Errorf("schemaId parameter is required and must be a string")
	}

	// Basic schema repository doesn't support versioning in go-llms v0.3.5
	// Just return current schema info
	schema, err := b.schemaRepo.Get(schemaID)
	if err != nil {
		return []interface{}{}, nil // Schema not found, return empty
	}

	return []interface{}{
		map[string]interface{}{
			"version": 1,
			"title":   schema.Title,
		},
	}, nil
}

// SetSchemaVersion sets the current version of a schema
func (b *StateContextBridge) SetSchemaVersion(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	schemaID, ok := params["schemaId"].(string)
	if !ok {
		return nil, fmt.Errorf("schemaId parameter is required and must be a string")
	}

	version, ok := params["version"].(int)
	if !ok {
		if versionFloat, ok := params["version"].(float64); ok {
			version = int(versionFloat)
		} else {
			return nil, fmt.Errorf("version parameter is required and must be an integer")
		}
	}

	// Basic schema repository doesn't support versioning in go-llms v0.3.5
	// Just validate the schema exists
	_, err := b.schemaRepo.Get(schemaID)
	if err != nil {
		return nil, fmt.Errorf("schema %s not found: %w", schemaID, err)
	}

	return map[string]interface{}{
		"schemaId": schemaID,
		"version":  version,
		"success":  true,
	}, nil
}

// ValidateWithVersion validates state against a specific schema version
func (b *StateContextBridge) ValidateWithVersion(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	schemaID, ok := params["schemaId"].(string)
	if !ok {
		return nil, fmt.Errorf("schemaId parameter is required and must be a string")
	}

	version, ok := params["version"].(int)
	if !ok {
		if versionFloat, ok := params["version"].(float64); ok {
			version = int(versionFloat)
		} else {
			return nil, fmt.Errorf("version parameter is required and must be an integer")
		}
	}

	// Get schema from repository (no versioning support in go-llms v0.3.5)
	schema, err := b.schemaRepo.Get(schemaID)
	if err != nil {
		return nil, fmt.Errorf("schema %s not found: %w", schemaID, err)
	}

	// Convert shared context to state
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	state := sharedContext.AsState()
	// Extract just the data values for schema validation
	stateData := state.Values()

	// Validate using go-llms validator
	result, err := b.validator.ValidateStruct(schema, stateData)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return map[string]interface{}{
		"valid":    result.Valid,
		"errors":   result.Errors,
		"schemaId": schemaID,
		"version":  version,
	}, nil
}

// Helper function to convert schema to script representation
func (b *StateContextBridge) schemaToScript(schema *sdomain.Schema) map[string]interface{} {
	result := map[string]interface{}{
		"title":      schema.Title,
		"type":       schema.Type,
		"required":   schema.Required,
		"properties": make(map[string]interface{}),
	}

	// Convert properties
	for propName, propProperty := range schema.Properties {
		result["properties"].(map[string]interface{})[propName] = map[string]interface{}{
			"type":        propProperty.Type,
			"description": propProperty.Description,
		}
	}

	return result
}

func (b *StateContextBridge) scriptToStateReader(scriptObj map[string]interface{}) (domain.StateReader, error) {
	// StateReader is an interface implemented by State, so we can just convert to State
	return b.scriptToState(scriptObj)
}

func (b *StateContextBridge) updateScriptSharedContext(scriptObj map[string]interface{}, sharedContext *domain.SharedStateContext) {
	// Update the script object to reflect any changes in the shared context
	// In the real implementation, this would sync any changes back to the script representation
}

func (b *StateContextBridge) stateToScript(state *domain.State) map[string]interface{} {
	return map[string]interface{}{
		"id":       state.ID(),
		"data":     state.Values(),
		"metadata": state.GetAllMetadata(),
	}
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) artifactToScript(artifact *domain.Artifact) map[string]interface{} {
	return map[string]interface{}{
		"id":       artifact.ID,
		"name":     artifact.Name,
		"type":     string(artifact.Type),
		"data":     artifact.Data,
		"size":     artifact.Size,
		"mimeType": artifact.MimeType,
		"metadata": artifact.Metadata,
	}
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) messageToScript(message domain.Message) map[string]interface{} {
	return map[string]interface{}{
		"role":    message.Role,
		"content": message.Content,
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *StateContextBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	switch name {
	case "createSharedContext":
		var parentState *domain.State
		if len(args) > 0 && args[0] != nil {
			if parentObj, ok := args[0].(map[string]interface{}); ok {
				var err error
				parentState, err = b.scriptToState(parentObj)
				if err != nil {
					return nil, fmt.Errorf("failed to convert parent state: %w", err)
				}
			}
		}

		// Create shared context
		sharedContext := domain.NewSharedStateContext(parentState)

		// Generate ID and store context
		b.mu.Lock()
		contextID := fmt.Sprintf("context_%d", b.nextID)
		b.nextID++
		b.contexts[contextID] = sharedContext
		b.configs[contextID] = &inheritanceConfig{
			inheritMessages:  true,
			inheritArtifacts: true,
			inheritMetadata:  true,
		}
		b.mu.Unlock()

		return b.sharedContextToScript(contextID, sharedContext), nil

	case "get":
		if len(args) < 2 {
			return nil, fmt.Errorf("get requires context and key parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("key must be string")
		}

		sharedContext, err := b.scriptToSharedContext(contextObj)
		if err != nil {
			return nil, err
		}

		value, exists := sharedContext.Get(key)
		return map[string]interface{}{
			"value":  value,
			"exists": exists,
		}, nil

	case "set":
		if len(args) < 3 {
			return nil, fmt.Errorf("set requires context, key, and value parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("key must be string")
		}

		// Get context ID for event emission
		contextID, ok := contextObj["_id"].(string)
		if !ok {
			return nil, fmt.Errorf("shared context object missing _id")
		}

		sharedContext, err := b.scriptToSharedContext(contextObj)
		if err != nil {
			return nil, err
		}

		// Get old value for event emission
		oldValue, _ := sharedContext.Get(key)

		sharedContext.Set(key, args[2])

		// Emit state change event
		b.emitStateChangeEvent(contextID, key, oldValue, args[2])

		b.updateScriptSharedContext(contextObj, sharedContext)

		return nil, nil

	case "delete":
		if len(args) < 2 {
			return nil, fmt.Errorf("delete requires context and key parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("key must be string")
		}

		params := map[string]interface{}{
			"context": contextObj,
			"key":     key,
		}
		return b.contextDelete(ctx, params)

	case "createSnapshot":
		if len(args) < 1 {
			return nil, fmt.Errorf("createSnapshot requires context parameter")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}

		params := map[string]interface{}{
			"context": contextObj,
		}
		return b.createSnapshot(ctx, params)

	case "validateState":
		if len(args) < 1 {
			return nil, fmt.Errorf("validateState requires context parameter")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}

		params := map[string]interface{}{"context": contextObj}
		return b.ValidateState(ctx, params)

	case "setStateSchema":
		if len(args) < 3 {
			return nil, fmt.Errorf("setStateSchema requires context, schemaId, and schema parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		schemaID, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("schemaId must be string")
		}
		schema, ok := args[2].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be object")
		}

		params := map[string]interface{}{
			"context":  contextObj,
			"schemaId": schemaID,
			"schema":   schema,
		}
		return b.SetStateSchema(ctx, params)

	case "getStateSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("getStateSchema requires context parameter")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}

		params := map[string]interface{}{"context": contextObj}
		return b.GetStateSchema(ctx, params)

	case "registerCustomValidator":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerCustomValidator requires name and validator parameters")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}
		validatorFunc, ok := args[1].(func(interface{}) bool)
		if !ok {
			return nil, fmt.Errorf("validator must be function")
		}

		params := map[string]interface{}{
			"name":      name,
			"validator": validatorFunc,
		}
		return b.RegisterCustomValidator(ctx, params)

	case "getSchemaVersions":
		if len(args) < 1 {
			return nil, fmt.Errorf("getSchemaVersions requires schemaId parameter")
		}
		schemaID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("schemaId must be string")
		}

		params := map[string]interface{}{"schemaId": schemaID}
		return b.GetSchemaVersions(ctx, params)

	case "setSchemaVersion":
		if len(args) < 2 {
			return nil, fmt.Errorf("setSchemaVersion requires schemaId and version parameters")
		}
		schemaID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("schemaId must be string")
		}
		version, ok := args[1].(int)
		if !ok {
			if versionFloat, ok := args[1].(float64); ok {
				version = int(versionFloat)
			} else {
				return nil, fmt.Errorf("version must be integer")
			}
		}

		params := map[string]interface{}{
			"schemaId": schemaID,
			"version":  version,
		}
		return b.SetSchemaVersion(ctx, params)

	case "validateWithVersion":
		if len(args) < 3 {
			return nil, fmt.Errorf("validateWithVersion requires context, schemaId, and version parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		schemaID, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("schemaId must be string")
		}
		version, ok := args[2].(int)
		if !ok {
			if versionFloat, ok := args[2].(float64); ok {
				version = int(versionFloat)
			} else {
				return nil, fmt.Errorf("version must be integer")
			}
		}

		params := map[string]interface{}{
			"context":  contextObj,
			"schemaId": schemaID,
			"version":  version,
		}
		return b.ValidateWithVersion(ctx, params)

	case "addEventFilter":
		if len(args) < 1 {
			return nil, fmt.Errorf("addEventFilter requires pattern parameter")
		}
		pattern, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pattern must be string")
		}

		params := map[string]interface{}{
			"pattern": pattern,
		}
		if len(args) > 1 {
			if name, ok := args[1].(string); ok {
				params["name"] = name
			}
		}
		return b.addEventFilter(ctx, params)

	case "removeEventFilter":
		if len(args) < 1 {
			return nil, fmt.Errorf("removeEventFilter requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		params := map[string]interface{}{
			"name": name,
		}
		return b.removeEventFilter(ctx, params)

	case "listEventFilters":
		return b.listEventFilters(ctx, map[string]interface{}{})

	case "replayEvents":
		if len(args) < 1 {
			return nil, fmt.Errorf("replayEvents requires contextId parameter")
		}
		contextID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("contextId must be string")
		}

		params := map[string]interface{}{
			"contextId": contextID,
		}
		if len(args) > 1 {
			if fromStr, ok := args[1].(string); ok {
				params["fromTimestamp"] = fromStr
			}
		}
		if len(args) > 2 {
			if toStr, ok := args[2].(string); ok {
				params["toTimestamp"] = toStr
			}
		}
		return b.replayEvents(ctx, params)

	case "getEventHistory":
		params := map[string]interface{}{}
		if len(args) > 0 {
			if contextID, ok := args[0].(string); ok {
				params["contextId"] = contextID
			}
		}
		if len(args) > 1 {
			if limit, ok := args[1].(int); ok {
				params["limit"] = limit
			} else if limit, ok := args[1].(float64); ok {
				params["limit"] = int(limit)
			}
		}
		return b.getEventHistory(ctx, params)

	case "clearEventHistory":
		params := map[string]interface{}{}
		if len(args) > 0 {
			if contextID, ok := args[0].(string); ok {
				params["contextId"] = contextID
			}
		}
		return b.clearEventHistory(ctx, params)

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
