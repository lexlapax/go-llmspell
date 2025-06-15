// ABOUTME: Schema bridge providing access to go-llms schema validation system
// ABOUTME: Wraps go-llms schema functionality for script-based validation and generation

package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for schema functionality
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/generator"
	"github.com/lexlapax/go-llms/pkg/schema/repository"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// SchemaBridge provides access to go-llms schema validation system
type SchemaBridge struct {
	mu          sync.RWMutex
	initialized bool
	validator   schemaDomain.Validator
	generator   schemaDomain.SchemaGenerator
	repository  schemaDomain.SchemaRepository
	schemas     map[string]*schemaDomain.Schema // Simple in-memory storage
}

// NewSchemaBridge creates a new schema bridge
func NewSchemaBridge() *SchemaBridge {
	return &SchemaBridge{}
}

// GetID returns the bridge ID
func (b *SchemaBridge) GetID() string {
	return "schema"
}

// GetMetadata returns bridge metadata
func (b *SchemaBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "Schema Bridge",
		Version:     "1.0.0",
		Description: "Provides access to go-llms schema validation and generation system",
		Author:      "go-llmspell",
	}
}

// Initialize initializes the bridge
func (b *SchemaBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize schema components from go-llms
	b.validator = validation.NewValidator()
	b.generator = generator.NewReflectionSchemaGenerator()
	b.repository = repository.NewInMemorySchemaRepository()

	b.initialized = true
	return nil
}

// Cleanup performs cleanup
func (b *SchemaBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	b.validator = nil
	b.generator = nil
	b.repository = nil
	b.schemas = nil
	return nil
}

// IsInitialized checks if the bridge is initialized
func (b *SchemaBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine
func (b *SchemaBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge
func (b *SchemaBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Schema creation and manipulation
		{
			Name:        "createSchema",
			Description: "Create a new schema object",
			Parameters: []engine.ParameterInfo{
				{Name: "definition", Type: "object", Description: "Schema definition", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "createProperty",
			Description: "Create a property definition",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Description: "Property type", Required: true},
				{Name: "options", Type: "object", Description: "Property options", Required: false},
			},
			ReturnType: "object",
		},
		// Validation
		{
			Name:        "validateJSON",
			Description: "Validate JSON data against a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "schema", Type: "object", Description: "Schema object", Required: true},
				{Name: "data", Type: "string", Description: "JSON data to validate", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateStruct",
			Description: "Validate a struct against a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "schema", Type: "object", Description: "Schema object", Required: true},
				{Name: "struct", Type: "object", Description: "Struct to validate", Required: true},
			},
			ReturnType: "object",
		},
		// Schema generation
		{
			Name:        "generateSchemaFromType",
			Description: "Generate schema from a type definition",
			Parameters: []engine.ParameterInfo{
				{Name: "typeInfo", Type: "object", Description: "Type information", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "convertJSONSchema",
			Description: "Convert JSON Schema string to schema object",
			Parameters: []engine.ParameterInfo{
				{Name: "jsonSchema", Type: "string", Description: "JSON Schema string", Required: true},
			},
			ReturnType: "object",
		},
		// Repository operations
		{
			Name:        "saveSchema",
			Description: "Save a schema to the repository",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
				{Name: "schema", Type: "object", Description: "Schema object", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getSchema",
			Description: "Get a schema from the repository",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "deleteSchema",
			Description: "Delete a schema from the repository",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
			},
			ReturnType: "void",
		},
	}
}

// TypeMappings returns type conversion mappings
func (b *SchemaBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"Schema": {
			GoType:     "Schema",
			ScriptType: "object",
		},
		"Property": {
			GoType:     "Property",
			ScriptType: "object",
		},
		"ValidationResult": {
			GoType:     "ValidationResult",
			ScriptType: "object",
		},
		"Validator": {
			GoType:     "Validator",
			ScriptType: "object",
		},
		"SchemaRepository": {
			GoType:     "SchemaRepository",
			ScriptType: "object",
		},
		"SchemaGenerator": {
			GoType:     "SchemaGenerator",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls
func (b *SchemaBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions
func (b *SchemaBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "schema",
			Actions:     []string{"create", "validate", "store"},
			Description: "Schema creation and validation",
		},
	}
}

// ExecuteMethod executes a bridge method
func (b *SchemaBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, fmt.Errorf("bridge not initialized")
	}

	switch name {
	case "createSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("createSchema requires definition parameter")
		}
		def, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("definition must be object")
		}

		schema, err := scriptToSchema(def)
		if err != nil {
			return nil, fmt.Errorf("failed to create schema: %w", err)
		}

		return schemaToScript(schema), nil

	case "createProperty":
		if len(args) < 1 {
			return nil, fmt.Errorf("createProperty requires type parameter")
		}
		propType, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("type must be string")
		}

		prop := &schemaDomain.Property{
			Type: propType,
		}

		// Apply options if provided
		if len(args) > 1 {
			if options, ok := args[1].(map[string]interface{}); ok {
				applyPropertyOptions(prop, options)
			}
		}

		return propertyToScript(prop), nil

	case "validateJSON":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateJSON requires schema and data parameters")
		}

		schemaDef, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be object")
		}

		data, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("data must be string")
		}

		schema, err := scriptToSchema(schemaDef)
		if err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}

		result, err := b.validator.Validate(schema, data)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		return validationResultToScript(result), nil

	case "validateStruct":
		if len(args) < 2 {
			return nil, fmt.Errorf("validateStruct requires schema and struct parameters")
		}

		schemaDef, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be object")
		}

		structData := args[1]

		schema, err := scriptToSchema(schemaDef)
		if err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}

		result, err := b.validator.ValidateStruct(schema, structData)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		return validationResultToScript(result), nil

	case "generateSchemaFromType":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateSchemaFromType requires typeInfo parameter")
		}

		// This would need more sophisticated type handling
		// For now, return a basic implementation
		return nil, fmt.Errorf("schema generation not yet implemented")

	case "convertJSONSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("convertJSONSchema requires jsonSchema parameter")
		}

		jsonSchema, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("jsonSchema must be string")
		}

		var schema schemaDomain.Schema
		if err := json.Unmarshal([]byte(jsonSchema), &schema); err != nil {
			return nil, fmt.Errorf("invalid JSON schema: %w", err)
		}

		return schemaToScript(&schema), nil

	case "saveSchema":
		if len(args) < 2 {
			return nil, fmt.Errorf("saveSchema requires id and schema parameters")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		schemaDef, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be object")
		}

		schema, err := scriptToSchema(schemaDef)
		if err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}

		if err := b.repository.Save(id, schema); err != nil {
			return nil, fmt.Errorf("failed to save schema: %w", err)
		}

		return nil, nil

	case "getSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("getSchema requires id parameter")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		schema, err := b.repository.Get(id)
		if err != nil {
			return nil, fmt.Errorf("schema not found: %w", err)
		}

		return schemaToScript(schema), nil

	case "deleteSchema":
		if len(args) < 1 {
			return nil, fmt.Errorf("deleteSchema requires id parameter")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		if err := b.repository.Delete(id); err != nil {
			return nil, fmt.Errorf("failed to delete schema: %w", err)
		}

		return nil, nil

	default:
		return nil, fmt.Errorf("method not found: %s", name)
	}
}

// Helper functions for type conversion

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
				if err != nil {
					return nil, fmt.Errorf("invalid property %s: %w", name, err)
				}
				schema.Properties[name] = *prop
			}
		}
	}

	if addProps, ok := def["additionalProperties"].(bool); ok {
		schema.AdditionalProperties = &addProps
	}

	return schema, nil
}

func scriptToProperty(def map[string]interface{}) (*schemaDomain.Property, error) {
	prop := &schemaDomain.Property{}

	if t, ok := def["type"].(string); ok {
		prop.Type = t
	}

	applyPropertyOptions(prop, def)

	return prop, nil
}

func applyPropertyOptions(prop *schemaDomain.Property, options map[string]interface{}) {
	if desc, ok := options["description"].(string); ok {
		prop.Description = desc
	}

	if format, ok := options["format"].(string); ok {
		prop.Format = format
	}

	if pattern, ok := options["pattern"].(string); ok {
		prop.Pattern = pattern
	}

	// Numeric constraints
	if min := getNumericValue(options["minimum"]); min != nil {
		prop.Minimum = min
	}

	if max := getNumericValue(options["maximum"]); max != nil {
		prop.Maximum = max
	}

	// String constraints
	if minLen := getNumericValue(options["minLength"]); minLen != nil {
		intVal := int(*minLen)
		prop.MinLength = &intVal
	}

	if maxLen := getNumericValue(options["maxLength"]); maxLen != nil {
		intVal := int(*maxLen)
		prop.MaxLength = &intVal
	}

	// Array constraints
	if minItems := getNumericValue(options["minItems"]); minItems != nil {
		intVal := int(*minItems)
		prop.MinItems = &intVal
	}

	if maxItems := getNumericValue(options["maxItems"]); maxItems != nil {
		intVal := int(*maxItems)
		prop.MaxItems = &intVal
	}

	if unique, ok := options["uniqueItems"].(bool); ok {
		prop.UniqueItems = &unique
	}

	// Enum values
	if enum, ok := options["enum"].([]interface{}); ok {
		prop.Enum = make([]string, 0, len(enum))
		for _, e := range enum {
			if s, ok := e.(string); ok {
				prop.Enum = append(prop.Enum, s)
			}
		}
	}
}

func schemaToScript(schema *schemaDomain.Schema) map[string]interface{} {
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
		result["required"] = schema.Required
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for name, prop := range schema.Properties {
			props[name] = propertyToScript(&prop)
		}
		result["properties"] = props
	}

	if schema.AdditionalProperties != nil {
		result["additionalProperties"] = *schema.AdditionalProperties
	}

	return result
}

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

	if prop.MinLength != nil {
		result["minLength"] = float64(*prop.MinLength)
	}

	if prop.MaxLength != nil {
		result["maxLength"] = float64(*prop.MaxLength)
	}

	if prop.MinItems != nil {
		result["minItems"] = float64(*prop.MinItems)
	}

	if prop.MaxItems != nil {
		result["maxItems"] = float64(*prop.MaxItems)
	}

	if prop.UniqueItems != nil {
		result["uniqueItems"] = *prop.UniqueItems
	}

	if len(prop.Enum) > 0 {
		result["enum"] = prop.Enum
	}

	return result
}

func validationResultToScript(result *schemaDomain.ValidationResult) map[string]interface{} {
	scriptResult := map[string]interface{}{
		"valid": result.Valid,
	}

	if len(result.Errors) > 0 {
		scriptResult["errors"] = result.Errors
	}

	return scriptResult
}

// ConvertJSONSchemaToSchema converts a JSON schema string or object to domain.Schema
// This is a public helper function for use by other bridges (like tools)
func ConvertJSONSchemaToSchema(jsonSchema interface{}) (*schemaDomain.Schema, error) {
	switch v := jsonSchema.(type) {
	case string:
		var schema schemaDomain.Schema
		if err := json.Unmarshal([]byte(v), &schema); err != nil {
			return nil, fmt.Errorf("invalid JSON schema string: %w", err)
		}
		return &schema, nil

	case map[string]interface{}:
		return scriptToSchema(v)

	default:
		return nil, fmt.Errorf("jsonSchema must be string or object, got %T", jsonSchema)
	}
}

// getNumericValue extracts a numeric value that could be int, float64, or other numeric types
func getNumericValue(val interface{}) *float64 {
	switch v := val.(type) {
	case int:
		f := float64(v)
		return &f
	case int32:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case float32:
		f := float64(v)
		return &f
	case float64:
		return &v
	default:
		return nil
	}
}
