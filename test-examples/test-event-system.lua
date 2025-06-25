-- Minimal EventSystem test
local EventSystem = {}

function EventSystem:new()
    local obj = {
        handlers = {}
    }
    setmetatable(obj, {__index = self})
    return obj
end

function EventSystem:on(event_name, handler)
    if not self.handlers[event_name] then
        self.handlers[event_name] = {}
    end
    table.insert(self.handlers[event_name], handler)
end

-- Test it
local system = EventSystem:new()
system:on("test", function() print("Handler called") end)

print("EventSystem test successful")
return "Success"