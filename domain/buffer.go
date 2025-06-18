package domain

import (
	"strings"
	"unicode/utf8"
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

type Buffer struct {
	name     string
	content  []string
	cursor   Position
	modified bool
}

type Position struct {
	Row int
	Col int
}

func NewBuffer(name string) *Buffer {
	return &Buffer{
		name:    name,
		content: []string{""},
		cursor:  Position{Row: 0, Col: 0},
	}
}

func (b *Buffer) Name() string {
	return b.name
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
	log.Debug("Inserting rune %c into line %q at byte pos %d", ch, line, b.cursor.Col)
	
	// Insert at byte position (cursor.Col is in bytes)
	newLine := line[:b.cursor.Col] + string(ch) + line[b.cursor.Col:]
	b.content[b.cursor.Row] = newLine
	
	// Move cursor by the byte length of the inserted character
	b.cursor.Col += utf8.RuneLen(ch)
	b.modified = true
	
	log.Debug("After insert: line=%q, cursor byte pos=%d", newLine, b.cursor.Col)
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