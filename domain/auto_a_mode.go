package domain

// AutoAMode is a simple minor mode that automatically adds 'a' after each newline
type AutoAMode struct {
	name     string
	priority int
	enabled  map[*Buffer]bool
}

// NewAutoAMode creates a new AutoAMode instance
func NewAutoAMode() *AutoAMode {
	return &AutoAMode{
		name:     "auto-a-mode",
		priority: 10, // Medium priority
		enabled:  make(map[*Buffer]bool),
	}
}

// Name returns the mode name
func (am *AutoAMode) Name() string {
	return am.name
}

// KeyBindings returns nil (no special key bindings)
func (am *AutoAMode) KeyBindings() *KeyBindingMap {
	return nil
}

// Commands returns mode-specific commands
func (am *AutoAMode) Commands() map[string]*Command {
	commands := make(map[string]*Command)
	
	// Add toggle command
	commands["auto-a-mode"] = NewCommand("auto-a-mode", func(editor *Editor) error {
		buffer := editor.CurrentBuffer()
		if buffer == nil {
			return &ModeError{Message: "No current buffer"}
		}
		
		return editor.ModeManager().ToggleMinorMode(buffer, "auto-a-mode")
	})
	
	return commands
}

// Enable enables the minor mode for a buffer
func (am *AutoAMode) Enable(buffer *Buffer) error {
	am.enabled[buffer] = true
	buffer.EnableMinorMode(am)
	
	// Hook into the buffer's newline insertion
	am.setupHooks(buffer)
	
	return nil
}

// Disable disables the minor mode for a buffer
func (am *AutoAMode) Disable(buffer *Buffer) error {
	am.enabled[buffer] = false
	buffer.DisableMinorMode(am.name)
	
	// Remove hooks (in a real implementation, we'd need proper hook management)
	
	return nil
}

// IsEnabled checks if the mode is enabled for a buffer
func (am *AutoAMode) IsEnabled(buffer *Buffer) bool {
	return am.enabled[buffer]
}

// Priority returns the mode priority
func (am *AutoAMode) Priority() int {
	return am.priority
}

// setupHooks sets up the auto-'a' insertion logic
func (am *AutoAMode) setupHooks(buffer *Buffer) {
	// In a real implementation, we'd use a proper hook system
	// For now, we'll modify the buffer's InsertChar method via a wrapper
	// This is a simplified demonstration
}

// ProcessNewline processes newline insertion and adds 'a' if mode is enabled
func (am *AutoAMode) ProcessNewline(buffer *Buffer) {
	if !am.IsEnabled(buffer) {
		return
	}
	
	// Add 'a' after the newline
	buffer.InsertChar('a')
}