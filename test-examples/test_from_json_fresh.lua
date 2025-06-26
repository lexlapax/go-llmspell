-- ABOUTME: Test data.from_json alias in fresh runtime
-- ABOUTME: Verify alias is available after adding to data.lua

-- Don't cache, load fresh
package.loaded["data"] = nil
local data = require("data")

-- Test both functions
local json_str = '{"test":"value","number":42}'

-- Test parse_json
local parse_result = data.parse_json(json_str)

-- Test from_json alias
local from_result = nil
if data.from_json then
    from_result = data.from_json(json_str)
end

return {
    has_parse_json = type(data.parse_json) == "function",
    has_from_json = type(data.from_json) == "function", 
    parse_result = parse_result,
    from_result = from_result,
    both_work = parse_result and from_result and parse_result.test == from_result.test
}