-- Test utils module
local utils = require("utils")

print("Testing utils module...")

-- Test file operations
print("\n1. Writing file with utils...")
local write_result = utils.file_write("/tmp/test-utils.txt", "Hello from utils module!")
print("Write result:", write_result)
if write_result and write_result.error then
    print("Write error:", write_result.error)
end

-- Test file exists
print("\n2. Checking if file exists...")
local exists = utils.file_exists("/tmp/test-utils.txt")
print("File exists:", exists)

-- Test file read
print("\n3. Reading file...")
local content, err = utils.file_read("/tmp/test-utils.txt")
if content then
    print("File content:", content)
else
    print("Read error:", err)
end

-- Test sleep
print("\n4. Testing sleep (1 second)...")
local start = os.time()
utils.sleep(1)
local elapsed = os.time() - start
print("Elapsed time:", elapsed, "seconds")

-- Test env
print("\n5. Testing environment variable...")
local path_env = utils.env("PATH")
if path_env then
    print("PATH found (truncated):", path_env:sub(1, 50) .. "...")
else
    print("PATH not found")
end

return "Utils test completed"