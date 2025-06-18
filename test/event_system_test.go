package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

func TestEventQueue(t *testing.T) {
	editor := domain.NewEditor()
	queue := editor.EventQueue()
	
	testEvent := events.KeyEventData{Rune: 'A', Key: "A"}
	queue.Push(testEvent)
	
	event, hasEvent := queue.Pop()
	if !hasEvent {
		t.Fatal("Expected event in queue")
	}
	
	keyEvent, ok := event.(events.KeyEventData)
	if !ok {
		t.Fatal("Expected KeyEventData")
	}
	
	if keyEvent.Rune != 'A' {
		t.Errorf("Expected rune 'A', got '%c'", keyEvent.Rune)
	}
}

func TestResizeEvent(t *testing.T) {
	editor := domain.NewEditor()
	
	resizeEvent := events.ResizeEventData{
		Width:  100,
		Height: 30,
	}
	
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	width, height := window.Size()
	
	if width != 100 || height != 30 {
		t.Errorf("Expected size 100x30, got %dx%d", width, height)
	}
}

func TestQuitEvent(t *testing.T) {
	editor := domain.NewEditor()
	
	if !editor.IsRunning() {
		t.Fatal("Editor should be running initially")
	}
	
	quitEvent := events.QuitEventData{}
	editor.HandleEvent(quitEvent)
	
	if editor.IsRunning() {
		t.Error("Editor should have quit after QuitEvent")
	}
}

func TestEventQueueCapacity(t *testing.T) {
	queue := events.NewEventQueue(2)
	
	// Fill the queue
	queue.Push(events.KeyEventData{Rune: 'A', Key: "A"})
	queue.Push(events.KeyEventData{Rune: 'B', Key: "B"})
	queue.Push(events.KeyEventData{Rune: 'C', Key: "C"}) // This should be dropped
	
	// Pop first event
	event, hasEvent := queue.Pop()
	if !hasEvent {
		t.Fatal("Expected first event")
	}
	if keyEvent := event.(events.KeyEventData); keyEvent.Rune != 'A' {
		t.Errorf("Expected 'A', got '%c'", keyEvent.Rune)
	}
	
	// Pop second event
	event, hasEvent = queue.Pop()
	if !hasEvent {
		t.Fatal("Expected second event")
	}
	if keyEvent := event.(events.KeyEventData); keyEvent.Rune != 'B' {
		t.Errorf("Expected 'B', got '%c'", keyEvent.Rune)
	}
	
	// Queue should be empty now
	_, hasEvent = queue.Pop()
	if hasEvent {
		t.Error("Queue should be empty")
	}
}

func BenchmarkEventProcessing(b *testing.B) {
	editor := domain.NewEditor()
	event := events.KeyEventData{Rune: 'a', Key: "a"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		editor.HandleEvent(event)
	}
}