package test

import (
	"strings"
	"testing"
	"time"

	"github.com/TakahashiShuuhei/gmacs/events"
	"github.com/TakahashiShuuhei/gmacs/log"
)

/**
 * @spec プラグインシステム/BufferInterface API検証
 * @scenario BufferInterface APIのテスト
 * @description テスト用プラグインを使用してBufferInterface APIが正常に動作することを確認
 * @given テスト用プラグインがロードされた状態のエディタ
 * @when BufferInterface の各メソッドを実行
 * @then 期待される結果が返される
 * @implementation plugin/host_api.go BufferInterface実装
 */
func TestPluginBufferAPI(t *testing.T) {
	// テスト用プラグインをインストール
	err := InstallTestPlugin("../test-plugins/buffer-test-plugin")
	if err != nil {
		t.Fatalf("Failed to install test plugin: %v", err)
	}

	// テスト用プラグインが読み込まれるエディタを作成
	editor := NewEditorWithTestPlugins()
	defer editor.Cleanup()

	// プラグインロードの待機
	time.Sleep(100 * time.Millisecond)

	testCases := []struct {
		command         string
		expectedMessage string
		description     string
	}{
		{
			command:         "buffer-test-create",
			expectedMessage: "Buffer created successfully",
			description:     "Buffer creation test",
		},
		{
			command:         "buffer-test-content",
			expectedMessage: "Buffer content test PASSED",
			description:     "Buffer content operations test",
		},
		{
			command:         "buffer-test-cursor",
			expectedMessage: "Cursor test PASSED",
			description:     "Buffer cursor operations test",
		},
		{
			command:         "buffer-test-switch",
			expectedMessage: "Buffer switch test PASSED",
			description:     "Buffer switching test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// M-x コマンドを実行
			escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
			editor.HandleEvent(escEvent)
			xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
			editor.HandleEvent(xEvent)

			// コマンド名を入力
			for _, ch := range tc.command {
				event := events.KeyEventData{Key: string(ch), Rune: ch}
				editor.HandleEvent(event)
			}

			// Enterで実行
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)

			// コマンド実行の待機
			time.Sleep(50 * time.Millisecond)

			// 結果を確認
			minibuffer := editor.Minibuffer()
			message := minibuffer.Message()

			log.Debug("Command %s result: %s", tc.command, message)

			if message == "" {
				t.Errorf("Command %s produced no message", tc.command)
			} else if !strings.Contains(message, tc.expectedMessage) {
				t.Errorf("Command %s: expected message containing '%s', got '%s'", tc.command, tc.expectedMessage, message)
			} else {
				t.Logf("✓ %s: %s", tc.description, message)
			}
		})
	}
}

/**
 * @spec プラグインシステム/FileInterface API検証
 * @scenario ファイル操作APIのテスト
 * @description テスト用プラグインを使用してファイル操作APIが正常に動作することを確認
 * @given テスト用プラグインがロードされた状態のエディタ
 * @when ファイル操作の各メソッドを実行
 * @then 期待される結果が返される
 * @implementation plugin/host_api.go ファイル操作実装
 */
func TestPluginFileAPI(t *testing.T) {
	// テスト用プラグインをインストール
	err := InstallTestPlugin("../test-plugins/file-test-plugin")
	if err != nil {
		t.Fatalf("Failed to install test plugin: %v", err)
	}

	// テスト用プラグインが読み込まれるエディタを作成
	editor := NewEditorWithTestPlugins()
	defer editor.Cleanup()

	// プラグインロードの待機
	time.Sleep(100 * time.Millisecond)

	// Test all file operations in one command to avoid RPC state issues
	testCommand := "file-test-all"
	expectedMessage := "File all test PASSED"
	description := "File operations (create, open, content, save) test"

	// M-x コマンドを実行
	escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
	editor.HandleEvent(escEvent)
	xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
	editor.HandleEvent(xEvent)

	// コマンド名を入力
	for _, ch := range testCommand {
		event := events.KeyEventData{Key: string(ch), Rune: ch}
		editor.HandleEvent(event)
	}

	// Enterで実行
	enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
	editor.HandleEvent(enterEvent)

	// コマンド実行の待機
	time.Sleep(100 * time.Millisecond)

	// 結果を確認
	minibuffer := editor.Minibuffer()
	message := minibuffer.Message()

	log.Debug("Command %s result: %s", testCommand, message)

	if message == "" {
		t.Errorf("Command %s produced no message", testCommand)
	} else if !strings.Contains(message, expectedMessage) {
		t.Errorf("Command %s: expected message containing '%s', got '%s'", testCommand, expectedMessage, message)
	} else {
		t.Logf("✓ %s: %s", description, message)
	}

	// Also test individual commands for backwards compatibility
	individualTests := []struct {
		command         string
		expectedMessage string
		description     string
	}{
		{
			command:         "file-test-create",
			expectedMessage: "File create/open test PASSED",
			description:     "File creation and opening test",
		},
		{
			command:         "file-test-content",
			expectedMessage: "File content test PASSED",
			description:     "File content operations test",
		},
		{
			command:         "file-test-save",
			expectedMessage: "File save test PASSED",
			description:     "File saving test",
		},
	}

	for _, tc := range individualTests {
		t.Run(tc.description, func(t *testing.T) {
			// M-x コマンドを実行
			escEvent := events.KeyEventData{Key: "\x1b", Rune: 0}
			editor.HandleEvent(escEvent)
			xEvent := events.KeyEventData{Key: "x", Rune: 'x'}
			editor.HandleEvent(xEvent)

			// コマンド名を入力
			for _, ch := range tc.command {
				event := events.KeyEventData{Key: string(ch), Rune: ch}
				editor.HandleEvent(event)
			}

			// Enterで実行
			enterEvent := events.KeyEventData{Key: "Enter", Rune: '\n'}
			editor.HandleEvent(enterEvent)

			// コマンド実行の待機
			time.Sleep(50 * time.Millisecond)

			// 結果を確認
			minibuffer := editor.Minibuffer()
			message := minibuffer.Message()

			log.Debug("Command %s result: %s", tc.command, message)

			if message == "" {
				t.Errorf("Command %s produced no message", tc.command)
			} else if !strings.Contains(message, tc.expectedMessage) {
				t.Errorf("Command %s: expected message containing '%s', got '%s'", tc.command, tc.expectedMessage, message)
			} else {
				t.Logf("✓ %s: %s", tc.description, message)
			}
		})
	}
}

/**
 * @spec プラグインシステム/テストプラグイン分離
 * @scenario テスト環境でのプラグイン分離確認
 * @description テスト用プラグインがグローバルプラグインと分離されていることを確認
 * @given テスト用プラグインディレクトリを指定したエディタ
 * @when プラグインリストを取得
 * @then テスト用プラグインのみが読み込まれている
 * @implementation plugin/manager.go NewPluginManagerWithPaths
 */
func TestPluginIsolation(t *testing.T) {
	// テスト用プラグインをインストール
	err := InstallTestPlugin("../test-plugins/buffer-test-plugin")
	if err != nil {
		t.Fatalf("Failed to install test plugin: %v", err)
	}

	// テスト用エディタを作成
	editor := NewEditorWithTestPlugins()
	defer editor.Cleanup()

	// プラグインロードの待機
	time.Sleep(100 * time.Millisecond)

	// プラグインマネージャからプラグインリストを取得
	pluginManager := editor.PluginManager()
	if pluginManager == nil {
		t.Fatal("Plugin manager is nil")
	}

	plugins := pluginManager.ListPlugins()
	t.Logf("Found %d plugins in test environment", len(plugins))

	// テスト用プラグインが読み込まれていることを確認
	var foundTestPlugin bool
	for _, plugin := range plugins {
		t.Logf("Plugin: %s (%s)", plugin.Name, plugin.Description)
		if plugin.Name == "buffer-test-plugin" {
			foundTestPlugin = true
		}
	}

	if !foundTestPlugin {
		t.Error("Test plugin 'buffer-test-plugin' not found in test environment")
	}

	// Test環境には特定のテスト用プラグインのみが読み込まれていることを確認
	// example-pluginはテスト環境でも使用されるためグローバルプラグインとは見なさない
	expectedTestPlugins := []string{"buffer-test-plugin", "example-plugin", "file-test-plugin"}
	foundCount := 0
	
	for _, expected := range expectedTestPlugins {
		found := false
		for _, plugin := range plugins {
			if plugin.Name == expected {
				found = true
				foundCount++
				break
			}
		}
		if !found {
			t.Errorf("Expected test plugin '%s' not found in test environment", expected)
		}
	}
	
	// 予期しないプラグインが読み込まれていないことを確認
	if len(plugins) > len(expectedTestPlugins) {
		t.Errorf("Unexpected plugins found: expected %d plugins, got %d", len(expectedTestPlugins), len(plugins))
		for _, plugin := range plugins {
			found := false
			for _, expected := range expectedTestPlugins {
				if plugin.Name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unexpected plugin found: %s", plugin.Name)
			}
		}
	}

	t.Logf("Plugin isolation test passed: found %d expected test plugins, no unexpected plugins", foundCount)
}