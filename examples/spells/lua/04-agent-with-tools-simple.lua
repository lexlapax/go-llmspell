-- ABOUTME: Simplified agent with tools example demonstrating tool-enabled agents
-- ABOUTME: Shows how agents can be given tool access for autonomous task completion

-- Required modules
local agent = require("agent")

-- Agent With Tools Example
-- This spell demonstrates creating and using tool-enabled agents

-- Parameters
local task = params.task or "Research Lua programming and summarize its key features"
local model = params.model or "gpt-3.5-turbo"

print("=== Agent With Tools Example (Simplified) ===")
print("Task: " .. task)
print("Model: " .. model)
print()

-- Example 1: Basic Tool-Enabled Agent
print("=== Example 1: Basic Tool-Enabled Agent ===")

-- Create a research assistant with tools
local researcher = agent.create("Research Assistant", {
    model = model,
    system = "You are a helpful research assistant with access to tools.",
    tools = {"web_search", "calculator"},  -- Specify which tools the agent can use
    temperature = 0.7
})

-- Run the research task
print("\nExecuting research task...")
local result = researcher:run(task)
print("\nResearch Result:")
print(tostring(result))
print()

-- Example 2: Multi-Agent Collaboration with Tools
print("=== Example 2: Multi-Agent Collaboration ===")

-- Create specialized agents with different tools
local data_analyst = agent.create("Data Analyst", {
    model = model,
    system = "You are a data analyst. Use calculations to analyze information.",
    tools = {"calculator"},
    temperature = 0.3
})

local writer = agent.create("Writer", {
    model = model,
    system = "You are a technical writer. Create clear, concise summaries.",
    temperature = 0.7
})

-- Collaborate on a task
print("\nAnalyst examining data...")
local analysis = data_analyst:run("Calculate the factorial of 7 and explain its significance")
print("Analysis: " .. tostring(analysis))

print("\nWriter creating summary...")
local summary = writer:run("Based on this analysis: " .. tostring(analysis) .. "\nCreate a brief summary.")
print("Summary: " .. tostring(summary))

-- Return results
return {
    success = true,
    agents_created = {"researcher", "data_analyst", "writer"},
    tools_demonstrated = {"web_search", "calculator"},
    task_completed = true
}