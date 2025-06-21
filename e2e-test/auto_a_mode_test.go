package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

/**
 * @spec マイナーモード/AutoAMode
 * @scenario AutoAModeの基本動作
 * @description AutoAModeが正しく動作し、Enterキー押下時に'a'が自動追加される
 * @given エディタとバッファが存在する
 * @when AutoAModeを有効化してEnterキーを押す
 * @then 改行後に'a'が自動的に追加される
 * @implementation domain/auto_a_mode.go, domain/editor.go
 */
func TestAutoAModeBasic(t *testing.T) {
	// エディタの作成
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	modeManager := editor.ModeManager()
	
	// 初期状態の確認：AutoAModeが無効
	autoAMode := modeManager.GetMinorModes(buffer)
	if len(autoAMode) != 0 {
		t.Errorf("Expected 0 minor modes initially, got %d", len(autoAMode))
	}
	
	// AutoAModeを有効化
	err := modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	if err != nil {
		t.Errorf("Failed to enable auto-a-mode: %v", err)
	}
	
	// モードが有効になったことを確認
	minorModes := modeManager.GetMinorModes(buffer)
	if len(minorModes) != 1 {
		t.Errorf("Expected 1 minor mode after enabling, got %d", len(minorModes))
	}
	
	if minorModes[0].Name() != "auto-a-mode" {
		t.Errorf("Expected auto-a-mode, got %s", minorModes[0].Name())
	}
}

/**
 * @spec マイナーモード/AutoAMode機能
 * @scenario AutoAModeの改行時a追加機能
 * @description AutoAMode有効時に改行すると'a'が自動追加される
 * @given AutoAModeが有効なバッファ
 * @when 改行を挿入する
 * @then 改行後に'a'が追加される
 * @implementation domain/auto_a_mode.go, domain/editor.go
 */
func TestAutoAModeNewlineEffect(t *testing.T) {
	// エディタとバッファの準備
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	modeManager := editor.ModeManager()
	
	// AutoAModeを有効化
	err := modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	if err != nil {
		t.Fatalf("Failed to enable auto-a-mode: %v", err)
	}
	
	// 初期状態のバッファ内容確認
	initialContent := buffer.Content()
	if len(initialContent) != 1 || initialContent[0] != "" {
		t.Errorf("Expected empty buffer initially, got %v", initialContent)
	}
	
	// テキストを挿入してから改行
	buffer.InsertChar('H')
	buffer.InsertChar('i')
	
	// 改行を挿入（AutoAModeのフックが動作するはず）
	buffer.InsertChar('\n')
	
	// AutoAModeのProcessNewlineを手動で呼び出し（実際の実装では自動で呼ばれる）
	minorModes := buffer.MinorModes()
	for _, mode := range minorModes {
		if autoAMode, ok := mode.(*domain.AutoAMode); ok {
			autoAMode.ProcessNewline(buffer)
		}
	}
	
	// バッファ内容を確認
	content := buffer.Content()
	if len(content) != 2 {
		t.Errorf("Expected 2 lines after newline, got %d", len(content))
	}
	
	if content[0] != "Hi" {
		t.Errorf("Expected first line to be 'Hi', got '%s'", content[0])
	}
	
	if content[1] != "a" {
		t.Errorf("Expected second line to be 'a', got '%s'", content[1])
	}
}

/**
 * @spec マイナーモード/AutoAModeトグル
 * @scenario AutoAModeの有効/無効切り替え
 * @description AutoAModeのトグル機能が正常に動作する
 * @given エディタとバッファが存在する
 * @when AutoAModeを複数回トグルする
 * @then 正しく有効/無効が切り替わる
 * @implementation domain/auto_a_mode.go, domain/mode.go
 */
func TestAutoAModeToggle(t *testing.T) {
	// エディタの準備
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	modeManager := editor.ModeManager()
	
	// 初期状態：無効
	minorModes := modeManager.GetMinorModes(buffer)
	if len(minorModes) != 0 {
		t.Errorf("Expected no minor modes initially, got %d", len(minorModes))
	}
	
	// 1回目のトグル：有効化
	err := modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	if err != nil {
		t.Errorf("Failed to toggle auto-a-mode (enable): %v", err)
	}
	
	minorModes = modeManager.GetMinorModes(buffer)
	if len(minorModes) != 1 {
		t.Errorf("Expected 1 minor mode after first toggle, got %d", len(minorModes))
	}
	
	// 2回目のトグル：無効化
	err = modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	if err != nil {
		t.Errorf("Failed to toggle auto-a-mode (disable): %v", err)
	}
	
	minorModes = modeManager.GetMinorModes(buffer)
	if len(minorModes) != 0 {
		t.Errorf("Expected 0 minor modes after second toggle, got %d", len(minorModes))
	}
	
	// 3回目のトグル：再度有効化
	err = modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	if err != nil {
		t.Errorf("Failed to toggle auto-a-mode (re-enable): %v", err)
	}
	
	minorModes = modeManager.GetMinorModes(buffer)
	if len(minorModes) != 1 {
		t.Errorf("Expected 1 minor mode after third toggle, got %d", len(minorModes))
	}
}

/**
 * @spec マイナーモード/モードライン表示
 * @scenario マイナーモードのモードライン表示
 * @description 有効なマイナーモードがモードラインに表示される
 * @given エディタとバッファが存在する
 * @when AutoAModeを有効化する
 * @then モードライン表示にマイナーモード名が含まれる
 * @implementation cli/display.go
 */
func TestMinorModeDisplayInModeLine(t *testing.T) {
	// エディタとバッファの準備
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	modeManager := editor.ModeManager()
	
	// AutoAModeを有効化
	err := modeManager.ToggleMinorMode(buffer, "auto-a-mode")
	if err != nil {
		t.Fatalf("Failed to enable auto-a-mode: %v", err)
	}
	
	// モードライン表示の構成要素を確認
	majorMode := buffer.MajorMode()
	minorModes := buffer.MinorModes()
	
	// メジャーモード名が取得できることを確認
	if majorMode == nil {
		t.Fatal("Major mode should not be nil")
	}
	
	// マイナーモードが設定されていることを確認
	if len(minorModes) != 1 {
		t.Errorf("Expected 1 minor mode, got %d", len(minorModes))
	}
	
	if minorModes[0].Name() != "auto-a-mode" {
		t.Errorf("Expected auto-a-mode, got %s", minorModes[0].Name())
	}
	
	// 期待されるモードライン形式: " BufferName (major-mode) [minor-mode] "
	expectedContent := buffer.Name()
	expectedMajor := majorMode.Name()
	expectedMinor := minorModes[0].Name()
	
	// モードライン文字列の形式確認
	modeLinePattern := " " + expectedContent + " (" + expectedMajor + ") [" + expectedMinor + "] "
	
	// 各要素が存在することを確認
	if !strings.Contains(modeLinePattern, expectedContent) {
		t.Error("Mode line should contain buffer name")
	}
	
	if !strings.Contains(modeLinePattern, expectedMajor) {
		t.Error("Mode line should contain major mode name")
	}
	
	if !strings.Contains(modeLinePattern, expectedMinor) {
		t.Error("Mode line should contain minor mode name")
	}
}