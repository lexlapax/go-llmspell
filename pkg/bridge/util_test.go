// ABOUTME: Test suite for the essential utilities bridge that provides common helper functions.
// ABOUTME: Tests JSON utilities, environment access, auth utilities, and error handling capabilities.

package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

// Tests for UtilBridge
func TestNewUtilBridge(t *testing.T) {
	bridge := NewUtilBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util", bridge.GetID())
}

func TestUtilBridgeMetadata(t *testing.T) {
	bridge := NewUtilBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util", metadata.Name)
	assert.NotEmpty(t, metadata.Version)
	assert.NotEmpty(t, metadata.Description)
	assert.Contains(t, metadata.Description, "utilities")
}

func TestJSONUtilities(t *testing.T) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	t.Run("Parse JSON", func(t *testing.T) {
		jsonStr := `{"name": "test", "value": 42, "active": true}`

		result, err := bridge.ParseJSON(jsonStr)
		assert.NoError(t, err)

		data, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test", data["name"])
		assert.Equal(t, float64(42), data["value"])
		assert.Equal(t, true, data["active"])
	})

	t.Run("Parse JSON Array", func(t *testing.T) {
		jsonStr := `[1, "two", true, null]`

		result, err := bridge.ParseJSON(jsonStr)
		assert.NoError(t, err)

		arr, ok := result.([]interface{})
		assert.True(t, ok)
		assert.Len(t, arr, 4)
		assert.Equal(t, float64(1), arr[0])
		assert.Equal(t, "two", arr[1])
		assert.Equal(t, true, arr[2])
		assert.Nil(t, arr[3])
	})

	t.Run("Parse Invalid JSON", func(t *testing.T) {
		jsonStr := `{invalid json}`

		_, err := bridge.ParseJSON(jsonStr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("Stringify JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"name":   "test",
			"value":  42,
			"active": true,
			"nested": map[string]interface{}{
				"key": "value",
			},
		}

		result, err := bridge.StringifyJSON(data, false)
		assert.NoError(t, err)

		// Parse it back to verify
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(result), &parsed)
		assert.NoError(t, err)
		assert.Equal(t, "test", parsed["name"])
	})

	t.Run("Stringify JSON Pretty", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "test",
			"value": 42,
		}

		result, err := bridge.StringifyJSON(data, true)
		assert.NoError(t, err)
		assert.Contains(t, result, "\n")
		assert.Contains(t, result, "  ")
	})

	t.Run("Validate JSON Schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []string{"name"},
		}

		validData := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		err := bridge.ValidateJSONSchema(validData, schema)
		assert.NoError(t, err)

		invalidData := map[string]interface{}{
			"age": 30, // missing required "name"
		}

		err = bridge.ValidateJSONSchema(invalidData, schema)
		assert.Error(t, err)
	})
}

func TestEnvironmentUtilities(t *testing.T) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	t.Run("Get Environment Variable", func(t *testing.T) {
		// Set a test env var
		_ = os.Setenv("TEST_VAR", "test_value")
		defer func() { _ = os.Unsetenv("TEST_VAR") }()

		value := bridge.GetEnv("TEST_VAR", "default")
		assert.Equal(t, "test_value", value)
	})

	t.Run("Get Non-existent Environment Variable", func(t *testing.T) {
		value := bridge.GetEnv("NON_EXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", value)
	})

	t.Run("Set Environment Variable", func(t *testing.T) {
		err := bridge.SetEnv("TEST_SET_VAR", "new_value")
		assert.NoError(t, err)

		value := os.Getenv("TEST_SET_VAR")
		assert.Equal(t, "new_value", value)

		_ = os.Unsetenv("TEST_SET_VAR")
	})

	t.Run("List Environment Variables", func(t *testing.T) {
		_ = os.Setenv("TEST_LIST_VAR1", "value1")
		_ = os.Setenv("TEST_LIST_VAR2", "value2")
		defer func() { _ = os.Unsetenv("TEST_LIST_VAR1") }()
		defer func() { _ = os.Unsetenv("TEST_LIST_VAR2") }()

		envMap := bridge.ListEnv("TEST_LIST")
		assert.Contains(t, envMap, "TEST_LIST_VAR1")
		assert.Contains(t, envMap, "TEST_LIST_VAR2")
		assert.Equal(t, "value1", envMap["TEST_LIST_VAR1"])
		assert.Equal(t, "value2", envMap["TEST_LIST_VAR2"])
	})

	t.Run("Expand Environment Variables", func(t *testing.T) {
		_ = os.Setenv("USER", "testuser")
		_ = os.Setenv("HOME", "/home/testuser")
		defer func() { _ = os.Unsetenv("USER") }()
		defer func() { _ = os.Unsetenv("HOME") }()

		template := "Hello ${USER}, your home is ${HOME}"
		expanded := bridge.ExpandEnv(template)
		assert.Equal(t, "Hello testuser, your home is /home/testuser", expanded)
	})
}

func TestAuthUtilities(t *testing.T) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	t.Run("Create Basic Auth Header", func(t *testing.T) {
		header := bridge.CreateBasicAuth("username", "password")
		assert.Equal(t, "Basic dXNlcm5hbWU6cGFzc3dvcmQ=", header)
	})

	t.Run("Parse Basic Auth Header", func(t *testing.T) {
		header := "Basic dXNlcm5hbWU6cGFzc3dvcmQ="
		username, password, err := bridge.ParseBasicAuth(header)
		assert.NoError(t, err)
		assert.Equal(t, "username", username)
		assert.Equal(t, "password", password)
	})

	t.Run("Parse Invalid Basic Auth Header", func(t *testing.T) {
		// Missing "Basic " prefix
		_, _, err := bridge.ParseBasicAuth("dXNlcm5hbWU6cGFzc3dvcmQ=")
		assert.Error(t, err)

		// Invalid base64
		_, _, err = bridge.ParseBasicAuth("Basic invalid!!!")
		assert.Error(t, err)

		// Missing colon separator
		_, _, err = bridge.ParseBasicAuth("Basic " + bridge.Base64Encode("noColon"))
		assert.Error(t, err)
	})

	t.Run("Create Bearer Token Header", func(t *testing.T) {
		header := bridge.CreateBearerAuth("my-token-123")
		assert.Equal(t, "Bearer my-token-123", header)
	})

	t.Run("Parse Bearer Token Header", func(t *testing.T) {
		header := "Bearer my-token-123"
		token, err := bridge.ParseBearerAuth(header)
		assert.NoError(t, err)
		assert.Equal(t, "my-token-123", token)
	})

	t.Run("Parse Invalid Bearer Token Header", func(t *testing.T) {
		// Missing "Bearer " prefix
		_, err := bridge.ParseBearerAuth("my-token-123")
		assert.Error(t, err)

		// Empty token
		_, err = bridge.ParseBearerAuth("Bearer ")
		assert.Error(t, err)
	})

	t.Run("Base64 Encoding", func(t *testing.T) {
		data := "Hello, World!"
		encoded := bridge.Base64Encode(data)
		assert.Equal(t, "SGVsbG8sIFdvcmxkIQ==", encoded)

		decoded, err := bridge.Base64Decode(encoded)
		assert.NoError(t, err)
		assert.Equal(t, data, decoded)
	})

	t.Run("Base64 URL Encoding", func(t *testing.T) {
		data := "Hello, World!?/"
		encoded := bridge.Base64URLEncode(data)
		assert.NotContains(t, encoded, "+")
		assert.NotContains(t, encoded, "/")

		decoded, err := bridge.Base64URLDecode(encoded)
		assert.NoError(t, err)
		assert.Equal(t, data, decoded)
	})
}

func TestErrorHandling(t *testing.T) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	t.Run("Create Error", func(t *testing.T) {
		err := bridge.CreateError("test error", map[string]interface{}{
			"code":   "ERR001",
			"detail": "Something went wrong",
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error")

		// Check error details
		utilErr, ok := err.(*UtilError)
		assert.True(t, ok)
		assert.Equal(t, "ERR001", utilErr.Code)
		assert.Equal(t, "Something went wrong", utilErr.Details["detail"])
	})

	t.Run("Wrap Error", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := bridge.WrapError(originalErr, "wrapper message", map[string]interface{}{
			"context": "testing",
		})

		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "wrapper message")
		assert.Contains(t, wrappedErr.Error(), "original error")

		// Check unwrapping
		utilErr, ok := wrappedErr.(*UtilError)
		assert.True(t, ok)
		assert.Equal(t, originalErr, utilErr.Unwrap())
		assert.Equal(t, "testing", utilErr.Details["context"])
	})

	t.Run("Error Type Checking", func(t *testing.T) {
		err1 := bridge.CreateError("type1 error", map[string]interface{}{
			"type": "validation",
		})

		err2 := bridge.CreateError("type2 error", map[string]interface{}{
			"type": "network",
		})

		assert.True(t, bridge.IsErrorType(err1, "validation"))
		assert.False(t, bridge.IsErrorType(err1, "network"))
		assert.True(t, bridge.IsErrorType(err2, "network"))
		assert.False(t, bridge.IsErrorType(err2, "validation"))

		// Non-util error
		regularErr := errors.New("regular error")
		assert.False(t, bridge.IsErrorType(regularErr, "any"))
	})

	t.Run("Error Chain", func(t *testing.T) {
		err1 := errors.New("root cause")
		err2 := bridge.WrapError(err1, "middle layer", nil)
		err3 := bridge.WrapError(err2, "top layer", nil)

		chain := bridge.GetErrorChain(err3)
		assert.Len(t, chain, 3)
		assert.Contains(t, chain[0], "top layer")
		assert.Contains(t, chain[1], "middle layer")
		assert.Contains(t, chain[2], "root cause")
	})
}

func TestUtilBridgeEngineIntegration(t *testing.T) {
	t.Run("Register With Engine", func(t *testing.T) {
		bridge := NewUtilBridge()
		engine := &mockScriptEngine{}

		err := bridge.RegisterWithEngine(engine)
		assert.NoError(t, err)
		assert.Len(t, engine.bridges, 1)
		assert.Equal(t, bridge, engine.bridges[0])
	})

	t.Run("Bridge Methods", func(t *testing.T) {
		bridge := NewUtilBridge()
		methods := bridge.Methods()

		// Check that all utility methods are exposed
		methodNames := make([]string, len(methods))
		for i, m := range methods {
			methodNames[i] = m.Name
		}

		// JSON utilities
		assert.Contains(t, methodNames, "parseJSON")
		assert.Contains(t, methodNames, "stringifyJSON")
		assert.Contains(t, methodNames, "validateJSONSchema")

		// Environment utilities
		assert.Contains(t, methodNames, "getEnv")
		assert.Contains(t, methodNames, "setEnv")
		assert.Contains(t, methodNames, "listEnv")
		assert.Contains(t, methodNames, "expandEnv")

		// Auth utilities
		assert.Contains(t, methodNames, "createBasicAuth")
		assert.Contains(t, methodNames, "parseBasicAuth")
		assert.Contains(t, methodNames, "createBearerAuth")
		assert.Contains(t, methodNames, "parseBearerAuth")
		assert.Contains(t, methodNames, "base64Encode")
		assert.Contains(t, methodNames, "base64Decode")

		// Error utilities
		assert.Contains(t, methodNames, "createError")
		assert.Contains(t, methodNames, "wrapError")
		assert.Contains(t, methodNames, "isErrorType")
		assert.Contains(t, methodNames, "getErrorChain")
	})

	t.Run("Type Mappings", func(t *testing.T) {
		bridge := NewUtilBridge()
		mappings := bridge.TypeMappings()

		assert.Contains(t, mappings, "JSONValue")
		assert.Contains(t, mappings, "ErrorDetails")
		assert.Contains(t, mappings, "EnvMap")
	})

	t.Run("Required Permissions", func(t *testing.T) {
		bridge := NewUtilBridge()
		permissions := bridge.RequiredPermissions()

		// Should require filesystem permission for env vars
		hasFS := false
		for _, p := range permissions {
			if p.Type == engine.PermissionFileSystem {
				hasFS = true
				break
			}
		}
		assert.True(t, hasFS, "Util bridge should require filesystem permission for env vars")
	})
}

func TestConcurrentUtilAccess(t *testing.T) {
	t.Run("Concurrent JSON Operations", func(t *testing.T) {
		bridge := NewUtilBridge()
		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Run concurrent JSON parse/stringify operations
		done := make(chan bool, 20)

		for i := 0; i < 10; i++ {
			go func(id int) {
				data := map[string]interface{}{
					"id":    id,
					"value": "test",
				}
				str, _ := bridge.StringifyJSON(data, false)
				parsed, _ := bridge.ParseJSON(str)
				_ = parsed
				done <- true
			}(i)

			go func(id int) {
				jsonStr := `{"id": ` + string(rune(id)) + `, "active": true}`
				parsed, _ := bridge.ParseJSON(jsonStr)
				_ = parsed
				done <- true
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < 20; i++ {
			<-done
		}
	})

	t.Run("Concurrent Environment Operations", func(t *testing.T) {
		bridge := NewUtilBridge()
		ctx := context.Background()
		_ = bridge.Initialize(ctx)

		// Run concurrent env operations
		done := make(chan bool, 15)

		for i := 0; i < 5; i++ {
			go func(id int) {
				_ = bridge.SetEnv("CONCURRENT_TEST_"+string(rune(id)), "value")
				done <- true
			}(i)

			go func(id int) {
				bridge.GetEnv("CONCURRENT_TEST_"+string(rune(id)), "default")
				done <- true
			}(i)

			go func(id int) {
				bridge.ListEnv("CONCURRENT_TEST")
				done <- true
			}(i)
		}

		// Wait for all operations
		for i := 0; i < 15; i++ {
			<-done
		}

		// Cleanup
		for i := 0; i < 5; i++ {
			_ = os.Unsetenv("CONCURRENT_TEST_" + string(rune(i)))
		}
	})
}

func TestEdgeCasesUtil(t *testing.T) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	t.Run("Empty JSON", func(t *testing.T) {
		result, err := bridge.ParseJSON("")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Null JSON", func(t *testing.T) {
		result, err := bridge.ParseJSON("null")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Large JSON", func(t *testing.T) {
		// Create a large nested structure
		data := make(map[string]interface{})
		current := data
		for i := 0; i < 100; i++ {
			next := make(map[string]interface{})
			current["nested"] = next
			current["value"] = i
			current = next
		}

		str, err := bridge.StringifyJSON(data, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, str)

		parsed, err := bridge.ParseJSON(str)
		assert.NoError(t, err)
		assert.NotNil(t, parsed)
	})

	t.Run("Special Characters in Env", func(t *testing.T) {
		// Test with special characters
		err := bridge.SetEnv("TEST_SPECIAL", "value with spaces & symbols!")
		assert.NoError(t, err)

		value := bridge.GetEnv("TEST_SPECIAL", "")
		assert.Equal(t, "value with spaces & symbols!", value)

		_ = os.Unsetenv("TEST_SPECIAL")
	})

	t.Run("Unicode in Auth", func(t *testing.T) {
		username := "用户"
		password := "密码"

		header := bridge.CreateBasicAuth(username, password)
		assert.NotEmpty(t, header)

		parsedUser, parsedPass, err := bridge.ParseBasicAuth(header)
		assert.NoError(t, err)
		assert.Equal(t, username, parsedUser)
		assert.Equal(t, password, parsedPass)
	})
}

// Benchmark tests
func BenchmarkJSONParse(b *testing.B) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	jsonStr := `{"name": "test", "value": 42, "nested": {"key": "value"}, "array": [1, 2, 3]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bridge.ParseJSON(jsonStr)
	}
}

func BenchmarkJSONStringify(b *testing.B) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	data := map[string]interface{}{
		"name":  "test",
		"value": 42,
		"nested": map[string]interface{}{
			"key": "value",
		},
		"array": []int{1, 2, 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bridge.StringifyJSON(data, false)
	}
}

func BenchmarkBase64Encode(b *testing.B) {
	bridge := NewUtilBridge()
	ctx := context.Background()
	_ = bridge.Initialize(ctx)

	data := "This is a test string for base64 encoding benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bridge.Base64Encode(data)
	}
}
