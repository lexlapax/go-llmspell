-- Simple test without file operations
local result = {
    status = "success",
    message = "Basic Lua execution works",
    calculations = {
        sum = 10 + 20,
        product = 5 * 6,
        concat = "Hello " .. "World"
    }
}

return result