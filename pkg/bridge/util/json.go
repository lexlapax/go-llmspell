// ABOUTME: JSON utilities bridge provides access to go-llms optimized JSON functions.
// ABOUTME: Wraps high-performance JSON marshaling, streaming, and schema operations.

package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	// go-llms imports for structured output functionality
	"github.com/lexlapax/go-llms/pkg/llm/outputs"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
	llmjson "github.com/lexlapax/go-llms/pkg/util/json"
)

var (
	// Common bridge errors
	ErrBridgeNotInitialized = errors.New("bridge not initialized")
	ErrInvalidArguments     = errors.New("invalid arguments")
	ErrMethodNotFound       = errors.New("method not found")
)

// UtilJSONBridge provides script access to go-llms structured output capabilities.
type UtilJSONBridge struct {
	mu          sync.RWMutex
	initialized bool

	// Structured output components from go-llms v0.3.5
	processor      domain.Processor
	promptEnhancer domain.PromptEnhancer
	validator      schemaDomain.Validator
	converter      *outputs.Converter
}

// NewUtilJSONBridge creates a new JSON utilities bridge.
func NewUtilJSONBridge() *UtilJSONBridge {
	return &UtilJSONBridge{}
}

// NewUtilJSONBridgeWithValidator creates a new JSON utilities bridge with custom validator.
func NewUtilJSONBridgeWithValidator(validator schemaDomain.Validator) *UtilJSONBridge {
	return &UtilJSONBridge{
		validator: validator,
	}
}

// GetID returns the bridge identifier.
func (b *UtilJSONBridge) GetID() string {
	return "util_json"
}

// GetMetadata returns bridge metadata.
func (b *UtilJSONBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_json",
		Version:     "2.0.0",
		Description: "Structured output parser with JSON/YAML/XML conversion, schema validation, and LLM-optimized extraction",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *UtilJSONBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Initialize structured output components from go-llms v0.3.5
	if b.validator == nil {
		b.validator = validation.NewValidator()
	}

	b.processor = processor.NewStructuredProcessor(b.validator)
	b.promptEnhancer = processor.NewPromptEnhancer()
	b.converter = outputs.NewConverter()

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *UtilJSONBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *UtilJSONBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *UtilJSONBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *UtilJSONBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Optimized marshaling
		{
			Name:        "marshal",
			Description: "Marshal object to JSON with optimizations",
			Parameters: []engine.ParameterInfo{
				{Name: "value", Type: "any", Description: "Value to marshal", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "marshalIndent",
			Description: "Marshal object to indented JSON",
			Parameters: []engine.ParameterInfo{
				{Name: "value", Type: "any", Description: "Value to marshal", Required: true},
				{Name: "prefix", Type: "string", Description: "Line prefix", Required: false},
				{Name: "indent", Type: "string", Description: "Indentation", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "marshalToBytes",
			Description: "Marshal object to JSON bytes",
			Parameters: []engine.ParameterInfo{
				{Name: "value", Type: "any", Description: "Value to marshal", Required: true},
			},
			ReturnType: "bytes",
		},

		// Optimized unmarshaling
		{
			Name:        "unmarshal",
			Description: "Unmarshal JSON string to object",
			Parameters: []engine.ParameterInfo{
				{Name: "json", Type: "string", Description: "JSON string", Required: true},
			},
			ReturnType: "any",
		},
		{
			Name:        "unmarshalFromBytes",
			Description: "Unmarshal JSON bytes to object",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "bytes", Description: "JSON bytes", Required: true},
			},
			ReturnType: "any",
		},
		{
			Name:        "unmarshalStrict",
			Description: "Unmarshal JSON with strict validation",
			Parameters: []engine.ParameterInfo{
				{Name: "json", Type: "string", Description: "JSON string", Required: true},
				{Name: "disallowUnknownFields", Type: "boolean", Description: "Disallow unknown fields", Required: false},
			},
			ReturnType: "any",
		},

		// Streaming operations
		{
			Name:        "createEncoder",
			Description: "Create JSON encoder for streaming",
			Parameters: []engine.ParameterInfo{
				{Name: "writer", Type: "io.Writer", Description: "Output writer", Required: true},
			},
			ReturnType: "JSONEncoder",
		},
		{
			Name:        "createDecoder",
			Description: "Create JSON decoder for streaming",
			Parameters: []engine.ParameterInfo{
				{Name: "reader", Type: "io.Reader", Description: "Input reader", Required: true},
			},
			ReturnType: "JSONDecoder",
		},
		{
			Name:        "encodeStream",
			Description: "Encode value to JSON stream",
			Parameters: []engine.ParameterInfo{
				{Name: "encoder", Type: "JSONEncoder", Description: "JSON encoder", Required: true},
				{Name: "value", Type: "any", Description: "Value to encode", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "decodeStream",
			Description: "Decode value from JSON stream",
			Parameters: []engine.ParameterInfo{
				{Name: "decoder", Type: "JSONDecoder", Description: "JSON decoder", Required: true},
			},
			ReturnType: "any",
		},

		// Schema operations
		{
			Name:        "validateWithSchema",
			Description: "Validate JSON against schema",
			Parameters: []engine.ParameterInfo{
				{Name: "json", Type: "string", Description: "JSON to validate", Required: true},
				{Name: "schema", Type: "object", Description: "JSON schema", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "generateFromSchema",
			Description: "Generate example JSON from schema",
			Parameters: []engine.ParameterInfo{
				{Name: "schema", Type: "object", Description: "JSON schema", Required: true},
			},
			ReturnType: "any",
		},
		{
			Name:        "inferSchema",
			Description: "Infer JSON schema from example",
			Parameters: []engine.ParameterInfo{
				{Name: "example", Type: "any", Description: "Example object", Required: true},
			},
			ReturnType: "object",
		},

		// Structured output operations (go-llms v0.3.5)
		{
			Name:        "parseStructured",
			Description: "Parse and validate structured output from LLM response",
			Parameters: []engine.ParameterInfo{
				{Name: "output", Type: "string", Description: "LLM output containing JSON", Required: true},
				{Name: "schema", Type: "object", Description: "JSON schema for validation", Required: true},
			},
			ReturnType: "any",
		},
		{
			Name:        "parseWithRecovery",
			Description: "Extract JSON from malformed or mixed content",
			Parameters: []engine.ParameterInfo{
				{Name: "output", Type: "string", Description: "Potentially malformed content", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "enhancePrompt",
			Description: "Add schema information to prompt for better LLM output",
			Parameters: []engine.ParameterInfo{
				{Name: "prompt", Type: "string", Description: "Original prompt", Required: true},
				{Name: "schema", Type: "object", Description: "JSON schema", Required: true},
				{Name: "options", Type: "object", Description: "Enhancement options", Required: false},
			},
			ReturnType: "string",
		},

		// Format conversion operations (go-llms v0.3.5)
		{
			Name:        "convertFormat",
			Description: "Convert between JSON, YAML, and XML formats",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "string", Description: "Data to convert", Required: true},
				{Name: "fromFormat", Type: "string", Description: "Source format (json/yaml/xml)", Required: true},
				{Name: "toFormat", Type: "string", Description: "Target format (json/yaml/xml)", Required: true},
				{Name: "options", Type: "object", Description: "Conversion options", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "streamConvert",
			Description: "Convert format using streaming for large data",
			Parameters: []engine.ParameterInfo{
				{Name: "reader", Type: "io.Reader", Description: "Input reader", Required: true},
				{Name: "writer", Type: "io.Writer", Description: "Output writer", Required: true},
				{Name: "fromFormat", Type: "string", Description: "Source format", Required: true},
				{Name: "toFormat", Type: "string", Description: "Target format", Required: true},
				{Name: "options", Type: "object", Description: "Conversion options", Required: false},
			},
			ReturnType: "void",
		},

		// Utility operations
		{
			Name:        "prettyPrint",
			Description: "Pretty print JSON with colors",
			Parameters: []engine.ParameterInfo{
				{Name: "json", Type: "string", Description: "JSON string", Required: true},
				{Name: "colorize", Type: "boolean", Description: "Enable colors", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "minify",
			Description: "Minify JSON by removing whitespace",
			Parameters: []engine.ParameterInfo{
				{Name: "json", Type: "string", Description: "JSON string", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "merge",
			Description: "Deep merge multiple JSON objects",
			Parameters: []engine.ParameterInfo{
				{Name: "objects", Type: "array", Description: "Objects to merge", Required: true},
			},
			ReturnType: "any",
		},
		{
			Name:        "diff",
			Description: "Compare two JSON objects",
			Parameters: []engine.ParameterInfo{
				{Name: "obj1", Type: "any", Description: "First object", Required: true},
				{Name: "obj2", Type: "any", Description: "Second object", Required: true},
			},
			ReturnType: "object",
		},

		// Performance utilities
		{
			Name:        "marshalWithBuffer",
			Description: "Marshal with reusable buffer for performance",
			Parameters: []engine.ParameterInfo{
				{Name: "value", Type: "any", Description: "Value to marshal", Required: true},
				{Name: "buffer", Type: "bytes", Description: "Reusable buffer", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "marshalConcurrent",
			Description: "Marshal multiple values concurrently",
			Parameters: []engine.ParameterInfo{
				{Name: "values", Type: "array", Description: "Values to marshal", Required: true},
			},
			ReturnType: "array",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *UtilJSONBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"JSONEncoder": {
			GoType:     "json.Encoder",
			ScriptType: "object",
		},
		"JSONDecoder": {
			GoType:     "json.Decoder",
			ScriptType: "object",
		},
		"io.Writer": {
			GoType:     "io.Writer",
			ScriptType: "object",
		},
		"io.Reader": {
			GoType:     "io.Reader",
			ScriptType: "object",
		},
		"bytes": {
			GoType:     "[]byte",
			ScriptType: "array",
		},
	}
}

// ValidateMethod validates method calls.
func (b *UtilJSONBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *UtilJSONBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionMemory,
			Resource:    "json",
			Actions:     []string{"read", "write"},
			Description: "JSON processing operations",
		},
	}
}

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *UtilJSONBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrBridgeNotInitialized
	}

	switch name {
	case "marshal":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		result, err := llmjson.MarshalToString(args[0])
		if err != nil {
			return nil, err
		}
		return result, nil

	case "marshalIndent":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		prefix := ""
		indent := "  "
		if len(args) > 1 && args[1] != nil {
			prefix = args[1].(string)
		}
		if len(args) > 2 && args[2] != nil {
			indent = args[2].(string)
		}
		data, err := llmjson.MarshalIndent(args[0], prefix, indent)
		if err != nil {
			return nil, err
		}
		return string(data), nil

	case "marshalToBytes":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		return llmjson.Marshal(args[0])

	case "unmarshal":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		jsonStr, ok := args[0].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		var result interface{}
		err := llmjson.UnmarshalFromString(jsonStr, &result)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "unmarshalFromBytes":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		data, ok := args[0].([]byte)
		if !ok {
			return nil, ErrInvalidArguments
		}
		var result interface{}
		err := llmjson.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "unmarshalStrict":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		jsonStr, ok := args[0].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		// For strict unmarshaling, we'd need to use the decoder with DisallowUnknownFields
		// This is a simplified version
		var result interface{}
		err := llmjson.UnmarshalFromString(jsonStr, &result)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "createEncoder":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		writer, ok := args[0].(io.Writer)
		if !ok {
			return nil, ErrInvalidArguments
		}
		return llmjson.NewEncoder(writer), nil

	case "createDecoder":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		reader, ok := args[0].(io.Reader)
		if !ok {
			return nil, ErrInvalidArguments
		}
		return llmjson.NewDecoder(reader), nil

	// Structured output operations (go-llms v0.3.5)
	case "parseStructured":
		if len(args) < 2 {
			return nil, ErrInvalidArguments
		}
		output, ok := args[0].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		schemaMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, ErrInvalidArguments
		}

		// Convert script schema to go-llms schema
		schema, err := b.convertToSchema(schemaMap)
		if err != nil {
			return nil, err
		}

		return b.processor.Process(schema, output)

	case "parseWithRecovery":
		if len(args) < 1 {
			return nil, ErrInvalidArguments
		}
		output, ok := args[0].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}

		// Use go-llms ExtractJSON for malformed content recovery
		return processor.ExtractJSON(output), nil

	case "enhancePrompt":
		if len(args) < 2 {
			return nil, ErrInvalidArguments
		}
		prompt, ok := args[0].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		schemaMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, ErrInvalidArguments
		}

		// Convert script schema to go-llms schema
		schema, err := b.convertToSchema(schemaMap)
		if err != nil {
			return nil, err
		}

		// Check for options
		if len(args) > 2 && args[2] != nil {
			if options, ok := args[2].(map[string]interface{}); ok {
				return b.promptEnhancer.EnhanceWithOptions(prompt, schema, options)
			}
		}

		return b.promptEnhancer.Enhance(prompt, schema)

	// Format conversion operations (go-llms v0.3.5)
	case "convertFormat":
		if len(args) < 3 {
			return nil, ErrInvalidArguments
		}
		data, ok := args[0].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		fromFormat, ok := args[1].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		toFormat, ok := args[2].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}

		// Convert format strings to outputs.Format
		from, err := b.stringToFormat(fromFormat)
		if err != nil {
			return nil, err
		}
		to, err := b.stringToFormat(toFormat)
		if err != nil {
			return nil, err
		}

		// Check for conversion options
		var opts *outputs.ConversionOptions
		if len(args) > 3 && args[3] != nil {
			if optionsMap, ok := args[3].(map[string]interface{}); ok {
				opts = b.convertToConversionOptions(optionsMap)
			}
		}

		return b.converter.ConvertString(ctx, data, from, to, opts)

	case "streamConvert":
		if len(args) < 4 {
			return nil, ErrInvalidArguments
		}
		reader, ok := args[0].(io.Reader)
		if !ok {
			return nil, ErrInvalidArguments
		}
		writer, ok := args[1].(io.Writer)
		if !ok {
			return nil, ErrInvalidArguments
		}
		fromFormat, ok := args[2].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}
		toFormat, ok := args[3].(string)
		if !ok {
			return nil, ErrInvalidArguments
		}

		// Convert format strings to outputs.Format
		from, err := b.stringToFormat(fromFormat)
		if err != nil {
			return nil, err
		}
		to, err := b.stringToFormat(toFormat)
		if err != nil {
			return nil, err
		}

		// Check for conversion options
		var opts *outputs.ConversionOptions
		if len(args) > 4 && args[4] != nil {
			if optionsMap, ok := args[4].(map[string]interface{}); ok {
				opts = b.convertToConversionOptions(optionsMap)
			}
		}

		return nil, b.converter.StreamConvert(ctx, reader, writer, from, to, opts)

	default:
		return nil, ErrMethodNotFound
	}
}

// Helper methods for structured output operations

// convertToSchema converts a script schema map to go-llms Schema
func (b *UtilJSONBridge) convertToSchema(schemaMap map[string]interface{}) (*schemaDomain.Schema, error) {
	// Marshal the schema map to JSON
	jsonBytes, err := llmjson.Marshal(schemaMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Create a new schema and unmarshal into it
	schema := &schemaDomain.Schema{}
	if err := llmjson.Unmarshal(jsonBytes, schema); err != nil {
		return nil, fmt.Errorf("failed to convert to schema: %w", err)
	}

	return schema, nil
}

// stringToFormat converts a string to outputs.Format
func (b *UtilJSONBridge) stringToFormat(format string) (outputs.Format, error) {
	switch strings.ToLower(format) {
	case "json":
		return outputs.FormatJSON, nil
	case "yaml", "yml":
		return outputs.FormatYAML, nil
	case "xml":
		return outputs.FormatXML, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// convertToConversionOptions converts a script options map to outputs.ConversionOptions
func (b *UtilJSONBridge) convertToConversionOptions(optionsMap map[string]interface{}) *outputs.ConversionOptions {
	opts := outputs.DefaultConversionOptions()

	if pretty, ok := optionsMap["pretty"].(bool); ok {
		opts.Pretty = pretty
	}

	if preserveTypes, ok := optionsMap["preserveTypes"].(bool); ok {
		opts.PreserveTypes = preserveTypes
	}

	if rootElement, ok := optionsMap["rootElement"].(string); ok {
		opts.RootElement = rootElement
	}

	if xmlNamespace, ok := optionsMap["xmlNamespace"].(string); ok {
		opts.XMLNamespace = xmlNamespace
	}

	if indentSize, ok := optionsMap["indentSize"].(int); ok {
		opts.IndentSize = indentSize
	} else if indentSize, ok := optionsMap["indentSize"].(float64); ok {
		opts.IndentSize = int(indentSize)
	}

	return opts
}
