# Web Summarizer Spell

Fetches web pages and generates summaries using LLM.

## Features

- Fetches content from any HTTP/HTTPS URL
- Offers multiple summary styles
- Saves summaries to storage
- Handles errors gracefully

## Usage

Basic usage (brief summary):
```bash
llmspell run web-summarizer --param url="https://example.com/article"
```

Detailed summary:
```bash
llmspell run web-summarizer \
  --param url="https://example.com/article" \
  --param style="detailed"
```

Bullet-point summary:
```bash
llmspell run web-summarizer \
  --param url="https://example.com/article" \
  --param style="bullet-points"
```

## Parameters

- `url` (string, required): The URL of the web page to summarize
- `style` (string, optional): Summary style - "brief", "detailed", or "bullet-points" (default: "brief")

## Summary Styles

### Brief
A concise 2-3 sentence summary capturing the main idea.

### Detailed
A comprehensive summary including:
- Main topic or purpose
- Key points or arguments
- Important facts or figures
- Conclusions or takeaways

### Bullet Points
Key information organized as easy-to-scan bullet points.

## Example Output

```
Web Summarizer
==============
URL: https://example.com/article-about-ai
Style: brief

Fetching web page...
Page fetched successfully (12543 bytes)

Generating summary...

================================================================================
SUMMARY
================================================================================
The article discusses recent advances in artificial intelligence, particularly 
focusing on large language models and their applications in various industries. 
It highlights both the potential benefits and ethical considerations that need 
to be addressed as AI becomes more prevalent in society.

Summary saved to: summary_example.com_article-about-ai.txt
```

## Notes

- The spell requires an active internet connection
- Large web pages may take longer to process
- Some websites may block automated requests
- The quality of summaries depends on the LLM provider and model used