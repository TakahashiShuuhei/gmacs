package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestCtrlCQuit(t *testing.T) {
	editor := domain.NewEditor()
	
	if !editor.IsRunning() {
		t.Fatal("Editor should be running initially")
	}
	
	event := events.KeyEventData{
		Key:  "c",
		Ctrl: true,
	}
	editor.HandleEvent(event)
	
	if editor.IsRunning() {
		t.Error("Editor should have quit after Ctrl+C")
	}
}


func TestCtrlModifierDoesNotInsertText(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	event := events.KeyEventData{
		Key:  "a",
		Rune: 'a',
		Ctrl: true,
	}
	editor.HandleEvent(event)
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}
	
	if lines[0] != "" {
		t.Errorf("Expected empty line after Ctrl+a, got '%s'", lines[0])
	}
}

func TestMetaModifierDoesNotInsertText(t *testing.T) {
	editor := domain.NewEditor()
	renderer := &TestRenderer{}
	
	event := events.KeyEventData{
		Key:  "x",
		Rune: 'x',
		Meta: true,
	}
	editor.HandleEvent(event)
	
	renderer.Render(editor)
	lines := renderer.GetLastRender()
	
	if len(lines) == 0 {
		t.Fatal("Expected at least one line")
	}
	
	if lines[0] != "" {
		t.Errorf("Expected empty line after Meta+x, got '%s'", lines[0])
	}
}