package test

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/util"
)

type MockDisplay struct {
	width       int
	height      int
	content     []string
	cursorRow   int
	cursorCol   int
	modeLine    string
	renderCount int
}

func NewMockDisplay(width, height int) *MockDisplay {
	return &MockDisplay{
		width:  width,
		height: height,
		content: make([]string, height-1), // -1 for mode line
	}
}

func (d *MockDisplay) Render(editor *domain.Editor) {
	d.renderCount++
	
	window := editor.CurrentWindow()
	if window == nil {
		return
	}
	
	lines := window.VisibleLines()
	
	// Clear content
	for i := range d.content {
		d.content[i] = ""
	}
	
	// Render lines
	for i := 0; i < d.height-1; i++ {
		if i < len(lines) {
			line := lines[i]
			
			// Truncate by display width, not rune count
			if util.StringWidth(line) > d.width {
				line = truncateToWidthMock(line, d.width)
			}
			d.content[i] = line
		}
	}
	
	// Render mode line
	buffer := editor.CurrentBuffer()
	if buffer != nil {
		modeLine := fmt.Sprintf(" %s ", buffer.Name())
		modeLineWidth := util.StringWidth(modeLine)
		paddingLength := d.width - modeLineWidth
		if paddingLength < 0 {
			paddingLength = 0
		}
		padding := strings.Repeat("-", paddingLength)
		d.modeLine = modeLine + padding
	}
	
	// Get cursor position
	d.cursorRow, d.cursorCol = window.CursorPosition()
}

func (d *MockDisplay) GetContent() []string {
	return d.content
}

func (d *MockDisplay) GetModeLine() string {
	return d.modeLine
}

func (d *MockDisplay) GetCursorPosition() (int, int) {
	return d.cursorRow, d.cursorCol
}

func (d *MockDisplay) GetRenderCount() int {
	return d.renderCount
}

// Get full screen representation as string
func (d *MockDisplay) GetScreenText() string {
	var result strings.Builder
	
	for i, line := range d.content {
		// Pad line to full width
		lineRunes := utf8.RuneCountInString(line)
		padding := strings.Repeat(" ", d.width-lineRunes)
		result.WriteString(line + padding)
		
		if i < len(d.content)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// Get screen with cursor marked
func (d *MockDisplay) GetScreenWithCursor() string {
	lines := make([]string, len(d.content))
	copy(lines, d.content)
	
	// Mark cursor position
	if d.cursorRow >= 0 && d.cursorRow < len(lines) {
		line := lines[d.cursorRow]
		runes := []rune(line)
		
		// Insert cursor marker at position
		if d.cursorCol >= 0 && d.cursorCol <= len(runes) {
			before := string(runes[:d.cursorCol])
			after := string(runes[d.cursorCol:])
			lines[d.cursorRow] = before + "|" + after
		}
	}
	
	var result strings.Builder
	for i, line := range lines {
		// Pad line to full width (accounting for cursor marker)
		lineRunes := utf8.RuneCountInString(line)
		targetWidth := d.width
		if strings.Contains(line, "|") {
			targetWidth++ // cursor marker takes extra space
		}
		padding := strings.Repeat(" ", targetWidth-lineRunes)
		result.WriteString(line + padding)
		
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// Get detailed screen info for debugging
func (d *MockDisplay) GetScreenInfo() string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("Screen size: %dx%d\n", d.width, d.height))
	result.WriteString(fmt.Sprintf("Cursor: (%d, %d)\n", d.cursorRow, d.cursorCol))
	result.WriteString(fmt.Sprintf("Render count: %d\n", d.renderCount))
	result.WriteString("Content:\n")
	
	for i, line := range d.content {
		result.WriteString(fmt.Sprintf("  Line %d: %q (runes: %d)\n", 
			i, line, utf8.RuneCountInString(line)))
	}
	
	result.WriteString(fmt.Sprintf("Mode line: %q\n", d.modeLine))
	
	return result.String()
}

// truncateToWidthMock truncates a string to fit within the specified display width
func truncateToWidthMock(s string, maxWidth int) string {
	if util.StringWidth(s) <= maxWidth {
		return s
	}
	
	runes := []rune(s)
	width := 0
	
	for i, r := range runes {
		charWidth := util.RuneWidth(r)
		if width + charWidth > maxWidth {
			return string(runes[:i])
		}
		width += charWidth
	}
	
	return s
}