// ABOUTME: State Context Bridge implementation that exposes go-llms SharedStateContext to script engines
// ABOUTME: Provides parent-child state sharing with configurable inheritance for multi-agent systems

package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// StateContextBridge bridges go-llms SharedStateContext to script engines
type StateContextBridge struct {
	mu       sync.RWMutex
	contexts map[string]*domain.SharedStateContext
	configs  map[string]*inheritanceConfig // Track inheritance configs
	nextID   int
}

// inheritanceConfig tracks inheritance settings for a shared context
type inheritanceConfig struct {
	inheritMessages  bool
	inheritArtifacts bool
	inheritMetadata  bool
}

// NewStateContextBridge creates a new state context bridge
func NewStateContextBridge() (*StateContextBridge, error) {
	return &StateContextBridge{
		contexts: make(map[string]*domain.SharedStateContext),
		configs:  make(map[string]*inheritanceConfig),
		nextID:   1,
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

	// Convert script object to shared context
	sharedContext, err := b.scriptToSharedContext(contextObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert shared context: %w", err)
	}

	// Set in local state
	sharedContext.Set(key, value)

	// Update the script object
	b.updateScriptSharedContext(contextObj, sharedContext)

	return nil, nil
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
