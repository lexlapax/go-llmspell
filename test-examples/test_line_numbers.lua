-- ABOUTME: Test if line numbers are preserved in errors
-- ABOUTME: Add deliberate errors at known lines

print("Line 4: Starting test")

-- Line 6: This should work
local x = 5 + 3
print("Line 8: x =", x)

-- Line 10: This will cause an error - calling a number
local y = 42
print("Line 12: About to cause error on next line")
y()  -- Line 13: This should error

print("Line 15: This won't be reached")

return {unreachable = true}