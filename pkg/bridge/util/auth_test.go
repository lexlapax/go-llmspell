// ABOUTME: Tests for auth utilities bridge with ScriptValue-based API
// ABOUTME: Validates auth configuration, OAuth2 operations, and credential management

package util

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUtilAuthBridgeInitialization(t *testing.T) {
	bridge := NewUtilAuthBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_auth", bridge.GetID())
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

func TestUtilAuthBridgeMetadata(t *testing.T) {
	bridge := NewUtilAuthBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util_auth", metadata.Name)
	assert.Equal(t, "2.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Enhanced authentication")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilAuthBridgeMethods(t *testing.T) {
	bridge := NewUtilAuthBridge()
	methods := bridge.Methods()

	// Check that all expected auth methods are present
	expectedMethods := []string{
		// Auth configuration
		"createAuthConfig",
		"createAuthFromEnv",
		"createAuthFromState",
		// HTTP request authentication
		"applyAuth",
		"applyAuthToHeaders",
		// Auth scheme utilities
		"detectAuthScheme",
		"parseAuthHeader",
		"validateAuthConfig",
		// OAuth2 utilities
		"createOAuth2Config",
		"refreshOAuth2Token",
		"discoverOAuth2Endpoints",
		"validateOAuth2Token",
		"parseJWTClaims",
		"autoRefreshToken",
		// Multi-scheme authentication
		"registerAuthScheme",
		"getAuthSchemes",
		"selectBestAuthScheme",
		// Credential serialization
		"serializeCredentials",
		"deserializeCredentials",
		"cacheCredentials",
		// Auth event logging
		"logAuthEvent",
		"getAuthEventHistory",
		"subscribeToAuthEvents",
		// Session management
		"createAuthSession",
		"validateSession",
		// Credential management
		"maskCredentials",
		"rotateAPIKey",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m.Name] = true
	}

	for _, expected := range expectedMethods {
		assert.True(t, methodMap[expected], "Method %s not found", expected)
	}
}

func TestUtilAuthBridgeCreateAuthConfig(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        []engine.ScriptValue
		wantErr     bool
		checkResult func(t *testing.T, result engine.ScriptValue)
	}{
		{
			name: "create bearer auth config",
			args: []engine.ScriptValue{
				sv("bearer"),
				svMap(map[string]interface{}{
					"token": "test-token-123",
				}),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue) {
				require.NotNil(t, result)
				assert.Equal(t, engine.TypeObject, result.Type())
				obj := result.(engine.ObjectValue).Fields()
				assert.Equal(t, "bearer", obj["type"].(engine.StringValue).Value())
				assert.NotNil(t, obj["data"])
			},
		},
		{
			name: "create api key auth config",
			args: []engine.ScriptValue{
				sv("apiKey"),
				svMap(map[string]interface{}{
					"key":    "api-key-456",
					"header": "X-API-Key",
				}),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result engine.ScriptValue) {
				require.NotNil(t, result)
				assert.Equal(t, engine.TypeObject, result.Type())
				obj := result.(engine.ObjectValue).Fields()
				assert.Equal(t, "apiKey", obj["type"].(engine.StringValue).Value())
			},
		},
		{
			name: "missing auth type",
			args: []engine.ScriptValue{
				sv(nil),
				svMap(map[string]interface{}{}),
			},
			wantErr: true,
		},
		{
			name: "missing credentials",
			args: []engine.ScriptValue{
				sv("bearer"),
			},
			wantErr: true,
		},
		{
			name: "invalid credentials type",
			args: []engine.ScriptValue{
				sv("bearer"),
				sv("not-an-object"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bridge.ExecuteMethod(ctx, "createAuthConfig", tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

func TestUtilAuthBridgeOAuth2Operations(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	t.Run("discoverOAuth2Endpoints", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "discoverOAuth2Endpoints", []engine.ScriptValue{
			sv("https://auth.example.com"),
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, engine.TypeObject, result.Type())

		obj := result.(engine.ObjectValue).Fields()
		assert.Equal(t, "https://auth.example.com", obj["issuer"].(engine.StringValue).Value())
		assert.Equal(t, "https://auth.example.com/authorize", obj["authorization_endpoint"].(engine.StringValue).Value())
		assert.Equal(t, "https://auth.example.com/token", obj["token_endpoint"].(engine.StringValue).Value())

		// Check arrays
		responseTypes := obj["response_types_supported"].(engine.ArrayValue).Elements()
		assert.Len(t, responseTypes, 3)
		assert.Equal(t, "code", responseTypes[0].(engine.StringValue).Value())
	})

	t.Run("parseJWTClaims", func(t *testing.T) {
		// This would fail with a real JWT parsing, so we expect an error
		// In a real test, we'd use a valid JWT token
		_, err := bridge.ExecuteMethod(ctx, "parseJWTClaims", []engine.ScriptValue{
			sv("invalid-jwt"),
		})
		assert.Error(t, err)
	})

	t.Run("autoRefreshToken", func(t *testing.T) {
		result, err := bridge.ExecuteMethod(ctx, "autoRefreshToken", []engine.ScriptValue{
			svMap(map[string]interface{}{
				"type": "oauth2",
				"data": map[string]interface{}{
					"access_token": "test-token",
				},
			}),
			sv(600), // 10 minutes
		})
		require.NoError(t, err)
		require.NotNil(t, result)

		obj := result.(engine.ObjectValue).Fields()
		assert.True(t, obj["enabled"].(engine.BoolValue).Value())
		assert.Equal(t, float64(600), obj["refreshBefore"].(engine.NumberValue).Value())
		assert.NotNil(t, obj["nextRefresh"])
	})
}

func TestUtilAuthBridgeMultiSchemeAuth(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Register auth scheme
	result, err := bridge.ExecuteMethod(ctx, "registerAuthScheme", []engine.ScriptValue{
		sv("/api/v1"),
		svMap(map[string]interface{}{
			"type":        "bearer",
			"description": "Bearer token authentication",
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.(engine.BoolValue).Value())

	// Get auth schemes
	result, err = bridge.ExecuteMethod(ctx, "getAuthSchemes", []engine.ScriptValue{
		sv("/api/v1/users"),
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	schemes := result.(engine.ArrayValue).Elements()
	assert.Len(t, schemes, 1)
	scheme := schemes[0].(engine.ObjectValue).Fields()
	assert.Equal(t, "bearer", scheme["type"].(engine.StringValue).Value())
}

func TestUtilAuthBridgeCredentialSerialization(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	authConfig := svMap(map[string]interface{}{
		"type": "apiKey",
		"data": map[string]interface{}{
			"key": "secret-key-123",
		},
	})

	// Serialize credentials
	serialized, err := bridge.ExecuteMethod(ctx, "serializeCredentials", []engine.ScriptValue{
		authConfig,
	})
	require.NoError(t, err)
	require.NotNil(t, serialized)
	assert.Equal(t, engine.TypeString, serialized.Type())

	// Deserialize credentials
	deserialized, err := bridge.ExecuteMethod(ctx, "deserializeCredentials", []engine.ScriptValue{
		serialized,
	})
	require.NoError(t, err)
	require.NotNil(t, deserialized)
	assert.Equal(t, engine.TypeObject, deserialized.Type())
}

func TestUtilAuthBridgeCredentialCaching(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	authConfig := svMap(map[string]interface{}{
		"type": "bearer",
		"data": map[string]interface{}{
			"token": "cached-token",
		},
	})

	// Cache credentials
	result, err := bridge.ExecuteMethod(ctx, "cacheCredentials", []engine.ScriptValue{
		sv("test-cache-key"),
		authConfig,
		sv(1800), // 30 minutes TTL
	})
	require.NoError(t, err)
	assert.True(t, result.(engine.BoolValue).Value())

	// Verify cache entry exists
	bridge.mu.RLock()
	entry, exists := bridge.credentialCache["test-cache-key"]
	bridge.mu.RUnlock()

	assert.True(t, exists)
	assert.NotNil(t, entry)
	assert.Equal(t, 1800, entry.Metadata["ttl"])
	assert.WithinDuration(t, time.Now(), entry.CreatedAt, 1*time.Second)
}

func TestUtilAuthBridgeEventLogging(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()
	err := bridge.Initialize(ctx)
	require.NoError(t, err)

	// Log auth event
	result, err := bridge.ExecuteMethod(ctx, "logAuthEvent", []engine.ScriptValue{
		sv("login"),
		svMap(map[string]interface{}{
			"user":     "test-user",
			"provider": "oauth2",
			"success":  true,
		}),
	})
	require.NoError(t, err)
	assert.True(t, result.IsNil())
}

func TestUtilAuthBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilAuthBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("createAuthConfig", []engine.ScriptValue{
		sv("bearer"),
		svMap(map[string]interface{}{}),
	})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", []engine.ScriptValue{})
	assert.NoError(t, err)
}

func TestUtilAuthBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilAuthBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 3)

	// Check for expected permissions
	hasProcess := false
	hasNetwork := false
	hasMemory := false

	for _, perm := range permissions {
		switch perm.Type {
		case engine.PermissionProcess:
			if perm.Resource == "environment" {
				hasProcess = true
				assert.Contains(t, perm.Actions, "read")
			}
		case engine.PermissionNetwork:
			if perm.Resource == "oauth2" {
				hasNetwork = true
				assert.Contains(t, perm.Actions, "token")
			}
		case engine.PermissionMemory:
			if perm.Resource == "credentials" {
				hasMemory = true
				assert.Contains(t, perm.Actions, "read")
				assert.Contains(t, perm.Actions, "mask")
			}
		}
	}

	assert.True(t, hasProcess, "Process permission not found")
	assert.True(t, hasNetwork, "Network permission not found")
	assert.True(t, hasMemory, "Memory permission not found")
}

func TestUtilAuthBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilAuthBridge()
	mappings := bridge.TypeMappings()

	// Check expected type mappings
	assert.Contains(t, mappings, "AuthConfig")
	assert.Contains(t, mappings, "AuthScheme")
	assert.Contains(t, mappings, "OAuth2Config")

	// Verify mapping properties
	authConfigMapping := mappings["AuthConfig"]
	assert.Equal(t, "AuthConfig", authConfigMapping.GoType)
	assert.Equal(t, "object", authConfigMapping.ScriptType)
}

func TestUtilAuthBridgeErrorHandling(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()

	// Test method execution before initialization
	_, err := bridge.ExecuteMethod(ctx, "createAuthConfig", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Equal(t, ErrBridgeNotInitialized, err)

	// Initialize bridge
	err = bridge.Initialize(ctx)
	require.NoError(t, err)

	// Test unknown method
	_, err = bridge.ExecuteMethod(ctx, "unknownMethod", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "method not found")

	// Test invalid arguments
	_, err = bridge.ExecuteMethod(ctx, "createAuthConfig", []engine.ScriptValue{})
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArguments, err)
}

func TestUtilAuthBridgeWithEventEmitter(t *testing.T) {
	// Create a mock event emitter
	// In a real test, you'd use a proper mock or test double
	bridge := NewUtilAuthBridgeWithEventEmitter(nil)
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_auth", bridge.GetID())
}
