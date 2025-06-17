package display

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/config"
	"github.com/TakahashiShuuhei/gmacs/internal/input"
	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
	"github.com/TakahashiShuuhei/gmacs/internal/window"
)

// PluginRegistrar is a function type for registering plugin commands
type PluginRegistrar func(editor *Editor, registry *command.Registry)

// Global registry for plugin registrars
var pluginRegistrars []PluginRegistrar

// RegisterPlugin registers a plugin command registrar
func RegisterPlugin(registrar PluginRegistrar) {
	pluginRegistrars = append(pluginRegistrars, registrar)
}

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
	// バッファ管理
	buffers []*buffer.Buffer // バッファリスト
	// ウィンドウサイズ変更検知用
	resizeSignal chan os.Signal // SIGWINCHシグナル受信用チャンネル
}

// NewEditor creates a new editor instance
func NewEditor() *Editor {
	terminal := NewStandardTerminal()
	keyboard := input.CreateStandardKeyboard()
	rawKeyboard, _ := input.NewRawKeyboard() // Ignore error for now
	minibuffer := NewMinibuffer(terminal, keyboard)

	// Create a default scratch buffer
	scratchBuf := buffer.New("*scratch*")
	scratchBuf.SetText("Welcome to gmacs!\n\nThis is the scratch buffer.\nType M-x to execute commands.")

	width, height := terminal.Size()
	fmt.Printf("DEBUG: Initial terminal size: %dx%d\n", width, height) // DEBUG
	win := window.New(scratchBuf, height-2, width)                     // Reserve space for minibuffer

	// Create global keymap
	km := keymap.New("global")

	editor := &Editor{
		terminal:     terminal,
		keyboard:     keyboard,
		rawKeyboard:  rawKeyboard,
		minibuffer:   minibuffer,
		currentWin:   win,
		keymap:       km,
		running:      true,
		registry:     command.GetGlobalRegistry(),
		buffers:      []*buffer.Buffer{scratchBuf}, // バッファリストに追加
		resizeSignal: make(chan os.Signal, 1),      // ウィンドウサイズ変更検知用
	}

	// SIGWINCHシグナル（ウィンドウサイズ変更）を監視
	signal.Notify(editor.resizeSignal, syscall.SIGWINCH)

	// Register basic editor commands
	editor.registerEditorCommands()

	// Register all plugins
	editor.registerPlugins()

	// Load Lua configuration (includes default key bindings)
	editor.loadLuaConfig()

	return editor
}

// GetCurrentBuffer returns the currently active buffer
func (e *Editor) GetCurrentBuffer() *buffer.Buffer {
	return e.currentWin.Buffer()
}

// FindBuffer finds a buffer by name, returns index or -1 if not found
func (e *Editor) FindBuffer(name string) int {
	for i, buf := range e.buffers {
		if buf.Name() == name {
			return i
		}
	}
	return -1
}

// AddBuffer adds a new buffer to the buffer list
func (e *Editor) AddBuffer(buf *buffer.Buffer) int {
	e.buffers = append(e.buffers, buf)
	return len(e.buffers) - 1
}

// SwitchToBuffer switches to the buffer at the given index
func (e *Editor) SwitchToBuffer(index int) error {
	if index < 0 || index >= len(e.buffers) {
		return fmt.Errorf("buffer index %d out of range", index)
	}

	buf := e.buffers[index]
	e.currentWin.SetBuffer(buf)

	return nil
}

// SwitchToBufferByName switches to a buffer by name, creates if not found
func (e *Editor) SwitchToBufferByName(name string) error {
	index := e.FindBuffer(name)
	if index == -1 {
		// Buffer doesn't exist, create new one
		buf := buffer.New(name)
		index = e.AddBuffer(buf)
	}

	return e.SwitchToBuffer(index)
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
		(r >= 0xFF00 && r <= 0xFFEF) // Halfwidth and Fullwidth Forms
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

	// Use channels for concurrent input handling
	keyEventChan := make(chan *input.KeyEvent, 1)
	lineChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Start input goroutine
	go func() {
		for e.running {
			if rawModeEnabled {
				keyEvent, err := e.rawKeyboard.ReadKey()
				if err != nil {
					errorChan <- err
					return
				}
				keyEventChan <- keyEvent
			} else {
				line, err := e.keyboard.ReadLine()
				if err != nil {
					errorChan <- err
					return
				}
				lineChan <- line
			}
		}
	}()

	for e.running {
		e.redraw()

		// Handle resize signals and input concurrently
		select {
		case <-e.resizeSignal:
			e.handleWindowResize()
			// Continue immediately to redraw with new size
			continue

		case keyEvent := <-keyEventChan:
			// Handle key event from raw mode
			if err := e.handleKeyEvent(keyEvent); err != nil {
				e.minibuffer.ShowError(err)
			}

		case line := <-lineChan:
			// Handle line input from fallback mode
			if err := e.handleInput(line); err != nil {
				e.minibuffer.ShowError(err)
			}

		case err := <-errorChan:
			// Handle input errors
			return err
		}
	}

	// Stop signal monitoring
	signal.Stop(e.resizeSignal)
	close(e.resizeSignal)

	return nil
}

// handleWindowResize handles terminal window resize events
func (e *Editor) handleWindowResize() {
	fmt.Printf("DEBUG: SIGWINCH signal received\n") // DEBUG

	// Update terminal size
	e.terminal.UpdateSize()

	// Get new terminal size
	newWidth, newHeight := e.terminal.Size()
	fmt.Printf("DEBUG: New terminal size: %dx%d\n", newWidth, newHeight) // DEBUG

	// Update window size (reserve space for minibuffer)
	if e.currentWin != nil {
		oldHeight := e.currentWin.Height()
		oldWidth := e.currentWin.Width()
		e.currentWin.SetSize(newHeight-2, newWidth)
		fmt.Printf("DEBUG: Window size changed from %dx%d to %dx%d\n", oldWidth, oldHeight, newWidth, newHeight-2) // DEBUG
	}

	// Clear screen to avoid artifacts
	e.terminal.Clear()

	// Show resize notification in minibuffer (with debug info)
	e.minibuffer.ShowMessage(fmt.Sprintf("Terminal resized to %dx%d (W=%d H=%d)", newWidth, newHeight, newWidth, newHeight))
}

// handleKeyEvent processes a key event from raw keyboard input
func (e *Editor) handleKeyEvent(keyEvent *input.KeyEvent) error {
	// ミニバッファがアクティブな場合はキーイベントを処理しない
	// ミニバッファが独自に入力を処理する
	if e.minibuffer.IsActive() {
		return nil
	}

	// Handle key sequence building
	return e.handleKeySequence(keyEvent.Key)
}

// handleKeySequence handles multi-key sequences (prefix keys)
func (e *Editor) handleKeySequence(key keymap.Key) error {
	// Add the key to current sequence
	e.keySequence = append(e.keySequence, key)

	// Check for C-g (quit/cancel sequence)
	keyStr := key.String()
	if keyStr == "C-g" {
		e.keySequence = nil
		e.minibuffer.ShowMessage("Quit")
		return nil
	}

	// Try to look up current sequence
	if binding, exists := e.keymap.Lookup(e.keySequence); exists {
		// Found a complete binding
		e.keySequence = nil // Reset sequence

		// Handle special commands
		switch binding.Command {
		case "execute-extended-command":
			return e.executeExtendedCommand()
		default:
			return e.registry.Execute(binding.Command, binding.Args...)
		}
	}

	// Check if this could be a prefix for a longer sequence
	if e.isPrefixSequence(e.keySequence) {
		// Show current sequence in minibuffer
		seqStr := e.keySequence.String()
		e.minibuffer.ShowMessage(seqStr + "-")
		return nil
	}

	// No binding found and not a prefix
	if len(e.keySequence) == 1 {
		// Single key - check if it's printable
		if key.Special == "" && key.Char >= 32 && key.Char <= 126 && !key.Ctrl && !key.Alt {
			e.keySequence = nil
			return e.selfInsertCommand(key.Char)
		}
	}

	// Unknown sequence
	seqStr := e.keySequence.String()
	e.keySequence = nil // Reset sequence
	e.minibuffer.ShowMessage(fmt.Sprintf("'%s' is undefined", seqStr))
	return nil
}

// isPrefixSequence checks if the current sequence could be a prefix for a longer binding
func (e *Editor) isPrefixSequence(seq keymap.KeySequence) bool {
	seqStr := seq.String()
	allBindings := e.keymap.GetAllBindings()

	for bindingStr := range allBindings {
		if len(bindingStr) > len(seqStr) && strings.HasPrefix(bindingStr, seqStr+" ") {
			return true
		}
	}
	return false
}

// handleInput processes user input (fallback for non-raw mode)
func (e *Editor) handleInput(input string) error {
	// ミニバッファがアクティブな場合はキーイベントを処理しない
	if e.minibuffer.IsActive() {
		return nil
	}

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

	// コマンドが正常に実行された場合、コマンド自身がメッセージを表示する
	return nil
}

// redraw redraws the editor interface
func (e *Editor) redraw() {
	// 常に全画面再描画を使用（差分描画は一旦無効）
	e.fullRedraw()
}

// fullRedraw performs a complete redraw of the interface
func (e *Editor) fullRedraw() {
	e.terminal.Clear()
	e.drawBuffer()
	e.drawStatusLine()
	e.drawMinibuffer()
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

			if displayWidth+charWidth > width-1 {
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
	statusLine := height - 1 // Status line is second from bottom (minibuffer is at bottom)

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

	// Version command
	e.registry.Register("version", "Show gmacs version", "", func(args ...interface{}) error {
		e.minibuffer.ShowMessage("gmacs version 0.0.1 - Go Emacs-like Editor")
		return nil
	})

	// Terminal size command (debug)
	e.registry.Register("terminal-size", "Show current terminal size", "", func(args ...interface{}) error {
		width, height := e.terminal.Size()
		e.minibuffer.ShowMessage(fmt.Sprintf("Terminal size: %dx%d (width x height)", width, height))
		return nil
	})

	// Manual resize command (debug)
	e.registry.Register("force-resize", "Force terminal resize handling", "", func(args ...interface{}) error {
		e.handleWindowResize()
		return nil
	})

	// Compile commands
	e.registry.Register("compile", "Compile current project", "", func(args ...interface{}) error {
		e.minibuffer.ShowMessage("Compilation started... (not implemented yet)")
		return nil
	})

	e.registry.Register("kill-compilation", "Kill running compilation", "", func(args ...interface{}) error {
		e.minibuffer.ShowMessage("Compilation killed (not implemented yet)")
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

	// newlineコマンドを登録（Enterキー用）
	e.registry.Register("newline", "Insert newline or complete input", "", func(args ...interface{}) error {
		// ミニバッファがアクティブな場合は何もしない（ミニバッファが処理する）
		if e.minibuffer.IsActive() {
			return nil
		}

		// 通常モードでは改行を挿入
		// ただし、ミニバッファにメッセージが表示されている場合はクリアする
		if e.minibuffer.HasMessage() {
			e.minibuffer.Clear()
			return nil
		}

		// バッファに改行を挿入
		return e.selfInsertCommand('\n')
	})

}

// registerPlugins registers all registered plugins
func (e *Editor) registerPlugins() {
	for _, registrar := range pluginRegistrars {
		registrar(e, e.registry)
	}
}

// drawMinibuffer draws the minibuffer based on its current state
func (e *Editor) drawMinibuffer() {
	// ミニバッファがメッセージを持っている場合のみ描画
	if e.minibuffer.HasMessage() {
		// ShowMessage("") を呼んで現在のメッセージを再描画
		// これにより一貫した描画処理を使用
		currentMessage := e.minibuffer.prompt
		if currentMessage != "" {
			width, height := e.terminal.Size()
			e.terminal.MoveCursor(height, 1)
			e.terminal.ClearLine()

			if len(currentMessage) > width-1 {
				currentMessage = currentMessage[:width-4] + "..."
			}
			e.terminal.Print(currentMessage)
		}
	}
}
