-- ABOUTME: Example demonstrating agent creation without tools, focusing on reasoning
-- ABOUTME: Shows agent personalities, chain of thought, and multi-agent conversations

-- Required modules
local agent = require("agent")

-- Agent Without Tools Example
-- This spell demonstrates creating and using agents without tool access:
-- 1. Basic agent creation
-- 2. Agent personalities and system prompts
-- 3. Chain of thought reasoning
-- 4. Multi-agent conversations
-- 5. Agent memory and context

-- Parameters
local topic = params.topic or "the future of artificial intelligence"
local model = params.model or "gpt-3.5-turbo"

print("=== Agent Without Tools Example ===")
print("Topic: " .. topic)
print("Model: " .. model)
print()

-- Example 1: Basic Agent Creation
print("=== Example 1: Basic Agent Creation ===")

-- Create a simple analyst agent
local analyst = agent.create({
    name = "Data Analyst",
    model = model,
    system = "You are a data analyst who provides clear, insightful analysis. Focus on facts and logical reasoning.",
    temperature = 0.3  -- Lower temperature for more focused responses
})

-- Use the agent
local analysis = analyst:run("Analyze the topic: " .. topic .. ". Provide 3 key insights.")
print("Analyst's response:")
print(analysis)
print()

-- Example 2: Agent with Chain of Thought
print("=== Example 2: Chain of Thought Agent ===")

-- Create an agent that uses chain of thought reasoning
local thinker = agent.create({
    name = "Deep Thinker",
    model = model,
    system = [[You are a careful thinker who uses step-by-step reasoning.
Always structure your response as:
1. Understanding: Restate the problem
2. Reasoning: Think through the problem step by step
3. Conclusion: Provide your final answer
Show your thinking process clearly.]],
    temperature = 0.5
})

local problem = "If all roses are flowers, and some flowers fade quickly, can we conclude that some roses fade quickly?"
local reasoning = thinker:run(problem)
print("Chain of Thought Response:")
print(reasoning)
print()

-- Example 3: Multiple Agents with Different Personalities
print("=== Example 3: Multiple Agent Personalities ===")

-- Create agents with different perspectives
local optimist = agent.create({
    name = "Optimist",
    model = model,
    system = "You are an optimistic futurist who sees the positive potential in everything. Focus on opportunities and benefits.",
    temperature = 0.7
})

local pessimist = agent.create({
    name = "Pessimist", 
    model = model,
    system = "You are a cautious analyst who identifies risks and potential problems. Focus on challenges and concerns.",
    temperature = 0.7
})

local realist = agent.create({
    name = "Realist",
    model = model,
    system = "You are a balanced analyst who considers both opportunities and challenges objectively. Provide nuanced perspectives.",
    temperature = 0.5
})

-- Get each agent's perspective on the topic
print("Three perspectives on: " .. topic)
print("\nOptimist's view:")
local optimist_view = optimist:run("What are your thoughts on " .. topic .. "? (Keep it to 3 sentences)")
print(optimist_view)

print("\nPessimist's view:")
local pessimist_view = pessimist:run("What are your thoughts on " .. topic .. "? (Keep it to 3 sentences)")
print(pessimist_view)

print("\nRealist's view:")
local realist_view = realist:run("What are your thoughts on " .. topic .. "? (Keep it to 3 sentences)")
print(realist_view)
print()

-- Example 4: Agent Conversation/Debate
print("=== Example 4: Agent Debate ===")

-- Create a moderator
local moderator = agent.create({
    name = "Moderator",
    model = model,
    system = "You are a debate moderator. Keep discussions focused and balanced. Summarize key points.",
    temperature = 0.3
})

-- Conduct a mini-debate
local debate_topic = "Should AI development be accelerated or slowed down?"
print("Debate topic: " .. debate_topic)

-- Opening statements
local optimist_opening = optimist:run("Give your opening statement on: " .. debate_topic .. " (2-3 sentences)")
local pessimist_opening = pessimist:run("Give your opening statement on: " .. debate_topic .. " (2-3 sentences)")

print("\nOptimist opening: " .. optimist_opening)
print("\nPessimist opening: " .. pessimist_opening)

-- Realist responds to both
local realist_response = realist:run(
    "After hearing these perspectives:\n" ..
    "Optimist: " .. optimist_opening .. "\n" ..
    "Pessimist: " .. pessimist_opening .. "\n" ..
    "What is your balanced view? (2-3 sentences)"
)
print("\nRealist response: " .. realist_response)

-- Moderator summary
local debate_summary = moderator:run(
    "Summarize this debate:\n" ..
    "Optimist: " .. optimist_opening .. "\n" ..
    "Pessimist: " .. pessimist_opening .. "\n" ..
    "Realist: " .. realist_response .. "\n" ..
    "Provide a balanced conclusion."
)
print("\nModerator's summary: " .. debate_summary)
print()

-- Example 5: Agent with Memory/Context
print("=== Example 5: Agent with Memory ===")

-- Create a teaching agent that remembers previous interactions
local teacher = agent.create({
    name = "Teacher",
    model = model,
    system = "You are a patient teacher who builds on previous concepts. Remember what was discussed and reference it.",
    temperature = 0.5
})

-- Simulate a learning session with context
local lesson_parts = {
    "Explain what recursion is in programming.",
    "Now give me a simple example of recursion.",
    "How does the example you gave actually work step by step?",
    "What are the potential problems with recursion?"
}

local conversation_history = {}

for i, question in ipairs(lesson_parts) do
    print("Question " .. i .. ": " .. question)
    
    -- Build context from history
    local context = ""
    if #conversation_history > 0 then
        context = "Previous discussion:\n"
        for _, exchange in ipairs(conversation_history) do
            context = context .. "Q: " .. exchange.question .. "\n"
            context = context .. "A: " .. exchange.answer .. "\n\n"
        end
        context = context .. "New question: " .. question
    else
        context = question
    end
    
    local answer = teacher:run(context)
    print("Answer: " .. answer)
    print()
    
    -- Add to history
    table.insert(conversation_history, {
        question = question,
        answer = answer
    })
end

-- Example 6: Creative Agents
print("=== Example 6: Creative Agents ===")

-- Create agents for creative tasks
local poet = agent.create({
    name = "Poet",
    model = model,
    system = "You are a creative poet who writes meaningful, evocative poetry. Use vivid imagery and metaphors.",
    temperature = 0.9  -- High temperature for creativity
})

local storyteller = agent.create({
    name = "Storyteller",
    model = model,
    system = "You are a master storyteller who creates engaging micro-stories. Every story should have a beginning, middle, and end.",
    temperature = 0.8
})

local comedian = agent.create({
    name = "Comedian",
    model = model,
    system = "You are a witty comedian who finds humor in everyday situations. Keep it clean and clever.",
    temperature = 0.8
})

-- Get creative outputs about the topic
print("Creative interpretations of: " .. topic)

print("\nPoet's creation:")
local poem = poet:run("Write a short 4-line poem about " .. topic)
print(poem)

print("\nStoryteller's creation:")
local story = storyteller:run("Tell a 3-sentence story about " .. topic)
print(story)

print("\nComedian's take:")
local joke = comedian:run("Make a clever observation or joke about " .. topic)
print(joke)
print()

-- Example 7: Analytical Agent Pipeline
print("=== Example 7: Analytical Pipeline ===")

-- Create a pipeline of analytical agents
local researcher = agent.create({
    name = "Researcher",
    model = model,
    system = "You are a researcher who identifies key facts and questions about a topic.",
    temperature = 0.4
})

local critic = agent.create({
    name = "Critic",
    model = model,
    system = "You are a critical thinker who identifies gaps, assumptions, and potential flaws in reasoning.",
    temperature = 0.4
})

local synthesizer = agent.create({
    name = "Synthesizer",
    model = model,
    system = "You are a synthesizer who combines different viewpoints into coherent conclusions.",
    temperature = 0.3
})

-- Run the analytical pipeline
print("Analyzing: " .. topic)

local research = researcher:run("Research and list 3 key facts about " .. topic)
print("\nResearch findings:")
print(research)

local critique = critic:run("Critique these findings and identify what's missing:\n" .. research)
print("\nCritical analysis:")
print(critique)

local synthesis = synthesizer:run(
    "Synthesize these perspectives:\n" ..
    "Research: " .. research .. "\n" ..
    "Critique: " .. critique .. "\n" ..
    "Provide a balanced conclusion."
)
print("\nFinal synthesis:")
print(synthesis)
print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Basic agent creation without tools")
print("2. Chain of thought reasoning")
print("3. Multiple agent personalities")
print("4. Agent debates and conversations")
print("5. Agents with memory/context")
print("6. Creative agents")
print("7. Analytical agent pipelines")
print()
print("Key insight: Agents without tools can still perform complex reasoning,")
print("analysis, and creative tasks through careful prompt engineering and")
print("structured interactions.")

-- Return summary of agent interactions
return {
    agents_created = {
        "analyst", "thinker", "optimist", "pessimist", "realist",
        "moderator", "teacher", "poet", "storyteller", "comedian",
        "researcher", "critic", "synthesizer"
    },
    topic_analyzed = topic,
    total_agents = 13,
    debate_summary = debate_summary,
    final_synthesis = synthesis
}