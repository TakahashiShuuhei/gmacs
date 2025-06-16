// Package display provides UI components for gmacs
//
// simple_ui.go - Temporary workaround for terminal key input issues
// TODO: Remove this file once proper raw mode terminal input is implemented
package display

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/buffer"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/window"
)

// SimpleEditor is a simplified text-based editor for testing/debugging
type SimpleEditor struct {
	currentWin *window.Window
	registry   *command.Registry
	running    bool
	input      *bufio.Scanner
	output     io.Writer
}

// NewSimpleEditor creates a new simple editor for testing
func NewSimpleEditor() *SimpleEditor {
	// Create a default buffer and window
	buf := buffer.New("*scratch*")
	buf.SetText("Welcome to gmacs (Simple Mode)!\n\nThis is the scratch buffer.\nType 'M-x' (literally) to execute commands.\nType 'help' for available commands.\nType 'quit' to exit.")
	
	win := window.New(buf, 20, 80)
	
	editor := &SimpleEditor{
		currentWin: win,
		registry:   command.GetGlobalRegistry(),
		running:    true,
		input:      bufio.NewScanner(os.Stdin),
		output:     os.Stdout,
	}
	
	// Register basic commands
	editor.registerCommands()
	
	return editor
}

// Run starts the simple editor loop
func (e *SimpleEditor) Run() error {
	fmt.Fprintln(e.output, "=== gmacs Simple Mode ===")
	fmt.Fprintln(e.output, "Type commands directly:")
	fmt.Fprintln(e.output, "- 'M-x' to enter command mode")
	fmt.Fprintln(e.output, "- 'help' for help")
	fmt.Fprintln(e.output, "- 'quit' to exit")
	fmt.Fprintln(e.output, "- 'show' to show buffer content")
	fmt.Fprintln(e.output)
	
	for e.running {
		fmt.Fprint(e.output, "gmacs> ")
		
		if !e.input.Scan() {
			break
		}
		
		line := strings.TrimSpace(e.input.Text())
		if line == "" {
			continue
		}
		
		if err := e.handleCommand(line); err != nil {
			fmt.Fprintf(e.output, "Error: %v\n", err)
		}
	}
	
	return nil
}

// handleCommand processes a command
func (e *SimpleEditor) handleCommand(input string) error {
	switch input {
	case "M-x":
		return e.executeExtendedCommand()
	case "help":
		e.showHelp()
		return nil
	case "quit", "exit":
		e.running = false
		fmt.Fprintln(e.output, "Goodbye!")
		return nil
	case "show":
		e.showBuffer()
		return nil
	case "clear":
		// Simple clear
		for i := 0; i < 10; i++ {
			fmt.Fprintln(e.output)
		}
		return nil
	default:
		// Try to execute as a direct command
		return e.registry.Execute(input)
	}
}

// executeExtendedCommand handles M-x style command execution
func (e *SimpleEditor) executeExtendedCommand() error {
	fmt.Fprint(e.output, "M-x ")
	
	if !e.input.Scan() {
		return fmt.Errorf("failed to read command")
	}
	
	commandLine := strings.TrimSpace(e.input.Text())
	if commandLine == "" {
		fmt.Fprintln(e.output, "Quit")
		return nil
	}
	
	// Parse command and arguments
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return nil
	}
	
	commandName := parts[0]
	
	// Convert string arguments to interface{} slice
	var args []interface{}
	for _, arg := range parts[1:] {
		args = append(args, arg)
	}
	
	// Execute the command
	err := e.registry.Execute(commandName, args...)
	if err != nil {
		fmt.Fprintf(e.output, "Error: %v\n", err)
		return nil
	}
	
	fmt.Fprintf(e.output, "Executed: %s\n", commandName)
	return nil
}

// showHelp displays help information
func (e *SimpleEditor) showHelp() {
	fmt.Fprintln(e.output, "\n=== gmacs Simple Mode Help ===")
	fmt.Fprintln(e.output, "Available commands:")
	fmt.Fprintln(e.output, "  M-x          - Execute extended command")
	fmt.Fprintln(e.output, "  help         - Show this help")
	fmt.Fprintln(e.output, "  quit/exit    - Exit the editor")
	fmt.Fprintln(e.output, "  show         - Show buffer content")
	fmt.Fprintln(e.output, "  clear        - Clear screen")
	fmt.Fprintln(e.output)
	
	fmt.Fprintln(e.output, "Extended commands (use with M-x):")
	commands := e.registry.List()
	allFunctions := e.registry.GetAll()
	for _, name := range commands {
		fn := allFunctions[name]
		if fn.Description != "" {
			fmt.Fprintf(e.output, "  %-20s - %s\n", name, fn.Description)
		} else {
			fmt.Fprintf(e.output, "  %s\n", name)
		}
	}
	fmt.Fprintln(e.output)
}

// showBuffer displays the current buffer content
func (e *SimpleEditor) showBuffer() {
	fmt.Fprintln(e.output, "\n=== Buffer Content ===")
	fmt.Fprintf(e.output, "Buffer: %s", e.currentWin.Buffer().Name())
	if e.currentWin.Buffer().IsModified() {
		fmt.Fprint(e.output, " *")
	}
	fmt.Fprintln(e.output)
	
	lines := e.currentWin.GetVisibleText()
	for i, line := range lines {
		fmt.Fprintf(e.output, "%3d: %s\n", i+1, line)
	}
	
	cursor := e.currentWin.Cursor()
	fmt.Fprintf(e.output, "Cursor: Line %d, Column %d\n", cursor.Line()+1, cursor.Col()+1)
	fmt.Fprintln(e.output, "======================")
}

// registerCommands registers basic commands
func (e *SimpleEditor) registerCommands() {
	// Version command
	e.registry.Register("version", "Show gmacs version", "", func(args ...interface{}) error {
		fmt.Fprintln(e.output, "gmacs version 0.0.1 - Go Emacs-like Editor (Simple Mode)")
		return nil
	})
	
	// Hello command
	e.registry.Register("hello", "Say hello", "", func(args ...interface{}) error {
		name := "World"
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		fmt.Fprintf(e.output, "Hello, %s!\n", name)
		return nil
	})
	
	// List commands
	e.registry.Register("list-commands", "List all available commands", "", func(args ...interface{}) error {
		commands := e.registry.List()
		fmt.Fprintf(e.output, "Available commands (%d): %s\n", len(commands), strings.Join(commands, ", "))
		return nil
	})
	
	// Buffer info
	e.registry.Register("buffer-info", "Show buffer information", "", func(args ...interface{}) error {
		buf := e.currentWin.Buffer()
		fmt.Fprintf(e.output, "Buffer: %s\n", buf.Name())
		fmt.Fprintf(e.output, "Lines: %d\n", buf.LineCount())
		fmt.Fprintf(e.output, "Modified: %t\n", buf.IsModified())
		if buf.Filename() != "" {
			fmt.Fprintf(e.output, "File: %s\n", buf.Filename())
		}
		return nil
	})
	
	// Echo command for testing
	e.registry.Register("echo", "Echo arguments", "", func(args ...interface{}) error {
		var parts []string
		for _, arg := range args {
			parts = append(parts, fmt.Sprintf("%v", arg))
		}
		fmt.Fprintf(e.output, "Echo: %s\n", strings.Join(parts, " "))
		return nil
	})
}