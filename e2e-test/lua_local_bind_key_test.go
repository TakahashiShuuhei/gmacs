package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	luaconfig "github.com/TakahashiShuuhei/gmacs/core/lua-config"
	gmacslog "github.com/TakahashiShuuhei/gmacs/core/log"
)

/**
 * @spec lua-config/local_bind_key_basic
 * @scenario Lua gmacs.local_bind_key() 基本動作
 * @description gmacs.local_bind_key()でモード固有のキーバインドを設定する
 * @given エディタを作成し、Lua設定システムを初期化
 * @when gmacs.local_bind_key("fundamental-mode", "C-t", "version")を実行
 * @then fundamental-modeにC-tキーバインドが登録される
 * @implementation lua-config/api_bindings.go, domain/editor.go
 */
func TestLuaLocalBindKeyBasic(t *testing.T) {
	// Initialize logger for test
	if err := gmacslog.Init(); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer gmacslog.Close()
	
	// Create editor with Lua config
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	// Register API bindings
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Execute Lua code to bind key
	luaCode := `gmacs.local_bind_key("fundamental-mode", "C-t", "version")`
	err = configLoader.GetVM().ExecuteString(luaCode)
	if err != nil {
		t.Fatalf("Failed to execute Lua code: %v", err)
	}
	
	t.Log("Successfully executed local_bind_key Lua command")
	
	// Get the fundamental mode and check if the key binding was added
	modeManager := editor.ModeManager()
	fundamentalMode, exists := modeManager.GetMajorModeByName("fundamental-mode")
	if !exists {
		t.Fatal("fundamental-mode should exist")
	}
	
	t.Logf("Found fundamental-mode: %s (addr: %p)", fundamentalMode.Name(), fundamentalMode)
	
	keyBindings := fundamentalMode.KeyBindings()
	if keyBindings == nil {
		t.Fatal("fundamental-mode should have key bindings")
	}
	
	t.Logf("fundamental-mode has key bindings object (addr: %p)", keyBindings)
	
	// Check if the key binding exists by simulating the key press
	// "C-t" means Ctrl+t, so call ProcessKeyPress with key="t", ctrl=true, meta=false
	cmd, matched, continuing := keyBindings.ProcessKeyPress("t", true, false)
	if !matched {
		t.Error("Key binding C-t should be registered in fundamental-mode")
	} else {
		t.Log("Key binding C-t found in fundamental-mode")
	}
	
	if continuing {
		t.Error("C-t should be a complete binding, not a prefix")
	}
	
	// Verify the command works by executing it
	if cmd != nil {
		err := cmd(editor)
		if err != nil {
			t.Errorf("Command should execute without error: %v", err)
		}
		
		// Check if version message is displayed
		minibuffer := editor.Minibuffer()
		if !minibuffer.IsActive() {
			t.Error("Minibuffer should be active after version command")
		}
		
		if !strings.Contains(minibuffer.Message(), "gmacs") {
			t.Errorf("Expected version message, got: %s", minibuffer.Message())
		}
	}
}

/**
 * @spec lua-config/local_bind_key_unknown_mode
 * @scenario 未知のモードへのキーバインド試行
 * @description 存在しないモードにキーバインドを設定しようとした場合のエラー処理
 * @given エディタとLua設定システムを初期化
 * @when gmacs.local_bind_key("non-existent-mode", "C-t", "version")を実行
 * @then Luaエラーが発生し、Unknown modeエラーメッセージが含まれる
 * @implementation lua-config/api_bindings.go, domain/editor.go
 */
func TestLuaLocalBindKeyUnknownMode(t *testing.T) {
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Try to bind key to non-existent mode
	luaCode := `gmacs.local_bind_key("non-existent-mode", "C-t", "version")`
	err = configLoader.GetVM().ExecuteString(luaCode)
	if err == nil {
		t.Error("Expected error when binding to unknown mode")
	}
	
	if !strings.Contains(err.Error(), "Unknown mode") {
		t.Errorf("Expected 'Unknown mode' error, got: %v", err)
	}
}

/**
 * @spec lua-config/local_bind_key_unknown_command
 * @scenario 未知のコマンドでのキーバインド試行
 * @description 存在しないコマンドにキーバインドを設定しようとした場合のエラー処理
 * @given エディタとLua設定システムを初期化
 * @when gmacs.local_bind_key("fundamental-mode", "C-t", "non-existent-command")を実行
 * @then Luaエラーが発生し、Unknown commandエラーメッセージが含まれる
 * @implementation lua-config/api_bindings.go, domain/editor.go
 */
func TestLuaLocalBindKeyUnknownCommand(t *testing.T) {
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Try to bind key to non-existent command
	luaCode := `gmacs.local_bind_key("fundamental-mode", "C-t", "non-existent-command")`
	err = configLoader.GetVM().ExecuteString(luaCode)
	if err == nil {
		t.Error("Expected error when binding to unknown command")
	}
	
	if !strings.Contains(err.Error(), "Unknown command") {
		t.Errorf("Expected 'Unknown command' error, got: %v", err)
	}
}

/**
 * @spec lua-config/local_bind_key_minor_mode
 * @scenario マイナーモードへのキーバインド設定
 * @description gmacs.local_bind_key()でマイナーモードにキーバインドを設定する
 * @given エディタとLua設定システムを初期化
 * @when gmacs.local_bind_key("auto-a-mode", "C-a", "version")を実行
 * @then auto-a-modeにC-aキーバインドが登録される
 * @implementation lua-config/api_bindings.go, domain/editor.go
 */
func TestLuaLocalBindKeyMinorMode(t *testing.T) {
	configLoader := luaconfig.NewConfigLoader()
	defer configLoader.Close()
	
	hookManager := luaconfig.NewHookManager()
	editor := domain.NewEditorWithConfig(configLoader, hookManager)
	
	apiBindings := luaconfig.NewAPIBindings(editor, configLoader.GetVM())
	err := apiBindings.RegisterGmacsAPI()
	if err != nil {
		t.Fatalf("Failed to register gmacs API: %v", err)
	}
	
	// Bind key to minor mode
	luaCode := `gmacs.local_bind_key("auto-a-mode", "C-a", "version")`
	err = configLoader.GetVM().ExecuteString(luaCode)
	if err != nil {
		t.Fatalf("Failed to execute Lua code: %v", err)
	}
	
	// Get the auto-a-mode and check if the key binding was added
	modeManager := editor.ModeManager()
	autoAMode, exists := modeManager.GetMinorModeByName("auto-a-mode")
	if !exists {
		t.Fatal("auto-a-mode should exist")
	}
	
	keyBindings := autoAMode.KeyBindings()
	if keyBindings == nil {
		t.Fatal("auto-a-mode should have key bindings")
	}
	
	// Check if the key binding exists by simulating the key press
	// "C-a" means Ctrl+a, so call ProcessKeyPress with key="a", ctrl=true, meta=false
	_, matched, continuing := keyBindings.ProcessKeyPress("a", true, false)
	if !matched {
		t.Error("Key binding C-a should be registered in auto-a-mode")
	}
	
	if continuing {
		t.Error("C-a should be a complete binding, not a prefix")
	}
}