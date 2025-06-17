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
	input         io.Reader
	output        io.Writer
	width         int
	height        int
	cursorLine    int
	cursorCol     int
	inRawMode     bool
	originalState []byte
	// ダブルバッファリング用
	buffer        *strings.Builder
	isBuffering   bool
}

// NewTerminal creates a new terminal instance
func NewTerminal(input io.Reader, output io.Writer) *Terminal {
	t := &Terminal{
		input:  input,
		output: output,
		buffer: &strings.Builder{},
	}
	
	// Get terminal size
	t.UpdateSize()
	
	return t
}

// UpdateSize updates the terminal size (public method for resize handling)
func (t *Terminal) UpdateSize() {
	// Try to get terminal size using ioctl syscall
	width, height := getTerminalSize()
	
	if width > 0 && height > 0 {
		t.width = width
		t.height = height
		return
	}
	
	// Try to get terminal size using stty as fallback
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

// getTerminalSize gets terminal size using ioctl syscall
func getTerminalSize() (width, height int) {
	// Try to get size from stdout file descriptor
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(0x5413), // TIOCGWINSZ
		uintptr(unsafe.Pointer(ws)))
	
	if retCode == 0 {
		return int(ws.Col), int(ws.Row)
	}
	
	// If stdout fails, try stderr
	retCode, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stderr),
		uintptr(0x5413), // TIOCGWINSZ
		uintptr(unsafe.Pointer(ws)))
	
	if retCode == 0 {
		return int(ws.Col), int(ws.Row)
	}
	
	// If both fail, try stdin
	retCode, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(0x5413), // TIOCGWINSZ
		uintptr(unsafe.Pointer(ws)))
	
	if retCode == 0 {
		return int(ws.Col), int(ws.Row)
	}
	
	// All failed
	_ = errno // suppress unused variable warning
	return 0, 0
}

// Clear clears the terminal screen (with buffering support)
func (t *Terminal) Clear() {
	if t.isBuffering {
		t.buffer.WriteString("\033[2J\033[H")
	} else {
		fmt.Fprint(t.output, "\033[2J\033[H")
	}
}

// MoveCursor moves the cursor to the specified position (1-based) (with buffering support)
func (t *Terminal) MoveCursor(line, col int) {
	text := fmt.Sprintf("\033[%d;%dH", line, col)
	if t.isBuffering {
		t.buffer.WriteString(text)
	} else {
		fmt.Fprint(t.output, text)
	}
	t.cursorLine = line
	t.cursorCol = col
}

// GetCursorPos returns the current cursor position
func (t *Terminal) GetCursorPos() (line, col int) {
	return t.cursorLine, t.cursorCol
}

// Print prints text at the current cursor position (with buffering support)
func (t *Terminal) Print(text string) {
	if t.isBuffering {
		t.buffer.WriteString(text)
	} else {
		fmt.Fprint(t.output, text)
	}
}

// Printf prints formatted text at the current cursor position (with buffering support)
func (t *Terminal) Printf(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	if t.isBuffering {
		t.buffer.WriteString(text)
	} else {
		fmt.Fprint(t.output, text)
	}
}

// PrintAt prints text at the specified position
func (t *Terminal) PrintAt(line, col int, text string) {
	t.MoveCursor(line, col)
	t.Print(text)
}

// ClearLine clears the current line (with buffering support)
func (t *Terminal) ClearLine() {
	if t.isBuffering {
		t.buffer.WriteString("\033[2K")
	} else {
		fmt.Fprint(t.output, "\033[2K")
	}
}

// ClearToEndOfLine clears from cursor to end of line (with buffering support)
func (t *Terminal) ClearToEndOfLine() {
	if t.isBuffering {
		t.buffer.WriteString("\033[K")
	} else {
		fmt.Fprint(t.output, "\033[K")
	}
}

// ShowCursor shows the cursor (with buffering support)
func (t *Terminal) ShowCursor() {
	if t.isBuffering {
		t.buffer.WriteString("\033[?25h")
	} else {
		fmt.Fprint(t.output, "\033[?25h")
	}
}

// HideCursor hides the cursor (with buffering support)
func (t *Terminal) HideCursor() {
	if t.isBuffering {
		t.buffer.WriteString("\033[?25l")
	} else {
		fmt.Fprint(t.output, "\033[?25l")
	}
}

// Flush flushes the output (handles buffering)
func (t *Terminal) Flush() {
	if t.isBuffering && t.buffer.Len() > 0 {
		// Write buffered content to output
		fmt.Fprint(t.output, t.buffer.String())
		t.buffer.Reset()
	}
	
	if f, ok := t.output.(*os.File); ok {
		f.Sync()
	}
}

// SetColor sets the text color (basic 8 colors) (with buffering support)
func (t *Terminal) SetColor(fg, bg int) {
	var colorCode string
	if fg >= 0 && fg <= 7 {
		colorCode += fmt.Sprintf("\033[3%dm", fg)
	}
	if bg >= 0 && bg <= 7 {
		colorCode += fmt.Sprintf("\033[4%dm", bg)
	}
	
	if colorCode != "" {
		if t.isBuffering {
			t.buffer.WriteString(colorCode)
		} else {
			fmt.Fprint(t.output, colorCode)
		}
	}
}

// ResetColor resets colors to default (with buffering support)
func (t *Terminal) ResetColor() {
	if t.isBuffering {
		t.buffer.WriteString("\033[0m")
	} else {
		fmt.Fprint(t.output, "\033[0m")
	}
}

// SetBold sets bold text (with buffering support)
func (t *Terminal) SetBold(bold bool) {
	var code string
	if bold {
		code = "\033[1m"
	} else {
		code = "\033[22m"
	}
	
	if t.isBuffering {
		t.buffer.WriteString(code)
	} else {
		fmt.Fprint(t.output, code)
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

// バッファリング制御メソッド

// StartBuffering starts buffering terminal output
func (t *Terminal) StartBuffering() {
	t.isBuffering = true
	t.buffer.Reset()
}

// StopBuffering stops buffering and flushes any buffered output
func (t *Terminal) StopBuffering() {
	if t.isBuffering {
		t.Flush()
		t.isBuffering = false
	}
}

// IsBuffering returns whether terminal output is being buffered
func (t *Terminal) IsBuffering() bool {
	return t.isBuffering
}