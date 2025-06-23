-- ABOUTME: Agent handoff example showing how to transfer control between specialized agents
-- ABOUTME: Demonstrates agent coordination, context passing, and seamless handoff patterns

-- Agent Handoff Example
-- This spell demonstrates advanced agent coordination including:
-- 1. Specialized agent creation with specific capabilities
-- 2. Context preservation during handoffs
-- 3. Decision making for agent selection
-- 4. Seamless conversation flow between agents
-- 5. Result aggregation from multiple agents

-- Required modules
local agent = require("agent")
local llm = require("llm")
local tools = require("tools")
local state = require("state")
local log = require("log")
local data = require("data")
local core = require("core")
local errors = require("errors")

-- Example 1: Basic Agent Handoff
print("=== Example 1: Basic Agent Handoff ===")

-- Create specialized agents
local customer_service_agent = agent.create({
    name = "Customer Service Agent",
    model = "gpt-3.5-turbo",
    instructions = [[
        You are a friendly customer service representative.
        You help with general inquiries, account questions, and basic support.
        If the customer needs technical help, say "I'll transfer you to our technical support specialist."
        If they need billing help, say "I'll connect you with our billing department."
    ]]
})

local technical_support_agent = agent.create({
    name = "Technical Support Agent",
    model = "gpt-4",
    instructions = [[
        You are a technical support specialist with deep knowledge of software and hardware issues.
        Provide detailed technical solutions and troubleshooting steps.
        If the issue is resolved, say "Is there anything else I can help you with?"
    ]],
    tools = {"web_search", "code_interpreter"}  -- Has additional tools
})

local billing_agent = agent.create({
    name = "Billing Agent",
    model = "gpt-3.5-turbo",
    instructions = [[
        You handle billing inquiries, payment processing, and subscription management.
        Be professional and precise with financial information.
        Always confirm any changes with the customer.
    ]]
})

-- Router function to determine which agent to use
local function route_to_agent(query, context)
    -- Use a simple LLM call to classify the request
    local classification = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {
            {role = "system", content = [[
                Classify the customer query into one of these categories:
                - general: General questions, greetings, account info
                - technical: Software bugs, hardware issues, technical problems
                - billing: Payments, subscriptions, invoices, refunds
                
                Respond with only the category name.
            ]]},
            {role = "user", content = query}
        },
        max_tokens = 10
    })
    
    local category = string.lower(string.gsub(classification.content, "%s+", ""))
    log.debug("Classified query as: " .. category)
    
    -- Return appropriate agent
    if string.find(category, "technical") then
        return technical_support_agent, "technical"
    elseif string.find(category, "billing") then
        return billing_agent, "billing"
    else
        return customer_service_agent, "general"
    end
end

-- Handle a customer interaction with handoffs
local function handle_customer_interaction(initial_query)
    local conversation_context = {
        history = {},
        current_agent = nil,
        handoff_count = 0
    }
    
    -- Start with routing the initial query
    local current_agent, category = route_to_agent(initial_query, conversation_context)
    conversation_context.current_agent = current_agent.name
    
    print("\n[" .. current_agent.name .. " handling query]")
    
    -- Process the query
    local response = current_agent:generate({
        prompt = initial_query,
        max_tokens = 200
    })
    
    -- Add to conversation history
    table.insert(conversation_context.history, {
        agent = current_agent.name,
        user_message = initial_query,
        agent_response = response.content
    })
    
    print("Response: " .. response.content)
    
    -- Check if handoff is needed
    if string.find(string.lower(response.content), "transfer you to") or 
       string.find(string.lower(response.content), "connect you with") then
        conversation_context.handoff_count = conversation_context.handoff_count + 1
        print("\n[Handoff detected - transferring...]")
        
        -- Simulate follow-up technical question
        if string.find(string.lower(response.content), "technical") then
            local tech_query = "My application crashes when I try to export files"
            print("\nCustomer: " .. tech_query)
            
            local tech_response = technical_support_agent:generate({
                prompt = tech_query .. "\n\nContext: Customer was just transferred from customer service.",
                max_tokens = 300
            })
            
            print("\n[" .. technical_support_agent.name .. " responds]")
            print("Response: " .. tech_response.content)
            
            table.insert(conversation_context.history, {
                agent = technical_support_agent.name,
                user_message = tech_query,
                agent_response = tech_response.content
            })
        end
    end
    
    return conversation_context
end

-- Test the handoff system
local test_queries = {
    "I'm having trouble with my software crashing",
    "Hello, I need help with my account"
}

for _, query in ipairs(test_queries) do
    print("\n--- New Customer Interaction ---")
    print("Customer: " .. query)
    local context = handle_customer_interaction(query)
    print("Total handoffs: " .. context.handoff_count)
end

print()

-- Example 2: Complex Multi-Agent Workflow
print("=== Example 2: Complex Multi-Agent Workflow ===")

-- Create a team of specialized agents for a research project
local research_coordinator = agent.create({
    name = "Research Coordinator",
    model = "gpt-4",
    instructions = [[
        You coordinate research projects by:
        1. Breaking down research questions into subtasks
        2. Assigning tasks to appropriate specialist agents
        3. Synthesizing results from multiple agents
        4. Ensuring comprehensive coverage of the topic
    ]]
})

local data_analyst = agent.create({
    name = "Data Analyst",
    model = "gpt-4",
    instructions = [[
        You analyze data and statistics related to research topics.
        Provide quantitative insights, trends, and data-driven conclusions.
        Always cite your sources and explain your methodology.
    ]],
    tools = {"calculator", "web_search"}
})

local literature_reviewer = agent.create({
    name = "Literature Reviewer",
    model = "gpt-3.5-turbo",
    instructions = [[
        You review academic literature and research papers.
        Summarize key findings, identify research gaps, and synthesize knowledge.
        Focus on peer-reviewed sources and recent publications.
    ]],
    tools = {"web_search"}
})

local domain_expert = agent.create({
    name = "Domain Expert",
    model = "gpt-4",
    instructions = [[
        You provide expert domain knowledge and practical insights.
        Explain complex concepts clearly and relate them to real-world applications.
        Highlight important considerations and potential challenges.
    ]]
})

-- Coordinate a research project
local function coordinate_research(topic)
    print("\nResearch Topic: " .. topic)
    print("\n[Research Coordinator planning...]")
    
    -- Initialize shared state for the research
    state.set("research_project", {
        topic = topic,
        subtasks = {},
        findings = {},
        status = "planning"
    })
    
    -- Step 1: Break down the research
    local breakdown = research_coordinator:generate({
        prompt = string.format([[
            Create a research plan for: %s
            
            Break it down into 3 specific subtasks that can be handled by:
            1. Data Analyst - for quantitative analysis
            2. Literature Reviewer - for academic research
            3. Domain Expert - for practical insights
            
            Format each subtask clearly.
        ]], topic),
        max_tokens = 300
    })
    
    print("Research plan created:")
    print(breakdown.content)
    
    -- Store subtasks
    state.set("research_project.subtasks", {
        data_analysis = "Analyze quantitative data about " .. topic,
        literature_review = "Review recent research on " .. topic,
        domain_expertise = "Provide practical insights on " .. topic
    })
    
    state.set("research_project.status", "in_progress")
    
    -- Step 2: Execute subtasks in parallel (simulated)
    print("\n[Distributing tasks to specialist agents...]")
    
    -- Data Analyst work
    print("\n[Data Analyst working...]")
    local data_findings = data_analyst:generate({
        prompt = state.get("research_project.subtasks.data_analysis"),
        max_tokens = 250
    })
    state.set("research_project.findings.data_analysis", data_findings.content)
    print("✓ Data analysis complete")
    
    -- Literature Reviewer work
    print("\n[Literature Reviewer working...]")
    local literature_findings = literature_reviewer:generate({
        prompt = state.get("research_project.subtasks.literature_review"),
        max_tokens = 250
    })
    state.set("research_project.findings.literature_review", literature_findings.content)
    print("✓ Literature review complete")
    
    -- Domain Expert work
    print("\n[Domain Expert working...]")
    local expert_findings = domain_expert:generate({
        prompt = state.get("research_project.subtasks.domain_expertise"),
        max_tokens = 250
    })
    state.set("research_project.findings.domain_expertise", expert_findings.content)
    print("✓ Domain expertise gathered")
    
    -- Step 3: Synthesize results
    print("\n[Research Coordinator synthesizing findings...]")
    
    local findings = state.get("research_project.findings")
    local synthesis = research_coordinator:generate({
        prompt = string.format([[
            Synthesize these research findings on "%s":
            
            DATA ANALYSIS:
            %s
            
            LITERATURE REVIEW:
            %s
            
            DOMAIN EXPERTISE:
            %s
            
            Create a comprehensive summary with key insights and recommendations.
        ]], topic, findings.data_analysis, findings.literature_review, findings.domain_expertise),
        max_tokens = 400
    })
    
    state.set("research_project.synthesis", synthesis.content)
    state.set("research_project.status", "completed")
    
    return {
        topic = topic,
        synthesis = synthesis.content,
        agent_contributions = 4,
        status = "completed"
    }
end

-- Run a research project
local research_result = coordinate_research("The impact of AI on software development workflows")
print("\n=== Research Summary ===")
print(string.sub(research_result.synthesis, 1, 300) .. "...")
print("\nAgents involved: " .. research_result.agent_contributions)
print()

-- Example 3: Dynamic Agent Selection with Context
print("=== Example 3: Dynamic Agent Selection with Context ===")

-- Create a pool of specialized agents
local agent_pool = {
    python_expert = agent.create({
        name = "Python Expert",
        model = "gpt-4",
        instructions = "You are a Python programming expert. Help with Python code, libraries, and best practices."
    }),
    
    javascript_expert = agent.create({
        name = "JavaScript Expert",
        model = "gpt-4",
        instructions = "You are a JavaScript/Node.js expert. Help with JS, frameworks, and web development."
    }),
    
    database_expert = agent.create({
        name = "Database Expert",
        model = "gpt-3.5-turbo",
        instructions = "You are a database expert. Help with SQL, NoSQL, database design, and optimization."
    }),
    
    devops_expert = agent.create({
        name = "DevOps Expert",
        model = "gpt-3.5-turbo",
        instructions = "You are a DevOps expert. Help with CI/CD, containerization, cloud services, and deployment."
    })
}

-- Smart agent selector based on context
local function select_expert(query, context)
    -- Build context string
    local context_str = ""
    if context and context.previous_topics then
        context_str = "Previous topics: " .. table.concat(context.previous_topics, ", ")
    end
    
    -- Use LLM to select the best expert
    local selection = llm.complete({
        model = "gpt-3.5-turbo",
        messages = {
            {role = "system", content = [[
                Select the best expert for this query from:
                - python_expert: Python programming
                - javascript_expert: JavaScript/web development
                - database_expert: Database/SQL
                - devops_expert: DevOps/deployment
                
                Consider the context and respond with only the expert key.
            ]]},
            {role = "user", content = "Query: " .. query .. "\n" .. context_str}
        },
        max_tokens = 20
    })
    
    local expert_key = string.lower(string.gsub(selection.content, "%s+", ""))
    expert_key = string.gsub(expert_key, "[^%w_]", "")  -- Clean the key
    
    -- Fallback to python expert if key not found
    if not agent_pool[expert_key] then
        log.warn("Unknown expert key: " .. expert_key .. ", falling back to python_expert")
        expert_key = "python_expert"
    end
    
    return agent_pool[expert_key], expert_key
end

-- Interactive help system with dynamic routing
local function provide_help(queries)
    local context = {
        previous_topics = {},
        experts_used = {}
    }
    
    for i, query in ipairs(queries) do
        print("\n--- Query " .. i .. " ---")
        print("User: " .. query)
        
        -- Select appropriate expert
        local expert, expert_key = select_expert(query, context)
        print("[Routing to " .. expert.name .. "]")
        
        -- Get response
        local response = expert:generate({
            prompt = query,
            max_tokens = 200
        })
        
        print("Expert: " .. string.sub(response.content, 1, 150) .. "...")
        
        -- Update context
        table.insert(context.previous_topics, expert_key)
        context.experts_used[expert_key] = (context.experts_used[expert_key] or 0) + 1
    end
    
    -- Summary
    print("\n=== Session Summary ===")
    print("Experts consulted:")
    for expert, count in pairs(context.experts_used) do
        print("  " .. expert .. ": " .. count .. " times")
    end
    
    return context
end

-- Test dynamic routing
local test_questions = {
    "How do I create a virtual environment in Python?",
    "What's the best way to handle async operations in Node.js?",
    "How can I optimize a slow SQL query with multiple joins?",
    "How do I set up a CI/CD pipeline for my Python project?"
}

provide_help(test_questions)
print()

-- Example 4: Agent Handoff with State Transfer
print("=== Example 4: Agent Handoff with State Transfer ===")

-- Create agents that maintain and transfer state
local intake_agent = agent.create({
    name = "Intake Agent",
    model = "gpt-3.5-turbo",
    instructions = [[
        You gather initial information from users.
        Collect: name, issue type, urgency level, and any relevant details.
        Summarize the information before handoff.
    ]]
})

local resolution_agent = agent.create({
    name = "Resolution Agent",
    model = "gpt-4",
    instructions = [[
        You resolve issues based on intake information.
        Review the provided context and offer specific solutions.
        Be thorough and actionable in your recommendations.
    ]]
})

local followup_agent = agent.create({
    name = "Followup Agent",
    model = "gpt-3.5-turbo",
    instructions = [[
        You handle post-resolution followup.
        Check if the issue was resolved, gather feedback, and document outcomes.
        Be friendly and ensure customer satisfaction.
    ]]
})

-- Process a support ticket through multiple agents
local function process_support_ticket(initial_message)
    print("\n[Starting support ticket process]")
    
    -- Phase 1: Intake
    print("\n[Intake Agent gathering information...]")
    local intake_response = intake_agent:generate({
        prompt = "Customer says: " .. initial_message .. "\n\nGather necessary information and create a summary.",
        max_tokens = 200
    })
    
    -- Extract and structure intake data
    local ticket_data = {
        initial_message = initial_message,
        intake_summary = intake_response.content,
        status = "intake_complete",
        timestamp = os.time()
    }
    
    print("Intake complete. Summary created.")
    
    -- Phase 2: Resolution
    print("\n[Handing off to Resolution Agent...]")
    local resolution_response = resolution_agent:generate({
        prompt = string.format([[
            Support ticket information:
            
            Customer Issue: %s
            
            Intake Summary: %s
            
            Please provide a detailed resolution.
        ]], initial_message, intake_response.content),
        max_tokens = 300
    })
    
    ticket_data.resolution = resolution_response.content
    ticket_data.status = "resolved"
    
    print("Resolution provided.")
    
    -- Phase 3: Followup
    print("\n[Handing off to Followup Agent...]")
    local followup_response = followup_agent:generate({
        prompt = string.format([[
            Ticket Summary:
            - Original Issue: %s
            - Resolution Provided: %s
            
            Create a followup message to ensure customer satisfaction.
        ]], initial_message, string.sub(resolution_response.content, 1, 100)),
        max_tokens = 150
    })
    
    ticket_data.followup = followup_response.content
    ticket_data.status = "complete"
    
    print("\n=== Ticket Processing Complete ===")
    print("Agents involved: Intake → Resolution → Followup")
    print("\nFollowup message: " .. followup_response.content)
    
    return ticket_data
end

-- Process a support ticket
local ticket = process_support_ticket("My API calls are returning 429 errors intermittently")
print()

-- Example 5: Parallel Agent Collaboration
print("=== Example 5: Parallel Agent Collaboration ===")

-- Create agents that work in parallel
local security_auditor = agent.create({
    name = "Security Auditor",
    model = "gpt-4",
    instructions = "You analyze code and systems for security vulnerabilities. Be thorough and specific."
})

local performance_auditor = agent.create({
    name = "Performance Auditor", 
    model = "gpt-3.5-turbo",
    instructions = "You analyze code and systems for performance issues. Focus on optimization opportunities."
})

local code_reviewer = agent.create({
    name = "Code Reviewer",
    model = "gpt-4",
    instructions = "You review code for quality, maintainability, and best practices. Be constructive."
})

-- Conduct parallel code review
local function parallel_code_review(code_snippet)
    print("\n[Initiating parallel code review...]")
    
    local review_results = {
        security = nil,
        performance = nil,
        quality = nil
    }
    
    -- In a real implementation, these would run concurrently
    -- For this example, we'll simulate parallel execution
    
    print("\n[3 agents analyzing code simultaneously...]")
    
    -- Security audit
    review_results.security = security_auditor:generate({
        prompt = "Review this code for security issues:\n\n" .. code_snippet,
        max_tokens = 200
    })
    print("✓ Security audit complete")
    
    -- Performance audit
    review_results.performance = performance_auditor:generate({
        prompt = "Review this code for performance issues:\n\n" .. code_snippet,
        max_tokens = 200
    })
    print("✓ Performance audit complete")
    
    -- Code quality review
    review_results.quality = code_reviewer:generate({
        prompt = "Review this code for quality and best practices:\n\n" .. code_snippet,
        max_tokens = 200
    })
    print("✓ Code quality review complete")
    
    -- Aggregate results
    print("\n=== Consolidated Review Results ===")
    print("\nSecurity Findings:")
    print(string.sub(review_results.security.content, 1, 150) .. "...")
    print("\nPerformance Findings:")
    print(string.sub(review_results.performance.content, 1, 150) .. "...")
    print("\nCode Quality Findings:")
    print(string.sub(review_results.quality.content, 1, 150) .. "...")
    
    return review_results
end

-- Example code for review
local sample_code = [[
def get_user_data(user_id):
    query = f"SELECT * FROM users WHERE id = {user_id}"
    result = db.execute(query)
    return result[0]
]]

parallel_code_review(sample_code)

-- Return summary
return {
    success = true,
    examples_demonstrated = {
        basic_handoff = true,
        multi_agent_workflow = true,
        dynamic_selection = true,
        state_transfer = true,
        parallel_collaboration = true
    },
    total_agents_created = 16,
    handoff_patterns = {"router-based", "sequential", "parallel", "context-aware"}
}