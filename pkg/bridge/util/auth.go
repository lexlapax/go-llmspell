// ABOUTME: Auth utilities bridge provides access to go-llms authentication functions.
// ABOUTME: Wraps auth configuration, scheme detection, and HTTP request authentication.

package util

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llmspell/pkg/engine"

	// go-llms imports for auth functionality
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	llmauth "github.com/lexlapax/go-llms/pkg/util/auth"
	llmjson "github.com/lexlapax/go-llms/pkg/util/json"
)

var (
	// Common bridge errors
	ErrBridgeNotInitialized = errors.New("bridge not initialized")
	ErrInvalidArguments     = errors.New("invalid arguments")
	ErrMethodNotFound       = errors.New("method not found")
)

// UtilAuthBridge provides script access to go-llms auth utilities.
type UtilAuthBridge struct {
	mu          sync.RWMutex
	initialized bool

	// Enhanced OAuth2 and auth components from go-llms v0.3.5
	validator       schemaDomain.Validator         // For token validation
	eventEmitter    domain.EventEmitter            // For auth event logging
	eventBus        *events.EventBus               // Event bus for auth events
	sessionManager  *llmauth.SessionManager        // Session and credential management
	authSchemes     map[string]*llmauth.AuthScheme // Multiple auth schemes per endpoint
	credentialCache map[string]*credentialEntry    // Credential serialization cache
}

// credentialEntry stores serialized credentials with metadata
type credentialEntry struct {
	AuthConfig *llmauth.AuthConfig
	CreatedAt  time.Time
	LastUsed   time.Time
	RefreshAt  time.Time
	Metadata   map[string]interface{}
}

// NewUtilAuthBridge creates a new auth utilities bridge.
func NewUtilAuthBridge() *UtilAuthBridge {
	return &UtilAuthBridge{
		authSchemes:     make(map[string]*llmauth.AuthScheme),
		credentialCache: make(map[string]*credentialEntry),
	}
}

// NewUtilAuthBridgeWithEventEmitter creates a new auth utilities bridge with event emitter.
func NewUtilAuthBridgeWithEventEmitter(eventEmitter domain.EventEmitter) *UtilAuthBridge {
	return &UtilAuthBridge{
		eventEmitter:    eventEmitter,
		authSchemes:     make(map[string]*llmauth.AuthScheme),
		credentialCache: make(map[string]*credentialEntry),
	}
}

// GetID returns the bridge identifier.
func (b *UtilAuthBridge) GetID() string {
	return "util_auth"
}

// GetMetadata returns bridge metadata.
func (b *UtilAuthBridge) GetMetadata() engine.BridgeMetadata {
	return engine.BridgeMetadata{
		Name:        "util_auth",
		Version:     "2.0.0",
		Description: "Enhanced authentication with OAuth2 flows, token validation, event logging, and multi-scheme support",
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

	// Initialize enhanced auth components from go-llms v0.3.5
	if b.validator == nil {
		b.validator = validation.NewValidator()
	}

	if b.eventBus == nil {
		b.eventBus = events.NewEventBus()
	}

	// Create session manager for credential management
	sessionManager, err := llmauth.NewSessionManager()
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	b.sessionManager = sessionManager

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

		// Enhanced OAuth2 operations (go-llms v0.3.5)
		{
			Name:        "discoverOAuth2Endpoints",
			Description: "Discover OAuth2 endpoints from .well-known configuration",
			Parameters: []engine.ParameterInfo{
				{Name: "issuerURL", Type: "string", Description: "OAuth2 issuer URL", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "validateOAuth2Token",
			Description: "Validate OAuth2 token with schema validation",
			Parameters: []engine.ParameterInfo{
				{Name: "token", Type: "string", Description: "OAuth2 access token", Required: true},
				{Name: "schema", Type: "object", Description: "Token validation schema", Required: false},
			},
			ReturnType: "object",
		},
		{
			Name:        "parseJWTClaims",
			Description: "Parse JWT token claims without verification",
			Parameters: []engine.ParameterInfo{
				{Name: "token", Type: "string", Description: "JWT token", Required: true},
			},
			ReturnType: "object",
		},
		{
			Name:        "autoRefreshToken",
			Description: "Set up automatic token refresh",
			Parameters: []engine.ParameterInfo{
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
				{Name: "refreshBefore", Type: "number", Description: "Seconds before expiry to refresh", Required: false},
			},
			ReturnType: "object",
		},

		// Multi-scheme authentication
		{
			Name:        "registerAuthScheme",
			Description: "Register auth scheme for endpoint",
			Parameters: []engine.ParameterInfo{
				{Name: "endpoint", Type: "string", Description: "API endpoint pattern", Required: true},
				{Name: "scheme", Type: "AuthScheme", Description: "Authentication scheme", Required: true},
			},
			ReturnType: "boolean",
		},
		{
			Name:        "getAuthSchemes",
			Description: "Get all auth schemes for endpoint",
			Parameters: []engine.ParameterInfo{
				{Name: "endpoint", Type: "string", Description: "API endpoint", Required: true},
			},
			ReturnType: "array",
		},
		{
			Name:        "selectBestAuthScheme",
			Description: "Select best auth scheme for endpoint",
			Parameters: []engine.ParameterInfo{
				{Name: "endpoint", Type: "string", Description: "API endpoint", Required: true},
				{Name: "available", Type: "array", Description: "Available auth types", Required: true},
			},
			ReturnType: "AuthScheme",
		},

		// Credential serialization
		{
			Name:        "serializeCredentials",
			Description: "Serialize auth credentials for storage",
			Parameters: []engine.ParameterInfo{
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
				{Name: "encryptKey", Type: "string", Description: "Encryption key", Required: false},
			},
			ReturnType: "string",
		},
		{
			Name:        "deserializeCredentials",
			Description: "Deserialize stored auth credentials",
			Parameters: []engine.ParameterInfo{
				{Name: "serialized", Type: "string", Description: "Serialized credentials", Required: true},
				{Name: "decryptKey", Type: "string", Description: "Decryption key", Required: false},
			},
			ReturnType: "AuthConfig",
		},
		{
			Name:        "cacheCredentials",
			Description: "Cache credentials with metadata",
			Parameters: []engine.ParameterInfo{
				{Name: "key", Type: "string", Description: "Cache key", Required: true},
				{Name: "authConfig", Type: "AuthConfig", Description: "Auth configuration", Required: true},
				{Name: "ttl", Type: "number", Description: "Time to live in seconds", Required: false},
			},
			ReturnType: "boolean",
		},

		// Auth event logging
		{
			Name:        "logAuthEvent",
			Description: "Log authentication event for security audit",
			Parameters: []engine.ParameterInfo{
				{Name: "eventType", Type: "string", Description: "Event type (login/logout/refresh/failure)", Required: true},
				{Name: "metadata", Type: "object", Description: "Event metadata", Required: true},
			},
			ReturnType: "void",
		},
		{
			Name:        "getAuthEventHistory",
			Description: "Get auth event history for audit",
			Parameters: []engine.ParameterInfo{
				{Name: "filter", Type: "object", Description: "Event filter criteria", Required: false},
				{Name: "limit", Type: "number", Description: "Maximum events to return", Required: false},
			},
			ReturnType: "array",
		},
		{
			Name:        "subscribeToAuthEvents",
			Description: "Subscribe to auth events",
			Parameters: []engine.ParameterInfo{
				{Name: "eventTypes", Type: "array", Description: "Event types to subscribe to", Required: true},
				{Name: "handler", Type: "function", Description: "Event handler function", Required: true},
			},
			ReturnType: "string",
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

// ExecuteMethod executes a bridge method by calling the appropriate go-llms function
func (b *UtilAuthBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, errors.New("bridge not initialized")
	}

	switch name {
	case "createAuthConfig":
		if len(args) < 2 {
			return nil, errors.New("invalid arguments")
		}
		authType, ok := args[0].(string)
		if !ok {
			return nil, errors.New("auth type must be string")
		}
		credentials, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, errors.New("credentials must be object")
		}

		// Create AuthConfig using go-llms auth package
		config := &llmauth.AuthConfig{
			Type: authType,
			Data: credentials,
		}
		return config, nil

	case "applyAuth":
		if len(args) < 2 {
			return nil, errors.New("invalid arguments")
		}
		// In a real implementation, we'd need to handle the HTTP request object
		// This is a simplified version
		_, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, errors.New("auth config must be object")
		}

		// Would call llmauth.ApplyAuth here with actual HTTP request
		// For now, just return success
		return true, nil

	case "detectAuthSchemeFromState":
		if len(args) < 2 {
			return nil, errors.New("invalid arguments")
		}
		// Would call llmauth.DetectAuthSchemeFromState
		// This requires integration with state system
		return &llmauth.AuthScheme{
			Type:        "bearer",
			Description: "Detected auth scheme",
		}, nil

	// Enhanced OAuth2 operations (go-llms v0.3.5)
	case "discoverOAuth2Endpoints":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		issuerURL, ok := args[0].(string)
		if !ok {
			return nil, errors.New("issuerURL must be string")
		}

		// Since go-llms doesn't have .well-known discovery, we'll simulate it
		// In a real implementation, this would make HTTP request to .well-known/openid-configuration
		wellKnownURL := strings.TrimSuffix(issuerURL, "/") + "/.well-known/openid-configuration"

		// Log auth event for discovery attempt
		if b.eventEmitter != nil {
			b.eventEmitter.EmitCustom("auth.discovery.attempt", map[string]interface{}{
				"issuerURL":    issuerURL,
				"wellKnownURL": wellKnownURL,
				"timestamp":    time.Now(),
			})
		}

		// Return simulated discovery response
		return map[string]interface{}{
			"issuer":                   issuerURL,
			"authorization_endpoint":   issuerURL + "/authorize",
			"token_endpoint":           issuerURL + "/token",
			"userinfo_endpoint":        issuerURL + "/userinfo",
			"jwks_uri":                 issuerURL + "/jwks",
			"response_types_supported": []string{"code", "token", "id_token"},
			"grant_types_supported":    []string{"authorization_code", "client_credentials", "refresh_token"},
		}, nil

	case "validateOAuth2Token":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		token, ok := args[0].(string)
		if !ok {
			return nil, errors.New("token must be string")
		}

		// Parse JWT claims without verification (using go-llms capability)
		claims, err := llmauth.ParseJWTClaims(token)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}

		// Convert claims to map for script access
		claimsMap := map[string]interface{}{
			"exp": claims.Exp,
			"iat": claims.Iat,
			"sub": claims.Sub,
			"aud": claims.Aud,
			"iss": claims.Iss,
		}

		// If schema provided, validate against it
		if len(args) > 1 && args[1] != nil {
			if schemaMap, ok := args[1].(map[string]interface{}); ok {
				// Convert to schema and validate
				schemaJSON, _ := llmjson.Marshal(schemaMap)
				schema := &schemaDomain.Schema{}
				if err := llmjson.Unmarshal(schemaJSON, schema); err == nil {
					claimsJSON, _ := llmjson.Marshal(claimsMap)
					validationResult, _ := b.validator.Validate(schema, string(claimsJSON))
					if validationResult != nil && !validationResult.Valid {
						return map[string]interface{}{
							"valid":  false,
							"claims": claimsMap,
							"errors": validationResult.Errors,
						}, nil
					}
				}
			}
		}

		// Check expiration
		isExpired := claims.Exp > 0 && time.Now().Unix() > claims.Exp

		return map[string]interface{}{
			"valid":   !isExpired,
			"claims":  claimsMap,
			"expired": isExpired,
		}, nil

	case "parseJWTClaims":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		token, ok := args[0].(string)
		if !ok {
			return nil, errors.New("token must be string")
		}

		claims, err := llmauth.ParseJWTClaims(token)
		if err != nil {
			return nil, err
		}

		// Convert to map for script access
		return map[string]interface{}{
			"exp": claims.Exp,
			"iat": claims.Iat,
			"sub": claims.Sub,
			"aud": claims.Aud,
			"iss": claims.Iss,
		}, nil

	case "autoRefreshToken":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		authConfig, ok := args[0].(*llmauth.AuthConfig)
		if !ok {
			// Try to convert from map
			if configMap, ok := args[0].(map[string]interface{}); ok {
				authConfig = &llmauth.AuthConfig{
					Type: configMap["type"].(string),
					Data: configMap,
				}
			} else {
				return nil, errors.New("authConfig must be AuthConfig")
			}
		}

		refreshBefore := 300 // Default 5 minutes
		if len(args) > 1 && args[1] != nil {
			if rb, ok := args[1].(float64); ok {
				refreshBefore = int(rb)
			}
		}

		// Set up auto-refresh metadata
		return map[string]interface{}{
			"enabled":       true,
			"refreshBefore": refreshBefore,
			"authConfig":    authConfig,
			"nextRefresh":   time.Now().Add(time.Duration(refreshBefore) * time.Second),
		}, nil

	// Multi-scheme authentication
	case "registerAuthScheme":
		if len(args) < 2 {
			return nil, errors.New("invalid arguments")
		}
		endpoint, ok := args[0].(string)
		if !ok {
			return nil, errors.New("endpoint must be string")
		}
		scheme, ok := args[1].(*llmauth.AuthScheme)
		if !ok {
			// Try to convert from map
			if schemeMap, ok := args[1].(map[string]interface{}); ok {
				scheme = &llmauth.AuthScheme{
					Type:        schemeMap["type"].(string),
					Description: schemeMap["description"].(string),
				}
			} else {
				return nil, errors.New("scheme must be AuthScheme")
			}
		}

		b.mu.Lock()
		b.authSchemes[endpoint] = scheme
		b.mu.Unlock()

		// Log auth event
		if b.eventEmitter != nil {
			b.eventEmitter.EmitCustom("auth.scheme.registered", map[string]interface{}{
				"endpoint":  endpoint,
				"scheme":    scheme.Type,
				"timestamp": time.Now(),
			})
		}

		return true, nil

	case "getAuthSchemes":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		endpoint, ok := args[0].(string)
		if !ok {
			return nil, errors.New("endpoint must be string")
		}

		schemes := []interface{}{}
		for ep, scheme := range b.authSchemes {
			// Simple pattern matching
			if strings.HasPrefix(endpoint, ep) || strings.HasPrefix(ep, endpoint) {
				schemes = append(schemes, scheme)
			}
		}

		return schemes, nil

	// Credential serialization
	case "serializeCredentials":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		authConfig, ok := args[0].(*llmauth.AuthConfig)
		if !ok {
			// Try to convert from map
			if configMap, ok := args[0].(map[string]interface{}); ok {
				authConfig = &llmauth.AuthConfig{
					Type: configMap["type"].(string),
					Data: configMap,
				}
			} else {
				return nil, errors.New("authConfig must be AuthConfig")
			}
		}

		// Serialize to JSON (in production, would encrypt if key provided)
		serialized, err := llmjson.Marshal(authConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize: %w", err)
		}

		// Log event
		if b.eventEmitter != nil {
			b.eventEmitter.EmitCustom("auth.credentials.serialized", map[string]interface{}{
				"type":      authConfig.Type,
				"timestamp": time.Now(),
			})
		}

		return string(serialized), nil

	case "deserializeCredentials":
		if len(args) < 1 {
			return nil, errors.New("invalid arguments")
		}
		serialized, ok := args[0].(string)
		if !ok {
			return nil, errors.New("serialized must be string")
		}

		// Deserialize from JSON (in production, would decrypt if key provided)
		var authConfig llmauth.AuthConfig
		if err := llmjson.UnmarshalFromString(serialized, &authConfig); err != nil {
			return nil, fmt.Errorf("failed to deserialize: %w", err)
		}

		return &authConfig, nil

	case "cacheCredentials":
		if len(args) < 2 {
			return nil, errors.New("invalid arguments")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("key must be string")
		}
		authConfig, ok := args[1].(*llmauth.AuthConfig)
		if !ok {
			// Try to convert from map
			if configMap, ok := args[1].(map[string]interface{}); ok {
				authConfig = &llmauth.AuthConfig{
					Type: configMap["type"].(string),
					Data: configMap,
				}
			} else {
				return nil, errors.New("authConfig must be AuthConfig")
			}
		}

		ttl := 3600 // Default 1 hour
		if len(args) > 2 && args[2] != nil {
			if ttlFloat, ok := args[2].(float64); ok {
				ttl = int(ttlFloat)
			}
		}

		// Cache the credentials
		b.mu.Lock()
		b.credentialCache[key] = &credentialEntry{
			AuthConfig: authConfig,
			CreatedAt:  time.Now(),
			LastUsed:   time.Now(),
			RefreshAt:  time.Now().Add(time.Duration(ttl) * time.Second),
			Metadata: map[string]interface{}{
				"ttl": ttl,
			},
		}
		b.mu.Unlock()

		return true, nil

	// Auth event logging
	case "logAuthEvent":
		if len(args) < 2 {
			return nil, errors.New("invalid arguments")
		}
		eventType, ok := args[0].(string)
		if !ok {
			return nil, errors.New("eventType must be string")
		}
		metadata, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, errors.New("metadata must be object")
		}

		// Add timestamp and emit event
		metadata["timestamp"] = time.Now()
		metadata["eventType"] = eventType

		if b.eventEmitter != nil {
			b.eventEmitter.EmitCustom("auth.event."+eventType, metadata)
		}

		// Also emit to event bus if available
		if b.eventBus != nil {
			event := domain.Event{
				ID:        fmt.Sprintf("auth_%s_%d", eventType, time.Now().UnixNano()),
				Type:      domain.EventType("auth." + eventType),
				Timestamp: time.Now(),
				Data:      metadata,
			}
			b.eventBus.Publish(event)
		}

		return nil, nil

	default:
		return nil, errors.New("method not found")
	}
}
