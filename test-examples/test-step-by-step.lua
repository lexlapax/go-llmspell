-- Test step by step
print("Step 1: Load agent module")
local agent = require("agent")

print("Step 2: Check agent.create")
print("  Type:", type(agent.create))

print("Step 3: Prepare parameters")
local name = "Test"
local config = {model = "gpt-3.5-turbo"}
print("  name:", name)
print("  config:", config)

print("Step 4: Call agent.create directly")
-- Add a wrapper to catch the exact error
local function create_wrapper()
    return agent.create(name, config)
end

local ok, err = pcall(create_wrapper)
print("  Result:", ok)
if not ok then
    print("  Error:", err)
    
    -- Try to extract more info
    local err_str = tostring(err)
    print("  Error string:", err_str)
    
    -- Check if it's the validate_required error
    if string.find(err_str, "validate_required") then
        print("  ERROR: validate_required issue")
    elseif string.find(err_str, "attempt to call") then
        print("  ERROR: Trying to call non-function")
        -- The error says line 62, let's see what that could be
    end
end

return "done"