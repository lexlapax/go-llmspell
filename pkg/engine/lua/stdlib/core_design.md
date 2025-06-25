# Core Utilities Library Design

## Overview
The Core Utilities Library provides essential utility functions for Lua scripts in go-llmspell. It focuses on string manipulation, collection utilities, cryptographic functions, and time/date handling that aren't covered by other stdlib modules.

## Design Principles
1. **No Duplication**: Avoid reimplementing functions that exist in other modules
2. **Clean API**: Extend Lua's built-in types naturally (string, table, os)
3. **Performance**: Efficient implementations for common operations
4. **Type Safety**: Clear parameter validation and error messages
5. **Integration**: Work seamlessly with other stdlib modules

## Core API Design

### 1. String Utilities (Extend string table)
```lua
-- Template string with variable substitution
string.template(template, variables)
-- Example: string.template("Hello {{name}}!", {name = "World"}) -> "Hello World!"

-- Convert to URL-safe slug
string.slugify(text)
-- Example: string.slugify("Hello World!") -> "hello-world"

-- Truncate string with optional suffix
string.truncate(text, length, suffix)
-- Example: string.truncate("Long text here", 10, "...") -> "Long te..."

-- Additional useful string utilities
string.split(str, delimiter) -- Split string by delimiter
string.trim(str) -- Remove leading/trailing whitespace
string.capitalize(str) -- Capitalize first letter
string.camelcase(str) -- Convert to camelCase
string.snakecase(str) -- Convert to snake_case
```

### 2. Table Utilities (Extend table table)
```lua
-- Extract keys from table
table.keys(tbl)
-- Example: table.keys({a=1, b=2}) -> {"a", "b"}

-- Extract values from table
table.values(tbl)
-- Example: table.values({a=1, b=2}) -> {1, 2}

-- Shallow merge (delegates to data.merge for deep merge)
table.merge(t1, t2, ...)
-- Example: table.merge({a=1}, {b=2}) -> {a=1, b=2}

-- Deep copy (delegates to data.clone)
table.deep_copy(tbl)

-- Additional table utilities
table.is_empty(tbl) -- Check if table is empty
table.slice(tbl, start, end) -- Array slice
table.reverse(tbl) -- Reverse array in-place
table.shuffle(tbl) -- Shuffle array in-place
table.contains(tbl, value) -- Check if value exists
```

### 3. Crypto Utilities (New crypto namespace)
```lua
local crypto = {}

-- Generate UUID (delegates to testing.data.uuid for implementation)
crypto.uuid()
-- Example: crypto.uuid() -> "550e8400-e29b-41d4-a716-446655440000"

-- Hash data using various algorithms
crypto.hash(data, algorithm)
-- Example: crypto.hash("hello", "sha256") -> "2cf24dba5fb0a30e26e83b2ac5b9e29e..."
-- Supported: md5, sha1, sha256, sha512

-- Generate random string (delegates to testing.data.random_string)
crypto.random_string(length, charset)
-- Example: crypto.random_string(16) -> "Xy7kP9mN2qRt5Lw8"

-- Base64 encoding/decoding
crypto.base64_encode(data)
crypto.base64_decode(encoded)

-- URL-safe base64
crypto.base64url_encode(data)
crypto.base64url_decode(encoded)
```

### 4. Time Utilities (Extend os table)
```lua
-- Get current timestamp in seconds
os.now()
-- Example: os.now() -> 1634567890.123

-- Format timestamp using strftime patterns
os.format(timestamp, format)
-- Example: os.format(os.now(), "%Y-%m-%d %H:%M:%S") -> "2023-10-17 14:30:45"

-- Calculate duration between timestamps
os.duration(start_time, end_time)
-- Returns table: {seconds=123, minutes=2.05, hours=0.034, days=0.0014}

-- Parse time string to timestamp
os.parse_time(time_str, format)
-- Example: os.parse_time("2023-10-17", "%Y-%m-%d") -> 1697500800

-- Time arithmetic
os.add_time(timestamp, duration)
-- Example: os.add_time(os.now(), {days=1, hours=2}) -> future timestamp

-- Human-readable duration
os.humanize_duration(seconds)
-- Example: os.humanize_duration(3661) -> "1 hour 1 minute 1 second"
```

### 5. Miscellaneous Utilities
```lua
local core = {}

-- Type checking utilities
core.is_callable(value) -- Check if value is callable (function or table with __call)
core.is_array(value) -- Check if value is array-like table
core.is_object(value) -- Check if value is object-like table

-- Function utilities
core.debounce(func, delay) -- Create debounced function
core.throttle(func, delay) -- Create throttled function
core.memoize(func) -- Create memoized function

-- Error handling
core.try(func, catch_func) -- Try-catch pattern
core.safe_call(func, ...) -- Call function safely, return success, result

-- Export the module
return core
```

## Implementation Strategy

1. **Reuse Existing Code**: 
   - UUID generation uses `testing.data.uuid()`
   - Random strings use `testing.data.random_string()`
   - Deep copy uses `data.clone()`
   - Merge uses `data.merge()`

2. **Native Extensions**:
   - Extend Lua's built-in tables (string, table, os)
   - Maintain backward compatibility

3. **Performance Considerations**:
   - Use native Lua patterns where possible
   - Implement efficient algorithms for common operations
   - Avoid unnecessary table allocations

4. **Error Handling**:
   - Clear error messages for invalid inputs
   - Type validation for all parameters
   - Graceful degradation where appropriate

## Integration with Other Modules

- **data.lua**: Reuse merge and clone functions
- **testing.lua**: Reuse UUID and random string generation
- **crypto bridge**: May need to access crypto functions from go-llms
- **time functions**: May need to use Go's time package for complex operations

## Testing Requirements

1. **String utilities**: Test edge cases, Unicode handling, empty strings
2. **Table utilities**: Test with nested tables, metatables, circular references
3. **Crypto functions**: Test all hash algorithms, encoding/decoding
4. **Time utilities**: Test timezone handling, DST, various formats
5. **Performance**: Benchmark critical functions
6. **Thread safety**: Test concurrent usage where applicable