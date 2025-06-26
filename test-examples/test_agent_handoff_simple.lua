-- ABOUTME: Simplified agent handoff test
-- ABOUTME: Test basic agent creation and communication

local agent = require("agent")

print("=== Simple Agent Handoff Test ===")

-- Create two agents
print("\n1. Creating agents...")
local customer_agent = agent.create("Customer Service", {
    model = "gpt-3.5-turbo",
    system = "You are a friendly customer service agent. If asked about technical issues, say 'I need to transfer you to technical support.'"
})

local tech_agent = agent.create("Technical Support", {
    model = "gpt-3.5-turbo",
    system = "You are a technical support specialist. Help solve technical problems."
})

print("Created customer service agent:", customer_agent.name)
print("Created technical support agent:", tech_agent.name)

-- Test customer service agent
print("\n2. Testing customer service agent...")
local response1 = customer_agent:run("Hello, I need help")
print("Customer response:", type(response1) == "string" and response1 or (response1 and response1.content or "No response"))

-- Test with technical query
print("\n3. Testing with technical query...")
local response2 = customer_agent:run("My computer is crashing")
local response_text = type(response2) == "string" and response2 or (response2 and response2.content or "No response")
print("Customer response:", response_text)

-- Check if handoff is needed
if string.find(string.lower(response_text), "transfer") or string.find(string.lower(response_text), "technical support") then
    print("\n4. Handoff detected! Transferring to technical support...")
    
    local tech_response = tech_agent:run("Customer transferred: Computer is crashing")
    print("Tech response:", type(tech_response) == "string" and tech_response or (tech_response and tech_response.content or "No response"))
else
    print("\n4. No handoff needed")
end

print("\n=== Test Complete ===")

return {
    success = true,
    agents_created = 2
}