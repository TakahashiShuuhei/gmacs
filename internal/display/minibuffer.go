package display

import (
	"fmt"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/input"
	"github.com/TakahashiShuuhei/gmacs/internal/logging"
)

// Minibuffer represents the minibuffer (like Emacs minibuffer)
type Minibuffer struct {
	terminal      *Terminal
	keyboard      *input.Keyboard
	rawKeyboard   *input.RawKeyboard
	keyEventChan  chan *input.KeyEvent
	prompt        string
	input         string
	cursorPos     int
	completions   []string
	history       []string
	historyPos    int
	active        bool
}

// NewMinibuffer creates a new minibuffer
func NewMinibuffer(terminal *Terminal, keyboard *input.Keyboard, rawKeyboard *input.RawKeyboard, keyEventChan chan *input.KeyEvent) *Minibuffer {
	return &Minibuffer{
		terminal:     terminal,
		keyboard:     keyboard,
		rawKeyboard:  rawKeyboard,
		keyEventChan: keyEventChan,
		historyPos:   -1,
	}
}

// ReadCommand reads a command from the user with M-x prompt
func (mb *Minibuffer) ReadCommand() (string, error) {
	logging.Debug("ReadCommand started")
	
	mb.prompt = "M-x "
	mb.input = ""
	mb.cursorPos = 0
	// Don't set active here if it's already active (to avoid overriding)
	if !mb.active {
		mb.active = true
		logging.Debug("Minibuffer set to active in ReadCommand")
	} else {
		logging.Debug("Minibuffer already active, continuing")
	}
	
	logging.Debug("Displaying prompt")
	// Initial display
	mb.displayMinibuffer()
	
	logging.Debug("Starting raw input collection for minibuffer")
	line, err := mb.readRawInput()
	logging.Debug("Raw input collection returned: %q, err: %v", line, err)
	
	// 入力完了後、必ずミニバッファの状態をクリア
	mb.clearData()
	logging.Debug("Minibuffer cleared and deactivated")
	
	if err != nil {
		logging.LogError("ReadCommand", err)
		return "", err
	}
	
	// Handle special inputs
	switch line {
	case "": // Enter pressed - return empty to cancel
		return "", nil
	case "C-g", "\\C-g": // Cancel
		return "", fmt.Errorf("quit")
	case "C-c", "\\C-c": // Cancel
		return "", fmt.Errorf("quit")
	default:
		// Normal input - this is the command name
		if line != "" {
			mb.addToHistory(line)
			return line, nil
		}
		return "", nil
	}
}

// ReadString reads a string from the user with a custom prompt
func (mb *Minibuffer) ReadString(prompt string) (string, error) {
	mb.prompt = prompt
	mb.input = ""
	mb.cursorPos = 0
	mb.active = true
	
	defer func() {
		mb.clearData()
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

// clearData clears the minibuffer internal state
func (mb *Minibuffer) clearData() {
	logging.Debug("clearData called - deactivating minibuffer")
	mb.prompt = ""
	mb.input = ""
	mb.cursorPos = 0
	mb.active = false
}

// clearDisplay clears the minibuffer line on screen
func (mb *Minibuffer) clearDisplay() {
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
	
	// 確実にミニバッファ行をクリア
	mb.terminal.MoveCursor(height, 1)
	mb.terminal.ClearLine()
	
	// 空の場合は表示しない
	if message == "" {
		mb.prompt = ""
		mb.input = ""
		mb.terminal.Flush()
		return
	}
	
	// Truncate message if too long
	if len(message) > width-1 {
		message = message[:width-4] + "..."
	}
	
	mb.terminal.Print(message)
	mb.terminal.Flush()
	
	// メッセージ表示状態を記録（HasMessageで使用）
	mb.prompt = message
	mb.input = ""
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

// SetActive sets the minibuffer active state
func (mb *Minibuffer) SetActive(active bool) {
	logging.Debug("SetActive called with: %v", active)
	mb.active = active
}

// HasMessage returns whether the minibuffer is currently displaying a message
func (mb *Minibuffer) HasMessage() bool {
	// ミニバッファが非アクティブで、最後に何かメッセージが表示されている状態
	return !mb.active && (mb.prompt != "" || mb.input != "")
}

// Clear clears the minibuffer
func (mb *Minibuffer) Clear() {
	mb.clearData()
}

// readRawInput reads input character by character from the key event channel
func (mb *Minibuffer) readRawInput() (string, error) {
	logging.Debug("readRawInput started - using channel-based input")
	inputBuffer := ""
	
	for {
		logging.Debug("Waiting for key event from channel")
		keyEvent := <-mb.keyEventChan
		logging.Debug("Received key event: %s", keyEvent.Key.String())
		
		switch keyEvent.Key.String() {
		case "return", "enter":
			logging.Debug("Return key pressed, finishing input: %q", inputBuffer)
			return inputBuffer, nil
		case "C-g":
			logging.Debug("C-g pressed, canceling input")
			return "", fmt.Errorf("quit")
		case "backspace":
			if len(inputBuffer) > 0 {
				inputBuffer = inputBuffer[:len(inputBuffer)-1]
				mb.input = inputBuffer
				mb.cursorPos = len(inputBuffer)
				mb.displayMinibuffer()
				logging.Debug("Backspace pressed, new input: %q", inputBuffer)
			}
		default:
			// Handle printable characters
			if keyEvent.Key.Char != 0 && keyEvent.Key.Char >= 32 && keyEvent.Key.Char <= 126 {
				inputBuffer += string(keyEvent.Key.Char)
				mb.input = inputBuffer
				mb.cursorPos = len(inputBuffer)
				mb.displayMinibuffer()
				logging.Debug("Character added: %c, new input: %q", keyEvent.Key.Char, inputBuffer)
			} else {
				logging.Debug("Non-printable key ignored: %s", keyEvent.Key.String())
			}
		}
	}
}