-- gmacs Plugin Configuration
-- This file is loaded after init.lua and provides plugin management functionality

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

-- Bind plugin management commands to keys
gmacs.bind_key("C-c p l", "list-plugins-interactive")

-- Print startup message
gmacs.message("Plugin configuration loaded")