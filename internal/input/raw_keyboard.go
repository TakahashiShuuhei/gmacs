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
		
		nextByte := nextBuf[0]
		
		// Check for ANSI escape sequences like ESC[A (arrow keys)
		if nextByte == '[' {
			// Read the final character of the escape sequence
			finalBuf := make([]byte, 1)
			n, err := rk.input.Read(finalBuf)
			if err != nil || n == 0 {
				// Incomplete sequence, treat as ESC + [
				return &KeyEvent{
					Key:       keymap.NewAltKey('['),
					Raw:       []byte{b, nextByte},
					Printable: false,
				}, nil
			}
			
			finalByte := finalBuf[0]
			
			// Handle arrow keys and other escape sequences
			switch finalByte {
			case 'A': // Up arrow
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("up"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			case 'B': // Down arrow
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("down"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			case 'C': // Right arrow
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("right"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			case 'D': // Left arrow
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("left"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			case 'H': // Home key
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("home"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			case 'F': // End key
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("end"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			default:
				// Check for extended sequences like ESC[1~ (Home), ESC[4~ (End), etc.
				if finalByte >= '0' && finalByte <= '9' {
					// This might be a longer sequence, try to read more
					return rk.readExtendedEscapeSequence([]byte{b, nextByte, finalByte})
				}
				
				// Unknown escape sequence
				return &KeyEvent{
					Key:       keymap.NewSpecialKey(fmt.Sprintf("escape-sequence-%c", finalByte)),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			}
		}
		
		// Alt+key combination (ESC + single character)
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
	
	// Handle UTF-8 multi-byte characters
	if b >= 0x80 {
		// This is the start of a UTF-8 multi-byte sequence
		return rk.readUTF8Character(b)
	}
	
	// Single-byte printable ASCII character
	if b >= 32 && b <= 126 {
		char := rune(b)
		return &KeyEvent{
			Key:       keymap.NewKey(char),
			Raw:       []byte{b},
			Printable: true,
		}, nil
	}
	
	// Other single-byte characters (not printable)
	return &KeyEvent{
		Key:       keymap.NewSpecialKey(fmt.Sprintf("byte-%d", b)),
		Raw:       []byte{b},
		Printable: false,
	}, nil
}

// readUTF8Character reads a complete UTF-8 character starting with the given byte
func (rk *RawKeyboard) readUTF8Character(firstByte byte) (*KeyEvent, error) {
	// Determine the number of bytes needed for this UTF-8 character
	var numBytes int
	if firstByte&0x80 == 0 {
		numBytes = 1 // ASCII (should not reach here)
	} else if firstByte&0xE0 == 0xC0 {
		numBytes = 2 // 110xxxxx
	} else if firstByte&0xF0 == 0xE0 {
		numBytes = 3 // 1110xxxx
	} else if firstByte&0xF8 == 0xF0 {
		numBytes = 4 // 11110xxx
	} else {
		// Invalid UTF-8 start byte
		return &KeyEvent{
			Key:       keymap.NewSpecialKey(fmt.Sprintf("invalid-utf8-%d", firstByte)),
			Raw:       []byte{firstByte},
			Printable: false,
		}, nil
	}
	
	// Read the remaining bytes
	utf8Bytes := make([]byte, numBytes)
	utf8Bytes[0] = firstByte
	
	for i := 1; i < numBytes; i++ {
		buf := make([]byte, 1)
		n, err := rk.input.Read(buf)
		if err != nil || n == 0 {
			// Incomplete UTF-8 sequence
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("incomplete-utf8"),
				Raw:       utf8Bytes[:i],
				Printable: false,
			}, nil
		}
		utf8Bytes[i] = buf[0]
	}
	
	// Convert to rune
	runes := []rune(string(utf8Bytes))
	if len(runes) != 1 {
		// Invalid UTF-8 sequence
		return &KeyEvent{
			Key:       keymap.NewSpecialKey("invalid-utf8"),
			Raw:       utf8Bytes,
			Printable: false,
		}, nil
	}
	
	char := runes[0]
	return &KeyEvent{
		Key:       keymap.NewKey(char),
		Raw:       utf8Bytes,
		Printable: true,
	}, nil
}

// readExtendedEscapeSequence reads extended escape sequences like ESC[1~
func (rk *RawKeyboard) readExtendedEscapeSequence(prefix []byte) (*KeyEvent, error) {
	// Try to read the rest of the sequence (expecting ~ as terminator)
	buf := make([]byte, 1)
	n, err := rk.input.Read(buf)
	if err != nil || n == 0 {
		// Incomplete sequence
		return &KeyEvent{
			Key:       keymap.NewSpecialKey("incomplete-escape"),
			Raw:       prefix,
			Printable: false,
		}, nil
	}
	
	finalByte := buf[0]
	fullSequence := append(prefix, finalByte)
	
	if finalByte == '~' {
		// Complete extended sequence
		sequenceStr := string(prefix[2:len(prefix)]) // Extract the number part
		
		switch sequenceStr {
		case "1": // Home
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("home"),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		case "4": // End
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("end"),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		case "5": // Page Up
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("page-up"),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		case "6": // Page Down
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("page-down"),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		case "3": // Delete key
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("delete"),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		default:
			return &KeyEvent{
				Key:       keymap.NewSpecialKey(fmt.Sprintf("extended-escape-%s", sequenceStr)),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		}
	}
	
	// Not a recognized extended sequence
	return &KeyEvent{
		Key:       keymap.NewSpecialKey("unknown-escape"),
		Raw:       fullSequence,
		Printable: false,
	}, nil
}

// IsRawMode returns whether raw mode is enabled
func (rk *RawKeyboard) IsRawMode() bool {
	return rk.inRawMode
}