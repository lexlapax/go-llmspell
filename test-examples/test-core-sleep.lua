-- Test core.sleep
local core = require("core")

print("Type of core:", type(core))
print("Type of core.sleep:", type(core.sleep))

-- Test sleep
print("Before sleep:", os.time())
core.sleep(0.1)
print("After sleep:", os.time())

return "Success"