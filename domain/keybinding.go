package domain

import (
	"strings"
)

// KeyBinding represents a key combination and its associated command
type KeyBinding struct {
	Key     string
	Ctrl    bool
	Meta    bool
	Command CommandFunc
}

// KeySequenceBinding represents a multi-key sequence binding
type KeySequenceBinding struct {
	Sequence []KeyPress
	Command  CommandFunc
}

// KeyPress represents a single key press in a sequence
type KeyPress struct {
	Key  string
	Ctrl bool
	Meta bool
}

// KeyBindingMap manages key bindings
type KeyBindingMap struct {
	bindings         []KeyBinding
	sequenceBindings []KeySequenceBinding
	currentSequence  []KeyPress
}

func NewKeyBindingMap() *KeyBindingMap {
	kbm := &KeyBindingMap{
		bindings:         make([]KeyBinding, 0),
		sequenceBindings: make([]KeySequenceBinding, 0),
		currentSequence:  make([]KeyPress, 0),
	}
	
	// Register default Emacs-style key bindings
	kbm.registerDefaultBindings()
	
	return kbm
}

// NewEmptyKeyBindingMap creates a KeyBindingMap without default bindings for testing
func NewEmptyKeyBindingMap() *KeyBindingMap {
	return &KeyBindingMap{
		bindings:         make([]KeyBinding, 0),
		sequenceBindings: make([]KeySequenceBinding, 0),
		currentSequence:  make([]KeyPress, 0),
	}
}

func (kbm *KeyBindingMap) registerDefaultBindings() {
	// Cursor movement
	kbm.Bind("f", true, false, ForwardChar)      // C-f
	kbm.Bind("b", true, false, BackwardChar)     // C-b
	kbm.Bind("n", true, false, NextLine)        // C-n
	kbm.Bind("p", true, false, PreviousLine)    // C-p
	kbm.Bind("a", true, false, BeginningOfLine) // C-a
	kbm.Bind("e", true, false, EndOfLine)       // C-e
	
	// Scrolling
	kbm.Bind("v", true, false, PageDown)        // C-v (page down)
	kbm.BindSequence("\x1b[6~", PageDown)       // Page Down key
	kbm.BindSequence("\x1b[5~", PageUp)         // Page Up key
	
	// Arrow keys (ANSI escape sequences)
	kbm.BindSequence("\x1b[C", ForwardChar)     // Right arrow
	kbm.BindSequence("\x1b[D", BackwardChar)    // Left arrow
	kbm.BindSequence("\x1b[B", NextLine)        // Down arrow
	kbm.BindSequence("\x1b[A", PreviousLine)    // Up arrow
	
	// Multi-key sequences
	kbm.BindKeySequence("C-x C-c", Quit)        // C-x C-c: quit
}

// Bind adds a key binding
func (kbm *KeyBindingMap) Bind(key string, ctrl, meta bool, command CommandFunc) {
	binding := KeyBinding{
		Key:     key,
		Ctrl:    ctrl,
		Meta:    meta,
		Command: command,
	}
	kbm.bindings = append(kbm.bindings, binding)
}

// BindSequence adds a key sequence binding (like arrow keys)
func (kbm *KeyBindingMap) BindSequence(sequence string, command CommandFunc) {
	binding := KeyBinding{
		Key:     sequence,
		Ctrl:    false,
		Meta:    false,
		Command: command,
	}
	kbm.bindings = append(kbm.bindings, binding)
}

// Lookup finds a command for the given key combination
func (kbm *KeyBindingMap) Lookup(key string, ctrl, meta bool) (CommandFunc, bool) {
	for _, binding := range kbm.bindings {
		if binding.Key == key && binding.Ctrl == ctrl && binding.Meta == meta {
			return binding.Command, true
		}
	}
	return nil, false
}

// LookupSequence finds a command for the given key sequence
func (kbm *KeyBindingMap) LookupSequence(sequence string) (CommandFunc, bool) {
	for _, binding := range kbm.bindings {
		if binding.Key == sequence && !binding.Ctrl && !binding.Meta {
			return binding.Command, true
		}
	}
	return nil, false
}

// BindKeySequence adds a multi-key sequence binding like "C-x C-c"
func (kbm *KeyBindingMap) BindKeySequence(keySequence string, command CommandFunc) {
	sequence := parseKeySequence(keySequence)
	binding := KeySequenceBinding{
		Sequence: sequence,
		Command:  command,
	}
	kbm.sequenceBindings = append(kbm.sequenceBindings, binding)
}

// parseKeySequence parses a string like "C-x C-c" into []KeyPress
func parseKeySequence(keySequence string) []KeyPress {
	parts := strings.Fields(keySequence)
	sequence := make([]KeyPress, 0, len(parts))
	
	for _, part := range parts {
		keyPress := parseKeyPress(part)
		sequence = append(sequence, keyPress)
	}
	
	return sequence
}

// parseKeyPress parses a string like "C-x" or "M-x" into KeyPress
func parseKeyPress(keyStr string) KeyPress {
	parts := strings.Split(keyStr, "-")
	
	keyPress := KeyPress{
		Key:  "",
		Ctrl: false,
		Meta: false,
	}
	
	for i, part := range parts {
		switch part {
		case "C":
			keyPress.Ctrl = true
		case "M":
			keyPress.Meta = true
		default:
			// The last part is the actual key
			if i == len(parts)-1 {
				keyPress.Key = part
			}
		}
	}
	
	return keyPress
}

// ProcessKeyPress processes a key press and returns a command if a sequence is completed
func (kbm *KeyBindingMap) ProcessKeyPress(key string, ctrl, meta bool) (CommandFunc, bool, bool) {
	currentPress := KeyPress{Key: key, Ctrl: ctrl, Meta: meta}
	
	// Add to current sequence
	kbm.currentSequence = append(kbm.currentSequence, currentPress)
	
	// Check if current sequence matches any binding
	for _, binding := range kbm.sequenceBindings {
		if kbm.matchesSequence(binding.Sequence) {
			// Complete match - reset sequence and return command
			kbm.currentSequence = make([]KeyPress, 0)
			return binding.Command, true, false
		}
		
		if kbm.isPrefixOf(binding.Sequence) {
			// Partial match - continue sequence
			return nil, false, true
		}
	}
	
	// No match - reset sequence
	kbm.currentSequence = make([]KeyPress, 0)
	return nil, false, false
}

// matchesSequence checks if current sequence exactly matches the given sequence
func (kbm *KeyBindingMap) matchesSequence(sequence []KeyPress) bool {
	if len(kbm.currentSequence) != len(sequence) {
		return false
	}
	
	for i, press := range kbm.currentSequence {
		if press.Key != sequence[i].Key || 
		   press.Ctrl != sequence[i].Ctrl || 
		   press.Meta != sequence[i].Meta {
			return false
		}
	}
	
	return true
}

// isPrefixOf checks if current sequence is a prefix of the given sequence
func (kbm *KeyBindingMap) isPrefixOf(sequence []KeyPress) bool {
	if len(kbm.currentSequence) >= len(sequence) {
		return false
	}
	
	for i, press := range kbm.currentSequence {
		if press.Key != sequence[i].Key || 
		   press.Ctrl != sequence[i].Ctrl || 
		   press.Meta != sequence[i].Meta {
			return false
		}
	}
	
	return true
}

// ResetSequence resets the current key sequence
func (kbm *KeyBindingMap) ResetSequence() {
	kbm.currentSequence = make([]KeyPress, 0)
}