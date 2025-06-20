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
    - ✅ 2.3.3: Bridge Adapters [COMPLETED - 2025-06-19]
    - ✅ 2.3.4: Async/Coroutine Support [COMPLETED - 2025-06-19]
    - 🚧 2.3.5: Lua Standard Library [NOT STARTED]
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
✅ **COMPLETED [2025-06-19]** - All 24 tasks completed (Tasks 1-14 already in TODO-DONE.md)
**See TODO-DONE.md for complete task details and implementation history**

#### 2.3.4: Async/Coroutine Support
✅ **COMPLETED [2025-06-19]** - All 4 tasks completed. See TODO-DONE.md for implementation details.

#### 2.3.5: Lua Standard Library
Based on comprehensive research of all bridge adapters, these feature-oriented modules provide script-friendly APIs for complex operations. Each module requires comprehensive Go-based testing. **Progress: 12/18 tasks complete**

- [x] **Task 2.3.5.1: Lua stdlib - Promise & Async Library** ✅ COMPLETED [2025-06-19]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/promise.lua`)
    - [x] Implement Promise class with full async support
      - [x] Add `Promise.new(executor)` constructor
      - [x] Add `andThen/onError/onFinally` chain methods (renamed to avoid Lua keywords)
      - [x] Add `Promise.all(promises)` for concurrent execution
      - [x] Add `Promise.race(promises)` for first-wins scenarios
      - [x] Add `Promise.resolve(value)` and `Promise.reject(error)` helpers
    - [x] Add async/await syntax sugar
      - [x] Add `async(func)` wrapper for promise-returning functions
      - [x] Add `await(promise, timeout)` method with timeout support
      - [x] Add `sleep(duration)` utility for delays
    - [x] Add coroutine integration
      - [x] Add `spawn(func, args)` for concurrent execution
      - [x] Add `yield()` for cooperative multitasking
      - [x] Add channel-based communication helpers
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/promise_test.go`)
    - [x] Test promise constructor and executor behavior
    - [x] Test promise resolution/rejection with various types
    - [x] Test promise chaining (andThen/onError/onFinally)
    - [x] Test Promise.all concurrent execution
    - [x] Test Promise.race timing behavior
    - [x] Test timeout and cancellation
    - [x] Test error propagation through chains
    - [x] Test memory leaks in long chains
    - [x] Test coroutine integration
    - [x] Benchmark promise creation/resolution

- [x] **Task 2.3.5.2: Lua stdlib - LLM Operations Library** ✅ COMPLETED [2025-06-19]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/llm.lua`)
    - [x] High-level LLM operation helpers
      - [x] Add `llm.quick_prompt(prompt, options)` for simple prompting
      - [x] Add `llm.chat_session(system_prompt)` for conversation management
      - [x] Add `llm.streaming_response(prompt, callback)` for streaming
      - [x] Add `llm.batch_process(prompts, options)` for bulk operations
    - [x] Provider management utilities
      - [x] Add `llm.use_provider(name, config)` for easy provider switching
      - [x] Add `llm.compare_providers(prompt, providers)` for A/B testing
      - [x] Add `llm.fallback_chain(providers, prompt)` for reliability
    - [x] Model discovery helpers
      - [x] Add `llm.find_model(requirements)` for capability-based selection
      - [x] Add `llm.model_info(model_id)` for metadata access
      - [x] Add `llm.cost_estimate(operation, model)` for cost tracking
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/llm_test.go`)
    - [x] Test with mock LLM bridge
    - [x] Test streaming callbacks
    - [x] Test batch processing limits
    - [x] Test provider fallback chain
    - [x] Test cost estimation accuracy
    - [x] Test async operations with promises
    - [x] Test error handling and retries
    - [x] Test concurrent batch operations

- [x] **Task 2.3.5.3: Lua stdlib - Agent Management Library** ✅ COMPLETED [2025-06-19]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/agent.lua`)
    - [x] Added `agent.run(agent_id, input, options)` and `agent.run_async()` methods
    - [x] Agent lifecycle management
      - [x] Add `agent.create(name, config)` for agent creation
      - [x] Add `agent.configure(agent, settings)` for configuration
      - [x] Add `agent.clone(agent, modifications)` for agent templating
    - [x] Agent communication helpers
      - [x] Add `agent.conversation(agent, messages)` for multi-turn chat
      - [x] Add `agent.delegate(from_agent, to_agent, task)` for task delegation
      - [x] Add `agent.collaborate(agents, task)` for multi-agent workflows
    - [x] Agent tool integration
      - [x] Add `agent.add_tools(agent, tools)` for tool assignment
      - [x] Add `agent.create_tool(name, func, schema)` for custom tools
      - [x] Add `agent.tool_chain(tools, data)` for tool pipelines
    - [x] Workflow orchestration helpers (separate workflow bridge integration)
      - [x] Add `agent.workflow_create(name, steps)` for workflow definition
      - [x] Add `agent.workflow_run(workflow_id, input)` for execution
      - [x] Add `agent.workflow_parallel(steps, input)` for concurrent execution
      - [x] Add `agent.workflow_conditional(condition, then_step, else_step, input)` for branching
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/agent_test.go`)
    - [x] Test agent lifecycle state transitions
    - [x] Test multi-agent communication patterns
    - [x] Test tool assignment and execution
    - [x] Test conversation state management
    - [x] Test agent cloning with modifications
    - [x] Test delegation and collaboration
    - [x] Test concurrent agent operations
    - [x] Test workflow execution with branching
    - [x] Test parallel step coordination
    - [x] Test workflow cancellation
    - [x] Test error handling in agent workflows

- [x] **Task 2.3.5.4: Lua stdlib - State Management Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/state.lua`)
    - [x] Context and state utilities
      - [x] Add `state.create(initial_data)` for state creation
      - [x] Add `state.merge(state1, state2)` for state composition
      - [x] Add `state.snapshot(state)` for state capture
      - [x] Add `state.restore(snapshot)` for state restoration
    - [x] State persistence helpers
      - [x] Add `state.save(state, key)` for persistent storage
      - [x] Add `state.load(key, default)` for state retrieval
      - [x] Add `state.expire(key, duration)` for TTL support
    - [x] State transformation utilities
      - [x] Add `state.transform(state, transformer)` for state modification
      - [x] Add `state.filter(state, predicate)` for state filtering
      - [x] Add `state.validate(state, schema)` for state validation
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/state_test.go`)
    - [x] Test state persistence and retrieval
    - [x] Test TTL expiration behavior
    - [x] Test state merging conflict resolution
    - [x] Test schema validation errors
    - [x] Test concurrent state modifications
    - [x] Test snapshot/restore consistency
    - [x] Test state transformation chains
    - [x] Benchmark state operations
  - [x] Fixed mock method implementations for Lua colon syntax
    - [x] Updated all bridge mock methods to handle implicit self parameter
    - [x] Fixed promise constructor usage in expire function

- [x] **Task 2.3.5.5: Lua stdlib - Event & Hooks Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/events.lua`)
    - [x] Event system utilities
      - [x] Add `events.emit(event, data)` for event emission
      - [x] Add `events.on(event, handler)` for event subscription
      - [x] Add `events.once(event, handler)` for one-time handlers
      - [x] Add `events.off(event, handler)` for unsubscription
      - [x] Add `events.create_emitter()` for custom emitters
      - [x] Add `events.wait_for(event, timeout)` for promise-based waiting
      - [x] Add `events.aggregate(events, timeout)` for event collection
      - [x] Add `events.filter(pattern, handler)` for pattern matching
      - [x] Add `events.namespace(name)` for namespaced events
    - [x] Hook and lifecycle utilities
      - [x] Add `hooks.before(event, handler)` for pre-hooks
      - [x] Add `hooks.after(event, handler)` for post-hooks
      - [x] Add `hooks.around(event, wrapper)` for around-hooks
      - [x] Add `hooks.execute(event, fn, args)` for hook execution
      - [x] Add hook removal and clearing utilities
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/events_test.go`)
    - [x] Test event emission and subscription ordering
    - [x] Test one-time handler cleanup
    - [x] Test hook execution order (before/after/around)
    - [x] Test event handler errors
    - [x] Test memory leaks in event handlers
    - [x] Test advanced features (waiting, aggregation, filtering)
    - [x] Test performance benchmarks
  - [x] Fixed async execution issues in promise integration
    - [x] Resolved TestEventAggregation failure with improved promise handling
    - [x] Fixed TestEventWaitFor by removing problematic async timeouts
    - [x] Updated concurrent event handling test for Lua thread safety
    - [x] Fixed all Lua linter warnings (unused variables, shadowing, static methods)

- [x] **Task 2.3.5.6: Lua stdlib - Structured Data Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/data.lua`)
    - [x] JSON and data processing utilities
      - [x] Add `data.parse_json(text, schema)` for validated JSON parsing
      - [x] Add `data.to_json(object, format)` for pretty JSON serialization
      - [x] Add `data.extract_structured(text, schema)` for LLM output parsing
      - [x] Add `data.convert_format(data, from_format, to_format)` for format conversion
    - [x] Schema validation helpers
      - [x] Add `data.validate(data, schema)` for schema validation
      - [x] Add `data.infer_schema(data)` for schema generation
      - [x] Add `data.migrate_schema(data, old_schema, new_schema)` for migration
    - [x] Data transformation utilities
      - [x] Add `data.map(collection, mapper)` for data mapping
      - [x] Add `data.filter(collection, predicate)` for filtering
      - [x] Add `data.reduce(collection, reducer, initial)` for aggregation
      - [x] Add `data.clone(obj)` for deep cloning
      - [x] Add `data.merge(obj1, obj2)` for deep merging
      - [x] Add `data.get_path(obj, path)` and `data.set_path(obj, path, value)` for nested access
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/data_test.go`)
    - [x] Test JSON parsing and formatting operations
    - [x] Test schema validation and inference
    - [x] Test data transformation operations (map, filter, reduce)
    - [x] Test utility functions (clone, merge, path operations)
    - [x] Test format conversion functionality
    - [x] Test comprehensive error handling
    - [x] Test complex data processing pipelines
    - [x] Performance benchmarks for key operations

- [x] **Task 2.3.5.7: Lua stdlib - Tools & Registry Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/tools.lua`)
    - [x] Tool registration and management
      - [x] Add `tools.define(name, description, schema, func)` for tool creation
      - [x] Add `tools.register_library(library)` for tool library loading
      - [x] Add `tools.compose(tools)` for tool composition
    - [x] Tool execution utilities
      - [x] Add `tools.execute_safe(tool, params)` for safe execution with error handling
      - [x] Add `tools.pipeline(tools, data)` for tool pipelines
      - [x] Add `tools.parallel_execute(tools, params)` for concurrent execution
    - [x] Tool validation and testing
      - [x] Add `tools.validate_params(tool, params)` for parameter validation
      - [x] Add `tools.test_tool(tool, test_cases)` for tool testing
      - [x] Add `tools.benchmark_tool(tool, params)` for performance testing
    - [x] Tool discovery and information
      - [x] Add `tools.list()` for tool listing
      - [x] Add `tools.search(query)` for tool search
      - [x] Add `tools.get_info(name)` for tool information
      - [x] Add `tools.get_metrics(name)` for tool metrics
      - [x] Add `tools.get_history(limit)` for execution history
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/tools_test.go`)
    - [x] Test tool registration and discovery
    - [x] Test parameter validation errors
    - [x] Test tool composition behavior (pipeline, parallel, conditional)
    - [x] Test pipeline execution order
    - [x] Test parallel execution limits
    - [x] Test tool error handling
    - [x] Test tool benchmarking accuracy
    - [x] Test comprehensive bridge integration with mocks

- [x] **Task 2.3.5.8: Lua stdlib - Observability & Monitoring Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/observability.lua`)
    - [x] Metrics and monitoring utilities
      - [x] Add `observability.counter(name, description, tags)` for counter metrics
      - [x] Add `observability.gauge(name, description, tags)` for gauge metrics
      - [x] Add `observability.timer(name, description, tags)` for timing metrics
      - [x] Add `observability.ratio_counter(name, description, tags)` for ratio tracking
      - [x] Add `observability.track(func, name, options)` for automatic function tracking
    - [x] Tracing and debugging helpers
      - [x] Add `observability.start_span(name, options)` for traced execution
      - [x] Add `observability.trace(func, span_name, options)` for function tracing
      - [x] Add span methods for events, attributes, status, and error recording
    - [x] Structured logging utilities
      - [x] Add `observability.logger(name, options)` for custom loggers
      - [x] Add `observability.debug/info/warn/error(message, data)` for logging
      - [x] Add contextual logging with logger.with_context()
    - [x] Health monitoring and safety utilities
      - [x] Add `observability.health_check(name, check_func, options)` for health checks
      - [x] Add `observability.monitor_events(pattern, handler, options)` for event monitoring
      - [x] Add `observability.guardrail(name, validation_func, options)` for safety validation
    - [x] Performance monitoring
      - [x] Add comprehensive function tracking with metrics and tracing integration
      - [x] Add execution time measurement, error tracking, and metrics collection
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/observability_test.go`)
    - [x] Test metric collection accuracy (counters, gauges, timers, ratios)
    - [x] Test trace span propagation and lifecycle management
    - [x] Test performance monitoring and function tracking
    - [x] Test structured logging with custom loggers and context
    - [x] Test health checks for healthy and unhealthy scenarios
    - [x] Test event monitoring with pattern matching
    - [x] Test guardrail validation (both bridge-based and local fallback)
    - [x] Test error handling for all operations
    - [x] Test comprehensive integration scenarios with all bridges
    - [x] Test utility functions and system information retrieval

- [x] **Task 2.3.5.9: Lua stdlib - Authentication & Security Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/auth.lua`)
    - [x] Authentication utilities
      - [x] Add `auth.create_config(type, credentials)` for auth configuration
      - [x] Add `auth.from_env(provider)` for environment-based auth
      - [x] Add `auth.refresh_oauth2_token(config, refresh_token)` for token refresh
      - [x] Add `auth.validate_session(session_id)` for session validation
    - [x] OAuth and token management
      - [x] Add `auth.create_oauth2_config()` for OAuth2 flows
      - [x] Add `auth.parse_jwt_claims(token)` for JWT handling
      - [x] Add `auth.serialize_credentials()` for secure storage
      - [x] Add `auth.auto_refresh_token()` for automatic token refresh
    - [x] Permission and access control
      - [x] Add `auth.check_permission(permission, context)` for access control
      - [x] Add `auth.create_security_policy(name, rules)` for policy creation
      - [x] Add `auth.evaluate_policy(policy_name, context)` for policy evaluation
      - [x] Add `auth.log_event(event_type, metadata)` for audit logging
    - [x] Session management and multi-scheme authentication
      - [x] Add `auth.create_session(auth_config, session_id)` for sessions
      - [x] Add `auth.register_scheme(endpoint, scheme)` for multi-scheme support
      - [x] Add `auth.cache_credentials(key, auth_config, ttl)` for credential caching
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/auth_test.go`)
    - [x] Test authentication configuration and schemes
    - [x] Test OAuth2 token operations and JWT parsing
    - [x] Test session creation and validation
    - [x] Test security policy creation and evaluation (role-based, time-based, IP whitelist)
    - [x] Test credential serialization and caching
    - [x] Test audit logging and event handling
    - [x] Test multi-scheme authentication and error handling
    - [x] Test comprehensive integration with all auth bridges

- [x] **Task 2.3.5.10: Lua stdlib - Error Handling & Recovery Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/errors.lua`)
    - [x] Enhanced error handling
      - [x] Add `errors.try(func, catch_func, finally_func)` for try-catch-finally
      - [x] Add `errors.wrap(error, context)` for error wrapping
      - [x] Add `errors.chain(errors)` for error chaining
      - [x] Add `errors.create(message, code, context)` for custom error creation
    - [x] Retry and recovery mechanisms
      - [x] Add `errors.retry(func, options)` for retry logic with exponential/linear backoff
      - [x] Add `errors.circuit_breaker(func, config)` for fault tolerance
      - [x] Add `errors.fallback(primary, fallback)` for fallback strategies
      - [x] Add `errors.create_recovery_strategy(type, config)` for custom strategies
    - [x] Error categorization and reporting
      - [x] Add `errors.categorize(error)` for error classification
      - [x] Add `errors.is_retryable(error)` and `errors.is_fatal(error)` for error inspection
      - [x] Add `errors.aggregate(errors)` for error aggregation
      - [x] Add `errors.log_error(type, metadata)` for error reporting
    - [x] Serialization and context management
      - [x] Add `errors.to_json(error)` and `errors.from_json(json)` for serialization
      - [x] Add `errors.get_context(error)` and `errors.add_context(error, key, value)` for context
      - [x] Add `errors.register_category(name, matcher)` for custom categories
    - [x] Utility functions
      - [x] Add `errors.safe(func, default)` for safe function wrapping
      - [x] Add `errors.timeout(func, timeout_ms)` for timeout protection
      - [x] Add `errors.subscribe_to_errors(types, handler)` for event handling
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/errors_test.go`)
    - [x] Test try-catch-finally execution flow and error handling
    - [x] Test error wrapping, chaining, and context preservation  
    - [x] Test retry mechanisms with backoff strategies
    - [x] Test circuit breaker creation and execution
    - [x] Test fallback strategy implementation
    - [x] Test error categorization and property inspection
    - [x] Test error aggregation and serialization
    - [x] Test event handling and subscription mechanisms
    - [x] Test utility functions (safe, timeout) and system integration
    - [x] Test comprehensive error handling workflow integration

- [x] **Task 2.3.5.11: Lua stdlib - Logging & Debug Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/logging.lua`)
    - [x] Unified logging interface
      - [x] Add `log.info(message, context)` for info logging
      - [x] Add `log.warn(message, context)` for warning logging
      - [x] Add `log.error(message, context)` for error logging
      - [x] Add `log.debug(message, context)` for debug logging
    - [x] Structured logging utilities
      - [x] Add `log.with_context(context)` for context propagation
      - [x] Add `log.create_logger(component, level)` for component loggers
      - [x] Add `log.set_formatter(formatter)` for custom formatting
    - [x] Debug and diagnostics helpers
      - [x] Add `debug.trace_calls(func)` for call tracing (via component debug)
      - [x] Add `debug.memory_usage()` for memory monitoring (via system info)
      - [x] Add `debug.performance_profile(func)` for performance profiling
    - [x] Additional features implemented
      - [x] Hook integration for LLM operations monitoring
      - [x] Metrics collection (count, gauge, histogram)
      - [x] Audit logging with compliance support
      - [x] Error handling integration (catch, assert)
      - [x] Timer and profiling utilities
      - [x] Log search and statistics (framework in place)
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/logging_test.go`)
    - [x] Test log level filtering
    - [x] Test context propagation
    - [x] Test custom formatters
    - [x] Test call tracing accuracy (component debug)
    - [x] Test memory usage reporting (system info)
    - [x] Test performance profiling
    - [x] Test concurrent logging
    - [x] Test log rotation behavior (configuration)
    - [x] Additional tests
      - [x] Test hook registration and execution
      - [x] Test metrics collection
      - [x] Test audit logging and handlers
      - [x] Test error handling integration
      - [x] Test real-world usage scenarios
      - [x] Test graceful bridge failure handling
      - [x] Performance benchmarking

- [x] **Task 2.3.5.12: Lua stdlib - Testing & Validation Library** ✅ COMPLETED [2025-06-20]
  - [x] Implementation (`/pkg/engine/gopherlua/stdlib/testing.lua`)
    - [x] Test framework and assertions
      - [x] Add `testing.describe(name, tests)` for test grouping
      - [x] Add `testing.it(name, test_func)` for individual tests
      - [x] Add comprehensive assertion library (30+ assertion methods)
      - [x] Add `testing.assert.error(func, expected_error)` for error testing
    - [x] Mocking and stubbing utilities
      - [x] Add `testing.mock.func(name)` and `testing.mock.create(name)` for mocking
      - [x] Add `testing.stub(func, return_value)` for stubbing
      - [x] Add `testing.spy(func)` for function spying with call tracking
    - [x] Performance and load testing
      - [x] Add `testing.benchmark(func, iterations)` for benchmarking
      - [x] Add `testing.load_test(func, config)` for load testing
      - [x] Add `testing.memory_test(func)` for memory testing
  - [x] Testing (`/pkg/engine/gopherlua/stdlib/testing_test.go`)
    - [x] Test assertion functionality (all 30+ assertion types)
    - [x] Test mock behavior and control methods
    - [x] Test spy call tracking with metatable approach
    - [x] Test benchmark accuracy and statistics
    - [x] Test load test execution and metrics
    - [x] Test memory test functionality
    - [x] Test nested test groups and suite organization
    - [x] Test skip/only test functionality

- [ ] **Task 2.3.5.13: Lua stdlib - Core Utilities Library**
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

- [ ] **Task 2.3.5.14: Lua stdlib - Spell Framework Library**
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

- [ ] **Task 2.3.5.15: Lua stdlib - Documentation & Examples** (`/pkg/engine/gopherlua/stdlib/`)
  - [ ] Comprehensive documentation
    - [ ] Create `README.md` with library overview and philosophy
    - [ ] Create `API_REFERENCE.md` with complete function documentation
    - [ ] Create `EXAMPLES.md` with practical usage examples

- [ ] **Task 2.3.5.16: Lua stdlib - Test Infrastructure**
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

- [ ] **Task 2.3.5.17: Lua stdlib - Integration Testing**
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