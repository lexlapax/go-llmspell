// ABOUTME: Built-in spell templates for various use cases.
// ABOUTME: Contains basic, advanced, agent-based, workflow, and interactive templates.

package template

// createBasicTemplate creates a basic spell template
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

security:
  profile: sandbox
  permissions:
    - llm:chat
    - file:read

parameters:
  prompt:
    type: string
    description: The prompt to send to the LLM
    required: true
  model:
    type: string
    description: The model to use
    default: gpt-3.5-turbo
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

// createAdvancedTemplate creates an advanced spell template
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

security:
  profile: development
  permissions:
    - llm:*
    - file:*
    - state:*
    - hooks:*

parameters:
  mode:
    type: string
    description: Operation mode
    enum: [chat, analyze, summarize]
    default: chat
  input_file:
    type: string
    description: Input file path (optional)
  output_file:
    type: string
    description: Output file path (optional)
  model:
    type: string
    description: The model to use
    default: gpt-4
  temperature:
    type: number
    description: Temperature for generation
    default: 0.7
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

// createAgentTemplate creates an agent-based spell template
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

security:
  profile: development
  permissions:
    - llm:*
    - agent:*
    - tools:*
    - file:*
    - network:read

parameters:
  task:
    type: string
    description: The task for the agent to complete
    required: true
  tools:
    type: array
    description: List of tools to enable
    default: ["calculator", "web_search", "file_reader"]
  max_iterations:
    type: number
    description: Maximum iterations for the agent
    default: 10
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

// createWorkflowTemplate creates a workflow spell template
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

security:
  profile: development
  permissions:
    - llm:*
    - workflow:*
    - state:*
    - events:*
    - file:*

parameters:
  workflow:
    type: string
    description: The workflow to execute
    enum: [process_document, generate_report, analyze_data]
    default: process_document
  input:
    type: string
    description: Input data or file path
    required: true
  output_dir:
    type: string
    description: Output directory
    default: ./output
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

// createInteractiveTemplate creates an interactive spell template
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

security:
  profile: development
  permissions:
    - llm:*
    - state:*
    - hooks:*
    - ui:terminal

parameters:
  mode:
    type: string
    description: Interaction mode
    enum: [chat, quiz, assistant]
    default: assistant
  personality:
    type: string
    description: Assistant personality
    default: helpful
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
