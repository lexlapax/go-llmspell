# LLM.lua CamelCase to snake_case Conversions

## Summary

All camelCase functions in llm.lua have been converted to snake_case, with backward compatibility aliases added for each one. The namespace references have also been updated to use the new snake_case functions.

## Conversion List

### Core Functions
1. `generateMessage` → `generate_message` ✓
2. `countTokens` → `count_tokens` ✓
3. `createAgent` → `create_agent` ✓
4. `agentComplete` → `agent_complete` ✓
5. `agentStream` → `agent_stream` ✓
6. `batchComplete` → `batch_complete` ✓

### Provider Functions (providers*)
7. `providersCreate` → `providers_create` ✓
8. `providersGet` → `providers_get` ✓
9. `providersList` → `providers_list` ✓
10. `providersGetTemplate` → `providers_get_template` ✓
11. `providersCreateMulti` → `providers_create_multi` ✓
12. `providersCreateFromEnvironment` → `providers_create_from_environment` ✓
13. `providersRemove` → `providers_remove` ✓
14. `providersTemplatesList` → `providers_templates_list` ✓
15. `providersTemplatesValidate` → `providers_templates_validate` ✓
16. `providersConfigureMulti` → `providers_configure_multi` ✓
17. `providersGetMulti` → `providers_get_multi` ✓
18. `providersCreateMock` → `providers_create_mock` ✓
19. `providersGenerateWith` → `providers_generate_with` ✓
20. `providersExportConfig` → `providers_export_config` ✓
21. `providersImportConfig` → `providers_import_config` ✓
22. `providersSetMetadata` → `providers_set_metadata` ✓
23. `providersGetMetadata` → `providers_get_metadata` ✓
24. `providersListByCapability` → `providers_list_by_capability` ✓

### Pool Functions (pool*)
25. `poolCreate` → `pool_create` ✓
26. `poolGetHealth` → `pool_get_health` ✓
27. `poolGenerate` → `pool_generate` ✓
28. `poolGetMetrics` → `pool_get_metrics` ✓
29. `poolGet` → `pool_get` ✓
30. `poolList` → `pool_list` ✓
31. `poolRemove` → `pool_remove` ✓
32. `poolGetProviderHealth` → `pool_get_provider_health` ✓
33. `poolResetMetrics` → `pool_reset_metrics` ✓
34. `poolGenerateMessage` → `pool_generate_message` ✓
35. `poolStream` → `pool_stream` ✓
36. `poolGetResponse` → `pool_get_response` ✓
37. `poolReturnResponse` → `pool_return_response` ✓
38. `poolGetToken` → `pool_get_token` ✓
39. `poolReturnToken` → `pool_return_token` ✓
40. `poolGetChannel` → `pool_get_channel` ✓
41. `poolReturnChannel` → `pool_return_channel` ✓

### Model Functions (models*)
42. `modelsList` → `models_list` ✓
43. `modelsGetInfo` → `models_get_info` ✓
44. `modelsCheckCapabilities` → `models_check_capabilities` ✓

## Implementation Details

### Backward Compatibility
- Each renamed function has a camelCase alias pointing to the new snake_case version
- Example: `llm.generateMessage = llm.generate_message`

### Namespace Updates
All three namespaces have been updated to reference the new snake_case functions:
- `llm.providers.*`
- `llm.pool.*`
- `llm.models.*`

### Testing Recommendations
1. Run existing tests to ensure backward compatibility aliases work
2. Create new tests using snake_case functions
3. Verify namespace access works with both naming conventions

## Migration Guide

### For Users
- Existing code using camelCase will continue to work
- New code should use snake_case for consistency
- Both styles can be mixed during migration

### Examples
```lua
-- Old style (still works)
local response = llm.generateMessage(messages, options)
local tokens = llm.countTokens(text)

-- New style (recommended)
local response = llm.generate_message(messages, options)
local tokens = llm.count_tokens(text)

-- Namespace access (both work)
llm.providers.create(...)  -- uses snake_case internally
llm.providersCreate(...)    -- still works via alias
```