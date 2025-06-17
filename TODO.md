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
  - Phase 2.3: Bridge Integration Layer 🔄 NEXT
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

#### 2.3.1: Module System Architecture ✅ COMPLETED [2025-06-19]

#### 2.3.2: Async/Coroutine Support
**Note**: This section moved before Bridge Adapters to provide async foundation for bridge operations.

- ✅ **Task 2.3.2.1: Async Runtime** (`/pkg/engine/gopherlua/async.go`) [COMPLETED - 2025-06-19]
  - ✅ Implement `AsyncRuntime` for coroutine management
  - ✅ Add promise-coroutine integration
  - ✅ Create async execution context
  - ✅ Implement cancellation support
  - ✅ Add timeout handling

- ✅ **Task 2.3.2.2: Channel Integration** (`/pkg/engine/gopherlua/channels.go`) [COMPLETED - 2025-06-19]
  - ✅ Implement Go channel ↔ LChannel bridge
  - ✅ Add select operation support
  - ✅ Create buffered channel support
  - ✅ Implement channel closing
  - ✅ Add deadlock detection

- [ ] **Task 2.3.2.0: ScriptValue Type System Refactoring** [CRITICAL - Foundation for all bridge operations]
  - ✅ **Phase 1: Define ScriptValue Types** (`/pkg/engine/value_types.go`) [COMPLETED - 2025-06-19]
    - ✅ Create ScriptValue interface with Type(), IsNil(), String(), ToGo(), Equals()
    - ✅ Define ScriptValueType enum (Nil, Bool, Number, String, Array, Object, Function, Error, Channel, Custom)
    - ✅ Implement concrete types: NilValue, BoolValue, NumberValue, StringValue
    - ✅ Implement collection types: ArrayValue, ObjectValue
    - ✅ Implement special types: FunctionValue, ErrorValue, ChannelValue
    - ✅ Add constructor functions: NewStringValue(), NewNumberValue(), etc.
  
  - ✅ **Phase 2: Update Core Interfaces** (`/pkg/engine/interface.go`) [COMPLETED - 2025-06-19]
    - ✅ Change ToNative(interface{}) to ToNative(ScriptValue) 
    - ✅ Change FromNative return to (ScriptValue, error)
    - ✅ Update Bridge.ValidateMethod to use []ScriptValue
    - ✅ Update Bridge.ExecuteMethod to use ScriptValue params/returns
    - [ ] Update Execute methods to use ScriptValue in params map
  
  - ✅ **Phase 3: Update TypeConverter** [COMPLETED - 2025-06-19]
    - ✅ Changed Convert() to accept and return ScriptValue
    - ✅ Updated TypeMapping definitions
    - ✅ Added ScriptValue-aware conversion functions
  
  - [ ] **Phase 4: Update Bridge Package** [IN PROGRESS - 2025-06-19]
    - [ ] Update all bridge implementations to use ScriptValue (no backward compatibility needed)
    - [ ] Replace []interface{} with []ScriptValue in method args
    - [ ] Convert return values to appropriate ScriptValue types
    - [ ] Update type mappings for each bridge
    - ✅ ModelInfoBridge - Updated ValidateMethod and ExecuteMethod
    - ✅ SchemaBridge - Already had updated signatures 
    - ✅ GuardrailsBridge - Updated ValidateMethod, added ExecuteMethod, updated all methods
    - ✅ MetricsBridge - ValidateMethod and ExecuteMethod updated, all methods converted
    - ✅ TracingBridge - ValidateMethod and ExecuteMethod updated, all methods converted
    - [ ] Agent package bridges (6 bridges)
      - ✅ agent.go - Already updated signatures
      - [ ] tools.go - Partially updated, needs method implementation fixes
      - [ ] hooks.go - Needs update
      - [ ] events.go - Needs update
      - [ ] workflow.go - Needs update
      - [ ] tool_registry.go - Needs update
    - [ ] LLM package bridges (3 bridges) - Need updates
    - [ ] State package bridges (2 bridges) - Need updates
    - [ ] Util package bridges (8 bridges) - Need updates
  
  - [ ] **Phase 5: Update GopherLua Engine**
    - [ ] Create LValueToScriptValue(lua.LValue) ScriptValue converter
    - [ ] Create ScriptValueToLValue(ScriptValue) lua.LValue converter
    - [ ] Update existing converter.go to use ScriptValue internally
    - [ ] Maintain circular reference detection
    - [ ] Update caching to work with ScriptValue
    - [ ] Refactor engine.go ToNative/FromNative to use ScriptValue
    - [ ] Update Execute to convert map[string]interface{} to map[string]ScriptValue
    - [ ] Update ExecuteFile to use ScriptValue
    - [ ] Update ExecuteScript to return ScriptValue in ExecutionResult
    - [ ] Fix engine_bridge.go to use ScriptValue for method calls
    - [ ] Update bridge_adapter.go to use ScriptValue throughout
    - [ ] Fix convertArgs to work with ScriptValue
    - [ ] Update all method wrappers to handle ScriptValue
    - [ ] Ensure error propagation with ErrorValue
  
  - [ ] **Phase 6: Update Other Engines**
    - [ ] Update JavaScript engine when implemented
    - [ ] Update Tengo engine when implemented
    - [ ] Ensure consistency across all engines
  
  - [ ] **Phase 7: Update Tests**
    - [ ] Update all test files to use ScriptValue
    - [ ] Fix mock implementations in tests
    - [ ] Update example scripts to demonstrate new type system
    - [ ] Add comprehensive ScriptValue conversion tests
  
  - [ ] **Phase 8: Cleanup and Documentation**
    - [ ] Document ScriptValue type system
    - [ ] Create migration guide for bridge authors
    - [ ] Update architecture documentation
    - [ ] Add performance benchmarks comparing old vs new
    - [ ] Remove old interface{} based code where possible

- ✅ **Task 2.3.2.3: Async Bridge Methods** (`/pkg/engine/gopherlua/async_bridges.go`) [COMPLETED - 2025-06-19]
  - ✅ Wrap bridge methods for async execution
  - ✅ Add automatic promisification
  - ✅ Implement streaming support
  - ✅ Add progress callbacks
  - ✅ Create cancellation tokens

- ✅ **Task 2.3.2.4: Async Testing** (`/pkg/engine/gopherlua/async_test.go`) [COMPLETED - 2025-06-19]
  - ✅ Test coroutine lifecycle
  - ✅ Test promise integration
  - ✅ Test channel operations
  - ✅ Test cancellation and timeouts
  - ✅ Test concurrent async operations

#### 2.3.3: Bridge Adapters
- ✅ **Task 2.3.3.1: Bridge Adapter Base** (`/pkg/engine/gopherlua/bridge_adapter.go`) [COMPLETED - 2025-06-19]
  - ✅ Defined `BridgeAdapter` struct with engine.Bridge wrapping
  - ✅ Implemented base adapter with common functionality
  - ✅ Added method discovery and wrapping
  - ✅ Created error handling standards
  - ✅ Implemented type conversion integration

- [ ] **Task 2.3.3.2: LLM and Provider Bridge Adapter** (`/pkg/engine/gopherlua/adapters/llm.go`)
  - [ ] Create LLM module with basic generation methods
    - [ ] Implement `generate(prompt, options)` method
    - [ ] Implement `generateMessage(messages, options)` method  
    - [ ] Add streaming support with `stream(prompt, options)`
    - [ ] Add token counting utilities
  - [ ] Integrate Provider Registry functionality (from providers.go)
    - [ ] Add `createProvider(type, name, config)` method
    - [ ] Add `getProvider(name)` method
    - [ ] Add `listProviders()` method
    - [ ] Add provider template support
    - [ ] Add multi-provider support with strategies
  - [ ] Integrate Provider Pool functionality (from pool.go)
    - [ ] Add `createPool(name, providers, strategy)` method
    - [ ] Add pool health monitoring methods
    - [ ] Add `generateWithPool(poolName, prompt, options)` method
    - [ ] Add object pool utilities for optimization
  - [ ] Implement model selection and info
    - [ ] Add `listModels()` method
    - [ ] Add `getModelInfo(modelName)` method
    - [ ] Add model capability checking

- [ ] **Task 2.3.3.3: State Bridge Adapter** (`/pkg/engine/gopherlua/adapters/state.go`)
  - [ ] Create state and context management module
  - [ ] Implement get/set operations
  - [ ] Add transform functions
  - [ ] Implement persistence methods
  - [ ] Add state merging capabilities

- [ ] **Task 2.3.3.4: Events Bridge Adapter** (`/pkg/engine/gopherlua/adapters/events.go`)
  - [ ] Create event module
  - [ ] Implement event subscription
  - [ ] Add event emission
  - [ ] Implement filtering
  - [ ] Add event correlation

- [ ] **Task 2.3.3.5: Structure Bridge Adapter** (`/pkg/engine/gopherlua/adapters/structured.go`)
  - [ ] Create structured output module
    - [ ] Implement JSON schema validation
    - [ ] Add structured generation methods
    - [ ] Implement response parsing
    - [ ] Add schema registry support
  - [ ] Implement structured tools
    - [ ] Add tool schema definitions
    - [ ] Implement tool parameter validation
    - [ ] Add tool result parsing
  - [ ] Add structured streaming support
    - [ ] Implement structured stream parsing
    - [ ] Add partial object assembly
    - [ ] Handle incomplete structured data

- [ ] **Task 2.3.3.6: Agent Bridge Adapter** (`/pkg/engine/gopherlua/adapters/agent.go`)
  - [ ] Create agent module with agent lifecycle
    - [ ] Implement `createAgent(name, provider, options)` method
    - [ ] Add `getAgent(name)` method
    - [ ] Add `listAgents()` method
    - [ ] Implement agent configuration methods
  - [ ] Implement agent communication
    - [ ] Add `agent:complete(prompt, options)` method
    - [ ] Add `agent:generateMessage(messages, options)` method
    - [ ] Add `agent:stream(prompt, options)` method
  - [ ] Add agent tool integration
    - [ ] Implement `agent:addTool(tool)` method
    - [ ] Add `agent:removeTool(toolName)` method
    - [ ] Add `agent:listTools()` method
    - [ ] Implement tool execution within agent context
  - [ ] Implement agent state management
    - [ ] Add `agent:getState()` method
    - [ ] Add `agent:setState(state)` method
    - [ ] Implement state persistence for agents

- [ ] **Task 2.3.3.7: Hooks Bridge Adapter** (`/pkg/engine/gopherlua/adapters/hooks.go`)
  - [ ] Create hooks module for lifecycle events
    - [ ] Implement `registerHook(event, callback)` method
    - [ ] Add `unregisterHook(event, hookId)` method
    - [ ] Add `listHooks(event)` method
    - [ ] Implement hook priority system
  - [ ] Add pre/post generation hooks
    - [ ] Implement `beforeGenerate` hook
    - [ ] Implement `afterGenerate` hook
    - [ ] Add request/response modification support
    - [ ] Implement hook context passing
  - [ ] Add streaming hooks
    - [ ] Implement `onStreamStart` hook
    - [ ] Implement `onStreamChunk` hook
    - [ ] Implement `onStreamEnd` hook
    - [ ] Add stream modification support
  - [ ] Add error and retry hooks
    - [ ] Implement `onError` hook
    - [ ] Implement `beforeRetry` hook
    - [ ] Add error recovery strategies
    - [ ] Implement custom retry logic

- [ ] **Task 2.3.3.8: Workflow Bridge Adapter** (`/pkg/engine/gopherlua/adapters/workflow.go`)
  - [ ] Create workflow module
  - [ ] Implement workflow builders
  - [ ] Add step definitions
  - [ ] Implement execution methods
  - [ ] Add state passing between steps

- [ ] **Task 2.3.3.9: Tools Bridge Adapter** (`/pkg/engine/gopherlua/adapters/tools.go`)
  - [ ] Create tools module
  - [ ] Implement tool registration
  - [ ] Add tool execution
  - [ ] Implement parameter validation
  - [ ] Add custom tool support

- [ ] **Task 2.3.3.10: Observability Bridge Adapters** (`/pkg/engine/gopherlua/adapters/observability.go`)
  - [ ] Implement Guardrails Bridge Adapter
    - [ ] Add `enableGuardrails(config)` method for safety system configuration
    - [ ] Add `validateContent(content, type)` method for content filtering
    - [ ] Add `addBehavioralConstraint(constraint)` method for behavioral limits
    - [ ] Add `checkCompliance(request)` method for compliance validation
  - [ ] Implement Metrics Bridge Adapter
    - [ ] Add `createCounter(name, labels)` method for counter metrics
    - [ ] Add `createGauge(name, labels)` method for gauge metrics
    - [ ] Add `createTimer(name, labels)` method for timing metrics
    - [ ] Add `recordMetric(name, value, labels)` method for metric recording
    - [ ] Add `getMetrics()` method for metric aggregation
  - [ ] Implement Tracing Bridge Adapter
    - [ ] Add `startSpan(name, options)` method for trace span creation
    - [ ] Add `addSpanEvent(span, name, attributes)` method for span events
    - [ ] Add `setSpanAttribute(span, key, value)` method for span attributes
    - [ ] Add `endSpan(span)` method for span completion
    - [ ] Add OpenTelemetry-compatible interface

- [ ] **Task 2.3.3.11: Schema Bridge Adapter** (`/pkg/engine/gopherlua/adapters/schema.go`)
  - [ ] Create schema validation module
    - [ ] Add `validateJSON(data, schema)` method for JSON schema validation
    - [ ] Add `generateSchema(data, options)` method for schema generation
    - [ ] Add `registerSchema(name, schema)` method for schema registration
    - [ ] Add `getSchema(name)` method for schema retrieval
  - [ ] Implement structured tools support
    - [ ] Add `validateStructuredOutput(output, schema)` method
    - [ ] Add `parseStructuredResponse(response, schema)` method
    - [ ] Add schema-based tool parameter validation
  - [ ] Add schema versioning and migration
    - [ ] Add `migrateSchema(oldSchema, newSchema)` method
    - [ ] Add `versionSchema(schema, version)` method
    - [ ] Add backward compatibility checking

- [ ] **Task 2.3.3.12: ModelInfo Bridge Adapter** (`/pkg/engine/gopherlua/adapters/modelinfo.go`)
  - [ ] Create model discovery module
    - [ ] Add `registerModelRegistry(name, registry)` method for registry management
    - [ ] Add `listModels()` method for listing all available models
    - [ ] Add `listModelsByRegistry(registryName)` method for registry-specific models
    - [ ] Add `getModel(modelId)` method for specific model retrieval
    - [ ] Add `listRegistries()` method for registry enumeration
  - [ ] Implement model capability queries
    - [ ] Add `getModelCapabilities(modelId)` method for capability discovery
    - [ ] Add `findModelsByCapability(capability)` method for capability-based search
    - [ ] Add model metadata access methods
  - [ ] Add model selection helpers
    - [ ] Add `suggestModel(requirements)` method for recommendation
    - [ ] Add `compareModels(modelIds)` method for model comparison
    - [ ] Add cost and performance estimation

- [ ] **Task 2.3.3.13: Utility Bridge Adapters** (`/pkg/engine/gopherlua/adapters/utils.go`)
  - [ ] Implement Auth Bridge Adapter
    - [ ] Add `authenticate(credentials, scheme)` method for authentication
    - [ ] Add `validateToken(token, options)` method for token validation
    - [ ] Add `refreshToken(refreshToken)` method for token refresh
    - [ ] Add OAuth2 flow support methods
  - [ ] Implement Debug Bridge Adapter
    - [ ] Add `setDebugLevel(component, level)` method for debug control
    - [ ] Add `debugLog(component, message, data)` method for debug logging
    - [ ] Add `getDebugConfig()` method for configuration retrieval
    - [ ] Add environment-based debug configuration
  - [ ] Implement Errors Bridge Adapter
    - [ ] Add `createError(message, code, category)` method for error creation
    - [ ] Add `wrapError(error, context)` method for error wrapping
    - [ ] Add `aggregateErrors(errors)` method for error aggregation
    - [ ] Add `categorizeError(error)` method for error categorization
    - [ ] Add error recovery strategy support
  - [ ] Implement JSON Bridge Adapter
    - [ ] Add `parseJSON(text, options)` method for JSON parsing
    - [ ] Add `toJSON(data, options)` method for JSON serialization
    - [ ] Add `validateJSONSchema(data, schema)` method for validation
    - [ ] Add `extractStructuredData(text, schema)` method for LLM output parsing
    - [ ] Add format conversion support (JSON/YAML/XML)
  - [ ] Implement LLM Utils Bridge Adapter
    - [ ] Add `createProvider(type, config)` method for provider creation
    - [ ] Add `generateTyped(prompt, schema, options)` method for typed generation
    - [ ] Add `getModelCapabilities(model)` method for capability queries
    - [ ] Add `trackCost(operation, tokens, model)` method for cost tracking
    - [ ] Add streaming with event support
  - [ ] Implement Script Logger Bridge Adapter
    - [ ] Add `createLogger(component, config)` method for logger creation
    - [ ] Add `log(level, message, context)` method for unified logging
    - [ ] Add `setLogLevel(component, level)` method for level control
    - [ ] Add context propagation support
  - [ ] Implement Slog Bridge Adapter
    - [ ] Add `info(message, fields)` method for info logging
    - [ ] Add `warn(message, fields)` method for warning logging
    - [ ] Add `error(message, fields)` method for error logging
    - [ ] Add `debug(message, fields)` method for debug logging
    - [ ] Add emoji enhancement and structured logging hooks
  - [ ] Implement General Util Bridge Adapter
    - [ ] Add `generateUUID()` method for UUID generation
    - [ ] Add `hash(data, algorithm)` method for hashing
    - [ ] Add `retry(operation, options)` method for retry logic
    - [ ] Add `sleep(duration)` method for delays
    - [ ] Add string and time utilities

- [ ] **Task 2.3.3.14: Adapter Testing** (`/pkg/engine/gopherlua/adapters/adapters_test.go`)
  - [ ] Test each adapter functionality
  - [ ] Test cross-adapter interaction
  - [ ] Test error propagation
  - [ ] Test type conversions



#### 2.3.4: Lua Standard Library
Based on comprehensive research of all bridge adapters, these feature-oriented modules provide script-friendly APIs for complex operations.

- [ ] **Task 2.3.4.1: Promise & Async Library** (`/pkg/engine/gopherlua/stdlib/promise.lua`)
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

- [ ] **Task 2.3.4.2: LLM Operations Library** (`/pkg/engine/gopherlua/stdlib/llm.lua`)
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

- [ ] **Task 2.3.4.3: Agent Management Library** (`/pkg/engine/gopherlua/stdlib/agent.lua`)
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

- [ ] **Task 2.3.4.4: State Management Library** (`/pkg/engine/gopherlua/stdlib/state.lua`)
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

- [ ] **Task 2.3.4.5: Event & Workflow Library** (`/pkg/engine/gopherlua/stdlib/events.lua`)
  - [ ] Event system utilities
    - [ ] Add `events.emit(event, data)` for event emission
    - [ ] Add `events.on(event, handler)` for event subscription
    - [ ] Add `events.once(event, handler)` for one-time handlers
    - [ ] Add `events.off(event, handler)` for unsubscription
  - [ ] Workflow orchestration helpers
    - [ ] Add `workflow.create(steps)` for workflow definition
    - [ ] Add `workflow.run(workflow, input)` for execution
    - [ ] Add `workflow.parallel(steps)` for concurrent execution
    - [ ] Add `workflow.conditional(condition, then_step, else_step)` for branching
  - [ ] Hook and lifecycle utilities
    - [ ] Add `hooks.before(event, handler)` for pre-hooks
    - [ ] Add `hooks.after(event, handler)` for post-hooks
    - [ ] Add `hooks.around(event, wrapper)` for around-hooks

- [ ] **Task 2.3.4.6: Structured Data Library** (`/pkg/engine/gopherlua/stdlib/data.lua`)
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

- [ ] **Task 2.3.4.7: Tools & Registry Library** (`/pkg/engine/gopherlua/stdlib/tools.lua`)
  - [ ] Tool registration and management
    - [ ] Add `tools.define(name, description, schema, func)` for tool creation
    - [ ] Add `tools.register_library(library)` for tool library loading
    - [ ] Add `tools.compose(tools)` for tool composition
  - [ ] Tool execution utilities
    - [ ] Add `tools.execute_safe(tool, params)` for safe execution
    - [ ] Add `tools.pipeline(tools, data)` for tool pipelines
    - [ ] Add `tools.parallel_execute(tools, params)` for concurrent execution
  - [ ] Tool validation and testing
    - [ ] Add `tools.validate_params(tool, params)` for parameter validation
    - [ ] Add `tools.test_tool(tool, test_cases)` for tool testing
    - [ ] Add `tools.benchmark_tool(tool, params)` for performance testing

- [ ] **Task 2.3.4.8: Observability & Monitoring Library** (`/pkg/engine/gopherlua/stdlib/observability.lua`)
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

- [ ] **Task 2.3.4.9: Authentication & Security Library** (`/pkg/engine/gopherlua/stdlib/auth.lua`)
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

- [ ] **Task 2.3.4.10: Error Handling & Recovery Library** (`/pkg/engine/gopherlua/stdlib/errors.lua`)
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

- [ ] **Task 2.3.4.11: Logging & Debug Library** (`/pkg/engine/gopherlua/stdlib/logging.lua`)
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

- [ ] **Task 2.3.4.12: Testing & Validation Library** (`/pkg/engine/gopherlua/stdlib/testing.lua`)
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

- [ ] **Task 2.3.4.13: Core Utilities Library** (`/pkg/engine/gopherlua/stdlib/core.lua`)
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

- [ ] **Task 2.3.4.14: Spell Framework Library** (`/pkg/engine/gopherlua/stdlib/spell.lua`)
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

- [ ] **Task 2.3.4.15: Documentation & Examples** (`/pkg/engine/gopherlua/stdlib/`)
  - [ ] Comprehensive documentation
    - [ ] Create `README.md` with library overview and philosophy
    - [ ] Create `API_REFERENCE.md` with complete function documentation
    - [ ] Create `EXAMPLES.md` with practical usage examples
    - [ ] Create `BEST_PRACTICES.md` with performance and security guidelines
  - [ ] Interactive examples and tutorials
    - [ ] Create `examples/` directory with working examples for each library
    - [ ] Create `tutorials/` directory with step-by-step guides
    - [ ] Create `templates/` directory with spell templates
  - [ ] Integration guides
    - [ ] Create bridge integration examples showing stdlib + bridge usage
    - [ ] Create performance optimization guides
    - [ ] Create security configuration examples

#### 2.3.4: Async/Coroutine Support
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