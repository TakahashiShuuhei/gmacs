package domain

import (
	"unicode/utf8"
)

type Window struct {
	buffer      *Buffer
	width       int
	height      int
	scrollTop   int
	cursorRow   int
	cursorCol   int
}

func NewWindow(buffer *Buffer, width, height int) *Window {
	return &Window{
		buffer:    buffer,
		width:     width,
		height:    height,
		scrollTop: 0,
	}
}

func (w *Window) Buffer() *Buffer {
	return w.buffer
}

func (w *Window) SetBuffer(buffer *Buffer) {
	w.buffer = buffer
}

func (w *Window) Resize(width, height int) {
	w.width = width
	w.height = height
}

func (w *Window) Size() (int, int) {
	return w.width, w.height
}

func (w *Window) ScrollTop() int {
	return w.scrollTop
}

func (w *Window) SetScrollTop(top int) {
	if top < 0 {
		top = 0
	}
	maxScroll := len(w.buffer.content) - w.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if top > maxScroll {
		top = maxScroll
	}
	w.scrollTop = top
}

func (w *Window) VisibleLines() []string {
	content := w.buffer.Content()
	start := w.scrollTop
	end := start + w.height
	
	if start >= len(content) {
		return []string{}
	}
	if end > len(content) {
		end = len(content)
	}
	
	return content[start:end]
}

func (w *Window) CursorPosition() (int, int) {
	bufferPos := w.buffer.Cursor()
	screenRow := bufferPos.Row - w.scrollTop
	
	// Convert byte position to rune position for display
	if bufferPos.Row < len(w.buffer.content) {
		line := w.buffer.content[bufferPos.Row]
		if bufferPos.Col <= len(line) {
			// Count runes up to cursor position
			runeCol := utf8.RuneCountInString(line[:bufferPos.Col])
			return screenRow, runeCol
		}
	}
	
	return screenRow, bufferPos.Col
}