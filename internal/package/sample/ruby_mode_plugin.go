// +build plugin

package main

import (
	"github.com/yuin/gopher-lua"
	pkg "github.com/TakahashiShuuhei/gmacs/internal/package"
)

// RubyModePlugin implements the Package interface for dynamic loading
type RubyModePlugin struct {
	enabled bool
}

// GetInfo returns package information
func (r *RubyModePlugin) GetInfo() pkg.PackageInfo {
	return pkg.PackageInfo{
		Name:        "ruby-mode",
		Version:     "1.0.0",
		Description: "Ruby editing support for gmacs",
		Author:      "gmacs contributors",
		URL:         "github.com/TakahashiShuuhei/gmacs/internal/package/sample/ruby-mode",
		Keywords:    []string{"ruby", "programming", "syntax"},
	}
}

// Initialize initializes the package
func (r *RubyModePlugin) Initialize() error {
	return nil
}

// Cleanup cleans up package resources
func (r *RubyModePlugin) Cleanup() error {
	return nil
}

// IsEnabled returns whether the package is currently enabled
func (r *RubyModePlugin) IsEnabled() bool {
	return r.enabled
}

// Enable enables the package
func (r *RubyModePlugin) Enable() error {
	r.enabled = true
	return nil
}

// Disable disables the package
func (r *RubyModePlugin) Disable() error {
	r.enabled = false
	return nil
}

// ExtendLuaAPI adds Ruby-specific functions to Lua environment
func (r *RubyModePlugin) ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error {
	// Add ruby.run_script function
	luaTable.RawSetString("run_script", vm.NewFunction(r.luaRunScript))
	
	// Add ruby.check_syntax function
	luaTable.RawSetString("check_syntax", vm.NewFunction(r.luaCheckSyntax))
	
	// Add ruby.show_doc function
	luaTable.RawSetString("show_doc", vm.NewFunction(r.luaShowDoc))
	
	return nil
}

// GetNamespace returns the Lua namespace for this package
func (r *RubyModePlugin) GetNamespace() string {
	return "ruby"
}

// Lua function implementations

func (r *RubyModePlugin) luaRunScript(L *lua.LState) int {
	script := L.CheckString(1)
	// In a real implementation, this would execute Ruby script
	result := "Executed Ruby script: " + script
	L.Push(lua.LString(result))
	return 1
}

func (r *RubyModePlugin) luaCheckSyntax(L *lua.LState) int {
	code := L.CheckString(1)
	// In a real implementation, this would check Ruby syntax
	isValid := len(code) > 0 // Simple validation for demo
	L.Push(lua.LBool(isValid))
	return 1
}

func (r *RubyModePlugin) luaShowDoc(L *lua.LState) int {
	symbol := L.CheckString(1)
	// In a real implementation, this would show Ruby documentation
	doc := "Documentation for Ruby symbol: " + symbol
	L.Push(lua.LString(doc))
	return 1
}

// Plugin entry point - this is what gets loaded by plugin.Open()
var NewPackage = func() pkg.Package {
	return &RubyModePlugin{}
}

// Also export as LuaAPIExtender
var NewLuaAPIExtender = func() pkg.LuaAPIExtender {
	return &RubyModePlugin{}
}