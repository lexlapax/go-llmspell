// ABOUTME: State Manager Bridge implementation that exposes go-llms StateManager to script engines
// ABOUTME: Provides comprehensive state lifecycle, transforms, validation, and merging operations

package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// StateManagerBridge bridges go-llms StateManager to script engines
type StateManagerBridge struct {
	manager bridge.StateManager
}

// NewStateManagerBridge creates a new state manager bridge
func NewStateManagerBridge(manager bridge.StateManager) (*StateManagerBridge, error) {
	if manager == nil {
		return nil, fmt.Errorf("state manager cannot be nil")
	}

	return &StateManagerBridge{
		manager: manager,
	}, nil
}

// Name returns the bridge name
func (b *StateManagerBridge) Name() string {
	return "state_manager"
}

// Methods returns the methods exposed by this bridge
func (b *StateManagerBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{Name: "createState", Description: "Create a new state object", ReturnType: "State"},
		{Name: "saveState", Description: "Save a state object to persistence", ReturnType: "void"},
		{Name: "loadState", Description: "Load a state object by ID", ReturnType: "State"},
		{Name: "deleteState", Description: "Delete a state object by ID", ReturnType: "void"},
		{Name: "listStates", Description: "List all state IDs", ReturnType: "string[]"},
		{Name: "registerTransform", Description: "Register a state transform function", ReturnType: "void"},
		{Name: "applyTransform", Description: "Apply a transform to a state", ReturnType: "State"},
		{Name: "registerValidator", Description: "Register a state validator function", ReturnType: "void"},
		{Name: "validateState", Description: "Validate a state using a validator", ReturnType: "void"},
		{Name: "mergeStates", Description: "Merge multiple states using a strategy", ReturnType: "State"},
		{Name: "get", Description: "Get a value from state", ReturnType: "any"},
		{Name: "set", Description: "Set a value in state", ReturnType: "void"},
		{Name: "delete", Description: "Delete a key from state", ReturnType: "void"},
		{Name: "has", Description: "Check if state has a key", ReturnType: "boolean"},
		{Name: "keys", Description: "Get all keys from state", ReturnType: "string[]"},
		{Name: "values", Description: "Get all values from state", ReturnType: "any[]"},
		{Name: "setMetadata", Description: "Set metadata on state", ReturnType: "void"},
		{Name: "getMetadata", Description: "Get metadata from state", ReturnType: "any"},
		{Name: "getAllMetadata", Description: "Get all metadata from state", ReturnType: "object"},
		{Name: "addArtifact", Description: "Add artifact to state", ReturnType: "void"},
		{Name: "getArtifact", Description: "Get artifact from state", ReturnType: "Artifact"},
		{Name: "artifacts", Description: "Get all artifacts from state", ReturnType: "Artifact[]"},
		{Name: "addMessage", Description: "Add message to state", ReturnType: "void"},
		{Name: "messages", Description: "Get all messages from state", ReturnType: "Message[]"},
	}
}

// TypeMappings returns type mappings for this bridge
func (b *StateManagerBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"State": {
			GoType:     "State",
			ScriptType: "object",
		},
		"StateManager": {
			GoType:     "StateManager",
			ScriptType: "object",
		},
		"Artifact": {
			GoType:     "Artifact",
			ScriptType: "object",
		},
		"Message": {
			GoType:     "Message",
			ScriptType: "object",
		},
	}
}

// GetID returns the bridge ID
func (b *StateManagerBridge) GetID() string {
	return "state_manager"
}

// GetMetadata returns bridge metadata
func (b *StateManagerBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "State Manager Bridge",
		Version:     "1.0.0",
		Description: "Bridges go-llms StateManager to script engines",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge
func (b *StateManagerBridge) Initialize(ctx context.Context) error {
	b.registerBuiltinTransforms()
	return nil
}

// Cleanup cleans up bridge resources
func (b *StateManagerBridge) Cleanup(ctx context.Context) error {
	return nil
}

// IsInitialized returns whether the bridge is initialized
func (b *StateManagerBridge) IsInitialized() bool {
	return true
}

// RegisterWithEngine registers this bridge with a script engine
func (b *StateManagerBridge) RegisterWithEngine(scriptEngine engine.ScriptEngine) error {
	return scriptEngine.RegisterBridge(b)
}

// ValidateMethod validates a method call
func (b *StateManagerBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Basic validation - method exists
	for _, method := range b.Methods() {
		if method.Name == name {
			return nil
		}
	}
	return fmt.Errorf("method %s not found", name)
}

// RequiredPermissions returns required permissions
func (b *StateManagerBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "state_management",
			Actions:     []string{"read", "write"},
			Description: "Access to state management operations",
		},
	}
}

// State lifecycle operations

func (b *StateManagerBridge) createState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// go-llms doesn't have CreateState on StateManager, create directly
	state := domain.NewState()
	return b.stateToScript(state), nil
}

func (b *StateManagerBridge) saveState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	err = b.manager.SaveState(state)
	if err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return nil, nil
}

func (b *StateManagerBridge) loadState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id parameter is required and must be a string")
	}

	state, err := b.manager.LoadState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return b.stateToScript(state), nil
}

func (b *StateManagerBridge) deleteState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id parameter is required and must be a string")
	}

	err := b.manager.DeleteState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete state: %w", err)
	}

	return nil, nil
}

func (b *StateManagerBridge) listStates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	ids := b.manager.ListStates()

	// Convert to interface slice for script engines
	result := make([]interface{}, len(ids))
	for i, id := range ids {
		result[i] = id
	}

	return result, nil
}

// Transform operations

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) registerTransform(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	transformFunc, ok := params["transform"].(func(context.Context, *domain.State) (*domain.State, error))
	if !ok {
		return nil, fmt.Errorf("transform parameter is required and must be a function")
	}

	b.manager.RegisterTransform(name, transformFunc)
	return nil, nil
}

func (b *StateManagerBridge) applyTransform(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	transformedState, err := b.manager.ApplyTransform(ctx, name, state)
	if err != nil {
		return nil, fmt.Errorf("failed to apply transform: %w", err)
	}

	return b.stateToScript(transformedState), nil
}

// Validation operations

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) registerValidator(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	validatorFunc, ok := params["validator"].(func(*domain.State) error)
	if !ok {
		return nil, fmt.Errorf("validator parameter is required and must be a function")
	}

	b.manager.RegisterValidator(name, validatorFunc)
	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) validateState(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter is required and must be a string")
	}

	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	err = b.manager.ValidateState(name, state)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Merge operations

func (b *StateManagerBridge) mergeStates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	statesParam, ok := params["states"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("states parameter is required and must be an array")
	}

	strategyStr, ok := params["strategy"].(string)
	if !ok {
		return nil, fmt.Errorf("strategy parameter is required and must be a string")
	}

	// Convert strategy string to enum
	var strategy bridge.MergeStrategy
	switch strings.ToLower(strategyStr) {
	case "last":
		strategy = bridge.MergeStrategyLast
	case "merge_all":
		strategy = bridge.MergeStrategyMergeAll
	case "union":
		strategy = bridge.MergeStrategyUnion
	default:
		return nil, fmt.Errorf("invalid merge strategy: %s", strategyStr)
	}

	// Convert script states to Go states
	states := make([]bridge.State, len(statesParam))
	for i, stateParam := range statesParam {
		stateObj, ok := stateParam.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("state at index %d must be an object", i)
		}

		state, err := b.scriptToState(stateObj)
		if err != nil {
			return nil, fmt.Errorf("failed to convert state at index %d: %w", i, err)
		}

		states[i] = state
	}

	mergedState, err := b.manager.MergeStates(states, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to merge states: %w", err)
	}

	return b.stateToScript(mergedState), nil
}

// State data operations

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) get(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	value, exists := state.Get(key)
	return map[string]interface{}{
		"value":  value,
		"exists": exists,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) set(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	value := params["value"]

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	state.Set(key, value)

	// Update the script object with new state data
	b.updateScriptState(stateObj, state)

	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) delete(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	state.Delete(key)

	// Update the script object with new state data
	b.updateScriptState(stateObj, state)

	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) has(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	return state.Has(key), nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) keys(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	keys := state.Keys()

	// Convert to interface slice for script engines
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = key
	}

	return result, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) values(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	return state.Values(), nil
}

// Metadata operations

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) setMetadata(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	value := params["value"]

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	state.SetMetadata(key, value)

	// Update the script object with new state data
	b.updateScriptState(stateObj, state)

	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) getMetadata(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	key, ok := params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key parameter is required and must be a string")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	value, exists := state.GetMetadata(key)
	return map[string]interface{}{
		"value":  value,
		"exists": exists,
	}, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) getAllMetadata(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	return state.GetAllMetadata(), nil
}

// Artifact operations

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) addArtifact(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	artifactObj, ok := params["artifact"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("artifact parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	artifact, err := b.scriptToArtifact(artifactObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert artifact: %w", err)
	}

	// Add artifact to state
	state.AddArtifact(artifact)

	// Update the script object with new state data
	b.updateScriptState(stateObj, state)

	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) getArtifact(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id parameter is required and must be a string")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	artifact, exists := state.GetArtifact(id)
	if !exists {
		return nil, nil
	}

	return b.artifactToScript(artifact), nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) artifacts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	artifacts := state.Artifacts()
	result := make(map[string]interface{})
	for id, artifact := range artifacts {
		result[id] = b.artifactToScript(artifact)
	}

	return result, nil
}

// Message operations

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) addMessage(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	messageObj, ok := params["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("message parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	message, err := b.scriptToMessage(messageObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message: %w", err)
	}

	state.AddMessage(message)

	// Update the script object with new state data
	b.updateScriptState(stateObj, state)

	return nil, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) messages(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	stateObj, ok := params["state"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state parameter is required and must be an object")
	}

	state, err := b.scriptToState(stateObj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	messages := state.Messages()
	result := make([]interface{}, len(messages))
	for i, message := range messages {
		result[i] = b.messageToScript(message)
	}

	return result, nil
}

// Helper functions for type conversion

func (b *StateManagerBridge) stateToScript(state bridge.State) map[string]interface{} {
	return map[string]interface{}{
		"id":       state.ID(),
		"created":  state.Created().Format(time.RFC3339),
		"modified": state.Modified().Format(time.RFC3339),
		"version":  state.Version(),
		"parentID": state.ParentID(),
		"data":     state.Values(),
		"metadata": state.GetAllMetadata(),
		"__state":  state, // Store the actual state object for round-trip conversion
	}
}

func (b *StateManagerBridge) scriptToState(scriptObj map[string]interface{}) (bridge.State, error) {
	// First check if we have the actual state object stored
	if state, ok := scriptObj["__state"].(bridge.State); ok {
		return state, nil
	}

	// Otherwise try to reconstruct from the data
	// This is a simplified conversion - in a real implementation,
	// we would need to properly reconstruct the state object
	// For now, we'll return an error if we can't find the state
	return nil, fmt.Errorf("cannot convert script object to state: missing __state reference")
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) updateScriptState(scriptObj map[string]interface{}, state bridge.State) {
	// Update the script object to reflect state changes
	scriptObj["data"] = state.Values()
	scriptObj["metadata"] = state.GetAllMetadata()
	scriptObj["modified"] = state.Modified().Format(time.RFC3339)
	scriptObj["version"] = state.Version()
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) artifactToScript(artifact *domain.Artifact) map[string]interface{} {
	data, _ := artifact.Data()
	return map[string]interface{}{
		"id":       artifact.ID,
		"name":     artifact.Name,
		"type":     string(artifact.Type),
		"data":     data,
		"size":     artifact.Size,
		"mimeType": artifact.MimeType,
		"created":  artifact.Created.Format(time.RFC3339),
		"metadata": artifact.Metadata,
	}
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) scriptToArtifact(scriptObj map[string]interface{}) (*domain.Artifact, error) {
	// Extract artifact properties from script object
	id, ok := scriptObj["id"].(string)
	if !ok {
		return nil, fmt.Errorf("artifact missing required id field")
	}

	name, ok := scriptObj["name"].(string)
	if !ok {
		return nil, fmt.Errorf("artifact missing required name field")
	}

	typeStr, ok := scriptObj["type"].(string)
	if !ok {
		return nil, fmt.Errorf("artifact missing required type field")
	}

	// Get data if provided
	var data []byte
	if dataStr, ok := scriptObj["data"].(string); ok {
		data = []byte(dataStr)
	} else if dataBytes, ok := scriptObj["data"].([]byte); ok {
		data = dataBytes
	}

	// Create artifact
	artifact := domain.NewArtifact(name, domain.ArtifactType(typeStr), data)

	// Override ID to match script object
	// Note: go-llms Artifact ID field is exported, so we can set it
	artifact.ID = id

	// Set MIME type if provided
	if mimeType, ok := scriptObj["mimeType"].(string); ok {
		artifact.WithMimeType(mimeType)
	}

	// Set metadata if provided
	if metadata, ok := scriptObj["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			artifact.WithMetadata(k, v)
		}
	}

	return artifact, nil
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) messageToScript(message domain.Message) map[string]interface{} {
	return map[string]interface{}{
		"role":    string(message.Role),
		"content": message.Content,
	}
}

//nolint:unused // Will be used by script engines
func (b *StateManagerBridge) scriptToMessage(scriptObj map[string]interface{}) (domain.Message, error) {
	role, ok := scriptObj["role"].(string)
	if !ok {
		return domain.Message{}, fmt.Errorf("message role is required and must be a string")
	}

	content, ok := scriptObj["content"].(string)
	if !ok {
		return domain.Message{}, fmt.Errorf("message content is required and must be a string")
	}

	return domain.Message{
		Role:    domain.Role(role),
		Content: content,
	}, nil
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *StateManagerBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	switch name {
	case "createState":
		result, err := b.createState(ctx, nil)
		if err != nil {
			return nil, err
		}
		// Convert result to ScriptValue
		if result == nil {
			return engine.NewNilValue(), nil
		}
		if resultMap, ok := result.(map[string]interface{}); ok {
			scriptMap := make(map[string]engine.ScriptValue)
			for k, v := range resultMap {
				// Handle special fields
				if k == "__state" {
					// Preserve the state object as a custom value
					scriptMap[k] = engine.NewCustomValue("State", v)
				} else {
					// Simple conversion for other fields
					switch val := v.(type) {
					case string:
						scriptMap[k] = engine.NewStringValue(val)
					case float64:
						scriptMap[k] = engine.NewNumberValue(val)
					case bool:
						scriptMap[k] = engine.NewBoolValue(val)
					case map[string]interface{}:
						scriptMap[k] = engine.ConvertToScriptValue(val)
					default:
						scriptMap[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
					}
				}
			}
			return engine.NewObjectValue(scriptMap), nil
		}
		return engine.NewStringValue(fmt.Sprintf("%v", result)), nil

	case "saveState":
		if len(args) < 1 {
			return nil, fmt.Errorf("saveState requires state parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		_, err := b.saveState(ctx, map[string]interface{}{"state": stateObj})
		if err != nil {
			return nil, err
		}
		// For now, saveState typically returns nil on success
		return engine.NewNilValue(), nil

	case "loadState":
		if len(args) < 1 {
			return nil, fmt.Errorf("loadState requires id parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("id must be string")
		}
		id := args[0].(engine.StringValue).Value()
		result, err := b.loadState(ctx, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}
		// Convert state to ScriptValue - similar to createState
		if result == nil {
			return engine.NewNilValue(), nil
		}
		if resultMap, ok := result.(map[string]interface{}); ok {
			scriptMap := make(map[string]engine.ScriptValue)
			for k, v := range resultMap {
				// Handle special fields
				if k == "__state" {
					// Preserve the state object as a custom value
					scriptMap[k] = engine.NewCustomValue("State", v)
				} else {
					// Simple conversion for other fields
					switch val := v.(type) {
					case string:
						scriptMap[k] = engine.NewStringValue(val)
					case float64:
						scriptMap[k] = engine.NewNumberValue(val)
					case bool:
						scriptMap[k] = engine.NewBoolValue(val)
					case map[string]interface{}:
						scriptMap[k] = engine.ConvertToScriptValue(val)
					default:
						scriptMap[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
					}
				}
			}
			return engine.NewObjectValue(scriptMap), nil
		}
		return engine.NewStringValue(fmt.Sprintf("%v", result)), nil

	case "deleteState":
		if len(args) < 1 {
			return nil, fmt.Errorf("deleteState requires id parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("id must be string")
		}
		id := args[0].(engine.StringValue).Value()
		_, err := b.deleteState(ctx, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "listStates":
		result, err := b.listStates(ctx, nil)
		if err != nil {
			return nil, err
		}
		// Convert result array to ScriptValue
		if resultArray, ok := result.([]interface{}); ok {
			scriptArray := make([]engine.ScriptValue, len(resultArray))
			for i, v := range resultArray {
				switch val := v.(type) {
				case string:
					scriptArray[i] = engine.NewStringValue(val)
				default:
					scriptArray[i] = engine.NewStringValue(fmt.Sprintf("%v", v))
				}
			}
			return engine.NewArrayValue(scriptArray), nil
		}
		return engine.NewArrayValue([]engine.ScriptValue{}), nil

	case "applyTransform":
		if len(args) < 2 {
			return nil, fmt.Errorf("applyTransform requires name and state parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := make(map[string]interface{})
		for k, v := range args[1].(engine.ObjectValue).Fields() {
			stateObj[k] = v.ToGo()
		}
		result, err := b.applyTransform(ctx, map[string]interface{}{"name": name, "state": stateObj})
		if err != nil {
			return nil, err
		}
		// Convert transformed state to ScriptValue
		if result == nil {
			return engine.NewNilValue(), nil
		}
		if resultMap, ok := result.(map[string]interface{}); ok {
			scriptMap := make(map[string]engine.ScriptValue)
			for k, v := range resultMap {
				// Handle special fields
				if k == "__state" {
					// Preserve the state object as a custom value
					scriptMap[k] = engine.NewCustomValue("State", v)
				} else {
					// Simple conversion for other fields
					switch val := v.(type) {
					case string:
						scriptMap[k] = engine.NewStringValue(val)
					case float64:
						scriptMap[k] = engine.NewNumberValue(val)
					case bool:
						scriptMap[k] = engine.NewBoolValue(val)
					case map[string]interface{}:
						scriptMap[k] = engine.ConvertToScriptValue(val)
					default:
						scriptMap[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
					}
				}
			}
			return engine.NewObjectValue(scriptMap), nil
		}
		return engine.NewStringValue(fmt.Sprintf("%v", result)), nil

	case "mergeStates":
		if len(args) < 2 {
			return nil, fmt.Errorf("mergeStates requires states and strategy parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeArray {
			return nil, fmt.Errorf("states must be array")
		}
		states := make([]interface{}, 0)
		for _, v := range args[0].(engine.ArrayValue).Elements() {
			states = append(states, v.ToGo())
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("strategy must be string")
		}
		strategy := args[1].(engine.StringValue).Value()
		result, err := b.mergeStates(ctx, map[string]interface{}{"states": states, "strategy": strategy})
		if err != nil {
			return nil, err
		}
		// Convert merged state to ScriptValue
		if result == nil {
			return engine.NewNilValue(), nil
		}
		if resultMap, ok := result.(map[string]interface{}); ok {
			scriptMap := make(map[string]engine.ScriptValue)
			for k, v := range resultMap {
				// Handle special fields
				if k == "__state" {
					// Preserve the state object as a custom value
					scriptMap[k] = engine.NewCustomValue("State", v)
				} else {
					// Simple conversion for other fields
					switch val := v.(type) {
					case string:
						scriptMap[k] = engine.NewStringValue(val)
					case float64:
						scriptMap[k] = engine.NewNumberValue(val)
					case bool:
						scriptMap[k] = engine.NewBoolValue(val)
					case map[string]interface{}:
						scriptMap[k] = engine.ConvertToScriptValue(val)
					default:
						scriptMap[k] = engine.NewStringValue(fmt.Sprintf("%v", v))
					}
				}
			}
			return engine.NewObjectValue(scriptMap), nil
		}
		return engine.NewStringValue(fmt.Sprintf("%v", result)), nil

	case "get":
		if len(args) < 2 {
			return nil, fmt.Errorf("get requires state and key parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("key must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		key := args[1].(engine.StringValue).Value()
		result, err := b.get(ctx, map[string]interface{}{"state": stateObj, "key": key})
		if err != nil {
			return nil, err
		}
		// Convert result to ScriptValue
		if resultMap, ok := result.(map[string]interface{}); ok {
			scriptMap := make(map[string]engine.ScriptValue)
			for k, v := range resultMap {
				switch val := v.(type) {
				case string:
					scriptMap[k] = engine.NewStringValue(val)
				case bool:
					scriptMap[k] = engine.NewBoolValue(val)
				default:
					scriptMap[k] = engine.ConvertToScriptValue(v)
				}
			}
			return engine.NewObjectValue(scriptMap), nil
		}
		return engine.ConvertToScriptValue(result), nil

	case "set":
		if len(args) < 3 {
			return nil, fmt.Errorf("set requires state, key, and value parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("key must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		key := args[1].(engine.StringValue).Value()
		value := args[2].ToGo()
		_, err := b.set(ctx, map[string]interface{}{"state": stateObj, "key": key, "value": value})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "delete":
		if len(args) < 2 {
			return nil, fmt.Errorf("delete requires state and key parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("key must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		key := args[1].(engine.StringValue).Value()
		_, err := b.delete(ctx, map[string]interface{}{"state": stateObj, "key": key})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "has":
		if len(args) < 2 {
			return nil, fmt.Errorf("has requires state and key parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("key must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		key := args[1].(engine.StringValue).Value()
		result, err := b.has(ctx, map[string]interface{}{"state": stateObj, "key": key})
		if err != nil {
			return nil, err
		}
		if boolResult, ok := result.(bool); ok {
			return engine.NewBoolValue(boolResult), nil
		}
		return engine.NewBoolValue(false), nil

	case "keys":
		if len(args) < 1 {
			return nil, fmt.Errorf("keys requires state parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		result, err := b.keys(ctx, map[string]interface{}{"state": stateObj})
		if err != nil {
			return nil, err
		}
		if arrayResult, ok := result.([]interface{}); ok {
			scriptArray := make([]engine.ScriptValue, len(arrayResult))
			for i, v := range arrayResult {
				if strVal, ok := v.(string); ok {
					scriptArray[i] = engine.NewStringValue(strVal)
				} else {
					scriptArray[i] = engine.NewStringValue(fmt.Sprintf("%v", v))
				}
			}
			return engine.NewArrayValue(scriptArray), nil
		}
		return engine.NewArrayValue([]engine.ScriptValue{}), nil

	case "values":
		if len(args) < 1 {
			return nil, fmt.Errorf("values requires state parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		result, err := b.values(ctx, map[string]interface{}{"state": stateObj})
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil

	case "setMetadata":
		if len(args) < 3 {
			return nil, fmt.Errorf("setMetadata requires state, key, and value parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("key must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		key := args[1].(engine.StringValue).Value()
		value := args[2].ToGo()
		_, err := b.setMetadata(ctx, map[string]interface{}{"state": stateObj, "key": key, "value": value})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "getMetadata":
		if len(args) < 2 {
			return nil, fmt.Errorf("getMetadata requires state and key parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("key must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		key := args[1].(engine.StringValue).Value()
		result, err := b.getMetadata(ctx, map[string]interface{}{"state": stateObj, "key": key})
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil

	case "getAllMetadata":
		if len(args) < 1 {
			return nil, fmt.Errorf("getAllMetadata requires state parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		result, err := b.getAllMetadata(ctx, map[string]interface{}{"state": stateObj})
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil

	case "addArtifact":
		if len(args) < 2 {
			return nil, fmt.Errorf("addArtifact requires state and artifact parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return nil, fmt.Errorf("artifact must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		artifactObj := make(map[string]interface{})
		for k, v := range args[1].(engine.ObjectValue).Fields() {
			artifactObj[k] = v.ToGo()
		}
		_, err := b.addArtifact(ctx, map[string]interface{}{"state": stateObj, "artifact": artifactObj})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "getArtifact":
		if len(args) < 2 {
			return nil, fmt.Errorf("getArtifact requires state and id parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("id must be string")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		id := args[1].(engine.StringValue).Value()
		result, err := b.getArtifact(ctx, map[string]interface{}{"state": stateObj, "id": id})
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil

	case "artifacts":
		if len(args) < 1 {
			return nil, fmt.Errorf("artifacts requires state parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		result, err := b.artifacts(ctx, map[string]interface{}{"state": stateObj})
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil

	case "addMessage":
		if len(args) < 2 {
			return nil, fmt.Errorf("addMessage requires state and message parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return nil, fmt.Errorf("message must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		messageObj := make(map[string]interface{})
		for k, v := range args[1].(engine.ObjectValue).Fields() {
			messageObj[k] = v.ToGo()
		}
		_, err := b.addMessage(ctx, map[string]interface{}{"state": stateObj, "message": messageObj})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "messages":
		if len(args) < 1 {
			return nil, fmt.Errorf("messages requires state parameter")
		}
		if args[0] == nil || args[0].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		stateObj := b.extractStateObject(args[0].(engine.ObjectValue))
		result, err := b.messages(ctx, map[string]interface{}{"state": stateObj})
		if err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil

	case "registerTransform":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerTransform requires name and transform parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		transform := args[1].ToGo()
		_, err := b.registerTransform(ctx, map[string]interface{}{"name": name, "transform": transform})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "registerValidator":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerValidator requires name and validator parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		name := args[0].(engine.StringValue).Value()
		validator := args[1].ToGo()
		_, err := b.registerValidator(ctx, map[string]interface{}{"name": name, "validator": validator})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	case "validateState":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateState requires name and state parameters")
		}
		if args[0] == nil || args[0].Type() != engine.TypeString {
			return nil, fmt.Errorf("name must be string")
		}
		if args[1] == nil || args[1].Type() != engine.TypeObject {
			return nil, fmt.Errorf("state must be object")
		}
		name := args[0].(engine.StringValue).Value()
		stateObj := b.extractStateObject(args[1].(engine.ObjectValue))
		_, err := b.validateState(ctx, map[string]interface{}{"name": name, "state": stateObj})
		if err != nil {
			return nil, err
		}
		return engine.NewNilValue(), nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// extractStateObject safely extracts a state object from ScriptValue, preserving the __state field
func (b *StateManagerBridge) extractStateObject(obj engine.ObjectValue) map[string]interface{} {
	stateObj := make(map[string]interface{})
	for k, v := range obj.Fields() {
		if k == "__state" && v.Type() == engine.TypeCustom {
			// Preserve the actual state object
			stateObj[k] = v.ToGo()
		} else {
			stateObj[k] = v.ToGo()
		}
	}
	return stateObj
}

func (b *StateManagerBridge) registerBuiltinTransforms() {
	// Register built-in filter transform
	b.manager.RegisterTransform("filter", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()
		// Filter implementation would go here
		return newState, nil
	})

	// Register built-in flatten transform
	b.manager.RegisterTransform("flatten", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()
		// Flatten implementation would go here
		return newState, nil
	})

	// Register built-in sanitize transform
	b.manager.RegisterTransform("sanitize", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()
		// Sanitize implementation would go here
		return newState, nil
	})
}
