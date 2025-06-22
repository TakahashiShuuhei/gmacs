package test

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/util"
)

type MockDisplay struct {
	width        int
	height       int
	content      []string
	cursorRow    int
	cursorCol    int
	modeLine     string
	minibuffer   string
	renderCount  int
}

func NewMockDisplay(width, height int) *MockDisplay {
	// MockDisplay should match actual Display behavior
	// The content area size should match what the window reports
	contentHeight := height - 2  // Reserve 2 lines for mode line and minibuffer
	return &MockDisplay{
		width:   width,
		height:  height,
		content: make([]string, contentHeight),
	}
}

func (d *MockDisplay) Render(editor *domain.Editor) {
	d.renderCount++
	
	layout := editor.Layout()
	if layout == nil {
		return
	}
	
	// Initialize content to terminal size
	d.content = make([]string, d.height)
	for i := range d.content {
		d.content[i] = strings.Repeat(" ", d.width)
	}
	
	// Get all window nodes for rendering
	windowNodes := layout.GetAllWindowNodes()
	
	// Render all windows
	for _, node := range windowNodes {
		if node.Window != nil {
			d.renderWindow(node)
			d.renderWindowModeLine(node)
		}
	}
	
	// Render window borders
	d.renderWindowBorders(layout)
	
	// Always render mode line
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
	
	// Render minibuffer (separate from mode line)
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() {
		// Show minibuffer content
		content := minibuffer.GetDisplayText()
		contentWidth := util.StringWidth(content)
		
		// Truncate if too long
		if contentWidth > d.width {
			content = truncateToWidthMock(content, d.width)
		}
		
		// Pad with spaces
		paddingLength := d.width - util.StringWidth(content)
		if paddingLength > 0 {
			padding := strings.Repeat(" ", paddingLength)
			content += padding
		}
		
		d.minibuffer = content
	} else {
		// Empty minibuffer
		d.minibuffer = strings.Repeat(" ", d.width)
	}
	
	// Get cursor position (minibuffer or main window)
	if minibuffer.IsActive() && minibuffer.Mode() == domain.MinibufferCommand {
		// Position cursor in minibuffer (last line)
		promptLen := util.StringWidth(minibuffer.Prompt())
		d.cursorRow = d.height - 1
		d.cursorCol = promptLen + minibuffer.CursorPosition()
	} else {
		// Position cursor in current window
		currentWindow := editor.CurrentWindow()
		if currentWindow != nil {
			// Find the window node for the current window
			for _, node := range windowNodes {
				if node.Window == currentWindow {
					// Get cursor position relative to window content
					cursorRow, cursorCol := currentWindow.CursorPosition()
					// Convert to absolute screen position
					d.cursorRow = node.Y + cursorRow
					d.cursorCol = node.X + cursorCol
					break
				}
			}
		}
	}
}

func (d *MockDisplay) GetContent() []string {
	return d.content
}

func (d *MockDisplay) GetModeLine() string {
	return d.modeLine
}

func (d *MockDisplay) GetMinibuffer() string {
	return d.minibuffer
}

func (d *MockDisplay) GetCursorPosition() (int, int) {
	return d.cursorRow, d.cursorCol
}

func (d *MockDisplay) GetRenderCount() int {
	return d.renderCount
}

func (d *MockDisplay) Size() (int, int) {
	return d.width, d.height
}

func (d *MockDisplay) Resize(width, height int) {
	d.width = width
	d.height = height
	d.content = make([]string, height-2) // Update content array size
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

// renderWindow renders a single window at its designated position
func (d *MockDisplay) renderWindow(node *domain.WindowLayoutNode) {
	if node.Window == nil {
		return
	}
	
	window := node.Window
	lines := window.VisibleLines()
	_, windowContentHeight := window.Size()
	
	// Render each line of the window content
	for i := 0; i < windowContentHeight; i++ {
		row := node.Y + i
		if row >= d.height {
			break
		}
		
		if i < len(lines) {
			line := lines[i]
			
			// Truncate line to fit window width
			if util.StringWidth(line) > node.Width {
				line = truncateToWidthMock(line, node.Width)
			}
			
			// Insert the line into the display content at the correct position
			d.insertStringAt(row, node.X, line, node.Width)
		}
	}
}

// renderWindowModeLine renders the mode line for a specific window
func (d *MockDisplay) renderWindowModeLine(node *domain.WindowLayoutNode) {
	if node.Window == nil || node.Window.Buffer() == nil {
		return
	}
	
	buffer := node.Window.Buffer()
	_, windowContentHeight := node.Window.Size()
	
	// Mode line appears right after the window content
	modeLineRow := node.Y + windowContentHeight
	if modeLineRow >= d.height {
		return
	}
	
	// Create mode line content
	modeLine := fmt.Sprintf(" %s ", buffer.Name())
	modeLineWidth := util.StringWidth(modeLine)
	
	// Calculate padding to fill window width
	paddingLength := node.Width - modeLineWidth
	if paddingLength < 0 {
		// Mode line too long, truncate
		modeLine = truncateToWidthMock(modeLine, node.Width)
		paddingLength = 0
	}
	
	padding := strings.Repeat("-", paddingLength)
	fullModeLine := modeLine + padding
	
	// Insert the mode line into the display content
	d.insertStringAt(modeLineRow, node.X, fullModeLine, node.Width)
}

// renderWindowBorders renders borders between split windows
func (d *MockDisplay) renderWindowBorders(layout *domain.WindowLayout) {
	d.renderBordersForNode(layout.Root())
}

// renderBordersForNode recursively renders borders for split nodes
func (d *MockDisplay) renderBordersForNode(node *domain.WindowLayoutNode) {
	if node == nil {
		return
	}
	
	// Render borders for child nodes first
	if node.Left != nil {
		d.renderBordersForNode(node.Left)
	}
	if node.Right != nil {
		d.renderBordersForNode(node.Right)
	}
	
	// If this is a split node, render the border between children
	if node.SplitType == domain.SplitVertical && node.Left != nil && node.Right != nil {
		// Vertical split: draw vertical line between left and right
		borderX := node.Left.X + node.Left.Width
		if borderX >= d.width {
			return
		}
		
		for y := node.Y; y < node.Y+node.Height-1 && y < d.height; y++ { // -1 to avoid overwriting mode line
			d.insertCharAt(y, borderX, '│')
		}
	} else if node.SplitType == domain.SplitHorizontal && node.Left != nil && node.Right != nil {
		// Horizontal split: mode lines provide sufficient visual separation
		// No additional horizontal line needed (was obscuring bottom window content)
	}
}

// insertStringAt inserts a string at the specified position in the display content
func (d *MockDisplay) insertStringAt(row, col int, text string, maxWidth int) {
	if row < 0 || row >= len(d.content) || col < 0 || col >= d.width {
		return
	}
	
	// Convert the existing line to runes for proper character handling
	existingRunes := []rune(d.content[row])
	textRunes := []rune(text)
	
	// Ensure we have enough space in the existing line
	for len(existingRunes) < d.width {
		existingRunes = append(existingRunes, ' ')
	}
	
	// Insert the text, respecting width limits
	insertedWidth := 0
	for i, r := range textRunes {
		if col+i >= d.width || insertedWidth >= maxWidth {
			break
		}
		existingRunes[col+i] = r
		insertedWidth += util.RuneWidth(r)
	}
	
	d.content[row] = string(existingRunes)
}

// insertCharAt inserts a single character at the specified position
func (d *MockDisplay) insertCharAt(row, col int, char rune) {
	if row < 0 || row >= len(d.content) || col < 0 || col >= d.width {
		return
	}
	
	// Convert the existing line to runes for proper character handling
	existingRunes := []rune(d.content[row])
	
	// Ensure we have enough space in the existing line
	for len(existingRunes) <= col {
		existingRunes = append(existingRunes, ' ')
	}
	
	existingRunes[col] = char
	d.content[row] = string(existingRunes)
}