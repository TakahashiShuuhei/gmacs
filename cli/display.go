package cli

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/log"
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
			
			// Truncate by rune count, not byte count
			runes := []rune(line)
			if len(runes) > width {
				runes = runes[:width]
				line = string(runes)
				log.Debug("Truncated line %d to %d runes: %q", i, width, line)
			}
			fmt.Print(line)
		}
		if i < height-2 {
			fmt.Print("\r\n")
		}
	}
	
	d.renderModeLine(editor)
	
	cursorRow, cursorCol := window.CursorPosition()
	if cursorRow >= 0 && cursorRow < height-1 {
		d.MoveCursor(cursorRow, cursorCol)
	}
}

func (d *Display) renderModeLine(editor *domain.Editor) {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return
	}
	
	modeLine := fmt.Sprintf(" %s ", buffer.Name())
	// Calculate padding based on rune count, not byte count
	modeLineRunes := utf8.RuneCountInString(modeLine)
	paddingLength := d.width - modeLineRunes
	if paddingLength < 0 {
		paddingLength = 0
	}
	padding := strings.Repeat("-", paddingLength)
	
	fmt.Printf("\r\n%s%s", modeLine, padding)
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