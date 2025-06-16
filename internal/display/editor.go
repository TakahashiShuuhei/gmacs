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
	rawKeyboard *input.RawKeyboard
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
	rawKeyboard, _ := input.NewRawKeyboard() // Ignore error for now
	minibuffer := NewMinibuffer(terminal, keyboard)
	
	// Create a default buffer and window
	buf := buffer.New("*scratch*")
	buf.SetText("Welcome to gmacs!\n\nThis is the scratch buffer.\nType M-x to execute commands.\n\nAvailable commands:\n- version: Show version\n- hello: Say hello\n- list-commands: List all commands\n- quit: Exit the editor")
	
	width, height := terminal.Size()
	win := window.New(buf, height-2, width) // Reserve space for minibuffer
	
	// Create global keymap
	km := keymap.New("global")
	
	editor := &Editor{
		terminal:    terminal,
		keyboard:    keyboard,
		rawKeyboard: rawKeyboard,
		minibuffer:  minibuffer,
		currentWin:  win,
		keymap:      km,
		running:     true,
		registry:    command.GetGlobalRegistry(),
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
	
	// Try to enable raw mode for proper key detection
	rawModeEnabled := false
	if err := e.rawKeyboard.EnableRawMode(); err != nil {
		// Fallback to line-based input if raw mode fails
		e.minibuffer.ShowMessage("Raw mode unavailable, using fallback mode")
		rawModeEnabled = false
	} else {
		rawModeEnabled = true
		defer e.rawKeyboard.DisableRawMode()
	}
	
	for e.running {
		e.redraw()
		
		if rawModeEnabled {
			// Read key input in raw mode
			keyEvent, err := e.rawKeyboard.ReadKey()
			if err != nil {
				return err
			}
			
			// Handle the key event
			if err := e.handleKeyEvent(keyEvent); err != nil {
				e.minibuffer.ShowError(err)
			}
		} else {
			// Fallback to line-based input
			line, err := e.keyboard.ReadLine()
			if err != nil {
				return err
			}
			
			// Handle the input using old method
			if err := e.handleInput(line); err != nil {
				e.minibuffer.ShowError(err)
			}
		}
	}
	
	return nil
}

// handleKeyEvent processes a key event from raw keyboard input
func (e *Editor) handleKeyEvent(keyEvent *input.KeyEvent) error {
	// Handle special key combinations directly
	keyStr := keyEvent.Key.String()
	switch keyStr {
	case "M-x":
		return e.executeExtendedCommand()
	case "C-x":
		// Start multi-key sequence (for now, just show message)
		e.minibuffer.ShowMessage("C-x-")
		return nil
	case "C-g":
		e.minibuffer.ShowMessage("Quit")
		return nil
	case "C-c":
		// C-c prefix key (like Emacs)
		e.minibuffer.ShowMessage("C-c-")
		return nil
	default:
		// Try to look up binding
		seq := keymap.KeySequence{keyEvent.Key}
		if binding, exists := e.keymap.Lookup(seq); exists {
			return e.registry.Execute(binding.Command, binding.Args...)
		}
		
		// If it's a printable character, just show it was pressed
		if keyEvent.Printable {
			e.minibuffer.ShowMessage(fmt.Sprintf("Key: %c", keyEvent.Key.Char))
		} else {
			e.minibuffer.ShowMessage(fmt.Sprintf("'%s' is undefined", keyStr))
		}
		return nil
	}
}

// handleInput processes user input (fallback for non-raw mode)
func (e *Editor) handleInput(input string) error {
	if input == "" {
		// Just redraw on empty input
		return nil
	}
	
	// First, convert input using the same logic as simple mode
	processedInput := e.preprocessInput(input)
	
	// Parse the processed input as a key event
	keyEvent, err := e.parseKeyInput(processedInput)
	if err != nil {
		e.minibuffer.ShowMessage(fmt.Sprintf("Invalid key input: %s", input))
		return nil
	}
	
	// Handle special key combinations directly
	keyStr := keyEvent.String()
	switch keyStr {
	case "M-x":
		return e.executeExtendedCommand()
	case "C-x":
		// Start multi-key sequence (for now, just show message)
		e.minibuffer.ShowMessage("C-x-")
		return nil
	case "C-g":
		e.minibuffer.ShowMessage("Quit")
		return nil
	default:
		// Try to look up binding
		seq := keymap.KeySequence{keyEvent}
		if binding, exists := e.keymap.Lookup(seq); exists {
			return e.registry.Execute(binding.Command, binding.Args...)
		}
		
		// If not a key binding, show message
		e.minibuffer.ShowMessage(fmt.Sprintf("'%s' is undefined", keyStr))
		return nil
	}
}

// parseKeyInput parses raw input string into a Key
func (e *Editor) parseKeyInput(input string) (keymap.Key, error) {
	// Handle escape sequences (like ESC+x for M-x)
	if len(input) >= 2 && input[0] == 0x1b { // ESC character (27)
		if len(input) == 2 {
			// Alt+key combination: ESC + key
			char := rune(input[1])
			return keymap.NewAltKey(char), nil
		}
	}
	
	// Handle Ctrl+key combinations (single byte control characters)
	if len(input) == 1 {
		b := input[0]
		switch b {
		case 0x18: // Ctrl+X
			return keymap.NewCtrlKey('x'), nil
		case 0x03: // Ctrl+C
			return keymap.NewCtrlKey('c'), nil
		case 0x07: // Ctrl+G
			return keymap.NewCtrlKey('g'), nil
		case 0x06: // Ctrl+F
			return keymap.NewCtrlKey('f'), nil
		case 0x13: // Ctrl+S
			return keymap.NewCtrlKey('s'), nil
		}
		
		// Handle other control characters (0x01-0x1F)
		if b >= 0x01 && b <= 0x1F && b != 0x09 && b != 0x0A && b != 0x0D {
			// Convert control character back to letter
			char := rune('a' + b - 1)
			return keymap.NewCtrlKey(char), nil
		}
	}
	
	// Single printable character
	if len(input) == 1 && input[0] >= 32 && input[0] <= 126 {
		return keymap.NewKey(rune(input[0])), nil
	}
	
	// Try parsing as key sequence string (fallback)
	seq, err := keymap.ParseKeySequence(input)
	if err == nil && len(seq) == 1 {
		return seq[0], nil
	}
	
	return keymap.Key{}, fmt.Errorf("unable to parse key input: %s", input)
}

// preprocessInput processes raw input the same way as SimpleEditor
func (e *Editor) preprocessInput(input string) string {
	// Handle escape sequences (like ^[x for M-x)
	if len(input) >= 2 && input[0] == 0x1b { // ESC character (27)
		if len(input) == 2 {
			// Alt+key combination: ESC + key
			char := input[1]
			return fmt.Sprintf("M-%c", char)
		}
	}
	
	// Handle ^[x format (display format)
	if strings.HasPrefix(input, "^[") && len(input) == 3 {
		char := input[2]
		return fmt.Sprintf("M-%c", char)
	}
	
	// Handle the case where ^[x is literally displayed in terminal
	if input == "^[x" {
		return "M-x"
	}
	
	// Handle Ctrl+key combinations (single byte control characters)
	if len(input) == 1 {
		b := input[0]
		switch b {
		case 0x18: // Ctrl+X
			return "C-x"
		case 0x03: // Ctrl+C
			return "C-c"
		case 0x07: // Ctrl+G
			return "C-g"
		}
		
		// Handle other control characters (0x01-0x1F)
		if b >= 0x01 && b <= 0x1F && b != 0x09 && b != 0x0A && b != 0x0D {
			// Convert control character back to letter
			char := 'a' + b - 1
			return fmt.Sprintf("C-%c", char)
		}
	}
	
	// Return as-is
	return input
}


// executeExtendedCommand handles M-x command execution
func (e *Editor) executeExtendedCommand() error {
	// Temporarily disable raw mode to read line input
	if e.rawKeyboard.IsRawMode() {
		e.rawKeyboard.DisableRawMode()
		defer e.rawKeyboard.EnableRawMode()
	}
	
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
	e.minibuffer.ShowMessage("Welcome to gmacs! Type M-x and press Enter for commands. For better key support, use: gmacs --simple")
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