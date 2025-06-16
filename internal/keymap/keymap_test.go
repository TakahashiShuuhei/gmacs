package keymap

import (
	"testing"
)

func TestKey(t *testing.T) {
	// Test simple character key
	key := NewKey('a')
	if key.String() != "a" {
		t.Errorf("Expected 'a', got '%s'", key.String())
	}
	
	// Test Ctrl key
	ctrlKey := NewCtrlKey('x')
	if ctrlKey.String() != "C-x" {
		t.Errorf("Expected 'C-x', got '%s'", ctrlKey.String())
	}
	
	// Test Alt key
	altKey := NewAltKey('x')
	if altKey.String() != "M-x" {
		t.Errorf("Expected 'M-x', got '%s'", altKey.String())
	}
	
	// Test special key
	specialKey := NewSpecialKey("return")
	if specialKey.String() != "return" {
		t.Errorf("Expected 'return', got '%s'", specialKey.String())
	}
	
	// Test Ctrl+special key
	ctrlSpecial := NewCtrlSpecialKey("return")
	if ctrlSpecial.String() != "C-return" {
		t.Errorf("Expected 'C-return', got '%s'", ctrlSpecial.String())
	}
}

func TestKeySequence(t *testing.T) {
	seq := KeySequence{
		NewCtrlKey('x'),
		NewCtrlKey('f'),
	}
	
	expected := "C-x C-f"
	if seq.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, seq.String())
	}
	
	// Test empty sequence
	emptySeq := KeySequence{}
	if emptySeq.String() != "" {
		t.Errorf("Expected empty string for empty sequence, got '%s'", emptySeq.String())
	}
}

func TestKeymap(t *testing.T) {
	km := New("test-keymap")
	
	if km.Name() != "test-keymap" {
		t.Errorf("Expected name 'test-keymap', got '%s'", km.Name())
	}
	
	// Test binding
	seq := KeySequence{NewCtrlKey('x'), NewCtrlKey('f')}
	err := km.Bind(seq, "find-file")
	if err != nil {
		t.Fatalf("Failed to bind key: %v", err)
	}
	
	// Test lookup
	binding, exists := km.Lookup(seq)
	if !exists {
		t.Fatal("Binding should exist")
	}
	if binding.Command != "find-file" {
		t.Errorf("Expected command 'find-file', got '%s'", binding.Command)
	}
	
	// Test unbind
	err = km.Unbind(seq)
	if err != nil {
		t.Fatalf("Failed to unbind key: %v", err)
	}
	
	_, exists = km.Lookup(seq)
	if exists {
		t.Error("Binding should not exist after unbinding")
	}
}

func TestKeymapParent(t *testing.T) {
	parent := New("parent")
	child := New("child")
	
	// Bind in parent
	parentSeq := KeySequence{NewCtrlKey('a')}
	parent.Bind(parentSeq, "parent-command")
	
	// Bind in child
	childSeq := KeySequence{NewCtrlKey('b')}
	child.Bind(childSeq, "child-command")
	
	// Set parent
	child.SetParent(parent)
	
	// Test lookup in child (should find local binding)
	binding, exists := child.Lookup(childSeq)
	if !exists || binding.Command != "child-command" {
		t.Error("Should find child binding")
	}
	
	// Test lookup in child (should find parent binding)
	binding, exists = child.Lookup(parentSeq)
	if !exists || binding.Command != "parent-command" {
		t.Error("Should find parent binding")
	}
	
	// Test override: bind same key in child
	child.Bind(parentSeq, "overridden-command")
	binding, exists = child.Lookup(parentSeq)
	if !exists || binding.Command != "overridden-command" {
		t.Error("Child binding should override parent")
	}
}

func TestKeymapErrors(t *testing.T) {
	km := New("test")
	
	// Test empty key sequence
	err := km.Bind(KeySequence{}, "command")
	if err == nil {
		t.Error("Should error on empty key sequence")
	}
	
	// Test empty command
	seq := KeySequence{NewKey('a')}
	err = km.Bind(seq, "")
	if err == nil {
		t.Error("Should error on empty command")
	}
	
	// Test unbinding non-existent key
	err = km.Unbind(KeySequence{NewKey('z')})
	if err == nil {
		t.Error("Should error on unbinding non-existent key")
	}
}

func TestParseKeySequence(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"C-x", "C-x"},
		{"M-x", "M-x"},
		{"C-x C-f", "C-x C-f"},
		{"return", "return"},
		{"C-return", "C-return"},
		{"M-return", "M-return"},
		{"a", "a"},
	}
	
	for _, tc := range testCases {
		seq, err := ParseKeySequence(tc.input)
		if err != nil {
			t.Errorf("Failed to parse '%s': %v", tc.input, err)
			continue
		}
		
		result := seq.String()
		if result != tc.expected {
			t.Errorf("Parse '%s': expected '%s', got '%s'", tc.input, tc.expected, result)
		}
	}
}

func TestParseKeySequenceErrors(t *testing.T) {
	errorCases := []string{
		"",          // empty string
		"X-a",       // invalid modifier
		"C-",        // incomplete
	}
	
	for _, input := range errorCases {
		_, err := ParseKeySequence(input)
		if err == nil {
			t.Errorf("Should error on input '%s'", input)
		}
	}
}

func TestGetBindings(t *testing.T) {
	parent := New("parent")
	child := New("child")
	
	// Add bindings to parent
	parent.Bind(KeySequence{NewCtrlKey('a')}, "parent-a")
	parent.Bind(KeySequence{NewCtrlKey('b')}, "parent-b")
	
	// Add bindings to child
	child.Bind(KeySequence{NewCtrlKey('c')}, "child-c")
	child.Bind(KeySequence{NewCtrlKey('a')}, "child-a") // override parent
	
	child.SetParent(parent)
	
	// Test local bindings
	localBindings := child.GetLocalBindings()
	if len(localBindings) != 2 {
		t.Errorf("Expected 2 local bindings, got %d", len(localBindings))
	}
	
	// Test all bindings (including inherited)
	allBindings := child.GetAllBindings()
	if len(allBindings) != 3 { // child-c, child-a (overrides parent-a), parent-b
		t.Errorf("Expected 3 total bindings, got %d", len(allBindings))
	}
	
	// Verify override
	if allBindings["C-a"].Command != "child-a" {
		t.Error("Child binding should override parent binding")
	}
}