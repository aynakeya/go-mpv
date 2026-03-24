//go:build cgo
// +build cgo

package mpv

/*
#include <mpv/client.h>
*/
import "C"
import (
	"unsafe"
)

type EventId int

const (
	EVENT_NONE               EventId = C.MPV_EVENT_NONE
	EVENT_SHUTDOWN           EventId = C.MPV_EVENT_SHUTDOWN
	EVENT_LOG_MESSAGE        EventId = C.MPV_EVENT_LOG_MESSAGE
	EVENT_GET_PROPERTY_REPLY EventId = C.MPV_EVENT_GET_PROPERTY_REPLY
	EVENT_SET_PROPERTY_REPLY EventId = C.MPV_EVENT_SET_PROPERTY_REPLY
	EVENT_COMMAND_REPLY      EventId = C.MPV_EVENT_COMMAND_REPLY
	EVENT_START_FILE         EventId = C.MPV_EVENT_START_FILE
	EVENT_END_FILE           EventId = C.MPV_EVENT_END_FILE
	EVENT_FILE_LOADED        EventId = C.MPV_EVENT_FILE_LOADED
	EVENT_CLIENT_MESSAGE     EventId = C.MPV_EVENT_CLIENT_MESSAGE
	EVENT_VIDEO_RECONFIG     EventId = C.MPV_EVENT_VIDEO_RECONFIG
	EVENT_AUDIO_RECONFIG     EventId = C.MPV_EVENT_AUDIO_RECONFIG
	EVENT_SEEK               EventId = C.MPV_EVENT_SEEK
	EVENT_PLAYBACK_RESTART   EventId = C.MPV_EVENT_PLAYBACK_RESTART
	EVENT_PROPERTY_CHANGE    EventId = C.MPV_EVENT_PROPERTY_CHANGE
	EVENT_QUEUE_OVERFLOW     EventId = C.MPV_EVENT_QUEUE_OVERFLOW
	EVENT_HOOK               EventId = C.MPV_EVENT_HOOK
	// @deprecated
	//EVENT_IDLE                  EventId = C.MPV_EVENT_IDLE
	//EVENT_TICK                  EventId = C.MPV_EVENT_TICK
)

// EventName returns the symbolic event name.
// C: const char *mpv_event_name(mpv_event_id event);
func EventName(id EventId) string {
	return C.GoString(C.mpv_event_name(C.mpv_event_id(id)))
}

type Event struct {
	EventId       EventId
	Error         error
	ReplyUserData uint64
	Data          unsafe.Pointer
	cEvent        *C.mpv_event
}

type EventLogMessage struct {
	Prefix   string
	Level    string
	Text     string
	LogLevel LogLevel
}

// LogMessage convert data to EventLogMessage
// MPV_EVENT_LOG_MESSAGE
func (e *Event) LogMessage() EventLogMessage {
	if e.EventId != EVENT_LOG_MESSAGE {
		panic("not a log message event")
	}
	s := (*C.mpv_event_log_message)(e.Data)
	return EventLogMessage{
		Prefix:   C.GoString(s.prefix),
		Level:    C.GoString(s.level),
		Text:     C.GoString(s.text),
		LogLevel: LogLevel(s.log_level),
	}
}

type EventProperty struct {
	Name   string
	Format Format
	Data   interface{}
}

// Property convert data to EventProperty
// MPV_EVENT_GET_PROPERTY_REPLY & MPV_EVENT_PROPERTY_CHANGE
func (e *Event) Property() EventProperty {
	if e.EventId != EVENT_PROPERTY_CHANGE && e.EventId != EVENT_GET_PROPERTY_REPLY {
		panic("not a property event")
	}
	s := (*C.mpv_event_property)(e.Data)
	p := EventProperty{
		Name:   C.GoString(s.name),
		Format: Format(s.format),
	}
	switch p.Format {
	case FORMAT_NONE:
		p.Data = nil
	case FORMAT_STRING, FORMAT_OSD_STRING:
		p.Data = C.GoString(*(**C.char)(s.data))
		// seems like i don't need to free lol
		// fmt.Println(s.data, *(**C.char)(s.data))
	case FORMAT_INT64:
		p.Data = int64(*(*C.int64_t)(s.data))
	case FORMAT_DOUBLE:
		p.Data = float64(*(*C.double)(s.data))
	case FORMAT_FLAG:
		p.Data = int(*(*C.int)(s.data)) == 1
	case FORMAT_NODE:
		p.Data = newNode((*C.mpv_node)(s.data))
	case FORMAT_BYTE_ARRAY:
		ba := (*C.mpv_byte_array)(s.data)
		if ba == nil || ba.size == 0 {
			p.Data = ByteArray{}
			break
		}
		p.Data = ByteArray(C.GoBytes(ba.data, C.int(ba.size)))
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
	s := (*C.mpv_event_hook)(e.Data)
	return EventHook{
		Name: C.GoString(s.name),
		ID:   uint64(s.id),
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
	s := (*C.mpv_event_command)(e.Data)
	return EventCommand{
		Result: newNode(&s.result),
	}
}

// ToNode converts the current event into a Node map.
// C: int mpv_event_to_node(mpv_node *dst, mpv_event *src);
func (e *Event) ToNode() (Node, error) {
	if e.cEvent == nil {
		return Node{}, ERROR_INVALID_PARAMETER
	}
	var dst C.mpv_node
	if err := newError(C.mpv_event_to_node(&dst, e.cEvent)); err != nil {
		return Node{}, err
	}
	defer C.mpv_free_node_contents(&dst)
	return newNode(&dst), nil
}
