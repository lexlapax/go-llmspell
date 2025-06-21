// ABOUTME: Script content generators for different template types and engines.
// ABOUTME: Provides engine-specific implementations for Lua, JavaScript, and Tengo.

package template

import "fmt"

// getBasicScriptContent returns the basic script content based on engine
func (g *Generator) getBasicScriptContent() string {
	return `{{if eq .Engine "lua"}}-- {{.Name}}
-- {{.Description}}

local llm = require("llm")
local json = require("json")

-- Get parameters
local prompt = params.prompt or error("Prompt is required")
local model = params.model or "gpt-3.5-turbo"

-- Create LLM client
local client = llm.new({
    model = model,
    temperature = 0.7
})

-- Send prompt to LLM
print("Sending prompt to " .. model .. "...")
local response, err = client:complete(prompt)

if err then
    error("LLM request failed: " .. tostring(err))
end

-- Print response
print("\nResponse:")
print(response)

-- Return response for further processing
return response
{{else if eq .Engine "javascript"}}// {{.Name}}
// {{.Description}}

const llm = require('llm');

// Get parameters
const prompt = params.prompt || throw new Error('Prompt is required');
const model = params.model || 'gpt-3.5-turbo';

// Create LLM client
const client = llm.new({
    model: model,
    temperature: 0.7
});

// Send prompt to LLM
console.log(` + "`Sending prompt to ${model}...`" + `);
const response = await client.complete(prompt);

// Print response
console.log('\nResponse:');
console.log(response);

// Return response for further processing
return response;
{{else if eq .Engine "tengo"}}// {{.Name}}
// {{.Description}}

llm := import("llm")
fmt := import("fmt")
os := import("os")

// Get parameters
prompt := params.prompt
if !prompt {
    os.exit(1, "Prompt is required")
}
model := params.model || "gpt-3.5-turbo"

// Create LLM client
client := llm.new({
    model: model,
    temperature: 0.7
})

// Send prompt to LLM
fmt.println("Sending prompt to", model, "...")
response := client.complete(prompt)

// Print response
fmt.println("\nResponse:")
fmt.println(response)

// Return response for further processing
return response
{{end}}`
}

// getAdvancedScriptContent returns the advanced script content
func (g *Generator) getAdvancedScriptContent() string {
	return `{{if eq .Engine "lua"}}-- {{.Name}}
-- {{.Description}}

local llm = require("llm")
local state = require("state")
local hooks = require("hooks")
local json = require("json")
local utils = require("lib.utils")
local prompts = require("lib.prompts")

-- Initialize state
local conversation_state = state.new("conversation")

-- Get parameters
local mode = params.mode or "chat"
local input_file = params.input_file
local output_file = params.output_file
local model = params.model or "gpt-4"
local temperature = params.temperature or 0.7

-- Setup hooks
hooks.register("before_llm_call", function(data)
    print("Preparing LLM request...")
    return data
end)

hooks.register("after_llm_call", function(response)
    -- Save to conversation history
    conversation_state:set("last_response", response)
    return response
end)

-- Main function
function main()
    local input_text = ""
    
    -- Read input if file provided
    if input_file then
        input_text = utils.read_file(input_file)
    else
        input_text = params.prompt or error("No input provided")
    end
    
    -- Get appropriate prompt based on mode
    local system_prompt = prompts.get_system_prompt(mode)
    
    -- Create LLM client
    local client = llm.new({
        model = model,
        temperature = temperature,
        system = system_prompt
    })
    
    -- Process based on mode
    local result
    if mode == "chat" then
        result = handle_chat(client, input_text)
    elseif mode == "analyze" then
        result = handle_analyze(client, input_text)
    elseif mode == "summarize" then
        result = handle_summarize(client, input_text)
    else
        error("Unknown mode: " .. mode)
    end
    
    -- Save output if specified
    if output_file then
        utils.write_file(output_file, result)
        print("Output saved to: " .. output_file)
    end
    
    return result
end

function handle_chat(client, input)
    -- Get conversation history
    local history = conversation_state:get("history") or {}
    
    -- Add user message
    table.insert(history, {role = "user", content = input})
    
    -- Get response
    local response = client:chat(history)
    
    -- Update history
    table.insert(history, {role = "assistant", content = response})
    conversation_state:set("history", history)
    
    return response
end

function handle_analyze(client, input)
    local prompt = prompts.get_analysis_prompt(input)
    return client:complete(prompt)
end

function handle_summarize(client, input)
    local prompt = prompts.get_summary_prompt(input)
    return client:complete(prompt)
end

-- Run main function with error handling
local success, result = pcall(main)
if not success then
    print("Error: " .. tostring(result))
    os.exit(1)
end

print("\nResult:")
print(result)
return result
{{else if eq .Engine "javascript"}}// {{.Name}}
// {{.Description}}

const llm = require('llm');
const state = require('state');
const hooks = require('hooks');
const utils = require('./lib/utils');
const prompts = require('./lib/prompts');

// Initialize state
const conversationState = state.new('conversation');

// Get parameters
const mode = params.mode || 'chat';
const inputFile = params.input_file;
const outputFile = params.output_file;
const model = params.model || 'gpt-4';
const temperature = params.temperature || 0.7;

// Setup hooks
hooks.register('before_llm_call', (data) => {
    console.log('Preparing LLM request...');
    return data;
});

hooks.register('after_llm_call', (response) => {
    // Save to conversation history
    conversationState.set('last_response', response);
    return response;
});

// Main function
async function main() {
    let inputText = '';
    
    // Read input if file provided
    if (inputFile) {
        inputText = await utils.readFile(inputFile);
    } else {
        inputText = params.prompt || throw new Error('No input provided');
    }
    
    // Get appropriate prompt based on mode
    const systemPrompt = prompts.getSystemPrompt(mode);
    
    // Create LLM client
    const client = llm.new({
        model: model,
        temperature: temperature,
        system: systemPrompt
    });
    
    // Process based on mode
    let result;
    switch (mode) {
        case 'chat':
            result = await handleChat(client, inputText);
            break;
        case 'analyze':
            result = await handleAnalyze(client, inputText);
            break;
        case 'summarize':
            result = await handleSummarize(client, inputText);
            break;
        default:
            throw new Error(` + "`Unknown mode: ${mode}`" + `);
    }
    
    // Save output if specified
    if (outputFile) {
        await utils.writeFile(outputFile, result);
        console.log(` + "`Output saved to: ${outputFile}`" + `);
    }
    
    return result;
}

async function handleChat(client, input) {
    // Get conversation history
    let history = conversationState.get('history') || [];
    
    // Add user message
    history.push({role: 'user', content: input});
    
    // Get response
    const response = await client.chat(history);
    
    // Update history
    history.push({role: 'assistant', content: response});
    conversationState.set('history', history);
    
    return response;
}

async function handleAnalyze(client, input) {
    const prompt = prompts.getAnalysisPrompt(input);
    return await client.complete(prompt);
}

async function handleSummarize(client, input) {
    const prompt = prompts.getSummaryPrompt(input);
    return await client.complete(prompt);
}

// Run main function with error handling
try {
    const result = await main();
    console.log('\nResult:');
    console.log(result);
    return result;
} catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
}
{{else if eq .Engine "tengo"}}// {{.Name}}
// {{.Description}}

llm := import("llm")
state := import("state")
hooks := import("hooks")
fmt := import("fmt")
os := import("os")
utils := import("lib/utils")
prompts := import("lib/prompts")

// Initialize state
conversation_state := state.new("conversation")

// Get parameters
mode := params.mode || "chat"
input_file := params.input_file
output_file := params.output_file
model := params.model || "gpt-4"
temperature := params.temperature || 0.7

// Setup hooks
hooks.register("before_llm_call", func(data) {
    fmt.println("Preparing LLM request...")
    return data
})

hooks.register("after_llm_call", func(response) {
    // Save to conversation history
    conversation_state.set("last_response", response)
    return response
})

// Process based on mode
input_text := ""
if input_file {
    input_text = utils.read_file(input_file)
} else if params.prompt {
    input_text = params.prompt
} else {
    os.exit(1, "No input provided")
}

// Get appropriate prompt based on mode
system_prompt := prompts.get_system_prompt(mode)

// Create LLM client
client := llm.new({
    model: model,
    temperature: temperature,
    system: system_prompt
})

// Process based on mode
result := undefined
if mode == "chat" {
    // Get conversation history
    history := conversation_state.get("history") || []
    
    // Add user message
    history = append(history, {role: "user", content: input_text})
    
    // Get response
    result = client.chat(history)
    
    // Update history
    history = append(history, {role: "assistant", content: result})
    conversation_state.set("history", history)
} else if mode == "analyze" {
    prompt := prompts.get_analysis_prompt(input_text)
    result = client.complete(prompt)
} else if mode == "summarize" {
    prompt := prompts.get_summary_prompt(input_text)
    result = client.complete(prompt)
} else {
    os.exit(1, "Unknown mode: " + mode)
}

// Save output if specified
if output_file {
    utils.write_file(output_file, result)
    fmt.println("Output saved to:", output_file)
}

fmt.println("\nResult:")
fmt.println(result)
return result
{{end}}`
}

// getUtilsScriptContent returns utility functions
func (g *Generator) getUtilsScriptContent() string {
	return `{{if eq .Engine "lua"}}-- Utility functions

local utils = {}

function utils.read_file(path)
    local file = io.open(path, "r")
    if not file then
        error("Could not open file: " .. path)
    end
    local content = file:read("*all")
    file:close()
    return content
end

function utils.write_file(path, content)
    local file = io.open(path, "w")
    if not file then
        error("Could not open file for writing: " .. path)
    end
    file:write(content)
    file:close()
end

function utils.split(str, delimiter)
    local result = {}
    for match in (str..delimiter):gmatch("(.-)"..delimiter) do
        table.insert(result, match)
    end
    return result
end

function utils.trim(str)
    return str:match("^%s*(.-)%s*$")
end

return utils
{{else if eq .Engine "javascript"}}// Utility functions

const fs = require('fs').promises;

async function readFile(path) {
    try {
        return await fs.readFile(path, 'utf8');
    } catch (error) {
        throw new Error(` + "`Could not open file: ${path}`" + `);
    }
}

async function writeFile(path, content) {
    try {
        await fs.writeFile(path, content, 'utf8');
    } catch (error) {
        throw new Error(` + "`Could not write file: ${path}`" + `);
    }
}

function split(str, delimiter) {
    return str.split(delimiter);
}

function trim(str) {
    return str.trim();
}

module.exports = {
    readFile,
    writeFile,
    split,
    trim
};
{{else if eq .Engine "tengo"}}// Utility functions

os := import("os")
text := import("text")

read_file := func(path) {
    bytes := os.read_file(path)
    if is_error(bytes) {
        error("Could not open file: " + path)
    }
    return string(bytes)
}

write_file := func(path, content) {
    err := os.write_file(path, bytes(content))
    if is_error(err) {
        error("Could not write file: " + path)
    }
}

split := func(str, delimiter) {
    return text.split(str, delimiter)
}

trim := func(str) {
    return text.trim(str, " \t\n\r")
}

export {
    read_file: read_file,
    write_file: write_file,
    split: split,
    trim: trim
}
{{end}}`
}

// getPromptsScriptContent returns prompt templates
func (g *Generator) getPromptsScriptContent() string {
	return `{{if eq .Engine "lua"}}-- Prompt templates

local prompts = {}

prompts.system_prompts = {
    chat = "You are a helpful assistant. Engage in natural conversation and provide thoughtful responses.",
    analyze = "You are an expert analyst. Provide detailed analysis with insights and recommendations.",
    summarize = "You are a summarization expert. Create concise, accurate summaries that capture key points."
}

function prompts.get_system_prompt(mode)
    return prompts.system_prompts[mode] or prompts.system_prompts.chat
end

function prompts.get_analysis_prompt(text)
    return string.format([[
Please analyze the following text and provide:
1. Key themes and topics
2. Important insights
3. Potential concerns or issues
4. Recommendations

Text to analyze:
%s
]], text)
end

function prompts.get_summary_prompt(text)
    return string.format([[
Please provide a comprehensive summary of the following text.
Include main points, key findings, and important details.
Keep the summary concise but complete.

Text to summarize:
%s
]], text)
end

return prompts
{{else if eq .Engine "javascript"}}// Prompt templates

const systemPrompts = {
    chat: "You are a helpful assistant. Engage in natural conversation and provide thoughtful responses.",
    analyze: "You are an expert analyst. Provide detailed analysis with insights and recommendations.",
    summarize: "You are a summarization expert. Create concise, accurate summaries that capture key points."
};

function getSystemPrompt(mode) {
    return systemPrompts[mode] || systemPrompts.chat;
}

function getAnalysisPrompt(text) {
    return ` + "`" + `
Please analyze the following text and provide:
1. Key themes and topics
2. Important insights
3. Potential concerns or issues
4. Recommendations

Text to analyze:
${text}
` + "`" + `;
}

function getSummaryPrompt(text) {
    return ` + "`" + `
Please provide a comprehensive summary of the following text.
Include main points, key findings, and important details.
Keep the summary concise but complete.

Text to summarize:
${text}
` + "`" + `;
}

module.exports = {
    getSystemPrompt,
    getAnalysisPrompt,
    getSummaryPrompt
};
{{else if eq .Engine "tengo"}}// Prompt templates

system_prompts := {
    chat: "You are a helpful assistant. Engage in natural conversation and provide thoughtful responses.",
    analyze: "You are an expert analyst. Provide detailed analysis with insights and recommendations.",
    summarize: "You are a summarization expert. Create concise, accurate summaries that capture key points."
}

get_system_prompt := func(mode) {
    return system_prompts[mode] || system_prompts.chat
}

get_analysis_prompt := func(text) {
    return format("Please analyze the following text and provide:\n1. Key themes and topics\n2. Important insights\n3. Potential concerns or issues\n4. Recommendations\n\nText to analyze:\n%s", text)
}

get_summary_prompt := func(text) {
    return format("Please provide a comprehensive summary of the following text.\nInclude main points, key findings, and important details.\nKeep the summary concise but complete.\n\nText to summarize:\n%s", text)
}

export {
    get_system_prompt: get_system_prompt,
    get_analysis_prompt: get_analysis_prompt,
    get_summary_prompt: get_summary_prompt
}
{{end}}`
}

// Additional script content methods for agent, workflow, and interactive templates
func (g *Generator) getAgentScriptContent() string {
	return fmt.Sprintf(`{{if eq .Engine "lua"}}%s{{else if eq .Engine "javascript"}}%s{{else if eq .Engine "tengo"}}%s{{end}}`,
		g.getLuaAgentScript(),
		g.getJavaScriptAgentScript(),
		g.getTengoAgentScript())
}

func (g *Generator) getLuaAgentScript() string {
	return `-- {{.Name}} - Agent-based spell
-- {{.Description}}

local agent = require("agent")
local tools = require("tools")
local json = require("json")

-- Load tools
local calculator = require("tools.calculator")
local web_search = require("tools.web_search")
local file_reader = require("tools.file_reader")

-- Get parameters
local task = params.task or error("Task is required")
local enabled_tools = params.tools or {"calculator", "web_search", "file_reader"}
local max_iterations = params.max_iterations or 10

-- Register tools
local tool_registry = tools.new()

for _, tool_name in ipairs(enabled_tools) do
    if tool_name == "calculator" then
        tool_registry:register("calculator", calculator)
    elseif tool_name == "web_search" then
        tool_registry:register("web_search", web_search)
    elseif tool_name == "file_reader" then
        tool_registry:register("file_reader", file_reader)
    end
end

-- Create agent
local agent_instance = agent.new({
    model = "gpt-4",
    tools = tool_registry,
    max_iterations = max_iterations,
    system = "You are a helpful AI assistant with access to tools. Use the tools when needed to complete tasks."
})

-- Execute task
print("Starting agent task: " .. task)
print("Enabled tools: " .. table.concat(enabled_tools, ", "))
print("Max iterations: " .. max_iterations)
print("\n" .. string.rep("-", 50) .. "\n")

local result, err = agent_instance:run(task)

if err then
    error("Agent execution failed: " .. tostring(err))
end

print("\n" .. string.rep("-", 50) .. "\n")
print("Task completed!")
print("\nFinal result:")
print(json.encode(result, {indent = true}))

return result`
}

func (g *Generator) getJavaScriptAgentScript() string {
	return `// {{.Name}} - Agent-based spell
// {{.Description}}

const agent = require('agent');
const tools = require('tools');

// Load tools
const calculator = require('./tools/calculator');
const webSearch = require('./tools/web_search');
const fileReader = require('./tools/file_reader');

// Get parameters
const task = params.task || throw new Error('Task is required');
const enabledTools = params.tools || ['calculator', 'web_search', 'file_reader'];
const maxIterations = params.max_iterations || 10;

// Register tools
const toolRegistry = tools.new();

for (const toolName of enabledTools) {
    switch (toolName) {
        case 'calculator':
            toolRegistry.register('calculator', calculator);
            break;
        case 'web_search':
            toolRegistry.register('web_search', webSearch);
            break;
        case 'file_reader':
            toolRegistry.register('file_reader', fileReader);
            break;
    }
}

// Create agent
const agentInstance = agent.new({
    model: 'gpt-4',
    tools: toolRegistry,
    maxIterations: maxIterations,
    system: 'You are a helpful AI assistant with access to tools. Use the tools when needed to complete tasks.'
});

// Execute task
console.log(` + "`Starting agent task: ${task}`" + `);
console.log(` + "`Enabled tools: ${enabledTools.join(', ')}`" + `);
console.log(` + "`Max iterations: ${maxIterations}`" + `);
console.log('\n' + '-'.repeat(50) + '\n');

const result = await agentInstance.run(task);

console.log('\n' + '-'.repeat(50) + '\n');
console.log('Task completed!');
console.log('\nFinal result:');
console.log(JSON.stringify(result, null, 2));

return result;`
}

func (g *Generator) getTengoAgentScript() string {
	return `// {{.Name}} - Agent-based spell
// {{.Description}}

agent := import("agent")
tools := import("tools")
fmt := import("fmt")
json := import("json")
os := import("os")

// Load tools
calculator := import("tools/calculator")
web_search := import("tools/web_search")
file_reader := import("tools/file_reader")

// Get parameters
task := params.task
if !task {
    os.exit(1, "Task is required")
}
enabled_tools := params.tools || ["calculator", "web_search", "file_reader"]
max_iterations := params.max_iterations || 10

// Register tools
tool_registry := tools.new()

for tool_name in enabled_tools {
    if tool_name == "calculator" {
        tool_registry.register("calculator", calculator)
    } else if tool_name == "web_search" {
        tool_registry.register("web_search", web_search)
    } else if tool_name == "file_reader" {
        tool_registry.register("file_reader", file_reader)
    }
}

// Create agent
agent_instance := agent.new({
    model: "gpt-4",
    tools: tool_registry,
    max_iterations: max_iterations,
    system: "You are a helpful AI assistant with access to tools. Use the tools when needed to complete tasks."
})

// Execute task
fmt.println("Starting agent task:", task)
fmt.println("Enabled tools:", enabled_tools)
fmt.println("Max iterations:", max_iterations)
fmt.println("\n" + string(bytes("-" * 50)) + "\n")

result := agent_instance.run(task)

fmt.println("\n" + string(bytes("-" * 50)) + "\n")
fmt.println("Task completed!")
fmt.println("\nFinal result:")
fmt.println(json.encode(result))

return result`
}

// Tool content generators
func (g *Generator) getCalculatorToolContent() string {
	return `{{if eq .Engine "lua"}}-- Calculator tool

local tool = {}

tool.name = "calculator"
tool.description = "Performs mathematical calculations"

tool.parameters = {
    {
        name = "expression",
        type = "string",
        description = "Mathematical expression to evaluate",
        required = true
    }
}

function tool.execute(params)
    local expression = params.expression
    
    -- Basic safety check (in production, use proper sandboxing)
    if expression:match("[^%d%+%-%*/%(%)%.%s]") then
        return nil, "Invalid characters in expression"
    end
    
    -- Evaluate expression
    local fn, err = load("return " .. expression)
    if not fn then
        return nil, "Invalid expression: " .. tostring(err)
    end
    
    local success, result = pcall(fn)
    if not success then
        return nil, "Calculation error: " .. tostring(result)
    end
    
    return {result = result, expression = expression}
end

return tool
{{else if eq .Engine "javascript"}}// Calculator tool

module.exports = {
    name: 'calculator',
    description: 'Performs mathematical calculations',
    
    parameters: [
        {
            name: 'expression',
            type: 'string',
            description: 'Mathematical expression to evaluate',
            required: true
        }
    ],
    
    execute: async (params) => {
        const expression = params.expression;
        
        // Basic safety check (in production, use proper sandboxing)
        if (!/^[\d\+\-\*\/\(\)\.\s]+$/.test(expression)) {
            throw new Error('Invalid characters in expression');
        }
        
        try {
            // Evaluate expression (use math parser in production)
            const result = eval(expression);
            return { result, expression };
        } catch (error) {
            throw new Error(` + "`Calculation error: ${error.message}`" + `);
        }
    }
};
{{else if eq .Engine "tengo"}}// Calculator tool

name := "calculator"
description := "Performs mathematical calculations"

parameters := [
    {
        name: "expression",
        type: "string",
        description: "Mathematical expression to evaluate",
        required: true
    }
]

execute := func(params) {
    expression := params.expression
    
    // Note: Tengo doesn't have eval, so we'd need a math parser
    // This is a simplified example
    return {
        result: "Calculation would be performed here",
        expression: expression
    }
}

export {
    name: name,
    description: description,
    parameters: parameters,
    execute: execute
}
{{end}}`
}

func (g *Generator) getWebSearchToolContent() string {
	return `{{if eq .Engine "lua"}}-- Web search tool (mock implementation)

local tool = {}

tool.name = "web_search"
tool.description = "Searches the web for information"

tool.parameters = {
    {
        name = "query",
        type = "string",
        description = "Search query",
        required = true
    },
    {
        name = "max_results",
        type = "number",
        description = "Maximum number of results",
        required = false,
        default = 5
    }
}

function tool.execute(params)
    local query = params.query
    local max_results = params.max_results or 5
    
    -- Mock search results (in production, use actual API)
    local results = {}
    for i = 1, max_results do
        table.insert(results, {
            title = "Result " .. i .. " for: " .. query,
            url = "https://example.com/result" .. i,
            snippet = "This is a mock search result for the query: " .. query
        })
    end
    
    return {
        query = query,
        results = results,
        count = #results
    }
end

return tool
{{else if eq .Engine "javascript"}}// Web search tool (mock implementation)

module.exports = {
    name: 'web_search',
    description: 'Searches the web for information',
    
    parameters: [
        {
            name: 'query',
            type: 'string',
            description: 'Search query',
            required: true
        },
        {
            name: 'max_results',
            type: 'number',
            description: 'Maximum number of results',
            required: false,
            default: 5
        }
    ],
    
    execute: async (params) => {
        const query = params.query;
        const maxResults = params.max_results || 5;
        
        // Mock search results (in production, use actual API)
        const results = [];
        for (let i = 1; i <= maxResults; i++) {
            results.push({
                title: ` + "`Result ${i} for: ${query}`" + `,
                url: ` + "`https://example.com/result${i}`" + `,
                snippet: ` + "`This is a mock search result for the query: ${query}`" + `
            });
        }
        
        return {
            query: query,
            results: results,
            count: results.length
        };
    }
};
{{else if eq .Engine "tengo"}}// Web search tool (mock implementation)

name := "web_search"
description := "Searches the web for information"

parameters := [
    {
        name: "query",
        type: "string",
        description: "Search query",
        required: true
    },
    {
        name: "max_results",
        type: "number",
        description: "Maximum number of results",
        required: false,
        default: 5
    }
]

execute := func(params) {
    query := params.query
    max_results := params.max_results || 5
    
    // Mock search results (in production, use actual API)
    results := []
    for i := 1; i <= max_results; i++ {
        results = append(results, {
            title: "Result " + string(i) + " for: " + query,
            url: "https://example.com/result" + string(i),
            snippet: "This is a mock search result for the query: " + query
        })
    }
    
    return {
        query: query,
        results: results,
        count: len(results)
    }
}

export {
    name: name,
    description: description,
    parameters: parameters,
    execute: execute
}
{{end}}`
}

func (g *Generator) getFileReaderToolContent() string {
	return `{{if eq .Engine "lua"}}-- File reader tool

local tool = {}

tool.name = "file_reader"
tool.description = "Reads and processes local files"

tool.parameters = {
    {
        name = "path",
        type = "string",
        description = "File path to read",
        required = true
    },
    {
        name = "encoding",
        type = "string",
        description = "File encoding",
        required = false,
        default = "utf-8"
    }
}

function tool.execute(params)
    local path = params.path
    local encoding = params.encoding or "utf-8"
    
    -- Read file
    local file, err = io.open(path, "r")
    if not file then
        return nil, "Could not open file: " .. tostring(err)
    end
    
    local content = file:read("*all")
    file:close()
    
    -- Get file info
    local info = {
        path = path,
        size = #content,
        lines = 0
    }
    
    -- Count lines
    for _ in content:gmatch("\n") do
        info.lines = info.lines + 1
    end
    
    return {
        content = content,
        info = info
    }
end

return tool
{{else if eq .Engine "javascript"}}// File reader tool

const fs = require('fs').promises;
const path = require('path');

module.exports = {
    name: 'file_reader',
    description: 'Reads and processes local files',
    
    parameters: [
        {
            name: 'path',
            type: 'string',
            description: 'File path to read',
            required: true
        },
        {
            name: 'encoding',
            type: 'string',
            description: 'File encoding',
            required: false,
            default: 'utf-8'
        }
    ],
    
    execute: async (params) => {
        const filePath = params.path;
        const encoding = params.encoding || 'utf-8';
        
        try {
            // Read file
            const content = await fs.readFile(filePath, encoding);
            
            // Get file info
            const stats = await fs.stat(filePath);
            const info = {
                path: filePath,
                size: stats.size,
                lines: content.split('\n').length
            };
            
            return {
                content: content,
                info: info
            };
        } catch (error) {
            throw new Error(` + "`Could not read file: ${error.message}`" + `);
        }
    }
};
{{else if eq .Engine "tengo"}}// File reader tool

os := import("os")
text := import("text")

name := "file_reader"
description := "Reads and processes local files"

parameters := [
    {
        name: "path",
        type: "string",
        description: "File path to read",
        required: true
    },
    {
        name: "encoding",
        type: "string",
        description: "File encoding",
        required: false,
        default: "utf-8"
    }
]

execute := func(params) {
    path := params.path
    
    // Read file
    bytes := os.read_file(path)
    if is_error(bytes) {
        error("Could not open file: " + string(bytes))
    }
    
    content := string(bytes)
    
    // Get file info
    info := {
        path: path,
        size: len(content),
        lines: len(text.split(content, "\n"))
    }
    
    return {
        content: content,
        info: info
    }
}

export {
    name: name,
    description: description,
    parameters: parameters,
    execute: execute
}
{{end}}`
}

// Workflow template script generators
func (g *Generator) getWorkflowScriptContent() string {
	return `{{if eq .Engine "lua"}}-- {{.Name}} - Workflow-based spell
-- {{.Description}}

local workflow = require("workflow")
local state = require("state")
local events = require("events")

-- Load workflows
local workflows = {
    process_document = require("workflows.process_document"),
    generate_report = require("workflows.generate_report"),
    analyze_data = require("workflows.analyze_data")
}

-- Get parameters
local workflow_name = params.workflow or "process_document"
local input = params.input or error("Input is required")
local output_dir = params.output_dir or "./output"

-- Validate workflow
local selected_workflow = workflows[workflow_name]
if not selected_workflow then
    error("Unknown workflow: " .. workflow_name)
end

-- Create output directory
os.execute("mkdir -p " .. output_dir)

-- Initialize workflow state
local workflow_state = state.new("workflow_" .. workflow_name)
workflow_state:set("input", input)
workflow_state:set("output_dir", output_dir)
workflow_state:set("start_time", os.time())

-- Setup event handlers
events.on("workflow:step:start", function(data)
    print(string.format("\n[Step %d/%d] %s", data.current, data.total, data.name))
    print(string.rep("-", 50))
end)

events.on("workflow:step:complete", function(data)
    print(string.format("✓ Step completed in %.2fs", data.duration))
    workflow_state:set("last_step", data.name)
    workflow_state:set("last_result", data.result)
end)

events.on("workflow:error", function(data)
    print(string.format("✗ Error in step '%s': %s", data.step, data.error))
end)

-- Execute workflow
print("Starting workflow: " .. workflow_name)
print("Input: " .. input)
print("Output directory: " .. output_dir)
print("\n" .. string.rep("=", 60) .. "\n")

local success, result = pcall(function()
    return selected_workflow.execute(workflow_state)
end)

if not success then
    print("\n" .. string.rep("=", 60) .. "\n")
    print("Workflow failed: " .. tostring(result))
    os.exit(1)
end

-- Save final results
local duration = os.time() - workflow_state:get("start_time")
print("\n" .. string.rep("=", 60) .. "\n")
print(string.format("Workflow completed in %d seconds", duration))
print("Results saved to: " .. output_dir)

return result
{{else if eq .Engine "javascript"}}// {{.Name}} - Workflow-based spell
// {{.Description}}

const workflow = require('workflow');
const state = require('state');
const events = require('events');
const fs = require('fs').promises;

// Load workflows
const workflows = {
    process_document: require('./workflows/process_document'),
    generate_report: require('./workflows/generate_report'),
    analyze_data: require('./workflows/analyze_data')
};

// Get parameters
const workflowName = params.workflow || 'process_document';
const input = params.input || throw new Error('Input is required');
const outputDir = params.output_dir || './output';

// Validate workflow
const selectedWorkflow = workflows[workflowName];
if (!selectedWorkflow) {
    throw new Error(` + "`Unknown workflow: ${workflowName}`" + `);
}

// Create output directory
await fs.mkdir(outputDir, { recursive: true });

// Initialize workflow state
const workflowState = state.new(` + "`workflow_${workflowName}`" + `);
workflowState.set('input', input);
workflowState.set('output_dir', outputDir);
workflowState.set('start_time', Date.now());

// Setup event handlers
events.on('workflow:step:start', (data) => {
    console.log(` + "`\n[Step ${data.current}/${data.total}] ${data.name}`" + `);
    console.log('-'.repeat(50));
});

events.on('workflow:step:complete', (data) => {
    console.log(` + "`✓ Step completed in ${data.duration.toFixed(2)}s`" + `);
    workflowState.set('last_step', data.name);
    workflowState.set('last_result', data.result);
});

events.on('workflow:error', (data) => {
    console.log(` + "`✗ Error in step '${data.step}': ${data.error}`" + `);
});

// Execute workflow
console.log(` + "`Starting workflow: ${workflowName}`" + `);
console.log(` + "`Input: ${input}`" + `);
console.log(` + "`Output directory: ${outputDir}`" + `);
console.log('\n' + '='.repeat(60) + '\n');

try {
    const result = await selectedWorkflow.execute(workflowState);
    
    // Save final results
    const duration = (Date.now() - workflowState.get('start_time')) / 1000;
    console.log('\n' + '='.repeat(60) + '\n');
    console.log(` + "`Workflow completed in ${duration.toFixed(1)} seconds`" + `);
    console.log(` + "`Results saved to: ${outputDir}`" + `);
    
    return result;
} catch (error) {
    console.log('\n' + '='.repeat(60) + '\n');
    console.log(` + "`Workflow failed: ${error.message}`" + `);
    process.exit(1);
}
{{else if eq .Engine "tengo"}}// {{.Name}} - Workflow-based spell
// {{.Description}}

workflow := import("workflow")
state := import("state")
events := import("events")
fmt := import("fmt")
os := import("os")
times := import("times")

// Load workflows
workflows := {
    process_document: import("workflows/process_document"),
    generate_report: import("workflows/generate_report"),
    analyze_data: import("workflows/analyze_data")
}

// Get parameters
workflow_name := params.workflow || "process_document"
input := params.input
if !input {
    os.exit(1, "Input is required")
}
output_dir := params.output_dir || "./output"

// Validate workflow
selected_workflow := workflows[workflow_name]
if !selected_workflow {
    os.exit(1, "Unknown workflow: " + workflow_name)
}

// Create output directory
os.mkdir_all(output_dir, 0755)

// Initialize workflow state
workflow_state := state.new("workflow_" + workflow_name)
workflow_state.set("input", input)
workflow_state.set("output_dir", output_dir)
workflow_state.set("start_time", times.now())

// Setup event handlers
events.on("workflow:step:start", func(data) {
    fmt.printf("\n[Step %d/%d] %s\n", data.current, data.total, data.name)
    fmt.println(string(bytes("-" * 50)))
})

events.on("workflow:step:complete", func(data) {
    fmt.printf("✓ Step completed in %.2fs\n", data.duration)
    workflow_state.set("last_step", data.name)
    workflow_state.set("last_result", data.result)
})

events.on("workflow:error", func(data) {
    fmt.printf("✗ Error in step '%s': %s\n", data.step, data.error)
})

// Execute workflow
fmt.println("Starting workflow:", workflow_name)
fmt.println("Input:", input)
fmt.println("Output directory:", output_dir)
fmt.println("\n" + string(bytes("=" * 60)) + "\n")

result := selected_workflow.execute(workflow_state)

// Save final results
duration := times.since(workflow_state.get("start_time"))
fmt.println("\n" + string(bytes("=" * 60)) + "\n")
fmt.printf("Workflow completed in %d seconds\n", duration)
fmt.println("Results saved to:", output_dir)

return result
{{end}}`
}

// Workflow script generators
func (g *Generator) getProcessDocumentWorkflow() string {
	return `{{if eq .Engine "lua"}}-- Process Document Workflow

local llm = require("llm")
local json = require("json")
local events = require("events")

local workflow = {}

workflow.name = "Process Document"
workflow.description = "Extract, analyze, and summarize documents"

workflow.steps = {
    {name = "Extract Text", handler = "extract_text"},
    {name = "Analyze Content", handler = "analyze_content"},
    {name = "Generate Summary", handler = "generate_summary"},
    {name = "Create Insights", handler = "create_insights"}
}

function workflow.execute(state)
    local results = {}
    local total_steps = #workflow.steps
    
    for i, step in ipairs(workflow.steps) do
        events.emit("workflow:step:start", {
            current = i,
            total = total_steps,
            name = step.name
        })
        
        local start_time = os.clock()
        local handler = workflow[step.handler]
        
        if not handler then
            error("Missing handler: " .. step.handler)
        end
        
        local success, result = pcall(handler, state, results)
        
        if not success then
            events.emit("workflow:error", {
                step = step.name,
                error = tostring(result)
            })
            error("Step failed: " .. step.name)
        end
        
        results[step.handler] = result
        
        events.emit("workflow:step:complete", {
            current = i,
            total = total_steps,
            name = step.name,
            duration = os.clock() - start_time,
            result = result
        })
    end
    
    -- Save final results
    local output_dir = state:get("output_dir")
    local output_file = output_dir .. "/document_analysis.json"
    
    local file = io.open(output_file, "w")
    file:write(json.encode(results, {indent = true}))
    file:close()
    
    return results
end

function workflow.extract_text(state, results)
    local input = state:get("input")
    
    -- In a real implementation, this would extract text from PDF, DOCX, etc.
    -- For now, we'll assume the input is already text or a text file
    local content
    
    local file = io.open(input, "r")
    if file then
        content = file:read("*all")
        file:close()
    else
        content = input -- Assume direct text input
    end
    
    return {
        content = content,
        length = #content,
        words = #(content:gmatch("%S+"))
    }
end

function workflow.analyze_content(state, results)
    local content = results.extract_text.content
    
    local client = llm.new({model = "gpt-4"})
    
    local prompt = [[
Analyze the following document and provide:
1. Main topics and themes
2. Key entities (people, organizations, places)
3. Sentiment analysis
4. Document type classification

Document:
]] .. content

    local analysis = client:complete(prompt)
    
    return {
        analysis = analysis,
        timestamp = os.date()
    }
end

function workflow.generate_summary(state, results)
    local content = results.extract_text.content
    
    local client = llm.new({model = "gpt-4"})
    
    local prompt = [[
Create a comprehensive summary of this document.
Include:
- Executive summary (2-3 sentences)
- Key points (bullet list)
- Important details
- Conclusions

Document:
]] .. content

    local summary = client:complete(prompt)
    
    return {
        summary = summary
    }
end

function workflow.create_insights(state, results)
    local analysis = results.analyze_content.analysis
    local summary = results.generate_summary.summary
    
    local client = llm.new({model = "gpt-4"})
    
    local prompt = [[
Based on the analysis and summary, provide:
1. Key insights and implications
2. Actionable recommendations
3. Areas requiring further investigation
4. Potential risks or concerns

Analysis:
]] .. analysis .. [[

Summary:
]] .. summary

    local insights = client:complete(prompt)
    
    return {
        insights = insights,
        generated_at = os.date()
    }
end

return workflow
{{else if eq .Engine "javascript"}}// Process Document Workflow

const llm = require('llm');
const events = require('events');
const fs = require('fs').promises;

const workflow = {
    name: 'Process Document',
    description: 'Extract, analyze, and summarize documents',
    
    steps: [
        {name: 'Extract Text', handler: 'extractText'},
        {name: 'Analyze Content', handler: 'analyzeContent'},
        {name: 'Generate Summary', handler: 'generateSummary'},
        {name: 'Create Insights', handler: 'createInsights'}
    ],
    
    async execute(state) {
        const results = {};
        const totalSteps = this.steps.length;
        
        for (let i = 0; i < this.steps.length; i++) {
            const step = this.steps[i];
            
            events.emit('workflow:step:start', {
                current: i + 1,
                total: totalSteps,
                name: step.name
            });
            
            const startTime = Date.now();
            
            try {
                const handler = this[step.handler];
                if (!handler) {
                    throw new Error(` + "`Missing handler: ${step.handler}`" + `);
                }
                
                const result = await handler.call(this, state, results);
                results[step.handler] = result;
                
                events.emit('workflow:step:complete', {
                    current: i + 1,
                    total: totalSteps,
                    name: step.name,
                    duration: (Date.now() - startTime) / 1000,
                    result: result
                });
            } catch (error) {
                events.emit('workflow:error', {
                    step: step.name,
                    error: error.message
                });
                throw new Error(` + "`Step failed: ${step.name}`" + `);
            }
        }
        
        // Save final results
        const outputDir = state.get('output_dir');
        const outputFile = ` + "`${outputDir}/document_analysis.json`" + `;
        
        await fs.writeFile(outputFile, JSON.stringify(results, null, 2));
        
        return results;
    },
    
    async extractText(state, results) {
        const input = state.get('input');
        
        // In a real implementation, this would extract text from PDF, DOCX, etc.
        let content;
        
        try {
            content = await fs.readFile(input, 'utf8');
        } catch {
            content = input; // Assume direct text input
        }
        
        return {
            content: content,
            length: content.length,
            words: content.split(/\s+/).length
        };
    },
    
    async analyzeContent(state, results) {
        const content = results.extractText.content;
        
        const client = llm.new({model: 'gpt-4'});
        
        const prompt = ` + "`" + `
Analyze the following document and provide:
1. Main topics and themes
2. Key entities (people, organizations, places)
3. Sentiment analysis
4. Document type classification

Document:
${content}` + "`" + `;

        const analysis = await client.complete(prompt);
        
        return {
            analysis: analysis,
            timestamp: new Date().toISOString()
        };
    },
    
    async generateSummary(state, results) {
        const content = results.extractText.content;
        
        const client = llm.new({model: 'gpt-4'});
        
        const prompt = ` + "`" + `
Create a comprehensive summary of this document.
Include:
- Executive summary (2-3 sentences)
- Key points (bullet list)
- Important details
- Conclusions

Document:
${content}` + "`" + `;

        const summary = await client.complete(prompt);
        
        return {
            summary: summary
        };
    },
    
    async createInsights(state, results) {
        const analysis = results.analyzeContent.analysis;
        const summary = results.generateSummary.summary;
        
        const client = llm.new({model: 'gpt-4'});
        
        const prompt = ` + "`" + `
Based on the analysis and summary, provide:
1. Key insights and implications
2. Actionable recommendations
3. Areas requiring further investigation
4. Potential risks or concerns

Analysis:
${analysis}

Summary:
${summary}` + "`" + `;

        const insights = await client.complete(prompt);
        
        return {
            insights: insights,
            generated_at: new Date().toISOString()
        };
    }
};

module.exports = workflow;
{{else if eq .Engine "tengo"}}// Process Document Workflow

llm := import("llm")
json := import("json")
events := import("events")
os := import("os")
text := import("text")
times := import("times")

name := "Process Document"
description := "Extract, analyze, and summarize documents"

steps := [
    {name: "Extract Text", handler: "extract_text"},
    {name: "Analyze Content", handler: "analyze_content"},
    {name: "Generate Summary", handler: "generate_summary"},
    {name: "Create Insights", handler: "create_insights"}
]

execute := func(state) {
    results := {}
    total_steps := len(steps)
    
    for i, step in steps {
        events.emit("workflow:step:start", {
            current: i + 1,
            total: total_steps,
            name: step.name
        })
        
        start_time := times.now()
        
        // Execute step handler
        handler := undefined
        if step.handler == "extract_text" {
            handler = extract_text
        } else if step.handler == "analyze_content" {
            handler = analyze_content
        } else if step.handler == "generate_summary" {
            handler = generate_summary
        } else if step.handler == "create_insights" {
            handler = create_insights
        }
        
        if !handler {
            error("Missing handler: " + step.handler)
        }
        
        result := handler(state, results)
        results[step.handler] = result
        
        events.emit("workflow:step:complete", {
            current: i + 1,
            total: total_steps,
            name: step.name,
            duration: times.since(start_time),
            result: result
        })
    }
    
    // Save final results
    output_dir := state.get("output_dir")
    output_file := output_dir + "/document_analysis.json"
    
    os.write_file(output_file, bytes(json.encode(results)))
    
    return results
}

extract_text := func(state, results) {
    input := state.get("input")
    
    // Try to read as file
    content_bytes := os.read_file(input)
    content := undefined
    
    if is_error(content_bytes) {
        content = input // Assume direct text input
    } else {
        content = string(content_bytes)
    }
    
    return {
        content: content,
        length: len(content),
        words: len(text.split(content, " "))
    }
}

analyze_content := func(state, results) {
    content := results.extract_text.content
    
    client := llm.new({model: "gpt-4"})
    
    prompt := "Analyze the following document and provide:\n" +
              "1. Main topics and themes\n" +
              "2. Key entities (people, organizations, places)\n" +
              "3. Sentiment analysis\n" +
              "4. Document type classification\n\n" +
              "Document:\n" + content
    
    analysis := client.complete(prompt)
    
    return {
        analysis: analysis,
        timestamp: times.format(times.now())
    }
}

generate_summary := func(state, results) {
    content := results.extract_text.content
    
    client := llm.new({model: "gpt-4"})
    
    prompt := "Create a comprehensive summary of this document.\n" +
              "Include:\n" +
              "- Executive summary (2-3 sentences)\n" +
              "- Key points (bullet list)\n" +
              "- Important details\n" +
              "- Conclusions\n\n" +
              "Document:\n" + content
    
    summary := client.complete(prompt)
    
    return {
        summary: summary
    }
}

create_insights := func(state, results) {
    analysis := results.analyze_content.analysis
    summary := results.generate_summary.summary
    
    client := llm.new({model: "gpt-4"})
    
    prompt := "Based on the analysis and summary, provide:\n" +
              "1. Key insights and implications\n" +
              "2. Actionable recommendations\n" +
              "3. Areas requiring further investigation\n" +
              "4. Potential risks or concerns\n\n" +
              "Analysis:\n" + analysis + "\n\n" +
              "Summary:\n" + summary
    
    insights := client.complete(prompt)
    
    return {
        insights: insights,
        generated_at: times.format(times.now())
    }
}

export {
    name: name,
    description: description,
    steps: steps,
    execute: execute
}
{{end}}`
}

func (g *Generator) getGenerateReportWorkflow() string {
	return `{{if eq .Engine "lua"}}-- Generate Report Workflow

local workflow = {}

workflow.name = "Generate Report"
workflow.description = "Create comprehensive reports from data"

-- Implementation would follow similar pattern to process_document
workflow.execute = function(state)
    return {
        report = "Report generation workflow placeholder",
        status = "success"
    }
end

return workflow
{{else if eq .Engine "javascript"}}// Generate Report Workflow

module.exports = {
    name: 'Generate Report',
    description: 'Create comprehensive reports from data',
    
    // Implementation would follow similar pattern to process_document
    async execute(state) {
        return {
            report: 'Report generation workflow placeholder',
            status: 'success'
        };
    }
};
{{else if eq .Engine "tengo"}}// Generate Report Workflow

name := "Generate Report"
description := "Create comprehensive reports from data"

// Implementation would follow similar pattern to process_document
execute := func(state) {
    return {
        report: "Report generation workflow placeholder",
        status: "success"
    }
}

export {
    name: name,
    description: description,
    execute: execute
}
{{end}}`
}

func (g *Generator) getAnalyzeDataWorkflow() string {
	return `{{if eq .Engine "lua"}}-- Analyze Data Workflow

local workflow = {}

workflow.name = "Analyze Data"
workflow.description = "Perform data analysis"

-- Implementation would follow similar pattern to process_document
workflow.execute = function(state)
    return {
        analysis = "Data analysis workflow placeholder",
        status = "success"
    }
end

return workflow
{{else if eq .Engine "javascript"}}// Analyze Data Workflow

module.exports = {
    name: 'Analyze Data',
    description: 'Perform data analysis',
    
    // Implementation would follow similar pattern to process_document
    async execute(state) {
        return {
            analysis: 'Data analysis workflow placeholder',
            status: 'success'
        };
    }
};
{{else if eq .Engine "tengo"}}// Analyze Data Workflow

name := "Analyze Data"
description := "Perform data analysis"

// Implementation would follow similar pattern to process_document
execute := func(state) {
    return {
        analysis: "Data analysis workflow placeholder",
        status: "success"
    }
}

export {
    name: name,
    description: description,
    execute: execute
}
{{end}}`
}

// Interactive script content
func (g *Generator) getInteractiveScriptContent() string {
	return `{{if eq .Engine "lua"}}-- {{.Name}} - Interactive spell
-- {{.Description}}

local llm = require("llm")
local state = require("state")
local hooks = require("hooks")
local json = require("json")

-- Get parameters
local mode = params.mode or "assistant"
local personality = params.personality or "helpful"

-- Initialize conversation state
local conversation = state.new("conversation")
conversation:set("history", {})
conversation:set("mode", mode)

-- Create LLM client with personality
local system_prompts = {
    helpful = "You are a helpful and friendly assistant.",
    professional = "You are a professional and formal assistant.",
    creative = "You are a creative and imaginative assistant.",
    technical = "You are a technical expert assistant."
}

local client = llm.new({
    model = "gpt-4",
    system = system_prompts[personality] or system_prompts.helpful
})

-- Command handlers
local commands = {
    ["/help"] = function()
        print([[
Available commands:
  /help     - Show this help message
  /clear    - Clear conversation history
  /save     - Save conversation to file
  /load     - Load conversation from file
  /mode     - Change interaction mode
  /exit     - Exit the program
        ]])
    end,
    
    ["/clear"] = function()
        conversation:set("history", {})
        print("Conversation history cleared.")
    end,
    
    ["/save"] = function()
        local history = conversation:get("history")
        local filename = "conversation_" .. os.date("%Y%m%d_%H%M%S") .. ".json"
        local file = io.open(filename, "w")
        file:write(json.encode(history, {indent = true}))
        file:close()
        print("Conversation saved to: " .. filename)
    end,
    
    ["/exit"] = function()
        print("Goodbye!")
        os.exit(0)
    end
}

-- Mode handlers
local mode_handlers = {
    assistant = function(input, history)
        table.insert(history, {role = "user", content = input})
        local response = client:chat(history)
        table.insert(history, {role = "assistant", content = response})
        return response
    end,
    
    chat = function(input, history)
        -- Similar to assistant but more conversational
        return mode_handlers.assistant(input, history)
    end,
    
    quiz = function(input, history)
        -- Quiz mode would implement Q&A logic
        print("Quiz mode not yet implemented")
        return "Quiz functionality coming soon!"
    end
}

-- Main interaction loop
print("Welcome to " .. params.name .. "!")
print("Mode: " .. mode .. " | Personality: " .. personality)
print("Type /help for commands or start chatting!\n")

while true do
    io.write("> ")
    local input = io.read()
    
    if not input then
        break
    end
    
    input = input:gsub("^%s*(.-)%s*$", "%1") -- trim
    
    if input == "" then
        goto continue
    end
    
    -- Check for commands
    if input:sub(1, 1) == "/" then
        local cmd = commands[input:match("^(/[%w]+)")]
        if cmd then
            cmd()
        else
            print("Unknown command. Type /help for available commands.")
        end
    else
        -- Process input based on mode
        local handler = mode_handlers[mode]
        if handler then
            local history = conversation:get("history")
            local response = handler(input, history)
            conversation:set("history", history)
            print("\n" .. response .. "\n")
        else
            print("Unknown mode: " .. mode)
        end
    end
    
    ::continue::
end

return "Interactive session ended"
{{else if eq .Engine "javascript"}}// {{.Name}} - Interactive spell
// {{.Description}}

const llm = require('llm');
const state = require('state');
const hooks = require('hooks');
const readline = require('readline');
const fs = require('fs').promises;

// Get parameters
const mode = params.mode || 'assistant';
const personality = params.personality || 'helpful';

// Initialize conversation state
const conversation = state.new('conversation');
conversation.set('history', []);
conversation.set('mode', mode);

// Create LLM client with personality
const systemPrompts = {
    helpful: 'You are a helpful and friendly assistant.',
    professional: 'You are a professional and formal assistant.',
    creative: 'You are a creative and imaginative assistant.',
    technical: 'You are a technical expert assistant.'
};

const client = llm.new({
    model: 'gpt-4',
    system: systemPrompts[personality] || systemPrompts.helpful
});

// Command handlers
const commands = {
    '/help': () => {
        console.log(` + "`" + `
Available commands:
  /help     - Show this help message
  /clear    - Clear conversation history
  /save     - Save conversation to file
  /load     - Load conversation from file
  /mode     - Change interaction mode
  /exit     - Exit the program
        ` + "`" + `);
    },
    
    '/clear': () => {
        conversation.set('history', []);
        console.log('Conversation history cleared.');
    },
    
    '/save': async () => {
        const history = conversation.get('history');
        const filename = ` + "`conversation_${new Date().toISOString().replace(/[:.]/g, '-')}.json`" + `;
        await fs.writeFile(filename, JSON.stringify(history, null, 2));
        console.log(` + "`Conversation saved to: ${filename}`" + `);
    },
    
    '/exit': () => {
        console.log('Goodbye!');
        process.exit(0);
    }
};

// Mode handlers
const modeHandlers = {
    assistant: async (input, history) => {
        history.push({role: 'user', content: input});
        const response = await client.chat(history);
        history.push({role: 'assistant', content: response});
        return response;
    },
    
    chat: async (input, history) => {
        // Similar to assistant but more conversational
        return modeHandlers.assistant(input, history);
    },
    
    quiz: async (input, history) => {
        // Quiz mode would implement Q&A logic
        console.log('Quiz mode not yet implemented');
        return 'Quiz functionality coming soon!';
    }
};

// Create readline interface
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: '> '
});

// Main interaction
console.log(` + "`Welcome to ${params.name}!`" + `);
console.log(` + "`Mode: ${mode} | Personality: ${personality}`" + `);
console.log('Type /help for commands or start chatting!\n');

rl.prompt();

rl.on('line', async (input) => {
    input = input.trim();
    
    if (!input) {
        rl.prompt();
        return;
    }
    
    // Check for commands
    if (input.startsWith('/')) {
        const cmd = input.split(' ')[0];
        const handler = commands[cmd];
        if (handler) {
            await handler();
        } else {
            console.log('Unknown command. Type /help for available commands.');
        }
    } else {
        // Process input based on mode
        const handler = modeHandlers[mode];
        if (handler) {
            const history = conversation.get('history');
            const response = await handler(input, history);
            conversation.set('history', history);
            console.log(` + "`\n${response}\n`" + `);
        } else {
            console.log(` + "`Unknown mode: ${mode}`" + `);
        }
    }
    
    rl.prompt();
});

rl.on('close', () => {
    console.log('\\nGoodbye!');
    process.exit(0);
});

// Keep process alive
process.stdin.resume();
{{else if eq .Engine "tengo"}}// {{.Name}} - Interactive spell
// {{.Description}}

llm := import("llm")
state := import("state")
fmt := import("fmt")
os := import("os")
json := import("json")
text := import("text")

// Get parameters
mode := params.mode || "assistant"
personality := params.personality || "helpful"

// Initialize conversation state
conversation := state.new("conversation")
conversation.set("history", [])
conversation.set("mode", mode)

// Create LLM client with personality
system_prompts := {
    helpful: "You are a helpful and friendly assistant.",
    professional: "You are a professional and formal assistant.",
    creative: "You are a creative and imaginative assistant.",
    technical: "You are a technical expert assistant."
}

client := llm.new({
    model: "gpt-4",
    system: system_prompts[personality] || system_prompts.helpful
})

// Simple interactive loop (Tengo doesn't have readline)
fmt.println("Welcome to", params.name + "!")
fmt.println("Mode:", mode, "| Personality:", personality)
fmt.println("Type 'exit' to quit\n")

handle_input := func(input, history) {
    if input == "exit" {
        os.exit(0)
    }
    
    if input == "help" {
        fmt.println("Commands: help, clear, exit")
        return undefined
    }
    
    if input == "clear" {
        conversation.set("history", [])
        fmt.println("Conversation history cleared.")
        return undefined
    }
    
    // Process based on mode
    history = append(history, {role: "user", content: input})
    response := client.chat(history)
    history = append(history, {role: "assistant", content: response})
    conversation.set("history", history)
    
    return response
}

// Note: Tengo doesn't have interactive input capabilities
// This is a simplified version
fmt.println("Interactive mode requires external input handling")
fmt.println("Use this spell with a wrapper script for full interactivity")

// Process single input if provided
if params.input {
    history := conversation.get("history")
    response := handle_input(params.input, history)
    if response {
        fmt.println("\n" + response + "\n")
    }
}

return "Interactive session ended"
{{end}}`
}
