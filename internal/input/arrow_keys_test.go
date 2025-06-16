package input

import (
	"bytes"
	"io"
	"testing"
	
	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
)

func TestArrowKeyDetection(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectedKey string
	}{
		{"Up arrow", []byte{0x1b, '[', 'A'}, "up"},
		{"Down arrow", []byte{0x1b, '[', 'B'}, "down"},
		{"Right arrow", []byte{0x1b, '[', 'C'}, "right"},
		{"Left arrow", []byte{0x1b, '[', 'D'}, "left"},
		{"Home key (ESC[H)", []byte{0x1b, '[', 'H'}, "home"},
		{"End key (ESC[F)", []byte{0x1b, '[', 'F'}, "end"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test keyboard with test input
			inputReader := bytes.NewReader(tc.input)
			rk := &testRawKeyboard{
				reader:    inputReader,
				inRawMode: true,
			}
			
			keyEvent, err := rk.ReadKey()
			if err != nil {
				t.Errorf("ReadKey failed: %v", err)
				return
			}
			
			if keyEvent.Key.String() != tc.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tc.expectedKey, keyEvent.Key.String())
			}
			
			if keyEvent.Printable {
				t.Errorf("Arrow key should not be printable")
			}
		})
	}
}

func TestExtendedEscapeSequences(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectedKey string
	}{
		{"Home key (ESC[1~)", []byte{0x1b, '[', '1', '~'}, "home"},
		{"End key (ESC[4~)", []byte{0x1b, '[', '4', '~'}, "end"},
		{"Page Up (ESC[5~)", []byte{0x1b, '[', '5', '~'}, "page-up"},
		{"Page Down (ESC[6~)", []byte{0x1b, '[', '6', '~'}, "page-down"},
		{"Delete key (ESC[3~)", []byte{0x1b, '[', '3', '~'}, "delete"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test keyboard with test input
			inputReader := bytes.NewReader(tc.input)
			rk := &testRawKeyboard{
				reader:    inputReader,
				inRawMode: true,
			}
			
			keyEvent, err := rk.ReadKey()
			if err != nil {
				t.Errorf("ReadKey failed: %v", err)
				return
			}
			
			if keyEvent.Key.String() != tc.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tc.expectedKey, keyEvent.Key.String())
			}
			
			if keyEvent.Printable {
				t.Errorf("Special key should not be printable")
			}
		})
	}
}

func TestAltKeyVsArrowKey(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectedKey string
	}{
		{"Alt+A", []byte{0x1b, 'A'}, "M-A"},
		{"Arrow Up", []byte{0x1b, '[', 'A'}, "up"},
		{"Alt+x", []byte{0x1b, 'x'}, "M-x"},
		{"Arrow Right", []byte{0x1b, '[', 'C'}, "right"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test keyboard with test input
			inputReader := bytes.NewReader(tc.input)
			rk := &testRawKeyboard{
				reader:    inputReader,
				inRawMode: true,
			}
			
			keyEvent, err := rk.ReadKey()
			if err != nil {
				t.Errorf("ReadKey failed: %v", err)
				return
			}
			
			if keyEvent.Key.String() != tc.expectedKey {
				t.Errorf("Expected key '%s', got '%s'", tc.expectedKey, keyEvent.Key.String())
			}
		})
	}
}

// testRawKeyboard is a test version that uses a reader instead of os.File
type testRawKeyboard struct {
	reader    io.Reader
	inRawMode bool
}

func (trk *testRawKeyboard) EnableRawMode() error {
	trk.inRawMode = true
	return nil
}

func (trk *testRawKeyboard) DisableRawMode() error {
	trk.inRawMode = false
	return nil
}

func (trk *testRawKeyboard) IsRawMode() bool {
	return trk.inRawMode
}

func (trk *testRawKeyboard) ReadKey() (*KeyEvent, error) {
	if !trk.inRawMode {
		return nil, io.ErrUnexpectedEOF
	}
	
	// Read one byte at a time
	buf := make([]byte, 1)
	n, err := trk.reader.Read(buf)
	if err != nil {
		return nil, err
	}
	
	if n == 0 {
		return nil, io.EOF
	}
	
	b := buf[0]
	
	// Handle escape sequences (simplified version of RawKeyboard logic)
	if b == 0x1b { // ESC
		// Read the next byte
		nextBuf := make([]byte, 1)
		n, err := trk.reader.Read(nextBuf)
		if err != nil || n == 0 {
			return &KeyEvent{
				Key:       keymap.NewSpecialKey("escape"),
				Raw:       []byte{b},
				Printable: false,
			}, nil
		}
		
		nextByte := nextBuf[0]
		
		// Check for ANSI escape sequences
		if nextByte == '[' {
			finalBuf := make([]byte, 1)
			n, err := trk.reader.Read(finalBuf)
			if err != nil || n == 0 {
				return &KeyEvent{
					Key:       keymap.NewAltKey('['),
					Raw:       []byte{b, nextByte},
					Printable: false,
				}, nil
			}
			
			finalByte := finalBuf[0]
			
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
				// Check for extended sequences
				if finalByte >= '0' && finalByte <= '9' {
					return trk.readExtendedEscapeSequence([]byte{b, nextByte, finalByte})
				}
				return &KeyEvent{
					Key:       keymap.NewSpecialKey("unknown"),
					Raw:       []byte{b, nextByte, finalByte},
					Printable: false,
				}, nil
			}
		}
		
		// Alt+key combination
		char := rune(nextByte)
		return &KeyEvent{
			Key:       keymap.NewAltKey(char),
			Raw:       []byte{b, nextByte},
			Printable: false,
		}, nil
	}
	
	// Regular character
	char := rune(b)
	return &KeyEvent{
		Key:       keymap.NewKey(char),
		Raw:       []byte{b},
		Printable: true,
	}, nil
}

func (trk *testRawKeyboard) readExtendedEscapeSequence(prefix []byte) (*KeyEvent, error) {
	buf := make([]byte, 1)
	n, err := trk.reader.Read(buf)
	if err != nil || n == 0 {
		return &KeyEvent{
			Key:       keymap.NewSpecialKey("incomplete-escape"),
			Raw:       prefix,
			Printable: false,
		}, nil
	}
	
	finalByte := buf[0]
	fullSequence := append(prefix, finalByte)
	
	if finalByte == '~' {
		sequenceStr := string(prefix[2:len(prefix)])
		
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
				Key:       keymap.NewSpecialKey("unknown-extended"),
				Raw:       fullSequence,
				Printable: false,
			}, nil
		}
	}
	
	return &KeyEvent{
		Key:       keymap.NewSpecialKey("unknown-escape"),
		Raw:       fullSequence,
		Printable: false,
	}, nil
}