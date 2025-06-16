package command

import (
	"strings"
	"testing"
)

func TestRegistry(t *testing.T) {
	registry := NewRegistry()
	
	// Test registering a command
	executed := false
	err := registry.Register("test-command", "A test command", "", func(args ...interface{}) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}
	
	// Test command exists
	fn, exists := registry.Get("test-command")
	if !exists {
		t.Fatal("Command should exist")
	}
	if fn.Name != "test-command" {
		t.Errorf("Expected name 'test-command', got '%s'", fn.Name)
	}
	
	// Test executing command
	err = registry.Execute("test-command")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}
	if !executed {
		t.Error("Command should have been executed")
	}
	
	// Test listing commands
	commands := registry.List()
	if len(commands) != 1 || commands[0] != "test-command" {
		t.Errorf("Expected ['test-command'], got %v", commands)
	}
	
	// Test unregistering
	err = registry.Unregister("test-command")
	if err != nil {
		t.Fatalf("Failed to unregister command: %v", err)
	}
	
	_, exists = registry.Get("test-command")
	if exists {
		t.Error("Command should not exist after unregistering")
	}
}

func TestRegistryErrors(t *testing.T) {
	registry := NewRegistry()
	
	// Test empty name
	err := registry.Register("", "desc", "", func(args ...interface{}) error { return nil })
	if err == nil {
		t.Error("Should error on empty command name")
	}
	
	// Test nil handler
	err = registry.Register("test", "desc", "", nil)
	if err == nil {
		t.Error("Should error on nil handler")
	}
	
	// Test duplicate registration
	handler := func(args ...interface{}) error { return nil }
	err = registry.Register("test", "desc", "", handler)
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}
	
	err = registry.Register("test", "desc", "", handler)
	if err == nil {
		t.Error("Should error on duplicate registration")
	}
	
	// Test executing non-existent command
	err = registry.Execute("non-existent")
	if err == nil {
		t.Error("Should error on non-existent command")
	}
}

func TestListWithPrefix(t *testing.T) {
	registry := NewRegistry()
	
	handler := func(args ...interface{}) error { return nil }
	registry.Register("find-file", "Find file", "", handler)
	registry.Register("find-grep", "Find grep", "", handler)
	registry.Register("save-buffer", "Save buffer", "", handler)
	registry.Register("save-some-buffers", "Save some buffers", "", handler)
	
	// Test prefix matching
	findCommands := registry.ListWithPrefix("find")
	if len(findCommands) != 2 {
		t.Errorf("Expected 2 commands with 'find' prefix, got %d", len(findCommands))
	}
	
	saveCommands := registry.ListWithPrefix("save")
	if len(saveCommands) != 2 {
		t.Errorf("Expected 2 commands with 'save' prefix, got %d", len(saveCommands))
	}
	
	noCommands := registry.ListWithPrefix("xyz")
	if len(noCommands) != 0 {
		t.Errorf("Expected 0 commands with 'xyz' prefix, got %d", len(noCommands))
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Clear any existing registrations
	globalRegistry = NewRegistry()
	
	executed := false
	err := Register("global-test", "Global test command", "", func(args ...interface{}) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to register global command: %v", err)
	}
	
	err = Execute("global-test")
	if err != nil {
		t.Fatalf("Failed to execute global command: %v", err)
	}
	if !executed {
		t.Error("Global command should have been executed")
	}
	
	commands := List()
	if len(commands) != 1 || commands[0] != "global-test" {
		t.Errorf("Expected ['global-test'], got %v", commands)
	}
}

func TestExecutor(t *testing.T) {
	registry := NewRegistry()
	
	executed := false
	registry.Register("test-cmd", "Test command", "", func(args ...interface{}) error {
		executed = true
		return nil
	})
	
	input := strings.NewReader("test-cmd\n")
	output := &strings.Builder{}
	
	executor := NewExecutor(registry, input, output)
	
	err := executor.PromptAndExecute()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}
	
	if !executed {
		t.Error("Command should have been executed")
	}
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "M-x") {
		t.Error("Output should contain M-x prompt")
	}
}

func TestExecutorCompletion(t *testing.T) {
	registry := NewRegistry()
	
	handler := func(args ...interface{}) error { return nil }
	registry.Register("find-file", "Find file", "", handler)
	registry.Register("find-grep", "Find grep", "", handler)
	registry.Register("save-buffer", "Save buffer", "", handler)
	
	input := strings.NewReader("")
	output := &strings.Builder{}
	executor := NewExecutor(registry, input, output)
	
	completions := executor.CompleteCommand("find")
	if len(completions) != 2 {
		t.Errorf("Expected 2 completions for 'find', got %d", len(completions))
	}
	
	completions = executor.CompleteCommand("save")
	if len(completions) != 1 {
		t.Errorf("Expected 1 completion for 'save', got %d", len(completions))
	}
}

func TestExecutorHelp(t *testing.T) {
	registry := NewRegistry()
	
	registry.Register("test-cmd", "A test command for help", "s", func(args ...interface{}) error {
		return nil
	})
	
	input := strings.NewReader("")
	output := &strings.Builder{}
	executor := NewExecutor(registry, input, output)
	
	executor.ShowHelp("test-cmd")
	outputStr := output.String()
	
	if !strings.Contains(outputStr, "test-cmd") {
		t.Error("Help should contain command name")
	}
	if !strings.Contains(outputStr, "A test command for help") {
		t.Error("Help should contain description")
	}
	if !strings.Contains(outputStr, "Arguments: s") {
		t.Error("Help should contain argument spec")
	}
}