package input

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
)

// KeyEvent represents a key input event
type KeyEvent struct {
	Key       keymap.Key
	Raw       []byte
	Printable bool
}

// Keyboard handles keyboard input
type Keyboard struct {
	input      io.Reader
	rawMode    bool
	oldState   []byte
	buffer     []byte
	scanner    *bufio.Scanner
}

// NewKeyboard creates a new keyboard handler
func NewKeyboard(input io.Reader) *Keyboard {
	kb := &Keyboard{
		input:   input,
		scanner: bufio.NewScanner(input),
	}
	
	return kb
}

// ReadKey reads a single key press (non-blocking in raw mode)
func (kb *Keyboard) ReadKey() (*KeyEvent, error) {
	// For now, use line-based input for simplicity
	// In a full implementation, we would use raw mode for individual key presses
	
	if !kb.scanner.Scan() {
		return nil, io.EOF
	}
	
	line := kb.scanner.Text()
	if line == "" {
		return &KeyEvent{
			Key:       keymap.NewSpecialKey("return"),
			Raw:       []byte{'\n'},
			Printable: false,
		}, nil
	}
	
	// Parse the input as a key sequence
	key, err := kb.parseInput(line)
	if err != nil {
		// If parsing fails, treat as literal text
		if len(line) == 1 {
			return &KeyEvent{
				Key:       keymap.NewKey(rune(line[0])),
				Raw:       []byte(line),
				Printable: true,
			}, nil
		}
		return &KeyEvent{
			Key:       keymap.Key{Special: line},
			Raw:       []byte(line),
			Printable: false,
		}, nil
	}
	
	return &KeyEvent{
		Key:       key,
		Raw:       []byte(line),
		Printable: len(line) == 1 && line[0] >= 32 && line[0] <= 126,
	}, nil
}

// ReadLine reads a complete line of input
func (kb *Keyboard) ReadLine() (string, error) {
	if !kb.scanner.Scan() {
		return "", io.EOF
	}
	return kb.scanner.Text(), nil
}

// parseInput attempts to parse input as a key sequence
func (kb *Keyboard) parseInput(input string) (keymap.Key, error) {
	// Handle escape sequences (like ESC+x for M-x)
	if len(input) >= 2 && input[0] == 0x1b { // ESC character (27)
		if len(input) == 2 {
			// Alt+key combination: ESC + key
			char := rune(input[1])
			return keymap.NewAltKey(char), nil
		}
	}
	
	// Handle Ctrl+key combinations (single byte control characters)
	if len(input) == 1 {
		b := input[0]
		switch b {
		case 0x18: // Ctrl+X
			return keymap.NewCtrlKey('x'), nil
		case 0x03: // Ctrl+C
			return keymap.NewCtrlKey('c'), nil
		case 0x07: // Ctrl+G
			return keymap.NewCtrlKey('g'), nil
		case 0x06: // Ctrl+F
			return keymap.NewCtrlKey('f'), nil
		case 0x13: // Ctrl+S
			return keymap.NewCtrlKey('s'), nil
		case 0x01: // Ctrl+A
			return keymap.NewCtrlKey('a'), nil
		case 0x05: // Ctrl+E
			return keymap.NewCtrlKey('e'), nil
		case 0x0E: // Ctrl+N
			return keymap.NewCtrlKey('n'), nil
		case 0x10: // Ctrl+P
			return keymap.NewCtrlKey('p'), nil
		}
		
		// Handle other control characters (0x01-0x1F)
		if b >= 0x01 && b <= 0x1F && b != 0x09 && b != 0x0A && b != 0x0D {
			// Convert control character back to letter
			char := rune('a' + b - 1)
			return keymap.NewCtrlKey(char), nil
		}
	}
	
	// Handle ^[x format (sometimes displayed as such in terminals)
	if strings.HasPrefix(input, "^[") && len(input) == 3 {
		char := rune(input[2])
		return keymap.NewAltKey(char), nil
	}
	
	// Handle special key names
	switch strings.ToLower(input) {
	case "return", "enter":
		return keymap.NewSpecialKey("return"), nil
	case "tab":
		return keymap.NewSpecialKey("tab"), nil
	case "backspace", "bs":
		return keymap.NewSpecialKey("backspace"), nil
	case "delete", "del":
		return keymap.NewSpecialKey("delete"), nil
	case "escape", "esc":
		return keymap.NewSpecialKey("escape"), nil
	case "space":
		return keymap.NewKey(' '), nil
	}
	
	// Handle Ctrl+key combinations (C-x format)
	if strings.HasPrefix(input, "C-") || strings.HasPrefix(input, "c-") {
		if len(input) == 3 {
			char := rune(strings.ToLower(input)[2])
			return keymap.NewCtrlKey(char), nil
		}
	}
	
	// Handle Alt+key combinations (M-x format)
	if strings.HasPrefix(input, "M-") || strings.HasPrefix(input, "m-") {
		if len(input) == 3 {
			char := rune(strings.ToLower(input)[2])
			return keymap.NewAltKey(char), nil
		}
	}
	
	// Single character
	if len(input) == 1 {
		return keymap.NewKey(rune(input[0])), nil
	}
	
	// Try parsing as a key sequence
	seq, err := keymap.ParseKeySequence(input)
	if err != nil {
		return keymap.Key{}, err
	}
	
	if len(seq) == 1 {
		return seq[0], nil
	}
	
	return keymap.Key{}, err
}

// EnableRawMode enables raw mode for character-by-character input
func (kb *Keyboard) EnableRawMode() error {
	if f, ok := kb.input.(*os.File); ok {
		// Save current state
		cmd := exec.Command("stty", "-g")
		cmd.Stdin = f
		if output, err := cmd.Output(); err == nil {
			kb.oldState = output
		}
		
		// Enable raw mode
		cmd = exec.Command("stty", "raw", "-echo")
		cmd.Stdin = f
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err == nil {
			kb.rawMode = true
		}
		return err
	}
	
	return syscall.ENOTTY
}

// DisableRawMode disables raw mode and restores original terminal settings
func (kb *Keyboard) DisableRawMode() error {
	if !kb.rawMode {
		return nil
	}
	
	if f, ok := kb.input.(*os.File); ok && len(kb.oldState) > 0 {
		// Restore original state
		cmd := exec.Command("stty", string(kb.oldState))
		cmd.Stdin = f
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err == nil {
			kb.rawMode = false
		}
		return err
	}
	
	return nil
}

// IsRawMode returns whether raw mode is enabled
func (kb *Keyboard) IsRawMode() bool {
	return kb.rawMode
}

// ReadKeySequence reads a sequence of keys (for multi-key bindings like C-x C-f)
func (kb *Keyboard) ReadKeySequence() (keymap.KeySequence, error) {
	// For now, read a single key and return it as a sequence
	keyEvent, err := kb.ReadKey()
	if err != nil {
		return nil, err
	}
	
	return keymap.KeySequence{keyEvent.Key}, nil
}

// CreateStandardKeyboard creates a keyboard using standard input
func CreateStandardKeyboard() *Keyboard {
	return NewKeyboard(os.Stdin)
}