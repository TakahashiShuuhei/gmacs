package main

import (
	"fmt"
	"time"

	"github.com/TakahashiShuuhei/gmacs/events"
	"github.com/TakahashiShuuhei/gmacs/lua-config"
	"github.com/TakahashiShuuhei/gmacs/plugin"
)

func runPluginTest() {
	fmt.Println("=== Manual Plugin Test ===")
	
	// Create editor with plugins (using real plugin paths)
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	
	editor := plugin.CreateEditorWithPlugins(configLoader, hookManager)
	defer editor.Cleanup()
	
	fmt.Println("Editor created with plugin system")
	
	// Wait a moment for plugins to load
	time.Sleep(1 * time.Second)
	
	// Check if plugin commands are registered
	cmdRegistry := editor.CommandRegistry()
	
	pluginCommands := []string{"example-greet", "example-info", "example-insert-timestamp"}
	for _, cmdName := range pluginCommands {
		_, exists := cmdRegistry.Get(cmdName)
		if exists {
			fmt.Printf("✓ Plugin command '%s' is registered\n", cmdName)
		} else {
			fmt.Printf("✗ Plugin command '%s' NOT found\n", cmdName)
		}
	}
	
	// Try to execute plugin command directly
	fmt.Println("\n=== Executing example-greet command ===")
	
	// Simulate M-x example-greet
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)
	
	// Type "example-greet"
	for _, ch := range "example-greet" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}
	
	// Press Enter to execute
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)
	
	fmt.Println("Command execution attempted")
	
	// Check minibuffer for result
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() {
		fmt.Printf("Minibuffer message: %s\n", minibuffer.Message())
	} else {
		fmt.Println("No minibuffer message")
	}
	
	// Wait a moment to see any async results
	time.Sleep(2 * time.Second)
	
	fmt.Println("=== Test completed ===")
}

func main() {
	runPluginTest()
}