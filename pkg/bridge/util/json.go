// ABOUTME: JSON utilities bridge provides access to go-llms optimized JSON functions.
// ABOUTME: Wraps high-performance JSON marshaling, streaming, and schema operations.

package util

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// UtilJSONBridge provides script access to go-llms JSON utilities.
type UtilJSONBridge struct {
	mu          sync.RWMutex
	initialized bool
}

// NewUtilJSONBridge creates a new JSON utilities bridge.
func NewUtilJSONBridge() *UtilJSONBridge {
	return &UtilJSONBridge{}
}

// GetID returns the bridge identifier.
func (b *UtilJSONBridge) GetID() string {
	return "util_json"
}

// GetMetadata returns bridge metadata.
func (b *UtilJSONBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_json",
		Version:     "1.0.0",
		Description: "Optimized JSON utilities for high-performance marshaling and streaming",
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

// The actual method implementations would be provided by the script engine
// which would call the appropriate go-llms/pkg/util/json functions.
// For example:
// - marshal would call json.Marshal with optimizations
// - unmarshal would call json.Unmarshal
// - createEncoder would return json.NewEncoder
// etc.
