package input

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
)

func TestKeyboard(t *testing.T) {
	input := strings.NewReader("hello\n")
	kb := NewKeyboard(input)
	
	line, err := kb.ReadLine()
	if err != nil {
		t.Fatalf("Failed to read line: %v", err)
	}
	
	if line != "hello" {
		t.Errorf("Expected 'hello', got '%s'", line)
	}
}

func TestKeyboardParseInput(t *testing.T) {
	input := strings.NewReader("")
	kb := NewKeyboard(input)
	
	testCases := []struct {
		input    string
		expected keymap.Key
	}{
		{"a", keymap.NewKey('a')},
		{"C-x", keymap.NewCtrlKey('x')},
		{"M-x", keymap.NewAltKey('x')},
		{"return", keymap.NewSpecialKey("return")},
		{"tab", keymap.NewSpecialKey("tab")},
	}
	
	for _, tc := range testCases {
		key, err := kb.parseInput(tc.input)
		if err != nil {
			t.Errorf("Failed to parse '%s': %v", tc.input, err)
			continue
		}
		
		if key.String() != tc.expected.String() {
			t.Errorf("Parse '%s': expected '%s', got '%s'", 
				tc.input, tc.expected.String(), key.String())
		}
	}
}

func TestKeyboardReadKey(t *testing.T) {
	input := strings.NewReader("a\nC-x\n\n")
	kb := NewKeyboard(input)
	
	// Read 'a'
	keyEvent, err := kb.ReadKey()
	if err != nil {
		t.Fatalf("Failed to read key: %v", err)
	}
	if keyEvent.Key.String() != "a" {
		t.Errorf("Expected 'a', got '%s'", keyEvent.Key.String())
	}
	if !keyEvent.Printable {
		t.Error("'a' should be printable")
	}
	
	// Read 'C-x'
	keyEvent, err = kb.ReadKey()
	if err != nil {
		t.Fatalf("Failed to read key: %v", err)
	}
	if keyEvent.Key.String() != "C-x" {
		t.Errorf("Expected 'C-x', got '%s'", keyEvent.Key.String())
	}
	if keyEvent.Printable {
		t.Error("'C-x' should not be printable")
	}
	
	// Read empty line (return)
	keyEvent, err = kb.ReadKey()
	if err != nil {
		t.Fatalf("Failed to read key: %v", err)
	}
	if keyEvent.Key.String() != "return" {
		t.Errorf("Expected 'return', got '%s'", keyEvent.Key.String())
	}
}

func TestKeyboardReadKeySequence(t *testing.T) {
	input := strings.NewReader("C-x\n")
	kb := NewKeyboard(input)
	
	seq, err := kb.ReadKeySequence()
	if err != nil {
		t.Fatalf("Failed to read key sequence: %v", err)
	}
	
	if len(seq) != 1 {
		t.Errorf("Expected sequence length 1, got %d", len(seq))
	}
	
	if seq[0].String() != "C-x" {
		t.Errorf("Expected 'C-x', got '%s'", seq[0].String())
	}
}