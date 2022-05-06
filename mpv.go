package mpv

/*
#include <mpv/client.h>
#include <stdlib.h>
#include <stdint.h>
#cgo LDFLAGS: -L./lib -lmpv
//#cgo LDFLAGS: -lmpv
*/
import "C"
import (
	"unsafe"
)

/*
mpv_client_api_version
*/

func ClientApiVersion() uint32 {
	return uint32(C.mpv_client_api_version())
}

// Mpv mpv client.
type Mpv struct {
	handle *C.mpv_handle
}

func Create() *Mpv {
	ctx := C.mpv_create()
	if ctx == nil {
		return nil
	}
	return &Mpv{ctx}
}

func (m *Mpv) ClientName() string {
	return C.GoString(C.mpv_client_name(m.handle))
}

// not exists in linux
//func (m *Mpv) ClientId() int64 {
//	return int64(C.mpv_client_id(m.handle))
//}

func (m *Mpv) Initialize() error {
	return newError(C.mpv_initialize(m.handle))
}

func (m *Mpv) Destroy() {
	C.mpv_destroy(m.handle)
}

func (m *Mpv) TerminateDestroy() {
	C.mpv_terminate_destroy(m.handle)
}

func (m *Mpv) CreateClient(name string) *Mpv {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	client := C.mpv_create_client(m.handle, cname)
	if client == nil {
		return nil
	}
	return &Mpv{client}
}

func (m *Mpv) CreateWeakClient(name string) *Mpv {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	client := C.mpv_create_weak_client(m.handle, cname)
	if client == nil {
		return nil
	}
	return &Mpv{client}
}

func (m *Mpv) LoadConfigFile(filename string) int {
	cfn := C.CString(filename)
	defer C.free(unsafe.Pointer(cfn))
	return int(C.mpv_load_config_file(m.handle, cfn))
}

func (m *Mpv) GetTimeUS() int64 {
	return int64(C.mpv_get_time_us(m.handle))
}

/*
mpv_set_option
mpv_set_option_string
*/

func (m *Mpv) SetOption(name string, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ptr := getMpvDataPointer(format, data)
	defer C.free(ptr)
	return newError(C.mpv_set_option(m.handle, cname, C.mpv_format(format), ptr))
}

func (m *Mpv) SetOptionString(name, data string) error {
	cname := C.CString(name)
	cdata := C.CString(data)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cdata))
	return newError(C.mpv_set_option_string(m.handle, cname, cdata))
}

/*
mpv_set_property
mpv_set_property_async
mpv_set_property_string
*/

func (m *Mpv) SetProperty(name string, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ptr := getMpvDataPointer(format, data)
	defer C.free(ptr)
	return newError(C.mpv_set_property(m.handle, cname, C.mpv_format(format), ptr))
}

func (m *Mpv) SetPropertyString(name, data string) error {
	cname := C.CString(name)
	cdata := C.CString(data)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cdata))
	return newError(C.mpv_set_property_string(m.handle, cname, cdata))
}

func (m *Mpv) SetPropertyAsync(name string, replyUserdata uint64, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ptr := getMpvDataPointer(format, data)
	defer C.free(ptr)
	return newError(C.mpv_set_property_async(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format), ptr))
}

/*
mpv_get_property
mpv_get_property_async
mpv_get_property_osd_string
mpv_get_property_string
*/

func (m *Mpv) GetProperty(name string, format Format) (interface{}, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	switch format {
	case FORMAT_NONE:
		{
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), nil))
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
	case FORMAT_STRING, FORMAT_OSD_STRING:
		{
			var cval *C.char
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), unsafe.Pointer(&cval)))
			if err != nil {
				return nil, err
			}
			defer C.mpv_free(unsafe.Pointer(cval))
			return C.GoString(cval), nil
		}
	case FORMAT_INT64:
		{
			var cval C.int64_t
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), unsafe.Pointer(&cval)))
			if err != nil {
				return nil, err
			}
			return int64(cval), nil
		}
	case FORMAT_DOUBLE:
		{
			var cval C.double
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), unsafe.Pointer(&cval)))
			if err != nil {
				return nil, err
			}
			return float64(cval), nil
		}
	case FORMAT_FLAG:
		{
			var cval C.int
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), unsafe.Pointer(&cval)))
			if err != nil {
				return nil, err
			}
			return cval == 1, nil
		}
	case FORMAT_NODE:
		// FORMAT_NODE_ARRAY, FORMAT_NODE_MAP can only used within FORMAT_NODE
		{
			var cval C.mpv_node
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), unsafe.Pointer(&cval)))
			if err != nil {
				return nil, err
			}
			defer C.mpv_free_node_contents(&cval)
			return NewNode(&cval), nil
		}
	case FORMAT_BYTE_ARRAY:
		panic("Not implement")
	default:
		panic("No such Format")
	}
}

func (m *Mpv) GetPropertyString(name string) string {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cstr := C.mpv_get_property_string(m.handle, cname)
	if cstr != nil {
		defer C.mpv_free(unsafe.Pointer(cstr))
		return C.GoString(cstr)
	}
	return ""
}

func (m *Mpv) GetPropertyOsdString(name string) string {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cstr := C.mpv_get_property_osd_string(m.handle, cname)
	if cstr != nil {
		defer C.mpv_free(unsafe.Pointer(cstr))
		return C.GoString(cstr)
	}
	return ""
}

func (m *Mpv) GetPropertyAsync(name string, replyUserdata uint64, format Format) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return newError(C.mpv_get_property_async(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format)))
}

/*
mpv_observe_property
mpv_unobserve_property
*/

func (m *Mpv) ObserveProperty(replyUserdata uint64, name string, format Format) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return newError(C.mpv_observe_property(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format)))
}

func (m *Mpv) UnObserveProperty(registeredReplyUserdata uint64) error {
	return newError(C.mpv_unobserve_property(m.handle, C.uint64_t(registeredReplyUserdata)))
}

/*
mpv_request_event
mpv_request_log_messages
*/

func (m *Mpv) RequestEvent(event EventId, enable bool) error {
	return newError(C.mpv_request_event(m.handle, C.mpv_event_id(event), boolToCInt(enable)))
}

func (m *Mpv) RequestLogMessages(minLevel LogLevel) error {
	clevel := C.CString(LOG_LEVEL_STRING[minLevel])
	defer C.free(unsafe.Pointer(clevel))
	return newError(C.mpv_request_log_messages(m.handle, clevel))
}

/*
mpv_wait_async_requests
*/

func (m *Mpv) WaitAsyncRequests() {
	C.mpv_wait_async_requests(m.handle)
}

/*
mpv_wakeup
*/

func (m *Mpv) Wakeup() {
	C.mpv_wakeup(m.handle)
}

/*
mpv_wait_event
*/

func (m *Mpv) WaitEvent(timeout float64) *Event {
	var cevent *C.mpv_event
	cevent = C.mpv_wait_event(m.handle, C.double(timeout))
	if cevent == nil {
		return nil
	}

	event := Event{
		EventId:       EventId(cevent.event_id),
		Error:         newError(cevent.error),
		ReplyUserData: uint64(cevent.reply_userdata),
		Data:          unsafe.Pointer(cevent.data),
	}

	return &event
}

/*
mpv_command
mpv_command_async
mpv_command_string
*/

func (m *Mpv) Command(commands []string) error {
	carr := make([]*C.char, len(commands), len(commands)+1)
	for i, s := range commands {
		cStr := C.CString(s)
		carr[i] = cStr
		defer C.free(unsafe.Pointer(cStr))
	}
	return newError(C.mpv_command(m.handle, &carr[0]))
}

func (m *Mpv) CommandAsync(replyUserdata uint64, commands []string) error {
	carr := make([]*C.char, len(commands), len(commands)+1)
	for i, s := range commands {
		cStr := C.CString(s)
		carr[i] = cStr
		defer C.free(unsafe.Pointer(cStr))
	}
	return newError(C.mpv_command_async(m.handle, C.uint64_t(replyUserdata), &carr[0]))
}

func (m *Mpv) CommandString(command string) error {
	ccommand := C.CString(command)
	defer C.free(unsafe.Pointer(ccommand))
	return newError(C.mpv_command_string(m.handle, ccommand))
}

/*
mpv_abort_async_command
*/

func (m *Mpv) AbortAsyncCommand(replyUserdata uint64) {
	C.mpv_abort_async_command(m.handle, C.uint64_t(replyUserdata))
}
