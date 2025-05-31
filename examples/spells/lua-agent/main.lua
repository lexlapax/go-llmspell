-- Lua Agent Example
-- Demonstrates creating custom agents that orchestrate LLM calls with tools

print("=== Lua Agent Example ===\n")

-- Example 1: Research Agent with web_fetch tool
print("1. Creating a Research Agent that uses tools and LLM...")

local research_agent = {
    system_prompt = "You are a research assistant that gathers information from the web and provides comprehensive summaries.",
    
    execute = function(self, input, options)
        -- First, extract what to search for using the LLM
        local extract_prompt = "Extract the main topic or URL from this request. If it's a URL, return just the URL. If it's a topic, suggest a good URL to research it. User request: " .. input
        
        local url_or_topic, err = llm.chat(extract_prompt)
        if err then
            return "Failed to process request: " .. err
        end
        
        -- Check if we have web_fetch tool available
        local tools_list = tools.list()
        local has_web_fetch = false
        for _, tool in ipairs(tools_list) do
            if tool.name == "web_fetch" then
                has_web_fetch = true
                break
            end
        end
        
        if not has_web_fetch then
            return "Web fetch tool not available. Please ensure built-in tools are enabled."
        end
        
        -- Use web_fetch to get content
        print("  → Fetching web content...")
        local fetch_result, fetch_err = tools.execute("web_fetch", {
            url = url_or_topic:match("^https?://") and url_or_topic or "https://en.wikipedia.org/wiki/" .. url_or_topic:gsub(" ", "_")
        })
        
        if fetch_err then
            return "Failed to fetch web content: " .. fetch_err
        end
        
        -- Now use LLM to summarize the fetched content
        local summary_prompt = string.format(
            "Based on the following web content, provide a comprehensive summary about '%s':\n\n%s\n\nProvide a clear, informative summary.",
            input,
            fetch_result.content or fetch_result
        )
        
        local summary, summary_err = llm.chat(summary_prompt)
        if summary_err then
            return "Failed to generate summary: " .. summary_err
        end
        
        return summary
    end
}

-- Register the research agent
local success, err = agents.register("research-agent", research_agent)
if success then
    print("✓ Research agent registered successfully")
else
    print("✗ Failed to register research agent: " .. (err or "unknown error"))
    return
end

-- Test the research agent
print("\nTesting research agent:")
local result, err = agents.execute("research-agent", "Tell me about the Lua programming language")
if result then
    print("Research Result:")
    print(result)
else
    print("Error: " .. (err or "unknown error"))
end

print("\n" .. string.rep("-", 80) .. "\n")

-- Example 2: Code Analysis Agent with custom Lua tool
print("2. Creating a Code Analysis Agent with custom tools...")

-- First, create a custom tool for code analysis
tools.register("code_analyzer", "Analyzes code complexity and structure", {
    type = "object",
    properties = {
        code = {
            type = "string",
            description = "The code to analyze"
        },
        language = {
            type = "string",
            description = "Programming language",
            default = "lua"
        }
    },
    required = {"code"}
}, function(params)
    -- Simple code analysis
    local code = params.code
    local lines = 0
    local functions = 0
    local loops = 0
    
    -- Count lines
    for _ in code:gmatch("[^\n]+") do
        lines = lines + 1
    end
    
    -- Count functions (Lua specific)
    for _ in code:gmatch("function%s+%w+") do
        functions = functions + 1
    end
    for _ in code:gmatch("=%s*function") do
        functions = functions + 1
    end
    
    -- Count loops
    for _ in code:gmatch("for%s+") do
        loops = loops + 1
    end
    for _ in code:gmatch("while%s+") do
        loops = loops + 1
    end
    
    return {
        lines = lines,
        functions = functions,
        loops = loops,
        complexity = functions + loops
    }
end)

-- Create the code analysis agent
local code_agent = {
    system_prompt = "You are a code review expert that analyzes code quality and provides actionable feedback.",
    
    execute = function(self, input, options)
        -- First analyze the code structure
        local analysis, analysis_err = tools.execute("code_analyzer", {
            code = input,
            language = "lua"
        })
        
        if analysis_err then
            return "Failed to analyze code: " .. analysis_err
        end
        
        -- Use LLM to provide detailed review based on metrics
        local review_prompt = string.format([[
Based on the following code metrics and the code itself, provide a detailed code review:

Metrics:
- Lines of code: %d
- Functions: %d  
- Loops: %d
- Complexity score: %d

Code to review:
```lua
%s
```

Please provide:
1. Code quality assessment
2. Potential improvements
3. Best practices recommendations
4. Any bugs or issues found
]], analysis.lines, analysis.functions, analysis.loops, analysis.complexity, input)
        
        local review, review_err = llm.chat(review_prompt)
        if review_err then
            return "Failed to generate code review: " .. review_err
        end
        
        return string.format("=== Code Analysis Results ===\nMetrics: %d lines, %d functions, %d loops (complexity: %d)\n\n%s",
            analysis.lines, analysis.functions, analysis.loops, analysis.complexity, review)
    end
}

-- Register the code agent
success, err = agents.register("code-agent", code_agent)
if success then
    print("✓ Code analysis agent registered successfully")
else
    print("✗ Failed to register code agent: " .. (err or "unknown error"))
    return
end

-- Test with sample code
local sample_code = [[
function fibonacci(n)
    if n <= 1 then
        return n
    end
    
    local a, b = 0, 1
    for i = 2, n do
        a, b = b, a + b
    end
    
    return b
end

function factorial(n)
    local result = 1
    for i = 2, n do
        result = result * i
    end
    return result
end
]]

print("\nTesting code analysis agent:")
result, err = agents.execute("code-agent", sample_code)
if result then
    print(result)
else
    print("Error: " .. (err or "unknown error"))
end

print("\n" .. string.rep("-", 80) .. "\n")

-- Example 3: Multi-step Planning Agent
print("3. Creating a Multi-step Planning Agent...")

local planning_agent = {
    system_prompt = "You are a planning assistant that breaks down complex tasks into actionable steps.",
    
    execute = function(self, input, options)
        -- Step 1: Break down the task using LLM
        local breakdown_prompt = "Break down this task into 3-5 concrete steps: " .. input
        local steps, err = llm.chat(breakdown_prompt)
        if err then
            return "Failed to create plan: " .. err
        end
        
        -- Step 2: For each step, determine if tools are needed
        local detailed_plan_prompt = string.format([[
For this task: "%s"

I've identified these steps:
%s

Now, for each step, identify:
1. What specific actions are needed
2. What tools might help (available: web_fetch, code_analyzer)
3. Expected outcomes

Format as a detailed action plan.
]], input, steps)
        
        local detailed_plan, plan_err = llm.chat(detailed_plan_prompt)
        if plan_err then
            return "Failed to create detailed plan: " .. plan_err
        end
        
        return string.format("=== Task Planning Results ===\nTask: %s\n\n%s", input, detailed_plan)
    end
}

-- Register the planning agent
success, err = agents.register("planning-agent", planning_agent)
if success then
    print("✓ Planning agent registered successfully")
else
    print("✗ Failed to register planning agent: " .. (err or "unknown error"))
    return
end

-- Test the planning agent
print("\nTesting planning agent:")
result, err = agents.execute("planning-agent", "Create a web scraper that extracts article titles from a news website")
if result then
    print(result)
else
    print("Error: " .. (err or "unknown error"))
end

print("\n=== Summary ===")
print("This example demonstrated proper agent patterns:")
print("1. Research Agent - Combines web_fetch tool with LLM summarization")
print("2. Code Analysis Agent - Uses custom Lua tool + LLM for code review")  
print("3. Planning Agent - Multi-step LLM orchestration for task breakdown")
print("\nKey concepts:")
print("- Agents orchestrate multiple LLM calls")
print("- Agents can use both built-in and custom tools")
print("- Agents provide higher-level functionality than raw LLM calls")