package cli

import (
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/term"
	
	"github.com/TakahashiShuuhei/gmacs/core/events"
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
	var err error
	t.oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	
	signal.Notify(t.sigChan, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
	
	go t.handleSignals()
	go t.readInput()
	
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
	for sig := range t.sigChan {
		switch sig {
		case syscall.SIGWINCH:
			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				t.eventChan <- events.ResizeEventData{
					Width:  width,
					Height: height,
				}
			}
		case syscall.SIGINT, syscall.SIGTERM:
			t.eventChan <- events.QuitEventData{}
			return
		}
	}
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