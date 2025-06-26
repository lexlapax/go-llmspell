-- ABOUTME: Test module requires one by one
-- ABOUTME: Find which require is causing the error

print("Testing module requires...")

print("\n1. Requiring state...")
local ok, state = pcall(require, "state")
print("state require:", ok)

print("\n2. Requiring agent...")  
ok, agent = pcall(require, "agent")
print("agent require:", ok)

print("\n3. Requiring data...")
ok, data = pcall(require, "data")
print("data require:", ok)

print("\n4. Requiring utils...")
ok, utils = pcall(require, "utils")
print("utils require:", ok)

print("\nAll requires completed")

return {
    all_ok = true
}