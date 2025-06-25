-- ABOUTME: Minimal test script with no module dependencies
-- ABOUTME: Tests basic Lua execution without any go-llmspell features

-- Simple test that should work with any feature set
print("=== Minimal Test ===")
print("This test has no dependencies")

-- Basic calculations
local a = 10
local b = 20
local sum = a + b
print("10 + 20 = " .. sum)

-- Simple table operations
local data = {
    name = "test",
    value = 42,
    items = {1, 2, 3}
}

print("Data name: " .. data.name)
print("Data value: " .. data.value)
print("Items count: " .. #data.items)

-- Return a result
return {
    success = true,
    message = "Minimal test completed",
    result = sum
}