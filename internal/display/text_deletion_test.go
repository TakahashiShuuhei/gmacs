package display

import (
	"testing"
)

func TestDeleteChar(t *testing.T) {
	editor := NewEditor()
	
	// Set up test content
	buf := editor.currentWin.Buffer()
	buf.SetText("Hello World\nこんにちは\nLine 3")
	
	cursor := editor.currentWin.Cursor()
	
	testCases := []struct {
		name           string
		initialLine    int
		initialCol     int
		expectedText   string
		expectedLine   int
		expectedCol    int
		expectedMsg    string
	}{
		{
			"Delete char in middle",
			0, 1, // Position at 'e' in "Hello"
			"Hllo World\nこんにちは\nLine 3",
			0, 1, // Cursor stays at same position
			"",
		},
		{
			"Delete at end of line",
			0, 11, // After "Hello World" 
			"Hello Worldこんにちは\nLine 3", // Merges with next line
			0, 11, // Cursor stays at same position
			"",
		},
		{
			"Delete Japanese character",
			1, 0, // At 'こ'
			"Hello World\nんにちは\nLine 3",
			1, 0, // Cursor stays at same position
			"",
		},
		{
			"Delete at end of buffer",
			2, 6, // After "Line 3"
			"Hello World\nこんにちは\nLine 3", // No change at end of buffer
			2, 6, // Cursor stays at same position
			"End of buffer",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset buffer and cursor for each test
			buf.SetText("Hello World\nこんにちは\nLine 3")
			cursor.SetLine(tc.initialLine)
			cursor.SetCol(tc.initialCol)
			
			err := editor.deleteChar()
			if err != nil {
				t.Errorf("deleteChar failed: %v", err)
				return
			}
			
			if buf.GetText() != tc.expectedText {
				t.Errorf("Expected text %q, got %q", tc.expectedText, buf.GetText())
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

func TestBackwardDeleteChar(t *testing.T) {
	editor := NewEditor()
	
	// Set up test content
	buf := editor.currentWin.Buffer()
	buf.SetText("Hello World\nこんにちは\nLine 3")
	
	cursor := editor.currentWin.Cursor()
	
	testCases := []struct {
		name           string
		initialLine    int
		initialCol     int
		expectedText   string
		expectedLine   int
		expectedCol    int
		expectedMsg    string
	}{
		{
			"Backspace in middle",
			0, 2, // Position after 'e' in "Hello"
			"Hllo World\nこんにちは\nLine 3",
			0, 1, // Cursor moves back
			"",
		},
		{
			"Backspace at beginning of line",
			1, 0, // Beginning of second line
			"Hello Worldこんにちは\nLine 3", // Merges with previous line
			0, 11, // Cursor moves to end of previous line
			"",
		},
		{
			"Backspace Japanese character",
			1, 1, // After 'こ'
			"Hello World\nんにちは\nLine 3",
			1, 0, // Cursor moves back
			"",
		},
		{
			"Backspace at beginning of buffer",
			0, 0, // Very beginning
			"Hello World\nこんにちは\nLine 3", // No change
			0, 0, // Cursor stays same
			"Beginning of buffer",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset buffer and cursor for each test
			buf.SetText("Hello World\nこんにちは\nLine 3")
			cursor.SetLine(tc.initialLine)
			cursor.SetCol(tc.initialCol)
			
			err := editor.backwardDeleteChar()
			if err != nil {
				t.Errorf("backwardDeleteChar failed: %v", err)
				return
			}
			
			if buf.GetText() != tc.expectedText {
				t.Errorf("Expected text %q, got %q", tc.expectedText, buf.GetText())
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

func TestDeleteCharWithNewlines(t *testing.T) {
	editor := NewEditor()
	
	// Test deleting newlines specifically
	buf := editor.currentWin.Buffer()
	buf.SetText("Line1\nLine2\nLine3")
	
	cursor := editor.currentWin.Cursor()
	
	// Position at end of first line
	cursor.SetLine(0)
	cursor.SetCol(5) // After "Line1"
	
	err := editor.deleteChar()
	if err != nil {
		t.Errorf("deleteChar failed: %v", err)
		return
	}
	
	expectedText := "Line1Line2\nLine3"
	if buf.GetText() != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, buf.GetText())
	}
	
	// Cursor should stay at same position
	if cursor.Line() != 0 || cursor.Col() != 5 {
		t.Errorf("Expected cursor at (0, 5), got (%d, %d)", cursor.Line(), cursor.Col())
	}
}

func TestBackwardDeleteCharWithNewlines(t *testing.T) {
	editor := NewEditor()
	
	// Test deleting newlines specifically
	buf := editor.currentWin.Buffer()
	buf.SetText("Line1\nLine2\nLine3")
	
	cursor := editor.currentWin.Cursor()
	
	// Position at beginning of second line
	cursor.SetLine(1)
	cursor.SetCol(0) // Beginning of "Line2"
	
	err := editor.backwardDeleteChar()
	if err != nil {
		t.Errorf("backwardDeleteChar failed: %v", err)
		return
	}
	
	expectedText := "Line1Line2\nLine3"
	if buf.GetText() != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, buf.GetText())
	}
	
	// Cursor should move to end of previous line
	if cursor.Line() != 0 || cursor.Col() != 5 {
		t.Errorf("Expected cursor at (0, 5), got (%d, %d)", cursor.Line(), cursor.Col())
	}
}