-- Test promise availability
print("Testing promise module...")

-- Check global promise
print("_G.promise:", _G.promise)

-- Try to require promise
local success, promise = pcall(require, "promise")
print("require('promise') success:", success)
if success then
    print("promise module:", promise)
else
    print("error:", promise)
end

return "done"