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

func (d *Display) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row+1, col+1)
}

func (d *Display) Render(editor *domain.Editor) {
	d.Clear()
	
	window := editor.CurrentWindow()
	if window == nil {
		return
	}
	
	lines := window.VisibleLines()
	width, windowContentHeight := window.Size()
	
	// Render buffer content (window content area is already adjusted for mode line and minibuffer)
	for i := 0; i < windowContentHeight; i++ {
		if i < len(lines) {
			line := lines[i]
			
			// Truncate by display width, not rune count
			if util.StringWidth(line) > width {
				line = truncateToWidth(line, width)
				log.Debug("Truncated line %d to width %d: %q", i, width, line)
			}
			fmt.Print(line)
		}
		// Add newline after each line (to position cursor for next line)
		if i < windowContentHeight-1 {
			fmt.Print("\r\n")
		}
	}
	
	// Render mode line and minibuffer
	d.renderModeLine(editor)
	d.renderMinibuffer(editor)
	
	// Position cursor based on whether minibuffer is active
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() && minibuffer.Mode() == domain.MinibufferCommand {
		// Position cursor in minibuffer (last line)
		promptLen := util.StringWidth(minibuffer.Prompt())
		cursorPos := promptLen + minibuffer.CursorPosition()
		terminalHeight := d.height
		log.Debug("Moving cursor to minibuffer position (%d, %d)", terminalHeight-1, cursorPos)
		d.MoveCursor(terminalHeight-1, cursorPos)
	} else {
		// Position cursor in main window (buffer area)
		cursorRow, cursorCol := window.CursorPosition()
		// Window content area is height-2 (excluding mode line and minibuffer)
		// But we need to get the actual window content height from the window itself
		_, windowContentHeight := window.Size()
		if cursorRow >= 0 && cursorRow < windowContentHeight {
			log.Debug("Moving cursor to screen position (%d, %d)", cursorRow, cursorCol)
			d.MoveCursor(cursorRow, cursorCol)
		} else {
			log.Debug("Cursor outside visible area: screen row %d, content height %d", cursorRow, windowContentHeight)
		}
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
	
	if minibuffer.IsActive() {
		// Show minibuffer content
		content := minibuffer.GetDisplayText()
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