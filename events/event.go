package events

type EventType int

const (
	KeyEvent EventType = iota
	ResizeEvent
	QuitEvent
)

type Event interface {
	Type() EventType
}

type KeyEventData struct {
	Key      string
	Rune     rune
	Ctrl     bool
	Meta     bool
	Alt      bool
	Shift    bool
	Raw      []byte
}

func (k KeyEventData) Type() EventType {
	return KeyEvent
}

type ResizeEventData struct {
	Width  int
	Height int
}

func (r ResizeEventData) Type() EventType {
	return ResizeEvent
}

type QuitEventData struct{}

func (q QuitEventData) Type() EventType {
	return QuitEvent
}

type EventQueue struct {
	events chan Event
}

func NewEventQueue(bufferSize int) *EventQueue {
	return &EventQueue{
		events: make(chan Event, bufferSize),
	}
}

func (eq *EventQueue) Push(event Event) {
	select {
	case eq.events <- event:
	default:
	}
}

func (eq *EventQueue) Pop() (Event, bool) {
	select {
	case event := <-eq.events:
		return event, true
	default:
		return nil, false
	}
}

func (eq *EventQueue) PopBlocking() Event {
	return <-eq.events
}

func (eq *EventQueue) Close() {
	close(eq.events)
}