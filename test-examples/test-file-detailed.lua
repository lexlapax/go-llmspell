-- Test file tools with detailed output
local tools = require("tools")

print("Testing file tools with detailed output...")

-- Write a file
print("\n1. Writing file...")
local write_result = tools.execute_safe("file_write", {
    path = "/tmp/test-detailed.txt",
    content = "This is a test file with some content.",
    create_dirs = true
})

print("Write result type:", type(write_result))
if write_result then
    for k, v in pairs(write_result) do
        print("  " .. tostring(k) .. ":", tostring(v))
    end
end

-- Read the file
print("\n2. Reading file...")
local read_result = tools.execute_safe("file_read", {
    path = "/tmp/test-detailed.txt"
})

print("Read result type:", type(read_result))
if read_result then
    for k, v in pairs(read_result) do
        print("  " .. tostring(k) .. ":", tostring(v), "(" .. type(v) .. ")")
    end
end

-- List files
print("\n3. Listing files...")
local list_result = tools.execute_safe("file_list", {
    path = "/tmp",
    pattern = "test-*.txt"
})

print("List result type:", type(list_result))
if list_result then
    for k, v in pairs(list_result) do
        print("  " .. tostring(k) .. ":", tostring(v), "(" .. type(v) .. ")")
    end
end

return "Test completed"