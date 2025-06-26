-- ABOUTME: Test util_core bridge directly
-- ABOUTME: Check if sleep method exists and is callable

print("Testing util_core bridge...")

if not bridges then
    print("ERROR: bridges not available")
    return {error = "no bridges"}
end

print("bridges exists")

if not bridges.util_core then
    print("ERROR: util_core bridge not available")
    return {error = "no util_core"}
end

print("util_core bridge exists")

-- Check type
print("util_core type:", type(bridges.util_core))

-- Check if sleep exists
print("Checking sleep method...")
if bridges.util_core.sleep then
    print("sleep exists, type:", type(bridges.util_core.sleep))
    
    -- Try calling it
    print("Attempting to call sleep...")
    local ok, err = pcall(function()
        return bridges.util_core.sleep(100)
    end)
    
    print("Sleep call result:", ok, err)
else
    print("ERROR: sleep method does not exist on util_core")
    
    -- List what methods do exist
    print("\nAvailable methods on util_core:")
    for k, v in pairs(bridges.util_core) do
        print("  -", k, ":", type(v))
    end
end

return {
    bridge_exists = bridges.util_core ~= nil,
    sleep_exists = bridges.util_core.sleep ~= nil
}