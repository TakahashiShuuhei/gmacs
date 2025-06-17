package display

import (
	"fmt"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/config"
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
	keySequence keymap.KeySequence // For multi-key sequences
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
	
	// Load Lua configuration (includes default key bindings)
	editor.loadLuaConfig()
	
	return editor
}

// isFullWidth checks if a character is full-width (typically CJK characters)
func isFullWidth(r rune) bool {
	// Simplified check for common full-width ranges
	return (r >= 0x1100 && r <= 0x11FF) || // Hangul Jamo
		   (r >= 0x2E80 && r <= 0x2EFF) || // CJK Radicals Supplement
		   (r >= 0x2F00 && r <= 0x2FDF) || // Kangxi Radicals
		   (r >= 0x3000 && r <= 0x303F) || // CJK Symbols and Punctuation
		   (r >= 0x3040 && r <= 0x309F) || // Hiragana
		   (r >= 0x30A0 && r <= 0x30FF) || // Katakana
		   (r >= 0x3100 && r <= 0x312F) || // Bopomofo
		   (r >= 0x3200 && r <= 0x32FF) || // Enclosed CJK Letters and Months
		   (r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
		   (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		   (r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		   (r >= 0xFF00 && r <= 0xFFEF)    // Halfwidth and Fullwidth Forms
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
		
		// If it's a printable character, insert it
		if keyEvent.Printable {
			return e.selfInsertCommand(keyEvent.Key.Char)
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
	
	// Check if this is a multi-character string (likely from IME)
	inputRunes := []rune(input)
	processedRunes := []rune(processedInput)
	
	if len(inputRunes) > 1 && len(processedRunes) > 1 {
		// This is a multi-character input, insert it as a string
		return e.selfInsertStringCommand(input)
	}
	
	// Parse the processed input as a key event
	keyEvent, err := e.parseKeyInput(processedInput)
	if err != nil {
		// If parsing fails, try to insert as string if it contains printable characters
		if len(inputRunes) > 0 {
			return e.selfInsertStringCommand(input)
		}
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
		
		// This part will be handled by the raw input processing
		// Multi-character strings are handled in handleInput
		
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


// selfInsertCommand inserts a character at the current cursor position
func (e *Editor) selfInsertCommand(char rune) error {
	// Get current buffer and cursor
	buffer := e.currentWin.Buffer()
	cursor := e.currentWin.Cursor()
	
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
		
		// Truncate line if too long (considering full-width characters)
		runes := []rune(line)
		displayWidth := 0
		cutIndex := len(runes)
		
		for i, char := range runes {
			charWidth := 1
			if isFullWidth(char) {
				charWidth = 2
			}
			
			if displayWidth + charWidth > width-1 {
				cutIndex = i
				break
			}
			displayWidth += charWidth
		}
		
		if cutIndex < len(runes) {
			line = string(runes[:cutIndex])
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
	
	// Fill with dashes (UTF-8 safe)
	leftRunes := []rune(leftPart)
	rightRunes := []rune(rightPart)
	padding := width - len(leftRunes) - len(rightRunes)
	if padding < 0 {
		padding = 0
	}
	
	status := leftPart + strings.Repeat("-", padding) + rightPart
	statusRunes := []rune(status)
	if len(statusRunes) > width {
		status = string(statusRunes[:width])
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

// findFile opens a file for editing (C-x C-f)
func (e *Editor) findFile() error {
	// Temporarily disable raw mode to read line input
	if e.rawKeyboard.IsRawMode() {
		e.rawKeyboard.DisableRawMode()
		defer e.rawKeyboard.EnableRawMode()
	}
	
	// Prompt for filename
	e.minibuffer.ShowMessage("Find file: ")
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
	
	// Switch to the new buffer
	width, height := e.terminal.Size()
	e.currentWin = window.New(buf, height-2, width)
	
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
	// Temporarily disable raw mode to read line input
	if e.rawKeyboard.IsRawMode() {
		e.rawKeyboard.DisableRawMode()
		defer e.rawKeyboard.EnableRawMode()
	}
	
	buf := e.currentWin.Buffer()
	currentFilename := buf.Filename()
	
	// Prompt for filename
	prompt := "Write file: "
	if currentFilename != "" {
		prompt = fmt.Sprintf("Write file (default %s): ", currentFilename)
	}
	
	e.minibuffer.ShowMessage(prompt)
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

// quit quits the editor
func (e *Editor) quit() error {
	e.running = false
	e.terminal.Clear()
	return nil
}

// BindKey binds a key sequence to a command
func (e *Editor) BindKey(keySeq string, command string) error {
	// Parse the key sequence
	keys, err := keymap.ParseKeySequence(keySeq)
	if err != nil {
		return fmt.Errorf("failed to parse key sequence '%s': %v", keySeq, err)
	}
	
	// Bind to keymap
	err = e.keymap.Bind(keys, command)
	if err != nil {
		return fmt.Errorf("failed to bind key sequence '%s': %v", keySeq, err)
	}
	
	return nil
}

// GetMinibuffer returns the minibuffer (implements EditorInterface)
func (e *Editor) GetMinibuffer() config.MinibufferInterface {
	return e.minibuffer
}

// GetCommandRegistry returns the command registry (implements EditorInterface)
func (e *Editor) GetCommandRegistry() *command.Registry {
	return e.registry
}

// loadLuaConfig loads the Lua configuration
func (e *Editor) loadLuaConfig() {
	luaConfig := config.NewLuaConfig(e)
	
	err := luaConfig.LoadConfig()
	if err != nil {
		e.minibuffer.ShowError(fmt.Errorf("failed to load Lua config: %v", err))
	}
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
	
	// Text insertion commands
	e.registry.Register("self-insert-command", "Insert typed character", "", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("self-insert-command requires a character argument")
		}
		char, ok := args[0].(rune)
		if !ok {
			return fmt.Errorf("self-insert-command argument must be a character")
		}
		return e.selfInsertCommand(char)
	})
	
	e.registry.Register("insert-string", "Insert string at cursor", "", func(args ...interface{}) error {
		if len(args) == 0 {
			return fmt.Errorf("insert-string requires a string argument")
		}
		text, ok := args[0].(string)
		if !ok {
			return fmt.Errorf("insert-string argument must be a string")
		}
		return e.selfInsertStringCommand(text)
	})
	
	// Cursor movement commands
	e.registry.Register("forward-char", "Move cursor forward one character", "", func(args ...interface{}) error {
		return e.forwardChar()
	})
	
	e.registry.Register("backward-char", "Move cursor backward one character", "", func(args ...interface{}) error {
		return e.backwardChar()
	})
	
	e.registry.Register("next-line", "Move cursor to next line", "", func(args ...interface{}) error {
		return e.nextLine()
	})
	
	e.registry.Register("previous-line", "Move cursor to previous line", "", func(args ...interface{}) error {
		return e.previousLine()
	})
	
	// Text deletion commands
	e.registry.Register("delete-char", "Delete character at cursor", "", func(args ...interface{}) error {
		return e.deleteChar()
	})
	
	e.registry.Register("backward-delete-char", "Delete character before cursor", "", func(args ...interface{}) error {
		return e.backwardDeleteChar()
	})
	
	// File I/O commands
	e.registry.Register("find-file", "Open a file", "", func(args ...interface{}) error {
		return e.findFile()
	})
	
	e.registry.Register("save-buffer", "Save current buffer to file", "", func(args ...interface{}) error {
		return e.saveBuffer()
	})
	
	e.registry.Register("write-file", "Save buffer to a specified file", "", func(args ...interface{}) error {
		return e.writeFile()
	})
}

