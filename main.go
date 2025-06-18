package main

import (
	"log"
	"time"

	"github.com/TakahashiShuuhei/gmacs/core/cli"
	"github.com/TakahashiShuuhei/gmacs/core/domain"
	gmacslog "github.com/TakahashiShuuhei/gmacs/core/log"
)

func main() {
	if err := gmacslog.Init(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer gmacslog.Close()

	gmacslog.Info("gmacs starting up")

	editor := domain.NewEditor()
	display := cli.NewDisplay()
	terminal := cli.NewTerminal()

	gmacslog.Debug("Initializing terminal")
	if err := terminal.Init(); err != nil {
		gmacslog.Error("Failed to initialize terminal: %v", err)
		log.Fatal("Failed to initialize terminal:", err)
	}
	defer terminal.Restore()

	gmacslog.Debug("Initial render")
	display.Render(editor)

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	gmacslog.Info("Entering main loop")
	for editor.IsRunning() {
		select {
		case event := <-terminal.EventChan():
			gmacslog.Debug("Received event: %T", event)
			editor.EventQueue().Push(event)
		case <-ticker.C:
			for {
				event, hasEvent := editor.EventQueue().Pop()
				if !hasEvent {
					break
				}
				gmacslog.Debug("Processing event: %T", event)
				editor.HandleEvent(event)
			}
			display.Render(editor)
		}
	}

	gmacslog.Info("gmacs shutting down")
}