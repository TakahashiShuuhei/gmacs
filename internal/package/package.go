package pkg

import (
	"github.com/TakahashiShuuhei/gmacs/pkg/gmacs"
)

// Package is an alias for the public interface
type Package = gmacs.Package

// PackageInfo is an alias for the public type
type PackageInfo = gmacs.PackageInfo

// LuaAPIExtender is an alias for the public interface
type LuaAPIExtender = gmacs.LuaAPIExtender

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