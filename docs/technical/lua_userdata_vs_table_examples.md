# UserData vs Table Bridge Implementation Examples

This document provides practical examples of implementing bridge objects using UserData, Tables, and hybrid approaches in GopherLua.

## UserData Implementation

### Basic UserData Bridge

```go
package userdata

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
    "github.com/your/go-llms/bridge"
)

// Bridge wrapper for UserData
type LLMBridgeUD struct {
    bridge bridge.LLMBridge
    id     string
}

// Register UserData type
func RegisterLLMBridgeType(L *lua.LState) {
    mt := L.NewTypeMetatable("LLMBridge")
    L.SetGlobal("LLMBridge", mt)
    
    // Methods
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), llmBridgeMethods))
    L.SetField(mt, "__tostring", L.NewFunction(llmBridgeToString))
    L.SetField(mt, "__eq", L.NewFunction(llmBridgeEqual))
    
    // Prevent modification
    L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
        L.RaiseError("cannot modify LLMBridge")
        return 0
    }))
}

var llmBridgeMethods = map[string]lua.LGFunction{
    "createAgent": llmCreateAgent,
    "listModels":  llmListModels,
    "generate":    llmGenerate,
    "stream":      llmStream,
    "getInfo":     llmGetInfo,
}

// Constructor
func NewLLMBridge(L *lua.LState) int {
    config := L.CheckTable(1)
    
    // Extract configuration
    provider := getStringField(L, config, "provider", "openai")
    apiKey := getStringField(L, config, "apiKey", "")
    
    // Create actual bridge
    bridgeImpl, err := bridge.NewLLMBridge(bridge.Config{
        Provider: provider,
        APIKey:   apiKey,
    })
    if err != nil {
        L.RaiseError("failed to create LLM bridge: %v", err)
        return 0
    }
    
    // Wrap in UserData
    ud := L.NewUserData()
    ud.Value = &LLMBridgeUD{
        bridge: bridgeImpl,
        id:     fmt.Sprintf("llm-%s-%d", provider, time.Now().Unix()),
    }
    L.SetMetatable(ud, L.GetTypeMetatable("LLMBridge"))
    L.Push(ud)
    
    return 1
}

// Method implementations
func llmCreateAgent(L *lua.LState) int {
    b := checkLLMBridge(L, 1)
    config := L.CheckTable(2)
    
    model := getStringField(L, config, "model", "gpt-4")
    temperature := getNumberField(L, config, "temperature", 0.7)
    
    agent, err := b.bridge.CreateAgent(bridge.AgentConfig{
        Model:       model,
        Temperature: temperature,
    })
    if err != nil {
        L.Push(lua.LNil)
        L.Push(lua.LString(err.Error()))
        return 2
    }
    
    // Wrap agent in UserData
    agentUD := L.NewUserData()
    agentUD.Value = &AgentUD{agent: agent}
    L.SetMetatable(agentUD, L.GetTypeMetatable("Agent"))
    L.Push(agentUD)
    
    return 1
}

func llmGenerate(L *lua.LState) int {
    b := checkLLMBridge(L, 1)
    prompt := L.CheckString(2)
    options := L.OptTable(3, L.NewTable())
    
    // Extract options
    opts := bridge.GenerateOptions{
        MaxTokens: int(getNumberField(L, options, "maxTokens", 1000)),
        Stream:    getBoolField(L, options, "stream", false),
    }
    
    response, err := b.bridge.Generate(prompt, opts)
    if err != nil {
        L.Push(lua.LNil)
        L.Push(lua.LString(err.Error()))
        return 2
    }
    
    L.Push(lua.LString(response))
    return 1
}

// Helper to check UserData type
func checkLLMBridge(L *lua.LState, idx int) *LLMBridgeUD {
    ud := L.CheckUserData(idx)
    if v, ok := ud.Value.(*LLMBridgeUD); ok {
        return v
    }
    L.ArgError(idx, "LLMBridge expected")
    return nil
}

// String representation
func llmBridgeToString(L *lua.LState) int {
    b := checkLLMBridge(L, 1)
    L.Push(lua.LString(fmt.Sprintf("LLMBridge<%s>", b.id)))
    return 1
}

// Example usage
func ExampleUserDataBridge() {
    L := lua.NewState()
    defer L.Close()
    
    RegisterLLMBridgeType(L)
    L.SetGlobal("newLLMBridge", L.NewFunction(NewLLMBridge))
    
    err := L.DoString(`
        -- Create bridge
        local bridge = newLLMBridge({
            provider = "openai",
            apiKey = "sk-..."
        })
        
        print(bridge) -- LLMBridge<llm-openai-1234567890>
        
        -- Create agent
        local agent, err = bridge:createAgent({
            model = "gpt-4",
            temperature = 0.7
        })
        
        if err then
            error("Failed to create agent: " .. err)
        end
        
        -- Generate response
        local response = bridge:generate("Hello, world!", {
            maxTokens = 100
        })
        
        print("Response:", response)
        
        -- This will error
        -- bridge.newField = "value" -- Error: cannot modify LLMBridge
    `)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Advanced UserData with Nested Objects

```go
package advanced

// Complex bridge with nested objects
type WorkflowBridgeUD struct {
    bridge   bridge.WorkflowBridge
    id       string
    metadata map[string]interface{}
}

func RegisterWorkflowBridgeType(L *lua.LState) {
    mt := L.NewTypeMetatable("WorkflowBridge")
    
    // Method table
    methods := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "createFlow":    workflowCreateFlow,
        "execute":       workflowExecute,
        "getState":      workflowGetState,
        "addStep":       workflowAddStep,
        "removeStep":    workflowRemoveStep,
        "validate":      workflowValidate,
        "getMetadata":   workflowGetMetadata,
        "setMetadata":   workflowSetMetadata,
    })
    
    L.SetField(mt, "__index", methods)
    
    // Allow iteration over steps
    L.SetField(mt, "__pairs", L.NewFunction(workflowPairs))
    L.SetField(mt, "__len", L.NewFunction(workflowLen))
}

// Iteration support
func workflowPairs(L *lua.LState) int {
    w := checkWorkflowBridge(L, 1)
    
    // Create iterator function
    iter := L.NewFunction(func(L *lua.LState) int {
        w := checkWorkflowBridge(L, 1)
        idx := L.OptInt(2, 0)
        
        steps := w.bridge.GetSteps()
        if idx >= len(steps) {
            L.Push(lua.LNil)
            return 1
        }
        
        L.Push(lua.LNumber(idx + 1))
        L.Push(convertStepToLua(L, steps[idx]))
        return 2
    })
    
    L.Push(iter)
    L.Push(L.Get(1)) // workflow object
    L.Push(lua.LNumber(0)) // initial index
    return 3
}

// Metadata access with dot notation
func workflowGetMetadata(L *lua.LState) int {
    w := checkWorkflowBridge(L, 1)
    key := L.CheckString(2)
    
    if val, ok := w.metadata[key]; ok {
        L.Push(goValueToLua(L, val))
    } else {
        L.Push(lua.LNil)
    }
    return 1
}

func workflowSetMetadata(L *lua.LState) int {
    w := checkWorkflowBridge(L, 1)
    key := L.CheckString(2)
    value := L.Get(3)
    
    w.metadata[key] = luaValueToGo(value)
    return 0
}
```

## Table Implementation

### Basic Table Bridge

```go
package table

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
    "github.com/your/go-llms/bridge"
)

// Create table-based bridge
func CreateLLMBridgeTable(L *lua.LState, b bridge.LLMBridge) *lua.LTable {
    tbl := L.NewTable()
    
    // Type information
    L.SetField(tbl, "_type", lua.LString("LLMBridge"))
    L.SetField(tbl, "_id", lua.LString(generateID()))
    
    // Store bridge reference (as light userdata)
    L.SetField(tbl, "_bridge", lua.LLightUserData(unsafe.Pointer(&b)))
    
    // Public properties
    info := b.GetInfo()
    L.SetField(tbl, "provider", lua.LString(info.Provider))
    L.SetField(tbl, "version", lua.LString(info.Version))
    L.SetField(tbl, "capabilities", convertCapabilities(L, info.Capabilities))
    
    // Methods
    methods := L.NewTable()
    L.SetField(methods, "createAgent", L.NewFunction(createAgentTableMethod))
    L.SetField(methods, "generate", L.NewFunction(generateTableMethod))
    L.SetField(methods, "listModels", L.NewFunction(listModelsTableMethod))
    L.SetField(tbl, "methods", methods)
    
    // Metatable for method syntax
    mt := L.NewTable()
    L.SetField(mt, "__index", L.NewFunction(tableBridgeIndex))
    L.SetField(mt, "__tostring", L.NewFunction(tableBridgeToString))
    L.SetMetatable(tbl, mt)
    
    return tbl
}

// Index metamethod for method syntax
func tableBridgeIndex(L *lua.LState) int {
    tbl := L.CheckTable(1)
    key := L.CheckString(2)
    
    // Check if it's a method
    L.GetField(tbl, "methods")
    methods := L.CheckTable(-1)
    L.GetField(methods, key)
    
    if !L.IsNil(-1) {
        // Return bound method
        return 1
    }
    L.Pop(2)
    
    // Regular field access
    L.GetField(tbl, key)
    return 1
}

// Method implementation
func generateTableMethod(L *lua.LState) int {
    tbl := checkBridgeTable(L, 1, "LLMBridge")
    prompt := L.CheckString(2)
    options := L.OptTable(3, L.NewTable())
    
    // Get bridge from table
    bridge := getBridgeFromTable(L, tbl)
    
    // Execute
    response, err := bridge.Generate(prompt, convertOptions(L, options))
    if err != nil {
        L.Push(lua.LNil)
        L.Push(lua.LString(err.Error()))
        return 2
    }
    
    L.Push(lua.LString(response))
    return 1
}

// Helper functions
func checkBridgeTable(L *lua.LState, idx int, expectedType string) *lua.LTable {
    tbl := L.CheckTable(idx)
    
    L.GetField(tbl, "_type")
    if typ := L.Get(-1); lua.LVAsString(typ) != expectedType {
        L.ArgError(idx, fmt.Sprintf("expected %s table", expectedType))
    }
    L.Pop(1)
    
    return tbl
}

func getBridgeFromTable(L *lua.LState, tbl *lua.LTable) bridge.LLMBridge {
    L.GetField(tbl, "_bridge")
    if ptr, ok := L.Get(-1).(lua.LLightUserData); ok {
        L.Pop(1)
        return *(*bridge.LLMBridge)(unsafe.Pointer(ptr))
    }
    L.Pop(1)
    L.RaiseError("invalid bridge reference")
    return nil
}

// Example with rich table structure
func ExampleTableBridge() {
    L := lua.NewState()
    defer L.Close()
    
    // Create bridge
    bridgeImpl, _ := bridge.NewLLMBridge(bridge.Config{Provider: "openai"})
    bridgeTbl := CreateLLMBridgeTable(L, bridgeImpl)
    L.SetGlobal("bridge", bridgeTbl)
    
    err := L.DoString(`
        -- Inspect bridge
        print("Provider:", bridge.provider)
        print("Version:", bridge.version)
        
        -- Check capabilities
        for cap, enabled in pairs(bridge.capabilities) do
            print("  " .. cap .. ":", enabled)
        end
        
        -- Use methods
        local response = bridge:generate("Hello!", {maxTokens = 50})
        print("Response:", response)
        
        -- Add custom fields (allowed with tables)
        bridge.customData = {
            createdAt = os.time(),
            tags = {"production", "gpt-4"}
        }
        
        -- Iterate over bridge properties
        for k, v in pairs(bridge) do
            if not k:match("^_") then -- Skip internal fields
                print(k, "=", v)
            end
        end
    `)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Table with Validation

```go
package validation

// Validated table bridge
func CreateValidatedBridgeTable(L *lua.LState, bridge bridge.Bridge) *lua.LTable {
    tbl := L.NewTable()
    
    // Schema definition
    schema := map[string]FieldDef{
        "type":     {Type: "string", ReadOnly: true, Required: true},
        "id":       {Type: "string", ReadOnly: true, Required: true},
        "name":     {Type: "string", ReadOnly: false, Required: false},
        "config":   {Type: "table", ReadOnly: false, Required: false},
        "metadata": {Type: "table", ReadOnly: false, Required: false},
    }
    
    // Initialize required fields
    L.SetField(tbl, "type", lua.LString(bridge.Type()))
    L.SetField(tbl, "id", lua.LString(bridge.ID()))
    
    // Validation metatable
    mt := L.NewTable()
    
    L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
        tbl := L.CheckTable(1)
        key := L.CheckString(2)
        value := L.Get(3)
        
        // Validate against schema
        if def, ok := schema[key]; ok {
            if def.ReadOnly {
                L.RaiseError("field '%s' is read-only", key)
                return 0
            }
            
            if !validateType(value, def.Type) {
                L.RaiseError("field '%s' must be of type %s", key, def.Type)
                return 0
            }
        }
        
        // Allow setting
        L.RawSet(tbl, lua.LString(key), value)
        return 0
    }))
    
    L.SetMetatable(tbl, mt)
    return tbl
}

type FieldDef struct {
    Type     string
    ReadOnly bool
    Required bool
}

func validateType(value lua.LValue, expectedType string) bool {
    switch expectedType {
    case "string":
        return value.Type() == lua.LTString
    case "number":
        return value.Type() == lua.LTNumber
    case "boolean":
        return value.Type() == lua.LTBool
    case "table":
        return value.Type() == lua.LTTable
    case "function":
        return value.Type() == lua.LTFunction
    default:
        return true
    }
}
```

## Hybrid Implementation

### UserData Core with Table Interface

```go
package hybrid

type HybridBridge struct {
    core       bridge.Bridge
    properties *lua.LTable
    methods    map[string]lua.LGFunction
}

func CreateHybridBridge(L *lua.LState, core bridge.Bridge) lua.LValue {
    hybrid := &HybridBridge{
        core:       core,
        properties: L.NewTable(),
        methods:    make(map[string]lua.LGFunction),
    }
    
    // Register methods
    hybrid.methods["execute"] = hybridExecute
    hybrid.methods["getState"] = hybridGetState
    hybrid.methods["configure"] = hybridConfigure
    
    // Initialize properties
    L.SetField(hybrid.properties, "type", lua.LString(core.Type()))
    L.SetField(hybrid.properties, "id", lua.LString(core.ID()))
    L.SetField(hybrid.properties, "created", lua.LNumber(time.Now().Unix()))
    
    // Create UserData
    ud := L.NewUserData()
    ud.Value = hybrid
    
    // Metatable with table-like behavior
    mt := L.NewTypeMetatable("HybridBridge")
    
    // Index for property/method access
    L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
        h := checkHybridBridge(L, 1)
        key := L.CheckString(2)
        
        // Check methods
        if method, ok := h.methods[key]; ok {
            L.Push(L.NewClosure(method, ud))
            return 1
        }
        
        // Check properties
        L.GetField(h.properties, key)
        if !L.IsNil(-1) {
            return 1
        }
        L.Pop(1)
        
        // Core properties
        switch key {
        case "info":
            L.Push(convertBridgeInfo(L, h.core.GetInfo()))
            return 1
        }
        
        L.Push(lua.LNil)
        return 1
    }))
    
    // NewIndex for property setting
    L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
        h := checkHybridBridge(L, 1)
        key := L.CheckString(2)
        value := L.Get(3)
        
        // Protect core properties
        if key == "type" || key == "id" {
            L.RaiseError("cannot modify core property: %s", key)
            return 0
        }
        
        L.SetField(h.properties, key, value)
        return 0
    }))
    
    // Pairs for iteration
    L.SetField(mt, "__pairs", L.NewFunction(func(L *lua.LState) int {
        h := checkHybridBridge(L, 1)
        
        // Merge core and custom properties
        merged := L.NewTable()
        
        // Copy properties
        h.properties.ForEach(func(k, v lua.LValue) {
            L.SetField(merged, k.String(), v)
        })
        
        // Add methods as properties
        for name := range h.methods {
            L.SetField(merged, name, lua.LString("<method>"))
        }
        
        // Use default pairs on merged table
        L.Push(L.GetGlobal("pairs"))
        L.Push(merged)
        L.Call(1, 3)
        return 3
    }))
    
    // String representation
    L.SetField(mt, "__tostring", L.NewFunction(func(L *lua.LState) int {
        h := checkHybridBridge(L, 1)
        L.Push(lua.LString(fmt.Sprintf("HybridBridge<%s:%s>", 
            h.core.Type(), h.core.ID())))
        return 1
    }))
    
    L.SetMetatable(ud, mt)
    return ud
}

// Example usage
func ExampleHybridBridge() {
    L := lua.NewState()
    defer L.Close()
    
    // Register constructor
    L.SetGlobal("createBridge", L.NewFunction(func(L *lua.LState) int {
        bridgeType := L.CheckString(1)
        
        var core bridge.Bridge
        switch bridgeType {
        case "llm":
            core, _ = bridge.NewLLMBridge(bridge.Config{})
        case "workflow":
            core, _ = bridge.NewWorkflowBridge(bridge.Config{})
        default:
            L.RaiseError("unknown bridge type: %s", bridgeType)
            return 0
        }
        
        L.Push(CreateHybridBridge(L, core))
        return 1
    }))
    
    err := L.DoString(`
        -- Create hybrid bridge
        local bridge = createBridge("llm")
        
        -- Access core properties
        print("Type:", bridge.type)
        print("ID:", bridge.id)
        
        -- Add custom properties
        bridge.name = "Production LLM"
        bridge.tags = {"gpt-4", "production"}
        bridge.stats = {
            requests = 0,
            errors = 0,
            avgLatency = 0
        }
        
        -- Use methods
        local state = bridge:getState()
        
        -- Iterate over all properties
        print("\nAll properties:")
        for k, v in pairs(bridge) do
            print("  " .. k .. ":", v)
        end
        
        -- Update stats
        bridge.stats.requests = bridge.stats.requests + 1
        
        -- This will error
        -- bridge.type = "new-type" -- Error: cannot modify core property
    `)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Proxy Pattern Implementation

```go
package proxy

type BridgeProxy struct {
    target     bridge.Bridge
    interceptor func(method string, args []lua.LValue) ([]lua.LValue, error)
    cache      map[string]lua.LValue
    mu         sync.RWMutex
}

func CreateProxyBridge(L *lua.LState, target bridge.Bridge, opts ProxyOptions) lua.LValue {
    proxy := &BridgeProxy{
        target: target,
        cache:  make(map[string]lua.LValue),
    }
    
    if opts.EnableCaching {
        proxy.interceptor = createCachingInterceptor(proxy)
    }
    if opts.EnableLogging {
        oldInterceptor := proxy.interceptor
        proxy.interceptor = createLoggingInterceptor(oldInterceptor)
    }
    
    ud := L.NewUserData()
    ud.Value = proxy
    
    mt := L.NewTypeMetatable("BridgeProxy")
    
    L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
        p := checkBridgeProxy(L, 1)
        key := L.CheckString(2)
        
        // Check cache first
        p.mu.RLock()
        if cached, ok := p.cache[key]; ok {
            p.mu.RUnlock()
            L.Push(cached)
            return 1
        }
        p.mu.RUnlock()
        
        // Create method wrapper
        if method := getTargetMethod(p.target, key); method != nil {
            wrapper := L.NewFunction(func(L *lua.LState) int {
                // Collect arguments
                args := make([]lua.LValue, L.GetTop()-1)
                for i := 0; i < len(args); i++ {
                    args[i] = L.Get(i + 2)
                }
                
                // Apply interceptor
                if p.interceptor != nil {
                    results, err := p.interceptor(key, args)
                    if err != nil {
                        L.RaiseError("proxy error: %v", err)
                        return 0
                    }
                    
                    for _, r := range results {
                        L.Push(r)
                    }
                    return len(results)
                }
                
                // Direct call
                return method(L)
            })
            
            // Cache the wrapper
            p.mu.Lock()
            p.cache[key] = wrapper
            p.mu.Unlock()
            
            L.Push(wrapper)
            return 1
        }
        
        L.Push(lua.LNil)
        return 1
    }))
    
    L.SetMetatable(ud, mt)
    return ud
}

// Caching interceptor
func createCachingInterceptor(proxy *BridgeProxy) func(string, []lua.LValue) ([]lua.LValue, error) {
    resultCache := make(map[string][]lua.LValue)
    
    return func(method string, args []lua.LValue) ([]lua.LValue, error) {
        // Create cache key
        key := fmt.Sprintf("%s:%v", method, args)
        
        if cached, ok := resultCache[key]; ok {
            return cached, nil
        }
        
        // Execute actual method
        results, err := executeMethod(proxy.target, method, args)
        if err != nil {
            return nil, err
        }
        
        // Cache results
        resultCache[key] = results
        return results, nil
    }
}
```

## Performance Comparison

### Benchmark Implementation

```go
package benchmark

import (
    "testing"
    lua "github.com/yuin/gopher-lua"
)

func BenchmarkBridgeImplementations(b *testing.B) {
    implementations := []struct {
        name   string
        create func(L *lua.LState) lua.LValue
    }{
        {"UserData", createUserDataBridge},
        {"Table", createTableBridge},
        {"Hybrid", createHybridBridge},
        {"Proxy", createProxyBridge},
    }
    
    for _, impl := range implementations {
        b.Run(impl.name+"_MethodCall", func(b *testing.B) {
            L := lua.NewState()
            defer L.Close()
            
            bridge := impl.create(L)
            L.SetGlobal("bridge", bridge)
            
            L.DoString(`
                function benchmark()
                    return bridge:execute("test")
                end
            `)
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                L.CallByParam(lua.P{
                    Fn:      L.GetGlobal("benchmark"),
                    NRet:    1,
                    Protect: false,
                })
                L.Pop(1)
            }
        })
        
        b.Run(impl.name+"_PropertyAccess", func(b *testing.B) {
            L := lua.NewState()
            defer L.Close()
            
            bridge := impl.create(L)
            L.SetGlobal("bridge", bridge)
            
            L.DoString(`
                function benchmark()
                    return bridge.type
                end
            `)
            
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                L.CallByParam(lua.P{
                    Fn:      L.GetGlobal("benchmark"),
                    NRet:    1,
                    Protect: false,
                })
                L.Pop(1)
            }
        })
    }
}

// Memory usage comparison
func TestMemoryUsage(t *testing.T) {
    L := lua.NewState()
    defer L.Close()
    
    implementations := []struct {
        name   string
        create func(L *lua.LState, i int) lua.LValue
    }{
        {"UserData", func(L *lua.LState, i int) lua.LValue {
            return createIndexedUserDataBridge(L, i)
        }},
        {"Table", func(L *lua.LState, i int) lua.LValue {
            return createIndexedTableBridge(L, i)
        }},
    }
    
    for _, impl := range implementations {
        initialMem := L.MemUsage()
        
        // Create 1000 objects
        for i := 0; i < 1000; i++ {
            obj := impl.create(L, i)
            L.SetGlobal(fmt.Sprintf("%s%d", impl.name, i), obj)
        }
        
        finalMem := L.MemUsage()
        memPerObject := (finalMem - initialMem) / 1000
        
        t.Logf("%s: %d bytes per object", impl.name, memPerObject)
        
        // Cleanup
        for i := 0; i < 1000; i++ {
            L.SetGlobal(fmt.Sprintf("%s%d", impl.name, i), lua.LNil)
        }
        L.DoString(`collectgarbage("collect")`)
    }
}
```

## Migration Examples

### Migrating from Table to UserData

```go
package migration

func CreateMigrationLayer(L *lua.LState) {
    // Register both implementations
    RegisterUserDataBridge(L)
    RegisterTableBridge(L)
    
    // Compatibility function
    L.SetGlobal("createBridge", L.NewFunction(func(L *lua.LState) int {
        config := L.CheckTable(1)
        
        // Check for version hint
        L.GetField(config, "version")
        version := L.OptString(-1, "v1")
        L.Pop(1)
        
        switch version {
        case "v1":
            // Old table-based API
            L.Push(createTableBridge(L, config))
        case "v2":
            // New UserData API
            L.Push(createUserDataBridge(L, config))
        default:
            // Auto-detect based on feature usage
            if shouldUseUserData(L, config) {
                L.Push(createUserDataBridge(L, config))
            } else {
                L.Push(createTableBridge(L, config))
            }
        }
        
        return 1
    }))
    
    // Migration helper
    L.SetGlobal("migrateBridge", L.NewFunction(func(L *lua.LState) int {
        old := L.Get(1)
        
        switch old.Type() {
        case lua.LTTable:
            // Convert table to UserData
            L.Push(tableToUserData(L, old.(*lua.LTable)))
        case lua.LTUserData:
            // Already UserData
            L.Push(old)
        default:
            L.RaiseError("invalid bridge type")
        }
        
        return 1
    }))
}

// Gradual migration with compatibility wrapper
func CreateCompatibilityWrapper(L *lua.LState, ud lua.LValue) *lua.LTable {
    wrapper := L.NewTable()
    
    // Store UserData internally
    L.SetField(wrapper, "_ud", ud)
    
    mt := L.NewTable()
    
    // Forward method calls to UserData
    L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
        wrapper := L.CheckTable(1)
        key := L.CheckString(2)
        
        // Get UserData
        L.GetField(wrapper, "_ud")
        ud := L.Get(-1)
        L.Pop(1)
        
        // Forward property access
        L.GetField(ud, key)
        return 1
    }))
    
    // Forward method calls
    L.SetField(mt, "__call", L.NewFunction(func(L *lua.LState) int {
        wrapper := L.CheckTable(1)
        
        // Get UserData
        L.GetField(wrapper, "_ud")
        ud := L.Get(-1)
        L.Remove(-1)
        
        // Replace wrapper with UserData in arguments
        L.Replace(1)
        
        // Call UserData method
        return L.CallMeta(ud, "__call")
    }))
    
    L.SetMetatable(wrapper, mt)
    return wrapper
}
```

## Summary

These examples demonstrate various approaches to bridge implementation:

1. **UserData**: Type-safe, performant, encapsulated
2. **Table**: Flexible, transparent, Lua-idiomatic
3. **Hybrid**: Combines benefits of both approaches
4. **Proxy**: Advanced patterns for caching and interception

Choose based on your specific requirements for type safety, performance, flexibility, and compatibility.