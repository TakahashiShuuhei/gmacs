package keymap

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Key represents a key input (similar to Emacs key representation)
type Key struct {
	Char     rune
	Ctrl     bool
	Alt      bool
	Shift    bool
	Special  string // For special keys like "return", "tab", "backspace", etc.
}

// String returns a string representation of the key
func (k Key) String() string {
	var parts []string
	
	if k.Ctrl {
		parts = append(parts, "C")
	}
	if k.Alt {
		parts = append(parts, "M")
	}
	if k.Shift && k.Special != "" {
		parts = append(parts, "S")
	}
	
	if k.Special != "" {
		parts = append(parts, k.Special)
	} else {
		parts = append(parts, string(k.Char))
	}
	
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "-") + "-" + parts[len(parts)-1]
	}
	return parts[0]
}

// KeySequence represents a sequence of keys
type KeySequence []Key

// String returns a string representation of the key sequence
func (ks KeySequence) String() string {
	if len(ks) == 0 {
		return ""
	}
	
	var parts []string
	for _, key := range ks {
		parts = append(parts, key.String())
	}
	
	return strings.Join(parts, " ")
}

// Binding represents a key binding to a command
type Binding struct {
	KeySeq    KeySequence
	Command   string
	Args      []interface{}
}

// Keymap represents a keymap (similar to Emacs keymap)
type Keymap struct {
	mu       sync.RWMutex
	bindings map[string]*Binding // key sequence string -> binding
	parent   *Keymap             // parent keymap for inheritance
	name     string
}

// New creates a new keymap
func New(name string) *Keymap {
	return &Keymap{
		bindings: make(map[string]*Binding),
		name:     name,
	}
}

// SetParent sets the parent keymap for inheritance
func (km *Keymap) SetParent(parent *Keymap) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.parent = parent
}

// Bind binds a key sequence to a command
func (km *Keymap) Bind(keySeq KeySequence, command string, args ...interface{}) error {
	km.mu.Lock()
	defer km.mu.Unlock()
	
	if len(keySeq) == 0 {
		return errors.New("key sequence cannot be empty")
	}
	
	if command == "" {
		return errors.New("command cannot be empty")
	}
	
	keyStr := keySeq.String()
	km.bindings[keyStr] = &Binding{
		KeySeq:  keySeq,
		Command: command,
		Args:    args,
	}
	
	return nil
}

// Unbind removes a key binding
func (km *Keymap) Unbind(keySeq KeySequence) error {
	km.mu.Lock()
	defer km.mu.Unlock()
	
	keyStr := keySeq.String()
	if _, exists := km.bindings[keyStr]; !exists {
		return fmt.Errorf("key sequence %s not bound", keyStr)
	}
	
	delete(km.bindings, keyStr)
	return nil
}

// Lookup looks up a key sequence and returns the binding
func (km *Keymap) Lookup(keySeq KeySequence) (*Binding, bool) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	
	keyStr := keySeq.String()
	
	// Check local bindings first
	if binding, exists := km.bindings[keyStr]; exists {
		return binding, true
	}
	
	// Check parent keymap
	if km.parent != nil {
		return km.parent.Lookup(keySeq)
	}
	
	return nil, false
}

// GetAllBindings returns all bindings in this keymap (including inherited)
func (km *Keymap) GetAllBindings() map[string]*Binding {
	km.mu.RLock()
	defer km.mu.RUnlock()
	
	result := make(map[string]*Binding)
	
	// Add parent bindings first
	if km.parent != nil {
		parentBindings := km.parent.GetAllBindings()
		for k, v := range parentBindings {
			result[k] = v
		}
	}
	
	// Add local bindings (override parent)
	for k, v := range km.bindings {
		result[k] = v
	}
	
	return result
}

// GetLocalBindings returns only the local bindings (not inherited)
func (km *Keymap) GetLocalBindings() map[string]*Binding {
	km.mu.RLock()
	defer km.mu.RUnlock()
	
	result := make(map[string]*Binding)
	for k, v := range km.bindings {
		result[k] = v
	}
	
	return result
}

// Name returns the keymap name
func (km *Keymap) Name() string {
	return km.name
}

// Helper functions for creating common keys

// NewKey creates a simple character key
func NewKey(char rune) Key {
	return Key{Char: char}
}

// NewCtrlKey creates a Ctrl+key combination
func NewCtrlKey(char rune) Key {
	return Key{Char: char, Ctrl: true}
}

// NewAltKey creates an Alt+key combination  
func NewAltKey(char rune) Key {
	return Key{Char: char, Alt: true}
}

// NewSpecialKey creates a special key (like return, tab, etc.)
func NewSpecialKey(special string) Key {
	return Key{Special: special}
}

// NewCtrlSpecialKey creates a Ctrl+special key combination
func NewCtrlSpecialKey(special string) Key {
	return Key{Special: special, Ctrl: true}
}

// ParseKeySequence parses a string representation of a key sequence
// Examples: "C-x C-f", "M-x", "return", "C-return"
func ParseKeySequence(keyStr string) (KeySequence, error) {
	if keyStr == "" {
		return nil, errors.New("empty key string")
	}
	
	parts := strings.Fields(keyStr)
	var seq KeySequence
	
	for _, part := range parts {
		key, err := parseKey(part)
		if err != nil {
			return nil, fmt.Errorf("invalid key %s: %v", part, err)
		}
		seq = append(seq, key)
	}
	
	return seq, nil
}

// parseKey parses a single key string like "C-x", "M-return", "a"
func parseKey(keyStr string) (Key, error) {
	parts := strings.Split(keyStr, "-")
	if len(parts) == 0 {
		return Key{}, errors.New("empty key")
	}
	
	var key Key
	
	// Parse modifiers
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "C":
			key.Ctrl = true
		case "M":
			key.Alt = true
		case "S":
			key.Shift = true
		default:
			return Key{}, fmt.Errorf("unknown modifier: %s", parts[i])
		}
	}
	
	// Parse the main key
	mainKey := parts[len(parts)-1]
	if mainKey == "" {
		return Key{}, errors.New("missing key after modifier")
	}
	
	if len(mainKey) == 1 {
		key.Char = rune(mainKey[0])
	} else {
		key.Special = mainKey
	}
	
	return key, nil
}