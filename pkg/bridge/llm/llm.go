// ABOUTME: LLM bridge provides access to language model providers through go-llms interfaces.
// ABOUTME: Wraps go-llms Provider interface for script engine access without reimplementation.

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/bridge"
	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for LLM functionality
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"

	// go-llms imports for schema validation
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/generator"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	structuredDomain "github.com/lexlapax/go-llms/pkg/structured/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// LLMBridge provides script access to language model functionality via go-llms.
type LLMBridge struct {
	mu             sync.RWMutex
	providers      map[string]bridge.Provider
	activeProvider string
	initialized    bool

	// Schema validation components from go-llms v0.3.5
	responseSchemas map[string]*schemaDomain.Schema // Schema cache by name
	schemaRepo      schemaDomain.SchemaRepository   // Schema storage
	schemaValidator schemaDomain.Validator          // Schema validator
	schemaGenerator schemaDomain.SchemaGenerator    // Schema generator
	schemaCache     *processor.SchemaCache          // Performance cache
	promptEnhancer  structuredDomain.PromptEnhancer // Prompt enhancement
	structProcessor structuredDomain.Processor      // JSON extraction
}

// NewLLMBridge creates a new LLM bridge.
func NewLLMBridge() *LLMBridge {
	return &LLMBridge{
		providers:       make(map[string]bridge.Provider),
		responseSchemas: make(map[string]*schemaDomain.Schema),
	}
}

// GetID returns the bridge identifier.
func (b *LLMBridge) GetID() string {
	return "llm"
}

// GetMetadata returns bridge metadata.
func (b *LLMBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "llm",
		Version:     "2.0.0",
		Description: "LLM provider access bridge with v0.3.5 schema validation support",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *LLMBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize schema validation components from go-llms v0.3.5
	b.schemaRepo = repository.NewInMemorySchemaRepository()
	b.schemaValidator = validation.NewValidator(validation.WithCoercion(true))
	b.schemaGenerator = generator.NewReflectionSchemaGenerator()
	b.schemaCache = processor.NewSchemaCache()
	b.promptEnhancer = processor.NewPromptEnhancer()
	b.structProcessor = processor.NewStructuredProcessor(b.schemaValidator)

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *LLMBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

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
func (b *LLMBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *LLMBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		{
			Name:        "registerProvider",
			Description: "Register an LLM provider",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Provider name", Required: true},
				{Name: "provider", Type: "Provider", Description: "Provider instance", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "setActiveProvider",
			Description: "Set the active LLM provider",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "generate",
			Description: "Generate text using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Description: "Input prompt", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "generateMessage",
			Description: "Generate response from messages using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Description: "Input messages", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "stream",
			Description: "Stream text generation using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Description: "Input prompt", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "channel",
		},
		{
			Name:        "streamMessage",
			Description: "Stream response from messages using the active provider",
			Parameters: []engine.ParameterInfo{
				{Name: "messages", Type: "array", Description: "Input messages", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "channel",
		},
		{
			Name:        "listProviders",
			Description: "List all registered providers",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "getActiveProvider",
			Description: "Get the name of the active provider",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "string",
		},
		// Schema validation methods (v0.3.5)
		{
			Name:        "generateWithSchema",
			Description: "Generate structured output with schema validation",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Description: "Input prompt", Required: true},
				{Name: "schema", Type: "object", Description: "JSON Schema for validation", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "registerSchema",
			Description: "Register a named schema for reuse",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Schema name", Required: true},
				{Name: "schema", Type: "object", Description: "JSON Schema definition", Required: true},
				{Name: "version", Type: "string", Description: "Schema version", Required: false},
			},
			ReturnType: "void",
		},
		{
			Name:        "getSchema",
			Description: "Get a registered schema by name",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Schema name", Required: true},
				{Name: "version", Type: "string", Description: "Schema version (optional)", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "listSchemas",
			Description: "List all registered schemas",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "array",
		},
		{
			Name:        "validateWithSchema",
			Description: "Validate data against a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "any", Description: "Data to validate", Required: true},
				{Name: "schema", Type: "object", Description: "JSON Schema or schema name", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "generateSchemaFromExample",
			Description: "Generate a JSON Schema from example data",
			Parameters: []engine.ParameterInfo{
				{Name: "example", Type: "any", Description: "Example data", Required: true},
				{Name: "options", Type: "object", Description: "Generation options", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "clearSchemaCache",
			Description: "Clear the schema cache",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "void",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *LLMBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Provider": {
			GoType:     "Provider",
			ScriptType: "object",
		},
		"Message": {
			GoType:     "Message",
			ScriptType: "object",
		},
		"Response": {
			GoType:     "Response",
			ScriptType: "object",
		},
		"ProviderOptions": {
			GoType:     "ProviderOptions",
			ScriptType: "object",
		},
		"Schema": {
			GoType:     "Schema",
			ScriptType: "object",
		},
		"ValidationResult": {
			GoType:     "ValidationResult",
			ScriptType: "object",
		},
		"SchemaInfo": {
			GoType:     "SchemaInfo",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls.
func (b *LLMBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *LLMBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionNetwork,
			Resource:    "llm",
			Actions:     []string{"access"},
			Description: "Access to LLM providers",
		},
	}
}

// RegisterProvider registers an LLM provider.
func (b *LLMBridge) RegisterProvider(name string, provider bridge.Provider) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.providers[name] = provider
	return nil
}

// SetActiveProvider sets the active provider.
func (b *LLMBridge) SetActiveProvider(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	b.activeProvider = name
	return nil
}

// GetActiveProvider returns the active provider.
func (b *LLMBridge) GetActiveProvider() bridge.Provider {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.activeProvider == "" {
		return nil
	}

	return b.providers[b.activeProvider]
}

// ListProviders returns all registered provider names.
func (b *LLMBridge) ListProviders() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.providers))
	for name := range b.providers {
		names = append(names, name)
	}
	return names
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *LLMBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createProvider":
		if len(args) < 2 {
			return nil, fmt.Errorf("createProvider requires name and config parameters")
		}
		providerName, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("provider name must be string")
		}
		config, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config must be object")
		}

		// Extract provider type
		providerType, ok := config["type"].(string)
		if !ok {
			return nil, fmt.Errorf("provider type required in config")
		}

		// Create provider based on type
		var llmProvider bridge.Provider

		// Get API key and model
		apiKey, _ := config["apiKey"].(string)
		model, _ := config["model"].(string)
		if model == "" {
			model = "gpt-4" // default
		}

		switch providerType {
		case "openai":
			// Create OpenAI provider
			llmProvider = provider.NewOpenAIProvider(apiKey, model)
		case "anthropic":
			// Create Anthropic provider
			llmProvider = provider.NewAnthropicProvider(apiKey, model)
		default:
			return nil, fmt.Errorf("unsupported provider type: %s", providerType)
		}

		// Store provider
		b.providers[providerName] = llmProvider
		if b.activeProvider == "" {
			b.activeProvider = providerName
		}

		return map[string]interface{}{
			"name": providerName,
			"type": providerType,
		}, nil

	case "generate":
		if len(args) < 1 {
			return nil, fmt.Errorf("generate requires prompt parameter")
		}

		// Get active provider
		if b.activeProvider == "" {
			return nil, fmt.Errorf("no active provider set")
		}
		provider, exists := b.providers[b.activeProvider]
		if !exists {
			return nil, fmt.Errorf("active provider not found: %s", b.activeProvider)
		}

		// Get prompt
		prompt, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("prompt must be string")
		}

		// Generate response
		response, err := provider.Generate(ctx, prompt, nil)
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// Return response
		return map[string]interface{}{
			"content": response,
		}, nil

	case "generateMessage":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateMessage requires messages parameter")
		}

		// Get active provider
		if b.activeProvider == "" {
			return nil, fmt.Errorf("no active provider set")
		}
		provider, exists := b.providers[b.activeProvider]
		if !exists {
			return nil, fmt.Errorf("active provider not found: %s", b.activeProvider)
		}

		// Convert messages
		var messages []llmdomain.Message
		if msgList, ok := args[0].([]interface{}); ok {
			for _, msg := range msgList {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					role, _ := msgMap["role"].(string)
					content, _ := msgMap["content"].(string)
					messages = append(messages, llmdomain.NewTextMessage(llmdomain.Role(role), content))
				}
			}
		}

		// Generate response
		response, err := provider.GenerateMessage(ctx, messages, nil)
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// Return response
		return map[string]interface{}{
			"content": response.Content,
		}, nil

	case "setActiveProvider":
		if len(args) < 1 {
			return nil, fmt.Errorf("setActiveProvider requires name parameter")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		if _, exists := b.providers[name]; !exists {
			return nil, fmt.Errorf("provider %s not found", name)
		}

		b.activeProvider = name
		return nil, nil

	case "getActiveProvider":
		return b.activeProvider, nil

	case "listProviders":
		providers := make([]map[string]interface{}, 0, len(b.providers))
		for name := range b.providers {
			providers = append(providers, map[string]interface{}{
				"name":   name,
				"active": name == b.activeProvider,
			})
		}
		return providers, nil

	case "generateWithSchema":
		if len(args) < 2 {
			return nil, fmt.Errorf("generateWithSchema requires prompt and schema parameters")
		}

		// Get active provider
		if b.activeProvider == "" {
			return nil, fmt.Errorf("no active provider set")
		}
		provider, exists := b.providers[b.activeProvider]
		if !exists {
			return nil, fmt.Errorf("active provider not found: %s", b.activeProvider)
		}

		// Get prompt
		prompt, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("prompt must be string")
		}

		// Get schema
		schemaMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be object")
		}

		// Convert schema map to Schema struct
		schemaJSON, err := json.Marshal(schemaMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %w", err)
		}
		var schema schemaDomain.Schema
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}

		// Enhance prompt with schema
		enhancedPrompt, err := b.promptEnhancer.Enhance(prompt, &schema)
		if err != nil {
			return nil, fmt.Errorf("prompt enhancement failed: %w", err)
		}

		// Generate response
		response, err := provider.Generate(ctx, enhancedPrompt, nil)
		if err != nil {
			return nil, fmt.Errorf("generation failed: %w", err)
		}

		// Extract and validate structured output
		structuredData, err := b.structProcessor.Process(&schema, response)
		if err != nil {
			return nil, fmt.Errorf("structured output processing failed: %w", err)
		}

		return map[string]interface{}{
			"data":      structuredData,
			"rawOutput": response,
			"validated": true,
		}, nil

	case "registerSchema":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerSchema requires name and schema parameters")
		}

		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		schemaMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be object")
		}

		// Convert schema map to Schema struct
		schemaJSON, err := json.Marshal(schemaMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %w", err)
		}
		var schema schemaDomain.Schema
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}

		// Get optional version (removed since repo doesn't support versions)
		// Version tracking could be added as metadata in the schema description

		// Store in repository
		if err := b.schemaRepo.Save(name, &schema); err != nil {
			return nil, fmt.Errorf("failed to save schema: %w", err)
		}

		// Store in local cache
		b.responseSchemas[name] = &schema

		// Cache for performance (cache stores JSON bytes)
		b.schemaCache.Set(uint64(len(name)), schemaJSON)

		return nil, nil

	case "getSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("getSchema requires name parameter")
		}

		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		// Check local map first
		if schema, found := b.responseSchemas[name]; found {
			return schema, nil
		}

		// Get from repository
		schema, err := b.schemaRepo.Get(name)
		if err != nil {
			return nil, fmt.Errorf("schema not found: %w", err)
		}

		// Store in local cache
		b.responseSchemas[name] = schema

		return schema, nil

	case "listSchemas":
		// List from local cache
		result := make([]map[string]interface{}, 0, len(b.responseSchemas))
		for name, schema := range b.responseSchemas {
			result = append(result, map[string]interface{}{
				"name":        name,
				"description": schema.Description,
				"title":       schema.Title,
			})
		}
		return result, nil

	case "validateWithSchema":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateWithSchema requires data and schema parameters")
		}

		data := args[0]

		// Get schema (could be object or name)
		var schema *schemaDomain.Schema
		switch v := args[1].(type) {
		case string:
			// Schema name
			var err error
			schema, err = b.schemaRepo.Get(v)
			if err != nil {
				return nil, fmt.Errorf("schema not found: %w", err)
			}
		case map[string]interface{}:
			// Schema object
			schemaJSON, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal schema: %w", err)
			}
			schema = &schemaDomain.Schema{}
			if err := json.Unmarshal(schemaJSON, schema); err != nil {
				return nil, fmt.Errorf("invalid schema: %w", err)
			}
		default:
			return nil, fmt.Errorf("schema must be string (name) or object")
		}

		// Convert data to JSON string for validation
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		// Validate
		result, err := b.schemaValidator.Validate(schema, string(dataJSON))
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}

		return map[string]interface{}{
			"valid":  result.Valid,
			"errors": result.Errors,
		}, nil

	case "generateSchemaFromExample":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateSchemaFromExample requires example parameter")
		}

		example := args[0]

		// Generate schema from example using reflection generator
		schema, err := b.schemaGenerator.GenerateSchema(example)
		if err != nil {
			return nil, fmt.Errorf("schema generation failed: %w", err)
		}

		return schema, nil

	case "clearSchemaCache":
		b.schemaCache.Clear()
		return nil, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}
