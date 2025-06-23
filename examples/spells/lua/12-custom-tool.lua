-- ABOUTME: Custom tool creation example showing how to define and use custom tools in Lua
-- ABOUTME: Demonstrates creating tools with schemas, validation, and integration with agents

-- Custom Tool Creation Example
-- This spell demonstrates how to create custom tools including:
-- 1. Tool definition with schemas
-- 2. Input validation and error handling
-- 3. Tool registration and discovery
-- 4. Integration with agents
-- 5. Complex tool compositions

-- Required modules
local tools = require("tools")
local agent = require("agent")
local llm = require("llm")
local data = require("data")
local errors = require("errors")
local log = require("log")
local core = require("core")

-- Example 1: Basic Custom Tool
print("=== Example 1: Basic Custom Tool ===")

-- Define a calculator tool
local calculator_tool = {
    name = "calculator",
    description = "Performs basic mathematical operations",
    schema = {
        type = "object",
        properties = {
            operation = {
                type = "string",
                enum = {"add", "subtract", "multiply", "divide"},
                description = "The mathematical operation to perform"
            },
            a = {
                type = "number",
                description = "First operand"
            },
            b = {
                type = "number", 
                description = "Second operand"
            }
        },
        required = {"operation", "a", "b"}
    },
    
    -- Implementation
    execute = function(params)
        local a = params.a
        local b = params.b
        
        if params.operation == "add" then
            return {result = a + b, operation = "addition"}
        elseif params.operation == "subtract" then
            return {result = a - b, operation = "subtraction"}
        elseif params.operation == "multiply" then
            return {result = a * b, operation = "multiplication"}
        elseif params.operation == "divide" then
            if b == 0 then
                error("Division by zero")
            end
            return {result = a / b, operation = "division"}
        else
            error("Unknown operation: " .. params.operation)
        end
    end
}

-- Register the tool
tools.register(calculator_tool)
print("Registered calculator tool")

-- Test the tool directly
local calc_result = tools.execute("calculator", {
    operation = "multiply",
    a = 7,
    b = 8
})
print("Calculator result: 7 * 8 = " .. calc_result.result)
print()

-- Example 2: Advanced Tool with External Data
print("=== Example 2: Advanced Tool with External Data ===")

-- Create a weather tool (simulated)
local weather_data = {
    ["New York"] = {temp = 72, condition = "Sunny", humidity = 65},
    ["London"] = {temp = 59, condition = "Cloudy", humidity = 80},
    ["Tokyo"] = {temp = 77, condition = "Partly Cloudy", humidity = 70},
    ["Sydney"] = {temp = 68, condition = "Clear", humidity = 55}
}

local weather_tool = {
    name = "get_weather",
    description = "Get current weather for a city",
    schema = {
        type = "object",
        properties = {
            city = {
                type = "string",
                description = "City name"
            },
            units = {
                type = "string",
                enum = {"fahrenheit", "celsius"},
                description = "Temperature units",
                default = "fahrenheit"
            }
        },
        required = {"city"}
    },
    
    execute = function(params)
        local city = params.city
        local units = params.units or "fahrenheit"
        
        -- Simulate API call delay
        core.sleep(0.1)
        
        local data = weather_data[city]
        if not data then
            -- If city not in cache, generate random data
            data = {
                temp = math.random(50, 90),
                condition = ({"Sunny", "Cloudy", "Rainy", "Partly Cloudy"})[math.random(1, 4)],
                humidity = math.random(40, 90)
            }
            weather_data[city] = data
        end
        
        -- Convert temperature if needed
        local temp = data.temp
        if units == "celsius" then
            temp = (temp - 32) * 5/9
        end
        
        return {
            city = city,
            temperature = temp,
            units = units,
            condition = data.condition,
            humidity = data.humidity,
            timestamp = os.time()
        }
    end
}

tools.register(weather_tool)

-- Test weather tool
local weather = tools.execute("get_weather", {city = "New York", units = "celsius"})
print(string.format("Weather in %s: %.1f°C, %s, %d%% humidity",
    weather.city, weather.temperature, weather.condition, weather.humidity))
print()

-- Example 3: Tool with Validation and Error Handling
print("=== Example 3: Tool with Validation and Error Handling ===")

-- Create a text analysis tool
local text_analyzer_tool = {
    name = "analyze_text",
    description = "Analyzes text for various metrics and patterns",
    schema = {
        type = "object",
        properties = {
            text = {
                type = "string",
                description = "Text to analyze",
                minLength = 1,
                maxLength = 10000
            },
            analyses = {
                type = "array",
                items = {
                    type = "string",
                    enum = {"word_count", "char_count", "sentiment", "language", "readability"}
                },
                description = "Types of analysis to perform",
                default = {"word_count", "char_count"}
            }
        },
        required = {"text"}
    },
    
    -- Custom validation
    validate = function(params)
        if not params.text or #params.text == 0 then
            return false, "Text cannot be empty"
        end
        
        if #params.text > 10000 then
            return false, "Text too long (max 10000 characters)"
        end
        
        -- Validate analyses
        if params.analyses then
            for _, analysis in ipairs(params.analyses) do
                local valid = false
                for _, allowed in ipairs({"word_count", "char_count", "sentiment", "language", "readability"}) do
                    if analysis == allowed then
                        valid = true
                        break
                    end
                end
                if not valid then
                    return false, "Invalid analysis type: " .. analysis
                end
            end
        end
        
        return true
    end,
    
    execute = function(params)
        local text = params.text
        local analyses = params.analyses or {"word_count", "char_count"}
        local results = {}
        
        for _, analysis in ipairs(analyses) do
            if analysis == "word_count" then
                local count = 0
                for word in string.gmatch(text, "%S+") do
                    count = count + 1
                end
                results.word_count = count
                
            elseif analysis == "char_count" then
                results.char_count = #text
                
            elseif analysis == "sentiment" then
                -- Simple sentiment analysis (mock)
                local positive_words = {"good", "great", "excellent", "happy", "wonderful"}
                local negative_words = {"bad", "terrible", "awful", "sad", "horrible"}
                local pos_count, neg_count = 0, 0
                
                local lower_text = string.lower(text)
                for _, word in ipairs(positive_words) do
                    if string.find(lower_text, word) then
                        pos_count = pos_count + 1
                    end
                end
                for _, word in ipairs(negative_words) do
                    if string.find(lower_text, word) then
                        neg_count = neg_count + 1
                    end
                end
                
                if pos_count > neg_count then
                    results.sentiment = "positive"
                elseif neg_count > pos_count then
                    results.sentiment = "negative"
                else
                    results.sentiment = "neutral"
                end
                
            elseif analysis == "language" then
                -- Detect language (simplified)
                if string.find(text, "[あ-ん]") then
                    results.language = "Japanese"
                elseif string.find(text, "[一-龯]") then
                    results.language = "Chinese"
                else
                    results.language = "English"  -- Default
                end
                
            elseif analysis == "readability" then
                -- Simple readability score
                local word_count = results.word_count or 0
                for word in string.gmatch(text, "%S+") do
                    word_count = word_count + 1
                end
                
                local sentence_count = 0
                for _ in string.gmatch(text, "[.!?]+") do
                    sentence_count = sentence_count + 1
                end
                sentence_count = math.max(1, sentence_count)
                
                local avg_words_per_sentence = word_count / sentence_count
                
                if avg_words_per_sentence < 10 then
                    results.readability = "very easy"
                elseif avg_words_per_sentence < 15 then
                    results.readability = "easy"
                elseif avg_words_per_sentence < 20 then
                    results.readability = "moderate"
                else
                    results.readability = "difficult"
                end
            end
        end
        
        results.text_sample = string.sub(text, 1, 50) .. "..."
        return results
    end
}

tools.register(text_analyzer_tool)

-- Test with validation
local analysis = tools.execute("analyze_text", {
    text = "This is a great example of custom tool creation. It's wonderful!",
    analyses = {"word_count", "sentiment", "readability"}
})

print("Text analysis results:")
for k, v in pairs(analysis) do
    print("  " .. k .. ": " .. tostring(v))
end
print()

-- Example 4: Composite Tool (Using Other Tools)
print("=== Example 4: Composite Tool (Using Other Tools) ===")

-- Create a research assistant tool that uses other tools
local research_tool = {
    name = "research_assistant",
    description = "Researches a topic using multiple sources and tools",
    schema = {
        type = "object",
        properties = {
            topic = {
                type = "string",
                description = "Research topic"
            },
            depth = {
                type = "string",
                enum = {"quick", "standard", "deep"},
                description = "Research depth",
                default = "standard"
            }
        },
        required = {"topic"}
    },
    
    execute = function(params)
        local topic = params.topic
        local depth = params.depth or "standard"
        
        log.info("Starting research on: " .. topic)
        
        -- Step 1: Generate research questions
        local questions_response = llm.complete({
            model = "gpt-3.5-turbo",
            messages = {
                {role = "system", content = "You are a research assistant."},
                {role = "user", content = string.format(
                    "Generate %d key questions about: %s",
                    depth == "quick" and 2 or (depth == "deep" and 5 or 3),
                    topic
                )}
            },
            max_tokens = 200
        })
        
        -- Step 2: Analyze the topic text
        local topic_analysis = tools.execute("analyze_text", {
            text = topic,
            analyses = {"word_count", "language"}
        })
        
        -- Step 3: Simulate gathering data (would use web search in real implementation)
        local research_data = {
            topic = topic,
            depth = depth,
            questions = questions_response.content,
            analysis = topic_analysis,
            sources = {
                {type = "article", title = "Introduction to " .. topic, relevance = 0.9},
                {type = "paper", title = "Advanced " .. topic .. " Techniques", relevance = 0.8},
                {type = "tutorial", title = topic .. " for Beginners", relevance = 0.7}
            },
            timestamp = os.time()
        }
        
        -- Step 4: Generate summary
        local summary_response = llm.complete({
            model = "gpt-3.5-turbo",
            messages = {
                {role = "user", content = string.format(
                    "Summarize research on '%s' with these questions: %s",
                    topic, questions_response.content
                )}
            },
            max_tokens = 150
        })
        
        research_data.summary = summary_response.content
        
        return research_data
    end
}

tools.register(research_tool)

-- Use the composite tool
local research = tools.execute("research_assistant", {
    topic = "Custom Tools in LLM Applications",
    depth = "standard"
})

print("Research Results:")
print("Topic: " .. research.topic)
print("Summary: " .. string.sub(research.summary, 1, 100) .. "...")
print("Sources found: " .. #research.sources)
print()

-- Example 5: Tool Integration with Agents
print("=== Example 5: Tool Integration with Agents ===")

-- Create an agent that can use our custom tools
local tool_agent = agent.create({
    name = "Tool Master",
    model = "gpt-4",
    instructions = [[
        You are an assistant that helps users by using custom tools.
        Available tools:
        - calculator: for math operations
        - get_weather: for weather information
        - analyze_text: for text analysis
        - research_assistant: for researching topics
        
        Always use tools when appropriate.
    ]],
    tools = {"calculator", "get_weather", "analyze_text", "research_assistant"}
})

-- Test agent with tools
print("Agent using calculator tool:")
local calc_response = tool_agent:generate({
    prompt = "What is 158 multiplied by 37?",
    max_tokens = 100
})
print("Agent: " .. calc_response.content)

-- Check if tool was used
if calc_response.tool_calls and #calc_response.tool_calls > 0 then
    print("Tool used: " .. calc_response.tool_calls[1].name)
end
print()

-- Example 6: Dynamic Tool Creation
print("=== Example 6: Dynamic Tool Creation ===")

-- Function to create custom converters dynamically
local function create_unit_converter(from_unit, to_unit, conversion_factor)
    local tool_name = string.format("convert_%s_to_%s", from_unit, to_unit)
    
    return {
        name = tool_name,
        description = string.format("Convert %s to %s", from_unit, to_unit),
        schema = {
            type = "object",
            properties = {
                value = {
                    type = "number",
                    description = string.format("Value in %s", from_unit)
                }
            },
            required = {"value"}
        },
        execute = function(params)
            return {
                original = params.value,
                converted = params.value * conversion_factor,
                from_unit = from_unit,
                to_unit = to_unit
            }
        end
    }
end

-- Create several unit converters
local converters = {
    create_unit_converter("miles", "kilometers", 1.60934),
    create_unit_converter("pounds", "kilograms", 0.453592),
    create_unit_converter("fahrenheit", "celsius", function(f) return (f - 32) * 5/9 end)
}

-- Register all converters
for _, converter in ipairs(converters) do
    tools.register(converter)
    print("Registered: " .. converter.name)
end

-- List all available tools
local all_tools = tools.list()
print("\nAll registered tools (" .. #all_tools .. " total):")
for _, tool in ipairs(all_tools) do
    print("  - " .. tool.name .. ": " .. tool.description)
end

-- Return summary
return {
    success = true,
    tools_created = #all_tools,
    calculator_tested = calc_result ~= nil,
    weather_tested = weather ~= nil,
    analysis_tested = analysis ~= nil,
    research_tested = research ~= nil,
    agent_integrated = tool_agent ~= nil
}