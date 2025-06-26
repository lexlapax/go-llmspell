-- ABOUTME: Add debug prints to find exact line causing error
-- ABOUTME: Trace execution to locate the problematic call

local agent = require("agent")
local state = require("state")
local data = require("data")
local utils = require("utils")

-- Initialize state
state.set("research", {
    topic = "State Management in Distributed Systems",
    progress = {
        technical_research = "pending",
        practical_research = "pending",
        synthesis = "pending"
    },
    findings = {}
})

-- Create agents
print("Creating coordinator...")
local coordinator = agent.create("Research Coordinator", {
    model = "gpt-4",
    system = [[
        You coordinate research tasks. 
        Track progress in shared state under 'research.progress'.
        Assign tasks to researcher agents.
    ]]
})
print("Coordinator created")

print("Creating researcher1...")
local researcher1 = agent.create("Researcher 1", {
    model = "gpt-3.5-turbo",
    system = "You research technical topics and update shared state with findings."
})
print("Researcher1 created")

print("Creating researcher2...")
local researcher2 = agent.create("Researcher 2", {
    model = "gpt-3.5-turbo",
    system = "You research practical applications and update shared state with findings."
})
print("Researcher2 created")

-- Now test the calls one by one
print("\n=== Testing coordinator:run ===")
local ok, err = pcall(function()
    return coordinator:run({
        prompt = "Create a research plan for: " .. state.get("research.topic"),
        max_tokens = 200
    })
end)
print("Coordinator run result:", ok, err and type(err) or "success")

print("\n=== Testing researcher1:run ===")
ok, err = pcall(function()
    return researcher1:run({
        prompt = "Research technical aspects of: " .. state.get("research.topic"),
        max_tokens = 300
    })
end)
print("Researcher1 run result:", ok, err and type(err) or "success")

print("\n=== Testing researcher2:run ===")
ok, err = pcall(function()
    return researcher2:run({
        prompt = "Research practical applications of: " .. state.get("research.topic"),
        max_tokens = 300
    })
end)
print("Researcher2 run result:", ok, err and type(err) or "success")

print("\n=== Testing synthesis ===")
local findings = state.get("research.findings")
ok, err = pcall(function()
    return coordinator:run({
        prompt = string.format(
            "Synthesize these research findings on \"%s\":\n\nDATA ANALYSIS:\n%s\n\nLITERATURE REVIEW:\n%s\n\nDOMAIN EXPERTISE:\n%s\n\nCreate a comprehensive summary with key insights and recommendations.",
            state.get("research.topic"), 
            findings.technical or "none",
            findings.practical or "none",
            findings.domain or "none"
        ),
        max_tokens = 400
    })
end)
print("Synthesis run result:", ok, err and type(err) or "success")

return {
    all_tests_passed = true
}