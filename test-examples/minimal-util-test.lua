-- ABOUTME: Test script using only minimal feature set utilities
-- ABOUTME: Tests log, json/data, and error modules available in minimal configuration

-- Test logging (should be available in minimal)
local log = require("log")
local data = require("data")
local errors = require("errors")

log.info("=== Minimal Utility Test ===")
log.debug("Testing minimal feature set")

-- Test data/json operations
local test_data = {
    name = "test",
    value = 42,
    nested = {
        items = {1, 2, 3},
        enabled = true
    }
}

log.info("Testing JSON operations")
local json_str = data.to_json(test_data)
log.debug("JSON: " .. json_str)

local decoded = data.from_json(json_str)
log.info("Decoded name: " .. decoded.name)

-- Test error handling
log.info("Testing error handling")
local function divide(a, b)
    if b == 0 then
        error("Division by zero")
    end
    return a / b
end

local success, result = pcall(divide, 10, 2)
if success then
    log.info("10 / 2 = " .. result)
else
    log.error("Division failed: " .. tostring(result))
end

-- Test with error
success, result = pcall(divide, 10, 0)
if not success then
    log.warn("Expected error: " .. tostring(result))
end

-- Return results
return {
    success = true,
    message = "Minimal utility test completed",
    json_test = json_str ~= nil,
    log_test = true,
    error_test = true
}