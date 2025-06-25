# Lua Testing & Validation Library Design

## Overview
A comprehensive testing and validation library for Lua scripts in go-llmspell that provides:
- Test framework with assertions
- Mocking and stubbing utilities
- Performance and load testing
- Integration with existing validation infrastructure

## Core API Design

### 1. Test Framework and Organization
```lua
-- Test suite definition
testing.describe(name, function()
    -- Setup/teardown hooks
    testing.before_each(setup_func)
    testing.after_each(teardown_func)
    testing.before_all(setup_func)
    testing.after_all(teardown_func)
    
    -- Individual tests
    testing.it(test_name, test_func)
    testing.test(test_name, test_func)  -- alias
    
    -- Nested test groups
    testing.describe(nested_name, nested_tests)
    
    -- Skip/focus tests
    testing.skip(test_name, test_func)
    testing.only(test_name, test_func)
end)

-- Run tests
testing.run(options)
testing.run_file(file_path)
testing.run_suite(suite_name)
```

### 2. Assertions
```lua
-- Basic assertions
testing.assert.equals(actual, expected, message)
testing.assert.not_equals(actual, expected, message)
testing.assert.true(value, message)
testing.assert.false(value, message)
testing.assert.nil(value, message)
testing.assert.not_nil(value, message)

-- Type assertions
testing.assert.type(value, expected_type, message)
testing.assert.table(value, message)
testing.assert.function(value, message)
testing.assert.string(value, message)
testing.assert.number(value, message)
testing.assert.boolean(value, message)

-- Comparison assertions
testing.assert.greater_than(actual, expected, message)
testing.assert.less_than(actual, expected, message)
testing.assert.greater_or_equal(actual, expected, message)
testing.assert.less_or_equal(actual, expected, message)
testing.assert.between(value, min, max, message)

-- String assertions
testing.assert.contains(haystack, needle, message)
testing.assert.matches(string, pattern, message)
testing.assert.starts_with(string, prefix, message)
testing.assert.ends_with(string, suffix, message)

-- Table assertions
testing.assert.deep_equals(actual, expected, message)
testing.assert.has_key(table, key, message)
testing.assert.has_value(table, value, message)
testing.assert.length(table, expected_length, message)
testing.assert.empty(table, message)
testing.assert.not_empty(table, message)

-- Error assertions
testing.assert.error(func, expected_error, message)
testing.assert.no_error(func, message)
testing.assert.error_matches(func, pattern, message)

-- Async assertions
testing.assert.async(promise, timeout)
testing.assert.resolves(promise, expected_value, timeout)
testing.assert.rejects(promise, expected_error, timeout)
```

### 3. Mocking and Stubbing
```lua
-- Mock creation
local mock = testing.mock.create(name)
local mock = testing.mock.object({
    method1 = testing.mock.func(),
    method2 = testing.mock.func()
})

-- Mock configuration
mock:returns(value)
mock:returns_sequence(value1, value2, ...)
mock:throws(error)
mock:calls_through()
mock:with_args(arg1, arg2):returns(value)

-- Stub functions
local stub = testing.stub(original_func, replacement)
stub:restore()

-- Spy on functions
local spy = testing.spy(func)
spy:called()
spy:called_times(n)
spy:called_with(arg1, arg2)
spy:get_calls()
spy:reset()

-- Mock expectations
mock:expect_call(method, args)
mock:verify()
mock:verify_all()

-- Global mocks
testing.mock.global(name, value)
testing.mock.restore_all()
```

### 4. Performance Testing
```lua
-- Benchmarking
local result = testing.benchmark(name, func, options)
result.iterations
result.total_time
result.avg_time
result.min_time
result.max_time
result.ops_per_sec

-- Options
{
    iterations = 1000,
    warmup = 10,
    timeout = 30,
    measure_memory = true
}

-- Load testing
testing.load_test(name, func, config)
{
    concurrent = 10,
    duration = "30s",
    ramp_up = "5s",
    target_rps = 100
}

-- Memory testing
local mem_result = testing.memory_test(func, options)
mem_result.initial_memory
mem_result.peak_memory
mem_result.final_memory
mem_result.leaked
mem_result.allocations
```

### 5. Test Utilities
```lua
-- Test data generation
testing.data.random_string(length)
testing.data.random_number(min, max)
testing.data.random_table(depth, size)
testing.data.uuid()
testing.data.sample(array, count)

-- Test fixtures
testing.fixture.create(name, data)
testing.fixture.load(name)
testing.fixture.save(name, data)
testing.fixture.clear()

-- Test doubles
testing.double.llm_response(template)
testing.double.api_response(status, body)
testing.double.error_response(code, message)

-- Time control
testing.time.freeze(timestamp)
testing.time.advance(seconds)
testing.time.restore()

-- Environment control
testing.env.set(key, value)
testing.env.get(key)
testing.env.snapshot()
testing.env.restore()
```

### 6. Test Reporting
```lua
-- Report configuration
testing.reporter.set("json")  -- json, tap, junit, console
testing.reporter.output("/path/to/report")

-- Custom reporters
testing.reporter.register("custom", {
    on_suite_start = function(suite) end,
    on_test_start = function(test) end,
    on_test_pass = function(test) end,
    on_test_fail = function(test, error) end,
    on_suite_end = function(suite, results) end
})

-- Test results
local results = testing.get_results()
results.total
results.passed
results.failed
results.skipped
results.duration
results.tests -- array of test results
```

### 7. Validation Integration
```lua
-- Schema validation (bridge to go-llms)
testing.validate.schema(data, schema)
testing.validate.json_schema(data, schema_path)

-- Custom validators
testing.validate.register("email", function(value)
    return string.match(value, "^[%w._%+-]+@[%w.-]+%.[%w]+$") ~= nil
end)

testing.validate.email(value)
testing.validate.url(value)
testing.validate.uuid(value)

-- Validation chains
testing.validate.chain()
    :required()
    :string()
    :min_length(5)
    :max_length(50)
    :matches("^[a-zA-Z]+$")
    :validate(value)
```

### 8. Async Testing
```lua
-- Async test support
testing.it("async test", async(function()
    local result = await(some_async_operation())
    testing.assert.equals(result, expected)
end))

-- Promise testing
testing.it("promise test", function()
    local promise = create_promise()
    return testing.assert.resolves(promise, expected_value)
end)

-- Timeout control
testing.it("timeout test", function()
    -- Test timeout
end, {timeout = 5000})
```

### 9. Test Helpers
```lua
-- Setup/teardown helpers
testing.helpers.create_test_context()
testing.helpers.create_test_logger()
testing.helpers.create_test_state()

-- Cleanup tracking
testing.helpers.cleanup(func)
testing.helpers.cleanup_all()

-- Error capture
testing.helpers.capture_errors(func)
testing.helpers.suppress_errors(func)

-- Output capture
testing.helpers.capture_output(func)
testing.helpers.capture_logs(func)
```

## Implementation Notes

1. **Bridge Integration**: Integrate with existing validation bridge
2. **Async Support**: Use promise library for async test support
3. **Performance**: Minimize overhead in test execution
4. **Isolation**: Ensure test isolation and cleanup
5. **Compatibility**: Work with existing error and logging libraries

## Example Usage

```lua
local testing = require("testing")

testing.describe("User Authentication", function()
    local mock_db
    
    testing.before_each(function()
        mock_db = testing.mock.object({
            get_user = testing.mock.func(),
            save_user = testing.mock.func()
        })
    end)
    
    testing.after_each(function()
        testing.mock.restore_all()
    end)
    
    testing.it("should authenticate valid user", function()
        -- Arrange
        mock_db.get_user:returns({
            id = "123",
            username = "testuser",
            password_hash = hash("password")
        })
        
        -- Act
        local result = auth.login("testuser", "password", mock_db)
        
        -- Assert
        testing.assert.true(result.success)
        testing.assert.equals(result.user_id, "123")
        testing.assert.called(mock_db.get_user, 1)
    end)
    
    testing.it("should reject invalid password", function()
        mock_db.get_user:returns({
            id = "123",
            username = "testuser", 
            password_hash = hash("correct_password")
        })
        
        testing.assert.error(function()
            auth.login("testuser", "wrong_password", mock_db)
        end, "Invalid credentials")
    end)
    
    testing.describe("Performance", function()
        testing.it("should handle concurrent logins", function()
            local result = testing.load_test("login", function()
                return auth.login("user", "pass", mock_db)
            end, {
                concurrent = 100,
                duration = "10s"
            })
            
            testing.assert.greater_than(result.requests_per_sec, 1000)
            testing.assert.less_than(result.p95_latency, 50)
        end)
    end)
end)

-- Run tests
testing.run()
```