package domain

import (
	"strings"
	"github.com/TakahashiShuuhei/gmacs/core/log"
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

func (cr *CommandRegistry) Register(cmd *Command) {
	cr.commands[cmd.Name()] = cmd
	log.Debug("Registered command: %s", cmd.Name())
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
	return names
}