package pkg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPluginSupport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gmacs_plugin_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewManager(tempDir)

	// Test platform support detection
	expectedSupport := runtime.GOOS == "linux" || runtime.GOOS == "freebsd" || runtime.GOOS == "darwin"
	actualSupport := manager.isPluginSupported()

	if actualSupport != expectedSupport {
		t.Errorf("Expected plugin support %v for platform %s, got %v", expectedSupport, runtime.GOOS, actualSupport)
	}
}

func TestPluginLoadingFlow(t *testing.T) {
	// Skip if plugin loading not supported
	if runtime.GOOS != "linux" && runtime.GOOS != "freebsd" && runtime.GOOS != "darwin" {
		t.Skip("Plugin loading not supported on this platform")
	}

	tempDir, err := os.MkdirTemp("", "gmacs_plugin_flow_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewManager(tempDir)

	// Create a mock package structure for testing
	testURL := "github.com/test/ruby-mode"
	packageDir := filepath.Join(tempDir, "cache", "github.com", "test", "ruby-mode@v1.0.0")
	err = os.MkdirAll(packageDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create package directory: %v", err)
	}

	// Copy our sample plugin to the test location
	pluginSource := `// +build plugin

package main

import (
	"github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/pkg/gmacs"
)

type TestPlugin struct {
	enabled bool
}

func (r *TestPlugin) GetInfo() gmacs.PackageInfo {
	return gmacs.PackageInfo{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "Test plugin for unit testing",
	}
}

func (r *TestPlugin) Initialize() error { return nil }
func (r *TestPlugin) Cleanup() error { return nil }
func (r *TestPlugin) IsEnabled() bool { return r.enabled }
func (r *TestPlugin) Enable() error { r.enabled = true; return nil }
func (r *TestPlugin) Disable() error { r.enabled = false; return nil }

func (r *TestPlugin) ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error {
	luaTable.RawSetString("test_function", vm.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString("test result"))
		return 1
	}))
	return nil
}

func (r *TestPlugin) GetNamespace() string { return "test" }

var NewPackage = func() interface{} {
	return &TestPlugin{}
}
`

	pluginFile := filepath.Join(packageDir, "ruby_mode_plugin.go")
	err = os.WriteFile(pluginFile, []byte(pluginSource), 0644)
	if err != nil {
		t.Fatalf("Failed to write plugin source: %v", err)
	}

	// Test buildPlugin function
	pluginPath, err := manager.buildPlugin(packageDir, testURL)
	if err != nil {
		t.Errorf("Failed to build plugin: %v", err)
		return
	}

	// Check that plugin file was created
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Errorf("Plugin file was not created at %s", pluginPath)
		return
	}

	// Test loadPlugin function
	pkg, err := manager.loadPlugin(pluginPath)
	if err != nil {
		t.Errorf("Failed to load plugin: %v", err)
		return
	}

	// Verify plugin functionality
	info := pkg.GetInfo()
	if info.Name != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", info.Name)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected plugin version '1.0.0', got '%s'", info.Version)
	}

	// Test plugin enable/disable
	if pkg.IsEnabled() {
		t.Error("Plugin should not be enabled initially")
	}

	err = pkg.Enable()
	if err != nil {
		t.Errorf("Failed to enable plugin: %v", err)
	}

	if !pkg.IsEnabled() {
		t.Error("Plugin should be enabled after Enable() call")
	}
}

func TestUnsupportedPlatform(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gmacs_unsupported_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := NewManager(tempDir)

	// Note: We can't actually change runtime.GOOS at runtime,
	// so we'll just test the mock package functionality

	// Test that mock package is returned for unsupported platforms
	mockPkg, err := manager.loadMockPackage("github.com/test/package")
	if err != nil {
		t.Errorf("Failed to create mock package: %v", err)
	}

	info := mockPkg.GetInfo()
	if info.Name != "mock-package" {
		t.Errorf("Expected mock package name 'mock-package', got '%s'", info.Name)
	}

	if info.Description != "Mock package (plugin loading not supported on this platform)" {
		t.Errorf("Unexpected mock package description: %s", info.Description)
	}
}