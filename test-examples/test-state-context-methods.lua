-- Test state_context bridge methods in detail
print("Testing state_context bridge methods...")

local bridge = bridges.state_context

-- Test createSharedContext
print("\n1. Testing createSharedContext...")
local success, context = pcall(function()
    return bridge.createSharedContext("test_context")
end)
print("createSharedContext result:", success)
if success then
    print("Context type:", type(context))
    print("Context:", context)
end

-- Test set method
print("\n2. Testing set method...")
if success and context then
    local set_success, set_result = pcall(function()
        return bridge.set(context, "test_key", "test_value")
    end)
    print("set result:", set_success, set_result)
    
    -- Test get method
    print("\n3. Testing get method...")
    local get_success, get_result = pcall(function()
        return bridge.get(context, "test_key")
    end)
    print("get result:", get_success, get_result)
end

return "done"