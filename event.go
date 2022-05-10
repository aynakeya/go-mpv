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

func EventName(id EventId) string {
	return C.GoString(C.mpv_event_name(C.mpv_event_id(id)))
}

type Event struct {
	EventId       EventId
	Error         error
	ReplyUserData uint64
	Data          unsafe.Pointer
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
	if e.EventId != EVENT_PROPERTY_CHANGE || e.EventId != EVENT_SET_PROPERTY_REPLY {
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
		defer C.mpv_free_node_contents((*C.mpv_node)(s.data))
	case FORMAT_BYTE_ARRAY:
		panic("Not implement")
	default:
		panic("No such Format")
	}
	return p
}
