-- ABOUTME: Example demonstrating agents with tool access for autonomous task completion
-- ABOUTME: Shows tool selection, error recovery, and complex multi-step operations

-- Required modules
local agent = require("agent")

-- Agent With Tools Example
-- This spell demonstrates creating and using tool-enabled agents:
-- 1. Basic tool-enabled agent
-- 2. Agent tool selection and usage
-- 3. Multi-step autonomous tasks
-- 4. Error handling and recovery
-- 5. Tool usage optimization

-- Parameters
local task = params.task or "Research the latest developments in renewable energy and create a summary report"
local model = params.model or "gpt-4"  -- GPT-4 recommended for better tool usage
local output_dir = params.output_dir or "./agent-output"

print("=== Agent With Tools Example ===")
print("Task: " .. task)
print("Model: " .. model)
print("Output directory: " .. output_dir)
print()

-- Ensure output directory exists
if not tools.file_exists(output_dir) then
    tools.create_directory(output_dir)
end

-- Example 1: Basic Tool-Enabled Agent
print("=== Example 1: Basic Tool-Enabled Agent ===")

-- Create a research assistant with basic tools
local researcher = agent.create({
    name = "Research Assistant",
    model = model,
    system = "You are a helpful research assistant. Use the available tools to gather information and complete tasks efficiently.",
    tools = {"web_search", "file_write", "file_read"},
    temperature = 0.3
})

-- Simple task
local simple_result = researcher:run("Search for information about the Moon's temperature and save it to moon-facts.txt")
print("Simple task result:")
print(simple_result)
print()

-- Example 2: Agent with Calculator and Analysis Tools
print("=== Example 2: Agent with Calculation Tools ===")

local analyst = agent.create({
    name = "Data Analyst",
    model = model,
    system = "You are a data analyst who can perform calculations and analyze information. Be precise with numbers.",
    tools = {"calculator", "file_write", "file_read"},
    temperature = 0.2
})

-- Complex calculation task
local calc_task = [[
Calculate the following and save the results:
1. The compound interest on $10,000 at 5% annual rate for 10 years
2. The monthly payment for a $200,000 mortgage at 4% for 30 years
3. The break-even point if fixed costs are $50,000, variable cost per unit is $25, and selling price is $75

Save your calculations and results to 'financial-calculations.txt'
]]

local calc_result = analyst:run(calc_task)
print("Calculation task completed:")
print(calc_result)
print()

-- Example 3: Multi-Tool Research Agent
print("=== Example 3: Multi-Tool Research Agent ===")

local advanced_researcher = agent.create({
    name = "Advanced Researcher",
    model = model,
    system = [[You are an advanced research agent who can:
1. Search the web for information
2. Read and write files
3. Perform calculations
4. Format dates and times
Always cite your sources and organize information clearly.]],
    tools = {"web_search", "file_read", "file_write", "calculator", "datetime_now", "datetime_format"},
    temperature = 0.4
})

-- Execute the main research task
print("Executing main research task...")
local research_result = advanced_researcher:run(task)
print("\nResearch completed:")
print(research_result)
print()

-- Example 4: Agent Team with Specialized Tools
print("=== Example 4: Specialized Agent Team ===")

-- Create specialized agents
local web_researcher = agent.create({
    name = "Web Specialist",
    model = model,
    system = "You specialize in web research. Find the most relevant and recent information.",
    tools = {"web_search"},
    temperature = 0.3
})

local data_processor = agent.create({
    name = "Data Processor",
    model = model,
    system = "You process and organize information. Create structured summaries and extract key facts.",
    tools = {"file_read", "file_write", "calculator"},
    temperature = 0.2
})

local report_writer = agent.create({
    name = "Report Writer",
    model = model,
    system = "You create professional reports. Use proper formatting and clear structure.",
    tools = {"file_read", "file_write", "datetime_now", "datetime_format"},
    temperature = 0.4
})

-- Coordinate the team
print("Coordinating agent team...")

-- Step 1: Web research
local search_topic = "artificial intelligence in healthcare 2024"
print("Step 1: Web research on '" .. search_topic .. "'")
local web_findings = web_researcher:run("Research: " .. search_topic .. ". Find at least 3 relevant facts or developments.")
print("Web findings: " .. web_findings:sub(1, 200) .. "...")

-- Step 2: Process findings
print("\nStep 2: Processing findings")
tools.file_write(output_dir .. "/raw-findings.txt", web_findings)
local processed_data = data_processor:run(
    "Read the file 'raw-findings.txt' and create a structured summary with bullet points. " ..
    "Save it as 'processed-findings.txt'"
)
print("Processing complete")

-- Step 3: Create final report
print("\nStep 3: Creating final report")
local final_report = report_writer:run(
    "Read 'processed-findings.txt' and create a professional report about " .. search_topic .. 
    ". Include the current date and time. Save as 'final-report.md'"
)
print("Report created")
print()

-- Example 5: Error Recovery Agent
print("=== Example 5: Agent with Error Recovery ===")

local robust_agent = agent.create({
    name = "Robust Worker",
    model = model,
    system = [[You are a careful agent who handles errors gracefully. 
If a tool fails, try alternative approaches. 
Always explain what went wrong and how you recovered.]],
    tools = {"web_search", "file_read", "file_write", "calculator"},
    temperature = 0.3
})

-- Task that might fail
local risky_task = [[
1. Try to read a file called 'important-data.txt'
2. If it doesn't exist, create it with some default data
3. Search the web for "OpenAI GPT-4 capabilities"
4. If the search fails, use your knowledge to write about GPT-4
5. Calculate the square root of -1 (this will fail)
6. When the calculation fails, explain why and provide the correct mathematical answer
]]

print("Executing task with potential failures...")
local recovery_result = robust_agent:run(risky_task)
print("Result with error recovery:")
print(recovery_result)
print()

-- Example 6: Tool Usage Optimization
print("=== Example 6: Optimized Tool Usage ===")

local efficient_agent = agent.create({
    name = "Efficient Worker",
    model = model,
    system = [[You are an efficient agent who minimizes tool usage.
Before using a tool, consider if you already have the information.
Batch operations when possible.
Explain your tool usage strategy.]],
    tools = {"file_read", "file_write", "calculator", "web_search"},
    temperature = 0.2
})

-- Task requiring optimization
local optimization_task = [[
Create a report about prime numbers:
1. List the first 20 prime numbers (use calculator efficiently)
2. Calculate the sum of these primes
3. Search for interesting facts about prime numbers
4. Create a report combining all information
Minimize the number of tool calls by batching operations.
]]

print("Executing optimized task...")
local optimized_result = efficient_agent:run(optimization_task)
print("Optimized execution complete")
print()

-- Example 7: Complex Autonomous Task
print("=== Example 7: Complex Autonomous Task ===")

local autonomous_agent = agent.create({
    name = "Autonomous Worker",
    model = model,
    system = [[You are a fully autonomous agent capable of complex multi-step tasks.
Break down complex tasks into steps.
Use tools strategically.
Verify your work.
Create organized outputs.]],
    tools = {"web_search", "file_read", "file_write", "calculator", "datetime_now", "datetime_format"},
    temperature = 0.4
})

-- Complex task
local complex_task = [[
Create a comprehensive analysis of renewable energy:
1. Research current global renewable energy statistics
2. Calculate the growth rate over the past 5 years
3. Find the top 3 countries in renewable energy adoption
4. Create projections for the next 10 years
5. Write a structured report with:
   - Executive summary
   - Current state analysis
   - Growth calculations
   - Country rankings
   - Future projections
   - Timestamp and sources
Save the report as 'renewable-energy-analysis.md'
]]

print("Executing complex autonomous task...")
print("This may take a moment...")
local autonomous_result = autonomous_agent:run(complex_task)
print("\nComplex task completed!")
print("Agent's summary: " .. autonomous_result:sub(1, 300) .. "...")
print()

-- Example 8: Tool Usage Monitoring
print("=== Example 8: Tool Usage Analysis ===")

-- Create an agent that reports on its tool usage
local transparent_agent = agent.create({
    name = "Transparent Worker",
    model = model,
    system = [[You are a transparent agent who tracks and reports tool usage.
For each tool you use, note:
1. Why you chose that tool
2. What you expected to achieve
3. Whether it succeeded
At the end, summarize your tool usage patterns.]],
    tools = {"web_search", "file_write", "calculator", "datetime_now"},
    temperature = 0.3
})

local monitoring_task = [[
Complete these tasks and report on your tool usage:
1. Get the current date and time
2. Search for today's weather (this might fail if no location is specified)
3. Calculate the number of seconds in a day
4. Write a summary to 'tool-usage-report.txt'
Include a detailed analysis of which tools you used and why.
]]

local tool_report = transparent_agent:run(monitoring_task)
print("Tool usage analysis:")
print(tool_report)
print()

-- Summary and statistics
print("=== Summary ===")
print("This example demonstrated:")
print("1. Basic tool-enabled agents")
print("2. Agents with calculation capabilities")
print("3. Multi-tool research agents")
print("4. Specialized agent teams")
print("5. Error recovery strategies")
print("6. Tool usage optimization")
print("7. Complex autonomous tasks")
print("8. Tool usage monitoring")
print()

-- List all files created
print("Files created in " .. output_dir .. ":")
local created_files = tools.list_files(output_dir)
for _, file in ipairs(created_files) do
    print("  - " .. file)
end

-- Return summary
return {
    agents_created = 8,
    primary_task = task,
    output_directory = output_dir,
    files_created = created_files,
    demonstrated_tools = {
        "web_search", "file_read", "file_write",
        "calculator", "datetime_now", "datetime_format"
    }
}