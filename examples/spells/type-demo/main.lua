-- ABOUTME: Example spell demonstrating the ScriptValue type system
-- ABOUTME: Shows how different types are handled and converted between Lua and Go

-- This example demonstrates how the ScriptValue type system works
-- All values passed between Lua and Go bridges use the ScriptValue interface

print("=== ScriptValue Type System Demo ===\n")

-- 1. Basic Types
print("1. Basic Types:")
print("   Nil:", nil) -- ScriptValue: NilValue
print("   Boolean:", true) -- ScriptValue: BoolValue
print("   Number:", 42.5) -- ScriptValue: NumberValue
print("   String:", "hello world") -- ScriptValue: StringValue

-- 2. Collections
print("\n2. Collections:")
local array = { 1, 2, "three", true } -- ScriptValue: ArrayValue
print("   Array:", table.concat(array, ", "))

local object = { -- ScriptValue: ObjectValue
    name = "test",
    age = 25,
    active = true,
}
print("   Object fields:")
for k, v in pairs(object) do
    print("     " .. k .. ":", v)
end

-- 3. Nested Structures
print("\n3. Nested Structures:")
local nested = {
    users = {
        { name = "Alice", age = 30 },
        { name = "Bob", age = 25 },
    },
    settings = {
        theme = "dark",
        notifications = true,
    },
}
print("   Complex nested structure created")

-- 4. Functions (when passed to bridges)
print("\n4. Functions:")
local function myCallback(result)
    print("   Callback received:", result)
    return "processed"
end
print("   Function defined (becomes FunctionValue when passed to bridge)")

-- 5. Error Handling
print("\n5. Error Handling:")
-- When bridges return errors, they become ErrorValue
-- In Lua, we typically see them as second return values
local success, err = pcall(function()
    error("example error")
end)
if not success then
    print("   Error caught:", err)
end

-- 6. Type Safety in Bridge Calls
print("\n6. Type Safety in Bridge Calls:")
print("   When calling bridge methods, arguments are converted to ScriptValue")
print("   The bridge validates types and returns appropriate errors if mismatched")

-- Example with util module if available
if util then
    print("\n7. Real Bridge Example (util module):")

    -- JSON encoding/decoding shows type conversion
    local data = {
        name = "test",
        count = 42,
        enabled = true,
        items = { "a", "b", "c" },
    }

    local json_str = util.json_encode(data)
    print("   Encoded to JSON:", json_str)

    local decoded = util.json_decode(json_str)
    print("   Decoded back - type preserved")
    print("     name:", decoded.name, "(type:", type(decoded.name), ")")
    print("     count:", decoded.count, "(type:", type(decoded.count), ")")
    print("     enabled:", decoded.enabled, "(type:", type(decoded.enabled), ")")
end

-- 8. Type Conversion Edge Cases
print("\n8. Type Conversion Edge Cases:")
print("   Empty table: {}") -- Could be ArrayValue or ObjectValue
print("   Mixed array: {1, 'two', true}") -- ArrayValue with mixed types
print("   Sparse array: {[1]=1, [3]=3}") -- Becomes ObjectValue
print("   Float vs int: 42.0 vs 42") -- Both are NumberValue

print("\n=== Demo Complete ===")

-- Return a complex value to show final conversion
return {
    status = "success",
    results = {
        types_demonstrated = 8,
        message = "ScriptValue system provides type safety between Lua and Go",
    },
}
