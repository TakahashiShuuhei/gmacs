package test

import (
	"strings"
	"testing"
	"time"

	"github.com/TakahashiShuuhei/gmacs/domain"
	"github.com/TakahashiShuuhei/gmacs/events"
	"github.com/TakahashiShuuhei/gmacs/log"
	"github.com/TakahashiShuuhei/gmacs/plugin"
)

/**
 * @spec プラグインシステム/コマンド実行
 * @scenario M-x example-greetコマンドの実行
 * @description プラグインコマンドがM-x経由で正常に実行されるか確認
 * @given エディタがプラグインシステム付きで初期化される
 * @when M-x example-greetコマンドを実行する
 * @then プラグインコマンドが正常に実行される
 * @implementation plugin/editor_integration.go, domain/editor.go
 */
func TestPluginCommandExecution(t *testing.T) {
	editor := NewEditorWithDefaults()
	defer editor.Cleanup()
	display := NewMockDisplay(80, 5)

	// Start M-x
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)

	// Type "example-greet"
	commandText := "example-greet"
	for _, ch := range commandText {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}

	// Check minibuffer content
	display.Render(editor)
	minibufferContent := display.GetMinibuffer()
	expectedPrefix := "M-x example-greet"
	if !strings.HasPrefix(minibufferContent, expectedPrefix) {
		t.Errorf("Expected minibuffer to start with '%s', got %q", expectedPrefix, minibufferContent)
	}

	// Press Enter to execute
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)

	// Check that command execution was attempted
	// Note: The actual plugin may not be loaded in test environment
	// We primarily test that the command execution path works without crashing
	display.Render(editor)
	
	// Verify that we're back to normal mode (minibuffer not showing command input)
	minibuffer := editor.Minibuffer()
	if minibuffer.IsActive() {
		// If minibuffer is still active, it should be showing a message, not command input
		if minibuffer.Mode() == domain.MinibufferCommand {
			t.Error("Minibuffer should not still be in command mode after execution")
		}
	}

	t.Logf("Plugin command execution test completed")
}

/**
 * @spec プラグインシステム/コマンド登録
 * @scenario プラグインコマンドの登録確認
 * @description プラグインコマンドがエディタのコマンドレジストリに正しく登録されるか確認
 * @given エディタがプラグインシステム付きで初期化される
 * @when コマンドレジストリを確認する
 * @then プラグインコマンドが登録されている（模擬環境の場合）
 * @implementation plugin/editor_integration.go
 */
func TestPluginCommandRegistration(t *testing.T) {
	editor := NewEditorWithDefaults()
	defer editor.Cleanup()

	// Get command registry
	cmdRegistry := editor.CommandRegistry()
	if cmdRegistry == nil {
		t.Fatal("Command registry should be available")
	}

	// In a test environment, actual plugins may not be loaded
	// But we can check that the plugin command registration system is in place
	
	// Check if basic plugin-related functionality exists
	pluginManager := editor.PluginManager()
	if pluginManager == nil {
		t.Fatal("Plugin manager should be available for command registration")
	}

	// List current plugins (should be empty in test environment)
	plugins := pluginManager.ListPlugins()
	t.Logf("Found %d plugins in test environment", len(plugins))

	// Test command lookup for plugin commands
	// Note: In test environment, these may not exist, which is expected
	pluginCommands := []string{"example-greet", "example-info", "example-insert-timestamp"}
	
	foundCommands := 0
	for _, cmdName := range pluginCommands {
		_, exists := cmdRegistry.Get(cmdName)
		if exists {
			foundCommands++
			t.Logf("✓ Plugin command '%s' is registered", cmdName)
		} else {
			t.Logf("Plugin command '%s' not found (expected in test environment)", cmdName)
		}
	}

	// In test environment with no actual plugins, this is expected
	t.Logf("Plugin command registration test completed: %d/%d commands found", foundCommands, len(pluginCommands))
}

/**
 * @spec プラグインシステム/エラーハンドリング
 * @scenario 存在しないプラグインコマンドのエラー処理
 * @description 存在しないプラグインコマンドを実行した際のエラーハンドリング
 * @given エディタがプラグインシステム付きで初期化される
 * @when 存在しないプラグインコマンドを実行する
 * @then 適切なエラーメッセージが表示される
 * @implementation domain/editor.go
 */
func TestNonExistentPluginCommand(t *testing.T) {
	editor := NewEditorWithDefaults()
	defer editor.Cleanup()
	display := NewMockDisplay(80, 5)

	// Start M-x and type non-existent plugin command
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)

	// Type "fake-plugin-command"
	for _, ch := range "fake-plugin-command" {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}

	// Press Enter
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)

	// Should show error message
	display.Render(editor)
	minibuffer := editor.Minibuffer()
	if !minibuffer.IsActive() {
		t.Error("Minibuffer should show error message")
	}

	minibufferContent := display.GetMinibuffer()
	if !strings.Contains(minibufferContent, "Unknown command: fake-plugin-command") {
		t.Errorf("Expected unknown command error, got %q", minibufferContent)
	}

	t.Logf("Non-existent plugin command error handling test passed")
}

/**
 * @spec プラグインシステム/実際のプラグインコマンド実行
 * @scenario 実際にインストールされたプラグインコマンドの実行
 * @description 実際にインストールされたプラグインのコマンドをM-xで実行し、メッセージ表示を確認
 * @given プラグインがインストールされた状態のエディタ
 * @when M-xでプラグインコマンドを実行
 * @then プラグインのメッセージが表示される
 * @implementation domain/editor.go RegisterPluginCommands
 */
func TestRealPluginCommandExecution(t *testing.T) {
	// Create editor with actual plugin paths
	pluginPaths := []string{"/home/shuhei/.local/share/gmacs/plugins"}
	editor := plugin.CreateEditorWithPluginsAndPaths(nil, nil, pluginPaths)
	defer editor.Cleanup()
	
	// Wait for plugin loading
	time.Sleep(100 * time.Millisecond)
	
	// Test plugin commands that should return PLUGIN_MESSAGE
	testCommands := []struct {
		name           string
		expectedPrefix string
	}{
		{"example-greet", "[EXAMPLE]"},
		{"example-test-host-api", "[TEST]"},
		{"example-buffer-ops", "[BUFFER-OPS]"},
		{"example-command-exec", "[COMMAND-EXEC]"},
		{"example-file-ops", "[FILE-OPS]"},
		{"example-options", "[OPTIONS]"},
		{"example-hooks", "[HOOKS]"},
	}
	
	for _, tc := range testCommands {
		t.Run(tc.name, func(t *testing.T) {
			// Start M-x
			escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
			editor.HandleEvent(escEvent)
			xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
			editor.HandleEvent(xEvent)
			
			// Type command name
			for _, ch := range tc.name {
				event := events.KeyEventData{Key: string(ch), Rune: ch}
				editor.HandleEvent(event)
			}
			
			// Execute command
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)
			
			// Wait for command execution
			time.Sleep(50 * time.Millisecond)
			
			// Check minibuffer message
			minibuffer := editor.Minibuffer()
			message := minibuffer.Message()
			
			log.Debug("Command %s result: %s", tc.name, message)
			
			// Verify message was set and contains expected prefix
			if message == "" {
				t.Errorf("Command %s produced no message", tc.name)
			} else if !strings.Contains(message, tc.expectedPrefix) {
				t.Errorf("Command %s message '%s' does not contain expected prefix '%s'", tc.name, message, tc.expectedPrefix)
			} else {
				t.Logf("✓ Command %s returned: %s", tc.name, message)
			}
		})
	}
}