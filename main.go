package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TakahashiShuuhei/gmacs/cli"
	"github.com/TakahashiShuuhei/gmacs/events"
	gmacslog "github.com/TakahashiShuuhei/gmacs/log"
	"github.com/TakahashiShuuhei/gmacs/lua-config"
	"github.com/TakahashiShuuhei/gmacs/plugin"
)

//go:embed lua-config/default.lua
var defaultConfig string

func main() {
	if err := gmacslog.Init(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer gmacslog.Close()

	// Check for subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "plugin":
			handlePluginCommand()
			return
		case "--help", "-h", "help":
			showHelp()
			return
		case "--version", "-v", "version":
			showVersion()
			return
		}
	}

	gmacslog.Info("gmacs starting up")

	display := cli.NewDisplay()
	terminal := cli.NewTerminal()

	gmacslog.Debug("Initializing terminal")
	if err := terminal.Init(); err != nil {
		gmacslog.Error("Failed to initialize terminal: %v", err)
		log.Fatal("Failed to initialize terminal:", err)
	}
	defer terminal.Restore()

	// Always create editor with Lua configuration and plugin support
	configLoader := luaconfig.NewConfigLoader()
	hookManager := luaconfig.NewHookManager()
	
	// Resolve plugin directories (main.go controls this for testability)
	pluginPaths := resolvePluginPaths()
	gmacslog.Info("Plugin search paths: %v", pluginPaths)
	
	// Create editor with plugin system
	editor := plugin.CreateEditorWithPluginsAndPaths(configLoader, hookManager, pluginPaths)
	
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
	
	// Load plugin configuration if available
	if err := plugin.LoadPluginConfigIfExists(configLoader); err != nil {
		gmacslog.Error("Failed to load plugin config: %v", err)
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

// resolvePluginPaths resolves plugin search paths for the current environment
func resolvePluginPaths() []string {
	// Check if we're in test mode (no plugins for tests)
	if isTestMode() {
		gmacslog.Info("Test mode detected, using empty plugin paths")
		return []string{} // Empty paths for tests
	}
	
	// Use default plugin paths for normal operation
	return plugin.GetDefaultPluginPaths()
}

// isTestMode checks if we're running in test mode
func isTestMode() bool {
	// Check for test-specific environment variables or conditions
	if os.Getenv("GMACS_TEST_MODE") == "1" {
		return true
	}
	
	// Check if we're being run by go test
	if len(os.Args) > 0 && filepath.Base(os.Args[0]) == "gmacs.test" {
		return true
	}
	
	// Check for test binary names
	executable, err := os.Executable()
	if err == nil {
		baseName := filepath.Base(executable)
		if baseName == "gmacs.test" || baseName == "__debug_bin" {
			return true
		}
	}
	
	return false
}

// handlePluginCommand handles plugin subcommands
func handlePluginCommand() {
	if len(os.Args) < 3 {
		showPluginHelp()
		return
	}

	cli, err := plugin.NewPluginCLI()
	if err != nil {
		gmacslog.Error("Failed to create plugin CLI: %v", err)
		log.Fatal("Failed to create plugin CLI:", err)
	}

	subcommand := os.Args[2]
	args := os.Args[3:]

	switch subcommand {
	case "install":
		err = cli.InstallCommand(args)
	case "update":
		err = cli.UpdateCommand(args)
	case "remove", "uninstall":
		err = cli.RemoveCommand(args)
	case "list":
		err = cli.ListCommand(args)
	case "info":
		err = cli.InfoCommand(args)
	case "help", "--help", "-h":
		err = cli.HelpCommand(args)
	default:
		log.Printf("Unknown plugin command: %s\n", subcommand)
		showPluginHelp()
		os.Exit(1)
	}

	if err != nil {
		gmacslog.Error("Plugin command failed: %v", err)
		log.Fatal("Plugin command failed:", err)
	}
}

// showPluginHelp shows plugin command help
func showPluginHelp() {
	log.Println("Usage: gmacs plugin <command> [args...]")
	log.Println()
	log.Println("Available commands:")
	log.Println("  install <repo|path> [ref]  Install plugin from repository or local path")
	log.Println("  update <name|repo> [ref]   Update installed plugin")
	log.Println("  remove <name>              Remove installed plugin")
	log.Println("  list                       List all installed plugins")
	log.Println("  info <name>                Show detailed plugin information")
	log.Println("  help                       Show this help message")
	log.Println()
	log.Println("Examples:")
	log.Println("  gmacs plugin install github.com/user/my-plugin")
	log.Println("  gmacs plugin install ./local-plugin")
	log.Println("  gmacs plugin list")
	log.Println("  gmacs plugin remove my-plugin")
}

// showHelp shows general gmacs help
func showHelp() {
	log.Println("gmacs - Emacs-like text editor written in Go")
	log.Println()
	log.Println("Usage:")
	log.Println("  gmacs [file...]           Start the editor")
	log.Println("  gmacs plugin <command>    Manage plugins")
	log.Println("  gmacs --help              Show this help")
	log.Println("  gmacs --version           Show version information")
	log.Println()
	log.Println("Plugin commands:")
	log.Println("  gmacs plugin list         List installed plugins")
	log.Println("  gmacs plugin install     Install a plugin")
	log.Println("  gmacs plugin remove      Remove a plugin")
	log.Println("  gmacs plugin help        Show plugin help")
	log.Println()
	log.Println("For more information, visit: https://github.com/TakahashiShuuhei/gmacs")
}

// showVersion shows version information
func showVersion() {
	log.Println("gmacs version 0.1.0")
	log.Println("Go-based Emacs-like text editor with plugin support")
}