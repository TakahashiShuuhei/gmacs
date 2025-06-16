package input

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
)

// RawKeyboard handles raw terminal input for proper key detection
type RawKeyboard struct {
	input      *os.File
	oldState   *term.State
	inRawMode  bool
}

// NewRawKeyboard creates a new raw keyboard handler
func NewRawKeyboard() (*RawKeyboard, error) {
	// os.Stdin is already *os.File
	return &RawKeyboard{
		input: os.Stdin,
	}, nil
}

// EnableRawMode enables raw mode for character-by-character input
func (rk *RawKeyboard) EnableRawMode() error {
	if rk.inRawMode {
		return nil
	}
	
	// Check if input is a terminal
	if !term.IsTerminal(int(rk.input.Fd())) {
		return fmt.Errorf("input is not a terminal")
	}
	
	// Save current state
	state, err := term.GetState(int(rk.input.Fd()))
	if err != nil {
		return fmt.Errorf("failed to get terminal state: %v", err)
	}
	rk.oldState = state
	
	// Enable raw mode
	_, err = term.MakeRaw(int(rk.input.Fd()))
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %v", err)
	}
	
	rk.inRawMode = true
	return nil
}

// DisableRawMode restores the original terminal state
func (rk *RawKeyboard) DisableRawMode() error {
	if !rk.inRawMode || rk.oldState == nil {
		return nil
	}
	
	err := term.Restore(int(rk.input.Fd()), rk.oldState)
	if err != nil {
		return fmt.Errorf("failed to restore terminal state: %v", err)
	}
	
	rk.inRawMode = false
	return nil
}

// ReadKey reads a single key press in raw mode
func (rk *RawKeyboard) ReadKey() (*KeyEvent, error) {
	if !rk.inRawMode {
		return nil, fmt.Errorf("not in raw mode")
	}
	
	// Read one byte at a time
	buf := make([]byte, 1)
	n, err := rk.input.Read(buf)
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read from input: %v", err)
	}
	
	if n == 0 {
		return nil, fmt.Errorf("no data read")
	}
	
	b := buf[0]
	
	// Handle escape sequences (multi-byte)
	if b == 0x1b { // ESC
		// Read the next byte to see if it's an escape sequence
		nextBuf := make([]byte, 1)
		n, err := rk.input.Read(nextBuf)
		if err != nil || n == 0 {
			// Just ESC key by itself
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("escape"),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		}
		
		// Alt+key combination
		nextByte := nextBuf[0]
		char := rune(nextByte)
		
		return &KeyEvent{
			Key:       keymap.NewAltKey(char),
			Raw:       []byte{b, nextByte},
			Printable: false,
		}, nil
	}
	
	// Handle control characters
	if b < 32 {
		switch b {
		case 0x03: // Ctrl+C
			return &KeyEvent{
				Key:       keymap.NewCtrlKey('c'),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		case 0x07: // Ctrl+G
			return &KeyEvent{
				Key:       keymap.NewCtrlKey('g'),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		case 0x18: // Ctrl+X
			return &KeyEvent{
				Key:       keymap.NewCtrlKey('x'),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		case 0x0D, 0x0A: // Enter/Return
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("return"),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		case 0x09: // Tab
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("tab"),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		case 0x7F, 0x08: // Backspace/Delete
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("backspace"),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		default:
			// Other control characters (Ctrl+A through Ctrl+Z)
			if b >= 1 && b <= 26 {
				char := rune('a' + b - 1)
				return &KeyEvent{
					Key:       keymap.NewCtrlKey(char),
					Raw:       []byte{b},
					Printable: false,
				}, nil
			}
		}
	}
	
	// Printable character
	char := rune(b)
	return &KeyEvent{
		Key:       keymap.NewKey(char),
		Raw:       []byte{b},
		Printable: true,
	}, nil
}

// IsRawMode returns whether raw mode is enabled
func (rk *RawKeyboard) IsRawMode() bool {
	return rk.inRawMode
}