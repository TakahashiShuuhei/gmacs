package domain


// MinibufferMode represents the current mode of the minibuffer
type MinibufferMode int

const (
	MinibufferInactive MinibufferMode = iota
	MinibufferCommand                 // M-x command input
	MinibufferMessage                 // Displaying a message
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
	if mb.mode != MinibufferCommand {
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
	if mb.mode != MinibufferCommand || mb.cursor == 0 {
		return
	}
	
	runes := []rune(mb.content)
	before := runes[:mb.cursor-1]
	after := runes[mb.cursor:]
	
	mb.content = string(append(before, after...))
	mb.cursor--
	
}

// GetDisplayText returns the text to display in the minibuffer
func (mb *Minibuffer) GetDisplayText() string {
	switch mb.mode {
	case MinibufferCommand:
		return mb.prompt + mb.content
	case MinibufferMessage:
		return mb.message
	default:
		return ""
	}
}