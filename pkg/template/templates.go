// ABOUTME: Built-in spell templates for various use cases.
// ABOUTME: Contains basic, advanced, agent-based, workflow, and interactive templates.

// Package template provides spell template generation functionality.
// It includes templates for various spell types including basic, advanced,
// agent-based, workflow, and interactive spells.
package template

// createBasicTemplate creates a basic spell template.
// It generates a simple spell structure for basic LLM interactions
// with minimal configuration and a straightforward script.
func (g *Generator) createBasicTemplate() *SpellTemplate {
	return &SpellTemplate{
		Name:        "Basic Spell",
		Description: "A simple spell for basic LLM interactions",
		Type:        TemplateTypeBasic,
		Files: map[string]FileTemplate{
			"spell.yaml": {
				Path:     "spell.yaml",
				Template: true,
				Content: `name: {{.Name}}
description: {{.Description}}
author: {{.Author}}
license: {{.License}}
version: 1.0.0
engine: {{.Engine}}
entry_point: main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}}
timeout: 60s

# Security configuration
# Note: These can be overridden by --security-level and --feature-set CLI flags
security_profile: untrusted  # Maps to security level
# feature_set is controlled by CLI flag, not spell.yaml

# Dependencies (optional)
dependencies: []

# Parameters that can be passed to the spell
parameters:
  - name: prompt
    type: string
    description: The prompt to send to the LLM
    required: true
  - name: model
    type: string
    description: The model to use
    default: gpt-3.5-turbo

# Tags for categorization (optional)
tags:
  - llm
  - basic

# Additional metadata (optional)
metadata:
  category: example
  difficulty: beginner
`,
			},
			"main.script": {
				Path:     "main.script",
				Template: true,
				Content:  g.getBasicScriptContent(),
			},
			"README.md": {
				Path:     "README.md",
				Template: true,
				Content: `# {{.Name}}

{{.Description}}

## Usage

` + "```bash" + `
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --prompt "Your prompt here"
` + "```" + `

## Parameters

- ` + "`prompt`" + `: The prompt to send to the LLM (required)
- ` + "`model`" + `: The model to use (default: gpt-3.5-turbo)

## Author

{{.Author}}

## License

{{.License}}
`,
			},
		},
	}
}

// createAdvancedTemplate creates an advanced spell template.
// It includes state management, error handling, multiple operation modes,
// and library modules for more sophisticated LLM applications.
func (g *Generator) createAdvancedTemplate() *SpellTemplate {
	return &SpellTemplate{
		Name:        "Advanced Spell",
		Description: "An advanced spell with state management and error handling",
		Type:        TemplateTypeAdvanced,
		Files: map[string]FileTemplate{
			"spell.yaml": {
				Path:     "spell.yaml",
				Template: true,
				Content: `name: {{.Name}}
description: {{.Description}}
author: {{.Author}}
license: {{.License}}
version: 1.0.0
engine: {{.Engine}}
entry_point: main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}}
timeout: 300s

# Security configuration  
# Note: For advanced features, consider using --security-level trusted --feature-set full
security_profile: trusted

# Dependencies
dependencies:
  - lib/utils
  - lib/prompts

# Parameters
parameters:
  - name: mode
    type: string
    description: Operation mode
    enum: [chat, analyze, summarize]
    default: chat
  - name: input_file
    type: string
    description: Input file path (optional)
    required: false
  - name: output_file
    type: string
    description: Output file path (optional)
    required: false
  - name: model
    type: string
    description: The model to use
    default: gpt-4
  - name: temperature
    type: number
    description: Temperature for generation
    default: 0.7
    validation: "value >= 0 and value <= 2"

# Tags
tags:
  - llm
  - advanced
  - stateful

# Metadata
metadata:
  category: example
  difficulty: intermediate
  features:
    - state-management
    - error-handling
    - file-operations
`,
			},
			"main.script": {
				Path:     "main.script",
				Template: true,
				Content:  g.getAdvancedScriptContent(),
			},
			"lib/utils.script": {
				Path:     "lib/utils.script",
				Template: true,
				Content:  g.getUtilsScriptContent(),
			},
			"lib/prompts.script": {
				Path:     "lib/prompts.script",
				Template: true,
				Content:  g.getPromptsScriptContent(),
			},
			"config/default.yaml": {
				Path:     "config/default.yaml",
				Template: false,
				Content: `# Default configuration
models:
  default: gpt-4
  fallback: gpt-3.5-turbo

prompts:
  system: "You are a helpful assistant."
  
retry:
  max_attempts: 3
  delay: 1000
`,
			},
			"README.md": {
				Path:     "README.md",
				Template: true,
				Content: `# {{.Name}}

{{.Description}}

## Features

- Multiple operation modes (chat, analyze, summarize)
- State management for conversation history
- File input/output support
- Error handling and retry logic
- Configurable prompts and models

## Usage

` + "```bash" + `
# Interactive chat mode
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --mode chat

# Analyze a file
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --mode analyze --input_file document.txt

# Summarize with output
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --mode summarize --input_file article.txt --output_file summary.txt
` + "```" + `

## Configuration

Edit ` + "`config/default.yaml`" + ` to customize default settings.

## Author

{{.Author}}

## License

{{.License}}
`,
			},
		},
	}
}

// createAgentTemplate creates an agent-based spell template.
// It provides a structure for autonomous agents that can use tools,
// make decisions, and complete complex tasks iteratively.
func (g *Generator) createAgentTemplate() *SpellTemplate {
	return &SpellTemplate{
		Name:        "Agent Spell",
		Description: "An agent-based spell with tool usage",
		Type:        TemplateTypeAgent,
		Files: map[string]FileTemplate{
			"spell.yaml": {
				Path:     "spell.yaml",
				Template: true,
				Content: `name: {{.Name}}
description: {{.Description}}
author: {{.Author}}
license: {{.License}}
version: 1.0.0
engine: {{.Engine}}
entry_point: main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}}
timeout: 600s  # 10 minutes for agent tasks

# Security configuration
# Note: Agents need --security-level trusted --feature-set agent for full functionality
security_profile: trusted

# Dependencies
dependencies:
  - tools/calculator
  - tools/web_search
  - tools/file_reader

# Parameters
parameters:
  - name: task
    type: string
    description: The task for the agent to complete
    required: true
  - name: tools
    type: array
    description: List of tools to enable
    default: ["calculator", "web_search", "file_reader"]
  - name: max_iterations
    type: number
    description: Maximum iterations for the agent
    default: 10
    validation: "value > 0 and value <= 100"

# Tags
tags:
  - llm
  - agent
  - autonomous

# Metadata
metadata:
  category: example
  difficulty: advanced
  features:
    - agent-workflow
    - tool-usage
    - iterative-reasoning
`,
			},
			"main.script": {
				Path:     "main.script",
				Template: true,
				Content:  g.getAgentScriptContent(),
			},
			"tools/calculator.script": {
				Path:     "tools/calculator.script",
				Template: true,
				Content:  g.getCalculatorToolContent(),
			},
			"tools/web_search.script": {
				Path:     "tools/web_search.script",
				Template: true,
				Content:  g.getWebSearchToolContent(),
			},
			"tools/file_reader.script": {
				Path:     "tools/file_reader.script",
				Template: true,
				Content:  g.getFileReaderToolContent(),
			},
			"README.md": {
				Path:     "README.md",
				Template: true,
				Content: `# {{.Name}}

{{.Description}}

## Features

- Agent-based task execution
- Multiple tool support (calculator, web search, file reader)
- Customizable tool selection
- Iteration limits for safety

## Usage

` + "```bash" + `
# Run with default tools
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --task "Research the latest AI developments and summarize them"

# Run with specific tools
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --task "Calculate the compound interest" --tools '["calculator"]'
` + "```" + `

## Tools

### Calculator
Performs mathematical calculations.

### Web Search
Searches the web for information (requires API key).

### File Reader
Reads and processes local files.

## Adding Custom Tools

1. Create a new tool file in the ` + "`tools/`" + ` directory
2. Register the tool in ` + "`main.script`" + `
3. Add the tool to the available tools list

## Author

{{.Author}}

## License

{{.License}}
`,
			},
		},
	}
}

// createWorkflowTemplate creates a workflow spell template.
// It supports multi-step processes with checkpoints, state persistence,
// and different workflow types for document processing and data analysis.
func (g *Generator) createWorkflowTemplate() *SpellTemplate {
	return &SpellTemplate{
		Name:        "Workflow Spell",
		Description: "A workflow-based spell for complex multi-step processes",
		Type:        TemplateTypeWorkflow,
		Files: map[string]FileTemplate{
			"spell.yaml": {
				Path:     "spell.yaml",
				Template: true,
				Content: `name: {{.Name}}
description: {{.Description}}
author: {{.Author}}
license: {{.License}}
version: 1.0.0
engine: {{.Engine}}
entry_point: main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}}
timeout: 900s  # 15 minutes for complex workflows

# Security configuration
# Note: Workflows need --security-level trusted --feature-set full for all features
security_profile: trusted

# Dependencies
dependencies:
  - workflows/process_document
  - workflows/generate_report
  - workflows/analyze_data

# Parameters
parameters:
  - name: workflow
    type: string
    description: The workflow to execute
    enum: [process_document, generate_report, analyze_data]
    default: process_document
  - name: input
    type: string
    description: Input data or file path
    required: true
  - name: output_dir
    type: string
    description: Output directory
    default: ./output

# Tags
tags:
  - llm
  - workflow
  - multi-step
  - stateful

# Metadata
metadata:
  category: example
  difficulty: advanced
  features:
    - workflow-orchestration
    - state-persistence
    - checkpoint-recovery
`,
			},
			"main.script": {
				Path:     "main.script",
				Template: true,
				Content:  g.getWorkflowScriptContent(),
			},
			"workflows/process_document.script": {
				Path:     "workflows/process_document.script",
				Template: true,
				Content:  g.getProcessDocumentWorkflow(),
			},
			"workflows/generate_report.script": {
				Path:     "workflows/generate_report.script",
				Template: true,
				Content:  g.getGenerateReportWorkflow(),
			},
			"workflows/analyze_data.script": {
				Path:     "workflows/analyze_data.script",
				Template: true,
				Content:  g.getAnalyzeDataWorkflow(),
			},
			"README.md": {
				Path:     "README.md",
				Template: true,
				Content: `# {{.Name}}

{{.Description}}

## Features

- Multiple workflow support
- Step-by-step execution with checkpoints
- State persistence between steps
- Event-driven architecture
- Error recovery and rollback

## Workflows

### Process Document
Processes documents through multiple stages:
1. Extract text
2. Analyze content
3. Generate summary
4. Create insights

### Generate Report
Creates comprehensive reports:
1. Gather data
2. Analyze trends
3. Generate visualizations
4. Compile report

### Analyze Data
Performs data analysis:
1. Load data
2. Clean and preprocess
3. Statistical analysis
4. Generate conclusions

## Usage

` + "```bash" + `
# Process a document
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --workflow process_document --input document.pdf

# Generate a report
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --workflow generate_report --input data.csv --output_dir reports/

# Analyze data
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --workflow analyze_data --input dataset.json
` + "```" + `

## Author

{{.Author}}

## License

{{.License}}
`,
			},
		},
	}
}

// createInteractiveTemplate creates an interactive spell template.
// It provides a terminal-based interface for real-time user interaction,
// supporting chat, quiz, and assistant modes with conversation history.
func (g *Generator) createInteractiveTemplate() *SpellTemplate {
	return &SpellTemplate{
		Name:        "Interactive Spell",
		Description: "An interactive spell with user input and dynamic responses",
		Type:        TemplateTypeInteractive,
		Files: map[string]FileTemplate{
			"spell.yaml": {
				Path:     "spell.yaml",
				Template: true,
				Content: `name: {{.Name}}
description: {{.Description}}
author: {{.Author}}
license: {{.License}}
version: 1.0.0
engine: {{.Engine}}
entry_point: main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}}
timeout: 3600s  # 1 hour for interactive sessions

# Security configuration
# Note: Interactive mode needs --security-level trusted --feature-set full
security_profile: trusted

# Dependencies (none for basic interactive)
dependencies: []

# Parameters
parameters:
  - name: mode
    type: string
    description: Interaction mode
    enum: [chat, quiz, assistant]
    default: assistant
  - name: personality
    type: string
    description: Assistant personality
    default: helpful
    validation: "value in ['helpful', 'professional', 'friendly', 'concise']"

# Tags
tags:
  - llm
  - interactive
  - conversational

# Metadata
metadata:
  category: example
  difficulty: intermediate
  features:
    - terminal-ui
    - conversation-history
    - command-system
`,
			},
			"main.script": {
				Path:     "main.script",
				Template: true,
				Content:  g.getInteractiveScriptContent(),
			},
			"README.md": {
				Path:     "README.md",
				Template: true,
				Content: `# {{.Name}}

{{.Description}}

## Features

- Interactive terminal interface
- Multiple interaction modes
- Conversation history
- Customizable personalities
- Command system

## Modes

### Assistant
A helpful assistant that can answer questions and perform tasks.

### Chat
Free-form conversation mode.

### Quiz
Interactive quiz mode with scoring.

## Usage

` + "```bash" + `
# Start in assistant mode
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}}

# Start in quiz mode
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --mode quiz

# Use a specific personality
llmspell run main.{{if eq .Engine "javascript"}}js{{else if eq .Engine "js"}}js{{else if eq .Engine "tengo"}}tengo{{else}}lua{{end}} --personality professional
` + "```" + `

## Commands

- ` + "`/help`" + ` - Show available commands
- ` + "`/clear`" + ` - Clear conversation history
- ` + "`/save`" + ` - Save conversation
- ` + "`/load`" + ` - Load previous conversation
- ` + "`/exit`" + ` - Exit the program

## Author

{{.Author}}

## License

{{.License}}
`,
			},
		},
	}
}
