package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLuaConfig_ValidationIntegration(t *testing.T) {
	// Create a temporary config file with validation errors
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "init.lua")

	invalidConfig := `-- Invalid configuration for testing

-- Unknown function (semantic error)
gmacs.unknown_function("arg1", "arg2")

-- Invalid package URL (package error)
gmacs.use_package("invalid-url", "v1.0.0")

-- Invalid key sequence (keybinding error)
gmacs.global_set_key("invalid-key", "find-file")

-- Unknown command (keybinding error)
gmacs.global_set_key("C-x C-u", "unknown-command")

-- Valid stuff that should pass
gmacs.global_set_key("C-x C-s", "save-buffer")`

	err := os.WriteFile(configPath, []byte(invalidConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test validation through LuaConfig
	luaConfig := NewLuaConfig(nil) // No editor for testing
	
	result, err := luaConfig.ValidateConfigFile(configPath)
	if err != nil {
		t.Fatalf("ValidateConfigFile failed: %v", err)
	}

	// Should have validation errors
	if result.Valid {
		t.Error("Expected validation to fail, but it passed")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected validation errors, but got none")
	}

	// Check error types
	syntaxErrors := result.GetErrorsByType("syntax")
	semanticErrors := result.GetErrorsByType("semantic")
	packageErrors := result.GetErrorsByType("package")
	keybindingErrors := result.GetErrorsByType("keybinding")

	t.Logf("Found %d total errors:", len(result.Errors))
	t.Logf("  Syntax: %d", len(syntaxErrors))
	t.Logf("  Semantic: %d", len(semanticErrors))
	t.Logf("  Package: %d", len(packageErrors))
	t.Logf("  Keybinding: %d", len(keybindingErrors))

	// Note: No syntax errors in this config, that's tested separately

	// Should have semantic errors
	if len(semanticErrors) == 0 {
		t.Error("Expected at least one semantic error")
	}

	// Should have package errors
	if len(packageErrors) == 0 {
		t.Error("Expected at least one package error")
	}

	// Should have keybinding errors
	if len(keybindingErrors) == 0 {
		t.Error("Expected at least one keybinding error")
	}
}

func TestLuaConfig_ValidationSyntaxErrors(t *testing.T) {
	// Test configuration with syntax errors
	configWithSyntaxErrors := `-- Configuration with syntax errors

-- Syntax error: missing quote  
gmacs.global_set_key("C-x C-f, "find-file")

-- Another syntax error: missing parenthesis
gmacs.global_set_key("C-x C-s", "save-buffer"`

	luaConfig := NewLuaConfig(nil)
	result := luaConfig.ValidateConfig(configWithSyntaxErrors)

	if result.Valid {
		t.Error("Expected validation to fail due to syntax errors")
	}

	syntaxErrors := result.GetErrorsByType("syntax")
	if len(syntaxErrors) == 0 {
		t.Error("Expected at least one syntax error")
	}

	t.Logf("Found %d syntax errors", len(syntaxErrors))
	for _, err := range syntaxErrors {
		t.Logf("  %s", err.Error())
	}
}

func TestLuaConfig_ValidationWithCustomCommands(t *testing.T) {
	// Test configuration with custom commands
	validConfigWithCustomCommands := `-- Valid configuration with custom commands

-- Register a custom command
function my_custom_command()
    gmacs.message("Hello from custom command!")
end

gmacs.register_command("my-custom", my_custom_command, "My custom command")

-- Bind the custom command (should be valid)
gmacs.global_set_key("C-c m", "my-custom")

-- Valid package declaration
gmacs.use_package("github.com/user/test-package", "v1.0.0")

-- Valid built-in command binding
gmacs.global_set_key("C-x C-f", "find-file")`

	luaConfig := NewLuaConfig(nil)
	result := luaConfig.ValidateConfig(validConfigWithCustomCommands)

	if !result.Valid {
		t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
	}

	// Should have no errors
	if result.HasErrors() {
		t.Errorf("Expected no errors, but got: %v", result.Errors)
	}
}

func TestLuaConfig_LoadConfigWithValidation(t *testing.T) {
	// Create a temporary config file with minor issues (warnings only)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "init.lua")

	configWithWarnings := `-- Configuration with warnings

-- Unknown command (warning)
gmacs.global_set_key("C-x C-t", "unknown-test-command")

-- Valid stuff
gmacs.global_set_key("C-x C-f", "find-file")
gmacs.global_set_key("C-x C-s", "save-buffer")`

	err := os.WriteFile(configPath, []byte(configWithWarnings), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading config with validation
	luaConfig := NewLuaConfig(nil)
	luaConfig.SetConfigPath(configPath)

	// Should load successfully despite warnings
	err = luaConfig.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig should succeed with warnings only, but got error: %v", err)
	}
}

func TestLuaConfig_LoadConfigWithErrors(t *testing.T) {
	// Create a temporary config file with syntax errors
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "init.lua")

	configWithErrors := `-- Configuration with syntax errors

-- Syntax error: missing quote
gmacs.global_set_key("C-x C-f, "find-file")

-- Valid stuff (should not be reached due to syntax error)
gmacs.global_set_key("C-x C-s", "save-buffer")`

	err := os.WriteFile(configPath, []byte(configWithErrors), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading config with validation errors
	luaConfig := NewLuaConfig(nil)
	luaConfig.SetConfigPath(configPath)

	// Should fail to load due to validation errors
	err = luaConfig.LoadConfig()
	if err == nil {
		t.Error("Expected LoadConfig to fail with syntax errors, but it succeeded")
	}

	t.Logf("LoadConfig correctly failed with error: %v", err)
}

func TestValidationError_Methods(t *testing.T) {
	err := &ValidationError{
		Type:     "syntax",
		Message:  "test error",
		Line:     10,
		Column:   5,
		Severity: "error",
	}

	// Test Error() method
	errorStr := err.Error()
	expected := "syntax:10:5: error: test error"
	if errorStr != expected {
		t.Errorf("Error() = %q, want %q", errorStr, expected)
	}

	// Test error without line/column
	errNoLine := &ValidationError{
		Type:     "semantic",
		Message:  "test error",
		Severity: "warning",
	}

	errorStr = errNoLine.Error()
	expected = "semantic: warning: test error"
	if errorStr != expected {
		t.Errorf("Error() = %q, want %q", errorStr, expected)
	}
}

func TestConfigValidator_EdgeCases(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name        string
		content     string
		expectValid bool
		description string
	}{
		{
			name:        "empty config",
			content:     "",
			expectValid: true,
			description: "Empty config should be valid",
		},
		{
			name:        "comments only",
			content:     "-- This is just a comment\n-- Another comment",
			expectValid: true,
			description: "Comments-only config should be valid",
		},
		{
			name: "mixed valid and invalid",
			content: `-- Valid and invalid mixed
gmacs.global_set_key("C-x C-f", "find-file")  -- valid
gmacs.unknown_func()  -- invalid
gmacs.global_set_key("C-x C-s", "save-buffer")  -- valid`,
			expectValid: false,
			description: "Mixed valid/invalid should be invalid overall",
		},
		{
			name: "lua constructs",
			content: `-- Lua language constructs
if true then
    gmacs.message("Hello")
end

for i=1,5 do
    gmacs.message("Count: " .. i)
end

local function helper()
    return "test"
end`,
			expectValid: true,
			description: "Standard Lua constructs should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.content)
			
			if result.Valid != tt.expectValid {
				t.Errorf("%s: got valid=%v, want %v", tt.description, result.Valid, tt.expectValid)
				if len(result.Errors) > 0 {
					t.Logf("Errors found:")
					for _, err := range result.Errors {
						t.Logf("  %s", err.Error())
					}
				}
			}
		})
	}
}