// ABOUTME: Tests for JSON utilities bridge with ScriptValue-based API
// ABOUTME: Validates JSON operations, schema handling, and format conversions

package util

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUtilJSONBridgeInitialization(t *testing.T) {
	bridge := NewUtilJSONBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_json", bridge.GetID())
	assert.False(t, bridge.IsInitialized())

	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test double initialization
	err = bridge.Initialize(ctx)
	assert.NoError(t, err)

	// Test cleanup
	err = bridge.Cleanup(ctx)
	require.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestUtilJSONBridgeMetadata(t *testing.T) {
	bridge := NewUtilJSONBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util_json", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Structured output parser")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilJSONBridgeMethods(t *testing.T) {
	bridge := NewUtilJSONBridge()
	methods := bridge.Methods()

	// Check that all expected methods are present
	expectedMethods := []string{
		// Marshaling
		"marshal",
		"marshalIndent",
		"marshalToBytes",
		// Unmarshaling
		"unmarshal",
		"unmarshalFromBytes",
		"unmarshalStrict",
		// Streaming
		"createEncoder",
		"createDecoder",
		"encodeStream",
		"decodeStream",
		// Schema
		"validateWithSchema",
		"generateFromSchema",
		"inferSchema",
		// Structured output
		"parseStructured",
		"parseWithRecovery",
		"enhancePrompt",
		// Format conversion
		"convertFormat",
		"streamConvert",
		// Utilities
		"prettyPrint",
		"minify",
		"merge",
		"diff",
		// Performance
		"marshalWithBuffer",
		"marshalConcurrent",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestUtilJSONBridgeMarshal(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		args     []engine.ScriptValue
		wantJSON string
		wantErr  bool
	}{
		{
			name: "marshal simple object",
			args: []engine.ScriptValue{
				svMap(map[string]interface{}{
					"name": "test",
					"age":  25,
				}),
			},
			wantJSON: `{"age":25,"name":"test"}`,
			wantErr:  false,
		},
		{
			name: "marshal array",
			args: []engine.ScriptValue{
				svArray(1, 2, 3),
			},
			wantJSON: `[1,2,3]`,
			wantErr:  false,
		},
		{
			name: "marshal string",
			args: []engine.ScriptValue{
				sv("hello world"),
			},
			wantJSON: `"hello world"`,
			wantErr:  false,
		},
		{
			name: "marshal null",
			args: []engine.ScriptValue{
				sv(nil),
			},
			wantJSON: `null`,
			wantErr:  false,
		},
		{
			name:    "missing value",
			args:    []engine.ScriptValue{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "marshal", tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, engine.TypeString, result.Type())
				assert.Equal(t, tt.wantJSON, result.(engine.StringValue).Value())
			}
		})
	}
}

func TestUtilJSONBridgeMarshalIndent(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	obj := svMap(map[string]interface{}{
		"name": "test",
		"age":  25,
	})

	// Test default indentation
	result, err := bridge.ExecuteMethod(ctx, "marshalIndent", []engine.ScriptValue{obj})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())
	jsonStr := result.(engine.StringValue).Value()
	assert.Contains(t, jsonStr, "\n")
	assert.Contains(t, jsonStr, "  ")

	// Test custom indentation
	result, err = bridge.ExecuteMethod(ctx, "marshalIndent", []engine.ScriptValue{
		obj,
		sv(">>"),
		sv("\t"),
	})
	require.NoError(t, err)
	jsonStr = result.(engine.StringValue).Value()
	assert.Contains(t, jsonStr, ">>")
	assert.Contains(t, jsonStr, "\t")
}

func TestUtilJSONBridgeUnmarshal(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		json    string
		wantErr bool
		check   func(t *testing.T, result engine.ScriptValue)
	}{
		{
			name:    "unmarshal object",
			json:    `{"name":"test","age":25}`,
			wantErr: false,
			check: func(t *testing.T, result engine.ScriptValue) {
				assert.Equal(t, engine.TypeObject, result.Type())
				obj := result.(engine.ObjectValue).Fields()
				assert.Equal(t, "test", obj["name"].(engine.StringValue).Value())
				assert.Equal(t, float64(25), obj["age"].(engine.NumberValue).Value())
			},
		},
		{
			name:    "unmarshal array",
			json:    `[1,2,3]`,
			wantErr: false,
			check: func(t *testing.T, result engine.ScriptValue) {
				assert.Equal(t, engine.TypeArray, result.Type())
				arr := result.(engine.ArrayValue).Elements()
				assert.Len(t, arr, 3)
				assert.Equal(t, float64(1), arr[0].(engine.NumberValue).Value())
			},
		},
		{
			name:    "unmarshal string",
			json:    `"hello"`,
			wantErr: false,
			check: func(t *testing.T, result engine.ScriptValue) {
				assert.Equal(t, engine.TypeString, result.Type())
				assert.Equal(t, "hello", result.(engine.StringValue).Value())
			},
		},
		{
			name:    "unmarshal null",
			json:    `null`,
			wantErr: false,
			check: func(t *testing.T, result engine.ScriptValue) {
				assert.True(t, result.IsNil())
			},
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "unmarshal", []engine.ScriptValue{
				sv(tt.json),
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.check(t, result)
			}
		})
	}
}

func TestUtilJSONBridgeMarshalToBytes(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	obj := svMap(map[string]interface{}{
		"test": true,
	})

	result, err := bridge.ExecuteMethod(ctx, "marshalToBytes", []engine.ScriptValue{obj})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeArray, result.Type())

	// Convert back to bytes and verify
	elements := result.(engine.ArrayValue).Elements()
	data := make([]byte, len(elements))
	for i, elem := range elements {
		data[i] = byte(elem.(engine.NumberValue).Value())
	}

	var decoded map[string]bool
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.True(t, decoded["test"])
}

func TestUtilJSONBridgeUnmarshalFromBytes(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create byte array ScriptValue
	jsonBytes := []byte(`{"success":true}`)
	bytes := make([]interface{}, len(jsonBytes))
	for i, b := range jsonBytes {
		bytes[i] = float64(b)
	}

	result, err := bridge.ExecuteMethod(ctx, "unmarshalFromBytes", []engine.ScriptValue{
		svArray(bytes...),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, result.Type())

	obj := result.(engine.ObjectValue).Fields()
	assert.True(t, obj["success"].(engine.BoolValue).Value())
}

func TestUtilJSONBridgeUnmarshalStrict(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test normal unmarshal
	result, err := bridge.ExecuteMethod(ctx, "unmarshalStrict", []engine.ScriptValue{
		sv(`{"name":"test"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, result.Type())

	// Test with disallow unknown fields
	// This would fail with unknown fields in a strictly typed struct
	_, err = bridge.ExecuteMethod(ctx, "unmarshalStrict", []engine.ScriptValue{
		sv(`{"name":"test","unknown":"field"}`),
		sv(true),
	})
	// For generic interface{} unmarshaling, this will still succeed
	// as we're not unmarshaling into a struct with defined fields
	require.NoError(t, err)
}

func TestUtilJSONBridgeStreaming(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Create encoder
	var buf bytes.Buffer
	writer := engine.NewCustomValue("io.Writer", &buf)
	encoder, err := bridge.ExecuteMethod(ctx, "createEncoder", []engine.ScriptValue{writer})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, encoder.Type())

	// Encode value
	obj := svMap(map[string]interface{}{
		"stream": true,
	})
	result, err := bridge.ExecuteMethod(ctx, "encodeStream", []engine.ScriptValue{encoder, obj})
	require.NoError(t, err)
	assert.True(t, result.IsNil())

	// Verify encoded data
	assert.Contains(t, buf.String(), "stream")
	assert.Contains(t, buf.String(), "true")

	// Create decoder
	reader := engine.NewCustomValue("io.Reader", strings.NewReader(buf.String()))
	decoder, err := bridge.ExecuteMethod(ctx, "createDecoder", []engine.ScriptValue{reader})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeCustom, decoder.Type())

	// Decode value
	decoded, err := bridge.ExecuteMethod(ctx, "decodeStream", []engine.ScriptValue{decoder})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, decoded.Type())

	decodedObj := decoded.(engine.ObjectValue).Fields()
	assert.True(t, decodedObj["stream"].(engine.BoolValue).Value())
}

func TestUtilJSONBridgeParseStructured(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	schema := svMap(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"age": map[string]interface{}{
				"type": "number",
			},
		},
	})

	output := "Here's the JSON: {\"name\":\"John\",\"age\":30}"

	result, err := bridge.ExecuteMethod(ctx, "parseStructured", []engine.ScriptValue{
		sv(output),
		schema,
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeObject, result.Type())
}

func TestUtilJSONBridgeParseWithRecovery(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test with malformed content
	malformed := `Some text before {"valid":"json"} and some after`
	result, err := bridge.ExecuteMethod(ctx, "parseWithRecovery", []engine.ScriptValue{
		sv(malformed),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())
	assert.Contains(t, result.(engine.StringValue).Value(), "valid")
}

func TestUtilJSONBridgeEnhancePrompt(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	prompt := "Extract user information"
	schema := svMap(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	})

	// Test basic enhancement
	result, err := bridge.ExecuteMethod(ctx, "enhancePrompt", []engine.ScriptValue{
		sv(prompt),
		schema,
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())
	enhanced := result.(engine.StringValue).Value()
	assert.Contains(t, enhanced, prompt)

	// Test with options
	options := svMap(map[string]interface{}{
		"examples": true,
	})
	result, err = bridge.ExecuteMethod(ctx, "enhancePrompt", []engine.ScriptValue{
		sv(prompt),
		schema,
		options,
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())
}

func TestUtilJSONBridgeConvertFormat(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	jsonData := `{"name":"test","value":123}`

	// Test JSON to YAML conversion
	result, err := bridge.ExecuteMethod(ctx, "convertFormat", []engine.ScriptValue{
		sv(jsonData),
		sv("json"),
		sv("yaml"),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())
	yamlStr := result.(engine.StringValue).Value()
	assert.Contains(t, yamlStr, "name:")
	assert.Contains(t, yamlStr, "test")

	// Test with options
	options := svMap(map[string]interface{}{
		"pretty":     true,
		"indentSize": 4,
	})
	result, err = bridge.ExecuteMethod(ctx, "convertFormat", []engine.ScriptValue{
		sv(jsonData),
		sv("json"),
		sv("yaml"),
		options,
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())
}

func TestUtilJSONBridgePrettyPrint(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	compactJSON := `{"a":1,"b":2,"c":{"d":3}}`

	result, err := bridge.ExecuteMethod(ctx, "prettyPrint", []engine.ScriptValue{
		sv(compactJSON),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())

	prettyJSON := result.(engine.StringValue).Value()
	assert.Contains(t, prettyJSON, "\n")
	assert.Contains(t, prettyJSON, "  ")
	assert.Contains(t, prettyJSON, `"a": 1`)
}

func TestUtilJSONBridgeMinify(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	prettyJSON := `{
		"a": 1,
		"b": 2,
		"c": {
			"d": 3
		}
	}`

	result, err := bridge.ExecuteMethod(ctx, "minify", []engine.ScriptValue{
		sv(prettyJSON),
	})
	require.NoError(t, err)
	assert.Equal(t, engine.TypeString, result.Type())

	minified := result.(engine.StringValue).Value()
	assert.NotContains(t, minified, "\n")
	assert.NotContains(t, minified, "  ")
	assert.Equal(t, `{"a":1,"b":2,"c":{"d":3}}`, minified)
}

func TestUtilJSONBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilJSONBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("marshal", []engine.ScriptValue{
		svMap(map[string]interface{}{}),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

func TestUtilJSONBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilJSONBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 1)

	// Check for expected permissions
	hasMemory := false

	for _, perm := range permissions {
		if perm.Type == engine.PermissionMemory && perm.Resource == "json" {
			hasMemory = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "write")
		}
	}

	assert.True(t, hasMemory, "Memory permission not found")
}

func TestUtilJSONBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilJSONBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	expectedTypes := []string{
		"JSONEncoder",
		"JSONDecoder",
		"io.Writer",
		"io.Reader",
		"bytes",
	}

	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.NotEmpty(t, mapping.ScriptType)
	}
}

func TestUtilJSONBridgeErrorHandling(t *testing.T) {
	bridge := NewUtilJSONBridge()
	ctx := context.Background()

	// Test method execution before initialization
	_, err := bridge.ExecuteMethod(ctx, "marshal", []engine.ScriptValue{
		svMap(map[string]interface{}{}),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test unknown method
	_, err = bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method not found")

	// Test invalid arguments
	_, err = bridge.ExecuteMethod(ctx, "marshal", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}
