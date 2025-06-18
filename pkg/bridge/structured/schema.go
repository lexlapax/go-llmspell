// ABOUTME: Schema bridge providing access to go-llms schema validation system
// ABOUTME: Wraps go-llms schema functionality for script-based validation and generation

package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for schema functionality
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/generator"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// validationMetrics tracks performance metrics for validation operations
type validationMetrics struct {
	TotalValidations      int64         `json:"totalValidations"`
	SuccessfulValidations int64         `json:"successfulValidations"`
	FailedValidations     int64         `json:"failedValidations"`
	AverageLatency        time.Duration `json:"averageLatency"`
	CacheHits             int64         `json:"cacheHits"`
	CacheMisses           int64         `json:"cacheMisses"`
	AsyncValidations      int64         `json:"asyncValidations"`
	mutex                 sync.RWMutex
}

// asyncValidationRequest represents an async validation request
type asyncValidationRequest struct {
	ID       string               `json:"id"`
	Schema   *schemaDomain.Schema `json:"schema"`
	Data     interface{}          `json:"data"`
	Callback interface{}          `json:"callback"`
	Created  time.Time            `json:"created"`
}

// validationCacheEntry represents a cached validation result
type validationCacheEntry struct {
	Result  *schemaDomain.ValidationResult `json:"result"`
	Created time.Time                      `json:"created"`
	TTL     time.Duration                  `json:"ttl"`
}

// SchemaBridge provides access to go-llms schema validation system
type SchemaBridge struct {
	mu                    sync.RWMutex
	initialized           bool
	validator             schemaDomain.Validator
	generator             schemaDomain.SchemaGenerator
	repository            schemaDomain.SchemaRepository
	fileRepo              schemaDomain.SchemaRepository        // File-based repository for persistence
	migrators             map[string]repository.SchemaMigrator // Migration registry
	tagGenerator          schemaDomain.SchemaGenerator         // Tag-based generator
	schemas               map[string]*schemaDomain.Schema      // Simple in-memory storage
	customValidators      map[string]interface{}               // Script-based custom validators
	validationCache       sync.Map                             // Validation result cache
	validationMetrics     *validationMetrics                   // Performance metrics
	asyncValidationQueue  chan *asyncValidationRequest         // Async validation queue
	conditionalValidators map[string]interface{}               // Conditional validation functions
}

// NewSchemaBridge creates a new schema bridge
func NewSchemaBridge() *SchemaBridge {
	return &SchemaBridge{
		migrators:             make(map[string]repository.SchemaMigrator),
		schemas:               make(map[string]*schemaDomain.Schema),
		customValidators:      make(map[string]interface{}),
		conditionalValidators: make(map[string]interface{}),
		validationMetrics:     &validationMetrics{},
		asyncValidationQueue:  make(chan *asyncValidationRequest, 100),
	}
}

// GetID returns the bridge ID
func (b *SchemaBridge) GetID() string {
	return "schema"
}

// GetMetadata returns bridge metadata
func (b *SchemaBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Schema Bridge",
		Version:     "2.0.0",
		Description: "Provides access to go-llms schema validation, generation, versioning, and migration system",
		Author:      "go-llmspell",
		License:     "MIT",
		Dependencies: []string{
			"github.com/lexlapax/go-llms/pkg/schema/domain",
			"github.com/lexlapax/go-llms/pkg/schema/generator",
			"github.com/lexlapax/go-llms/pkg/schema/repository",
			"github.com/lexlapax/go-llms/pkg/schema/validation",
		},
	}
}

// Initialize initializes the bridge
func (b *SchemaBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize go-llms components
	b.validator = validation.NewValidator()
	b.generator = generator.NewReflectionSchemaGenerator()
	b.repository = repository.NewInMemorySchemaRepository()
	b.tagGenerator = generator.NewTagSchemaGenerator()

	// Start async validation worker
	go b.processAsyncValidations(ctx)

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources
func (b *SchemaBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.fileRepo != nil {
		// Clean up file repository if initialized
		if closer, ok := b.fileRepo.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	// Close async validation queue
	close(b.asyncValidationQueue)

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *SchemaBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *SchemaBridge) RegisterWithEngine(e engine.ScriptEngine) error {
	return e.RegisterBridge(b)
}

// Methods returns all available methods
func (b *SchemaBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Core Schema Operations
		{Name: "createSchema", Description: "Create a new schema object", Parameters: []engine.ParameterInfo{{Name: "schemaData", Type: "object", Required: true, Description: "Schema definition object"}}, ReturnType: "object"},
		{Name: "createProperty", Description: "Create a property definition", Parameters: []engine.ParameterInfo{{Name: "propertyType", Type: "string", Required: true, Description: "Property type"}, {Name: "constraints", Type: "object", Required: false, Description: "Property constraints"}}, ReturnType: "object"},
		{Name: "validateJSON", Description: "Validate JSON data against a schema", Parameters: []engine.ParameterInfo{{Name: "schema", Type: "object", Required: true, Description: "Schema to validate against"}, {Name: "data", Type: "object", Required: true, Description: "Data to validate"}}, ReturnType: "object"},
		{Name: "validateStruct", Description: "Validate a struct against a schema", Parameters: []engine.ParameterInfo{{Name: "schema", Type: "object", Required: true, Description: "Schema to validate against"}, {Name: "data", Type: "object", Required: true, Description: "Struct data to validate"}}, ReturnType: "object"},
		{Name: "generateSchemaFromType", Description: "Generate schema from a type definition", Parameters: []engine.ParameterInfo{{Name: "typeInfo", Type: "object", Required: true, Description: "Type information"}}, ReturnType: "object"},
		{Name: "convertJSONSchema", Description: "Convert JSON Schema string to schema object", Parameters: []engine.ParameterInfo{{Name: "jsonSchema", Type: "string", Required: true, Description: "JSON Schema string"}}, ReturnType: "object"},

		// Repository Operations
		{Name: "saveSchema", Description: "Save a schema to the repository", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}, {Name: "schema", Type: "object", Required: true, Description: "Schema object"}}, ReturnType: "void"},
		{Name: "getSchema", Description: "Get a schema from the repository", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}}, ReturnType: "object"},
		{Name: "deleteSchema", Description: "Delete a schema from the repository", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}}, ReturnType: "void"},

		// Versioning and Migration
		{Name: "initializeFileRepository", Description: "Initialize file-based repository for schema persistence", Parameters: []engine.ParameterInfo{{Name: "directory", Type: "string", Required: true, Description: "Repository directory path"}}, ReturnType: "void"},
		{Name: "saveSchemaVersion", Description: "Save a specific version of a schema", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}, {Name: "schema", Type: "object", Required: true, Description: "Schema object"}, {Name: "version", Type: "number", Required: true, Description: "Version number"}}, ReturnType: "void"},
		{Name: "getSchemaVersion", Description: "Get a specific version of a schema", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}, {Name: "version", Type: "number", Required: true, Description: "Version number"}}, ReturnType: "object"},
		{Name: "listSchemaVersions", Description: "List all versions of a schema", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}}, ReturnType: "array"},
		{Name: "setCurrentSchemaVersion", Description: "Set the current version of a schema", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}, {Name: "version", Type: "number", Required: true, Description: "Version number"}}, ReturnType: "void"},
		{Name: "registerMigrator", Description: "Register a schema migrator", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Migrator name"}, {Name: "migrator", Type: "object", Required: true, Description: "Migrator configuration"}}, ReturnType: "void"},
		{Name: "migrateSchema", Description: "Migrate a schema between versions", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Schema name"}, {Name: "fromVersion", Type: "number", Required: true, Description: "Source version"}, {Name: "toVersion", Type: "number", Required: true, Description: "Target version"}}, ReturnType: "object"},
		{Name: "exportRepository", Description: "Export entire repository to JSON", Parameters: []engine.ParameterInfo{}, ReturnType: "string"},
		{Name: "importRepository", Description: "Import repository from JSON", Parameters: []engine.ParameterInfo{{Name: "data", Type: "object", Required: true, Description: "Repository data"}}, ReturnType: "object"},

		// Tag-Based Schema Generation
		{Name: "generateFromTags", Description: "Generate schema from struct tags", Parameters: []engine.ParameterInfo{{Name: "structData", Type: "object", Required: true, Description: "Struct with tags"}}, ReturnType: "object"},
		{Name: "setTagPriority", Description: "Set the order in which tags are processed", Parameters: []engine.ParameterInfo{{Name: "tags", Type: "array", Required: true, Description: "Ordered list of tag names"}}, ReturnType: "void"},
		{Name: "registerTagParser", Description: "Register a custom tag parser", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Parser name"}, {Name: "parser", Type: "object", Required: true, Description: "Parser configuration"}}, ReturnType: "void"},
		{Name: "extractValidationRules", Description: "Extract validation rules from struct tags", Parameters: []engine.ParameterInfo{{Name: "structData", Type: "object", Required: true, Description: "Struct with validation tags"}}, ReturnType: "object"},
		{Name: "generateWithDocumentation", Description: "Generate schema with embedded documentation from tags", Parameters: []engine.ParameterInfo{{Name: "structData", Type: "object", Required: true, Description: "Struct with doc tags"}, {Name: "includeExamples", Type: "boolean", Required: false, Description: "Include example values"}}, ReturnType: "object"},

		// Import/Export Operations
		{Name: "exportToJSONSchema", Description: "Export schema to JSON Schema format", Parameters: []engine.ParameterInfo{{Name: "schema", Type: "object", Required: true, Description: "Schema to export"}}, ReturnType: "string"},
		{Name: "exportToOpenAPI", Description: "Export schema to OpenAPI schema format", Parameters: []engine.ParameterInfo{{Name: "schema", Type: "object", Required: true, Description: "Schema to export"}}, ReturnType: "string"},
		{Name: "importFromFile", Description: "Import schema from file", Parameters: []engine.ParameterInfo{{Name: "filePath", Type: "string", Required: true, Description: "File path"}, {Name: "format", Type: "string", Required: false, Description: "File format"}}, ReturnType: "object"},
		{Name: "importFromString", Description: "Import schema from string content", Parameters: []engine.ParameterInfo{{Name: "content", Type: "string", Required: true, Description: "Schema content"}, {Name: "format", Type: "string", Required: false, Description: "Content format"}}, ReturnType: "object"},
		{Name: "convertFormat", Description: "Convert schema between different formats", Parameters: []engine.ParameterInfo{{Name: "schema", Type: "object", Required: true, Description: "Source schema"}, {Name: "fromFormat", Type: "string", Required: true, Description: "Source format"}, {Name: "toFormat", Type: "string", Required: true, Description: "Target format"}}, ReturnType: "object"},
		{Name: "mergeSchemas", Description: "Merge multiple schemas into one", Parameters: []engine.ParameterInfo{{Name: "schemas", Type: "array", Required: true, Description: "Array of schemas"}, {Name: "strategy", Type: "string", Required: false, Description: "Merge strategy"}}, ReturnType: "object"},
		{Name: "generateDiff", Description: "Generate diff between two schemas", Parameters: []engine.ParameterInfo{{Name: "oldSchema", Type: "object", Required: true, Description: "Original schema"}, {Name: "newSchema", Type: "object", Required: true, Description: "Updated schema"}}, ReturnType: "object"},
		{Name: "exportCollection", Description: "Export multiple schemas as a collection", Parameters: []engine.ParameterInfo{{Name: "schemaIds", Type: "array", Required: true, Description: "List of schema IDs"}, {Name: "format", Type: "string", Required: false, Description: "Export format"}}, ReturnType: "object"},
		{Name: "importCollection", Description: "Import a collection of schemas", Parameters: []engine.ParameterInfo{{Name: "collection", Type: "object", Required: true, Description: "Schema collection"}, {Name: "overwrite", Type: "boolean", Required: false, Description: "Overwrite existing schemas"}}, ReturnType: "array"},

		// Custom Validation
		{Name: "registerCustomValidator", Description: "Register a custom validation function", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Validator name"}, {Name: "validator", Type: "object", Required: true, Description: "Validator configuration"}}, ReturnType: "void"},
		{Name: "unregisterCustomValidator", Description: "Unregister a custom validator", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Validator name"}}, ReturnType: "void"},
		{Name: "listCustomValidators", Description: "List all registered custom validators", Parameters: []engine.ParameterInfo{}, ReturnType: "array"},
		{Name: "validateWithCustom", Description: "Validate data using custom validators", Parameters: []engine.ParameterInfo{{Name: "data", Type: "object", Required: true, Description: "Data to validate"}, {Name: "validatorName", Type: "string", Required: true, Description: "Custom validator name"}}, ReturnType: "object"},
		{Name: "validateAsync", Description: "Perform asynchronous validation", Parameters: []engine.ParameterInfo{{Name: "schema", Type: "object", Required: true, Description: "Schema to validate against"}, {Name: "data", Type: "object", Required: true, Description: "Data to validate"}, {Name: "callback", Type: "object", Required: false, Description: "Callback configuration"}}, ReturnType: "object"},
		{Name: "getValidationMetrics", Description: "Get validation performance metrics", Parameters: []engine.ParameterInfo{}, ReturnType: "object"},
		{Name: "clearValidationCache", Description: "Clear validation cache entries", Parameters: []engine.ParameterInfo{}, ReturnType: "void"},
		{Name: "registerConditionalValidator", Description: "Register a conditional validator", Parameters: []engine.ParameterInfo{{Name: "name", Type: "string", Required: true, Description: "Validator name"}, {Name: "validator", Type: "object", Required: true, Description: "Conditional validator configuration"}}, ReturnType: "void"},
		{Name: "validateConditional", Description: "Validate data with conditional validators", Parameters: []engine.ParameterInfo{{Name: "data", Type: "object", Required: true, Description: "Data to validate"}, {Name: "validatorName", Type: "string", Required: true, Description: "Conditional validator name"}}, ReturnType: "object"},
	}
}

// ValidateMethod validates method parameters
func (b *SchemaBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	if !b.IsInitialized() {
		return fmt.Errorf("schema bridge not initialized")
	}

	// Basic parameter count validation
	methods := b.Methods()
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

// ExecuteMethod executes a bridge method with ScriptValue support
func (b *SchemaBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if err := b.ValidateMethod(name, args); err != nil {
		return engine.NewErrorValue(err), nil
	}

	switch name {
	// Core Schema Operations
	case "createSchema":
		return b.createSchema(ctx, args)
	case "createProperty":
		return b.createProperty(ctx, args)
	case "validateJSON":
		return b.validateJSON(ctx, args)
	case "validateStruct":
		return b.validateStruct(ctx, args)
	case "generateSchemaFromType":
		return b.generateSchemaFromType(ctx, args)
	case "convertJSONSchema":
		return b.convertJSONSchema(ctx, args)

	// Repository Operations
	case "saveSchema":
		return b.saveSchema(ctx, args)
	case "getSchema":
		return b.getSchema(ctx, args)
	case "deleteSchema":
		return b.deleteSchema(ctx, args)

	// Versioning and Migration
	case "initializeFileRepository":
		return b.initializeFileRepository(ctx, args)
	case "saveSchemaVersion":
		return b.saveSchemaVersion(ctx, args)
	case "getSchemaVersion":
		return b.getSchemaVersion(ctx, args)
	case "listSchemaVersions":
		return b.listSchemaVersions(ctx, args)
	case "setCurrentSchemaVersion":
		return b.setCurrentSchemaVersion(ctx, args)
	case "registerMigrator":
		return b.registerMigrator(ctx, args)
	case "migrateSchema":
		return b.migrateSchema(ctx, args)
	case "exportRepository":
		return b.exportRepository(ctx, args)
	case "importRepository":
		return b.importRepository(ctx, args)

	// Tag-Based Schema Generation
	case "generateFromTags":
		return b.generateFromTags(ctx, args)
	case "setTagPriority":
		return b.setTagPriority(ctx, args)
	case "registerTagParser":
		return b.registerTagParser(ctx, args)
	case "extractValidationRules":
		return b.extractValidationRules(ctx, args)
	case "generateWithDocumentation":
		return b.generateWithDocumentation(ctx, args)

	// Import/Export Operations
	case "exportToJSONSchema":
		return b.exportToJSONSchema(ctx, args)
	case "exportToOpenAPI":
		return b.exportToOpenAPI(ctx, args)
	case "importFromFile":
		return b.importFromFile(ctx, args)
	case "importFromString":
		return b.importFromString(ctx, args)
	case "convertFormat":
		return b.convertFormat(ctx, args)
	case "mergeSchemas":
		return b.mergeSchemas(ctx, args)
	case "generateDiff":
		return b.generateDiff(ctx, args)
	case "exportCollection":
		return b.exportCollection(ctx, args)
	case "importCollection":
		return b.importCollection(ctx, args)

	// Custom Validation
	case "registerCustomValidator":
		return b.registerCustomValidator(ctx, args)
	case "unregisterCustomValidator":
		return b.unregisterCustomValidator(ctx, args)
	case "listCustomValidators":
		return b.listCustomValidators(ctx, args)
	case "validateWithCustom":
		return b.validateWithCustom(ctx, args)
	case "validateAsync":
		return b.validateAsync(ctx, args)
	case "getValidationMetrics":
		return b.getValidationMetrics(ctx, args)
	case "clearValidationCache":
		return b.clearValidationCache(ctx, args)
	case "registerConditionalValidator":
		return b.registerConditionalValidator(ctx, args)
	case "validateConditional":
		return b.validateConditional(ctx, args)

	default:
		return engine.NewErrorValue(fmt.Errorf("unknown method: %s", name)), nil
	}
}

// Implementation methods start here

// Core Schema Operations
func (b *SchemaBridge) createSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("expected object for schema data, got %s", args[0].Type())), nil
	}

	schemaData := args[0].(engine.ObjectValue).ToGo().(map[string]interface{})
	
	// Create schema using script conversion
	schema, err := scriptToSchema(schemaData)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to create schema: %w", err)), nil
	}

	result := map[string]interface{}{
		"schema":    schemaToScript(schema),
		"created":   true,
		"timestamp": time.Now(),
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *SchemaBridge) createProperty(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("expected string for property type, got %s", args[0].Type())), nil
	}

	propertyType := args[0].(engine.StringValue).Value()
	
	var constraints map[string]interface{}
	if len(args) > 1 && args[1].Type() == engine.TypeObject {
		constraints = args[1].(engine.ObjectValue).ToGo().(map[string]interface{})
	}

	// Create property definition
	property := map[string]interface{}{
		"type":        propertyType,
		"constraints": constraints,
		"created":     time.Now(),
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(property)), nil
}

func (b *SchemaBridge) validateJSON(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("expected object for schema, got %s", args[0].Type())), nil
	}
	if args[1].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("expected object for data, got %s", args[1].Type())), nil
	}

	schemaData := args[0].(engine.ObjectValue).ToGo().(map[string]interface{})
	data := args[1].(engine.ObjectValue).ToGo().(map[string]interface{})

	// Convert to go-llms schema
	schema, err := scriptToSchema(schemaData)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("invalid schema: %w", err)), nil
	}

	// Convert data to JSON string for validation
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to marshal data to JSON: %w", err)), nil
	}

	result, err := b.validator.Validate(schema, string(dataJSON))
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("validation failed: %w", err)), nil
	}

	validationResult := map[string]interface{}{
		"valid":   result.Valid,
		"errors":  validationErrorsToScript(result.Errors),
		"schema":  schemaToScript(schema),
		"data":    data,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(validationResult)), nil
}

func (b *SchemaBridge) validateStruct(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Similar to validateJSON but for struct validation
	return b.validateJSON(ctx, args)
}

func (b *SchemaBridge) generateSchemaFromType(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("expected object for type info, got %s", args[0].Type())), nil
	}

	typeInfo := args[0].(engine.ObjectValue).ToGo().(map[string]interface{})
	
	// Generate schema from type information using the reflection generator
	schema, err := b.generator.GenerateSchema(typeInfo)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to generate schema: %w", err)), nil
	}

	result := map[string]interface{}{
		"schema":    schemaToScript(schema),
		"generated": true,
		"source":    "type",
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *SchemaBridge) convertJSONSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("expected string for JSON schema, got %s", args[0].Type())), nil
	}

	jsonSchema := args[0].(engine.StringValue).Value()
	
	var schemaData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonSchema), &schemaData); err != nil {
		return engine.NewErrorValue(fmt.Errorf("invalid JSON schema: %w", err)), nil
	}

	schema, err := b.generator.GenerateSchema(schemaData)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("failed to convert JSON schema: %w", err)), nil
	}

	result := map[string]interface{}{
		"schema":    schemaToScript(schema),
		"converted": true,
		"source":    "json",
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

// Repository Operations
func (b *SchemaBridge) saveSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("expected string for schema name, got %s", args[0].Type())), nil
	}
	if args[1].Type() != engine.TypeObject {
		return engine.NewErrorValue(fmt.Errorf("expected object for schema, got %s", args[1].Type())), nil
	}

	name := args[0].(engine.StringValue).Value()
	schemaData := args[1].(engine.ObjectValue).ToGo().(map[string]interface{})

	schema, err := scriptToSchema(schemaData)
	if err != nil {
		return engine.NewErrorValue(fmt.Errorf("invalid schema: %w", err)), nil
	}

	b.mu.Lock()
	b.schemas[name] = schema
	b.mu.Unlock()

	// Also save to repository if available
	if b.repository != nil {
		if err := b.repository.Save(name, schema); err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to save to repository: %w", err)), nil
		}
	}

	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) getSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("expected string for schema name, got %s", args[0].Type())), nil
	}

	name := args[0].(engine.StringValue).Value()

	b.mu.RLock()
	schema, exists := b.schemas[name]
	b.mu.RUnlock()

	if !exists {
		// Try to load from repository
		if b.repository != nil {
			var err error
			schema, err = b.repository.Get(name)
			if err != nil {
				return engine.NewErrorValue(fmt.Errorf("schema not found: %s", name)), nil
			}
		} else {
			return engine.NewErrorValue(fmt.Errorf("schema not found: %s", name)), nil
		}
	}

	result := map[string]interface{}{
		"name":   name,
		"schema": schemaToScript(schema),
		"found":  true,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(result)), nil
}

func (b *SchemaBridge) deleteSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if args[0].Type() != engine.TypeString {
		return engine.NewErrorValue(fmt.Errorf("expected string for schema name, got %s", args[0].Type())), nil
	}

	name := args[0].(engine.StringValue).Value()

	b.mu.Lock()
	delete(b.schemas, name)
	b.mu.Unlock()

	// Also delete from repository if available
	if b.repository != nil {
		if err := b.repository.Delete(name); err != nil {
			return engine.NewErrorValue(fmt.Errorf("failed to delete from repository: %w", err)), nil
		}
	}

	return engine.NewNilValue(), nil
}

// Stub implementations for remaining methods (would be implemented similarly)
func (b *SchemaBridge) initializeFileRepository(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) saveSchemaVersion(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) getSchemaVersion(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) listSchemaVersions(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewArrayValue([]engine.ScriptValue{}), nil
}

func (b *SchemaBridge) setCurrentSchemaVersion(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) registerMigrator(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) migrateSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) exportRepository(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) importRepository(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) generateFromTags(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) setTagPriority(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) registerTagParser(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) extractValidationRules(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) generateWithDocumentation(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) exportToJSONSchema(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) exportToOpenAPI(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) importFromFile(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) importFromString(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) convertFormat(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) mergeSchemas(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) generateDiff(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) exportCollection(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) importCollection(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) registerCustomValidator(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) unregisterCustomValidator(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) listCustomValidators(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewArrayValue([]engine.ScriptValue{}), nil
}

func (b *SchemaBridge) validateWithCustom(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) validateAsync(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) getValidationMetrics(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.validationMetrics.mutex.RLock()
	defer b.validationMetrics.mutex.RUnlock()

	metrics := map[string]interface{}{
		"totalValidations":      b.validationMetrics.TotalValidations,
		"successfulValidations": b.validationMetrics.SuccessfulValidations,
		"failedValidations":     b.validationMetrics.FailedValidations,
		"averageLatency":        b.validationMetrics.AverageLatency.Milliseconds(),
		"cacheHits":             b.validationMetrics.CacheHits,
		"cacheMisses":           b.validationMetrics.CacheMisses,
		"asyncValidations":      b.validationMetrics.AsyncValidations,
	}

	return engine.NewObjectValue(engine.ConvertMapToScriptValue(metrics)), nil
}

func (b *SchemaBridge) clearValidationCache(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.validationCache = sync.Map{}
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) registerConditionalValidator(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

func (b *SchemaBridge) validateConditional(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	// Implementation would go here
	return engine.NewNilValue(), nil
}

// Helper functions

func (b *SchemaBridge) processAsyncValidations(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-b.asyncValidationQueue:
			if !ok {
				return
			}
			
			// Process async validation request
			// Convert data to JSON string for validation
			dataJSON, err := json.Marshal(req.Data)
			if err != nil {
				continue
			}
			
			result, err := b.validator.Validate(req.Schema, string(dataJSON))
			if err != nil {
				// Handle error
				continue
			}
			
			// Update metrics
			b.validationMetrics.mutex.Lock()
			b.validationMetrics.AsyncValidations++
			if result.Valid {
				b.validationMetrics.SuccessfulValidations++
			} else {
				b.validationMetrics.FailedValidations++
			}
			b.validationMetrics.mutex.Unlock()
		}
	}
}

// schemaToScript converts a go-llms schema to a script-friendly format
func schemaToScript(schema *schemaDomain.Schema) map[string]interface{} {
	if schema == nil {
		return map[string]interface{}{}
	}

	result := map[string]interface{}{
		"type": schema.Type,
	}

	if schema.Description != "" {
		result["description"] = schema.Description
	}

	if schema.Title != "" {
		result["title"] = schema.Title
	}

	if len(schema.Required) > 0 {
		// Convert []string to []interface{} for script compatibility
		required := make([]interface{}, len(schema.Required))
		for i, req := range schema.Required {
			required[i] = req
		}
		result["required"] = required
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for name, prop := range schema.Properties {
			props[name] = propertyToScript(&prop)
		}
		result["properties"] = props
	}

	return result
}

// propertyToScript converts a property to script format
func propertyToScript(prop *schemaDomain.Property) map[string]interface{} {
	result := map[string]interface{}{
		"type": prop.Type,
	}

	if prop.Description != "" {
		result["description"] = prop.Description
	}

	if prop.Format != "" {
		result["format"] = prop.Format
	}

	if prop.Pattern != "" {
		result["pattern"] = prop.Pattern
	}

	if prop.Minimum != nil {
		result["minimum"] = *prop.Minimum
	}

	if prop.Maximum != nil {
		result["maximum"] = *prop.Maximum
	}

	return result
}

// scriptToSchema converts script format to go-llms schema
func scriptToSchema(def map[string]interface{}) (*schemaDomain.Schema, error) {
	schema := &schemaDomain.Schema{}

	if t, ok := def["type"].(string); ok {
		schema.Type = t
	}

	if desc, ok := def["description"].(string); ok {
		schema.Description = desc
	}

	if title, ok := def["title"].(string); ok {
		schema.Title = title
	}

	// Handle required field - could be []interface{} or []string
	if required := def["required"]; required != nil {
		switch req := required.(type) {
		case []interface{}:
			schema.Required = make([]string, 0, len(req))
			for _, r := range req {
				if s, ok := r.(string); ok {
					schema.Required = append(schema.Required, s)
				}
			}
		case []string:
			schema.Required = req
		}
	}

	if props, ok := def["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]schemaDomain.Property)
		for name, propDef := range props {
			if propMap, ok := propDef.(map[string]interface{}); ok {
				prop, err := scriptToProperty(propMap)
				if err == nil {
					schema.Properties[name] = *prop
				}
			}
		}
	}

	return schema, nil
}

// scriptToProperty converts script format to go-llms property
func scriptToProperty(def map[string]interface{}) (*schemaDomain.Property, error) {
	prop := &schemaDomain.Property{}

	if t, ok := def["type"].(string); ok {
		prop.Type = t
	}

	if desc, ok := def["description"].(string); ok {
		prop.Description = desc
	}

	if format, ok := def["format"].(string); ok {
		prop.Format = format
	}

	if pattern, ok := def["pattern"].(string); ok {
		prop.Pattern = pattern
	}

	if min, ok := def["minimum"].(float64); ok {
		prop.Minimum = &min
	}

	if max, ok := def["maximum"].(float64); ok {
		prop.Maximum = &max
	}

	return prop, nil
}

// validationErrorsToScript converts validation errors to script format
func validationErrorsToScript(errors []string) []interface{} {
	result := make([]interface{}, len(errors))
	for i, err := range errors {
		result[i] = map[string]interface{}{
			"message": err,
			"type":    "validation_error",
		}
	}
	return result
}

// TypeMappings returns type conversion hints
func (b *SchemaBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"schema": {
			GoType:     "*schemaDomain.Schema",
			ScriptType: "object",
			Converter:  "schemaConverter",
		},
		"validationResult": {
			GoType:     "*schemaDomain.ValidationResult",
			ScriptType: "object",
			Converter:  "validationResultConverter",
		},
	}
}

// RequiredPermissions returns required permissions
func (b *SchemaBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionFileSystem,
			Resource:    "schema.files",
			Actions:     []string{"read", "write"},
			Description: "Access to schema files for persistence",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "schema.cache",
			Actions:     []string{"read", "write"},
			Description: "Cache for validation results",
		},
	}
}