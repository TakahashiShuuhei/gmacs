package gmacs

import "github.com/yuin/gopher-lua"

// Package represents a gmacs extension package (public interface)
type Package interface {
	GetInfo() PackageInfo
	Initialize() error
	Cleanup() error
	IsEnabled() bool
	Enable() error
	Disable() error
}

// PackageInfo contains metadata about a package
type PackageInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	URL         string   `json:"url"`
	Dependencies []string `json:"dependencies"`
	Keywords    []string `json:"keywords"`
}

// LuaAPIExtender represents a package that can extend Lua API
type LuaAPIExtender interface {
	Package
	ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error
	GetNamespace() string
}