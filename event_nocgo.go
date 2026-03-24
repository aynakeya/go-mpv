//go:build !cgo
// +build !cgo

package mpv

import "unsafe"

type EventId int

const (
	EVENT_NONE               EventId = 0
	EVENT_SHUTDOWN           EventId = 1
	EVENT_LOG_MESSAGE        EventId = 2
	EVENT_GET_PROPERTY_REPLY EventId = 3
	EVENT_SET_PROPERTY_REPLY EventId = 4
	EVENT_COMMAND_REPLY      EventId = 5
	EVENT_START_FILE         EventId = 6
	EVENT_END_FILE           EventId = 7
	EVENT_FILE_LOADED        EventId = 8
	EVENT_CLIENT_MESSAGE     EventId = 16
	EVENT_VIDEO_RECONFIG     EventId = 17
	EVENT_AUDIO_RECONFIG     EventId = 18
	EVENT_SEEK               EventId = 20
	EVENT_PLAYBACK_RESTART   EventId = 21
	EVENT_PROPERTY_CHANGE    EventId = 22
	EVENT_QUEUE_OVERFLOW     EventId = 24
	EVENT_HOOK               EventId = 25
)

// EventName returns the symbolic event name.
func EventName(id EventId) string {
	if ensureBackend() != nil {
		return ""
	}
	return goString(backend.eventName(int32(id)))
}

type Event struct {
	EventId       EventId
	Error         error
	ReplyUserData uint64
	Data          unsafe.Pointer
	cEvent        *mpvEventRaw
}

type EventLogMessage struct {
	Prefix   string
	Level    string
	Text     string
	LogLevel LogLevel
}

// LogMessage converts event data to EventLogMessage.
func (e *Event) LogMessage() EventLogMessage {
	if e.EventId != EVENT_LOG_MESSAGE {
		panic("not a log message event")
	}
	s := (*mpvEventLogMessageRaw)(e.Data)
	return EventLogMessage{
		Prefix:   goString(s.Prefix),
		Level:    goString(s.Level),
		Text:     goString(s.Text),
		LogLevel: LogLevel(s.LogLevel),
	}
}

type EventProperty struct {
	Name   string
	Format Format
	Data   interface{}
}

// Property converts event data to EventProperty.
func (e *Event) Property() EventProperty {
	if e.EventId != EVENT_PROPERTY_CHANGE && e.EventId != EVENT_GET_PROPERTY_REPLY {
		panic("not a property event")
	}
	s := (*mpvEventPropertyRaw)(e.Data)
	p := EventProperty{
		Name:   goString(s.Name),
		Format: Format(s.Format),
	}
	switch p.Format {
	case FORMAT_NONE:
		p.Data = nil
	case FORMAT_STRING, FORMAT_OSD_STRING:
		p.Data = goString(*(**byte)(s.Data))
	case FORMAT_INT64:
		p.Data = *(*int64)(s.Data)
	case FORMAT_DOUBLE:
		p.Data = *(*float64)(s.Data)
	case FORMAT_FLAG:
		p.Data = *(*int32)(s.Data) == 1
	case FORMAT_NODE:
		p.Data = nodeFromRaw((*mpvNodeRaw)(s.Data))
	case FORMAT_BYTE_ARRAY:
		ba := (*mpvByteArrayRaw)(s.Data)
		if ba == nil {
			p.Data = ByteArray{}
		} else {
			p.Data = ByteArray(copyBytes(ba.Data, ba.Size))
		}
	default:
		panic("No such Format")
	}
	return p
}

type EventHook struct {
	Name string
	ID   uint64
}

// Hook converts event data to EventHook.
func (e *Event) Hook() EventHook {
	if e.EventId != EVENT_HOOK {
		panic("not a hook event")
	}
	s := (*mpvEventHookRaw)(e.Data)
	return EventHook{
		Name: goString(s.Name),
		ID:   s.ID,
	}
}

type EventCommand struct {
	Result Node
}

// Command converts event data to EventCommand.
func (e *Event) Command() EventCommand {
	if e.EventId != EVENT_COMMAND_REPLY {
		panic("not a command reply event")
	}
	s := (*mpvEventCommandRaw)(e.Data)
	return EventCommand{
		Result: nodeFromRaw(&s.Result),
	}
}

// ToNode converts event to node map.
func (e *Event) ToNode() (Node, error) {
	if e.cEvent == nil || ensureBackend() != nil {
		return Node{}, ERROR_INVALID_PARAMETER
	}
	var out mpvNodeRaw
	if err := newError(backend.eventToNode(&out, e.cEvent)); err != nil {
		return Node{}, err
	}
	defer backend.freeNodeContents(&out)
	return nodeFromRaw(&out), nil
}
