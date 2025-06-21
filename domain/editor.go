package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

// ConfigLoader interface for Lua configuration loading
type ConfigLoader interface {
	LoadConfig(configPath string) error
	Close()
}

// HookManager interface for event hooks  
type HookManager interface {
	AddHook(event string, fn func(...interface{}) error)
	TriggerHook(event string, args ...interface{})
}

type Editor struct {
	buffers         []*Buffer
	layout          *WindowLayout
	eventQueue      *events.EventQueue
	running         bool
	minibuffer      *Minibuffer
	commandRegistry *CommandRegistry
	keyBindings     *KeyBindingMap
	metaPressed     bool
	modeManager     *ModeManager
	configLoader    ConfigLoader
	hookManager     HookManager
	options         map[string]interface{}
}

// EditorConfig holds configuration options for editor initialization
type EditorConfig struct {
	ConfigLoader ConfigLoader // Optional Lua config loader
	HookManager  HookManager  // Optional hook manager
}

// NewEditor creates a new editor with default configuration
func NewEditor() *Editor {
	editor := newEditorWithConfig(EditorConfig{})
	
	// Register built-in commands directly for backward compatibility
	editor.RegisterBuiltinCommands()
	
	return editor
}

// NewEditorWithConfig creates a new editor with the specified configuration
func NewEditorWithConfig(configLoader ConfigLoader, hookManager HookManager) *Editor {
	return newEditorWithConfig(EditorConfig{
		ConfigLoader: configLoader,
		HookManager:  hookManager,
	})
}

// newEditorWithConfig is the internal constructor that handles configuration
func newEditorWithConfig(config EditorConfig) *Editor {
	buffer := NewBuffer("*scratch*")
	window := NewWindow(buffer, 80, 22) // 24-2 for mode line and minibuffer
	layout := NewWindowLayout(window, 80, 24) // Total terminal size
	
	editor := &Editor{
		buffers:         []*Buffer{buffer},
		layout:          layout,
		eventQueue:      events.NewEventQueue(100),
		running:         true,
		minibuffer:      NewMinibuffer(),
		commandRegistry: NewCommandRegistry(),
		keyBindings:     NewKeyBindingMap(),
		metaPressed:     false,
		modeManager:     NewModeManager(),
		configLoader:    config.ConfigLoader,
		hookManager:     config.HookManager,
		options:         make(map[string]interface{}),
	}
	
	// Built-in commands are now registered via Lua configuration
	
	// Initialize buffer with fundamental mode
	err := editor.modeManager.SetMajorMode(buffer, "fundamental-mode")
	if err != nil {
		// TODO: Handle error appropriately
	}
	
	return editor
}

// RegisterBuiltinCommands registers all built-in editor commands for backward compatibility
func (e *Editor) RegisterBuiltinCommands() {
	e.registerCoreCommands()
	e.registerCursorCommands()
	e.registerScrollCommands()
	e.registerBufferCommands()
	e.registerWindowCommands()
	e.registerMinorModeCommands()
}

func (e *Editor) registerCoreCommands() {
	// Register core system commands
	e.commandRegistry.RegisterFunc("quit", Quit)
	e.commandRegistry.RegisterFunc("keyboard-quit", KeyboardQuit)
	e.commandRegistry.RegisterFunc("find-file", FindFile)
	e.commandRegistry.RegisterFunc("delete-backward-char", DeleteBackwardChar)
	e.commandRegistry.RegisterFunc("delete-char", DeleteChar)
}

func (e *Editor) registerCursorCommands() {
	// Register cursor movement commands as M-x interactive functions
	e.commandRegistry.RegisterFunc("forward-char", ForwardChar)
	e.commandRegistry.RegisterFunc("backward-char", BackwardChar)
	e.commandRegistry.RegisterFunc("next-line", NextLine)
	e.commandRegistry.RegisterFunc("previous-line", PreviousLine)
	e.commandRegistry.RegisterFunc("beginning-of-line", BeginningOfLine)
	e.commandRegistry.RegisterFunc("end-of-line", EndOfLine)
}

func (e *Editor) registerScrollCommands() {
	// Register scrolling commands as M-x interactive functions
	e.commandRegistry.RegisterFunc("scroll-up", ScrollUp)
	e.commandRegistry.RegisterFunc("scroll-down", ScrollDown)
	e.commandRegistry.RegisterFunc("page-up", PageUp)
	e.commandRegistry.RegisterFunc("page-down", PageDown)
	e.commandRegistry.RegisterFunc("toggle-truncate-lines", ToggleLineWrap)
	e.commandRegistry.RegisterFunc("debug-info", ShowDebugInfo)
}

func (e *Editor) registerBufferCommands() {
	// Register buffer commands as M-x interactive functions
	e.commandRegistry.RegisterFunc("switch-to-buffer", SwitchToBufferInteractive)
	e.commandRegistry.RegisterFunc("list-buffers", ListBuffersInteractive)
	e.commandRegistry.RegisterFunc("kill-buffer", KillBufferInteractive)
}

func (e *Editor) registerWindowCommands() {
	// Register window commands as M-x interactive functions
	e.commandRegistry.RegisterFunc("split-window-right", SplitWindowRight)
	e.commandRegistry.RegisterFunc("split-window-below", SplitWindowBelow)
	e.commandRegistry.RegisterFunc("other-window", OtherWindow)
	e.commandRegistry.RegisterFunc("delete-window", DeleteWindow)
	e.commandRegistry.RegisterFunc("delete-other-windows", DeleteOtherWindows)
}

// Cleanup closes any resources when the editor is shutting down
func (e *Editor) Cleanup() {
	if e.configLoader != nil {
		e.configLoader.Close()
		e.configLoader = nil
	}
}

// EditorInterface implementation for Lua API

// BindKey implements global key binding
func (e *Editor) BindKey(sequence, command string) error {
	cmd, exists := e.commandRegistry.Get(command)
	if !exists {
		return &ConfigError{Message: "Unknown command: " + command}
	}
	
	// Convert command to function
	cmdFunc := func(editor *Editor) error {
		return cmd.Execute(editor)
	}
	
	e.keyBindings.BindKeySequence(sequence, cmdFunc)
	return nil
}

// LocalBindKey implements mode-specific key binding
func (e *Editor) LocalBindKey(modeName, sequence, command string) error {
	// Check if command exists
	cmd, exists := e.commandRegistry.Get(command)
	if !exists {
		return &ConfigError{Message: "Unknown command: " + command}
	}
	
	// Convert command to function
	cmdFunc := func(editor *Editor) error {
		return cmd.Execute(editor)
	}
	
	// Try to find major mode first
	if majorMode, exists := e.modeManager.GetMajorModeByName(modeName); exists {
		keyBindings := majorMode.KeyBindings()
		if keyBindings != nil {
			keyBindings.BindKeySequence(sequence, cmdFunc)
			return nil
		}
	}
	
	// Try to find minor mode
	if minorMode, exists := e.modeManager.GetMinorModeByName(modeName); exists {
		keyBindings := minorMode.KeyBindings()
		if keyBindings != nil {
			keyBindings.BindKeySequence(sequence, cmdFunc)
			return nil
		}
	}
	
	return &ConfigError{Message: "Unknown mode: " + modeName}
}

// RegisterCommand implements custom command registration
func (e *Editor) RegisterCommand(name string, fn func() error) error {
	// Convert to CommandFunc
	cmdFunc := func(editor *Editor) error {
		return fn()
	}
	
	e.commandRegistry.RegisterFunc(name, cmdFunc)
	return nil
}

// SetOption implements option setting
func (e *Editor) SetOption(name string, value interface{}) error {
	e.options[name] = value
	return nil
}

// GetOption implements option getting
func (e *Editor) GetOption(name string) (interface{}, error) {
	value, exists := e.options[name]
	if !exists {
		return nil, &ConfigError{Message: "Unknown option: " + name}
	}
	return value, nil
}

// RegisterMajorMode implements major mode registration
func (e *Editor) RegisterMajorMode(name string, config map[string]interface{}) error {
	// TODO: Implement dynamic major mode registration
	return nil
}

// RegisterMinorMode implements minor mode registration
func (e *Editor) RegisterMinorMode(name string, config map[string]interface{}) error {
	// TODO: Implement dynamic minor mode registration
	return nil
}

// AddHook implements hook registration
func (e *Editor) AddHook(event string, fn func(...interface{}) error) error {
	if e.hookManager != nil {
		e.hookManager.AddHook(event, fn)
	}
	return nil
}

// TriggerHook triggers hooks for an event
func (e *Editor) TriggerHook(event string, args ...interface{}) {
	if e.hookManager != nil {
		e.hookManager.TriggerHook(event, args...)
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

func (e *Editor) CurrentBuffer() *Buffer {
	window := e.CurrentWindow()
	if window != nil {
		return window.Buffer()
	}
	return nil
}

func (e *Editor) CurrentWindow() *Window {
	if e.layout != nil {
		return e.layout.CurrentWindow()
	}
	return nil
}

func (e *Editor) Layout() *WindowLayout {
	return e.layout
}

func (e *Editor) EventQueue() *events.EventQueue {
	return e.eventQueue
}

func (e *Editor) HandleEvent(event events.Event) {
	switch ev := event.(type) {
	case events.KeyEventData:
		e.handleKeyEvent(ev)
	case events.ResizeEventData:
		e.handleResizeEvent(ev)
	case events.QuitEventData:
		e.running = false
	default:
		// Unknown event type
	}
}

func (e *Editor) handleKeyEvent(event events.KeyEventData) {
	
	// Always process key sequences first to handle multi-key sequences correctly
	cmd, matched, continuing := e.keyBindings.ProcessKeyPress(event.Key, event.Ctrl, event.Meta)
	
	// If we have a continuing sequence, always handle it first
	if continuing {
		return
	}
	
	// Handle minibuffer input first if active (except for special cases)
	if e.minibuffer.IsActive() {
		// Special commands that should always execute immediately
		if event.Ctrl && event.Key == "g" { // C-g (KeyboardQuit)
			if matched {
				cmd(e)
			}
			return
		}
		
		// Try minibuffer handling first
		handled := e.handleMinibufferInput(event)
		if handled {
			return
		}
	}
	
	// If not handled by minibuffer, check for matched global commands
	if matched {
		cmd(e)
		return
	}
	
	// Handle Meta key press detection
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.metaPressed = true
		// Reset key sequence on Escape
		e.keyBindings.ResetSequence()
		return
	}
	
	// Handle M-x command
	if e.metaPressed && event.Key == "x" {
		e.minibuffer.StartCommandInput()
		e.metaPressed = false
		return
	}
	
	// Reset meta state for other keys
	if e.metaPressed {
		e.metaPressed = false
	}
	
	// Check for any remaining key bindings through the unified system
	// (single keys and raw sequences that weren't caught by sequence processing)
	if cmd, found := e.keyBindings.LookupSequence(event.Key); found {
		cmd(e)
		return
	}
	
	// Regular text input
	buffer := e.CurrentBuffer()
	if buffer == nil {
		return
	}
	
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		if event.Key == "Enter" || event.Key == "Return" {
			buffer.InsertChar('\n')
			
			// Check for auto-a-mode and add 'a' if enabled
			e.processMinorModeHooks(buffer, "newline")
			
			EnsureCursorVisible(e)
		} else {
			buffer.InsertChar(event.Rune)
			EnsureCursorVisible(e)
		}
	}
}

func (e *Editor) handleResizeEvent(event events.ResizeEventData) {
	if e.layout != nil {
		e.layout.Resize(event.Width, event.Height)
	}
}

func (e *Editor) IsRunning() bool {
	return e.running
}

func (e *Editor) Quit() {
	e.running = false
}

func (e *Editor) Minibuffer() *Minibuffer {
	return e.minibuffer
}

func (e *Editor) SetMinibufferMessage(message string) {
	e.minibuffer.SetMessage(message)
}

// GetKeySequenceInProgress returns the current key sequence in progress, if any
func (e *Editor) GetKeySequenceInProgress() string {
	sequence := e.keyBindings.GetCurrentSequence()
	return FormatSequence(sequence)
}



func (e *Editor) handleMinibufferInput(event events.KeyEventData) bool {
	switch e.minibuffer.Mode() {
	case MinibufferCommand:
		return e.handleMinibufferAsBuffer(event, e.executeMinibufferCommand)
	case MinibufferFile:
		return e.handleMinibufferAsBuffer(event, e.executeFileOpen)
	case MinibufferBufferSelection:
		e.HandleBufferSelectionInput(event)
		return true
	case MinibufferMessage:
		// Any key clears the message, but allow the key to continue being processed
		e.minibuffer.Clear()
		return false
	}
	return false
}

// handleMinibufferAsBuffer treats minibuffer like a regular buffer, using unified commands
func (e *Editor) handleMinibufferAsBuffer(event events.KeyEventData, onEnter func()) bool {
	// Handle Enter - execute the completion action
	if event.Key == "Enter" || event.Key == "Return" {
		onEnter()
		return true
	}
	
	// Handle Escape - cancel
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.minibuffer.Clear()
		return true
	}
	
	// Handle Backspace as delete-backward-char
	if event.Key == "Backspace" || event.Key == "\x7f" {
		e.executeCommandOnMinibuffer("delete-backward-char")
		return true
	}
	
	// Handle Ctrl commands using the unified command system
	if event.Ctrl {
		switch event.Key {
		case "h":
			e.executeCommandOnMinibuffer("delete-backward-char")
			return true
		case "d":
			e.executeCommandOnMinibuffer("delete-forward-char")
			return true
		case "f":
			e.executeCommandOnMinibuffer("forward-char")
			return true
		case "b":
			e.executeCommandOnMinibuffer("backward-char")
			return true
		case "a":
			e.executeCommandOnMinibuffer("beginning-of-line")
			return true
		case "e":
			e.executeCommandOnMinibuffer("end-of-line")
			return true
		}
	}
	
	// Handle normal character input
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		e.executeCommandOnMinibuffer("self-insert-command", event.Rune)
		return true
	}
	
	return false
}

// executeCommandOnMinibuffer executes a command in the context of the minibuffer
func (e *Editor) executeCommandOnMinibuffer(commandName string, rune ...rune) {
	// Handle special minibuffer-specific commands first
	switch commandName {
	case "delete-backward-char":
		e.minibuffer.DeleteBackward()
	case "delete-forward-char":
		e.minibuffer.DeleteForward()
	case "forward-char":
		e.minibuffer.MoveCursorForward()
	case "backward-char":
		e.minibuffer.MoveCursorBackward()
	case "beginning-of-line":
		e.minibuffer.MoveCursorToBeginning()
	case "end-of-line":
		e.minibuffer.MoveCursorToEnd()
	case "self-insert-command":
		if len(rune) > 0 {
			e.minibuffer.InsertChar(rune[0])
		}
	default:
		// Try to execute as a regular command if it exists
		if cmd, exists := e.commandRegistry.Get(commandName); exists {
			// Execute command - this allows minibuffer to use same commands as main buffer
			cmd.Execute(e)
		}
	}
}

// executeMinibufferCommand handles M-x command execution
func (e *Editor) executeMinibufferCommand() {
	commandName := e.minibuffer.Content()
	
	if cmd, exists := e.commandRegistry.Get(commandName); exists {
		// Clear command input first
		e.minibuffer.Clear()
		
		// Execute command (command can set its own message)
		err := cmd.Execute(e)
		if err != nil {
			e.minibuffer.SetMessage("Command failed: " + err.Error())
		}
	} else {
		e.minibuffer.SetMessage("Unknown command: " + commandName)
	}
}

// executeFileOpen handles C-x C-f file opening
func (e *Editor) executeFileOpen() {
	filepath := e.minibuffer.Content()
	
	// Try to load the file
	buffer, err := NewBufferFromFile(filepath)
	if err != nil {
		e.minibuffer.SetMessage("Cannot open file: " + filepath)
	} else {
		// Add buffer to editor and switch to it
		e.AddBuffer(buffer)
		e.SwitchToBuffer(buffer)
		e.minibuffer.SetMessage("Opened: " + filepath)
	}
}






// AddBuffer adds a new buffer to the editor
func (e *Editor) AddBuffer(buffer *Buffer) {
	e.buffers = append(e.buffers, buffer)
	
	// Auto-detect and set major mode for new buffer
	if buffer.MajorMode() == nil {
		mode, err := e.modeManager.AutoDetectMajorMode(buffer)
		if err == nil {
			e.modeManager.SetMajorMode(buffer, mode.Name())
		}
	}
}

// SwitchToBuffer switches the current window to the specified buffer
func (e *Editor) SwitchToBuffer(buffer *Buffer) {
	window := e.CurrentWindow()
	if window != nil {
		window.SetBuffer(buffer)
	}
}

// FindBuffer finds a buffer by name
func (e *Editor) FindBuffer(name string) *Buffer {
	for _, buffer := range e.buffers {
		if buffer.Name() == name {
			return buffer
		}
	}
	return nil
}


// ModeManager returns the mode manager
func (e *Editor) ModeManager() *ModeManager {
	return e.modeManager
}

func (e *Editor) registerMinorModeCommands() {
	// Register auto-a-mode command
	e.commandRegistry.RegisterFunc("auto-a-mode", func(editor *Editor) error {
		buffer := editor.CurrentBuffer()
		if buffer == nil {
			return &ModeError{Message: "No current buffer"}
		}
		
		// Check if already enabled
		modeManager := editor.ModeManager()
		autoAMode := modeManager.minorModes["auto-a-mode"]
		if autoAMode == nil {
			return &ModeError{Message: "auto-a-mode not available"}
		}
		
		if autoAMode.IsEnabled(buffer) {
			editor.SetMinibufferMessage("Auto-A mode disabled")
		} else {
			editor.SetMinibufferMessage("Auto-A mode enabled")
		}
		
		return modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	})
}

func (e *Editor) processMinorModeHooks(buffer *Buffer, event string) {
	// Process minor mode hooks for specific events
	minorModes := buffer.MinorModes()
	
	for _, mode := range minorModes {
		if autoAMode, ok := mode.(*AutoAMode); ok && event == "newline" {
			autoAMode.ProcessNewline(buffer)
		}
	}
}