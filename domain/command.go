package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

// Command represents an interactive command that can be executed
type Command interface {
	Name() string
	Execute(editor *Editor) error
}

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}
	
	// Register built-in commands
	registry.Register(&VersionCommand{})
	
	return registry
}

func (cr *CommandRegistry) Register(cmd Command) {
	cr.commands[cmd.Name()] = cmd
	log.Debug("Registered command: %s", cmd.Name())
}

func (cr *CommandRegistry) Get(name string) (Command, bool) {
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

// VersionCommand displays version information
type VersionCommand struct{}

func (vc *VersionCommand) Name() string {
	return "version"
}

func (vc *VersionCommand) Execute(editor *Editor) error {
	version := "gmacs 0.1.0 - Emacs-like text editor in Go"
	log.Info("Executing version command")
	
	// Display version in minibuffer
	editor.SetMinibufferMessage(version)
	
	return nil
}