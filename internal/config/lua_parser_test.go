package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLuaParser_ParsePackageDeclarations(t *testing.T) {
	parser := NewLuaParser()

	testCases := []struct {
		name     string
		content  string
		expected []PackageDeclaration
	}{
		{
			name: "Simple package with version",
			content: `
				gmacs.use_package("github.com/user/package", "v1.0.0")
			`,
			expected: []PackageDeclaration{
				{
					URL:     "github.com/user/package",
					Version: "v1.0.0",
					Config:  map[string]any{},
				},
			},
		},
		{
			name: "Package without version",
			content: `
				gmacs.use_package("github.com/user/package")
			`,
			expected: []PackageDeclaration{
				{
					URL:     "github.com/user/package",
					Version: "latest",
					Config:  map[string]any{},
				},
			},
		},
		{
			name: "Package with config",
			content: `
				gmacs.use_package("github.com/user/package", {
					enabled = true,
					theme = "dark",
					count = 42
				})
			`,
			expected: []PackageDeclaration{
				{
					URL:     "github.com/user/package",
					Version: "latest",
					Config: map[string]any{
						"enabled": true,
						"theme":   "dark",
						"count":   42,
					},
				},
			},
		},
		{
			name: "Multiple packages",
			content: `
				gmacs.use_package("github.com/user/package1", "v1.0.0")
				gmacs.use_package("github.com/user/package2")
				gmacs.use_package("github.com/user/package3", {enabled = false})
			`,
			expected: []PackageDeclaration{
				{
					URL:     "github.com/user/package1",
					Version: "v1.0.0",
					Config:  map[string]any{},
				},
				{
					URL:     "github.com/user/package2",
					Version: "latest",
					Config:  map[string]any{},
				},
				{
					URL:     "github.com/user/package3",
					Version: "latest",
					Config: map[string]any{
						"enabled": false,
					},
				},
			},
		},
		{
			name: "Mixed with other Lua code",
			content: `
				-- Configuration file
				gmacs.set_variable("theme", "dark")
				
				gmacs.use_package("github.com/user/ruby-mode", "v2.1.0")
				
				function my_function()
					gmacs.message("hello")
				end
				
				gmacs.use_package("github.com/user/git-mode", {
					auto_stage = true,
					show_status = false
				})
				
				gmacs.global_set_key("C-x g", "git-status")
			`,
			expected: []PackageDeclaration{
				{
					URL:     "github.com/user/ruby-mode",
					Version: "v2.1.0",
					Config:  map[string]any{},
				},
				{
					URL:     "github.com/user/git-mode",
					Version: "latest",
					Config: map[string]any{
						"auto_stage":  true,
						"show_status": false,
					},
				},
			},
		},
		{
			name:     "No packages",
			content:  `gmacs.set_variable("theme", "dark")`,
			expected: []PackageDeclaration{},
		},
		{
			name:     "Empty content",
			content:  "",
			expected: []PackageDeclaration{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.parsePackageDeclarationsFromContent(tc.content)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d declarations, got %d", len(tc.expected), len(result))
				return
			}

			for i, expected := range tc.expected {
				actual := result[i]

				if actual.URL != expected.URL {
					t.Errorf("Declaration %d: Expected URL %s, got %s", i, expected.URL, actual.URL)
				}

				if actual.Version != expected.Version {
					t.Errorf("Declaration %d: Expected version %s, got %s", i, expected.Version, actual.Version)
				}

				if len(actual.Config) != len(expected.Config) {
					t.Errorf("Declaration %d: Expected %d config items, got %d", i, len(expected.Config), len(actual.Config))
					continue
				}

				for key, expectedValue := range expected.Config {
					actualValue, exists := actual.Config[key]
					if !exists {
						t.Errorf("Declaration %d: Expected config key %s not found", i, key)
						continue
					}

					if actualValue != expectedValue {
						t.Errorf("Declaration %d: Expected config[%s] = %v, got %v", i, key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestLuaParser_ParseFromFile(t *testing.T) {
	// Create temporary file
	tempDir, err := os.MkdirTemp("", "gmacs_parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "init.lua")
	configContent := `
		-- Test config file
		gmacs.use_package("github.com/test/package", "v1.0.0")
		gmacs.set_variable("test", "value")
	`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	parser := NewLuaParser()
	declarations, err := parser.ParsePackageDeclarations(configFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(declarations) != 1 {
		t.Errorf("Expected 1 declaration, got %d", len(declarations))
		return
	}

	expected := PackageDeclaration{
		URL:     "github.com/test/package",
		Version: "v1.0.0",
		Config:  map[string]any{},
	}

	actual := declarations[0]
	if actual.URL != expected.URL || actual.Version != expected.Version {
		t.Errorf("Expected %+v, got %+v", expected, actual)
	}
}

func TestLuaParser_NonExistentFile(t *testing.T) {
	parser := NewLuaParser()
	declarations, err := parser.ParsePackageDeclarations("/nonexistent/file.lua")
	
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
		return
	}

	if len(declarations) != 0 {
		t.Errorf("Expected empty declarations for non-existent file, got %d", len(declarations))
	}
}

func TestLuaParser_ParseSimpleLuaTable(t *testing.T) {
	parser := NewLuaParser()

	testCases := []struct {
		name     string
		input    string
		expected map[string]any
		hasError bool
	}{
		{
			name:  "Simple table",
			input: `{enabled = true, name = "test"}`,
			expected: map[string]any{
				"enabled": true,
				"name":    "test",
			},
		},
		{
			name:  "Mixed types",
			input: `{count = 42, rate = 3.14, active = false}`,
			expected: map[string]any{
				"count":  42,
				"rate":   3.14,
				"active": false,
			},
		},
		{
			name:     "Empty table",
			input:    `{}`,
			expected: map[string]any{},
		},
		{
			name:     "Invalid format",
			input:    `not a table`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.parseSimpleLuaTable(tc.input)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d items, got %d", len(tc.expected), len(result))
				return
			}

			for key, expectedValue := range tc.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key %s not found", key)
					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Expected %s = %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}