-- ABOUTME: Example demonstrating use of built-in tools from go-llms
-- ABOUTME: Shows how to use web_fetch and other pre-registered tools

-- List all available tools
print("Available built-in tools:")
local tools_list = tools.list()
for _, tool in ipairs(tools_list) do
    print(string.format("- %s: %s", tool.name, tool.description))
    if tool.parameters then
        print("  Parameters: " .. json.encode(tool.parameters))
    end
end
print()

-- Example 1: Using web_fetch tool
print("Example 1: Fetching content from a URL")
local web_fetch = tools.get("web_fetch")
if web_fetch then
    print("web_fetch tool found!")
    print("Description: " .. web_fetch.description)
    print("Parameters: " .. json.encode(web_fetch.parameters))
    
    -- Test with a simple URL (example.com is a safe test site)
    local params = {
        url = "https://example.com"
    }
    
    -- Validate parameters first
    local ok, err = pcall(tools.validate, "web_fetch", params)
    if not ok then
        print("Parameter validation failed: " .. tostring(err))
    else
        print("Parameters validated successfully")
        
        -- Execute the tool
        print("\nFetching https://example.com...")
        local success, result = pcall(tools.execute, "web_fetch", params)
        if success then
            print("Success! Status: " .. tostring(result.status))
            print("Content length: " .. string.len(result.content) .. " bytes")
            -- Print first 200 characters of content
            if string.len(result.content) > 200 then
                print("First 200 chars: " .. string.sub(result.content, 1, 200) .. "...")
            else
                print("Content: " .. result.content)
            end
        else
            print("Fetch failed: " .. tostring(result))
        end
    end
else
    print("web_fetch tool not found!")
end

print("\n" .. string.rep("-", 50) .. "\n")

-- Example 2: Demonstrating parameter validation
print("Example 2: Parameter validation")

-- Try with invalid parameters
print("Testing web_fetch with invalid URL...")
local invalid_params = {
    url = "not-a-valid-url"
}

local success, result = pcall(tools.execute, "web_fetch", invalid_params)
if not success then
    print("Expected error: " .. tostring(result))
else
    print("Unexpected success")
end

-- Try with missing required parameter
print("\nTesting web_fetch with missing URL...")
local missing_params = {}

local ok, err = pcall(tools.validate, "web_fetch", missing_params)
if not ok then
    print("Expected validation error: " .. tostring(err))
else
    print("Unexpected validation success")
end

print("\n" .. string.rep("-", 50) .. "\n")

-- Example 3: Working with tool information
print("Example 3: Tool information")

-- Get detailed info about a specific tool
local tool_info = tools.get("web_fetch")
if tool_info then
    print("Tool: " .. tool_info.name)
    print("Description: " .. tool_info.description)
    
    if tool_info.parameters and tool_info.parameters.properties then
        print("Parameters:")
        for param_name, param_info in pairs(tool_info.parameters.properties) do
            print(string.format("  - %s (%s): %s", 
                param_name, 
                param_info.type or "unknown",
                param_info.description or "No description"))
            if param_info.format then
                print("    Format: " .. param_info.format)
            end
        end
    end
    
    if tool_info.parameters and tool_info.parameters.required then
        print("Required parameters: " .. table.concat(tool_info.parameters.required, ", "))
    end
end

print("\nScript completed successfully!")