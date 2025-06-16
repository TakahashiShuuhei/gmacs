package window

import (
	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
	"github.com/TakahashiShuuhei/gmacs/internal/cursor"
)

// Window represents a view into a buffer (similar to Emacs window)
type Window struct {
	buffer     *buffer.Buffer
	cursor     *cursor.Cursor
	topLine    int // first visible line (0-indexed)
	height     int // window height in lines
	width      int // window width in characters
	leftMargin int // left margin for line numbers, etc.
}

// New creates a new window displaying the given buffer
func New(buf *buffer.Buffer, height, width int) *Window {
	return &Window{
		buffer:     buf,
		cursor:     cursor.New(),
		topLine:    0,
		height:     height,
		width:      width,
		leftMargin: 0,
	}
}

// Buffer returns the buffer displayed in this window
func (w *Window) Buffer() *buffer.Buffer {
	return w.buffer
}

// SetBuffer sets the buffer to display in this window
func (w *Window) SetBuffer(buf *buffer.Buffer) {
	w.buffer = buf
	// Reset cursor position when switching buffers
	w.cursor.SetPoint(0, 0)
	w.topLine = 0
}

// Cursor returns the cursor for this window
func (w *Window) Cursor() *cursor.Cursor {
	return w.cursor
}

// TopLine returns the first visible line
func (w *Window) TopLine() int {
	return w.topLine
}

// SetTopLine sets the first visible line
func (w *Window) SetTopLine(line int) {
	if line < 0 {
		line = 0
	}
	maxTop := w.buffer.LineCount() - w.height
	if maxTop < 0 {
		maxTop = 0
	}
	if line > maxTop {
		line = maxTop
	}
	w.topLine = line
}

// Height returns the window height
func (w *Window) Height() int {
	return w.height
}

// Width returns the window width
func (w *Window) Width() int {
	return w.width
}

// SetSize sets the window size
func (w *Window) SetSize(height, width int) {
	if height < 1 {
		height = 1
	}
	if width < 1 {
		width = 1
	}
	w.height = height
	w.width = width
}

// LeftMargin returns the left margin width
func (w *Window) LeftMargin() int {
	return w.leftMargin
}

// SetLeftMargin sets the left margin width
func (w *Window) SetLeftMargin(margin int) {
	if margin < 0 {
		margin = 0
	}
	w.leftMargin = margin
}

// VisibleLines returns the range of visible line numbers
func (w *Window) VisibleLines() (start, end int) {
	start = w.topLine
	end = w.topLine + w.height - 1
	if end >= w.buffer.LineCount() {
		end = w.buffer.LineCount() - 1
	}
	return start, end
}

// IsLineVisible returns whether the given line is visible in the window
func (w *Window) IsLineVisible(line int) bool {
	start, end := w.VisibleLines()
	return line >= start && line <= end
}

// EnsureCursorVisible scrolls the window to ensure the cursor is visible
func (w *Window) EnsureCursorVisible() {
	cursorLine := w.cursor.Line()
	
	// If cursor is above the visible area, scroll up
	if cursorLine < w.topLine {
		w.topLine = cursorLine
		return
	}
	
	// If cursor is below the visible area, scroll down
	if cursorLine >= w.topLine+w.height {
		w.topLine = cursorLine - w.height + 1
		if w.topLine < 0 {
			w.topLine = 0
		}
	}
}

// ScrollUp scrolls the window up by the specified number of lines
func (w *Window) ScrollUp(lines int) {
	w.topLine -= lines
	if w.topLine < 0 {
		w.topLine = 0
	}
}

// ScrollDown scrolls the window down by the specified number of lines
func (w *Window) ScrollDown(lines int) {
	w.topLine += lines
	maxTop := w.buffer.LineCount() - w.height
	if maxTop < 0 {
		maxTop = 0
	}
	if w.topLine > maxTop {
		w.topLine = maxTop
	}
}

// GetVisibleText returns the text visible in the window with line numbers
func (w *Window) GetVisibleText() []string {
	start, end := w.VisibleLines()
	lines := make([]string, 0, end-start+1)
	
	for i := start; i <= end; i++ {
		line := w.buffer.GetLine(i)
		lines = append(lines, line)
	}
	
	return lines
}

// CursorScreenPosition returns the cursor position relative to the window
func (w *Window) CursorScreenPosition() (screenLine, screenCol int) {
	screenLine = w.cursor.Line() - w.topLine
	
	// Calculate the actual display width up to the cursor position
	line := w.buffer.GetLine(w.cursor.Line())
	runes := []rune(line)
	cursorPos := w.cursor.Col()
	
	// Calculate display width up to cursor position
	displayWidth := 0
	for i := 0; i < cursorPos && i < len(runes); i++ {
		char := runes[i]
		if isFullWidth(char) {
			displayWidth += 2 // Full-width characters take 2 columns
		} else {
			displayWidth += 1 // Half-width characters take 1 column
		}
	}
	
	screenCol = displayWidth + w.leftMargin
	return screenLine, screenCol
}

// isFullWidth checks if a character is full-width (typically CJK characters)
func isFullWidth(r rune) bool {
	// Simplified check for common full-width ranges
	// This covers most CJK characters, but may not be perfect
	return (r >= 0x1100 && r <= 0x11FF) || // Hangul Jamo
		   (r >= 0x2E80 && r <= 0x2EFF) || // CJK Radicals Supplement
		   (r >= 0x2F00 && r <= 0x2FDF) || // Kangxi Radicals
		   (r >= 0x3000 && r <= 0x303F) || // CJK Symbols and Punctuation
		   (r >= 0x3040 && r <= 0x309F) || // Hiragana
		   (r >= 0x30A0 && r <= 0x30FF) || // Katakana
		   (r >= 0x3100 && r <= 0x312F) || // Bopomofo
		   (r >= 0x3200 && r <= 0x32FF) || // Enclosed CJK Letters and Months
		   (r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
		   (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		   (r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		   (r >= 0xFF00 && r <= 0xFFEF)    // Halfwidth and Fullwidth Forms
}