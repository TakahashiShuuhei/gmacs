package sample

import (
	"fmt"
	"strings"
	
	"github.com/yuin/gopher-lua"
	pkg "github.com/TakahashiShuuhei/gmacs/internal/package"
)

// RubyMode is a sample package that provides Ruby language support
type RubyMode struct {
	enabled    bool
	rubyPath   string
	enableLint bool
}

// NewRubyMode creates a new Ruby mode package
func NewRubyMode() *RubyMode {
	return &RubyMode{
		rubyPath:   "/usr/bin/ruby",
		enableLint: true,
	}
}

// Package interface implementation

func (rm *RubyMode) GetInfo() pkg.PackageInfo {
	return pkg.PackageInfo{
		Name:        "ruby-mode",
		Version:     "1.0.0",
		Description: "Ruby language support for gmacs",
		Author:      "gmacs-dev",
		URL:         "github.com/gmacs-dev/gmacs-ruby-mode",
		Dependencies: []string{},
		Keywords:    []string{"ruby", "language", "syntax"},
	}
}

func (rm *RubyMode) Initialize() error {
	fmt.Printf("Initializing Ruby mode...\n")
	return nil
}

func (rm *RubyMode) Cleanup() error {
	fmt.Printf("Cleaning up Ruby mode...\n")
	return nil
}

func (rm *RubyMode) IsEnabled() bool {
	return rm.enabled
}

func (rm *RubyMode) Enable() error {
	rm.enabled = true
	fmt.Printf("Ruby mode enabled\n")
	return nil
}

func (rm *RubyMode) Disable() error {
	rm.enabled = false
	fmt.Printf("Ruby mode disabled\n")
	return nil
}

// ConfigurablePackage interface implementation

func (rm *RubyMode) SetConfig(config map[string]interface{}) error {
	if rubyPath, ok := config["ruby_path"].(string); ok {
		rm.rubyPath = rubyPath
	}
	
	if enableLint, ok := config["enable_lint"].(bool); ok {
		rm.enableLint = enableLint
	}
	
	return nil
}

func (rm *RubyMode) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"ruby_path":   rm.rubyPath,
		"enable_lint": rm.enableLint,
	}
}

func (rm *RubyMode) GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"ruby_path":   "/usr/bin/ruby",
		"enable_lint": true,
	}
}

// LuaAPIExtender interface implementation

func (rm *RubyMode) ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error {
	// Ruby-specific functions
	luaTable.RawSetString("show_doc", vm.NewFunction(rm.luaShowDoc))
	luaTable.RawSetString("goto_definition", vm.NewFunction(rm.luaGotoDefinition))
	luaTable.RawSetString("run_script", vm.NewFunction(rm.luaRunScript))
	luaTable.RawSetString("format_code", vm.NewFunction(rm.luaFormatCode))
	luaTable.RawSetString("get_gem_info", vm.NewFunction(rm.luaGetGemInfo))
	
	return nil
}

func (rm *RubyMode) GetNamespace() string {
	return "ruby"
}

// Lua function implementations

func (rm *RubyMode) luaShowDoc(L *lua.LState) int {
	method := L.CheckString(1)
	
	// Mock Ruby documentation lookup
	doc := rm.getRubyDocumentation(method)
	L.Push(lua.LString(doc))
	return 1
}

func (rm *RubyMode) luaGotoDefinition(L *lua.LState) int {
	symbol := L.CheckString(1)
	
	// Mock definition location lookup
	location := rm.findDefinition(symbol)
	
	// Return location as table
	locationTable := L.NewTable()
	locationTable.RawSetString("file", lua.LString(location.File))
	locationTable.RawSetString("line", lua.LNumber(location.Line))
	locationTable.RawSetString("column", lua.LNumber(location.Column))
	
	L.Push(locationTable)
	return 1
}

func (rm *RubyMode) luaRunScript(L *lua.LState) int {
	scriptPath := L.CheckString(1)
	
	// Mock script execution
	output := rm.runRubyScript(scriptPath)
	L.Push(lua.LString(output))
	return 1
}

func (rm *RubyMode) luaFormatCode(L *lua.LState) int {
	code := L.CheckString(1)
	
	// Mock code formatting
	formatted := rm.formatRubyCode(code)
	L.Push(lua.LString(formatted))
	return 1
}

func (rm *RubyMode) luaGetGemInfo(L *lua.LState) int {
	gemName := L.CheckString(1)
	
	// Mock gem information lookup
	gemInfo := rm.getGemInformation(gemName)
	
	// Return gem info as table
	gemTable := L.NewTable()
	gemTable.RawSetString("name", lua.LString(gemInfo.Name))
	gemTable.RawSetString("version", lua.LString(gemInfo.Version))
	gemTable.RawSetString("description", lua.LString(gemInfo.Description))
	
	L.Push(gemTable)
	return 1
}

// Helper types and methods

type Location struct {
	File   string
	Line   int
	Column int
}

type GemInfo struct {
	Name        string
	Version     string
	Description string
}

func (rm *RubyMode) getRubyDocumentation(method string) string {
	// Mock documentation - in real implementation, this would query ri or online docs
	docs := map[string]string{
		"String#gsub":    "gsub(pattern, replacement) → Returns a copy of str with all occurrences of pattern substituted",
		"Array#each":     "each {|item| block } → array. Calls the given block once for each element in self",
		"Hash#merge":     "merge(other_hash) → new_hash. Returns a new hash containing the contents of other_hash and self",
		"File.read":      "read(name, [length [, offset]]) → string. Opens the file and returns its contents",
	}
	
	if doc, exists := docs[method]; exists {
		return doc
	}
	
	return fmt.Sprintf("No documentation found for %s", method)
}

func (rm *RubyMode) findDefinition(symbol string) Location {
	// Mock definition lookup - in real implementation, this would use language server or ctags
	definitions := map[string]Location{
		"User":         {File: "app/models/user.rb", Line: 1, Column: 1},
		"Post":         {File: "app/models/post.rb", Line: 1, Column: 1},
		"authenticate": {File: "app/controllers/sessions_controller.rb", Line: 15, Column: 5},
	}
	
	if loc, exists := definitions[symbol]; exists {
		return loc
	}
	
	return Location{File: "unknown", Line: 0, Column: 0}
}

func (rm *RubyMode) runRubyScript(scriptPath string) string {
	// Mock script execution - in real implementation, this would execute the script
	return fmt.Sprintf("Executed %s successfully\nHello from Ruby script!", scriptPath)
}

func (rm *RubyMode) formatRubyCode(code string) string {
	// Mock code formatting - in real implementation, this would use rubocop or similar
	lines := strings.Split(code, "\n")
	formatted := make([]string, len(lines))
	
	for i, line := range lines {
		// Simple mock formatting: trim whitespace and add proper indentation
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "end") || strings.HasPrefix(trimmed, "}") {
			formatted[i] = trimmed
		} else if strings.Contains(trimmed, "def ") || strings.Contains(trimmed, "class ") {
			formatted[i] = trimmed
		} else {
			formatted[i] = "  " + trimmed
		}
	}
	
	return strings.Join(formatted, "\n")
}

func (rm *RubyMode) getGemInformation(gemName string) GemInfo {
	// Mock gem information lookup - in real implementation, this would query rubygems.org
	gems := map[string]GemInfo{
		"rails":    {Name: "rails", Version: "7.0.0", Description: "Full-stack web application framework"},
		"rspec":    {Name: "rspec", Version: "3.11.0", Description: "BDD testing framework for Ruby"},
		"rubocop":  {Name: "rubocop", Version: "1.30.0", Description: "Ruby static code analyzer and formatter"},
		"sidekiq":  {Name: "sidekiq", Version: "6.5.0", Description: "Background job processing for Ruby"},
	}
	
	if gem, exists := gems[gemName]; exists {
		return gem
	}
	
	return GemInfo{Name: gemName, Version: "unknown", Description: "Gem not found"}
}