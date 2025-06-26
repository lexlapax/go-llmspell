-- ABOUTME: Debug test to check if print works
-- ABOUTME: Most basic test possible

print("Hello from Lua!")
print("If you see this, Lua is working")

-- Try basic Lua operations
local x = 1 + 1
print("1 + 1 =", x)

-- Check if we can see errors
print("About to test error handling...")
local ok, err = pcall(function()
    error("This is a test error")
end)
print("Error caught:", not ok)
print("Error message:", err)

print("Test complete!")