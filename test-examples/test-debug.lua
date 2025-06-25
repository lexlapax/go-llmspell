-- Test debug module
print("Testing debug module...")

-- Check if debug module exists
local success, debug_mod = pcall(require, "debug")
if success then
    print("Debug module loaded:", debug_mod ~= nil)
    if debug_mod then
        print("Debug functions:")
        for k, v in pairs(debug_mod) do
            print("  ", k, type(v))
        end
    end
else
    print("Failed to load debug module:", debug_mod)
end

-- Check standard Lua debug
if debug then
    print("\nStandard Lua debug available:")
    for k, v in pairs(debug) do
        print("  ", k, type(v))
    end
end

-- Check log module
local success2, log = pcall(require, "log")
if success2 then
    print("\nLog module available:", log ~= nil)
end

return "done"