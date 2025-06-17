# TODO: Go-LLMSpell Bridge-First Implementation

## Overview
Based on the bridge-first architecture in `docs/MIGRATION_PLAN_V0.3.3.md`, this TODO focuses on **bridging existing go-llms functionality** rather than reimplementing features. Our value is making go-llms scriptable through Lua, JavaScript, and Tengo.

## Key Principles
1. **Fundamental Rule**: If it's not in go-llms, we don't implement it in go-llmspell
2. **Bridge, Don't Build**: We ONLY bridge existing go-llms functionality. Bridging also means imports from go-llms and implementing the bridge function calls in the bridge.
3. **Clean Architecture**: Just `pkg/engine/` and `pkg/bridge/` - no business logic
4. **Script Infrastructure Only**: We only build what's needed for scripting (engines, type conversion, sandboxing)
5. **Type Safety**: Maintain type conversions at bridge boundaries

## Migration Status
- ✅ Updated go-llms to v0.3.5
- ✅ Phase 1: Engine and Bridge Foundation [COMPLETED - 2025-06-17]
  - 38+ bridges across 13 categories with comprehensive test coverage
  - Pure bridge architecture: zero business logic duplication
- 🚧 Phase 2: Lua Engine Implementation - NOT STARTED
- 🚧 Phase 3: JavaScript Engine Implementation - NOT STARTED
- 🚧 Phase 4: Tengo Engine Implementation - NOT STARTED
- 🚧 Phase 5: Integration and Examples - NOT STARTED

---
## DEFERRED TASKS from different Phases - For Revisit 
- See `TODO-DONE-ARCHIVE.md` for completed tasks history

### Section 1.3.
  - [ ] **Task 1.3.20: Support for async/promise-based tool execution** (**[DEFERRED]** to script engine implementation)
  - [ ] **Task 1.3.21: Test cross-engine compatibility** (**[DEFERRED]** to script engine implementation)

#### ⏸️ 1.4.6 Model Info Bridge Intelligence **[DEFERRED]** - Features not in go-llms
**Status**: Tasks deferred - missing features documented in `go-llms-upstream-request.md`

- [ ] **Task 1.4.6.1: Add Model Performance Analytics** ⏸️ **[DEFERRED]**
  - Missing from go-llms: Model performance tracking, analytics, metrics
  - Documented in upstream request #1

- [ ] **Task 1.4.6.2: Add Model Recommendation Engine** ⏸️ **[DEFERRED]**  
  - Missing from go-llms: Recommendation algorithms, model selection
  - Documented in upstream request #2

- [ ] **Task 1.4.6.3: Add Model Catalog Export** ⏸️ **[DEFERRED]**
  - Missing from go-llms: Catalog export, OpenAPI generation for models
  - Documented in upstream request #3
- [ ] **Task 1.5.8: Memory Bridge** ⏸️ **[DEFERRED]** - Not in go-llms yet
  - [ ] Will implement when available in go-llms

### Section 1.5
- [ ] **Task 1.5.9: Conversation Bridge** ⏸️ **[DEFERRED]** - Not in go-llms yet
  - [ ] Will implement when available in go-llms


---

## Phase 2: Lua Engine Implementation
### 2.1 Lua Engine Research and Planning
- [x] 2.1.1. Research gopher-lua integration with go and add additional TODO.md entries as needed 
- [x] 2.1.2. Analyze LState management and pooling strategies
- [x] 2.1.3. Design ScriptValue ↔ LValue type conversion system 
- [x] 2.1.4. Plan goroutine and channel integration
- [x] 2.1.5. Design security sandboxing approach
- [x] 2.1.6. Research compiled chunk caching for performance
- [x] 2.1.7. Investigate instruction count limits and timeout mechanisms
- [x] 2.1.8. Study memory limits via registry configuration
- [x] 2.1.9. Research module preloading and lazy initialization
- [x] 2.1.10. Design error handling and stack trace preservation
- [x] 2.1.11. Plan LState lifecycle management (creation, pooling, cleanup)
- [x] 2.1.12. Research UserData vs Table for bridge object representation
- [x] 2.1.13. Investigate coroutine support for async bridge operations
- [ ] 2.1.14. Combine all research documents `docs/technical/lua*.md` and re-synthesize into one gopherlua_engine_architecture_design.md  based on `docs/technical/architecture.md`, requires megathink

### 2.2 Lua Engine Core
- [ ] **Task 2.2.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/gopherlua/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test GopherLua LState creation and management
  - [ ] Test resource limits (instruction count, memory, stack depth)
  - [ ] Test script execution timeout mechanisms
  - [ ] Test state cleanup and isolation between executions
  - [ ] Create `/pkg/engine/gopherlua/engine.go`
  - [ ] Implement ScriptEngine interface for Lua
  - [ ] Integrate GopherLua with LState management
  - [ ] Implement resource limits and timeouts
  - [ ] Add compiled chunk caching
  - [ ] Implement state pooling with sync.Pool

- [ ] **Task 2.2.2: Type Converter**
  - [ ] Create test file `/pkg/engine/gopherlua/converter_test.go`
  - [ ] Test ScriptValue ↔ LValue conversions
  - [ ] Test Lua tables ↔ Go maps/structs
  - [ ] Test LUserData for complex types
  - [ ] Test nil/error handling
  - [ ] Test circular reference detection
  - [ ] Create `/pkg/engine/gopherlua/converter.go`
  - [ ] Implement bidirectional type converter
  - [ ] Handle all LValue types (nil, bool, number, string, function, table, userdata)
  - [ ] Implement efficient table traversal
  - [ ] Add type validation and error reporting
  - [ ] Optimize for minimal allocations

- [ ] **Task 2.2.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/gopherlua/sandbox_test.go`
  - [ ] Test dangerous libraries disabled (io, os, debug)
  - [ ] Test instruction count limits
  - [ ] Test memory limits via registry size
  - [ ] Test function whitelist/blacklist
  - [ ] Test module loading restrictions
  - [ ] Create `/pkg/engine/gopherlua/sandbox.go`
  - [ ] Remove dangerous standard libraries
  - [ ] Implement instruction counting
  - [ ] Configure memory limits
  - [ ] Create safe function whitelist
  - [ ] Control module access

### 2.3 Lua Bridge Integration
- [ ] **Task 2.3.1: Bridge Module System**
  - [ ] Create `/pkg/engine/gopherlua/modules.go` - Module registration system
  - [ ] Implement bridge-to-module adapter
  - [ ] Add lazy loading support
  - [ ] Create module documentation generator

- [ ] **Task 2.3.2: Core Bridge Modules**
  - [ ] Create `/pkg/engine/gopherlua/stdlib/core.lua` - Core utilities and helpers
  - [ ] Create `/pkg/engine/gopherlua/stdlib/llm.lua` - LLM bridge wrapper
  - [ ] Create `/pkg/engine/gopherlua/stdlib/tools.lua` - Tools bridge wrapper
  - [ ] Create `/pkg/engine/gopherlua/stdlib/workflow.lua` - Workflow bridge wrapper
  - [ ] Create `/pkg/engine/gopherlua/stdlib/state.lua` - State bridge wrapper
  - [ ] Create `/pkg/engine/gopherlua/stdlib/events.lua` - Events bridge wrapper
  - [ ] Create `/pkg/engine/gopherlua/stdlib/hooks.lua` - Hooks bridge wrapper
  - [ ] Create `/pkg/engine/gopherlua/stdlib/logging.lua` - Logging bridge wrapper

- [ ] **Task 2.3.3: Lua Idioms and Patterns**
  - [ ] Design Lua-idiomatic APIs for bridges
  - [ ] Create promise/async patterns for Lua
  - [ ] Implement error handling conventions
  - [ ] Document Lua best practices

### 2.4 Lua Engine Advanced Features
- [ ] **Task 2.4.1: Performance Optimization**
  - [ ] Implement LState pooling
  - [ ] Add compiled script caching
  - [ ] Optimize hot paths in type conversion
  - [ ] Benchmark against performance targets

- [ ] **Task 2.4.2: Development Tools**
  - [ ] Add Lua script debugger support
  - [ ] Implement script profiling
  - [ ] Create REPL for testing
  - [ ] Add script validation tools

---

## Phase 3: JavaScript Engine Implementation

### 3.1 JavaScript Engine Research and Planning
- [ ] 3.1.1. Research goja (https://github.com/dop251/goja) go. Find the best javascript engine to work with in go-llmspell (There are others). 
- [ ] 3.1.2. Research how to integrate the chosen javascript engine into this go-llmspell library. add additional TODO.md entries as needed 
- [ ] 3.1.3. Analyze state management and memory integration
- [ ] 3.1.4. Design ScriptValue ↔ javascript type conversion system 
- [ ] 3.1.5. Plan goroutine integration for async operations
- [ ] 3.1.6. Design security sandboxing approach
- [ ] 3.1.7. Create detailed implementation roadmap
- [ ] 3.1.8. Research  bytecode validation and security implications - may not apply to gopher-lua
- [ ] 3.1.9. Investigate warning system integration 
- [ ] 3.1.10. Study generational GC vs incremental GC trade-offs if it applies
- [ ] 3.1.11. Research goja debug introspection capabilities for development tools
- [ ] 3.1.12. Combine all research documents and re-synthesize into one javascript_engine_architecture.md based on `docs/technical/architecture.md` and a detailed implementation roadmap

### 3.2 JavaScript Engine Core
- [ ] **Task 3.2.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/javascript/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test Goja integration
  - [ ] Test ES6+ or ES5.1+ whichever is the lstest support
  - [ ] Create `/pkg/engine/javascript/engine.go`
  - [ ] Implement ScriptEngine interface for JS
  - [ ] Integrate Goja
  - [ ] Add ES6+ support

- [ ] **Task 3.2.2: Type Converter**
  - [ ] Create test file `/pkg/engine/javascript/converter_test.go`
  - [ ] Test JS ↔ Go type conversions
  - [ ] Test Promise handling
  - [ ] Create `/pkg/engine/javascript/converter.go`
  - [ ] Implement type conversions
  - [ ] Handle async patterns

- [ ] **Task 3.2.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/javascript/sandbox_test.go`
  - [ ] Test global access restrictions
  - [ ] Test resource limits
  - [ ] Create `/pkg/engine/javascript/sandbox.go`
  - [ ] Restrict global access
  - [ ] Implement CSP-like policies

### 3.3 JavaScript Standard Library
- [ ] **Task 3.3.1: Core Modules**
  - [ ] Create `/pkg/engine/javascript/stdlib/core.js`
  - [ ] Create `/pkg/engine/javascript/stdlib/llm.js` - LLM bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/tools.js` - Tools bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/workflow.js` - Workflow bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/state.js` - State bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/events.js` - Events bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/hooks.js` - Hooks bridge wrapper

---

## Phase 4: Tengo Engine Implementation

### 4.1 Tengo Engine Core
- [ ] **Task 4.1.1: Engine Implementation**
  - [ ] Create `/pkg/engine/tengo/engine.go`
  - [ ] Implement ScriptEngine interface for Tengo
  - [ ] Integrate Tengo VM
  - [ ] Optimize for performance

- [ ] **Task 4.1.2: Type Converter**
  - [ ] Create `/pkg/engine/tengo/converter.go`
  - [ ] Implement Tengo ↔ Go conversions
  - [ ] Handle Tengo objects

- [ ] **Task 4.1.3: Security Sandbox**
  - [ ] Create `/pkg/engine/tengo/sandbox.go`
  - [ ] Implement Tengo restrictions
  - [ ] Add import controls

---

## Phase 5: Integration and Examples

### 5.1 Example Spells
- [ ] **Task 5.1.1: Basic Examples**
  - [ ] Hello World spell (all engines)
  - [ ] LLM chat spell
  - [ ] Tool usage spell
  - [ ] State management spell

- [ ] **Task 5.1.2: Advanced Examples**
  - [ ] Multi-agent orchestration spell
  - [ ] Complex workflow spell
  - [ ] Event-driven spell
  - [ ] Hook-based customization spell

### 5.2 Testing
- [ ] **Task 5.2.1: Cross-Engine Tests**
  - [ ] Create conformance test suite
  - [ ] Verify API compatibility
  - [ ] Test performance characteristics

- [ ] **Task 5.2.2: Integration Tests**
  - [ ] Test bridge functionality
  - [ ] Test type conversions
  - [ ] Test error handling

---

## Documentation

### API Documentation
- [ ] Bridge API reference
- [ ] Engine-specific features
- [ ] Type conversion guide

### User Guides
- [ ] Getting started guide
- [ ] Migration from direct go-llms usage
- [ ] Best practices

### Tutorials
- [ ] First spell tutorial
- [ ] Using go-llms agents from scripts
- [ ] Building workflows in scripts

---

## Success Metrics

### Development
- [ ] Zero duplicate implementations of go-llms features
- [ ] Clean two-package architecture maintained
- [ ] All bridges properly tested

### Performance
- [ ] < 5% overhead from bridging
- [ ] Type conversions optimized
- [ ] Memory usage minimal

### Adoption
- [ ] Clear examples for all major features
- [ ] Comprehensive documentation
- [ ] Easy migration path

---

## Notes

### Development Order
1. Complete core bridges (llm_agent, workflow, events, tools)
2. Implement provider and pool bridges
3. Complete Lua engine and stdlib
4. Add JavaScript engine
5. Add Tengo engine
6. Create comprehensive examples

### Testing Strategy
- TDD for all new code
- Test bridges thoroughly
- Cross-engine conformance tests
- Performance benchmarks

### What We DON'T Build (CRITICAL)
- ❌ **NO LLM Logic**: No provider implementations, no API calls, no response parsing
- ❌ **NO Agent Logic**: No agent orchestration, no tool execution logic
- ❌ **NO State Management**: No state storage, transforms, or merging logic
- ❌ **NO Workflow Engine**: No workflow execution or state passing
- ❌ **NO Event System**: No event dispatching or subscription logic
- ❌ **NO Tools Implementation**: No tool logic, only bridging to go-llms tools
- ❌ **NO Business Features**: If it should be in go-llms, contribute it there first
- ❌ **NO Custom Abstractions**: No "improved" versions of go-llms features

### What We DO Build (Our ONLY Value-Add)
- ✅ **Script Engines**: Lua, JavaScript, Tengo execution environments
- ✅ **Type Converters**: Script ↔ Go type conversion infrastructure
- ✅ **Bridge Interfaces**: Thin wrappers that expose go-llms to scripts
- ✅ **Security Sandboxes**: Script execution isolation and resource limits
- ✅ **Language Bindings**: Idiomatic script APIs for each language
- ✅ **Examples/Documentation**: How to use go-llms from scripts

### If You're Tempted to Implement Something...
1. **STOP**: Does it exist in go-llms? → Bridge it
2. **STOP**: Should it exist in go-llms? → Contribute upstream first
3. **STOP**: Is it script-specific? → Only then implement it here

---

**Remember**: If it exists in go-llms, we bridge it. We only build what's unique to our scripting layer.