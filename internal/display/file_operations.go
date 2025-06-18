package display

import (
	"fmt"

	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
)

// init registers the file operations plugin
func init() {
	RegisterPlugin(func(editor *Editor, registry *command.Registry) {
		editor.registerFileOperationsCommands(registry)
	})
}

// findFile opens a file for editing (C-x C-f)
func (e *Editor) findFile() error {
	// Use minibuffer's channel-based input system
	// Set minibuffer active to use unified input channel
	e.minibuffer.SetActive(true)
	
	// Prompt for filename using ReadCommand (which uses channel-based input)
	filename, err := e.minibuffer.ReadString("Find file: ")
	if err != nil {
		if err.Error() == "quit" {
			e.minibuffer.ShowMessage("Quit")
			return nil
		}
		return fmt.Errorf("failed to read filename: %v", err)
	}

	if filename == "" {
		e.minibuffer.ShowMessage("No filename specified")
		return nil
	}

	// Check if buffer already exists
	bufIndex := e.FindBuffer(filename)
	if bufIndex != -1 {
		// Buffer already exists, switch to it
		err := e.SwitchToBuffer(bufIndex)
		if err != nil {
			return err
		}
		e.minibuffer.ShowMessage(fmt.Sprintf("Switched to %s", filename))
		return nil
	}

	// Create new buffer and load file
	buf := buffer.NewFromFile(filename)
	err = buf.LoadFromFile(filename)
	if err != nil {
		// If file doesn't exist, create empty buffer with that name
		e.minibuffer.ShowMessage(fmt.Sprintf("(New file) %s", filename))
		buf.SetFilename(filename)
	} else {
		e.minibuffer.ShowMessage(fmt.Sprintf("Loaded %s", filename))
	}

	// Add buffer to buffer list and switch to it
	bufIndex = e.AddBuffer(buf)
	err = e.SwitchToBuffer(bufIndex)
	if err != nil {
		return err
	}

	return nil
}

// saveBuffer saves the current buffer to its file (C-x C-s)
func (e *Editor) saveBuffer() error {
	buf := e.currentWin.Buffer()

	if buf.Filename() == "" {
		// No filename, need to prompt for one
		return e.writeFile()
	}

	err := buf.Save()
	if err != nil {
		e.minibuffer.ShowError(fmt.Errorf("failed to save buffer: %v", err))
		return nil
	}

	e.minibuffer.ShowMessage(fmt.Sprintf("Wrote %s", buf.Filename()))
	return nil
}

// writeFile saves the buffer to a specified file (C-x C-w)
func (e *Editor) writeFile() error {
	// Use minibuffer's channel-based input system
	// Set minibuffer active to use unified input channel
	e.minibuffer.SetActive(true)

	buf := e.currentWin.Buffer()
	currentFilename := buf.Filename()

	// Prompt for filename
	prompt := "Write file: "
	if currentFilename != "" {
		prompt = fmt.Sprintf("Write file (default %s): ", currentFilename)
	}

	filename, err := e.minibuffer.ReadString(prompt)
	if err != nil {
		if err.Error() == "quit" {
			e.minibuffer.ShowMessage("Quit")
			return nil
		}
		return fmt.Errorf("failed to read filename: %v", err)
	}

	// Use current filename if none specified
	if filename == "" {
		if currentFilename == "" {
			e.minibuffer.ShowMessage("No filename specified")
			return nil
		}
		filename = currentFilename
	}

	err = buf.SaveToFile(filename)
	if err != nil {
		e.minibuffer.ShowError(fmt.Errorf("failed to save to %s: %v", filename, err))
		return nil
	}

	e.minibuffer.ShowMessage(fmt.Sprintf("Wrote %s", filename))
	return nil
}

// registerFileOperationsCommands registers file operations commands
func (e *Editor) registerFileOperationsCommands(registry *command.Registry) {
	// File I/O commands
	registry.Register("find-file", "Open a file", "", func(args ...interface{}) error {
		return e.findFile()
	})

	registry.Register("save-buffer", "Save current buffer to file", "", func(args ...interface{}) error {
		return e.saveBuffer()
	})

	registry.Register("write-file", "Save buffer to a specified file", "", func(args ...interface{}) error {
		return e.writeFile()
	})
}