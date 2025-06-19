// ABOUTME: Structured bridge adapter that exposes go-llms schema validation and generation functionality to Lua scripts
// ABOUTME: Provides schema creation, validation, generation, repository operations, import/export, and custom validation features

package adapters

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// StructuredAdapter specializes BridgeAdapter for structured output functionality
type StructuredAdapter struct {
	*gopherlua.BridgeAdapter
}

// NewStructuredAdapter creates a new structured adapter
func NewStructuredAdapter(bridge engine.Bridge) *StructuredAdapter {
	// Create structured adapter
	adapter := &StructuredAdapter{}

	// Create base adapter if bridge is provided
	if bridge != nil {
		adapter.BridgeAdapter = gopherlua.NewBridgeAdapter(bridge)
	}

	// Add structured-specific methods if not already present
	adapter.ensureStructuredMethods()

	return adapter
}

// ensureStructuredMethods ensures structured-specific methods are available
func (sa *StructuredAdapter) ensureStructuredMethods() {
	// These methods should already be exposed by the bridge
	// For now, this is a placeholder for future validation
	// In production, this could validate that expected structured methods exist
}

// CreateLuaModule creates a Lua module with structured-specific enhancements
func (sa *StructuredAdapter) CreateLuaModule() lua.LGFunction {
	return func(L *lua.LState) int {
		// Create module table
		module := L.NewTable()

		// Add base bridge methods if bridge adapter exists
		if sa.BridgeAdapter != nil {
			// Call base module loader to get the base module
			baseLoader := sa.BridgeAdapter.CreateLuaModule()
			err := L.CallByParam(lua.P{
				Fn:      L.NewFunction(baseLoader),
				NRet:    1,
				Protect: true,
			})
			if err != nil {
				L.RaiseError("failed to create base module: %v", err)
				return 0
			}

			// Get the base module and copy its methods
			baseModule := L.Get(-1).(*lua.LTable)
			L.Pop(1)

			// Copy base module methods to our module
			baseModule.ForEach(func(k, v lua.LValue) {
				module.RawSet(k, v)
			})
		}

		// Add our own metadata
		L.SetField(module, "_adapter", lua.LString("structured"))
		L.SetField(module, "_version", lua.LString("2.0.0"))

		// Add structured-specific enhancements
		sa.addStructuredEnhancements(L, module)

		// Add flattened methods from all namespaces
		sa.addFlattenedMethods(L, module)

		// Add convenience methods
		sa.addConvenienceMethods(L, module)

		// Push the module and return it
		L.Push(module)
		return 1
	}
}

// addStructuredEnhancements adds structured-specific enhancements to the module
func (sa *StructuredAdapter) addStructuredEnhancements(L *lua.LState, module *lua.LTable) {
	// Add structured constants
	sa.addStructuredConstants(L, module)
}

// addFlattenedMethods adds all flattened methods from all namespaces
func (sa *StructuredAdapter) addFlattenedMethods(L *lua.LState, module *lua.LTable) {
	// Validation methods - flattened to module level
	L.SetField(module, "validationValidateJSON", L.NewFunction(func(L *lua.LState) int {
		schema := L.CheckTable(1)
		data := L.CheckTable(2)

		schemaMap := sa.tableToMap(L, schema)
		dataMap := sa.tableToMap(L, data)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewObjectValue(schemaMap),
			engine.NewObjectValue(dataMap),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "validateJSON", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "validationValidateStruct", L.NewFunction(func(L *lua.LState) int {
		schema := L.CheckTable(1)
		data := L.CheckTable(2)

		schemaMap := sa.tableToMap(L, schema)
		dataMap := sa.tableToMap(L, data)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewObjectValue(schemaMap),
			engine.NewObjectValue(dataMap),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "validateStruct", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Generation methods - flattened to module level
	L.SetField(module, "generationFromType", L.NewFunction(func(L *lua.LState) int {
		typeInfo := L.CheckTable(1)

		typeInfoMap := sa.tableToMap(L, typeInfo)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(typeInfoMap)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "generateSchemaFromType", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "generationFromTags", L.NewFunction(func(L *lua.LState) int {
		structData := L.CheckTable(1)

		structDataMap := sa.tableToMap(L, structData)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(structDataMap)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "generateFromTags", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "generationFromJSONSchema", L.NewFunction(func(L *lua.LState) int {
		jsonSchema := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(jsonSchema)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "convertJSONSchema", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Repository methods - flattened to module level
	L.SetField(module, "repositorySave", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		schema := L.CheckTable(2)

		schemaMap := sa.tableToMap(L, schema)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(name),
			engine.NewObjectValue(schemaMap),
		}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "saveSchema", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	L.SetField(module, "repositoryGet", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(name)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "getSchema", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "repositoryDelete", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(name)}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "deleteSchema", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	L.SetField(module, "repositoryInitializeFile", L.NewFunction(func(L *lua.LState) int {
		directory := L.CheckString(1)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewStringValue(directory)}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "initializeFileRepository", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	// ImportExport methods - flattened to module level
	L.SetField(module, "importExportToJSONSchema", L.NewFunction(func(L *lua.LState) int {
		schema := L.CheckTable(1)

		schemaMap := sa.tableToMap(L, schema)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(schemaMap)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "exportToJSONSchema", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "importExportToOpenAPI", L.NewFunction(func(L *lua.LState) int {
		schema := L.CheckTable(1)

		schemaMap := sa.tableToMap(L, schema)

		ctx := context.Background()
		args := []engine.ScriptValue{engine.NewObjectValue(schemaMap)}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "exportToOpenAPI", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "importExportFromFile", L.NewFunction(func(L *lua.LState) int {
		filePath := L.CheckString(1)
		format := L.OptString(2, "json")

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(filePath),
			engine.NewStringValue(format),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "importFromFile", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "importExportMerge", L.NewFunction(func(L *lua.LState) int {
		schemas := L.CheckTable(1)
		strategy := L.OptString(2, "merge_all")

		// Convert schemas table to array
		var schemaValues []engine.ScriptValue
		schemas.ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTTable {
				tableMap := sa.tableToMap(L, v.(*lua.LTable))
				schemaValues = append(schemaValues, engine.NewObjectValue(tableMap))
			}
		})

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewArrayValue(schemaValues),
			engine.NewStringValue(strategy),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "mergeSchemas", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Custom methods - flattened to module level
	L.SetField(module, "customRegisterValidator", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		validator := L.CheckTable(2)

		validatorMap := sa.tableToMap(L, validator)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewStringValue(name),
			engine.NewObjectValue(validatorMap),
		}

		_, err := sa.GetBridge().ExecuteMethod(ctx, "registerCustomValidator", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LNil)
		return 1
	}))

	L.SetField(module, "customValidate", L.NewFunction(func(L *lua.LState) int {
		data := L.CheckTable(1)
		validatorName := L.CheckString(2)

		dataMap := sa.tableToMap(L, data)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewObjectValue(dataMap),
			engine.NewStringValue(validatorName),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "validateWithCustom", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "customListValidators", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := sa.GetBridge().ExecuteMethod(ctx, "listCustomValidators", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		// Handle array results - they return as multiple values
		if result != nil && result.Type() == engine.TypeArray {
			if arrayResult, ok := result.(engine.ArrayValue); ok {
				elements := arrayResult.Elements()
				for _, elem := range elements {
					lval, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, elem)
					if err != nil {
						L.Push(lua.LNil)
						L.Push(lua.LString(err.Error()))
						return 2
					}
					L.Push(lval)
				}
				return len(elements)
			}
		}

		// Single return fallback
		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		return 1
	}))

	L.SetField(module, "customValidateAsync", L.NewFunction(func(L *lua.LState) int {
		schema := L.CheckTable(1)
		data := L.CheckTable(2)

		schemaMap := sa.tableToMap(L, schema)
		dataMap := sa.tableToMap(L, data)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewObjectValue(schemaMap),
			engine.NewObjectValue(dataMap),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "validateAsync", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	L.SetField(module, "customGetMetrics", L.NewFunction(func(L *lua.LState) int {
		ctx := context.Background()

		result, err := sa.GetBridge().ExecuteMethod(ctx, "getValidationMetrics", []engine.ScriptValue{})
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))

	// Utils methods - flattened to module level
	L.SetField(module, "utilsGenerateDiff", L.NewFunction(func(L *lua.LState) int {
		oldSchema := L.CheckTable(1)
		newSchema := L.CheckTable(2)

		oldSchemaMap := sa.tableToMap(L, oldSchema)
		newSchemaMap := sa.tableToMap(L, newSchema)

		ctx := context.Background()
		args := []engine.ScriptValue{
			engine.NewObjectValue(oldSchemaMap),
			engine.NewObjectValue(newSchemaMap),
		}

		result, err := sa.GetBridge().ExecuteMethod(ctx, "generateDiff", args)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(luaResult)
		L.Push(lua.LNil)
		return 2
	}))
}

// addConvenienceMethods adds convenience methods to the module
func (sa *StructuredAdapter) addConvenienceMethods(L *lua.LState, module *lua.LTable) {
	// Add createProperty method if not already present
	if module.RawGetString("createProperty") == lua.LNil {
		L.SetField(module, "createProperty", L.NewFunction(func(L *lua.LState) int {
			propertyType := L.CheckString(1)
			constraints := L.OptTable(2, L.NewTable())

			constraintsMap := sa.tableToMap(L, constraints)

			ctx := context.Background()
			args := []engine.ScriptValue{
				engine.NewStringValue(propertyType),
				engine.NewObjectValue(constraintsMap),
			}

			result, err := sa.GetBridge().ExecuteMethod(ctx, "createProperty", args)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			luaResult, err := sa.BridgeAdapter.GetTypeConverter().FromLuaScriptValue(L, result)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(luaResult)
			L.Push(lua.LNil)
			return 2
		}))
	}
}

// addStructuredConstants adds structured-related constants to the module
func (sa *StructuredAdapter) addStructuredConstants(L *lua.LState, module *lua.LTable) {
	// Add schema types
	types := L.NewTable()
	L.SetField(types, "STRING", lua.LString("string"))
	L.SetField(types, "NUMBER", lua.LString("number"))
	L.SetField(types, "INTEGER", lua.LString("integer"))
	L.SetField(types, "BOOLEAN", lua.LString("boolean"))
	L.SetField(types, "OBJECT", lua.LString("object"))
	L.SetField(types, "ARRAY", lua.LString("array"))
	L.SetField(types, "NULL", lua.LString("null"))
	L.SetField(module, "TYPES", types)

	// Add format types
	formats := L.NewTable()
	L.SetField(formats, "EMAIL", lua.LString("email"))
	L.SetField(formats, "DATE", lua.LString("date"))
	L.SetField(formats, "DATETIME", lua.LString("date-time"))
	L.SetField(formats, "TIME", lua.LString("time"))
	L.SetField(formats, "URI", lua.LString("uri"))
	L.SetField(formats, "UUID", lua.LString("uuid"))
	L.SetField(formats, "IPV4", lua.LString("ipv4"))
	L.SetField(formats, "IPV6", lua.LString("ipv6"))
	L.SetField(module, "FORMATS", formats)

	// Add validation operators
	operators := L.NewTable()
	L.SetField(operators, "AND", lua.LString("and"))
	L.SetField(operators, "OR", lua.LString("or"))
	L.SetField(operators, "NOT", lua.LString("not"))
	L.SetField(module, "OPERATORS", operators)
}

// WrapMethod wraps a bridge method with structured-specific handling
func (sa *StructuredAdapter) WrapMethod(methodName string) lua.LGFunction {
	// Get base wrapped method if available
	if sa.BridgeAdapter != nil {
		baseWrapped := sa.BridgeAdapter.WrapMethod(methodName)

		// Add structured-specific handling for certain methods
		switch methodName {
		case "createSchema", "validateJSON", "validateStruct", "generateSchemaFromType":
			return sa.wrapStructuredOperation(methodName, baseWrapped)
		default:
			return baseWrapped
		}
	}

	// Return a simple function that returns an error when no bridge is available
	return func(L *lua.LState) int {
		L.Push(lua.LNil)
		L.Push(lua.LString("method not available - no bridge adapter"))
		return 2
	}
}

// wrapStructuredOperation adds structured operation handling
func (sa *StructuredAdapter) wrapStructuredOperation(_ string, baseFn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Ensure at least one parameter is provided for structured operations
		if L.GetTop() == 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("structured operation requires parameters"))
			return 2
		}

		return baseFn(L)
	}
}

// tableToMap converts a Lua table to a map[string]engine.ScriptValue
func (sa *StructuredAdapter) tableToMap(L *lua.LState, table *lua.LTable) map[string]engine.ScriptValue {
	result := make(map[string]engine.ScriptValue)

	table.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			// Convert value to ScriptValue
			var converter *gopherlua.LuaTypeConverter
			if sa.BridgeAdapter != nil {
				converter = sa.GetTypeConverter()
			} else {
				converter = gopherlua.NewLuaTypeConverter()
			}

			sv, err := converter.ToLuaScriptValue(L, v)
			if err == nil {
				result[string(key)] = sv
			}
		}
	})

	return result
}

// RegisterAsModule registers the adapter as a module in the module system
func (sa *StructuredAdapter) RegisterAsModule(ms *gopherlua.ModuleSystem, name string) error {
	// Get bridge metadata
	var bridgeMetadata engine.BridgeMetadata
	if sa.GetBridge() != nil {
		bridgeMetadata = sa.GetBridge().GetMetadata()
	} else {
		bridgeMetadata = engine.BridgeMetadata{
			Name:        "Structured Adapter",
			Description: "Schema validation and generation functionality",
		}
	}

	// Create module definition using our overridden CreateLuaModule
	module := gopherlua.ModuleDefinition{
		Name:         name,
		Description:  bridgeMetadata.Description,
		Dependencies: []string{},           // Structured module has no dependencies by default
		LoadFunc:     sa.CreateLuaModule(), // Use our enhanced module creator
	}

	// Register the module
	return ms.Register(module)
}

// GetBridge returns the underlying bridge
func (sa *StructuredAdapter) GetBridge() engine.Bridge {
	if sa.BridgeAdapter != nil {
		return sa.BridgeAdapter.GetBridge()
	}
	return nil
}

// GetMethods returns the available methods
func (sa *StructuredAdapter) GetMethods() []string {
	// Get base methods if bridge adapter exists
	var methods []string
	if sa.BridgeAdapter != nil {
		methods = sa.BridgeAdapter.GetMethods()
	}

	// Add structured-specific flattened methods if not already present
	structuredMethods := []string{
		// Standard convenience methods
		"createSchema", "createProperty",

		// Validation methods (flattened)
		"validationValidateJSON", "validationValidateStruct",

		// Generation methods (flattened)
		"generationFromType", "generationFromTags", "generationFromJSONSchema",

		// Repository methods (flattened)
		"repositorySave", "repositoryGet", "repositoryDelete", "repositoryInitializeFile",

		// ImportExport methods (flattened)
		"importExportToJSONSchema", "importExportToOpenAPI", "importExportFromFile", "importExportMerge",

		// Custom methods (flattened)
		"customRegisterValidator", "customValidate", "customListValidators", "customValidateAsync", "customGetMetrics",

		// Utils methods (flattened)
		"utilsGenerateDiff",
	}

	methodMap := make(map[string]bool)
	for _, m := range methods {
		methodMap[m] = true
	}

	for _, m := range structuredMethods {
		if !methodMap[m] {
			methods = append(methods, m)
		}
	}

	return methods
}
