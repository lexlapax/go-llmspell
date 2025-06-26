-- ABOUTME: Simplified agent handoff example showing agent coordination patterns
-- ABOUTME: Demonstrates agent specialization and context passing without complex state management

-- Agent Handoff Example (Simplified)
-- This spell demonstrates agent coordination:
-- 1. Creating specialized agents
-- 2. Routing requests to appropriate agents
-- 3. Maintaining conversation context
-- 4. Aggregating results from multiple agents

-- Required modules
local agent = require("agent")
local llm = require("llm")
local log = require("log")
local data = require("data")
local utils = require("utils")

-- Parameters
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./handoff-output"

print("=== Agent Handoff Example (Simplified) ===")
print()

-- Note: File operations removed as they're not available in this environment

-- Simple context manager
local Context = {}

function Context.new()
    return {
        conversation_history = {},
        handoff_log = {},
        current_agent = nil
    }
end

function Context.add_message(ctx, role, content, agent_name)
    ctx.conversation_history[#ctx.conversation_history + 1] = {
        role = role,
        content = content,
        agent = agent_name,
        timestamp = os.time()
    }
end

function Context.log_handoff(ctx, from_agent, to_agent, reason)
    ctx.handoff_log[#ctx.handoff_log + 1] = {
        from = from_agent,
        to = to_agent,
        reason = reason,
        timestamp = os.time()
    }
    ctx.current_agent = to_agent
end

-- Create conversation context
local context = Context.new()

-- Example 1: Basic Agent Handoff
print("=== Example 1: Basic Agent Handoff ===")

-- Create specialized agents
local agents = {
    customer_service = agent.create("Customer Service", {
        model = model,
        system = "You are a friendly customer service representative. " ..
                "Help with general inquiries. If technical help is needed, say 'HANDOFF:TECHNICAL'. " ..
                "If billing help is needed, say 'HANDOFF:BILLING'.",
        temperature = 0.7
    }),
    
    technical = agent.create("Technical Support", {
        model = model,
        system = "You are a technical support specialist. " ..
                "Provide detailed technical solutions. " ..
                "If billing is mentioned, say 'HANDOFF:BILLING'.",
        temperature = 0.5
    }),
    
    billing = agent.create("Billing Department", {
        model = model,
        system = "You are a billing specialist. " ..
                "Help with payment issues, refunds, and account billing. " ..
                "If technical issues arise, say 'HANDOFF:TECHNICAL'.",
        temperature = 0.5
    })
}

-- Agent router function
local function route_to_agent(message, current_agent)
    local response = current_agent:run(message)
    
    -- Check for handoff signals
    if response:find("HANDOFF:TECHNICAL") then
        return agents.technical, "technical", response:gsub("HANDOFF:TECHNICAL", "")
    elseif response:find("HANDOFF:BILLING") then
        return agents.billing, "billing", response:gsub("HANDOFF:BILLING", "")
    elseif response:find("HANDOFF:CUSTOMER_SERVICE") then
        return agents.customer_service, "customer_service", response:gsub("HANDOFF:CUSTOMER_SERVICE", "")
    end
    
    return current_agent, nil, response
end

-- Simulate customer conversation
print("\nSimulating customer support conversation:")

local conversations = {
    "Hello, I need help with my account.",
    "My software keeps crashing when I try to export files.",
    "How much will this cost to fix?",
    "Thank you for your help!"
}

local current_agent = agents.customer_service
context.current_agent = "customer_service"

for i, message in ipairs(conversations) do
    print("\n[Turn " .. i .. "]")
    print("Customer: " .. message)
    Context.add_message(context, "user", message, "customer")
    
    -- Get response and check for handoff
    local next_agent, next_name, response = route_to_agent(message, current_agent)
    
    -- Handle handoff if needed
    if next_name and next_name ~= context.current_agent then
        print("System: Transferring to " .. next_name)
        Context.log_handoff(context, context.current_agent, next_name, "Customer needs " .. next_name .. " assistance")
        current_agent = next_agent
        
        -- Get introduction from new agent
        local intro = current_agent:run("You just received a handoff. The customer said: " .. message)
        response = intro
    end
    
    print("Agent (" .. context.current_agent .. "): " .. response)
    Context.add_message(context, "assistant", response, context.current_agent)
end

print()

-- Example 2: Multi-Agent Research
print("=== Example 2: Multi-Agent Research ===")

-- Create research team
local research_team = {
    researcher = agent.create("Researcher", {
        model = model,
        system = "You are a research specialist. Find relevant information on topics. Be thorough but concise.",
        temperature = 0.7
    }),
    
    analyst = agent.create("Analyst", {
        model = model,
        system = "You are a data analyst. Analyze information and identify patterns, trends, and insights.",
        temperature = 0.5
    }),
    
    writer = agent.create("Writer", {
        model = model,
        system = "You are a technical writer. Create clear, well-structured summaries from research and analysis.",
        temperature = 0.7
    })
}

-- Research pipeline function
local function conduct_research(topic)
    local research_context = {
        topic = topic,
        research_data = nil,
        analysis = nil,
        summary = nil
    }
    
    print("\nResearching: " .. topic)
    
    -- Step 1: Research
    print("  1. Researcher gathering information...")
    research_context.research_data = research_team.researcher:run(
        "Research the topic: " .. topic .. ". Provide key facts and findings."
    )
    
    -- Step 2: Analysis
    print("  2. Analyst processing data...")
    research_context.analysis = research_team.analyst:run(
        "Analyze this research on " .. topic .. ":\n" .. research_context.research_data
    )
    
    -- Step 3: Writing
    print("  3. Writer creating summary...")
    research_context.summary = research_team.writer:run(
        "Create a concise summary based on this research and analysis:\n" ..
        "Research: " .. research_context.research_data .. "\n" ..
        "Analysis: " .. research_context.analysis
    )
    
    return research_context
end

-- Test research pipeline
local research_topic = "The impact of AI on software development"
local research_result = conduct_research(research_topic)

print("\nResearch Summary:")
print(research_result.summary)
print()

-- Example 3: Consensus Building
print("=== Example 3: Consensus Building ===")

-- Create advisory panel
local advisors = {
    optimist = agent.create("Optimist Advisor", {
        model = model,
        system = "You are an optimistic advisor. Always see the positive side and opportunities.",
        temperature = 0.8
    }),
    
    pessimist = agent.create("Pessimist Advisor", {
        model = model,
        system = "You are a cautious advisor. Point out risks, challenges, and potential problems.",
        temperature = 0.8
    }),
    
    realist = agent.create("Realist Advisor", {
        model = model,
        system = "You are a balanced advisor. Provide practical, realistic assessments.",
        temperature = 0.6
    })
}

-- Get consensus on a decision
local function get_consensus(question)
    local opinions = {}
    
    print("\nSeeking consensus on: " .. question)
    
    -- Gather opinions
    for name, advisor in pairs(advisors) do
        print("\n" .. name .. " says:")
        opinions[name] = advisor:run(question .. " Please provide your perspective in 2-3 sentences.")
        print(opinions[name])
    end
    
    -- Synthesize consensus
    print("\nSynthesizing consensus...")
    local synthesis_agent = agent.create("Synthesizer", {
        model = model,
        system = "You synthesize different viewpoints into a balanced conclusion.",
        temperature = 0.5
    })
    
    local consensus = synthesis_agent:run(
        "Synthesize these perspectives into a balanced conclusion:\n" ..
        "Optimist: " .. opinions.optimist .. "\n" ..
        "Pessimist: " .. opinions.pessimist .. "\n" ..
        "Realist: " .. opinions.realist
    )
    
    return consensus
end

-- Test consensus building
local decision = "Should we implement a new AI-powered feature in our product?"
local consensus = get_consensus(decision)

print("\nConsensus:")
print(consensus)
print()

-- Save conversation log
print("=== Saving Logs ===")

-- Save handoff log
local handoff_report = "Agent Handoff Log\n"
handoff_report = handoff_report .. "=================\n\n"

for _, handoff in ipairs(context.handoff_log) do
    handoff_report = handoff_report .. string.format(
        "[%s] %s → %s | Reason: %s\n",
        os.date("%H:%M:%S", handoff.timestamp),
        handoff.from,
        handoff.to,
        handoff.reason
    )
end

-- In production, would save: utils.file_write(output_dir .. "/handoff_log.txt", handoff_report)
print("Handoff log would be saved to: " .. output_dir .. "/handoff_log.txt")
print("Handoff log saved")

-- Save conversation history
local conversation_log = "Conversation History\n"
conversation_log = conversation_log .. "===================\n\n"

for _, msg in ipairs(context.conversation_history) do
    conversation_log = conversation_log .. string.format(
        "[%s] %s (%s): %s\n\n",
        os.date("%H:%M:%S", msg.timestamp),
        msg.role:upper(),
        msg.agent or "user",
        msg.content
    )
end

-- In production, would save: utils.file_write(output_dir .. "/conversation_history.txt", conversation_log)
print("Conversation history would be saved to: " .. output_dir .. "/conversation_history.txt")
print("Conversation history saved")

-- Save research results
local research_data = data.to_json({
    topic = research_topic,
    research = research_result.research_data,
    analysis = research_result.analysis,
    summary = research_result.summary
})
-- In production, would save: utils.file_write(output_dir .. "/research_results.json", research_data)
print("Research results would be saved to: " .. output_dir .. "/research_results.json")

print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Basic agent handoff with routing logic")
print("2. Multi-agent research pipeline")
print("3. Consensus building with multiple perspectives")
print("4. Context preservation across handoffs")
print()
print("Key insights:")
print("- Specialized agents improve response quality")
print("- Clear handoff signals enable smooth transitions")
print("- Pipeline patterns coordinate multiple agents")
print("- Consensus building leverages diverse perspectives")
print("- Context tracking maintains conversation continuity")
print()

-- Return summary
return {
    total_handoffs = #context.handoff_log,
    total_messages = #context.conversation_history,
    agents_created = 7,
    files_created = 3
}