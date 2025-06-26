-- ABOUTME: Test if print statements work
-- ABOUTME: Simple test to debug output issues

print("=== Testing Print Statements ===")
print("Line 1: Basic print")
print("Line 2: Number:", 42)
print("Line 3: Table:", {a = 1, b = 2})

-- Also test with different approaches
io.write("Line 4: Using io.write\n")
io.stdout:write("Line 5: Using io.stdout:write\n")

-- Return something to see if that shows
return {
    test = "complete",
    printed_lines = 5
}