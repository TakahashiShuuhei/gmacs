package domain

import (
	"bufio"
	"os"
	"strings"
	"unicode/utf8"
)

type Buffer struct {
	name     string
	content  []string
	cursor   Position
	modified bool
	filepath string // File path if buffer is associated with a file
}

type Position struct {
	Row int
	Col int
}

func NewBuffer(name string) *Buffer {
	return &Buffer{
		name:     name,
		content:  []string{""},
		cursor:   Position{Row: 0, Col: 0},
		filepath: "",
	}
}

// NewBufferFromFile creates a new buffer and loads content from a file
func NewBufferFromFile(filepath string) (*Buffer, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	// If file is empty, ensure at least one empty line
	if len(lines) == 0 {
		lines = []string{""}
	}
	
	// Extract filename from path for buffer name
	name := filepath
	if lastSlash := strings.LastIndex(filepath, "/"); lastSlash != -1 {
		name = filepath[lastSlash+1:]
	}
	
	return &Buffer{
		name:     name,
		content:  lines,
		cursor:   Position{Row: 0, Col: 0},
		modified: false,
		filepath: filepath,
	}, nil
}

func (b *Buffer) Name() string {
	return b.name
}

func (b *Buffer) Filepath() string {
	return b.filepath
}

func (b *Buffer) Content() []string {
	return b.content
}

func (b *Buffer) Cursor() Position {
	return b.cursor
}

func (b *Buffer) SetCursor(pos Position) {
	if pos.Row < 0 {
		pos.Row = 0
	}
	if pos.Row >= len(b.content) {
		pos.Row = len(b.content) - 1
	}
	if pos.Col < 0 {
		pos.Col = 0
	}
	if pos.Col > len(b.content[pos.Row]) {
		pos.Col = len(b.content[pos.Row])
	}
	b.cursor = pos
}

func (b *Buffer) InsertChar(ch rune) {
	if ch == '\n' {
		b.insertNewline()
		return
	}
	
	line := b.content[b.cursor.Row]
	
	// Insert at byte position (cursor.Col is in bytes)
	newLine := line[:b.cursor.Col] + string(ch) + line[b.cursor.Col:]
	b.content[b.cursor.Row] = newLine
	
	// Move cursor by the byte length of the inserted character
	b.cursor.Col += utf8.RuneLen(ch)
	b.modified = true
	
}

func (b *Buffer) insertNewline() {
	line := b.content[b.cursor.Row]
	beforeCursor := line[:b.cursor.Col]
	afterCursor := line[b.cursor.Col:]
	
	b.content[b.cursor.Row] = beforeCursor
	
	newContent := make([]string, 0, len(b.content)+1)
	newContent = append(newContent, b.content[:b.cursor.Row+1]...)
	newContent = append(newContent, afterCursor)
	newContent = append(newContent, b.content[b.cursor.Row+1:]...)
	
	b.content = newContent
	b.cursor.Row++
	b.cursor.Col = 0
	b.modified = true
}

func (b *Buffer) InsertString(s string) {
	lines := strings.Split(s, "\n")
	if len(lines) == 1 {
		b.InsertChar(rune(s[0]))
		return
	}
	
	currentLine := b.content[b.cursor.Row]
	beforeCursor := currentLine[:b.cursor.Col]
	afterCursor := currentLine[b.cursor.Col:]
	
	lines[0] = beforeCursor + lines[0]
	lines[len(lines)-1] = lines[len(lines)-1] + afterCursor
	
	newContent := make([]string, 0, len(b.content)+len(lines)-1)
	newContent = append(newContent, b.content[:b.cursor.Row]...)
	newContent = append(newContent, lines...)
	newContent = append(newContent, b.content[b.cursor.Row+1:]...)
	
	b.content = newContent
	b.cursor.Row += len(lines) - 1
	b.cursor.Col = len(lines[len(lines)-1]) - len(afterCursor)
	b.modified = true
}

func (b *Buffer) Clear() {
	b.content = []string{""}
	b.cursor = Position{Row: 0, Col: 0}
	b.modified = true
}

// DeleteBackward deletes the character before the cursor (backspace)
func (b *Buffer) DeleteBackward() {
	if b.cursor.Row == 0 && b.cursor.Col == 0 {
		// At beginning of buffer, nothing to delete
		return
	}
	
	line := b.content[b.cursor.Row]
	
	if b.cursor.Col == 0 {
		// At beginning of line, join with previous line
		if b.cursor.Row > 0 {
			prevLine := b.content[b.cursor.Row-1]
			
			// Join lines
			b.content[b.cursor.Row-1] = prevLine + line
			
			// Remove current line
			newContent := make([]string, 0, len(b.content)-1)
			newContent = append(newContent, b.content[:b.cursor.Row]...)
			newContent = append(newContent, b.content[b.cursor.Row+1:]...)
			b.content = newContent
			
			// Move cursor to end of previous line
			b.cursor.Row--
			b.cursor.Col = len(prevLine) // Use byte position for cursor
			b.modified = true
		}
	} else {
		// Delete character before cursor
		runes := []rune(line)
		if len(runes) > 0 {
			// Find the rune position before cursor
			bytePos := 0
			runeIndex := 0
			for bytePos < b.cursor.Col && runeIndex < len(runes) {
				bytePos += utf8.RuneLen(runes[runeIndex])
				runeIndex++
			}
			
			if runeIndex > 0 {
				// Remove the previous rune
				newRunes := append(runes[:runeIndex-1], runes[runeIndex:]...)
				b.content[b.cursor.Row] = string(newRunes)
				
				// Move cursor back by the byte length of the deleted rune
				deletedRune := runes[runeIndex-1]
				b.cursor.Col -= utf8.RuneLen(deletedRune)
				b.modified = true
			}
		}
	}
}

// DeleteForward deletes the character at the cursor position (delete)
func (b *Buffer) DeleteForward() {
	if b.cursor.Row >= len(b.content) {
		return
	}
	
	line := b.content[b.cursor.Row]
	
	if b.cursor.Col >= len(line) {
		// At end of line, join with next line
		if b.cursor.Row < len(b.content)-1 {
			nextLine := b.content[b.cursor.Row+1]
			
			// Join lines
			b.content[b.cursor.Row] = line + nextLine
			
			// Remove next line
			newContent := make([]string, 0, len(b.content)-1)
			newContent = append(newContent, b.content[:b.cursor.Row+1]...)
			newContent = append(newContent, b.content[b.cursor.Row+2:]...)
			b.content = newContent
			b.modified = true
		}
	} else {
		// Delete character at cursor
		runes := []rune(line)
		if len(runes) > 0 {
			// Find the rune position at cursor
			bytePos := 0
			runeIndex := 0
			for bytePos < b.cursor.Col && runeIndex < len(runes) {
				bytePos += utf8.RuneLen(runes[runeIndex])
				runeIndex++
			}
			
			if runeIndex < len(runes) {
				// Remove the rune at cursor position
				newRunes := append(runes[:runeIndex], runes[runeIndex+1:]...)
				b.content[b.cursor.Row] = string(newRunes)
				b.modified = true
			}
		}
	}
}