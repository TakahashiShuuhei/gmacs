package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/events"
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

type Editor struct {
	buffers     []*Buffer
	windows     []*Window
	currentWin  int
	eventQueue  *events.EventQueue
	running     bool
}

func NewEditor() *Editor {
	log.Debug("Creating new editor")
	buffer := NewBuffer("*scratch*")
	window := NewWindow(buffer, 80, 24)
	
	editor := &Editor{
		buffers:    []*Buffer{buffer},
		windows:    []*Window{window},
		currentWin: 0,
		eventQueue: events.NewEventQueue(100),
		running:    true,
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
	buffer := e.CurrentBuffer()
	if buffer == nil {
		log.Warn("No current buffer for key event")
		return
	}
	
	log.Debug("Key event: key=%s, rune=%c, ctrl=%t, meta=%t", event.Key, event.Rune, event.Ctrl, event.Meta)
	
	if event.Ctrl && event.Key == "c" {
		log.Info("Ctrl+C pressed, shutting down")
		e.running = false
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