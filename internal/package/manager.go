package pkg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"
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

// DownloaderInterface defines interface for package downloader
type DownloaderInterface interface {
	DownloadPackage(url, version string) error
	GetPackagePath(url string) (string, error)
	GetDownloadedPackages() ([]PackageInfo, error)
	InitializeWorkspace() error
	validatePackageURL(url string) error
}

// Manager manages gmacs packages
type Manager struct {
	mu               sync.RWMutex
	declaredPackages []PackageDeclaration
	loadedPackages   map[string]*LoadedPackage
	luaConfig        LuaConfigInterface
	downloadDir      string
	downloader       DownloaderInterface
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

// SetDownloader sets the package downloader (for testing)
func (pm *Manager) SetDownloader(downloader DownloaderInterface) {
	pm.downloader = downloader
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
	// Check if plugin loading is supported on this platform
	if !pm.isPluginSupported() {
		return pm.loadMockPackage(url)
	}

	// Get package path
	packagePath, err := pm.downloader.GetPackagePath(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get package path: %v", err)
	}

	// Build plugin
	pluginPath, err := pm.buildPlugin(packagePath, url)
	if err != nil {
		return nil, fmt.Errorf("failed to build plugin: %v", err)
	}

	// Load plugin
	pkg, err := pm.loadPlugin(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %v", err)
	}

	return pkg, nil
}

// isPluginSupported checks if plugin loading is supported on current platform
func (pm *Manager) isPluginSupported() bool {
	switch runtime.GOOS {
	case "linux", "freebsd", "darwin": // darwin = macOS
		return true
	default:
		return false
	}
}

// buildPlugin builds a package as a plugin (.so file)
func (pm *Manager) buildPlugin(packagePath, url string) (string, error) {
	// Create plugins directory
	pluginsDir := filepath.Join(pm.downloadDir, "plugins")
	err := os.MkdirAll(pluginsDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create plugins directory: %v", err)
	}

	// Generate plugin filename
	pluginName := filepath.Base(url) + ".so"
	pluginPath := filepath.Join(pluginsDir, pluginName)

	// Find the plugin source file (look for common plugin file patterns)
	pluginSourcePath, err := pm.findPluginSourceFile(packagePath)
	if err != nil {
		return "", fmt.Errorf("plugin source not found: %v", err)
	}

	// Check if plugin already exists and is up-to-date
	if pm.isPluginUpToDate(pluginPath, pluginSourcePath) {
		return pluginPath, nil
	}

	// Create temporary build directory
	tempBuildDir, err := os.MkdirTemp("", "gmacs-plugin-build-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp build directory: %v", err)
	}
	defer os.RemoveAll(tempBuildDir)

	// Copy source files to temp directory
	tempSourcePath := filepath.Join(tempBuildDir, filepath.Base(pluginSourcePath))
	err = pm.copyFile(pluginSourcePath, tempSourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to copy plugin source: %v", err)
	}

	// Create compatible go.mod in temp directory with main project versions
	tempGoModPath := filepath.Join(tempBuildDir, "go.mod")
	err = pm.createCompatibleGoMod(tempGoModPath)
	if err != nil {
		return "", fmt.Errorf("failed to create compatible go.mod: %v", err)
	}

	// Initialize dependencies in temp directory
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tempBuildDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go mod tidy failed: %v\nOutput: %s", err, string(output))
	}

	// Build plugin using go build -buildmode=plugin
	cmd = exec.Command("go", "build", "-buildmode=plugin", "-o", pluginPath, tempSourcePath)
	cmd.Dir = tempBuildDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("plugin build failed: %v\nOutput: %s", err, string(output))
	}

	// Plugin built successfully
	return pluginPath, nil
}

// ensurePluginGoMod creates or updates go.mod in plugin directory
func (pm *Manager) ensurePluginGoMod(packagePath string) error {
	goModPath := filepath.Join(packagePath, "go.mod")
	
	// Check if main project go.mod exists
	mainGoModPath := filepath.Join(pm.getProjectRoot(), "go.mod")
	if _, err := os.Stat(mainGoModPath); err != nil {
		return fmt.Errorf("main go.mod not found: %v", err)
	}

	// Create go.mod for plugin (simplified version)
	pluginGoModContent := `module gmacs-plugin

go 1.21

require (
	github.com/yuin/gopher-lua v0.0.0-20220504180219-658193537a64
	github.com/TakahashiShuuhei/gmacs v0.0.0
)

replace github.com/TakahashiShuuhei/gmacs => ` + pm.getProjectRoot() + `
`

	err := os.WriteFile(goModPath, []byte(pluginGoModContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create plugin go.mod: %v", err)
	}

	// Run go mod tidy to resolve dependencies
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = packagePath
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod tidy failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// getProjectRoot returns the root directory of the gmacs project
func (pm *Manager) getProjectRoot() string {
	// Find the directory containing go.mod
	currentDir, _ := os.Getwd()
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir
		}
		
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}
	
	// Fallback: assume we're in the project somewhere
	return filepath.Join(currentDir, "../../../")
}

// loadPlugin loads a .so plugin file
func (pm *Manager) loadPlugin(pluginPath string) (Package, error) {
	// Check if plugin file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file does not exist: %s", pluginPath)
	}
	
	// Open the plugin
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %v", pluginPath, err)
	}

	// Look for NewPackage symbol
	newPackageSymbol, err := plug.Lookup("NewPackage")
	if err != nil {
		return nil, fmt.Errorf("NewPackage symbol not found in plugin: %v", err)
	}

	// Convert to function and call it
	newPackageFunc, ok := newPackageSymbol.(func() interface{})
	if !ok {
		return nil, fmt.Errorf("NewPackage symbol is not a function returning interface{}")
	}
	
	// Call the function and type assert to our Package interface
	pluginInstance := newPackageFunc()
	pkg, ok := pluginInstance.(Package)
	if !ok {
		return nil, fmt.Errorf("plugin instance does not implement Package interface, got: %T", pluginInstance)
	}

	// Plugin package loaded successfully
	return pkg, nil
}

// loadMockPackage returns a mock package for unsupported platforms
func (pm *Manager) loadMockPackage(url string) (Package, error) {
	return &mockPackage{
		info: PackageInfo{
			Name:        "mock-package",
			URL:         url,
			Version:     "1.0.0",
			Description: "Mock package (plugin loading not supported on this platform)",
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

// copyFile copies a file from src to dst
func (pm *Manager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// createCompatibleGoMod creates a go.mod with versions compatible with main project
func (pm *Manager) createCompatibleGoMod(goModPath string) error {
	// Read main project's go.mod to get compatible versions
	mainGoModPath := filepath.Join(pm.getProjectRoot(), "go.mod")
	mainGoModContent, err := os.ReadFile(mainGoModPath)
	if err != nil {
		return fmt.Errorf("failed to read main go.mod: %v", err)
	}

	// Extract gopher-lua version from main go.mod
	gopherLuaVersion := "v0.0.0-20220504180219-658193537a64" // default fallback
	lines := strings.Split(string(mainGoModContent), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "github.com/yuin/gopher-lua") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				gopherLuaVersion = parts[1]
			}
			break
		}
	}

	// Create compatible go.mod content with replace directive to use local gmacs
	projectRoot := pm.getProjectRoot()
	goModContent := fmt.Sprintf(`module gmacs-plugin-build

go 1.21

require (
	github.com/yuin/gopher-lua %s
	github.com/TakahashiShuuhei/gmacs v0.1.0
)

replace github.com/TakahashiShuuhei/gmacs => %s
`, gopherLuaVersion, projectRoot)

	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write go.mod: %v", err)
	}

	return nil
}

// isPluginUpToDate checks if the plugin file is newer than the source file and dependencies
func (pm *Manager) isPluginUpToDate(pluginPath, sourcePath string) bool {
	// Check if plugin file exists
	pluginStat, err := os.Stat(pluginPath)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false // If we can't stat the plugin file, rebuild it
	}

	// Check source file modification time
	sourceStat, err := os.Stat(sourcePath)
	if err != nil {
		return false // If we can't stat the source file, something's wrong
	}

	// Plugin must be newer than source file
	if pluginStat.ModTime().Before(sourceStat.ModTime()) {
		return false
	}

	// Check if main project's go.mod is newer (dependency changes)
	mainGoModPath := filepath.Join(pm.getProjectRoot(), "go.mod")
	if mainGoModStat, err := os.Stat(mainGoModPath); err == nil {
		if pluginStat.ModTime().Before(mainGoModStat.ModTime()) {
			return false
		}
	}

	// Check package's go.mod if it exists
	packageDir := filepath.Dir(sourcePath)
	packageGoModPath := filepath.Join(packageDir, "go.mod")
	if packageGoModStat, err := os.Stat(packageGoModPath); err == nil {
		if pluginStat.ModTime().Before(packageGoModStat.ModTime()) {
			return false
		}
	}

	return true
}

// findPluginSourceFile finds the plugin source file in the package directory
func (pm *Manager) findPluginSourceFile(packagePath string) (string, error) {
	// Common plugin file patterns to look for
	patterns := []string{
		"*_plugin.go",
		"plugin.go", 
		"main.go",
	}
	
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(packagePath, pattern))
		if err != nil {
			continue
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	
	return "", fmt.Errorf("no plugin source file found in %s", packagePath)
}