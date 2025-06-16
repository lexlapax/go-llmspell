// ABOUTME: Guardrails bridge for go-llms safety system and content filtering
// ABOUTME: Provides script-accessible guardrail validation and behavioral constraints

package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	// go-llms imports for guardrails functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"

	// Internal bridge imports
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// GuardrailsBridge provides script access to go-llms safety system
type GuardrailsBridge struct {
	initialized   bool
	guardrails    map[string]domain.Guardrail
	chains        map[string]*domain.GuardrailChain
	asyncChannels map[string]<-chan error
	mu            sync.RWMutex
}

// NewGuardrailsBridge creates a new guardrails bridge
func NewGuardrailsBridge() *GuardrailsBridge {
	return &GuardrailsBridge{
		guardrails:    make(map[string]domain.Guardrail),
		chains:        make(map[string]*domain.GuardrailChain),
		asyncChannels: make(map[string]<-chan error),
	}
}

// GetID returns the bridge identifier
func (gb *GuardrailsBridge) GetID() string {
	return "guardrails"
}

// GetMetadata returns bridge metadata
func (gb *GuardrailsBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:         "guardrails",
		Version:      "v1.0.0",
		Description:  "Bridge for go-llms safety system with content filtering and behavioral constraints",
		Author:       "go-llmspell",
		License:      "MIT",
		Dependencies: []string{"github.com/lexlapax/go-llms/pkg/agent/domain"},
	}
}

// Initialize sets up the guardrails bridge
func (gb *GuardrailsBridge) Initialize(ctx context.Context) error {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	gb.initialized = true
	return nil
}

// Cleanup performs bridge cleanup
func (gb *GuardrailsBridge) Cleanup(ctx context.Context) error {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	// Clear all stored data
	gb.guardrails = make(map[string]domain.Guardrail)
	gb.chains = make(map[string]*domain.GuardrailChain)
	gb.asyncChannels = make(map[string]<-chan error)
	gb.initialized = false

	return nil
}

// IsInitialized returns initialization status
func (gb *GuardrailsBridge) IsInitialized() bool {
	gb.mu.RLock()
	defer gb.mu.RUnlock()
	return gb.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (gb *GuardrailsBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(gb)
}

// Methods returns available bridge methods
func (gb *GuardrailsBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "createGuardrailFunc",
			Description: "Create a guardrail from a validation function",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Guardrail name"},
				{Name: "type", Type: "string", Required: true, Description: "Guardrail type: 'input', 'output', 'both'"},
				{Name: "validationFunc", Type: "function", Required: true, Description: "Validation function"},
			},
			ReturnType: "object",
			Examples:   []string{"createGuardrailFunc('content_check', 'input', function(state) { return true; })"},
		},
		{
			Name:        "createGuardrailChain",
			Description: "Create a chain of guardrails",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Chain name"},
				{Name: "type", Type: "string", Required: true, Description: "Chain type: 'input', 'output', 'both'"},
				{Name: "failFast", Type: "boolean", Required: true, Description: "Stop on first failure"},
			},
			ReturnType: "object",
			Examples:   []string{"createGuardrailChain('safety_chain', 'both', true)"},
		},
		{
			Name:        "addGuardrailToChain",
			Description: "Add a guardrail to a chain",
			Parameters: []engine.ParameterInfo{
				{Name: "chainID", Type: "string", Required: true, Description: "Chain identifier"},
				{Name: "guardrailID", Type: "string", Required: true, Description: "Guardrail identifier"},
			},
			ReturnType: "void",
			Examples:   []string{"addGuardrailToChain(chainID, guardrailID)"},
		},
		{
			Name:        "validateGuardrail",
			Description: "Validate state against a guardrail",
			Parameters: []engine.ParameterInfo{
				{Name: "guardrailID", Type: "string", Required: true, Description: "Guardrail identifier"},
				{Name: "state", Type: "object", Required: true, Description: "State to validate"},
			},
			ReturnType: "void",
			Examples:   []string{"validateGuardrail(guardrailID, {key: 'value'})"},
		},
		{
			Name:        "validateGuardrailAsync",
			Description: "Validate state asynchronously",
			Parameters: []engine.ParameterInfo{
				{Name: "guardrailID", Type: "string", Required: true, Description: "Guardrail identifier"},
				{Name: "state", Type: "object", Required: true, Description: "State to validate"},
				{Name: "timeoutSeconds", Type: "number", Required: true, Description: "Timeout in seconds"},
			},
			ReturnType: "object",
			Examples:   []string{"validateGuardrailAsync(guardrailID, state, 5.0)"},
		},
		{
			Name:        "validateChain",
			Description: "Validate state against a guardrail chain",
			Parameters: []engine.ParameterInfo{
				{Name: "chainID", Type: "string", Required: true, Description: "Chain identifier"},
				{Name: "state", Type: "object", Required: true, Description: "State to validate"},
			},
			ReturnType: "void",
			Examples:   []string{"validateChain(chainID, {key: 'value'})"},
		},
		{
			Name:        "createRequiredKeysGuardrail",
			Description: "Create guardrail that requires specific keys",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Guardrail name"},
				{Name: "keys", Type: "array", Required: true, Description: "Required keys"},
			},
			ReturnType: "object",
			Examples:   []string{"createRequiredKeysGuardrail('required_fields', ['name', 'email'])"},
		},
		{
			Name:        "createContentModerationGuardrail",
			Description: "Create guardrail for content moderation",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Guardrail name"},
				{Name: "prohibitedWords", Type: "array", Required: true, Description: "List of prohibited words"},
			},
			ReturnType: "object",
			Examples:   []string{"createContentModerationGuardrail('content_filter', ['spam', 'inappropriate'])"},
		},
		{
			Name:        "createMessageCountGuardrail",
			Description: "Create guardrail that limits message count",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Guardrail name"},
				{Name: "maxMessages", Type: "number", Required: true, Description: "Maximum message count"},
			},
			ReturnType: "object",
			Examples:   []string{"createMessageCountGuardrail('message_limit', 100)"},
		},
		{
			Name:        "createMaxStateSizeGuardrail",
			Description: "Create guardrail that limits state size",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Guardrail name"},
				{Name: "maxBytes", Type: "number", Required: true, Description: "Maximum size in bytes"},
			},
			ReturnType: "object",
			Examples:   []string{"createMaxStateSizeGuardrail('size_limit', 1048576)"},
		},
	}
}

// ValidateMethod validates method calls
func (gb *GuardrailsBridge) ValidateMethod(name string, args []interface{}) error {
	if !gb.IsInitialized() {
		return fmt.Errorf("guardrails bridge not initialized")
	}

	methods := gb.Methods()
	for _, method := range methods {
		if method.Name == name {
			requiredCount := 0
			for _, param := range method.Parameters {
				if param.Required {
					requiredCount++
				}
			}
			if len(args) < requiredCount {
				return fmt.Errorf("method %s requires at least %d arguments, got %d", name, requiredCount, len(args))
			}
			return nil
		}
	}
	return fmt.Errorf("unknown method: %s", name)
}

// TypeMappings returns type conversion mappings
func (gb *GuardrailsBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"guardrail": {
			GoType:     "domain.Guardrail",
			ScriptType: "object",
			Converter:  "guardrailConverter",
			Metadata:   map[string]interface{}{"description": "Guardrail validation interface"},
		},
		"guardrail_chain": {
			GoType:     "*domain.GuardrailChain",
			ScriptType: "object",
			Converter:  "guardrailChainConverter",
			Metadata:   map[string]interface{}{"description": "Chain of guardrails"},
		},
		"guardrail_type": {
			GoType:     "domain.GuardrailType",
			ScriptType: "string",
			Converter:  "guardrailTypeConverter",
			Metadata:   map[string]interface{}{"description": "Guardrail application type"},
		},
	}
}

// RequiredPermissions returns required permissions
func (gb *GuardrailsBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "guardrails.validation",
			Actions:     []string{"create", "validate"},
			Description: "Create and validate guardrails",
		},
		{
			Type:        engine.PermissionProcess,
			Resource:    "guardrails.chains",
			Actions:     []string{"create", "modify"},
			Description: "Create and modify guardrail chains",
		},
	}
}

// Bridge method implementations

// createGuardrailFunc creates a guardrail from a validation function
func (gb *GuardrailsBridge) createGuardrailFunc(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("createGuardrailFunc", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail name must be a string")
	}

	typeStr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail type must be a string")
	}

	// Convert string to GuardrailType
	var guardType domain.GuardrailType
	switch typeStr {
	case "input":
		guardType = domain.GuardrailTypeInput
	case "output":
		guardType = domain.GuardrailTypeOutput
	case "both":
		guardType = domain.GuardrailTypeBoth
	default:
		return nil, fmt.Errorf("invalid guardrail type: %s", typeStr)
	}

	// Extract validation function (this would be adapted based on script engine)
	validationFunc, ok := args[2].(func(interface{}) bool)
	if !ok {
		return nil, fmt.Errorf("validation function must be a callable function")
	}

	// Create the guardrail function wrapper
	guardrailFunc := func(ctx context.Context, state *domain.State) error {
		// Convert state to script-friendly format
		stateData, err := gb.stateToMap(state)
		if err != nil {
			return fmt.Errorf("failed to convert state: %w", err)
		}

		// Call validation function
		if !validationFunc(stateData) {
			return fmt.Errorf("guardrail validation failed")
		}

		return nil
	}

	// Create the guardrail
	guardrail := domain.NewGuardrailFunc(name, guardType, guardrailFunc)

	// Store it
	guardrailID := fmt.Sprintf("guardrail-%s-%d", name, time.Now().UnixNano())
	gb.mu.Lock()
	gb.guardrails[guardrailID] = guardrail
	gb.mu.Unlock()

	return map[string]interface{}{
		"id":      guardrailID,
		"name":    name,
		"type":    typeStr,
		"created": time.Now(),
	}, nil
}

// createGuardrailChain creates a new guardrail chain
func (gb *GuardrailsBridge) createGuardrailChain(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("createGuardrailChain", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("chain name must be a string")
	}

	typeStr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("chain type must be a string")
	}

	failFast, ok := args[2].(bool)
	if !ok {
		return nil, fmt.Errorf("fail fast must be a boolean")
	}

	// Convert string to GuardrailType
	var guardType domain.GuardrailType
	switch typeStr {
	case "input":
		guardType = domain.GuardrailTypeInput
	case "output":
		guardType = domain.GuardrailTypeOutput
	case "both":
		guardType = domain.GuardrailTypeBoth
	default:
		return nil, fmt.Errorf("invalid guardrail type: %s", typeStr)
	}

	// Create the chain
	chain := domain.NewGuardrailChain(name, guardType, failFast)

	// Store it
	chainID := fmt.Sprintf("chain-%s-%d", name, time.Now().UnixNano())
	gb.mu.Lock()
	gb.chains[chainID] = chain
	gb.mu.Unlock()

	return map[string]interface{}{
		"id":        chainID,
		"name":      name,
		"type":      typeStr,
		"fail_fast": failFast,
		"created":   time.Now(),
	}, nil
}

// addGuardrailToChain adds a guardrail to a chain
func (gb *GuardrailsBridge) addGuardrailToChain(ctx context.Context, args []interface{}) error {
	if err := gb.ValidateMethod("addGuardrailToChain", args); err != nil {
		return err
	}

	chainID, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("chain ID must be a string")
	}

	guardrailID, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("guardrail ID must be a string")
	}

	gb.mu.Lock()
	defer gb.mu.Unlock()

	chain, exists := gb.chains[chainID]
	if !exists {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	guardrail, exists := gb.guardrails[guardrailID]
	if !exists {
		return fmt.Errorf("guardrail not found: %s", guardrailID)
	}

	chain.Add(guardrail)

	return nil
}

// validateGuardrail validates state against a guardrail
func (gb *GuardrailsBridge) validateGuardrail(ctx context.Context, args []interface{}) error {
	if err := gb.ValidateMethod("validateGuardrail", args); err != nil {
		return err
	}

	guardrailID, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("guardrail ID must be a string")
	}

	stateData, ok := args[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("state must be an object")
	}

	gb.mu.RLock()
	guardrail, exists := gb.guardrails[guardrailID]
	gb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("guardrail not found: %s", guardrailID)
	}

	// Convert map to domain.State
	state, err := gb.mapToState(stateData)
	if err != nil {
		return fmt.Errorf("failed to convert state: %w", err)
	}

	// Validate
	return guardrail.Validate(ctx, state)
}

// validateGuardrailAsync validates state asynchronously
func (gb *GuardrailsBridge) validateGuardrailAsync(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("validateGuardrailAsync", args); err != nil {
		return nil, err
	}

	guardrailID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail ID must be a string")
	}

	stateData, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("state must be an object")
	}

	timeoutSeconds, ok := args[2].(float64)
	if !ok {
		return nil, fmt.Errorf("timeout must be a number")
	}

	gb.mu.RLock()
	guardrail, exists := gb.guardrails[guardrailID]
	gb.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("guardrail not found: %s", guardrailID)
	}

	// Convert map to domain.State
	state, err := gb.mapToState(stateData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert state: %w", err)
	}

	timeout := time.Duration(timeoutSeconds * float64(time.Second))
	errCh := guardrail.ValidateAsync(ctx, state, timeout)

	// Store the channel for later retrieval
	channelID := fmt.Sprintf("async-%s-%d", guardrailID, time.Now().UnixNano())
	gb.mu.Lock()
	gb.asyncChannels[channelID] = errCh
	gb.mu.Unlock()

	return map[string]interface{}{
		"channel_id": channelID,
		"timeout":    timeoutSeconds,
		"started":    time.Now(),
	}, nil
}

// validateChain validates state against a guardrail chain
// Used via bridge reflection system, not directly called in Go code
//
//nolint:unused // Bridge method called via reflection
func (gb *GuardrailsBridge) validateChain(ctx context.Context, args []interface{}) error {
	if err := gb.ValidateMethod("validateChain", args); err != nil {
		return err
	}

	chainID, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("chain ID must be a string")
	}

	stateData, ok := args[1].(map[string]interface{})
	if !ok {
		return fmt.Errorf("state must be an object")
	}

	gb.mu.RLock()
	chain, exists := gb.chains[chainID]
	gb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	// Convert map to domain.State
	state, err := gb.mapToState(stateData)
	if err != nil {
		return fmt.Errorf("failed to convert state: %w", err)
	}

	// Validate
	return chain.Validate(ctx, state)
}

// Built-in guardrail creation methods

// createRequiredKeysGuardrail creates a guardrail that requires specific keys
func (gb *GuardrailsBridge) createRequiredKeysGuardrail(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("createRequiredKeysGuardrail", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail name must be a string")
	}

	keysInterface, ok := args[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("keys must be an array")
	}

	// Convert to string slice
	keys := make([]string, len(keysInterface))
	for i, key := range keysInterface {
		keyStr, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("key %d must be a string", i)
		}
		keys[i] = keyStr
	}

	// Create the guardrail
	guardrail := domain.RequiredKeysGuardrail(name, keys...)

	// Store it
	guardrailID := fmt.Sprintf("required-keys-%s-%d", name, time.Now().UnixNano())
	gb.mu.Lock()
	gb.guardrails[guardrailID] = guardrail
	gb.mu.Unlock()

	return map[string]interface{}{
		"id":           guardrailID,
		"name":         name,
		"type":         "input",
		"builtin_type": "required_keys",
		"keys":         keys,
		"created":      time.Now(),
	}, nil
}

// createContentModerationGuardrail creates a content moderation guardrail
func (gb *GuardrailsBridge) createContentModerationGuardrail(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("createContentModerationGuardrail", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail name must be a string")
	}

	wordsInterface, ok := args[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("prohibited words must be an array")
	}

	// Convert to string slice
	words := make([]string, len(wordsInterface))
	for i, word := range wordsInterface {
		wordStr, ok := word.(string)
		if !ok {
			return nil, fmt.Errorf("word %d must be a string", i)
		}
		words[i] = wordStr
	}

	// Create the guardrail
	guardrail := domain.ContentModerationGuardrail(name, words)

	// Store it
	guardrailID := fmt.Sprintf("content-mod-%s-%d", name, time.Now().UnixNano())
	gb.mu.Lock()
	gb.guardrails[guardrailID] = guardrail
	gb.mu.Unlock()

	return map[string]interface{}{
		"id":               guardrailID,
		"name":             name,
		"type":             "both",
		"builtin_type":     "content_moderation",
		"prohibited_words": words,
		"created":          time.Now(),
	}, nil
}

// createMessageCountGuardrail creates a message count guardrail
func (gb *GuardrailsBridge) createMessageCountGuardrail(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("createMessageCountGuardrail", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail name must be a string")
	}

	maxMessagesFloat, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("max messages must be a number")
	}

	maxMessages := int(maxMessagesFloat)

	// Create the guardrail
	guardrail := domain.MessageCountGuardrail(name, maxMessages)

	// Store it
	guardrailID := fmt.Sprintf("msg-count-%s-%d", name, time.Now().UnixNano())
	gb.mu.Lock()
	gb.guardrails[guardrailID] = guardrail
	gb.mu.Unlock()

	return map[string]interface{}{
		"id":           guardrailID,
		"name":         name,
		"type":         "both",
		"builtin_type": "message_count",
		"max_messages": maxMessages,
		"created":      time.Now(),
	}, nil
}

// createMaxStateSizeGuardrail creates a max state size guardrail
func (gb *GuardrailsBridge) createMaxStateSizeGuardrail(ctx context.Context, args []interface{}) (interface{}, error) {
	if err := gb.ValidateMethod("createMaxStateSizeGuardrail", args); err != nil {
		return nil, err
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("guardrail name must be a string")
	}

	maxBytesFloat, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("max bytes must be a number")
	}

	maxBytes := int64(maxBytesFloat)

	// Create the guardrail
	guardrail := domain.MaxStateSizeGuardrail(name, maxBytes)

	// Store it
	guardrailID := fmt.Sprintf("state-size-%s-%d", name, time.Now().UnixNano())
	gb.mu.Lock()
	gb.guardrails[guardrailID] = guardrail
	gb.mu.Unlock()

	return map[string]interface{}{
		"id":           guardrailID,
		"name":         name,
		"type":         "both",
		"builtin_type": "max_state_size",
		"max_bytes":    maxBytes,
		"created":      time.Now(),
	}, nil
}

// Helper methods for state conversion

// stateToMap converts a domain.State to a script-friendly map
func (gb *GuardrailsBridge) stateToMap(state *domain.State) (map[string]interface{}, error) {
	if state == nil {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})

	// Add state values
	for key, value := range state.Values() {
		result[key] = value
	}

	// Add messages
	messages := make([]map[string]interface{}, len(state.Messages()))
	for i, msg := range state.Messages() {
		messages[i] = map[string]interface{}{
			"role":    string(msg.Role),
			"content": msg.Content,
		}
	}
	result["_messages"] = messages

	// Add artifacts
	artifacts := make(map[string]interface{})
	for name, artifact := range state.Artifacts() {
		// Get artifact data
		data, err := artifact.Data()
		if err != nil {
			// If we can't get data, just include metadata
			artifacts[name] = map[string]interface{}{
				"type":      string(artifact.Type),
				"mime_type": artifact.MimeType,
				"size":      artifact.Size,
				"error":     err.Error(),
			}
		} else {
			artifacts[name] = map[string]interface{}{
				"type":      string(artifact.Type),
				"mime_type": artifact.MimeType,
				"size":      artifact.Size,
				"content":   data,
			}
		}
	}
	result["_artifacts"] = artifacts

	return result, nil
}

// mapToState converts a script map to a domain.State
func (gb *GuardrailsBridge) mapToState(data map[string]interface{}) (*domain.State, error) {
	state := domain.NewState()

	// Add regular values (skip internal fields)
	for key, value := range data {
		if key[0] != '_' {
			state.Set(key, value)
		}
	}

	// Add messages if present
	if messagesInterface, exists := data["_messages"]; exists {
		if messagesList, ok := messagesInterface.([]interface{}); ok {
			for _, msgInterface := range messagesList {
				if msgMap, ok := msgInterface.(map[string]interface{}); ok {
					if role, hasRole := msgMap["role"].(string); hasRole {
						if content, hasContent := msgMap["content"].(string); hasContent {
							msg := domain.Message{
								Role:    domain.Role(role),
								Content: content,
							}
							state.AddMessage(msg)
						}
					}
				}
			}
		}
	}

	return state, nil
}
