package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/log"
	"github.com/TakahashiShuuhei/gmacs/core/util"
)

type Display struct {
	width  int
	height int
}

func NewDisplay() *Display {
	return &Display{
		width:  80,
		height: 24,
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
	width, height := window.Size()
	
	for i := 0; i < height-1; i++ {
		if i < len(lines) {
			line := lines[i]
			
			// Truncate by display width, not rune count
			if util.StringWidth(line) > width {
				line = truncateToWidth(line, width)
				log.Debug("Truncated line %d to width %d: %q", i, width, line)
			}
			fmt.Print(line)
		}
		if i < height-2 {
			fmt.Print("\r\n")
		}
	}
	
	d.renderBottomLine(editor)
	
	// Position cursor based on whether minibuffer is active
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() && minibuffer.Mode() == domain.MinibufferCommand {
		// Position cursor in minibuffer
		promptLen := util.StringWidth(minibuffer.Prompt())
		cursorPos := promptLen + minibuffer.CursorPosition()
		log.Debug("Moving cursor to minibuffer position (%d, %d)", height-1, cursorPos)
		d.MoveCursor(height-1, cursorPos)
	} else {
		// Position cursor in main window
		cursorRow, cursorCol := window.CursorPosition()
		if cursorRow >= 0 && cursorRow < height-1 {
			log.Debug("Moving cursor to screen position (%d, %d)", cursorRow, cursorCol)
			d.MoveCursor(cursorRow, cursorCol)
		}
	}
}

func (d *Display) renderBottomLine(editor *domain.Editor) {
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
		// Show normal mode line
		buffer := editor.CurrentBuffer()
		if buffer == nil {
			return
		}
		
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