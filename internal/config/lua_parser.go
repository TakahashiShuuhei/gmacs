package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// PackageDeclaration represents a package declaration from Lua config
type PackageDeclaration struct {
	URL     string
	Version string
	Config  map[string]any
}

// LuaParser parses Lua configuration files to extract package declarations
type LuaParser struct{}

// NewLuaParser creates a new Lua parser
func NewLuaParser() *LuaParser {
	return &LuaParser{}
}

// ParsePackageDeclarations extracts use_package declarations from a Lua file
func (p *LuaParser) ParsePackageDeclarations(configPath string) ([]PackageDeclaration, error) {
	// Read the Lua file
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file, return empty declarations
			return []PackageDeclaration{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	return p.parsePackageDeclarationsFromContent(string(content))
}

// ParsePackageDeclarationsFromContent parses package declarations from Lua content (exported for testing)
func (p *LuaParser) ParsePackageDeclarationsFromContent(content string) ([]PackageDeclaration, error) {
	return p.parsePackageDeclarationsFromContent(content)
}

// parsePackageDeclarationsFromContent parses package declarations from Lua content
func (p *LuaParser) parsePackageDeclarationsFromContent(content string) ([]PackageDeclaration, error) {
	var declarations []PackageDeclaration

	// Process line by line to exclude comments
	lines := strings.Split(content, "\n")
	var activeLines []string
	
	for _, line := range lines {
		// Remove comments (everything after --)
		if commentIndex := strings.Index(line, "--"); commentIndex >= 0 {
			line = line[:commentIndex]
		}
		// Only add non-empty lines
		line = strings.TrimSpace(line)
		if line != "" {
			activeLines = append(activeLines, line)
		}
	}
	
	// Join active lines back for pattern matching
	activeContent := strings.Join(activeLines, "\n")

	// Regular expression to match use_package calls
	// Matches both:
	// gmacs.use_package("url", "version")
	// gmacs.use_package("url", {config})
	// gmacs.use_package("url")
	usePackageRegex := regexp.MustCompile(`gmacs\.use_package\s*\(\s*"([^"]+)"(?:\s*,\s*([^)]+))?\s*\)`)

	matches := usePackageRegex.FindAllStringSubmatch(activeContent, -1)

	for _, match := range matches {
		url := match[1]
		if url == "" {
			continue
		}

		decl := PackageDeclaration{
			URL:     url,
			Version: "latest", // default version
			Config:  make(map[string]any),
		}

		// Parse second parameter (version or config)
		if len(match) > 2 && match[2] != "" {
			secondParam := strings.TrimSpace(match[2])
			
			if strings.HasPrefix(secondParam, "\"") && strings.HasSuffix(secondParam, "\"") {
				// It's a version string
				version := strings.Trim(secondParam, "\"")
				if version != "" {
					decl.Version = version
				}
			} else if strings.HasPrefix(secondParam, "{") && strings.HasSuffix(secondParam, "}") {
				// It's a config table
				config, err := p.parseSimpleLuaTable(secondParam)
				if err != nil {
					return nil, fmt.Errorf("failed to parse config for package %s: %v", url, err)
				}
				decl.Config = config
			}
		}

		declarations = append(declarations, decl)
	}

	return declarations, nil
}

// parseSimpleLuaTable parses a simple Lua table into a Go map
// This is a simplified parser that handles basic key-value pairs
// It supports string, number, and boolean values
func (p *LuaParser) parseSimpleLuaTable(tableStr string) (map[string]any, error) {
	result := make(map[string]any)

	// Remove braces
	tableStr = strings.TrimSpace(tableStr)
	if !strings.HasPrefix(tableStr, "{") || !strings.HasSuffix(tableStr, "}") {
		return nil, fmt.Errorf("invalid table format")
	}
	
	content := strings.TrimSpace(tableStr[1 : len(tableStr)-1])
	if content == "" {
		return result, nil
	}

	// Split by commas (simple approach)
	parts := strings.Split(content, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Look for key = value pattern
		if strings.Contains(part, "=") {
			keyValue := strings.SplitN(part, "=", 2)
			if len(keyValue) != 2 {
				continue
			}

			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])

			// Remove quotes from key if present
			if strings.HasPrefix(key, "\"") && strings.HasSuffix(key, "\"") {
				key = strings.Trim(key, "\"")
			}

			// Parse value
			parsedValue, err := p.parseValue(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value for key %s: %v", key, err)
			}

			result[key] = parsedValue
		}
	}

	return result, nil
}

// parseValue parses a Lua value (string, number, boolean)
func (p *LuaParser) parseValue(value string) (any, error) {
	value = strings.TrimSpace(value)

	// String value
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\""), nil
	}

	// Boolean values
	if value == "true" {
		return true, nil
	}
	if value == "false" {
		return false, nil
	}

	// Number value
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal, nil
	}

	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal, nil
	}

	// If all else fails, treat as string
	return value, nil
}