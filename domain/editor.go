package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/events"
	"github.com/TakahashiShuuhei/gmacs/core/log"
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
}

func NewEditor() *Editor {
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
	}
	
	// Register cursor movement commands as interactive functions
	editor.registerCursorCommands()
	
	// Register scrolling commands as interactive functions
	editor.registerScrollCommands()
	
	// Register buffer commands as interactive functions
	editor.registerBufferCommands()
	
	// Register window commands as interactive functions
	editor.registerWindowCommands()
	
	log.Info("Editor created with buffer: %s", buffer.Name())
	return editor
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