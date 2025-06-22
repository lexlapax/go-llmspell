-- ABOUTME: Example demonstrating usage of built-in tools without LLM interaction
-- ABOUTME: Shows file operations, web fetching, datetime, calculator, and tool discovery

-- Built-in Tools Usage Example
-- This spell demonstrates how to use various built-in tools:
-- 1. File operations (read, write, list, exists, delete)
-- 2. Web operations (fetch)
-- 3. DateTime utilities
-- 4. Calculator
-- 5. Tool discovery and inspection

-- Parameters
local output_dir = params.output_dir or "./tool-examples-output"
local test_url = params.test_url or "https://api.github.com/repos/anthropics/claude-docs"

print("=== Built-in Tools Usage Examples ===")
print("Output directory: " .. output_dir)
print()

-- Example 1: Tool Discovery
print("=== Example 1: Tool Discovery ===")

-- List all available tools
local available_tools = tools.list()
print("Available tools:")
for i, tool in ipairs(available_tools) do
    print(string.format("  %d. %s - %s", i, tool.name, tool.description))
end
print()

-- Get detailed info about a specific tool
local file_write_info = tools.get_info("file_write")
if file_write_info then
    print("Detailed info for 'file_write' tool:")
    print("  Description: " .. file_write_info.description)
    print("  Parameters: " .. data.to_json(file_write_info.parameters))
end
print()

-- Example 2: File Operations
print("=== Example 2: File Operations ===")

-- Create output directory if it doesn't exist
if not tools.file_exists(output_dir) then
    print("Creating output directory: " .. output_dir)
    tools.create_directory(output_dir)
end

-- Write a file
local test_content = [[
This is a test file created by go-llmspell tools example.
Created at: ]] .. os.date() .. [[

Here's some JSON data:
{
    "example": true,
    "timestamp": "]] .. os.date() .. [[",
    "values": [1, 2, 3, 4, 5]
}
]]

local test_file = output_dir .. "/test-file.txt"
print("Writing to file: " .. test_file)
tools.file_write(test_file, test_content)

-- Read the file back
print("Reading file contents:")
local read_content = tools.file_read(test_file)
print("File contents (first 100 chars): " .. read_content:sub(1, 100) .. "...")

-- List files in directory
print("\nListing files in output directory:")
local files = tools.list_files(output_dir)
for _, file in ipairs(files) do
    print("  - " .. file)
end

-- Check if file exists
local exists = tools.file_exists(test_file)
print("\nFile exists check: " .. tostring(exists))

-- Copy file
local copy_file = output_dir .. "/test-file-copy.txt"
print("Copying file to: " .. copy_file)
tools.file_copy(test_file, copy_file)

-- Move/rename file
local moved_file = output_dir .. "/test-file-moved.txt"
print("Moving file to: " .. moved_file)
tools.file_move(copy_file, moved_file)

-- Delete file
print("Deleting moved file...")
tools.file_delete(moved_file)

print()

-- Example 3: DateTime Operations
print("=== Example 3: DateTime Operations ===")

-- Get current datetime
local now = tools.datetime_now()
print("Current datetime: " .. now)

-- Format datetime
local formatted = tools.datetime_format(now, "YYYY-MM-DD HH:mm:ss")
print("Formatted: " .. formatted)

-- Parse datetime
local parsed = tools.datetime_parse("2024-01-15 14:30:00", "YYYY-MM-DD HH:mm:ss")
print("Parsed datetime: " .. parsed)

-- Add/subtract time
local tomorrow = tools.datetime_add(now, 1, "day")
print("Tomorrow: " .. tools.datetime_format(tomorrow, "YYYY-MM-DD"))

local last_week = tools.datetime_subtract(now, 7, "days")
print("Last week: " .. tools.datetime_format(last_week, "YYYY-MM-DD"))

-- Calculate difference
local diff_days = tools.datetime_diff(now, last_week, "days")
print("Difference in days: " .. diff_days)

-- Get timestamp
local timestamp = tools.datetime_timestamp(now)
print("Unix timestamp: " .. timestamp)

print()

-- Example 4: Calculator Operations
print("=== Example 4: Calculator Operations ===")

-- Basic arithmetic
local calculations = {
    "2 + 2",
    "10 * 5",
    "100 / 4",
    "2 ^ 8",
    "sqrt(16)",
    "sin(3.14159 / 2)",
    "log(100)",
    "abs(-42)",
    "round(3.14159, 2)",
    "max(10, 20, 30)",
    "min(10, 20, 30)",
    "(5 + 3) * 2"
}

for _, expr in ipairs(calculations) do
    local result = tools.calculator(expr)
    print(string.format("  %s = %s", expr, tostring(result)))
end

-- Complex calculation
local complex_expr = "sqrt(3^2 + 4^2)"  -- Pythagorean theorem
local complex_result = tools.calculator(complex_expr)
print(string.format("\nComplex calculation: %s = %s", complex_expr, complex_result))

print()

-- Example 5: Web Fetch Operations
print("=== Example 5: Web Fetch Operations ===")

print("Fetching from: " .. test_url)

-- Fetch web content
local web_content = tools.web_fetch(test_url)
print("Response length: " .. #web_content .. " characters")

-- Try to parse as JSON (GitHub API returns JSON)
local success, parsed_data = pcall(function()
    return data.from_json(web_content)
end)

if success and parsed_data then
    print("Successfully parsed JSON response:")
    print("  Repository: " .. (parsed_data.full_name or "N/A"))
    print("  Description: " .. (parsed_data.description or "N/A"))
    print("  Stars: " .. tostring(parsed_data.stargazers_count or "N/A"))
    print("  Language: " .. (parsed_data.language or "N/A"))
    
    -- Save the response
    local response_file = output_dir .. "/web-response.json"
    tools.file_write(response_file, data.to_json(parsed_data, {pretty = true}))
    print("\nSaved response to: " .. response_file)
else
    print("Response was not JSON or failed to parse")
    -- Save raw response
    local response_file = output_dir .. "/web-response.txt"
    tools.file_write(response_file, web_content:sub(1, 1000))  -- First 1000 chars
    print("Saved raw response preview to: " .. response_file)
end

print()

-- Example 6: Combined Tool Usage
print("=== Example 6: Combined Tool Usage ===")

-- Create a report combining multiple tools
local report = {
    title = "Tool Usage Report",
    generated_at = tools.datetime_format(tools.datetime_now(), "YYYY-MM-DD HH:mm:ss"),
    system_info = {
        timestamp = tools.datetime_timestamp(tools.datetime_now()),
        calculations = {
            random_number = tools.calculator("random() * 100"),
            pi_squared = tools.calculator("3.14159 ^ 2")
        }
    },
    files_created = {},
    web_fetch_summary = nil
}

-- Add file info
for _, file in ipairs(tools.list_files(output_dir)) do
    table.insert(report.files_created, {
        name = file,
        path = output_dir .. "/" .. file
    })
end

-- Add web fetch summary
if parsed_data then
    report.web_fetch_summary = {
        url = test_url,
        repo_name = parsed_data.full_name,
        stars = parsed_data.stargazers_count
    }
end

-- Save report
local report_file = output_dir .. "/tools-usage-report.json"
tools.file_write(report_file, data.to_json(report, {pretty = true}))
print("Generated comprehensive report: " .. report_file)

-- Example 7: Error Handling with Tools
print("\n=== Example 7: Tool Error Handling ===")

-- Try to read non-existent file
local success, error_msg = pcall(function()
    return tools.file_read("/this/file/does/not/exist.txt")
end)
if not success then
    print("Expected error reading non-existent file: " .. tostring(error_msg))
end

-- Try invalid calculator expression
success, error_msg = pcall(function()
    return tools.calculator("invalid expression !!!")
end)
if not success then
    print("Expected error with invalid calculation: " .. tostring(error_msg))
end

-- Safe tool wrapper function
local function safe_tool_call(tool_fn, ...)
    local args = {...}
    local success, result = pcall(function()
        return tool_fn(table.unpack(args))
    end)
    
    if success then
        return {success = true, result = result}
    else
        return {success = false, error = tostring(result)}
    end
end

-- Use safe wrapper
local safe_result = safe_tool_call(tools.file_read, "/another/fake/file.txt")
if not safe_result.success then
    print("Safely handled error: " .. safe_result.error)
end

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Tool discovery and inspection")
print("2. File operations (read, write, copy, move, delete)")
print("3. DateTime utilities")
print("4. Calculator functions")
print("5. Web fetching")
print("6. Combined tool usage")
print("7. Error handling")
print()
print("Check the output directory for generated files: " .. output_dir)

-- Return summary data
return {
    tools_demonstrated = {
        "file_read", "file_write", "file_exists", "file_copy",
        "file_move", "file_delete", "list_files", "create_directory",
        "datetime_now", "datetime_format", "datetime_parse",
        "datetime_add", "datetime_subtract", "datetime_diff",
        "calculator", "web_fetch"
    },
    files_created = report.files_created,
    output_directory = output_dir
}