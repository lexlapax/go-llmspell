-- Minimal agent module for testing
local agent = {}

function agent.create(name, config)
    print("In agent.create, name:", name)
    print("In agent.create, config:", config)
    
    -- Just return a simple table
    return {
        id = "test_id",
        name = name,
        run = function(self, input)
            return "Agent " .. self.name .. " ran with input: " .. tostring(input)
        end
    }
end

return agent