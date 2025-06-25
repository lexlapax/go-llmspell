-- ABOUTME: Utility module wrapping multi-bridge UtilsAdapter for auth, debug, errors, json, llm utils, logging, slog, and general utilities
-- ABOUTME: Provides comprehensive utility functions through 8 specialized bridges with graceful degradation

local utils = {}

-- ============================================================================
-- Bridge Accessors
-- ============================================================================

-- Get auth bridge for authentication operations
local function get_auth_bridge()
    if bridges and bridges.util_auth then
        return bridges.util_auth
    end
    return nil
end

-- Get debug bridge for debugging operations
local function get_debug_bridge()
    if bridges and bridges.util_debug then
        return bridges.util_debug
    end
    return nil
end

-- Get errors bridge for error handling
local function get_errors_bridge()
    if bridges and bridges.util_errors then
        return bridges.util_errors
    end
    return nil
end

-- Get json bridge for JSON operations
local function get_json_bridge()
    if bridges and bridges.util_json then
        return bridges.util_json
    end
    return nil
end

-- Get llm bridge for LLM utilities
local function get_llm_bridge()
    if bridges and bridges.util_llm then
        return bridges.util_llm
    end
    return nil
end

-- Get logger bridge for logging operations
local function get_logger_bridge()
    if bridges and bridges.util_script_logger then
        return bridges.util_script_logger
    end
    return nil
end

-- Get slog bridge for structured logging
local function get_slog_bridge()
    if bridges and bridges.util_slog then
        return bridges.util_slog
    end
    return nil
end

-- Get util bridge for general utilities
local function get_util_bridge()
    if bridges and bridges.util_core then
        return bridges.util_core
    end
    return nil
end

-- ============================================================================
-- Authentication Methods
-- ============================================================================

function utils.auth_authenticate(credentials)
    local auth_bridge = get_auth_bridge()
    if not auth_bridge then
        error("Auth bridge not available for authenticate")
    end
    return auth_bridge.authenticate(credentials)
end

function utils.auth_validate_token(token)
    local auth_bridge = get_auth_bridge()
    if not auth_bridge then
        error("Auth bridge not available for validate_token")
    end
    return auth_bridge.validateToken(token)
end

function utils.auth_refresh_token(token)
    local auth_bridge = get_auth_bridge()
    if not auth_bridge then
        error("Auth bridge not available for refresh_token")
    end
    return auth_bridge.refreshToken(token)
end

function utils.auth_generate_token(claims)
    local auth_bridge = get_auth_bridge()
    if not auth_bridge then
        error("Auth bridge not available for generate_token")
    end
    return auth_bridge.generateToken(claims)
end

function utils.auth_hash_password(password, algorithm)
    local auth_bridge = get_auth_bridge()
    if not auth_bridge then
        error("Auth bridge not available for hash_password")
    end
    return auth_bridge.hashPassword(password, algorithm)
end

function utils.auth_verify_password(password, hash)
    local auth_bridge = get_auth_bridge()
    if not auth_bridge then
        error("Auth bridge not available for verify_password")
    end
    return auth_bridge.verifyPassword(password, hash)
end

-- ============================================================================
-- Debug Methods
-- ============================================================================

function utils.debug_set_level(level)
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for set_level")
    end
    return debug_bridge.setLevel(level)
end

function utils.debug_log(level, message, data)
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for log")
    end
    return debug_bridge.log(level, message, data)
end

function utils.debug_get_config()
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for get_config")
    end
    return debug_bridge.getConfig()
end

function utils.debug_trace(operation)
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for trace")
    end
    return debug_bridge.trace(operation)
end

function utils.debug_profile(name, fn)
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for profile")
    end
    return debug_bridge.profile(name, fn)
end

function utils.debug_dump(variable, name)
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for dump")
    end
    return debug_bridge.dump(variable, name)
end

function utils.debug_assert(condition, message)
    local debug_bridge = get_debug_bridge()
    if not debug_bridge then
        error("Debug bridge not available for assert")
    end
    return debug_bridge.assert(condition, message)
end

-- ============================================================================
-- Error Handling Methods
-- ============================================================================

function utils.errors_create_error(message, category, metadata)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for create_error")
    end
    return errors_bridge.createError(message, category, metadata)
end

function utils.errors_wrap_error(err, message)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for wrap_error")
    end
    return errors_bridge.wrapError(err, message)
end

function utils.errors_aggregate_errors(errors)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for aggregate_errors")
    end
    return errors_bridge.aggregateErrors(errors)
end

function utils.errors_categorize_error(err)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for categorize_error")
    end
    return errors_bridge.categorizeError(err)
end

function utils.errors_wrap(err, message)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for wrap")
    end
    return errors_bridge.wrap(err, message)
end

function utils.errors_unwrap(err)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for unwrap")
    end
    return errors_bridge.unwrap(err)
end

function utils.errors_is_type(err, error_type)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for is_type")
    end
    return errors_bridge.isType(err, error_type)
end

function utils.errors_get_stack(err)
    local errors_bridge = get_errors_bridge()
    if not errors_bridge then
        error("Errors bridge not available for get_stack")
    end
    return errors_bridge.getStack(err)
end

-- ============================================================================
-- JSON Methods
-- ============================================================================

function utils.json_parse(json_string)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for parse")
    end
    return json_bridge.parse(json_string)
end

function utils.json_to_json(value, pretty)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for to_json")
    end
    return json_bridge.toJSON(value, pretty)
end

function utils.json_validate_json_schema(json_data, schema)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for validate_json_schema")
    end
    return json_bridge.validateJSONSchema(json_data, schema)
end

function utils.json_extract_structured_data(text, schema)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for extract_structured_data")
    end
    return json_bridge.extractStructuredData(text, schema)
end

function utils.json_encode(value)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for encode")
    end
    return json_bridge.encode(value)
end

function utils.json_decode(json_string)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for decode")
    end
    return json_bridge.decode(json_string)
end

function utils.json_validate(json_string)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for validate")
    end
    return json_bridge.validate(json_string)
end

function utils.json_format(json_string, indent)
    local json_bridge = get_json_bridge()
    if not json_bridge then
        error("JSON bridge not available for format")
    end
    return json_bridge.format(json_string, indent)
end

-- ============================================================================
-- LLM Utility Methods
-- ============================================================================

function utils.llm_create_provider(name, config)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for create_provider")
    end
    return llm_bridge.createProvider(name, config)
end

function utils.llm_generate_typed(prompt, schema, options)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for generate_typed")
    end
    return llm_bridge.generateTyped(prompt, schema, options)
end

function utils.llm_track_cost(model, tokens, cost)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for track_cost")
    end
    return llm_bridge.trackCost(model, tokens, cost)
end

function utils.llm_parse_response(response, format)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for parse_response")
    end
    return llm_bridge.parseResponse(response, format)
end

function utils.llm_format_prompt(template, variables)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for format_prompt")
    end
    return llm_bridge.formatPrompt(template, variables)
end

function utils.llm_count_tokens(text, model)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for count_tokens")
    end
    return llm_bridge.countTokens(text, model)
end

function utils.llm_split_message(text, max_tokens, model)
    local llm_bridge = get_llm_bridge()
    if not llm_bridge then
        error("LLM bridge not available for split_message")
    end
    return llm_bridge.splitMessage(text, max_tokens, model)
end

-- ============================================================================
-- Logger Methods
-- ============================================================================

function utils.logger_create_logger(name, config)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for create_logger")
    end
    return logger_bridge.createLogger(name, config)
end

function utils.logger_log(level, message, data)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for log")
    end
    return logger_bridge.log(level, message, data)
end

function utils.logger_set_log_level(level)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for set_log_level")
    end
    return logger_bridge.setLogLevel(level)
end

function utils.logger_error(message, data)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for error")
    end
    return logger_bridge.error(message, data)
end

function utils.logger_warn(message, data)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for warn")
    end
    return logger_bridge.warn(message, data)
end

function utils.logger_info(message, data)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for info")
    end
    return logger_bridge.info(message, data)
end

function utils.logger_debug(message, data)
    local logger_bridge = get_logger_bridge()
    if not logger_bridge then
        error("Logger bridge not available for debug")
    end
    return logger_bridge.debug(message, data)
end

-- ============================================================================
-- Structured Logging (slog) Methods
-- ============================================================================

function utils.slog_info(message, fields)
    local slog_bridge = get_slog_bridge()
    if not slog_bridge then
        error("Slog bridge not available for info")
    end
    return slog_bridge.info(message, fields)
end

function utils.slog_warn(message, fields)
    local slog_bridge = get_slog_bridge()
    if not slog_bridge then
        error("Slog bridge not available for warn")
    end
    return slog_bridge.warn(message, fields)
end

function utils.slog_error(message, fields)
    local slog_bridge = get_slog_bridge()
    if not slog_bridge then
        error("Slog bridge not available for error")
    end
    return slog_bridge.error(message, fields)
end

function utils.slog_debug(message, fields)
    local slog_bridge = get_slog_bridge()
    if not slog_bridge then
        error("Slog bridge not available for debug")
    end
    return slog_bridge.debug(message, fields)
end

function utils.slog_with_fields(fields)
    local slog_bridge = get_slog_bridge()
    if not slog_bridge then
        error("Slog bridge not available for with_fields")
    end
    return slog_bridge.withFields(fields)
end

-- ============================================================================
-- General Utility Methods
-- ============================================================================

function utils.general_generate_uuid()
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for generate_uuid")
    end
    return util_bridge.generateUUID()
end

function utils.general_hash(data, algorithm)
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for hash")
    end
    return util_bridge.hash(data, algorithm)
end

function utils.general_retry(fn, options)
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for retry")
    end
    return util_bridge.retry(fn, options)
end

function utils.general_sleep(milliseconds)
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for sleep")
    end
    return util_bridge.sleep(milliseconds)
end

function utils.general_uuid()
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for uuid")
    end
    return util_bridge.uuid()
end

function utils.general_encode(data, encoding)
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for encode")
    end
    return util_bridge.encode(data, encoding)
end

function utils.general_decode(data, encoding)
    local util_bridge = get_util_bridge()
    if not util_bridge then
        error("Util bridge not available for decode")
    end
    return util_bridge.decode(data, encoding)
end

-- ============================================================================
-- Constants
-- ============================================================================

-- Log levels
utils.LOG_LEVELS = {
    TRACE = "trace",
    DEBUG = "debug",
    INFO = "info",
    WARN = "warn",
    ERROR = "error",
    FATAL = "fatal"
}

-- Authentication schemes
utils.AUTH_SCHEMES = {
    BASIC = "basic",
    BEARER = "bearer",
    DIGEST = "digest",
    OAUTH2 = "oauth2",
    API_KEY = "apikey"
}

-- Hash algorithms
utils.HASH_ALGORITHMS = {
    MD5 = "md5",
    SHA1 = "sha1",
    SHA256 = "sha256",
    SHA512 = "sha512",
    BCRYPT = "bcrypt",
    ARGON2 = "argon2"
}

-- Error categories
utils.ERROR_CATEGORIES = {
    VALIDATION = "validation",
    AUTHENTICATION = "authentication",
    AUTHORIZATION = "authorization",
    NOT_FOUND = "not_found",
    CONFLICT = "conflict",
    RATE_LIMIT = "rate_limit",
    INTERNAL = "internal",
    EXTERNAL = "external",
    TIMEOUT = "timeout",
    NETWORK = "network"
}

-- Export module
return utils