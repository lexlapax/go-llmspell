-- ABOUTME: Minimal test to verify state module works
-- ABOUTME: Tests basic state.get() and state.set() operations

print("=== Minimal State Test ===")

-- Try to require state module
local ok, state = pcall(require, "state")
if not ok then
    print("ERROR: Failed to load state module:", state)
    os.exit(1)
end

print("✓ State module loaded successfully")

-- Test 1: Simple set and get
print("\n--- Test 1: Simple set/get ---")
state.set("test_key", "test_value")
local value = state.get("test_key")
print("Set 'test_key' to 'test_value'")
print("Got back:", value)

if value == "test_value" then
    print("✓ Simple set/get works!")
else
    print("✗ Simple set/get failed!")
end

-- Test 2: Set complex object
print("\n--- Test 2: Complex object ---")
state.set("user", {
    name = "Test User",
    age = 25,
    preferences = {
        color = "blue",
        theme = "dark"
    }
})

local user = state.get("user")
if user then
    print("Got user object:")
    print("  name:", user.name)
    print("  age:", user.age)
    if user.preferences then
        print("  preferences.color:", user.preferences.color)
        print("  preferences.theme:", user.preferences.theme)
    end
    print("✓ Complex object works!")
else
    print("✗ Failed to get complex object")
end

-- Test 3: Nested path
print("\n--- Test 3: Nested path access ---")
local color = state.get("user.preferences.color")
print("Got user.preferences.color:", color)

if color == "blue" then
    print("✓ Nested path access works!")
else
    print("✗ Nested path access failed!")
end

print("\n=== State Module Basic Tests Complete ===")