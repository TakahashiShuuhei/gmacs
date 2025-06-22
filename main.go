package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TakahashiShuuhei/gmacs/cli"
	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
	gmacslog "github.com/TakahashiShuuhei/gmacs/log"
	"github.com/TakahashiShuuhei/gmacs/lua-config"
)

//go:embed lua-config/default.lua
var defaultConfig string

func main() {
	if err := gmacslog.Init(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer gmacslog.Close()

	gmacslog.Info("gmacs starting up")

	display := cli.NewDisplay()
	terminal := cli.NewTerminal()

	gmacslog.Debug("Initializing terminal")
	if err := terminal.Init(); err != nil {
		gmacslog.Error("Failed to initialize terminal: %v", err)
		log.Fatal("Failed to initialize terminal:", err)
	}
	defer terminal.Restore()

	// Always create editor with Lua configuration support
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Register Lua API (this also registers built-in commands)
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	if err := apiBindings.RegisterGmacsAPI(); err != nil {
		gmacslog.Error("Failed to register Lua API: %v", err)
		log.Fatal("Failed to register Lua API:", err)
	}
	
	// Load default configuration first
	gmacslog.Info("Loading default configuration")
	if err := configLoader.GetVM().ExecuteString(defaultConfig); err != nil {
		gmacslog.Error("Failed to load default config: %v", err)
		log.Fatal("Failed to load default config:", err)
	}
	
	// Then load user configuration if available
	configPath := findConfigFile()
	if configPath != "" {
		gmacslog.Info("Loading user config: %s", configPath)
		if err := configLoader.LoadConfig(configPath); err != nil {
			gmacslog.Error("Failed to load user config: %v", err)
		}
	} else {
		gmacslog.Info("No user config file found, using defaults only")
	}
	
	// Ensure cleanup on exit
	defer editor.Cleanup()
	
	width, height := display.Size()
	resizeEvent := events.ResizeEventData{
		Width:  width,
		Height: height,
	}
	editor.HandleEvent(resizeEvent)

	gmacslog.Debug("Initial render")
	display.Render(editor)

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	gmacslog.Info("Entering main loop")
	needsRender := false
	
	for editor.IsRunning() {
		select {
		case event := <-terminal.EventChan():
			gmacslog.Debug("Received event: %T", event)
			editor.EventQueue().Push(event)
		case <-ticker.C:
			for {
				event, hasEvent := editor.EventQueue().Pop()
				if !hasEvent {
					break
				}
				gmacslog.Debug("Processing event: %T", event)
				
				// Handle resize events for display as well
				if resizeEvent, ok := event.(events.ResizeEventData); ok {
					display.Resize(resizeEvent.Width, resizeEvent.Height)
				}
				
				editor.HandleEvent(event)
				needsRender = true
			}
			
			// Only render if there were events or if we need to render
			if needsRender {
				display.Render(editor)
				needsRender = false
			}
		}
	}

	// Clear screen and reset terminal state before exiting
	gmacslog.Debug("Clearing screen on exit")
	display.ClearAndExit()

	gmacslog.Info("gmacs shutting down")
}

// findConfigFile searches for a configuration file in standard locations
func findConfigFile() string {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		gmacslog.Warn("Could not get user home directory: %v", err)
		return ""
	}
	
	// Check standard config file locations
	configPaths := []string{
		filepath.Join(homeDir, ".gmacs", "init.lua"),
		filepath.Join(homeDir, ".gmacs.lua"),
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			gmacslog.Info("Found config file: %s", path)
			return path
		}
	}
	
	gmacslog.Info("No config file found in standard locations")
	return ""
}