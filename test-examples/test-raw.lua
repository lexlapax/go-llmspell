-- Raw test without modules
print("Raw test starting...")

-- Define what we need inline
local function validate_required(param, name)
    if not param or param == "" then
        error(name .. " is required")
    end
end

local function create_agent(name, config)
    print("In create_agent")
    validate_required(name, "agent name")
    
    local bridge = bridges.agent_core
    if not bridge then
        error("No agent bridge")
    end
    
    local agent_id = "agent_" .. name .. "_" .. tostring(os.time())
    config = config or {}
    config.name = name
    
    print("Calling bridge.createAgent...")
    local agent_obj = bridge.createAgent(agent_id, config)
    
    return agent_obj
end

-- Test it
print("Creating agent...")
local agent = create_agent("TestAgent", {model = "gpt-3.5-turbo"})
print("Agent created:", agent)

return "done"