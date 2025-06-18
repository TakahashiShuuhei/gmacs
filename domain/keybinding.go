package domain


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
	
	// Scrolling
	kbm.Bind("v", true, false, PageDown)        // C-v (page down)
	kbm.BindSequence("\x1b[6~", PageDown)       // Page Down key
	kbm.BindSequence("\x1b[5~", PageUp)         // Page Up key
	
	// Arrow keys (ANSI escape sequences)
	kbm.BindSequence("\x1b[C", ForwardChar)     // Right arrow
	kbm.BindSequence("\x1b[D", BackwardChar)    // Left arrow
	kbm.BindSequence("\x1b[B", NextLine)        // Down arrow
	kbm.BindSequence("\x1b[A", PreviousLine)    // Up arrow
	
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