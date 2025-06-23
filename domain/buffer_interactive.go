package domain

import (
	"fmt"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/events"
	"github.com/TakahashiShuuhei/gmacs/log"
)

// MinibufferBufferSelection represents buffer selection mode
const MinibufferBufferSelection MinibufferMode = 4

// Interactive buffer functions

// SwitchToBufferInteractive implements C-x b (switch-to-buffer)
func SwitchToBufferInteractive(e *Editor) error {
	// Start buffer selection in minibuffer
	e.minibuffer.mode = MinibufferBufferSelection
	e.minibuffer.content = ""
	e.minibuffer.prompt = "Switch to buffer: "
	e.minibuffer.message = ""
	e.minibuffer.cursor = 0
	
	log.Info("Started interactive buffer switch")
	return nil
}

// ListBuffersInteractive implements C-x C-b (list-buffers)
func ListBuffersInteractive(e *Editor) error {
	// Create or find *Buffer List* buffer
	bufferListName := "*Buffer List*"
	bufferListBuffer := e.FindBuffer(bufferListName)
	if bufferListBuffer == nil {
		bufferListBuffer = NewBuffer(bufferListName)
		e.AddBuffer(bufferListBuffer)
	}
	
	// Clear the buffer and populate with buffer list
	bufferListBuffer.Clear()
	
	// Create the buffer list content in Emacs format
	content := e.formatBufferList()
	for _, line := range content {
		for _, ch := range line {
			bufferListBuffer.InsertChar(ch)
		}
		bufferListBuffer.InsertChar('\n')
	}
	
	// Switch to the *Buffer List* buffer
	e.SwitchToBuffer(bufferListBuffer)
	
	// Move cursor to beginning
	bufferListBuffer.SetCursor(Position{Row: 0, Col: 0})
	
	log.Info("Listed buffers in *Buffer List* buffer")
	return nil
}

// KillBufferInteractive implements C-x k (kill-buffer)
func KillBufferInteractive(e *Editor) error {
	currentBuffer := e.CurrentBuffer()
	if currentBuffer == nil {
		e.minibuffer.SetMessage("No buffer to kill")
		return nil
	}
	
	// Don't kill the last buffer
	if len(e.buffers) <= 1 {
		e.minibuffer.SetMessage("Cannot kill the last buffer")
		return nil
	}
	
	// Remove the buffer from the list
	bufferName := currentBuffer.Name()
	for i, buffer := range e.buffers {
		if buffer == currentBuffer {
			e.buffers = append(e.buffers[:i], e.buffers[i+1:]...)
			break
		}
	}
	
	// Switch to the first remaining buffer
	if len(e.buffers) > 0 {
		e.SwitchToBuffer(e.buffers[0])
	}
	
	e.minibuffer.SetMessage("Killed buffer: " + bufferName)
	log.Info("Killed buffer: %s", bufferName)
	return nil
}

// GetOrCreateBuffer finds an existing buffer or creates a new one
func (e *Editor) GetOrCreateBuffer(name string) *Buffer {
	// Try to find existing buffer
	for _, buffer := range e.buffers {
		if buffer.Name() == name {
			return buffer
		}
	}
	
	// Create new buffer
	buffer := NewBuffer(name)
	e.AddBuffer(buffer)
	log.Info("Created new buffer: %s", name)
	return buffer
}

// GetBufferNames returns a list of all buffer names
func (e *Editor) GetBufferNames() []string {
	names := make([]string, len(e.buffers))
	for i, buffer := range e.buffers {
		names[i] = buffer.Name()
	}
	return names
}

// HandleBufferSelectionInput handles input during buffer selection
func (e *Editor) HandleBufferSelectionInput(event events.KeyEventData) {
	// Handle Enter - switch to the buffer
	if event.Key == "Enter" || event.Key == "Return" {
		bufferName := e.minibuffer.Content()
		log.Info("Switching to buffer: %s", bufferName)
		
		if bufferName == "" {
			// If empty, stay with current buffer
			e.minibuffer.Clear()
			return
		}
		
		// Find or create buffer
		buffer := e.GetOrCreateBuffer(bufferName)
		e.SwitchToBuffer(buffer)
		
		e.minibuffer.SetMessage("Switched to buffer: " + bufferName)
		return
	}
	
	// Handle Escape - cancel buffer selection
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.minibuffer.Clear()
		return
	}
	
	// Handle Tab - buffer name completion
	if event.Key == "Tab" || event.Key == "\t" {
		e.handleBufferCompletion()
		return
	}
	
	// Handle Ctrl commands in minibuffer (same as other minibuffer modes)
	if event.Ctrl {
		switch event.Key {
		case "h":
			e.minibuffer.DeleteBackward()
			return
		case "d":
			e.minibuffer.DeleteForward()
			return
		case "f":
			e.minibuffer.MoveCursorForward()
			return
		case "b":
			e.minibuffer.MoveCursorBackward()
			return
		case "a":
			e.minibuffer.MoveCursorToBeginning()
			return
		case "e":
			e.minibuffer.MoveCursorToEnd()
			return
		}
	}
	
	// Handle Backspace
	if event.Key == "Backspace" || event.Key == "\x7f" {
		e.minibuffer.DeleteBackward()
		return
	}
	
	// Handle normal character input
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		e.minibuffer.InsertChar(event.Rune)
	}
}

// handleBufferCompletion implements buffer name completion
func (e *Editor) handleBufferCompletion() {
	input := e.minibuffer.Content()
	matches := []string{}
	
	// Find matching buffer names
	for _, name := range e.GetBufferNames() {
		if strings.HasPrefix(name, input) {
			matches = append(matches, name)
		}
	}
	
	if len(matches) == 0 {
		// No matches
		return
	} else if len(matches) == 1 {
		// Single match - complete it
		e.minibuffer.content = matches[0]
		e.minibuffer.cursor = len([]rune(matches[0]))
	} else {
		// Multiple matches - find common prefix
		commonPrefix := findCommonPrefix(matches)
		if len(commonPrefix) > len(input) {
			e.minibuffer.content = commonPrefix
			e.minibuffer.cursor = len([]rune(commonPrefix))
		} else {
			// Show all matches in message
			e.minibuffer.SetMessage("Matches: " + strings.Join(matches, ", "))
		}
	}
}

// findCommonPrefix finds the common prefix of a list of strings
func findCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	
	prefix := strs[0]
	for _, str := range strs[1:] {
		for i := 0; i < len(prefix) && i < len(str); i++ {
			if prefix[i] != str[i] {
				prefix = prefix[:i]
				break
			}
		}
		if len(str) < len(prefix) {
			prefix = str
		}
	}
	
	return prefix
}

// formatBufferList creates buffer list content in Emacs format
func (e *Editor) formatBufferList() []string {
	var lines []string
	
	// Header line
	header := "CRM Buffer                Size  Mode              File"
	lines = append(lines, header)
	
	currentBuffer := e.CurrentBuffer()
	
	// Sort buffers by usage order (most recent first)
	// For now, keep original order but put current buffer first
	var sortedBuffers []*Buffer
	if currentBuffer != nil {
		sortedBuffers = append(sortedBuffers, currentBuffer)
	}
	for _, buffer := range e.buffers {
		if buffer != currentBuffer {
			sortedBuffers = append(sortedBuffers, buffer)
		}
	}
	
	// Format each buffer line
	for _, buffer := range sortedBuffers {
		line := e.formatBufferLine(buffer, currentBuffer)
		lines = append(lines, line)
	}
	
	return lines
}

// formatBufferLine formats a single buffer line
func (e *Editor) formatBufferLine(buffer *Buffer, currentBuffer *Buffer) string {
	// CRM indicators
	c := " " // C - Current buffer indicator
	r := " " // R - Read-only indicator  
	m := " " // M - Modified indicator
	
	// Current buffer gets "."
	if buffer == currentBuffer {
		c = "."
	}
	
	// Modified buffer gets "*"
	if buffer.IsModified() {
		m = "*"
	}
	
	// Read-only buffers get "%" (for now, assume all buffers are writable)
	// Special buffers like *Buffer List*, *scratch* could be marked read-only
	if strings.HasPrefix(buffer.Name(), "*") && buffer.Name() != "*scratch*" {
		r = "%"
	}
	
	// Buffer name (truncated to fit column width)
	name := buffer.Name()
	if len(name) > 20 {
		name = name[:17] + "..."
	}
	
	// Buffer size
	size := e.getBufferSize(buffer)
	
	// Mode (simplified for now)
	mode := e.getBufferMode(buffer)
	
	// File path
	filepath := buffer.Filepath()
	if filepath == "" {
		filepath = ""
	}
	
	// Format the line with proper spacing
	line := fmt.Sprintf("%s%s%s %-20s %5d  %-16s %s", 
		c, r, m, name, size, mode, filepath)
	
	return line
}

// getBufferSize calculates buffer size in characters
func (e *Editor) getBufferSize(buffer *Buffer) int {
	total := 0
	for _, line := range buffer.Content() {
		total += len(line) + 1 // +1 for newline
	}
	return total
}

// getBufferMode returns the buffer mode
func (e *Editor) getBufferMode(buffer *Buffer) string {
	name := buffer.Name()
	
	// Special buffer modes
	if name == "*scratch*" {
		return "Lisp Interaction"
	}
	if name == "*Buffer List*" {
		return "Buffer Menu"
	}
	if strings.HasPrefix(name, "*") {
		return "Special"
	}
	
	// File-based modes (simplified)
	filepath := buffer.Filepath()
	if filepath != "" {
		if strings.HasSuffix(filepath, ".go") {
			return "Go"
		}
		if strings.HasSuffix(filepath, ".md") {
			return "Markdown"
		}
		if strings.HasSuffix(filepath, ".txt") {
			return "Text"
		}
		return "Fundamental"
	}
	
	return "Fundamental"
}

// RegisterBufferCommands registers buffer management commands
func RegisterBufferCommands(registry *CommandRegistry) {
	registry.RegisterFunc("switch-to-buffer", SwitchToBufferInteractive)
	registry.RegisterFunc("list-buffers", ListBuffersInteractive)
	registry.RegisterFunc("kill-buffer", KillBufferInteractive)
}