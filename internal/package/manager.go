package pkg

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"
	
	"github.com/yuin/gopher-lua"
)

// LuaConfigInterface defines interface for Lua configuration
type LuaConfigInterface interface {
	RegisterAPIExtension(ext LuaAPIExtension) error
}

// LuaAPIExtension represents a package that can extend Lua API
type LuaAPIExtension interface {
	ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error
	GetNamespace() string
}

// Manager manages gmacs packages
type Manager struct {
	mu               sync.RWMutex
	declaredPackages []PackageDeclaration
	loadedPackages   map[string]*LoadedPackage
	luaConfig        LuaConfigInterface
	downloadDir      string
	downloader       *Downloader
}

// NewManager creates a new package manager
func NewManager(downloadDir string) *Manager {
	return &Manager{
		loadedPackages: make(map[string]*LoadedPackage),
		downloadDir:    downloadDir,
		downloader:     NewDownloader(filepath.Join(downloadDir, "cache")),
	}
}

// SetLuaConfig sets the Lua configuration manager
func (pm *Manager) SetLuaConfig(luaConfig LuaConfigInterface) {
	pm.luaConfig = luaConfig
}

// DeclarePackage adds a package to the declaration list (doesn't load it yet)
func (pm *Manager) DeclarePackage(url, version string, config map[string]any) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.declaredPackages = append(pm.declaredPackages, PackageDeclaration{
		URL:     url,
		Version: version,
		Config:  config,
	})
}

// GetDeclaredPackages returns the list of declared packages
func (pm *Manager) GetDeclaredPackages() []PackageDeclaration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	result := make([]PackageDeclaration, len(pm.declaredPackages))
	copy(result, pm.declaredPackages)
	return result
}

// LoadDeclaredPackages downloads and loads all declared packages
func (pm *Manager) LoadDeclaredPackages() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	for _, decl := range pm.declaredPackages {
		err := pm.loadPackage(decl)
		if err != nil {
			// Store error for the package
			pm.loadedPackages[decl.URL] = &LoadedPackage{
				Status: PackageStatusError,
				Error:  err,
				LoadedAt: time.Now().Unix(),
			}
			return fmt.Errorf("failed to load package %s: %v", decl.URL, err)
		}
	}
	
	return nil
}

// loadPackage loads a single package (internal method)
func (pm *Manager) loadPackage(decl PackageDeclaration) error {
	// Mark as loading
	pm.loadedPackages[decl.URL] = &LoadedPackage{
		Status:   PackageStatusLoading,
		LoadedAt: time.Now().Unix(),
	}
	
	// Step 1: Download package
	err := pm.downloadPackage(decl.URL, decl.Version)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	
	// Step 2: Load package binary/plugin
	pkg, err := pm.loadPackageBinary(decl.URL)
	if err != nil {
		return fmt.Errorf("load binary failed: %v", err)
	}
	
	// Step 3: Initialize package
	err = pkg.Initialize()
	if err != nil {
		return fmt.Errorf("initialization failed: %v", err)
	}
	
	// Step 4: Configure package if needed
	if configurablePkg, ok := pkg.(ConfigurablePackage); ok && decl.Config != nil {
		err = configurablePkg.SetConfig(decl.Config)
		if err != nil {
			return fmt.Errorf("configuration failed: %v", err)
		}
	}
	
	// Step 5: Register Lua API extension if applicable
	if luaExtPkg, ok := pkg.(LuaAPIExtender); ok && pm.luaConfig != nil {
		err = pm.luaConfig.RegisterAPIExtension(luaExtPkg)
		if err != nil {
			return fmt.Errorf("Lua API registration failed: %v", err)
		}
	}
	
	// Step 6: Enable package
	err = pkg.Enable()
	if err != nil {
		return fmt.Errorf("enable failed: %v", err)
	}
	
	// Success: Update status
	pm.loadedPackages[decl.URL] = &LoadedPackage{
		Package:  pkg,
		Status:   PackageStatusEnabled,
		LoadedAt: time.Now().Unix(),
	}
	
	return nil
}

// AddLoadedPackage manually adds a loaded package (for testing)
func (pm *Manager) AddLoadedPackage(url string, pkg Package) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	pm.loadedPackages[url] = &LoadedPackage{
		Package:  pkg,
		Status:   PackageStatusEnabled,
		LoadedAt: time.Now().Unix(),
	}
}

// downloadPackage downloads a package using go get
func (pm *Manager) downloadPackage(url, version string) error {
	return pm.downloader.DownloadPackage(url, version)
}

// loadPackageBinary loads the package binary/plugin
func (pm *Manager) loadPackageBinary(url string) (Package, error) {
	// TODO: Implement actual plugin loading
	// This would involve Go plugin system or other dynamic loading mechanism
	// For now, return a mock package
	return &mockPackage{
		info: PackageInfo{
			Name:        "mock-package",
			URL:         url,
			Version:     "1.0.0",
			Description: "Mock package for testing",
		},
	}, nil
}

// GetLoadedPackage returns a loaded package by URL
func (pm *Manager) GetLoadedPackage(url string) (*LoadedPackage, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	pkg, exists := pm.loadedPackages[url]
	return pkg, exists
}

// GetAllLoadedPackages returns all loaded packages
func (pm *Manager) GetAllLoadedPackages() map[string]*LoadedPackage {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	result := make(map[string]*LoadedPackage)
	for k, v := range pm.loadedPackages {
		result[k] = v
	}
	return result
}

// GetDownloadedPackages returns list of downloaded packages
func (pm *Manager) GetDownloadedPackages() ([]PackageInfo, error) {
	return pm.downloader.GetDownloadedPackages()
}

// GetPackagePath returns local path for a downloaded package
func (pm *Manager) GetPackagePath(url string) (string, error) {
	return pm.downloader.GetPackagePath(url)
}

// EnablePackage enables a loaded package
func (pm *Manager) EnablePackage(url string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	loadedPkg, exists := pm.loadedPackages[url]
	if !exists {
		return fmt.Errorf("package %s not loaded", url)
	}
	
	if loadedPkg.Package == nil {
		return fmt.Errorf("package %s has no valid package object", url)
	}
	
	err := loadedPkg.Package.Enable()
	if err != nil {
		return err
	}
	
	loadedPkg.Status = PackageStatusEnabled
	return nil
}

// DisablePackage disables a loaded package
func (pm *Manager) DisablePackage(url string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	loadedPkg, exists := pm.loadedPackages[url]
	if !exists {
		return fmt.Errorf("package %s not loaded", url)
	}
	
	if loadedPkg.Package == nil {
		return fmt.Errorf("package %s has no valid package object", url)
	}
	
	err := loadedPkg.Package.Disable()
	if err != nil {
		return err
	}
	
	loadedPkg.Status = PackageStatusDisabled
	return nil
}

// Cleanup cleans up all loaded packages
func (pm *Manager) Cleanup() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	var errors []error
	
	for url, loadedPkg := range pm.loadedPackages {
		if loadedPkg.Package != nil {
			err := loadedPkg.Package.Cleanup()
			if err != nil {
				errors = append(errors, fmt.Errorf("cleanup failed for %s: %v", url, err))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}
	
	return nil
}

// mockPackage is a mock implementation for testing
type mockPackage struct {
	info    PackageInfo
	enabled bool
}

func (mp *mockPackage) GetInfo() PackageInfo {
	return mp.info
}

func (mp *mockPackage) Initialize() error {
	return nil
}

func (mp *mockPackage) Cleanup() error {
	return nil
}

func (mp *mockPackage) IsEnabled() bool {
	return mp.enabled
}

func (mp *mockPackage) Enable() error {
	mp.enabled = true
	return nil
}

func (mp *mockPackage) Disable() error {
	mp.enabled = false
	return nil
}