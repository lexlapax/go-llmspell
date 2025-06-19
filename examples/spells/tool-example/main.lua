-- Tool System Example
-- Demonstrates how to register and use tools in Lua scripts

-- Register a simple calculator tool
tools.register("calculator", "Performs basic arithmetic operations", {
    type = "object",
    properties = {
        operation = { type = "string", enum = { "add", "subtract", "multiply", "divide" } },
        a = { type = "number" },
        b = { type = "number" },
    },
    required = { "operation", "a", "b" },
}, function(params)
    local op = params.operation
    local a = params.a
    local b = params.b

    if op == "add" then
        return a + b
    elseif op == "subtract" then
        return a - b
    elseif op == "multiply" then
        return a * b
    elseif op == "divide" then
        if b == 0 then
            return nil, "Division by zero"
        end
        return a / b
    else
        return nil, "Unknown operation: " .. op
    end
end)

-- Register a string manipulation tool
tools.register("string_tools", "Provides string manipulation utilities", {
    type = "object",
    properties = {
        action = { type = "string", enum = { "upper", "lower", "reverse", "length" } },
        text = { type = "string" },
    },
    required = { "action", "text" },
}, function(params)
    local action = params.action
    local text = params.text

    if action == "upper" then
        return string.upper(text)
    elseif action == "lower" then
        return string.lower(text)
    elseif action == "reverse" then
        return string.reverse(text)
    elseif action == "length" then
        return string.len(text)
    else
        return nil, "Unknown action: " .. action
    end
end)

-- Register a JSON processing tool
tools.register("json_processor", "Processes JSON data", {
    type = "object",
    properties = {
        action = { type = "string", enum = { "parse", "stringify", "get_field" } },
        data = { type = "string" },
        field = { type = "string" },
    },
    required = { "action" },
}, function(params)
    local action = params.action

    if action == "parse" then
        if not params.data then
            return nil, "data field is required for parse action"
        end
        return json.decode(params.data)
    elseif action == "stringify" then
        if not params.data then
            return nil, "data field is required for stringify action"
        end
        -- For stringify, we expect already parsed data
        return json.encode(params.data)
    elseif action == "get_field" then
        if not params.data or not params.field then
            return nil, "data and field are required for get_field action"
        end
        local parsed = json.decode(params.data)
        return parsed[params.field]
    else
        return nil, "Unknown action: " .. action
    end
end)

-- List all registered tools
print("=== Registered Tools ===")
local toolList = tools.list()
for i, tool in ipairs(toolList) do
    print(string.format("%d. %s - %s", i, tool.name, tool.description))
end
print()

-- Test calculator tool
print("=== Calculator Tool Tests ===")

-- Test addition
local result, err = tools.execute("calculator", { operation = "add", a = 10, b = 5 })
if err then
    print("Error:", err)
else
    print("10 + 5 =", result)
end

-- Test division
result, err = tools.execute("calculator", { operation = "divide", a = 20, b = 4 })
if err then
    print("Error:", err)
else
    print("20 / 4 =", result)
end

-- Test division by zero
result, err = tools.execute("calculator", { operation = "divide", a = 10, b = 0 })
if err then
    print("Division by zero error:", err)
else
    print("Result:", result)
end

print()

-- Test string tools
print("=== String Tools Tests ===")

result, err = tools.execute("string_tools", { action = "upper", text = "hello world" })
if err then
    print("Error:", err)
else
    print("Upper case:", result)
end

result, err = tools.execute("string_tools", { action = "reverse", text = "hello" })
if err then
    print("Error:", err)
else
    print("Reversed:", result)
end

print()

-- Test JSON processor
print("=== JSON Processor Tests ===")

local jsonData = '{"name": "Alice", "age": 30, "city": "New York"}'
result, err = tools.execute("json_processor", { action = "parse", data = jsonData })
if err then
    print("Error:", err)
else
    print("Parsed JSON:")
    for k, v in pairs(result) do
        print("  " .. k .. ":", v)
    end
end

result, err =
    tools.execute("json_processor", { action = "get_field", data = jsonData, field = "name" })
if err then
    print("Error:", err)
else
    print("Name field:", result)
end

print()

-- Test parameter validation
print("=== Parameter Validation Tests ===")

-- Valid parameters
local isValid, err = tools.validate("calculator", { operation = "add", a = 1, b = 2 })
if isValid then
    print("Valid parameters for calculator")
else
    print("Invalid parameters:", err)
end

-- Missing required parameter
isValid, err = tools.validate("calculator", { operation = "add", a = 1 })
if isValid then
    print("Valid parameters")
else
    print("Invalid parameters:", err)
end

-- Wrong type
isValid, err = tools.validate("string_tools", { action = "upper", text = 123 })
if isValid then
    print("Valid parameters")
else
    print("Invalid parameters:", err)
end

print()

-- Get tool information
print("=== Tool Information ===")
local info, err = tools.get("calculator")
if err then
    print("Error:", err)
else
    print("Calculator tool info:")
    print("  Name:", info.name)
    print("  Description:", info.description)
    print("  Parameters:", json.encode(info.parameters))
end

-- Clean up - remove one tool
print("\n=== Cleanup ===")
local success, err = tools.remove("json_processor")
if success then
    print("Successfully removed json_processor tool")
else
    print("Error removing tool:", err)
end

-- List tools again to confirm removal
print("\nRemaining tools:")
toolList = tools.list()
for i, tool in ipairs(toolList) do
    print(string.format("%d. %s", i, tool.name))
end
