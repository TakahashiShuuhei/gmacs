package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
	pkg "github.com/TakahashiShuuhei/gmacs/internal/package"
)

//go:embed default.lua
var defaultConfig embed.FS

// EditorInterface defines interface for editor functionality needed by Lua config
type EditorInterface interface {
	// GetMinibuffer returns the minibuffer for message display
	GetMinibuffer() MinibufferInterface
	
	// BindKey binds a key sequence to a command
	BindKey(keySeq string, command string) error
	
	// GetCommandRegistry returns the command registry
	GetCommandRegistry() *command.Registry
}

// MinibufferInterface defines interface for minibuffer message display
type MinibufferInterface interface {
	ShowMessage(message string)
}

// LuaConfig manages Lua-based configuration
type LuaConfig struct {
	vm                      *lua.LState
	editor                  EditorInterface
	configPath              string // For testing purposes
	apiExtensions           []pkg.LuaAPIExtension
	packageLoadedCallbacks  []*lua.LFunction
	packageManager          *pkg.Manager
	parser                  *LuaParser
}

// NewLuaConfig creates a new Lua configuration manager
func NewLuaConfig(editor EditorInterface) *LuaConfig {
	// Create package manager with default download directory
	homeDir, _ := os.UserHomeDir()
	packagesDir := filepath.Join(homeDir, ".config", "gmacs", "packages")
	
	packageManager := pkg.NewManager(packagesDir)
	
	lc := &LuaConfig{
		editor:         editor,
		packageManager: packageManager,
		parser:         NewLuaParser(),
	}
	
	// Set up package manager to use this config for Lua API extensions
	packageManager.SetLuaConfig(lc)
	
	return lc
}

// RegisterAPIExtension registers a Lua API extension from a package
func (lc *LuaConfig) RegisterAPIExtension(ext pkg.LuaAPIExtension) error {
	lc.apiExtensions = append(lc.apiExtensions, ext)
	return nil
}

// SetConfigPath sets a custom config path (for testing)
func (lc *LuaConfig) SetConfigPath(path string) {
	lc.configPath = path
}

// LoadConfig loads and executes the Lua configuration file using D案 architecture
// Steps: 1. Parse package declarations, 2. Download packages, 3. Execute config with full APIs
func (lc *LuaConfig) LoadConfig() error {
	// Get config path
	var configPath string
	var err error
	if lc.configPath != "" {
		configPath = lc.configPath
	} else {
		configPath, err = lc.getConfigPath()
		if err != nil {
			return err
		}
	}

	// Step 1: Parse package declarations from config file (without executing)
	packageDeclarations, err := lc.parser.ParsePackageDeclarations(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse package declarations: %v", err)
	}

	// Step 2: Declare and load packages
	for _, decl := range packageDeclarations {
		lc.packageManager.DeclarePackage(decl.URL, decl.Version, decl.Config)
	}

	if len(packageDeclarations) > 0 {
		err = lc.packageManager.LoadDeclaredPackages()
		if err != nil {
			return fmt.Errorf("failed to load packages: %v", err)
		}
	}

	// Step 3: Initialize Lua VM with all APIs (basic + package extensions)
	if lc.vm != nil {
		lc.vm.Close()
	}
	lc.vm = lua.NewState()

	// Expose gmacs API to Lua (now includes package extensions)
	lc.exposeGmacsAPI()

	// Load default key bindings first
	err = lc.loadDefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to load default config: %v", err)
	}

	// Step 4: Execute the user config file (with all APIs available)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No user config file, that's OK
		return nil
	}

	return lc.vm.DoFile(configPath)
}

// Close closes the Lua VM
func (lc *LuaConfig) Close() {
	if lc.vm != nil {
		lc.vm.Close()
		lc.vm = nil
	}
}

// getConfigPath returns the path to the configuration file
func (lc *LuaConfig) getConfigPath() (string, error) {
	// Try XDG_CONFIG_HOME first
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		// Fall back to ~/.config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configHome, "gmacs", "init.lua"), nil
}

// loadDefaultConfig loads the default key bindings from embedded Lua file
func (lc *LuaConfig) loadDefaultConfig() error {
	// Read embedded default.lua file
	content, err := defaultConfig.ReadFile("default.lua")
	if err != nil {
		return fmt.Errorf("failed to read default config: %v", err)
	}

	// Execute the default configuration
	err = lc.vm.DoString(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute default config: %v", err)
	}
	return nil
}

// exposeGmacsAPI exposes gmacs functionality to Lua
func (lc *LuaConfig) exposeGmacsAPI() {
	// Create global 'gmacs' table
	gmacsTable := lc.vm.NewTable()

	// Configuration functions
	gmacsTable.RawSetString("set_variable", lc.vm.NewFunction(lc.luaSetVariable))
	gmacsTable.RawSetString("get_variable", lc.vm.NewFunction(lc.luaGetVariable))

	// Key binding functions
	gmacsTable.RawSetString("global_set_key", lc.vm.NewFunction(lc.luaGlobalSetKey))
	gmacsTable.RawSetString("local_set_key", lc.vm.NewFunction(lc.luaLocalSetKey))

	// Command functions
	gmacsTable.RawSetString("register_command", lc.vm.NewFunction(lc.luaRegisterCommand))
	gmacsTable.RawSetString("execute_command", lc.vm.NewFunction(lc.luaExecuteCommand))
	gmacsTable.RawSetString("list_commands", lc.vm.NewFunction(lc.luaListCommands))

	// Message functions
	gmacsTable.RawSetString("message", lc.vm.NewFunction(lc.luaMessage))
	gmacsTable.RawSetString("error", lc.vm.NewFunction(lc.luaError))

	// Buffer/Editor functions
	gmacsTable.RawSetString("current_buffer", lc.vm.NewFunction(lc.luaCurrentBuffer))
	gmacsTable.RawSetString("current_word", lc.vm.NewFunction(lc.luaCurrentWord))
	gmacsTable.RawSetString("current_char", lc.vm.NewFunction(lc.luaCurrentChar))

	// Hook functions
	gmacsTable.RawSetString("add_hook", lc.vm.NewFunction(lc.luaAddHook))

	// Package management functions
	gmacsTable.RawSetString("use_package", lc.vm.NewFunction(lc.luaUsePackage))
	gmacsTable.RawSetString("after_packages_loaded", lc.vm.NewFunction(lc.luaAfterPackagesLoaded))
	
	// Extend with registered API extensions
	for _, ext := range lc.apiExtensions {
		namespace := ext.GetNamespace()
		nsTable := lc.vm.NewTable()
		err := ext.ExtendLuaAPI(nsTable, lc.vm)
		if err != nil {
			fmt.Printf("Warning: Failed to extend Lua API for %s: %v\n", namespace, err)
			continue
		}
		gmacsTable.RawSetString(namespace, nsTable)
	}

	// Set global 'gmacs' table
	lc.vm.SetGlobal("gmacs", gmacsTable)
}

// Lua function implementations

func (lc *LuaConfig) luaSetVariable(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckAny(2)
	
	// TODO: Implement variable storage
	fmt.Printf("Setting variable %s = %v\n", key, value)
	return 0
}

func (lc *LuaConfig) luaGetVariable(L *lua.LState) int {
	key := L.CheckString(1)
	
	// TODO: Implement variable retrieval
	fmt.Printf("Getting variable %s\n", key)
	L.Push(lua.LNil)
	return 1
}

func (lc *LuaConfig) luaGlobalSetKey(L *lua.LState) int {
	keySeq := L.CheckString(1)
	command := L.CheckString(2)
	
	if lc.editor != nil {
		err := lc.editor.BindKey(keySeq, command)
		if err != nil {
			L.RaiseError("Failed to bind key '%s' to command '%s': %v", keySeq, command, err)
		}
	}
	return 0
}

func (lc *LuaConfig) luaLocalSetKey(L *lua.LState) int {
	keySeq := L.CheckString(1)
	command := L.CheckString(2)
	
	// TODO: Implement local key binding
	fmt.Printf("Local binding key %s to command %s\n", keySeq, command)
	return 0
}

func (lc *LuaConfig) luaRegisterCommand(L *lua.LState) int {
	name := L.CheckString(1)
	luaFunc := L.CheckFunction(2)
	description := ""
	
	// Optional description parameter
	if L.GetTop() >= 3 {
		description = L.CheckString(3)
	}
	
	// Register Lua function as a command
	registry := lc.editor.GetCommandRegistry()
	err := registry.Register(name, description, "", func(args ...any) error {
		// Create a new Lua state for execution to avoid conflicts
		execVM := lua.NewState()
		defer execVM.Close()
		
		// Copy the function to the execution VM
		execVM.SetGlobal("user_func", luaFunc)
		
		// Execute the function
		err := execVM.CallByParam(lua.P{
			Fn:      luaFunc,
			NRet:    0,
			Protect: true,
		})
		
		return err
	})
	
	if err != nil {
		L.RaiseError("Failed to register command '%s': %v", name, err)
	}
	
	return 0
}

func (lc *LuaConfig) luaExecuteCommand(L *lua.LState) int {
	name := L.CheckString(1)
	
	registry := lc.editor.GetCommandRegistry()
	err := registry.Execute(name)
	if err != nil {
		L.RaiseError("Failed to execute command '%s': %v", name, err)
	}
	
	return 0
}

func (lc *LuaConfig) luaListCommands(L *lua.LState) int {
	registry := lc.editor.GetCommandRegistry()
	commands := registry.List()
	
	// Create Lua table with command names
	table := L.NewTable()
	for i, cmd := range commands {
		table.RawSetInt(i+1, lua.LString(cmd))
	}
	
	L.Push(table)
	return 1
}

func (lc *LuaConfig) luaMessage(L *lua.LState) int {
	message := L.CheckString(1)
	
	if lc.editor != nil {
		minibuffer := lc.editor.GetMinibuffer()
		minibuffer.ShowMessage(message)
	} else {
		// Fallback to stdout if no editor is set
		fmt.Printf("Message: %s\n", message)
	}
	return 0
}

func (lc *LuaConfig) luaError(L *lua.LState) int {
	message := L.CheckString(1)
	
	if lc.editor != nil {
		minibuffer := lc.editor.GetMinibuffer()
		minibuffer.ShowMessage("Error: " + message)
	} else {
		// Fallback to stdout if no editor is set
		fmt.Printf("Error: %s\n", message)
	}
	return 0
}

func (lc *LuaConfig) luaCurrentBuffer(L *lua.LState) int {
	// TODO: Implement current buffer access
	// Return a table with buffer information
	bufferTable := L.NewTable()
	bufferTable.RawSetString("filename", lua.LString("example.txt"))
	bufferTable.RawSetString("modified", lua.LBool(false))
	
	L.Push(bufferTable)
	return 1
}

func (lc *LuaConfig) luaCurrentWord(L *lua.LState) int {
	// TODO: Implement current word detection
	L.Push(lua.LString("example_word"))
	return 1
}

func (lc *LuaConfig) luaCurrentChar(L *lua.LState) int {
	// TODO: Implement current character detection
	L.Push(lua.LString("a"))
	return 1
}

func (lc *LuaConfig) luaAddHook(L *lua.LState) int {
	hookName := L.CheckString(1)
	luaFunc := L.CheckFunction(2)
	
	// TODO: Implement hook system
	fmt.Printf("Adding hook %s\n", hookName)
	_ = luaFunc // Use the function parameter
	return 0
}

func (lc *LuaConfig) luaUsePackage(L *lua.LState) int {
	packageName := L.CheckString(1)
	version := ""
	
	// Optional version parameter
	if L.GetTop() >= 2 {
		version = L.CheckString(2)
	}
	
	// TODO: Implement package management integration
	// This will be connected to PackageManager
	fmt.Printf("Using package %s@%s\n", packageName, version)
	return 0
}

func (lc *LuaConfig) luaAfterPackagesLoaded(L *lua.LState) int {
	callback := L.CheckFunction(1)
	
	// Store callback to be executed after packages are loaded
	lc.packageLoadedCallbacks = append(lc.packageLoadedCallbacks, callback)
	return 0
}

// ExecutePackageLoadedCallbacks executes all stored callbacks (public method for testing)
func (lc *LuaConfig) ExecutePackageLoadedCallbacks() error {
	return lc.executePackageLoadedCallbacks()
}

// executePackageLoadedCallbacks executes all stored callbacks
func (lc *LuaConfig) executePackageLoadedCallbacks() error {
	for _, callback := range lc.packageLoadedCallbacks {
		err := lc.vm.CallByParam(lua.P{
			Fn:      callback,
			NRet:    0,
			Protect: true,
		})
		if err != nil {
			return fmt.Errorf("failed to execute package loaded callback: %v", err)
		}
	}
	return nil
}

// ExecuteCode executes Lua code (public method for testing)
func (lc *LuaConfig) ExecuteCode(code string) error {
	if lc.vm == nil {
		return fmt.Errorf("Lua VM not initialized")
	}
	return lc.vm.DoString(code)
}