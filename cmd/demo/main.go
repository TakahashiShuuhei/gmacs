package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/TakahashiShuuhei/gmacs/internal/command"
	"github.com/TakahashiShuuhei/gmacs/internal/keymap"
)

func main() {
	fmt.Println("gmacs Command System Demo")
	fmt.Println("========================")
	
	// Register some basic commands
	registerBasicCommands()
	
	// Create and test keymap
	testKeymap()
	
	// Test M-x functionality
	testMxCommand()
}

func registerBasicCommands() {
	fmt.Println("\n--- Registering Commands ---")
	
	// Register a simple hello command
	command.Register("hello", "Say hello", "", func(args ...interface{}) error {
		name := "World"
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		fmt.Printf("Hello, %s!\n", name)
		return nil
	})
	
	// Register a version command
	command.Register("version", "Show version", "", func(args ...interface{}) error {
		fmt.Println("gmacs version 0.0.1")
		return nil
	})
	
	// Register a list-commands command
	command.Register("list-commands", "List all available commands", "", func(args ...interface{}) error {
		commands := command.List()
		fmt.Println("Available commands:")
		for _, cmd := range commands {
			fmt.Printf("  %s\n", cmd)
		}
		return nil
	})
	
	// Register a quit command
	command.Register("quit", "Quit the application", "", func(args ...interface{}) error {
		fmt.Println("Goodbye!")
		os.Exit(0)
		return nil
	})
	
	fmt.Printf("Registered %d commands\n", len(command.List()))
}

func testKeymap() {
	fmt.Println("\n--- Testing Keymap ---")
	
	// Create a keymap
	km := keymap.New("global")
	
	// Parse and bind some key sequences
	testBindings := []struct {
		keys    string
		command string
	}{
		{"C-x C-f", "find-file"},
		{"C-x C-s", "save-buffer"},
		{"M-x", "execute-extended-command"},
		{"C-g", "keyboard-quit"},
	}
	
	for _, binding := range testBindings {
		seq, err := keymap.ParseKeySequence(binding.keys)
		if err != nil {
			fmt.Printf("Error parsing key sequence '%s': %v\n", binding.keys, err)
			continue
		}
		
		err = km.Bind(seq, binding.command)
		if err != nil {
			fmt.Printf("Error binding '%s': %v\n", binding.keys, err)
			continue
		}
		
		fmt.Printf("Bound %s -> %s\n", binding.keys, binding.command)
	}
	
	// Test lookup
	fmt.Println("\nTesting key lookups:")
	testSeq, _ := keymap.ParseKeySequence("C-x C-f")
	if binding, exists := km.Lookup(testSeq); exists {
		fmt.Printf("C-x C-f -> %s ✓\n", binding.Command)
	} else {
		fmt.Printf("C-x C-f -> not found ✗\n")
	}
	
	// Show all bindings
	fmt.Println("\nAll key bindings:")
	allBindings := km.GetAllBindings()
	for keyStr, binding := range allBindings {
		fmt.Printf("  %-10s -> %s\n", keyStr, binding.Command)
	}
}

func testMxCommand() {
	fmt.Println("\n--- Testing M-x Command System ---")
	
	// Simulate M-x input
	testCommands := []string{
		"version",
		"hello",
		"hello Go",
		"list-commands",
	}
	
	for _, cmdLine := range testCommands {
		fmt.Printf("\nSimulating: M-x %s\n", cmdLine)
		
		parts := strings.Fields(cmdLine)
		if len(parts) == 0 {
			continue
		}
		
		commandName := parts[0]
		var args []interface{}
		for _, arg := range parts[1:] {
			args = append(args, arg)
		}
		
		err := command.Execute(commandName, args...)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
	
	// Test completion
	fmt.Println("\n--- Testing Command Completion ---")
	testPrefixes := []string{"h", "list", "v", "xyz"}
	
	for _, prefix := range testPrefixes {
		completions := command.ListWithPrefix(prefix)
		fmt.Printf("Completions for '%s': %v\n", prefix, completions)
	}
	
	// Test help
	fmt.Println("\n--- Testing Help System ---")
	executor := command.CreateGlobalExecutor(os.Stdin, os.Stdout)
	executor.ShowHelp("hello")
}