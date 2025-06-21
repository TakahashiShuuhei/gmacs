package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/events"
	"github.com/TakahashiShuuhei/gmacs/core/log"
	luaconfig "github.com/TakahashiShuuhei/gmacs/core/lua-config"
)

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
	configLoader    *luaconfig.ConfigLoader
	hookManager     *luaconfig.HookManager
	options         map[string]interface{}
}

// EditorConfig holds configuration options for editor initialization
type EditorConfig struct {
	ConfigPath string // Path to Lua config file (empty means no config)
}

// NewEditor creates a new editor without loading any configuration file
func NewEditor() *Editor {
	return newEditorWithConfig(EditorConfig{})
}

// NewEditorWithConfig creates a new editor and loads the specified configuration file
func NewEditorWithConfig(configPath string) *Editor {
	return newEditorWithConfig(EditorConfig{
		ConfigPath: configPath,
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
		configLoader:    nil, // Will be initialized if config is loaded
		hookManager:     luaconfig.NewHookManager(),
		options:         make(map[string]interface{}),
	}
	
	// Register all built-in commands
	editor.registerBuiltinCommands()
	
	// Load configuration if specified
	if config.ConfigPath != "" {
		err := editor.loadConfig(config.ConfigPath)
		if err != nil {
			log.Error("Failed to load config from %s: %v", config.ConfigPath, err)
			// Continue without config rather than failing
		} else {
			log.Info("Successfully applied config from: %s", config.ConfigPath)
		}
	}
	
	// Initialize buffer with fundamental mode
	err := editor.modeManager.SetMajorMode(buffer, "fundamental-mode")
	if err != nil {
		log.Error("Failed to set fundamental mode: %v", err)
	}
	
	log.Info("Editor created with buffer: %s", buffer.Name())
	return editor
}

// registerBuiltinCommands registers all built-in editor commands
func (e *Editor) registerBuiltinCommands() {
	e.registerCursorCommands()
	e.registerScrollCommands()
	e.registerBufferCommands()
	e.registerWindowCommands()
	e.registerMinorModeCommands()
}

// loadConfig loads a Lua configuration file
func (e *Editor) loadConfig(configPath string) error {
	log.Info("Starting to load config from: %s", configPath)
	
	// Initialize config loader
	e.configLoader = luaconfig.NewConfigLoader()
	
	// Register API bindings first
	apiBindings := luaconfig.NewAPIBindings(e, e.configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		e.configLoader.Close()
		e.configLoader = nil
		return err
	}
	log.Info("Registered gmacs API successfully")
	
	// Load the configuration file
	err = e.configLoader.LoadConfig(configPath)
	if err != nil {
		e.configLoader.Close()
		e.configLoader = nil
		return err
	}
	
	log.Info("Successfully loaded and executed Lua configuration")
	return nil
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
	// TODO: Implement mode-specific key bindings
	// For now, just use global bindings
	return e.BindKey(sequence, command)
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
	log.Info("Set option %s = %v", name, value)
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
	log.Info("Major mode registration not yet implemented: %s", name)
	return nil
}

// RegisterMinorMode implements minor mode registration
func (e *Editor) RegisterMinorMode(name string, config map[string]interface{}) error {
	// TODO: Implement dynamic minor mode registration
	log.Info("Minor mode registration not yet implemented: %s", name)
	return nil
}

// AddHook implements hook registration
func (e *Editor) AddHook(event string, fn func(...interface{}) error) error {
	e.hookManager.AddHook(event, fn)
	return nil
}

// TriggerHook triggers hooks for an event
func (e *Editor) TriggerHook(event string, args ...interface{}) {
	e.hookManager.TriggerHook(event, args...)
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
		log.Info("Quit event received")
		e.running = false
	default:
		log.Warn("Unknown event type: %T", event)
	}
}

func (e *Editor) handleKeyEvent(event events.KeyEventData) {
	
	// Always process key sequences first to handle multi-key sequences correctly
	cmd, matched, continuing := e.keyBindings.ProcessKeyPress(event.Key, event.Ctrl, event.Meta)
	
	// If we have a continuing sequence, always handle it first
	if continuing {
		log.Info("Key sequence in progress")
		return
	}
	
	// If we have a matched command, check if it should be handled or deferred to minibuffer
	if matched {
		// Special commands that should always execute immediately
		if event.Ctrl && event.Key == "g" { // C-g (KeyboardQuit)
			log.Info("Keyboard quit command, executing immediately")
			err := cmd(e)
			if err != nil {
				log.Error("Keyboard quit command failed: %v", err)
			}
			return
		}
		
		// Check if this is likely a multi-key sequence command (like C-x C-c)
		// by seeing if the command is different from basic editing commands
		// For now, we'll use a simple heuristic: if minibuffer is active and 
		// this is a single Ctrl+key that could be editing, defer to minibuffer
		if e.minibuffer.IsActive() && e.isSingleKeyEditCommand(event) {
			// Try minibuffer input handling first
			handled := e.handleMinibufferInput(event)
			if handled {
				return
			}
		}
		
		// Execute the matched command
		log.Info("Key sequence matched, executing command")
		err := cmd(e)
		if err != nil {
			log.Error("Key sequence command failed: %v", err)
		}
		return
	}
	
	// No key sequence match - handle minibuffer input if active
	if e.minibuffer.IsActive() {
		handled := e.handleMinibufferInput(event)
		if handled {
			return
		}
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
		err := cmd(e)
		if err != nil {
			log.Error("Key sequence command failed: %v", err)
		}
		return
	}
	
	// Regular text input
	buffer := e.CurrentBuffer()
	if buffer == nil {
		log.Warn("No current buffer for key event")
		return
	}
	
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		if event.Key == "Enter" || event.Key == "Return" {
			buffer.InsertChar('\n')
			log.Info("SCROLL_TIMING: Text inserted (newline), calling EnsureCursorVisible at cursor (%d,%d)", buffer.Cursor().Row, buffer.Cursor().Col)
			
			// Check for auto-a-mode and add 'a' if enabled
			e.processMinorModeHooks(buffer, "newline")
			
			EnsureCursorVisible(e)
		} else {
			buffer.InsertChar(event.Rune)
			log.Info("SCROLL_TIMING: Text inserted (char %c), calling EnsureCursorVisible at cursor (%d,%d)", event.Rune, buffer.Cursor().Row, buffer.Cursor().Col)
			EnsureCursorVisible(e)
		}
	}
}

func (e *Editor) handleResizeEvent(event events.ResizeEventData) {
	log.Info("Window resize: %dx%d", event.Width, event.Height)
	if e.layout != nil {
		e.layout.Resize(event.Width, event.Height)
	} else {
		log.Warn("No layout for resize event")
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

// isSingleKeyEditCommand checks if this is a single-key editing command that should be handled by minibuffer
func (e *Editor) isSingleKeyEditCommand(event events.KeyEventData) bool {
	if !event.Ctrl {
		return false
	}
	
	// These are single-key editing commands that should be handled by minibuffer when active
	editKeys := []string{"h", "d", "f", "b", "a", "e"}
	for _, key := range editKeys {
		if event.Key == key {
			return true
		}
	}
	return false
}


func (e *Editor) handleMinibufferInput(event events.KeyEventData) bool {
	switch e.minibuffer.Mode() {
	case MinibufferCommand:
		e.handleCommandInput(event)
		return true
	case MinibufferFile:
		e.handleFileInput(event)
		return true
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
	e.commandRegistry.RegisterFunc("scroll-left", ScrollLeftChar)
	e.commandRegistry.RegisterFunc("scroll-right", ScrollRightChar)
	e.commandRegistry.RegisterFunc("toggle-truncate-lines", ToggleLineWrap)
	e.commandRegistry.RegisterFunc("page-up", PageUp)
	e.commandRegistry.RegisterFunc("page-down", PageDown)
	e.commandRegistry.RegisterFunc("debug-info", ShowDebugInfo)
}

func (e *Editor) registerBufferCommands() {
	// Register buffer commands as M-x interactive functions
	e.commandRegistry.RegisterFunc("switch-to-buffer", SwitchToBufferInteractive)
	e.commandRegistry.RegisterFunc("list-buffers", ListBuffersInteractive)
	e.commandRegistry.RegisterFunc("kill-buffer", KillBufferInteractive)
	
	// Set up keybindings for buffer functions
	e.keyBindings.BindKeySequence("C-x b", SwitchToBufferInteractive)
	e.keyBindings.BindKeySequence("C-x C-b", ListBuffersInteractive)
	e.keyBindings.BindKeySequence("C-x k", KillBufferInteractive)
}

func (e *Editor) registerWindowCommands() {
	// Register window commands as M-x interactive functions
	e.commandRegistry.RegisterFunc("split-window-right", SplitWindowRight)
	e.commandRegistry.RegisterFunc("split-window-below", SplitWindowBelow)
	e.commandRegistry.RegisterFunc("other-window", OtherWindow)
	e.commandRegistry.RegisterFunc("delete-window", DeleteWindow)
	e.commandRegistry.RegisterFunc("delete-other-windows", DeleteOtherWindows)
	
	// Set up keybindings for window functions
	e.keyBindings.BindKeySequence("C-x 3", SplitWindowRight)
	e.keyBindings.BindKeySequence("C-x 2", SplitWindowBelow)
	e.keyBindings.BindKeySequence("C-x o", OtherWindow)
	e.keyBindings.BindKeySequence("C-x 0", DeleteWindow)
	e.keyBindings.BindKeySequence("C-x 1", DeleteOtherWindows)
}

func (e *Editor) handleCommandInput(event events.KeyEventData) {
	// Handle Enter - execute command
	if event.Key == "Enter" || event.Key == "Return" {
		commandName := e.minibuffer.Content()
		log.Info("Executing command: %s", commandName)
		
		if cmd, exists := e.commandRegistry.Get(commandName); exists {
			// Clear command input first
			e.minibuffer.Clear()
			
			// Execute command (command can set its own message)
			err := cmd.Execute(e)
			if err != nil {
				log.Error("Command execution failed: %v", err)
				e.minibuffer.SetMessage("Command failed: " + err.Error())
			}
		} else {
			log.Warn("Unknown command: %s", commandName)
			e.minibuffer.SetMessage("Unknown command: " + commandName)
		}
		return
	}
	
	// Handle Escape - cancel command
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.minibuffer.Clear()
		return
	}
	
	// Handle Ctrl commands in minibuffer
	if event.Ctrl {
		switch event.Key {
		case "h":
			e.minibuffer.DeleteBackward()
			return
		case "d":
			e.minibuffer.DeleteForward()
			return
		case "f":
			e.minibuffer.MoveCursorForward()
			return
		case "b":
			e.minibuffer.MoveCursorBackward()
			return
		case "a":
			e.minibuffer.MoveCursorToBeginning()
			return
		case "e":
			e.minibuffer.MoveCursorToEnd()
			return
		}
	}
	
	// Handle Backspace
	if event.Key == "Backspace" || event.Key == "\x7f" {
		e.minibuffer.DeleteBackward()
		return
	}
	
	// Handle normal character input
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		e.minibuffer.InsertChar(event.Rune)
	}
}

// AddBuffer adds a new buffer to the editor
func (e *Editor) AddBuffer(buffer *Buffer) {
	e.buffers = append(e.buffers, buffer)
	
	// Auto-detect and set major mode for new buffer
	if buffer.MajorMode() == nil {
		mode, err := e.modeManager.AutoDetectMajorMode(buffer)
		if err != nil {
			log.Error("Failed to auto-detect major mode: %v", err)
		} else {
			err = e.modeManager.SetMajorMode(buffer, mode.Name())
			if err != nil {
				log.Error("Failed to set major mode: %v", err)
			}
		}
	}
}

// SwitchToBuffer switches the current window to the specified buffer
func (e *Editor) SwitchToBuffer(buffer *Buffer) {
	window := e.CurrentWindow()
	if window != nil {
		window.SetBuffer(buffer)
		log.Info("Switched to buffer: %s", buffer.Name())
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

// handleFileInput handles file path input for C-x C-f
func (e *Editor) handleFileInput(event events.KeyEventData) {
	// Handle Enter - open file
	if event.Key == "Enter" || event.Key == "Return" {
		filepath := e.minibuffer.Content()
		log.Info("Opening file: %s", filepath)
		
		// Try to load the file
		buffer, err := NewBufferFromFile(filepath)
		if err != nil {
			log.Error("Failed to open file %s: %v", filepath, err)
			e.minibuffer.SetMessage("Cannot open file: " + filepath)
		} else {
			// Add buffer to editor and switch to it
			e.AddBuffer(buffer)
			e.SwitchToBuffer(buffer)
			e.minibuffer.SetMessage("Opened: " + filepath)
		}
		return
	}
	
	// Handle Escape - cancel file input
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.minibuffer.Clear()
		return
	}
	
	// Handle Ctrl commands in minibuffer
	if event.Ctrl {
		switch event.Key {
		case "h":
			e.minibuffer.DeleteBackward()
			return
		case "d":
			e.minibuffer.DeleteForward()
			return
		case "f":
			e.minibuffer.MoveCursorForward()
			return
		case "b":
			e.minibuffer.MoveCursorBackward()
			return
		case "a":
			e.minibuffer.MoveCursorToBeginning()
			return
		case "e":
			e.minibuffer.MoveCursorToEnd()
			return
		}
	}
	
	// Handle Backspace
	if event.Key == "Backspace" || event.Key == "\x7f" {
		e.minibuffer.DeleteBackward()
		return
	}
	
	// Handle normal character input
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		e.minibuffer.InsertChar(event.Rune)
	}
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