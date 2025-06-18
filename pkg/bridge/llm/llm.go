// ABOUTME: LLM bridge provides access to language model providers through go-llms interfaces.
// ABOUTME: Wraps go-llms Provider interface for script engine access without reimplementation.

package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// LLMBridge provides script access to language model functionality via go-llms.
type LLMBridge struct {
	mu             sync.RWMutex
	providers      map[string]bridge.Provider
	activeProvider string
	initialized    bool

	// Fallback chain configuration
	fallbackChain []string // Ordered list of provider names for fallback

	// Performance monitoring
	metrics map[string]*ProviderMetrics
}

// ProviderMetrics tracks performance metrics for each provider
type ProviderMetrics struct {
	TotalRequests   int64
	SuccessfulCalls int64
	FailedCalls     int64
	TotalLatency    time.Duration
	AverageLatency  time.Duration
	LastError       error
	LastErrorTime   time.Time
}

// NewLLMBridge creates a new LLM bridge.
func NewLLMBridge() *LLMBridge {
	return &LLMBridge{
		providers: make(map[string]bridge.Provider),
		metrics:   make(map[string]*ProviderMetrics),
	}
}

// GetID returns the bridge ID.
func (b *LLMBridge) GetID() string {
	return "llm"
}

// GetMetadata returns bridge metadata.
func (b *LLMBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "llm",
		Version:     "1.0.0",
		Description: "Language model provider bridge for text generation",
		Author:      "go-llmspell",
		License:     "MIT",
		Dependencies: []string{
			"github.com/lexlapax/go-llms/pkg/llm/domain",
		},
	}
}

// Initialize initializes the bridge.
func (b *LLMBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup performs cleanup operations.
func (b *LLMBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.providers = make(map[string]bridge.Provider)
	b.activeProvider = ""
	b.initialized = false

	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *LLMBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *LLMBridge) RegisterWithEngine(e engine.ScriptEngine) error {
	return e.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *LLMBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Provider management
		{
			Name:        "setProvider",
			Description: "Set the active LLM provider",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Provider name"},
				{Name: "config", Type: "object", Required: false, Description: "Provider configuration"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.setProvider("openai", {model: "gpt-4"})`,
			},
		},
		{
			Name:        "getProvider",
			Description: "Get information about the active provider",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples: []string{
				`llm.getProvider()`,
			},
		},
		{
			Name:        "listProviders",
			Description: "List all registered providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples: []string{
				`llm.listProviders()`,
			},
		},

		// Text generation
		{
			Name:        "generate",
			Description: "Generate text using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Required: true, Description: "Input prompt"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "string",
			Examples: []string{
				`llm.generate("Tell me a joke")`,
				`llm.generate("Explain quantum computing", {temperature: 0.7, max_tokens: 500})`,
			},
		},
		{
			Name:        "generateMessage",
			Description: "Generate a message response",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Required: true, Description: "Array of message objects"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.generateMessage([{role: "user", content: "Hello"}])`,
			},
		},
		{
			Name:        "stream",
			Description: "Stream text generation",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Required: true, Description: "Input prompt"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.stream("Write a story")`,
			},
		},

		// Structured generation
		{
			Name:        "generateWithSchema",
			Description: "Generate structured output with schema validation",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Required: true, Description: "Input prompt"},
				{Name: "schema", Type: "string", Required: true, Description: "Schema name"},
				{Name: "options", Type: "object", Required: false, Description: "Generation options"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.generateWithSchema("List 3 colors", "color_list")`,
			},
		},

		// Schema management
		{
			Name:        "addResponseSchema",
			Description: "Add a response schema for structured generation",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Schema name"},
				{Name: "schema", Type: "object", Required: true, Description: "JSON schema"},
			},
			ReturnType: "void",
			Examples: []string{
				`llm.addResponseSchema("person", {type: "object", properties: {name: {type: "string"}}})`,
			},
		},
		{
			Name:        "getResponseSchema",
			Description: "Get a registered response schema",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Schema name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.getResponseSchema("person")`,
			},
		},
		{
			Name:        "listResponseSchemas",
			Description: "List all registered schemas",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples: []string{
				`llm.listResponseSchemas()`,
			},
		},
		{
			Name:        "validateWithSchema",
			Description: "Validate data against a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "object", Required: true, Description: "Data to validate"},
				{Name: "schema", Type: "string", Required: true, Description: "Schema name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.validateWithSchema({name: "John"}, "person")`,
			},
		},

		// Provider capabilities
		{
			Name:        "getCapabilities",
			Description: "Get capabilities of the active provider",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "object",
			Examples: []string{
				`llm.getCapabilities()`,
			},
		},
		{
			Name:        "getModelInfo",
			Description: "Get information about a model",
			Parameters: []engine.ParameterInfo{
				{Name: "model", Type: "string", Required: true, Description: "Model name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.getModelInfo("gpt-4")`,
			},
		},
		{
			Name:        "listModels",
			Description: "List available models",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples: []string{
				`llm.listModels()`,
			},
		},
		{
			Name:        "checkCapability",
			Description: "Check if provider supports a capability",
			Parameters: []engine.ParameterInfo{
				{Name: "capability", Type: "string", Required: true, Description: "Capability name"},
			},
			ReturnType: "boolean",
			Examples: []string{
				`llm.checkCapability("streaming")`,
			},
		},

		// Streaming
		{
			Name:        "streamMessage",
			Description: "Stream a message response",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Required: true, Description: "Array of messages"},
				{Name: "options", Type: "object", Required: false, Description: "Stream options"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.streamMessage([{role: "user", content: "Hello"}])`,
			},
		},
		{
			Name:        "readStream",
			Description: "Read from an active stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamId", Type: "string", Required: true, Description: "Stream ID"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.readStream("stream-123")`,
			},
		},
		{
			Name:        "closeStream",
			Description: "Close an active stream",
			Parameters: []engine.ParameterInfo{
				{Name: "streamId", Type: "string", Required: true, Description: "Stream ID"},
			},
			ReturnType: "void",
			Examples: []string{
				`llm.closeStream("stream-123")`,
			},
		},

		// Fallback configuration
		{
			Name:        "setFallbackChain",
			Description: "Set provider fallback chain",
			Parameters: []engine.ParameterInfo{
				{Name: "providers", Type: "array", Required: true, Description: "Ordered list of provider names"},
			},
			ReturnType: "void",
			Examples: []string{
				`llm.setFallbackChain(["primary", "secondary", "tertiary"])`,
			},
		},
		{
			Name:        "getFallbackChain",
			Description: "Get current fallback chain",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
			Examples: []string{
				`llm.getFallbackChain()`,
			},
		},

		// Metrics
		{
			Name:        "getProviderMetrics",
			Description: "Get metrics for a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.getProviderMetrics("openai")`,
			},
		},
		{
			Name:        "resetProviderMetrics",
			Description: "Reset metrics for a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "void",
			Examples: []string{
				`llm.resetProviderMetrics("openai")`,
			},
		},

		// Schema generation
		{
			Name:        "generateSchemaFromExample",
			Description: "Generate a schema from example data",
			Parameters: []engine.ParameterInfo{
				{Name: "example", Type: "object", Required: true, Description: "Example data"},
				{Name: "name", Type: "string", Required: true, Description: "Schema name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.generateSchemaFromExample({name: "John", age: 30}, "person")`,
			},
		},

		// Provider information
		{
			Name:        "getProviderInfo",
			Description: "Get detailed provider information",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.getProviderInfo("openai")`,
			},
		},
		{
			Name:        "testProviderConnection",
			Description: "Test connection to a provider",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Required: true, Description: "Provider name"},
			},
			ReturnType: "object",
			Examples: []string{
				`llm.testProviderConnection("openai")`,
			},
		},
	}
}

// ValidateMethod validates method parameters.
func (b *LLMBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	if !b.IsInitialized() {
		return fmt.Errorf("llm bridge not initialized")
	}

	methods := b.Methods()
	for _, method := range methods {
		if method.Name == name {
			// Count required parameters
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

// ExecuteMethod executes a bridge method.
func (b *LLMBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod(name, args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	switch name {
	// Provider management
	case "setProvider":
		return b.setProvider(ctx, args)
	case "getProvider":
		return b.getProvider(ctx, args)
	case "listProviders":
		return b.listProviders(ctx, args)

	// Text generation
	case "generate":
		return b.generate(ctx, args)
	case "generateMessage":
		return b.generateMessage(ctx, args)
	case "stream":
		return b.stream(ctx, args)

	// Structured generation
	case "generateWithSchema":
		return b.generateWithSchema(ctx, args)

	// Schema management
	case "addResponseSchema":
		return b.addResponseSchema(ctx, args)
	case "getResponseSchema":
		return b.getResponseSchema(ctx, args)
	case "listResponseSchemas":
		return b.listResponseSchemas(ctx, args)
	case "validateWithSchema":
		return b.validateWithSchema(ctx, args)

	// Provider capabilities
	case "getCapabilities":
		return b.getCapabilities(ctx, args)
	case "getModelInfo":
		return b.getModelInfo(ctx, args)
	case "listModels":
		return b.listModels(ctx, args)
	case "checkCapability":
		return b.checkCapability(ctx, args)

	// Streaming
	case "streamMessage":
		return b.streamMessage(ctx, args)
	case "readStream":
		return b.readStream(ctx, args)
	case "closeStream":
		return b.closeStream(ctx, args)

	// Fallback configuration
	case "setFallbackChain":
		return b.setFallbackChain(ctx, args)
	case "getFallbackChain":
		return b.getFallbackChain(ctx, args)

	// Metrics
	case "getProviderMetrics":
		return b.getProviderMetrics(ctx, args)
	case "resetProviderMetrics":
		return b.resetProviderMetrics(ctx, args)

	// Schema generation
	case "generateSchemaFromExample":
		return b.generateSchemaFromExample(ctx, args)

	// Provider information
	case "getProviderInfo":
		return b.getProviderInfo(ctx, args)
	case "testProviderConnection":
		return b.testProviderConnection(ctx, args)

	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// TypeMappings returns type conversion hints.
func (b *LLMBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"provider": {
			GoType:     "bridge.Provider",
			ScriptType: "object",
			Converter:  "providerConverter",
		},
		"response": {
			GoType:     "*bridge.Response",
			ScriptType: "object",
			Converter:  "responseConverter",
		},
		"message": {
			GoType:     "bridge.Message",
			ScriptType: "object",
			Converter:  "messageConverter",
		},
		"schema": {
			GoType:     "*bridge.Schema",
			ScriptType: "object",
			Converter:  "schemaConverter",
		},
	}
}

// RequiredPermissions returns required permissions.
func (b *LLMBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm.providers",
			Actions:     []string{"read", "write"},
			Description: "Access to LLM provider APIs",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "llm.cache",
			Actions:     []string{"read", "write"},
			Description: "Cache for LLM responses",
		},
	}
}

// Implementation methods

func (b *LLMBridge) setProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	name := args[0].(engine.StringValue).Value()

	var config map[string]interface{}
	if len(args) > 1 {
		config = args[1].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// For now, we'll create a mock provider
	// In a real implementation, this would interface with go-llms
	b.activeProvider = name
	b.providers[name] = nil // Placeholder for actual provider

	// Initialize metrics for the provider
	if _, exists := b.metrics[name]; !exists {
		b.metrics[name] = &ProviderMetrics{}
	}

	result := map[string]interface{}{
		"name":   name,
		"config": config,
		"active": true,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) getProvider(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.activeProvider == "" {
		return engine.NewErrorValue(fmt.Errorf("no active provider set")), nil
	}

	result := map[string]interface{}{
		"name":   b.activeProvider,
		"active": true,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) listProviders(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	providers := make([]interface{}, 0, len(b.providers))
	for name := range b.providers {
		providers = append(providers, map[string]interface{}{
			"name":   name,
			"active": name == b.activeProvider,
		})
	}

	return engine.NewArrayValue(engine.ConvertSliceToScriptValue(providers)), nil
}

func (b *LLMBridge) generate(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("expected string for prompt, got %s", args[0].Type())), nil
	}
	prompt := args[0].(engine.StringValue).Value()

	var options map[string]interface{}
	if len(args) > 1 {
		options = args[1].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	b.mu.RLock()
	providerName := b.activeProvider
	b.mu.RUnlock()

	if providerName == "" {
		return engine.NewErrorValue(fmt.Errorf("no active provider set")), nil
	}

	// Update metrics
	b.updateMetrics(providerName, true, 0)

	// Mock response for now
	response := fmt.Sprintf("Generated response for: %s (options: %v)", prompt, options)

	return engine.NewStringValue(response), nil
}

func (b *LLMBridge) generateMessage(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	messages := args[0].(engine.ArrayValue).ToGo().([]interface{})

	var options map[string]interface{}
	if len(args) > 1 {
		options = args[1].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	b.mu.RLock()
	providerName := b.activeProvider
	b.mu.RUnlock()

	if providerName == "" {
		return engine.NewErrorValue(fmt.Errorf("no active provider set")), nil
	}

	// Update metrics
	b.updateMetrics(providerName, true, 0)

	// Mock response
	result := map[string]interface{}{
		"message": map[string]interface{}{
			"role":    "assistant",
			"content": fmt.Sprintf("Response to %d messages (options: %v)", len(messages), options),
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) stream(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	prompt := args[0].(engine.StringValue).Value()

	var options map[string]interface{}
	if len(args) > 1 {
		options = args[1].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	b.mu.RLock()
	providerName := b.activeProvider
	b.mu.RUnlock()

	if providerName == "" {
		return engine.NewErrorValue(fmt.Errorf("no active provider set")), nil
	}

	// Create a mock stream ID
	streamID := fmt.Sprintf("stream-%d", time.Now().UnixNano())

	result := map[string]interface{}{
		"stream_id": streamID,
		"prompt":    prompt,
		"options":   options,
		"active":    true,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) generateWithSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	prompt := args[0].(engine.StringValue).Value()
	schemaName := args[1].(engine.StringValue).Value()

	var options map[string]interface{}
	if len(args) > 2 {
		options = args[2].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	// Mock structured response
	result := map[string]interface{}{
		"raw":     fmt.Sprintf("Generated for: %s with schema: %s", prompt, schemaName),
		"parsed":  map[string]interface{}{"example": "data"},
		"valid":   true,
		"options": options,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) addResponseSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	name := args[0].(engine.StringValue).Value()
	schema := args[1].(engine.ObjectValue).ToGo().(map[string]interface{})

	// In a real implementation, this would store the schema
	_ = name
	_ = schema

	return engine.NewNilValue(), nil
}

func (b *LLMBridge) getResponseSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	name := args[0].(engine.StringValue).Value()

	// Mock schema
	result := map[string]interface{}{
		"name": name,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"example": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) listResponseSchemas(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Mock schema list
	schemas := []interface{}{
		map[string]interface{}{"name": "person", "type": "object"},
		map[string]interface{}{"name": "color_list", "type": "array"},
	}

	return engine.NewArrayValue(engine.ConvertSliceToScriptValue(schemas)), nil
}

func (b *LLMBridge) validateWithSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	data := args[0].(engine.ObjectValue).ToGo().(map[string]interface{})
	schemaName := args[1].(engine.StringValue).Value()

	// Mock validation
	result := map[string]interface{}{
		"valid":  true,
		"errors": []interface{}{},
		"data":   data,
		"schema": schemaName,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) getCapabilities(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	providerName := b.activeProvider
	b.mu.RUnlock()

	if providerName == "" {
		return engine.NewErrorValue(fmt.Errorf("no active provider set")), nil
	}

	// Mock capabilities
	capabilities := map[string]interface{}{
		"streaming":          true,
		"structured_output":  true,
		"function_calling":   true,
		"embeddings":         false,
		"max_context_length": 4096,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(capabilities)), nil
}

func (b *LLMBridge) getModelInfo(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	model := args[0].(engine.StringValue).Value()

	// Mock model info
	info := map[string]interface{}{
		"name":         model,
		"description":  "Language model",
		"context_size": 4096,
		"capabilities": []interface{}{"text-generation", "chat"},
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(info)), nil
}

func (b *LLMBridge) listModels(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Mock model list
	models := []interface{}{
		map[string]interface{}{
			"name": "gpt-3.5-turbo",
			"type": "chat",
		},
		map[string]interface{}{
			"name": "gpt-4",
			"type": "chat",
		},
	}

	return engine.NewArrayValue(engine.ConvertSliceToScriptValue(models)), nil
}

func (b *LLMBridge) checkCapability(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	capability := args[0].(engine.StringValue).Value()

	// Mock capability check
	supported := map[string]bool{
		"streaming":         true,
		"structured_output": true,
		"function_calling":  true,
		"embeddings":        false,
	}

	hasCapability, exists := supported[capability]
	if !exists {
		hasCapability = false
	}

	return engine.NewBoolValue(hasCapability), nil
}

func (b *LLMBridge) streamMessage(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	messages := args[0].(engine.ArrayValue).ToGo().([]interface{})

	var options map[string]interface{}
	if len(args) > 1 {
		options = args[1].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	// Create a mock stream
	streamID := fmt.Sprintf("stream-%d", time.Now().UnixNano())

	result := map[string]interface{}{
		"stream_id": streamID,
		"messages":  messages,
		"options":   options,
		"active":    true,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) readStream(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	streamID := args[0].(engine.StringValue).Value()

	// Mock stream chunk
	chunk := map[string]interface{}{
		"stream_id": streamID,
		"chunk":     "Next chunk of text",
		"done":      false,
		"index":     1,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(chunk)), nil
}

func (b *LLMBridge) closeStream(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	streamID := args[0].(engine.StringValue).Value()

	// Mock close stream
	_ = streamID

	return engine.NewNilValue(), nil
}

func (b *LLMBridge) setFallbackChain(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	providers := args[0].(engine.ArrayValue).ToGo().([]interface{})

	b.mu.Lock()
	defer b.mu.Unlock()

	b.fallbackChain = make([]string, 0, len(providers))
	for _, p := range providers {
		if name, ok := p.(string); ok {
			b.fallbackChain = append(b.fallbackChain, name)
		}
	}

	return engine.NewNilValue(), nil
}

func (b *LLMBridge) getFallbackChain(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	chain := make([]interface{}, len(b.fallbackChain))
	for i, name := range b.fallbackChain {
		chain[i] = name
	}

	return engine.NewArrayValue(engine.ConvertSliceToScriptValue(chain)), nil
}

func (b *LLMBridge) getProviderMetrics(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	provider := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	metrics, exists := b.metrics[provider]
	b.mu.RUnlock()

	if !exists {
		return engine.NewErrorValue(fmt.Errorf("no metrics for provider: %s", provider)), nil
	}

	result := map[string]interface{}{
		"total_requests":   metrics.TotalRequests,
		"successful_calls": metrics.SuccessfulCalls,
		"failed_calls":     metrics.FailedCalls,
		"average_latency":  metrics.AverageLatency.Milliseconds(),
		"success_rate":     float64(metrics.SuccessfulCalls) / float64(metrics.TotalRequests) * 100,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *LLMBridge) resetProviderMetrics(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	provider := args[0].(engine.StringValue).Value()

	b.mu.Lock()
	defer b.mu.Unlock()

	b.metrics[provider] = &ProviderMetrics{}

	return engine.NewNilValue(), nil
}

func (b *LLMBridge) generateSchemaFromExample(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	example := args[0].(engine.ObjectValue).ToGo().(map[string]interface{})
	name := args[1].(engine.StringValue).Value()

	// Mock schema generation
	schema := map[string]interface{}{
		"name": name,
		"type": "object",
		"properties": map[string]interface{}{
			// Simplified schema generation
			"example": map[string]interface{}{
				"type": "object",
			},
		},
		"generated_from": example,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(schema)), nil
}

func (b *LLMBridge) getProviderInfo(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	provider := args[0].(engine.StringValue).Value()

	// Mock provider info
	info := map[string]interface{}{
		"name":         provider,
		"type":         "llm",
		"version":      "1.0.0",
		"capabilities": []interface{}{"text-generation", "chat", "streaming"},
		"models":       []interface{}{"model1", "model2"},
		"status":       "active",
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(info)), nil
}

func (b *LLMBridge) testProviderConnection(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	provider := args[0].(engine.StringValue).Value()

	// Mock connection test
	result := map[string]interface{}{
		"provider":   provider,
		"connected":  true,
		"latency_ms": 42,
		"status":     "healthy",
		"tested_at":  time.Now().Format(time.RFC3339),
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

// Helper methods

func (b *LLMBridge) updateMetrics(provider string, success bool, latency time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	metrics, exists := b.metrics[provider]
	if !exists {
		metrics = &ProviderMetrics{}
		b.metrics[provider] = metrics
	}

	metrics.TotalRequests++
	if success {
		metrics.SuccessfulCalls++
	} else {
		metrics.FailedCalls++
	}

	metrics.TotalLatency += latency
	if metrics.TotalRequests > 0 {
		metrics.AverageLatency = metrics.TotalLatency / time.Duration(metrics.TotalRequests)
	}
}

// getProvider returns the active provider
