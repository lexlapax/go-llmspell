-- ABOUTME: Promise & Async Library for go-llmspell Lua standard library
-- ABOUTME: Provides Promise class, async/await syntax sugar, and coroutine integration

local promise = {}

-- Promise class implementation
local Promise = {}
Promise.__index = Promise

-- Promise states
local PENDING = "pending"
local FULFILLED = "fulfilled"
local REJECTED = "rejected"

-- Create a new Promise with an executor function
-- executor: function(resolve, reject) - the function to execute
function Promise.new(executor)
    if type(executor) ~= "function" then
        error("Promise executor must be a function")
    end
    
    local self = setmetatable({
        state = PENDING,
        value = nil,
        reason = nil,
        handlers = {}
    }, Promise)
    
    -- Resolve function
    local function resolve(value)
        if self.state == PENDING then
            self.state = FULFILLED
            self.value = value
            self:_handleChain()
        end
    end
    
    -- Reject function
    local function reject(reason)
        if self.state == PENDING then
            self.state = REJECTED
            self.reason = reason
            self:_handleChain()
        end
    end
    
    -- Execute the executor function
    local success, err = pcall(executor, resolve, reject)
    if not success then
        reject(err)
    end
    
    return self
end

-- Handle promise chain (then/catch)
function Promise:_handleChain()
    for _, handler in ipairs(self.handlers) do
        if self.state == FULFILLED and handler.onFulfilled then
            local success, result = pcall(handler.onFulfilled, self.value)
            if success then
                handler.resolve(result)
            else
                handler.reject(result)
            end
        elseif self.state == REJECTED and handler.onRejected then
            local success, result = pcall(handler.onRejected, self.reason)
            if success then
                handler.resolve(result)
            else
                handler.reject(result)
            end
        elseif self.state == REJECTED and not handler.onRejected then
            handler.reject(self.reason)
        end
    end
    self.handlers = {} -- Clear handlers after execution
end

-- Add success handler (andThen to avoid 'then' keyword conflict)
function Promise:andThen(onFulfilled, onRejected)
    return Promise.new(function(resolve, reject)
        local handler = {
            onFulfilled = onFulfilled,
            onRejected = onRejected,
            resolve = resolve,
            reject = reject
        }
        
        if self.state == PENDING then
            table.insert(self.handlers, handler)
        else
            -- Promise already settled, handle immediately
            if self.state == FULFILLED and onFulfilled then
                local success, result = pcall(onFulfilled, self.value)
                if success then
                    resolve(result)
                else
                    reject(result)
                end
            elseif self.state == REJECTED then
                if onRejected then
                    local success, result = pcall(onRejected, self.reason)
                    if success then
                        resolve(result)
                    else
                        reject(result)
                    end
                else
                    reject(self.reason)
                end
            else
                resolve(self.value)
            end
        end
    end)
end

-- Add error handler (onError to avoid catch keyword)
function Promise:onError(onRejected)
    return self:andThen(nil, onRejected)
end

-- Add finally handler  
function Promise:onFinally(onFinally)
    return self:andThen(
        function(value)
            if onFinally then onFinally() end
            return value
        end,
        function(reason)
            if onFinally then onFinally() end
            error(reason)
        end
    )
end

-- Promise.resolve - create a resolved promise
function Promise.resolve(value)
    return Promise.new(function(resolve)
        resolve(value)
    end)
end

-- Promise.reject - create a rejected promise
function Promise.reject(reason)
    return Promise.new(function(_, reject)
        reject(reason)
    end)
end

-- Promise.all - wait for all promises to resolve
function Promise.all(promises)
    if type(promises) ~= "table" then
        return Promise.reject("Promise.all requires a table of promises")
    end
    
    return Promise.new(function(resolve, reject)
        local results = {}
        local count = #promises
        local completed = 0
        
        if count == 0 then
            resolve({})
            return
        end
        
        for i, promise in ipairs(promises) do
            if type(promise) == "table" and promise.andThen then
                promise:andThen(function(value)
                    results[i] = value
                    completed = completed + 1
                    if completed == count then
                        resolve(results)
                    end
                end, function(reason)
                    reject(reason)
                end)
            else
                -- Not a promise, treat as resolved value
                results[i] = promise
                completed = completed + 1
                if completed == count then
                    resolve(results)
                end
            end
        end
    end)
end

-- Promise.race - resolve with the first promise that settles
function Promise.race(promises)
    if type(promises) ~= "table" then
        return Promise.reject("Promise.race requires a table of promises")
    end
    
    return Promise.new(function(resolve, reject)
        local settled = false
        
        for _, promise in ipairs(promises) do
            if type(promise) == "table" and promise.andThen then
                promise:andThen(function(value)
                    if not settled then
                        settled = true
                        resolve(value)
                    end
                end, function(reason)
                    if not settled then
                        settled = true
                        reject(reason)
                    end
                end)
            else
                -- Not a promise, resolve immediately
                if not settled then
                    settled = true
                    resolve(promise)
                end
                break
            end
        end
    end)
end

-- Async/await syntax sugar

-- async: wrap a function to return a promise
function promise.async(func)
    if type(func) ~= "function" then
        error("async() requires a function")
    end
    
    return function(...)
        local args = {...}
        return Promise.new(function(resolve, reject)
            local success, result = pcall(func, table.unpack(args))
            if success then
                resolve(result)
            else
                reject(result)
            end
        end)
    end
end

-- await: wait for a promise to resolve (with optional timeout)
function promise.await(promiseOrFunc, timeout)
    local targetPromise
    
    -- If it's a function, call it to get the promise
    if type(promiseOrFunc) == "function" then
        targetPromise = promiseOrFunc()
    else
        targetPromise = promiseOrFunc
    end
    
    if not (type(targetPromise) == "table" and targetPromise.andThen) then
        error("await() requires a promise")
    end
    
    -- Simple blocking implementation for now
    -- In a real implementation, this would yield to the coroutine scheduler
    local completed = false
    local result = nil
    local error_msg = nil
    
    targetPromise:andThen(function(value)
        result = value
        completed = true
    end, function(reason)
        error_msg = reason
        completed = true
    end)
    
    -- Simple spin-wait (not ideal for production)
    local start_time = os.clock()
    local timeout_seconds = timeout or 30 -- Default 30 second timeout
    
    while not completed do
        if timeout and (os.clock() - start_time) > timeout_seconds then
            error("Promise timeout after " .. timeout_seconds .. " seconds")
        end
        -- Yield briefly to prevent busy waiting
        coroutine.yield()
    end
    
    if error_msg then
        error(error_msg)
    end
    
    return result
end

-- sleep: create a promise that resolves after a delay
function promise.sleep(duration)
    if type(duration) ~= "number" or duration < 0 then
        error("sleep() requires a positive number (seconds)")
    end
    
    return Promise.new(function(resolve)
        -- In a real implementation, this would use the Go async runtime
        -- For now, use a simple timer approach
        local start_time = os.clock()
        while (os.clock() - start_time) < duration do
            coroutine.yield()
        end
        resolve()
    end)
end

-- Coroutine integration helpers

-- spawn: create a new coroutine for concurrent execution
function promise.spawn(func, ...)
    if type(func) ~= "function" then
        error("spawn() requires a function")
    end
    
    local args = {...}
    return Promise.new(function(resolve, reject)
        local co = coroutine.create(function()
            local success, result = pcall(func, table.unpack(args))
            if success then
                resolve(result)
            else
                reject(result)
            end
        end)
        
        -- Start the coroutine
        local success, err = coroutine.resume(co)
        if not success then
            reject(err)
        end
    end)
end

-- yield: cooperative yield for multitasking
function promise.yield()
    coroutine.yield()
end

-- Channel-based communication helpers (simplified)
local channels = {}

-- create_channel: create a new communication channel
function promise.create_channel(name, buffer_size)
    if type(name) ~= "string" then
        error("create_channel() requires a string name")
    end
    
    buffer_size = buffer_size or 0
    channels[name] = {
        buffer = {},
        buffer_size = buffer_size,
        waiting_senders = {},
        waiting_receivers = {}
    }
    
    return name
end

-- send: send a value through a channel
function promise.send(channel_name, value)
    local channel = channels[channel_name]
    if not channel then
        error("Channel not found: " .. tostring(channel_name))
    end
    
    return Promise.new(function(resolve, reject)
        if #channel.buffer < channel.buffer_size then
            table.insert(channel.buffer, value)
            resolve()
        else
            table.insert(channel.waiting_senders, {value = value, resolve = resolve})
        end
    end)
end

-- receive: receive a value from a channel
function promise.receive(channel_name)
    local channel = channels[channel_name]
    if not channel then
        error("Channel not found: " .. tostring(channel_name))
    end
    
    return Promise.new(function(resolve, reject)
        if #channel.buffer > 0 then
            local value = table.remove(channel.buffer, 1)
            resolve(value)
            
            -- Process waiting senders
            if #channel.waiting_senders > 0 then
                local sender = table.remove(channel.waiting_senders, 1)
                table.insert(channel.buffer, sender.value)
                sender.resolve()
            end
        else
            table.insert(channel.waiting_receivers, resolve)
        end
    end)
end

-- close_channel: close a communication channel
function promise.close_channel(channel_name)
    channels[channel_name] = nil
end

-- Export the Promise class and utility functions
promise.Promise = Promise
return promise