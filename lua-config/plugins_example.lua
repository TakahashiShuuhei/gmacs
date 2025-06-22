-- gmacs Plugin Configuration Example
-- 
-- This is an EXAMPLE file demonstrating plugin configuration syntax.
-- The actual plugin configuration should be placed in one of these locations:
--   ~/.config/gmacs/plugins.lua  (XDG compliant)
--   ~/.gmacs/plugins.lua         (legacy location)
--
-- This file is used for testing and as a reference for plugin API usage.

-- Plugin configuration examples

-- Load plugins on startup
--[[ 
gmacs.load_plugin("example-plugin")
gmacs.load_plugin("syntax-highlighter")
]]

-- Check if a plugin is loaded
--[[
if gmacs.plugin_loaded("example-plugin") then
    gmacs.message("Example plugin is loaded!")
end
]]

-- List all loaded plugins
--[[
local plugins = gmacs.list_plugins()
for i, plugin in ipairs(plugins) do
    print("Plugin: " .. plugin.name .. " v" .. plugin.version)
end
]]

-- Example plugin configuration
--[[
if gmacs.plugin_loaded("syntax-highlighter") then
    -- Configure syntax highlighter plugin
    gmacs.set_option("syntax.theme", "dark")
    gmacs.set_option("syntax.line_numbers", true)
end
]]

-- Plugin-specific key bindings
--[[
if gmacs.plugin_loaded("file-tree") then
    gmacs.bind_key("C-x t", "toggle-file-tree")
end
]]

-- Auto-load plugins based on file type
--[[
gmacs.add_hook("file-opened", function(filename)
    if string.match(filename, "%.go$") then
        if not gmacs.plugin_loaded("go-mode") then
            gmacs.load_plugin("go-mode")
        end
    elseif string.match(filename, "%.py$") then
        if not gmacs.plugin_loaded("python-mode") then
            gmacs.load_plugin("python-mode")
        end
    end
end)
]]

-- Example of defining custom commands that use plugins
--[[
gmacs.defun("reload-plugins", function()
    -- Reload all currently loaded plugins
    local plugins = gmacs.list_plugins()
    for i, plugin in ipairs(plugins) do
        if plugin.enabled then
            gmacs.unload_plugin(plugin.name)
            gmacs.load_plugin(plugin.name)
        end
    end
    gmacs.message("Plugins reloaded!")
end)
]]

-- Plugin management commands
gmacs.defun("list-plugins-interactive", function()
    local plugins = gmacs.list_plugins()
    if #plugins == 0 then
        gmacs.message("No plugins loaded")
    else
        local message = "Loaded plugins: "
        for i, plugin in ipairs(plugins) do
            if i > 1 then
                message = message .. ", "
            end
            message = message .. plugin.name .. " (" .. plugin.version .. ")"
        end
        gmacs.message(message)
    end
end)

-- Enhanced plugin setup example
gmacs.defun("setup-example-plugin", function()
    -- Example of plugin configuration
    local success = gmacs.setup_plugin("example-plugin", {
        theme = "dark",
        line_numbers = true,
        auto_save = false,
        indent_size = 4
    })
    
    if success then
        gmacs.message("Example plugin configured successfully")
    else
        gmacs.message("Failed to configure example plugin")
    end
end)

-- Toggle plugin enabled/disabled state
gmacs.defun("toggle-plugin", function()
    local plugin_name = "example-plugin" -- This would be interactive in real usage
    
    if gmacs.plugin_loaded(plugin_name) then
        local success = gmacs.disable_plugin(plugin_name)
        if success then
            gmacs.message("Plugin " .. plugin_name .. " disabled")
        else
            gmacs.message("Failed to disable plugin " .. plugin_name)
        end
    else
        local success = gmacs.enable_plugin(plugin_name)
        if success then
            gmacs.message("Plugin " .. plugin_name .. " enabled")
        else
            gmacs.message("Failed to enable plugin " .. plugin_name)
        end
    end
end)

-- Get and display plugin configuration
gmacs.defun("show-plugin-config", function()
    local plugin_name = "example-plugin"
    local config = gmacs.get_plugin_config(plugin_name)
    
    if config then
        local message = "Plugin " .. plugin_name .. " configuration:"
        for key, value in pairs(config) do
            message = message .. "\n  " .. key .. " = " .. tostring(value)
        end
        gmacs.message(message)
    else
        gmacs.message("No configuration found for plugin " .. plugin_name)
    end
end)

-- Bind plugin management commands to keys
gmacs.bind_key("C-c p l", "list-plugins-interactive")
gmacs.bind_key("C-c p s", "setup-example-plugin")
gmacs.bind_key("C-c p t", "toggle-plugin")
gmacs.bind_key("C-c p c", "show-plugin-config")

-- Print startup message
gmacs.message("Plugin configuration loaded")