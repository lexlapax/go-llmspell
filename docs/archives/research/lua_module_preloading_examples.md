# Lua Module Preloading Examples

This document provides practical examples of implementing module preloading and lazy initialization in GopherLua.

## Basic Module Preloading

### Simple Module Preloading

```go
package main

import (
    "fmt"
    lua "github.com/yuin/gopher-lua"
)

func main() {
    L := lua.NewState()
    defer L.Close()
    
    // Preload a simple module
    L.PreloadModule("hello", func(L *lua.LState) int {
        // Create module table
        mod := L.NewTable()
        
        // Add module functions
        L.SetField(mod, "greet", L.NewFunction(func(L *lua.LState) int {
            name := L.CheckString(1)
            L.Push(lua.LString(fmt.Sprintf("Hello, %s!", name)))
            return 1
        }))
        
        L.SetField(mod, "version", lua.LString("1.0.0"))
        
        // Return module
        L.Push(mod)
        return 1
    })
    
    // Use the module
    err := L.DoString(`
        local hello = require("hello")
        print(hello.greet("World"))
        print("Version:", hello.version)
    `)
    
    if err != nil {
        panic(err)
    }
}
```

### Bridge Module Preloading

```go
package bridgemodules

import (
    lua "github.com/yuin/gopher-lua"
    "github.com/your/go-llms/bridge"
)

func PreloadLLMBridge(L *lua.LState, llmBridge bridge.LLMBridge) {
    L.PreloadModule("llm", func(L *lua.LState) int {
        mod := L.NewTable()
        
        // Create agent function
        L.SetField(mod, "create_agent", L.NewFunction(func(L *lua.LState) int {
            config := L.CheckTable(1)
            
            // Convert Lua table to config
            agentConfig := bridge.AgentConfig{
                Provider: getStringField(L, config, "provider", "openai"),
                Model:    getStringField(L, config, "model", "gpt-4"),
            }
            
            // Create agent via bridge
            agent, err := llmBridge.CreateAgent(agentConfig)
            if err != nil {
                L.RaiseError("failed to create agent: %v", err)
                return 0
            }
            
            // Wrap agent for Lua
            L.Push(wrapAgent(L, agent))
            return 1
        }))
        
        // Generate function
        L.SetField(mod, "generate", L.NewFunction(func(L *lua.LState) int {
            agent := checkAgent(L, 1)
            prompt := L.CheckString(2)
            
            response, err := agent.Generate(prompt)
            if err != nil {
                L.RaiseError("generation failed: %v", err)
                return 0
            }
            
            L.Push(lua.LString(response))
            return 1
        }))
        
        L.Push(mod)
        return 1
    })
}
```

## Lazy Module Loading

### On-Demand Module System

```go
package lazymodules

import (
    "sync"
    lua "github.com/yuin/gopher-lua"
)

type LazyModuleSystem struct {
    modules    map[string]ModuleDefinition
    loaded     map[string]bool
    loadMutex  sync.Mutex
}

type ModuleDefinition struct {
    Name         string
    Dependencies []string
    InitFunc     func() error
    LoadFunc     lua.LGFunction
}

func NewLazyModuleSystem() *LazyModuleSystem {
    return &LazyModuleSystem{
        modules: make(map[string]ModuleDefinition),
        loaded:  make(map[string]bool),
    }
}

func (lms *LazyModuleSystem) RegisterModule(def ModuleDefinition) {
    lms.modules[def.Name] = def
}

func (lms *LazyModuleSystem) PreloadAll(L *lua.LState) {
    for name, def := range lms.modules {
        name := name // Capture
        def := def   // Capture
        
        L.PreloadModule(name, func(L *lua.LState) int {
            // Ensure module is loaded only once
            lms.loadMutex.Lock()
            defer lms.loadMutex.Unlock()
            
            if !lms.loaded[name] {
                // Load dependencies first
                for _, dep := range def.Dependencies {
                    if !lms.loaded[dep] {
                        L.DoString(fmt.Sprintf(`require("%s")`, dep))
                    }
                }
                
                // Initialize module if needed
                if def.InitFunc != nil {
                    if err := def.InitFunc(); err != nil {
                        L.RaiseError("module %s initialization failed: %v", name, err)
                        return 0
                    }
                }
                
                lms.loaded[name] = true
            }
            
            // Load module content
            return def.LoadFunc(L)
        })
    }
}

// Example usage
func SetupModules(lms *LazyModuleSystem) {
    // Core module (no dependencies)
    lms.RegisterModule(ModuleDefinition{
        Name: "core",
        LoadFunc: func(L *lua.LState) int {
            mod := L.NewTable()
            L.SetField(mod, "initialized", lua.LBool(true))
            L.Push(mod)
            return 1
        },
    })
    
    // LLM module (depends on core)
    lms.RegisterModule(ModuleDefinition{
        Name:         "llm",
        Dependencies: []string{"core"},
        InitFunc: func() error {
            // Initialize LLM connections
            fmt.Println("Initializing LLM module...")
            return nil
        },
        LoadFunc: LoadLLMModule,
    })
    
    // Tools module (depends on core and llm)
    lms.RegisterModule(ModuleDefinition{
        Name:         "tools",
        Dependencies: []string{"core", "llm"},
        LoadFunc:     LoadToolsModule,
    })
}
```

### Progressive Module Loader

```go
package progressive

import (
    "context"
    "log"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type ProgressiveLoader struct {
    stages   []LoadStage
    progress chan LoadProgress
}

type LoadStage struct {
    Name     string
    Priority int
    Timeout  time.Duration
    Modules  []ModuleSpec
}

type ModuleSpec struct {
    Name     string
    Required bool
    LoadFunc lua.LGFunction
}

type LoadProgress struct {
    Stage      string
    Module     string
    TotalSteps int
    Current    int
    Error      error
}

func (pl *ProgressiveLoader) LoadModules(ctx context.Context, L *lua.LState) error {
    totalModules := 0
    for _, stage := range pl.stages {
        totalModules += len(stage.Modules)
    }
    
    currentModule := 0
    
    for _, stage := range pl.stages {
        stageCtx, cancel := context.WithTimeout(ctx, stage.Timeout)
        
        for _, module := range stage.Modules {
            currentModule++
            
            // Report progress
            if pl.progress != nil {
                pl.progress <- LoadProgress{
                    Stage:      stage.Name,
                    Module:     module.Name,
                    TotalSteps: totalModules,
                    Current:    currentModule,
                }
            }
            
            // Check context
            select {
            case <-stageCtx.Done():
                cancel()
                return fmt.Errorf("stage %s timed out", stage.Name)
            default:
            }
            
            // Preload module
            L.PreloadModule(module.Name, module.LoadFunc)
            
            // Load required modules immediately
            if module.Required {
                err := L.DoString(fmt.Sprintf(`require("%s")`, module.Name))
                if err != nil {
                    cancel()
                    if pl.progress != nil {
                        pl.progress <- LoadProgress{
                            Stage:   stage.Name,
                            Module:  module.Name,
                            Error:   err,
                        }
                    }
                    return fmt.Errorf("failed to load required module %s: %w", 
                        module.Name, err)
                }
            }
        }
        
        cancel()
    }
    
    return nil
}

// Example usage with progress monitoring
func LoadWithProgress() {
    L := lua.NewState()
    defer L.Close()
    
    progress := make(chan LoadProgress, 10)
    loader := &ProgressiveLoader{
        stages: []LoadStage{
            {
                Name:     "critical",
                Priority: 1,
                Timeout:  5 * time.Second,
                Modules: []ModuleSpec{
                    {Name: "core", Required: true, LoadFunc: LoadCoreModule},
                    {Name: "error", Required: true, LoadFunc: LoadErrorModule},
                },
            },
            {
                Name:     "standard",
                Priority: 2,
                Timeout:  10 * time.Second,
                Modules: []ModuleSpec{
                    {Name: "llm", Required: false, LoadFunc: LoadLLMModule},
                    {Name: "tools", Required: false, LoadFunc: LoadToolsModule},
                },
            },
            {
                Name:     "optional",
                Priority: 3,
                Timeout:  15 * time.Second,
                Modules: []ModuleSpec{
                    {Name: "workflow", Required: false, LoadFunc: LoadWorkflowModule},
                    {Name: "events", Required: false, LoadFunc: LoadEventsModule},
                },
            },
        },
        progress: progress,
    }
    
    // Monitor progress in background
    go func() {
        for p := range progress {
            if p.Error != nil {
                log.Printf("Error loading %s: %v", p.Module, p.Error)
            } else {
                log.Printf("Loading %s (%d/%d) - %s", 
                    p.Module, p.Current, p.TotalSteps, p.Stage)
            }
        }
    }()
    
    ctx := context.Background()
    if err := loader.LoadModules(ctx, L); err != nil {
        log.Fatal(err)
    }
    
    close(progress)
}
```

## Module Bundling

### Bundle Multiple Modules

```go
package bundling

import (
    lua "github.com/yuin/gopher-lua"
)

type ModuleBundle struct {
    Name        string
    Description string
    Modules     map[string]lua.LGFunction
}

func CreateAIBundle() ModuleBundle {
    return ModuleBundle{
        Name:        "ai",
        Description: "AI and LLM functionality",
        Modules: map[string]lua.LGFunction{
            "llm":    LoadLLMModule,
            "tools":  LoadToolsModule,
            "agents": LoadAgentsModule,
        },
    }
}

func (mb ModuleBundle) Preload(L *lua.LState) {
    L.PreloadModule(mb.Name, func(L *lua.LState) int {
        bundle := L.NewTable()
        
        // Add metadata
        meta := L.NewTable()
        L.SetField(meta, "name", lua.LString(mb.Name))
        L.SetField(meta, "description", lua.LString(mb.Description))
        L.SetField(bundle, "_meta", meta)
        
        // Lazy loading for bundle modules
        for modName, loadFunc := range mb.Modules {
            modName := modName
            loadFunc := loadFunc
            
            // Create lazy getter
            L.SetField(bundle, modName, L.NewFunction(func(L *lua.LState) int {
                // Check if already loaded
                L.GetField(bundle, "_loaded_"+modName)
                if !L.IsNil(-1) {
                    return 1 // Return cached module
                }
                L.Pop(1)
                
                // Load module
                loadFunc(L)
                module := L.Get(-1)
                
                // Cache it
                L.SetField(bundle, "_loaded_"+modName, module)
                
                return 1
            }))
        }
        
        L.Push(bundle)
        return 1
    })
}

// Usage example
func UseBundles() {
    L := lua.NewState()
    defer L.Close()
    
    // Create and preload bundles
    bundles := []ModuleBundle{
        CreateAIBundle(),
        CreateDataBundle(),
        CreateUtilsBundle(),
    }
    
    for _, bundle := range bundles {
        bundle.Preload(L)
    }
    
    // Use bundled modules
    L.DoString(`
        local ai = require("ai")
        
        -- Modules are loaded on first access
        local llm = ai.llm()
        local tools = ai.tools()
        
        -- Subsequent access uses cached version
        local llm2 = ai.llm() -- Same instance
    `)
}
```

## Conditional Loading

### Profile-Based Loading

```go
package profiles

import (
    lua "github.com/yuin/gopher-lua"
)

type LoadProfile struct {
    Name        string
    Modules     []string
    Features    map[string]bool
    MemoryLimit int
}

var Profiles = map[string]LoadProfile{
    "minimal": {
        Name:        "minimal",
        Modules:     []string{"core"},
        Features:    map[string]bool{"sandbox": true},
        MemoryLimit: 10 * 1024 * 1024, // 10MB
    },
    "standard": {
        Name:        "standard", 
        Modules:     []string{"core", "llm", "tools"},
        Features:    map[string]bool{"sandbox": true, "async": true},
        MemoryLimit: 50 * 1024 * 1024, // 50MB
    },
    "full": {
        Name:        "full",
        Modules:     []string{"core", "llm", "tools", "workflow", "events", "state"},
        Features:    map[string]bool{"sandbox": false, "async": true, "debug": true},
        MemoryLimit: 200 * 1024 * 1024, // 200MB
    },
}

type ProfileLoader struct {
    availableModules map[string]lua.LGFunction
}

func (pl *ProfileLoader) LoadProfile(L *lua.LState, profileName string) error {
    profile, exists := Profiles[profileName]
    if !exists {
        return fmt.Errorf("unknown profile: %s", profileName)
    }
    
    // Set global profile info
    profileTable := L.NewTable()
    L.SetField(profileTable, "name", lua.LString(profile.Name))
    L.SetField(profileTable, "memory_limit", lua.LNumber(profile.MemoryLimit))
    
    featuresTable := L.NewTable()
    for feature, enabled := range profile.Features {
        L.SetField(featuresTable, feature, lua.LBool(enabled))
    }
    L.SetField(profileTable, "features", featuresTable)
    L.SetGlobal("_PROFILE", profileTable)
    
    // Load modules for profile
    for _, modName := range profile.Modules {
        if loadFunc, ok := pl.availableModules[modName]; ok {
            L.PreloadModule(modName, loadFunc)
        }
    }
    
    // Load profile-specific initialization
    return L.DoString(fmt.Sprintf(`
        -- Profile: %s
        _G.profile = _PROFILE
        
        -- Conditional features
        if profile.features.sandbox then
            -- Enable sandboxing
            _G.io = nil
            _G.os = nil
            _G.debug = nil
        end
        
        if profile.features.debug then
            -- Enable debug helpers
            _G.trace = function(msg) print("[TRACE]", msg) end
        else
            _G.trace = function() end
        end
    `, profile.Name))
}
```

## Caching and Optimization

### Module Cache Implementation

```go
package caching

import (
    "crypto/sha256"
    "encoding/hex"
    "sync"
    "time"
    lua "github.com/yuin/gopher-lua"
)

type ModuleCache struct {
    compiled map[string]*CachedModule
    mu       sync.RWMutex
    maxAge   time.Duration
}

type CachedModule struct {
    Hash         string
    Proto        *lua.FunctionProto
    Dependencies []string
    LoadedAt     time.Time
    AccessCount  int64
}

func NewModuleCache(maxAge time.Duration) *ModuleCache {
    mc := &ModuleCache{
        compiled: make(map[string]*CachedModule),
        maxAge:   maxAge,
    }
    
    // Start cleanup goroutine
    go mc.cleanupLoop()
    
    return mc
}

func (mc *ModuleCache) LoadModule(L *lua.LState, name, source string) error {
    hash := mc.hashSource(source)
    
    mc.mu.RLock()
    cached, exists := mc.compiled[name]
    mc.mu.RUnlock()
    
    if exists && cached.Hash == hash {
        // Use cached version
        cached.AccessCount++
        fn := L.NewFunctionFromProto(cached.Proto)
        L.Push(fn)
        L.Call(0, 1)
        return nil
    }
    
    // Compile new version
    chunk, err := parse.Parse(strings.NewReader(source), name)
    if err != nil {
        return err
    }
    
    proto, err := Compile(chunk, name)
    if err != nil {
        return err
    }
    
    // Cache compiled module
    mc.mu.Lock()
    mc.compiled[name] = &CachedModule{
        Hash:     hash,
        Proto:    proto,
        LoadedAt: time.Now(),
    }
    mc.mu.Unlock()
    
    // Execute
    fn := L.NewFunctionFromProto(proto)
    L.Push(fn)
    L.Call(0, 1)
    
    return nil
}

func (mc *ModuleCache) hashSource(source string) string {
    h := sha256.New()
    h.Write([]byte(source))
    return hex.EncodeToString(h.Sum(nil))
}

func (mc *ModuleCache) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        mc.mu.Lock()
        now := time.Now()
        for name, cached := range mc.compiled {
            if now.Sub(cached.LoadedAt) > mc.maxAge && cached.AccessCount == 0 {
                delete(mc.compiled, name)
            }
        }
        mc.mu.Unlock()
    }
}
```

## Complete Example: Modular Script Engine

### Full Implementation

```go
package engine

import (
    "context"
    "fmt"
    "path/filepath"
    "sync"
    lua "github.com/yuin/gopher-lua"
)

type ModularScriptEngine struct {
    pool          *LStatePool
    moduleSystem  *LazyModuleSystem
    cache         *ModuleCache
    profiles      map[string]LoadProfile
    currentProfile string
}

func NewModularScriptEngine(config EngineConfig) *ModularScriptEngine {
    engine := &ModularScriptEngine{
        pool:         NewLStatePool(config.PoolSize),
        moduleSystem: NewLazyModuleSystem(),
        cache:        NewModuleCache(config.CacheMaxAge),
        profiles:     make(map[string]LoadProfile),
    }
    
    // Register standard modules
    engine.RegisterStandardModules()
    
    // Setup profiles
    engine.SetupProfiles()
    
    return engine
}

func (e *ModularScriptEngine) RegisterStandardModules() {
    // Core modules
    e.moduleSystem.RegisterModule(ModuleDefinition{
        Name:     "spell",
        LoadFunc: e.loadSpellModule,
    })
    
    // Bridge modules
    bridges := []string{"llm", "tools", "workflow", "state", "events"}
    for _, name := range bridges {
        name := name
        e.moduleSystem.RegisterModule(ModuleDefinition{
            Name:         name,
            Dependencies: []string{"spell"},
            LoadFunc:     e.createBridgeLoader(name),
        })
    }
    
    // Utility modules
    e.moduleSystem.RegisterModule(ModuleDefinition{
        Name:     "async",
        LoadFunc: e.loadAsyncModule,
    })
    
    e.moduleSystem.RegisterModule(ModuleDefinition{
        Name:         "http",
        Dependencies: []string{"async"},
        LoadFunc:     e.loadHTTPModule,
    })
}

func (e *ModularScriptEngine) loadSpellModule(L *lua.LState) int {
    spell := L.NewTable()
    
    // Version info
    L.SetField(spell, "version", lua.LString("1.0.0"))
    
    // Module loader helper
    L.SetField(spell, "load", L.NewFunction(func(L *lua.LState) int {
        modules := L.CheckTable(1)
        
        loaded := L.NewTable()
        modules.ForEach(func(k, v lua.LValue) {
            if modName, ok := v.(lua.LString); ok {
                L.DoString(fmt.Sprintf(`_LOADED[%q] = require(%q)`, modName, modName))
                L.GetGlobal("_LOADED")
                L.GetField(-1, string(modName))
                L.SetField(loaded, string(modName), L.Get(-1))
                L.Pop(2)
            }
        })
        
        L.Push(loaded)
        return 1
    }))
    
    L.Push(spell)
    return 1
}

func (e *ModularScriptEngine) createBridgeLoader(bridgeName string) lua.LGFunction {
    return func(L *lua.LState) int {
        // Get bridge instance
        bridge := e.getBridge(bridgeName)
        if bridge == nil {
            L.RaiseError("bridge %s not available", bridgeName)
            return 0
        }
        
        // Create module table
        module := L.NewTable()
        
        // Wrap bridge methods
        e.wrapBridgeForLua(L, module, bridge)
        
        L.Push(module)
        return 1
    }
}

func (e *ModularScriptEngine) ExecuteFile(ctx context.Context, filename string, profile string) error {
    // Get or create state
    L := e.pool.Get()
    defer e.pool.Put(L)
    
    // Load profile
    if err := e.LoadProfile(L, profile); err != nil {
        return fmt.Errorf("failed to load profile %s: %w", profile, err)
    }
    
    // Setup context
    L.SetContext(ctx)
    
    // Check cache
    source, err := e.readFile(filename)
    if err != nil {
        return err
    }
    
    // Try cached version first
    if err := e.cache.LoadModule(L, filename, source); err != nil {
        // Fall back to standard loading
        if err := L.DoFile(filename); err != nil {
            return err
        }
    }
    
    return nil
}

func (e *ModularScriptEngine) LoadProfile(L *lua.LState, profileName string) error {
    profile, exists := e.profiles[profileName]
    if !exists {
        profile = e.profiles["default"]
    }
    
    // Preload modules for profile
    e.moduleSystem.PreloadForProfile(L, profile)
    
    // Set profile globals
    L.SetGlobal("PROFILE", lua.LString(profileName))
    
    e.currentProfile = profileName
    return nil
}

// Example usage
func ExampleUsage() {
    engine := NewModularScriptEngine(EngineConfig{
        PoolSize:    10,
        CacheMaxAge: 1 * time.Hour,
    })
    
    ctx := context.Background()
    
    // Execute with minimal profile
    err := engine.ExecuteFile(ctx, "scripts/simple.lua", "minimal")
    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    // Execute with full profile
    err = engine.ExecuteFile(ctx, "scripts/complex.lua", "full")
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
```

## Summary

These examples demonstrate comprehensive module preloading and lazy initialization patterns:
1. Basic preloading for simple modules
2. Lazy loading with dependency management
3. Progressive loading with timeout control
4. Module bundling for logical grouping
5. Profile-based conditional loading
6. Caching and optimization strategies
7. Complete integration in a modular engine

Key benefits include reduced startup time, lower memory usage, and flexible deployment configurations.