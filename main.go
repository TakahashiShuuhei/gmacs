package main

import (
	"log"
	"time"

	"github.com/TakahashiShuuhei/gmacs/core/cli"
	"github.com/TakahashiShuuhei/gmacs/core/domain"
	"github.com/TakahashiShuuhei/gmacs/core/events"
	gmacslog "github.com/TakahashiShuuhei/gmacs/core/log"
)

func main() {
	if err := gmacslog.Init(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer gmacslog.Close()

	gmacslog.Info("gmacs starting up")

	display := cli.NewDisplay()
	terminal := cli.NewTerminal()

	gmacslog.Debug("Initializing terminal")
	if err := terminal.Init(); err != nil {
		gmacslog.Error("Failed to initialize terminal: %v", err)
		log.Fatal("Failed to initialize terminal:", err)
	}
	defer terminal.Restore()

	// Create editor and set initial size from display
	editor := domain.NewEditor()
	width, height := display.Size()
	resizeEvent := events.ResizeEventData{
		Width:  width,
		Height: height,
	}
	editor.HandleEvent(resizeEvent)

	gmacslog.Debug("Initial render")
	display.Render(editor)

	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	gmacslog.Info("Entering main loop")
	needsRender := false
	
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
				
				// Handle resize events for display as well
				if resizeEvent, ok := event.(events.ResizeEventData); ok {
					display.Resize(resizeEvent.Width, resizeEvent.Height)
				}
				
				editor.HandleEvent(event)
				needsRender = true
			}
			
			// Only render if there were events or if we need to render
			if needsRender {
				display.Render(editor)
				needsRender = false
			}
		}
	}

	gmacslog.Info("gmacs shutting down")
}