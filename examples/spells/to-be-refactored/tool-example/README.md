# Tool System Example

This example demonstrates the tool system capabilities in go-llmspell, showing how to:

1. Register custom tools with parameter schemas
2. Execute tools with validated parameters
3. Handle tool errors gracefully
4. List and inspect available tools
5. Remove tools when no longer needed

## Tools Demonstrated

### Calculator Tool
- Performs basic arithmetic operations (add, subtract, multiply, divide)
- Validates numeric inputs
- Handles division by zero errors

### String Tools
- Provides string manipulation utilities (upper, lower, reverse, length)
- Works with text inputs

### JSON Processor
- Parses and processes JSON data
- Can extract specific fields
- Demonstrates more complex parameter handling

## Key Features

- **Parameter Validation**: Tools define JSON schemas for their parameters
- **Error Handling**: Tools can return errors for invalid inputs
- **Tool Discovery**: Scripts can list and inspect available tools
- **Dynamic Registration**: Tools can be registered and removed at runtime

## Running the Example

```bash
# From the project root:
./llmspell run examples/spells/tool-example

# Or if llmspell is in your PATH:
llmspell run examples/spells/tool-example
```

## Expected Output

The script will:
1. Register three custom tools
2. List all registered tools
3. Execute various tool operations
4. Demonstrate parameter validation
5. Show tool information retrieval
6. Remove a tool and confirm removal