-- Test with xpcall to get full stack trace
local function trace_handler(err)
    return debug.traceback(err, 2)
end

print("Loading agent module with trace...")
local ok, result = xpcall(function()
    return require("agent")
end, trace_handler)

if not ok then
    print("Failed to load agent module:")
    print(result)
    return "failed"
end

local agent = result
print("Agent module loaded successfully")

-- Now try to create
print("\nCreating agent with trace...")
local ok2, result2 = xpcall(function()
    return agent.create("Test", {model = "gpt-3.5-turbo"})
end, trace_handler)

if not ok2 then
    print("Failed to create agent:")
    print(result2)
    return "failed"
end

print("Agent created:", result2)
return "done"