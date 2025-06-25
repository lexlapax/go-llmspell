# Utils Adapter Analysis: utils.lua vs UtilsAdapter

## Current Status
**FILE**: `pkg/engine/lua/stdlib/utils.lua`
**ADAPTER**: `pkg/engine/lua/adapters/impl/utils.go`

## Architecture Comparison

### UtilsAdapter Architecture
- **Multi-bridge pattern**: Uses 8 different bridge interfaces
- **Bridge Dependencies**: auth, debug, errors, json, llm, logger, slog, util
- **API Style**: Flattened namespace methods with prefixes (e.g., `authAuthenticate`, `jsonParse`)
- **Method Count**: ~50 flattened methods across all bridge namespaces
- **Focus**: Comprehensive utility functionality through bridge system

### utils.lua Architecture  
- **Tool-based pattern**: Uses `tools` module for built-in operations
- **Dependencies**: Only requires `tools` module
- **API Style**: Simple function names for file/system operations
- **Method Count**: ~25 utility functions for file/system/time operations
- **Focus**: Basic file system and utility operations

## Key Differences

### 1. API Philosophy
**UtilsAdapter**: Multi-bridge utility system with comprehensive functionality
**utils.lua**: Simple tool-based utilities for basic operations

### 2. Bridge Dependencies
**UtilsAdapter**: Requires 8 bridges (auth, debug, errors, json, llm, logger, slog, util)
**utils.lua**: No bridge dependencies - uses tools system

### 3. Functionality Scope
**UtilsAdapter**: ~50 methods across 8 namespaces
**utils.lua**: ~25 methods for file/system/time operations

## Implementation Assessment

### ❌ INCOMPATIBLE ARCHITECTURES
The current utils.lua implementation is **fundamentally incompatible** with the UtilsAdapter pattern:

1. **Different dependency models**: Tool-based vs multi-bridge
2. **Different method signatures**: Simple utilities vs bridge wrappers
3. **Different scope**: Basic operations vs comprehensive utility system
4. **Different bridge requirements**: None vs 8 specific bridges

### ❌ MISSING ADAPTER COVERAGE
Current utils.lua provides **zero coverage** of UtilsAdapter functionality:

**Missing Auth Methods** (6 methods):
- `authAuthenticate` - missing entirely
- `authValidateToken` - missing entirely  
- `authRefreshToken` - missing entirely
- `authGenerateToken` - missing entirely
- `authHashPassword` - missing entirely
- `authVerifyPassword` - missing entirely

**Missing Debug Methods** (7 methods):
- `debugSetLevel` - missing entirely
- `debugLog` - missing entirely
- `debugGetConfig` - missing entirely
- `debugTrace` - missing entirely
- `debugProfile` - missing entirely
- `debugDump` - missing entirely
- `debugAssert` - missing entirely

**Missing Error Methods** (8 methods):
- `errorsCreateError` - missing entirely
- `errorsWrapError` - missing entirely
- `errorsAggregateErrors` - missing entirely
- `errorsCategorizeError` - missing entirely
- `errorsWrap` - missing entirely
- `errorsUnwrap` - missing entirely
- `errorsIsType` - missing entirely
- `errorsGetStack` - missing entirely

**Missing JSON Methods** (8 methods):
- `jsonParse` - missing entirely
- `jsonToJSON` - missing entirely
- `jsonValidateJSONSchema` - missing entirely
- `jsonExtractStructuredData` - missing entirely
- `jsonEncode` - missing entirely
- `jsonDecode` - missing entirely
- `jsonValidate` - missing entirely
- `jsonPrettify` - missing entirely

**Missing LLM Methods** (7 methods):
- `llmCreateProvider` - missing entirely
- `llmGenerateTyped` - missing entirely
- `llmTrackCost` - missing entirely
- `llmParseResponse` - missing entirely
- `llmFormatPrompt` - missing entirely
- `llmCountTokens` - missing entirely
- `llmSplitMessage` - missing entirely

**Missing Logger Methods** (7 methods):
- `loggerCreateLogger` - missing entirely
- `loggerLog` - missing entirely
- `loggerSetLogLevel` - missing entirely
- `loggerError` - missing entirely
- `loggerWarn` - missing entirely
- `loggerInfo` - missing entirely
- `loggerDebug` - missing entirely

**Missing Slog Methods** (5 methods):
- `slogInfo` - missing entirely
- `slogWarn` - missing entirely
- `slogError` - missing entirely
- `slogDebug` - missing entirely
- `slogWithFields` - missing entirely

**Missing General Methods** (7 methods):
- `generalGenerateUUID` - missing entirely
- `generalHash` - missing entirely
- `generalRetry` - missing entirely
- `generalSleep` - `utils.sleep()` exists but different signature
- `generalUuid` - missing entirely
- `generalEncode` - missing entirely
- `generalDecode` - missing entirely

### ❌ TEST FAILURES
Current tests fail due to missing bridge dependencies:
- Tools bridge not available error
- File extension regex logic error
- File operations fail without proper bridge setup

## Required Changes

### COMPLETE REWRITE NEEDED
To match the UtilsAdapter pattern, utils.lua requires a **complete rewrite**:

1. **Remove tool-based architecture**
2. **Implement multi-bridge pattern** with 8 bridge accessors
3. **Add all 50+ missing adapter methods** with proper bridge routing
4. **Implement bridge availability checks** for graceful degradation
5. **Add adapter constants** (LOG_LEVELS, AUTH_SCHEMES, HASH_ALGORITHMS, ERROR_CATEGORIES)
6. **Update tests** to use mock bridges instead of tools

### Method Implementation Required

**Auth Bridge Methods** (6):
```lua
function utils.authAuthenticate(credentials, scheme)
function utils.authValidateToken(token, options)
function utils.authRefreshToken(refreshToken)
function utils.authGenerateToken(userData, options)
function utils.authHashPassword(password)
function utils.authVerifyPassword(password, hash)
```

**Debug Bridge Methods** (7):
```lua
function utils.debugSetLevel(component, level)
function utils.debugLog(component, message, data)
function utils.debugGetConfig()
function utils.debugTrace(message)
function utils.debugProfile(name, action)
function utils.debugDump(value)
function utils.debugAssert(condition, message)
```

**Error Bridge Methods** (8):
```lua
function utils.errorsCreateError(message, code, category)
function utils.errorsWrapError(originalError, contextData)
function utils.errorsAggregateErrors(errorsTable)
function utils.errorsCategorizeError(errorData)
function utils.errorsWrap(originalError, contextData)
function utils.errorsUnwrap(errorData)
function utils.errorsIsType(errorData, errorType)
function utils.errorsGetStack(errorData)
```

**JSON Bridge Methods** (8):
```lua
function utils.jsonParse(text, options)
function utils.jsonToJSON(data, options)
function utils.jsonValidateJSONSchema(data, schema)
function utils.jsonExtractStructuredData(text, schema)
function utils.jsonEncode(data)
function utils.jsonDecode(text)
function utils.jsonValidate(text)
function utils.jsonPrettify(text)
```

**LLM Bridge Methods** (7):
```lua
function utils.llmCreateProvider(providerType, config)
function utils.llmGenerateTyped(prompt, schema, options)
function utils.llmTrackCost(operation, tokens, model)
function utils.llmParseResponse(response)
function utils.llmFormatPrompt(template, data)
function utils.llmCountTokens(text, model)
function utils.llmSplitMessage(message, maxTokens)
```

**Logger Bridge Methods** (7):
```lua
function utils.loggerCreateLogger(component, config)
function utils.loggerLog(level, message, contextTable)
function utils.loggerSetLogLevel(component, level)
function utils.loggerError(message, contextTable)
function utils.loggerWarn(message, contextTable)
function utils.loggerInfo(message, contextTable)
function utils.loggerDebug(message, contextTable)
```

**Slog Bridge Methods** (5):
```lua
function utils.slogInfo(message, fields)
function utils.slogWarn(message, fields)
function utils.slogError(message, fields)
function utils.slogDebug(message, fields)
function utils.slogWithFields(fields)
```

**General Bridge Methods** (7):
```lua
function utils.generalGenerateUUID()
function utils.generalHash(data, algorithm)
function utils.generalRetry(operation, options)
function utils.generalSleep(duration)
function utils.generalUuid()
function utils.generalEncode(data, encoding)
function utils.generalDecode(data, encoding)
```

## Recommendation

**COMPLETE REWRITE REQUIRED**

The current utils.lua implementation has **zero compatibility** with the UtilsAdapter pattern and requires a complete rewrite:

1. **Architecture Change**: Tool-based → Multi-bridge pattern
2. **Scope Expansion**: 25 methods → 50+ adapter methods
3. **Bridge Integration**: Add 8 bridge accessor functions
4. **Method Implementation**: Add all missing adapter methods with proper signatures
5. **Constants Addition**: Add LOG_LEVELS, AUTH_SCHEMES, HASH_ALGORITHMS, ERROR_CATEGORIES
6. **Test Updates**: Rewrite tests to use mock bridges instead of tools
7. **Error Handling**: Implement bridge availability checks and graceful degradation

This represents the most significant gap between any stdlib module and its corresponding adapter.