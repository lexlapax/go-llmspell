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
- ✅ Phase 2: Lua Engine Implementation [COMPLETED - 2025-06-20]
  - ✅ Phase 2.1: Research and Planning [COMPLETED - 2025-06-17]
  - ✅ Phase 2.2: Core Engine Components [COMPLETED - 2025-06-18]
  - ✅ Phase 2.3: Bridge Integration Layer [COMPLETED - 2025-06-20]
    - ✅ 2.3.1: Module System Architecture [COMPLETED - 2025-06-19]
    - ✅ 2.3.2: Async/Coroutine Support [COMPLETED - 2025-06-19]
    - ✅ 2.3.2.0: ScriptValue Type System Refactoring [COMPLETED - 2025-06-19]
    - ✅ 2.3.2.0.X: Fix ScriptValue Bridge Test Failures [COMPLETED - 2025-06-19]
    - ✅ 2.3.2.5: Test Utilities Extraction [COMPLETED - 2025-06-19]
    - ✅ 2.3.3: Bridge Adapters [COMPLETED - 2025-06-19]
    - ✅ 2.3.4: Async/Coroutine Support [COMPLETED - 2025-06-19]
    - ✅ 2.3.5: Lua Standard Library [COMPLETED - 2025-06-20]
  - ✅ Phase 2.4: Advanced Features & Optimization (Partial) [COMPLETED - 2025-06-20]
    - ✅ 2.4.3.1: Debugger Support (100% coverage)
    - ✅ 2.4.3.2: Script Validator (100% coverage)
- ✅ Phase 3: Spell Runner CLI [COMPLETED - 2025-06-21]
  - ✅ All core CLI functionality implemented with comprehensive testing and documentation
  - ✅ See TODO-DONE.md for complete implementation details
- 🚧 Phase 4: JavaScript Engine Implementation - NOT STARTED
- 🚧 Phase 5: Tengo Engine Implementation - NOT STARTED
- 🚧 Phase 6: Integration and Examples - NOT STARTED

---

## Phase 2: Lua Engine Implementation
### 2.1 Lua Engine Research and Planning
✅ **COMPLETED [2025-06-17]** - All 14 research tasks completed. See TODO-DONE.md for details.

### Phase 2.2: Core Engine Components
✅ **COMPLETED [2025-06-18]** - All components implemented. See TODO-DONE.md for details.

### Phase 2.3: Bridge Integration Layer

#### 2.3.1: Module System Architecture 
✅ **COMPLETED [2025-06-18]** - See TODO-DONE.md for details.

#### 2.3.2: Async/Coroutine Support
✅ **COMPLETED [2025-06-18]** - All async/coroutine tasks completed. See TODO-DONE.md for details.

#### 2.3.2.5: Test Utilities Extraction
✅ **COMPLETED [2025-06-18]** - See TODO-DONE.md for complete details

#### 2.3.3: Bridge Adapters
✅ **COMPLETED [2025-06-19]** - All 24 tasks completed (Tasks 1-14 already in TODO-DONE.md)
**See TODO-DONE.md for complete task details and implementation history**

#### 2.3.4: Async/Coroutine Support
✅ **COMPLETED [2025-06-19]** - All 4 tasks completed. See TODO-DONE.md for implementation details.

#### 2.3.5: Lua Standard Library
✅ **PHASE COMPLETED [2025-06-20]** - All 18 tasks complete. See TODO-DONE.md for details.

### Phase 2.4: Advanced Features & Optimization
#### 2.4.1: Performance Optimization
#### 2.4.3: Development Tools

#### 2.4.4: Production Readiness
- [x] **Task 2.4.4.1: Comprehensive Testing** **[COMPLETED - 2025-06-23]**
- [x] **Task 2.4.4.2: Error Handling Enhancement** **[COMPLETED - 2025-06-23]**
- [x] **Task 2.4.4.3: Stdlib Module Loading Fix** (Option 1: Embed and Preload) **[COMPLETED - 2025-06-23]**
  - [ ] Documentation updates
    - [ ] Document embedded module system
    - [ ] Update troubleshooting guide
    - [ ] Add notes about deployment considerations
- [x] **Task 2.4.4.4: Parameter Injection Fix** (Option 1: Create params table) **[COMPLETED - 2025-06-23]**

    - [ ] **Phase 6.1 - Individual example test results**
        - [x] done in a separate file ADAPTER-TODO.md
    - [ ] 6.2 **run all examples using command line and make sure output matches what's in the script like print()**


#### 2.4.5: Documentation & Examples
- [x] **Task 2.4.5.1: CODE documentation** **[COMPLETED - 2025-06-22]**

- [x] **Task 2.4.5.2: User Guide** (`/docs/user-guide/`) **[COMPLETED - 2025-06-22]**
  - [x] Getting started with Lua spells (`lua-spells.md`) **[COMPLETED - 2025-06-22]**
  - [x] Complete API reference (`api-reference.md`) **[COMPLETED - 2025-06-22]**
  - [x] Common patterns and idioms (`common-patterns.md`) **[COMPLETED - 2025-06-22]**
  - [x] Troubleshooting guide (`troubleshooting.md`) **[COMPLETED - 2025-06-22]**
  - [x] Migration from pure Lua (`migration-from-pure-lua.md`) **[COMPLETED - 2025-06-22]**

- [x] **Task 2.4.5.2: Example Spells** (`/examples/spells/lua/`) **[COMPLETED - 2025-06-22]**
  - [x] Calling builtin tools by themselves (`01-tools-usage.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Basic LLM interaction (`02-basic-llm.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Agent without tools (plain llm) (`03-agent-plain.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Agent with tools (`04-agent-with-tools.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Agent with tools, one of which is an agent wrapped as a tool (`05-agent-as-tool.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Complex workflows (`06-complex-workflows.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Event-driven spells (`07-event-driven.lua`) **[COMPLETED - 2025-06-22]**
  - [x] Performance patterns (`08-performance-patterns.lua`) **[COMPLETED - 2025-06-22]**
  - [x] State management example (`09-state-management.lua`) **[COMPLETED - 2025-06-23]**
  - [x] Hooks Example (`10-hooks.lua`) **[COMPLETED - 2025-06-23]**
  - [x] Debug usage (`11-debug-usage.lua`) **[COMPLETED - 2025-06-23]**
  - [x] Custom Tool creation and use in lua (`12-custom-tool.lua`) **[COMPLETED - 2025-06-23]**
  - [x] Agent handoff to another agent example (`13-agent-handoff.lua`) **[COMPLETED - 2025-06-23]**

- [ ] **Task 2.4.5.3: Developer Documentation** (`/docs/technical/`)
  - [ ] Architecture deep dive
  - [ ] Extension guide
  - [ ] Performance tuning
  - [ ] Security best practices
  - [ ] Contribution guide

---
## Phase 3: Spell Runner CLI - COMPLETED [2025-06-21]
---
## Phase 4: JavaScript Engine Implementation

### 4.1 JavaScript Engine Research and Planning
- [ ] 4.1.1. Research goja (https://github.com/dop251/goja) go. Find the best javascript engine to work with in go-llmspell (There are others). 
- [ ] 4.1.2. Research how to integrate the chosen javascript engine into this go-llmspell library. add additional TODO.md entries as needed 
- [ ] 4.1.3. Analyze state management and memory integration
- [ ] 4.1.4. Design ScriptValue ↔ javascript type conversion system 
- [ ] 4.1.5. Plan goroutine integration for async operations
- [ ] 4.1.6. Design security sandboxing approach
- [ ] 4.1.7. Create detailed implementation roadmap
- [ ] 4.1.8. Research  bytecode validation and security implications - may not apply to gopher-lua
- [ ] 4.1.9. Investigate warning system integration 
- [ ] 4.1.10. Study generational GC vs incremental GC trade-offs if it applies
- [ ] 4.1.11. Research goja debug introspection capabilities for development tools
- [ ] 4.1.12. Combine all research documents and re-synthesize into one javascript_engine_architecture.md based on `docs/technical/architecture.md` and a detailed implementation roadmap

### 4.2 JavaScript Engine Core
- [ ] **Task 4.2.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/javascript/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test Goja integration
  - [ ] Test ES6+ or ES5.1+ whichever is the lstest support
  - [ ] Create `/pkg/engine/javascript/engine.go`
  - [ ] Implement ScriptEngine interface for JS
  - [ ] Integrate Goja
  - [ ] Add ES6+ support
  - [ ] Update engine to use ScriptValue type system

- [ ] **Task 4.2.2: Type Converter**
  - [ ] Create test file `/pkg/engine/javascript/converter_test.go`
  - [ ] Test JS ↔ Go type conversions
  - [ ] Test Promise handling
  - [ ] Create `/pkg/engine/javascript/converter.go`
  - [ ] Implement type conversions
  - [ ] Handle async patterns
  - [ ] Implement ScriptValue ↔ JS value converters

- [ ] **Task 4.2.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/javascript/sandbox_test.go`
  - [ ] Test global access restrictions
  - [ ] Test resource limits
  - [ ] Create `/pkg/engine/javascript/sandbox.go`
  - [ ] Restrict global access
  - [ ] Implement CSP-like policies

### 3.3 JavaScript Standard Library
- [ ] **Task 4.3.1: Core Modules**
  - [ ] Create `/pkg/engine/javascript/stdlib/core.js`
  - [ ] Create `/pkg/engine/javascript/stdlib/llm.js` - LLM bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/tools.js` - Tools bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/workflow.js` - Workflow bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/state.js` - State bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/events.js` - Events bridge wrapper
  - [ ] Create `/pkg/engine/javascript/stdlib/hooks.js` - Hooks bridge wrapper

---

## Phase 5: Tengo Engine Implementation

### 5.1 Tengo Engine Core
- [ ] **Task 5.1.1: Engine Implementation**
  - [ ] Create `/pkg/engine/tengo/engine.go`
  - [ ] Implement ScriptEngine interface for Tengo
  - [ ] Integrate Tengo VM
  - [ ] Optimize for performance
  - [ ] Update engine to use ScriptValue type system

- [ ] **Task 5.1.2: Type Converter**
  - [ ] Create `/pkg/engine/tengo/converter.go`
  - [ ] Implement Tengo ↔ Go conversions
  - [ ] Handle Tengo objects
  - [ ] Implement ScriptValue ↔ Tengo converters

- [ ] **Task 5.1.3: Security Sandbox**
  - [ ] Create `/pkg/engine/tengo/sandbox.go`
  - [ ] Implement Tengo restrictions
  - [ ] Add import controls

---
## Phase 6: Security Level & Feature Set Separation (Current Priority)

Based on `/docs/technical/security-feature-separation-implementation.md`, this phase implements the breaking change from `--profile` to `--security-level` and `--feature-set` flags with single-source enum definitions.

**See Task 2.4.4.5.7 above for detailed implementation plan.**

---
## Phase 7: Integration and Examples

### 7.1 Example Spells
- [ ] **Task 7.1.1: Basic Examples**
  - [ ] Hello World spell (all engines)
  - [ ] LLM chat spell
  - [ ] Tool usage spell
  - [ ] State management spell

- [ ] **Task 7.1.2: Advanced Examples**
  - [ ] Multi-agent orchestration spell
  - [ ] Complex workflow spell
  - [ ] Event-driven spell
  - [ ] Hook-based customization spell

### 7.2 Testing
- [ ] **Task 7.2.1: Cross-Engine Tests** **[DEFERRED from 1.3.21]**
  - [ ] Create conformance test suite
  - [ ] Verify API compatibility
  - [ ] Test performance characteristics

- [ ] **Task 7.2.2: Integration Tests**
  - [ ] Test bridge functionality
  - [ ] Test type conversions
  - [ ] Test error handling

### 7.3 Comprehensive Integration Testing (Moved from Task 2.4.4.5.6 Phase 6)
- [ ] **Bridge System Integration** (from Task 2.4.4.5.5):
  - [x] Test all renamed bridges are accessible with lazy loading
  - [x] Test all stdlib modules can load with correct bridge names and lazy initialization
  - [ ] Verify `tools.list()` works correctly with on-demand bridge loading (from 2.4.4.5.5)
  - [ ] Test other stdlib modules that require bridges work with lazy loading (from 2.4.4.5.5)
  - [ ] Ensure bridge globals are properly set when engine is loaded on-demand (from 2.4.4.5.5)
  - [ ] Test backward compatibility if needed
- [ ] **Example Spells Integration** (from Task 2.4.4.5.5):
  - [ ] Test all 13 example spells work with new bridge names AND lazy bridge initialization (from 2.4.4.5.5)
  - [ ] Run each spell with different security profiles (`--profile=development`, `--profile=sandbox`)
  - [ ] Verify bridge-dependent functionality works correctly with on-demand loading (from 2.4.4.5.5)
  - [ ] Test that only needed bridges are loaded for each spell (lazy loading verification from 2.4.4.5.5)
  - [ ] Verify unified globals (`state`, `observability`) work with new bridge names
  - [ ] Document any breaking changes
- [ ] **Repl Integration** (from Task 2.4.4.5.5):
  - [ ] Test that repl works with lazy loading for different profiles
- [ ] **Performance Integration** (from Task 2.4.4.5.5):
  - [ ] Measure startup time improvement with lazy loading (from 2.4.4.5.5)
  - [ ] Verify memory usage is reduced for simple scripts (from 2.4.4.5.5)
  - [ ] Test CLI responsiveness for engine listing commands (from 2.4.4.5.5)
  - [ ] Ensure bridge renaming doesn't impact performance
  - [ ] Verify lazy loading still works with new bridge names
- [ ] **Security Profile Integration**:
  - [ ] Test bridge profile mapping works with lazy loading
  - [ ] Verify security profiles load correct bridge sets on-demand
  - [ ] Test that bridge profiles respect lazy loading behavior

---


## Phase 8: Deferred Tasks from Previous Phases
**DEFERRED TASKS from different Phases - For Revisit from previous Phases**
- See `TODO-DONE-ARCHIVE.md` for completed tasks history

### 8.1 More Production Readiness
- [ ] **Task 8.1.1: Monitoring & Metrics** **DEFERRED from  2.4.4.3**
  - [ ] Add Prometheus metrics
  - [ ] Implement health checks
  - [ ] Create performance dashboards
  - [ ] Add distributed tracing
  - [ ] Implement alerting rules

- [ ] **Task 8.1.2: Security Hardening** **DEFERRED from 2.4.4.4**
  - [ ] Conduct security audit
  - [ ] Add input validation
  - [ ] Implement rate limiting
  - [ ] Create security benchmarks
  - [ ] Add CVE scanning

### 8.2 Model Info Bridge Intelligence **[DEFERRED from  1.4.6 ]** - Features not in go-llms

- [ ] **Task 8.2.1: Add Model Performance Analytics** ⏸️ **[DEFERRED from 1.4.6.1]**
  - Missing from go-llms: Model performance tracking, analytics, metrics
  - Documented in upstream request #1

- [ ] **Task 8.2.2: Add Model Recommendation Engine** ⏸️ **[DEFERRED from 1.4.6.2]**  
  - Missing from go-llms: Recommendation algorithms, model selection
  - Documented in upstream request #2

- [ ] **Task 8.2.3: Add Model Catalog Export** ⏸️ **[DEFERRED from 1.4.6.3]**
  - Missing from go-llms: Catalog export, OpenAPI generation for models
  - Documented in upstream request #3

### 8.3 Additional bridgest from go-llms 

- [ ] **Task 8.3.1: Memory Bridge** ⏸️ **[DEFERRED from 1.5.8]** - Not in go-llms yet
  - [ ] Will implement when available in go-llms

- [ ] **Task 8.3.2: Conversation Bridge** ⏸️ **[DEFERRED from 1.5.9]** - Not in go-llms yet
  - [ ] Will implement when available in go-llms

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