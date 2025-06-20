-- ABOUTME: Event & Hooks Library for go-llmspell Lua standard library
-- ABOUTME: Provides event emitter pattern, hooks system, and lifecycle management

local events = {}

-- No need for compatibility - gopher-lua uses table.unpack

-- Import promise library for async operations
local promise = _G.promise or require("promise")

-- Internal tracking
local global_emitter = nil -- Global event emitter

-- EventEmitter class
local EventEmitter = {}
EventEmitter.__index = EventEmitter

-- Create a new EventEmitter instance
function EventEmitter.new()
    local instance = setmetatable({
        _events = {}, -- event name -> array of handlers
        _once_handlers = {}, -- track one-time handlers
        _max_listeners = 10, -- default max listeners per event
        _warning_printed = {}, -- track warning state per event
    }, EventEmitter)
    return instance
end

-- Add an event listener
function EventEmitter:on(event, handler)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type(handler) ~= "function" then
        error("Handler must be a function")
    end

    -- Initialize event handler list if needed
    if not self._events[event] then
        self._events[event] = {}
    end

    -- Add handler
    table.insert(self._events[event], handler)

    -- Check max listeners
    local count = #self._events[event]
    if count > self._max_listeners and not self._warning_printed[event] then
        print(
            string.format(
                "Warning: Event '%s' has %d listeners. Possible memory leak?",
                event,
                count
            )
        )
        self._warning_printed[event] = true
    end

    return self
end

-- Add a one-time event listener
function EventEmitter:once(event, handler)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type(handler) ~= "function" then
        error("Handler must be a function")
    end

    -- Create wrapper that removes itself after execution
    local function wrapper(...)
        self:off(event, wrapper)
        return handler(...)
    end

    -- Track the wrapper for this handler
    if not self._once_handlers[handler] then
        self._once_handlers[handler] = {}
    end
    self._once_handlers[handler][event] = wrapper

    -- Add the wrapper
    return self:on(event, wrapper)
end

-- Remove an event listener
function EventEmitter:off(event, handler)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end

    if not self._events[event] then
        return self
    end

    if handler == nil then
        -- Remove all handlers for this event
        self._events[event] = nil
        -- Clean up once handlers
        for _, event_map in pairs(self._once_handlers) do
            event_map[event] = nil
        end
    else
        -- Remove specific handler
        local handlers = self._events[event]
        for i = #handlers, 1, -1 do
            if handlers[i] == handler then
                table.remove(handlers, i)
            end
        end

        -- Also check for once wrappers
        if self._once_handlers[handler] and self._once_handlers[handler][event] then
            local wrapper = self._once_handlers[handler][event]
            for i = #handlers, 1, -1 do
                if handlers[i] == wrapper then
                    table.remove(handlers, i)
                end
            end
            self._once_handlers[handler][event] = nil
        end

        -- Clean up empty handler list
        if #handlers == 0 then
            self._events[event] = nil
        end
    end

    return self
end

-- Emit an event
function EventEmitter:emit(event, ...)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end

    if not self._events[event] then
        return false
    end

    -- Make a copy of handlers to avoid issues with modifications during emit
    local handlers = {}
    for i, handler in ipairs(self._events[event]) do
        handlers[i] = handler
    end

    local args = { ... }
    local has_listeners = #handlers > 0

    -- Call each handler
    for i = 1, #handlers do
        local handler = handlers[i]
        if handler and type(handler) == "function" then
            -- Protected call to prevent one handler from breaking others
            local success, err
            if #args == 0 then
                success, err = pcall(handler)
            elseif #args == 1 then
                success, err = pcall(handler, args[1])
            elseif #args == 2 then
                success, err = pcall(handler, args[1], args[2])
            elseif #args == 3 then
                success, err = pcall(handler, args[1], args[2], args[3])
            else
                -- For more args, use a wrapper function
                success, err = pcall(function()
                    local result = { handler(args[1], args[2], args[3], args[4], args[5]) }
                    return table.unpack(result)
                end)
            end
            if not success then
                -- Emit error event if handler fails
                if event ~= "error" then
                    -- Schedule error emission to avoid recursive call during iteration
                    local error_args = { err, event, handler }
                    -- Use a simple timeout to emit error after current iteration
                    if self._events["error"] then
                        -- Call error handlers directly to avoid iteration issues
                        local error_handlers = {}
                        for j, h in ipairs(self._events["error"]) do
                            error_handlers[j] = h
                        end
                        for _, error_handler in ipairs(error_handlers) do
                            if type(error_handler) == "function" then
                                pcall(error_handler, error_args[1], error_args[2], error_args[3])
                            end
                        end
                    end
                else
                    -- Prevent infinite recursion on error event
                    print("Error in error handler: " .. tostring(err))
                end
            end
        end
    end

    return has_listeners
end

-- Get listener count for an event
function EventEmitter:listenerCount(event)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end

    if not self._events[event] then
        return 0
    end

    return #self._events[event]
end

-- Get all listeners for an event
function EventEmitter:listeners(event)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end

    if not self._events[event] then
        return {}
    end

    -- Return a copy to prevent external modification
    local copy = {}
    for i, handler in ipairs(self._events[event]) do
        copy[i] = handler
    end
    return copy
end

-- Remove all listeners
function EventEmitter:removeAllListeners(event)
    if event then
        self._events[event] = nil
    else
        self._events = {}
        self._once_handlers = {}
    end
    return self
end

-- Set max listeners (0 = unlimited)
function EventEmitter:setMaxListeners(n)
    if type(n) ~= "number" or n < 0 then
        error("Max listeners must be a non-negative number")
    end
    self._max_listeners = n
    return self
end

-- Get event names
function EventEmitter:eventNames()
    local names = {}
    for event in pairs(self._events) do
        table.insert(names, event)
    end
    return names
end

-- Hook System
local hooks = {}
local hook_handlers = {
    before = {}, -- event -> array of before handlers
    after = {}, -- event -> array of after handlers
    around = {}, -- event -> array of around handlers
}

-- Add a before hook
function hooks.before(event, handler)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type(handler) ~= "function" then
        error("Handler must be a function")
    end

    if not hook_handlers.before[event] then
        hook_handlers.before[event] = {}
    end

    table.insert(hook_handlers.before[event], handler)
    return #hook_handlers.before[event] -- Return handler ID
end

-- Add an after hook
function hooks.after(event, handler)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type(handler) ~= "function" then
        error("Handler must be a function")
    end

    if not hook_handlers.after[event] then
        hook_handlers.after[event] = {}
    end

    table.insert(hook_handlers.after[event], handler)
    return #hook_handlers.after[event] -- Return handler ID
end

-- Add an around hook
function hooks.around(event, wrapper)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type(wrapper) ~= "function" then
        error("Wrapper must be a function")
    end

    if not hook_handlers.around[event] then
        hook_handlers.around[event] = {}
    end

    table.insert(hook_handlers.around[event], wrapper)
    return #hook_handlers.around[event] -- Return handler ID
end

-- Execute hooks for an event
function hooks.execute(event, fn, ...)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type(fn) ~= "function" then
        error("Function must be a function")
    end

    local args = { ... }

    -- Execute before hooks
    if hook_handlers.before[event] then
        for _, handler in ipairs(hook_handlers.before[event]) do
            local success, result
            if #args == 0 then
                success, result = pcall(handler)
            elseif #args == 1 then
                success, result = pcall(handler, args[1])
            elseif #args == 2 then
                success, result = pcall(handler, args[1], args[2])
            elseif #args == 3 then
                success, result = pcall(handler, args[1], args[2], args[3])
            else
                success, result = pcall(handler, args[1], args[2], args[3], args[4], args[5])
            end
            if not success then
                error("Before hook failed: " .. tostring(result))
            end
            -- Before hooks can modify arguments
            if result ~= nil then
                args = { result }
            end
        end
    end

    -- Build the execution chain with around hooks
    local exec_fn = fn
    if hook_handlers.around[event] then
        -- Apply around hooks in reverse order (last registered wraps first)
        for i = #hook_handlers.around[event], 1, -1 do
            local wrapper = hook_handlers.around[event][i]
            local prev_fn = exec_fn
            exec_fn = function(...)
                return wrapper(prev_fn, ...)
            end
        end
    end

    -- Execute the function
    local results
    if #args == 0 then
        results = { pcall(exec_fn) }
    elseif #args == 1 then
        results = { pcall(exec_fn, args[1]) }
    elseif #args == 2 then
        results = { pcall(exec_fn, args[1], args[2]) }
    elseif #args == 3 then
        results = { pcall(exec_fn, args[1], args[2], args[3]) }
    else
        results = { pcall(exec_fn, args[1], args[2], args[3], args[4], args[5]) }
    end
    local success = table.remove(results, 1)

    if not success then
        error("Function execution failed: " .. tostring(results[1]))
    end

    -- Execute after hooks
    if hook_handlers.after[event] then
        for _, handler in ipairs(hook_handlers.after[event]) do
            local hook_success, hook_result
            if #results == 0 then
                hook_success, hook_result = pcall(handler)
            elseif #results == 1 then
                hook_success, hook_result = pcall(handler, results[1])
            elseif #results == 2 then
                hook_success, hook_result = pcall(handler, results[1], results[2])
            elseif #results == 3 then
                hook_success, hook_result = pcall(handler, results[1], results[2], results[3])
            else
                hook_success, hook_result =
                    pcall(handler, results[1], results[2], results[3], results[4], results[5])
            end
            if not hook_success then
                error("After hook failed: " .. tostring(hook_result))
            end
            -- After hooks can modify results
            if hook_result ~= nil then
                results = { hook_result }
            end
        end
    end

    if #results == 0 then
        return
    elseif #results == 1 then
        return results[1]
    elseif #results == 2 then
        return results[1], results[2]
    elseif #results == 3 then
        return results[1], results[2], results[3]
    else
        return results[1], results[2], results[3], results[4], results[5]
    end
end

-- Remove a hook
function hooks.remove(event, type, handler_id)
    if type(event) ~= "string" then
        error("Event name must be a string")
    end
    if type ~= "before" and type ~= "after" and type ~= "around" then
        error("Hook type must be 'before', 'after', or 'around'")
    end
    if type(handler_id) ~= "number" then
        error("Handler ID must be a number")
    end

    if hook_handlers[type][event] and hook_handlers[type][event][handler_id] then
        hook_handlers[type][event][handler_id] = nil
        return true
    end

    return false
end

-- Clear all hooks for an event
function hooks.clear(event, type)
    if event then
        if type then
            if hook_handlers[type] then
                hook_handlers[type][event] = nil
            end
        else
            -- Clear all types for this event
            hook_handlers.before[event] = nil
            hook_handlers.after[event] = nil
            hook_handlers.around[event] = nil
        end
    else
        -- Clear all hooks
        if type then
            hook_handlers[type] = {}
        else
            hook_handlers.before = {}
            hook_handlers.after = {}
            hook_handlers.around = {}
        end
    end
end

-- Global Event System Functions

-- Get or create the global event emitter
local function get_global_emitter()
    if not global_emitter then
        global_emitter = EventEmitter.new()
        global_emitter:setMaxListeners(0) -- Unlimited for global
    end
    return global_emitter
end

-- Emit a global event
function events.emit(event, ...)
    return get_global_emitter():emit(event, ...)
end

-- Subscribe to global events
function events.on(event, handler)
    return get_global_emitter():on(event, handler)
end

-- Subscribe once to global events
function events.once(event, handler)
    return get_global_emitter():once(event, handler)
end

-- Unsubscribe from global events
function events.off(event, handler)
    return get_global_emitter():off(event, handler)
end

-- Create a new event emitter
function events.create_emitter()
    return EventEmitter.new()
end

-- Advanced Event Features

-- Wait for an event (returns a promise)
function events.wait_for(emitter_or_event, event_or_timeout, timeout)
    -- Handle overloaded parameters
    local emitter, event_name, timeout_ms

    if type(emitter_or_event) == "string" then
        -- Called as events.wait_for(event, timeout)
        emitter = get_global_emitter()
        event_name = emitter_or_event
        timeout_ms = event_or_timeout
    else
        -- Called as events.wait_for(emitter, event, timeout)
        emitter = emitter_or_event
        event_name = event_or_timeout
        timeout_ms = timeout
    end

    return promise.Promise.new(function(resolve, reject)
        local handler
        local timer_handle
        local resolved = false

        -- Create handler that resolves the promise
        handler = function(...)
            if not resolved then
                resolved = true
                if timer_handle then
                    timer_handle.cancelled = true
                end
                resolve({ ... })
            end
        end

        -- Set up one-time listener
        emitter:once(event_name, handler)

        -- Set up timeout if specified
        if timeout_ms and timeout_ms > 0 then
            timer_handle = { cancelled = false }
            promise.spawn(function()
                promise.sleep(timeout_ms)
                if not resolved and not timer_handle.cancelled then
                    resolved = true
                    -- Remove handler on timeout
                    emitter:off(event_name, handler)
                    reject("Timeout waiting for event: " .. event_name)
                end
            end)
        end
    end)
end

-- Event aggregation
function events.aggregate(emitter_or_events, events_or_timeout, timeout)
    -- Handle overloaded parameters
    local emitter, event_list, timeout_ms

    if type(emitter_or_events) == "table" and not emitter_or_events.emit then
        -- Called as events.aggregate(events, timeout)
        emitter = get_global_emitter()
        event_list = emitter_or_events
        timeout_ms = events_or_timeout
    else
        -- Called as events.aggregate(emitter, events, timeout)
        emitter = emitter_or_events
        event_list = events_or_timeout
        timeout_ms = timeout
    end

    return promise.Promise.new(function(resolve, reject)
        local results = {}
        local remaining = #event_list
        local completed = false
        local timer_handle

        -- Create handler for each event
        for _, event_name in ipairs(event_list) do
            emitter:once(event_name, function(...)
                if not completed then
                    results[event_name] = { ... }
                    remaining = remaining - 1

                    if remaining == 0 then
                        completed = true
                        if timer_handle then
                            timer_handle.cancelled = true
                        end
                        resolve(results)
                    end
                end
            end)
        end

        -- Set up timeout if specified
        if timeout_ms and timeout_ms > 0 then
            timer_handle = { cancelled = false }
            promise.spawn(function()
                promise.sleep(timeout_ms)
                if not completed and not timer_handle.cancelled then
                    completed = true
                    reject("Timeout waiting for events")
                end
            end)
        end
    end)
end

-- Event filtering with pattern matching
function events.filter(emitter_or_pattern, pattern_or_handler, handler)
    -- Handle overloaded parameters
    local emitter, pattern, callback

    if type(pattern_or_handler) == "function" then
        -- Called as events.filter(pattern, handler)
        emitter = get_global_emitter()
        pattern = emitter_or_pattern
        callback = pattern_or_handler
    else
        -- Called as events.filter(emitter, pattern, handler)
        emitter = emitter_or_pattern
        pattern = pattern_or_handler
        callback = handler
    end

    -- Convert pattern to regex if it's a string with wildcards
    local regex_pattern
    if pattern:find("*") then
        -- Convert wildcard pattern to Lua pattern
        regex_pattern = "^"
            .. pattern:gsub("([%.%+%-%*%?%[%]%^%$%(%)%%])", "%%%1"):gsub("%%%*", ".*")
            .. "$"
    else
        regex_pattern = "^" .. pattern .. "$"
    end

    -- Subscribe to all events and filter
    local original_emit = emitter.emit
    emitter.emit = function(self, event, ...)
        if event:match(regex_pattern) then
            callback(event, ...)
        end
        return original_emit(self, event, ...)
    end

    -- Return unsubscribe function
    return function()
        emitter.emit = original_emit
    end
end

-- Event namespacing
function events.namespace(namespace)
    if type(namespace) ~= "string" then
        error("Namespace must be a string")
    end

    local emitter = EventEmitter.new()

    -- Override emit to add namespace
    local original_emit = emitter.emit
    emitter.emit = function(self, event, ...)
        return original_emit(self, namespace .. ":" .. event, ...)
    end

    -- Override on/once/off to add namespace
    local original_on = emitter.on
    emitter.on = function(self, event, handler)
        return original_on(self, namespace .. ":" .. event, handler)
    end

    local original_once = emitter.once
    emitter.once = function(self, event, handler)
        return original_once(self, namespace .. ":" .. event, handler)
    end

    local original_off = emitter.off
    emitter.off = function(self, event, handler)
        return original_off(self, namespace .. ":" .. event, handler)
    end

    return emitter
end

-- Export event emitter class
events.EventEmitter = EventEmitter

-- Export hooks
events.hooks = hooks

-- Bridge integration helpers
events.bridge = {}

-- Use bridge event system if available
function events.bridge.emit(event_type, data)
    if _G.events then
        -- Create event object
        local event = {
            type = event_type,
            timestamp = os.time(),
            data = data,
        }

        -- Publish through bridge
        return _G.events:publishEvent(event)
    else
        -- Fallback to local emission
        return events.emit(event_type, data)
    end
end

-- Subscribe through bridge if available
function events.bridge.subscribe(pattern, handler)
    if _G.events then
        return _G.events:subscribe(pattern, handler)
    else
        -- Fallback to local subscription with pattern support
        return events.filter(pattern, handler)
    end
end

-- Export the module
return events
