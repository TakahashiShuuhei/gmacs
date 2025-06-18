package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

// KeyBinding represents a key combination and its associated command
type KeyBinding struct {
	Key     string
	Ctrl    bool
	Meta    bool
	Command CommandFunc
}

// KeyBindingMap manages key bindings
type KeyBindingMap struct {
	bindings []KeyBinding
}

func NewKeyBindingMap() *KeyBindingMap {
	kbm := &KeyBindingMap{
		bindings: make([]KeyBinding, 0),
	}
	
	// Register default Emacs-style key bindings
	kbm.registerDefaultBindings()
	
	return kbm
}

func (kbm *KeyBindingMap) registerDefaultBindings() {
	// Cursor movement
	kbm.Bind("f", true, false, ForwardChar)      // C-f
	kbm.Bind("b", true, false, BackwardChar)     // C-b
	kbm.Bind("n", true, false, NextLine)        // C-n
	kbm.Bind("p", true, false, PreviousLine)    // C-p
	kbm.Bind("a", true, false, BeginningOfLine) // C-a
	kbm.Bind("e", true, false, EndOfLine)       // C-e
	
	// Arrow keys (ANSI escape sequences)
	kbm.BindSequence("\x1b[C", ForwardChar)     // Right arrow
	kbm.BindSequence("\x1b[D", BackwardChar)    // Left arrow
	kbm.BindSequence("\x1b[B", NextLine)        // Down arrow
	kbm.BindSequence("\x1b[A", PreviousLine)    // Up arrow
	
	log.Debug("Registered default key bindings")
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
	log.Debug("Bound key: %s (ctrl=%t, meta=%t)", key, ctrl, meta)
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
	log.Debug("Bound key sequence: %q", sequence)
}

// Lookup finds a command for the given key combination
func (kbm *KeyBindingMap) Lookup(key string, ctrl, meta bool) (CommandFunc, bool) {
	for _, binding := range kbm.bindings {
		if binding.Key == key && binding.Ctrl == ctrl && binding.Meta == meta {
			log.Debug("Found binding for key: %s (ctrl=%t, meta=%t)", key, ctrl, meta)
			return binding.Command, true
		}
	}
	return nil, false
}

// LookupSequence finds a command for the given key sequence
func (kbm *KeyBindingMap) LookupSequence(sequence string) (CommandFunc, bool) {
	for _, binding := range kbm.bindings {
		if binding.Key == sequence && !binding.Ctrl && !binding.Meta {
			log.Debug("Found binding for sequence: %q", sequence)
			return binding.Command, true
		}
	}
	return nil, false
}