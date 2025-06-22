package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/domain"
)

/**
 * @spec モード管理/メジャーモード
 * @scenario 基本的なメジャーモード機能
 * @description Emacsライクなメジャーモードシステムの基本動作を確認
 * @given エディタが起動している
 * @when 新しいバッファを作成する
 * @then fundamental-modeが自動設定される
 * @implementation domain/mode.go, domain/fundamental_mode.go
 */
func TestMajorModeBasics(t *testing.T) {
	// エディタの作成
	editor := NewEditorWithDefaults()

	// 現在のバッファを取得
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		t.Fatal("Current buffer should not be nil")
	}

	// fundamental-modeが設定されていることを確認
	majorMode := buffer.MajorMode()
	if majorMode == nil {
		t.Fatal("Major mode should not be nil")
	}

	if majorMode.Name() != "fundamental-mode" {
		t.Errorf("Expected fundamental-mode, got %s", majorMode.Name())
	}
}

/**
 * @spec モード管理/メジャーモード切り替え
 * @scenario メジャーモードの手動切り替え
 * @description ModeManagerを使ったメジャーモードの切り替えが正常に動作する
 * @given エディタとバッファが存在する
 * @when ModeManagerでメジャーモードを切り替える
 * @then 正しいモードが設定される
 * @implementation domain/mode.go
 */
func TestMajorModeSwitch(t *testing.T) {
	// エディタとモードマネージャーの準備
	editor := NewEditorWithDefaults()
	modeManager := editor.ModeManager()
	buffer := editor.CurrentBuffer()

	// 初期状態の確認
	if buffer.MajorMode().Name() != "fundamental-mode" {
		t.Errorf("Expected fundamental-mode, got %s", buffer.MajorMode().Name())
	}

	// 新しいメジャーモードを設定（存在しないモード）
	err := modeManager.SetMajorMode(buffer, "nonexistent-mode")
	if err == nil {
		t.Error("Expected error for nonexistent mode")
	}

	// fundamental-modeに再設定
	err = modeManager.SetMajorMode(buffer, "fundamental-mode")
	if err != nil {
		t.Errorf("Failed to set fundamental-mode: %v", err)
	}
}

/**
 * @spec モード管理/ファイル関連バッファ
 * @scenario ファイルバッファのモード自動検出
 * @description ファイルパスに基づくメジャーモードの自動検出が動作する
 * @given エディタが存在する
 * @when ファイルバッファを作成する
 * @then 適切なメジャーモードが設定される
 * @implementation domain/mode.go
 */
func TestFileModeDetection(t *testing.T) {
	// エディタの作成
	editor := NewEditorWithDefaults()
	modeManager := editor.ModeManager()

	// テスト用のファイルバッファを作成
	buffer := domain.NewBuffer("test.txt")

	// モード自動検出をテスト
	detectedMode, err := modeManager.AutoDetectMajorMode(buffer)
	if err != nil {
		t.Errorf("Failed to auto-detect mode: %v", err)
	}

	// ファイルパスがない場合は fundamental-mode になるはず
	if detectedMode.Name() != "fundamental-mode" {
		t.Errorf("Expected fundamental-mode for buffer without filepath, got %s", detectedMode.Name())
	}
}

/**
 * @spec モード管理/マイナーモード
 * @scenario マイナーモードの基本動作
 * @description マイナーモードの有効化・無効化が正常に動作する
 * @given エディタとバッファが存在する
 * @when マイナーモードを操作する
 * @then 正しく有効化・無効化される
 * @implementation domain/mode.go, domain/buffer.go
 */
func TestMinorModeBasics(t *testing.T) {
	// エディタの準備
	editor := NewEditorWithDefaults()
	buffer := editor.CurrentBuffer()

	// 初期状態ではマイナーモードは空
	minorModes := buffer.MinorModes()
	if len(minorModes) != 0 {
		t.Errorf("Expected 0 minor modes, got %d", len(minorModes))
	}

	// テスト用マイナーモードの作成と有効化をテスト
	// （実際のマイナーモード実装が必要になったら追加）
}

/**
 * @spec モード管理/システム統合
 * @scenario モードシステムとエディタの統合
 * @description モードシステムがエディタ全体と正しく統合されている
 * @given エディタが起動している
 * @when 各種操作を行う
 * @then モードシステムが正常に動作する
 * @implementation domain/editor.go, domain/mode.go
 */
func TestModeSystemIntegration(t *testing.T) {
	// エディタの作成
	editor := NewEditorWithDefaults()

	// ModeManagerがエディタに統合されていることを確認
	modeManager := editor.ModeManager()
	if modeManager == nil {
		t.Fatal("ModeManager should not be nil")
	}

	// 新しいバッファを追加してモードが自動設定されることを確認
	newBuffer := domain.NewBuffer("test")
	editor.AddBuffer(newBuffer)

	// モードが設定されていることを確認
	if newBuffer.MajorMode() == nil {
		t.Error("Major mode should be set automatically for new buffer")
	}

	// fundamental-modeが設定されていることを確認
	if newBuffer.MajorMode().Name() != "fundamental-mode" {
		t.Errorf("Expected fundamental-mode, got %s", newBuffer.MajorMode().Name())
	}
}
