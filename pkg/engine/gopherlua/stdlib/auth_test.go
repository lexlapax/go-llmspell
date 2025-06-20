// ABOUTME: Comprehensive test suite for Authentication & Security Library in Lua standard library
// ABOUTME: Tests authentication, OAuth2, sessions, security policies, permissions, and audit logging

package stdlib

import (
	"path/filepath"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// setupAuthLibrary loads the auth library and sets up required bridges
func setupAuthLibrary(t *testing.T, L *lua.LState) {
	t.Helper()

	// Set up mock auth bridge
	authTable := L.NewTable()

	// Mock auth configuration methods
	authTable.RawSetString("createAuthConfig", L.NewFunction(func(L *lua.LState) int {
		authType := L.CheckString(1)
		credentials := L.CheckTable(2)
		_ = authType
		_ = credentials

		config := L.NewTable()
		config.RawSetString("type", lua.LString(authType))
		config.RawSetString("id", lua.LString("auth_config_123"))
		L.Push(config)
		return 1
	}))

	authTable.RawSetString("createAuthFromEnv", L.NewFunction(func(L *lua.LState) int {
		provider := L.CheckString(1)
		_ = provider

		config := L.NewTable()
		config.RawSetString("type", lua.LString("api_key"))
		config.RawSetString("provider", lua.LString(provider))
		config.RawSetString("token", lua.LString("env_token_123"))
		L.Push(config)
		return 1
	}))

	authTable.RawSetString("createAuthFromState", L.NewFunction(func(L *lua.LState) int {
		state := L.CheckTable(1)
		provider := L.CheckString(2)
		_ = state
		_ = provider

		config := L.NewTable()
		config.RawSetString("type", lua.LString("bearer"))
		config.RawSetString("provider", lua.LString(provider))
		config.RawSetString("token", lua.LString("state_token_456"))
		L.Push(config)
		return 1
	}))

	authTable.RawSetString("detectAuthScheme", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		L.Push(lua.LString("api_key"))
		return 1
	}))

	// Mock HTTP request authentication
	authTable.RawSetString("applyAuth", L.NewFunction(func(L *lua.LState) int {
		request := L.CheckTable(1)
		authConfig := L.CheckTable(2)
		_ = authConfig

		// Modify request with auth
		request.RawSetString("authenticated", lua.LTrue)
		headers := L.NewTable()
		headers.RawSetString("Authorization", lua.LString("Bearer test_token"))
		request.RawSetString("headers", headers)
		L.Push(request)
		return 1
	}))

	authTable.RawSetString("applyAuthToHeaders", L.NewFunction(func(L *lua.LState) int {
		headers := L.CheckTable(1)
		_ = L.CheckTable(2)

		headers.RawSetString("Authorization", lua.LString("Bearer test_token"))
		L.Push(headers)
		return 1
	}))

	authTable.RawSetString("validateAuthConfig", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		L.Push(lua.LTrue)
		return 1
	}))

	authTable.RawSetString("parseAuthHeader", L.NewFunction(func(L *lua.LState) int {
		header := L.CheckString(1)
		_ = header

		result := L.NewTable()
		result.RawSetString("scheme", lua.LString("Bearer"))
		result.RawSetString("token", lua.LString("parsed_token"))
		L.Push(result)
		return 1
	}))

	// Mock OAuth2 methods
	authTable.RawSetString("createOAuth2Config", L.NewFunction(func(L *lua.LState) int {
		clientID := L.CheckString(1)
		clientSecret := L.CheckString(2)
		tokenURL := L.CheckString(3)
		_ = clientID
		_ = clientSecret
		_ = tokenURL

		config := L.NewTable()
		config.RawSetString("client_id", lua.LString(clientID))
		config.RawSetString("token_url", lua.LString(tokenURL))
		L.Push(config)
		return 1
	}))

	authTable.RawSetString("refreshOAuth2Token", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		_ = L.CheckString(2)

		result := L.NewTable()
		result.RawSetString("access_token", lua.LString("new_access_token"))
		result.RawSetString("refresh_token", lua.LString("new_refresh_token"))
		result.RawSetString("expires_in", lua.LNumber(3600))
		L.Push(result)
		return 1
	}))

	authTable.RawSetString("discoverOAuth2Endpoints", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)

		endpoints := L.NewTable()
		endpoints.RawSetString("authorization_endpoint", lua.LString("https://provider.com/auth"))
		endpoints.RawSetString("token_endpoint", lua.LString("https://provider.com/token"))
		endpoints.RawSetString("userinfo_endpoint", lua.LString("https://provider.com/userinfo"))
		L.Push(endpoints)
		return 1
	}))

	authTable.RawSetString("validateOAuth2Token", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)

		result := L.NewTable()
		result.RawSetString("valid", lua.LTrue)
		result.RawSetString("expires_at", lua.LNumber(1234567890))
		result.RawSetString("scope", lua.LString("read write"))
		L.Push(result)
		return 1
	}))

	authTable.RawSetString("parseJWTClaims", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)

		claims := L.NewTable()
		claims.RawSetString("sub", lua.LString("user123"))
		claims.RawSetString("iss", lua.LString("https://provider.com"))
		claims.RawSetString("exp", lua.LNumber(1234567890))
		claims.RawSetString("iat", lua.LNumber(1234567000))
		L.Push(claims)
		return 1
	}))

	authTable.RawSetString("autoRefreshToken", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		_ = L.OptNumber(2, 300)

		result := L.NewTable()
		result.RawSetString("refresh_scheduled", lua.LTrue)
		result.RawSetString("next_refresh", lua.LNumber(1234567890))
		L.Push(result)
		return 1
	}))

	// Mock multi-scheme authentication
	authTable.RawSetString("registerAuthScheme", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckTable(2)
		L.Push(lua.LTrue)
		return 1
	}))

	authTable.RawSetString("getAuthSchemes", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)

		schemes := L.NewTable()
		scheme1 := L.NewTable()
		scheme1.RawSetString("type", lua.LString("api_key"))
		scheme1.RawSetString("priority", lua.LNumber(1))
		schemes.RawSetInt(1, scheme1)

		scheme2 := L.NewTable()
		scheme2.RawSetString("type", lua.LString("bearer"))
		scheme2.RawSetString("priority", lua.LNumber(2))
		schemes.RawSetInt(2, scheme2)

		L.Push(schemes)
		return 1
	}))

	authTable.RawSetString("selectBestAuthScheme", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckTable(2)

		scheme := L.NewTable()
		scheme.RawSetString("type", lua.LString("bearer"))
		scheme.RawSetString("priority", lua.LNumber(1))
		L.Push(scheme)
		return 1
	}))

	// Mock credential management
	authTable.RawSetString("serializeCredentials", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckTable(1)
		L.Push(lua.LString("serialized_credentials_data"))
		return 1
	}))

	authTable.RawSetString("deserializeCredentials", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)

		config := L.NewTable()
		config.RawSetString("type", lua.LString("api_key"))
		config.RawSetString("token", lua.LString("deserialized_token"))
		L.Push(config)
		return 1
	}))

	authTable.RawSetString("cacheCredentials", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckTable(2)
		L.Push(lua.LTrue)
		return 1
	}))

	// Mock event logging
	authTable.RawSetString("logAuthEvent", L.NewFunction(func(L *lua.LState) int {
		_ = L.CheckString(1)
		_ = L.CheckTable(2)
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("util_auth", authTable)

	// Set up optional mock security manager
	securityTable := L.NewTable()

	securityTable.RawSetString("getSecurityLevel", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("standard"))
		return 1
	}))

	securityTable.RawSetString("getResourceLimits", L.NewFunction(func(L *lua.LState) int {
		limits := L.NewTable()
		limits.RawSetString("max_memory", lua.LNumber(50*1024*1024))
		limits.RawSetString("max_duration", lua.LNumber(60))
		limits.RawSetString("max_instructions", lua.LNumber(10000000))
		L.Push(limits)
		return 1
	}))

	securityTable.RawSetString("getAllowedLibraries", L.NewFunction(func(L *lua.LState) int {
		libraries := L.NewTable()
		libraries.RawSetInt(1, lua.LString("base"))
		libraries.RawSetInt(2, lua.LString("table"))
		libraries.RawSetInt(3, lua.LString("string"))
		libraries.RawSetInt(4, lua.LString("math"))
		L.Push(libraries)
		return 1
	}))

	securityTable.RawSetString("isFunctionDenied", L.NewFunction(func(L *lua.LState) int {
		funcName := L.CheckString(1)
		// Simulate some denied functions
		denied := funcName == "os.execute" || funcName == "os.exit" || funcName == "debug.debug"
		L.Push(lua.LBool(denied))
		return 1
	}))

	L.SetGlobal("security", securityTable)

	// Load the auth library
	authPath := filepath.Join(".", "auth.lua")
	err := L.DoFile(authPath)
	if err != nil {
		t.Fatalf("Failed to load auth library: %v", err)
	}
	auth := L.Get(-1)
	L.SetGlobal("auth", auth)
}

// TestAuthLibraryLoading tests that the auth library can be loaded
func TestAuthLibraryLoading(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	// Check that auth table exists and has expected functions
	script := `
		if type(auth) ~= "table" then
			error("Auth module should be a table")
		end
		
		local required_functions = {
			"create_config", "from_env", "from_state", "detect_scheme",
			"apply_to_request", "apply_to_headers", "validate_config", "parse_header",
			"create_oauth2_config", "refresh_oauth2_token", "discover_oauth2_endpoints",
			"validate_oauth2_token", "parse_jwt_claims", "auto_refresh_token",
			"register_scheme", "get_schemes", "select_best_scheme",
			"create_session", "validate_session", "get_session", "list_sessions",
			"serialize_credentials", "deserialize_credentials", "cache_credentials",
			"create_security_policy", "evaluate_policy", "check_permission", "check_role",
			"log_event", "get_event_history", "subscribe_to_events",
			"get_security_level", "get_resource_limits", "get_allowed_libraries",
			"generate_random_string", "hash_password", "verify_password",
			"get_system_info", "cleanup"
		}
		
		for _, func_name in ipairs(required_functions) do
			if type(auth[func_name]) ~= "function" then
				error("Function " .. func_name .. " should be available")
			end
		end
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Auth library structure test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected library structure test to pass")
	}
}

// TestAuthConfiguration tests authentication configuration creation and management
func TestAuthConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_basic_auth_config",
			script: `
				local config = auth.create_config("api_key", {
					token = "test_token_123",
					header = "X-API-Key"
				})
				
				return config.type == "api_key" and 
				       type(config.config) == "table" and
				       type(config.apply) == "function" and
				       type(config.validate) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected auth config creation to work, got %v", result)
				}
			},
		},
		{
			name: "create_auth_from_env",
			script: `
				local config = auth.from_env("github")
				return config.type == "detected" and type(config.config) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected auth from env to work, got %v", result)
				}
			},
		},
		{
			name: "create_auth_from_state",
			script: `
				local state = {provider_tokens = {github = "token123"}}
				local config = auth.from_state(state, "github")
				return config.type == "detected" and type(config.config) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected auth from state to work, got %v", result)
				}
			},
		},
		{
			name: "detect_auth_scheme",
			script: `
				local scheme = auth.detect_scheme({
					api_key = "test123",
					endpoint = "https://api.example.com"
				})
				return scheme == "api_key"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected scheme detection to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestHTTPAuthentication tests HTTP request authentication
func TestHTTPAuthentication(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "apply_auth_to_request",
			script: `
				local request = {url = "https://api.example.com", method = "GET"}
				local auth_config = {type = "bearer", token = "test123"}
				
				local authed_request = auth.apply_to_request(request, auth_config)
				return authed_request.authenticated == true and
				       type(authed_request.headers) == "table"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected request authentication to work, got %v", result)
				}
			},
		},
		{
			name: "apply_auth_to_headers",
			script: `
				local headers = {["Content-Type"] = "application/json"}
				local auth_config = {type = "bearer", token = "test123"}
				
				local authed_headers = auth.apply_to_headers(headers, auth_config)
				return type(authed_headers.Authorization) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected header authentication to work, got %v", result)
				}
			},
		},
		{
			name: "validate_auth_config",
			script: `
				local auth_config = {type = "api_key", token = "valid_token"}
				local is_valid = auth.validate_config(auth_config)
				return is_valid == true
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected auth config validation to work, got %v", result)
				}
			},
		},
		{
			name: "parse_auth_header",
			script: `
				local parsed = auth.parse_header("Bearer test_token_123")
				return parsed.scheme == "Bearer" and parsed.token == "parsed_token"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected auth header parsing to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestOAuth2Authentication tests OAuth2 functionality
func TestOAuth2Authentication(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_oauth2_config",
			script: `
				local config = auth.create_oauth2_config(
					"client123",
					"secret456",
					"https://provider.com/token",
					{"read", "write"}
				)
				return config.client_id == "client123" and
				       config.token_url == "https://provider.com/token"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected OAuth2 config creation to work, got %v", result)
				}
			},
		},
		{
			name: "refresh_oauth2_token",
			script: `
				local oauth2_config = {client_id = "test"}
				local token_response = auth.refresh_oauth2_token(oauth2_config, "refresh123")
				
				return token_response.access_token == "new_access_token" and
				       token_response.refresh_token == "new_refresh_token" and
				       token_response.expires_in == 3600
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected OAuth2 token refresh to work, got %v", result)
				}
			},
		},
		{
			name: "discover_oauth2_endpoints",
			script: `
				local endpoints = auth.discover_oauth2_endpoints("https://provider.com")
				
				return endpoints.authorization_endpoint == "https://provider.com/auth" and
				       endpoints.token_endpoint == "https://provider.com/token" and
				       endpoints.userinfo_endpoint == "https://provider.com/userinfo"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected OAuth2 endpoint discovery to work, got %v", result)
				}
			},
		},
		{
			name: "validate_oauth2_token",
			script: `
				local validation = auth.validate_oauth2_token("token123")
				
				return validation.valid == true and
				       validation.expires_at == 1234567890 and
				       validation.scope == "read write"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected OAuth2 token validation to work, got %v", result)
				}
			},
		},
		{
			name: "parse_jwt_claims",
			script: `
				local claims = auth.parse_jwt_claims("jwt.token.here")
				
				return claims.sub == "user123" and
				       claims.iss == "https://provider.com" and
				       claims.exp == 1234567890 and
				       claims.iat == 1234567000
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected JWT claims parsing to work, got %v", result)
				}
			},
		},
		{
			name: "auto_refresh_token",
			script: `
				local auth_config = {type = "oauth2", refresh_token = "refresh123"}
				local refresh_setup = auth.auto_refresh_token(auth_config, 600)
				
				return refresh_setup.refresh_scheduled == true and
				       type(refresh_setup.next_refresh) == "number"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected auto token refresh setup to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestMultiSchemeAuthentication tests multi-scheme authentication
func TestMultiSchemeAuthentication(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	script := `
		-- Register multiple auth schemes
		local api_key_scheme = {type = "api_key", priority = 1}
		local bearer_scheme = {type = "bearer", priority = 2}
		
		auth.register_scheme("https://api.example.com", api_key_scheme)
		auth.register_scheme("https://api.example.com", bearer_scheme)
		
		-- Get registered schemes
		local schemes = auth.get_schemes("https://api.example.com")
		
		-- Select best scheme
		local best_scheme = auth.select_best_scheme("https://api.example.com", {"bearer", "api_key"})
		
		return #schemes == 2 and 
		       schemes[1].type == "api_key" and
		       schemes[2].type == "bearer" and
		       best_scheme.type == "bearer"
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Multi-scheme test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected multi-scheme authentication to work correctly, got %v", result)
	}
}

// TestSessionManagement tests session creation and management
func TestSessionManagement(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_and_validate_session",
			script: `
				local auth_config = {type = "bearer", token = "session_token"}
				local session = auth.create_session(auth_config, "session123", {ttl = 3600})
				
				local is_valid, err = auth.validate_session("session123")
				
				return session.id == "session123" and
				       is_valid == true and
				       type(session.validate) == "function" and
				       type(session.refresh) == "function" and
				       type(session.destroy) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected session creation and validation to work, got %v", result)
				}
			},
		},
		{
			name: "list_active_sessions",
			script: `
				-- Create multiple sessions
				local auth_config = {type = "api_key", token = "test"}
				auth.create_session(auth_config, "session1")
				auth.create_session(auth_config, "session2")
				auth.create_session(auth_config, "session3")
				
				local sessions = auth.list_sessions()
				return #sessions >= 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected session listing to work, got %v", result)
				}
			},
		},
		{
			name: "session_refresh_and_destroy",
			script: `
				local auth_config = {type = "bearer", token = "refresh_test"}
				local session = auth.create_session(auth_config, "refresh_session")
				
				-- Refresh session
				session.refresh(7200) -- 2 hours
				
				-- Destroy session
				session.destroy()
				
				local is_valid, err = auth.validate_session("refresh_session")
				return is_valid == false and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected session refresh and destroy to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestCredentialManagement tests credential serialization and caching
func TestCredentialManagement(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	script := `
		local auth_config = {type = "api_key", token = "secret123"}
		
		-- Serialize credentials
		local serialized = auth.serialize_credentials(auth_config, "encrypt_key_123")
		
		-- Deserialize credentials
		local deserialized = auth.deserialize_credentials(serialized, "encrypt_key_123")
		
		-- Cache credentials
		local cached = auth.cache_credentials("cache_key_1", auth_config, 3600)
		
		return type(serialized) == "string" and
		       serialized == "serialized_credentials_data" and
		       type(deserialized) == "table" and
		       deserialized.type == "api_key" and
		       cached == true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Credential management test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected credential management to work correctly, got %v", result)
	}
}

// TestSecurityPolicies tests security policy creation and evaluation
func TestSecurityPolicies(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "create_and_evaluate_permission_policy",
			script: `
				local rules = {
					{type = "permission", permission = "read:users"},
					{type = "permission", permission = "write:posts"}
				}
				
				local policy = auth.create_security_policy("user_policy", rules)
				
				-- Test with user having permissions
				local context = {permissions = {"read:users", "write:posts", "delete:comments"}}
				local allowed, reason = auth.evaluate_policy("user_policy", context)
				
				return policy.name == "user_policy" and
				       allowed == true and
				       type(policy.evaluate) == "function" and
				       type(policy.update) == "function"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected permission policy to work, got %v", result)
				}
			},
		},
		{
			name: "role_based_policy",
			script: `
				local rules = {
					{type = "role", role = "admin"},
					{type = "role", role = "moderator"}
				}
				
				auth.create_security_policy("admin_policy", rules)
				
				-- Test with admin role
				local admin_context = {roles = {"admin", "user"}}
				local admin_allowed, _ = auth.evaluate_policy("admin_policy", admin_context)
				
				-- Test with user role only
				local user_context = {roles = {"user"}}
				local user_allowed, _ = auth.evaluate_policy("admin_policy", user_context)
				
				return admin_allowed == true and user_allowed == false
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected role-based policy to work, got %v", result)
				}
			},
		},
		{
			name: "time_based_policy",
			script: `
				local current_time = os.time()
				local rules = {
					{
						type = "time_based",
						constraint = {
							start_time = current_time - 3600, -- 1 hour ago
							end_time = current_time + 3600,   -- 1 hour from now
							allowed_hours = {9, 10, 11, 12, 13, 14, 15, 16, 17} -- Business hours
						}
					}
				}
				
				auth.create_security_policy("business_hours_policy", rules)
				
				local context = {}
				local allowed, reason = auth.evaluate_policy("business_hours_policy", context)
				
				-- Should be allowed if within time window (may fail based on current hour)
				return type(allowed) == "boolean" and type(reason) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected time-based policy to work, got %v", result)
				}
			},
		},
		{
			name: "ip_whitelist_policy",
			script: `
				local rules = {
					{
						type = "ip_whitelist",
						allowed_ips = {"192.168.1.100", "10.0.0.0/8", "127.0.0.1"}
					}
				}
				
				auth.create_security_policy("ip_policy", rules)
				
				-- Test with allowed IP
				local allowed_context = {client_ip = "192.168.1.100"}
				local allowed, _ = auth.evaluate_policy("ip_policy", allowed_context)
				
				-- Test with disallowed IP
				local denied_context = {client_ip = "203.0.113.1"}
				local denied, _ = auth.evaluate_policy("ip_policy", denied_context)
				
				return allowed == true and denied == false
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected IP whitelist policy to work, got %v", result)
				}
			},
		},
		{
			name: "list_policies",
			script: `
				-- Create several policies
				auth.create_security_policy("policy1", {{type = "permission", permission = "read"}})
				auth.create_security_policy("policy2", {{type = "role", role = "admin"}})
				auth.create_security_policy("policy3", {{type = "permission", permission = "write"}})
				
				local policies = auth.list_policies()
				return #policies >= 3
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected policy listing to work, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestAuditLogging tests audit and event logging
func TestAuditLogging(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	script := `
		-- Log various events
		auth.log_event("login", {user_id = "user123", ip = "192.168.1.1"})
		auth.log_event("logout", {user_id = "user123", session_duration = 3600})
		auth.log_event("token_refresh", {user_id = "user123", token_type = "oauth2"})
		auth.log_event("permission_denied", {user_id = "user456", resource = "/admin"})
		
		-- Get event history
		local all_events = auth.get_event_history()
		
		-- Get filtered events
		local login_events = auth.get_event_history({event_type = "login"}, 10)
		
		-- Test event subscription
		local events_received = {}
		local subscription = auth.subscribe_to_events({"login", "logout"}, function(event)
			table.insert(events_received, event)
		end)
		
		-- Generate more events
		auth.log_event("login", {user_id = "user789"})
		auth.log_event("admin_action", {action = "delete_user"}) -- Should not trigger
		auth.log_event("logout", {user_id = "user789"})
		
		subscription.unsubscribe()
		
		return #all_events >= 4 and
		       #login_events >= 1 and
		       #events_received >= 2 and
		       type(subscription.unsubscribe) == "function"
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Audit logging test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected audit logging to work correctly, got %v", result)
	}
}

// TestSecurityInformation tests security information retrieval
func TestSecurityInformation(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	script := `
		-- Get security level
		local security_level = auth.get_security_level()
		
		-- Get resource limits
		local limits = auth.get_resource_limits()
		
		-- Get allowed libraries
		local libraries = auth.get_allowed_libraries()
		
		-- Check denied functions
		local execute_denied = auth.is_function_denied("os.execute")
		local print_denied = auth.is_function_denied("print")
		
		-- Get system info
		local system_info = auth.get_system_info()
		
		return security_level == "standard" and
		       type(limits) == "table" and
		       limits.max_memory == 52428800 and -- 50MB
		       type(libraries) == "table" and
		       #libraries >= 4 and
		       execute_denied == true and
		       print_denied == false and
		       type(system_info) == "table" and
		       system_info.bridges_available.auth == true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Security information test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected security information retrieval to work correctly, got %v", result)
	}
}

// TestAuthUtilityFunctions tests utility functions
func TestAuthUtilityFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	script := `
		-- Generate random string
		local random1 = auth.generate_random_string(16)
		local random2 = auth.generate_random_string(32, "0123456789")
		
		-- Hash password
		local password_hash = auth.hash_password("secret123")
		
		-- Verify password
		local verify_correct = auth.verify_password("secret123", password_hash)
		local verify_wrong = auth.verify_password("wrong_password", password_hash)
		
		-- Cleanup
		auth.cleanup()
		
		return #random1 == 16 and
		       #random2 == 32 and
		       type(password_hash) == "table" and
		       password_hash.algorithm == "simple_hash" and
		       type(password_hash.salt) == "string" and
		       verify_correct == true and
		       verify_wrong == false
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Utility functions test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected utility functions to work correctly, got %v", result)
	}
}

// TestAuthErrorHandling tests error handling in auth operations
func TestAuthErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		script string
		check  func(t *testing.T, result lua.LValue)
	}{
		{
			name: "missing_required_parameters",
			script: `
				local success, err = pcall(function()
					auth.create_config(nil, {})
				end)
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing parameters, got %v", result)
				}
			},
		},
		{
			name: "invalid_parameter_types",
			script: `
				local success, err = pcall(function()
					auth.validate_session(123) -- Should be string
				end)
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid parameter types, got %v", result)
				}
			},
		},
		{
			name: "invalid_security_policy",
			script: `
				local success, err = pcall(function()
					auth.create_security_policy(123, "not_a_table") -- Invalid types
				end)
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for invalid security policy, got %v", result)
				}
			},
		},
		{
			name: "missing_bridge_graceful_handling",
			script: `
				-- Temporarily remove bridge
				local original_bridge = _G.util_auth
				_G.util_auth = nil
				
				local success, err = pcall(function()
					auth.create_config("api_key", {token = "test"})
				end)
				
				-- Restore bridge
				_G.util_auth = original_bridge
				
				return not success and type(err) == "string"
			`,
			check: func(t *testing.T, result lua.LValue) {
				if result != lua.LTrue {
					t.Errorf("Expected error for missing bridge, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(t, L)

			err := L.DoString(tt.script)
			if err != nil {
				t.Fatalf("Test script failed: %v", err)
			}

			result := L.Get(-1)
			tt.check(t, result)
		})
	}
}

// TestAuthIntegration tests integration between different auth components
func TestAuthIntegration(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	setupAuthLibrary(t, L)

	script := `
		-- Create a comprehensive authentication and security setup
		local auth_config = auth.create_config("oauth2", {
			client_id = "app123",
			client_secret = "secret456",
			token_url = "https://provider.com/token",
			access_token = "current_token",
			refresh_token = "refresh_token"
		})
		
		-- Create session with the auth config
		local session = auth.create_session(auth_config.config, "integration_session")
		
		-- Create security policy for this session
		local policy_rules = {
			{type = "permission", permission = "api:read"},
			{type = "time_based", constraint = {
				start_time = os.time() - 3600,
				end_time = os.time() + 3600
			}}
		}
		local policy = auth.create_security_policy("session_policy", policy_rules)
		
		-- Validate session and policy
		local session_valid, _ = auth.validate_session("integration_session")
		local policy_context = {permissions = {"api:read", "api:write"}}
		local policy_valid, _ = auth.evaluate_policy("session_policy", policy_context)
		
		-- Apply auth to a request
		local request = {url = "https://api.example.com/data", method = "GET"}
		local authed_request = auth.apply_to_request(request, auth_config.config)
		
		-- Log the successful authentication
		auth.log_event("api_request_authenticated", {
			session_id = "integration_session",
			endpoint = request.url,
			method = request.method
		})
		
		-- Get system overview
		local system_info = auth.get_system_info()
		
		return session_valid == true and
		       policy_valid == true and
		       authed_request.authenticated == true and
		       system_info.active_sessions >= 1 and
		       system_info.security_policies >= 1 and
		       system_info.audit_events >= 1
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected integration test to work correctly, got %v", result)
	}
}

// TestAuthPackageRequire tests that the module can be required as a package
func TestAuthPackageRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Set up mock bridges
	setupAuthLibrary(t, L)

	script := `
		-- Test that auth is available globally
		if type(auth) ~= "table" then
			error("Auth module should be available globally")
		end
		
		-- Test basic functionality
		local config = auth.create_config("api_key", {token = "require_test"})
		local session = auth.create_session(config.config, "require_test_session")
		
		local system_info = auth.get_system_info()
		auth.log_event("module_test", {test = "require_success"})
		
		return true
	`

	err := L.DoString(script)
	if err != nil {
		t.Fatalf("Package availability test failed: %v", err)
	}

	result := L.Get(-1)
	if result != lua.LTrue {
		t.Errorf("Expected package test to pass, got %v", result)
	}
}

// BenchmarkAuthOperations benchmarks key authentication operations
func BenchmarkAuthOperations(b *testing.B) {
	benchmarks := []struct {
		name   string
		script string
	}{
		{
			name: "auth_config_creation",
			script: `
				local config = auth.create_config("api_key", {token = "bench_token"})
				return config.type == "api_key"
			`,
		},
		{
			name: "session_management",
			script: `
				local auth_config = {type = "bearer", token = "bench_token"}
				local session = auth.create_session(auth_config, "bench_session")
				local valid, _ = auth.validate_session("bench_session")
				return valid
			`,
		},
		{
			name: "security_policy_evaluation",
			script: `
				local rules = {{type = "permission", permission = "read"}}
				local policy = auth.create_security_policy("bench_policy", rules)
				local context = {permissions = {"read", "write"}}
				local allowed, _ = auth.evaluate_policy("bench_policy", context)
				return allowed
			`,
		},
		{
			name: "oauth2_operations",
			script: `
				local config = auth.create_oauth2_config("client", "secret", "https://token.url")
				local claims = auth.parse_jwt_claims("jwt.token")
				return config.client_id == "client" and claims.sub == "user123"
			`,
		},
		{
			name: "audit_logging",
			script: `
				auth.log_event("benchmark_event", {operation = "test"})
				local events = auth.get_event_history({}, 10)
				return #events > 0
			`,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			L := lua.NewState()
			defer L.Close()

			setupAuthLibrary(nil, L) // Skip t.Helper in benchmark

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := L.DoString(bm.script)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
				L.Pop(1) // Clean stack
			}
		})
	}
}
