package display

import (
	"testing"
)

func TestCursorMovementCommands(t *testing.T) {
	editor := NewEditor()
	
	// Set up test content
	buf := editor.currentWin.Buffer()
	buf.SetText("Hello World\nこんにちは世界\nLine 3")
	
	cursor := editor.currentWin.Cursor()
	cursor.SetLine(0)
	cursor.SetCol(0)
	
	testCases := []struct {
		name           string
		command        func() error
		expectedLine   int
		expectedCol    int
		expectedMsg    string
	}{
		{"forward-char from start", editor.forwardChar, 0, 1, ""},
		{"forward-char again", editor.forwardChar, 0, 2, ""},
		{"backward-char", editor.backwardChar, 0, 1, ""},
		{"backward-char to start", editor.backwardChar, 0, 0, ""},
		{"next-line", editor.nextLine, 1, 0, ""},
		{"forward-char on Japanese", editor.forwardChar, 1, 1, ""}, // こ
		{"forward-char on Japanese", editor.forwardChar, 1, 2, ""}, // ん
		{"previous-line", editor.previousLine, 0, 2, ""}, // Back to line 0, maintain col
		{"next-line twice", func() error { editor.nextLine(); return editor.nextLine() }, 2, 2, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.command()
			if err != nil {
				t.Errorf("Command failed: %v", err)
				return
			}
			
			if cursor.Line() != tc.expectedLine {
				t.Errorf("Expected line %d, got %d", tc.expectedLine, cursor.Line())
			}
			
			if cursor.Col() != tc.expectedCol {
				t.Errorf("Expected col %d, got %d", tc.expectedCol, cursor.Col())
			}
		})
	}
}

func TestCursorMovementBoundaries(t *testing.T) {
	editor := NewEditor()
	
	// Set up single line content
	buf := editor.currentWin.Buffer()
	buf.SetText("Hello")
	
	cursor := editor.currentWin.Cursor()
	cursor.SetLine(0)
	cursor.SetCol(0)
	
	// Test moving backward at beginning
	err := editor.backwardChar()
	if err != nil {
		t.Errorf("backwardChar failed: %v", err)
	}
	
	// Should still be at beginning
	if cursor.Line() != 0 || cursor.Col() != 0 {
		t.Errorf("Cursor should stay at beginning, got (%d, %d)", cursor.Line(), cursor.Col())
	}
	
	// Move to end
	cursor.SetCol(5) // After "Hello"
	
	// Test moving forward at end
	err = editor.forwardChar()
	if err != nil {
		t.Errorf("forwardChar failed: %v", err)
	}
	
	// Should still be at end
	if cursor.Line() != 0 || cursor.Col() != 5 {
		t.Errorf("Cursor should stay at end, got (%d, %d)", cursor.Line(), cursor.Col())
	}
}

func TestCursorMovementAcrossLines(t *testing.T) {
	editor := NewEditor()
	
	// Set up multi-line content
	buf := editor.currentWin.Buffer()
	buf.SetText("Short\nVery long line here\nEnd")
	
	cursor := editor.currentWin.Cursor()
	
	// Start at end of first line
	cursor.SetLine(0)
	cursor.SetCol(5) // After "Short"
	
	// Move forward should go to next line
	err := editor.forwardChar()
	if err != nil {
		t.Errorf("forwardChar failed: %v", err)
	}
	
	if cursor.Line() != 1 || cursor.Col() != 0 {
		t.Errorf("Expected (1, 0), got (%d, %d)", cursor.Line(), cursor.Col())
	}
	
	// Move backward should go to previous line end
	err = editor.backwardChar()
	if err != nil {
		t.Errorf("backwardChar failed: %v", err)
	}
	
	if cursor.Line() != 0 || cursor.Col() != 5 {
		t.Errorf("Expected (0, 5), got (%d, %d)", cursor.Line(), cursor.Col())
	}
}

func TestCursorMovementWithJapanese(t *testing.T) {
	editor := NewEditor()
	
	// Set up Japanese content
	buf := editor.currentWin.Buffer()
	buf.SetText("あいう\nかきく")
	
	cursor := editor.currentWin.Cursor()
	cursor.SetLine(0)
	cursor.SetCol(0)
	
	// Test forward movement through Japanese characters
	moves := []struct {
		expectedLine int
		expectedCol  int
	}{
		{0, 1}, // あ → い
		{0, 2}, // い → う
		{0, 3}, // う → end of line
		{1, 0}, // next line start
		{1, 1}, // か → き
	}
	
	for i, move := range moves {
		err := editor.forwardChar()
		if err != nil {
			t.Errorf("Move %d failed: %v", i, err)
			continue
		}
		
		if cursor.Line() != move.expectedLine || cursor.Col() != move.expectedCol {
			t.Errorf("Move %d: expected (%d, %d), got (%d, %d)", 
				i, move.expectedLine, move.expectedCol, cursor.Line(), cursor.Col())
		}
	}
}