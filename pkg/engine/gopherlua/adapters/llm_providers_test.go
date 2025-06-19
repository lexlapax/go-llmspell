// ABOUTME: Tests for LLM bridge adapter providers functionality that extends the LLM bridge
// ABOUTME: Validates provider creation, templates, multi-provider support, and metadata management

package adapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	lua "github.com/yuin/gopher-lua"

	"github.com/lexlapax/go-llmspell/pkg/engine"
	"github.com/lexlapax/go-llmspell/pkg/engine/gopherlua"
)

// Mock providers bridge for testing
type mockProvidersBridge struct {
	id          string
	initialized bool
	methods     map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error)
}

func (m *mockProvidersBridge) GetID() string                                  { return m.id }
func (m *mockProvidersBridge) GetMetadata() engine.BridgeMetadata             { return engine.BridgeMetadata{} }
func (m *mockProvidersBridge) Initialize(ctx context.Context) error           { m.initialized = true; return nil }
func (m *mockProvidersBridge) Cleanup(ctx context.Context) error              { m.initialized = false; return nil }
func (m *mockProvidersBridge) IsInitialized() bool                            { return m.initialized }
func (m *mockProvidersBridge) RegisterWithEngine(e engine.ScriptEngine) error { return nil }
func (m *mockProvidersBridge) Methods() []engine.MethodInfo                   { return nil }
func (m *mockProvidersBridge) ValidateMethod(name string, args []engine.ScriptValue) error {
	return nil
}
func (m *mockProvidersBridge) TypeMappings() map[string]engine.TypeMapping { return nil }
func (m *mockProvidersBridge) RequiredPermissions() []engine.Permission    { return nil }

func (m *mockProvidersBridge) ExecuteMethod(ctx context.Context, name string, args []engine.ScriptValue) (engine.ScriptValue, error) {
	if method, ok := m.methods[name]; ok {
		return method(args)
	}
	return engine.NewErrorValue(fmt.Errorf("method not found: %s", name)), nil
}

func TestLLMAdapter_ProvidersEnhancement(t *testing.T) {
	t.Run("providers_methods_with_bridge", func(t *testing.T) {
		// Create mock LLM bridge
		llmBridge := &mockLLMBridge{
			id: "llm",
			metadata: engine.BridgeMetadata{
				Name: "LLM Bridge",
			},
		}

		// Create mock providers bridge
		providersBridge := &mockProvidersBridge{
			id:          "providers",
			initialized: true,
			methods: map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error){
				"createProviderFromEnvironment": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":   args[1],
						"type":   args[0],
						"source": engine.NewStringValue("environment"),
					}), nil
				},
				"removeProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"listProviderTemplates": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"type":        engine.NewStringValue("openai"),
							"description": engine.NewStringValue("OpenAI GPT models"),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"type":        engine.NewStringValue("anthropic"),
							"description": engine.NewStringValue("Anthropic Claude models"),
						}),
					}), nil
				},
				"validateProviderConfig": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"valid":  engine.NewBoolValue(true),
						"errors": engine.NewArrayValue([]engine.ScriptValue{}),
					}), nil
				},
				"configureMultiProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"getMultiProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":     args[0],
						"strategy": engine.NewStringValue("primary"),
						"providers": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewObjectValue(map[string]engine.ScriptValue{
								"name":    engine.NewStringValue("provider1"),
								"weight":  engine.NewNumberValue(1.0),
								"primary": engine.NewBoolValue(true),
							}),
						}),
					}), nil
				},
				"createMockProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":      args[0],
						"type":      engine.NewStringValue("mock"),
						"responses": engine.NewNumberValue(3),
					}), nil
				},
			},
		}

		// Create adapter with providers bridge
		adapter := NewLLMAdapter(llmBridge, providersBridge, nil)

		// Should have provider methods
		methods := adapter.GetMethods()
		assert.Contains(t, methods, "providersCreateFromEnvironment")
		assert.Contains(t, methods, "providersRemove")
		assert.Contains(t, methods, "providersTemplatesList")
		assert.Contains(t, methods, "providersTemplatesValidate")
		assert.Contains(t, methods, "providersConfigureMulti")
		assert.Contains(t, methods, "providersGetMulti")
		assert.Contains(t, methods, "providersCreateMock")
	})

	t.Run("providers_methods_in_lua", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				switch method {
				case "createProvider":
					return map[string]interface{}{
						"name":    args[1],
						"type":    args[0],
						"created": true,
					}, nil
				case "getProvider":
					return map[string]interface{}{
						"name": args[0],
						"type": "openai",
					}, nil
				case "listProviders":
					return []interface{}{"provider1", "provider2"}, nil
				case "getProviderTemplate":
					return map[string]interface{}{
						"type":            args[0],
						"description":     "Template description",
						"requiredEnvVars": []interface{}{"API_KEY"},
					}, nil
				}
				return nil, fmt.Errorf("unknown method: %s", method)
			},
		}

		providersBridge := &mockProvidersBridge{
			id:          "providers",
			initialized: true,
			methods: map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error){
				"createProviderFromEnvironment": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":    args[1],
						"type":    args[0],
						"source":  engine.NewStringValue("environment"),
						"created": engine.NewBoolValue(true),
					}), nil
				},
				"removeProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"listProviderTemplates": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"type": engine.NewStringValue("openai"),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"type": engine.NewStringValue("anthropic"),
						}),
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"type": engine.NewStringValue("mock"),
						}),
					}), nil
				},
				"validateProviderConfig": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					// Simulate validation
					configMap := args[1].ToGo().(map[string]interface{})
					if _, hasKey := configMap["api_key"]; hasKey {
						return engine.NewObjectValue(map[string]engine.ScriptValue{
							"valid":  engine.NewBoolValue(true),
							"errors": engine.NewArrayValue([]engine.ScriptValue{}),
						}), nil
					}
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"valid": engine.NewBoolValue(false),
						"errors": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewStringValue("missing required field: api_key"),
						}),
					}), nil
				},
				"createMockProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":      args[0],
						"type":      engine.NewStringValue("mock"),
						"responses": engine.NewNumberValue(float64(len(args[1].ToGo().([]interface{})))),
					}), nil
				},
				"generateWithProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewStringValue(fmt.Sprintf("Generated from %s: %s", args[0].ToGo(), args[1].ToGo())), nil
				},
				"exportProviderConfig": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"providers": engine.NewObjectValue(map[string]engine.ScriptValue{
							"provider1": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("openai"),
							}),
						}),
						"templates": engine.NewObjectValue(map[string]engine.ScriptValue{
							"openai": engine.NewObjectValue(map[string]engine.ScriptValue{
								"type": engine.NewStringValue("openai"),
							}),
						}),
					}), nil
				},
				"importProviderConfig": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"setProviderMetadata": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"getProviderMetadata": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name": args[0],
						"metadata": engine.NewObjectValue(map[string]engine.ScriptValue{
							"description": engine.NewStringValue("Test provider"),
							"version":     engine.NewStringValue("1.0.0"),
						}),
					}), nil
				},
				"listProvidersByCapability": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewArrayValue([]engine.ScriptValue{
						engine.NewObjectValue(map[string]engine.ScriptValue{
							"name":       engine.NewStringValue("provider1"),
							"capability": args[0],
						}),
					}), nil
				},
			},
		}

		adapter := NewLLMAdapter(llmBridge, providersBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test all provider methods
		err = L.DoString(`
			local llm = require("llm")
			
			-- Test basic provider creation
			local provider1, err = llm.providersCreate("openai", "test-provider", {api_key = "test-key"})
			assert(err == nil, "should not error")
			assert(provider1.name == "test-provider", "should have correct name")
			assert(provider1.created == true, "should be created")
			
			-- Test environment-based creation
			local provider2, err = llm.providersCreateFromEnvironment("openai", "env-provider")
			assert(err == nil, "should not error")
			assert(provider2.source == "environment", "should be from environment")
			
			-- Test provider listing
			local providers, err = llm.providersList()
			assert(err == nil, "should not error")
			assert(#providers == 2, "should have two providers")
			
			-- Test provider get
			local provider, err = llm.providersGet("test-provider")
			assert(err == nil, "should not error")
			assert(provider.name == "test-provider", "should get correct provider")
			
			-- Test template operations
			local template, err = llm.providersGetTemplate("openai")
			assert(err == nil, "should not error")
			assert(template.type == "openai", "should get correct template")
			assert(#template.requiredEnvVars == 1, "should have required vars")
			
			local templates, err = llm.providersTemplatesList()
			assert(err == nil, "should not error")
			assert(#templates == 3, "should have three templates")
			
			-- Test config validation
			local valid_result, err = llm.providersTemplatesValidate("openai", {api_key = "test"})
			assert(err == nil, "should not error")
			assert(valid_result.valid == true, "should be valid")
			assert(#valid_result.errors == 0, "should have no errors")
			
			local invalid_result, err = llm.providersTemplatesValidate("openai", {})
			assert(err == nil, "should not error")
			assert(invalid_result.valid == false, "should be invalid")
			assert(#invalid_result.errors == 1, "should have one error")
			
			-- Test mock provider
			local mock, err = llm.providersCreateMock("mock-provider", {"Response 1", "Response 2", "Response 3"})
			assert(err == nil, "should not error")
			assert(mock.type == "mock", "should be mock type")
			assert(mock.responses == 3, "should have 3 responses")
			
			-- Test generate with provider
			local result, err = llm.providersGenerateWith("mock-provider", "Test prompt", {})
			assert(err == nil, "should not error")
			assert(string.find(result, "Generated from"), "should generate response")
			
			-- Test config export/import
			local config, err = llm.providersExportConfig()
			assert(err == nil, "should not error")
			assert(config.providers ~= nil, "should have providers")
			assert(config.templates ~= nil, "should have templates")
			
			local success, err = llm.providersImportConfig(config)
			assert(err == nil, "should not error")
			assert(success == true, "should import successfully")
			
			-- Test metadata operations
			local success, err = llm.providersSetMetadata("test-provider", {description = "Test", version = "1.0.0"})
			assert(err == nil, "should not error")
			assert(success == true, "should set metadata")
			
			local metadata, err = llm.providersGetMetadata("test-provider")
			assert(err == nil, "should not error")
			assert(metadata.metadata.description == "Test provider", "should have metadata")
			
			-- Test capability listing
			local capable, err = llm.providersListByCapability("generate")
			assert(err == nil, "should not error")
			assert(#capable == 1, "should have one capable provider")
			assert(capable[1].capability == "generate", "should have capability")
			
			-- Test provider removal
			local success, err = llm.providersRemove("test-provider")
			assert(err == nil, "should not error")
			assert(success == true, "should remove successfully")
		`)
		assert.NoError(t, err)
	})

	t.Run("multi_provider_functionality", func(t *testing.T) {
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "createMultiProvider" {
					return map[string]interface{}{
						"name":     args[0],
						"strategy": args[2],
						"created":  true,
					}, nil
				}
				return nil, fmt.Errorf("unknown method: %s", method)
			},
		}

		providersBridge := &mockProvidersBridge{
			id:          "providers",
			initialized: true,
			methods: map[string]func(args []engine.ScriptValue) (engine.ScriptValue, error){
				"configureMultiProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewNilValue(), nil
				},
				"getMultiProvider": func(args []engine.ScriptValue) (engine.ScriptValue, error) {
					return engine.NewObjectValue(map[string]engine.ScriptValue{
						"name":     args[0],
						"strategy": engine.NewStringValue("fastest"),
						"providers": engine.NewArrayValue([]engine.ScriptValue{
							engine.NewObjectValue(map[string]engine.ScriptValue{
								"name":    engine.NewStringValue("openai"),
								"weight":  engine.NewNumberValue(0.7),
								"primary": engine.NewBoolValue(false),
							}),
							engine.NewObjectValue(map[string]engine.ScriptValue{
								"name":    engine.NewStringValue("anthropic"),
								"weight":  engine.NewNumberValue(0.3),
								"primary": engine.NewBoolValue(false),
							}),
						}),
						"config": engine.NewObjectValue(map[string]engine.ScriptValue{
							"consensusThreshold": engine.NewNumberValue(0.5),
							"timeout":            engine.NewNumberValue(30),
							"retryOnFailure":     engine.NewBoolValue(true),
						}),
					}), nil
				},
			},
		}

		adapter := NewLLMAdapter(llmBridge, providersBridge, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Test multi-provider functionality
		err = L.DoString(`
			local llm = require("llm")
			
			-- Create multi-provider
			local multi, err = llm.providersCreateMulti("multi-test", {"openai", "anthropic"}, "fastest", {})
			assert(err == nil, "should not error")
			assert(multi.name == "multi-test", "should have correct name")
			assert(multi.strategy == "fastest", "should have correct strategy")
			
			-- Configure multi-provider
			local success, err = llm.providersConfigureMulti("multi-test", {
				consensusThreshold = 0.7,
				timeout = 60,
				retryOnFailure = false
			})
			assert(err == nil, "should not error")
			assert(success == true, "should configure successfully")
			
			-- Get multi-provider info
			local info, err = llm.providersGetMulti("multi-test")
			assert(err == nil, "should not error")
			assert(info.name == "multi-test", "should have correct name")
			assert(info.strategy == "fastest", "should have strategy")
			assert(#info.providers == 2, "should have two providers")
			assert(info.providers[1].weight == 0.7, "should have correct weight")
			assert(info.config.timeout == 30, "should have timeout")
		`)
		assert.NoError(t, err)
	})

	t.Run("providers_methods_without_bridge", func(t *testing.T) {
		// Create adapter without providers bridge
		llmBridge := &mockLLMBridge{
			id: "llm",
			callFunc: func(method string, args ...interface{}) (interface{}, error) {
				if method == "createProvider" {
					return map[string]interface{}{"created": true}, nil
				}
				if method == "listProviders" {
					return []interface{}{"provider1"}, nil
				}
				return nil, fmt.Errorf("unknown method")
			},
		}

		adapter := NewLLMAdapter(llmBridge, nil, nil)
		L := lua.NewState()
		defer L.Close()

		// Register module
		ms := gopherlua.NewModuleSystem()
		err := adapter.RegisterAsModule(ms, "llm")
		require.NoError(t, err)

		err = ms.LoadModule(L, "llm")
		require.NoError(t, err)

		// Enhanced provider methods should not exist
		err = L.DoString(`
			local llm = require("llm")
			
			-- Basic provider methods should still work
			local provider, err = llm.providersCreate("openai", "test", {})
			assert(err == nil, "providersCreate should work")
			assert(provider.created == true)
			
			local list, err = llm.providersList()
			assert(err == nil, "providersList should work")
			assert(#list == 1)
			
			-- Enhanced provider methods should not exist
			assert(llm.providersCreateFromEnvironment == nil, "providersCreateFromEnvironment should not exist")
			assert(llm.providersRemove == nil, "providersRemove should not exist")
			assert(llm.providersTemplatesList == nil, "providersTemplatesList should not exist")
			assert(llm.providersTemplatesValidate == nil, "providersTemplatesValidate should not exist")
			assert(llm.providersConfigureMulti == nil, "providersConfigureMulti should not exist")
			assert(llm.providersGetMulti == nil, "providersGetMulti should not exist")
			assert(llm.providersCreateMock == nil, "providersCreateMock should not exist")
			assert(llm.providersGenerateWith == nil, "providersGenerateWith should not exist")
		`)
		assert.NoError(t, err)
	})
}
