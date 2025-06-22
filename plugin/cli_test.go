package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewPluginCLI(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	if cli == nil {
		t.Fatal("NewPluginCLI() returned nil")
	}

	if cli.manager == nil {
		t.Error("PluginCLI.manager is nil")
	}

	if cli.builder == nil {
		t.Error("PluginCLI.builder is nil")
	}
}

func TestPluginCLI_InstallCommand_InvalidArgs(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 引数なしでエラーになることを確認
	err = cli.InstallCommand([]string{})
	if err == nil {
		t.Error("Expected error for empty args, but got nil")
	}

	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("Expected usage error, but got: %v", err)
	}
}

func TestPluginCLI_InstallCommand_LocalPath(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// シンプルなプラグインを作成
	files := map[string]string{
		"main.go": `package main
import "fmt"
func main() {
	fmt.Println("Hello from CLI test plugin!")
}`,
		"go.mod": `module cli-test-plugin
go 1.22.2`,
		"manifest.json": `{
	"name": "cli-test-plugin",
	"version": "1.0.0",
	"description": "CLI test plugin",
	"author": "Test",
	"binary": "cli-test-plugin"
}`,
	}

	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// ローカルパスからインストール
	err = cli.InstallCommand([]string{tempDir})
	if err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// インストールされたプラグインを確認
	plugins := cli.manager.ListInstalledPlugins()

	found := false
	for _, plugin := range plugins {
		if strings.Contains(plugin.Name, "local-source-") {
			found = true
			t.Logf("Installed plugin: %s", plugin.Name)
			break
		}
	}

	if !found {
		t.Error("Plugin was not found in the installed plugins list")
	}
}

func TestPluginCLI_UpdateCommand_InvalidArgs(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 引数なしでエラーになることを確認
	err = cli.UpdateCommand([]string{})
	if err == nil {
		t.Error("Expected error for empty args, but got nil")
	}

	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("Expected usage error, but got: %v", err)
	}
}

func TestPluginCLI_RemoveCommand_InvalidArgs(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 引数なしでエラーになることを確認
	err = cli.RemoveCommand([]string{})
	if err == nil {
		t.Error("Expected error for empty args, but got nil")
	}

	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("Expected usage error, but got: %v", err)
	}
}

func TestPluginCLI_RemoveCommand_NotFound(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 存在しないプラグインの削除
	err = cli.RemoveCommand([]string{"non-existent-plugin"})
	if err == nil {
		t.Error("Expected error for non-existent plugin, but got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, but got: %v", err)
	}
}

func TestPluginCLI_ListCommand(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 引数は無視されるはずなので、任意の引数で実行
	err = cli.ListCommand([]string{})
	if err != nil {
		t.Errorf("ListCommand() error = %v", err)
	}

	// 引数ありでも動作するはず
	err = cli.ListCommand([]string{"ignored"})
	if err != nil {
		t.Errorf("ListCommand() with args error = %v", err)
	}
}

func TestPluginCLI_InfoCommand_InvalidArgs(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 引数なしでエラーになることを確認
	err = cli.InfoCommand([]string{})
	if err == nil {
		t.Error("Expected error for empty args, but got nil")
	}

	if !strings.Contains(err.Error(), "usage:") {
		t.Errorf("Expected usage error, but got: %v", err)
	}
}

func TestPluginCLI_InfoCommand_NotFound(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// 存在しないプラグインの情報取得
	err = cli.InfoCommand([]string{"non-existent-plugin"})
	if err == nil {
		t.Error("Expected error for non-existent plugin, but got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, but got: %v", err)
	}
}

func TestPluginCLI_HelpCommand(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// ヘルプコマンドは常に成功するはず
	err = cli.HelpCommand([]string{})
	if err != nil {
		t.Errorf("HelpCommand() error = %v", err)
	}

	// 引数ありでも動作するはず
	err = cli.HelpCommand([]string{"ignored"})
	if err != nil {
		t.Errorf("HelpCommand() with args error = %v", err)
	}
}

func TestPluginStateToString(t *testing.T) {
	tests := []struct {
		state    PluginState
		expected string
	}{
		{PluginStateUnloaded, "Unloaded"},
		{PluginStateLoading, "Loading"},
		{PluginStateLoaded, "Loaded"},
		{PluginStateError, "Error"},
		{PluginState(999), "Unknown"}, // 未知の状態
	}

	for _, test := range tests {
		result := pluginStateToString(test.state)
		if result != test.expected {
			t.Errorf("pluginStateToString(%v) = %s, expected %s", test.state, result, test.expected)
		}
	}
}

func TestPluginCLI_InstallAndRemove_Integration(t *testing.T) {
	cli, err := NewPluginCLI()
	if err != nil {
		t.Fatalf("NewPluginCLI() error = %v", err)
	}

	// テスト用の一時ディレクトリを作成
	tempDir := t.TempDir()

	// シンプルなプラグインを作成
	files := map[string]string{
		"main.go": `package main
import "fmt"
func main() {
	fmt.Println("Hello from integration test plugin!")
}`,
		"go.mod": `module integration-test-plugin
go 1.22.2`,
		"manifest.json": `{
	"name": "integration-test-plugin",
	"version": "1.0.0",
	"description": "Integration test plugin",
	"author": "Test",
	"binary": "integration-test-plugin"
}`,
	}

	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// 1. インストール
	err = cli.InstallCommand([]string{tempDir})
	if err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// 2. 一覧で確認
	plugins := cli.manager.ListInstalledPlugins()

	var installedPluginName string
	for _, plugin := range plugins {
		if strings.Contains(plugin.Name, "local-source-") {
			installedPluginName = plugin.Name
			break
		}
	}

	if installedPluginName == "" {
		t.Fatal("Plugin was not found after installation")
	}

	t.Logf("Installed plugin name: %s", installedPluginName)

	// 3. 情報表示
	err = cli.InfoCommand([]string{installedPluginName})
	if err != nil {
		t.Errorf("InfoCommand() error = %v", err)
	}

	// 4. 削除
	err = cli.RemoveCommand([]string{installedPluginName})
	if err != nil {
		t.Errorf("RemoveCommand() error = %v", err)
	}

	// 5. 削除後の確認
	plugins = cli.manager.ListPlugins()

	for _, plugin := range plugins {
		if plugin.Name == installedPluginName {
			t.Errorf("Plugin %s still exists after removal", installedPluginName)
		}
	}
}