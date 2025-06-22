package cli

import (
	"os"
	"os/signal"
	"syscall"
	"golang.org/x/term"
	
	"github.com/TakahashiShuuhei/gmacs/events"
	"github.com/TakahashiShuuhei/gmacs/log"
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
	
	// Check for ANSI escape sequences first
	if len(data) >= 3 && data[0] == 27 && data[1] == '[' {
		// Arrow keys and other ANSI sequences
		sequence := string(data)
		log.Debug("Detected ANSI escape sequence: %q", sequence)
		
		event := events.KeyEventData{
			Raw: data,
			Key: sequence,
		}
		
		switch sequence {
		case "\x1b[A":
			log.Debug("Recognized Up arrow")
		case "\x1b[B":
			log.Debug("Recognized Down arrow")
		case "\x1b[C":
			log.Debug("Recognized Right arrow")
		case "\x1b[D":
			log.Debug("Recognized Left arrow")
		default:
			log.Debug("Unknown ANSI sequence: %q", sequence)
		}
		
		log.Debug("Sending ANSI event: key=%s", event.Key)
		t.eventChan <- event
		return
	}
	
	// Try to parse as UTF-8 for multi-byte characters
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
		case 1: // Ctrl+A
			event.Key = "a"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+A")
		case 2: // Ctrl+B
			event.Key = "b"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+B")
		case 3: // Ctrl+C
			event.Key = "c"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+C")
		case 5: // Ctrl+E
			event.Key = "e"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+E")
		case 6: // Ctrl+F
			event.Key = "f"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+F")
		case 14: // Ctrl+N
			event.Key = "n"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+N")
		case 16: // Ctrl+P
			event.Key = "p"
			event.Ctrl = true
			log.Debug("Recognized Ctrl+P")
		case 13: // Enter
			event.Key = "Enter"
			event.Rune = '\n'
			log.Debug("Recognized Enter")
		case 27: // ESC
			event.Key = "\x1b"
			log.Debug("Recognized Escape")
		case 127: // Backspace
			event.Key = "Backspace"
			log.Debug("Recognized Backspace")
		default:
			if b >= 32 && b <= 126 {
				event.Rune = rune(b)
				event.Key = string(rune(b))
				log.Debug("ASCII character: %c", b)
			} else if b >= 1 && b <= 26 {
				// Other Ctrl combinations
				event.Key = string(rune('a' + b - 1))
				event.Ctrl = true
				log.Debug("Recognized Ctrl+%c", 'A'+b-1)
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