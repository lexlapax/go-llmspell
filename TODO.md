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
- 🚧 Phase 2: Lua Engine Implementation - IN PROGRESS
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
  - 🚧 Phase 2.4: Advanced Features & Optimization - NOT STARTED
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

#### 2.4.2: Spell Runner command line (`/cmd/llmspell/main.go`)
  - [x] **Task 2.4.2.1: Research** research how to create a script runner that can run spells **[COMPLETED - 2025-06-20]**
    - [x] look at existing code
    - [x] figure out how to do script detection (lua, javascript) - file extension based
    - [x] figure out how to load the right script engine bridge - via registry
    - [x] research what commands of commands do we need
    - [x] arg parsing engine with flags - Kong recommended
    - [x] come up with implementation plan - see `/docs/spell-runner-research.md`
  
  - [ ] **Task 2.4.2.2: Core Implementation**
    - [ ] **2.4.2.2.1: Core Runner Package** (`/pkg/runner/`)
      - [ ] Create `/pkg/runner/runner.go` with Runner interface
      - [ ] Create `/pkg/runner/spell_loader.go` to parse spell.yaml files
      - [ ] Create `/pkg/runner/engine_selector.go` for engine detection
      - [ ] Create `/pkg/runner/executor.go` for script execution
      - [ ] Add parameter passing mechanism to scripts
      - [ ] Write comprehensive tests for runner package
    
    - [ ] **2.4.2.2.2: CLI Structure with Kong** (`/cmd/llmspell/`)
      - [ ] Set up Kong dependency (`go get github.com/alecthomas/kong`)
      - [ ] Create `/cmd/llmspell/main.go` with Kong CLI struct
      - [ ] Create `/cmd/llmspell/commands/` directory
      - [ ] Implement `/cmd/llmspell/commands/run.go` (default command)
      - [ ] Implement `/cmd/llmspell/commands/validate.go`
      - [ ] Implement `/cmd/llmspell/commands/engines.go`
      - [ ] Implement `/cmd/llmspell/commands/version.go`
      - [ ] create `/cmd/llmspell/commands/repl.go` which uses `/pkg/repl` see tasks 2.4.2.2.4
      - [ ] Add global flags (--debug, --config, --quiet, --verbose)
      - [ ] Set up Kong help formatting and error handling
    
    - [ ] **2.4.2.2.3: Configuration with Koanf** (`/pkg/config/`)
      - [ ] Set up Koanf v2 dependencies
      - [ ] Create `/pkg/config/config.go` with Config struct
      - [ ] Create `/pkg/config/loader.go` for loading configs
      - [ ] Implement default configuration values
      - [ ] Add config file support (~/.llmspell/config.yaml)
      - [ ] Add environment variable support (LLMSPELL_*)
      - [ ] Add command-line flag integration
      - [ ] Implement config validation
      - [ ] Add Watch() support for config reloading
      - [ ] Create `/cmd/llmspell/commands/config.go` command
    
    - [ ] **2.4.2.2.4: REPL Implementation** (`/pkg/repl/`)
      - [ ] Set up readline dependency (`go get github.com/chzyer/readline`)
      - [ ] Create `/pkg/repl/repl.go` with REPL interface
      - [ ] Create `/pkg/repl/base_repl.go` with common functionality
      - [ ] Implement `/pkg/repl/lua_repl.go` with Lua-specific features
      - [ ] Implement `/pkg/repl/completer.go` for auto-completion
      - [ ] Add REPL commands (.help, .exit, .clear, .load, .save)
      - [ ] Add history persistence (~/.llmspell_history)
      - [ ] Add syntax highlighting support
      - [ ] implement `/cmd/llmspell/commands/repl.go` command
      - [ ] Write tests for REPL functionality
    
    - [ ] **2.4.2.2.5: Security & Validation**
      - [ ] Create `/pkg/security/profiles.go` for security profiles
      - [ ] Implement sandbox, development, and production profiles
      - [ ] Create `/pkg/validator/validator.go` interface
      - [ ] use `/pkg/engine/gopherlua/validator.go`
      - [ ] Add spell.yaml schema validation
      - [ ] Create `/cmd/llmspell/commands/security.go` command
      - [ ] Add security checks to runner
    
    - [ ] **2.4.2.2.6: Testing & Integration**
      - [ ] Create integration test suite
      - [ ] Add example spells for testing
      - [ ] Test all commands with various inputs
      - [ ] Test configuration layering
      - [ ] Test REPL functionality
      - [ ] Add benchmarks for performance
      - [ ] Ensure cross-platform compatibility
  
  - [ ] **Task 2.4.2.3: Documentation**
    - [ ] Create `/docs/cli-usage.md` with command reference
    - [ ] Create `/docs/configuration.md` with config guide
    - [ ] Create `/docs/repl-guide.md` for REPL usage
    - [ ] Update main README.md with CLI examples
    - [ ] Add inline help text for all commands
    - [ ] Create man page generation
    - [ ] Add shell completion scripts


#### 2.4.3: Development Tools
- [ ] **Task 2.4.3.1: Debugger Support** (`/pkg/engine/gopherlua/debug.go`)
  - [ ] Implement breakpoint support
  - [ ] Add step debugging
  - [ ] Create variable inspection
  - [ ] Implement stack trace visualization
  - [ ] Add watch expressions

- [x] **Task 2.4.3.2: Script Validator** (`/pkg/engine/gopherlua/validator.go`) **[COMPLETED - 2025-06-20]**
  - [x] Implement syntax validation
  - [x] Add type checking where possible (limited by Lua's dynamic nature)
  - [x] Create linting rules
  - [x] Implement security validation
  - [x] Add performance warnings

- [ ] **Task 2.4.3.3: Documentation Generator** (`/pkg/engine/gopherlua/docs.go`)
  - [ ] Extract API from bridges
  - [ ] Generate Lua documentation
  - [ ] Create example extraction
  - [ ] Add type annotations
  - [ ] Generate completion data

#### 2.4.4: Production Readiness
- [ ] **Task 2.4.4.1: Comprehensive Testing**
  - [ ] Achieve 90%+ test coverage
  - [ ] Add integration test suite
  - [ ] Create stress tests
  - [ ] Implement chaos testing
  - [ ] Add regression test suite

- [ ] **Task 2.4.4.2: Error Handling Enhancement**
  - [ ] Standardize error types
  - [ ] Add error categorization
  - [ ] Implement error recovery
  - [ ] Create error reporting
  - [ ] Add error metrics

- [ ] **Task 2.4.4.3: Monitoring & Metrics**
  - [ ] Add Prometheus metrics
  - [ ] Implement health checks
  - [ ] Create performance dashboards
  - [ ] Add distributed tracing
  - [ ] Implement alerting rules

- [ ] **Task 2.4.4.4: Security Hardening**
  - [ ] Conduct security audit
  - [ ] Add input validation
  - [ ] Implement rate limiting
  - [ ] Create security benchmarks
  - [ ] Add CVE scanning

#### 2.4.5: Documentation & Examples
- [ ] **Task 2.4.5.1: User Guide** (`/docs/user-guide/lua/`)
  - [ ] Getting started with Lua spells
  - [ ] Complete API reference
  - [ ] Common patterns and idioms
  - [ ] Troubleshooting guide
  - [ ] Migration from pure Lua

- [ ] **Task 2.4.5.2: Example Spells** (`/examples/lua/`)
  - [ ] Basic LLM interaction
  - [ ] Agent with tools
  - [ ] Complex workflows
  - [ ] Event-driven spells
  - [ ] Performance patterns

- [ ] **Task 2.4.5.3: Developer Documentation**
  - [ ] Architecture deep dive
  - [ ] Extension guide
  - [ ] Performance tuning
  - [ ] Security best practices
  - [ ] Contribution guide

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
  - [ ] Update engine to use ScriptValue type system

- [ ] **Task 3.2.2: Type Converter**
  - [ ] Create test file `/pkg/engine/javascript/converter_test.go`
  - [ ] Test JS ↔ Go type conversions
  - [ ] Test Promise handling
  - [ ] Create `/pkg/engine/javascript/converter.go`
  - [ ] Implement type conversions
  - [ ] Handle async patterns
  - [ ] Implement ScriptValue ↔ JS value converters

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
  - [ ] Update engine to use ScriptValue type system

- [ ] **Task 4.1.2: Type Converter**
  - [ ] Create `/pkg/engine/tengo/converter.go`
  - [ ] Implement Tengo ↔ Go conversions
  - [ ] Handle Tengo objects
  - [ ] Implement ScriptValue ↔ Tengo converters

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