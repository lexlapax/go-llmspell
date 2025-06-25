-- Test available bridges
print("Testing available bridges...")

if bridges then
    print("Bridges found:")
    for k, v in pairs(bridges) do
        print("  ", k, type(v))
    end
else
    print("No bridges global found")
end

-- Check _G for bridges
print("\nChecking _G:")
for k, v in pairs(_G) do
    if k:find("bridge") or k:find("state") then
        print("  ", k, type(v))
    end
end

return "done"