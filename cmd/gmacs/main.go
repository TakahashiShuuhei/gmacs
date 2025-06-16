package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TakahashiShuuhei/gmacs/internal/display"
)

func main() {
	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		showHelp()
		return
	}
	
	// Check for simple mode flag
	// TODO: Remove simple mode once proper terminal key input is implemented
	if len(os.Args) > 1 && (os.Args[1] == "-s" || os.Args[1] == "--simple") {
		fmt.Println("Starting gmacs in simple mode...")
		editor := display.NewSimpleEditor()
		if err := editor.Run(); err != nil {
			log.Fatalf("Editor error: %v", err)
		}
		return
	}
	
	// Create and run the normal editor
	fmt.Println("Starting gmacs in normal mode...")
	fmt.Println("Note: If you have issues with key input, try: gmacs --simple")
	editor := display.NewEditor()
	
	if err := editor.Run(); err != nil {
		log.Fatalf("Editor error: %v", err)
	}
}

func showHelp() {
	fmt.Println("gmacs - Go Emacs-like Editor")
	fmt.Println("Version 0.0.1")
	fmt.Println()
	fmt.Println("Usage: gmacs [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println("  -s, --simple  Start in simple mode (recommended for testing)")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  Normal mode:  Full terminal UI (may have key input issues)")
	fmt.Println("  Simple mode:  Text-based interface (better compatibility)")
	fmt.Println()
	fmt.Println("Key bindings (Normal mode):")
	fmt.Println("  M-x           Execute extended command")
	fmt.Println("  C-x C-c       Quit")
	fmt.Println("  C-g           Cancel/Quit current operation")
	fmt.Println()
	fmt.Println("Commands (Simple mode):")
	fmt.Println("  M-x           Enter command mode")
	fmt.Println("  help          Show help")
	fmt.Println("  quit          Quit the editor")
	fmt.Println("  show          Show buffer content")
	fmt.Println()
	fmt.Println("Available extended commands:")
	fmt.Println("  version       Show version information")
	fmt.Println("  hello         Say hello")
	fmt.Println("  list-commands List all available commands")
	fmt.Println("  buffer-info   Show buffer information")
	fmt.Println("  echo          Echo arguments")
}