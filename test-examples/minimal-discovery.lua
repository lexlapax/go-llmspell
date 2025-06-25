-- ABOUTME: Discover what's available in minimal feature set
-- ABOUTME: Lists all available modules and functions

print("=== Minimal Feature Discovery ===")

-- Check what's in the global environment
print("\nGlobal modules:")
for k, v in pairs(_G) do
    if type(v) == "table" and k ~= "_G" and k ~= "package" then
        print("  " .. k .. " (" .. type(v) .. ")")
    end
end

-- Try to see what modules can be required
local modules_to_try = {
    "log", "data", "errors", "debug", "core", 
    "llm", "agent", "state", "tools", "json"
}

print("\nTrying to require modules:")
for _, mod in ipairs(modules_to_try) do
    local success, result = pcall(require, mod)
    if success then
        print("  ✓ " .. mod .. " - available")
        -- Check if it has functions
        if type(result) == "table" then
            local count = 0
            for k, v in pairs(result) do
                if type(v) == "function" then
                    count = count + 1
                end
            end
            if count > 0 then
                print("    (" .. count .. " functions)")
            end
        end
    else
        print("  ✗ " .. mod .. " - not available")
    end
end

-- Check if params is available
print("\nChecking params:")
if params then
    print("  params is available")
    for k, v in pairs(params) do
        print("    " .. k .. " = " .. tostring(v))
    end
else
    print("  params is not available")
end

-- Collect results
local results = {
    globals = {},
    modules = {},
    params_available = params ~= nil
}

-- Collect globals
for k, v in pairs(_G) do
    if type(v) == "table" and k ~= "_G" and k ~= "package" then
        results.globals[k] = type(v)
    end
end

-- Collect module availability
for _, mod in ipairs(modules_to_try) do
    local success, result = pcall(require, mod)
    results.modules[mod] = success
end

return {
    success = true,
    message = "Discovery completed",
    results = results
}