# gmacs Configuration Examples

This directory contains example configuration files for gmacs.

## Getting Started

1. Copy the example configuration to your config directory:
   ```bash
   cp examples/init.lua.example ~/.config/gmacs/init.lua
   ```

2. Edit the configuration file to customize gmacs:
   ```bash
   $EDITOR ~/.config/gmacs/init.lua
   ```

3. Restart gmacs to load the new configuration.

## Configuration Files

### `init.lua.example`
Basic configuration template with:
- Theme and editor settings
- Common key bindings
- Custom command examples
- File type hooks (planned)
- Package management examples (planned)

## Lua API Reference

### Settings
- `gmacs.set_variable(key, value)` - Set a configuration variable
- `gmacs.get_variable(key)` - Get a configuration variable

### Key Bindings
- `gmacs.global_set_key(key_sequence, command)` - Set global key binding
- `gmacs.local_set_key(key_sequence, command)` - Set local key binding (mode-specific)

### Commands
- `gmacs.register_command(name, function, description)` - Register a custom command
- `gmacs.execute_command(command_name)` - Execute a command
- `gmacs.list_commands()` - List all available commands

### Messages
- `gmacs.message(text)` - Display a message in the minibuffer
- `gmacs.error(text)` - Display an error message

### Buffer/Editor
- `gmacs.current_buffer()` - Get current buffer information
- `gmacs.current_word()` - Get word at cursor
- `gmacs.current_char()` - Get character at cursor

### Hooks (Planned)
- `gmacs.add_hook(hook_name, function)` - Add a hook function

### Package Management (Planned)
- `gmacs.use_package(package_name, version)` - Use a package

## Example Customizations

### Custom Delete Function
```lua
function smart_delete()
    local char = gmacs.current_char()
    if char == " " then
        -- Delete whitespace
        while gmacs.current_char() == " " do
            gmacs.execute_command("delete-char")
        end
    else
        -- Delete single character
        gmacs.execute_command("delete-char")
    end
end

gmacs.register_command("smart-delete", smart_delete)
gmacs.global_set_key("C-d", "smart-delete")
```

### Conditional Key Bindings
```lua
-- Different bindings for different systems
if os.getenv("OS") and os.getenv("OS"):match("Windows") then
    gmacs.global_set_key("C-z", "undo")  -- Windows style
else
    gmacs.global_set_key("C-/", "undo")  -- Unix style
end
```

### File Type Specific Settings
```lua
function setup_programming_mode()
    gmacs.set_variable("show-line-numbers", true, "local")
    gmacs.set_variable("highlight-current-line", true, "local")
end

-- This will be available when mode system is implemented
-- gmacs.add_hook("prog-mode-hook", setup_programming_mode)
```