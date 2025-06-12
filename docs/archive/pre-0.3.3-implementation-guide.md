# Implementation Guide

This guide provides a roadmap for implementing the go-llmspell architecture.

## Phase 1: Core Infrastructure

### 1.1 Script Engine Interface Implementation

Create the base interface and registry system:

```go
// pkg/engine/interface.go
package engine

import (
    "context"
    "io"
    "time"
)

type Engine interface {
    Name() string
    LoadScript(reader io.Reader) error
    LoadFile(path string) error
    Execute(ctx context.Context) (*ExecutionResult, error)
    RegisterBinding(name string, binding interface{}) error
    SetGlobal(name string, value interface{}) error
}

type ExecutionResult struct {
    Output   string
    Error    error
    Logs     []LogEntry
    Metrics  map[string]interface{}
    Duration time.Duration
}

type LogEntry struct {
    Level     string
    Message   string
    Timestamp time.Time
    Fields    map[string]interface{}
}
```

### 1.2 Engine Registry

```go
// pkg/engine/registry.go
package engine

import (
    "fmt"
    "sync"
)

type Registry struct {
    engines map[string]Factory
    mu      sync.RWMutex
}

type Factory func(config *Config) (Engine, error)

func NewRegistry() *Registry {
    return &Registry{
        engines: make(map[string]Factory),
    }
}

func (r *Registry) Register(name string, factory Factory) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.engines[name]; exists {
        return fmt.Errorf("engine %s already registered", name)
    }
    
    r.engines[name] = factory
    return nil
}

func (r *Registry) Create(name string, config *Config) (Engine, error) {
    r.mu.RLock()
    factory, exists := r.engines[name]
    r.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("engine %s not found", name)
    }
    
    return factory(config)
}
```

### 1.3 Bridge Infrastructure

Base bridge interface:

```go
// pkg/bridge/interface.go
package bridge

type Bridge interface {
    Name() string
    Register(engine engine.Engine) error
}

type Set struct {
    bridges map[string]Bridge
    mu      sync.RWMutex
}

func NewSet() *Set {
    return &Set{
        bridges: make(map[string]Bridge),
    }
}

func (s *Set) Add(bridge Bridge) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.bridges[bridge.Name()] = bridge
    return nil
}

func (s *Set) RegisterAll(engine engine.Engine) error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    for _, bridge := range s.bridges {
        if err := bridge.Register(engine); err != nil {
            return fmt.Errorf("failed to register bridge %s: %w", bridge.Name(), err)
        }
    }
    
    return nil
}
```

## Phase 2: Lua Engine Implementation

### 2.1 GopherLua Engine

```go
// pkg/engine/lua/engine.go
package lua

import (
    "context"
    "io"
    "io/ioutil"
    
    "github.com/yuin/gopher-lua"
    "github.com/lexlapax/go-llmspell/pkg/engine"
)

type LuaEngine struct {
    state   *lua.LState
    config  *engine.Config
    preload []func(*lua.LState) error
}

func New(config *engine.Config) (engine.Engine, error) {
    return &LuaEngine{
        config:  config,
        preload: make([]func(*lua.LState) error, 0),
    }, nil
}

func (e *LuaEngine) Name() string {
    return "lua"
}

func (e *LuaEngine) LoadScript(reader io.Reader) error {
    content, err := ioutil.ReadAll(reader)
    if err != nil {
        return err
    }
    
    e.state = lua.NewState()
    
    // Apply preload functions
    for _, fn := range e.preload {
        if err := fn(e.state); err != nil {
            return err
        }
    }
    
    return e.state.DoString(string(content))
}

func (e *LuaEngine) RegisterBinding(name string, binding interface{}) error {
    e.preload = append(e.preload, func(L *lua.LState) error {
        // Convert Go value to Lua value
        L.SetGlobal(name, toLuaValue(L, binding))
        return nil
    })
    return nil
}
```

### 2.2 Lua Bridge Adapters

```go
// pkg/bridge/lua/llm.go
package lua

import (
    lua "github.com/yuin/gopher-lua"
    "github.com/lexlapax/go-llmspell/pkg/bridge"
)

type LLMBridge struct {
    llm *bridge.LLMBridge
}

func (b *LLMBridge) RegisterLua(L *lua.LState) {
    llmMod := L.NewTable()
    
    L.SetField(llmMod, "chat", L.NewFunction(b.chat))
    L.SetField(llmMod, "complete", L.NewFunction(b.complete))
    L.SetField(llmMod, "stream", L.NewFunction(b.stream))
    
    L.SetGlobal("llm", llmMod)
}

func (b *LLMBridge) chat(L *lua.LState) int {
    prompt := L.CheckString(1)
    
    response, err := b.llm.Chat(context.Background(), prompt)
    if err != nil {
        L.Push(lua.LNil)
        L.Push(lua.LString(err.Error()))
        return 2
    }
    
    L.Push(lua.LString(response))
    return 1
}
```

## Phase 3: Tool System

### 3.1 Tool Interface

```go
// pkg/tools/interface.go
package tools

type Tool interface {
    Name() string
    Description() string
    Parameters() Schema
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

type Schema struct {
    Type       string                 `json:"type"`
    Properties map[string]Property    `json:"properties"`
    Required   []string              `json:"required"`
}

type Property struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Default     interface{} `json:"default,omitempty"`
}
```

### 3.2 Tool Registry

```go
// pkg/tools/registry.go
package tools

type Registry struct {
    tools map[string]Tool
    mu    sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        tools: make(map[string]Tool),
    }
}

func (r *Registry) Register(tool Tool) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.tools[tool.Name()] = tool
    return nil
}

func (r *Registry) Get(name string) (Tool, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    tool, exists := r.tools[name]
    if !exists {
        return nil, fmt.Errorf("tool %s not found", name)
    }
    
    return tool, nil
}
```

## Phase 4: Agent System

### 4.1 Agent Interface

```go
// pkg/agents/interface.go
package agents

type Agent interface {
    Name() string
    Run(ctx context.Context, input string) (string, error)
    AddTool(tool tools.Tool) error
    SetSystemPrompt(prompt string)
}

type Config struct {
    Name         string
    SystemPrompt string
    Tools        []string
    Provider     string
    Model        string
    MaxTokens    int
    Temperature  float64
}
```

### 4.2 Agent Implementation

```go
// pkg/agents/agent.go
package agents

import (
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llmspell/pkg/tools"
)

type BaseAgent struct {
    config   Config
    agent    *workflow.Agent
    toolReg  *tools.Registry
}

func New(config Config, toolReg *tools.Registry) (*BaseAgent, error) {
    // Create go-llms agent
    // Map tools from registry
    // Configure with system prompt
    return &BaseAgent{
        config:  config,
        toolReg: toolReg,
    }, nil
}
```

## Phase 5: Spell System

### 5.1 Spell Loader

```go
// pkg/spells/loader.go
package spells

type Loader struct {
    directories []string
    engines     *engine.Registry
}

type Spell struct {
    Name        string
    Path        string
    Engine      string
    Metadata    Metadata
}

type Metadata struct {
    Version     string
    Description string
    Author      string
    Parameters  []Parameter
}

func (l *Loader) Discover() ([]Spell, error) {
    var spells []Spell
    
    for _, dir := range l.directories {
        // Walk directory structure
        // Load spell.yaml metadata
        // Categorize by engine
    }
    
    return spells, nil
}
```

### 5.2 Spell Runner

```go
// pkg/spells/runner.go
package spells

type Runner struct {
    engines  *engine.Registry
    bridges  *bridge.Set
    security *security.Policy
}

func (r *Runner) Execute(spell Spell, params map[string]interface{}) (*engine.ExecutionResult, error) {
    // Create engine instance
    engine, err := r.engines.Create(spell.Engine, &engine.Config{
        Timeout:     30 * time.Second,
        MemoryLimit: 100 * 1024 * 1024, // 100MB
    })
    if err != nil {
        return nil, err
    }
    
    // Register bridges
    if err := r.bridges.RegisterAll(engine); err != nil {
        return nil, err
    }
    
    // Load spell
    if err := engine.LoadFile(spell.Path); err != nil {
        return nil, err
    }
    
    // Apply security policy
    ctx := r.security.Apply(context.Background())
    
    // Execute with parameters
    for k, v := range params {
        engine.SetGlobal(k, v)
    }
    
    return engine.Execute(ctx)
}
```

## Phase 6: Security Implementation

### 6.1 Security Policy

```go
// pkg/security/policy.go
package security

type Policy struct {
    FileSystem   FileSystemPolicy
    Network      NetworkPolicy
    Resources    ResourcePolicy
}

type FileSystemPolicy struct {
    Enabled      bool
    JailPath     string
    AllowedPaths []string
    DeniedPaths  []string
    MaxFileSize  int64
}

type NetworkPolicy struct {
    Enabled         bool
    AllowedDomains  []string
    DeniedDomains   []string
    MaxRequests     int
    RequestTimeout  time.Duration
}

type ResourcePolicy struct {
    MaxMemory     int64
    MaxCPUTime    time.Duration
    MaxGoroutines int
}
```

### 6.2 Sandbox Implementation

```go
// pkg/security/sandbox.go
package security

type Sandbox struct {
    policy Policy
}

func (s *Sandbox) ValidateFileAccess(path string) error {
    // Check jail path
    // Validate against allowed/denied lists
    // Prevent directory traversal
    return nil
}

func (s *Sandbox) ValidateNetworkAccess(domain string) error {
    // Check domain allowlist
    // Validate against denied list
    // Check request count
    return nil
}
```

## Phase 7: CLI Integration

### 7.1 Main Command

```go
// cmd/llmspell/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    
    "github.com/lexlapax/go-llmspell/pkg/spells"
)

func main() {
    var (
        spellPath = flag.String("spell", "", "Path to spell file")
        listSpells = flag.Bool("list", false, "List available spells")
        engine = flag.String("engine", "", "Force specific engine")
    )
    
    flag.Parse()
    
    runner := spells.NewRunner()
    
    if *listSpells {
        spells, _ := runner.List()
        for _, spell := range spells {
            fmt.Printf("%s (%s): %s\n", spell.Name, spell.Engine, spell.Description)
        }
        return
    }
    
    if *spellPath != "" {
        result, err := runner.ExecuteFile(*spellPath)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Println(result.Output)
    }
}
```

## Implementation Timeline

### Week 1-2: Core Infrastructure
- Script engine interface
- Registry system
- Basic bridge infrastructure
- LLM bridge completion

### Week 3-4: Lua Integration
- GopherLua engine implementation
- Lua bridge adapters
- Basic spell loading

### Week 5-6: Tool and Agent Systems
- Tool interface and registry
- Agent implementation
- Tool bridge for scripts

### Week 7-8: JavaScript and Tengo
- Goja engine implementation
- Tengo engine implementation
- Cross-engine testing

### Week 9-10: Security and Polish
- Security policy implementation
- Sandboxing
- Performance optimization
- Documentation

### Week 11-12: CLI and Examples
- CLI commands
- Example spells
- Integration testing
- Release preparation

## Testing Strategy

1. **Unit Tests**: Test each component in isolation
2. **Integration Tests**: Test bridge interactions
3. **Engine Tests**: Verify script execution across engines
4. **Security Tests**: Validate sandboxing and policies
5. **Performance Tests**: Benchmark script execution
6. **Example Tests**: Ensure all examples work correctly