-- Test basic Lua functionality
local function test()
    return "Hello from test function"
end

-- This should work
local result = test()

-- Try using io.write if available  
if io then
    io.write("Result: " .. result .. "\n")
else
    -- Fall back to return
    return result
end