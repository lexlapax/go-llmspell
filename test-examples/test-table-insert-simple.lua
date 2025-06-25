-- Test table.insert
print("Testing table.insert...")

local t = {}
print("Type of table:", type(table))
print("Type of table.insert:", type(table.insert))

-- Try direct insert
table.insert(t, "hello")
print("After insert, #t =", #t)

return "Test completed"