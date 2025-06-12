// ABOUTME: Essential utilities bridge provides common helper functions for scripts.
// ABOUTME: Includes JSON utilities, environment access, auth utilities, and error handling.

package bridge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// UtilError represents a structured error with additional details.
type UtilError struct {
	Message string
	Code    string
	Details map[string]interface{}
	Cause   error
}

func (e *UtilError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *UtilError) Unwrap() error {
	return e.Cause
}

// UtilBridge provides script access to essential utilities.
type UtilBridge struct {
	mu          sync.RWMutex
	initialized bool
}

// NewUtilBridge creates a new utilities bridge.
func NewUtilBridge() *UtilBridge {
	return &UtilBridge{}
}

// GetID returns the bridge identifier.
func (b *UtilBridge) GetID() string {
	return "util"
}

// GetMetadata returns bridge metadata.
func (b *UtilBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util",
		Version:     "1.0.0",
		Description: "Essential utilities bridge providing JSON, environment, auth, and error handling functions",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *UtilBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *UtilBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *UtilBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *UtilBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *UtilBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// JSON utilities
		{
			Name:        "parseJSON",
			Description: "Parse a JSON string into a data structure",
			Parameters: []engine.ParameterInfo{
				{Name: "jsonString", Type: "string", Required: true, Description: "JSON string to parse"},
			},
			ReturnType: "any",
		},
		{
			Name:        "stringifyJSON",
			Description: "Convert a data structure to JSON string",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "any", Required: true, Description: "Data to stringify"},
				{Name: "pretty", Type: "boolean", Required: false, Default: false, Description: "Pretty print JSON"},
			},
			ReturnType: "string",
		},
		{
			Name:        "validateJSONSchema",
			Description: "Validate data against a JSON schema",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "any", Required: true, Description: "Data to validate"},
				{Name: "schema", Type: "object", Required: true, Description: "JSON schema"},
			},
			ReturnType: "void",
		},
		// Environment utilities
		{
			Name:        "getEnv",
			Description: "Get environment variable value",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Environment variable name"},
				{Name: "defaultValue", Type: "string", Required: false, Default: "", Description: "Default value if not found"},
			},
			ReturnType: "string",
		},
		{
			Name:        "setEnv",
			Description: "Set environment variable value",
			Parameters: []engine.ParameterInfo{
				{Name: "name", Type: "string", Required: true, Description: "Environment variable name"},
				{Name: "value", Type: "string", Required: true, Description: "Value to set"},
			},
			ReturnType: "void",
		},
		{
			Name:        "listEnv",
			Description: "List environment variables matching prefix",
			Parameters: []engine.ParameterInfo{
				{Name: "prefix", Type: "string", Required: false, Default: "", Description: "Prefix to filter by"},
			},
			ReturnType: "object",
		},
		{
			Name:        "expandEnv",
			Description: "Expand environment variables in a template string",
			Parameters: []engine.ParameterInfo{
				{Name: "template", Type: "string", Required: true, Description: "Template string with ${VAR} placeholders"},
			},
			ReturnType: "string",
		},
		// Auth utilities
		{
			Name:        "createBasicAuth",
			Description: "Create Basic authentication header value",
			Parameters: []engine.ParameterInfo{
				{Name: "username", Type: "string", Required: true, Description: "Username"},
				{Name: "password", Type: "string", Required: true, Description: "Password"},
			},
			ReturnType: "string",
		},
		{
			Name:        "parseBasicAuth",
			Description: "Parse Basic authentication header value",
			Parameters: []engine.ParameterInfo{
				{Name: "header", Type: "string", Required: true, Description: "Authorization header value"},
			},
			ReturnType: "object",
		},
		{
			Name:        "createBearerAuth",
			Description: "Create Bearer token authentication header value",
			Parameters: []engine.ParameterInfo{
				{Name: "token", Type: "string", Required: true, Description: "Bearer token"},
			},
			ReturnType: "string",
		},
		{
			Name:        "parseBearerAuth",
			Description: "Parse Bearer token authentication header value",
			Parameters: []engine.ParameterInfo{
				{Name: "header", Type: "string", Required: true, Description: "Authorization header value"},
			},
			ReturnType: "string",
		},
		{
			Name:        "base64Encode",
			Description: "Encode string to base64",
			Parameters: []engine.ParameterInfo{
				{Name: "data", Type: "string", Required: true, Description: "Data to encode"},
			},
			ReturnType: "string",
		},
		{
			Name:        "base64Decode",
			Description: "Decode base64 string",
			Parameters: []engine.ParameterInfo{
				{Name: "encoded", Type: "string", Required: true, Description: "Base64 encoded string"},
			},
			ReturnType: "string",
		},
		// Error utilities
		{
			Name:        "createError",
			Description: "Create a structured error",
			Parameters: []engine.ParameterInfo{
				{Name: "message", Type: "string", Required: true, Description: "Error message"},
				{Name: "details", Type: "object", Required: false, Description: "Additional error details"},
			},
			ReturnType: "error",
		},
		{
			Name:        "wrapError",
			Description: "Wrap an existing error with additional context",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Required: true, Description: "Original error"},
				{Name: "message", Type: "string", Required: true, Description: "Wrapper message"},
				{Name: "details", Type: "object", Required: false, Description: "Additional details"},
			},
			ReturnType: "error",
		},
		{
			Name:        "isErrorType",
			Description: "Check if error is of specific type",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Required: true, Description: "Error to check"},
				{Name: "type", Type: "string", Required: true, Description: "Expected error type"},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "getErrorChain",
			Description: "Get error chain as array of messages",
			Parameters: []engine.ParameterInfo{
				{Name: "error", Type: "error", Required: true, Description: "Error to unwrap"},
			},
			ReturnType: "array",
		},
	}
}

// ValidateMethod validates method parameters.
func (b *UtilBridge) ValidateMethod(name string, args []interface{}) error {
	// Basic validation - detailed validation would be method-specific
	switch name {
	case "parseJSON", "base64Decode", "expandEnv":
		if len(args) < 1 {
			return fmt.Errorf("missing required argument")
		}
	case "stringifyJSON", "setEnv", "createBasicAuth", "parseBasicAuth",
		"createBearerAuth", "parseBearerAuth", "base64Encode", "createError",
		"wrapError", "isErrorType", "getErrorChain":
		if len(args) < 1 {
			return fmt.Errorf("missing required arguments")
		}
	}
	return nil
}

// TypeMappings returns type conversion mappings.
func (b *UtilBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"JSONValue": {
			GoType:     "interface{}",
			ScriptType: "any",
			Converter:  "standard",
		},
		"ErrorDetails": {
			GoType:     "map[string]interface{}",
			ScriptType: "object",
			Converter:  "standard",
		},
		"EnvMap": {
			GoType:     "map[string]string",
			ScriptType: "object",
			Converter:  "standard",
		},
	}
}

// RequiredPermissions returns required permissions.
func (b *UtilBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionFileSystem,
			Resource:    "environment",
			Actions:     []string{"read", "write"},
			Description: "Access to environment variables",
		},
	}
}

// JSON Utilities

// ParseJSON parses a JSON string into a Go value.
func (b *UtilBridge) ParseJSON(jsonStr string) (interface{}, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("empty JSON string")
	}

	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return result, nil
}

// StringifyJSON converts a Go value to JSON string.
func (b *UtilBridge) StringifyJSON(data interface{}, pretty bool) (string, error) {
	var bytes []byte
	var err error

	if pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("failed to stringify JSON: %w", err)
	}

	return string(bytes), nil
}

// ValidateJSONSchema validates data against a simple schema.
// Note: This is a simplified implementation. For full JSON Schema support,
// a proper validation library would be needed.
func (b *UtilBridge) ValidateJSONSchema(data interface{}, schema map[string]interface{}) error {
	// Check type
	if schemaType, ok := schema["type"].(string); ok {
		if !b.validateType(data, schemaType) {
			return fmt.Errorf("type mismatch: expected %s", schemaType)
		}
	}

	// Check required fields for objects
	if dataMap, ok := data.(map[string]interface{}); ok {
		if required, ok := schema["required"].([]string); ok {
			for _, field := range required {
				if _, exists := dataMap[field]; !exists {
					return fmt.Errorf("missing required field: %s", field)
				}
			}
		} else if required, ok := schema["required"].([]interface{}); ok {
			for _, field := range required {
				if fieldStr, ok := field.(string); ok {
					if _, exists := dataMap[fieldStr]; !exists {
						return fmt.Errorf("missing required field: %s", fieldStr)
					}
				}
			}
		}
	}

	return nil
}

func (b *UtilBridge) validateType(data interface{}, expectedType string) bool {
	switch expectedType {
	case "object":
		_, ok := data.(map[string]interface{})
		return ok
	case "array":
		_, ok := data.([]interface{})
		return ok
	case "string":
		_, ok := data.(string)
		return ok
	case "number":
		_, ok1 := data.(float64)
		_, ok2 := data.(int)
		return ok1 || ok2
	case "boolean":
		_, ok := data.(bool)
		return ok
	case "null":
		return data == nil
	default:
		return false
	}
}

// Environment Utilities

// GetEnv gets an environment variable value.
func (b *UtilBridge) GetEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

// SetEnv sets an environment variable.
func (b *UtilBridge) SetEnv(name, value string) error {
	return os.Setenv(name, value)
}

// ListEnv lists environment variables matching a prefix.
func (b *UtilBridge) ListEnv(prefix string) map[string]string {
	result := make(map[string]string)

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			if prefix == "" || strings.HasPrefix(parts[0], prefix) {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result
}

// ExpandEnv expands environment variables in a template string.
func (b *UtilBridge) ExpandEnv(template string) string {
	return os.ExpandEnv(template)
}

// Auth Utilities

// CreateBasicAuth creates a Basic authentication header value.
func (b *UtilBridge) CreateBasicAuth(username, password string) string {
	auth := username + ":" + password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + encoded
}

// ParseBasicAuth parses a Basic authentication header value.
func (b *UtilBridge) ParseBasicAuth(header string) (username, password string, err error) {
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", fmt.Errorf("invalid Basic auth header")
	}

	encoded := strings.TrimPrefix(header, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("invalid base64 encoding: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid Basic auth format")
	}

	return parts[0], parts[1], nil
}

// CreateBearerAuth creates a Bearer token authentication header value.
func (b *UtilBridge) CreateBearerAuth(token string) string {
	return "Bearer " + token
}

// ParseBearerAuth parses a Bearer token authentication header value.
func (b *UtilBridge) ParseBearerAuth(header string) (string, error) {
	if !strings.HasPrefix(header, "Bearer ") {
		return "", fmt.Errorf("invalid Bearer auth header")
	}

	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("empty Bearer token")
	}

	return token, nil
}

// Base64Encode encodes data to base64.
func (b *UtilBridge) Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64Decode decodes base64 data.
func (b *UtilBridge) Base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid base64 encoding: %w", err)
	}
	return string(decoded), nil
}

// Base64URLEncode encodes data to base64 URL encoding.
func (b *UtilBridge) Base64URLEncode(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// Base64URLDecode decodes base64 URL encoded data.
func (b *UtilBridge) Base64URLDecode(encoded string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("invalid base64 URL encoding: %w", err)
	}
	return string(decoded), nil
}

// Error Handling Utilities

// CreateError creates a structured error.
func (b *UtilBridge) CreateError(message string, details map[string]interface{}) error {
	utilErr := &UtilError{
		Message: message,
		Details: details,
	}

	if details != nil {
		if code, ok := details["code"].(string); ok {
			utilErr.Code = code
		}
	}

	return utilErr
}

// WrapError wraps an existing error with additional context.
func (b *UtilBridge) WrapError(err error, message string, details map[string]interface{}) error {
	utilErr := &UtilError{
		Message: message,
		Cause:   err,
		Details: details,
	}

	if details != nil {
		if code, ok := details["code"].(string); ok {
			utilErr.Code = code
		}
	}

	return utilErr
}

// IsErrorType checks if an error is of a specific type.
func (b *UtilBridge) IsErrorType(err error, errorType string) bool {
	utilErr, ok := err.(*UtilError)
	if !ok {
		return false
	}

	if utilErr.Details == nil {
		return false
	}

	if typeVal, ok := utilErr.Details["type"].(string); ok {
		return typeVal == errorType
	}

	return false
}

// GetErrorChain returns the chain of error messages.
func (b *UtilBridge) GetErrorChain(err error) []string {
	var chain []string

	for err != nil {
		chain = append(chain, err.Error())

		// Try to unwrap
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}

	return chain
}
