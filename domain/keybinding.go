package domain

import (
	"strings"
)

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

// RawSequenceBinding represents raw escape sequences (like arrow keys)
type RawSequenceBinding struct {
	Sequence string
	Command  CommandFunc
}

// KeyBindingMap manages key bindings
type KeyBindingMap struct {
	sequenceBindings    []KeySequenceBinding
	rawSequenceBindings []RawSequenceBinding
	currentSequence     []KeyPress
}

func NewKeyBindingMap() *KeyBindingMap {
	kbm := &KeyBindingMap{
		sequenceBindings:    make([]KeySequenceBinding, 0),
		rawSequenceBindings: make([]RawSequenceBinding, 0),
		currentSequence:     make([]KeyPress, 0),
	}
	
	// Register default Emacs-style key bindings
	kbm.registerDefaultBindings()
	
	return kbm
}

// NewEmptyKeyBindingMap creates a KeyBindingMap without default bindings for testing
func NewEmptyKeyBindingMap() *KeyBindingMap {
	return &KeyBindingMap{
		sequenceBindings:    make([]KeySequenceBinding, 0),
		rawSequenceBindings: make([]RawSequenceBinding, 0),
		currentSequence:     make([]KeyPress, 0),
	}
}

func (kbm *KeyBindingMap) registerDefaultBindings() {
	// Cursor movement
	kbm.BindKeySequence("C-f", ForwardChar)      // C-f
	kbm.BindKeySequence("C-b", BackwardChar)     // C-b
	kbm.BindKeySequence("C-n", NextLine)        // C-n
	kbm.BindKeySequence("C-p", PreviousLine)    // C-p
	kbm.BindKeySequence("C-a", BeginningOfLine) // C-a
	kbm.BindKeySequence("C-e", EndOfLine)       // C-e
	
	// Deletion
	kbm.BindKeySequence("C-h", DeleteBackwardChar) // C-h: backspace
	kbm.BindKeySequence("C-d", DeleteChar)         // C-d: delete-char
	
	// Cancel/Quit
	kbm.BindKeySequence("C-g", KeyboardQuit)       // C-g: keyboard-quit
	
	// Scrolling
	kbm.BindKeySequence("C-v", PageDown)        // C-v (page down)
	kbm.BindRawSequence("\x1b[6~", PageDown)    // Page Down key
	kbm.BindRawSequence("\x1b[5~", PageUp)      // Page Up key
	
	// Arrow keys (ANSI escape sequences)
	kbm.BindRawSequence("\x1b[C", ForwardChar)  // Right arrow
	kbm.BindRawSequence("\x1b[D", BackwardChar) // Left arrow
	kbm.BindRawSequence("\x1b[B", NextLine)     // Down arrow
	kbm.BindRawSequence("\x1b[A", PreviousLine) // Up arrow
	
	// Multi-key sequences
	kbm.BindKeySequence("C-x C-c", Quit)        // C-x C-c: quit
	kbm.BindKeySequence("C-x C-f", FindFile)    // C-x C-f: find-file
}

// BindRawSequence adds a raw key sequence binding (like arrow keys)
func (kbm *KeyBindingMap) BindRawSequence(sequence string, command CommandFunc) {
	binding := RawSequenceBinding{
		Sequence: sequence,
		Command:  command,
	}
	kbm.rawSequenceBindings = append(kbm.rawSequenceBindings, binding)
}

// LookupSequence finds a command for the given key sequence (both raw and parsed)
func (kbm *KeyBindingMap) LookupSequence(sequence string) (CommandFunc, bool) {
	// First check raw sequence bindings (escape sequences)
	for _, binding := range kbm.rawSequenceBindings {
		if binding.Sequence == sequence {
			return binding.Command, true
		}
	}
	
	// Then check parsed key sequence bindings (like M-a, C-x C-c)
	parsedSequence := parseKeySequence(sequence)
	for _, binding := range kbm.sequenceBindings {
		if kbm.sequencesEqual(binding.Sequence, parsedSequence) {
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

// GetCurrentSequence returns the current key sequence in progress
func (kbm *KeyBindingMap) GetCurrentSequence() []KeyPress {
	return kbm.currentSequence
}

// HasKeySequenceBinding checks if a key sequence binding exists
func (kbm *KeyBindingMap) HasKeySequenceBinding(keySequence string) (CommandFunc, bool) {
	sequence := parseKeySequence(keySequence)
	
	for _, binding := range kbm.sequenceBindings {
		if kbm.sequencesEqual(binding.Sequence, sequence) {
			return binding.Command, true
		}
	}
	return nil, false
}

// sequencesEqual checks if two key sequences are equal
func (kbm *KeyBindingMap) sequencesEqual(seq1, seq2 []KeyPress) bool {
	if len(seq1) != len(seq2) {
		return false
	}
	
	for i, press := range seq1 {
		if press.Key != seq2[i].Key || 
		   press.Ctrl != seq2[i].Ctrl || 
		   press.Meta != seq2[i].Meta {
			return false
		}
	}
	return true
}

// FormatSequence formats a key sequence as a display string
func FormatSequence(sequence []KeyPress) string {
	if len(sequence) == 0 {
		return ""
	}
	
	parts := make([]string, len(sequence))
	for i, press := range sequence {
		var keyStr string
		if press.Ctrl && press.Meta {
			keyStr = "C-M-" + press.Key
		} else if press.Ctrl {
			keyStr = "C-" + press.Key
		} else if press.Meta {
			keyStr = "M-" + press.Key
		} else {
			keyStr = press.Key
		}
		parts[i] = keyStr
	}
	
	return strings.Join(parts, " ") + " -"
}