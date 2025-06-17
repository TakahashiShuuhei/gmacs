package display

import (
	"fmt"
)

// forwardChar moves cursor forward one character (C-f)
func (e *Editor) forwardChar() error {
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	line := buffer.GetLine(cursor.Line())
	lineRunes := []rune(line)
	
	// Check if we can move forward
	if cursor.Col() < len(lineRunes) {
		cursor.SetCol(cursor.Col() + 1)
		e.minibuffer.ShowMessage("")
	} else {
		// At end of line, try to move to beginning of next line
		if cursor.Line() < buffer.LineCount()-1 {
			cursor.SetLine(cursor.Line() + 1)
			cursor.SetCol(0)
			e.currentWin.EnsureCursorVisible()
			e.minibuffer.ShowMessage("")
		} else {
			e.minibuffer.ShowMessage("End of buffer")
		}
	}
	
	return nil
}

// backwardChar moves cursor backward one character (C-b)
func (e *Editor) backwardChar() error {
	cursor := e.currentWin.Cursor()
	
	// Check if we can move backward
	if cursor.Col() > 0 {
		cursor.SetCol(cursor.Col() - 1)
		e.minibuffer.ShowMessage("")
	} else {
		// At beginning of line, try to move to end of previous line
		if cursor.Line() > 0 {
			buffer := e.currentWin.Buffer()
			prevLine := buffer.GetLine(cursor.Line() - 1)
			prevLineRunes := []rune(prevLine)
			
			cursor.SetLine(cursor.Line() - 1)
			cursor.SetCol(len(prevLineRunes))
			e.currentWin.EnsureCursorVisible()
			e.minibuffer.ShowMessage("")
		} else {
			e.minibuffer.ShowMessage("Beginning of buffer")
		}
	}
	
	return nil
}

// nextLine moves cursor to next line (C-n)
func (e *Editor) nextLine() error {
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	// Check if there is a next line
	if cursor.Line() < buffer.LineCount()-1 {
		nextLineNum := cursor.Line() + 1
		nextLine := buffer.GetLine(nextLineNum)
		nextLineRunes := []rune(nextLine)
		
		cursor.SetLine(nextLineNum)
		
		// Try to maintain column position, but clamp to line length
		if cursor.Col() > len(nextLineRunes) {
			cursor.SetCol(len(nextLineRunes))
		}
		
		e.currentWin.EnsureCursorVisible()
		e.minibuffer.ShowMessage("")
	} else {
		e.minibuffer.ShowMessage("End of buffer")
	}
	
	return nil
}

// previousLine moves cursor to previous line (C-p)
func (e *Editor) previousLine() error {
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	// Check if there is a previous line
	if cursor.Line() > 0 {
		prevLineNum := cursor.Line() - 1
		prevLine := buffer.GetLine(prevLineNum)
		prevLineRunes := []rune(prevLine)
		
		cursor.SetLine(prevLineNum)
		
		// Try to maintain column position, but clamp to line length
		if cursor.Col() > len(prevLineRunes) {
			cursor.SetCol(len(prevLineRunes))
		}
		
		e.currentWin.EnsureCursorVisible()
		e.minibuffer.ShowMessage("")
	} else {
		e.minibuffer.ShowMessage("Beginning of buffer")
	}
	
	return nil
}

// selfInsertCommand inserts a character at the current cursor position
func (e *Editor) selfInsertCommand(char rune) error {
	// Get current buffer and cursor
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	// Special handling for newline character
	if char == '\n' {
		return e.insertNewline()
	}
	
	// Insert the character at cursor position
	err := buffer.InsertChar(cursor.Line(), cursor.Col(), char)
	if err != nil {
		return fmt.Errorf("failed to insert character: %v", err)
	}
	
	// Move cursor forward
	oldCol := cursor.Col()
	newCol := oldCol + 1
	cursor.SetCol(newCol)
	
	// Clear any previous message
	e.minibuffer.ShowMessage("")
	
	return nil
}

// insertNewline inserts a newline and moves cursor to next line
func (e *Editor) insertNewline() error {
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	currentLine := cursor.Line()
	currentCol := cursor.Col()
	
	// Get the current line content
	if currentLine >= buffer.LineCount() {
		// Add a new empty line
		buffer.InsertLine(currentLine, "")
		cursor.SetLine(currentLine + 1)
		cursor.SetCol(0)
		return nil
	}
	
	lineContent := buffer.GetLine(currentLine)
	lineRunes := []rune(lineContent)
	
	// Split the line at cursor position
	if currentCol >= len(lineRunes) {
		// Cursor is at end of line, just add new line
		buffer.InsertLine(currentLine + 1, "")
	} else {
		// Split the line
		leftPart := string(lineRunes[:currentCol])
		rightPart := string(lineRunes[currentCol:])
		
		// Update current line with left part
		buffer.SetLine(currentLine, leftPart)
		
		// Insert new line with right part
		buffer.InsertLine(currentLine + 1, rightPart)
	}
	
	// Move cursor to next line, column 0
	cursor.SetLine(currentLine + 1)
	cursor.SetCol(0)
	
	// Clear any previous message
	e.minibuffer.ShowMessage("")
	
	return nil
}

// selfInsertStringCommand inserts a string at the current cursor position
func (e *Editor) selfInsertStringCommand(text string) error {
	// Get current buffer and cursor
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	// Insert the string at cursor position
	err := buffer.InsertString(cursor.Line(), cursor.Col(), text)
	if err != nil {
		return fmt.Errorf("failed to insert string: %v", err)
	}
	
	// Move cursor forward by the number of characters inserted
	textRunes := []rune(text)
	oldCol := cursor.Col()
	newCol := oldCol + len(textRunes)
	cursor.SetCol(newCol)
	
	// Clear any previous message
	e.minibuffer.ShowMessage("")
	
	return nil
}

// deleteChar deletes the character at the current cursor position (C-d)
func (e *Editor) deleteChar() error {
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	line := buffer.GetLine(cursor.Line())
	lineRunes := []rune(line)
	
	// Check if cursor is at end of line
	if cursor.Col() >= len(lineRunes) {
		// At end of line, try to merge with next line
		if cursor.Line() < buffer.LineCount()-1 {
			nextLine := buffer.GetLine(cursor.Line() + 1)
			
			// Merge current line with next line
			newLine := line + nextLine
			err := buffer.SetLine(cursor.Line(), newLine)
			if err != nil {
				return fmt.Errorf("failed to merge lines: %v", err)
			}
			
			// Delete the next line
			err = buffer.DeleteLine(cursor.Line() + 1)
			if err != nil {
				return fmt.Errorf("failed to delete line: %v", err)
			}
			
			e.minibuffer.ShowMessage("")
		} else {
			e.minibuffer.ShowMessage("End of buffer")
		}
	} else {
		// Delete character at cursor position
		newRunes := make([]rune, len(lineRunes)-1)
		copy(newRunes[:cursor.Col()], lineRunes[:cursor.Col()])
		copy(newRunes[cursor.Col():], lineRunes[cursor.Col()+1:])
		
		newLine := string(newRunes)
		err := buffer.SetLine(cursor.Line(), newLine)
		if err != nil {
			return fmt.Errorf("failed to delete character: %v", err)
		}
		
		e.minibuffer.ShowMessage("")
	}
	
	return nil
}

// backwardDeleteChar deletes the character before the cursor (backspace)
func (e *Editor) backwardDeleteChar() error {
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
	// Check if cursor is at beginning of line
	if cursor.Col() == 0 {
		// At beginning of line, try to merge with previous line
		if cursor.Line() > 0 {
			prevLine := buffer.GetLine(cursor.Line() - 1)
			currentLine := buffer.GetLine(cursor.Line())
			prevLineRunes := []rune(prevLine)
			
			// Merge previous line with current line
			newLine := prevLine + currentLine
			err := buffer.SetLine(cursor.Line()-1, newLine)
			if err != nil {
				return fmt.Errorf("failed to merge lines: %v", err)
			}
			
			// Delete the current line
			err = buffer.DeleteLine(cursor.Line())
			if err != nil {
				return fmt.Errorf("failed to delete line: %v", err)
			}
			
			// Move cursor to end of previous line
			cursor.SetLine(cursor.Line() - 1)
			cursor.SetCol(len(prevLineRunes))
			e.currentWin.EnsureCursorVisible()
			
			e.minibuffer.ShowMessage("")
		} else {
			e.minibuffer.ShowMessage("Beginning of buffer")
		}
	} else {
		// Delete character before cursor position
		line := buffer.GetLine(cursor.Line())
		lineRunes := []rune(line)
		
		newRunes := make([]rune, len(lineRunes)-1)
		copy(newRunes[:cursor.Col()-1], lineRunes[:cursor.Col()-1])
		copy(newRunes[cursor.Col()-1:], lineRunes[cursor.Col():])
		
		newLine := string(newRunes)
		err := buffer.SetLine(cursor.Line(), newLine)
		if err != nil {
			return fmt.Errorf("failed to delete character: %v", err)
		}
		
		// Move cursor backward
		cursor.SetCol(cursor.Col() - 1)
		
		e.minibuffer.ShowMessage("")
	}
	
	return nil
}