-- ABOUTME: Test what methods are available on agent objects
-- ABOUTME: Verify agent:run() works instead of agent:generate()

local agent = require("agent")

-- Check if we can create an agent
local test_agent = nil
local create_error = nil

local ok, result = pcall(function()
    return agent.create("Test Agent", {
        model = "gpt-3.5-turbo",
        system = "You are a test agent"
    })
end)

if ok then
    test_agent = result
else
    create_error = result
end

-- Check what's on the agent object
local methods = {}
if test_agent and type(test_agent) == "table" then
    -- Check direct properties
    for k, v in pairs(test_agent) do
        if type(v) == "function" then
            table.insert(methods, k .. " (direct function)")
        else
            table.insert(methods, k .. " (" .. type(v) .. ")")
        end
    end
    
    -- Check metatable methods
    local mt = getmetatable(test_agent)
    if mt and mt.__index then
        for k, v in pairs(mt.__index) do
            if type(v) == "function" then
                table.insert(methods, k .. " (metatable function)")
            end
        end
    end
end

return {
    agent_created = test_agent ~= nil,
    create_error = create_error and tostring(create_error) or nil,
    agent_type = type(test_agent),
    agent_id = test_agent and test_agent.id or nil,
    methods = methods,
    has_generate = test_agent and type(test_agent.generate) == "function",
    has_run = test_agent and type(test_agent.run) == "function",
    has_run_async = test_agent and type(test_agent.run_async) == "function"
}