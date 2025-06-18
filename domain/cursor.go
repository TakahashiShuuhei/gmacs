package domain

import (
	"unicode/utf8"
	"github.com/TakahashiShuuhei/gmacs/core/util"
)

// Cursor movement interactive functions following Emacs conventions

// ForwardChar moves cursor forward by one character (C-f)
func ForwardChar(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	
	cursor := buffer.Cursor()
	content := buffer.Content()
	
	// Check if we're at the end of current line
	if cursor.Row < len(content) {
		line := content[cursor.Row]
		if cursor.Col < len(line) {
			// Move within current line
			runes := []rune(line[cursor.Col:])
			if len(runes) > 0 {
				newCol := cursor.Col + utf8.RuneLen(runes[0])
				buffer.SetCursor(Position{Row: cursor.Row, Col: newCol})
				EnsureCursorVisible(editor)
			}
		} else if cursor.Row < len(content)-1 {
			// Move to beginning of next line
			buffer.SetCursor(Position{Row: cursor.Row + 1, Col: 0})
			EnsureCursorVisible(editor)
		}
	}
	
	return nil
}

// BackwardChar moves cursor backward by one character (C-b)
func BackwardChar(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	
	cursor := buffer.Cursor()
	content := buffer.Content()
	
	if cursor.Col > 0 {
		// Move within current line
		line := content[cursor.Row]
		beforeCursor := line[:cursor.Col]
		runes := []rune(beforeCursor)
		if len(runes) > 0 {
			lastRune := runes[len(runes)-1]
			newCol := cursor.Col - utf8.RuneLen(lastRune)
			buffer.SetCursor(Position{Row: cursor.Row, Col: newCol})
			EnsureCursorVisible(editor)
		}
	} else if cursor.Row > 0 {
		// Move to end of previous line
		prevLine := content[cursor.Row-1]
		buffer.SetCursor(Position{Row: cursor.Row - 1, Col: len(prevLine)})
		EnsureCursorVisible(editor)
	}
	
	return nil
}

// NextLine moves cursor to next line (C-n)
func NextLine(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	
	cursor := buffer.Cursor()
	content := buffer.Content()
	
	if cursor.Row < len(content)-1 {
		// Calculate target column in display width
		currentLine := content[cursor.Row]
		targetDisplayCol := calculateDisplayColumn(currentLine, cursor.Col)
		
		// Move to next line and find corresponding byte position
		nextLine := content[cursor.Row+1]
		newCol := findBytePositionFromDisplay(nextLine, targetDisplayCol)
		
		buffer.SetCursor(Position{Row: cursor.Row + 1, Col: newCol})
		EnsureCursorVisible(editor)
	}
	
	return nil
}

// PreviousLine moves cursor to previous line (C-p)
func PreviousLine(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	
	cursor := buffer.Cursor()
	content := buffer.Content()
	
	if cursor.Row > 0 {
		// Calculate target column in display width
		currentLine := content[cursor.Row]
		targetDisplayCol := calculateDisplayColumn(currentLine, cursor.Col)
		
		// Move to previous line and find corresponding byte position
		prevLine := content[cursor.Row-1]
		newCol := findBytePositionFromDisplay(prevLine, targetDisplayCol)
		
		buffer.SetCursor(Position{Row: cursor.Row - 1, Col: newCol})
		EnsureCursorVisible(editor)
	}
	
	return nil
}

// BeginningOfLine moves cursor to beginning of current line (C-a)
func BeginningOfLine(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	
	cursor := buffer.Cursor()
	buffer.SetCursor(Position{Row: cursor.Row, Col: 0})
	EnsureCursorVisible(editor)
	
	return nil
}

// EndOfLine moves cursor to end of current line (C-e)
func EndOfLine(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		return nil
	}
	
	cursor := buffer.Cursor()
	content := buffer.Content()
	
	if cursor.Row < len(content) {
		line := content[cursor.Row]
		buffer.SetCursor(Position{Row: cursor.Row, Col: len(line)})
		EnsureCursorVisible(editor)
	}
	
	return nil
}

// Helper functions

// calculateDisplayColumn calculates the display width from byte position
func calculateDisplayColumn(line string, bytePos int) int {
	if bytePos >= len(line) {
		return util.StringWidth(line)
	}
	return util.StringWidth(line[:bytePos])
}

// findBytePositionFromDisplay finds the byte position from display width
func findBytePositionFromDisplay(line string, targetWidth int) int {
	if targetWidth <= 0 {
		return 0
	}
	
	runes := []rune(line)
	currentWidth := 0
	
	for i, r := range runes {
		charWidth := util.RuneWidth(r)
		if currentWidth + charWidth > targetWidth {
			// Return position just before this character
			return len(string(runes[:i]))
		}
		currentWidth += charWidth
	}
	
	// Target width is beyond the line
	return len(line)
}