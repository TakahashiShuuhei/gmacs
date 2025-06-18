package cli

import (
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/term"
	
	"github.com/TakahashiShuuhei/gmacs/core/events"
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

type Terminal struct {
	oldState  *term.State
	eventChan chan events.Event
	sigChan   chan os.Signal
}

func NewTerminal() *Terminal {
	return &Terminal{
		eventChan: make(chan events.Event, 100),
		sigChan:   make(chan os.Signal, 1),
	}
}

func (t *Terminal) Init() error {
	log.Debug("Initializing terminal")
	var err error
	t.oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Error("Failed to make terminal raw: %v", err)
		return err
	}
	
	signal.Notify(t.sigChan, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
	
	log.Debug("Starting signal and input handlers")
	go t.handleSignals()
	go t.readInput()
	
	log.Info("Terminal initialized successfully")
	return nil
}

func (t *Terminal) Restore() error {
	if t.oldState != nil {
		return term.Restore(int(os.Stdin.Fd()), t.oldState)
	}
	return nil
}

func (t *Terminal) EventChan() <-chan events.Event {
	return t.eventChan
}

func (t *Terminal) handleSignals() {
	log.Debug("Signal handler started")
	for sig := range t.sigChan {
		log.Debug("Received signal: %v", sig)
		switch sig {
		case syscall.SIGWINCH:
			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				log.Info("Terminal resized to %dx%d", width, height)
				t.eventChan <- events.ResizeEventData{
					Width:  width,
					Height: height,
				}
			} else {
				log.Error("Failed to get terminal size: %v", err)
			}
		case syscall.SIGINT, syscall.SIGTERM:
			log.Info("Received termination signal: %v", sig)
			t.eventChan <- events.QuitEventData{}
			return
		}
	}
	log.Debug("Signal handler stopped")
}

func (t *Terminal) readInput() {
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		
		if n > 0 {
			t.parseInput(buf[:n])
		}
	}
}

func (t *Terminal) parseInput(data []byte) {
	for _, b := range data {
		event := events.KeyEventData{
			Raw: []byte{b},
		}
		
		switch b {
		case 3: // Ctrl+C
			event.Key = "c"
			event.Ctrl = true
		case 13: // Enter
			event.Key = "Enter"
			event.Rune = '\n'
		case 27: // ESC
			event.Key = "Escape"
		case 127: // Backspace
			event.Key = "Backspace"
		default:
			if b >= 32 && b <= 126 {
				event.Rune = rune(b)
				event.Key = string(rune(b))
			}
		}
		
		t.eventChan <- event
	}
}

func (t *Terminal) Close() {
	close(t.eventChan)
	signal.Stop(t.sigChan)
	close(t.sigChan)
}