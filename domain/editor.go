package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/events"
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

type Editor struct {
	buffers         []*Buffer
	windows         []*Window
	currentWin      int
	eventQueue      *events.EventQueue
	running         bool
	minibuffer      *Minibuffer
	commandRegistry *CommandRegistry
	keyBindings     *KeyBindingMap
	metaPressed     bool
	ctrlXPressed    bool
}

func NewEditor() *Editor {
	buffer := NewBuffer("*scratch*")
	window := NewWindow(buffer, 80, 22) // 24-2 for mode line and minibuffer
	
	editor := &Editor{
		buffers:         []*Buffer{buffer},
		windows:         []*Window{window},
		currentWin:      0,
		eventQueue:      events.NewEventQueue(100),
		running:         true,
		minibuffer:      NewMinibuffer(),
		commandRegistry: NewCommandRegistry(),
		keyBindings:     NewKeyBindingMap(),
		metaPressed:     false,
		ctrlXPressed:    false,
	}
	
	// Register cursor movement commands as interactive functions
	editor.registerCursorCommands()
	
	// Register scrolling commands as interactive functions
	editor.registerScrollCommands()
	
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
	if e.currentWin >= 0 && e.currentWin < len(e.windows) {
		return e.windows[e.currentWin]
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
	
	// Handle C-x prefix key sequences
	if e.ctrlXPressed {
		if event.Ctrl && event.Key == "c" {
			// C-x C-c: quit
			log.Info("C-x C-c pressed, shutting down")
			e.running = false
			e.ctrlXPressed = false
			return
		}
		// Reset C-x state for other keys
		e.ctrlXPressed = false
		// For now, just log unhandled C-x sequences
		log.Info("Unhandled C-x sequence: C-x %s", event.Key)
		return
	}
	
	// Handle C-x prefix key
	if event.Ctrl && event.Key == "x" {
		e.ctrlXPressed = true
		log.Info("C-x prefix pressed")
		return
	}
	
	// Handle minibuffer input if active
	if e.minibuffer.IsActive() {
		handled := e.handleMinibufferInput(event)
		if handled {
			return
		}
		// If not handled, continue processing as normal input
	}
	
	// Handle Meta key press detection
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.metaPressed = true
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
	
	// Check for key bindings first
	if cmd, found := e.keyBindings.Lookup(event.Key, event.Ctrl, event.Meta); found {
		err := cmd(e)
		if err != nil {
			log.Error("Key binding command failed: %v", err)
		}
		return
	}
	
	// Check for key sequence bindings (like arrow keys)
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
	window := e.CurrentWindow()
	if window != nil {
		// Reserve 2 lines for mode line and minibuffer
		contentHeight := event.Height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}
		window.Resize(event.Width, contentHeight)
	} else {
		log.Warn("No current window for resize event")
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

func (e *Editor) handleMinibufferInput(event events.KeyEventData) bool {
	switch e.minibuffer.Mode() {
	case MinibufferCommand:
		e.handleCommandInput(event)
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