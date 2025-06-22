package domain

import (
	"github.com/TakahashiShuuhei/gmacs/log"
	"github.com/TakahashiShuuhei/gmacs/util"
)

type Window struct {
	buffer      *Buffer
	width       int
	height      int
	scrollTop   int
	scrollLeft  int
	lineWrap    bool
	cursorRow   int
	cursorCol   int
}

func NewWindow(buffer *Buffer, width, height int) *Window {
	return &Window{
		buffer:     buffer,
		width:      width,
		height:     height,
		scrollTop:  0,
		scrollLeft: 0,
		lineWrap:   true, // Default to line wrapping enabled
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
	
	// Calculate max scroll based on window content mode
	var maxScroll int
	if w.lineWrap {
		// In line wrap mode, we need to consider how many screen lines the content takes
		// For bounds checking, we use a simpler approach: can't scroll past the last buffer line
		maxScroll = len(w.buffer.content) - 1
		if maxScroll < 0 {
			maxScroll = 0
		}
	} else {
		// In no-wrap mode, use the traditional calculation
		maxScroll = len(w.buffer.content) - w.height
		if maxScroll < 0 {
			maxScroll = 0
		}
	}
	
	if top > maxScroll {
		top = maxScroll
	}
	w.scrollTop = top
}

func (w *Window) ScrollLeft() int {
	return w.scrollLeft
}

func (w *Window) SetScrollLeft(left int) {
	if left < 0 {
		left = 0
	}
	w.scrollLeft = left
}

func (w *Window) LineWrap() bool {
	return w.lineWrap
}

func (w *Window) SetLineWrap(wrap bool) {
	w.lineWrap = wrap
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
	
	lines := content[start:end]
	result := make([]string, 0, len(lines))
	
	for _, line := range lines {
		if w.lineWrap {
			// Line wrapping: split long lines into multiple display lines
			wrappedLines := w.wrapLine(line)
			result = append(result, wrappedLines...)
		} else {
			// No wrapping: apply horizontal scrolling
			scrolledLine := w.applyHorizontalScroll(line)
			result = append(result, scrolledLine)
		}
	}
	
	// Ensure we don't exceed the window height
	if len(result) > w.height {
		result = result[:w.height]
	}
	
	return result
}

func (w *Window) wrapLine(line string) []string {
	if util.StringWidth(line) <= w.width {
		return []string{line}
	}
	
	var result []string
	runes := []rune(line)
	start := 0
	
	for start < len(runes) {
		width := 0
		end := start
		
		// Find how many characters fit in one line
		for end < len(runes) {
			charWidth := util.RuneWidth(runes[end])
			if width+charWidth > w.width {
				break
			}
			width += charWidth
			end++
		}
		
		// If we couldn't fit even one character, take at least one
		if end == start && start < len(runes) {
			end = start + 1
		}
		
		result = append(result, string(runes[start:end]))
		start = end
	}
	
	return result
}

func (w *Window) applyHorizontalScroll(line string) string {
	lineWidth := util.StringWidth(line)
	
	if w.scrollLeft == 0 {
		// No horizontal scrolling
		if lineWidth <= w.width {
			return line
		}
		// Line continues beyond window width - show continuation indicator
		truncated := w.truncateToWidth(line, w.width-1)
		return truncated + "\\"
	}
	
	// Apply horizontal scrolling
	runes := []rune(line)
	width := 0
	start := 0
	
	// Skip characters until we reach scrollLeft position
	for start < len(runes) && width < w.scrollLeft {
		width += util.RuneWidth(runes[start])
		start++
	}
	
	// Now get characters that fit in the window width
	if start >= len(runes) {
		return ""
	}
	
	end := start
	displayWidth := 0
	availableWidth := w.width
	
	// If there's content before scroll position, show left indicator
	showLeftIndicator := w.scrollLeft > 0
	if showLeftIndicator {
		availableWidth-- // Reserve space for left continuation indicator
	}
	
	// Check if there's content after what we can display
	hasContentAfter := false
	tempEnd := start
	tempWidth := 0
	for tempEnd < len(runes) {
		charWidth := util.RuneWidth(runes[tempEnd])
		if tempWidth+charWidth > availableWidth {
			hasContentAfter = true
			break
		}
		tempWidth += charWidth
		tempEnd++
	}
	
	// If there's content after, reserve space for right indicator
	if hasContentAfter {
		availableWidth--
	}
	
	// Now build the actual display line
	for end < len(runes) && displayWidth < availableWidth {
		charWidth := util.RuneWidth(runes[end])
		if displayWidth+charWidth > availableWidth {
			break
		}
		displayWidth += charWidth
		end++
	}
	
	result := string(runes[start:end])
	
	// Add continuation indicators
	if showLeftIndicator {
		result = "\\" + result
	}
	if hasContentAfter {
		result = result + "\\"
	}
	
	return result
}

func (w *Window) truncateToWidth(s string, maxWidth int) string {
	if util.StringWidth(s) <= maxWidth {
		return s
	}
	
	runes := []rune(s)
	width := 0
	
	for i, r := range runes {
		charWidth := util.RuneWidth(r)
		if width+charWidth > maxWidth {
			return string(runes[:i])
		}
		width += charWidth
	}
	
	return s
}

func (w *Window) CursorPosition() (int, int) {
	bufferPos := w.buffer.Cursor()
	log.Info("SCROLL_TIMING: CursorPosition calculation - buffer cursor at (%d,%d), scrollTop=%d", bufferPos.Row, bufferPos.Col, w.scrollTop)
	
	if bufferPos.Row < len(w.buffer.content) {
		line := w.buffer.content[bufferPos.Row]
		if bufferPos.Col <= len(line) {
			// Calculate display width up to cursor position
			displayCol := util.StringWidthUpTo(line, bufferPos.Col)
			
			if w.lineWrap {
				// Line wrapping mode: calculate which wrapped line the cursor is on
				screenRow, wrappedCol := w.calculateWrappedCursorPosition(bufferPos.Row, displayCol)
				log.Info("SCROLL_TIMING: CursorPosition result (wrapped) - screen (%d,%d)", screenRow, wrappedCol)
				return screenRow, wrappedCol
			} else {
				// No wrapping: apply horizontal scrolling
				screenRow := bufferPos.Row - w.scrollTop
				screenCol := displayCol - w.scrollLeft
				// Don't clamp screenCol to 0 - let it be negative if cursor is left of visible area
				log.Info("SCROLL_TIMING: CursorPosition result (no wrap) - screen (%d,%d)", screenRow, screenCol)
				return screenRow, screenCol
			}
		}
	}
	
	// Fallback
	screenRow := bufferPos.Row - w.scrollTop
	log.Info("SCROLL_TIMING: CursorPosition result (fallback) - screen (%d,%d)", screenRow, bufferPos.Col)
	return screenRow, bufferPos.Col
}

// calculateWrappedCursorPosition calculates the screen position when line wrapping is enabled
func (w *Window) calculateWrappedCursorPosition(bufferRow int, cursorDisplayCol int) (int, int) {
	content := w.buffer.Content()
	
	// If cursor is above scroll area, return negative screen row
	if bufferRow < w.scrollTop {
		// Simple calculation for lines above scroll area
		return bufferRow - w.scrollTop, cursorDisplayCol
	}
	
	// Count how many screen lines are used by buffer lines before this one
	screenRow := 0
	
	// Count wrapped lines from scroll top to cursor row
	for row := w.scrollTop; row < bufferRow && row < len(content); row++ {
		if row >= 0 {
			line := content[row]
			wrappedLines := w.wrapLine(line)
			screenRow += len(wrappedLines)
		}
	}
	
	// Now handle the cursor's line - find which wrapped segment it's in
	if bufferRow >= 0 && bufferRow < len(content) {
		line := content[bufferRow]
		wrappedLines := w.wrapLine(line)
		
		// Find which wrapped line contains our cursor
		currentWidth := 0
		for i, wrappedLine := range wrappedLines {
			lineWidth := util.StringWidth(wrappedLine)
			if cursorDisplayCol <= currentWidth + lineWidth {
				// Cursor is in this wrapped line
				colInWrappedLine := cursorDisplayCol - currentWidth
				return screenRow + i, colInWrappedLine
			}
			currentWidth += lineWidth
		}
		
		// Cursor is at the end of the last wrapped line
		if len(wrappedLines) > 0 {
			lastLineWidth := util.StringWidth(wrappedLines[len(wrappedLines)-1])
			return screenRow + len(wrappedLines) - 1, lastLineWidth
		}
	}
	
	return screenRow, 0
}