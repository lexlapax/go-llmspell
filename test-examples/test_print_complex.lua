-- Test print with more complex scenarios
print("=== Testing Print Output ===")

-- Test with tables
local data = {name = "test", value = 42}
print("Table:", data)

-- Test with functions
local function testFunc()
    return "test"
end
print("Function:", testFunc)

-- Test with nil
print("Nil value:", nil)

-- Test with boolean
print("Boolean:", true, false)

-- Test from within a function
local function innerPrint()
    print("  Print from inside function")
end
innerPrint()

-- Test capturing output (should still work for tests)
_test_output = {}
print("This should be captured AND printed")
print("Captured count:", #_test_output)

return "All tests complete"