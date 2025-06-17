package display

import (
	"fmt"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/command"
)

// init registers the buffer management plugin
func init() {
	RegisterPlugin(func(editor *Editor, registry *command.Registry) {
		editor.registerBufferManagementCommands(registry)
	})
}

// listBuffers lists all open buffers (C-x C-b)
func (e *Editor) listBuffers() error {
	currentBuf := e.currentWin.Buffer()
	bufferNames := make([]string, len(e.buffers))
	for i, buf := range e.buffers {
		mark := " "
		if buf == currentBuf {
			mark = "*"
		}
		bufferNames[i] = fmt.Sprintf("%s %s", mark, buf.Name())
	}
	message := fmt.Sprintf("Buffers (%d): %s", len(e.buffers), strings.Join(bufferNames, ", "))
	e.minibuffer.ShowMessage(message)
	return nil
}

// switchToBuffer switches to another buffer (C-x b)
func (e *Editor) switchToBuffer() error {
	// This would normally prompt for buffer name, but for now just cycle through buffers
	currentBuf := e.currentWin.Buffer()
	currentIndex := -1
	for i, buf := range e.buffers {
		if buf == currentBuf {
			currentIndex = i
			break
		}
	}
	nextBuffer := (currentIndex + 1) % len(e.buffers)
	return e.SwitchToBuffer(nextBuffer)
}

// registerBufferManagementCommands registers buffer management commands
func (e *Editor) registerBufferManagementCommands(registry *command.Registry) {
	// Buffer management commands
	registry.Register("list-buffers", "List all buffers", "", func(args ...interface{}) error {
		return e.listBuffers()
	})

	registry.Register("switch-to-buffer", "Switch to a buffer", "", func(args ...interface{}) error {
		return e.switchToBuffer()
	})
}