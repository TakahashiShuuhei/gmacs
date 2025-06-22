package luaconfig

import (
	"os"
	"path/filepath"
	"github.com/TakahashiShuuhei/gmacs/log"
)

// ConfigLoader handles loading and managing Lua configuration files
type ConfigLoader struct {
	vm *LuaVM
	configPath string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		vm: NewLuaVM(),
	}
}

// Close closes the configuration loader and its Lua VM
func (cl *ConfigLoader) Close() {
	if cl.vm != nil {
		cl.vm.Close()
		cl.vm = nil
	}
}

// LoadUserConfig loads the user's configuration file
func (cl *ConfigLoader) LoadUserConfig() error {
	configPath, err := cl.FindConfigFile()
	if err != nil {
		return err
	}
	
	if configPath == "" {
		log.Info("No configuration file found")
		return nil
	}
	
	return cl.LoadConfig(configPath)
}

// LoadConfig loads a specific configuration file
func (cl *ConfigLoader) LoadConfig(configPath string) error {
	if cl.vm == nil {
		return &ConfigError{Message: "ConfigLoader is closed"}
	}
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ConfigError{Message: "Config file does not exist: " + configPath}
	}
	
	cl.configPath = configPath
	return cl.vm.LoadConfig(configPath)
}

// FindConfigFile searches for a configuration file in standard locations
func (cl *ConfigLoader) FindConfigFile() (string, error) {
	// Check for ~/.gmacs/init.lua
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warn("Could not get user home directory: %v", err)
		return "", nil
	}
	
	configPaths := []string{
		filepath.Join(homeDir, ".gmacs", "init.lua"),
		filepath.Join(homeDir, ".gmacs.lua"),
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			log.Info("Found config file: %s", path)
			return path, nil
		}
	}
	
	log.Info("No config file found in standard locations")
	return "", nil
}

// FindPluginConfigFile searches for plugin configuration file in standard locations
func (cl *ConfigLoader) FindPluginConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warn("Could not get user home directory: %v", err)
		return "", nil
	}
	
	// XDG Base Directory specification compliant path
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(homeDir, ".config")
	}
	
	configPaths := []string{
		filepath.Join(xdgConfigHome, "gmacs", "plugins.lua"),
		filepath.Join(homeDir, ".gmacs", "plugins.lua"),
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			log.Info("Found plugin config file: %s", path)
			return path, nil
		}
	}
	
	log.Info("No plugin config file found in standard locations")
	return "", nil
}

// LoadPluginConfig loads the plugin configuration file
func (cl *ConfigLoader) LoadPluginConfig() error {
	configPath, err := cl.FindPluginConfigFile()
	if err != nil {
		return err
	}
	
	if configPath == "" {
		log.Info("No plugin configuration file found")
		return nil
	}
	
	return cl.LoadConfig(configPath)
}

// ReloadConfig reloads the current configuration file
func (cl *ConfigLoader) ReloadConfig() error {
	if cl.configPath == "" {
		return &ConfigError{Message: "No config file loaded"}
	}
	
	// Close and recreate VM to ensure clean state
	cl.vm.Close()
	cl.vm = NewLuaVM()
	
	return cl.vm.LoadConfig(cl.configPath)
}

// GetVM returns the Lua virtual machine
func (cl *ConfigLoader) GetVM() *LuaVM {
	return cl.vm
}

// GetConfigPath returns the path of the currently loaded config file
func (cl *ConfigLoader) GetConfigPath() string {
	return cl.configPath
}