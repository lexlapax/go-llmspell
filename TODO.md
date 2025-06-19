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
  - Phase 2.1: Research and Planning ✅ COMPLETED [2025-06-17]
  - Phase 2.2: Core Engine Components ✅ COMPLETED [2025-06-18]
  - Phase 2.3: Bridge Integration Layer 🚧 IN PROGRESS
    - ✅ 2.3.1: Module System Architecture [COMPLETED - 2025-06-19]
    - ✅ 2.3.2: Async/Coroutine Support [COMPLETED - 2025-06-19]
    - ✅ 2.3.2.0: ScriptValue Type System Refactoring [COMPLETED - 2025-06-19]
    - ✅ 2.3.2.0.X: Fix ScriptValue Bridge Test Failures [COMPLETED - 2025-06-19]
    - ✅ 2.3.2.5: Test Utilities Extraction [COMPLETED - 2025-06-19]
    - 🚧 2.3.3: Bridge Adapters [IN PROGRESS - 16 of 24 completed]
    - 🚧 2.3.4: Lua Standard Library [IN PROGRESS]
  - Phase 2.4: Advanced Features & Optimization - NOT STARTED
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
🚧 **IN PROGRESS** - 15 of 24 adapters implemented
**See TODO-DONE.md for completed task details**

**Design Decision**: Tasks 15-24 implement complete method flattening for consistency:
- ALL adapter methods use flat naming at Lua level (e.g., `poolCreate` instead of `pool.create`)
- Backend bridges maintain their namespace organization unchanged
- This provides consistent developer experience across ALL adapters
- Easier method discovery and simpler API surface
- Flattening 51 namespaces containing 200+ methods across 10 adapters

**Task Dependencies Analysis**:
- Tasks 16-17 MUST be completed before LLM namespace flattening (part of Task 18)
- Task 15 is independent and can be done anytime
- Tasks 18-24 depend on 16-17 for LLM adapter but are independent for other adapters
- Recommended order: 15 → 16-17 → 18-24

- [x] **Task 2.3.3.15: Tool Registry Bridge Enhancement** ✅ **[COMPLETED - 2025-06-19]** (enhance `/pkg/engine/gopherlua/adapters/tools.go`)
- [x] **Task 2.3.3.16: LLM Pool Bridge Enhancement** ✅ **[COMPLETED - 2025-06-19]** (enhance `/pkg/engine/gopherlua/adapters/llm.go`)
- [x] **Task 2.3.3.17: LLM Providers Bridge Enhancement** ✅ **[COMPLETED - 2025-06-19]** (enhance `/pkg/engine/gopherlua/adapters/llm.go`)

- [x] **Task 2.3.3.18: Events Adapter Namespace Flattening** ✅ **[COMPLETED - 2025-06-19]** (enhance `/pkg/engine/gopherlua/adapters/events.go`)
  - [x] Flatten bus namespace methods:
    - [x] events.bus.publish → events.busPublish
    - [x] events.bus.subscribe → events.busSubscribe  
    - [x] events.bus.unsubscribe → events.busUnsubscribe
  - [x] Flatten filters namespace methods:
    - [x] events.filters.create → events.filtersCreate
    - [x] events.filters.createComposite → events.filtersCreateComposite
  - [x] Flatten recording namespace methods:
    - [x] events.recording.start → events.recordingStart
    - [x] events.recording.stop → events.recordingStop
    - [x] events.recording.isRecording → events.recordingIsRecording
  - [x] Flatten replay namespace methods:
    - [x] events.replay.start → events.replayStart
    - [x] events.replay.pause → events.replayPause
    - [x] events.replay.resume → events.replayResume
    - [x] events.replay.stop → events.replayStop
  - [x] Flatten aggregation namespace methods:
    - [x] events.aggregation.create → events.aggregationCreate
    - [x] events.aggregation.getData → events.aggregationGetData
  - [x] Update tests in events_test.go

- [x] **Task 2.3.3.19: State Adapter Namespace Flattening** ✅ **[COMPLETED - 2025-06-19]** (enhance `/pkg/engine/gopherlua/adapters/state.go`)
  - [x] Flatten transforms namespace methods:
    - [x] state.transforms.register → state.transformsRegister
    - [x] state.transforms.apply → state.transformsApply
    - [x] state.transforms.chain → state.transformsChain
    - [x] state.transforms.validate → state.transformsValidate
    - [x] state.transforms.getAvailable → state.transformsGetAvailable
  - [x] Flatten context namespace methods:
    - [x] state.context.get → state.contextGet
    - [x] state.context.set → state.contextSet
    - [x] state.context.merge → state.contextMerge
    - [x] state.context.clear → state.contextClear
    - [x] state.context.createShared → state.contextCreateShared
    - [x] state.context.withInheritance → state.contextWithInheritance
  - [x] Flatten persistence namespace methods:
    - [x] state.persistence.save → state.persistenceSave
    - [x] state.persistence.load → state.persistenceLoad
    - [x] state.persistence.exists → state.persistenceExists
    - [x] state.persistence.delete → state.persistenceDelete
    - [x] state.persistence.listVersions → state.persistenceListVersions
  - [x] Update tests in state_test.go

- [x] **Task 2.3.3.20: Utils Adapter Namespace Flattening** (enhance `/pkg/engine/gopherlua/adapters/utils.go`) ✅ [2025-06-19]
  - [x] Flatten auth namespace methods:
    - [x] utils.auth.generateToken → utils.authGenerateToken
    - [x] utils.auth.validateToken → utils.authValidateToken
    - [x] utils.auth.hashPassword → utils.authHashPassword
    - [x] utils.auth.verifyPassword → utils.authVerifyPassword
  - [x] Flatten debug namespace methods:
    - [x] utils.debug.trace → utils.debugTrace
    - [x] utils.debug.profile → utils.debugProfile
    - [x] utils.debug.dump → utils.debugDump
    - [x] utils.debug.assert → utils.debugAssert
  - [x] Flatten errors namespace methods:
    - [x] utils.errors.wrap → utils.errorsWrap
    - [x] utils.errors.unwrap → utils.errorsUnwrap
    - [x] utils.errors.isType → utils.errorsIsType
    - [x] utils.errors.getStack → utils.errorsGetStack
  - [x] Flatten json namespace methods:
    - [x] utils.json.encode → utils.jsonEncode
    - [x] utils.json.decode → utils.jsonDecode
    - [x] utils.json.validate → utils.jsonValidate
    - [x] utils.json.prettify → utils.jsonPrettify
  - [x] Flatten llm namespace methods:
    - [x] utils.llm.parseResponse → utils.llmParseResponse
    - [x] utils.llm.formatPrompt → utils.llmFormatPrompt
    - [x] utils.llm.countTokens → utils.llmCountTokens
    - [x] utils.llm.splitMessage → utils.llmSplitMessage
  - [x] Flatten logger namespace methods:
    - [x] utils.logger.log → utils.loggerLog
    - [x] utils.logger.error → utils.loggerError
    - [x] utils.logger.warn → utils.loggerWarn
    - [x] utils.logger.info → utils.loggerInfo
    - [x] utils.logger.debug → utils.loggerDebug
  - [x] Flatten slog namespace methods:
    - [x] utils.slog.info → utils.slogInfo
    - [x] utils.slog.error → utils.slogError
    - [x] utils.slog.warn → utils.slogWarn
    - [x] utils.slog.debug → utils.slogDebug
    - [x] utils.slog.withFields → utils.slogWithFields
  - [x] Flatten general namespace methods:
    - [x] utils.general.uuid → utils.generalUuid
    - [x] utils.general.hash → utils.generalHash
    - [x] utils.general.encode → utils.generalEncode
    - [x] utils.general.decode → utils.generalDecode
  - [x] Update tests in utils_test.go

- [x] **Task 2.3.3.21: Agent Adapter Namespace Flattening** ✅ **[COMPLETED - 2025-06-19]** (enhance `/pkg/engine/gopherlua/adapters/agent.go`)
  - [x] check if agent bridge has addTool or addTools or similar method.. it should, check in go-llms agent methods and report back
    - Found: Agent has AddTool(tool Tool) method, no AddTools bulk method
    - Note: registerAgentTool is just an alias for registerTool
    - Pattern: To add agent as tool, wrap with AgentTool first, then use AddTool
  - [x] Flatten lifecycle namespace methods:
    - [x] agent.lifecycle.create → agent.lifecycleCreate
    - [x] agent.lifecycle.createLLM → agent.lifecycleCreateLLM
    - [x] agent.lifecycle.list → agent.lifecycleList
    - [x] agent.lifecycle.get → agent.lifecycleGet
    - [x] agent.lifecycle.remove → agent.lifecycleRemove
    - [x] agent.lifecycle.getMetrics → agent.lifecycleGetMetrics
  - [x] Flatten communication namespace methods:
    - communications methods can be shorted to omit the communication altogether.
    - [x] agent.communication.run → agent.run
    - [x] agent.communication.runAsync → agent.runAsync
    - [x] agent.communication.registerTool → agent.registerTool
    - [x] agent.communication.unregisterTool → agent.unregisterTool
    - [x] agent.communication.listTools → agent.listTools
  - [x] Flatten state namespace methods:
    - [x] agent.state.get → agent.stateGet
    - [x] agent.state.set → agent.stateSet
    - [x] agent.state.export → agent.stateExport
    - [x] agent.state.import → agent.stateImport
    - [x] agent.state.saveSnapshot → agent.stateSaveSnapshot
    - [x] agent.state.loadSnapshot → agent.stateLoadSnapshot
    - [x] agent.state.listSnapshots → agent.stateListSnapshots
  - [x] Flatten events namespace methods:
    - [x] agent.events.emit → agent.eventsEmit
    - [x] agent.events.subscribe → agent.eventsSubscribe
    - [x] agent.events.unsubscribe → agent.eventsUnsubscribe
    - [x] agent.events.startRecording → agent.eventsStartRecording
    - [x] agent.events.stopRecording → agent.eventsStopRecording
    - [x] agent.events.replay → agent.eventsReplay
  - [x] Flatten profiling namespace methods:
    - [x] agent.profiling.start → agent.profilingStart
    - [x] agent.profiling.stop → agent.profilingStop
    - [x] agent.profiling.getMetrics → agent.profilingGetMetrics
    - [x] agent.profiling.getReport → agent.profilingGetReport
  - [x] Flatten workflow namespace methods:
    - [x] agent.workflow.create → agent.workflowCreate
    - [x] agent.workflow.execute → agent.workflowExecute
    - [x] agent.workflow.addStep → agent.workflowAddStep
  - [x] Flatten hooks namespace methods:
    - [x] agent.hooks.register → agent.hooksRegister
    - [x] agent.hooks.set → agent.hooksSet
    - [x] agent.hooks.unregister → agent.hooksUnregister
  - [x] Flatten utils namespace methods:
    - [x] agent.utils.validateConfig → agent.utilsValidateConfig
  - [x] Update tests in agent_test.go

- [ ] **Task 2.3.3.22: Structured Adapter Namespace Flattening** (enhance `/pkg/engine/gopherlua/adapters/structured.go`)
  - [ ] Flatten validation namespace methods:
    - [ ] structured.validation.validate → structured.validationValidate
    - [ ] structured.validation.validatePartial → structured.validationValidatePartial
    - [ ] structured.validation.getErrors → structured.validationGetErrors
    - [ ] structured.validation.addCustom → structured.validationAddCustom
  - [ ] Flatten generation namespace methods:
    - [ ] structured.generation.fromType → structured.generationFromType
    - [ ] structured.generation.fromTags → structured.generationFromTags
    - [ ] structured.generation.fromJSONSchema → structured.generationFromJSONSchema
  - [ ] Flatten repository namespace methods:
    - [ ] structured.repository.save → structured.repositorySave
    - [ ] structured.repository.load → structured.repositoryLoad
    - [ ] structured.repository.list → structured.repositoryList
    - [ ] structured.repository.delete → structured.repositoryDelete
  - [ ] Flatten importExport namespace methods:
    - [ ] structured.importExport.toJSON → structured.importExportToJSON
    - [ ] structured.importExport.fromJSON → structured.importExportFromJSON
    - [ ] structured.importExport.toYAML → structured.importExportToYAML
    - [ ] structured.importExport.fromYAML → structured.importExportFromYAML
  - [ ] Flatten custom namespace methods:
    - [ ] structured.custom.register → structured.customRegister
    - [ ] structured.custom.execute → structured.customExecute
    - [ ] structured.custom.list → structured.customList
  - [ ] Flatten utils namespace methods:
    - [ ] structured.utils.merge → structured.utilsMerge
    - [ ] structured.utils.diff → structured.utilsDiff
    - [ ] structured.utils.transform → structured.utilsTransform
  - [ ] Update tests in structured_test.go

- [ ] **Task 2.3.3.23: ModelInfo Adapter Namespace Flattening** (enhance `/pkg/engine/gopherlua/adapters/modelinfo.go`)
  - [ ] Flatten discovery namespace methods:
    - [ ] modelinfo.discovery.scan → modelinfo.discoveryScan
    - [ ] modelinfo.discovery.refresh → modelinfo.discoveryRefresh
    - [ ] modelinfo.discovery.getProviders → modelinfo.discoveryGetProviders
    - [ ] modelinfo.discovery.getModels → modelinfo.discoveryGetModels
  - [ ] Flatten capabilities namespace methods:
    - [ ] modelinfo.capabilities.check → modelinfo.capabilitiesCheck
    - [ ] modelinfo.capabilities.list → modelinfo.capabilitiesList
    - [ ] modelinfo.capabilities.compare → modelinfo.capabilitiesCompare
    - [ ] modelinfo.capabilities.getDetails → modelinfo.capabilitiesGetDetails
  - [ ] Flatten selection namespace methods:
    - [ ] modelinfo.selection.find → modelinfo.selectionFind
    - [ ] modelinfo.selection.rank → modelinfo.selectionRank
    - [ ] modelinfo.selection.filter → modelinfo.selectionFilter
    - [ ] modelinfo.selection.recommend → modelinfo.selectionRecommend
  - [ ] Update tests in modelinfo_test.go

- [ ] **Task 2.3.3.24: Observability Adapter Namespace Flattening** (enhance `/pkg/engine/gopherlua/adapters/observability.go`)
  - [ ] Flatten guardrails namespace methods:
    - [ ] observability.guardrails.registerRule → observability.guardrailsRegisterRule
    - [ ] observability.guardrails.check → observability.guardrailsCheck
    - [ ] observability.guardrails.enableRule → observability.guardrailsEnableRule
    - [ ] observability.guardrails.disableRule → observability.guardrailsDisableRule
  - [ ] Flatten metrics namespace methods:
    - [ ] observability.metrics.increment → observability.metricsIncrement
    - [ ] observability.metrics.gauge → observability.metricsGauge
    - [ ] observability.metrics.histogram → observability.metricsHistogram
    - [ ] observability.metrics.getAll → observability.metricsGetAll
  - [ ] Flatten tracing namespace methods:
    - [ ] observability.tracing.startSpan → observability.tracingStartSpan
    - [ ] observability.tracing.endSpan → observability.tracingEndSpan
    - [ ] observability.tracing.addAttribute → observability.tracingAddAttribute
    - [ ] observability.tracing.getTrace → observability.tracingGetTrace
  - [ ] Update tests in observability_test.go

#### Summary of Complete Namespace Flattening Scope

**All adapters being flattened in Phase 2.3.3 (Tasks 15-24)**:
- Tools (Task 15): Registry methods added as flat methods
- LLM (Tasks 16-17): Pool and Provider namespaces flattened (~15 methods)
- Events (Task 18): 5 namespaces with 15 methods flattened
- State (Task 19): 3 namespaces with 14 methods flattened  
- Utils (Task 20): 8 namespaces with 36 methods flattened
- Agent (Task 21): 8 namespaces with 31 methods flattened
- Structured (Task 22): 6 namespaces with 20 methods flattened
- ModelInfo (Task 23): 3 namespaces with 12 methods flattened
- Observability (Task 24): 3 namespaces with 12 methods flattened

**Total refactoring scope**: 51 namespaces with 200+ methods across all 10 adapters
**No deferrals** - Complete consistency across entire codebase!



#### 2.3.4: Async/Coroutine Support
Foundation for async operations that Lua Standard Library will build upon. This resolves deferred Task 1.3.20 for async/promise-based tool execution.

- [ ] **Task 2.3.4.1: Async Runtime** (`/pkg/engine/gopherlua/async.go`)
  - [ ] Implement `AsyncRuntime` for coroutine management
  - [ ] Add promise-coroutine integration
  - [ ] Create async execution context
  - [ ] Implement cancellation support
  - [ ] Add timeout handling

- [ ] **Task 2.3.4.2: Channel Integration** (`/pkg/engine/gopherlua/channels.go`)
  - [ ] Implement Go channel ↔ LChannel bridge
  - [ ] Add select operation support
  - [ ] Create buffered channel support
  - [ ] Implement channel closing
  - [ ] Add deadlock detection

- [ ] **Task 2.3.4.3: Async Bridge Methods** (`/pkg/engine/gopherlua/async_bridges.go`)
  - [ ] Wrap bridge methods for async execution
  - [ ] Add automatic promisification
  - [ ] Implement streaming support
  - [ ] Add progress callbacks
  - [ ] Create cancellation tokens

- [ ] **Task 2.3.4.4: Async Testing** (`/pkg/engine/gopherlua/async_test.go`)
  - [ ] Test coroutine lifecycle
  - [ ] Test promise integration
  - [ ] Test channel operations
  - [ ] Test cancellation and timeouts
  - [ ] Test concurrent async operations

#### 2.3.5: Lua Standard Library
Based on comprehensive research of all bridge adapters, these feature-oriented modules provide script-friendly APIs for complex operations. Each module requires comprehensive Go-based testing.

- [ ] **Task 2.3.5.1: Promise & Async Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/promise.lua`)
    - [ ] Implement Promise class with full async support
      - [ ] Add `Promise.new(executor)` constructor
      - [ ] Add `then/catch/finally` chain methods
      - [ ] Add `Promise.all(promises)` for concurrent execution
      - [ ] Add `Promise.race(promises)` for first-wins scenarios
      - [ ] Add `Promise.resolve(value)` and `Promise.reject(error)` helpers
    - [ ] Add async/await syntax sugar
      - [ ] Add `async(func)` wrapper for promise-returning functions
      - [ ] Add `await(promise, timeout)` method with timeout support
      - [ ] Add `sleep(duration)` utility for delays
    - [ ] Add coroutine integration
      - [ ] Add `spawn(func, args)` for concurrent execution
      - [ ] Add `yield()` for cooperative multitasking
      - [ ] Add channel-based communication helpers
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/promise_test.go`)
    - [ ] Test promise constructor and executor behavior
    - [ ] Test promise resolution/rejection with various types
    - [ ] Test promise chaining (then/catch/finally)
    - [ ] Test Promise.all concurrent execution
    - [ ] Test Promise.race timing behavior
    - [ ] Test timeout and cancellation
    - [ ] Test error propagation through chains
    - [ ] Test memory leaks in long chains
    - [ ] Test coroutine integration
    - [ ] Benchmark promise creation/resolution

- [ ] **Task 2.3.5.2: LLM Operations Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/llm.lua`)
    - [ ] High-level LLM operation helpers
      - [ ] Add `llm.quick_prompt(prompt, options)` for simple prompting
      - [ ] Add `llm.chat_session(system_prompt)` for conversation management
      - [ ] Add `llm.streaming_response(prompt, callback)` for streaming
      - [ ] Add `llm.batch_process(prompts, options)` for bulk operations
    - [ ] Provider management utilities
      - [ ] Add `llm.use_provider(name, config)` for easy provider switching
      - [ ] Add `llm.compare_providers(prompt, providers)` for A/B testing
      - [ ] Add `llm.fallback_chain(providers, prompt)` for reliability
    - [ ] Model discovery helpers
      - [ ] Add `llm.find_model(requirements)` for capability-based selection
      - [ ] Add `llm.model_info(model_id)` for metadata access
      - [ ] Add `llm.cost_estimate(operation, model)` for cost tracking
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/llm_test.go`)
    - [ ] Test with mock LLM bridge
    - [ ] Test streaming callbacks
    - [ ] Test batch processing limits
    - [ ] Test provider fallback chain
    - [ ] Test cost estimation accuracy
    - [ ] Test async operations with promises
    - [ ] Test error handling and retries
    - [ ] Test concurrent batch operations

- [ ] **Task 2.3.5.3: Agent Management Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/agent.lua`)
    - [ ] where is `agent.run(..)` or `agent.runAsync(..)`?
    - [ ] Agent lifecycle management
      - [ ] Add `agent.create(name, config)` for agent creation
      - [ ] Add `agent.configure(agent, settings)` for configuration
      - [ ] Add `agent.clone(agent, modifications)` for agent templating
    - [ ] Agent communication helpers
      - [ ] Add `agent.conversation(agent, messages)` for multi-turn chat
      - [ ] Add `agent.delegate(from_agent, to_agent, task)` for task delegation
      - [ ] Add `agent.collaborate(agents, task)` for multi-agent workflows
    - [ ] Agent tool integration
      - [ ] Add `agent.add_tools(agent, tools)` for tool assignment
      - [ ] Add `agent.create_tool(name, func, schema)` for custom tools
      - [ ] Add `agent.tool_chain(tools, data)` for tool pipelines
    - [ ] Workflow orchestration helpers -  **think hard.. I think the workflow should not be in events.lua but agent.lua or separate**
      - [ ] where is `workflow.serial`.. or other tyupes of workflows?
      - [ ] Add `workflow.create(steps)` for workflow definition
      - [ ] Add `workflow.run(workflow, input)` for execution
      - [ ] Add `workflow.parallel(steps)` for concurrent execution
      - [ ] Add `workflow.conditional(condition, then_step, else_step)` for branching
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/agent_test.go`)
    - [ ] Test agent lifecycle state transitions
    - [ ] Test multi-agent communication patterns
    - [ ] Test tool assignment and execution
    - [ ] Test conversation state management
    - [ ] Test agent cloning with modifications
    - [ ] Test delegation and collaboration
    - [ ] Test concurrent agent operations
    - [ ] Test workflow execution with branching
    - [ ] Test parallel step coordination
    - [ ] Test workflow cancellation
    - [ ] Test error handling in agent workflows

- [ ] **Task 2.3.5.4: State Management Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/state.lua`)
    - [ ] Context and state utilities
      - [ ] Add `state.create(initial_data)` for state creation
      - [ ] Add `state.merge(state1, state2)` for state composition
      - [ ] Add `state.snapshot(state)` for state capture
      - [ ] Add `state.restore(snapshot)` for state restoration
    - [ ] State persistence helpers
      - [ ] Add `state.save(state, key)` for persistent storage
      - [ ] Add `state.load(key, default)` for state retrieval
      - [ ] Add `state.expire(key, duration)` for TTL support
    - [ ] State transformation utilities
      - [ ] Add `state.transform(state, transformer)` for state modification
      - [ ] Add `state.filter(state, predicate)` for state filtering
      - [ ] Add `state.validate(state, schema)` for state validation
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/state_test.go`)
    - [ ] Test state persistence and retrieval
    - [ ] Test TTL expiration behavior
    - [ ] Test state merging conflict resolution
    - [ ] Test schema validation errors
    - [ ] Test concurrent state modifications
    - [ ] Test snapshot/restore consistency
    - [ ] Test state transformation chains
    - [ ] Benchmark state operations

- [ ] **Task 2.3.5.5: Event & Hooks Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/events.lua`)
    - [ ] Event system utilities
      - [ ] Add `events.emit(event, data)` for event emission
      - [ ] Add `events.on(event, handler)` for event subscription
      - [ ] Add `events.once(event, handler)` for one-time handlers
      - [ ] Add `events.off(event, handler)` for unsubscription
    - [ ] Hook and lifecycle utilities - **this might need to be separate in hooks.lua**
      - [ ] Add `hooks.before(event, handler)` for pre-hooks
      - [ ] Add `hooks.after(event, handler)` for post-hooks
      - [ ] Add `hooks.around(event, wrapper)` for around-hooks
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/events_test.go`)
    - [ ] Test event emission and subscription ordering
    - [ ] Test one-time handler cleanup
    - [ ] Test hook execution order (before/after/around)
    - [ ] Test event handler errors
    - [ ] Test memory leaks in event handlers

- [ ] **Task 2.3.5.6: Structured Data Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/data.lua`)
    - [ ] JSON and data processing utilities
      - [ ] Add `data.parse_json(text, schema)` for validated JSON parsing
      - [ ] Add `data.to_json(object, format)` for pretty JSON serialization
      - [ ] Add `data.extract_structured(text, schema)` for LLM output parsing
      - [ ] Add `data.convert_format(data, from_format, to_format)` for format conversion
    - [ ] Schema validation helpers
      - [ ] Add `data.validate(data, schema)` for schema validation
      - [ ] Add `data.infer_schema(data)` for schema generation
      - [ ] Add `data.migrate_schema(data, old_schema, new_schema)` for migration
    - [ ] Data transformation utilities
      - [ ] Add `data.map(collection, mapper)` for data mapping
      - [ ] Add `data.filter(collection, predicate)` for filtering
      - [ ] Add `data.reduce(collection, reducer, initial)` for aggregation
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/data_test.go`)
    - [ ] Test JSON parsing with invalid schemas
    - [ ] Test LLM output extraction accuracy
    - [ ] Test format conversion edge cases
    - [ ] Test schema inference from complex data
    - [ ] Test schema migration compatibility
    - [ ] Test transformation performance
    - [ ] Test concurrent data operations
    - [ ] Test memory usage with large datasets

- [ ] **Task 2.3.5.7: Tools & Registry Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/tools.lua`)
    - [ ] Tool registration and management
      - [ ] Add `tools.define(name, description, schema, func)` for tool creation
      - [ ] Add `tools.register_library(library)` for tool library loading
      - [ ] Add `tools.compose(tools)` for tool composition
    - [ ] Tool execution utilities - **is there a tools.async_execute?**
      - [ ] Add `tools.execute_safe(tool, params)` for safe execution
      - [ ] Add `tools.pipeline(tools, data)` for tool pipelines
      - [ ] Add `tools.parallel_execute(tools, params)` for concurrent execution
    - [ ] Tool validation and testing
      - [ ] Add `tools.validate_params(tool, params)` for parameter validation
      - [ ] Add `tools.test_tool(tool, test_cases)` for tool testing
      - [ ] Add `tools.benchmark_tool(tool, params)` for performance testing
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/tools_test.go`)
    - [ ] Test tool registration and discovery
    - [ ] Test parameter validation errors
    - [ ] Test tool composition behavior
    - [ ] Test pipeline execution order
    - [ ] Test parallel execution limits
    - [ ] Test tool error handling
    - [ ] Test tool benchmarking accuracy
    - [ ] Test memory leaks in tool chains

- [ ] **Task 2.3.5.8: Observability & Monitoring Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/observability.lua`)
    - [ ] Metrics and monitoring utilities
      - [ ] Add `metrics.counter(name, tags)` for counter metrics
      - [ ] Add `metrics.gauge(name, value, tags)` for gauge metrics
      - [ ] Add `metrics.timer(name, duration, tags)` for timing metrics
      - [ ] Add `metrics.track(func, name)` for automatic function tracking
    - [ ] Tracing and debugging helpers
      - [ ] Add `trace.span(name, func)` for traced execution
      - [ ] Add `trace.add_event(name, attributes)` for span events
      - [ ] Add `trace.set_attribute(key, value)` for span attributes
    - [ ] Guardrails and safety utilities
      - [ ] Add `safety.check_content(content, rules)` for content validation
      - [ ] Add `safety.rate_limit(key, limit, window)` for rate limiting
      - [ ] Add `safety.circuit_breaker(name, config)` for fault tolerance
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/observability_test.go`)
    - [ ] Test metric collection accuracy
    - [ ] Test trace span propagation
    - [ ] Test rate limiting behavior
    - [ ] Test circuit breaker state transitions
    - [ ] Test content validation rules
    - [ ] Test metric aggregation
    - [ ] Test performance overhead
    - [ ] Test concurrent metric updates

- [ ] **Task 2.3.5.9: Authentication & Security Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/auth.lua`)
    - [ ] Authentication utilities
      - [ ] Add `auth.login(credentials, scheme)` for authentication
      - [ ] Add `auth.refresh_token(refresh_token)` for token refresh
      - [ ] Add `auth.validate_session(session)` for session validation
    - [ ] OAuth and token management
      - [ ] Add `auth.oauth_flow(provider, config)` for OAuth flows
      - [ ] Add `auth.jwt_decode(token, verify)` for JWT handling
      - [ ] Add `auth.secure_store(key, value)` for secure storage
    - [ ] Permission and access control
      - [ ] Add `auth.check_permission(user, resource, action)` for access control
      - [ ] Add `auth.create_policy(rules)` for policy creation
      - [ ] Add `auth.audit_log(action, context)` for audit logging
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/auth_test.go`)
    - [ ] Test authentication schemes
    - [ ] Test token refresh logic
    - [ ] Test session validation
    - [ ] Test OAuth flow states
    - [ ] Test JWT verification
    - [ ] Test permission checks
    - [ ] Test secure storage encryption
    - [ ] Test audit logging completeness

- [ ] **Task 2.3.5.10: Error Handling & Recovery Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/errors.lua`)
    - [ ] Enhanced error handling
      - [ ] Add `errors.try(func, catch_func, finally_func)` for try-catch-finally
      - [ ] Add `errors.wrap(error, context)` for error wrapping
      - [ ] Add `errors.chain(errors)` for error chaining
    - [ ] Retry and recovery mechanisms
      - [ ] Add `errors.retry(func, options)` for retry logic
      - [ ] Add `errors.circuit_breaker(func, config)` for fault tolerance
      - [ ] Add `errors.fallback(primary, fallback)` for fallback strategies
    - [ ] Error categorization and reporting
      - [ ] Add `errors.categorize(error)` for error classification
      - [ ] Add `errors.report(error, context)` for error reporting
      - [ ] Add `errors.aggregate(errors)` for error aggregation
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/errors_test.go`)
    - [ ] Test try-catch-finally execution order
    - [ ] Test error wrapping context preservation
    - [ ] Test retry with backoff strategies
    - [ ] Test circuit breaker state machine
    - [ ] Test fallback chain behavior
    - [ ] Test error categorization accuracy
    - [ ] Test error aggregation patterns
    - [ ] Test memory leaks in error chains

- [ ] **Task 2.3.5.11: Logging & Debug Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/logging.lua`)
    - [ ] Unified logging interface
      - [ ] Add `log.info(message, context)` for info logging
      - [ ] Add `log.warn(message, context)` for warning logging
      - [ ] Add `log.error(message, context)` for error logging
      - [ ] Add `log.debug(message, context)` for debug logging
    - [ ] Structured logging utilities
      - [ ] Add `log.with_context(context)` for context propagation
      - [ ] Add `log.create_logger(component, level)` for component loggers
      - [ ] Add `log.set_formatter(formatter)` for custom formatting
    - [ ] Debug and diagnostics helpers
      - [ ] Add `debug.trace_calls(func)` for call tracing
      - [ ] Add `debug.memory_usage()` for memory monitoring
      - [ ] Add `debug.performance_profile(func)` for performance profiling
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/logging_test.go`)
    - [ ] Test log level filtering
    - [ ] Test context propagation
    - [ ] Test custom formatters
    - [ ] Test call tracing accuracy
    - [ ] Test memory usage reporting
    - [ ] Test performance profiling
    - [ ] Test concurrent logging
    - [ ] Test log rotation behavior

- [ ] **Task 2.3.5.12: Testing & Validation Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/testing.lua`)
    - [ ] Test framework and assertions
      - [ ] Add `test.describe(name, tests)` for test grouping
      - [ ] Add `test.it(name, test_func)` for individual tests
      - [ ] Add `test.assert_equals(actual, expected)` for assertions
      - [ ] Add `test.assert_error(func, expected_error)` for error testing
    - [ ] Mocking and stubbing utilities
      - [ ] Add `test.mock(object, method, replacement)` for mocking
      - [ ] Add `test.stub(func, return_value)` for stubbing
      - [ ] Add `test.spy(func)` for function spying
    - [ ] Performance and load testing
      - [ ] Add `test.benchmark(func, iterations)` for benchmarking
      - [ ] Add `test.load_test(func, config)` for load testing
      - [ ] Add `test.memory_test(func)` for memory testing
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/testing_test.go`)
    - [ ] Test assertion functionality
    - [ ] Test mock behavior
    - [ ] Test spy call tracking
    - [ ] Test benchmark accuracy
    - [ ] Test load test execution
    - [ ] Test memory leak detection
    - [ ] Test nested test groups
    - [ ] Test async test support

- [ ] **Task 2.3.5.13: Core Utilities Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/core.lua`)
    - [ ] String and text utilities
      - [ ] Add `string.template(template, variables)` for string templating
      - [ ] Add `string.slugify(text)` for URL-safe strings
      - [ ] Add `string.truncate(text, length)` for text truncation
    - [ ] Collection and data utilities
      - [ ] Add `table.merge(t1, t2)` for table merging
      - [ ] Add `table.deep_copy(table)` for deep copying
      - [ ] Add `table.keys(table)` and `table.values(table)` for extraction
    - [ ] UUID, hashing, and crypto utilities
      - [ ] Add `crypto.uuid()` for UUID generation
      - [ ] Add `crypto.hash(data, algorithm)` for hashing
      - [ ] Add `crypto.random_string(length)` for random strings
    - [ ] Time and date utilities
      - [ ] Add `time.now()` for current timestamp
      - [ ] Add `time.format(timestamp, format)` for time formatting
      - [ ] Add `time.duration(start, end)` for duration calculation
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/core_test.go`)
    - [ ] Test string templating edge cases
    - [ ] Test table deep copy with cycles
    - [ ] Test UUID uniqueness
    - [ ] Test hash algorithm support
    - [ ] Test time formatting locales
    - [ ] Test duration calculations
    - [ ] Test random string entropy
    - [ ] Test concurrent utility usage

- [ ] **Task 2.3.5.14: Spell Framework Library**
  - [ ] Implementation (`/pkg/engine/gopherlua/stdlib/spell.lua`)
    - [ ] Spell lifecycle and framework
      - [ ] Add `spell.init(config)` for spell initialization
      - [ ] Add `spell.params(name, default, type)` for parameter handling
      - [ ] Add `spell.output(data, format)` for result output
    - [ ] Spell composition and reuse
      - [ ] Add `spell.include(spell_path)` for spell inclusion
      - [ ] Add `spell.compose(spells)` for spell composition
      - [ ] Add `spell.library(name, functions)` for library creation
    - [ ] Spell execution context
      - [ ] Add `spell.context()` for execution context access
      - [ ] Add `spell.config(key, default)` for configuration access
      - [ ] Add `spell.cache(key, value, ttl)` for caching
  - [ ] Testing (`/pkg/engine/gopherlua/stdlib/spell_test.go`)
    - [ ] Test spell initialization
    - [ ] Test parameter validation
    - [ ] Test spell composition
    - [ ] Test context isolation
    - [ ] Test cache TTL behavior
    - [ ] Test output formatting
    - [ ] Test library loading
    - [ ] Test spell error handling

- [ ] **Task 2.3.5.15: Documentation & Examples** (`/pkg/engine/gopherlua/stdlib/`)
  - [ ] Comprehensive documentation
    - [ ] Create `README.md` with library overview and philosophy
    - [ ] Create `API_REFERENCE.md` with complete function documentation
    - [ ] Create `EXAMPLES.md` with practical usage examples

- [ ] **Task 2.3.5.16: Test Infrastructure**
  - [ ] Create test helpers (`/pkg/engine/gopherlua/stdlib/stdlib_test_helpers.go`)
    - [ ] Lua module loading helpers
    - [ ] Lua table comparison utilities
    - [ ] Async test utilities
    - [ ] Error assertion helpers
    - [ ] Mock bridge creation utilities
    - [ ] Test fixture management
  - [ ] Create async test helpers (`/pkg/engine/gopherlua/stdlib/async_test_helpers.go`)
    - [ ] Promise assertion utilities
    - [ ] Coroutine lifecycle helpers
    - [ ] Timeout testing utilities
    - [ ] Concurrent operation validators
    - [ ] Memory leak detectors

- [ ] **Task 2.3.5.17: Integration Testing**
  - [ ] Cross-module tests (`/pkg/engine/gopherlua/stdlib/integration_test.go`)
    - [ ] Test Promise + LLM async operations
    - [ ] Test Agent + State + Events coordination
    - [ ] Test Workflow + Tools integration
    - [ ] Test Error handling across modules
    - [ ] Test module loading dependencies
    - [ ] Test sandbox security with all modules
    - [ ] Test resource cleanup across modules
    - [ ] Test performance with all modules loaded

- [ ] **Task 2.3.5.18: Performance Testing**
  - [ ] Benchmark suite (`/pkg/engine/gopherlua/stdlib/benchmark_test.go`)
    - [ ] Promise creation/resolution benchmarks
    - [ ] Module loading time benchmarks
    - [ ] Memory usage profiling
    - [ ] Concurrent operation stress tests
    - [ ] Event system throughput tests
    - [ ] State management scalability tests
    - [ ] Tool execution performance tests
    - [ ] Generate performance report

#### Testing Requirements for All Lua Standard Library Modules:
1. **Minimum 90% test coverage** for all modules
2. **Table-driven tests** using testutils patterns
3. **Both success and failure paths** must be tested
4. **Timeout tests** for all async operations
5. **Memory leak tests** for resource management
6. **Sandbox restriction verification** for security
7. **Concurrent execution tests** for thread safety
8. **Performance benchmarks** for critical paths
9. **Integration tests** between dependent modules
10. **Documentation examples** must be executable tests
    - [ ] Create `BEST_PRACTICES.md` with performance and security guidelines
  - [ ] Interactive examples and tutorials
    - [ ] Create `examples/` directory with working examples for each library
    - [ ] Create `tutorials/` directory with step-by-step guides
    - [ ] Create `templates/` directory with spell templates
  - [ ] Integration guides
    - [ ] Create bridge integration examples showing stdlib + bridge usage
    - [ ] Create performance optimization guides
    - [ ] Create security configuration examples

### Phase 2.4: Advanced Features & Optimization

#### 2.4.1: Performance Optimization
- [ ] **Task 2.4.1.1: Profiling Infrastructure** (`/pkg/engine/gopherlua/profiling.go`)
  - [ ] Add execution time tracking
  - [ ] Implement memory profiling
  - [ ] Create allocation tracking
  - [ ] Add hot path identification
  - [ ] Implement profiling API

- [ ] **Task 2.4.1.2: Type Conversion Optimization**
  - [ ] Implement conversion caching
  - [ ] Add fast paths for common types
  - [ ] Optimize table traversal
  - [ ] Reduce allocation in hot paths
  - [ ] Add benchmarks for all conversions

- [ ] **Task 2.4.1.3: State Pool Optimization**
  - [ ] Implement predictive scaling
  - [ ] Optimize state reset process
  - [ ] Add state pre-warming
  - [ ] Implement memory pooling
  - [ ] Create performance metrics

- [ ] **Task 2.4.1.4: Script Compilation Optimization**
  - [ ] Enhance chunk caching
  - [ ] Add AST optimization
  - [ ] Implement dead code elimination
  - [ ] Add constant folding
  - [ ] Create compilation benchmarks

#### 2.4.2: Development Tools
- [ ] **Task 2.4.2.1: REPL Implementation** (`/cmd/llmspell-lua/main.go`)
  - [ ] Create interactive Lua REPL
  - [ ] Add command history
  - [ ] Implement auto-completion
  - [ ] Add syntax highlighting
  - [ ] Create help system

- [ ] **Task 2.4.2.2: Debugger Support** (`/pkg/engine/gopherlua/debug.go`)
  - [ ] Implement breakpoint support
  - [ ] Add step debugging
  - [ ] Create variable inspection
  - [ ] Implement stack trace visualization
  - [ ] Add watch expressions

- [ ] **Task 2.4.2.3: Script Validator** (`/pkg/engine/gopherlua/validator.go`)
  - [ ] Implement syntax validation
  - [ ] Add type checking where possible
  - [ ] Create linting rules
  - [ ] Implement security validation
  - [ ] Add performance warnings

- [ ] **Task 2.4.2.4: Documentation Generator** (`/pkg/engine/gopherlua/docs.go`)
  - [ ] Extract API from bridges
  - [ ] Generate Lua documentation
  - [ ] Create example extraction
  - [ ] Add type annotations
  - [ ] Generate completion data

#### 2.4.3: Production Readiness
- [ ] **Task 2.4.3.1: Comprehensive Testing**
  - [ ] Achieve 90%+ test coverage
  - [ ] Add integration test suite
  - [ ] Create stress tests
  - [ ] Implement chaos testing
  - [ ] Add regression test suite

- [ ] **Task 2.4.3.2: Error Handling Enhancement**
  - [ ] Standardize error types
  - [ ] Add error categorization
  - [ ] Implement error recovery
  - [ ] Create error reporting
  - [ ] Add error metrics

- [ ] **Task 2.4.3.3: Monitoring & Metrics**
  - [ ] Add Prometheus metrics
  - [ ] Implement health checks
  - [ ] Create performance dashboards
  - [ ] Add distributed tracing
  - [ ] Implement alerting rules

- [ ] **Task 2.4.3.4: Security Hardening**
  - [ ] Conduct security audit
  - [ ] Add input validation
  - [ ] Implement rate limiting
  - [ ] Create security benchmarks
  - [ ] Add CVE scanning

#### 2.4.4: Documentation & Examples
- [ ] **Task 2.4.4.1: User Guide** (`/docs/user-guide/lua/`)
  - [ ] Getting started with Lua spells
  - [ ] Complete API reference
  - [ ] Common patterns and idioms
  - [ ] Troubleshooting guide
  - [ ] Migration from pure Lua

- [ ] **Task 2.4.4.2: Example Spells** (`/examples/lua/`)
  - [ ] Basic LLM interaction
  - [ ] Agent with tools
  - [ ] Complex workflows
  - [ ] Event-driven spells
  - [ ] Performance patterns

- [ ] **Task 2.4.4.3: Developer Documentation**
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