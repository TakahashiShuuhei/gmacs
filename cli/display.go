package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
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
			if len(line) > width {
				line = line[:width]
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
	padding := strings.Repeat("-", d.width-len(modeLine))
	
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