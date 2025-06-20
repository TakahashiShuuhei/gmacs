package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

type TestRenderer struct {
	lastRender []string
	renderCount int
}

func (tr *TestRenderer) Render(editor *domain.Editor) {
	window := editor.CurrentWindow()
	if window != nil {
		tr.lastRender = window.VisibleLines()
		tr.renderCount++
	}
}

func (tr *TestRenderer) GetLastRender() []string {
	return tr.lastRender
}

func (tr *TestRenderer) GetRenderCount() int {
	return tr.renderCount
}

/**
 * @spec editor/startup
 * @scenario エディタ初期化と基本状態の確認
 * @description エディタ起動時の初期状態（バッファ、ウィンドウ、レンダリング）を検証
 * @given エディタを新規作成する
 * @when エディタの初期状態を確認する
 * @then 実行中状態、*scratch*バッファ、ウィンドウが正しく設定される
 * @implementation domain/editor.go, domain/buffer.go, domain/window.go
 */
func TestEditorStartup(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	if !editor.IsRunning() {
		t.Fatal("Editor should be running after creation")
	}
	
	buffer := editor.CurrentBuffer()
	if buffer == nil {
		t.Fatal("Current buffer should not be nil")
	}
	
	if buffer.Name() != "*scratch*" {
		t.Errorf("Expected buffer name '*scratch*', got '%s'", buffer.Name())
	}
	
	window := editor.CurrentWindow()
	if window == nil {
		t.Fatal("Current window should not be nil")
	}
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Error("Expected at least one line in the buffer")
	}
	
	if lines[0] != "" {
		t.Errorf("Expected empty first line, got '%s'", lines[0])
	}
}


