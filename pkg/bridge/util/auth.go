// ABOUTME: Auth utilities bridge provides access to go-llms authentication functions.
// ABOUTME: Wraps auth configuration, scheme detection, and HTTP request authentication.

package util

import (
	"context"
	"sync"

	"github.com/lexlapax/go-llmspell/pkg/engine"
)

// UtilAuthBridge provides script access to go-llms auth utilities.
type UtilAuthBridge struct {
	mu          sync.RWMutex
	initialized bool
}

// NewUtilAuthBridge creates a new auth utilities bridge.
func NewUtilAuthBridge() *UtilAuthBridge {
	return &UtilAuthBridge{}
}

// GetID returns the bridge identifier.
func (b *UtilAuthBridge) GetID() string {
	return "util_auth"
}

// GetMetadata returns bridge metadata.
func (b *UtilAuthBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_auth",
		Version:     "1.0.0",
		Description: "Authentication utilities for HTTP requests and provider auth",
		Author:      "go-llmspell",
		License:     "MIT",
	}
}

// Initialize initializes the bridge.
func (b *UtilAuthBridge) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	b.initialized = true
	return nil
}

// Cleanup cleans up bridge resources.
func (b *UtilAuthBridge) Cleanup(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.initialized = false
	return nil
}

// IsInitialized checks if the bridge is initialized.
func (b *UtilAuthBridge) IsInitialized() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.initialized
}

// RegisterWithEngine registers the bridge with a script engine.
func (b *UtilAuthBridge) RegisterWithEngine(engine engine.ScriptEngine) error {
	return engine.RegisterBridge(b)
}

// Methods returns the methods exposed by this bridge.
func (b *UtilAuthBridge) Methods() []engine.MethodInfo {
	return []engine.MethodInfo{
		// Auth configuration
		{
			Name:        "createAuthConfig",
			Description: "Create authentication configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "type", Type: "string", Description: "Auth type (apiKey/bearer/basic/oauth2)", Required: true},
				{Name: "credentials", Type: "object", Description: "Auth credentials", Required: true},
			},
			ReturnType: "AuthConfig",
		},
		{
			Name:        "createAuthFromEnv",
			Description: "Create auth config from environment variables",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "AuthConfig",
		},
		{
			Name:        "createAuthFromState",
			Description: "Create auth config from agent state",
			Parameters: []engine.ParameterInfo{
				{Name: "state", Type: "State", Description: "Agent state", Required: true},
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "AuthConfig",
		},

		// HTTP request authentication
		{
			Name:        "applyAuth",
			Description: "Apply authentication to HTTP request",
			Parameters: []engine.ParameterInfo{
				{Name: "request", Type: "object", Description: "HTTP request", Required: true},
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "applyAuthToHeaders",
			Description: "Apply authentication to headers map",
			Parameters: []engine.ParameterInfo{
				{Name: "headers", Type: "object", Description: "Headers map", Required: true},
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
			},
			ReturnType: "object",
		},

		// Auth scheme utilities
		{
			Name:        "detectAuthScheme",
			Description: "Detect authentication scheme from configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "config", Type: "object", Description: "Configuration object", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "parseAuthHeader",
			Description: "Parse authentication header",
			Parameters: []engine.ParameterInfo{
				{Name: "header", Type: "string", Description: "Auth header value", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateAuthConfig",
			Description: "Validate authentication configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
			},
			ReturnType: "boolean",
		},

		// OAuth2 utilities
		{
			Name:        "createOAuth2Config",
			Description: "Create OAuth2 configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "clientID", Type: "string", Description: "OAuth2 client ID", Required: true},
				{Name: "clientSecret", Type: "string", Description: "OAuth2 client secret", Required: true},
				{Name: "tokenURL", Type: "string", Description: "Token endpoint URL", Required: true},
				{Name: "scopes", Type: "array", Description: "OAuth2 scopes", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "refreshOAuth2Token",
			Description: "Refresh OAuth2 access token",
			Parameters: []engine.ParameterInfo{
				{Name: "oauth2Config", Type: "object", Description: "OAuth2 configuration", Required: true},
				{Name: "refreshToken", Type: "string", Description: "Refresh token", Required: true},
			},
			ReturnType: "object",
		},

		// Session management
		{
			Name:        "createAuthSession",
			Description: "Create authentication session",
			Parameters: []engine.ParameterInfo{
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
				{Name: "sessionID", Type: "string", Description: "Session identifier", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateSession",
			Description: "Validate authentication session",
			Parameters: []engine.ParameterInfo{
				{Name: "session", Type: "object", Description: "Auth session", Required: true},
			},
			ReturnType: "boolean",
		},

		// Credential management
		{
			Name:        "maskCredentials",
			Description: "Mask sensitive credentials in logs",
			Parameters: []engine.ParameterInfo{
				{Name: "text", Type: "string", Description: "Text containing credentials", Required: true},
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
			},
			ReturnType: "string",
		},
		{
			Name:        "rotateAPIKey",
			Description: "Generate rotated API key suggestion",
			Parameters: []engine.ParameterInfo{
				{Name: "provider", Type: "string", Description: "Provider name", Required: true},
			},
			ReturnType: "string",
		},
	}
}

// TypeMappings returns type conversion mappings.
func (b *UtilAuthBridge) TypeMappings() map[string]engine.TypeMapping {
	return map[string]engine.TypeMapping{
		"AuthConfig": {
			GoType:     "AuthConfig",
			ScriptType: "object",
		},
		"AuthScheme": {
			GoType:     "AuthScheme",
			ScriptType: "object",
		},
		"OAuth2Config": {
			GoType:     "OAuth2Config",
			ScriptType: "object",
		},
	}
}

// ValidateMethod validates method calls.
func (b *UtilAuthBridge) ValidateMethod(name string, args []interface{}) error {
	// Method validation handled by engine based on Methods() metadata
	return nil
}

// RequiredPermissions returns required permissions.
func (b *UtilAuthBridge) RequiredPermissions() []engine.Permission {
	return []engine.Permission{
		{
			Type:        engine.PermissionProcess,
			Resource:    "environment",
			Actions:     []string{"read"},
			Description: "Read authentication credentials from environment",
		},
		{
			Type:        engine.PermissionNetwork,
			Resource:    "oauth2",
			Actions:     []string{"token"},
			Description: "OAuth2 token operations",
		},
		{
			Type:        engine.PermissionMemory,
			Resource:    "credentials",
			Actions:     []string{"read", "mask"},
			Description: "Handle authentication credentials",
		},
	}
}

// The actual method implementations would be provided by the script engine
// which would call the appropriate go-llms/pkg/util/auth functions.
// For example:
// - createAuthConfig would create an auth.AuthConfig
// - applyAuth would call auth.ApplyAuth
// - detectAuthScheme would parse auth schemes
// etc.
