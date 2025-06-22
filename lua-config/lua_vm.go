package luaconfig

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/log"
)

// LuaVM manages a Lua virtual machine for configuration
type LuaVM struct {
	state *lua.LState
}

// NewLuaVM creates a new Lua virtual machine
func NewLuaVM() *LuaVM {
	L := lua.NewState()
	
	// TODO: Add sandbox restrictions
	// - Limit memory usage
	// - Restrict file system access
	// - Set execution timeout
	
	return &LuaVM{
		state: L,
	}
}

// Close closes the Lua virtual machine
func (vm *LuaVM) Close() {
	if vm.state != nil {
		vm.state.Close()
		vm.state = nil
	}
}

// LoadConfig loads and executes a Lua configuration file
func (vm *LuaVM) LoadConfig(configPath string) error {
	if vm.state == nil {
		return &ConfigError{Message: "Lua VM is closed"}
	}
	
	log.Info("Loading Lua config: %s", configPath)
	
	err := vm.state.DoFile(configPath)
	if err != nil {
		return &ConfigError{Message: "Failed to load config: " + err.Error()}
	}
	
	log.Info("Successfully loaded Lua config")
	return nil
}

// ExecuteString executes a Lua code string
func (vm *LuaVM) ExecuteString(code string) error {
	if vm.state == nil {
		return &ConfigError{Message: "Lua VM is closed"}
	}
	
	err := vm.state.DoString(code)
	if err != nil {
		return &ConfigError{Message: "Failed to execute Lua code: " + err.Error()}
	}
	
	return nil
}

// GetGlobalFunction retrieves a global Lua function
func (vm *LuaVM) GetGlobalFunction(name string) lua.LValue {
	if vm.state == nil {
		return lua.LNil
	}
	
	return vm.state.GetGlobal(name)
}

// CallFunction calls a Lua function with arguments
func (vm *LuaVM) CallFunction(fn lua.LValue, args ...lua.LValue) error {
	if vm.state == nil {
		return &ConfigError{Message: "Lua VM is closed"}
	}
	
	err := vm.state.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, args...)
	
	if err != nil {
		return &ConfigError{Message: "Failed to call Lua function: " + err.Error()}
	}
	
	return nil
}

// GetState returns the underlying Lua state (for API bindings)
// Implements domain.VM interface
func (vm *LuaVM) GetState() interface{} {
	return vm.state
}

// ConfigError represents a configuration-related error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}