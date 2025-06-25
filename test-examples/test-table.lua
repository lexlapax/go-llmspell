print("Testing table functions:")
print("type(table):", type(table))
print("type(table.insert):", type(table.insert))

local t = {}
table.insert(t, "test")
print("Table after insert:", t[1])