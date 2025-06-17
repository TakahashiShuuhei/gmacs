package config

import (
	"strings"
	"testing"
)

func TestConfigValidator_ValidateSyntax(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name:      "valid lua syntax",
			content:   `gmacs.global_set_key("C-x C-f", "find-file")`,
			wantError: false,
		},
		{
			name:      "invalid lua syntax - missing quote",
			content:   `gmacs.global_set_key("C-x C-f, "find-file")`,
			wantError: true,
		},
		{
			name:      "invalid lua syntax - missing parenthesis",
			content:   `gmacs.global_set_key("C-x C-f", "find-file"`,
			wantError: true,
		},
		{
			name:      "valid multi-line config",
			content: `-- Configuration
gmacs.global_set_key("C-x C-f", "find-file")
gmacs.global_set_key("C-x C-s", "save-buffer")`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.content)
			
			hasError := result.HasErrors()
			if hasError != tt.wantError {
				t.Errorf("ValidateConfig() hasError = %v, want %v", hasError, tt.wantError)
				for _, err := range result.Errors {
					t.Logf("Error: %s", err.Error())
				}
			}
		})
	}
}

func TestConfigValidator_ValidateSemantics(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name         string
		content      string
		wantErrorType string
		expectError   bool
	}{
		{
			name:          "valid gmacs function",
			content:       `gmacs.global_set_key("C-x C-f", "find-file")`,
			expectError:   false,
		},
		{
			name:          "unknown gmacs function",
			content:       `gmacs.unknown_function("arg1", "arg2")`,
			wantErrorType: "semantic",
			expectError:   true,
		},
		{
			name:          "valid package declaration",
			content:       `gmacs.use_package("github.com/user/repo", "v1.0.0")`,
			expectError:   false,
		},
		{
			name:          "invalid package URL",
			content:       `gmacs.use_package("invalid-url", "v1.0.0")`,
			wantErrorType: "package",
			expectError:   true,
		},
		{
			name:          "valid key binding",
			content:       `gmacs.global_set_key("C-x C-f", "find-file")`,
			expectError:   false,
		},
		{
			name:          "invalid key sequence",
			content:       `gmacs.global_set_key("invalid-key", "find-file")`,
			wantErrorType: "keybinding",
			expectError:   true,
		},
		{
			name:          "unknown command",
			content:       `gmacs.global_set_key("C-x C-f", "unknown-command")`,
			wantErrorType: "keybinding",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateConfig(tt.content)
			
			if tt.expectError {
				if !result.HasErrors() && len(result.Errors) == 0 {
					t.Errorf("ValidateConfig() expected error but got none")
					return
				}
				
				if tt.wantErrorType != "" {
					found := false
					for _, err := range result.Errors {
						if err.Type == tt.wantErrorType {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ValidateConfig() expected error type %s but got %v", 
							tt.wantErrorType, result.Errors)
					}
				}
			} else {
				if result.HasErrors() {
					t.Errorf("ValidateConfig() unexpected error: %v", result.Errors)
				}
			}
		})
	}
}

func TestConfigValidator_KeySequenceValidation(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		keySeq string
		valid  bool
	}{
		{"C-x", true},
		{"M-x", true},
		{"C-c C-f", true},
		{"C-x C-s", true},
		{"C-c m", true},
		{"a", true},
		{"RET", true},
		{"invalid-key", false},
		{"", false},
		{"Z-x", false}, // Z- is not a valid modifier
	}

	for _, tt := range tests {
		t.Run(tt.keySeq, func(t *testing.T) {
			result := validator.isValidKeySequence(tt.keySeq)
			if result != tt.valid {
				t.Errorf("isValidKeySequence(%q) = %v, want %v", tt.keySeq, result, tt.valid)
			}
		})
	}
}

func TestConfigValidator_PackageURLValidation(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		url   string
		valid bool
	}{
		{"github.com/user/repo", true},
		{"gitlab.com/user/repo", true},
		{"bitbucket.org/user/repo", true},
		{"github.com/user/repo/subpath", true},
		{"invalid-url", false},
		{"github.com/user", false}, // missing repo
		{"github.com/", false},     // missing user and repo
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := validator.isValidPackageURL(tt.url)
			if result != tt.valid {
				t.Errorf("isValidPackageURL(%q) = %v, want %v", tt.url, result, tt.valid)
			}
		})
	}
}

func TestConfigValidator_VersionValidation(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		version string
		valid   bool
	}{
		{"v1.0.0", true},
		{"v2.1.3-beta", true},
		{"abc123def456789012345678901234567890abcd", true}, // 40-char commit hash
		{"abc123d", true}, // 7-char short hash
		{"abc123def45", true}, // 11-char hash
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := validator.isValidVersion(tt.version)
			if result != tt.valid {
				t.Errorf("isValidVersion(%q) = %v, want %v", tt.version, result, tt.valid)
			}
		})
	}
}

func TestConfigValidator_Suggestions(t *testing.T) {
	validator := NewConfigValidator()

	// Test function suggestions
	knownFunctions := map[string]bool{
		"global_set_key": true,
		"local_set_key":  true,
		"register_command": true,
	}

	suggestions := validator.suggestSimilarFunctions("global_key", knownFunctions)
	if len(suggestions) == 0 {
		t.Error("Expected suggestions for 'global_key' but got none")
	}

	found := false
	for _, suggestion := range suggestions {
		if suggestion == "global_set_key" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'global_set_key' in suggestions, got %v", suggestions)
	}
}

func TestValidationResult_Methods(t *testing.T) {
	result := &ValidationResult{
		Errors: []*ValidationError{
			{Type: "syntax", Severity: "error", Message: "syntax error"},
			{Type: "semantic", Severity: "warning", Message: "semantic warning"},
			{Type: "syntax", Severity: "error", Message: "another syntax error"},
		},
	}

	// Test HasErrors
	if !result.HasErrors() {
		t.Error("HasErrors() should return true when there are errors")
	}

	// Test GetErrorsByType
	syntaxErrors := result.GetErrorsByType("syntax")
	if len(syntaxErrors) != 2 {
		t.Errorf("GetErrorsByType('syntax') = %d errors, want 2", len(syntaxErrors))
	}

	semanticErrors := result.GetErrorsByType("semantic")
	if len(semanticErrors) != 1 {
		t.Errorf("GetErrorsByType('semantic') = %d errors, want 1", len(semanticErrors))
	}
}

func TestConfigValidator_ComplexConfig(t *testing.T) {
	validator := NewConfigValidator()

	complexConfig := `
-- This is a test configuration
gmacs.use_package("github.com/user/test-package", "v1.0.0")

-- Key bindings
gmacs.global_set_key("C-x C-f", "find-file")
gmacs.global_set_key("C-x C-s", "save-buffer")

-- Custom command
function my_custom_command()
    gmacs.message("Hello from custom command!")
end

gmacs.register_command("my-custom", my_custom_command, "My custom command")
gmacs.global_set_key("C-c m", "my-custom")

-- Some variables
gmacs.set_variable("auto-save", true)
gmacs.set_variable("tab-width", 4)
`

	result := validator.ValidateConfig(complexConfig)
	
	// This should be valid
	if result.HasErrors() {
		t.Errorf("Complex config should be valid, but got errors: %v", result.Errors)
	}

	// Should have no errors, but might have warnings
	errorCount := 0
	for _, err := range result.Errors {
		if err.Severity == "error" {
			errorCount++
		}
	}
	
	if errorCount > 0 {
		t.Errorf("Complex config should have no errors, but got %d", errorCount)
	}
}

func TestConfigValidator_ArgumentParsing(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name     string
		args     string
		expected []string
	}{
		{
			name:     "simple arguments",
			args:     `"arg1", "arg2"`,
			expected: []string{`"arg1"`, `"arg2"`},
		},
		{
			name:     "arguments with spaces",
			args:     `"first arg", "second arg"`,
			expected: []string{`"first arg"`, `"second arg"`},
		},
		{
			name:     "mixed quotes",
			args:     `"arg1", 'arg2'`,
			expected: []string{`"arg1"`, `'arg2'`},
		},
		{
			name:     "arguments with commas inside quotes",
			args:     `"arg1, with comma", "arg2"`,
			expected: []string{`"arg1, with comma"`, `"arg2"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseStringArguments(tt.args)
			
			if len(result) != len(tt.expected) {
				t.Errorf("parseStringArguments() got %d args, want %d", len(result), len(tt.expected))
				return
			}
			
			for i, expected := range tt.expected {
				if strings.TrimSpace(result[i]) != expected {
					t.Errorf("parseStringArguments() arg[%d] = %q, want %q", i, result[i], expected)
				}
			}
		})
	}
}