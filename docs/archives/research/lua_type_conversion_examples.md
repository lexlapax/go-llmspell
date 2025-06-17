# Lua Type Conversion Examples

## Basic Type Conversions

### Primitives
```go
// Go → Lua
nil         → lua.LNil
true/false  → lua.LTrue/lua.LFalse
42          → lua.LNumber(42)
3.14        → lua.LNumber(3.14)
"hello"     → lua.LString("hello")

// Lua → Go
lua.LNil    → nil
lua.LTrue   → true
lua.LFalse  → false
lua.LNumber → float64
lua.LString → string
```

### Collections
```go
// Go array → Lua table (1-indexed)
[]interface{}{1, 2, 3} → {[1]=1, [2]=2, [3]=3}

// Go map → Lua table
map[string]interface{}{
    "name": "John",
    "age": 30,
} → {name="John", age=30}

// Nested structures
map[string]interface{}{
    "user": map[string]interface{}{
        "id": 123,
        "roles": []interface{}{"admin", "user"},
    },
} → {user={id=123, roles={[1]="admin", [2]="user"}}}
```

## Bridge Object Conversions

### Bridge as UserData
```lua
-- In Lua, bridge objects appear as userdata with methods
local llm = require("llm")
local client = llm.new_client({provider="openai"})

-- Method calls work naturally
local response = client:generate({
    prompt = "Hello, world!",
    max_tokens = 100
})

-- Type checking
print(type(client))  -- "userdata"
print(tostring(client))  -- "LLMBridge{provider=openai}"
```

### Bridge Method Wrapping
```go
// Go bridge method
func (b *LLMBridge) Generate(ctx context.Context, params map[string]interface{}) (string, error)

// Becomes Lua method
function client:generate(params)
    -- Automatic type conversion happens here
    -- params table → Go map
    -- return string → Lua string
    -- errors → Lua error()
end
```

## Error Handling

### Go Error → Lua Error
```go
// Go function returns error
func doSomething() error {
    return fmt.Errorf("something went wrong: %w", io.EOF)
}

// In Lua
local ok, result = pcall(function()
    return bridge.doSomething()
end)

if not ok then
    print("Error:", result)  -- "Error: something went wrong: EOF"
end
```

### Lua Error → Go Error
```lua
-- Lua function throws error
function riskyOperation()
    error("operation failed")
end

// Go receives as error
err := L.PCall(0, 0, nil)
if err != nil {
    // err.Error() == "operation failed"
}
```

## Function Conversions

### Go Function → Lua Function
```go
// Register Go function
L.SetGlobal("formatNumber", L.NewFunction(func(L *lua.LState) int {
    num := L.CheckNumber(1)
    precision := L.OptInt(2, 2)
    
    result := fmt.Sprintf("%.*f", precision, float64(num))
    L.Push(lua.LString(result))
    return 1
}))
```

```lua
-- Use in Lua
print(formatNumber(3.14159))      -- "3.14"
print(formatNumber(3.14159, 4))   -- "3.1416"
```

### Lua Function → Go Function
```lua
-- Lua function
function add(a, b)
    return a + b
end
```

```go
// Call from Go
L.GetGlobal("add")
L.Push(lua.LNumber(10))
L.Push(lua.LNumber(20))
L.Call(2, 1)
result := L.Get(-1).(lua.LNumber)  // 30
L.Pop(1)
```

## Complex Type Examples

### Time/Duration Handling
```go
// Go time.Time → Lua userdata with methods
type TimeUserData struct {
    time.Time
}

// Register methods
func (t *TimeUserData) Format(L *lua.LState) int {
    format := L.CheckString(1)
    L.Push(lua.LString(t.Time.Format(format)))
    return 1
}

// Usage in Lua
local now = time.now()
print(now:format("2006-01-02"))  -- "2025-06-17"
```

### Struct Conversions
```go
// Go struct
type User struct {
    ID       int      `json:"id"`
    Name     string   `json:"name"`
    Email    string   `json:"email"`
    Tags     []string `json:"tags"`
    Metadata map[string]interface{} `json:"metadata"`
}

// Option 1: Convert to table
user := User{
    ID:    123,
    Name:  "John Doe",
    Email: "john@example.com",
    Tags:  []string{"vip", "subscriber"},
    Metadata: map[string]interface{}{
        "lastLogin": "2025-06-17",
        "score":     95.5,
    },
}

// Becomes Lua table:
{
    id = 123,
    name = "John Doe",
    email = "john@example.com",
    tags = {[1]="vip", [2]="subscriber"},
    metadata = {
        lastLogin = "2025-06-17",
        score = 95.5
    }
}

// Option 2: Keep as UserData with field access
local user = getUserData()
print(user.name)  -- "John Doe"
print(user.tags[1])  -- "vip"
```

## Channel Support

### Go Channel ↔ Lua Channel
```go
// Create channel in Go
ch := make(chan string, 10)
L.SetGlobal("ch", lua.LChannel(ch))
```

```lua
-- Use in Lua
local ch = ch  -- Global channel

-- Send
ch:send("hello")

-- Receive
local msg = ch:receive()
print(msg)  -- "hello"

-- Select with timeout
local msg, ok = ch:receive(1.0)  -- 1 second timeout
if ok then
    print("Got:", msg)
else
    print("Timeout")
end
```

## Performance Considerations

### Cached Conversions
```go
// Common strings are cached
"true", "false", "null", "" → Cached LString values

// Small integers are cached
-100 to 100 → Cached LNumber values

// Reduces allocation pressure
```

### Lazy Conversions
```go
// Large structures can use lazy conversion
type LazyTable struct {
    data map[string]interface{}
    converted map[string]lua.LValue
}

// Only convert accessed fields
func (lt *LazyTable) Get(L *lua.LState, key string) lua.LValue {
    if lv, exists := lt.converted[key]; exists {
        return lv
    }
    
    // Convert on demand
    lv, _ := converter.ToLValue(L, lt.data[key])
    lt.converted[key] = lv
    return lv
}
```

## Circular Reference Handling

```go
// Circular structure in Go
type Node struct {
    Value string
    Next  *Node
}

node1 := &Node{Value: "A"}
node2 := &Node{Value: "B", Next: node1}
node1.Next = node2  // Circular!

// Converter detects and handles
// First occurrence: creates table/userdata
// Second occurrence: returns same reference
// Result in Lua: tables maintain circular reference
```

## Type Validation Examples

```go
// Validate before conversion
params := map[string]interface{}{
    "prompt": "Hello",
    "max_tokens": "100",  // Wrong type!
    "temperature": 0.7,
}

// Converter can validate against schema
schema := map[string]string{
    "prompt": "string",
    "max_tokens": "number",
    "temperature": "number",
}

// Error: max_tokens: cannot convert string "100" to number
// (Though in this case, string→number conversion would work)
```

## Integration with Bridges

```lua
-- All bridge methods get automatic conversion
local llm = require("llm")
local agent = require("agent")

-- Complex nested structure automatically converted
local result = agent.create({
    name = "Assistant",
    llm = llm.new_client({provider = "openai"}),
    tools = {
        search = {
            description = "Search the web",
            parameters = {
                {name = "query", type = "string", required = true}
            }
        }
    },
    hooks = {
        before_generate = function(messages)
            print("Generating with", #messages, "messages")
            return messages  -- Can modify
        end
    }
})

-- All types converted seamlessly:
-- - Tables ↔ maps
-- - Functions ↔ Function interface
-- - Bridge objects ↔ UserData
```