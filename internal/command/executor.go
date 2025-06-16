package command

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Executor handles the execution of interactive commands (M-x functionality)
type Executor struct {
	registry   *Registry
	input      io.Reader
	output     io.Writer
	lastInput  string
}

// NewExecutor creates a new command executor
func NewExecutor(registry *Registry, input io.Reader, output io.Writer) *Executor {
	return &Executor{
		registry: registry,
		input:    input,
		output:   output,
	}
}

// PromptAndExecute prompts the user for a command and executes it (M-x functionality)
func (e *Executor) PromptAndExecute() error {
	fmt.Fprint(e.output, "M-x ")
	
	scanner := bufio.NewScanner(e.input)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read input")
	}
	
	commandLine := strings.TrimSpace(scanner.Text())
	if commandLine == "" {
		fmt.Fprintln(e.output, "Quit")
		return nil
	}
	
	e.lastInput = commandLine
	
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
		return err
	}
	
	return nil
}

// CompleteCommand provides command completion for the given prefix
func (e *Executor) CompleteCommand(prefix string) []string {
	return e.registry.ListWithPrefix(prefix)
}

// ShowHelp displays help for a command
func (e *Executor) ShowHelp(commandName string) {
	if commandName == "" {
		e.listAllCommands()
		return
	}
	
	fn, exists := e.registry.Get(commandName)
	if !exists {
		fmt.Fprintf(e.output, "Command '%s' not found\n", commandName)
		return
	}
	
	fmt.Fprintf(e.output, "Command: %s\n", fn.Name)
	if fn.Description != "" {
		fmt.Fprintf(e.output, "Description: %s\n", fn.Description)
	}
	if fn.ArgSpec != "" {
		fmt.Fprintf(e.output, "Arguments: %s\n", fn.ArgSpec)
	}
}

// listAllCommands lists all available commands
func (e *Executor) listAllCommands() {
	commands := e.registry.List()
	if len(commands) == 0 {
		fmt.Fprintln(e.output, "No commands registered")
		return
	}
	
	fmt.Fprintln(e.output, "Available commands:")
	
	allFunctions := e.registry.GetAll()
	for _, name := range commands {
		fn := allFunctions[name]
		if fn.Description != "" {
			fmt.Fprintf(e.output, "  %-20s - %s\n", name, fn.Description)
		} else {
			fmt.Fprintf(e.output, "  %s\n", name)
		}
	}
}

// ExecuteDirect executes a command directly by name
func (e *Executor) ExecuteDirect(commandName string, args ...interface{}) error {
	return e.registry.Execute(commandName, args...)
}

// GetLastInput returns the last input entered by the user
func (e *Executor) GetLastInput() string {
	return e.lastInput
}

// FindSimilarCommands finds commands similar to the given name (for typo suggestions)
func (e *Executor) FindSimilarCommands(name string, maxSuggestions int) []string {
	commands := e.registry.List()
	var suggestions []string
	
	// Simple similarity: check if command contains the input as substring
	for _, cmd := range commands {
		if strings.Contains(cmd, name) && cmd != name {
			suggestions = append(suggestions, cmd)
		}
	}
	
	// If not enough substring matches, try prefix matching
	if len(suggestions) < maxSuggestions {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, name) && cmd != name {
				found := false
				for _, existing := range suggestions {
					if existing == cmd {
						found = true
						break
					}
				}
				if !found {
					suggestions = append(suggestions, cmd)
				}
			}
		}
	}
	
	sort.Strings(suggestions)
	
	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}
	
	return suggestions
}

// CreateGlobalExecutor creates an executor using the global registry
func CreateGlobalExecutor(input io.Reader, output io.Writer) *Executor {
	return NewExecutor(globalRegistry, input, output)
}