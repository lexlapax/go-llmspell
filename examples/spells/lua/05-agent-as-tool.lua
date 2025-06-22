-- ABOUTME: Example demonstrating agents wrapped as tools for other agents
-- ABOUTME: Shows hierarchical agent systems, delegation, and specialized agent tools

-- Agent as Tool Example
-- This spell demonstrates advanced agent patterns:
-- 1. Wrapping agents as tools
-- 2. Hierarchical agent systems
-- 3. Agent delegation and specialization
-- 4. Complex multi-agent workflows
-- 5. Recursive agent structures

-- Parameters
local project = params.project or "Create a comprehensive business plan for a sustainable coffee shop"
local model = params.model or "gpt-4"
local output_dir = params.output_dir or "./agent-hierarchy-output"

print("=== Agent as Tool Example ===")
print("Project: " .. project)
print("Model: " .. model)
print()

-- Ensure output directory exists
if not tools.file_exists(output_dir) then
    tools.create_directory(output_dir)
end

-- Example 1: Simple Agent as Tool
print("=== Example 1: Basic Agent as Tool ===")

-- Create a specialized research agent
local market_researcher = agent.create({
    name = "Market Research Specialist",
    model = model,
    system = "You are a market research specialist. Provide detailed market analysis with data and trends.",
    tools = {"web_search"},
    temperature = 0.3
})

-- Wrap the agent as a tool
tools.register({
    name = "market_research",
    description = "Conduct detailed market research on any topic",
    parameters = {
        query = {type = "string", required = true, description = "What to research"}
    },
    execute = function(args)
        return market_researcher:run(args.query)
    end
})

-- Create a business analyst that uses the market research tool
local business_analyst = agent.create({
    name = "Business Analyst",
    model = model,
    system = "You are a business analyst who creates comprehensive reports. Use the market_research tool for data.",
    tools = {"market_research", "file_write", "calculator"},
    temperature = 0.4
})

-- Use the hierarchical system
local analysis_result = business_analyst:run(
    "Analyze the market opportunity for sustainable coffee shops. " ..
    "Use market research to find data about coffee consumption trends and sustainability preferences."
)
print("Business Analysis Result:")
print(analysis_result:sub(1, 500) .. "...")
print()

-- Example 2: Multiple Specialized Agent Tools
print("=== Example 2: Multiple Agent Tools ===")

-- Create specialized agents
local financial_analyst = agent.create({
    name = "Financial Analyst",
    model = model,
    system = "You are a financial analyst specializing in cost projections and ROI calculations.",
    tools = {"calculator"},
    temperature = 0.2
})

local legal_advisor = agent.create({
    name = "Legal Advisor",
    model = model,
    system = "You are a legal advisor who identifies regulatory requirements and compliance needs.",
    tools = {"web_search"},
    temperature = 0.2
})

local marketing_expert = agent.create({
    name = "Marketing Expert",
    model = model,
    system = "You are a marketing expert who creates compelling strategies and brand positioning.",
    tools = {"web_search"},
    temperature = 0.6
})

-- Register all specialist agents as tools
tools.register({
    name = "financial_analysis",
    description = "Get detailed financial analysis and projections",
    parameters = {
        request = {type = "string", required = true}
    },
    execute = function(args)
        return financial_analyst:run(args.request)
    end
})

tools.register({
    name = "legal_consultation",
    description = "Get legal and regulatory compliance advice",
    parameters = {
        request = {type = "string", required = true}
    },
    execute = function(args)
        return legal_advisor:run(args.request)
    end
})

tools.register({
    name = "marketing_strategy",
    description = "Get marketing strategy and branding advice",
    parameters = {
        request = {type = "string", required = true}
    },
    execute = function(args)
        return marketing_expert:run(args.request)
    end
})

-- Create a project manager that coordinates all specialists
local project_manager = agent.create({
    name = "Project Manager",
    model = model,
    system = [[You are a senior project manager coordinating a team of specialists.
You have access to:
- market_research: For market analysis
- financial_analysis: For financial projections
- legal_consultation: For compliance issues
- marketing_strategy: For marketing plans
- file_write: To save reports

Coordinate these specialists effectively to complete complex projects.
Always save important findings to files.]],
    tools = {"market_research", "financial_analysis", "legal_consultation", "marketing_strategy", "file_write"},
    temperature = 0.4
})

-- Execute complex project
print("Executing complex project with specialist team...")
local project_result = project_manager:run(project .. [[

Please coordinate your team to create:
1. Market analysis (use market_research)
2. Financial projections for 3 years (use financial_analysis)
3. Legal requirements checklist (use legal_consultation)
4. Marketing strategy outline (use marketing_strategy)

Save each section to a separate file and create a summary report.]])

print("Project completed!")
print("Manager's summary: " .. project_result:sub(1, 400) .. "...")
print()

-- Example 3: Recursive Agent Structure
print("=== Example 3: Recursive Agent Structure ===")

-- Create a review agent that can use other agents
local reviewer = agent.create({
    name = "Quality Reviewer",
    model = model,
    system = "You are a quality reviewer who ensures work meets high standards. Provide constructive feedback.",
    tools = {"file_read"},
    temperature = 0.3
})

-- Register the reviewer as a tool
tools.register({
    name = "quality_review",
    description = "Review and provide feedback on any work",
    parameters = {
        work = {type = "string", required = true}
    },
    execute = function(args)
        return reviewer:run("Review this work and provide feedback:\n" .. args.work)
    end
})

-- Create a writer agent that uses the reviewer
local writer = agent.create({
    name = "Content Writer",
    model = model,
    system = "You are a content writer. Use quality_review to check your work and iterate based on feedback.",
    tools = {"quality_review", "file_write"},
    temperature = 0.5
})

-- Register the writer as a tool (creating recursion potential)
tools.register({
    name = "content_writing",
    description = "Create written content with built-in quality review",
    parameters = {
        topic = {type = "string", required = true}
    },
    execute = function(args)
        return writer:run("Write about: " .. args.topic .. ". Review your work and revise based on feedback.")
    end
})

-- Use the recursive system
print("Testing recursive agent system...")
local content = writer:run(
    "Write a 200-word introduction for a sustainable coffee shop business plan. " ..
    "Use quality_review to check your work and improve it based on the feedback."
)
print("Content with review iterations:")
print(content:sub(1, 500) .. "...")
print()

-- Example 4: Dynamic Agent Team Assembly
print("=== Example 4: Dynamic Agent Teams ===")

-- Function to create specialist agent tools dynamically
local function create_specialist_tool(name, specialty, model_override)
    local specialist = agent.create({
        name = name,
        model = model_override or model,
        system = "You are a specialist in " .. specialty .. ". Provide expert insights in your domain.",
        tools = {"web_search", "calculator"},
        temperature = 0.4
    })
    
    tools.register({
        name = name:lower():gsub(" ", "_"),
        description = specialty .. " specialist",
        parameters = {
            query = {type = "string", required = true}
        },
        execute = function(args)
            return specialist:run(args.query)
        end
    })
    
    return name:lower():gsub(" ", "_")
end

-- Create a team dynamically
local specialties = {
    {name = "Sustainability Expert", specialty = "environmental sustainability and green practices"},
    {name = "Customer Experience Designer", specialty = "customer journey and experience design"},
    {name = "Supply Chain Analyst", specialty = "supply chain optimization and vendor management"},
    {name = "Technology Consultant", specialty = "technology solutions and digital transformation"}
}

local tool_names = {}
for _, spec in ipairs(specialties) do
    local tool_name = create_specialist_tool(spec.name, spec.specialty)
    table.insert(tool_names, tool_name)
    print("Created specialist tool: " .. tool_name)
end

-- Create a coordinator with all dynamic tools
local all_tools = {"file_write"}
for _, tool_name in ipairs(tool_names) do
    table.insert(all_tools, tool_name)
end

local coordinator = agent.create({
    name = "Team Coordinator",
    model = model,
    system = "You coordinate a dynamic team of specialists. Use their expertise to solve complex problems.",
    tools = all_tools,
    temperature = 0.4
})

-- Use the dynamic team
print("\nCoordinating dynamic team...")
local team_result = coordinator:run(
    "For a sustainable coffee shop, consult with: " ..
    "1. Sustainability expert about eco-friendly practices " ..
    "2. Customer experience designer about the ideal customer journey " ..
    "3. Supply chain analyst about sourcing ethical coffee " ..
    "4. Technology consultant about POS and customer apps " ..
    "Save a summary of all consultations."
)
print("Dynamic team result completed")
print()

-- Example 5: Agent Tool with State
print("=== Example 5: Stateful Agent Tools ===")

-- Create an agent with memory
local memory_agent = agent.create({
    name = "Memory Assistant",
    model = model,
    system = "You are an assistant with perfect memory. Remember all previous interactions and reference them.",
    tools = {"file_read", "file_write"},
    temperature = 0.3
})

-- Create a stateful wrapper
local conversation_history = {}

tools.register({
    name = "memory_assistant",
    description = "An assistant that remembers previous interactions",
    parameters = {
        message = {type = "string", required = true}
    },
    execute = function(args)
        -- Add context from history
        local context = "Previous conversations:\n"
        for i, entry in ipairs(conversation_history) do
            context = context .. "Q" .. i .. ": " .. entry.question .. "\n"
            context = context .. "A" .. i .. ": " .. entry.answer .. "\n\n"
        end
        context = context .. "New message: " .. args.message
        
        -- Get response
        local response = memory_agent:run(context)
        
        -- Update history
        table.insert(conversation_history, {
            question = args.message,
            answer = response
        })
        
        return response
    end
})

-- Create a main agent that uses the stateful assistant
local main_agent = agent.create({
    name = "Main Coordinator",
    model = model,
    system = "You coordinate tasks and use the memory_assistant for continuity across interactions.",
    tools = {"memory_assistant", "file_write"},
    temperature = 0.4
})

-- Test stateful interactions
print("Testing stateful agent tool...")
local tasks = {
    "Tell the memory assistant about our sustainable coffee shop project",
    "Ask the memory assistant what we discussed earlier",
    "Ask the memory assistant to summarize all our discussions"
}

for i, task in ipairs(tasks) do
    print("Task " .. i .. ": " .. task)
    local result = main_agent:run(task)
    print("Result: " .. result:sub(1, 300) .. "...")
    print()
end

-- Example 6: Agent Pipeline as Tool
print("=== Example 6: Agent Pipeline Tool ===")

-- Create a pipeline of agents as a single tool
local function create_analysis_pipeline()
    local data_collector = agent.create({
        name = "Data Collector",
        model = model,
        system = "You collect relevant data and statistics.",
        tools = {"web_search"},
        temperature = 0.3
    })
    
    local data_analyzer = agent.create({
        name = "Data Analyzer",
        model = model,
        system = "You analyze data and identify patterns and insights.",
        tools = {"calculator"},
        temperature = 0.3
    })
    
    local report_generator = agent.create({
        name = "Report Generator",
        model = model,
        system = "You create clear, professional reports from analyzed data.",
        tools = {"file_write"},
        temperature = 0.4
    })
    
    -- Create pipeline tool
    tools.register({
        name = "analysis_pipeline",
        description = "Complete analysis pipeline: collect, analyze, and report",
        parameters = {
            topic = {type = "string", required = true}
        },
        execute = function(args)
            print("  [Pipeline] Stage 1: Collecting data...")
            local data = data_collector:run("Collect data about: " .. args.topic)
            
            print("  [Pipeline] Stage 2: Analyzing...")
            local analysis = data_analyzer:run("Analyze this data: " .. data)
            
            print("  [Pipeline] Stage 3: Generating report...")
            local report = report_generator:run(
                "Create a report based on this analysis: " .. analysis .. 
                "\nSave the report as '" .. args.topic:gsub(" ", "-") .. "-report.md'"
            )
            
            return report
        end
    })
end

create_analysis_pipeline()

-- Use the pipeline tool
local executive = agent.create({
    name = "Executive",
    model = model,
    system = "You are an executive who delegates complex analysis tasks. Use the analysis_pipeline for research.",
    tools = {"analysis_pipeline", "file_read"},
    temperature = 0.4
})

print("Testing pipeline tool...")
local pipeline_result = executive:run(
    "Use the analysis pipeline to research 'coffee shop sustainability trends'. " ..
    "Then read the generated report and give me the key takeaways."
)
print("Pipeline execution complete")
print("Executive summary: " .. pipeline_result:sub(1, 400) .. "...")
print()

-- Summary
print("=== Summary ===")
print("This example demonstrated:")
print("1. Basic agents wrapped as tools")
print("2. Multiple specialized agent tools")
print("3. Recursive agent structures") 
print("4. Dynamic agent team assembly")
print("5. Stateful agent tools with memory")
print("6. Agent pipelines as single tools")
print()
print("Key insights:")
print("- Agents can be wrapped as tools for other agents")
print("- This enables hierarchical and specialized systems")
print("- State can be maintained across agent tool calls")
print("- Complex pipelines can be encapsulated as single tools")
print()

-- List created files
print("Files created in " .. output_dir .. ":")
local files = tools.list_files(output_dir)
for _, file in ipairs(files) do
    print("  - " .. file)
end

-- Return summary
return {
    hierarchies_demonstrated = 6,
    specialist_agents_created = #specialties + 7,  -- Dynamic + individual specialists
    project = project,
    files_created = files,
    advanced_patterns = {
        "agent_as_tool",
        "recursive_agents",
        "dynamic_teams",
        "stateful_agents",
        "agent_pipelines"
    }
}