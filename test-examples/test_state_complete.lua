-- ABOUTME: Complete test of state management functionality
-- ABOUTME: Work around from_json issue by using parse_json

-- Load modules
local state = require("state")
local data = require("data")

-- Initialize state
state.set("app.user_preferences", {
    theme = "dark",
    notifications = true,
    language = "en"
})

state.set("app.session_data", {
    user_id = "user123",
    login_time = os.time(),
    last_activity = os.time()
})

-- Test state operations
local theme = state.get("app.user_preferences.theme")
local has_prefs = state.has("app.user_preferences")
local app_keys = state.keys("app")

-- Save state to JSON
local current_state = state.get("app")
local state_json = data.to_json(current_state)
local json_size = state_json and #state_json or 0

-- Clear and restore
state.clear("app")
local cleared = state.get("app") == nil

-- Restore using parse_json (not from_json)
local restored_state = data.parse_json(state_json)
state.set("app", restored_state)

-- Verify restoration
local restored_theme = state.get("app.user_preferences.theme")
local restoration_success = restored_theme == "dark"

return {
    initial_theme = theme,
    has_prefs = has_prefs,
    app_keys_count = app_keys and #app_keys or 0,
    json_size = json_size,
    cleared = cleared,
    restored_theme = restored_theme,
    restoration_success = restoration_success,
    test_complete = true
}