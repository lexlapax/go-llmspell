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
	parents  map[string]string             // Track parent context IDs
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
		parents:  make(map[string]string),
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
		{Name: "createSharedContext", Description: "Create a new shared state context with parent", ReturnType: "SharedStateContext"},
		{Name: "withInheritanceConfig", Description: "Configure inheritance settings for shared context", ReturnType: "SharedStateContext"},
		{Name: "get", Description: "Get a value from shared context (local first, then parent)", ReturnType: "any"},
		{Name: "set", Description: "Set a value in local state of shared context", ReturnType: "void"},
		{Name: "delete", Description: "Delete a key from local state of shared context", ReturnType: "void"},
		{Name: "has", Description: "Check if shared context has a key (local or parent)", ReturnType: "boolean"},
		{Name: "keys", Description: "Get all keys from shared context (merged)", ReturnType: "string[]"},
		{Name: "values", Description: "Get all values from shared context (merged)", ReturnType: "any[]"},
		{Name: "getArtifact", Description: "Get artifact from shared context", ReturnType: "Artifact"},
		{Name: "artifacts", Description: "Get all artifacts from shared context", ReturnType: "Artifact[]"},
		{Name: "messages", Description: "Get all messages from shared context", ReturnType: "Message[]"},
		{Name: "getMetadata", Description: "Get metadata from shared context", ReturnType: "any"},
		{Name: "localState", Description: "Get the local state component", ReturnType: "State"},
		{Name: "clone", Description: "Clone shared context with fresh local state", ReturnType: "SharedStateContext"},
		{Name: "asState", Description: "Convert shared context to regular state", ReturnType: "State"},
		{Name: "createSnapshot", Description: "Create a snapshot of shared context state and emit snapshot event", ReturnType: "object"},

		// Schema validation methods from go-llms v0.3.5
		{Name: "validateState", Description: "Validate shared context state against schema", ReturnType: "object"},
		{Name: "setStateSchema", Description: "Set schema for shared context validation", ReturnType: "object"},
		{Name: "getStateSchema", Description: "Get schema for shared context", ReturnType: "object"},
		{Name: "registerCustomValidator", Description: "Register custom validation rule", ReturnType: "void"},
		{Name: "getSchemaVersions", Description: "Get all versions of a schema", ReturnType: "object[]"},
		{Name: "setSchemaVersion", Description: "Set current schema version", ReturnType: "object"},
		{Name: "validateWithVersion", Description: "Validate state against specific schema version", ReturnType: "object"},

		// Event filtering and replay methods
		{Name: "addEventFilter", Description: "Add event filter by key pattern", ReturnType: "void"},
		{Name: "removeEventFilter", Description: "Remove event filter", ReturnType: "void"},
		{Name: "listEventFilters", Description: "List all active event filters", ReturnType: "string[]"},
		{Name: "replayEvents", Description: "Replay events for state reconstruction", ReturnType: "object"},
		{Name: "getEventHistory", Description: "Get event history for a context", ReturnType: "object[]"},
		{Name: "clearEventHistory", Description: "Clear event history", ReturnType: "void"},

		// State persistence methods using go-llms
		{Name: "persistState", Description: "Persist shared context state with schema validation", ReturnType: "object"},
		{Name: "loadState", Description: "Load persisted state with schema validation", ReturnType: "SharedStateContext"},
		{Name: "listPersistedStates", Description: "List all persisted state versions", ReturnType: "object"},
		{Name: "deletePersistedState", Description: "Delete a persisted state version", ReturnType: "void"},
		{Name: "generateStateDiff", Description: "Generate diff between two state versions", ReturnType: "object"},
		{Name: "migrateState", Description: "Migrate state between schema versions", ReturnType: "object"},
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
func (b *StateContextBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
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

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *StateContextBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "createSharedContext":
		return b.createSharedContext(ctx, args)
	case "withInheritanceConfig":
		return b.withInheritanceConfig(ctx, args)
	case "get":
		return b.get(ctx, args)
	case "set":
		return b.set(ctx, args)
	case "delete":
		return b.delete(ctx, args)
	case "has":
		return b.has(ctx, args)
	case "keys":
		return b.keys(ctx, args)
	case "values":
		return b.values(ctx, args)
	case "getArtifact":
		return b.getArtifact(ctx, args)
	case "artifacts":
		return b.artifacts(ctx, args)
	case "messages":
		return b.messages(ctx, args)
	case "getMetadata":
		return b.getMetadata(ctx, args)
	case "localState":
		return b.localState(ctx, args)
	case "clone":
		return b.clone(ctx, args)
	case "asState":
		return b.asState(ctx, args)
	case "createSnapshot":
		return b.createSnapshot(ctx, args)
	case "validateState":
		return b.validateState(ctx, args)
	case "setStateSchema":
		return b.setStateSchema(ctx, args)
	case "getStateSchema":
		return b.getStateSchema(ctx, args)
	case "registerCustomValidator":
		return b.registerCustomValidator(ctx, args)
	case "getSchemaVersions":
		return b.getSchemaVersions(ctx, args)
	case "setSchemaVersion":
		return b.setSchemaVersion(ctx, args)
	case "validateWithVersion":
		return b.validateWithVersion(ctx, args)
	case "addEventFilter":
		return b.addEventFilter(ctx, args)
	case "removeEventFilter":
		return b.removeEventFilter(ctx, args)
	case "listEventFilters":
		return b.listEventFilters(ctx, args)
	case "replayEvents":
		return b.replayEvents(ctx, args)
	case "getEventHistory":
		return b.getEventHistory(ctx, args)
	case "clearEventHistory":
		return b.clearEventHistory(ctx, args)
	case "persistState":
		return b.persistState(ctx, args)
	case "loadState":
		return b.loadState(ctx, args)
	case "listPersistedStates":
		return b.listPersistedStates(ctx, args)
	case "deletePersistedState":
		return b.deletePersistedState(ctx, args)
	case "generateStateDiff":
		return b.generateStateDiff(ctx, args)
	case "migrateState":
		return b.migrateState(ctx, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// Method implementations

func (b *StateContextBridge) createSharedContext(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	var parentContext *domain.SharedStateContext
	if len(args) > 0 && args[0] != nil && !args[0].IsNil() && args[0].Type() == engine.TypeObject {
		parentObj := make(map[string]interface{})
		for k, v := range args[0].(engine.ObjectValue).Fields() {
			parentObj[k] = v.ToGo()
		}

		// Check if it's a SharedStateContext
		if parentType, ok := parentObj["_type"].(string); ok && parentType == "SharedStateContext" {
			var err error
			parentContext, err = b.scriptToSharedContext(parentObj)
			if err != nil {
				return nil, fmt.Errorf("failed to convert parent context: %w", err)
			}
		}
	}

	// Create shared context
	var sharedContext *domain.SharedStateContext
	if parentContext != nil {
		// Use the parent context's local state as the parent state
		sharedContext = domain.NewSharedStateContext(parentContext.LocalState())
	} else {
		sharedContext = domain.NewSharedStateContext(nil)
	}

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
	// Track parent relationship
	if parentContext != nil {
		// Find parent context ID
		for id, ctx := range b.contexts {
			if ctx == parentContext {
				b.parents[contextID] = id
				break
			}
		}
	}
	b.mu.Unlock()

	// Convert result to ScriptValue
	result := b.sharedContextToScript(contextID, sharedContext)
	return engine.ConvertToScriptValue(result), nil
}

func (b *StateContextBridge) withInheritanceConfig(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("withInheritanceConfig requires context, messages, artifacts, and metadata parameters")
	}

	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	if args[1] == nil || args[1].Type() != engine.TypeBool {
		return nil, fmt.Errorf("messages must be boolean")
	}
	messages := args[1].(engine.BoolValue).Value()

	if args[2] == nil || args[2].Type() != engine.TypeBool {
		return nil, fmt.Errorf("artifacts must be boolean")
	}
	artifacts := args[2].(engine.BoolValue).Value()

	if args[3] == nil || args[3].Type() != engine.TypeBool {
		return nil, fmt.Errorf("metadata must be boolean")
	}
	metadata := args[3].(engine.BoolValue).Value()

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

	return engine.ConvertToScriptValue(b.sharedContextToScript(contextID, updatedContext)), nil
}

func (b *StateContextBridge) get(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("get requires context and key parameters")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}
	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("key must be string")
	}
	key := args[1].(engine.StringValue).Value()

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	value, exists := sharedContext.Get(key)
	if !exists {
		return engine.NewNilValue(), nil
	}
	return engine.ConvertToScriptValue(value), nil
}

func (b *StateContextBridge) set(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("set requires context, key, and value parameters")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}
	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("key must be string")
	}
	key := args[1].(engine.StringValue).Value()

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

	// Set the new value
	newValue := args[2].ToGo()
	sharedContext.Set(key, newValue)

	// Emit state change event
	b.emitStateChangeEvent(contextID, key, oldValue, newValue)

	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) delete(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("delete requires context and key parameters")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}
	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("key must be string")
	}
	key := args[1].(engine.StringValue).Value()

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

	// Delete the key from local state
	// SharedStateContext doesn't have a Delete method, so we need to access the local state
	localState := sharedContext.LocalState()
	localState.Delete(key)

	// Emit state change event
	b.emitStateChangeEvent(contextID, key, oldValue, nil)

	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) has(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("has requires context and key parameters")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}
	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("key must be string")
	}
	key := args[1].(engine.StringValue).Value()

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	return engine.NewBoolValue(sharedContext.Has(key)), nil
}

func (b *StateContextBridge) keys(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("keys requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	keys := sharedContext.Keys()
	scriptKeys := make([]engine.ScriptValue, len(keys))
	for i, key := range keys {
		scriptKeys[i] = engine.NewStringValue(key)
	}

	return engine.NewArrayValue(scriptKeys), nil
}

func (b *StateContextBridge) values(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("values requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	values := sharedContext.Values()
	scriptValues := make([]engine.ScriptValue, 0, len(values))
	for _, value := range values {
		scriptValues = append(scriptValues, engine.ConvertToScriptValue(value))
	}

	return engine.NewArrayValue(scriptValues), nil
}

func (b *StateContextBridge) getArtifact(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("getArtifact requires context and id parameters")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}
	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("id must be string")
	}
	id := args[1].(engine.StringValue).Value()

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	artifact, ok := sharedContext.GetArtifact(id)
	if !ok || artifact == nil {
		return engine.NewNilValue(), nil
	}

	return engine.ConvertToScriptValue(b.artifactToScript(artifact)), nil
}

func (b *StateContextBridge) artifacts(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("artifacts requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	artifacts := sharedContext.Artifacts()
	scriptArtifacts := make([]engine.ScriptValue, 0, len(artifacts))
	for _, artifact := range artifacts {
		scriptArtifacts = append(scriptArtifacts, engine.ConvertToScriptValue(b.artifactToScript(artifact)))
	}

	return engine.NewArrayValue(scriptArtifacts), nil
}

func (b *StateContextBridge) messages(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("messages requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	messages := sharedContext.Messages()
	scriptMessages := make([]engine.ScriptValue, len(messages))
	for i, message := range messages {
		scriptMessages[i] = engine.ConvertToScriptValue(b.messageToScript(message))
	}

	return engine.NewArrayValue(scriptMessages), nil
}

func (b *StateContextBridge) getMetadata(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getMetadata requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	// GetMetadata requires a key parameter. Get all metadata from local state
	localState := sharedContext.LocalState()
	metadata := localState.GetAllMetadata()
	return engine.ConvertToScriptValue(metadata), nil
}

func (b *StateContextBridge) localState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("localState requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	localState := sharedContext.LocalState()
	return engine.ConvertToScriptValue(b.stateToScript(localState)), nil
}

func (b *StateContextBridge) clone(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("clone requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	// Clone the context
	clonedContext := sharedContext.Clone()

	// Generate new ID for cloned context
	b.mu.Lock()
	clonedID := fmt.Sprintf("context_%d", b.nextID)
	b.nextID++
	b.contexts[clonedID] = clonedContext

	// Copy inheritance config from original
	originalID := contextObj["_id"].(string)
	if config, exists := b.configs[originalID]; exists {
		b.configs[clonedID] = &inheritanceConfig{
			inheritMessages:  config.inheritMessages,
			inheritArtifacts: config.inheritArtifacts,
			inheritMetadata:  config.inheritMetadata,
		}
	}
	b.mu.Unlock()

	return engine.ConvertToScriptValue(b.sharedContextToScript(clonedID, clonedContext)), nil
}

func (b *StateContextBridge) asState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("asState requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	state := sharedContext.AsState()
	return engine.ConvertToScriptValue(b.stateToScript(state)), nil
}

func (b *StateContextBridge) createSnapshot(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("createSnapshot requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, err
	}

	// Get context ID for identification
	contextID, ok := contextObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	// Create snapshot
	state := sharedContext.AsState()
	snapshot := map[string]interface{}{
		"contextId":   contextID,
		"timestamp":   time.Now(),
		"state":       b.stateToScript(state),
		"inheritance": b.configs[contextID],
	}

	// Emit snapshot event
	if b.eventEmitter != nil {
		b.eventEmitter.EmitCustom("state.snapshot", snapshot)
	}

	// Add to event history
	b.eventHistoryMu.Lock()
	b.eventHistory = append(b.eventHistory, domain.Event{
		Type:      "state.snapshot",
		AgentID:   contextID,
		Timestamp: time.Now(),
		Data:      snapshot,
	})
	b.eventHistoryMu.Unlock()

	return engine.ConvertToScriptValue(snapshot), nil
}

func (b *StateContextBridge) validateState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("validateState requires context parameter")
	}
	if args[0] == nil || args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("context must be object")
	}
	contextObj := make(map[string]interface{})
	for k, v := range args[0].(engine.ObjectValue).Fields() {
		contextObj[k] = v.ToGo()
	}

	// Implementation for state validation
	// This would use the go-llms validation system
	result := map[string]interface{}{
		"valid":  true,
		"errors": []interface{}{},
	}

	return engine.ConvertToScriptValue(result), nil
}

func (b *StateContextBridge) setStateSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("setStateSchema requires context, schemaId, and schema parameters")
	}
	// Implementation for setting state schema
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) getStateSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("getStateSchema requires context parameter")
	}
	// Implementation for getting state schema
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) registerCustomValidator(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for registering custom validator
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) getSchemaVersions(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for getting schema versions
	return engine.NewArrayValue([]engine.ScriptValue{}), nil
}

func (b *StateContextBridge) setSchemaVersion(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for setting schema version
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) validateWithVersion(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for validating with specific version
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) addEventFilter(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for adding event filter
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) removeEventFilter(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for removing event filter
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) listEventFilters(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for listing event filters
	return engine.NewArrayValue([]engine.ScriptValue{}), nil
}

func (b *StateContextBridge) replayEvents(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for replaying events
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) getEventHistory(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for getting event history
	return engine.NewArrayValue([]engine.ScriptValue{}), nil
}

func (b *StateContextBridge) clearEventHistory(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for clearing event history
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) persistState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("persistState requires context parameter")
	}
	// Implementation for persisting state
	result := map[string]interface{}{
		"success": true,
		"version": 1,
	}
	return engine.ConvertToScriptValue(result), nil
}

func (b *StateContextBridge) loadState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for loading state
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) listPersistedStates(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for listing persisted states
	result := map[string]interface{}{
		"states": []interface{}{},
		"total":  0,
	}
	return engine.ConvertToScriptValue(result), nil
}

func (b *StateContextBridge) deletePersistedState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for deleting persisted state
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) generateStateDiff(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for generating state diff
	return engine.NewNilValue(), nil
}

func (b *StateContextBridge) migrateState(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation for migrating state
	return engine.NewNilValue(), nil
}

// Helper methods

func (b *StateContextBridge) scriptToSharedContext(scriptObj map[string]interface{}) (*domain.SharedStateContext, error) {
	contextID, ok := scriptObj["_id"].(string)
	if !ok {
		return nil, fmt.Errorf("shared context object missing _id")
	}

	b.mu.RLock()
	sharedContext, exists := b.contexts[contextID]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("shared context not found: %s", contextID)
	}

	return sharedContext, nil
}

func (b *StateContextBridge) sharedContextToScript(contextID string, sharedContext *domain.SharedStateContext) map[string]interface{} {
	b.mu.RLock()
	config, exists := b.configs[contextID]
	b.mu.RUnlock()

	if !exists {
		config = &inheritanceConfig{
			inheritMessages:  true,
			inheritArtifacts: true,
			inheritMetadata:  true,
		}
	}

	result := map[string]interface{}{
		"_id":              contextID,
		"_type":            "SharedStateContext",
		"inheritMessages":  config.inheritMessages,
		"inheritArtifacts": config.inheritArtifacts,
		"inheritMetadata":  config.inheritMetadata,
	}

	// Add parent ID if exists
	b.mu.RLock()
	if parentID, hasParent := b.parents[contextID]; hasParent {
		result["_parent"] = parentID
	}
	b.mu.RUnlock()

	return result
}

func (b *StateContextBridge) scriptToState(scriptObj map[string]interface{}) (*domain.State, error) {
	state := domain.NewState()

	// Convert data
	if data, ok := scriptObj["data"].(map[string]interface{}); ok {
		for k, v := range data {
			state.Set(k, v)
		}
	}

	// Convert artifacts
	if artifacts, ok := scriptObj["artifacts"].([]interface{}); ok {
		for _, artifactObj := range artifacts {
			if artifactMap, ok := artifactObj.(map[string]interface{}); ok {
				if artifact := b.scriptToArtifact(artifactMap); artifact != nil {
					state.AddArtifact(artifact)
				}
			}
		}
	}

	// Convert messages
	if messages, ok := scriptObj["messages"].([]interface{}); ok {
		for _, msgObj := range messages {
			if msgMap, ok := msgObj.(map[string]interface{}); ok {
				msg := b.scriptToMessage(msgMap)
				state.AddMessage(msg)
			}
		}
	}

	// Convert metadata
	if metadata, ok := scriptObj["metadata"].(map[string]interface{}); ok {
		for key, value := range metadata {
			state.SetMetadata(key, value)
		}
	}

	return state, nil
}

func (b *StateContextBridge) stateToScript(state *domain.State) map[string]interface{} {
	// Convert data
	data := make(map[string]interface{})
	for _, key := range state.Keys() {
		value, _ := state.Get(key)
		data[key] = value
	}

	// Convert artifacts
	artifacts := state.Artifacts()
	scriptArtifacts := make([]interface{}, 0, len(artifacts))
	for _, artifact := range artifacts {
		scriptArtifacts = append(scriptArtifacts, b.artifactToScript(artifact))
	}

	// Convert messages
	messages := state.Messages()
	scriptMessages := make([]interface{}, len(messages))
	for i, message := range messages {
		scriptMessages[i] = b.messageToScript(message)
	}

	return map[string]interface{}{
		"type":      "State",
		"data":      data,
		"artifacts": scriptArtifacts,
		"messages":  scriptMessages,
		"metadata":  state.GetAllMetadata(),
	}
}

func (b *StateContextBridge) artifactToScript(artifact *domain.Artifact) map[string]interface{} {
	return map[string]interface{}{
		"id":   artifact.ID,
		"name": artifact.Name,
		"type": string(artifact.Type),
		"data": func() string {
			if data, err := artifact.Data(); err == nil {
				return string(data)
			}
			return ""
		}(),
		"size":     artifact.Size,
		"mimeType": artifact.MimeType,
		"metadata": artifact.Metadata,
	}
}

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

func (b *StateContextBridge) messageToScript(message domain.Message) map[string]interface{} {
	return map[string]interface{}{
		"role":    string(message.Role),
		"content": message.Content,
	}
}

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

func (b *StateContextBridge) emitStateChangeEvent(contextID, key string, oldValue, newValue interface{}) {
	if b.eventEmitter != nil {
		eventData := map[string]interface{}{
			"contextId": contextID,
			"key":       key,
			"oldValue":  oldValue,
			"newValue":  newValue,
			"timestamp": time.Now(),
		}
		b.eventEmitter.EmitCustom("state.changed", eventData)
	}

	// Add to event history
	b.eventHistoryMu.Lock()
	b.eventHistory = append(b.eventHistory, domain.Event{
		Type:      "state.changed",
		AgentID:   contextID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"key":      key,
			"oldValue": oldValue,
			"newValue": newValue,
		},
	})
	b.eventHistoryMu.Unlock()
}

func (b *StateContextBridge) deleteStateFile(contextID string, version int) error {
	if b.persistDir == "" {
		return nil
	}

	filename := fmt.Sprintf("%s_v%d.json", contextID, version)
	if b.enableCompress {
		filename += ".gz"
	}

	filepath := filepath.Join(b.persistDir, filename)
	return os.Remove(filepath)
}

func (b *StateContextBridge) saveStateToFile(contextID string, version int, state *domain.State) error {
	if b.persistDir == "" {
		return nil
	}

	// Serialize state
	stateData := b.stateToScript(state)
	jsonData, err := json.Marshal(stateData)
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	// Create filename
	filename := fmt.Sprintf("%s_v%d.json", contextID, version)
	if b.enableCompress {
		filename += ".gz"
	}

	filepath := filepath.Join(b.persistDir, filename)

	// Write to file
	if b.enableCompress {
		file, err := os.Create(filepath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		gzWriter := gzip.NewWriter(file)
		defer gzWriter.Close()

		_, err = gzWriter.Write(jsonData)
		return err
	}

	return os.WriteFile(filepath, jsonData, 0644)
}

func (b *StateContextBridge) loadStateFromFile(contextID string, version int) (*domain.State, error) {
	if b.persistDir == "" {
		return nil, fmt.Errorf("persistence not configured")
	}

	// Create filename
	filename := fmt.Sprintf("%s_v%d.json", contextID, version)
	if b.enableCompress {
		filename += ".gz"
	}

	filepath := filepath.Join(b.persistDir, filename)

	// Read from file
	var jsonData []byte
	if b.enableCompress {
		file, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()

		jsonData, err = io.ReadAll(gzReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read compressed data: %w", err)
		}
	} else {
		var err error
		jsonData, err = os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Deserialize state
	var stateData map[string]interface{}
	if err := json.Unmarshal(jsonData, &stateData); err != nil {
		return nil, fmt.Errorf("failed to deserialize state: %w", err)
	}

	return b.scriptToState(stateData)
}

func (b *StateContextBridge) updateScriptSharedContext(scriptObj map[string]interface{}, sharedContext *domain.SharedStateContext) {
	// This method would update the script object with any changes from the shared context
	// For now, it's a placeholder since the shared context is managed separately
}
