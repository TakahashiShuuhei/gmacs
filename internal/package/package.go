package pkg

import (
	"github.com/yuin/gopher-lua"
)

// Package represents a gmacs extension package
type Package interface {
	// GetInfo returns basic package information
	GetInfo() PackageInfo
	
	// Initialize initializes the package (called after loading)
	Initialize() error
	
	// Cleanup cleans up package resources
	Cleanup() error
	
	// IsEnabled returns whether the package is currently enabled
	IsEnabled() bool
	
	// Enable enables the package
	Enable() error
	
	// Disable disables the package
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
	
	// ExtendLuaAPI adds package-specific functions to Lua environment
	ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error
	
	// GetNamespace returns the Lua namespace for this package (e.g., "ruby", "git")
	GetNamespace() string
}

// ConfigurablePackage represents a package that can be configured
type ConfigurablePackage interface {
	Package
	
	// SetConfig sets package configuration
	SetConfig(config map[string]any) error
	
	// GetConfig returns current package configuration
	GetConfig() map[string]any
	
	// GetDefaultConfig returns default configuration
	GetDefaultConfig() map[string]any
}

// PackageDeclaration represents a package declaration from Lua config
type PackageDeclaration struct {
	URL     string         `json:"url"`
	Version string         `json:"version"`
	Config  map[string]any `json:"config,omitempty"`
}

// PackageStatus represents the status of a package
type PackageStatus int

const (
	PackageStatusNotLoaded PackageStatus = iota
	PackageStatusLoading
	PackageStatusLoaded
	PackageStatusEnabled
	PackageStatusDisabled
	PackageStatusError
)

// String returns string representation of PackageStatus
func (ps PackageStatus) String() string {
	switch ps {
	case PackageStatusNotLoaded:
		return "not_loaded"
	case PackageStatusLoading:
		return "loading"
	case PackageStatusLoaded:
		return "loaded"
	case PackageStatusEnabled:
		return "enabled"
	case PackageStatusDisabled:
		return "disabled"
	case PackageStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// LoadedPackage represents a loaded package with its status
type LoadedPackage struct {
	Package    Package
	Status     PackageStatus
	Error      error
	LoadedAt   int64 // Unix timestamp
}