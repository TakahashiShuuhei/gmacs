package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetXDGDataHome(t *testing.T) {
	// 元の環境変数を保存
	originalXDGDataHome := os.Getenv("XDG_DATA_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_DATA_HOME", originalXDGDataHome)
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name            string
		xdgDataHome     string
		home            string
		expectedSuffix  string
	}{
		{
			name:            "XDG_DATA_HOME設定済み",
			xdgDataHome:     "/custom/data",
			home:            "/home/user",
			expectedSuffix:  "/custom/data",
		},
		{
			name:            "XDG_DATA_HOME未設定、HOME設定済み",
			xdgDataHome:     "",
			home:            "/home/user",
			expectedSuffix:  "/home/user/.local/share",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_DATA_HOME", tt.xdgDataHome)
			os.Setenv("HOME", tt.home)

			result := GetXDGDataHome()
			if result != tt.expectedSuffix {
				t.Errorf("GetXDGDataHome() = %v, want %v", result, tt.expectedSuffix)
			}
		})
	}
}

func TestGetXDGConfigHome(t *testing.T) {
	// 元の環境変数を保存
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name              string
		xdgConfigHome     string
		home              string
		expectedSuffix    string
	}{
		{
			name:              "XDG_CONFIG_HOME設定済み",
			xdgConfigHome:     "/custom/config",
			home:              "/home/user",
			expectedSuffix:    "/custom/config",
		},
		{
			name:              "XDG_CONFIG_HOME未設定、HOME設定済み",
			xdgConfigHome:     "",
			home:              "/home/user",
			expectedSuffix:    "/home/user/.config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			os.Setenv("HOME", tt.home)

			result := GetXDGConfigHome()
			if result != tt.expectedSuffix {
				t.Errorf("GetXDGConfigHome() = %v, want %v", result, tt.expectedSuffix)
			}
		})
	}
}

func TestGetDefaultPluginPaths(t *testing.T) {
	// 元の環境変数を保存
	originalXDGDataHome := os.Getenv("XDG_DATA_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_DATA_HOME", originalXDGDataHome)
		os.Setenv("HOME", originalHome)
	}()

	// テスト用の環境設定
	os.Setenv("XDG_DATA_HOME", "/test/data")
	os.Setenv("HOME", "/test/home")

	paths := GetDefaultPluginPaths()

	expectedPaths := []string{
		"/test/data/gmacs/plugins",
		"/usr/share/gmacs/plugins/",
		"/usr/local/share/gmacs/plugins/",
	}

	if len(paths) != len(expectedPaths) {
		t.Fatalf("期待されるパス数 %d, 実際 %d", len(expectedPaths), len(paths))
	}

	for i, expected := range expectedPaths {
		if paths[i] != expected {
			t.Errorf("paths[%d] = %v, want %v", i, paths[i], expected)
		}
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	// 元の環境変数を保存
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		os.Setenv("HOME", originalHome)
	}()

	// テスト用の環境設定
	os.Setenv("XDG_CONFIG_HOME", "/test/config")
	os.Setenv("HOME", "/test/home")

	result := GetDefaultConfigPath()
	expected := "/test/config/gmacs/init.lua"

	if result != expected {
		t.Errorf("GetDefaultConfigPath() = %v, want %v", result, expected)
	}
}

func TestGetPluginConfigPath(t *testing.T) {
	// 元の環境変数を保存
	originalXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", originalXDGConfigHome)
		os.Setenv("HOME", originalHome)
	}()

	// テスト用の環境設定
	os.Setenv("XDG_CONFIG_HOME", "/test/config")
	os.Setenv("HOME", "/test/home")

	result := GetPluginConfigPath()
	expected := "/test/config/gmacs/plugins.lua"

	if result != expected {
		t.Errorf("GetPluginConfigPath() = %v, want %v", result, expected)
	}
}

func TestEnsurePluginDir(t *testing.T) {
	// 一時ディレクトリでテスト
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test", "plugin", "dir")

	// ディレクトリが存在しないことを確認
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Fatalf("テストディレクトリが既に存在します: %v", testPath)
	}

	// ディレクトリ作成
	err := EnsurePluginDir(testPath)
	if err != nil {
		t.Fatalf("EnsurePluginDir() error = %v", err)
	}

	// ディレクトリが作成されたことを確認
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Fatalf("ディレクトリが作成されませんでした: %v", testPath)
	}

	// 既に存在するディレクトリに対してもエラーが出ないことを確認
	err = EnsurePluginDir(testPath)
	if err != nil {
		t.Fatalf("既存ディレクトリでEnsurePluginDir() error = %v", err)
	}
}

func TestIsPluginDir(t *testing.T) {
	// 一時ディレクトリでテスト
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupFunc      func(string) string
		expectedResult bool
	}{
		{
			name: "manifest.jsonが存在する",
			setupFunc: func(baseDir string) string {
				pluginDir := filepath.Join(baseDir, "valid-plugin")
				os.MkdirAll(pluginDir, 0755)
				manifestPath := filepath.Join(pluginDir, "manifest.json")
				os.WriteFile(manifestPath, []byte(`{"name": "test"}`), 0644)
				return pluginDir
			},
			expectedResult: true,
		},
		{
			name: "manifest.jsonが存在しない",
			setupFunc: func(baseDir string) string {
				pluginDir := filepath.Join(baseDir, "invalid-plugin")
				os.MkdirAll(pluginDir, 0755)
				return pluginDir
			},
			expectedResult: false,
		},
		{
			name: "ディレクトリが存在しない",
			setupFunc: func(baseDir string) string {
				return filepath.Join(baseDir, "nonexistent")
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := tt.setupFunc(tempDir)
			result := IsPluginDir(testPath)

			if result != tt.expectedResult {
				t.Errorf("IsPluginDir(%v) = %v, want %v", testPath, result, tt.expectedResult)
			}
		})
	}
}