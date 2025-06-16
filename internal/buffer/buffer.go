package buffer

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Buffer represents a text buffer (similar to Emacs buffer)
type Buffer struct {
	name     string
	filename string
	content  []string // lines of text
	modified bool
	readOnly bool
	created  time.Time
	lastMod  time.Time
}

// New creates a new buffer with the given name
func New(name string) *Buffer {
	now := time.Now()
	return &Buffer{
		name:    name,
		content: []string{""},
		created: now,
		lastMod: now,
	}
}

// NewFromFile creates a new buffer from a file
func NewFromFile(filename string) *Buffer {
	now := time.Now()
	return &Buffer{
		name:     filename,
		filename: filename,
		content:  []string{""},
		created:  now,
		lastMod:  now,
	}
}

// Name returns the buffer name
func (b *Buffer) Name() string {
	return b.name
}

// Filename returns the associated filename
func (b *Buffer) Filename() string {
	return b.filename
}

// SetFilename sets the associated filename
func (b *Buffer) SetFilename(filename string) {
	b.filename = filename
	b.markModified()
}

// LineCount returns the number of lines in the buffer
func (b *Buffer) LineCount() int {
	return len(b.content)
}

// GetLine returns the content of the specified line (0-indexed)
func (b *Buffer) GetLine(lineNum int) string {
	if lineNum < 0 || lineNum >= len(b.content) {
		return ""
	}
	return b.content[lineNum]
}

// SetLine sets the content of the specified line
func (b *Buffer) SetLine(lineNum int, text string) error {
	if lineNum < 0 || lineNum >= len(b.content) {
		return fmt.Errorf("line %d does not exist", lineNum)
	}
	b.content[lineNum] = text
	b.markModified()
	return nil
}

// InsertChar inserts a character at the specified position (UTF-8 compatible)
func (b *Buffer) InsertChar(lineNum, colNum int, char rune) error {
	// Ensure line exists
	if lineNum < 0 || lineNum >= len(b.content) {
		return fmt.Errorf("line %d does not exist", lineNum)
	}
	
	line := b.content[lineNum]
	runes := []rune(line) // Convert to rune slice for proper UTF-8 handling
	
	// Ensure column position is valid
	if colNum < 0 {
		colNum = 0
	}
	if colNum > len(runes) {
		colNum = len(runes)
	}
	
	// Insert the character
	newRunes := make([]rune, len(runes)+1)
	copy(newRunes[:colNum], runes[:colNum])
	newRunes[colNum] = char
	copy(newRunes[colNum+1:], runes[colNum:])
	
	// Convert back to string and update the line
	b.content[lineNum] = string(newRunes)
	b.markModified()
	
	return nil
}

// InsertString inserts a string at the specified position (UTF-8 compatible)
func (b *Buffer) InsertString(lineNum, colNum int, text string) error {
	// Ensure line exists
	if lineNum < 0 || lineNum >= len(b.content) {
		return fmt.Errorf("line %d does not exist", lineNum)
	}
	
	line := b.content[lineNum]
	runes := []rune(line) // Convert to rune slice for proper UTF-8 handling
	textRunes := []rune(text) // Convert input text to runes
	
	// Ensure column position is valid
	if colNum < 0 {
		colNum = 0
	}
	if colNum > len(runes) {
		colNum = len(runes)
	}
	
	// Insert the string
	newRunes := make([]rune, len(runes)+len(textRunes))
	copy(newRunes[:colNum], runes[:colNum])
	copy(newRunes[colNum:colNum+len(textRunes)], textRunes)
	copy(newRunes[colNum+len(textRunes):], runes[colNum:])
	
	// Convert back to string and update the line
	b.content[lineNum] = string(newRunes)
	b.markModified()
	
	return nil
}

// InsertLine inserts a new line at the specified position
func (b *Buffer) InsertLine(lineNum int, text string) {
	if lineNum < 0 {
		lineNum = 0
	}
	if lineNum > len(b.content) {
		lineNum = len(b.content)
	}
	
	// Insert new line
	newContent := make([]string, len(b.content)+1)
	copy(newContent[:lineNum], b.content[:lineNum])
	newContent[lineNum] = text
	copy(newContent[lineNum+1:], b.content[lineNum:])
	b.content = newContent
	b.markModified()
}

// DeleteLine deletes the specified line
func (b *Buffer) DeleteLine(lineNum int) error {
	if lineNum < 0 || lineNum >= len(b.content) {
		return fmt.Errorf("line %d does not exist", lineNum)
	}
	
	// Don't delete the last line if it's the only one
	if len(b.content) == 1 {
		b.content[0] = ""
		b.markModified()
		return nil
	}
	
	newContent := make([]string, len(b.content)-1)
	copy(newContent[:lineNum], b.content[:lineNum])
	copy(newContent[lineNum:], b.content[lineNum+1:])
	b.content = newContent
	b.markModified()
	return nil
}

// GetText returns all text in the buffer
func (b *Buffer) GetText() string {
	return strings.Join(b.content, "\n")
}

// SetText sets the entire buffer conten
func (b *Buffer) SetText(text string) {
	if text == "" {
		b.content = []string{""}
	} else {
		b.content = strings.Split(text, "\n")
	}
	b.markModified()
}

// IsModified returns whether the buffer has been modified
func (b *Buffer) IsModified() bool {
	return b.modified
}

// SetModified sets the modified flag
func (b *Buffer) SetModified(modified bool) {
	b.modified = modified
	if modified {
		b.lastMod = time.Now()
	}
}

// IsReadOnly returns whether the buffer is read-only
func (b *Buffer) IsReadOnly() bool {
	return b.readOnly
}

// SetReadOnly sets the read-only flag
func (b *Buffer) SetReadOnly(readOnly bool) {
	b.readOnly = readOnly
}

// markModified marks the buffer as modified
func (b *Buffer) markModified() {
	if !b.readOnly {
		b.modified = true
		b.lastMod = time.Now()
	}
}

// LoadFromFile loads the buffer content from a file
func (b *Buffer) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()
	
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	
	// Set the buffer content and metadata
	b.SetText(string(content))
	b.filename = filename
	b.name = filename
	b.modified = false // File just loaded, not modified
	
	return nil
}

// SaveToFile saves the buffer content to a file
func (b *Buffer) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer file.Close()
	
	content := b.GetText()
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filename, err)
	}
	
	// Update metadata
	b.filename = filename
	b.name = filename
	b.modified = false // File just saved, not modified
	
	return nil
}

// Save saves the buffer to its associated file
func (b *Buffer) Save() error {
	if b.filename == "" {
		return fmt.Errorf("no filename associated with buffer %s", b.name)
	}
	
	return b.SaveToFile(b.filename)
}