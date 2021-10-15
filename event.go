package gopool

// EventLevel event level
type EventLevel int

const (
	// EventLevelDebug the info event
	EventLevelDebug = iota
	// EventLevelInfo the debug event
	EventLevelInfo
	// EventLevelWarring the warring event
	EventLevelWarring
	// EventLevelError the error Event
	EventLevelError
	// EventJobName event job name
	EventJobName = "gopool_event"
)

// Event pool event
type Event struct {
	level    EventLevel
	msg      string
	callback func(event *Event)
}

var _ JobHandler = &Event{}

func newEvent(level EventLevel, msg string, callback func(event *Event)) *Event {
	return &Event{
		level:    level,
		msg:      msg,
		callback: callback,
	}
}

// Handle event job handle
func (e *Event) Handle() (interface{}, error) {
	e.callback(e)
	return nil, nil
}

func (e *Event) String() string {
	level := "[INFO] "

	switch e.level {
	case EventLevelDebug:
		level = "[DEBUG] "
	case EventLevelError:
		level = "[ERROR] "
	case EventLevelWarring:
		level = "[Warring] "
	}

	return level + e.msg
}
