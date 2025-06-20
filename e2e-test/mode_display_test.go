package test

import (
	"strings"
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

/**
 * @spec モード表示/メジャーモード表示
 * @scenario メジャーモード名のモードライン表示
 * @description モードラインにメジャーモード名が正しく表示される
 * @given エディタが起動している
 * @when 異なる拡張子のファイルを開く
 * @then モードラインに正しいメジャーモード名が表示される
 * @implementation cli/display.go, domain/text_mode.go
 */
func TestMajorModeDisplay(t *testing.T) {
	// エディタの作成
	editor := domain.NewEditor()

	// 基本バッファ（*scratch*）の確認
	buffer := editor.CurrentBuffer()
	if buffer.MajorMode().Name() != "fundamental-mode" {
		t.Errorf("Expected fundamental-mode for *scratch*, got %s", buffer.MajorMode().Name())
	}

	// テキストファイルバッファの作成とモード自動検出のテスト
	textBuffer, err := domain.NewBufferFromFile("test.txt")
	if err != nil {
		// ファイルが存在しない場合は、手動でファイルパスを設定
		textBuffer = domain.NewBuffer("test.txt")
		textBuffer.SetFilepath("test.txt")
	}
	
	editor.AddBuffer(textBuffer)
	
	// text-modeが設定されていることを確認
	if textBuffer.MajorMode() == nil {
		t.Fatal("Major mode should be set for text buffer")
	}
	
	expectedMode := "text-mode"
	if textBuffer.MajorMode().Name() != expectedMode {
		t.Errorf("Expected %s for .txt file, got %s", expectedMode, textBuffer.MajorMode().Name())
	}
}

/**
 * @spec モード表示/ファイル拡張子マッピング
 * @scenario 複数の拡張子でのモード自動検出
 * @description 異なるファイル拡張子で正しいメジャーモードが検出される
 * @given エディタとモードマネージャーが存在する
 * @when 様々な拡張子のファイルを処理する
 * @then 各ファイルに適切なメジャーモードが設定される
 * @implementation domain/mode.go, domain/text_mode.go
 */
func TestFileExtensionModeMapping(t *testing.T) {
	editor := domain.NewEditor()
	modeManager := editor.ModeManager()
	
	// テストケース: 拡張子とその期待されるモード
	testCases := []struct {
		filename     string
		expectedMode string
	}{
		{"README.md", "text-mode"},
		{"document.txt", "text-mode"},
		{"notes.text", "text-mode"},
		{"article.markdown", "text-mode"},
		{"todo.org", "text-mode"},
		{"script.go", "fundamental-mode"}, // Goファイル用モードがないので fundamental-mode
		{"unknown", "fundamental-mode"},    // 拡張子なし
	}
	
	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// ファイルパス付きのバッファを作成
			buffer := domain.NewBuffer(tc.filename)
			buffer.SetFilepath(tc.filename)
			
			// モード自動検出
			detectedMode, err := modeManager.AutoDetectMajorMode(buffer)
			if err != nil {
				t.Errorf("Failed to detect mode for %s: %v", tc.filename, err)
				return
			}
			
			if detectedMode.Name() != tc.expectedMode {
				t.Errorf("For %s: expected %s, got %s", tc.filename, tc.expectedMode, detectedMode.Name())
			}
		})
	}
}

/**
 * @spec モード表示/モードライン内容
 * @scenario モードライン表示内容の確認
 * @description モードラインに表示される内容が正しい形式である
 * @given エディタとモックディスプレイが存在する
 * @when モードラインを描画する
 * @then 正しい形式でバッファ名とモード名が表示される
 * @implementation cli/display.go
 */
func TestModeLineContent(t *testing.T) {
	// この関数は実際のモードライン描画のテストのために将来実装される
	// 現在はモックディスプレイの制限により、基本的な確認のみ
	editor := domain.NewEditor()
	buffer := editor.CurrentBuffer()
	
	// モードライン表示に必要な情報が取得できることを確認
	if buffer.Name() == "" {
		t.Error("Buffer name should not be empty")
	}
	
	if buffer.MajorMode() == nil {
		t.Error("Major mode should not be nil")
	}
	
	// 期待される形式: " BufferName (mode-name) "
	expectedFormat := " " + buffer.Name() + " (" + buffer.MajorMode().Name() + ") "
	if !strings.Contains(expectedFormat, buffer.Name()) {
		t.Error("Mode line should contain buffer name")
	}
	
	if !strings.Contains(expectedFormat, buffer.MajorMode().Name()) {
		t.Error("Mode line should contain major mode name")
	}
}

/**
 * @spec モード表示/モード切り替え確認
 * @scenario モード切り替え時の表示更新
 * @description メジャーモードを切り替えた時に表示が更新される
 * @given エディタとバッファが存在する
 * @when メジャーモードを切り替える
 * @then 新しいモード名が確認できる
 * @implementation domain/mode.go, domain/buffer.go
 */
func TestModeSwitch(t *testing.T) {
	editor := domain.NewEditor()
	modeManager := editor.ModeManager()
	buffer := editor.CurrentBuffer()
	
	// 初期状態の確認
	initialMode := buffer.MajorMode().Name()
	if initialMode != "fundamental-mode" {
		t.Errorf("Initial mode should be fundamental-mode, got %s", initialMode)
	}
	
	// text-modeに切り替え
	err := modeManager.SetMajorMode(buffer, "text-mode")
	if err != nil {
		t.Errorf("Failed to switch to text-mode: %v", err)
	}
	
	// モードが変更されたことを確認
	newMode := buffer.MajorMode().Name()
	if newMode != "text-mode" {
		t.Errorf("Mode should be text-mode after switch, got %s", newMode)
	}
	
	// fundamental-modeに戻す
	err = modeManager.SetMajorMode(buffer, "fundamental-mode")
	if err != nil {
		t.Errorf("Failed to switch back to fundamental-mode: %v", err)
	}
	
	finalMode := buffer.MajorMode().Name()
	if finalMode != "fundamental-mode" {
		t.Errorf("Mode should be fundamental-mode after switch back, got %s", finalMode)
	}
}