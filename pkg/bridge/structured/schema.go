// ABOUTME: Schema bridge providing access to go-llms schema validation system
// ABOUTME: Wraps go-llms schema functionality for script-based validation and generation

package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
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
	mu           sync.RWMutex
	initialized  bool
	validator    schemaDomain.Validator
	generator    schemaDomain.SchemaGenerator
	repository   schemaDomain.SchemaRepository
	fileRepo     schemaDomain.SchemaRepository        // File-based repository for persistence
	migrators    map[string]repository.SchemaMigrator // Migration registry
	tagGenerator schemaDomain.SchemaGenerator         // Tag-based generator
	schemas      map[string]*schemaDomain.Schema      // Simple in-memory storage
}

// NewSchemaBridge creates a new schema bridge
func NewSchemaBridge() *SchemaBridge {
	return &SchemaBridge{
		migrators: make(map[string]repository.SchemaMigrator),
		schemas:   make(map[string]*schemaDomain.Schema),
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
	b.tagGenerator = generator.NewTagSchemaGenerator()
	// Note: fileRepo will be initialized on demand when a directory is specified

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
		// Versioning and migration methods
		{
			Name:        "initializeFileRepository",
			Description: "Initialize file-based repository for schema persistence",
			Parameters: []engine.ParameterInfo{
				{Name: "directory", Type: "string", Description: "Directory path for schema storage", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "saveSchemaVersion",
			Description: "Save a specific version of a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
				{Name: "schema", Type: "object", Description: "Schema object", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "getSchemaVersion",
			Description: "Get a specific version of a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
				{Name: "version", Type: "number", Description: "Version number", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "listSchemaVersions",
			Description: "List all versions of a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "setCurrentSchemaVersion",
			Description: "Set the current version of a schema",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
				{Name: "version", Type: "number", Description: "Version number", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "registerMigrator",
			Description: "Register a schema migrator",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Description: "Migrator name", Required: true},
				{Name: "migrator", Type: "function", Description: "Migrator function", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "migrateSchema",
			Description: "Migrate a schema between versions",
			Parameters: []engine.ParameterInfo{
				{Name: "id", Type: "string", Description: "Schema ID", Required: true},
				{Name: "fromVersion", Type: "number", Description: "Source version", Required: true},
				{Name: "toVersion", Type: "number", Description: "Target version", Required: true},
				{Name: "migratorName", Type: "string", Description: "Migrator to use", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "exportRepository",
			Description: "Export entire repository to JSON",
			Parameters:  []engine.ParameterInfo{},
			ReturnType:  "string",
		},
		{
			Name:        "importRepository",
			Description: "Import repository from JSON",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "string", Description: "JSON data to import", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "generateFromTags",
			Description: "Generate schema from struct tags",
			Parameters: []engine.ParameterInfo{
				{Name: "source", Type: "any", Description: "Struct object or type to analyze", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "setTagPriority",
			Description: "Set the order in which tags are processed",
			Parameters: []engine.ParameterInfo{
				{Name: "tags", Type: "array", Description: "Array of tag names in priority order", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "registerTagParser",
			Description: "Register a custom tag parser",
			Parameters: []engine.ParameterInfo{
				{Name: "tagName", Type: "string", Description: "Name of the tag to parse", Required: true},
				{Name: "parser", Type: "function", Description: "Parser function", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "extractValidationRules",
			Description: "Extract validation rules from struct tags",
			Parameters: []engine.ParameterInfo{
				{Name: "source", Type: "any", Description: "Struct object or type to analyze", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "generateWithDocumentation",
			Description: "Generate schema with embedded documentation from tags",
			Parameters: []engine.ParameterInfo{
				{Name: "source", Type: "any", Description: "Struct object or type to analyze", Required: true},
				{Name: "includeExamples", Type: "boolean", Description: "Include example values from tags", Required: false},
			},
			ReturnType: "object",
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
		"SchemaVersion": {
			GoType:     "SchemaVersion",
			ScriptType: "object",
		},
		"SchemaMigrator": {
			GoType:     "SchemaMigrator",
			ScriptType: "function",
		},
		"SchemaMetadata": {
			GoType:     "SchemaMetadata",
			ScriptType: "object",
		},
		"TagParser": {
			GoType:     "TagParser",
			ScriptType: "function",
		},
		"ValidationRule": {
			GoType:     "ValidationRule",
			ScriptType: "object",
		},
		"ValidationExtractor": {
			GoType:     "ValidationExtractor",
			ScriptType: "function",
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

		// Use file repo if available, otherwise memory repo
		repo := b.repository
		if b.fileRepo != nil {
			repo = b.fileRepo
		}

		schema, err := repo.Get(id)
		if err != nil {
			// If file repo failed, try memory repo as fallback
			if b.fileRepo != nil && repo != b.repository {
				schema, err = b.repository.Get(id)
				if err != nil {
					return nil, fmt.Errorf("schema not found: %w", err)
				}
			} else {
				return nil, fmt.Errorf("schema not found: %w", err)
			}
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

	case "initializeFileRepository":
		if len(args) < 1 {
			return nil, fmt.Errorf("initializeFileRepository requires directory parameter")
		}

		directory, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("directory must be string")
		}

		// Create file-based repository
		fileRepo, err := repository.NewFileSchemaRepository(directory)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file repository: %w", err)
		}

		b.fileRepo = fileRepo
		return nil, nil

	case "saveSchemaVersion":
		if len(args) < 2 {
			return nil, fmt.Errorf("saveSchemaVersion requires id and schema parameters")
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

		// Use file repo if available, otherwise memory repo
		repo := b.repository
		if b.fileRepo != nil {
			repo = b.fileRepo
		}

		if err := repo.Save(id, schema); err != nil {
			return nil, fmt.Errorf("failed to save schema version: %w", err)
		}

		// Get version info if available from FileSchemaRepository metadata
		if fileRepo, ok := repo.(*repository.FileSchemaRepository); ok {
			// Read metadata file directly
			metadataPath := fmt.Sprintf("%s/%s/metadata.json", b.getBasePath(fileRepo), id)
			if data, err := os.ReadFile(metadataPath); err == nil {
				var metadata repository.SchemaMetadata
				if json.Unmarshal(data, &metadata) == nil {
					return map[string]interface{}{
						"id":             metadata.ID,
						"latestVersion":  metadata.LatestVersion,
						"currentVersion": metadata.CurrentVersion,
						"totalVersions":  metadata.TotalVersions,
					}, nil
				}
			}
		}

		return map[string]interface{}{"saved": true}, nil

	case "getSchemaVersion":
		if len(args) < 2 {
			return nil, fmt.Errorf("getSchemaVersion requires id and version parameters")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		version, err := getIntValue(args[1])
		if err != nil {
			return nil, fmt.Errorf("version must be number: %w", err)
		}

		// Use file repo if available
		repo := b.repository
		if b.fileRepo != nil {
			repo = b.fileRepo
		}

		// Check if repo supports versioning (both file and memory repos have GetVersion)
		if versionedRepo, ok := repo.(interface {
			GetVersion(string, int) (*schemaDomain.Schema, error)
		}); ok {
			schema, err := versionedRepo.GetVersion(id, version)
			if err != nil {
				return nil, fmt.Errorf("failed to get schema version: %w", err)
			}
			return schemaToScript(schema), nil
		}

		return nil, fmt.Errorf("repository does not support versioning")

	case "listSchemaVersions":
		if len(args) < 1 {
			return nil, fmt.Errorf("listSchemaVersions requires id parameter")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		// Use file repo if available
		repo := b.repository
		if b.fileRepo != nil {
			repo = b.fileRepo
		}

		// Check if repo supports versioning - file repo has ListVersions, memory repo has ListVersions
		if fileRepo, ok := repo.(*repository.FileSchemaRepository); ok {
			versions, err := fileRepo.ListVersions(id)
			if err != nil {
				return nil, fmt.Errorf("failed to list schema versions: %w", err)
			}
			return versions, nil
		}

		// For memory repo, get from versions
		if memRepo, ok := repo.(*repository.InMemorySchemaRepository); ok {
			versions, err := memRepo.ListVersions(id)
			if err != nil {
				return nil, fmt.Errorf("failed to list schema versions: %w", err)
			}
			// Extract just the version numbers
			var versionNumbers []int
			for _, v := range versions {
				versionNumbers = append(versionNumbers, v.Version)
			}
			return versionNumbers, nil
		}

		return nil, fmt.Errorf("repository does not support versioning")

	case "setCurrentSchemaVersion":
		if len(args) < 2 {
			return nil, fmt.Errorf("setCurrentSchemaVersion requires id and version parameters")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		version, err := getIntValue(args[1])
		if err != nil {
			return nil, fmt.Errorf("version must be number: %w", err)
		}

		// Use file repo if available
		repo := b.repository
		if b.fileRepo != nil {
			repo = b.fileRepo
		}

		// Check if repo supports versioning
		if fileRepo, ok := repo.(*repository.FileSchemaRepository); ok {
			if err := fileRepo.SetCurrentVersion(id, version); err != nil {
				return nil, fmt.Errorf("failed to set current version: %w", err)
			}
			return nil, nil
		}

		// For memory repo
		if memRepo, ok := repo.(*repository.InMemorySchemaRepository); ok {
			if err := memRepo.SetCurrentVersion(id, version); err != nil {
				return nil, fmt.Errorf("failed to set current version: %w", err)
			}
			return nil, nil
		}

		return nil, fmt.Errorf("repository does not support versioning")

	case "registerMigrator":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerMigrator requires name and migrator parameters")
		}

		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("name must be string")
		}

		// Create a script-based migrator wrapper
		migrator := &scriptMigrator{
			scriptFunc: args[1],
		}

		b.migrators[name] = migrator
		return nil, nil

	case "migrateSchema":
		if len(args) < 3 {
			return nil, fmt.Errorf("migrateSchema requires id, fromVersion, and toVersion parameters")
		}

		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("id must be string")
		}

		fromVersion, err := getIntValue(args[1])
		if err != nil {
			return nil, fmt.Errorf("fromVersion must be number: %w", err)
		}

		toVersion, err := getIntValue(args[2])
		if err != nil {
			return nil, fmt.Errorf("toVersion must be number: %w", err)
		}

		// Get migrator name if provided
		migratorName := "default"
		if len(args) > 3 {
			if name, ok := args[3].(string); ok {
				migratorName = name
			}
		}

		// Use file repo if available
		repo := b.repository
		if b.fileRepo != nil {
			repo = b.fileRepo
		}

		// Check if repo supports migration (only file repo has Migrate method)
		if fileRepo, ok := repo.(*repository.FileSchemaRepository); ok {
			migrator, exists := b.migrators[migratorName]
			if !exists {
				return nil, fmt.Errorf("migrator '%s' not found", migratorName)
			}

			if err := fileRepo.Migrate(id, fromVersion, toVersion, migrator); err != nil {
				return nil, fmt.Errorf("migration failed: %w", err)
			}

			// Get the migrated schema
			schema, err := fileRepo.GetVersion(id, toVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to get migrated schema: %w", err)
			}

			return schemaToScript(schema), nil
		}

		return nil, fmt.Errorf("repository does not support migration")

	case "exportRepository":
		// Check if repo supports export
		if memRepo, ok := b.repository.(*repository.InMemorySchemaRepository); ok {
			data, err := memRepo.Export()
			if err != nil {
				return nil, fmt.Errorf("failed to export repository: %w", err)
			}
			return string(data), nil
		}

		return nil, fmt.Errorf("repository does not support export")

	case "importRepository":
		if len(args) < 1 {
			return nil, fmt.Errorf("importRepository requires data parameter")
		}

		data, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("data must be string")
		}

		// Check if repo supports import
		if memRepo, ok := b.repository.(*repository.InMemorySchemaRepository); ok {
			if err := memRepo.Import([]byte(data)); err != nil {
				return nil, fmt.Errorf("failed to import repository: %w", err)
			}
			return nil, nil
		}

		return nil, fmt.Errorf("repository does not support import")

	case "generateFromTags":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateFromTags requires source parameter")
		}

		// Try to use the source directly - scripts would pass Go objects
		schema, err := b.tagGenerator.GenerateSchema(args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema from tags: %w", err)
		}

		return schemaToScript(schema), nil

	case "setTagPriority":
		if len(args) < 1 {
			return nil, fmt.Errorf("setTagPriority requires tags parameter")
		}

		tags, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("tags must be array")
		}

		// Convert to string slice
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagStr, ok := tag.(string)
			if !ok {
				return nil, fmt.Errorf("all tag names must be strings")
			}
			tagNames[i] = tagStr
		}

		// Set tag priority on the generator
		if tagGen, ok := b.tagGenerator.(*generator.TagSchemaGenerator); ok {
			tagGen.SetTagPriority(tagNames)
		} else {
			return nil, fmt.Errorf("tag generator does not support priority setting")
		}

		return nil, nil

	case "registerTagParser":
		if len(args) < 2 {
			return nil, fmt.Errorf("registerTagParser requires tagName and parser parameters")
		}

		tagName, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("tagName must be string")
		}

		// Create a wrapper for the script parser
		scriptParser := &scriptTagParser{
			scriptFunc: args[1],
		}

		// Register with the tag generator
		if tagGen, ok := b.tagGenerator.(*generator.TagSchemaGenerator); ok {
			tagGen.RegisterTagParser(tagName, scriptParser.Parse)
		} else {
			return nil, fmt.Errorf("tag generator does not support custom parsers")
		}

		return nil, nil

	case "extractValidationRules":
		if len(args) < 1 {
			return nil, fmt.Errorf("extractValidationRules requires source parameter")
		}

		// Extract validation rules using reflection
		rules, err := b.extractValidationRulesFromStruct(args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to extract validation rules: %w", err)
		}

		return rules, nil

	case "generateWithDocumentation":
		if len(args) < 1 {
			return nil, fmt.Errorf("generateWithDocumentation requires source parameter")
		}

		includeExamples := false
		if len(args) > 1 {
			if examples, ok := args[1].(bool); ok {
				includeExamples = examples
			}
		}

		// Generate schema with enhanced documentation
		schema, err := b.generateSchemaWithDocumentation(args[0], includeExamples)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema with documentation: %w", err)
		}

		return schemaToScript(schema), nil

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

// getIntValue converts a numeric value to int
func getIntValue(val interface{}) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("value must be numeric, got %T", val)
	}
}

// scriptMigrator wraps a script function as a SchemaMigrator
type scriptMigrator struct {
	scriptFunc interface{}
}

// Migrate implements repository.SchemaMigrator
func (m *scriptMigrator) Migrate(schema *schemaDomain.Schema, fromVersion, toVersion int) (*schemaDomain.Schema, error) {
	// This would need to be implemented based on the script engine
	// For now, return an error indicating script execution is needed
	return nil, fmt.Errorf("script-based migration not yet implemented - requires script engine integration")
}

// getBasePath extracts the base path from a FileSchemaRepository using reflection
func (b *SchemaBridge) getBasePath(repo *repository.FileSchemaRepository) string {
	v := reflect.ValueOf(repo).Elem()
	basePathField := v.FieldByName("basePath")
	if !basePathField.IsValid() {
		return ""
	}
	return basePathField.String()
}

// scriptTagParser wraps a script function as a TagParser
type scriptTagParser struct {
	scriptFunc interface{}
}

// Parse implements generator.TagParser (placeholder for script engine integration)
func (p *scriptTagParser) Parse(tagValue string, prop *schemaDomain.Property) error {
	// This would need to be implemented based on the script engine
	// For now, return an error indicating script execution is needed
	return fmt.Errorf("script-based tag parsing not yet implemented - requires script engine integration")
}

// extractValidationRulesFromStruct extracts validation rules from a struct using reflection
func (b *SchemaBridge) extractValidationRulesFromStruct(source interface{}) ([]interface{}, error) {
	t := reflect.TypeOf(source)
	if t == nil {
		return nil, fmt.Errorf("cannot extract rules from nil")
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only handle structs
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("validation rule extraction only supports structs, got %s", t.Kind())
	}

	var rules []interface{}

	// Process each field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Extract rules from tags
		fieldRules := b.extractRulesFromField(field)
		if len(fieldRules) > 0 {
			ruleMap := map[string]interface{}{
				"field": field.Name,
				"rules": fieldRules,
			}
			rules = append(rules, ruleMap)
		}
	}

	return rules, nil
}

// extractRulesFromField extracts validation rules from a specific field
func (b *SchemaBridge) extractRulesFromField(field reflect.StructField) []interface{} {
	var rules []interface{}

	// Check common validation tags
	validationTags := []string{"validate", "binding", "schema"}

	for _, tagName := range validationTags {
		tagValue := field.Tag.Get(tagName)
		if tagValue == "" {
			continue
		}

		// Parse tag value and extract rules
		tagRules := b.parseValidationTag(tagName, tagValue)
		rules = append(rules, tagRules...)
	}

	return rules
}

// parseValidationTag parses a validation tag and returns rules
func (b *SchemaBridge) parseValidationTag(tagName, tagValue string) []interface{} {
	var rules []interface{}

	switch tagName {
	case "validate", "binding":
		for _, rule := range strings.Split(tagValue, ",") {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			ruleObj := map[string]interface{}{
				"type":   rule,
				"source": tagName,
			}

			// Handle parameterized rules
			if idx := strings.Index(rule, "="); idx > 0 {
				ruleObj["type"] = rule[:idx]
				ruleObj["value"] = rule[idx+1:]
			}

			rules = append(rules, ruleObj)
		}

	case "schema":
		for _, part := range strings.Split(tagValue, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			if idx := strings.Index(part, "="); idx > 0 {
				ruleObj := map[string]interface{}{
					"type":   part[:idx],
					"value":  part[idx+1:],
					"source": tagName,
				}
				rules = append(rules, ruleObj)
			}
		}
	}

	return rules
}

// generateSchemaWithDocumentation generates a schema with enhanced documentation
func (b *SchemaBridge) generateSchemaWithDocumentation(source interface{}, includeExamples bool) (*schemaDomain.Schema, error) {
	// First generate the basic schema
	schema, err := b.tagGenerator.GenerateSchema(source)
	if err != nil {
		return nil, err
	}

	// Enhance with additional documentation
	t := reflect.TypeOf(source)
	if t == nil {
		return schema, nil
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return schema, nil
	}

	// Enhance properties with additional documentation
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Get field name (same logic as tag generator)
		fieldName := b.getFieldNameForDoc(field)
		if fieldName == "" {
			continue
		}

		// Enhance property if it exists
		if prop, exists := schema.Properties[fieldName]; exists {
			enhancedProp := b.enhancePropertyDocumentation(field, prop, includeExamples)
			schema.Properties[fieldName] = enhancedProp
		}
	}

	// Add schema-level documentation
	b.enhanceSchemaDocumentation(t, schema)

	return schema, nil
}

// getFieldNameForDoc gets the field name for documentation (similar to tag generator logic)
func (b *SchemaBridge) getFieldNameForDoc(field reflect.StructField) string {
	// Check JSON tag first
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] == "-" {
			return ""
		}
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Check schema tag
	if schemaTag := field.Tag.Get("schema"); schemaTag != "" {
		for _, part := range strings.Split(schemaTag, ",") {
			if strings.HasPrefix(part, "name=") {
				return strings.TrimPrefix(part, "name=")
			}
		}
	}

	return field.Name
}

// enhancePropertyDocumentation enhances a property with additional documentation
func (b *SchemaBridge) enhancePropertyDocumentation(field reflect.StructField, prop schemaDomain.Property, includeExamples bool) schemaDomain.Property {
	// Add examples if requested
	if includeExamples {
		if example := field.Tag.Get("example"); example != "" {
			// Note: JSON Schema examples would go here if supported
			if prop.Description != "" {
				prop.Description += fmt.Sprintf(" (Example: %s)", example)
			} else {
				prop.Description = fmt.Sprintf("Example: %s", example)
			}
		}
	}

	// Add deprecation info
	if deprecated := field.Tag.Get("deprecated"); deprecated != "" {
		if prop.Description != "" {
			prop.Description += " [DEPRECATED"
			if deprecated != "true" {
				prop.Description += ": " + deprecated
			}
			prop.Description += "]"
		} else {
			prop.Description = "DEPRECATED"
			if deprecated != "true" {
				prop.Description += ": " + deprecated
			}
		}
	}

	// Add version info
	if version := field.Tag.Get("version"); version != "" {
		if prop.Description != "" {
			prop.Description += fmt.Sprintf(" (Since: %s)", version)
		} else {
			prop.Description = fmt.Sprintf("Since: %s", version)
		}
	}

	return prop
}

// enhanceSchemaDocumentation enhances schema-level documentation
func (b *SchemaBridge) enhanceSchemaDocumentation(t reflect.Type, schema *schemaDomain.Schema) {
	// Create comprehensive description
	if schema.Description == "" {
		pkg := t.PkgPath()
		if pkg != "" {
			schema.Description = fmt.Sprintf("Generated from %s.%s", pkg, t.Name())
		} else {
			schema.Description = fmt.Sprintf("Generated from %s", t.Name())
		}
	}

	// Add version info if available (would come from build tags or constants)
	schema.Description += " - Generated by go-llmspell tag-based schema generator"
}
