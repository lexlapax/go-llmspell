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
- ✅ Phase 1.1: Script Engine Interface [COMPLETED]
- ✅ Phase 1.2: Core Bridge Foundation [COMPLETED]
- ✅ Phase 1.3: Core Bridge System [COMPLETED]
- ✅ Phase 1.4.1: Foundation Updates [COMPLETED - 2025-06-15]
- ✅ Phase 1.4.2: State Bridge Enhancements [COMPLETED - 2025-06-15]
- ✅ Phase 1.4.3: Utility Bridge Upgrades [COMPLETED - 2025-06-16]
- ✅ Phase 1.4.4: LLM Bridge Advanced Features [COMPLETED - 2025-06-16]
- ✅ Phase 1.4.5: Schema Bridge Full Implementation [COMPLETED - 2025-06-16]
- ⏸️ Phase 1.4.6: Model Info Bridge Intelligence [DEFERRED - Not in go-llms]
- ✅ Phase 1.4.7: Agent Bridge Advanced Features [COMPLETED - 2025-06-16]
- ✅ Phase 1.4.8: Event Bridge Replacement [COMPLETED - 2025-06-16]
- ✅ Phase 1.4.11: Engine Integration [COMPLETED - 2025-06-16]
- ✅ Phase 1.5: Additional Original Bridges [COMPLETED - 2025-06-16]
- 🚧 Phase 2-5: Engine Implementations - NOT STARTED

---

## Phase 1: Engine and Bridge Foundation

### ✅ 1.1 Script Engine Interface [COMPLETED]

### ✅ 1.2 Core Bridge Foundation [COMPLETED]

### ✅ 1.3 Core Bridge System [COMPLETED]
#### Items for revisit:
  - [ ] Support for async/promise-based tool execution (deferred to script engine implementation)
  - [ ] Test cross-engine compatibility (deferred to script engine implementation)

### 1.4 v0.3.5 Feature Integration

#### ✅ 1.4.1 Foundation Updates [COMPLETED - 2025-06-15]

All foundation updates for go-llms v0.3.5 integration completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.2 State Bridge Enhancements [COMPLETED - 2025-06-15]

All state bridge enhancements completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.3 Utility Bridge Upgrades [COMPLETED - 2025-06-16]

All utility bridge upgrades completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.4 LLM Bridge Advanced Features [COMPLETED - 2025-06-16]

All LLM Bridge advanced features completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.5 Schema Bridge Full Implementation [COMPLETED - 2025-06-16]

All schema bridge full implementation completed. See TODO-DONE.md for detailed completion summary.

#### ⏸️ 1.4.6 Model Info Bridge Intelligence [DEFERRED - Features not in go-llms]

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

**Next Action**: Contribute features to go-llms first, then implement bridges

#### ✅ 1.4.7 Agent Bridge Advanced Features [COMPLETED - 2025-06-16]

All agent bridge advanced features completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.8 Event Bridge Replacement [COMPLETED - 2025-06-16]

All event bridge replacement features completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.9 Tools Bridge Enhancement [COMPLETED - 2025-06-16]

All tools bridge enhancement features completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.10 Workflow Bridge Serialization [COMPLETED - 2025-06-16]

All workflow bridge serialization features completed. See TODO-DONE.md for detailed completion summary.

#### ✅ 1.4.11 Engine Integration [COMPLETED - 2025-06-16]

Enhanced engine capabilities for advanced scripting needs. Bridge go-llms core functionality for profiling, events, and API generation. See TODO-DONE.md for detailed completion summary.

### ✅ 1.5 Additional Original Bridges [COMPLETED - 2025-06-16]

All additional original bridges completed. See TODO-DONE.md for detailed completion summary.

#### Items for revisit:

- [ ] **Task 1.5.8: Memory Bridge** ⏸️ **[DEFERRED - Not in go-llms yet]**
  - [ ] Will implement when available in go-llms

- [ ] **Task 1.5.9: Conversation Bridge** ⏸️ **[DEFERRED - Not in go-llms yet]**
  - [ ] Will implement when available in go-llms

### 1.6 Logging Infrastructure

- [ ] **Task 1.6.1: Debug Logging Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/util/debug.go`
  - [ ] Bridge go-llms debug logging system (`pkg/internal/debug`)
  - [ ] Support component-based debug control via `GO_LLMS_DEBUG` environment variable
  - [ ] Expose debug.Printf and debug.Println to scripts
  - [ ] Enable conditional compilation support (debug vs production builds)
  - [ ] Support custom logger integration with debug.SetLogger
  - [ ] Check tests to use go-llms pkg/testutils and normalize for duplicate patterns

- [ ] **Task 1.6.2: Structured Logging Bridge**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/util/slog.go`
  - [ ] Bridge slog integration for structured logging (`pkg/agent/core/logging_hook.go`)
  - [ ] Expose log levels: Basic, Detailed, Debug with emoji enhancement
  - [ ] Support structured logging with key-value pairs from scripts
  - [ ] Enable log message truncation and JSON marshaling
  - [ ] Bridge LoggingHook for agent operations
  - [ ] Check tests to use go-llms pkg/testutils and normalize for duplicate patterns

- [ ] **Task 1.6.3: Script Logger Interface Design**
  - [ ] Ensure we leverage imports from go-llms pkg
  - [ ] Create `/pkg/bridge/util/script_logger.go`
  - [ ] Design unified script-friendly logging APIs
  - [ ] Combine debug and structured logging capabilities
  - [ ] Support context propagation through script calls
  - [ ] Integrate with existing bridge error handling
  - [ ] Support log formatting and output customization
  - [ ] Enable logger configuration from script environments
  - [ ] Check tests to use go-llms pkg/testutils and normalize for duplicate patterns

---

## Phase 2: Lua Engine Implementation
### 2.1 Lua Engine Research and Planning
- [ ] 2.1.1. Research gopher-lua integration with go and add additional TODO.md entries as needed 
- [ ] 2.1.2. Analyze lua_State management and memory integration
- [ ] 2.1.3. Design ScriptValue ↔ Lua type conversion system 
- [ ] 2.1.4. Plan coroutine integration for async operations
- [ ] 2.1.5. Design security sandboxing approach
- [ ] 2.1.6. Create detailed implementation roadmap
- [ ] 2.1.7. Research Lua bytecode validation and security implications - may not apply to gopher-lua
- [ ] 2.1.8. Investigate gopher-lua, Lua 5.4 warning system integration 
- [ ] 2.1.9. Study gopher-Lua generational GC vs incremental GC trade-offs
- [ ] 2.1.10. Research Lua debug introspection capabilities for development tools
- [ ] 2.1.11. Combine all research documents and re-synthesize into one lua_engine_architecture.md document

### 2.2 Lua Engine Core
- [ ] **Task 2.2.1: Engine Implementation**
  - [ ] Create test file `/pkg/engine/lua/engine_test.go`
  - [ ] Test ScriptEngine interface implementation
  - [ ] Test GopherLua integration
  - [ ] Test resource limits enforcement
  - [ ] Create `/pkg/engine/lua/engine.go`
  - [ ] Implement ScriptEngine interface for Lua
  - [ ] Integrate GopherLua
  - [ ] Implement resource limits

- [ ] **Task 2.2.2: Type Converter**
  - [ ] Create test file `/pkg/engine/lua/converter_test.go`
  - [ ] Test Lua ↔ Go type conversions
  - [ ] Test Lua tables → Go maps/arrays
  - [ ] Create `/pkg/engine/lua/converter.go`
  - [ ] Implement type conversions
  - [ ] Optimize for performance

- [ ] **Task 2.2.3: Security Sandbox**
  - [ ] Create test file `/pkg/engine/lua/sandbox_test.go`
  - [ ] Test dangerous functions disabled
  - [ ] Test file system restrictions
  - [ ] Create `/pkg/engine/lua/sandbox.go`
  - [ ] Disable dangerous Lua functions
  - [ ] Implement restrictions

### 2.3 Lua Standard Library
- [ ] **Task 2.3.1: Core Modules**
  - [ ] Create `/pkg/engine/lua/stdlib/core.lua`
  - [ ] Create `/pkg/engine/lua/stdlib/llm.lua` - LLM bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/tools.lua` - Tools bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/workflow.lua` - Workflow bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/state.lua` - State bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/events.lua` - Events bridge wrapper
  - [ ] Create `/pkg/engine/lua/stdlib/hooks.lua` - Hooks bridge wrapper

---

## Phase 3: JavaScript Engine Implementation

### 3.1 JavaScript Engine Research and Planning
- [ ] 3.1.1. Research goja (https://github.com/dop251/goja) integration with go and add additional TODO.md entries as needed 
- [ ] 3.1.2. Analyze state management and memory integration
- [ ] 3.1.3. Design ScriptValue ↔ javascript type conversion system 
- [ ] 3.1.4. Plan goroutine integration for async operations
- [ ] 3.1.5. Design security sandboxing approach
- [ ] 3.1.6. Create detailed implementation roadmap
- [ ] 3.1.7. Research  bytecode validation and security implications - may not apply to gopher-lua
- [ ] 3.1.8. Investigate warning system integration 
- [ ] 3.1.9. Study generational GC vs incremental GC trade-offs if it applies
- [ ] 3.1.10. Research goja debug introspection capabilities for development tools
- [ ] 3.1.11. Combine all research documents and re-synthesize into one lua_engine_architecture.md document

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