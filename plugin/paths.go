package plugin

import (
	"os"
	"path/filepath"
)

// XDG Base Directory Specification準拠のパス取得

// GetXDGDataHome returns XDG_DATA_HOME or fallback to ~/.local/share
func GetXDGDataHome() string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return dataHome
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share")
}

// GetXDGConfigHome returns XDG_CONFIG_HOME or fallback to ~/.config
func GetXDGConfigHome() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config")
}

// GetDefaultPluginPaths returns default plugin search paths
func GetDefaultPluginPaths() []string {
	userDataDir := GetXDGDataHome()
	
	return []string{
		filepath.Join(userDataDir, "gmacs", "plugins"),     // ユーザーローカル
		"/usr/share/gmacs/plugins/",                        // システムワイド
		"/usr/local/share/gmacs/plugins/",                  // ローカルシステム
	}
}

// GetDefaultConfigPath returns default config file path
func GetDefaultConfigPath() string {
	configHome := GetXDGConfigHome()
	return filepath.Join(configHome, "gmacs", "init.lua")
}

// GetPluginConfigPath returns plugin-specific config path
func GetPluginConfigPath() string {
	configHome := GetXDGConfigHome()
	return filepath.Join(configHome, "gmacs", "plugins.lua")
}

// EnsurePluginDir creates plugin directory if it doesn't exist
func EnsurePluginDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// IsPluginDir checks if directory contains a valid plugin
func IsPluginDir(path string) bool {
	manifestPath := filepath.Join(path, "manifest.json")
	if _, err := os.Stat(manifestPath); err != nil {
		return false
	}
	return true
}