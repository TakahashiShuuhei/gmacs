package domain

import (
	"sort"
	"strings"
	"github.com/TakahashiShuuhei/gmacs/log"
)

// CommandFunc represents a command function
type CommandFunc func(editor *Editor) error

// Command represents an interactive command that can be executed
type Command struct {
	name string
	fn   CommandFunc
}

// NewCommand creates a new command with the given name and function
func NewCommand(name string, fn CommandFunc) *Command {
	return &Command{
		name: name,
		fn:   fn,
	}
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Execute(editor *Editor) error {
	return c.fn(editor)
}

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands map[string]*Command
}

func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]*Command),
	}
	
	// Register built-in commands
	registry.RegisterFunc("version", func(editor *Editor) error {
		version := "gmacs 0.1.0 - Emacs-like text editor in Go"
		log.Info("Executing version command")
		editor.SetMinibufferMessage(version)
		return nil
	})
	
	registry.RegisterFunc("list-commands", func(editor *Editor) error {
		commands := registry.List()
		message := "Available commands: " + strings.Join(commands, ", ")
		log.Info("Listing available commands")
		editor.SetMinibufferMessage(message)
		return nil
	})
	
	registry.RegisterFunc("clear-buffer", func(editor *Editor) error {
		buffer := editor.CurrentBuffer()
		if buffer != nil {
			buffer.Clear()
			log.Info("Buffer cleared")
			editor.SetMinibufferMessage("Buffer cleared")
		}
		return nil
	})
	
	return registry
}

// Quit command for C-x C-c
func Quit(editor *Editor) error {
	log.Info("Quit command executed")
	editor.Quit()
	return nil
}

// DeleteBackwardChar command for C-h (backspace)
func DeleteBackwardChar(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer != nil {
		buffer.DeleteBackward()
		log.Debug("Deleted backward character")
	}
	return nil
}

// DeleteChar command for C-d (delete-char)
func DeleteChar(editor *Editor) error {
	buffer := editor.CurrentBuffer()
	if buffer != nil {
		buffer.DeleteForward()
		log.Debug("Deleted forward character")
	}
	return nil
}

// FindFile command for C-x C-f (find-file)
func FindFile(editor *Editor) error {
	editor.minibuffer.StartFileInput()
	log.Info("Find file command started")
	return nil
}

// KeyboardQuit command for C-g (keyboard-quit)
func KeyboardQuit(editor *Editor) error {
	// Clear minibuffer if active
	if editor.minibuffer.IsActive() {
		editor.minibuffer.Clear()
		log.Info("Keyboard quit: cleared minibuffer")
	} else {
		// Reset any partial key sequences
		editor.keyBindings.ResetSequence()
		log.Info("Keyboard quit: reset key sequences")
	}
	return nil
}

func (cr *CommandRegistry) Register(cmd *Command) {
	cr.commands[cmd.Name()] = cmd
}

func (cr *CommandRegistry) RegisterFunc(name string, fn CommandFunc) {
	cmd := NewCommand(name, fn)
	cr.Register(cmd)
}

func (cr *CommandRegistry) Get(name string) (*Command, bool) {
	cmd, exists := cr.commands[name]
	return cmd, exists
}

func (cr *CommandRegistry) List() []string {
	names := make([]string, 0, len(cr.commands))
	for name := range cr.commands {
		names = append(names, name)
	}
	sort.Strings(names) // Sort alphabetically for consistent order
	return names
}