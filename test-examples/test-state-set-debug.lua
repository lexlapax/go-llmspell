-- Debug state.set issue
local state = require("state")

print("Testing state.set...")

-- Try simple set
print("\n1. Simple set:")
local success1, err1 = pcall(function()
    state.set("test", "value")
end)
print("Result:", success1, err1)

-- Try setting a table
print("\n2. Setting a table:")
local success2, err2 = pcall(function()
    state.set("mytable", {a = 1, b = 2})
end)
print("Result:", success2, err2)

-- Try setting nested path
print("\n3. Setting nested path:")
local success3, err3 = pcall(function()
    state.set("app.name", "test app")
end)
print("Result:", success3, err3)

-- Try getting back
print("\n4. Getting values:")
print("test =", state.get("test"))
print("mytable =", state.get("mytable"))
print("app.name =", state.get("app.name"))

return "done"