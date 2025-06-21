package domain

import (
	"regexp"
)

// IndentFunc defines a function type for indentation logic
type IndentFunc func(buffer *Buffer, line int) int

// SyntaxHighlighter defines interface for syntax highlighting
type SyntaxHighlighter interface {
	HighlightLine(line string) []HighlightSegment
	GetTokens(content []string) []Token
}

// HighlightSegment represents a highlighted portion of text
type HighlightSegment struct {
	Start  int
	End    int
	Style  string
}

// Token represents a syntax token
type Token struct {
	Type     string
	Value    string
	Position Position
}

// MajorMode represents a major editing mode
type MajorMode interface {
	Name() string
	FilePattern() *regexp.Regexp
	KeyBindings() *KeyBindingMap
	Commands() map[string]*Command
	IndentFunction() IndentFunc
	SyntaxHighlighting() SyntaxHighlighter
	Initialize(buffer *Buffer) error
	OnActivate(buffer *Buffer) error
	OnDeactivate(buffer *Buffer) error
}

// MinorMode represents a minor editing mode
type MinorMode interface {
	Name() string
	KeyBindings() *KeyBindingMap
	Commands() map[string]*Command
	Enable(buffer *Buffer) error
	Disable(buffer *Buffer) error
	IsEnabled(buffer *Buffer) bool
	Priority() int // Higher priority modes override lower priority ones
}

// ModeManager manages major and minor modes
type ModeManager struct {
	majorModes map[string]MajorMode
	minorModes map[string]MinorMode
}

// NewModeManager creates a new mode manager
func NewModeManager() *ModeManager {
	mm := &ModeManager{
		majorModes: make(map[string]MajorMode),
		minorModes: make(map[string]MinorMode),
	}
	
	// Register default modes
	mm.registerDefaultModes()
	
	return mm
}

// RegisterMajorMode registers a major mode
func (mm *ModeManager) RegisterMajorMode(mode MajorMode) {
	mm.majorModes[mode.Name()] = mode
}

// RegisterMinorMode registers a minor mode
func (mm *ModeManager) RegisterMinorMode(mode MinorMode) {
	mm.minorModes[mode.Name()] = mode
}

// SetMajorMode sets the major mode for a buffer
func (mm *ModeManager) SetMajorMode(buffer *Buffer, modeName string) error {
	mode, exists := mm.majorModes[modeName]
	if !exists {
		return &ModeError{Message: "Unknown major mode: " + modeName}
	}
	
	// Deactivate current major mode if any
	if buffer.majorMode != nil {
		buffer.majorMode.OnDeactivate(buffer)
	}
	
	// Set new major mode
	buffer.majorMode = mode
	
	// Initialize and activate
	if err := mode.Initialize(buffer); err != nil {
		return err
	}
	
	return mode.OnActivate(buffer)
}

// ToggleMinorMode toggles a minor mode for a buffer
func (mm *ModeManager) ToggleMinorMode(buffer *Buffer, modeName string) error {
	mode, exists := mm.minorModes[modeName]
	if !exists {
		return &ModeError{Message: "Unknown minor mode: " + modeName}
	}
	
	if mode.IsEnabled(buffer) {
		return mode.Disable(buffer)
	} else {
		return mode.Enable(buffer)
	}
}

// GetEffectiveKeyBindings returns the effective key bindings for a buffer
func (mm *ModeManager) GetEffectiveKeyBindings(buffer *Buffer) *KeyBindingMap {
	// Start with global bindings
	effective := NewKeyBindingMap()
	
	// Add major mode bindings
	if buffer.majorMode != nil {
		majorBindings := buffer.majorMode.KeyBindings()
		if majorBindings != nil {
			effective = mm.mergeKeyBindings(effective, majorBindings)
		}
	}
	
	// Add minor mode bindings in priority order
	for _, mode := range buffer.getEnabledMinorModes() {
		minorBindings := mode.KeyBindings()
		if minorBindings != nil {
			effective = mm.mergeKeyBindings(effective, minorBindings)
		}
	}
	
	return effective
}

// GetMajorMode returns the major mode for a buffer
func (mm *ModeManager) GetMajorMode(buffer *Buffer) MajorMode {
	return buffer.majorMode
}

// GetMinorModes returns all enabled minor modes for a buffer
func (mm *ModeManager) GetMinorModes(buffer *Buffer) []MinorMode {
	return buffer.getEnabledMinorModes()
}

// AutoDetectMajorMode detects the appropriate major mode based on file extension
func (mm *ModeManager) AutoDetectMajorMode(buffer *Buffer) (MajorMode, error) {
	filepath := buffer.Filepath()
	if filepath == "" {
		// Default to fundamental mode for buffers without files
		return mm.majorModes["fundamental-mode"], nil
	}
	
	// Try to match file patterns
	for _, mode := range mm.majorModes {
		if pattern := mode.FilePattern(); pattern != nil {
			if pattern.MatchString(filepath) {
				return mode, nil
			}
		}
	}
	
	// Default to fundamental mode for unmatched files
	return mm.majorModes["fundamental-mode"], nil
}

// mergeKeyBindings merges two key binding maps, with the second taking precedence
func (mm *ModeManager) mergeKeyBindings(base, override *KeyBindingMap) *KeyBindingMap {
	// For now, return override (later we can implement proper merging)
	return override
}

// registerDefaultModes registers the default modes
func (mm *ModeManager) registerDefaultModes() {
	// Register fundamental mode
	mm.RegisterMajorMode(NewFundamentalMode())
	
	// Register text mode
	mm.RegisterMajorMode(NewTextMode())
	
	// Register minor modes
	mm.RegisterMinorMode(NewAutoAMode())
}

// ModeError represents an error in mode operations
type ModeError struct {
	Message string
}

func (e *ModeError) Error() string {
	return e.Message
}