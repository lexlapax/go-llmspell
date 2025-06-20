-- ABOUTME: Testing & Validation Library for go-llmspell Lua standard library
-- ABOUTME: Provides comprehensive testing framework with assertions, mocking, and performance testing

local testing = {}

-- Internal state
local test_suites = {}
local current_suite = nil
local current_test = nil -- luacheck: ignore current_test
local test_results = {
    total = 0,
    passed = 0,
    failed = 0,
    skipped = 0,
    tests = {},
    start_time = 0,
    end_time = 0,
}
local mocks = {} -- luacheck: ignore 241 (stores mocks for cleanup)
local spies = {} -- luacheck: ignore 241 (stores spies for cleanup)
local stubs = {}
-- These are not used at module level but are stored per suite
-- local before_each_hooks = {}
-- local after_each_hooks = {}
-- local before_all_hooks = {}
-- local after_all_hooks = {}

-- Helper function to validate required parameters
local function validate_required(param, name)
    if param == nil then
        error(name .. " is required")
    end
end

-- Helper function to deep copy tables
local function deep_copy(obj)
    if type(obj) ~= "table" then
        return obj
    end
    local copy = {}
    for k, v in pairs(obj) do
        copy[deep_copy(k)] = deep_copy(v)
    end
    return copy
end

-- Helper function to compare values
local function equals(a, b)
    if type(a) ~= type(b) then
        return false
    end
    if type(a) ~= "table" then
        return a == b
    end
    -- Deep comparison for tables
    for k, v in pairs(a) do
        if not equals(v, b[k]) then
            return false
        end
    end
    for k, v in pairs(b) do
        if not equals(v, a[k]) then
            return false
        end
    end
    return true
end

-- Test Suite Management

-- Create a test suite
function testing.describe(name, test_func)
    validate_required(name, "test suite name")
    validate_required(test_func, "test function")

    if type(test_func) ~= "function" then
        error("test function must be a function")
    end

    local suite = {
        name = name,
        tests = {},
        nested_suites = {},
        before_each = {},
        after_each = {},
        before_all = {},
        after_all = {},
        parent = current_suite,
    }

    -- Store previous suite
    local previous_suite = current_suite
    current_suite = suite

    -- Execute test definition
    local success, err = pcall(test_func)
    if not success then
        error("Failed to define test suite '" .. name .. "': " .. tostring(err))
    end

    -- Restore previous suite
    current_suite = previous_suite

    -- Add to parent or root
    if previous_suite then
        table.insert(previous_suite.nested_suites, suite)
    else
        table.insert(test_suites, suite)
    end

    return suite
end

-- Define an individual test
function testing.it(name, test_func, options)
    validate_required(name, "test name")
    validate_required(test_func, "test function")

    if not current_suite then
        error("it() must be called within describe()")
    end

    if type(test_func) ~= "function" then
        error("test function must be a function")
    end

    options = options or {}

    local test = {
        name = name,
        func = test_func,
        suite = current_suite,
        skip = options.skip or false,
        only = options.only or false,
        timeout = options.timeout or 5000,
    }

    table.insert(current_suite.tests, test)
    return test
end

-- Alias for it
testing.test = testing.it

-- Skip a test
function testing.skip(name, test_func, options)
    options = options or {}
    options.skip = true
    return testing.it(name, test_func, options)
end

-- Run only this test
function testing.only(name, test_func, options)
    options = options or {}
    options.only = true
    return testing.it(name, test_func, options)
end

-- Setup/Teardown Hooks

function testing.before_each(func)
    validate_required(func, "setup function")
    if not current_suite then
        error("before_each() must be called within describe()")
    end
    table.insert(current_suite.before_each, func)
end

function testing.after_each(func)
    validate_required(func, "teardown function")
    if not current_suite then
        error("after_each() must be called within describe()")
    end
    table.insert(current_suite.after_each, func)
end

function testing.before_all(func)
    validate_required(func, "setup function")
    if not current_suite then
        error("before_all() must be called within describe()")
    end
    table.insert(current_suite.before_all, func)
end

function testing.after_all(func)
    validate_required(func, "teardown function")
    if not current_suite then
        error("after_all() must be called within describe()")
    end
    table.insert(current_suite.after_all, func)
end

-- Assertion Library
testing.assert = {}

-- Basic assertions
function testing.assert.equals(actual, expected, message)
    if not equals(actual, expected) then
        error(
            (message or "Assertion failed")
                .. ": expected "
                .. tostring(expected)
                .. ", got "
                .. tostring(actual)
        )
    end
end

function testing.assert.not_equals(actual, expected, message)
    if equals(actual, expected) then
        error((message or "Assertion failed") .. ": expected values to be different")
    end
end

function testing.assert.truthy(value, message)
    if not value then
        error((message or "Assertion failed") .. ": expected truthy value, got " .. tostring(value))
    end
end

testing.assert["true"] = testing.assert.truthy

function testing.assert.falsy(value, message)
    if value then
        error((message or "Assertion failed") .. ": expected falsy value, got " .. tostring(value))
    end
end

testing.assert["false"] = testing.assert.falsy

function testing.assert.is_nil(value, message)
    if value ~= nil then
        error((message or "Assertion failed") .. ": expected nil, got " .. tostring(value))
    end
end

testing.assert["nil"] = testing.assert.is_nil

function testing.assert.not_nil(value, message)
    if value == nil then
        error((message or "Assertion failed") .. ": expected non-nil value")
    end
end

-- Type assertions
function testing.assert.type(value, expected_type, message)
    local actual_type = type(value)
    if actual_type ~= expected_type then
        error(
            (message or "Assertion failed")
                .. ": expected type "
                .. expected_type
                .. ", got "
                .. actual_type
        )
    end
end

function testing.assert.table(value, message)
    testing.assert.type(value, "table", message)
end

function testing.assert.func(value, message)
    testing.assert.type(value, "function", message)
end

testing.assert["function"] = testing.assert.func

function testing.assert.string(value, message)
    testing.assert.type(value, "string", message)
end

function testing.assert.number(value, message)
    testing.assert.type(value, "number", message)
end

function testing.assert.boolean(value, message)
    testing.assert.type(value, "boolean", message)
end

-- Comparison assertions
function testing.assert.greater_than(actual, expected, message)
    if actual <= expected then
        error(
            (message or "Assertion failed")
                .. ": "
                .. tostring(actual)
                .. " is not greater than "
                .. tostring(expected)
        )
    end
end

function testing.assert.less_than(actual, expected, message)
    if actual >= expected then
        error(
            (message or "Assertion failed")
                .. ": "
                .. tostring(actual)
                .. " is not less than "
                .. tostring(expected)
        )
    end
end

function testing.assert.greater_or_equal(actual, expected, message)
    if actual < expected then
        error(
            (message or "Assertion failed")
                .. ": "
                .. tostring(actual)
                .. " is not greater than or equal to "
                .. tostring(expected)
        )
    end
end

function testing.assert.less_or_equal(actual, expected, message)
    if actual > expected then
        error(
            (message or "Assertion failed")
                .. ": "
                .. tostring(actual)
                .. " is not less than or equal to "
                .. tostring(expected)
        )
    end
end

function testing.assert.between(value, min, max, message)
    if not (value >= min and value <= max) then
        error(
            (message or "Assertion failed")
                .. ": "
                .. tostring(value)
                .. " is not between "
                .. tostring(min)
                .. " and "
                .. tostring(max)
        )
    end
end

-- String assertions
function testing.assert.contains(haystack, needle, message)
    validate_required(haystack, "haystack")
    validate_required(needle, "needle")

    local found = false
    if type(haystack) == "string" then
        found = string.find(haystack, needle, 1, true) ~= nil
    elseif type(haystack) == "table" then
        for _, v in pairs(haystack) do
            if equals(v, needle) then
                found = true
                break
            end
        end
    else
        error("contains assertion requires string or table")
    end

    if not found then
        error(
            (message or "Assertion failed")
                .. ": "
                .. tostring(haystack)
                .. " does not contain "
                .. tostring(needle)
        )
    end
end

function testing.assert.matches(str, pattern, message)
    validate_required(str, "string")
    validate_required(pattern, "pattern")

    if type(str) ~= "string" then
        error("matches assertion requires string")
    end

    if not string.match(str, pattern) then
        error(
            (message or "Assertion failed")
                .. ": '"
                .. str
                .. "' does not match pattern '"
                .. pattern
                .. "'"
        )
    end
end

function testing.assert.starts_with(str, prefix, message)
    validate_required(str, "string")
    validate_required(prefix, "prefix")

    if type(str) ~= "string" or type(prefix) ~= "string" then
        error("starts_with assertion requires strings")
    end

    if string.sub(str, 1, #prefix) ~= prefix then
        error(
            (message or "Assertion failed")
                .. ": '"
                .. str
                .. "' does not start with '"
                .. prefix
                .. "'"
        )
    end
end

function testing.assert.ends_with(str, suffix, message)
    validate_required(str, "string")
    validate_required(suffix, "suffix")

    if type(str) ~= "string" or type(suffix) ~= "string" then
        error("ends_with assertion requires strings")
    end

    if string.sub(str, -#suffix) ~= suffix then
        error(
            (message or "Assertion failed")
                .. ": '"
                .. str
                .. "' does not end with '"
                .. suffix
                .. "'"
        )
    end
end

-- Table assertions
function testing.assert.deep_equals(actual, expected, message)
    if not equals(actual, expected) then
        error((message or "Assertion failed") .. ": deep equality check failed")
    end
end

function testing.assert.has_key(tbl, key, message)
    validate_required(tbl, "table")

    if type(tbl) ~= "table" then
        error("has_key assertion requires table")
    end

    if tbl[key] == nil then
        error(
            (message or "Assertion failed") .. ": table does not have key '" .. tostring(key) .. "'"
        )
    end
end

function testing.assert.has_value(tbl, value, message)
    validate_required(tbl, "table")

    if type(tbl) ~= "table" then
        error("has_value assertion requires table")
    end

    local found = false
    for _, v in pairs(tbl) do
        if equals(v, value) then
            found = true
            break
        end
    end

    if not found then
        error((message or "Assertion failed") .. ": table does not contain value")
    end
end

function testing.assert.length(tbl, expected_length, message)
    validate_required(tbl, "table or string")
    validate_required(expected_length, "expected length")

    local actual_length
    if type(tbl) == "table" then
        actual_length = #tbl
    elseif type(tbl) == "string" then
        actual_length = #tbl
    else
        error("length assertion requires table or string")
    end

    if actual_length ~= expected_length then
        error(
            (message or "Assertion failed")
                .. ": expected length "
                .. expected_length
                .. ", got "
                .. actual_length
        )
    end
end

function testing.assert.empty(tbl, message)
    if type(tbl) == "table" then
        if next(tbl) ~= nil then
            error((message or "Assertion failed") .. ": expected empty table")
        end
    elseif type(tbl) == "string" then
        if #tbl > 0 then
            error((message or "Assertion failed") .. ": expected empty string")
        end
    else
        error("empty assertion requires table or string")
    end
end

function testing.assert.not_empty(tbl, message)
    if type(tbl) == "table" then
        if next(tbl) == nil then
            error((message or "Assertion failed") .. ": expected non-empty table")
        end
    elseif type(tbl) == "string" then
        if #tbl == 0 then
            error((message or "Assertion failed") .. ": expected non-empty string")
        end
    else
        error("not_empty assertion requires table or string")
    end
end

-- Error assertions
function testing.assert.error(func, expected_error, message)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("error assertion requires function")
    end

    local success, err = pcall(func)

    if success then
        error((message or "Assertion failed") .. ": expected error but function succeeded")
    end

    if expected_error and not string.find(tostring(err), expected_error, 1, true) then
        error(
            (message or "Assertion failed")
                .. ": expected error containing '"
                .. expected_error
                .. "', got '"
                .. tostring(err)
                .. "'"
        )
    end
end

function testing.assert.no_error(func, message)
    validate_required(func, "function")

    if type(func) ~= "function" then
        error("no_error assertion requires function")
    end

    local success, err = pcall(func)

    if not success then
        error((message or "Assertion failed") .. ": expected no error, got: " .. tostring(err))
    end
end

function testing.assert.error_matches(func, pattern, message)
    validate_required(func, "function")
    validate_required(pattern, "error pattern")

    if type(func) ~= "function" then
        error("error_matches assertion requires function")
    end

    local success, err = pcall(func)

    if success then
        error((message or "Assertion failed") .. ": expected error but function succeeded")
    end

    if not string.match(tostring(err), pattern) then
        error(
            (message or "Assertion failed")
                .. ": error '"
                .. tostring(err)
                .. "' does not match pattern '"
                .. pattern
                .. "'"
        )
    end
end

-- Mocking and Stubbing
testing.mock = {}

-- Mock function creator
local function create_mock_func(name)
    local calls = {}
    local returns = {}
    local throws = nil
    local call_through = false
    local original_func = nil

    local mock_func = function(...)
        local args = { ... }
        table.insert(calls, args)

        if throws then
            error(throws)
        end

        if call_through and original_func then
            return original_func(...)
        end

        -- Check for specific return values based on arguments
        for _, config in ipairs(returns) do
            if config.args then
                local match = true
                for i, arg in ipairs(config.args) do
                    if not equals(arg, args[i]) then
                        match = false
                        break
                    end
                end
                if match then
                    if config.sequence then
                        local idx = config.sequence_index or 1
                        local value = config.values[idx]
                        config.sequence_index = idx + 1
                        if config.sequence_index > #config.values then
                            config.sequence_index = 1
                        end
                        return value
                    else
                        return config.value
                    end
                end
            end
        end

        -- Default return
        for _, config in ipairs(returns) do
            if not config.args then
                if config.sequence then
                    local idx = config.sequence_index or 1
                    local value = config.values[idx]
                    config.sequence_index = idx + 1
                    if config.sequence_index > #config.values then
                        config.sequence_index = 1
                    end
                    return value
                else
                    return config.value
                end
            end
        end
    end

    -- Mock control methods
    local control
    control = {
        name = name,
        calls = calls,

        returns = function(self, value) -- luacheck: ignore 212 (unused argument)
            table.insert(returns, { value = value })
            return control
        end,

        returns_sequence = function(self, ...) -- luacheck: ignore 212 (unused argument)
            local values = { ... }
            table.insert(returns, {
                sequence = true,
                values = values,
                sequence_index = 1,
            })
            return control
        end,

        throws = function(self, error_msg) -- luacheck: ignore 212 (unused argument)
            throws = error_msg
            return control
        end,

        calls_through = function(self, func) -- luacheck: ignore 212 (unused argument)
            call_through = true
            original_func = func
            return control
        end,

        with_args = function(self, ...) -- luacheck: ignore 212 (unused argument)
            local args = { ... }
            local config = { args = args }

            local methods = {
                returns = function(_, value)
                    config.value = value
                    table.insert(returns, config)
                    return control
                end,

                throws = function(_, error_msg)
                    config.throws = error_msg
                    table.insert(returns, config)
                    return control
                end,
            }

            return methods
        end,

        called = function() -- removed self parameter
            return #calls > 0
        end,

        called_times = function(n) -- removed self parameter
            return #calls == n
        end,

        called_with = function(...) -- removed self parameter
            local expected_args = { ... }
            for _, call_args in ipairs(calls) do
                local match = true
                for i, arg in ipairs(expected_args) do
                    if not equals(arg, call_args[i]) then
                        match = false
                        break
                    end
                end
                if match then
                    return true
                end
            end
            return false
        end,

        get_calls = function() -- removed self parameter
            return deep_copy(calls)
        end,

        reset = function() -- removed self parameter
            calls = {}
            control.calls = calls
            return control
        end,
    }

    -- Store mock for cleanup
    mocks[mock_func] = control

    return mock_func, control
end

-- Create a mock function
function testing.mock.func(name)
    name = name or "mock_func"
    local mock_func, _ = create_mock_func(name)
    return mock_func
end

-- Create a mock object
function testing.mock.create(name)
    name = name or "mock_object"
    local mock_obj = { _name = name }
    local mock_controls = {}

    local mt = {
        __index = function(t, k)
            if not mock_controls[k] then
                local mock_func, control = create_mock_func(name .. "." .. k)
                mock_controls[k] = control
                rawset(t, k, mock_func)
            end
            return rawget(t, k)
        end,
    }

    setmetatable(mock_obj, mt)

    -- Add control methods
    mock_obj._get_mock_control = function(method)
        return mock_controls[method]
    end

    mock_obj._verify_all = function()
        for _, _ in pairs(mock_controls) do
            -- Verify expectations if any
        end
    end

    return mock_obj
end

-- Create mock object with predefined methods
function testing.mock.object(definition)
    validate_required(definition, "mock definition")

    if type(definition) ~= "table" then
        error("mock definition must be a table")
    end

    local mock_obj = {}

    for k, v in pairs(definition) do
        if type(v) == "function" then
            mock_obj[k] = v
        else
            mock_obj[k] = testing.mock.func(k)
        end
    end

    return mock_obj
end

-- Stub a function
function testing.stub(original_func, replacement)
    validate_required(original_func, "original function")
    validate_required(replacement, "replacement")

    local stub_info = {
        original = original_func,
        replacement = replacement,
        restore = function()
            -- Restore would need access to where the function is stored
            -- This is simplified - in real implementation would need more context
        end,
    }

    table.insert(stubs, stub_info)
    return stub_info
end

-- Create a spy
function testing.spy(func)
    validate_required(func, "function to spy on")

    if type(func) ~= "function" then
        error("spy requires a function")
    end

    local calls = {}

    local spy_obj = {
        _is_spy = true,
        _func = func,
        _calls = calls,
    }

    local spy_control = {
        called = function()
            return #calls > 0
        end,

        called_times = function(n)
            return #calls == n
        end,

        called_with = function(...)
            local expected_args = { ... }
            for _, call in ipairs(calls) do
                local match = true
                for i, arg in ipairs(expected_args) do
                    if not equals(arg, call.args[i]) then
                        match = false
                        break
                    end
                end
                if match then
                    return true
                end
            end
            return false
        end,

        get_calls = function()
            return deep_copy(calls)
        end,

        reset = function()
            calls = {}
        end,
    }

    -- Add spy control to the object
    spy_obj.spy = spy_control

    -- Make it callable
    local mt = {
        __call = function(_, ...)
            local args = { ... }
            local result = { func(...) }
            table.insert(calls, { args = args, result = result })
            return unpack(result) -- luacheck: ignore 113 (unpack is global in Lua 5.1)
        end,
    }
    setmetatable(spy_obj, mt)

    -- Store spy for cleanup
    spies[spy_obj] = spy_control

    return spy_obj
end

-- Restore all mocks
function testing.mock.restore_all()
    mocks = {}
    spies = {}
    stubs = {}
end

-- Performance Testing

-- Benchmark a function
function testing.benchmark(name, func, options)
    validate_required(name, "benchmark name")
    validate_required(func, "function to benchmark")

    if type(func) ~= "function" then
        error("benchmark requires a function")
    end

    options = options or {}
    local iterations = options.iterations or 1000
    local warmup = options.warmup or 10

    -- Warmup
    for _ = 1, warmup do
        func()
    end

    -- Timing
    local times = {}
    local start_time = os.clock()

    for _ = 1, iterations do
        local iter_start = os.clock()
        func()
        local iter_time = os.clock() - iter_start
        table.insert(times, iter_time)
    end

    local total_time = os.clock() - start_time

    -- Calculate statistics
    local min_time = times[1]
    local max_time = times[1]
    local sum_time = 0

    for _, time in ipairs(times) do
        sum_time = sum_time + time
        if time < min_time then
            min_time = time
        end
        if time > max_time then
            max_time = time
        end
    end

    local avg_time = sum_time / iterations

    return {
        name = name,
        iterations = iterations,
        total_time = total_time,
        avg_time = avg_time,
        min_time = min_time,
        max_time = max_time,
        ops_per_sec = 1 / avg_time,
    }
end

-- Load test a function
function testing.load_test(name, func, config)
    validate_required(name, "load test name")
    validate_required(func, "function to test")
    validate_required(config, "load test config")

    if type(func) ~= "function" then
        error("load test requires a function")
    end

    -- Simplified load test - real implementation would use coroutines
    -- local concurrent = config.concurrent or 10
    local duration_str = config.duration or "10s"
    local iterations = config.iterations or 1000

    -- Parse duration string (simplified - assumes seconds)
    local duration_num = tonumber(string.match(duration_str, "%d+")) or 10

    local results = {
        name = name,
        requests = 0,
        errors = 0,
        latencies = {},
    }

    -- Simple sequential execution for now
    for _ = 1, iterations do
        local start = os.clock()
        local success, _ = pcall(func)
        local latency = (os.clock() - start) * 1000 -- Convert to ms

        results.requests = results.requests + 1
        if not success then
            results.errors = results.errors + 1
        end
        table.insert(results.latencies, latency)
    end

    -- Calculate percentiles
    table.sort(results.latencies)
    local p50_idx = math.floor(#results.latencies * 0.5)
    local p95_idx = math.floor(#results.latencies * 0.95)
    local p99_idx = math.floor(#results.latencies * 0.99)

    results.p50_latency = results.latencies[p50_idx] or 0
    results.p95_latency = results.latencies[p95_idx] or 0
    results.p99_latency = results.latencies[p99_idx] or 0
    results.requests_per_sec = results.requests / duration_num

    return results
end

-- Memory test a function
function testing.memory_test(func, options) -- luacheck: ignore 212 (unused argument)
    validate_required(func, "function to test")

    if type(func) ~= "function" then
        error("memory test requires a function")
    end

    -- options = options or {} -- for future use

    -- Get initial memory (simplified - Lua doesn't have direct memory access)
    collectgarbage("collect")
    local initial_memory = collectgarbage("count")

    -- Run function
    func()

    -- Get peak memory
    local peak_memory = collectgarbage("count")

    -- Force garbage collection
    collectgarbage("collect")
    local final_memory = collectgarbage("count")

    return {
        initial_memory = initial_memory or 0,
        peak_memory = peak_memory or 0,
        final_memory = final_memory or 0,
        leaked = (final_memory or 0) > (initial_memory or 0),
        allocations = (peak_memory or 0) - (initial_memory or 0),
    }
end

-- Test Runner

-- Run a single test
local function run_test(test)
    current_test = test
    test_results.total = test_results.total + 1

    local test_result = {
        name = test.name,
        suite = test.suite.name,
        status = "pending",
        error = nil,
        duration = 0,
    }

    if test.skip then
        test_result.status = "skipped"
        test_results.skipped = test_results.skipped + 1
        table.insert(test_results.tests, test_result)
        return
    end

    -- Run before_each hooks
    for _, hook in ipairs(test.suite.before_each) do
        local success, err = pcall(hook)
        if not success then
            test_result.status = "failed"
            test_result.error = "before_each failed: " .. tostring(err)
            test_results.failed = test_results.failed + 1
            table.insert(test_results.tests, test_result)
            return
        end
    end

    -- Run the test
    local start_time = os.clock()
    local success, err = pcall(test.func)
    test_result.duration = os.clock() - start_time

    if success then
        test_result.status = "passed"
        test_results.passed = test_results.passed + 1
    else
        test_result.status = "failed"
        test_result.error = tostring(err)
        test_results.failed = test_results.failed + 1
    end

    -- Run after_each hooks
    for _, hook in ipairs(test.suite.after_each) do
        pcall(hook) -- Don't fail test if cleanup fails
    end

    -- Cleanup mocks
    testing.mock.restore_all()

    table.insert(test_results.tests, test_result)
    current_test = nil
end

-- Run a test suite
local function run_suite(suite)
    -- Run before_all hooks
    for _, hook in ipairs(suite.before_all) do
        local success, err = pcall(hook)
        if not success then
            print("before_all failed for suite '" .. suite.name .. "': " .. tostring(err))
            return
        end
    end

    -- Check if any tests have "only"
    local has_only = false
    for _, test in ipairs(suite.tests) do
        if test.only then
            has_only = true
            break
        end
    end

    -- Run tests
    for _, test in ipairs(suite.tests) do
        if not has_only or test.only then
            run_test(test)
        elseif has_only and not test.only then
            -- Count non-only tests as skipped when there are only tests
            test_results.total = test_results.total + 1
            test_results.skipped = test_results.skipped + 1
            table.insert(test_results.tests, {
                name = test.name,
                suite = test.suite.name,
                status = "skipped",
                error = nil,
                duration = 0,
            })
        end
    end

    -- Run nested suites
    for _, nested in ipairs(suite.nested_suites) do
        run_suite(nested)
    end

    -- Run after_all hooks
    for _, hook in ipairs(suite.after_all) do
        pcall(hook) -- Don't fail if cleanup fails
    end
end

-- Run all tests
function testing.run(options)
    options = options or {}

    -- Reset results
    test_results = {
        total = 0,
        passed = 0,
        failed = 0,
        skipped = 0,
        tests = {},
        start_time = os.clock(),
        end_time = 0,
    }

    -- Run all suites
    for _, suite in ipairs(test_suites) do
        run_suite(suite)
    end

    test_results.end_time = os.clock()
    test_results.duration = test_results.end_time - test_results.start_time

    -- Print summary (simple console output)
    if not options.quiet then
        print("\nTest Results:")
        print(string.format("  Total:   %d", test_results.total))
        print(string.format("  Passed:  %d", test_results.passed))
        print(string.format("  Failed:  %d", test_results.failed))
        print(string.format("  Skipped: %d", test_results.skipped))
        print(string.format("  Time:    %.3fs\n", test_results.duration))

        -- Print failures
        if test_results.failed > 0 then
            print("Failures:")
            for _, test in ipairs(test_results.tests) do
                if test.status == "failed" then
                    print(string.format("  %s > %s", test.suite, test.name))
                    print("    " .. test.error)
                end
            end
        end
    end

    return test_results
end

-- Get test results
function testing.get_results()
    return deep_copy(test_results)
end

-- Reset testing state
function testing.reset()
    test_suites = {}
    current_suite = nil
    current_test = nil
    test_results = {
        total = 0,
        passed = 0,
        failed = 0,
        skipped = 0,
        tests = {},
        start_time = 0,
        end_time = 0,
    }
    testing.mock.restore_all()
end

-- Test Data Generation
testing.data = {}

function testing.data.random_string(length)
    length = length or 10
    local chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    local str = ""
    for _ = 1, length do
        local idx = math.random(1, #chars)
        str = str .. string.sub(chars, idx, idx)
    end
    return str
end

function testing.data.random_number(min, max)
    min = min or 0
    max = max or 100
    return math.random(min, max)
end

function testing.data.random_table(depth, size)
    depth = depth or 2
    size = size or 5

    if depth <= 0 then
        return testing.data.random_string(10)
    end

    local tbl = {}
    for i = 1, size do
        local key = "key" .. i
        if math.random() > 0.5 then
            tbl[key] = testing.data.random_table(depth - 1, size)
        else
            tbl[key] = testing.data.random_string(10)
        end
    end

    return tbl
end

function testing.data.uuid()
    -- Simple UUID v4 generation
    local template = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
    return string.gsub(template, "[xy]", function(c)
        local v = (c == "x") and math.random(0, 0xf) or math.random(8, 0xb)
        return string.format("%x", v)
    end)
end

function testing.data.sample(array, count)
    validate_required(array, "array")
    count = count or 1

    if type(array) ~= "table" then
        error("sample requires an array")
    end

    local sampled = {}
    local indices = {}

    -- Generate unique random indices
    while #sampled < count and #sampled < #array do
        local idx = math.random(1, #array)
        if not indices[idx] then
            indices[idx] = true
            table.insert(sampled, array[idx])
        end
    end

    return sampled
end

-- Validation utilities (simplified - would integrate with validation bridge)
testing.validate = {}

function testing.validate.schema(data, schema)
    -- This would integrate with the schema validation bridge
    -- For now, just basic validation
    if schema.type == "object" and type(data) ~= "table" then
        return false, "expected object"
    end
    if schema.type == "string" and type(data) ~= "string" then
        return false, "expected string"
    end
    if schema.type == "number" and type(data) ~= "number" then
        return false, "expected number"
    end
    return true
end

-- Export the module
return testing
