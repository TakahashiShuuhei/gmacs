package domain

import "github.com/TakahashiShuuhei/gmacs/core/events"

type Editor struct {
	buffers     []*Buffer
	windows     []*Window
	currentWin  int
	eventQueue  *events.EventQueue
	running     bool
}

func NewEditor() *Editor {
	buffer := NewBuffer("*scratch*")
	window := NewWindow(buffer, 80, 24)
	
	return &Editor{
		buffers:    []*Buffer{buffer},
		windows:    []*Window{window},
		currentWin: 0,
		eventQueue: events.NewEventQueue(100),
		running:    true,
	}
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
		e.running = false
	}
}

func (e *Editor) handleKeyEvent(event events.KeyEventData) {
	buffer := e.CurrentBuffer()
	if buffer == nil {
		return
	}
	
	if event.Ctrl && event.Key == "c" {
		e.running = false
		return
	}
	
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		if event.Key == "Enter" || event.Key == "Return" {
			buffer.InsertChar('\n')
		} else {
			buffer.InsertChar(event.Rune)
		}
	}
}

func (e *Editor) handleResizeEvent(event events.ResizeEventData) {
	window := e.CurrentWindow()
	if window != nil {
		window.Resize(event.Width, event.Height)
	}
}

func (e *Editor) IsRunning() bool {
	return e.running
}

func (e *Editor) Quit() {
	e.running = false
}