package main

import (
	"log"
	"time"

	"github.com/TakahashiShuuhei/gmacs/core/cli"
	"github.com/TakahashiShuuhei/gmacs/core/domain"
)

func main() {
	editor := domain.NewEditor()
	display := cli.NewDisplay()
	terminal := cli.NewTerminal()

	if err := terminal.Init(); err != nil {
		log.Fatal("Failed to initialize terminal:", err)
	}
	defer terminal.Restore()

	display.Render(editor)

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	for editor.IsRunning() {
		select {
		case event := <-terminal.EventChan():
			editor.EventQueue().Push(event)
		case <-ticker.C:
			for {
				event, hasEvent := editor.EventQueue().Pop()
				if !hasEvent {
					break
				}
				editor.HandleEvent(event)
			}
			display.Render(editor)
		}
	}
}