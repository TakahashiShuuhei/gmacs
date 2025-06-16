package display

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// Terminal represents a terminal interface
type Terminal struct {
	input       io.Reader
	output      io.Writer
	width       int
	height      int
	cursorLine  int
	cursorCol   int
	inRawMode   bool
	originalState []byte
}

// NewTerminal creates a new terminal instance
func NewTerminal(input io.Reader, output io.Writer) *Terminal {
	t := &Terminal{
		input:  input,
		output: output,
	}
	
	// Get terminal size
	t.updateSize()
	
	return t
}

// updateSize updates the terminal size
func (t *Terminal) updateSize() {
	// Try to get terminal size using stty
	if cmd := exec.Command("stty", "size"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			parts := strings.Fields(string(output))
			if len(parts) == 2 {
				if height, err := strconv.Atoi(parts[0]); err == nil {
					t.height = height
				}
				if width, err := strconv.Atoi(parts[1]); err == nil {
					t.width = width
				}
			}
		}
	}
	
	// Fallback to default size
	if t.width == 0 {
		t.width = 80
	}
	if t.height == 0 {
		t.height = 24
	}
}

// Size returns the terminal size
func (t *Terminal) Size() (width, height int) {
	return t.width, t.height
}

// Clear clears the terminal screen
func (t *Terminal) Clear() {
	fmt.Fprint(t.output, "\033[2J\033[H")
}

// MoveCursor moves the cursor to the specified position (1-based)
func (t *Terminal) MoveCursor(line, col int) {
	fmt.Fprintf(t.output, "\033[%d;%dH", line, col)
	t.cursorLine = line
	t.cursorCol = col
}

// GetCursorPos returns the current cursor position
func (t *Terminal) GetCursorPos() (line, col int) {
	return t.cursorLine, t.cursorCol
}

// Print prints text at the current cursor position
func (t *Terminal) Print(text string) {
	fmt.Fprint(t.output, text)
}

// Printf prints formatted text at the current cursor position
func (t *Terminal) Printf(format string, args ...interface{}) {
	fmt.Fprintf(t.output, format, args...)
}

// PrintAt prints text at the specified position
func (t *Terminal) PrintAt(line, col int, text string) {
	t.MoveCursor(line, col)
	t.Print(text)
}

// ClearLine clears the current line
func (t *Terminal) ClearLine() {
	fmt.Fprint(t.output, "\033[2K")
}

// ClearToEndOfLine clears from cursor to end of line
func (t *Terminal) ClearToEndOfLine() {
	fmt.Fprint(t.output, "\033[K")
}

// ShowCursor shows the cursor
func (t *Terminal) ShowCursor() {
	fmt.Fprint(t.output, "\033[?25h")
}

// HideCursor hides the cursor
func (t *Terminal) HideCursor() {
	fmt.Fprint(t.output, "\033[?25l")
}

// Flush flushes the output
func (t *Terminal) Flush() {
	if f, ok := t.output.(*os.File); ok {
		f.Sync()
	}
}

// SetColor sets the text color (basic 8 colors)
func (t *Terminal) SetColor(fg, bg int) {
	if fg >= 0 && fg <= 7 {
		fmt.Fprintf(t.output, "\033[3%dm", fg)
	}
	if bg >= 0 && bg <= 7 {
		fmt.Fprintf(t.output, "\033[4%dm", bg)
	}
}

// ResetColor resets colors to default
func (t *Terminal) ResetColor() {
	fmt.Fprint(t.output, "\033[0m")
}

// SetBold sets bold text
func (t *Terminal) SetBold(bold bool) {
	if bold {
		fmt.Fprint(t.output, "\033[1m")
	} else {
		fmt.Fprint(t.output, "\033[22m")
	}
}

// GetSize attempts to get the current terminal size using ioctl
func (t *Terminal) GetSize() (width, height int, err error) {
	// Try ioctl method for more accurate size detection
	if f, ok := t.output.(*os.File); ok {
		var ws winsize
		ret, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
			uintptr(f.Fd()),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(&ws)))
		
		if ret == 0 && errno == 0 {
			return int(ws.Col), int(ws.Row), nil
		}
	}
	
	// Fallback to stored values
	return t.width, t.height, nil
}

// winsize represents terminal window size
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// DrawBox draws a simple box with borders
func (t *Terminal) DrawBox(startLine, startCol, width, height int, title string) {
	// Top border
	t.MoveCursor(startLine, startCol)
	t.Print("┌")
	for i := 0; i < width-2; i++ {
		t.Print("─")
	}
	t.Print("┐")
	
	// Title
	if title != "" && len(title) < width-4 {
		titlePos := startCol + (width-len(title))/2
		t.MoveCursor(startLine, titlePos)
		t.Print(title)
	}
	
	// Side borders
	for i := 1; i < height-1; i++ {
		t.MoveCursor(startLine+i, startCol)
		t.Print("│")
		t.MoveCursor(startLine+i, startCol+width-1)
		t.Print("│")
	}
	
	// Bottom border
	t.MoveCursor(startLine+height-1, startCol)
	t.Print("└")
	for i := 0; i < width-2; i++ {
		t.Print("─")
	}
	t.Print("┘")
}

// DrawHorizontalLine draws a horizontal line
func (t *Terminal) DrawHorizontalLine(line, startCol, length int) {
	t.MoveCursor(line, startCol)
	for i := 0; i < length; i++ {
		t.Print("─")
	}
}

// Colors constants
const (
	ColorBlack = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// Alternative terminal creation for standard streams
func NewStandardTerminal() *Terminal {
	return NewTerminal(os.Stdin, os.Stdout)
}