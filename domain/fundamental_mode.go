package domain

import (
	"regexp"
)

// FundamentalMode implements the basic major mode
type FundamentalMode struct {
	name        string
	keyBindings *KeyBindingMap
	commands    map[string]*Command
}

// NewFundamentalMode creates a new fundamental mode instance
func NewFundamentalMode() *FundamentalMode {
	mode := &FundamentalMode{
		name:        "fundamental-mode",
		keyBindings: NewEmptyKeyBindingMap(),
		commands:    make(map[string]*Command),
	}
	
	// Register basic commands
	mode.registerCommands()
	
	// Set up basic key bindings
	mode.setupKeyBindings()
	
	return mode
}

// Name returns the mode name
func (fm *FundamentalMode) Name() string {
	return fm.name
}

// FilePattern returns the file pattern (nil for fundamental mode)
func (fm *FundamentalMode) FilePattern() *regexp.Regexp {
	return nil // Fundamental mode doesn't match specific file patterns
}

// KeyBindings returns the key bindings for this mode
func (fm *FundamentalMode) KeyBindings() *KeyBindingMap {
	return fm.keyBindings
}

// Commands returns the commands for this mode
func (fm *FundamentalMode) Commands() map[string]*Command {
	return fm.commands
}

// IndentFunction returns the indentation function
func (fm *FundamentalMode) IndentFunction() IndentFunc {
	return fm.basicIndent
}

// SyntaxHighlighting returns the syntax highlighter (nil for fundamental mode)
func (fm *FundamentalMode) SyntaxHighlighting() SyntaxHighlighter {
	return nil // No syntax highlighting in fundamental mode
}

// Initialize initializes the mode for a buffer
func (fm *FundamentalMode) Initialize(buffer *Buffer) error {
	// No special initialization needed for fundamental mode
	return nil
}

// OnActivate is called when the mode is activated
func (fm *FundamentalMode) OnActivate(buffer *Buffer) error {
	// No special activation needed for fundamental mode
	return nil
}

// OnDeactivate is called when the mode is deactivated
func (fm *FundamentalMode) OnDeactivate(buffer *Buffer) error {
	// No special deactivation needed for fundamental mode
	return nil
}

// registerCommands registers mode-specific commands
func (fm *FundamentalMode) registerCommands() {
	// Fundamental mode has no specific commands beyond global ones
	// This can be extended later
}

// setupKeyBindings sets up mode-specific key bindings
func (fm *FundamentalMode) setupKeyBindings() {
	// Fundamental mode uses only global key bindings
	// This can be extended later
}

// basicIndent provides basic indentation (no indentation)
func (fm *FundamentalMode) basicIndent(buffer *Buffer, line int) int {
	return 0 // No automatic indentation
}