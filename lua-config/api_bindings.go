package luaconfig

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/log"
)


// APIBindings manages the Lua API bindings for gmacs
type APIBindings struct {
	editor *domain.Editor
	vm     *LuaVM
}

// NewAPIBindings creates new API bindings
func NewAPIBindings(editor *domain.Editor, vm *LuaVM) *APIBindings {
	return &APIBindings{
		editor: editor,
		vm:     vm,
	}
}

// RegisterGmacsAPI registers all gmacs API functions in the Lua VM
func (api *APIBindings) RegisterGmacsAPI() error {
	if api.vm == nil || api.vm.GetState() == nil {
		return &ConfigError{Message: "Lua VM is not available"}
	}
	
	state := api.vm.GetState()
	L, ok := state.(*lua.LState)
	if !ok {
		return &ConfigError{Message: "Invalid Lua state"}
	}
	
	// Create gmacs table
	gmacsTable := L.NewTable()
	L.SetGlobal("gmacs", gmacsTable)
	
	// Register API functions
	L.SetField(gmacsTable, "bind_key", L.NewFunction(api.luaBindKey))
	L.SetField(gmacsTable, "local_bind_key", L.NewFunction(api.luaLocalBindKey))
	L.SetField(gmacsTable, "defun", L.NewFunction(api.luaDefun))
	L.SetField(gmacsTable, "set_option", L.NewFunction(api.luaSetOption))
	L.SetField(gmacsTable, "get_option", L.NewFunction(api.luaGetOption))
	L.SetField(gmacsTable, "major_mode", L.NewFunction(api.luaMajorMode))
	L.SetField(gmacsTable, "minor_mode", L.NewFunction(api.luaMinorMode))
	L.SetField(gmacsTable, "add_hook", L.NewFunction(api.luaAddHook))
	
	// Buffer and UI functions
	L.SetField(gmacsTable, "current_buffer", L.NewFunction(api.luaCurrentBuffer))
	L.SetField(gmacsTable, "message", L.NewFunction(api.luaMessage))
	L.SetField(gmacsTable, "toggle_minor_mode", L.NewFunction(api.luaToggleMinorMode))
	
	// Register all built-in commands
	api.registerBuiltinCommands()
	
	log.Info("Registered gmacs Lua API")
	return nil
}

// luaBindKey implements gmacs.bind_key(sequence, command)
func (api *APIBindings) luaBindKey(L *lua.LState) int {
	sequence := L.CheckString(1)
	command := L.CheckString(2)
	
	err := api.editor.BindKey(sequence, command)
	if err != nil {
		L.RaiseError("Error: " + err.Error())
		return 0 // This line won't be reached
	}
	
	log.Info("Lua: Bound key %s to %s", sequence, command)
	return 0
}

// luaLocalBindKey implements gmacs.local_bind_key(mode_name, sequence, command)
func (api *APIBindings) luaLocalBindKey(L *lua.LState) int {
	modeName := L.CheckString(1)
	sequence := L.CheckString(2)
	command := L.CheckString(3)
	
	err := api.editor.LocalBindKey(modeName, sequence, command)
	if err != nil {
		L.RaiseError("Error: " + err.Error())
		return 0 // This line won't be reached
	}
	
	log.Info("Lua: Bound key %s to %s in mode %s", sequence, command, modeName)
	return 0
}

// luaDefun implements gmacs.defun(name, function)
func (api *APIBindings) luaDefun(L *lua.LState) int {
	name := L.CheckString(1)
	fn := L.CheckFunction(2)
	
	// Create a wrapper function that calls the Lua function
	wrapper := func() error {
		err := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		})
		if err != nil {
			return &ConfigError{Message: "Lua function error: " + err.Error()}
		}
		return nil
	}
	
	err := api.editor.RegisterCommand(name, wrapper)
	if err != nil {
		L.Push(lua.LString("Error: " + err.Error()))
		return 1
	}
	
	log.Info("Lua: Registered command %s", name)
	return 0
}

// luaSetOption implements gmacs.set_option(name, value)
func (api *APIBindings) luaSetOption(L *lua.LState) int {
	name := L.CheckString(1)
	value := L.Get(2)
	
	var goValue interface{}
	switch value.Type() {
	case lua.LTString:
		goValue = lua.LVAsString(value)
	case lua.LTNumber:
		goValue = float64(lua.LVAsNumber(value))
	case lua.LTBool:
		goValue = lua.LVAsBool(value)
	default:
		L.Push(lua.LString("Error: Unsupported value type"))
		return 1
	}
	
	log.Info("Lua: About to set option %s = %v", name, goValue)
	err := api.editor.SetOption(name, goValue)
	if err != nil {
		log.Error("Lua: Failed to set option %s: %v", name, err)
		L.Push(lua.LString("Error: " + err.Error()))
		return 1
	}
	
	log.Info("Lua: Successfully set option %s = %v", name, goValue)
	return 0
}

// luaGetOption implements gmacs.get_option(name)
func (api *APIBindings) luaGetOption(L *lua.LState) int {
	name := L.CheckString(1)
	
	value, err := api.editor.GetOption(name)
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	
	// Convert Go value to Lua value
	switch v := value.(type) {
	case string:
		L.Push(lua.LString(v))
	case float64:
		L.Push(lua.LNumber(v))
	case bool:
		L.Push(lua.LBool(v))
	default:
		L.Push(lua.LNil)
	}
	
	return 1
}

// luaMajorMode implements gmacs.major_mode(name, config)
func (api *APIBindings) luaMajorMode(L *lua.LState) int {
	name := L.CheckString(1)
	configTable := L.CheckTable(2)
	
	// Convert Lua table to Go map
	config := make(map[string]interface{})
	configTable.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			config[string(keyStr)] = luaValueToGo(value)
		}
	})
	
	err := api.editor.RegisterMajorMode(name, config)
	if err != nil {
		L.Push(lua.LString("Error: " + err.Error()))
		return 1
	}
	
	log.Info("Lua: Registered major mode %s", name)
	return 0
}

// luaMinorMode implements gmacs.minor_mode(name, config)
func (api *APIBindings) luaMinorMode(L *lua.LState) int {
	name := L.CheckString(1)
	configTable := L.CheckTable(2)
	
	// Convert Lua table to Go map
	config := make(map[string]interface{})
	configTable.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			config[string(keyStr)] = luaValueToGo(value)
		}
	})
	
	err := api.editor.RegisterMinorMode(name, config)
	if err != nil {
		L.Push(lua.LString("Error: " + err.Error()))
		return 1
	}
	
	log.Info("Lua: Registered minor mode %s", name)
	return 0
}

// luaAddHook implements gmacs.add_hook(event, function)
func (api *APIBindings) luaAddHook(L *lua.LState) int {
	event := L.CheckString(1)
	fn := L.CheckFunction(2)
	
	// Create a wrapper function that calls the Lua function
	wrapper := func(args ...interface{}) error {
		// Convert Go arguments to Lua values
		luaArgs := make([]lua.LValue, len(args))
		for i, arg := range args {
			luaArgs[i] = goValueToLua(L, arg)
		}
		
		err := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}, luaArgs...)
		
		if err != nil {
			return &ConfigError{Message: "Lua hook error: " + err.Error()}
		}
		return nil
	}
	
	err := api.editor.AddHook(event, wrapper)
	if err != nil {
		L.Push(lua.LString("Error: " + err.Error()))
		return 1
	}
	
	log.Info("Lua: Added hook for event %s", event)
	return 0
}

// luaCurrentBuffer implements gmacs.current_buffer()
func (api *APIBindings) luaCurrentBuffer(L *lua.LState) int {
	buffer := api.editor.CurrentBuffer()
	if buffer == nil {
		L.Push(lua.LNil)
		return 1
	}
	
	// Return a table representing the buffer
	bufferTable := L.NewTable()
	L.SetField(bufferTable, "name", lua.LString(buffer.Name()))
	// Add more buffer properties as needed
	
	L.Push(bufferTable)
	return 1
}

// luaMessage implements gmacs.message(text)
func (api *APIBindings) luaMessage(L *lua.LState) int {
	message := L.CheckString(1)
	api.editor.SetMinibufferMessage(message)
	log.Info("Lua: Set message: %s", message)
	return 0
}

// luaToggleMinorMode implements gmacs.toggle_minor_mode(mode_name)
func (api *APIBindings) luaToggleMinorMode(L *lua.LState) int {
	modeName := L.CheckString(1)
	
	buffer := api.editor.CurrentBuffer()
	if buffer == nil {
		L.Push(lua.LBool(false))
		return 1
	}
	
	modeManager := api.editor.ModeManager()
	err := modeManager.ToggleMinorMode(buffer, modeName)
	if err != nil {
		log.Error("Lua: Failed to toggle minor mode %s: %v", modeName, err)
		L.Push(lua.LBool(false))
		return 1
	}
	
	// Check if mode is now enabled
	if minorMode, exists := modeManager.GetMinorModeByName(modeName); exists {
		enabled := minorMode.IsEnabled(buffer)
		L.Push(lua.LBool(enabled))
		log.Info("Lua: Toggled minor mode %s, now enabled: %v", modeName, enabled)
		return 1
	}
	
	L.Push(lua.LBool(false))
	return 1
}

// registerBuiltinCommands registers all built-in commands that were previously in editor.go
func (api *APIBindings) registerBuiltinCommands() {
	// Register cursor movement commands
	api.editor.RegisterCommand("forward-char", func() error { return domain.ForwardChar(api.editor) })
	api.editor.RegisterCommand("backward-char", func() error { return domain.BackwardChar(api.editor) })
	api.editor.RegisterCommand("next-line", func() error { return domain.NextLine(api.editor) })
	api.editor.RegisterCommand("previous-line", func() error { return domain.PreviousLine(api.editor) })
	api.editor.RegisterCommand("beginning-of-line", func() error { return domain.BeginningOfLine(api.editor) })
	api.editor.RegisterCommand("end-of-line", func() error { return domain.EndOfLine(api.editor) })
	
	// Register scrolling commands
	api.editor.RegisterCommand("scroll-up", func() error { return domain.ScrollUp(api.editor) })
	api.editor.RegisterCommand("scroll-down", func() error { return domain.ScrollDown(api.editor) })
	api.editor.RegisterCommand("page-up", func() error { return domain.PageUp(api.editor) })
	api.editor.RegisterCommand("page-down", func() error { return domain.PageDown(api.editor) })
	api.editor.RegisterCommand("toggle-truncate-lines", func() error { return domain.ToggleLineWrap(api.editor) })
	api.editor.RegisterCommand("debug-info", func() error { return domain.ShowDebugInfo(api.editor) })
	
	// Register buffer commands
	api.editor.RegisterCommand("switch-to-buffer", func() error { return domain.SwitchToBufferInteractive(api.editor) })
	api.editor.RegisterCommand("list-buffers", func() error { return domain.ListBuffersInteractive(api.editor) })
	api.editor.RegisterCommand("kill-buffer", func() error { return domain.KillBufferInteractive(api.editor) })
	
	// Register window commands
	api.editor.RegisterCommand("split-window-right", func() error { return domain.SplitWindowRight(api.editor) })
	api.editor.RegisterCommand("split-window-below", func() error { return domain.SplitWindowBelow(api.editor) })
	api.editor.RegisterCommand("other-window", func() error { return domain.OtherWindow(api.editor) })
	api.editor.RegisterCommand("delete-window", func() error { return domain.DeleteWindow(api.editor) })
	api.editor.RegisterCommand("delete-other-windows", func() error { return domain.DeleteOtherWindows(api.editor) })
	
	// Register minor mode commands
	api.registerMinorModeCommands()
	
	log.Info("Registered built-in commands via Lua API")
}

// registerMinorModeCommands registers minor mode commands
func (api *APIBindings) registerMinorModeCommands() {
	// auto-a-mode is now defined in default.lua
	// Other minor mode commands can be added here if needed
}

// Helper functions for type conversion

func luaValueToGo(value lua.LValue) interface{} {
	switch value.Type() {
	case lua.LTString:
		return lua.LVAsString(value)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(value))
	case lua.LTBool:
		return lua.LVAsBool(value)
	case lua.LTTable:
		table := value.(*lua.LTable)
		result := make(map[string]interface{})
		table.ForEach(func(key, val lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				result[string(keyStr)] = luaValueToGo(val)
			}
		})
		return result
	default:
		return nil
	}
}

func goValueToLua(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case string:
		return lua.LString(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case bool:
		return lua.LBool(v)
	default:
		return lua.LNil
	}
}