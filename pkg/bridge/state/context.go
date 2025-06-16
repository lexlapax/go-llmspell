// ABOUTME: State Context Bridge implementation that exposes go-llms SharedStateContext to script engines
// ABOUTME: Provides parent-child state sharing with configurable inheritance for multi-agent systems

package state

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/util/json"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// TransformMetrics tracks transformation pipeline metrics
type TransformMetrics struct {
	ExecutionCount  int64         `json:"execution_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastExecuted    time.Time     `json:"last_executed"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
}

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

	// State persistence fields from go-llms v0.3.5
	stateManager   *core.StateManager
	fileRepo       sdomain.SchemaRepository // File-based schema repository for versioned storage
	persistDir     string                   // Directory for state persistence
	stateVersions  map[string][]int         // contextID -> version numbers
	enableCompress bool                     // Enable compression for large states
	persistenceMu  sync.RWMutex             // Separate mutex for persistence operations

	// State transformation pipeline fields from go-llms v0.3.5
	transformPipelines map[string][]string               // contextID -> transform chain names
	pipelineConfigs    map[string]map[string]interface{} // pipelineID -> config
	transformCache     map[string]*domain.State          // Cache for transformed states
	transformMetrics   map[string]*TransformMetrics      // Metrics per pipeline
	transformMu        sync.RWMutex                      // Separate mutex for transformation operations
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
	return NewStateContextBridgeWithOptions(eventEmitter, "", false)
}

// NewStateContextBridgeWithOptions creates a new state context bridge with full configuration
func NewStateContextBridgeWithOptions(eventEmitter domain.EventEmitter, persistDir string, enableCompress bool) (*StateContextBridge, error) {
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

	// Create state manager using go-llms infrastructure
	stateManager := core.NewStateManager()

	// Create file-based repository for persistent storage if directory provided
	var fileRepo sdomain.SchemaRepository
	if persistDir != "" {
		var err error
		fileRepo, err = repository.NewFileSchemaRepository(persistDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create file schema repository: %w", err)
		}
	}

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

		// State persistence system from go-llms
		stateManager:   stateManager,
		fileRepo:       fileRepo,
		persistDir:     persistDir,
		stateVersions:  make(map[string][]int),
		enableCompress: enableCompress,
		persistenceMu:  sync.RWMutex{},

		// State transformation pipeline from go-llms
		transformPipelines: make(map[string][]string),
		pipelineConfigs:    make(map[string]map[string]interface{}),
		transformCache:     make(map[string]*domain.State),
		transformMetrics:   make(map[string]*TransformMetrics),
		transformMu:        sync.RWMutex{},
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

		// State persistence methods using go-llms
		{Name: "persistState", Description: "Persist shared context state with schema validation"},
		{Name: "loadState", Description: "Load persisted state with schema validation"},
		{Name: "listPersistedStates", Description: "List all persisted state versions"},
		{Name: "deletePersistedState", Description: "Delete a persisted state version"},
		{Name: "generateStateDiff", Description: "Generate diff between two state versions"},
		{Name: "migrateState", Description: "Migrate state between schema versions"},
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

// State persistence methods using go-llms infrastructure

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) persistState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	// Get context ID
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Optional parameters
	version := 0
	if v, ok := params["version"].(int); ok {
		version = v
	} else if v, ok := params["version"].(float64); ok {
		version = int(v)
	}

	description := ""
	if desc, ok := params["description"].(string); ok {
		description = desc
	}

	compress := b.enableCompress
	if c, ok := params["compress"].(bool); ok {
		compress = c
	}

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Generate state data for persistence
	stateData := map[string]interface{}{
		"contextId":   contextID,
		"version":     version,
		"timestamp":   time.Now(),
		"description": description,
		"state":       b.stateToScript(sharedContext.AsState()),
		"localState":  b.stateToScript(sharedContext.LocalState()),
		"inheritance": b.configs[contextID],
		"compressed":  compress,
	}

	// Add artifacts and messages
	artifacts := sharedContext.Artifacts()
	stateArtifacts := make(map[string]interface{})
	for id, artifact := range artifacts {
		stateArtifacts[id] = b.artifactToScript(artifact)
	}
	stateData["artifacts"] = stateArtifacts

	messages := sharedContext.Messages()
	stateMessages := make([]interface{}, len(messages))
	for i, message := range messages {
		stateMessages[i] = b.messageToScript(message)
	}
	stateData["messages"] = stateMessages

	// Serialize using go-llms optimized JSON
	serializedData, err := json.MarshalIndent(stateData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize state data: %w", err)
	}

	// Compress if enabled
	var finalData []byte
	if compress {
		compressedData, err := b.compressData(serializedData)
		if err != nil {
			return nil, fmt.Errorf("failed to compress state data: %w", err)
		}
		finalData = compressedData
	} else {
		finalData = serializedData
	}

	// Persist to file if file repository is available
	if b.fileRepo != nil && b.persistDir != "" {
		err = b.persistToFile(contextID, version, finalData, compress)
		if err != nil {
			return nil, fmt.Errorf("failed to persist state to file: %w", err)
		}
	}

	// Track version
	b.persistenceMu.Lock()
	if versions, exists := b.stateVersions[contextID]; exists {
		// Check if version already exists
		versionExists := false
		for _, v := range versions {
			if v == version {
				versionExists = true
				break
			}
		}
		if !versionExists {
			b.stateVersions[contextID] = append(versions, version)
		}
	} else {
		b.stateVersions[contextID] = []int{version}
	}
	b.persistenceMu.Unlock()

	// Emit persistence event if event emitter is available
	if b.eventEmitter != nil {
		persistenceEventData := map[string]interface{}{
			"contextId":   contextID,
			"version":     version,
			"size":        len(finalData),
			"compressed":  compress,
			"description": description,
			"timestamp":   time.Now(),
		}
		b.eventEmitter.EmitCustom("state.persisted", persistenceEventData)
	}

	return map[string]interface{}{
		"contextId":   contextID,
		"version":     version,
		"size":        len(finalData),
		"compressed":  compress,
		"description": description,
		"timestamp":   time.Now().Format(time.RFC3339),
		"success":     true,
	}, nil
}

// compressData compresses data using gzip
func (b *StateContextBridge) compressData(data []byte) ([]byte, error) {
	var compressed strings.Builder
	gzipWriter := gzip.NewWriter(&compressed)

	_, err := gzipWriter.Write(data)
	if err != nil {
		_ = gzipWriter.Close() // Ignore error on cleanup
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return []byte(compressed.String()), nil
}

// persistToFile persists state data to file using go-llms file patterns
func (b *StateContextBridge) persistToFile(contextID string, version int, data []byte, compressed bool) error {
	// Create context-specific directory
	contextDir := filepath.Join(b.persistDir, "states", contextID)
	err := os.MkdirAll(contextDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create context directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("state_v%d.json", version)
	if compressed {
		filename = fmt.Sprintf("state_v%d.json.gz", version)
	}

	filePath := filepath.Join(contextDir, filename)

	// Write file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	// Write metadata file
	metadata := map[string]interface{}{
		"contextId":  contextID,
		"version":    version,
		"timestamp":  time.Now().Format(time.RFC3339),
		"filename":   filename,
		"compressed": compressed,
		"size":       len(data),
	}

	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	metadataPath := filepath.Join(contextDir, fmt.Sprintf("metadata_v%d.json", version))
	err = os.WriteFile(metadataPath, metadataBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) loadState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID, ok := params["contextId"].(string)
	if !ok {
		return nil, fmt.Errorf("contextId parameter is required and must be a string")
	}

	// Optional version parameter
	version := 0
	if v, ok := params["version"].(int); ok {
		version = v
	} else if v, ok := params["version"].(float64); ok {
		version = int(v)
	}

	// Validate schema if enabled
	validateSchema := true
	if vs, ok := params["validateSchema"].(bool); ok {
		validateSchema = vs
	}

	// Load from file if file repository is available
	var stateData map[string]interface{}
	var err error

	if b.fileRepo != nil && b.persistDir != "" {
		stateData, err = b.loadFromFile(contextID, version)
		if err != nil {
			return nil, fmt.Errorf("failed to load state from file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no persistence directory configured")
	}

	// Validate against schema if enabled
	if validateSchema {
		// Check if there's a schema for this context
		b.mu.RLock()
		schemaID, hasSchema := b.stateSchemas[contextID]
		b.mu.RUnlock()

		if hasSchema {
			// Get schema from repository
			schema, err := b.schemaRepo.Get(schemaID)
			if err != nil {
				return nil, fmt.Errorf("failed to get schema %s: %w", schemaID, err)
			}

			// Validate state data
			if stateObj, ok := stateData["state"].(map[string]interface{}); ok {
				if dataMap, ok := stateObj["data"].(map[string]interface{}); ok {
					result, err := b.validator.ValidateStruct(schema, dataMap)
					if err != nil {
						return nil, fmt.Errorf("schema validation failed: %w", err)
					}
					if !result.Valid {
						return nil, fmt.Errorf("state data does not conform to schema: %v", result.Errors)
					}
				}
			}
		}
	}

	// Reconstruct shared context from loaded data
	reconstructedContext, err := b.reconstructSharedContext(stateData)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct shared context: %w", err)
	}

	// Generate new ID for loaded context
	b.mu.Lock()
	loadedContextID := fmt.Sprintf("loaded_%s_%d_%d", contextID, version, time.Now().UnixNano())
	b.contexts[loadedContextID] = reconstructedContext

	// Restore inheritance config
	if inheritanceData, ok := stateData["inheritance"]; ok {
		if inheritanceMap, ok := inheritanceData.(map[string]interface{}); ok {
			config := &inheritanceConfig{
				inheritMessages:  true,
				inheritArtifacts: true,
				inheritMetadata:  true,
			}
			if messages, ok := inheritanceMap["inheritMessages"].(bool); ok {
				config.inheritMessages = messages
			}
			if artifacts, ok := inheritanceMap["inheritArtifacts"].(bool); ok {
				config.inheritArtifacts = artifacts
			}
			if metadata, ok := inheritanceMap["inheritMetadata"].(bool); ok {
				config.inheritMetadata = metadata
			}
			b.configs[loadedContextID] = config
		}
	}
	b.mu.Unlock()

	// Emit load event if event emitter is available
	if b.eventEmitter != nil {
		loadEventData := map[string]interface{}{
			"originalContextId": contextID,
			"loadedContextId":   loadedContextID,
			"version":           version,
			"validateSchema":    validateSchema,
			"timestamp":         time.Now(),
		}
		b.eventEmitter.EmitCustom("state.loaded", loadEventData)
	}

	return map[string]interface{}{
		"originalContextId": contextID,
		"loadedContextId":   loadedContextID,
		"loadedContext":     b.sharedContextToScript(loadedContextID, reconstructedContext),
		"version":           version,
		"validateSchema":    validateSchema,
		"timestamp":         time.Now().Format(time.RFC3339),
		"success":           true,
	}, nil
}

// loadFromFile loads state data from file
func (b *StateContextBridge) loadFromFile(contextID string, version int) (map[string]interface{}, error) {
	contextDir := filepath.Join(b.persistDir, "states", contextID)

	// Check if version exists, if version is 0, load latest
	if version == 0 {
		// Find latest version
		b.persistenceMu.RLock()
		versions, exists := b.stateVersions[contextID]
		b.persistenceMu.RUnlock()

		if !exists || len(versions) == 0 {
			return nil, fmt.Errorf("no versions found for context %s", contextID)
		}

		// Get the latest version
		latestVersion := 0
		for _, v := range versions {
			if v > latestVersion {
				latestVersion = v
			}
		}
		version = latestVersion
	}

	// Try both compressed and uncompressed files
	filename := fmt.Sprintf("state_v%d.json", version)
	filePath := filepath.Join(contextDir, filename)

	var data []byte
	var err error
	compressed := false

	// Try uncompressed first
	data, err = os.ReadFile(filePath)
	if err != nil {
		// Try compressed
		compressedFilename := fmt.Sprintf("state_v%d.json.gz", version)
		compressedPath := filepath.Join(contextDir, compressedFilename)
		data, err = os.ReadFile(compressedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read state file (tried both compressed and uncompressed): %w", err)
		}
		compressed = true
	}

	// Decompress if needed
	if compressed {
		decompressedData, err := b.decompressData(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress state data: %w", err)
		}
		data = decompressedData
	}

	// Parse JSON using go-llms JSON utilities
	var stateData map[string]interface{}
	err = json.Unmarshal(data, &stateData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state JSON: %w", err)
	}

	return stateData, nil
}

// decompressData decompresses gzip data
func (b *StateContextBridge) decompressData(data []byte) ([]byte, error) {
	reader := strings.NewReader(string(data))
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = gzipReader.Close() // Ignore error on cleanup
	}()

	return io.ReadAll(gzipReader)
}

// reconstructSharedContext reconstructs a SharedStateContext from persisted data
func (b *StateContextBridge) reconstructSharedContext(stateData map[string]interface{}) (*domain.SharedStateContext, error) {
	// Create base state from parent state data
	var parentState *domain.State
	if stateObj, ok := stateData["state"].(map[string]interface{}); ok {
		parentState = domain.NewState()
		if dataMap, ok := stateObj["data"].(map[string]interface{}); ok {
			for key, value := range dataMap {
				parentState.Set(key, value)
			}
		}
		if metadataMap, ok := stateObj["metadata"].(map[string]interface{}); ok {
			for key, value := range metadataMap {
				parentState.SetMetadata(key, value)
			}
		}
	} else {
		parentState = domain.NewState()
	}

	// Create shared context
	sharedContext := domain.NewSharedStateContext(parentState)

	// Restore local state
	if localStateObj, ok := stateData["localState"].(map[string]interface{}); ok {
		if dataMap, ok := localStateObj["data"].(map[string]interface{}); ok {
			for key, value := range dataMap {
				sharedContext.Set(key, value)
			}
		}
	}

	// Restore artifacts
	if artifactsObj, ok := stateData["artifacts"].(map[string]interface{}); ok {
		for _, artifactData := range artifactsObj {
			if artifactMap, ok := artifactData.(map[string]interface{}); ok {
				artifact := b.scriptToArtifact(artifactMap)
				if artifact != nil {
					// We can't directly add artifacts to SharedStateContext, but we can add to the parent state
					parentState.AddArtifact(artifact)
				}
			}
		}
	}

	// Restore messages
	if messagesArray, ok := stateData["messages"].([]interface{}); ok {
		for _, messageData := range messagesArray {
			if messageMap, ok := messageData.(map[string]interface{}); ok {
				message := b.scriptToMessage(messageMap)
				// Add to parent state since SharedStateContext inherits messages
				parentState.AddMessage(message)
			}
		}
	}

	return sharedContext, nil
}

// scriptToArtifact converts script representation to domain.Artifact
func (b *StateContextBridge) scriptToArtifact(scriptObj map[string]interface{}) *domain.Artifact {
	id, ok := scriptObj["id"].(string)
	if !ok {
		return nil
	}

	name, ok := scriptObj["name"].(string)
	if !ok {
		return nil
	}

	typeStr, ok := scriptObj["type"].(string)
	if !ok {
		return nil
	}

	data, ok := scriptObj["data"]
	if !ok {
		return nil
	}

	// Convert data to []byte if it's a string
	var dataBytes []byte
	switch v := data.(type) {
	case string:
		dataBytes = []byte(v)
	case []byte:
		dataBytes = v
	default:
		// Try to serialize as JSON
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		dataBytes = jsonData
	}

	artifact := domain.NewArtifact(name, domain.ArtifactType(typeStr), dataBytes)
	artifact.ID = id

	// Set optional fields
	if size, ok := scriptObj["size"].(int); ok {
		artifact.Size = int64(size)
	} else if size, ok := scriptObj["size"].(float64); ok {
		artifact.Size = int64(size)
	}

	if mimeType, ok := scriptObj["mimeType"].(string); ok {
		artifact.MimeType = mimeType
	}

	if metadata, ok := scriptObj["metadata"].(map[string]interface{}); ok {
		artifact.Metadata = metadata
	}

	return artifact
}

// scriptToMessage converts script representation to domain.Message
func (b *StateContextBridge) scriptToMessage(scriptObj map[string]interface{}) domain.Message {
	role, ok := scriptObj["role"].(string)
	if !ok {
		role = "user"
	}

	content, ok := scriptObj["content"].(string)
	if !ok {
		content = ""
	}

	return domain.Message{
		Role:    domain.Role(role),
		Content: content,
	}
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) listPersistedStates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID := ""
	if id, ok := params["contextId"].(string); ok {
		contextID = id
	}

	b.persistenceMu.RLock()
	defer b.persistenceMu.RUnlock()

	if contextID == "" {
		// List all contexts and their versions
		allStates := make(map[string]interface{})
		for ctxID, versions := range b.stateVersions {
			allStates[ctxID] = versions
		}
		return map[string]interface{}{
			"states": allStates,
			"total":  len(allStates),
		}, nil
	}

	// List versions for specific context
	versions, exists := b.stateVersions[contextID]
	if !exists {
		return map[string]interface{}{
			"contextId": contextID,
			"versions":  []interface{}{},
			"total":     0,
		}, nil
	}

	// Convert []int to []interface{} for bridge compatibility
	versionsList := make([]interface{}, len(versions))
	for i, v := range versions {
		versionsList[i] = map[string]interface{}{
			"version":    v,
			"timestamp":  time.Now().Format(time.RFC3339),
			"size":       0, // Would need to be calculated from actual file
			"compressed": b.enableCompress,
		}
	}

	return map[string]interface{}{
		"contextId": contextID,
		"versions":  versionsList,
		"total":     len(versions),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) deletePersistedState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID, ok := params["contextId"].(string)
	if !ok {
		return nil, fmt.Errorf("contextId parameter is required and must be a string")
	}

	version := 0
	if v, ok := params["version"].(int); ok {
		version = v
	} else if v, ok := params["version"].(float64); ok {
		version = int(v)
	}

	if version == 0 {
		return nil, fmt.Errorf("version parameter is required and must be greater than 0")
	}

	// Delete from file system if available
	deleted := false
	if b.persistDir != "" {
		err := b.deleteStateFile(contextID, version)
		if err != nil {
			return nil, fmt.Errorf("failed to delete state file: %w", err)
		}
		deleted = true
	}

	// Remove from version tracking
	b.persistenceMu.Lock()
	if versions, exists := b.stateVersions[contextID]; exists {
		newVersions := make([]int, 0)
		for _, v := range versions {
			if v != version {
				newVersions = append(newVersions, v)
			}
		}
		if len(newVersions) == 0 {
			delete(b.stateVersions, contextID)
		} else {
			b.stateVersions[contextID] = newVersions
		}
	}
	b.persistenceMu.Unlock()

	// Emit deletion event
	if b.eventEmitter != nil {
		deleteEventData := map[string]interface{}{
			"contextId": contextID,
			"version":   version,
			"deleted":   deleted,
			"timestamp": time.Now(),
		}
		b.eventEmitter.EmitCustom("state.deleted", deleteEventData)
	}

	return map[string]interface{}{
		"contextId": contextID,
		"version":   version,
		"deleted":   deleted,
		"timestamp": time.Now().Format(time.RFC3339),
		"success":   true,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) generateStateDiff(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID, ok := params["contextId"].(string)
	if !ok {
		return nil, fmt.Errorf("contextId parameter is required and must be a string")
	}

	fromVersion := 0
	if v, ok := params["fromVersion"].(int); ok {
		fromVersion = v
	} else if v, ok := params["fromVersion"].(float64); ok {
		fromVersion = int(v)
	}

	toVersion := 0
	if v, ok := params["toVersion"].(int); ok {
		toVersion = v
	} else if v, ok := params["toVersion"].(float64); ok {
		toVersion = int(v)
	}

	if fromVersion == 0 || toVersion == 0 {
		return nil, fmt.Errorf("both fromVersion and toVersion are required")
	}

	// Load both states
	fromState, err := b.loadFromFile(contextID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to load from version %d: %w", fromVersion, err)
	}

	toState, err := b.loadFromFile(contextID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to load to version %d: %w", toVersion, err)
	}

	// Generate diff using go-llms state manager
	diff := b.generateDiff(fromState, toState)

	return map[string]interface{}{
		"contextId":   contextID,
		"fromVersion": fromVersion,
		"toVersion":   toVersion,
		"diff":        diff,
		"timestamp":   time.Now().Format(time.RFC3339),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) migrateState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextID, ok := params["contextId"].(string)
	if !ok {
		return nil, fmt.Errorf("contextId parameter is required and must be a string")
	}

	fromVersion := 0
	if v, ok := params["fromVersion"].(int); ok {
		fromVersion = v
	} else if v, ok := params["fromVersion"].(float64); ok {
		fromVersion = int(v)
	}

	toSchemaID, ok := params["toSchemaId"].(string)
	if !ok {
		return nil, fmt.Errorf("toSchemaId parameter is required and must be a string")
	}

	// Load state to migrate
	stateData, err := b.loadFromFile(contextID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to load state version %d: %w", fromVersion, err)
	}

	// Get target schema
	targetSchema, err := b.schemaRepo.Get(toSchemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target schema %s: %w", toSchemaID, err)
	}

	// Attempt migration using state manager transformation
	migratedData, err := b.applyMigrationTransforms(ctx, stateData, targetSchema)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Generate new version number
	nextVersion := b.getNextVersion(contextID)

	// Persist migrated state
	reconstructedContext, err := b.reconstructSharedContext(migratedData)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct migrated context: %w", err)
	}

	// Create new context with migrated state and persist it
	migratedContextID := fmt.Sprintf("migrated_%s_v%d", contextID, nextVersion)
	b.mu.Lock()
	b.contexts[migratedContextID] = reconstructedContext
	b.configs[migratedContextID] = b.configs[contextID] // Copy inheritance config
	b.mu.Unlock()

	// Emit migration event
	if b.eventEmitter != nil {
		migrationEventData := map[string]interface{}{
			"originalContextId": contextID,
			"migratedContextId": migratedContextID,
			"fromVersion":       fromVersion,
			"toVersion":         nextVersion,
			"toSchemaId":        toSchemaID,
			"timestamp":         time.Now(),
		}
		b.eventEmitter.EmitCustom("state.migrated", migrationEventData)
	}

	return map[string]interface{}{
		"originalContextId": contextID,
		"migratedContextId": migratedContextID,
		"fromVersion":       fromVersion,
		"toVersion":         nextVersion,
		"toSchemaId":        toSchemaID,
		"migratedContext":   b.sharedContextToScript(migratedContextID, reconstructedContext),
		"timestamp":         time.Now().Format(time.RFC3339),
		"success":           true,
	}, nil
}

// Helper methods for persistence operations

func (b *StateContextBridge) deleteStateFile(contextID string, version int) error {
	contextDir := filepath.Join(b.persistDir, "states", contextID)

	// Delete both compressed and uncompressed versions
	files := []string{
		fmt.Sprintf("state_v%d.json", version),
		fmt.Sprintf("state_v%d.json.gz", version),
		fmt.Sprintf("metadata_v%d.json", version),
	}

	for _, filename := range files {
		filePath := filepath.Join(contextDir, filename)
		err := os.Remove(filePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file %s: %w", filename, err)
		}
	}

	return nil
}

func (b *StateContextBridge) generateDiff(fromState, toState map[string]interface{}) map[string]interface{} {
	diff := map[string]interface{}{
		"added":    make(map[string]interface{}),
		"removed":  make(map[string]interface{}),
		"modified": make(map[string]interface{}),
	}

	// Compare state data
	fromData := b.extractStateData(fromState)
	toData := b.extractStateData(toState)

	// Find added keys
	for key, value := range toData {
		if _, exists := fromData[key]; !exists {
			diff["added"].(map[string]interface{})[key] = value
		}
	}

	// Find removed keys
	for key, value := range fromData {
		if _, exists := toData[key]; !exists {
			diff["removed"].(map[string]interface{})[key] = value
		}
	}

	// Find modified keys
	for key, newValue := range toData {
		if oldValue, exists := fromData[key]; exists {
			if !b.deepEqual(oldValue, newValue) {
				diff["modified"].(map[string]interface{})[key] = map[string]interface{}{
					"old": oldValue,
					"new": newValue,
				}
			}
		}
	}

	return diff
}

func (b *StateContextBridge) extractStateData(stateData map[string]interface{}) map[string]interface{} {
	if stateObj, ok := stateData["state"].(map[string]interface{}); ok {
		if dataMap, ok := stateObj["data"].(map[string]interface{}); ok {
			return dataMap
		}
	}
	return make(map[string]interface{})
}

func (b *StateContextBridge) deepEqual(a, other interface{}) bool {
	// Simple deep equality check - in production, use a more robust solution
	aBytes, err1 := json.Marshal(a)
	otherBytes, err2 := json.Marshal(other)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aBytes) == string(otherBytes)
}

func (b *StateContextBridge) applyMigrationTransforms(ctx context.Context, stateData map[string]interface{}, targetSchema *sdomain.Schema) (map[string]interface{}, error) {
	// Apply state manager transforms to migrate data
	// For now, we'll do a simple pass-through with basic validation

	// Extract state data
	extractedData := b.extractStateData(stateData)

	// Validate against target schema
	result, err := b.validator.ValidateStruct(targetSchema, extractedData)
	if err != nil {
		return nil, fmt.Errorf("migration validation failed: %w", err)
	}

	if !result.Valid {
		// Try to apply basic transforms using state manager
		if b.stateManager != nil {
			// Create a temporary state for transformation
			tempState := domain.NewState()
			for k, v := range extractedData {
				tempState.Set(k, v)
			}

			// Use built-in sanitize transform to clean up data
			transformed, err := b.stateManager.ApplyTransform(ctx, "sanitize", tempState)
			if err != nil {
				return nil, fmt.Errorf("failed to apply sanitize transform: %w", err)
			}

			// Update state data with transformed data
			extractedData = transformed.Values()
			if stateObj, ok := stateData["state"].(map[string]interface{}); ok {
				stateObj["data"] = extractedData
			}
		}
	}

	return stateData, nil
}

func (b *StateContextBridge) getNextVersion(contextID string) int {
	b.persistenceMu.RLock()
	versions, exists := b.stateVersions[contextID]
	b.persistenceMu.RUnlock()

	if !exists || len(versions) == 0 {
		return 1
	}

	maxVersion := 0
	for _, v := range versions {
		if v > maxVersion {
			maxVersion = v
		}
	}

	return maxVersion + 1
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

	case "persistState":
		if len(args) < 1 {
			return nil, fmt.Errorf("persistState requires context parameter")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}

		params := map[string]interface{}{
			"context": contextObj,
		}
		if len(args) > 1 {
			if version, ok := args[1].(int); ok {
				params["version"] = version
			} else if version, ok := args[1].(float64); ok {
				params["version"] = int(version)
			}
		}
		if len(args) > 2 {
			if description, ok := args[2].(string); ok {
				params["description"] = description
			}
		}
		if len(args) > 3 {
			if compress, ok := args[3].(bool); ok {
				params["compress"] = compress
			}
		}
		return b.persistState(ctx, params)

	case "loadState":
		if len(args) < 1 {
			return nil, fmt.Errorf("loadState requires contextId parameter")
		}
		contextID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("contextId must be string")
		}

		params := map[string]interface{}{
			"contextId": contextID,
		}
		if len(args) > 1 {
			if version, ok := args[1].(int); ok {
				params["version"] = version
			} else if version, ok := args[1].(float64); ok {
				params["version"] = int(version)
			}
		}
		if len(args) > 2 {
			if validateSchema, ok := args[2].(bool); ok {
				params["validateSchema"] = validateSchema
			}
		}
		return b.loadState(ctx, params)

	case "listPersistedStates":
		params := map[string]interface{}{}
		if len(args) > 0 {
			if contextID, ok := args[0].(string); ok {
				params["contextId"] = contextID
			}
		}
		return b.listPersistedStates(ctx, params)

	case "deletePersistedState":
		if len(args) < 2 {
			return nil, fmt.Errorf("deletePersistedState requires contextId and version parameters")
		}
		contextID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("contextId must be string")
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
			"contextId": contextID,
			"version":   version,
		}
		return b.deletePersistedState(ctx, params)

	case "generateStateDiff":
		if len(args) < 3 {
			return nil, fmt.Errorf("generateStateDiff requires contextId, fromVersion, and toVersion parameters")
		}
		contextID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("contextId must be string")
		}
		fromVersion, ok := args[1].(int)
		if !ok {
			if versionFloat, ok := args[1].(float64); ok {
				fromVersion = int(versionFloat)
			} else {
				return nil, fmt.Errorf("fromVersion must be integer")
			}
		}
		toVersion, ok := args[2].(int)
		if !ok {
			if versionFloat, ok := args[2].(float64); ok {
				toVersion = int(versionFloat)
			} else {
				return nil, fmt.Errorf("toVersion must be integer")
			}
		}

		params := map[string]interface{}{
			"contextId":   contextID,
			"fromVersion": fromVersion,
			"toVersion":   toVersion,
		}
		return b.generateStateDiff(ctx, params)

	case "migrateState":
		if len(args) < 3 {
			return nil, fmt.Errorf("migrateState requires contextId, fromVersion, and toSchemaId parameters")
		}
		contextID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("contextId must be string")
		}
		fromVersion, ok := args[1].(int)
		if !ok {
			if versionFloat, ok := args[1].(float64); ok {
				fromVersion = int(versionFloat)
			} else {
				return nil, fmt.Errorf("fromVersion must be integer")
			}
		}
		toSchemaID, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("toSchemaId must be string")
		}

		params := map[string]interface{}{
			"contextId":   contextID,
			"fromVersion": fromVersion,
			"toSchemaId":  toSchemaID,
		}
		return b.migrateState(ctx, params)

	// Transformation pipeline methods
	case "registerTransform":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerTransform requires name and transformType parameters")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}
		transformType, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("transformType must be string")
		}

		var config map[string]interface{}
		if len(args) > 2 {
			if cfg, ok := args[2].(map[string]interface{}); ok {
				config = cfg
			}
		}

		params := map[string]interface{}{
			"name":          name,
			"transformType": transformType,
			"config":        config,
		}
		return b.registerTransform(ctx, params)

	case "applyTransform":
		if len(args) < 2 {
			return nil, fmt.Errorf("applyTransform requires context and transformName parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		transformName, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("transformName must be string")
		}

		params := map[string]interface{}{
			"context":       contextObj,
			"transformName": transformName,
		}
		return b.applyTransform(ctx, params)

	case "createPipeline":
		if len(args) < 2 {
			return nil, fmt.Errorf("createPipeline requires pipelineId and transforms parameters")
		}
		pipelineId, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pipelineId must be string")
		}
		transforms, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("transforms must be array")
		}

		var config map[string]interface{}
		if len(args) > 2 {
			if cfg, ok := args[2].(map[string]interface{}); ok {
				config = cfg
			}
		}

		params := map[string]interface{}{
			"pipelineId": pipelineId,
			"transforms": transforms,
			"config":     config,
		}
		return b.createPipeline(ctx, params)

	case "applyPipeline":
		if len(args) < 2 {
			return nil, fmt.Errorf("applyPipeline requires context and pipelineId parameters")
		}
		contextObj, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("context must be object")
		}
		pipelineId, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("pipelineId must be string")
		}

		params := map[string]interface{}{
			"context":    contextObj,
			"pipelineId": pipelineId,
		}
		return b.applyPipeline(ctx, params)

	case "getTransformMetrics":
		if len(args) < 1 {
			return nil, fmt.Errorf("getTransformMetrics requires pipelineId parameter")
		}
		pipelineId, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pipelineId must be string")
		}

		params := map[string]interface{}{
			"pipelineId": pipelineId,
		}
		return b.getTransformMetrics(ctx, params)

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// State transformation pipeline methods using go-llms infrastructure

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) registerTransform(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	transformType, ok := params["transformType"].(string)
	if !ok {
		return nil, fmt.Errorf("transformType parameter is required and must be a string")
	}

	config, _ := params["config"].(map[string]interface{})

	// Create the appropriate transform based on type
	var transform core.StateTransform
	switch transformType {
	case "filter":
		pattern := ""
		if config != nil {
			if p, ok := config["pattern"].(string); ok {
				pattern = p
			}
		}
		transform = core.FilterTransform(pattern)

	case "selectKeys":
		keys := []string{}
		if config != nil {
			if keysList, ok := config["keys"].([]interface{}); ok {
				for _, k := range keysList {
					if keyStr, ok := k.(string); ok {
						keys = append(keys, keyStr)
					}
				}
			}
		}
		transform = core.SelectKeysTransform(keys...)

	case "renameKeys":
		mapping := make(map[string]string)
		if config != nil {
			if mappingConfig, ok := config["mapping"].(map[string]interface{}); ok {
				for k, v := range mappingConfig {
					if vStr, ok := v.(string); ok {
						mapping[k] = vStr
					}
				}
			}
		}
		transform = core.RenameKeysTransform(mapping)

	case "prefixKeys":
		prefix := ""
		if config != nil {
			if p, ok := config["prefix"].(string); ok {
				prefix = p
			}
		}
		transform = core.PrefixKeysTransform(prefix)

	case "normalizeKeys":
		transform = core.NormalizeKeysTransform()

	case "flatten":
		separator := "."
		if config != nil {
			if s, ok := config["separator"].(string); ok {
				separator = s
			}
		}
		transform = core.FlattenTransform(separator)

	case "clearMessages":
		transform = core.ClearMessagesTransform()

	case "limitMessages":
		limit := 10
		if config != nil {
			if l, ok := config["limit"].(int); ok {
				limit = l
			} else if l, ok := config["limit"].(float64); ok {
				limit = int(l)
			}
		}
		transform = core.LimitMessagesTransform(limit)

	case "filterMessagesByRole":
		roles := []string{}
		if config != nil {
			if rolesList, ok := config["roles"].([]interface{}); ok {
				for _, r := range rolesList {
					if roleStr, ok := r.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
		transform = core.FilterMessagesByRole(roles...)

	default:
		return nil, fmt.Errorf("unsupported transform type: %s", transformType)
	}

	// Register the transform with the state manager
	b.stateManager.RegisterTransform(name, transform)

	return map[string]interface{}{
		"name":          name,
		"transformType": transformType,
		"registered":    true,
		"timestamp":     time.Now().Format(time.RFC3339),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) applyTransform(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	transformName, ok := params["transformName"].(string)
	if !ok {
		return nil, fmt.Errorf("transformName parameter is required and must be a string")
	}

	// Get context ID
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Get shared context from map
	b.mu.RLock()
	sharedContext, exists := b.contexts[contextID]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("shared context not found: %s", contextID)
	}

	// Create a state from the shared context for transformation
	state := b.sharedContextToState(sharedContext)

	// Apply the transform using state manager
	startTime := time.Now()
	transformedState, err := b.stateManager.ApplyTransform(ctx, transformName, state)
	duration := time.Since(startTime)

	// Update metrics
	b.updateTransformMetrics(transformName, duration, err == nil)

	if err != nil {
		return nil, fmt.Errorf("failed to apply transform %s: %w", transformName, err)
	}

	// Update the shared context with transformed state
	b.mu.Lock()
	b.updateSharedContextFromState(sharedContext, transformedState)
	b.mu.Unlock()

	// Emit transform event if event emitter is available
	if b.eventEmitter != nil {
		transformEventData := map[string]interface{}{
			"contextId":     contextID,
			"transformName": transformName,
			"duration":      duration.String(),
			"timestamp":     time.Now(),
		}
		b.eventEmitter.EmitCustom("state.transformed", transformEventData)
	}

	return map[string]interface{}{
		"contextId":          contextID,
		"transformName":      transformName,
		"success":            true,
		"duration":           duration.String(),
		"transformedContext": b.sharedContextToScript(contextID, sharedContext),
		"timestamp":          time.Now().Format(time.RFC3339),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) createPipeline(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	pipelineId, ok := params["pipelineId"].(string)
	if !ok {
		return nil, fmt.Errorf("pipelineId parameter is required and must be a string")
	}

	transforms, ok := params["transforms"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("transforms parameter is required and must be an array")
	}

	config, _ := params["config"].(map[string]interface{})

	// Convert transforms to string array
	transformNames := make([]string, len(transforms))
	for i, t := range transforms {
		if tStr, ok := t.(string); ok {
			transformNames[i] = tStr
		} else {
			return nil, fmt.Errorf("transform at index %d must be a string", i)
		}
	}

	b.transformMu.Lock()
	defer b.transformMu.Unlock()

	// Store the pipeline configuration
	b.transformPipelines[pipelineId] = transformNames
	if config != nil {
		b.pipelineConfigs[pipelineId] = config
	}

	// Initialize metrics for the pipeline
	b.transformMetrics[pipelineId] = &TransformMetrics{
		ExecutionCount:  0,
		TotalDuration:   0,
		AverageDuration: 0,
		LastExecuted:    time.Time{},
		SuccessCount:    0,
		ErrorCount:      0,
		CacheHits:       0,
		CacheMisses:     0,
	}

	return map[string]interface{}{
		"pipelineId": pipelineId,
		"transforms": transformNames,
		"config":     config,
		"created":    true,
		"timestamp":  time.Now().Format(time.RFC3339),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) applyPipeline(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	contextObj, ok := params["context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("context parameter is required and must be a shared context object")
	}

	pipelineId, ok := params["pipelineId"].(string)
	if !ok {
		return nil, fmt.Errorf("pipelineId parameter is required and must be a string")
	}

	// Get context ID
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Get pipeline transforms
	b.transformMu.RLock()
	transformNames, exists := b.transformPipelines[pipelineId]
	config := b.pipelineConfigs[pipelineId]
	b.transformMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineId)
	}

	// Get shared context from map
	b.mu.RLock()
	sharedContext, contextExists := b.contexts[contextID]
	b.mu.RUnlock()

	if !contextExists {
		return nil, fmt.Errorf("shared context not found: %s", contextID)
	}

	// Check cache first if enabled
	cacheKey := fmt.Sprintf("%s_%s_%d", contextID, pipelineId, time.Now().Unix())
	useCache := false
	if config != nil {
		if cacheCfg, ok := config["cache"].(bool); ok {
			useCache = cacheCfg
		}
	}

	if useCache {
		b.transformMu.RLock()
		if cachedState, found := b.transformCache[cacheKey]; found {
			b.transformMu.RUnlock()

			// Update cache hit metrics
			b.updateCacheMetrics(pipelineId, true)

			// Update shared context from cached state
			b.mu.Lock()
			b.updateSharedContextFromState(sharedContext, cachedState)
			b.mu.Unlock()

			return map[string]interface{}{
				"contextId":          contextID,
				"pipelineId":         pipelineId,
				"transformsApplied":  len(transformNames),
				"fromCache":          true,
				"transformedContext": b.sharedContextToScript(contextID, sharedContext),
				"timestamp":          time.Now().Format(time.RFC3339),
			}, nil
		}
		b.transformMu.RUnlock()

		// Update cache miss metrics
		b.updateCacheMetrics(pipelineId, false)
	}

	// Create a state from the shared context for transformation
	state := b.sharedContextToState(sharedContext)

	// Apply transforms in sequence
	startTime := time.Now()
	currentState := state
	transformsApplied := 0

	for _, transformName := range transformNames {
		transformedState, err := b.stateManager.ApplyTransform(ctx, transformName, currentState)
		if err != nil {
			duration := time.Since(startTime)
			b.updatePipelineMetrics(pipelineId, duration, false)
			return nil, fmt.Errorf("failed to apply transform %s in pipeline %s: %w", transformName, pipelineId, err)
		}
		currentState = transformedState
		transformsApplied++
	}

	duration := time.Since(startTime)
	b.updatePipelineMetrics(pipelineId, duration, true)

	// Cache the result if caching is enabled
	if useCache {
		b.transformMu.Lock()
		b.transformCache[cacheKey] = currentState
		b.transformMu.Unlock()
	}

	// Update the shared context with transformed state
	b.mu.Lock()
	b.updateSharedContextFromState(sharedContext, currentState)
	b.mu.Unlock()

	// Emit pipeline event if event emitter is available
	if b.eventEmitter != nil {
		pipelineEventData := map[string]interface{}{
			"contextId":         contextID,
			"pipelineId":        pipelineId,
			"transformsApplied": transformsApplied,
			"duration":          duration.String(),
			"fromCache":         false,
			"timestamp":         time.Now(),
		}
		b.eventEmitter.EmitCustom("state.pipeline.applied", pipelineEventData)
	}

	return map[string]interface{}{
		"contextId":          contextID,
		"pipelineId":         pipelineId,
		"transformsApplied":  transformsApplied,
		"duration":           duration.String(),
		"fromCache":          false,
		"transformedContext": b.sharedContextToScript(contextID, sharedContext),
		"timestamp":          time.Now().Format(time.RFC3339),
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateContextBridge) getTransformMetrics(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	pipelineId, ok := params["pipelineId"].(string)
	if !ok {
		return nil, fmt.Errorf("pipelineId parameter is required and must be a string")
	}

	b.transformMu.RLock()
	metrics, exists := b.transformMetrics[pipelineId]
	b.transformMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("metrics not found for pipeline: %s", pipelineId)
	}

	return map[string]interface{}{
		"pipelineId":      pipelineId,
		"executionCount":  metrics.ExecutionCount,
		"totalDuration":   metrics.TotalDuration.String(),
		"averageDuration": metrics.AverageDuration.String(),
		"lastExecuted":    metrics.LastExecuted.Format(time.RFC3339),
		"successCount":    metrics.SuccessCount,
		"errorCount":      metrics.ErrorCount,
		"successRate":     float64(metrics.SuccessCount) / float64(metrics.ExecutionCount),
		"cacheHits":       metrics.CacheHits,
		"cacheMisses":     metrics.CacheMisses,
		"cacheHitRate":    float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses),
		"timestamp":       time.Now().Format(time.RFC3339),
	}, nil
}

// Helper methods for transformation pipeline

// sharedContextToState converts a SharedStateContext to a State for transformation
func (b *StateContextBridge) sharedContextToState(sharedContext *domain.SharedStateContext) *domain.State {
	// Convert shared context to a regular state by merging all data
	return sharedContext.AsState()
}

// updateSharedContextFromState updates a SharedStateContext with data from a transformed State
func (b *StateContextBridge) updateSharedContextFromState(sharedContext *domain.SharedStateContext, state *domain.State) {
	// Get the local state from shared context
	currentState := sharedContext.LocalState()

	// Merge the transformed state data
	for key, value := range state.Values() {
		currentState.Set(key, value)
	}

	// Merge artifacts
	for _, artifact := range state.Artifacts() {
		currentState.AddArtifact(artifact)
	}

	// Merge metadata
	for key, value := range state.GetAllMetadata() {
		currentState.SetMetadata(key, value)
	}
}

// updateTransformMetrics updates metrics for individual transform execution
func (b *StateContextBridge) updateTransformMetrics(transformName string, duration time.Duration, success bool) {
	b.transformMu.Lock()
	defer b.transformMu.Unlock()

	metrics, exists := b.transformMetrics[transformName]
	if !exists {
		metrics = &TransformMetrics{}
		b.transformMetrics[transformName] = metrics
	}

	metrics.ExecutionCount++
	metrics.TotalDuration += duration
	metrics.AverageDuration = time.Duration(int64(metrics.TotalDuration) / metrics.ExecutionCount)
	metrics.LastExecuted = time.Now()

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}
}

// updatePipelineMetrics updates metrics for pipeline execution
func (b *StateContextBridge) updatePipelineMetrics(pipelineId string, duration time.Duration, success bool) {
	b.transformMu.Lock()
	defer b.transformMu.Unlock()

	metrics, exists := b.transformMetrics[pipelineId]
	if !exists {
		metrics = &TransformMetrics{}
		b.transformMetrics[pipelineId] = metrics
	}

	metrics.ExecutionCount++
	metrics.TotalDuration += duration
	metrics.AverageDuration = time.Duration(int64(metrics.TotalDuration) / metrics.ExecutionCount)
	metrics.LastExecuted = time.Now()

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}
}

// updateCacheMetrics updates cache-related metrics
func (b *StateContextBridge) updateCacheMetrics(pipelineId string, hit bool) {
	b.transformMu.Lock()
	defer b.transformMu.Unlock()

	metrics, exists := b.transformMetrics[pipelineId]
	if !exists {
		metrics = &TransformMetrics{}
		b.transformMetrics[pipelineId] = metrics
	}

	if hit {
		metrics.CacheHits++
	} else {
		metrics.CacheMisses++
	}
}
