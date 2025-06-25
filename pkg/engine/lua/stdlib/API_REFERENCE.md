# go-llmspell Lua Standard Library API Reference

Complete API documentation for all modules in the go-llmspell Lua standard library.

## Table of Contents

- [Core Utilities (core.lua)](#core-utilities-corelua)
- [Promise & Async (promise.lua)](#promise--async-promiselua)
- [Spell Framework (spell.lua)](#spell-framework-spelllua)
- [LLM Operations (llm.lua)](#llm-operations-llmlua)
- [Agent Management (agent.lua)](#agent-management-agentlua)
- [Testing Framework (testing.lua)](#testing-framework-testinglua)
- [Event System (events.lua)](#event-system-eventslua)
- [State Management (state.lua)](#state-management-statelua)
- [Data Utilities (data.lua)](#data-utilities-datalua)
- [Tools Framework (tools.lua)](#tools-framework-toolslua)
- [Authentication (auth.lua)](#authentication-authlua)
- [Logging (logging.lua)](#logging-logginglua)
- [Error Handling (errors.lua)](#error-handling-errorslua)
- [Observability (observability.lua)](#observability-observabilitylua)

---

## Core Utilities (core.lua)

Extends Lua's built-in types with additional utility functions.

### String Extensions

#### `string.template(template, variables)`
Template string substitution with variable replacement.
- **Parameters:**
  - `template` (string): Template string with `${variable}` placeholders
  - `variables` (table): Variables for substitution
- **Returns:** (string) Rendered template
- **Example:**
  ```lua
  local result = string.template("Hello ${name}!", {name = "World"})
  -- result: "Hello World!"
  ```

#### `string.slugify(str)`
Convert string to URL-friendly slug.
- **Parameters:**
  - `str` (string): Input string
- **Returns:** (string) Slugified string
- **Example:**
  ```lua
  local slug = string.slugify("Hello World!")
  -- slug: "hello-world"
  ```

#### `string.truncate(str, length, suffix)`
Truncate string to specified length with optional suffix.
- **Parameters:**
  - `str` (string): Input string
  - `length` (number): Maximum length
  - `suffix` (string, optional): Suffix to append (default: "...")
- **Returns:** (string) Truncated string

#### `string.split(str, delimiter)`
Split string by delimiter.
- **Parameters:**
  - `str` (string): Input string
  - `delimiter` (string): Split delimiter
- **Returns:** (table) Array of substrings

#### `string.trim(str)`
Remove leading and trailing whitespace.
- **Parameters:**
  - `str` (string): Input string
- **Returns:** (string) Trimmed string

#### `string.capitalize(str)`
Capitalize first letter of each word.
- **Parameters:**
  - `str` (string): Input string
- **Returns:** (string) Capitalized string

#### `string.camelcase(str)`
Convert to camelCase.
- **Parameters:**
  - `str` (string): Input string
- **Returns:** (string) camelCase string

#### `string.snakecase(str)`
Convert to snake_case.
- **Parameters:**
  - `str` (string): Input string
- **Returns:** (string) snake_case string

### Table Extensions

#### `table.keys(tbl)`
Get all keys from table.
- **Parameters:**
  - `tbl` (table): Input table
- **Returns:** (table) Array of keys

#### `table.values(tbl)`
Get all values from table.
- **Parameters:**
  - `tbl` (table): Input table
- **Returns:** (table) Array of values

#### `table.merge(...)`
Merge multiple tables (shallow merge).
- **Parameters:**
  - `...` (table): Tables to merge
- **Returns:** (table) Merged table

#### `table.deep_copy(tbl)`
Deep copy table with circular reference handling.
- **Parameters:**
  - `tbl` (table): Table to copy
- **Returns:** (table) Deep copied table

#### `table.slice(tbl, start, length)`
Extract slice from array-like table.
- **Parameters:**
  - `tbl` (table): Input table
  - `start` (number): Start index (1-based)
  - `length` (number, optional): Number of elements
- **Returns:** (table) Sliced array

#### `table.reverse(tbl)`
Reverse array-like table in place.
- **Parameters:**
  - `tbl` (table): Array to reverse
- **Returns:** (table) Reversed array

#### `table.shuffle(tbl)`
Shuffle array-like table in place.
- **Parameters:**
  - `tbl` (table): Array to shuffle
- **Returns:** (table) Shuffled array

#### `table.contains(tbl, value)`
Check if table contains value.
- **Parameters:**
  - `tbl` (table): Table to search
  - `value` (any): Value to find
- **Returns:** (boolean) True if found

#### `table.is_empty(tbl)`
Check if table is empty.
- **Parameters:**
  - `tbl` (table): Table to check
- **Returns:** (boolean) True if empty

### Crypto Utilities

#### `core.crypto.uuid()`
Generate UUID v4.
- **Returns:** (string) UUID string

#### `core.crypto.random_string(length, charset)`
Generate random string.
- **Parameters:**
  - `length` (number): String length
  - `charset` (string, optional): Character set (default: alphanumeric)
- **Returns:** (string) Random string

### Time Utilities

#### `core.time.now()`
Get current timestamp.
- **Returns:** (number) Unix timestamp

#### `core.time.format(timestamp, format)`
Format timestamp.
- **Parameters:**
  - `timestamp` (number): Unix timestamp
  - `format` (string): Format string
- **Returns:** (string) Formatted time

#### `core.time.duration(seconds)`
Create duration object.
- **Parameters:**
  - `seconds` (number): Duration in seconds
- **Returns:** (table) Duration object with conversion methods

---

## Promise & Async (promise.lua)

Asynchronous programming support with Promise patterns and coroutine integration.

### Promise Class

#### `Promise.new(executor)`
Create new Promise.
- **Parameters:**
  - `executor` (function): Function with `(resolve, reject)` parameters
- **Returns:** (Promise) New Promise instance

#### `Promise.resolve(value)`
Create resolved Promise.
- **Parameters:**
  - `value` (any): Value to resolve with
- **Returns:** (Promise) Resolved Promise

#### `Promise.reject(error)`
Create rejected Promise.
- **Parameters:**
  - `error` (any): Error to reject with
- **Returns:** (Promise) Rejected Promise

#### `Promise.all(promises)`
Wait for all Promises to resolve.
- **Parameters:**
  - `promises` (table): Array of Promises
- **Returns:** (Promise) Promise that resolves with array of results

#### `Promise.race(promises)`
Wait for first Promise to resolve or reject.
- **Parameters:**
  - `promises` (table): Array of Promises
- **Returns:** (Promise) Promise that resolves/rejects with first result

### Promise Methods

#### `promise:andThen(onResolve, onReject)`
Chain Promise resolution.
- **Parameters:**
  - `onResolve` (function): Success callback
  - `onReject` (function, optional): Error callback
- **Returns:** (Promise) New Promise for chaining

#### `promise:onError(onReject)`
Handle Promise rejection.
- **Parameters:**
  - `onReject` (function): Error callback
- **Returns:** (Promise) New Promise for chaining

#### `promise:onFinally(onFinally)`
Execute callback regardless of outcome.
- **Parameters:**
  - `onFinally` (function): Cleanup callback
- **Returns:** (Promise) New Promise for chaining

### Async/Await

#### `async(func)`
Create async function that returns Promise.
- **Parameters:**
  - `func` (function): Function to make async
- **Returns:** (function) Async function

#### `await(promise, timeout)`
Wait for Promise to resolve.
- **Parameters:**
  - `promise` (Promise): Promise to wait for
  - `timeout` (number, optional): Timeout in milliseconds
- **Returns:** (any) Promise result
- **Throws:** Error if Promise rejects or times out

#### `sleep(duration)`
Sleep for specified duration.
- **Parameters:**
  - `duration` (number): Sleep duration in milliseconds
- **Returns:** (Promise) Promise that resolves after duration

### Coroutine Integration

#### `spawn(func, ...)`
Spawn coroutine for concurrent execution.
- **Parameters:**
  - `func` (function): Function to spawn
  - `...` (any): Arguments to pass
- **Returns:** (coroutine) Coroutine handle

#### `yield()`
Yield control to other coroutines.
- **Returns:** None

---

## Spell Framework (spell.lua)

Framework for spell lifecycle management, composition, and execution.

### Lifecycle Management

#### `spell.init(config)`
Initialize spell with configuration.
- **Parameters:**
  - `config` (table): Spell configuration
    - `name` (string): Spell name
    - `version` (string, optional): Version (default: "1.0.0")
    - `description` (string, optional): Description
    - `author` (string, optional): Author
    - `params` (table, optional): Parameter definitions
    - `timeout` (number, optional): Timeout in seconds (default: 300)
    - `max_retries` (number, optional): Max retries (default: 3)
    - `cache_ttl` (number, optional): Cache TTL in seconds (default: 300)
- **Returns:** (table) Spell configuration

#### `spell.params(name, config)`
Define or get spell parameter.
- **Parameters:**
  - `name` (string): Parameter name
  - `config` (table, optional): Parameter configuration
    - `type` (string): Parameter type ("string", "number", "boolean", "table", "array", "object")
    - `required` (boolean): Whether parameter is required
    - `default` (any): Default value
    - `enum` (table): Valid values for enum type
    - `description` (string): Parameter description
- **Returns:** (any) Parameter value if getting, config if defining

#### `spell.output(data, format, metadata)`
Output spell results.
- **Parameters:**
  - `data` (any): Result data
  - `format` (string, optional): Output format ("text", "json", "auto")
  - `metadata` (table, optional): Additional metadata
- **Returns:** (table) Output object with metadata

### Composition and Reuse

#### `spell.include(path_or_name)`
Include other spell or library.
- **Parameters:**
  - `path_or_name` (string): File path or registered library name
- **Returns:** (any) Included module or library

#### `spell.compose(spells_config)`
Compose multiple spells into workflow.
- **Parameters:**
  - `spells_config` (table): Array of spell step configurations
    - `name` (string, optional): Step name
    - `spell` (string): Spell name to execute
    - `params` (table, optional): Parameters with variable substitution
- **Returns:** (table) Results from all steps

#### `spell.library(name, functions)`
Create reusable library.
- **Parameters:**
  - `name` (string): Library name
  - `functions` (table): Table of functions
- **Returns:** (table) Registered library

### Context Management

#### `spell.context()`
Get current execution context.
- **Returns:** (table) Execution context with metadata

#### `spell.config(key, default)`
Get configuration value.
- **Parameters:**
  - `key` (string, optional): Configuration key
  - `default` (any, optional): Default value
- **Returns:** (any) Configuration value or entire config

#### `spell.cache(key, value, ttl)`
Cache operations.
- **Parameters:**
  - `key` (string): Cache key
  - `value` (any, optional): Value to cache
  - `ttl` (number, optional): TTL in seconds
- **Returns:** (any) Cached value if getting, value if setting

### Lifecycle Hooks

#### `spell.on_init(handler)`
Register initialization hook.
- **Parameters:**
  - `handler` (function): Hook function

#### `spell.on_start(handler)`
Register start hook.
- **Parameters:**
  - `handler` (function): Hook function

#### `spell.on_complete(handler)`
Register completion hook.
- **Parameters:**
  - `handler` (function): Hook function

#### `spell.on_cleanup(handler)`
Register cleanup hook.
- **Parameters:**
  - `handler` (function): Hook function

#### `spell.on_error(handler)`
Register error hook.
- **Parameters:**
  - `handler` (function): Hook function

### Environment and Resources

#### `spell.env(key, value)`
Environment variable operations.
- **Parameters:**
  - `key` (string): Environment key
  - `value` (any, optional): Value to set
- **Returns:** (any) Environment value

#### `spell.resource(name, config)`
Register managed resource.
- **Parameters:**
  - `name` (string): Resource name
  - `config` (table): Resource configuration
    - `type` (string): Resource type
    - `data` (any): Resource data
    - `cleanup` (function): Cleanup function
- **Returns:** (table) Resource object

#### `spell.cleanup_resources()`
Clean up all managed resources.
- **Returns:** None

#### `spell.sandbox(config)`
Create sandboxed execution environment.
- **Parameters:**
  - `config` (table, optional): Sandbox configuration
    - `globals` (table): Additional globals to include
- **Returns:** (table) Sandbox environment

### Utility Functions

#### `spell.validate_config(config)`
Validate spell configuration.
- **Parameters:**
  - `config` (table): Configuration to validate
- **Returns:** (boolean, string) Valid flag and error message

#### `spell.get_system_info()`
Get system information.
- **Returns:** (table) System information

#### `spell.get_cache_stats()`
Get cache statistics.
- **Returns:** (table) Cache statistics

#### `spell.clear_expired_cache()`
Clear expired cache entries.
- **Returns:** (number) Number of cleared entries

#### `spell.reset()`
Reset spell state (for testing).
- **Returns:** None

---

## LLM Operations (llm.lua)

High-level interface for LLM operations and provider management.

### Chat Operations

#### `llm.chat(message, options)`
Send chat message to LLM.
- **Parameters:**
  - `message` (string): Message to send
  - `options` (table, optional): Chat options
    - `model` (string): Model to use
    - `temperature` (number): Response randomness
    - `max_tokens` (number): Maximum response tokens
    - `system` (string): System prompt
- **Returns:** (Promise) Promise resolving to response

#### `llm.stream_chat(message, options)`
Stream chat response.
- **Parameters:**
  - `message` (string): Message to send
  - `options` (table, optional): Chat options
- **Returns:** (function) Iterator for response chunks

### Model Management

#### `llm.list_models()`
List available models.
- **Returns:** (Promise) Promise resolving to model list

#### `llm.get_model_info(model_name)`
Get model information.
- **Parameters:**
  - `model_name` (string): Model name
- **Returns:** (Promise) Promise resolving to model info

### Provider Operations

#### `llm.set_provider(provider_name, config)`
Set LLM provider.
- **Parameters:**
  - `provider_name` (string): Provider name
  - `config` (table): Provider configuration
- **Returns:** (boolean) Success status

#### `llm.get_providers()`
Get available providers.
- **Returns:** (table) List of providers

---

## Agent Management (agent.lua)

AI agent creation, management, and lifecycle operations.

### Agent Creation

#### `agent.create(config)`
Create new AI agent.
- **Parameters:**
  - `config` (table): Agent configuration
    - `name` (string): Agent name
    - `model` (string): LLM model to use
    - `system_prompt` (string): System prompt
    - `tools` (table, optional): Available tools
    - `memory_size` (number, optional): Conversation memory size
- **Returns:** (table) Agent instance

### Agent Operations

#### `agent.list()`
List all agents.
- **Returns:** (table) Array of agent instances

#### `agent.get(name)`
Get agent by name.
- **Parameters:**
  - `name` (string): Agent name
- **Returns:** (table) Agent instance or nil

#### `agent.remove(name)`
Remove agent.
- **Parameters:**
  - `name` (string): Agent name
- **Returns:** (boolean) Success status

### Agent Instance Methods

#### `agent_instance:chat(message, options)`
Send message to agent.
- **Parameters:**
  - `message` (string): Message to send
  - `options` (table, optional): Chat options
- **Returns:** (Promise) Promise resolving to response

#### `agent_instance:add_tool(tool_name, tool_config)`
Add tool to agent.
- **Parameters:**
  - `tool_name` (string): Tool name
  - `tool_config` (table): Tool configuration
- **Returns:** (boolean) Success status

#### `agent_instance:remove_tool(tool_name)`
Remove tool from agent.
- **Parameters:**
  - `tool_name` (string): Tool name
- **Returns:** (boolean) Success status

#### `agent_instance:get_metrics()`
Get agent metrics.
- **Returns:** (table) Agent performance metrics

---

## Testing Framework (testing.lua)

Comprehensive testing framework with assertions and test organization.

### Test Organization

#### `testing.describe(description, func)`
Create test suite.
- **Parameters:**
  - `description` (string): Suite description
  - `func` (function): Suite function
- **Returns:** None

#### `testing.it(description, func)`
Create test case.
- **Parameters:**
  - `description` (string): Test description
  - `func` (function): Test function
- **Returns:** None

#### `testing.before_each(func)`
Setup function run before each test.
- **Parameters:**
  - `func` (function): Setup function
- **Returns:** None

#### `testing.after_each(func)`
Teardown function run after each test.
- **Parameters:**
  - `func` (function): Teardown function
- **Returns:** None

### Assertions

#### `testing.assert.equal(actual, expected, message)`
Assert equality.
- **Parameters:**
  - `actual` (any): Actual value
  - `expected` (any): Expected value
  - `message` (string, optional): Error message

#### `testing.assert.not_equal(actual, expected, message)`
Assert inequality.
- **Parameters:**
  - `actual` (any): Actual value
  - `expected` (any): Expected value
  - `message` (string, optional): Error message

#### `testing.assert.is_true(value, message)`
Assert value is true.
- **Parameters:**
  - `value` (any): Value to check
  - `message` (string, optional): Error message

#### `testing.assert.is_false(value, message)`
Assert value is false.
- **Parameters:**
  - `value` (any): Value to check
  - `message` (string, optional): Error message

#### `testing.assert.is_nil(value, message)`
Assert value is nil.
- **Parameters:**
  - `value` (any): Value to check
  - `message` (string, optional): Error message

#### `testing.assert.is_not_nil(value, message)`
Assert value is not nil.
- **Parameters:**
  - `value` (any): Value to check
  - `message` (string, optional): Error message

#### `testing.assert.has_error(func, expected_error, message)`
Assert function throws error.
- **Parameters:**
  - `func` (function): Function to test
  - `expected_error` (string, optional): Expected error pattern
  - `message` (string, optional): Error message

#### `testing.assert.no_error(func, message)`
Assert function doesn't throw error.
- **Parameters:**
  - `func` (function): Function to test
  - `message` (string, optional): Error message

### Test Utilities

#### `testing.mock(original_func)`
Create mock function.
- **Parameters:**
  - `original_func` (function, optional): Original function to wrap
- **Returns:** (table) Mock object with call tracking

#### `testing.spy(target, method_name)`
Create spy for method.
- **Parameters:**
  - `target` (table): Target object
  - `method_name` (string): Method name to spy on
- **Returns:** (table) Spy object with call tracking

#### `testing.skip(reason)`
Skip current test.
- **Parameters:**
  - `reason` (string): Skip reason
- **Returns:** None

#### `testing.only()`
Mark test as only one to run.
- **Returns:** None

---

## Event System (events.lua)

Event-driven programming with hooks and lifecycle management.

### Event Operations

#### `events.emit(event, data)`
Emit event with data.
- **Parameters:**
  - `event` (string): Event name
  - `data` (any, optional): Event data
- **Returns:** None

#### `events.on(event, handler)`
Subscribe to event.
- **Parameters:**
  - `event` (string): Event name
  - `handler` (function): Event handler
- **Returns:** (string) Subscription ID

#### `events.once(event, handler)`
Subscribe to event (one-time).
- **Parameters:**
  - `event` (string): Event name
  - `handler` (function): Event handler
- **Returns:** (string) Subscription ID

#### `events.off(event, handler_or_id)`
Unsubscribe from event.
- **Parameters:**
  - `event` (string): Event name
  - `handler_or_id` (function|string): Handler function or subscription ID
- **Returns:** (boolean) Success status

#### `events.wait_for(event, timeout)`
Wait for event (Promise-based).
- **Parameters:**
  - `event` (string): Event name
  - `timeout` (number, optional): Timeout in milliseconds
- **Returns:** (Promise) Promise resolving with event data

### Event Utilities

#### `events.create_emitter()`
Create custom event emitter.
- **Returns:** (table) Event emitter instance

#### `events.aggregate(events, timeout)`
Collect multiple events.
- **Parameters:**
  - `events` (table): Array of event names
  - `timeout` (number, optional): Timeout in milliseconds
- **Returns:** (Promise) Promise resolving with collected events

#### `events.filter(pattern, handler)`
Filter events by pattern.
- **Parameters:**
  - `pattern` (string): Event pattern (supports wildcards)
  - `handler` (function): Event handler
- **Returns:** (string) Subscription ID

#### `events.namespace(name)`
Create namespaced event emitter.
- **Parameters:**
  - `name` (string): Namespace name
- **Returns:** (table) Namespaced emitter

### Hook System

#### `hooks.before(event, handler)`
Register pre-hook.
- **Parameters:**
  - `event` (string): Event name
  - `handler` (function): Hook handler
- **Returns:** (string) Hook ID

#### `hooks.after(event, handler)`
Register post-hook.
- **Parameters:**
  - `event` (string): Event name
  - `handler` (function): Hook handler
- **Returns:** (string) Hook ID

#### `hooks.around(event, wrapper)`
Register around-hook.
- **Parameters:**
  - `event` (string): Event name
  - `wrapper` (function): Wrapper function
- **Returns:** (string) Hook ID

#### `hooks.execute(event, func, args)`
Execute function with hooks.
- **Parameters:**
  - `event` (string): Event name
  - `func` (function): Function to execute
  - `args` (table): Function arguments
- **Returns:** (any) Function result

---

## State Management (state.lua)

State persistence, management, and transformation utilities.

### State Operations

#### `state.create(initial_data)`
Create new state.
- **Parameters:**
  - `initial_data` (table, optional): Initial state data
- **Returns:** (table) State object

#### `state.merge(state1, state2)`
Merge two states.
- **Parameters:**
  - `state1` (table): First state
  - `state2` (table): Second state
- **Returns:** (table) Merged state

#### `state.snapshot(state)`
Create state snapshot.
- **Parameters:**
  - `state` (table): State to snapshot
- **Returns:** (table) State snapshot

#### `state.restore(snapshot)`
Restore state from snapshot.
- **Parameters:**
  - `snapshot` (table): State snapshot
- **Returns:** (table) Restored state

### Persistence

#### `state.save(state, key)`
Save state persistently.
- **Parameters:**
  - `state` (table): State to save
  - `key` (string): Storage key
- **Returns:** (Promise) Promise resolving to success status

#### `state.load(key, default)`
Load state from storage.
- **Parameters:**
  - `key` (string): Storage key
  - `default` (table, optional): Default value
- **Returns:** (Promise) Promise resolving to state

#### `state.expire(key, duration)`
Set state expiration.
- **Parameters:**
  - `key` (string): Storage key
  - `duration` (number): Duration in seconds
- **Returns:** (Promise) Promise resolving to success status

### Transformation

#### `state.transform(state, transformer)`
Transform state.
- **Parameters:**
  - `state` (table): State to transform
  - `transformer` (function): Transformation function
- **Returns:** (table) Transformed state

#### `state.filter(state, predicate)`
Filter state properties.
- **Parameters:**
  - `state` (table): State to filter
  - `predicate` (function): Filter predicate
- **Returns:** (table) Filtered state

#### `state.validate(state, schema)`
Validate state against schema.
- **Parameters:**
  - `state` (table): State to validate
  - `schema` (table): Validation schema
- **Returns:** (boolean, table) Valid flag and validation errors

---

## Data Utilities (data.lua)

Data transformation, validation, and processing utilities.

### Data Transformation

#### `data.transform(input, transformer)`
Transform data using transformer function.
- **Parameters:**
  - `input` (any): Input data
  - `transformer` (function): Transformation function
- **Returns:** (any) Transformed data

#### `data.map(array, mapper)`
Map array elements.
- **Parameters:**
  - `array` (table): Input array
  - `mapper` (function): Mapping function
- **Returns:** (table) Mapped array

#### `data.filter(array, predicate)`
Filter array elements.
- **Parameters:**
  - `array` (table): Input array
  - `predicate` (function): Filter predicate
- **Returns:** (table) Filtered array

#### `data.reduce(array, reducer, initial)`
Reduce array to single value.
- **Parameters:**
  - `array` (table): Input array
  - `reducer` (function): Reducer function
  - `initial` (any): Initial value
- **Returns:** (any) Reduced value

### Data Validation

#### `data.validate(data, schema)`
Validate data against schema.
- **Parameters:**
  - `data` (any): Data to validate
  - `schema` (table): Validation schema
- **Returns:** (boolean, table) Valid flag and validation errors

#### `data.sanitize(data, rules)`
Sanitize data using rules.
- **Parameters:**
  - `data` (any): Data to sanitize
  - `rules` (table): Sanitization rules
- **Returns:** (any) Sanitized data

### Data Conversion

#### `data.to_json(data)`
Convert data to JSON string.
- **Parameters:**
  - `data` (any): Data to convert
- **Returns:** (string) JSON string

#### `data.from_json(json_string)`
Parse JSON string to data.
- **Parameters:**
  - `json_string` (string): JSON string
- **Returns:** (any) Parsed data

#### `data.to_csv(array, headers)`
Convert array to CSV string.
- **Parameters:**
  - `array` (table): Array of objects
  - `headers` (table, optional): Column headers
- **Returns:** (string) CSV string

#### `data.from_csv(csv_string, headers)`
Parse CSV string to array.
- **Parameters:**
  - `csv_string` (string): CSV string
  - `headers` (table, optional): Column headers
- **Returns:** (table) Array of objects

---

## Tools Framework (tools.lua)

Tool registration, execution, and management framework.

### Tool Registration

#### `tools.register(name, config)`
Register new tool.
- **Parameters:**
  - `name` (string): Tool name
  - `config` (table): Tool configuration
    - `description` (string): Tool description
    - `parameters` (table): Parameter schema
    - `execute` (function): Tool execution function
- **Returns:** (boolean) Success status

#### `tools.list()`
List registered tools.
- **Returns:** (table) Array of tool names

#### `tools.get(name)`
Get tool configuration.
- **Parameters:**
  - `name` (string): Tool name
- **Returns:** (table) Tool configuration

### Tool Execution

#### `tools.execute(name, params)`
Execute tool.
- **Parameters:**
  - `name` (string): Tool name
  - `params` (table): Tool parameters
- **Returns:** (Promise) Promise resolving to tool result

#### `tools.validate_params(name, params)`
Validate tool parameters.
- **Parameters:**
  - `name` (string): Tool name
  - `params` (table): Parameters to validate
- **Returns:** (boolean, table) Valid flag and validation errors

### Tool Utilities

#### `tools.search(query)`
Search tools by query.
- **Parameters:**
  - `query` (string): Search query
- **Returns:** (table) Matching tools

#### `tools.get_schema(name)`
Get tool parameter schema.
- **Parameters:**
  - `name` (string): Tool name
- **Returns:** (table) Parameter schema

#### `tools.unregister(name)`
Unregister tool.
- **Parameters:**
  - `name` (string): Tool name
- **Returns:** (boolean) Success status

---

## Authentication (auth.lua)

Authentication and authorization utilities.

### Authentication

#### `auth.login(credentials)`
Authenticate user.
- **Parameters:**
  - `credentials` (table): User credentials
- **Returns:** (Promise) Promise resolving to auth token

#### `auth.logout()`
Logout current user.
- **Returns:** (Promise) Promise resolving to success status

#### `auth.get_current_user()`
Get current authenticated user.
- **Returns:** (table) User information or nil

### Authorization

#### `auth.check_permission(permission)`
Check if user has permission.
- **Parameters:**
  - `permission` (string): Permission to check
- **Returns:** (boolean) Has permission

#### `auth.require_permission(permission)`
Require permission (throws error if missing).
- **Parameters:**
  - `permission` (string): Required permission
- **Returns:** None (throws error if missing)

### Token Management

#### `auth.get_token()`
Get current auth token.
- **Returns:** (string) Auth token or nil

#### `auth.set_token(token)`
Set auth token.
- **Parameters:**
  - `token` (string): Auth token
- **Returns:** None

#### `auth.refresh_token()`
Refresh auth token.
- **Returns:** (Promise) Promise resolving to new token

---

## Logging (logging.lua)

Structured logging with multiple levels and formatters.

### Log Levels

#### `logging.debug(message, context)`
Log debug message.
- **Parameters:**
  - `message` (string): Log message
  - `context` (table, optional): Additional context
- **Returns:** None

#### `logging.info(message, context)`
Log info message.
- **Parameters:**
  - `message` (string): Log message
  - `context` (table, optional): Additional context
- **Returns:** None

#### `logging.warn(message, context)`
Log warning message.
- **Parameters:**
  - `message` (string): Log message
  - `context` (table, optional): Additional context
- **Returns:** None

#### `logging.error(message, context)`
Log error message.
- **Parameters:**
  - `message` (string): Log message
  - `context` (table, optional): Additional context
- **Returns:** None

### Configuration

#### `logging.set_level(level)`
Set minimum log level.
- **Parameters:**
  - `level` (string): Log level ("debug", "info", "warn", "error")
- **Returns:** None

#### `logging.set_formatter(formatter)`
Set log formatter.
- **Parameters:**
  - `formatter` (function): Formatter function
- **Returns:** None

#### `logging.add_appender(appender)`
Add log appender.
- **Parameters:**
  - `appender` (table): Appender configuration
- **Returns:** None

---

## Error Handling (errors.lua)

Error handling, recovery, and debugging utilities.

### Error Creation

#### `errors.new(message, code, details)`
Create new error.
- **Parameters:**
  - `message` (string): Error message
  - `code` (string, optional): Error code
  - `details` (table, optional): Error details
- **Returns:** (table) Error object

#### `errors.wrap(original_error, message)`
Wrap existing error.
- **Parameters:**
  - `original_error` (any): Original error
  - `message` (string): Wrapper message
- **Returns:** (table) Wrapped error

### Error Handling

#### `errors.try(func, error_handler)`
Execute function with error handling.
- **Parameters:**
  - `func` (function): Function to execute
  - `error_handler` (function, optional): Error handler
- **Returns:** (any) Function result or error

#### `errors.retry(func, max_retries, delay)`
Retry function execution.
- **Parameters:**
  - `func` (function): Function to retry
  - `max_retries` (number): Maximum retry attempts
  - `delay` (number, optional): Delay between retries
- **Returns:** (any) Function result

### Error Analysis

#### `errors.is_type(error, error_type)`
Check error type.
- **Parameters:**
  - `error` (any): Error to check
  - `error_type` (string): Expected error type
- **Returns:** (boolean) Is specified type

#### `errors.get_stack_trace()`
Get current stack trace.
- **Returns:** (string) Stack trace

---

## Observability (observability.lua)

Metrics, monitoring, and debugging utilities.

### Metrics

#### `metrics.counter(name, value, tags)`
Increment counter metric.
- **Parameters:**
  - `name` (string): Metric name
  - `value` (number, optional): Increment value (default: 1)
  - `tags` (table, optional): Metric tags
- **Returns:** None

#### `metrics.gauge(name, value, tags)`
Set gauge metric.
- **Parameters:**
  - `name` (string): Metric name
  - `value` (number): Gauge value
  - `tags` (table, optional): Metric tags
- **Returns:** None

#### `metrics.histogram(name, value, tags)`
Record histogram value.
- **Parameters:**
  - `name` (string): Metric name
  - `value` (number): Value to record
  - `tags` (table, optional): Metric tags
- **Returns:** None

#### `metrics.timer(name, func, tags)`
Time function execution.
- **Parameters:**
  - `name` (string): Metric name
  - `func` (function): Function to time
  - `tags` (table, optional): Metric tags
- **Returns:** (any) Function result

### Monitoring

#### `monitor.health_check(name, check_func)`
Register health check.
- **Parameters:**
  - `name` (string): Check name
  - `check_func` (function): Health check function
- **Returns:** None

#### `monitor.get_health()`
Get overall health status.
- **Returns:** (table) Health status

#### `monitor.trace(operation, func)`
Trace operation execution.
- **Parameters:**
  - `operation` (string): Operation name
  - `func` (function): Function to trace
- **Returns:** (any) Function result

### Debugging

#### `debug.inspect(value, options)`
Inspect value for debugging.
- **Parameters:**
  - `value` (any): Value to inspect
  - `options` (table, optional): Inspection options
- **Returns:** (string) Inspection result

#### `debug.profile(func, options)`
Profile function execution.
- **Parameters:**
  - `func` (function): Function to profile
  - `options` (table, optional): Profiling options
- **Returns:** (table) Profiling results

---

## Type Definitions

### Common Types

#### Configuration Objects
Most functions accept configuration objects with optional parameters. Default values are provided for most options.

#### Promise Objects
Asynchronous operations return Promise objects that can be chained using `andThen`, `onError`, and `onFinally` methods.

#### Error Objects
Errors are structured objects with `message`, `code`, and `details` properties for consistent error handling.

#### Context Objects
Execution contexts contain metadata about the current operation, including timing, parameters, and environment information.

For practical usage examples, see [EXAMPLES.md](EXAMPLES.md).