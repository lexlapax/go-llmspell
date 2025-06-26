-- ABOUTME: Detailed test of agent:run() to find exact issue
-- ABOUTME: Step by step debugging of agent creation and execution

local agent = require("agent")

print("=== Agent Run Detailed Test ===")

-- Create agent
print("\n1. Creating agent...")
local test_agent = agent.create("Assistant", {
    model = "gpt-3.5-turbo",
    system = "You are a helpful assistant."
})

if not test_agent then
    print("ERROR: Failed to create agent")
    return {error = "agent creation failed"}
end

print("Agent created with ID:", test_agent.id)

-- Check agent metatable
print("\n2. Checking agent metatable...")
local mt = getmetatable(test_agent)
if mt then
    print("Metatable exists")
    if mt.__index then
        print("__index exists, type:", type(mt.__index))
        if type(mt.__index) == "table" then
            print("Methods in __index:")
            for k, v in pairs(mt.__index) do
                print("  -", k, ":", type(v))
            end
        end
    end
else
    print("No metatable found")
end

-- Check if run is accessible
print("\n3. Checking run method accessibility...")
print("test_agent.run type:", type(test_agent.run))
print("test_agent:run type:", type(test_agent.run))

-- Try calling run directly
print("\n4. Attempting to call run...")
local ok, err = pcall(function()
    -- Try the direct method call first
    local bridge = bridges.agent_core
    if not bridge then
        error("agent_core bridge not found")
    end
    
    print("Using bridge.run directly...")
    local result = bridge.run(test_agent.id, "Hello", {})
    return result
end)

if ok then
    print("Direct bridge call succeeded")
else
    print("Direct bridge call failed:", err)
end

-- Now try through the wrapper
print("\n5. Trying through agent wrapper...")
ok, err = pcall(function()
    return test_agent:run("Hello")
end)

if ok then
    print("Wrapper call succeeded")
else
    print("Wrapper call failed:", err)
end

return {
    agent_created = true,
    has_metatable = mt ~= nil,
    has_run_method = test_agent.run ~= nil,
    direct_call_ok = ok
}