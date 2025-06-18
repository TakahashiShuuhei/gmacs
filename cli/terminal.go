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
	log.Debug("Input reader started")
	buf := make([]byte, 256)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			log.Error("Failed to read input: %v", err)
			break
		}
		
		if n > 0 {
			log.Debug("Raw input received: %d bytes: %v (hex: %x)", n, buf[:n], buf[:n])
			t.parseInput(buf[:n])
		}
	}
	log.Debug("Input reader stopped")
}

func (t *Terminal) parseInput(data []byte) {
	log.Debug("Parsing input data: %d bytes", len(data))
	
	// Try to parse as UTF-8 first
	if len(data) > 1 {
		// Multi-byte sequence - could be UTF-8
		runes := []rune(string(data))
		log.Debug("UTF-8 parsing: %d runes from %d bytes: %+q", len(runes), len(data), runes)
		
		for _, r := range runes {
			if r != '\ufffd' { // Valid UTF-8 character
				event := events.KeyEventData{
					Raw:  data,
					Rune: r,
					Key:  string(r),
				}
				log.Debug("Created UTF-8 event: rune=%c (U+%04X), key=%s", r, r, event.Key)
				t.eventChan <- event
			}
		}
		return
	}
	
	// Single byte processing
	for i, b := range data {
		event := events.KeyEventData{
			Raw: []byte{b},
		}
		
		log.Debug("Processing byte %d: 0x%02x (%d)", i, b, b)
		
		switch b {
		case 3: // Ctrl+C
			event.Key = "c"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+C")
		case 13: // Enter
			event.Key = "Enter"
			event.Rune = '\n'
			log.Debug("Recognized Enter")
		case 27: // ESC
			event.Key = "Escape"
			log.Debug("Recognized Escape")
		case 127: // Backspace
			event.Key = "Backspace"
			log.Debug("Recognized Backspace")
		default:
			if b >= 32 && b <= 126 {
				event.Rune = rune(b)
				event.Key = string(rune(b))
				log.Debug("ASCII character: %c", b)
			} else {
				log.Debug("Non-printable byte: 0x%02x", b)
			}
		}
		
		log.Debug("Sending event: key=%s, rune=%c, ctrl=%t, raw=%v", event.Key, event.Rune, event.Ctrl, event.Raw)
		t.eventChan <- event
	}
}

func (t *Terminal) Close() {
	close(t.eventChan)
	signal.Stop(t.sigChan)
	close(t.sigChan)
}