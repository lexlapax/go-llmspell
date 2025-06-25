-- ABOUTME: Utility helpers for common operations using built-in tools
-- ABOUTME: Provides convenient wrappers around file, system, and other tool operations

local utils = {}

-- Get the tools module for executing built-in tools
local tools = require("tools")

-- ============================================================================
-- File Operations (using file_* built-in tools)
-- ============================================================================

-- Check if a file exists
-- @param path string: The file path to check
-- @return boolean: true if file exists, false otherwise
function utils.file_exists(path)
    if not path or path == "" then
        return false
    end
    
    -- Try to read 1 byte from the file
    local result = tools.execute_safe("file_read", {
        path = path,
        max_size = 1
    }, {silent = true})
    
    -- File exists if we got a result without error
    return result ~= nil and result.error == nil
end

-- Write content to a file
-- @param path string: The file path to write to
-- @param content string: The content to write
-- @param append boolean: Optional, append to file instead of overwrite
-- @return table: Result object with success/error
function utils.file_write(path, content, append)
    if not path or path == "" then
        return {error = "Path is required"}
    end
    
    return tools.execute_safe("file_write", {
        path = path,
        content = content or "",
        append = append or false,
        create_dirs = true,  -- Automatically create parent directories
        atomic = true        -- Use atomic writes for safety
    })
end

-- Read content from a file
-- @param path string: The file path to read
-- @param opts table: Optional parameters {max_size, line_start, line_end, include_meta}
-- @return string|nil, string|nil: Content and error (if any)
function utils.file_read(path, opts)
    if not path or path == "" then
        return nil, "Path is required"
    end
    
    opts = opts or {}
    local result = tools.execute_safe("file_read", {
        path = path,
        max_size = opts.max_size or 0,  -- 0 = unlimited
        line_start = opts.line_start,
        line_end = opts.line_end,
        include_meta = opts.include_meta or false
    })
    
    if result and not result.error then
        return result.content, nil
    else
        return nil, result and result.error or "Unknown error"
    end
end

-- Create a directory
-- @param path string: The directory path to create
-- @return table: Result object with success/error
function utils.mkdir(path)
    if not path or path == "" then
        return {error = "Path is required"}
    end
    
    -- Create directory by writing a .keep file
    return tools.execute_safe("file_write", {
        path = path .. "/.keep",
        content = "",
        create_dirs = true
    })
end

-- List files in a directory
-- @param path string: The directory path to list
-- @param pattern string: Optional glob pattern to filter files
-- @return table|nil, string|nil: File list and error (if any)
function utils.list_files(path, pattern)
    if not path or path == "" then
        return nil, "Path is required"
    end
    
    local params = {path = path}
    if pattern then
        params.pattern = pattern
    end
    
    local result = tools.execute_safe("file_list", params)
    
    if result and not result.error then
        return result.files or result.entries or {}, nil
    else
        return nil, result and result.error or "Unknown error"
    end
end

-- ============================================================================
-- System Operations (using system_* built-in tools)
-- ============================================================================

-- Sleep for specified seconds
-- @param seconds number: Number of seconds to sleep
-- @return table: Result object
function utils.sleep(seconds)
    if type(seconds) ~= "number" or seconds < 0 then
        return {error = "Invalid sleep duration"}
    end
    
    -- Use execute_command to run sleep command
    return tools.execute_safe("execute_command", {
        command = string.format("sleep %g", seconds),
        shell = "sh",
        timeout = math.ceil(seconds * 1000) + 1000  -- Add 1 second buffer
    })
end

-- Get environment variable
-- @param name string: Environment variable name
-- @return string|nil: Variable value or nil if not found
function utils.env(name)
    if not name or name == "" then
        return nil
    end
    
    local result = tools.execute_safe("get_environment_variable", {
        name = name
    }, {silent = true})
    
    if result and result.variables and #result.variables > 0 then
        return result.variables[1].value
    end
    
    return nil
end

-- Execute a system command
-- @param command string: Command to execute
-- @param opts table: Optional parameters {working_dir, environment, timeout, shell, input}
-- @return table: Result with stdout, stderr, exit_code, success
function utils.exec(command, opts)
    if not command or command == "" then
        return {error = "Command is required"}
    end
    
    opts = opts or {}
    
    return tools.execute_safe("execute_command", {
        command = command,
        working_dir = opts.working_dir,
        environment = opts.environment,
        timeout = opts.timeout or 30000,  -- Default 30 seconds
        shell = opts.shell or "sh",
        safe_mode = opts.safe_mode,
        input = opts.input
    })
end

-- Get system information
-- @param include_all boolean: Include all available information
-- @return table: System information
function utils.system_info(include_all)
    return tools.execute_safe("get_system_info", {
        include_environment = include_all,
        include_memory = include_all,
        include_runtime = include_all
    })
end

-- ============================================================================
-- Time Operations
-- ============================================================================

-- Get current timestamp
-- @return number: Current Unix timestamp
function utils.current_time()
    -- Use os.time() which should be available
    return os.time()
end

-- Format time
-- @param time number: Unix timestamp (optional, defaults to current time)
-- @param format string: Format string (optional, defaults to ISO format)
-- @return string: Formatted time string
function utils.format_time(time, format)
    time = time or os.time()
    format = format or "%Y-%m-%d %H:%M:%S"
    return os.date(format, time)
end

-- ============================================================================
-- Utility Functions
-- ============================================================================

-- Check if running on Windows
-- @return boolean: true if Windows, false otherwise
function utils.is_windows()
    local info = utils.system_info(false)
    if info and info.os then
        return info.os.name == "windows"
    end
    -- Fallback to checking path separator
    return package.config:sub(1,1) == '\\'
end

-- Join path components
-- @param ... string: Path components to join
-- @return string: Joined path
function utils.join_path(...)
    local sep = utils.is_windows() and '\\' or '/'
    local parts = {...}
    local result = {}
    
    for _, part in ipairs(parts) do
        if part and part ~= "" then
            -- Remove trailing separator from part
            part = part:gsub("[/\\]+$", "")
            table.insert(result, part)
        end
    end
    
    return table.concat(result, sep)
end

-- Get file extension
-- @param path string: File path
-- @return string: File extension (including dot) or empty string
function utils.file_extension(path)
    if not path then return "" end
    local ext = path:match("%.([^%.]+)$")
    return ext and ("." .. ext) or ""
end

-- Get file basename (filename without directory)
-- @param path string: File path
-- @return string: Basename
function utils.basename(path)
    if not path then return "" end
    return path:match("([^/\\]+)$") or path
end

-- Get directory name from path
-- @param path string: File path
-- @return string: Directory path
function utils.dirname(path)
    if not path then return "." end
    local dir = path:match("(.+)[/\\][^/\\]+$")
    return dir or "."
end

return utils