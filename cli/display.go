package cli

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/log"
	"github.com/TakahashiShuuhei/gmacs/core/util"
)

type Display struct {
	width  int
	height int
}

func NewDisplay() *Display {
	// Get initial terminal size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Warn("Failed to get terminal size, using defaults: %v", err)
		width, height = 80, 24
	}
	
	log.Info("Initial terminal size: %dx%d", width, height)
	
	return &Display{
		width:  width,
		height: height,
	}
}

func (d *Display) Clear() {
	fmt.Print("\033[2J\033[H")
}

// ClearAndExit clears the screen and prepares for clean exit
func (d *Display) ClearAndExit() {
	// Clear entire screen
	fmt.Print("\033[2J")
	// Move cursor to top-left
	fmt.Print("\033[H")
	// Show cursor (in case it was hidden)
	fmt.Print("\033[?25h")
	// Reset terminal attributes
	fmt.Print("\033[0m")
}

func (d *Display) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row+1, col+1)
}

func (d *Display) Render(editor *domain.Editor) {
	d.Clear()
	
	layout := editor.Layout()
	if layout == nil {
		return
	}
	
	// Get all window nodes for rendering
	windowNodes := layout.GetAllWindowNodes()
	currentWindow := editor.CurrentWindow()
	
	// Render all windows
	for _, node := range windowNodes {
		if node.Window != nil {
			d.renderWindow(node)
		}
	}
	
	// Render single mode line (for current buffer) and minibuffer at bottom
	d.renderModeLine(editor)
	d.renderMinibuffer(editor)
	
	// Position cursor based on whether minibuffer is active
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() && (minibuffer.Mode() == domain.MinibufferCommand || minibuffer.Mode() == domain.MinibufferFile) {
		// Position cursor in minibuffer (last line)
		promptLen := util.StringWidth(minibuffer.Prompt())
		
		// Calculate cursor position within the content up to cursor position
		content := minibuffer.Content()
		cursorPosInContent := minibuffer.CursorPosition()
		contentToCursor := string([]rune(content)[:cursorPosInContent])
		contentWidth := util.StringWidth(contentToCursor)
		
		cursorPos := promptLen + contentWidth
		terminalHeight := d.height
		log.Debug("Moving cursor to minibuffer position (%d, %d) - prompt=%q promptLen=%d, cursorInContent=%d, contentWidth=%d", 
			terminalHeight-1, cursorPos, minibuffer.Prompt(), promptLen, cursorPosInContent, contentWidth)
		d.MoveCursor(terminalHeight-1, cursorPos)
	} else if currentWindow != nil {
		// Position cursor in current window (buffer area)
		d.positionCursorInWindow(currentWindow, layout)
	}
}

func (d *Display) renderModeLine(editor *domain.Editor) {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return
	}
	
	fmt.Print("\r\n")
	
	// Always show mode line
	modeLine := fmt.Sprintf(" %s ", buffer.Name())
	// Calculate padding based on display width, not rune count
	modeLineWidth := util.StringWidth(modeLine)
	paddingLength := d.width - modeLineWidth
	if paddingLength < 0 {
		paddingLength = 0
	}
	padding := strings.Repeat("-", paddingLength)
	
	fmt.Printf("%s%s", modeLine, padding)
}

func (d *Display) renderMinibuffer(editor *domain.Editor) {
	minibuffer := editor.Minibuffer()
	
	fmt.Print("\r\n")
	
	var content string
	
	if minibuffer.IsActive() {
		// Show minibuffer content
		content = minibuffer.GetDisplayText()
	} else {
		// Check if there's a key sequence in progress
		keySequence := editor.GetKeySequenceInProgress()
		if keySequence != "" {
			content = keySequence
		}
	}
	
	if content != "" {
		contentWidth := util.StringWidth(content)
		
		// Truncate if too long
		if contentWidth > d.width {
			content = truncateToWidth(content, d.width)
		}
		
		fmt.Print(content)
		
		// Pad with spaces to clear any remaining text
		paddingLength := d.width - util.StringWidth(content)
		if paddingLength > 0 {
			padding := strings.Repeat(" ", paddingLength)
			fmt.Print(padding)
		}
	} else {
		// Empty minibuffer line
		padding := strings.Repeat(" ", d.width)
		fmt.Print(padding)
	}
}

func (d *Display) Size() (int, int) {
	return d.width, d.height
}

func (d *Display) Resize(width, height int) {
	d.width = width
	d.height = height
}

func (d *Display) ShowMessage(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

// truncateToWidth truncates a string to fit within the specified display width
func truncateToWidth(s string, maxWidth int) string {
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
func (d *Display) renderWindow(node *domain.WindowLayoutNode) {
	if node.Window == nil {
		return
	}
	
	window := node.Window
	lines := window.VisibleLines()
	_, windowContentHeight := window.Size()
	
	// Render each line of the window content
	for i := 0; i < windowContentHeight; i++ {
		// Position cursor at the start of this line
		d.MoveCursor(node.Y+i, node.X)
		
		if i < len(lines) {
			line := lines[i]
			
			// Truncate line to fit window width
			if util.StringWidth(line) > node.Width {
				line = truncateToWidth(line, node.Width)
			}
			
			// Pad line to window width to clear any previous content
			lineWidth := util.StringWidth(line)
			if lineWidth < node.Width {
				padding := strings.Repeat(" ", node.Width-lineWidth)
				line = line + padding
			}
			
			fmt.Print(line)
		} else {
			// Empty line - fill with spaces
			fmt.Print(strings.Repeat(" ", node.Width))
		}
	}
}

// renderWindowModeLine renders the mode line for a specific window
func (d *Display) renderWindowModeLine(node *domain.WindowLayoutNode, modeLineRow int) {
	if node.Window == nil || node.Window.Buffer() == nil {
		return
	}
	
	buffer := node.Window.Buffer()
	
	// Position cursor at the mode line position for this window
	d.MoveCursor(modeLineRow, node.X)
	
	// Create mode line content
	modeLine := fmt.Sprintf(" %s ", buffer.Name())
	modeLineWidth := util.StringWidth(modeLine)
	
	// Calculate padding to fill window width
	paddingLength := node.Width - modeLineWidth
	if paddingLength < 0 {
		// Mode line too long, truncate
		modeLine = truncateToWidth(modeLine, node.Width)
		paddingLength = 0
	}
	
	padding := strings.Repeat("-", paddingLength)
	fmt.Printf("%s%s", modeLine, padding)
}

// positionCursorInWindow positions the cursor in the current window
func (d *Display) positionCursorInWindow(currentWindow *domain.Window, layout *domain.WindowLayout) {
	// Find the window node for the current window
	windowNodes := layout.GetAllWindowNodes()
	var currentNode *domain.WindowLayoutNode
	
	for _, node := range windowNodes {
		if node.Window == currentWindow {
			currentNode = node
			break
		}
	}
	
	if currentNode == nil {
		return
	}
	
	// Get cursor position relative to window content
	cursorRow, cursorCol := currentWindow.CursorPosition()
	_, windowContentHeight := currentWindow.Size()
	
	// Check if cursor is within visible area
	if cursorRow >= 0 && cursorRow < windowContentHeight && cursorCol >= 0 {
		// Convert to absolute screen position
		absoluteRow := currentNode.Y + cursorRow
		absoluteCol := currentNode.X + cursorCol
		
		log.Debug("Moving cursor to window position (%d, %d) -> screen (%d, %d)", 
			cursorRow, cursorCol, absoluteRow, absoluteCol)
		d.MoveCursor(absoluteRow, absoluteCol)
	} else {
		log.Debug("Cursor outside visible area: window cursor (%d, %d), content height %d", 
			cursorRow, cursorCol, windowContentHeight)
	}
}