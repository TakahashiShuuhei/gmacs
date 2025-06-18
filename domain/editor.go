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
	metaPressed     bool
}

func NewEditor() *Editor {
	log.Debug("Creating new editor")
	buffer := NewBuffer("*scratch*")
	window := NewWindow(buffer, 80, 24)
	
	editor := &Editor{
		buffers:         []*Buffer{buffer},
		windows:         []*Window{window},
		currentWin:      0,
		eventQueue:      events.NewEventQueue(100),
		running:         true,
		minibuffer:      NewMinibuffer(),
		commandRegistry: NewCommandRegistry(),
		metaPressed:     false,
	}
	
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
	log.Debug("Handling event: %T", event)
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
	log.Debug("Key event: key=%s, rune=%c, ctrl=%t, meta=%t", event.Key, event.Rune, event.Ctrl, event.Meta)
	
	// Handle Ctrl+C for quit
	if event.Ctrl && event.Key == "c" {
		log.Info("Ctrl+C pressed, shutting down")
		e.running = false
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
		log.Debug("Meta key pressed")
		return
	}
	
	// Handle M-x command
	if e.metaPressed && event.Key == "x" {
		log.Debug("M-x pressed, starting command input")
		e.minibuffer.StartCommandInput()
		e.metaPressed = false
		return
	}
	
	// Reset meta state for other keys
	if e.metaPressed {
		e.metaPressed = false
	}
	
	// Regular text input
	buffer := e.CurrentBuffer()
	if buffer == nil {
		log.Warn("No current buffer for key event")
		return
	}
	
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		if event.Key == "Enter" || event.Key == "Return" {
			log.Debug("Inserting newline")
			buffer.InsertChar('\n')
		} else {
			log.Debug("Inserting character: %c", event.Rune)
			buffer.InsertChar(event.Rune)
		}
	}
}

func (e *Editor) handleResizeEvent(event events.ResizeEventData) {
	log.Info("Window resize: %dx%d", event.Width, event.Height)
	window := e.CurrentWindow()
	if window != nil {
		window.Resize(event.Width, event.Height)
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
		log.Debug("Command input cancelled")
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