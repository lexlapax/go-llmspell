# ScriptValue Migration Guide for Bridge Authors

## Overview

This guide helps bridge authors migrate from the old `interface{}`-based system to the new ScriptValue type system. The ScriptValue system provides compile-time type safety and eliminates runtime panics from type assertions.

## Why ScriptValue?

The original design used `interface{}` throughout, which led to:
- Runtime type errors and panics
- Inconsistent type handling across engines  
- Difficult debugging with unclear type mismatches
- No compile-time safety for bridge implementations

ScriptValue solves these issues by providing a universal type system that all script engines understand.

## Key Changes

### 1. Bridge Interface Methods

**Old (interface{}-based):**
```go
func (b *Bridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
    // Dangerous runtime assertions
    prompt := args[0].(string)  // PANIC if not string!
    options := args[1].(map[string]interface{})
    // ...
}

func (b *Bridge) ValidateMethod(name string, args []interface{}) error {
    // Manual type checking
}
```

**New (ScriptValue-based):**
```go
func (b *Bridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
    // Safe type checking
    if len(args) < 1 || args[0].Type() != engine.TypeString {
        return nil, fmt.Errorf("expected string prompt, got %v", args[0].Type())
    }
    prompt := args[0].(engine.StringValue).Value()
    // ...
    return engine.NewStringValue(result), nil
}

func (b *Bridge) ValidateMethod(name string, args []engine.ScriptValue) error {
    // Type-safe validation
}
```

### 2. Type Checking Patterns

**Checking argument types:**
```go
// Single type check
if args[0].Type() != engine.TypeString {
    return nil, fmt.Errorf("argument 1 must be string, got %s", args[0].Type())
}

// Multiple type options
switch args[0].Type() {
case engine.TypeString:
    // Handle string
case engine.TypeNumber:
    // Handle number
default:
    return nil, fmt.Errorf("unexpected type: %s", args[0].Type())
}
```

**Safe value extraction:**
```go
// String value
strVal := args[0].(engine.StringValue).Value()

// Number value
numVal := args[0].(engine.NumberValue).Value()

// Boolean value
boolVal := args[0].(engine.BoolValue).Value()

// Array elements
arrVal := args[0].(engine.ArrayValue)
for _, elem := range arrVal.Elements() {
    // Process each ScriptValue element
}

// Object fields
objVal := args[0].(engine.ObjectValue)
fields := objVal.Fields()
name := fields["name"].(engine.StringValue).Value()
```

### 3. Return Value Construction

**Creating return values:**
```go
// Basic types
return engine.NewNilValue(), nil
return engine.NewBoolValue(true), nil
return engine.NewNumberValue(42.5), nil
return engine.NewStringValue("result"), nil

// Arrays
elements := []engine.ScriptValue{
    engine.NewNumberValue(1),
    engine.NewStringValue("two"),
    engine.NewBoolValue(true),
}
return engine.NewArrayValue(elements), nil

// Objects
fields := map[string]engine.ScriptValue{
    "status": engine.NewStringValue("success"),
    "count":  engine.NewNumberValue(10),
    "data":   engine.NewArrayValue(dataElements),
}
return engine.NewObjectValue(fields), nil

// Errors (return as regular Go error, not ScriptValue)
return nil, fmt.Errorf("operation failed: %v", err)
```

### 4. Helper Functions

**Converting from go-llms types:**
```go
// Use the centralized converter
scriptValue := engine.ConvertToScriptValue(goValue)

// For maps
scriptMap := engine.ConvertMapToScriptValue(goMap)

// For slices
scriptSlice := engine.ConvertSliceToScriptValue(goSlice)
```

**Converting to go-llms types:**
```go
// Basic conversion
goValue := scriptValue.ToGo()

// For passing to go-llms functions
goMap := engine.ConvertScriptValueMap(scriptValueMap)
goSlice := engine.ConvertScriptValueSlice(scriptValueSlice)
```

## Migration Steps

### Step 1: Update Method Signatures

Change all bridge methods from `[]interface{}` to `[]engine.ScriptValue`:

```go
// Update these methods
func (b *Bridge) ValidateMethod(name string, args []engine.ScriptValue) error
func (b *Bridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error)
```

### Step 2: Replace Type Assertions

Find all `args[i].(type)` assertions and replace with safe ScriptValue checks:

```go
// Old
prompt := args[0].(string)

// New
if args[0].Type() != engine.TypeString {
    return nil, fmt.Errorf("expected string")
}
prompt := args[0].(engine.StringValue).Value()
```

### Step 3: Update Return Values

Replace all `return value, nil` with appropriate ScriptValue constructors:

```go
// Old
return "result", nil
return map[string]interface{}{"key": "value"}, nil

// New
return engine.NewStringValue("result"), nil
return engine.NewObjectValue(map[string]engine.ScriptValue{
    "key": engine.NewStringValue("value"),
}), nil
```

### Step 4: Handle Complex Types

For complex go-llms types, use the conversion helpers:

```go
// Convert go-llms response to ScriptValue
llmResponse := // ... from go-llms
return engine.ConvertToScriptValue(llmResponse), nil

// Convert ScriptValue to go-llms request
request := engine.ConvertFromScriptValue(args[0])
// Pass to go-llms function
```

## Common Patterns

### Optional Parameters

```go
func (b *Bridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
    // Required first parameter
    if len(args) < 1 || args[0].Type() != engine.TypeString {
        return nil, fmt.Errorf("first argument must be string")
    }
    prompt := args[0].(engine.StringValue).Value()
    
    // Optional second parameter with default
    maxTokens := 100 // default
    if len(args) > 1 {
        if args[1].Type() == engine.TypeNumber {
            maxTokens = int(args[1].(engine.NumberValue).Value())
        }
    }
    
    // Optional options object
    options := make(map[string]interface{})
    if len(args) > 2 && args[2].Type() == engine.TypeObject {
        options = args[2].ToGo().(map[string]interface{})
    }
    
    // ... use parameters
}
```

### Validating Complex Arguments

```go
func (b *Bridge) ValidateMethod(name string, args []engine.ScriptValue) error {
    switch name {
    case "createAgent":
        if len(args) != 1 {
            return fmt.Errorf("createAgent expects 1 argument")
        }
        if args[0].Type() != engine.TypeObject {
            return fmt.Errorf("argument must be an object")
        }
        
        // Validate required fields
        obj := args[0].(engine.ObjectValue)
        fields := obj.Fields()
        
        if _, ok := fields["model"]; !ok {
            return fmt.Errorf("missing required field: model")
        }
        if fields["model"].Type() != engine.TypeString {
            return fmt.Errorf("model must be a string")
        }
        
        return nil
    }
    return fmt.Errorf("unknown method: %s", name)
}
```

### Working with Callbacks

```go
func (b *Bridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
    if name == "streamChat" {
        if len(args) < 2 || args[1].Type() != engine.TypeFunction {
            return nil, fmt.Errorf("second argument must be callback function")
        }
        
        prompt := args[0].(engine.StringValue).Value()
        callback := args[1].(engine.FunctionValue).Value()
        
        // Use callback in streaming
        err := b.llm.StreamChat(prompt, func(chunk string) error {
            // Call the script callback
            // Engine-specific implementation
            return nil
        })
        
        if err != nil {
            return nil, err
        }
        return engine.NewNilValue(), nil
    }
}
```

## Testing Your Migration

1. **Update test mocks:**
```go
type MockBridge struct{}

func (m *MockBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
    // Return appropriate ScriptValue
    return engine.NewStringValue("mock result"), nil
}

func (m *MockBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
    return nil
}
```

2. **Write type safety tests:**
```go
func TestBridgeTypeSafety(t *testing.T) {
    bridge := NewMyBridge()
    
    // Test with wrong type
    args := []engine.ScriptValue{
        engine.NewNumberValue(42), // Should be string
    }
    
    _, err := bridge.ExecuteMethod(context.Background(), "chat", args)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected string")
}
```

3. **Benchmark performance:**
```go
func BenchmarkScriptValueVsInterface(b *testing.B) {
    bridge := NewMyBridge()
    args := []engine.ScriptValue{
        engine.NewStringValue("test prompt"),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bridge.ExecuteMethod(context.Background(), "chat", args)
    }
}
```

## Troubleshooting

### Common Errors

1. **"interface conversion: engine.ScriptValue is engine.StringValue, not *engine.StringValue"**
   - ScriptValue types are value types, not pointers
   - Use `args[0].(engine.StringValue)` not `args[0].(*engine.StringValue)`

2. **"cannot convert nil to ScriptValue"**
   - Use `engine.NewNilValue()` for nil values
   - Never return bare `nil` as ScriptValue

3. **Type check failures in tests**
   - Ensure mocks return proper ScriptValue types
   - Use `engine.ConvertToScriptValue()` for test data

### Best Practices

1. **Always validate types** before casting
2. **Use descriptive error messages** that include expected and actual types
3. **Leverage the conversion helpers** for complex types
4. **Write comprehensive tests** for type edge cases
5. **Document expected types** in your bridge's API documentation

## Example: Complete Bridge Migration

Here's a complete example of migrating a simple LLM bridge:

**Before:**
```go
type LLMBridge struct {
    llm llms.LLM
}

func (b *LLMBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
    switch name {
    case "chat":
        prompt := args[0].(string) // Dangerous!
        response, err := b.llm.Chat(ctx, prompt)
        if err != nil {
            return nil, err
        }
        return response.Content, nil
    }
    return nil, fmt.Errorf("unknown method: %s", name)
}
```

**After:**
```go
type LLMBridge struct {
    llm llms.LLM
}

func (b *LLMBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
    switch name {
    case "chat":
        // Validate arguments
        if len(args) < 1 || args[0].Type() != engine.TypeString {
            return nil, fmt.Errorf("chat requires string prompt, got %v", args[0].Type())
        }
        
        // Safe extraction
        prompt := args[0].(engine.StringValue).Value()
        
        // Call go-llms
        response, err := b.llm.Chat(ctx, prompt)
        if err != nil {
            return nil, err
        }
        
        // Return ScriptValue
        return engine.NewStringValue(response.Content), nil
    }
    return nil, fmt.Errorf("unknown method: %s", name)
}

func (b *LLMBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
    switch name {
    case "chat":
        if len(args) < 1 {
            return fmt.Errorf("chat requires at least 1 argument")
        }
        if args[0].Type() != engine.TypeString {
            return fmt.Errorf("first argument must be string, got %s", args[0].Type())
        }
        return nil
    default:
        return fmt.Errorf("unknown method: %s", name)
    }
}
```

## Conclusion

The ScriptValue migration provides significant benefits:
- **Compile-time type safety** catches errors early
- **Clear error messages** make debugging easier
- **Consistent behavior** across all script engines
- **Better performance** with less reflection

While the migration requires updating method signatures and type handling, the improved safety and maintainability make it worthwhile. Use this guide and the provided patterns to ensure a smooth transition.