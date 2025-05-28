package stdlib

import (
	"os"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestJSONModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	RegisterJSON(L)

	// Test encode
	err := L.DoString(`
		local data = {name = "test", value = 42, active = true}
		result = json.encode(data)
	`)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	result := L.GetGlobal("result").String()
	if !strings.Contains(result, `"name":"test"`) {
		t.Errorf("Expected result to contain name:test, got %s", result)
	}

	// Test decode
	err = L.DoString(`
		local json_str = '{"name":"test","value":42}'
		data, err = json.decode(json_str)
		if data then
			decoded_name = data.name
			decoded_value = data.value
		end
	`)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if L.GetGlobal("decoded_name").String() != "test" {
		t.Errorf("Expected decoded_name to be 'test', got %s", L.GetGlobal("decoded_name").String())
	}

	if L.GetGlobal("decoded_value").(lua.LNumber) != 42 {
		t.Errorf("Expected decoded_value to be 42, got %v", L.GetGlobal("decoded_value"))
	}

	// Test decode error
	err = L.DoString(`
		bad_data, bad_err = json.decode("{invalid json}")
	`)
	if err != nil {
		t.Fatalf("Failed to handle decode error: %v", err)
	}

	if L.GetGlobal("bad_data") != lua.LNil {
		t.Errorf("Expected bad_data to be nil")
	}

	if L.GetGlobal("bad_err") == lua.LNil {
		t.Errorf("Expected bad_err to be set")
	}
}

func TestStorageModule(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "llmspell-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		BaseDir:     tempDir,
		MaxFileSize: 1024,
		AllowedExts: []string{".txt", ".json"},
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	L := lua.NewState()
	defer L.Close()

	RegisterStorage(L, storage)

	// Test in-memory storage
	err = L.DoString(`
		storage.set("key1", "value1")
		result = storage.get("key1")
		missing = storage.get("nonexistent")
	`)
	if err != nil {
		t.Fatalf("Failed in-memory test: %v", err)
	}

	if L.GetGlobal("result").String() != "value1" {
		t.Errorf("Expected result to be 'value1', got %s", L.GetGlobal("result").String())
	}

	if L.GetGlobal("missing") != lua.LNil {
		t.Errorf("Expected missing to be nil")
	}

	// Test file operations
	err = L.DoString(`
		write_err = storage.write("test.txt", "Hello, World!")
		if not write_err then
			content, read_err = storage.read("test.txt")
			exists = storage.exists("test.txt")
		end
	`)
	if err != nil {
		t.Fatalf("Failed file operations: %v", err)
	}

	if L.GetGlobal("write_err") != lua.LNil {
		t.Errorf("Expected write_err to be nil, got %v", L.GetGlobal("write_err"))
	}

	if L.GetGlobal("content").String() != "Hello, World!" {
		t.Errorf("Expected content to be 'Hello, World!', got %s", L.GetGlobal("content").String())
	}

	if L.GetGlobal("exists") != lua.LTrue {
		t.Errorf("Expected file to exist")
	}

	// Test invalid extension
	err = L.DoString(`
		invalid_err = storage.write("test.exe", "content")
	`)
	if err != nil {
		t.Fatalf("Failed invalid extension test: %v", err)
	}

	if L.GetGlobal("invalid_err") == lua.LNil {
		t.Errorf("Expected error for invalid extension")
	}
}

func TestLogModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Use simple log for testing
	RegisterSimpleLog(L)

	// Just test that functions exist and don't error
	err := L.DoString(`
		log.info("test info message")
		log.error("test error message")
	`)
	if err != nil {
		t.Fatalf("Failed to run log functions: %v", err)
	}
}

func TestHTTPModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	httpClient := NewHTTPClient(nil)
	RegisterHTTP(L, httpClient)

	// Test URL validation
	err := L.DoString(`
		content, err = http.get("invalid://url")
	`)
	if err != nil {
		t.Fatalf("Failed URL validation test: %v", err)
	}

	if L.GetGlobal("content") != lua.LNil {
		t.Errorf("Expected content to be nil for invalid URL")
	}

	if L.GetGlobal("err") == lua.LNil {
		t.Errorf("Expected error for invalid URL")
	}
}

func TestRegisterAll(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "llmspell-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	L := lua.NewState()
	defer L.Close()

	config := &Config{
		Storage: &StorageConfig{
			BaseDir: tempDir,
		},
		HTTP:      DefaultHTTPConfig(),
		SpellName: "test",
	}

	err = RegisterAll(L, config)
	if err != nil {
		t.Fatalf("Failed to register all: %v", err)
	}

	// Test that all modules are available
	err = L.DoString(`
		-- Test JSON
		local data = json.encode({test = true})
		local decoded = json.decode(data)
		
		-- Test log
		log.info("test")
		
		-- Test storage
		storage.set("key", "value")
		local val = storage.get("key")
		
		-- Test HTTP exists
		assert(type(http.get) == "function", "http.get should be a function")
	`)
	if err != nil {
		t.Fatalf("Failed to test all modules: %v", err)
	}
}
