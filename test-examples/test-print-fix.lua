-- Test print() output after fix
print("=== Testing Print Output ===")
print("Line 1: Basic string")
print("Line 2: Number:", 42)
print("Line 3: Multiple", "values", "with", "tabs")
print("Line 4: Boolean:", true, "Nil:", nil)

-- Also return a value to confirm execution
return {
    status = "success",
    message = "Print fix verified"
}