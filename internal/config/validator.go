package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/yuin/gopher-lua"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Type        string // "syntax", "semantic", "package", "keybinding"
	Message     string
	Line        int
	Column      int
	Severity    string // "error", "warning", "info"
	Suggestions []string
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.Line > 0 {
		return fmt.Sprintf("%s:%d:%d: %s: %s", ve.Type, ve.Line, ve.Column, ve.Severity, ve.Message)
	}
	return fmt.Sprintf("%s: %s: %s", ve.Type, ve.Severity, ve.Message)
}

// ValidationResult contains the results of configuration validation
type ValidationResult struct {
	Valid   bool
	Errors  []*ValidationError
	FilePath string
}

// HasErrors returns true if there are any errors (not warnings)
func (vr *ValidationResult) HasErrors() bool {
	for _, err := range vr.Errors {
		if err.Severity == "error" {
			return true
		}
	}
	return false
}

// GetErrorsByType returns errors of a specific type
func (vr *ValidationResult) GetErrorsByType(errorType string) []*ValidationError {
	var filtered []*ValidationError
	for _, err := range vr.Errors {
		if err.Type == errorType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// ConfigValidator validates Lua configuration files
type ConfigValidator struct {
	knownCommands   map[string]bool
	knownVariables  map[string]bool
	packageManager  PackageManagerInterface
}

// PackageManagerInterface defines the interface for package validation
type PackageManagerInterface interface {
	ValidatePackageURL(url string) error
	GetAvailableVersions(url string) ([]string, error)
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		knownCommands: map[string]bool{
			"quit":                    true,
			"version":                 true,
			"hello":                   true,
			"list-commands":           true,
			"self-insert-command":     true,
			"forward-char":            true,
			"backward-char":           true,
			"next-line":               true,
			"previous-line":           true,
			"delete-char":             true,
			"backward-delete-char":    true,
			"find-file":               true,
			"save-buffer":             true,
		},
		knownVariables: map[string]bool{
			"auto-save":        true,
			"backup-directory": true,
			"tab-width":        true,
			"indent-tabs-mode": true,
			"show-line-numbers": true,
		},
	}
}

// SetPackageManager sets the package manager for package validation
func (cv *ConfigValidator) SetPackageManager(pm PackageManagerInterface) {
	cv.packageManager = pm
}

// ValidateFile validates a Lua configuration file
func (cv *ConfigValidator) ValidateFile(filePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		FilePath: filePath,
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, &ValidationError{
			Type:     "file",
			Message:  fmt.Sprintf("Configuration file does not exist: %s", filePath),
			Severity: "error",
		})
		result.Valid = false
		return result, nil
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, &ValidationError{
			Type:     "file",
			Message:  fmt.Sprintf("Failed to read configuration file: %v", err),
			Severity: "error",
		})
		result.Valid = false
		return result, nil
	}

	// Validate syntax
	cv.validateSyntax(string(content), result)

	// Validate semantics (even if there are syntax errors, to provide comprehensive feedback)
	cv.validateSemantics(string(content), result)

	// Set overall validity
	result.Valid = !result.HasErrors()

	return result, nil
}

// ValidateConfig validates configuration content directly
func (cv *ConfigValidator) ValidateConfig(content string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		FilePath: "<string>",
	}

	cv.validateSyntax(content, result)
	cv.validateSemantics(content, result)

	result.Valid = !result.HasErrors()
	return result
}

// validateSyntax checks Lua syntax validity
func (cv *ConfigValidator) validateSyntax(content string, result *ValidationResult) {
	// Create a new Lua state for syntax checking
	L := lua.NewState()
	defer L.Close()

	// Try to compile the Lua code
	_, err := L.LoadString(content)
	if err != nil {
		// Parse Lua error to extract line and column information
		errorMsg := err.Error()
		line, column := cv.parseLuaError(errorMsg)

		result.Errors = append(result.Errors, &ValidationError{
			Type:     "syntax",
			Message:  cv.cleanLuaError(errorMsg),
			Line:     line,
			Column:   column,
			Severity: "error",
		})
	}
}

// validateSemantics checks semantic correctness of the configuration
func (cv *ConfigValidator) validateSemantics(content string, result *ValidationResult) {
	lines := strings.Split(content, "\n")
	customCommands := make(map[string]bool)

	// First pass: find custom command registrations
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "register_command") {
			if commandName := cv.extractCommandName(line); commandName != "" {
				customCommands[commandName] = true
			}
		}
	}

	// Second pass: validate with custom commands included
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		// Validate gmacs function calls
		cv.validateGmacsAPICalls(line, lineNum+1, result)

		// Validate package declarations
		cv.validatePackageDeclarations(line, lineNum+1, result)

		// Validate key bindings (with custom commands)
		cv.validateKeyBindingsWithCustomCommands(line, lineNum+1, result, customCommands)
	}
}

// validateGmacsAPICalls validates calls to gmacs API functions
func (cv *ConfigValidator) validateGmacsAPICalls(line string, lineNum int, result *ValidationResult) {
	// Check for unknown gmacs functions
	if strings.Contains(line, "gmacs.") {
		// Extract function name
		if idx := strings.Index(line, "gmacs."); idx >= 0 {
			remaining := line[idx+6:] // Skip "gmacs."
			if dotIdx := strings.Index(remaining, "("); dotIdx >= 0 {
				funcName := remaining[:dotIdx]
				
				// Check if it's a known function
				knownFunctions := map[string]bool{
					"set_variable":       true,
					"get_variable":       true,
					"global_set_key":     true,
					"local_set_key":      true,
					"register_command":   true,
					"execute_command":    true,
					"list_commands":      true,
					"message":            true,
					"error":              true,
					"current_buffer":     true,
					"current_word":       true,
					"current_char":       true,
					"add_hook":           true,
					"use_package":        true,
					"after_packages_loaded": true,
				}

				if !knownFunctions[funcName] && !strings.Contains(funcName, ".") {
					result.Errors = append(result.Errors, &ValidationError{
						Type:     "semantic",
						Message:  fmt.Sprintf("Unknown gmacs function: %s", funcName),
						Line:     lineNum,
						Severity: "error",
						Suggestions: cv.suggestSimilarFunctions(funcName, knownFunctions),
					})
				}
			}
		}
	}
}

// validatePackageDeclarations validates use_package calls
func (cv *ConfigValidator) validatePackageDeclarations(line string, lineNum int, result *ValidationResult) {
	if strings.Contains(line, "use_package") {
		// Extract package URL and version
		if strings.Contains(line, "gmacs.use_package") {
			// Parse the arguments
			if idx := strings.Index(line, "("); idx >= 0 {
				args := line[idx+1:]
				if endIdx := strings.Index(args, ")"); endIdx >= 0 {
					args = args[:endIdx]
					parts := cv.parseStringArguments(args)
					
					if len(parts) >= 1 {
						packageURL := strings.Trim(parts[0], "\"'")
						
						// Validate package URL format
						if !cv.isValidPackageURL(packageURL) {
							result.Errors = append(result.Errors, &ValidationError{
								Type:     "package",
								Message:  fmt.Sprintf("Invalid package URL format: %s", packageURL),
								Line:     lineNum,
								Severity: "error",
								Suggestions: []string{
									"Package URL should be in format: github.com/user/repo",
									"Supported hosts: github.com, gitlab.com, bitbucket.org",
								},
							})
						}
						
						// Validate version if provided
						if len(parts) >= 2 {
							version := strings.Trim(parts[1], "\"'")
							if version != "latest" && !cv.isValidVersion(version) {
								result.Errors = append(result.Errors, &ValidationError{
									Type:     "package",
									Message:  fmt.Sprintf("Invalid version format: %s", version),
									Line:     lineNum,
									Severity: "warning",
									Suggestions: []string{
										"Use 'latest' for latest version",
										"Use semantic version format: v1.0.0",
										"Use commit hash for specific commits",
									},
								})
							}
						}
					}
				}
			}
		}
	}
}


// Helper functions

// parseLuaError extracts line and column from Lua error message
func (cv *ConfigValidator) parseLuaError(errorMsg string) (line, column int) {
	// Lua error format: "<string>:line: message"
	parts := strings.Split(errorMsg, ":")
	if len(parts) >= 2 {
		// Try to parse line number
		if n, err := fmt.Sscanf(parts[1], "%d", &line); n == 1 && err == nil {
			return line, 0
		}
	}
	return 0, 0
}

// cleanLuaError removes file prefix from Lua error message
func (cv *ConfigValidator) cleanLuaError(errorMsg string) string {
	if idx := strings.Index(errorMsg, ": "); idx >= 0 {
		return errorMsg[idx+2:]
	}
	return errorMsg
}

// parseStringArguments parses comma-separated string arguments
func (cv *ConfigValidator) parseStringArguments(args string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)
	
	for i := 0; i < len(args); i++ {
		char := args[i]
		
		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
			current.WriteByte(char)
		} else if inQuotes && char == quoteChar {
			inQuotes = false
			current.WriteByte(char)
		} else if !inQuotes && char == ',' {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(char)
		}
	}
	
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	
	return parts
}

// isValidPackageURL validates package URL format
func (cv *ConfigValidator) isValidPackageURL(url string) bool {
	validPrefixes := []string{
		"github.com/",
		"gitlab.com/",
		"bitbucket.org/",
		"git.sr.ht/",
		"codeberg.org/",
	}
	
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			// Check if it has at least user/repo format
			remaining := url[len(prefix):]
			parts := strings.Split(remaining, "/")
			return len(parts) >= 2 && parts[0] != "" && parts[1] != ""
		}
	}
	
	return false
}

// isValidVersion validates version format
func (cv *ConfigValidator) isValidVersion(version string) bool {
	// Accept semantic versions, git tags, or commit hashes
	if strings.HasPrefix(version, "v") && len(version) > 1 {
		return true // Assume v1.0.0 format is valid
	}
	if len(version) == 40 {
		// Check if it's a valid hex string (git commit hash)
		for _, char := range version {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
		return true
	}
	if len(version) >= 7 && len(version) <= 12 {
		// Check if it's a valid hex string (short commit hash)
		for _, char := range version {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
		return true
	}
	return false
}

// isValidKeySequence validates key sequence format
func (cv *ConfigValidator) isValidKeySequence(keySeq string) bool {
	// Basic validation for key sequences like C-x, M-x, C-c C-f
	if keySeq == "" {
		return false
	}
	
	// Check for valid modifiers
	validPrefixes := []string{"C-", "M-", "S-"}
	parts := strings.Split(keySeq, " ")
	
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		
		// Allow single characters without modifiers
		if len(part) == 1 {
			continue
		}
		
		// Allow special key names like RET, TAB, ESC
		specialKeys := []string{"RET", "TAB", "ESC", "SPC", "DEL", "BS"}
		isSpecialKey := false
		for _, special := range specialKeys {
			if part == special {
				isSpecialKey = true
				break
			}
		}
		if isSpecialKey {
			continue
		}
		
		// Check for modifier prefix
		hasValidPrefix := false
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(part, prefix) && len(part) > len(prefix) {
				hasValidPrefix = true
				break
			}
		}
		
		if !hasValidPrefix {
			return false
		}
	}
	
	return true
}

// suggestSimilarFunctions suggests similar function names
func (cv *ConfigValidator) suggestSimilarFunctions(input string, knownFunctions map[string]bool) []string {
	var suggestions []string
	
	for funcName := range knownFunctions {
		if cv.calculateSimilarity(input, funcName) > 0.6 {
			suggestions = append(suggestions, funcName)
		}
	}
	
	return suggestions
}

// suggestSimilarCommands suggests similar command names
func (cv *ConfigValidator) suggestSimilarCommands(input string) []string {
	var suggestions []string
	
	for command := range cv.knownCommands {
		if cv.calculateSimilarity(input, command) > 0.6 {
			suggestions = append(suggestions, command)
		}
	}
	
	return suggestions
}

// calculateSimilarity calculates string similarity (simple implementation)
func (cv *ConfigValidator) calculateSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	
	// Simple similarity based on common characters
	commonChars := 0
	for _, char := range a {
		if strings.ContainsRune(b, char) {
			commonChars++
		}
	}
	
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	
	return float64(commonChars) / float64(maxLen)
}

// extractCommandName extracts command name from register_command call
func (cv *ConfigValidator) extractCommandName(line string) string {
	if idx := strings.Index(line, "register_command"); idx >= 0 {
		if parenIdx := strings.Index(line[idx:], "("); parenIdx >= 0 {
			args := line[idx+parenIdx+1:]
			if endIdx := strings.Index(args, ")"); endIdx >= 0 {
				args = args[:endIdx]
				parts := cv.parseStringArguments(args)
				if len(parts) >= 1 {
					return strings.Trim(parts[0], "\"'")
				}
			}
		}
	}
	return ""
}

// validateKeyBindingsWithCustomCommands validates key bindings with custom commands included
func (cv *ConfigValidator) validateKeyBindingsWithCustomCommands(line string, lineNum int, result *ValidationResult, customCommands map[string]bool) {
	if strings.Contains(line, "global_set_key") || strings.Contains(line, "local_set_key") {
		if idx := strings.Index(line, "("); idx >= 0 {
			args := line[idx+1:]
			if endIdx := strings.Index(args, ")"); endIdx >= 0 {
				args = args[:endIdx]
				parts := cv.parseStringArguments(args)
				
				if len(parts) >= 2 {
					keySeq := strings.Trim(parts[0], "\"'")
					command := strings.Trim(parts[1], "\"'")
					
					// Validate key sequence format
					if !cv.isValidKeySequence(keySeq) {
						result.Errors = append(result.Errors, &ValidationError{
							Type:     "keybinding",
							Message:  fmt.Sprintf("Invalid key sequence format: %s", keySeq),
							Line:     lineNum,
							Severity: "error",
							Suggestions: []string{
								"Use format like: C-x, M-x, C-c C-f",
								"C- for Ctrl, M- for Meta/Alt",
							},
						})
					}
					
					// Validate command exists (including custom commands)
					if !cv.knownCommands[command] && !customCommands[command] {
						result.Errors = append(result.Errors, &ValidationError{
							Type:     "keybinding",
							Message:  fmt.Sprintf("Unknown command: %s", command),
							Line:     lineNum,
							Severity: "warning",
							Suggestions: cv.suggestSimilarCommands(command),
						})
					}
				}
			}
		}
	}
}