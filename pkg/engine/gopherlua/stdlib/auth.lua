-- ABOUTME: Authentication & Security Library for go-llmspell Lua standard library
-- ABOUTME: Provides unified access to authentication, authorization, security policies, and audit logging

local auth = {}

-- Import promise library for async operations (reserved for future use)
-- local promise = _G.promise or require("promise")

-- Internal state
local session_cache = {}
local auth_schemes = {}
local security_policies = {}
local audit_events = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to get auth bridge
local function get_auth_bridge()
    if not _G.util_auth then
        error("Authentication bridge not available. Ensure go-llmspell is properly initialized.")
    end
    return _G.util_auth
end

-- Helper function to get security manager (if available)
local function get_security_manager()
    -- Security manager is optional, return nil if not available
    return _G.security
end

-- Authentication Configuration

-- Create authentication configuration
function auth.create_config(auth_type, credentials, options)
    validate_required(auth_type, "authentication type")
    validate_required(credentials, "credentials")

    if type(auth_type) ~= "string" then
        error("authentication type must be a string")
    end

    if type(credentials) ~= "table" then
        error("credentials must be a table")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.createAuthConfig, auth_type, credentials)

    if not success then
        error("Failed to create auth config: " .. tostring(result))
    end

    return {
        type = auth_type,
        config = result,
        options = options or {},
        created_at = os.time(),
        apply = function(request)
            return auth.apply_to_request(request, result)
        end,
        validate = function()
            return auth.validate_config(result)
        end,
        serialize = function(encrypt_key)
            return auth.serialize_credentials(result, encrypt_key)
        end,
        refresh = function()
            if auth_type == "oauth2" and credentials.refresh_token then
                return auth.refresh_oauth2_token(result, credentials.refresh_token)
            else
                error("Token refresh not supported for auth type: " .. auth_type)
            end
        end,
    }
end

-- Create auth config from environment variables
function auth.from_env(provider, options)
    validate_required(provider, "provider name")

    if type(provider) ~= "string" then
        error("provider name must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.createAuthFromEnv, provider)

    if not success then
        error("Failed to create auth from environment: " .. tostring(result))
    end

    return auth.create_config("detected", result, options)
end

-- Create auth config from agent state
function auth.from_state(state, provider, options)
    validate_required(state, "agent state")
    validate_required(provider, "provider name")

    if type(provider) ~= "string" then
        error("provider name must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.createAuthFromState, state, provider)

    if not success then
        error("Failed to create auth from state: " .. tostring(result))
    end

    return auth.create_config("detected", result, options)
end

-- Auto-detect authentication scheme
function auth.detect_scheme(config)
    validate_required(config, "configuration")

    if type(config) ~= "table" then
        error("configuration must be a table")
    end

    local bridge = get_auth_bridge()
    local success, scheme = pcall(bridge.detectAuthScheme, config)

    if not success then
        error("Failed to detect auth scheme: " .. tostring(scheme))
    end

    return scheme
end

-- HTTP Request Authentication

-- Apply authentication to HTTP request
function auth.apply_to_request(request, auth_config)
    validate_required(request, "HTTP request")
    validate_required(auth_config, "auth configuration")

    if type(request) ~= "table" then
        error("request must be a table")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.applyAuth, request, auth_config)

    if not success then
        error("Failed to apply auth to request: " .. tostring(result))
    end

    return result
end

-- Apply authentication to headers
function auth.apply_to_headers(headers, auth_config)
    validate_required(headers, "headers")
    validate_required(auth_config, "auth configuration")

    if type(headers) ~= "table" then
        error("headers must be a table")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.applyAuthToHeaders, headers, auth_config)

    if not success then
        error("Failed to apply auth to headers: " .. tostring(result))
    end

    return result
end

-- Validate authentication configuration
function auth.validate_config(auth_config)
    validate_required(auth_config, "auth configuration")

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.validateAuthConfig, auth_config)

    if not success then
        error("Failed to validate auth config: " .. tostring(result))
    end

    return result
end

-- Parse authentication header
function auth.parse_header(header_value)
    validate_required(header_value, "header value")

    if type(header_value) ~= "string" then
        error("header value must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.parseAuthHeader, header_value)

    if not success then
        error("Failed to parse auth header: " .. tostring(result))
    end

    return result
end

-- OAuth2 Authentication

-- Create OAuth2 configuration
function auth.create_oauth2_config(client_id, client_secret, token_url, scopes)
    validate_required(client_id, "client ID")
    validate_required(client_secret, "client secret")
    validate_required(token_url, "token URL")

    if type(client_id) ~= "string" then
        error("client ID must be a string")
    end

    if type(client_secret) ~= "string" then
        error("client secret must be a string")
    end

    if type(token_url) ~= "string" then
        error("token URL must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result =
        pcall(bridge.createOAuth2Config, client_id, client_secret, token_url, scopes or {})

    if not success then
        error("Failed to create OAuth2 config: " .. tostring(result))
    end

    return result
end

-- Refresh OAuth2 token
function auth.refresh_oauth2_token(oauth2_config, refresh_token)
    validate_required(oauth2_config, "OAuth2 configuration")
    validate_required(refresh_token, "refresh token")

    if type(refresh_token) ~= "string" then
        error("refresh token must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.refreshOAuth2Token, oauth2_config, refresh_token)

    if not success then
        error("Failed to refresh OAuth2 token: " .. tostring(result))
    end

    return result
end

-- Discover OAuth2 endpoints from .well-known configuration
function auth.discover_oauth2_endpoints(issuer_url)
    validate_required(issuer_url, "issuer URL")

    if type(issuer_url) ~= "string" then
        error("issuer URL must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.discoverOAuth2Endpoints, issuer_url)

    if not success then
        error("Failed to discover OAuth2 endpoints: " .. tostring(result))
    end

    return result
end

-- Validate OAuth2 token
function auth.validate_oauth2_token(token, schema)
    validate_required(token, "OAuth2 token")

    if type(token) ~= "string" then
        error("token must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.validateOAuth2Token, token, schema)

    if not success then
        error("Failed to validate OAuth2 token: " .. tostring(result))
    end

    return result
end

-- Parse JWT token claims without verification
function auth.parse_jwt_claims(token)
    validate_required(token, "JWT token")

    if type(token) ~= "string" then
        error("token must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.parseJWTClaims, token)

    if not success then
        error("Failed to parse JWT claims: " .. tostring(result))
    end

    return result
end

-- Set up automatic token refresh
function auth.auto_refresh_token(auth_config, refresh_before)
    validate_required(auth_config, "auth configuration")

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.autoRefreshToken, auth_config, refresh_before or 300) -- 5 minutes default

    if not success then
        error("Failed to set up auto refresh: " .. tostring(result))
    end

    return result
end

-- Multi-Scheme Authentication

-- Register authentication scheme for endpoint
function auth.register_scheme(endpoint, scheme)
    validate_required(endpoint, "endpoint")
    validate_required(scheme, "auth scheme")

    if type(endpoint) ~= "string" then
        error("endpoint must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.registerAuthScheme, endpoint, scheme)

    if not success then
        error("Failed to register auth scheme: " .. tostring(result))
    end

    -- Store locally for quick access
    if not auth_schemes[endpoint] then
        auth_schemes[endpoint] = {}
    end
    table.insert(auth_schemes[endpoint], scheme)

    return result
end

-- Get authentication schemes for endpoint
function auth.get_schemes(endpoint)
    validate_required(endpoint, "endpoint")

    if type(endpoint) ~= "string" then
        error("endpoint must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.getAuthSchemes, endpoint)

    if not success then
        error("Failed to get auth schemes: " .. tostring(result))
    end

    return result
end

-- Select best authentication scheme for endpoint
function auth.select_best_scheme(endpoint, available_types)
    validate_required(endpoint, "endpoint")
    validate_required(available_types, "available types")

    if type(endpoint) ~= "string" then
        error("endpoint must be a string")
    end

    if type(available_types) ~= "table" then
        error("available types must be an array")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.selectBestAuthScheme, endpoint, available_types)

    if not success then
        error("Failed to select best auth scheme: " .. tostring(result))
    end

    return result
end

-- Session Management

-- Create authentication session
function auth.create_session(auth_config, session_id, options)
    validate_required(auth_config, "auth configuration")

    options = options or {}
    local sid = session_id or tostring(os.time()) .. "_" .. tostring(math.random(1000, 9999))

    local session = {
        id = sid,
        auth_config = auth_config,
        created_at = os.time(),
        last_used = os.time(),
        expires_at = options.expires_at or (os.time() + (options.ttl or 3600)), -- 1 hour default
        metadata = options.metadata or {},
    }

    -- Define methods after session table is created to avoid closure issues
    session.validate = function()
        if os.time() > session.expires_at then
            return false, "Session expired"
        end
        session.last_used = os.time()
        return true
    end

    session.refresh = function(extend_ttl)
        if extend_ttl then
            session.expires_at = os.time() + (extend_ttl or 3600)
        end
        session.last_used = os.time()
        return session
    end

    session.destroy = function()
        session_cache[sid] = nil
        auth.log_event("session_destroyed", {
            session_id = sid,
            duration = os.time() - session.created_at,
        })
    end

    session_cache[sid] = session

    auth.log_event("session_created", {
        session_id = sid,
        auth_type = auth_config.type,
    })

    return session
end

-- Validate session
function auth.validate_session(session_id)
    validate_required(session_id, "session ID")

    if type(session_id) ~= "string" then
        error("session ID must be a string")
    end

    local session = session_cache[session_id]
    if not session then
        return false, "Session not found"
    end

    return session.validate()
end

-- Get session
function auth.get_session(session_id)
    validate_required(session_id, "session ID")

    if type(session_id) ~= "string" then
        error("session ID must be a string")
    end

    return session_cache[session_id]
end

-- List active sessions
function auth.list_sessions()
    local sessions = {}
    for id, session in pairs(session_cache) do
        local valid, _ = session.validate()
        if valid then
            table.insert(sessions, {
                id = id,
                created_at = session.created_at,
                last_used = session.last_used,
                expires_at = session.expires_at,
            })
        else
            -- Clean up expired sessions
            session_cache[id] = nil
        end
    end
    return sessions
end

-- Credential Management

-- Serialize credentials for storage
function auth.serialize_credentials(auth_config, encrypt_key)
    validate_required(auth_config, "auth configuration")

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.serializeCredentials, auth_config, encrypt_key)

    if not success then
        error("Failed to serialize credentials: " .. tostring(result))
    end

    return result
end

-- Deserialize stored credentials
function auth.deserialize_credentials(serialized, decrypt_key)
    validate_required(serialized, "serialized credentials")

    if type(serialized) ~= "string" then
        error("serialized credentials must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.deserializeCredentials, serialized, decrypt_key)

    if not success then
        error("Failed to deserialize credentials: " .. tostring(result))
    end

    return result
end

-- Cache credentials with TTL
function auth.cache_credentials(key, auth_config, ttl)
    validate_required(key, "cache key")
    validate_required(auth_config, "auth configuration")

    if type(key) ~= "string" then
        error("cache key must be a string")
    end

    local bridge = get_auth_bridge()
    local success, result = pcall(bridge.cacheCredentials, key, auth_config, ttl or 3600)

    if not success then
        error("Failed to cache credentials: " .. tostring(result))
    end

    return result
end

-- Security Policy Management

-- Create security policy
function auth.create_security_policy(name, rules, options)
    validate_required(name, "policy name")
    validate_required(rules, "policy rules")

    if type(name) ~= "string" then
        error("policy name must be a string")
    end

    if type(rules) ~= "table" then
        error("policy rules must be a table")
    end

    options = options or {}

    local policy = {
        name = name,
        rules = rules,
        created_at = os.time(),
        enabled = options.enabled ~= false, -- Default to enabled
        metadata = options.metadata or {},
        evaluate = function(context)
            return auth.evaluate_policy(name, context)
        end,
        update = function(new_rules)
            local current_policy = security_policies[name]
            current_policy.rules = new_rules
            current_policy.updated_at = os.time()
            security_policies[name] = current_policy
            return current_policy
        end,
        disable = function()
            local current_policy = security_policies[name]
            current_policy.enabled = false
            return current_policy
        end,
        enable = function()
            local current_policy = security_policies[name]
            current_policy.enabled = true
            return current_policy
        end,
    }

    security_policies[name] = policy

    auth.log_event("security_policy_created", {
        policy_name = name,
        rules_count = #rules,
    })

    return policy
end

-- Evaluate security policy
function auth.evaluate_policy(policy_name, context)
    validate_required(policy_name, "policy name")
    validate_required(context, "evaluation context")

    if type(policy_name) ~= "string" then
        error("policy name must be a string")
    end

    if type(context) ~= "table" then
        error("context must be a table")
    end

    local policy = security_policies[policy_name]
    if not policy then
        return false, "Policy not found: " .. policy_name
    end

    if not policy.enabled then
        return true, "Policy disabled"
    end

    -- Group rules by type for proper evaluation logic
    local rule_groups = {}
    for _, rule in ipairs(policy.rules) do
        local rule_type = rule.type or "custom"
        if not rule_groups[rule_type] then
            rule_groups[rule_type] = {}
        end
        table.insert(rule_groups[rule_type], rule)
    end

    -- Evaluate rule groups (OR logic within groups, AND logic between groups)
    for rule_type, rules in pairs(rule_groups) do
        local group_passed = false

        -- For role/permission rules, use OR logic (any match passes)
        if rule_type == "role" or rule_type == "permission" then
            for _, rule in ipairs(rules) do
                if auth.evaluate_rule(rule, context) then
                    group_passed = true
                    break
                end
            end
        else
            -- For other rule types, use AND logic (all must pass)
            group_passed = true
            for _, rule in ipairs(rules) do
                if not auth.evaluate_rule(rule, context) then
                    group_passed = false
                    break
                end
            end
        end

        if not group_passed then
            auth.log_event("security_policy_violation", {
                policy_name = policy_name,
                rule_type = rule_type,
                context = context,
            })
            return false, "Policy violation: " .. rule_type .. " requirements not met"
        end
    end

    return true, "Policy satisfied"
end

-- Evaluate a single security rule
function auth.evaluate_rule(rule, context)
    if type(rule) ~= "table" then
        return false
    end

    -- Support different rule types
    if rule.type == "permission" then
        return auth.check_permission(rule.permission, context)
    elseif rule.type == "role" then
        return auth.check_role(rule.role, context)
    elseif rule.type == "custom" then
        if type(rule.evaluator) == "function" then
            return rule.evaluator(context)
        end
    elseif rule.type == "time_based" then
        return auth.check_time_constraint(rule.constraint, context)
    elseif rule.type == "ip_whitelist" then
        return auth.check_ip_whitelist(rule.allowed_ips, context)
    end

    return false
end

-- Permission checking
function auth.check_permission(permission, context)
    if type(permission) ~= "string" then
        return false
    end

    local user_permissions = context.permissions or {}

    -- Simple string match or pattern match
    for _, user_perm in ipairs(user_permissions) do
        if user_perm == permission or string.match(permission, user_perm) then
            return true
        end
    end

    return false
end

-- Role checking
function auth.check_role(required_role, context)
    if type(required_role) ~= "string" then
        return false
    end

    local user_roles = context.roles or {}

    for _, user_role in ipairs(user_roles) do
        if user_role == required_role then
            return true
        end
    end

    return false
end

-- Time-based constraint checking
function auth.check_time_constraint(constraint, context) -- luacheck: ignore 212
    if type(constraint) ~= "table" then
        return false
    end

    local current_time = os.time()

    if constraint.start_time and current_time < constraint.start_time then
        return false
    end

    if constraint.end_time and current_time > constraint.end_time then
        return false
    end

    if constraint.allowed_hours then
        local hour = tonumber(os.date("%H", current_time))
        local allowed = false
        for _, allowed_hour in ipairs(constraint.allowed_hours) do
            if hour == allowed_hour then
                allowed = true
                break
            end
        end
        if not allowed then
            return false
        end
    end

    return true
end

-- IP whitelist checking
function auth.check_ip_whitelist(allowed_ips, context)
    if type(allowed_ips) ~= "table" then
        return false
    end

    local client_ip = context.client_ip or context.ip
    if not client_ip then
        return false
    end

    for _, allowed_ip in ipairs(allowed_ips) do
        if client_ip == allowed_ip or string.match(client_ip, allowed_ip) then
            return true
        end
    end

    return false
end

-- List security policies
function auth.list_policies()
    local policies = {}
    for name, policy in pairs(security_policies) do
        table.insert(policies, {
            name = name,
            enabled = policy.enabled,
            created_at = policy.created_at,
            updated_at = policy.updated_at,
            rules_count = #policy.rules,
        })
    end
    return policies
end

-- Audit and Event Logging

-- Log authentication event
function auth.log_event(event_type, metadata)
    validate_required(event_type, "event type")
    validate_required(metadata, "event metadata")

    if type(event_type) ~= "string" then
        error("event type must be a string")
    end

    if type(metadata) ~= "table" then
        error("metadata must be a table")
    end

    local event = {
        type = event_type,
        timestamp = os.time(),
        metadata = metadata,
        id = tostring(os.time()) .. "_" .. tostring(math.random(10000, 99999)),
    }

    -- Store locally
    table.insert(audit_events, event)

    -- Also log through bridge if available
    local bridge = get_auth_bridge()
    if bridge and bridge.logAuthEvent then
        pcall(bridge.logAuthEvent, event_type, metadata)
    end

    return event
end

-- Get event history
function auth.get_event_history(filter, limit)
    filter = filter or {}
    limit = limit or 100

    local filtered_events = {}
    local count = 0

    -- Filter events in reverse chronological order
    for i = #audit_events, 1, -1 do
        if count >= limit then
            break
        end

        local event = audit_events[i]
        local include = true

        -- Apply filters
        if filter.event_type and event.type ~= filter.event_type then
            include = false
        end

        if filter.since and event.timestamp < filter.since then
            include = false
        end

        if filter.until_time and event.timestamp > filter.until_time then
            include = false
        end

        if include then
            table.insert(filtered_events, event)
            count = count + 1
        end
    end

    return filtered_events
end

-- Subscribe to authentication events
function auth.subscribe_to_events(event_types, handler)
    validate_required(event_types, "event types")
    validate_required(handler, "event handler")

    if type(event_types) ~= "table" then
        error("event types must be an array")
    end

    if type(handler) ~= "function" then
        error("handler must be a function")
    end

    -- Create subscription wrapper
    local subscription = {
        event_types = event_types,
        handler = handler,
        active = true,
    }

    -- Define unsubscribe method after table creation to avoid closure issues
    subscription.unsubscribe = function()
        subscription.active = false
    end

    -- Hook into event logging (simplified implementation)
    local original_log_event = auth.log_event
    auth.log_event = function(event_type, metadata)
        local event = original_log_event(event_type, metadata)

        -- Notify subscriber if interested in this event type
        if subscription.active then
            for _, interested_type in ipairs(event_types) do
                if event_type == interested_type then
                    local success, err = pcall(handler, event)
                    if not success then
                        auth.log_event("subscription_error", {
                            error = tostring(err),
                            event_type = event_type,
                        })
                    end
                    break
                end
            end
        end

        return event
    end

    return subscription
end

-- Security Information

-- Get current security level
function auth.get_security_level()
    local security_mgr = get_security_manager()
    if security_mgr then
        local success, level = pcall(security_mgr.getSecurityLevel)
        if success then
            return level
        end
    end
    return "unknown"
end

-- Get resource limits
function auth.get_resource_limits()
    local security_mgr = get_security_manager()
    if security_mgr then
        local success, limits = pcall(security_mgr.getResourceLimits)
        if success then
            return limits
        end
    end
    return {
        max_memory = "unknown",
        max_duration = "unknown",
        max_instructions = "unknown",
    }
end

-- Get allowed libraries
function auth.get_allowed_libraries()
    local security_mgr = get_security_manager()
    if security_mgr then
        local success, libraries = pcall(security_mgr.getAllowedLibraries)
        if success then
            return libraries
        end
    end
    return {}
end

-- Check if function is denied
function auth.is_function_denied(func_name)
    validate_required(func_name, "function name")

    if type(func_name) ~= "string" then
        error("function name must be a string")
    end

    local security_mgr = get_security_manager()
    if security_mgr then
        local success, denied = pcall(security_mgr.isFunctionDenied, func_name)
        if success then
            return denied
        end
    end
    return false
end

-- Utility Functions

-- Generate secure random string
function auth.generate_random_string(length, charset)
    length = length or 32
    charset = charset or "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    local result = {}
    for i = 1, length do
        local rand_index = math.random(1, #charset)
        result[i] = string.sub(charset, rand_index, rand_index)
    end

    return table.concat(result)
end

-- Hash password (basic implementation)
function auth.hash_password(password, salt)
    validate_required(password, "password")

    if type(password) ~= "string" then
        error("password must be a string")
    end

    -- Simple hash implementation (would use proper crypto in production)
    salt = salt or auth.generate_random_string(16)
    local combined = password .. salt

    -- Basic hash function (in production, use proper crypto libraries)
    local hash = 0
    for i = 1, #combined do
        hash = (hash * 31 + string.byte(combined, i)) % 2147483647
    end

    return {
        hash = tostring(hash),
        salt = salt,
        algorithm = "simple_hash",
    }
end

-- Verify password hash
function auth.verify_password(password, hash_data)
    validate_required(password, "password")
    validate_required(hash_data, "hash data")

    if type(password) ~= "string" then
        error("password must be a string")
    end

    if type(hash_data) ~= "table" then
        error("hash data must be a table")
    end

    local computed = auth.hash_password(password, hash_data.salt)
    return computed.hash == hash_data.hash
end

-- Get system information
function auth.get_system_info()
    return {
        lua_version = _VERSION,
        security_level = auth.get_security_level(),
        resource_limits = auth.get_resource_limits(),
        allowed_libraries = auth.get_allowed_libraries(),
        bridges_available = {
            auth = _G.util_auth ~= nil,
            security = _G.security ~= nil,
        },
        active_sessions = #auth.list_sessions(),
        security_policies = #auth.list_policies(),
        audit_events = #audit_events,
    }
end

-- Clean up resources
function auth.cleanup()
    -- Clean up expired sessions
    for id, session in pairs(session_cache) do
        local valid, _ = session.validate()
        if not valid then
            session_cache[id] = nil
        end
    end

    -- Trim audit events (keep last 1000)
    if #audit_events > 1000 then
        local trimmed = {}
        for i = #audit_events - 999, #audit_events do
            table.insert(trimmed, audit_events[i])
        end
        audit_events = trimmed
    end

    auth.log_event("auth_cleanup_completed", {
        active_sessions = #auth.list_sessions(),
        audit_events_count = #audit_events,
    })
end

-- Export the module
return auth
