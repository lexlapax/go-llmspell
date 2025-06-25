-- Test table.insert
local t = {}
print("Type of table:", type(table))
print("Type of table.insert:", type(table.insert))

-- Test basic insert
table.insert(t, "test")
print("After insert:", t[1])

-- Test insert with table
local handlers = {}
table.insert(handlers, {
    handler = function() print("Handler") end,
    priority = 0,
    once = false
})
print("Handlers count:", #handlers)

return "Success"