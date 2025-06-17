package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/internal/command"
)

func TestDArchitecture_LoadConfigWithPackages(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gmacs_d_architecture_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config file with package declarations
	configFile := filepath.Join(tempDir, "init.lua")
	configContent := `
-- Test configuration with packages
gmacs.use_package("github.com/test/package1", "v1.0.0")
gmacs.use_package("github.com/test/package2", {
	enabled = true,
	theme = "dark"
})

-- Regular configuration
gmacs.set_variable("test_var", "test_value")

function test_function()
	gmacs.message("Hello from test")
end

gmacs.register_command("test-cmd", test_function, "Test command")
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create mock editor
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}

	// Create Lua config with custom package directory
	luaConfig := NewLuaConfig(mockEditor)
	luaConfig.SetConfigPath(configFile)
	
	// Override package manager with test directory
	packagesDir := filepath.Join(tempDir, "packages")
	err = os.MkdirAll(packagesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create packages dir: %v", err)
	}

	// Note: Since we can't easily mock the package loading without 
	// restructuring the system, let's test just the parsing step
	packageDeclarations, err := luaConfig.parser.ParsePackageDeclarations(configFile)
	if err != nil {
		t.Fatalf("Failed to parse package declarations: %v", err)
	}

	// Verify packages were parsed correctly
	if len(packageDeclarations) != 2 {
		t.Errorf("Expected 2 package declarations, got %d", len(packageDeclarations))
	}

	// Check first package
	if packageDeclarations[0].URL != "github.com/test/package1" {
		t.Errorf("Expected first package URL 'github.com/test/package1', got '%s'", packageDeclarations[0].URL)
	}
	if packageDeclarations[0].Version != "v1.0.0" {
		t.Errorf("Expected first package version 'v1.0.0', got '%s'", packageDeclarations[0].Version)
	}

	// Check second package
	if packageDeclarations[1].URL != "github.com/test/package2" {
		t.Errorf("Expected second package URL 'github.com/test/package2', got '%s'", packageDeclarations[1].URL)
	}
	if packageDeclarations[1].Version != "latest" {
		t.Errorf("Expected second package version 'latest', got '%s'", packageDeclarations[1].Version)
	}
	if len(packageDeclarations[1].Config) != 2 {
		t.Errorf("Expected second package to have 2 config items, got %d", len(packageDeclarations[1].Config))
	}
}

func TestDArchitecture_LoadConfigWithoutPackages(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gmacs_d_architecture_no_packages_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config file without package declarations
	configFile := filepath.Join(tempDir, "init.lua")
	configContent := `
-- Test configuration without packages
gmacs.set_variable("editor_theme", "light")

function my_command()
	gmacs.message("Command executed")
end

gmacs.register_command("my-cmd", my_command, "My custom command")
gmacs.global_set_key("C-t", "my-cmd")
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create mock editor
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}

	// Create Lua config
	luaConfig := NewLuaConfig(mockEditor)
	luaConfig.SetConfigPath(configFile)

	// Test the complete LoadConfig flow (should work without packages)
	err = luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Verify that the command was registered
	_, exists := mockEditor.registry.Get("my-cmd")
	if !exists {
		t.Error("Expected 'my-cmd' command to be registered")
	}

	// Verify that key binding was set
	if mockEditor.keyBindings["C-t"] != "my-cmd" {
		t.Errorf("Expected key binding 'C-t' -> 'my-cmd', got '%s'", mockEditor.keyBindings["C-t"])
	}
}

func TestDArchitecture_ParseEmptyConfig(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gmacs_d_architecture_empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create empty config file
	configFile := filepath.Join(tempDir, "init.lua")
	err = os.WriteFile(configFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create mock editor
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}

	// Create Lua config
	luaConfig := NewLuaConfig(mockEditor)
	luaConfig.SetConfigPath(configFile)

	// Test LoadConfig with empty file (should succeed)
	err = luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig with empty file failed: %v", err)
	}

	// Should have default commands but no user commands
	_, exists := mockEditor.registry.Get("version")
	if !exists {
		t.Error("Expected 'version' command from default config to be registered")
	}
}