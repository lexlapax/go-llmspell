-- Test file read tool
local tools = require("tools")

print("Testing file_read tool...")

-- Create a test file first
local write_result = tools.execute_safe("file_write", {
    path = "/tmp/test-file.txt",
    content = "Hello from test file!",
    create_dirs = true
})

if write_result and not write_result.error then
    print("Created test file successfully")
else
    print("Error creating test file:", write_result and write_result.error or "unknown")
end

-- Now try to read it
local read_result = tools.execute_safe("file_read", {
    path = "/tmp/test-file.txt"
})

if read_result and not read_result.error then
    print("File content:", read_result.content)
else
    print("Error reading file:", read_result and read_result.error or "unknown")
end

return "Test completed"