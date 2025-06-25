-- ABOUTME: Example demonstrating usage of tools module for custom tool creation
-- ABOUTME: Shows tool discovery, definition, execution, and composition

-- Required modules
local data = require("data")
local tools = require("tools")

-- Tools Module Usage Example
-- This spell demonstrates the tools module API:
-- 1. Tool discovery with tools.list() and tools.get_info()
-- 2. Custom tool creation with tools.define()
-- 3. Safe tool execution with tools.execute_safe()
-- 4. Tool composition and metrics

-- Parameters
local iterations = tonumber(params and params.iterations) or 5

print("=== Tools Module Usage Examples ===")
print("Test iterations: " .. iterations)
print()

-- Example 1: Tool Discovery
print("=== Example 1: Tool Discovery ===")

-- List all available tools from bridge
local available_tools = tools.list()
print("Available tools from bridge: " .. #available_tools)

-- Show first 5 tools
print("\nFirst 5 tools:")
for i = 1, math.min(5, #available_tools) do
    local tool = available_tools[i]
    print(string.format("  %d. %s (%s)", i, tool.name, tool.category or "uncategorized"))
end

-- Get info about a specific tool if available
if #available_tools > 0 then
    local first_tool = available_tools[1]
    local tool_info = tools.get_info(first_tool.name)
    if tool_info then
        print("\nDetailed info for '" .. first_tool.name .. "':")
        print("  Description: " .. (tool_info.description or "N/A"))
        print("  Version: " .. (tool_info.version or "N/A"))
    end
end
print()

-- Example 2: Define Custom Tools
print("=== Example 2: Define Custom Tools ===")

-- Define a simple calculator tool
local calc_tool = tools.define(
    "my_calculator",
    "Simple calculator for basic operations",
    {
        parameters = {
            a = { type = "number", required = true },
            b = { type = "number", required = true },
            op = { type = "string", required = true }
        }
    },
    function(params)
        local a, b, op = params.a, params.b, params.op
        if op == "+" then return a + b
        elseif op == "-" then return a - b
        elseif op == "*" then return a * b
        elseif op == "/" then 
            if b == 0 then error("Division by zero") end
            return a / b
        else
            error("Unknown operation: " .. op)
        end
    end
)
print("Defined tool: " .. calc_tool.name)

-- Define a string processor tool
local string_tool = tools.define(
    "string_processor",
    "Process strings with various operations",
    {
        parameters = {
            text = { type = "string", required = true },
            operation = { type = "string", required = true }
        },
        examples = {
            {input = {text = "hello", operation = "upper"}, output = "HELLO"},
            {input = {text = "WORLD", operation = "lower"}, output = "world"}
        }
    },
    function(params)
        local text = params.text
        local op = params.operation
        
        if op == "upper" then return text:upper()
        elseif op == "lower" then return text:lower()
        elseif op == "reverse" then return text:reverse()
        elseif op == "length" then return #text
        else error("Unknown operation: " .. op)
        end
    end
)
print("Defined tool: " .. string_tool.name)
print()

-- Example 3: Execute Custom Tools
print("=== Example 3: Execute Custom Tools ===")

-- Execute calculator tool
print("\nCalculator tests:")
local result = tools.execute_safe("my_calculator", {a = 10, b = 5, op = "+"})
print("  10 + 5 = " .. tostring(result))

result = tools.execute_safe("my_calculator", {a = 20, b = 4, op = "/"})
print("  20 / 4 = " .. tostring(result))

-- Test error handling
local success, err = pcall(function()
    return tools.execute_safe("my_calculator", {a = 10, b = 0, op = "/"})
end)
if not success then
    print("  Division by zero caught: " .. tostring(err):match("([^:]+)$"))
end

-- Execute string processor
print("\nString processor tests:")
result = tools.execute_safe("string_processor", {text = "hello world", operation = "upper"})
print("  Upper: " .. tostring(result))

result = tools.execute_safe("string_processor", {text = "stressed", operation = "reverse"})
print("  Reverse: " .. tostring(result))

result = tools.execute_safe("string_processor", {text = "testing", operation = "length"})
print("  Length: " .. tostring(result))
print()

-- Example 4: Tool Composition
print("=== Example 4: Tool Composition ===")

-- Create a pipeline that processes data through multiple tools
local data_pipeline = tools.compose(
    {
        -- First: generate some text
        function(input)
            return "Hello World from Pipeline"
        end,
        -- Second: convert to uppercase
        function(text)
            return tools.execute_safe("string_processor", {
                text = text,
                operation = "upper"
            })
        end,
        -- Third: get the length
        function(text)
            return {
                original = text,
                length = tools.execute_safe("string_processor", {
                    text = text,
                    operation = "length"
                })
            }
        end
    },
    {name = "text_pipeline", mode = "pipeline"}
)

print("\nExecuting pipeline:")
result = tools.execute_safe(data_pipeline, nil)
print("  Original text: " .. tostring(result.original))
print("  Length: " .. tostring(result.length))
print()

-- Example 5: Tool Validation and Metrics
print("=== Example 5: Tool Validation and Metrics ===")

-- Validate tool parameters
print("\nParameter validation:")
local validation = tools.validate_params("my_calculator", {
    a = 10,
    b = "not a number",  -- This should fail
    op = "+"
})
print("  Invalid params: " .. (validation.valid and "Valid" or "Invalid"))
if not validation.valid then
    print("  Errors: " .. table.concat(validation.errors, "; "))
end

-- Valid parameters
validation = tools.validate_params("my_calculator", {
    a = 10,
    b = 20,
    op = "+"
})
print("  Valid params: " .. (validation.valid and "Valid" or "Invalid"))

-- Get tool metrics
print("\nTool metrics:")
local metrics = tools.get_metrics("my_calculator")
if metrics then
    print("  Total calls: " .. metrics.total_calls)
    print("  Successful: " .. metrics.successful_calls)
    print("  Failed: " .. metrics.failed_calls)
    print("  Avg duration: " .. string.format("%.4f", metrics.avg_duration) .. "s")
end
print()

-- Example 6: Advanced Tool Features
print("=== Example 6: Advanced Tool Features ===")

-- Create a tool with custom validation
local validated_tool = tools.define(
    "age_calculator",
    "Calculate age from birth year",
    {
        parameters = {
            birth_year = { type = "number", required = true },
            current_year = { type = "number", required = false }
        },
        constraints = {
            "birth_year must be between 1900 and current year",
            "birth_year must be a positive integer"
        }
    },
    function(params)
        local birth_year = params.birth_year
        local current_year = params.current_year or tonumber(os.date("%Y"))
        
        -- Validation
        if birth_year < 1900 or birth_year > current_year then
            error("Invalid birth year: must be between 1900 and " .. current_year)
        end
        
        return {
            age = current_year - birth_year,
            birth_year = birth_year,
            current_year = current_year
        }
    end
)

print("\nTesting age calculator:")
result = tools.execute_safe("age_calculator", {birth_year = 1990})
print("  Person born in 1990 is " .. result.age .. " years old")

-- Test with custom current year
result = tools.execute_safe("age_calculator", {birth_year = 2000, current_year = 2030})
print("  Person born in 2000 will be " .. result.age .. " years old in 2030")

-- Example 7: Batch Operations and Performance
print("\n=== Example 7: Batch Operations ===")

-- Test performance with multiple executions
print("\nRunning " .. iterations .. " iterations for performance test:")
local start_time = os.clock()

for i = 1, iterations do
    -- Execute calculator
    tools.execute_safe("my_calculator", {
        a = i,
        b = i * 2,
        op = "+"
    }, {silent = true})
    
    -- Execute string processor
    tools.execute_safe("string_processor", {
        text = "test" .. i,
        operation = "upper"
    }, {silent = true})
end

local elapsed = os.clock() - start_time
print(string.format("  Completed %d tool executions in %.4f seconds", iterations * 2, elapsed))
print(string.format("  Average time per execution: %.4f seconds", elapsed / (iterations * 2)))

-- Get execution history
local history = tools.get_history(5)
print("\nRecent executions (last 5):")
for i, exec in ipairs(history) do
    print(string.format("  %d. Tool: %s, Success: %s, Duration: %.4fs",
        i, exec.tool, tostring(exec.success), exec.duration))
end
print()

-- Summary
print("=== Summary ===")
print("This example demonstrated the tools module API:")
print("1. Tool discovery with tools.list() and tools.get_info()")
print("2. Custom tool creation with tools.define()")
print("3. Tool execution with tools.execute_safe()")
print("4. Tool composition with tools.compose()")
print("5. Parameter validation with tools.validate_params()")
print("6. Metrics tracking with tools.get_metrics()")
print("7. Batch operations and performance testing")
print()
print("Custom tools created: my_calculator, string_processor, age_calculator")

-- Get final metrics
local all_metrics = tools.get_metrics()
local total_executions = 0
for tool_name, metrics in pairs(all_metrics) do
    total_executions = total_executions + metrics.total_calls
end

-- Return summary data
return {
    tools_created = {
        "my_calculator",
        "string_processor",
        "age_calculator"
    },
    total_executions = total_executions,
    performance = {
        iterations = iterations,
        tools_per_iteration = 2
    }
}