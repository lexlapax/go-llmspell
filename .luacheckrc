-- Luacheck configuration for go-llmspell project
-- Static analysis rules for Lua code

-- Lua version and standard library
std = "lua54"
cache = true

-- Global variables that are allowed (will not trigger warnings)
globals = {
    "_G",
    -- Bridge globals that are injected by go-llmspell
    "llm_bridge",
    "provider_bridge", 
    "pool_bridge",
    "llm_util_bridge",
    "promise",
}

-- Patterns for allowed global access
allow_defined_top = true

-- Ignore specific warning codes
ignore = {
    -- "212", -- Unused argument (keep this enabled for better code quality)
    -- "213", -- Unused loop variable (keep this enabled)
    "631", -- Line is too long (we'll use stylua for formatting)
}

-- Maximum line length (stylua will handle this)
max_line_length = false

-- File-specific configurations
files["examples/**/*.lua"] = {
    -- Examples can be more lenient - suppress most warnings for development ease
    ignore = {
        "211", -- Unused local variable
        "212", -- Unused argument
        "213", -- Unused loop variable  
        "231", -- Variable is never accessed
        "311", -- Value assigned to variable is unused
        "411", -- Variable was previously defined
        "421", -- Shadowing upvalue
        "431", -- Shadowing upvalue argument  
        "511", -- Variable is never accessed
        "611", -- Line contains only whitespace
        "612", -- Line contains trailing whitespace
        "613", -- Trailing whitespace in a string
        "614", -- Trailing whitespace in a string
        "621", -- Line contains only whitespace (duplicate)
        "631", -- Line is too long (formatting handles this)
    },
    -- Additional globals for examples that are injected by runtime
    globals = {
        "_G",
        "params",     -- Runtime parameters
        "llm",        -- LLM bridge
        "tools",      -- Tools bridge 
        "agents",     -- Agents bridge
        "storage",    -- Storage bridge
        "http",       -- HTTP bridge
        "json",       -- JSON utilities
        "async",      -- Async utilities
        "util",       -- General utilities
        "continue",   -- Special control flow (non-standard Lua)
    }
}

files["pkg/engine/gopherlua/stdlib/*_test.lua"] = {
    -- Test files can access test-specific globals
    globals = {
        "test",
        "describe", 
        "it",
        "before_each",
        "after_each",
    }
}

-- Stdlib files should be strict
files["pkg/engine/gopherlua/stdlib/*.lua"] = {
    -- Standard library code should be very clean
    ignore = {},
    -- Additional globals for stdlib
    globals = {
        "_G",
        "llm_bridge",
        "provider_bridge", 
        "pool_bridge",
        "llm_util_bridge",
        "promise",
        "require",
        "module",
        "package",
    }
}

-- Specific configuration for promise.lua to handle intentional shadowing and unused parameters
files["pkg/engine/gopherlua/stdlib/promise.lua"] = {
    ignore = {
        "212", -- Unused argument (reject parameters in some promise methods)
        "421", -- Shadowing upvalue (intentional for promise variable in closures)
        "4..", -- All shadowing warnings (4xx series)
    },
    globals = {
        "_G",
        "require",
        "module",
        "package",
    }
}