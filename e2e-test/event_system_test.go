package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
)

/**
 * @spec events/queue_operations
 * @scenario イベントキューの基本操作
 * @description イベントキューのPush/Pop操作の基本動作検証
 * @given エディタを新規作成してイベントキューを取得
 * @when KeyEventData('A')をキューにプッシュし、ポップする
 * @then イベントが正しく取り出され、データが保持される
 * @implementation events/event_queue.go
 */
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

/**
 * @spec events/resize_handling
 * @scenario リサイズイベントの処理
 * @description ターミナルリサイズイベントの処理とウィンドウサイズ更新
 * @given エディタを新規作成する
 * @when 100x30サイズのリサイズイベントを送信
 * @then ウィンドウサイズが100x28（モードラインとミニバッファを除いたサイズ）に更新される
 * @implementation events/resize_event.go, domain/window.go
 */
func TestResizeEvent(t *testing.T) {
	editor := domain.NewEditor()
	
	resizeEvent := events.ResizeEventData{
		Width:  100,
		Height: 30,
	}
	
	editor.HandleEvent(resizeEvent)
	
	window := editor.CurrentWindow()
	width, height := window.Size()
	
	if width != 100 || height != 28 { // 30-2 for mode line and minibuffer
		t.Errorf("Expected size 100x28, got %dx%d", width, height)
	}
}

/**
 * @spec events/quit_handling
 * @scenario 終了イベントの処理
 * @description 終了イベントの処理とエディタ状態の変更
 * @given エディタが実行中の状態
 * @when QuitEventDataを送信する
 * @then エディタが終了状態に変更される
 * @implementation events/quit_event.go, domain/editor.go
 */
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

/**
 * @spec events/queue_capacity
 * @scenario イベントキューの容量制限
 * @description イベントキューの容量制限とオーバーフロー処理
 * @given 容量2のイベントキューを作成
 * @when 3つのイベント（A、B、C）を順次プッシュ
 * @then 最初の2つのイベント（A、B）のみが保持され、3番目（C）は破棄される
 * @implementation events/event_queue.go, 容量制限処理
 */
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

/**
 * @spec events/performance_benchmark
 * @scenario イベント処理のパフォーマンスベンチマーク
 * @description イベント処理のパフォーマンス測定
 * @given エディタとKeyEventData('a')を準備
 * @when N回繰り返してイベントを処理
 * @then イベント処理の平均時間を測定する
 * @implementation domain/editor.go, パフォーマンス測定
 */
func BenchmarkEventProcessing(b *testing.B) {
	editor := domain.NewEditor()
	event := events.KeyEventData{Rune: 'a', Key: "a"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		editor.HandleEvent(event)
	}
}