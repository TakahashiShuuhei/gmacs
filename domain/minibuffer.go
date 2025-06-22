package domain

import (
	"github.com/TakahashiShuuhei/gmacs/events"
)


// MinibufferMode represents the current mode of the minibuffer
type MinibufferMode int

const (
	MinibufferInactive MinibufferMode = iota
	MinibufferCommand                 // M-x command input
	MinibufferMessage                 // Displaying a message
	MinibufferFile                    // File path input (C-x C-f)
)

// Minibuffer manages the minibuffer state
type Minibuffer struct {
	mode     MinibufferMode
	content  string
	prompt   string
	message  string
	cursor   int
}

func NewMinibuffer() *Minibuffer {
	return &Minibuffer{
		mode:    MinibufferInactive,
		content: "",
		prompt:  "",
		message: "",
		cursor:  0,
	}
}

func (mb *Minibuffer) Mode() MinibufferMode {
	return mb.mode
}

func (mb *Minibuffer) IsActive() bool {
	return mb.mode != MinibufferInactive
}

func (mb *Minibuffer) Content() string {
	return mb.content
}

func (mb *Minibuffer) Prompt() string {
	return mb.prompt
}

func (mb *Minibuffer) Message() string {
	return mb.message
}

func (mb *Minibuffer) CursorPosition() int {
	return mb.cursor
}

// StartCommandInput starts M-x command input mode
func (mb *Minibuffer) StartCommandInput() {
	mb.mode = MinibufferCommand
	mb.content = ""
	mb.prompt = "M-x "
	mb.message = ""
	mb.cursor = 0
}

// StartFileInput starts file path input mode (C-x C-f)
func (mb *Minibuffer) StartFileInput() {
	mb.mode = MinibufferFile
	mb.content = ""
	mb.prompt = "Find file: "
	mb.message = ""
	mb.cursor = 0
}

// SetMessage displays a message in the minibuffer
func (mb *Minibuffer) SetMessage(message string) {
	mb.mode = MinibufferMessage
	mb.content = ""
	mb.prompt = ""
	mb.message = message
	mb.cursor = 0
}

// Clear clears the minibuffer
func (mb *Minibuffer) Clear() {
	mb.mode = MinibufferInactive
	mb.content = ""
	mb.prompt = ""
	mb.message = ""
	mb.cursor = 0
}

// InsertChar inserts a character at the cursor position
func (mb *Minibuffer) InsertChar(ch rune) {
	if mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection {
		return
	}
	
	runes := []rune(mb.content)
	before := runes[:mb.cursor]
	after := runes[mb.cursor:]
	
	newRunes := append(before, ch)
	newRunes = append(newRunes, after...)
	
	mb.content = string(newRunes)
	mb.cursor++
	
}

// DeleteBackward deletes the character before the cursor
func (mb *Minibuffer) DeleteBackward() {
	if (mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection) || mb.cursor == 0 {
		return
	}
	
	runes := []rune(mb.content)
	before := runes[:mb.cursor-1]
	after := runes[mb.cursor:]
	
	mb.content = string(append(before, after...))
	mb.cursor--
}

// DeleteForward deletes the character at the cursor position
func (mb *Minibuffer) DeleteForward() {
	if mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection {
		return
	}
	
	runes := []rune(mb.content)
	if mb.cursor >= len(runes) {
		return // Nothing to delete
	}
	
	before := runes[:mb.cursor]
	after := runes[mb.cursor+1:]
	
	mb.content = string(append(before, after...))
	// Cursor position stays the same
}

// MoveCursorForward moves cursor one position to the right
func (mb *Minibuffer) MoveCursorForward() {
	if mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection {
		return
	}
	
	runes := []rune(mb.content)
	if mb.cursor < len(runes) {
		mb.cursor++
	}
}

// MoveCursorBackward moves cursor one position to the left
func (mb *Minibuffer) MoveCursorBackward() {
	if mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection {
		return
	}
	
	if mb.cursor > 0 {
		mb.cursor--
	}
}

// MoveCursorToBeginning moves cursor to the beginning of the line
func (mb *Minibuffer) MoveCursorToBeginning() {
	if mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection {
		return
	}
	
	mb.cursor = 0
}

// MoveCursorToEnd moves cursor to the end of the line
func (mb *Minibuffer) MoveCursorToEnd() {
	if mb.mode != MinibufferCommand && mb.mode != MinibufferFile && mb.mode != MinibufferBufferSelection {
		return
	}
	
	runes := []rune(mb.content)
	mb.cursor = len(runes)
}

// GetDisplayText returns the text to display in the minibuffer
func (mb *Minibuffer) GetDisplayText() string {
	switch mb.mode {
	case MinibufferCommand:
		return mb.prompt + mb.content
	case MinibufferFile:
		return mb.prompt + mb.content
	case MinibufferBufferSelection:
		return mb.prompt + mb.content
	case MinibufferMessage:
		return mb.message
	default:
		return ""
	}
}

// HandleInput handles key input for the minibuffer
func (mb *Minibuffer) HandleInput(event events.KeyEventData, editor *Editor) bool {
	switch mb.mode {
	case MinibufferCommand:
		return mb.handleAsBuffer(event, func() { mb.executeCommand(editor) })
	case MinibufferFile:
		return mb.handleAsBuffer(event, func() { mb.executeFileOpen(editor) })
	case MinibufferBufferSelection:
		editor.HandleBufferSelectionInput(event)
		return true
	case MinibufferMessage:
		// Any key clears the message, but allow the key to continue being processed
		mb.Clear()
		return false
	}
	return false
}

// handleAsBuffer treats minibuffer like a regular buffer, using unified commands
func (mb *Minibuffer) handleAsBuffer(event events.KeyEventData, onEnter func()) bool {
	// Handle Enter - execute the completion action
	if event.Key == "Enter" || event.Key == "Return" {
		onEnter()
		return true
	}
	
	// Handle Escape - cancel
	if event.Key == "\x1b" || event.Key == "Escape" {
		mb.Clear()
		return true
	}
	
	// Handle Backspace as delete-backward-char
	if event.Key == "Backspace" || event.Key == "\x7f" {
		mb.executeCommandOnSelf("delete-backward-char")
		return true
	}
	
	// Handle Ctrl commands using the unified command system
	if event.Ctrl {
		switch event.Key {
		case "h":
			mb.executeCommandOnSelf("delete-backward-char")
			return true
		case "d":
			mb.executeCommandOnSelf("delete-forward-char")
			return true
		case "f":
			mb.executeCommandOnSelf("forward-char")
			return true
		case "b":
			mb.executeCommandOnSelf("backward-char")
			return true
		case "a":
			mb.executeCommandOnSelf("beginning-of-line")
			return true
		case "e":
			mb.executeCommandOnSelf("end-of-line")
			return true
		}
	}
	
	// Handle normal character input
	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		mb.executeCommandOnSelf("self-insert-command", event.Rune)
		return true
	}
	
	return false
}

// executeCommandOnSelf executes a command in the context of the minibuffer
func (mb *Minibuffer) executeCommandOnSelf(commandName string, rune ...rune) {
	// Handle special minibuffer-specific commands first
	switch commandName {
	case "delete-backward-char":
		mb.DeleteBackward()
	case "delete-forward-char":
		mb.DeleteForward()
	case "forward-char":
		mb.MoveCursorForward()
	case "backward-char":
		mb.MoveCursorBackward()
	case "beginning-of-line":
		mb.MoveCursorToBeginning()
	case "end-of-line":
		mb.MoveCursorToEnd()
	case "self-insert-command":
		if len(rune) > 0 {
			mb.InsertChar(rune[0])
		}
	}
}

// executeCommand handles M-x command execution
func (mb *Minibuffer) executeCommand(editor *Editor) {
	commandName := mb.content
	
	if cmd, exists := editor.commandRegistry.Get(commandName); exists {
		// Clear command input first
		mb.Clear()
		
		// Execute command (command can set its own message)
		err := cmd.Execute(editor)
		if err != nil {
			mb.SetMessage("Command failed: " + err.Error())
		}
	} else {
		mb.SetMessage("Unknown command: " + commandName)
	}
}

// executeFileOpen handles C-x C-f file opening
func (mb *Minibuffer) executeFileOpen(editor *Editor) {
	filepath := mb.content
	
	// Try to load the file
	buffer, err := NewBufferFromFile(filepath)
	if err != nil {
		mb.SetMessage("Cannot open file: " + filepath)
	} else {
		// Add buffer to editor and switch to it
		editor.AddBuffer(buffer)
		editor.SwitchToBuffer(buffer)
		mb.SetMessage("Opened: " + filepath)
	}
}