# Spell Development Guide

This guide explains how to create spells (scripts) for go-llmspell in various supported languages.

## What is a Spell?

A spell is a script that leverages LLM capabilities through go-llmspell's bridge system. Spells can:
- Interact with LLMs (chat, completion, streaming)
- Create and use tools
- Define and run agents
- Build complex workflows
- Access curated stdlib functions

## Spell Structure

### Basic Structure

```
my-spell/
├── spell.yaml      # Metadata (optional)
├── main.lua        # Main spell file
├── lib/           # Supporting libraries (optional)
│   └── helpers.lua
└── README.md      # Documentation (optional)
```

### Metadata Format

```yaml
# spell.yaml
name: "code-reviewer"
version: "1.0.0"
engine: "lua"  # or "javascript", "tengo"
description: "Automated code review assistant"
author: "your-name"
license: "MIT"

# Define parameters the spell accepts
parameters:
  - name: "file_path"
    type: "string"
    required: true
    description: "Path to the file to review"
  - name: "style_guide"
    type: "string"
    default: "standard"
    description: "Code style guide to follow"

# Dependencies on other spells or tools
dependencies:
  tools:
    - "code_analyzer"
    - "lint_runner"
  spells:
    - "comment_generator"

# Resource requirements
resources:
  memory: "50MB"
  timeout: "60s"
```

## Lua Spell Development

### Basic Example

```lua
-- hello-world.lua
-- A simple spell that greets the user

-- Get user input or use default
local name = params.name or "World"

-- Use LLM to generate a creative greeting
local prompt = string.format(
    "Generate a creative and friendly greeting for someone named %s. " ..
    "Make it unique and memorable, but keep it under 50 words.",
    name
)

local greeting = llm.chat(prompt)

-- Log the activity
log.info("Generated greeting for", {name = name})

-- Output the result
print(greeting)

-- Return structured data (optional)
return {
    success = true,
    greeting = greeting,
    recipient = name
}
```

### Using Built-in Tools

go-llmspell comes with several built-in tools from the go-llms library:

```lua
-- List available built-in tools
local tools_list = tools.list()
for _, tool in ipairs(tools_list) do
    print(string.format("- %s: %s", tool.name, tool.description))
end

-- Built-in tools include:
-- - web_fetch: Fetches content from a URL

-- Example: Using the web_fetch tool
local result = tools.execute("web_fetch", {
    url = "https://example.com"
})

print("Status:", result.status)
print("Content:", result.content)
```

Note: Additional built-in tools (execute_command, read_file, write_file) are available but disabled by default for security reasons.

### Advanced Example with Custom Tools

```lua
-- web-researcher.lua
-- Research a topic using custom tools

-- Create a custom tool
tools.register("summarize", "Summarizes text content", {
    type = "object",
    properties = {
        text = {
            type = "string",
            description = "Text to summarize"
        },
        max_sentences = {
            type = "number",
            description = "Maximum sentences in summary",
            default = 3
        }
    },
    required = {"text"}
}, function(params)
    -- Use LLM to summarize
    local prompt = string.format(
        "Summarize the following text in %d sentences:\n\n%s",
        params.max_sentences or 3,
        params.text
    )
    return llm.chat(prompt)
end)

-- Use built-in web_fetch with custom summarize
local url = params.url or "https://example.com"

-- Fetch content
local fetch_result = tools.execute("web_fetch", {url = url})

if fetch_result.status == 200 then
    -- Summarize the content
    local summary = tools.execute("summarize", {
        text = fetch_result.content,
        max_sentences = 5
    })
    
    print("Summary of", url)
    print(summary)
else
    print("Failed to fetch URL:", fetch_result.status)
end

-- Create a research agent with the tool
local researcher = agent.create({
    name = "web_researcher",
    tools = {"web_search"},
    system_prompt = [[
        You are a research assistant. When given a topic, search for relevant 
        information and provide a comprehensive summary. Always cite your sources.
    ]]
})

-- Main spell logic
local topic = params.topic or error("Topic parameter is required")

log.info("Starting research on topic", {topic = topic})

-- Run the research
local research_prompt = string.format(
    "Please research the following topic and provide a detailed summary: %s",
    topic
)

local result = researcher.run(research_prompt)

-- Save the results
local timestamp = os.date("%Y%m%d_%H%M%S")
local filename = string.format("research_%s_%s.md", 
    topic:gsub("%s+", "_"):lower(), 
    timestamp
)

fs.write(filename, result)

log.info("Research completed and saved", {file = filename})

return {
    success = true,
    topic = topic,
    summary = result,
    saved_to = filename
}
```

### Working with Workflows

```lua
-- blog-writer.lua
-- Automated blog post creation workflow

-- Define workflow steps
local workflow = workflow.create({
    name = "blog_writer",
    steps = {
        -- Step 1: Research the topic
        {
            name = "research",
            type = "agent",
            agent = agent.create({
                name = "researcher",
                tools = {"web_search"},
                system_prompt = "Research the given topic thoroughly"
            }),
            input = function(context)
                return "Research about: " .. context.topic
            end
        },
        
        -- Step 2: Create outline
        {
            name = "outline",
            type = "llm",
            prompt = function(context)
                return string.format(
                    "Based on this research:\n%s\n\n" ..
                    "Create a detailed blog post outline with 5-7 sections",
                    context.research.output
                )
            end
        },
        
        -- Step 3: Write the blog post
        {
            name = "write",
            type = "agent",
            agent = agent.create({
                name = "writer",
                system_prompt = [[
                    You are a professional blog writer. Write engaging,
                    informative content based on the provided outline.
                ]]
            }),
            input = function(context)
                return string.format(
                    "Write a blog post based on this outline:\n%s",
                    context.outline.output
                )
            end
        },
        
        -- Step 4: Review and edit
        {
            name = "edit",
            type = "llm",
            prompt = function(context)
                return string.format(
                    "Review and edit this blog post for clarity, " ..
                    "grammar, and engagement:\n\n%s",
                    context.write.output
                )
            end
        }
    }
})

-- Execute the workflow
local topic = params.topic or error("Topic parameter required")

local result = workflow.execute({
    topic = topic
})

-- Save the final blog post
local filename = string.format("blog_%s.md", 
    topic:gsub("%s+", "_"):lower()
)

fs.write(filename, result.edit.output)

return {
    success = true,
    topic = topic,
    blog_post = result.edit.output,
    saved_to = filename
}
```

## JavaScript Spell Development

### Basic Example

```javascript
// hello-world.js
// A simple spell that greets the user

async function main(params) {
    // Get user input or use default
    const name = params.name || "World";
    
    // Use LLM to generate a creative greeting
    const prompt = `Generate a creative and friendly greeting for someone named ${name}. 
                   Make it unique and memorable, but keep it under 50 words.`;
    
    const greeting = await llm.chat(prompt);
    
    // Log the activity
    log.info("Generated greeting for", {name});
    
    // Output the result
    console.log(greeting);
    
    // Return structured data
    return {
        success: true,
        greeting,
        recipient: name
    };
}

// Export for the spell runner
module.exports = main;
```

### Agent Example

```javascript
// code-reviewer.js
// Automated code review assistant

async function main(params) {
    const {filePath, styleGuide = "standard"} = params;
    
    if (!filePath) {
        throw new Error("filePath parameter is required");
    }
    
    // Read the code file
    const code = fs.read(filePath);
    
    // Create code review agent
    const reviewer = agent.create({
        name: "code_reviewer",
        systemPrompt: `You are an expert code reviewer. Review code for:
                      - Best practices
                      - Potential bugs
                      - Performance issues
                      - Security concerns
                      - Style guide compliance (${styleGuide})
                      Provide specific, actionable feedback.`
    });
    
    // Perform the review
    const reviewPrompt = `Please review the following code:\n\n${code}`;
    const review = await reviewer.run(reviewPrompt);
    
    // Create a summary report
    const summary = await llm.chat(
        `Summarize this code review in 3-5 bullet points:\n${review}`
    );
    
    // Save the review
    const reportPath = filePath.replace(/\.[^.]+$/, "_review.md");
    fs.write(reportPath, `# Code Review Report\n\n${review}\n\n## Summary\n\n${summary}`);
    
    log.info("Code review completed", {file: filePath, report: reportPath});
    
    return {
        success: true,
        file: filePath,
        review,
        summary,
        reportPath
    };
}

module.exports = main;
```

## Tengo Spell Development

### Basic Example

```go
// hello-world.tengo
// A simple spell that greets the user

fmt := import("fmt")

name := params.name || "World"

// Use LLM to generate a creative greeting
prompt := fmt.sprintf(
    "Generate a creative and friendly greeting for someone named %s. " +
    "Make it unique and memorable, but keep it under 50 words.",
    name
)

greeting := llm.chat(prompt)

// Log the activity
log.info("Generated greeting for " + name)

// Output the result
fmt.println(greeting)

// Return structured data
export {
    success: true,
    greeting: greeting,
    recipient: name
}
```

## Common Patterns

### Error Handling

```lua
-- Lua error handling
local function safe_execute(fn, ...)
    local success, result = pcall(fn, ...)
    if not success then
        log.error("Execution failed", {error = result})
        return nil, result
    end
    return result
end

-- Use it
local result, err = safe_execute(function()
    return llm.chat("Hello")
end)

if err then
    return {success = false, error = err}
end
```

### Configuration Management

```lua
-- Load configuration with defaults
local config = {
    max_retries = params.max_retries or 3,
    timeout = params.timeout or 30,
    model = params.model or "gpt-4",
    temperature = params.temperature or 0.7
}

-- Apply configuration
llm.set_model(config.model)
llm.set_option("temperature", config.temperature)
```

### Progress Reporting

```lua
-- Report progress for long-running operations
local function with_progress(steps)
    local total = #steps
    
    for i, step in ipairs(steps) do
        log.info(string.format("Progress: %d/%d - %s", i, total, step.name))
        
        local result = step.execute()
        
        if not result.success then
            log.error("Step failed", {step = step.name, error = result.error})
            return false
        end
    end
    
    return true
end
```

### Caching Results

```lua
-- Simple caching mechanism
local cache = {}

local function cached_llm_call(prompt)
    local cache_key = crypto.hash(prompt)
    
    if cache[cache_key] then
        log.debug("Cache hit", {key = cache_key})
        return cache[cache_key]
    end
    
    local result = llm.chat(prompt)
    cache[cache_key] = result
    
    return result
end
```

## Best Practices

### 1. Input Validation

Always validate input parameters:

```lua
-- Validate required parameters
assert(params.input, "Input parameter is required")
assert(type(params.max_tokens) == "number", "max_tokens must be a number")

-- Validate with defaults
local config = {
    temperature = tonumber(params.temperature) or 0.7,
    max_tokens = math.min(params.max_tokens or 1000, 4000)
}
```

### 2. Resource Management

Clean up resources properly:

```lua
-- Use finally blocks for cleanup
local file = fs.open("data.txt", "r")
local success, result = pcall(function()
    -- Process file
    return file:read("*all")
end)
file:close()

if not success then
    error(result)
end
```

### 3. Logging

Use appropriate log levels:

```lua
log.debug("Detailed information for debugging")
log.info("General information about spell execution")
log.warn("Warning about potential issues")
log.error("Error that doesn't stop execution")
-- Use error() for fatal errors that should stop execution
```

### 4. Documentation

Document your spells:

```lua
--[[
    Code Review Spell
    
    This spell performs automated code review using AI.
    
    Parameters:
    - file_path (string, required): Path to the file to review
    - style_guide (string, optional): Style guide to follow
    - severity (string, optional): Minimum severity level to report
    
    Returns:
    - success (boolean): Whether the review completed successfully
    - issues (array): List of identified issues
    - summary (string): Summary of the review
]]
```

### 5. Testing

Create test cases for your spells:

```lua
-- test_spell.lua
local function test_basic_functionality()
    local result = spell.execute({
        input = "test input",
        max_tokens = 100
    })
    
    assert(result.success == true, "Spell should succeed")
    assert(result.output ~= nil, "Should produce output")
end

-- Run tests if executed directly
if arg and arg[0]:match("test_spell.lua$") then
    test_basic_functionality()
    print("All tests passed!")
end
```

## Debugging Spells

### Enable Debug Logging

```lua
-- Set debug mode
if params.debug then
    log.set_level("debug")
end

-- Debug helpers
local function dump(value, name)
    log.debug(name or "Value", {
        type = type(value),
        value = json.encode(value)
    })
end
```

### Inspection Tools

```lua
-- Inspect LLM state
log.debug("Current model", {model = llm.get_model()})
log.debug("Available tools", {tools = tool.list()})

-- Trace execution
local function trace(fn, name)
    return function(...)
        log.debug("Calling " .. name, {args = {...}})
        local results = {fn(...)}
        log.debug("Returned from " .. name, {results = results})
        return unpack(results)
    end
end

-- Use tracing
llm.chat = trace(llm.chat, "llm.chat")
```

## Publishing Spells

### 1. Package Structure

```
my-awesome-spell/
├── spell.yaml          # Required metadata
├── main.lua           # Main entry point
├── lib/              # Supporting libraries
├── tests/            # Test files
├── examples/         # Example usage
├── README.md         # Documentation
└── LICENSE           # License file
```

### 2. Documentation Requirements

Your README.md should include:
- Description of what the spell does
- Installation instructions
- Usage examples
- Parameter documentation
- Return value documentation
- Requirements and dependencies

### 3. Version Management

Use semantic versioning in spell.yaml:
- MAJOR: Breaking changes
- MINOR: New features, backward compatible
- PATCH: Bug fixes

### 4. License

Choose an appropriate license and include it in your spell package.

## Spell Examples Repository

Find more spell examples at:
- Basic examples: `spells/examples/basic/`
- Advanced examples: `spells/examples/advanced/`
- Community spells: `spells/community/`

Each example includes:
- Full source code
- Detailed comments
- Test cases
- Usage documentation