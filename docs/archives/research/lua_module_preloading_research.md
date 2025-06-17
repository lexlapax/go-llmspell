# Lua Module Preloading and Lazy Initialization Research

This document investigates module preloading and lazy initialization strategies in GopherLua for optimizing startup time, memory usage, and module management.

## Executive Summary

GopherLua provides flexible module loading through `require`, module preloading via `PreloadModule`, and support for custom loaders. Effective lazy initialization combined with preloading can significantly improve performance and resource utilization.

## Module System in GopherLua

### Basic Module Loading

GopherLua implements Lua's module system with `require`:

```go
// Standard module loading
L.DoString(`
    local mymodule = require("mymodule")
    mymodule.doSomething()
`)

// Module is searched in:
// 1. Preloaded modules (_PRELOAD)
// 2. File system (using package.path)
// 3. Custom loaders
```

### Module Preloading API

```go
// Preload a Go module
L.PreloadModule("mymodule", func(L *lua.LState) int {
    mod := L.NewTable()
    L.SetField(mod, "version", lua.LString("1.0.0"))
    L.SetField(mod, "doSomething", L.NewFunction(doSomething))
    L.Push(mod)
    return 1
})

// Module is now available via require
```

## Preloading Strategies

### 1. Static Preloading

```go
type ModuleLoader struct {
    modules map[string]lua.LGFunction
}

func NewModuleLoader() *ModuleLoader {
    return &ModuleLoader{
        modules: map[string]lua.LGFunction{
            "llm":      LoadLLMModule,
            "tools":    LoadToolsModule,
            "workflow": LoadWorkflowModule,
            "state":    LoadStateModule,
            "events":   LoadEventsModule,
        },
    }
}

func (ml *ModuleLoader) PreloadAll(L *lua.LState) {
    for name, loader := range ml.modules {
        L.PreloadModule(name, loader)
    }
}

// Module implementation
func LoadLLMModule(L *lua.LState) int {
    exports := L.NewTable()
    
    // Module functions
    L.SetField(exports, "create", L.NewFunction(llmCreate))
    L.SetField(exports, "generate", L.NewFunction(llmGenerate))
    L.SetField(exports, "stream", L.NewFunction(llmStream))
    
    // Module metadata
    mt := L.NewTable()
    L.SetField(mt, "__tostring", L.NewFunction(func(L *lua.LState) int {
        L.Push(lua.LString("LLM Bridge Module v1.0"))
        return 1
    }))
    L.SetMetatable(exports, mt)
    
    L.Push(exports)
    return 1
}
```

### 2. Lazy Module Loading

```go
type LazyModuleLoader struct {
    definitions map[string]ModuleDefinition
    loaded      map[string]bool
    mu          sync.RWMutex
}

type ModuleDefinition struct {
    Name         string
    Dependencies []string
    Loader       lua.LGFunction
    InitOnce     sync.Once
    initialized  bool
}

func (lml *LazyModuleLoader) PreloadLazy(L *lua.LState) {
    for name, def := range lml.definitions {
        name := name // Capture for closure
        def := def   // Capture for closure
        
        L.PreloadModule(name, func(L *lua.LState) int {
            // Load dependencies first
            for _, dep := range def.Dependencies {
                L.DoString(fmt.Sprintf(`require("%s")`, dep))
            }
            
            // Initialize module once
            def.InitOnce.Do(func() {
                def.initialized = true
                lml.mu.Lock()
                lml.loaded[name] = true
                lml.mu.Unlock()
            })
            
            // Load the actual module
            return def.Loader(L)
        })
    }
}
```

### 3. Conditional Preloading

```go
type ConditionalLoader struct {
    profiles map[string][]string // Profile -> Module list
}

func (cl *ConditionalLoader) PreloadForProfile(L *lua.LState, profile string) {
    modules, ok := cl.profiles[profile]
    if !ok {
        modules = cl.profiles["default"]
    }
    
    for _, modName := range modules {
        switch modName {
        case "llm":
            if shouldLoadLLM(profile) {
                L.PreloadModule("llm", LoadLLMModule)
            }
        case "tools":
            if shouldLoadTools(profile) {
                L.PreloadModule("tools", LoadToolsModule)
            }
        // ... other modules
        }
    }
}

// Profile definitions
var LoadProfiles = map[string][]string{
    "minimal":  {"core"},
    "standard": {"core", "llm", "tools"},
    "full":     {"core", "llm", "tools", "workflow", "state", "events"},
    "agent":    {"core", "llm", "tools", "workflow"},
}
```

## Lazy Initialization Patterns

### 1. On-Demand Bridge Creation

```go
type LazyBridgeModule struct {
    bridgeFactory BridgeFactory
    bridges       map[string]interface{}
    mu            sync.RWMutex
}

func (lbm *LazyBridgeModule) LoadModule(L *lua.LState) int {
    module := L.NewTable()
    
    // Lazy getter for bridges
    L.SetField(module, "get", L.NewFunction(func(L *lua.LState) int {
        bridgeName := L.CheckString(1)
        
        lbm.mu.RLock()
        bridge, exists := lbm.bridges[bridgeName]
        lbm.mu.RUnlock()
        
        if !exists {
            // Create bridge on first access
            lbm.mu.Lock()
            bridge = lbm.bridgeFactory.Create(bridgeName)
            lbm.bridges[bridgeName] = bridge
            lbm.mu.Unlock()
        }
        
        // Wrap bridge for Lua
        L.Push(wrapBridgeForLua(L, bridge))
        return 1
    }))
    
    L.Push(module)
    return 1
}
```

### 2. Lazy Function Binding

```go
type LazyFunctionModule struct {
    functions map[string]FunctionDefinition
}

type FunctionDefinition struct {
    Name     string
    Impl     lua.LGFunction
    MinArgs  int
    MaxArgs  int
    LazyInit func() error
    initOnce sync.Once
}

func (lfm *LazyFunctionModule) CreateLazyFunction(def FunctionDefinition) lua.LGFunction {
    return func(L *lua.LState) int {
        // Initialize on first call
        var initErr error
        def.initOnce.Do(func() {
            if def.LazyInit != nil {
                initErr = def.LazyInit()
            }
        })
        
        if initErr != nil {
            L.RaiseError("failed to initialize %s: %v", def.Name, initErr)
            return 0
        }
        
        // Validate arguments
        argc := L.GetTop()
        if argc < def.MinArgs {
            L.RaiseError("%s: expected at least %d arguments, got %d", 
                def.Name, def.MinArgs, argc)
        }
        if def.MaxArgs > 0 && argc > def.MaxArgs {
            L.RaiseError("%s: expected at most %d arguments, got %d", 
                def.Name, def.MaxArgs, argc)
        }
        
        // Call actual implementation
        return def.Impl(L)
    }
}
```

### 3. Progressive Module Loading

```go
type ProgressiveLoader struct {
    stages []LoadStage
}

type LoadStage struct {
    Name     string
    Modules  []string
    Priority int
    Async    bool
}

func (pl *ProgressiveLoader) LoadModules(L *lua.LState) error {
    // Sort stages by priority
    sort.Slice(pl.stages, func(i, j int) bool {
        return pl.stages[i].Priority < pl.stages[j].Priority
    })
    
    for _, stage := range pl.stages {
        if stage.Async {
            // Load asynchronously
            go pl.loadStageAsync(L, stage)
        } else {
            // Load synchronously
            if err := pl.loadStage(L, stage); err != nil {
                return fmt.Errorf("failed to load stage %s: %w", stage.Name, err)
            }
        }
    }
    
    return nil
}

func (pl *ProgressiveLoader) loadStage(L *lua.LState, stage LoadStage) error {
    for _, modName := range stage.Modules {
        if err := pl.loadModule(L, modName); err != nil {
            return err
        }
    }
    return nil
}
```

## Module Caching and Reuse

### 1. Compiled Module Cache

```go
type CompiledModuleCache struct {
    cache map[string]*CompiledModule
    mu    sync.RWMutex
}

type CompiledModule struct {
    Name     string
    Proto    *lua.FunctionProto
    Metadata ModuleMetadata
    LoadTime time.Time
}

func (cmc *CompiledModuleCache) LoadOrCompile(L *lua.LState, name string, source string) error {
    cmc.mu.RLock()
    compiled, exists := cmc.cache[name]
    cmc.mu.RUnlock()
    
    if exists {
        // Use cached compiled module
        L.Push(L.NewFunctionFromProto(compiled.Proto))
        L.Call(0, 1)
        return nil
    }
    
    // Compile and cache
    proto, err := CompileLua(L, source, name)
    if err != nil {
        return err
    }
    
    cmc.mu.Lock()
    cmc.cache[name] = &CompiledModule{
        Name:     name,
        Proto:    proto,
        LoadTime: time.Now(),
    }
    cmc.mu.Unlock()
    
    L.Push(L.NewFunctionFromProto(proto))
    L.Call(0, 1)
    return nil
}
```

### 2. Module State Sharing

```go
type SharedModuleState struct {
    modules   map[string]*SharedModule
    mu        sync.RWMutex
}

type SharedModule struct {
    State     interface{} // Shared state
    RefCount  int32
    InitOnce  sync.Once
    CleanupFn func()
}

func (sms *SharedModuleState) GetOrCreateModule(name string, creator func() interface{}) interface{} {
    sms.mu.Lock()
    defer sms.mu.Unlock()
    
    mod, exists := sms.modules[name]
    if !exists {
        mod = &SharedModule{}
        sms.modules[name] = mod
    }
    
    mod.InitOnce.Do(func() {
        mod.State = creator()
    })
    
    atomic.AddInt32(&mod.RefCount, 1)
    return mod.State
}

func (sms *SharedModuleState) ReleaseModule(name string) {
    sms.mu.Lock()
    defer sms.mu.Unlock()
    
    if mod, exists := sms.modules[name]; exists {
        if atomic.AddInt32(&mod.RefCount, -1) == 0 {
            if mod.CleanupFn != nil {
                mod.CleanupFn()
            }
            delete(sms.modules, name)
        }
    }
}
```

## Performance Optimization

### 1. Module Bundling

```go
type ModuleBundle struct {
    Name    string
    Modules []string
    Loader  lua.LGFunction
}

func CreateModuleBundle(name string, modules []string) ModuleBundle {
    return ModuleBundle{
        Name:    name,
        Modules: modules,
        Loader: func(L *lua.LState) int {
            bundle := L.NewTable()
            
            // Load all modules in bundle
            for _, modName := range modules {
                L.DoString(fmt.Sprintf(`_BUNDLE["%s"] = require("%s")`, modName, modName))
            }
            
            // Return bundle table
            L.Push(bundle)
            return 1
        },
    }
}

// Usage
bundles := []ModuleBundle{
    CreateModuleBundle("ai", []string{"llm", "tools", "agents"}),
    CreateModuleBundle("data", []string{"state", "cache", "storage"}),
}
```

### 2. Ahead-of-Time Loading

```go
type AOTLoader struct {
    modules   []string
    loadGroup sync.WaitGroup
    errors    []error
    mu        sync.Mutex
}

func (aot *AOTLoader) PreloadInBackground(L *lua.LState) {
    for _, modName := range aot.modules {
        aot.loadGroup.Add(1)
        go func(name string) {
            defer aot.loadGroup.Done()
            
            // Create temporary state for loading
            tempL := lua.NewState()
            defer tempL.Close()
            
            err := tempL.DoString(fmt.Sprintf(`require("%s")`, name))
            if err != nil {
                aot.mu.Lock()
                aot.errors = append(aot.errors, fmt.Errorf("failed to preload %s: %w", name, err))
                aot.mu.Unlock()
            }
        }(modName)
    }
}

func (aot *AOTLoader) Wait() error {
    aot.loadGroup.Wait()
    if len(aot.errors) > 0 {
        return fmt.Errorf("preload errors: %v", aot.errors)
    }
    return nil
}
```

## Module Lifecycle Management

### 1. Module Initialization Hooks

```go
type ModuleLifecycle struct {
    preInit   []func(L *lua.LState) error
    postInit  []func(L *lua.LState) error
    preUnload []func(L *lua.LState) error
}

func (ml *ModuleLifecycle) LoadModuleWithLifecycle(L *lua.LState, name string, loader lua.LGFunction) error {
    // Pre-initialization
    for _, hook := range ml.preInit {
        if err := hook(L); err != nil {
            return fmt.Errorf("pre-init failed: %w", err)
        }
    }
    
    // Load module
    L.PreloadModule(name, loader)
    
    // Post-initialization
    for _, hook := range ml.postInit {
        if err := hook(L); err != nil {
            return fmt.Errorf("post-init failed: %w", err)
        }
    }
    
    // Register cleanup
    L.SetGlobal("__unload_"+name, L.NewFunction(func(L *lua.LState) int {
        for _, hook := range ml.preUnload {
            hook(L)
        }
        return 0
    }))
    
    return nil
}
```

### 2. Dependency Resolution

```go
type DependencyResolver struct {
    modules map[string]*ModuleDef
}

type ModuleDef struct {
    Name         string
    Dependencies []string
    Loader       lua.LGFunction
    Loaded       bool
}

func (dr *DependencyResolver) ResolveAndLoad(L *lua.LState, name string) error {
    visited := make(map[string]bool)
    return dr.loadWithDeps(L, name, visited)
}

func (dr *DependencyResolver) loadWithDeps(L *lua.LState, name string, visited map[string]bool) error {
    if visited[name] {
        return fmt.Errorf("circular dependency detected: %s", name)
    }
    visited[name] = true
    
    mod, exists := dr.modules[name]
    if !exists {
        return fmt.Errorf("module not found: %s", name)
    }
    
    if mod.Loaded {
        return nil // Already loaded
    }
    
    // Load dependencies first
    for _, dep := range mod.Dependencies {
        if err := dr.loadWithDeps(L, dep, visited); err != nil {
            return fmt.Errorf("failed to load dependency %s: %w", dep, err)
        }
    }
    
    // Load module
    L.PreloadModule(name, mod.Loader)
    mod.Loaded = true
    
    return nil
}
```

## Bridge Integration

### 1. Lazy Bridge Module

```go
func CreateLazyBridgeModule(bridges map[string]bridge.Bridge) lua.LGFunction {
    initialized := make(map[string]bool)
    var mu sync.RWMutex
    
    return func(L *lua.LState) int {
        module := L.NewTable()
        
        // Create lazy wrapper for each bridge
        for name, b := range bridges {
            name := name // Capture
            b := b       // Capture
            
            L.SetField(module, name, L.NewFunction(func(L *lua.LState) int {
                mu.RLock()
                isInit := initialized[name]
                mu.RUnlock()
                
                if !isInit {
                    mu.Lock()
                    if !initialized[name] {
                        // Initialize bridge on first use
                        if initializer, ok := b.(bridge.Initializable); ok {
                            if err := initializer.Initialize(); err != nil {
                                mu.Unlock()
                                L.RaiseError("failed to initialize bridge %s: %v", name, err)
                                return 0
                            }
                        }
                        initialized[name] = true
                    }
                    mu.Unlock()
                }
                
                // Return bridge wrapper
                L.Push(wrapBridge(L, b))
                return 1
            }))
        }
        
        L.Push(module)
        return 1
    }
}
```

### 2. Module Factory Pattern

```go
type ModuleFactory struct {
    factories map[string]func() lua.LGFunction
}

func (mf *ModuleFactory) Register(name string, factory func() lua.LGFunction) {
    mf.factories[name] = factory
}

func (mf *ModuleFactory) CreateLoader(config ModuleConfig) lua.LGFunction {
    return func(L *lua.LState) int {
        module := L.NewTable()
        
        // Only create requested modules
        for _, name := range config.EnabledModules {
            if factory, ok := mf.factories[name]; ok {
                loader := factory()
                L.SetField(module, name, L.NewFunction(loader))
            }
        }
        
        L.Push(module)
        return 1
    }
}
```

## Testing Strategies

### Module Loading Tests

```go
func TestLazyModuleLoading(t *testing.T) {
    L := lua.NewState()
    defer L.Close()
    
    loadCount := 0
    L.PreloadModule("test", func(L *lua.LState) int {
        loadCount++
        exports := L.NewTable()
        L.SetField(exports, "loaded", lua.LBool(true))
        L.Push(exports)
        return 1
    })
    
    // Module not loaded yet
    if loadCount != 0 {
        t.Error("module loaded prematurely")
    }
    
    // First require loads module
    L.DoString(`local m = require("test")`)
    if loadCount != 1 {
        t.Error("module not loaded on require")
    }
    
    // Second require uses cache
    L.DoString(`local m2 = require("test")`)
    if loadCount != 1 {
        t.Error("module loaded multiple times")
    }
}
```

## Best Practices

1. **Minimal Preloading**: Only preload essential modules
2. **Lazy Dependencies**: Load dependencies only when needed
3. **Module Bundling**: Group related modules for efficiency
4. **Cache Compiled Code**: Reuse compiled modules across states
5. **Profile-Based Loading**: Different module sets for different use cases

## Implementation Checklist

- [ ] Basic module preloading infrastructure
- [ ] Lazy loading mechanisms
- [ ] Dependency resolution system
- [ ] Module caching layer
- [ ] Lifecycle management hooks
- [ ] Profile-based loading
- [ ] Module bundling support
- [ ] Shared state management
- [ ] Performance monitoring
- [ ] Comprehensive test suite

## Summary

Effective module preloading and lazy initialization in GopherLua requires:
1. Strategic preloading of core modules
2. Lazy loading for optional functionality
3. Efficient caching and reuse mechanisms
4. Proper dependency management
5. Profile-based loading strategies

This approach minimizes startup time, reduces memory usage, and provides flexibility for different deployment scenarios.