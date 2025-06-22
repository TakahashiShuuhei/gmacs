package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

// PluginCLI はプラグインCLIコマンドを管理する
type PluginCLI struct {
	manager *PluginManager
	builder *PluginBuilder
}

// NewPluginCLI は新しいPluginCLIを作成する
func NewPluginCLI() (*PluginCLI, error) {
	manager := NewPluginManagerWithPaths(GetDefaultPluginPaths())

	builder, err := NewPluginBuilder()
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin builder: %v", err)
	}

	return &PluginCLI{
		manager: manager,
		builder: builder,
	}, nil
}

// InstallCommand はプラグインをインストールする
func (cli *PluginCLI) InstallCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: install <repository|local-path> [ref]")
	}

	target := args[0]
	ref := "main"
	if len(args) > 1 {
		ref = args[1]
	}

	fmt.Printf("Installing plugin from %s...\n", target)

	var req BuildRequest
	if strings.HasPrefix(target, "/") || strings.HasPrefix(target, "./") || strings.HasPrefix(target, "../") {
		// ローカルパス
		req = BuildRequest{
			LocalPath: target,
			Force:     false,
		}
	} else {
		// Gitリポジトリ
		req = BuildRequest{
			Repository: target,
			Ref:        ref,
			Force:      false,
		}
	}

	result, err := cli.builder.BuildFromRepository(req)
	if err != nil {
		return fmt.Errorf("installation failed: %v", err)
	}

	fmt.Printf("Plugin installed successfully:\n")
	fmt.Printf("  Name: %s\n", result.PluginName)
	fmt.Printf("  Version: %s\n", result.Version)
	fmt.Printf("  Binary: %s\n", result.BinaryPath)
	fmt.Printf("  From Cache: %t\n", result.FromCache)

	return nil
}

// UpdateCommand はプラグインを更新する
func (cli *PluginCLI) UpdateCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: update <plugin-name|repository> [ref]")
	}

	pluginName := args[0]
	ref := "main"
	if len(args) > 1 {
		ref = args[1]
	}

	// プラグインが既にインストールされているかチェック
	plugins := cli.manager.ListInstalledPlugins()

	var found *PluginInfo
	for _, plugin := range plugins {
		if plugin.Name == pluginName {
			found = &plugin
			break
		}
	}

	if found == nil {
		// 未インストールの場合はリポジトリとして扱う
		fmt.Printf("Plugin '%s' not found, installing from repository...\n", pluginName)
		req := BuildRequest{
			Repository: pluginName,
			Ref:        ref,
			Force:      true, // 更新時は強制ビルド
		}

		result, err := cli.builder.BuildFromRepository(req)
		if err != nil {
			return fmt.Errorf("installation failed: %v", err)
		}

		fmt.Printf("Plugin installed successfully:\n")
		fmt.Printf("  Name: %s\n", result.PluginName)
		fmt.Printf("  Version: %s\n", result.Version)
		fmt.Printf("  Binary: %s\n", result.BinaryPath)
		return nil
	}

	fmt.Printf("Updating plugin '%s'...\n", pluginName)
	
	// TODO: プラグインの元のリポジトリ情報を記録・取得して更新
	// 現在は簡易実装として再インストール指示
	fmt.Printf("To update plugin '%s', please reinstall with:\n", pluginName)
	fmt.Printf("  gmacs plugin install <original-repository> %s\n", ref)
	
	return nil
}

// RemoveCommand はプラグインを削除する
func (cli *PluginCLI) RemoveCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: remove <plugin-name>")
	}

	pluginName := args[0]

	// プラグインディレクトリを探す
	var pluginDir string
	for _, basePath := range GetDefaultPluginPaths() {
		candidatePath := filepath.Join(basePath, pluginName)
		if IsPluginDir(candidatePath) {
			pluginDir = candidatePath
			break
		}
	}

	if pluginDir == "" {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	fmt.Printf("Removing plugin '%s' from %s...\n", pluginName, pluginDir)

	// プラグインディレクトリを削除
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("failed to remove plugin directory: %v", err)
	}

	fmt.Printf("Plugin '%s' removed successfully.\n", pluginName)
	return nil
}

// ListCommand は利用可能なプラグインを一覧表示する
func (cli *PluginCLI) ListCommand(args []string) error {
	fmt.Println("Listing installed plugins...")

	plugins := cli.manager.ListInstalledPlugins()

	if len(plugins) == 0 {
		fmt.Println("No plugins installed.")
		return nil
	}

	// テーブル形式で表示
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tSTATE\tENABLED\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t-----\t-------\t-----------")

	for _, plugin := range plugins {
		state := pluginStateToString(plugin.State)
		enabled := "No"
		if plugin.Enabled {
			enabled = "Yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			plugin.Name,
			plugin.Version,
			state,
			enabled,
			plugin.Description)
	}

	w.Flush()
	return nil
}

// InfoCommand はプラグインの詳細情報を表示する
func (cli *PluginCLI) InfoCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: info <plugin-name>")
	}

	pluginName := args[0]

	// プラグインディレクトリを探す
	var pluginDir string
	for _, basePath := range GetDefaultPluginPaths() {
		candidatePath := filepath.Join(basePath, pluginName)
		if IsPluginDir(candidatePath) {
			pluginDir = candidatePath
			break
		}
	}

	if pluginDir == "" {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// マニフェストを読み込み
	manifest, err := cli.builder.loadManifest(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %v", err)
	}

	// バイナリファイル情報
	binaryPath := filepath.Join(pluginDir, manifest.Binary)
	var binaryInfo string
	if stat, err := os.Stat(binaryPath); err == nil {
		binaryInfo = fmt.Sprintf("%s (%d bytes, modified: %s)",
			binaryPath,
			stat.Size(),
			stat.ModTime().Format(time.RFC3339))
	} else {
		binaryInfo = fmt.Sprintf("%s (not found)", binaryPath)
	}

	// 情報表示
	fmt.Printf("Plugin Information: %s\n", pluginName)
	fmt.Printf("==========================================\n")
	fmt.Printf("Name:        %s\n", manifest.Name)
	fmt.Printf("Version:     %s\n", manifest.Version)
	fmt.Printf("Description: %s\n", manifest.Description)
	fmt.Printf("Author:      %s\n", manifest.Author)
	fmt.Printf("Directory:   %s\n", pluginDir)
	fmt.Printf("Binary:      %s\n", binaryInfo)
	fmt.Printf("Manifest:    %s\n", filepath.Join(pluginDir, "manifest.json"))

	if len(manifest.Dependencies) > 0 {
		fmt.Printf("Dependencies:\n")
		for _, dep := range manifest.Dependencies {
			fmt.Printf("  - %s\n", dep)
		}
	}

	if manifest.MinGmacs != "" {
		fmt.Printf("Min gmacs:   %s\n", manifest.MinGmacs)
	}

	return nil
}

// HelpCommand はCLIヘルプを表示する
func (cli *PluginCLI) HelpCommand(args []string) error {
	fmt.Println("gmacs plugin management commands:")
	fmt.Println()
	fmt.Println("  install <repo|path> [ref]  Install plugin from repository or local path")
	fmt.Println("  update <name|repo> [ref]   Update installed plugin")
	fmt.Println("  remove <name>              Remove installed plugin")
	fmt.Println("  list                       List all installed plugins")
	fmt.Println("  info <name>                Show detailed plugin information")
	fmt.Println("  help                       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gmacs plugin install github.com/user/my-plugin")
	fmt.Println("  gmacs plugin install ./local-plugin")
	fmt.Println("  gmacs plugin install github.com/user/plugin v1.2.3")
	fmt.Println("  gmacs plugin update my-plugin")
	fmt.Println("  gmacs plugin remove my-plugin")
	fmt.Println("  gmacs plugin list")
	fmt.Println("  gmacs plugin info my-plugin")

	return nil
}

// pluginStateToString converts PluginState to human-readable string
func pluginStateToString(state PluginState) string {
	switch state {
	case PluginStateUnloaded:
		return "Unloaded"
	case PluginStateLoading:
		return "Loading"
	case PluginStateLoaded:
		return "Loaded"
	case PluginStateError:
		return "Error"
	default:
		return "Unknown"
	}
}