package domain

import (
	"regexp"
)

// TextMode implements a text editing major mode
type TextMode struct {
	name        string
	keyBindings *KeyBindingMap
	commands    map[string]Command
	filePattern *regexp.Regexp
}

// NewTextMode creates a new text mode instance
func NewTextMode() *TextMode {
	mode := &TextMode{
		name:        "text-mode",
		keyBindings: NewEmptyKeyBindingMap(),
		commands:    make(map[string]Command),
		filePattern: regexp.MustCompile(`\.(txt|text|md|markdown|org)$`),
	}
	
	// Register text-specific commands
	mode.registerCommands()
	
	// Set up text-specific key bindings
	mode.setupKeyBindings()
	
	return mode
}

// Name returns the mode name
func (tm *TextMode) Name() string {
	return tm.name
}

// FilePattern returns the file pattern for text files
func (tm *TextMode) FilePattern() *regexp.Regexp {
	return tm.filePattern
}

// KeyBindings returns the key bindings for this mode
func (tm *TextMode) KeyBindings() *KeyBindingMap {
	return tm.keyBindings
}

// Commands returns the commands for this mode
func (tm *TextMode) Commands() map[string]Command {
	return tm.commands
}

// IndentFunction returns the indentation function for text mode
func (tm *TextMode) IndentFunction() IndentFunc {
	return tm.textIndent
}

// SyntaxHighlighting returns the syntax highlighter (nil for text mode)
func (tm *TextMode) SyntaxHighlighting() SyntaxHighlighter {
	return nil // No syntax highlighting in text mode
}

// Initialize initializes the mode for a buffer
func (tm *TextMode) Initialize(buffer *Buffer) error {
	// Set up text-specific settings
	// Could add text-specific initialization here
	return nil
}

// OnActivate is called when the mode is activated
func (tm *TextMode) OnActivate(buffer *Buffer) error {
	// Could add text-specific activation logic here
	return nil
}

// OnDeactivate is called when the mode is deactivated
func (tm *TextMode) OnDeactivate(buffer *Buffer) error {
	// Could add text-specific deactivation logic here
	return nil
}

// registerCommands registers text-specific commands
func (tm *TextMode) registerCommands() {
	// Text mode specific commands can be added here
	// For example: paragraph filling, word wrapping, etc.
}

// setupKeyBindings sets up text-specific key bindings
func (tm *TextMode) setupKeyBindings() {
	// Text mode specific key bindings can be added here
	// For now, inherits all global bindings
}

// textIndent provides basic text indentation (no automatic indentation)
func (tm *TextMode) textIndent(buffer *Buffer, line int) int {
	return 0 // No automatic indentation for text mode
}