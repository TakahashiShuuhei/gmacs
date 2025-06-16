package display

import (
	"fmt"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/input"
	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
	"github.com/TakahashiShuuhei/gmacs/internal/window"
)

// Editor represents the main editor interface
type Editor struct {
	terminal    *Terminal
	keyboard    *input.Keyboard
	minibuffer  *Minibuffer
	currentWin  *window.Window
	keymap      *keymap.Keymap
	running     bool
	registry    *command.Registry
}

// NewEditor creates a new editor instance
func NewEditor() *Editor {
	terminal := NewStandardTerminal()
	keyboard := input.CreateStandardKeyboard()
	minibuffer := NewMinibuffer(terminal, keyboard)
	
	// Create a default buffer and window
	buf := buffer.New("*scratch*")
	buf.SetText("Welcome to gmacs!\n\nThis is the scratch buffer.\nType M-x to execute commands.\n\nAvailable commands:\n- version: Show version\n- hello: Say hello\n- list-commands: List all commands\n- quit: Exit the editor")
	
	width, height := terminal.Size()
	win := window.New(buf, height-2, width) // Reserve space for minibuffer
	
	// Create global keymap
	km := keymap.New("global")
	
	editor := &Editor{
		terminal:   terminal,
		keyboard:   keyboard,
		minibuffer: minibuffer,
		currentWin: win,
		keymap:     km,
		running:    true,
		registry:   command.GetGlobalRegistry(),
	}
	
	// Register basic editor commands
	editor.registerEditorCommands()
	
	// Set up basic key bindings
	editor.setupKeyBindings()
	
	return editor
}

// Run starts the main editor loop
func (e *Editor) Run() error {
	e.terminal.Clear()
	e.showWelcomeMessage()
	
	for e.running {
		e.redraw()
		
		// Read key input
		line, err := e.keyboard.ReadLine()
		if err != nil {
			return err
		}
		
		// Handle the input
		if err := e.handleInput(line); err != nil {
			e.minibuffer.ShowError(err)
		}
	}
	
	return nil
}

// handleInput processes user input
func (e *Editor) handleInput(input string) error {
	// Handle special key combinations
	switch input {
	case "M-x", "\\M-x":
		return e.executeExtendedCommand()
	case "C-x C-c", "\\C-x \\C-c":
		return e.quit()
	case "C-g", "\\C-g":
		e.minibuffer.ShowMessage("Quit")
		return nil
	case "":
		// Just redraw on empty input
		return nil
	default:
		// Try to parse as key sequence and look up binding
		seq, err := keymap.ParseKeySequence(input)
		if err == nil {
			if binding, exists := e.keymap.Lookup(seq); exists {
				return e.registry.Execute(binding.Command, binding.Args...)
			}
		}
		
		// If not a key binding, show message
		e.minibuffer.ShowMessage(fmt.Sprintf("Unknown command: %s", input))
		return nil
	}
}

// executeExtendedCommand handles M-x command execution
func (e *Editor) executeExtendedCommand() error {
	commandName, err := e.minibuffer.ReadCommand()
	if err != nil {
		if err.Error() == "quit" {
			e.minibuffer.ShowMessage("Quit")
			return nil
		}
		return err
	}
	
	if commandName == "" {
		return nil
	}
	
	// Execute the command
	err = e.registry.Execute(commandName)
	if err != nil {
		e.minibuffer.ShowError(err)
		return nil
	}
	
	e.minibuffer.ShowMessage(fmt.Sprintf("Executed: %s", commandName))
	return nil
}

// redraw redraws the entire editor interface
func (e *Editor) redraw() {
	e.drawBuffer()
	e.drawStatusLine()
	e.drawCursor()
	e.terminal.Flush()
}

// drawBuffer draws the current buffer content
func (e *Editor) drawBuffer() {
	width, height := e.terminal.Size()
	bufferHeight := height - 2 // Reserve space for status line and minibuffer
	
	visibleLines := e.currentWin.GetVisibleText()
	
	// Clear buffer area
	for i := 1; i <= bufferHeight; i++ {
		e.terminal.MoveCursor(i, 1)
		e.terminal.ClearLine()
	}
	
	// Draw visible lines
	for i, line := range visibleLines {
		if i >= bufferHeight {
			break
		}
		
		e.terminal.MoveCursor(i+1, 1)
		
		// Truncate line if too long
		if len(line) > width {
			line = line[:width-1]
		}
		
		e.terminal.Print(line)
	}
}

// drawStatusLine draws the status line
func (e *Editor) drawStatusLine() {
	width, height := e.terminal.Size()
	statusLine := height - 1
	
	e.terminal.MoveCursor(statusLine, 1)
	e.terminal.ClearLine()
	
	// Create status line content
	bufferName := e.currentWin.Buffer().Name()
	modified := ""
	if e.currentWin.Buffer().IsModified() {
		modified = " *"
	}
	
	cursor := e.currentWin.Cursor()
	position := fmt.Sprintf("L%d C%d", cursor.Line()+1, cursor.Col()+1)
	
	// Format status line
	leftPart := fmt.Sprintf(" %s%s ", bufferName, modified)
	rightPart := fmt.Sprintf(" %s ", position)
	
	// Fill with dashes
	padding := width - len(leftPart) - len(rightPart)
	if padding < 0 {
		padding = 0
	}
	
	status := leftPart + strings.Repeat("-", padding) + rightPart
	if len(status) > width {
		status = status[:width]
	}
	
	// Draw status line with reverse video
	e.terminal.SetColor(ColorBlack, ColorWhite)
	e.terminal.Print(status)
	e.terminal.ResetColor()
}

// drawCursor positions the cursor in the buffer
func (e *Editor) drawCursor() {
	screenLine, screenCol := e.currentWin.CursorScreenPosition()
	// Adjust for 1-based terminal coordinates
	e.terminal.MoveCursor(screenLine+1, screenCol+1)
}

// quit quits the editor
func (e *Editor) quit() error {
	e.running = false
	e.terminal.Clear()
	return nil
}

// showWelcomeMessage shows a welcome message
func (e *Editor) showWelcomeMessage() {
	e.minibuffer.ShowMessage("Welcome to gmacs! Press M-x for commands, C-x C-c to quit")
}

// registerEditorCommands registers basic editor commands
func (e *Editor) registerEditorCommands() {
	// Quit command
	e.registry.Register("quit", "Quit gmacs", "", func(args ...interface{}) error {
		return e.quit()
	})
	
	// Version command
	e.registry.Register("version", "Show gmacs version", "", func(args ...interface{}) error {
		e.minibuffer.ShowMessage("gmacs version 0.0.1 - Go Emacs-like Editor")
		return nil
	})
	
	// Hello command
	e.registry.Register("hello", "Say hello", "", func(args ...interface{}) error {
		e.minibuffer.ShowMessage("Hello from gmacs!")
		return nil
	})
	
	// List commands
	e.registry.Register("list-commands", "List all available commands", "", func(args ...interface{}) error {
		commands := e.registry.List()
		message := fmt.Sprintf("Available commands (%d): %s", len(commands), strings.Join(commands, ", "))
		e.minibuffer.ShowMessage(message)
		return nil
	})
	
	// Redraw command
	e.registry.Register("redraw-display", "Redraw the display", "", func(args ...interface{}) error {
		e.terminal.Clear()
		e.redraw()
		e.minibuffer.ShowMessage("Display redrawn")
		return nil
	})
}

// setupKeyBindings sets up basic key bindings
func (e *Editor) setupKeyBindings() {
	// M-x -> execute-extended-command (handled specially in handleInput)
	// C-x C-c -> quit (handled specially in handleInput)
	// C-g -> cancel (handled specially in handleInput)
	
	// Add some basic bindings that call commands
	quitSeq, _ := keymap.ParseKeySequence("C-x C-c")
	e.keymap.Bind(quitSeq, "quit")
}