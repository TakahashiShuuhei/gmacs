package domain

import (
	"strings"

	"github.com/TakahashiShuuhei/gmacs/core/events"
	"github.com/TakahashiShuuhei/gmacs/core/log"
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
	// Create a buffer list message
	var bufferList strings.Builder
	bufferList.WriteString("Buffers: ")
	
	for i, buffer := range e.buffers {
		if i > 0 {
			bufferList.WriteString(", ")
		}
		bufferList.WriteString(buffer.Name())
		
		// Mark current buffer
		if buffer == e.CurrentBuffer() {
			bufferList.WriteString(" (current)")
		}
	}
	
	e.minibuffer.SetMessage(bufferList.String())
	log.Info("Listed buffers")
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