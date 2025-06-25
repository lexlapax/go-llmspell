# LLM Adapter Analysis: llm.lua vs LLMAdapter

## Current Status
**FILE**: `pkg/engine/lua/stdlib/llm.lua` 
**ADAPTER**: `pkg/engine/lua/adapters/impl/llm.go`

## Method Coverage Analysis

### ✅ EXISTING in llm.lua (High-level convenience methods)
- `llm.quick_prompt()` - Basic prompting
- `llm.chat_session()` - Conversation management
- `llm.streaming_response()` - Streaming with callback
- `llm.batch_process()` - Multiple prompts
- `llm.use_provider()` - Provider selection
- `llm.list_providers()` - Provider enumeration
- Basic provider management

### ❌ MISSING from llm.lua (LLMAdapter methods)

#### Provider Methods (Flattened naming from LLMAdapter)
- `llm.providersCreate(providerType, name, config)` 
- `llm.providersGet(name)`
- `llm.providersList()`
- `llm.providersGetTemplate(templateName)`
- `llm.providersCreateMulti(name, providerList, strategy, config)`
- `llm.providersCreateFromEnvironment(providerType, name)`
- `llm.providersRemove(name)`
- `llm.providersTemplatesList()`
- `llm.providersTemplatesValidate(providerType, config)`
- `llm.providersConfigureMulti(name, config)`
- `llm.providersGetMulti(name)`
- `llm.providersCreateMock(name, responses)`
- `llm.providersGenerateWith(providerName, prompt, options)`
- `llm.providersExportConfig()`
- `llm.providersImportConfig(config)`
- `llm.providersSetMetadata(providerName, metadata)`
- `llm.providersGetMetadata(providerName)`
- `llm.providersListByCapability(capability)`

#### Pool Methods (Flattened naming from LLMAdapter)  
- `llm.poolCreate(name, providers, strategy, config)`
- `llm.poolGetHealth(poolName)`
- `llm.poolGenerate(poolName, prompt, options)`
- `llm.poolGetMetrics(poolName)`
- `llm.poolGet(poolName)`
- `llm.poolList()`
- `llm.poolRemove(poolName)`
- `llm.poolGetProviderHealth(poolName)`
- `llm.poolResetMetrics(poolName)`
- `llm.poolGenerateMessage(poolName, messages, options)`
- `llm.poolStream(poolName, prompt, options)`
- `llm.poolGetResponse()` - Object pooling
- `llm.poolReturnResponse(response)`
- `llm.poolGetToken()`
- `llm.poolReturnToken(token)`
- `llm.poolGetChannel()`
- `llm.poolReturnChannel(channel)`

#### Model Methods (Flattened naming from LLMAdapter)
- `llm.modelsList(provider)`
- `llm.modelsGetInfo(modelName)`
- `llm.modelsCheckCapabilities(modelName, capability)`

#### Core LLM Methods (Base bridge methods)
- `llm.generate(prompt, options)`
- `llm.generateMessage(messages, options)`
- `llm.stream(prompt, options)`
- `llm.countTokens(text, model)`
- `llm.createAgent(config)` - Agent creation
- `llm.agentComplete(agentId, prompt, options)`
- `llm.agentStream(agentId, prompt, options)`

#### Convenience Methods from LLMAdapter
- `llm.quick(prompt)` - Quick completion with default model
- `llm.batchComplete(prompts, options)` - Batch completion
- `llm.Agent(config)` - Alias for createAgent

#### Constants from LLMAdapter
- `llm.MODELS` - Model constants (GPT4, CLAUDE3, etc.)
- `llm.DEFAULTS` - Default options
- `llm.ERRORS` - Error codes
- `llm.STRATEGIES` - Pool strategies

## Architecture Issues

### 1. Bridge Access Pattern
**CURRENT llm.lua**: Uses `bridges.llm_core` only
**LLMAdapter**: Uses multi-bridge pattern:
- Main bridge (`llm_core`)
- `providersBridge` (`llm_providers`) 
- `poolBridge` (`llm_pool`)

### 2. API Philosophy Mismatch
**CURRENT llm.lua**: High-level convenience wrappers
**LLMAdapter**: Comprehensive bridge method exposure + convenience

### 3. Missing Integration
- No provider bridge integration
- No pool bridge integration  
- Missing agent creation/management
- No object pooling support
- Limited model management

## Recommended Actions

1. **KEEP** existing high-level convenience methods in llm.lua
2. **ADD** all missing LLMAdapter methods to llm.lua
3. **IMPLEMENT** multi-bridge support (providers, pools)
4. **ADD** agent integration methods
5. **ADD** constants and defaults
6. **MAINTAIN** backward compatibility

## Implementation Strategy
- Add provider methods with `llm.providers.*` namespace + flat methods
- Add pool methods with `llm.pool.*` namespace + flat methods  
- Add model methods with `llm.models.*` namespace + flat methods
- Add core methods as flat methods
- Add constants as `llm.MODELS`, `llm.DEFAULTS`, etc.
- Keep existing convenience methods unchanged