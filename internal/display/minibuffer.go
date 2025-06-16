package display

import (
	"fmt"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/input"
)

// Minibuffer represents the minibuffer (like Emacs minibuffer)
type Minibuffer struct {
	terminal    *Terminal
	keyboard    *input.Keyboard
	prompt      string
	input       string
	cursorPos   int
	completions []string
	history     []string
	historyPos  int
	active      bool
}

// NewMinibuffer creates a new minibuffer
func NewMinibuffer(terminal *Terminal, keyboard *input.Keyboard) *Minibuffer {
	return &Minibuffer{
		terminal:   terminal,
		keyboard:   keyboard,
		historyPos: -1,
	}
}

// ReadCommand reads a command from the user with M-x prompt
func (mb *Minibuffer) ReadCommand() (string, error) {
	mb.prompt = "M-x "
	mb.input = ""
	mb.cursorPos = 0
	mb.active = true
	
	defer func() {
		mb.active = false
		mb.clearMinibuffer()
	}()
	
	for {
		mb.displayMinibuffer()
		
		line, err := mb.keyboard.ReadLine()
		if err != nil {
			return "", err
		}
		
		// Handle special inputs
		switch line {
		case "": // Enter pressed
			if mb.input != "" {
				mb.addToHistory(mb.input)
				return mb.input, nil
			}
			continue
		case "C-g", "\\C-g": // Cancel
			return "", fmt.Errorf("quit")
		case "C-c", "\\C-c": // Cancel
			return "", fmt.Errorf("quit")
		case "\\tab", "tab": // Tab completion
			mb.handleCompletion()
			continue
		default:
			// Normal input
			mb.input = line
			mb.cursorPos = len(mb.input)
		}
	}
}

// ReadString reads a string from the user with a custom prompt
func (mb *Minibuffer) ReadString(prompt string) (string, error) {
	mb.prompt = prompt
	mb.input = ""
	mb.cursorPos = 0
	mb.active = true
	
	defer func() {
		mb.active = false
		mb.clearMinibuffer()
	}()
	
	mb.displayMinibuffer()
	
	line, err := mb.keyboard.ReadLine()
	if err != nil {
		return "", err
	}
	
	if line == "C-g" || line == "\\C-g" {
		return "", fmt.Errorf("quit")
	}
	
	return line, nil
}

// displayMinibuffer displays the current minibuffer content
func (mb *Minibuffer) displayMinibuffer() {
	width, height := mb.terminal.Size()
	
	// Move to the last line (minibuffer line)
	mb.terminal.MoveCursor(height, 1)
	mb.terminal.ClearLine()
	
	// Display prompt and input
	display := mb.prompt + mb.input
	
	// Truncate if too long
	if len(display) > width-1 {
		display = display[:width-1]
	}
	
	mb.terminal.Print(display)
	
	// Position cursor
	cursorCol := len(mb.prompt) + mb.cursorPos + 1
	if cursorCol <= width {
		mb.terminal.MoveCursor(height, cursorCol)
	}
	
	mb.terminal.Flush()
}

// clearMinibuffer clears the minibuffer line
func (mb *Minibuffer) clearMinibuffer() {
	_, height := mb.terminal.Size()
	mb.terminal.MoveCursor(height, 1)
	mb.terminal.ClearLine()
	mb.terminal.Flush()
}

// handleCompletion handles tab completion for commands
func (mb *Minibuffer) handleCompletion() {
	if mb.prompt == "M-x " {
		// Command completion
		completions := command.ListWithPrefix(mb.input)
		mb.completions = completions
		
		if len(completions) == 1 {
			// Exact match, complete it
			mb.input = completions[0]
			mb.cursorPos = len(mb.input)
		} else if len(completions) > 1 {
			// Multiple matches, show them
			mb.showCompletions(completions)
		}
	}
}

// showCompletions displays completion candidates
func (mb *Minibuffer) showCompletions(completions []string) {
	if len(completions) == 0 {
		return
	}
	
	width, height := mb.terminal.Size()
	
	// Show completions above the minibuffer
	displayLine := height - 1
	
	// Clear the line above minibuffer
	mb.terminal.MoveCursor(displayLine, 1)
	mb.terminal.ClearLine()
	
	// Show completions (limit to terminal width)
	completionText := "Completions: " + strings.Join(completions, " ")
	if len(completionText) > width-1 {
		completionText = completionText[:width-4] + "..."
	}
	
	mb.terminal.Print(completionText)
}

// addToHistory adds a command to the history
func (mb *Minibuffer) addToHistory(cmd string) {
	if cmd == "" {
		return
	}
	
	// Remove duplicates
	for i, h := range mb.history {
		if h == cmd {
			mb.history = append(mb.history[:i], mb.history[i+1:]...)
			break
		}
	}
	
	// Add to the end
	mb.history = append(mb.history, cmd)
	
	// Limit history size
	if len(mb.history) > 100 {
		mb.history = mb.history[1:]
	}
	
	mb.historyPos = len(mb.history)
}

// ShowMessage displays a message in the minibuffer
func (mb *Minibuffer) ShowMessage(message string) {
	width, height := mb.terminal.Size()
	
	mb.terminal.MoveCursor(height, 1)
	mb.terminal.ClearLine()
	
	// Truncate message if too long
	if len(message) > width-1 {
		message = message[:width-4] + "..."
	}
	
	mb.terminal.Print(message)
	mb.terminal.Flush()
}

// ShowError displays an error message in the minibuffer
func (mb *Minibuffer) ShowError(err error) {
	mb.terminal.SetColor(ColorRed, -1)
	mb.ShowMessage(fmt.Sprintf("Error: %v", err))
	mb.terminal.ResetColor()
}

// IsActive returns whether the minibuffer is currently active
func (mb *Minibuffer) IsActive() bool {
	return mb.active
}

// Clear clears the minibuffer
func (mb *Minibuffer) Clear() {
	mb.clearMinibuffer()
}