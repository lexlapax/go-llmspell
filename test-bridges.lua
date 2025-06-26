-- Test bridges availability
local result = {}
result.bridges_found = bridges ~= nil

if bridges then
    result.bridge_types = {}
    for k, v in pairs(bridges) do
        result.bridge_types[k] = type(v)
    end
end

-- Try requiring agent
local ok, agent = pcall(require, "agent")
result.agent_loaded = ok

if ok then
    result.agent_type = type(agent)
    if type(agent) == "table" then
        result.agent_functions = {}
        for k, v in pairs(agent) do
            result.agent_functions[k] = type(v)
        end
    end
else
    result.agent_error = tostring(agent)
end

return result