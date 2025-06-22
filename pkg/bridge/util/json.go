// ABOUTME: JSON utilities bridge provides access to go-llms optimized JSON functions.
// ABOUTME: Wraps high-performance JSON marshaling, streaming, and schema operations.

package util

import (
	"context"
	"encoding/json"
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

// UtilJSONBridge provides script access to go-llms structured output capabilities.
// It manages JSON operations, schema validation, structured output parsing,
// format conversion, and streaming for efficient JSON processing.
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
// It initializes an empty bridge ready for JSON processing operations.
func NewUtilJSONBridge() *UtilJSONBridge {
	return &UtilJSONBridge{}
}

// NewUtilJSONBridgeWithValidator creates a new JSON utilities bridge with custom validator.
// It allows using a pre-configured schema validator for JSON validation.
func NewUtilJSONBridgeWithValidator(validator schemaDomain.Validator) *UtilJSONBridge {
	return &UtilJSONBridge{
		validator: validator,
	}
}

// GetID returns the bridge identifier.
// It implements the engine.Bridge interface.
func (b *UtilJSONBridge) GetID() string {
	return "util_json"
}

// GetMetadata returns bridge metadata.
// It provides information about the JSON utilities bridge including
// version, description, and supported JSON processing features.
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
// It sets up the validator, processor, prompt enhancer, and converter
// for comprehensive JSON processing capabilities.
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
// It releases resources and marks the bridge as uninitialized.
func (b *UtilJSONBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
// It returns true if the bridge has been initialized and is ready for use.
func (b *UtilJSONBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
// It enables the script engine to access JSON utilities through this bridge.
func (b *UtilJSONBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
// It provides metadata about all JSON-related methods available to scripts,
// including marshaling, streaming, schema operations, and format conversion.
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
// It defines how Go JSON types are mapped to script types
// for encoders, decoders, and IO interfaces.
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
// It delegates validation to the engine based on Methods() metadata.
func (b *UtilJSONBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
// It specifies permissions for JSON processing operations.
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

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function.
// It implements the engine.Bridge interface, routing method calls
// to the appropriate JSON processing operations.
func (b *UtilJSONBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrBridgeNotInitialized
	}

	switch name {
	case "marshal":
		return b.marshal(ctx, args)
	case "marshalIndent":
		return b.marshalIndent(ctx, args)
	case "marshalToBytes":
		return b.marshalToBytes(ctx, args)
	case "unmarshal":
		return b.unmarshal(ctx, args)
	case "unmarshalFromBytes":
		return b.unmarshalFromBytes(ctx, args)
	case "unmarshalStrict":
		return b.unmarshalStrict(ctx, args)
	case "createEncoder":
		return b.createEncoder(ctx, args)
	case "createDecoder":
		return b.createDecoder(ctx, args)
	case "encodeStream":
		return b.encodeStream(ctx, args)
	case "decodeStream":
		return b.decodeStream(ctx, args)
	case "parseStructured":
		return b.parseStructured(ctx, args)
	case "parseWithRecovery":
		return b.parseWithRecovery(ctx, args)
	case "enhancePrompt":
		return b.enhancePrompt(ctx, args)
	case "convertFormat":
		return b.convertFormat(ctx, args)
	case "streamConvert":
		return b.streamConvert(ctx, args)
	case "prettyPrint":
		return b.prettyPrint(ctx, args)
	case "minify":
		return b.minify(ctx, args)
	default:
		return nil, ErrMethodNotFound
	}
}

// Method implementations

// marshal converts a value to JSON string using optimized marshaling.
// It uses go-llms json package for better performance.
func (b *UtilJSONBridge) marshal(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}

	// Convert ScriptValue to native Go value
	value := args[0].ToGo()

	result, err := llmjson.MarshalToString(value)
	if err != nil {
		return nil, err
	}
	return engine.NewStringValue(result), nil
}

// marshalIndent converts a value to indented JSON string.
// It supports custom prefix and indentation for readable output.
func (b *UtilJSONBridge) marshalIndent(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}

	value := args[0].ToGo()
	prefix := ""
	indent := "  "

	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeString {
		prefix = args[1].(engine.StringValue).Value()
	}
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeString {
		indent = args[2].(engine.StringValue).Value()
	}

	// json-iterator doesn't support custom prefixes, so use standard library when prefix is non-empty
	if prefix != "" {
		// Use standard library json for prefix support
		data, err := json.MarshalIndent(value, prefix, indent)
		if err != nil {
			return nil, err
		}
		return engine.NewStringValue(string(data)), nil
	}

	// Use go-llms json for better performance when no prefix
	data, err := llmjson.MarshalIndent(value, "", indent)
	if err != nil {
		return nil, err
	}
	return engine.NewStringValue(string(data)), nil
}

// marshalToBytes converts a value to JSON byte array.
// It returns the JSON as an array of numbers for binary operations.
func (b *UtilJSONBridge) marshalToBytes(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}

	value := args[0].ToGo()
	data, err := llmjson.Marshal(value)
	if err != nil {
		return nil, err
	}

	// Convert []byte to array of numbers
	scriptBytes := make([]engine.ScriptValue, len(data))
	for i, b := range data {
		scriptBytes[i] = engine.NewNumberValue(float64(b))
	}
	return engine.NewArrayValue(scriptBytes), nil
}

// unmarshal parses JSON string into a value.
// It uses optimized parsing for better performance.
func (b *UtilJSONBridge) unmarshal(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, ErrInvalidArguments
	}

	jsonStr := args[0].(engine.StringValue).Value()
	var result interface{}
	err := llmjson.UnmarshalFromString(jsonStr, &result)
	if err != nil {
		return nil, err
	}

	return engine.ConvertToScriptValue(result), nil
}

// unmarshalFromBytes parses JSON byte array into a value.
// It converts the byte array back to JSON before parsing.
func (b *UtilJSONBridge) unmarshalFromBytes(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeArray {
		return nil, ErrInvalidArguments
	}

	// Convert array of numbers to []byte
	elements := args[0].(engine.ArrayValue).Elements()
	data := make([]byte, len(elements))
	for i, elem := range elements {
		if elem.Type() != engine.TypeNumber {
			return nil, fmt.Errorf("byte array must contain only numbers")
		}
		data[i] = byte(elem.(engine.NumberValue).Value())
	}

	var result interface{}
	err := llmjson.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return engine.ConvertToScriptValue(result), nil
}

// unmarshalStrict parses JSON with strict validation.
// It can optionally disallow unknown fields for stricter parsing.
func (b *UtilJSONBridge) unmarshalStrict(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, ErrInvalidArguments
	}

	jsonStr := args[0].(engine.StringValue).Value()

	// Check for disallowUnknownFields option
	disallowUnknown := false
	if len(args) > 1 && args[1] != nil && args[1].Type() == engine.TypeBool {
		disallowUnknown = args[1].(engine.BoolValue).Value()
	}

	// For strict unmarshaling with DisallowUnknownFields, we need to use standard decoder
	if disallowUnknown {
		decoder := json.NewDecoder(strings.NewReader(jsonStr))
		decoder.DisallowUnknownFields()

		var result interface{}
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		return engine.ConvertToScriptValue(result), nil
	}

	// Otherwise use go-llms unmarshal
	var result interface{}
	err := llmjson.UnmarshalFromString(jsonStr, &result)
	if err != nil {
		return nil, err
	}

	return engine.ConvertToScriptValue(result), nil
}

// createEncoder creates a JSON encoder for streaming output.
// It wraps an io.Writer for incremental JSON encoding.
func (b *UtilJSONBridge) createEncoder(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, ErrInvalidArguments
	}

	customVal := args[0].(engine.CustomValue)
	writer, ok := customVal.Value().(io.Writer)
	if !ok {
		return nil, fmt.Errorf("argument must be io.Writer")
	}

	encoder := llmjson.NewEncoder(writer)
	return engine.NewCustomValue("JSONEncoder", encoder), nil
}

// createDecoder creates a JSON decoder for streaming input.
// It wraps an io.Reader for incremental JSON decoding.
func (b *UtilJSONBridge) createDecoder(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, ErrInvalidArguments
	}

	customVal := args[0].(engine.CustomValue)
	reader, ok := customVal.Value().(io.Reader)
	if !ok {
		return nil, fmt.Errorf("argument must be io.Reader")
	}

	decoder := llmjson.NewDecoder(reader)
	return engine.NewCustomValue("JSONDecoder", decoder), nil
}

// encodeStream encodes a value to a JSON stream.
// It writes JSON data incrementally to the encoder's writer.
func (b *UtilJSONBridge) encodeStream(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, ErrInvalidArguments
	}

	// Get encoder
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("encoder must be JSONEncoder")
	}
	customVal := args[0].(engine.CustomValue)
	// json-iterator returns a concrete type, check the type name instead
	if customVal.TypeName() != "JSONEncoder" {
		return nil, fmt.Errorf("encoder must be JSONEncoder")
	}
	encoder := customVal.Value()

	// Get value to encode
	value := args[1].ToGo()

	// Use type assertion to call Encode - json-iterator returns a concrete encoder type
	switch enc := encoder.(type) {
	case interface{ Encode(interface{}) error }:
		if err := enc.Encode(value); err != nil {
			return nil, err
		}
	default:
		// Fall back to standard library json.Encoder
		if jsonEnc, ok := encoder.(*json.Encoder); ok {
			if err := jsonEnc.Encode(value); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("encoder does not implement Encode method")
		}
	}

	return engine.NewNilValue(), nil
}

// decodeStream decodes a value from a JSON stream.
// It reads JSON data incrementally from the decoder's reader.
func (b *UtilJSONBridge) decodeStream(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}

	// Get decoder
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("decoder must be JSONDecoder")
	}
	customVal := args[0].(engine.CustomValue)
	// json-iterator returns a concrete type, check the type name instead
	if customVal.TypeName() != "JSONDecoder" {
		return nil, fmt.Errorf("decoder must be JSONDecoder")
	}
	decoder := customVal.Value()

	// Use type assertion to call Decode - json-iterator returns a concrete decoder type
	var result interface{}
	switch dec := decoder.(type) {
	case interface{ Decode(interface{}) error }:
		if err := dec.Decode(&result); err != nil {
			return nil, err
		}
	default:
		// Fall back to standard library json.Decoder
		if jsonDec, ok := decoder.(*json.Decoder); ok {
			if err := jsonDec.Decode(&result); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("decoder does not implement Decode method")
		}
	}

	return engine.ConvertToScriptValue(result), nil
}

// parseStructured parses and validates structured output from LLM responses.
// It extracts JSON from LLM output and validates against a schema.
func (b *UtilJSONBridge) parseStructured(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("output must be string")
	}
	output := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("schema must be object")
	}
	schemaObj := args[1].(engine.ObjectValue).Fields()
	schemaMap := make(map[string]interface{})
	for k, v := range schemaObj {
		schemaMap[k] = v.ToGo()
	}

	// Convert script schema to go-llms schema
	schema, err := b.convertToSchema(schemaMap)
	if err != nil {
		return nil, err
	}

	result, err := b.processor.Process(schema, output)
	if err != nil {
		return nil, err
	}

	return engine.ConvertToScriptValue(result), nil
}

// parseWithRecovery extracts JSON from malformed or mixed content.
// It attempts to recover valid JSON from corrupted or partial data.
func (b *UtilJSONBridge) parseWithRecovery(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("output must be string")
	}
	output := args[0].(engine.StringValue).Value()

	// Use go-llms ExtractJSON for malformed content recovery
	extracted := processor.ExtractJSON(output)
	return engine.NewStringValue(extracted), nil
}

// enhancePrompt adds schema information to prompts for better LLM output.
// It modifies prompts to include JSON schema requirements for structured responses.
func (b *UtilJSONBridge) enhancePrompt(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 2 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("prompt must be string")
	}
	prompt := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeObject {
		return nil, fmt.Errorf("schema must be object")
	}
	schemaObj := args[1].(engine.ObjectValue).Fields()
	schemaMap := make(map[string]interface{})
	for k, v := range schemaObj {
		schemaMap[k] = v.ToGo()
	}

	// Convert script schema to go-llms schema
	schema, err := b.convertToSchema(schemaMap)
	if err != nil {
		return nil, err
	}

	// Check for options
	var enhanced string
	if len(args) > 2 && args[2] != nil && args[2].Type() == engine.TypeObject {
		optionsObj := args[2].(engine.ObjectValue).Fields()
		options := make(map[string]interface{})
		for k, v := range optionsObj {
			options[k] = v.ToGo()
		}
		enhanced, err = b.promptEnhancer.EnhanceWithOptions(prompt, schema, options)
	} else {
		enhanced, err = b.promptEnhancer.Enhance(prompt, schema)
	}

	if err != nil {
		return nil, err
	}

	return engine.NewStringValue(enhanced), nil
}

// convertFormat converts between JSON, YAML, and XML formats.
// It preserves data structure while changing the serialization format.
func (b *UtilJSONBridge) convertFormat(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 3 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("data must be string")
	}
	data := args[0].(engine.StringValue).Value()

	if args[1] == nil || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("fromFormat must be string")
	}
	fromFormat := args[1].(engine.StringValue).Value()

	if args[2] == nil || args[2].Type() != engine.TypeString {
		return nil, fmt.Errorf("toFormat must be string")
	}
	toFormat := args[2].(engine.StringValue).Value()

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
	if len(args) > 3 && args[3] != nil && args[3].Type() == engine.TypeObject {
		optionsObj := args[3].(engine.ObjectValue).Fields()
		optionsMap := make(map[string]interface{})
		for k, v := range optionsObj {
			optionsMap[k] = v.ToGo()
		}
		opts = b.convertToConversionOptions(optionsMap)
	}

	result, err := b.converter.ConvertString(ctx, data, from, to, opts)
	if err != nil {
		return nil, err
	}

	return engine.NewStringValue(result), nil
}

// streamConvert converts formats using streaming for large data.
// It processes data incrementally to handle files too large for memory.
func (b *UtilJSONBridge) streamConvert(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 4 {
		return nil, ErrInvalidArguments
	}

	// Get reader
	if args[0] == nil || args[0].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("reader must be io.Reader")
	}
	readerVal := args[0].(engine.CustomValue)
	reader, ok := readerVal.Value().(io.Reader)
	if !ok {
		return nil, fmt.Errorf("reader must be io.Reader")
	}

	// Get writer
	if args[1] == nil || args[1].Type() != engine.TypeCustom {
		return nil, fmt.Errorf("writer must be io.Writer")
	}
	writerVal := args[1].(engine.CustomValue)
	writer, ok := writerVal.Value().(io.Writer)
	if !ok {
		return nil, fmt.Errorf("writer must be io.Writer")
	}

	if args[2] == nil || args[2].Type() != engine.TypeString {
		return nil, fmt.Errorf("fromFormat must be string")
	}
	fromFormat := args[2].(engine.StringValue).Value()

	if args[3] == nil || args[3].Type() != engine.TypeString {
		return nil, fmt.Errorf("toFormat must be string")
	}
	toFormat := args[3].(engine.StringValue).Value()

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
	if len(args) > 4 && args[4] != nil && args[4].Type() == engine.TypeObject {
		optionsObj := args[4].(engine.ObjectValue).Fields()
		optionsMap := make(map[string]interface{})
		for k, v := range optionsObj {
			optionsMap[k] = v.ToGo()
		}
		opts = b.convertToConversionOptions(optionsMap)
	}

	err = b.converter.StreamConvert(ctx, reader, writer, from, to, opts)
	if err != nil {
		return nil, err
	}

	return engine.NewNilValue(), nil
}

// prettyPrint formats JSON with indentation for readability.
// It parses and reformats JSON with consistent spacing.
func (b *UtilJSONBridge) prettyPrint(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("json must be string")
	}
	jsonStr := args[0].(engine.StringValue).Value()

	// Parse and reformat with indentation
	var obj interface{}
	if err := llmjson.UnmarshalFromString(jsonStr, &obj); err != nil {
		return nil, err
	}

	data, err := llmjson.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, err
	}

	// Note: Color support would require additional terminal color library
	// For now, just return formatted JSON
	return engine.NewStringValue(string(data)), nil
}

// minify removes unnecessary whitespace from JSON.
// It creates compact JSON by removing formatting.
func (b *UtilJSONBridge) minify(ctx context.Context, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if len(args) < 1 {
		return nil, ErrInvalidArguments
	}
	if args[0] == nil || args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("json must be string")
	}
	jsonStr := args[0].(engine.StringValue).Value()

	// Parse and reformat without whitespace
	var obj interface{}
	if err := llmjson.UnmarshalFromString(jsonStr, &obj); err != nil {
		return nil, err
	}

	result, err := llmjson.MarshalToString(obj)
	if err != nil {
		return nil, err
	}

	return engine.NewStringValue(result), nil
}

// Helper methods

// convertToSchema converts a script schema map to go-llms Schema.
// It marshals and unmarshals to transform the schema format.
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

// stringToFormat converts a string to outputs.Format.
// It maps format names to go-llms format constants.
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

// convertToConversionOptions converts a script options map to outputs.ConversionOptions.
// It extracts conversion settings like pretty printing and XML namespaces.
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

// Helper function to convert interface{} to ScriptValue
// NOTE: Duplicate conversion functions removed - using centralized engine.ConvertToScriptValue() instead
