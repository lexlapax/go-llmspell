-- Test utils module
local utils = require("utils")

print("Utils module loaded:", type(utils))
print("\nChecking functions:")
print("  file_exists:", type(utils.file_exists))
print("  file_write:", type(utils.file_write))
print("  file_read:", type(utils.file_read))
print("  mkdir:", type(utils.mkdir))
print("  list_files:", type(utils.list_files))
print("  sleep:", type(utils.sleep))
print("  env:", type(utils.env))
print("  exec:", type(utils.exec))
print("  current_time:", type(utils.current_time))
print("  format_time:", type(utils.format_time))
print("  join_path:", type(utils.join_path))

-- Test some basic operations
print("\nTesting basic operations:")

-- Time operations (these should work without tools)
local now = utils.current_time()
print("  Current time:", now)
print("  Formatted:", utils.format_time(now))

-- Path operations (these should work without tools)
local path = utils.join_path("home", "user", "file.txt")
print("  Joined path:", path)
print("  Extension of 'test.lua':", utils.file_extension("test.lua"))
print("  Basename of '/path/to/file.txt':", utils.basename("/path/to/file.txt"))
print("  Dirname of '/path/to/file.txt':", utils.dirname("/path/to/file.txt"))

-- Try to get an environment variable
local user = utils.env("USER")
print("  USER env var:", user or "(not found)")

return "Utils module test completed"