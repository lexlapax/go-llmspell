// ABOUTME: Test suite for the auth utilities bridge that wraps go-llms authentication functions.
// ABOUTME: Tests bridge interface compliance and method definitions.

package util

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/stretchr/testify/assert"
)

func TestNewUtilAuthBridge(t *testing.T) {
	bridge := NewUtilAuthBridge()
	assert.NotNil(t, bridge)
	assert.Equal(t, "util_auth", bridge.GetID())
}

func TestUtilAuthBridgeMetadata(t *testing.T) {
	bridge := NewUtilAuthBridge()
	metadata := bridge.GetMetadata()

	assert.Equal(t, "util_auth", metadata.Name)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Contains(t, metadata.Description, "Authentication utilities")
	assert.Equal(t, "go-llmspell", metadata.Author)
	assert.Equal(t, "MIT", metadata.License)
}

func TestUtilAuthBridgeInitialization(t *testing.T) {
	bridge := NewUtilAuthBridge()
	ctx := context.Background()

	// Test initialization
	assert.False(t, bridge.IsInitialized())
	err := bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test double initialization
	err = bridge.Initialize(ctx)
	assert.NoError(t, err)
	assert.True(t, bridge.IsInitialized())

	// Test cleanup
	err = bridge.Cleanup(ctx)
	assert.NoError(t, err)
	assert.False(t, bridge.IsInitialized())
}

func TestUtilAuthBridgeMethods(t *testing.T) {
	bridge := NewUtilAuthBridge()
	methods := bridge.Methods()

	// Check that all expected method categories are present
	expectedMethods := map[string]bool{
		// Auth configuration
		"createAuthConfig":    false,
		"createAuthFromEnv":   false,
		"createAuthFromState": false,

		// HTTP request authentication
		"applyAuth":          false,
		"applyAuthToHeaders": false,

		// Auth scheme utilities
		"detectAuthScheme":   false,
		"parseAuthHeader":    false,
		"validateAuthConfig": false,

		// OAuth2 utilities
		"createOAuth2Config": false,
		"refreshOAuth2Token": false,

		// Session management
		"createAuthSession": false,
		"validateSession":   false,

		// Credential management
		"maskCredentials": false,
		"rotateAPIKey":    false,
	}

	for _, method := range methods {
		if _, ok := expectedMethods[method.Name]; ok {
			expectedMethods[method.Name] = true
		}
	}

	for method, found := range expectedMethods {
		assert.True(t, found, "Method %s not found", method)
	}
}

func TestUtilAuthBridgeMethodDetails(t *testing.T) {
	bridge := NewUtilAuthBridge()
	methods := bridge.Methods()

	// Verify createAuthConfig method details
	var createAuthMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "createAuthConfig" {
			createAuthMethod = &m
			break
		}
	}
	assert.NotNil(t, createAuthMethod)
	assert.Contains(t, createAuthMethod.Description, "authentication configuration")
	assert.Len(t, createAuthMethod.Parameters, 2)
	assert.Equal(t, "AuthConfig", createAuthMethod.ReturnType)

	// Verify applyAuth method details
	var applyAuthMethod *engine.MethodInfo
	for _, m := range methods {
		if m.Name == "applyAuth" {
			applyAuthMethod = &m
			break
		}
	}
	assert.NotNil(t, applyAuthMethod)
	assert.Contains(t, applyAuthMethod.Description, "HTTP request")
	assert.Len(t, applyAuthMethod.Parameters, 2)
	assert.Equal(t, "object", applyAuthMethod.ReturnType)
}

func TestUtilAuthBridgeTypeMappings(t *testing.T) {
	bridge := NewUtilAuthBridge()
	mappings := bridge.TypeMappings()

	// Check that expected type mappings are present
	expectedTypes := []string{"AuthConfig", "AuthScheme", "OAuth2Config"}
	for _, typeName := range expectedTypes {
		mapping, ok := mappings[typeName]
		assert.True(t, ok, "Type mapping for %s not found", typeName)
		assert.NotEmpty(t, mapping.GoType)
		assert.Equal(t, "object", mapping.ScriptType)
	}
}

func TestUtilAuthBridgeRequiredPermissions(t *testing.T) {
	bridge := NewUtilAuthBridge()
	permissions := bridge.RequiredPermissions()

	assert.GreaterOrEqual(t, len(permissions), 3)

	// Check for required permissions
	hasProcessPerm := false
	hasNetworkPerm := false
	hasMemoryPerm := false

	for _, perm := range permissions {
		if perm.Type == "process" && perm.Resource == "environment" {
			hasProcessPerm = true
			assert.Contains(t, perm.Actions, "read")
		}
		if perm.Type == "network" && perm.Resource == "oauth2" {
			hasNetworkPerm = true
			assert.Contains(t, perm.Actions, "token")
		}
		if perm.Type == "memory" && perm.Resource == "credentials" {
			hasMemoryPerm = true
			assert.Contains(t, perm.Actions, "read")
			assert.Contains(t, perm.Actions, "mask")
		}
	}

	assert.True(t, hasProcessPerm, "Process permission not found")
	assert.True(t, hasNetworkPerm, "Network permission not found")
	assert.True(t, hasMemoryPerm, "Memory permission not found")
}

func TestUtilAuthBridgeValidateMethod(t *testing.T) {
	bridge := NewUtilAuthBridge()

	// ValidateMethod should always return nil as validation is handled by engine
	err := bridge.ValidateMethod("createAuthConfig", []interface{}{"apiKey", map[string]interface{}{}})
	assert.NoError(t, err)

	err = bridge.ValidateMethod("unknownMethod", nil)
	assert.NoError(t, err)
}

// Note: Actual auth utility testing would require real go-llms implementations
// or would be done at integration test level with actual utilities
