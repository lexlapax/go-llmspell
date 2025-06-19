-- ABOUTME: Web page summarizer spell using HTTP fetch and LLM
-- ABOUTME: Demonstrates tool usage and content processing

-- Get parameters
local url = params.url
local style = params.style or "brief"

if not url then
    error("URL parameter is required")
end

-- Validate URL format
if not string.match(url, "^https?://") then
    error("Invalid URL format. Must start with http:// or https://")
end

print("Web Summarizer")
print("==============")
print("URL: " .. url)
print("Style: " .. style)
print("")

-- Fetch the web page
print("Fetching web page...")
local content, err = http.get(url)

if err then
    error("Failed to fetch URL: " .. err)
end

print("Page fetched successfully (" .. #content .. " bytes)")

-- Prepare the prompt based on style
local prompt = ""

if style == "brief" then
    prompt = string.format(
        [[
Please provide a brief summary (2-3 sentences) of the following web page content:

%s

Summary:]],
        content
    )
elseif style == "detailed" then
    prompt = string.format(
        [[
Please provide a detailed summary of the following web page content. Include:
1. Main topic or purpose
2. Key points or arguments
3. Important facts or figures
4. Conclusions or takeaways

Content:
%s

Detailed Summary:]],
        content
    )
elseif style == "bullet-points" then
    prompt = string.format(
        [[
Please summarize the following web page content as bullet points:

%s

Summary (bullet points):]],
        content
    )
else
    error("Unknown style: " .. style .. ". Use 'brief', 'detailed', or 'bullet-points'")
end

-- Send to LLM for summarization
print("\nGenerating summary...")
local summary, llm_err = llm.chat(prompt)

if llm_err then
    error("LLM error: " .. llm_err)
end

-- Display the summary
print("\n" .. string.rep("=", 80))
print("SUMMARY")
print(string.rep("=", 80))
print(summary)

-- Save summary to storage
if storage then
    -- Create filename from URL
    local filename = string.gsub(url, "https?://", "")
    filename = string.gsub(filename, "[^%w%-_.]", "_")
    filename = "summary_" .. filename .. ".txt"

    -- Save with metadata
    local output = string.format(
        [[
URL: %s
Date: [date unavailable]
Style: %s

%s
]],
        url,
        style,
        summary
    )

    local save_err = storage.write(filename, output)

    if save_err then
        print("\nWarning: Could not save summary: " .. save_err)
    else
        print("\nSummary saved to: " .. filename)
    end
end
