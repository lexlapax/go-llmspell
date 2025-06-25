-- Test table.sort
print("Type of table.sort:", type(table.sort))

local t = {3, 1, 4, 1, 5, 9}
table.sort(t)
print("Sorted:", table.concat(t, ", "))

-- Test with custom comparator
local items = {
    {name = "apple", priority = 3},
    {name = "banana", priority = 1},
    {name = "cherry", priority = 2}
}

table.sort(items, function(a, b)
    return a.priority > b.priority
end)

print("Sorted by priority:")
for _, item in ipairs(items) do
    print("  " .. item.name .. " (priority: " .. item.priority .. ")")
end

return "Success"