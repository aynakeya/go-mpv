//go:build cgo
// +build cgo

package mpv

/*
#include <mpv/client.h>
#include <stdlib.h>
#include <stdint.h>
//#cgo LDFLAGS: -L./lib -lmpv
#cgo LDFLAGS: -lmpv
*/
import "C"
import (
	"unsafe"
)

// ClientApiVersion returns the libmpv client API version.
// C: unsigned long mpv_client_api_version(void);
func ClientApiVersion() uint32 {
	return uint32(C.mpv_client_api_version())
}

// Mpv mpv client.
type Mpv struct {
	handle *C.mpv_handle
}

// Create creates a new mpv client handle.
// C: mpv_handle *mpv_create(void);
func Create() *Mpv {
	ctx := C.mpv_create()
	if ctx == nil {
		return nil
	}
	return &Mpv{ctx}
}

// ClientName returns the unique client name.
// C: const char *mpv_client_name(mpv_handle *ctx);
func (m *Mpv) ClientName() string {
	return C.GoString(C.mpv_client_name(m.handle))
}

// ClientId returns the unique client ID.
// C: int64_t mpv_client_id(mpv_handle *ctx);
func (m *Mpv) ClientId() int64 {
	return int64(C.mpv_client_id(m.handle))
}

// Initialize initializes an uninitialized mpv handle.
// C: int mpv_initialize(mpv_handle *ctx);
func (m *Mpv) Initialize() error {
	return newError(C.mpv_initialize(m.handle))
}

// Destroy destroys this mpv handle.
// C: void mpv_destroy(mpv_handle *ctx);
func (m *Mpv) Destroy() {
	C.mpv_destroy(m.handle)
}

// TerminateDestroy terminates playback and destroys the handle.
// C: void mpv_terminate_destroy(mpv_handle *ctx);
func (m *Mpv) TerminateDestroy() {
	C.mpv_terminate_destroy(m.handle)
}

// CreateClient creates a new client handle connected to the same core.
// C: mpv_handle *mpv_create_client(mpv_handle *ctx, const char *name);
func (m *Mpv) CreateClient(name string) *Mpv {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	client := C.mpv_create_client(m.handle, cname)
	if client == nil {
		return nil
	}
	return &Mpv{client}
}

// CreateWeakClient creates a new weak client handle.
// C: mpv_handle *mpv_create_weak_client(mpv_handle *ctx, const char *name);
func (m *Mpv) CreateWeakClient(name string) *Mpv {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	client := C.mpv_create_weak_client(m.handle, cname)
	if client == nil {
		return nil
	}
	return &Mpv{client}
}

// LoadConfigFile loads options from a config file.
// C: int mpv_load_config_file(mpv_handle *ctx, const char *filename);
func (m *Mpv) LoadConfigFile(filename string) int {
	cfn := C.CString(filename)
	defer C.free(unsafe.Pointer(cfn))
	return int(C.mpv_load_config_file(m.handle, cfn))
}

// GetTimeUS returns libmpv monotonic time in microseconds.
// C: int64_t mpv_get_time_us(mpv_handle *ctx);
func (m *Mpv) GetTimeUS() int64 {
	return int64(C.mpv_get_time_us(m.handle))
}

// GetTimeNS returns libmpv monotonic time in nanoseconds.
// C: int64_t mpv_get_time_ns(mpv_handle *ctx);
func (m *Mpv) GetTimeNS() int64 {
	return int64(C.mpv_get_time_ns(m.handle))
}

// SetOption sets an option value with an explicit format.
// C: int mpv_set_option(mpv_handle *ctx, const char *name, mpv_format format, void *data);
func (m *Mpv) SetOption(name string, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ptr := mallocMpvDataPointer(format, data)
	defer freeMpvDataPointer(format, ptr)
	return newError(C.mpv_set_option(m.handle, cname, C.mpv_format(format), ptr))
}

// SetOptionString sets an option value as string.
// C: int mpv_set_option_string(mpv_handle *ctx, const char *name, const char *data);
func (m *Mpv) SetOptionString(name, data string) error {
	cname := C.CString(name)
	cdata := C.CString(data)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cdata))
	return newError(C.mpv_set_option_string(m.handle, cname, cdata))
}

// SetProperty sets a property value with an explicit format.
// C: int mpv_set_property(mpv_handle *ctx, const char *name, mpv_format format, void *data);
func (m *Mpv) SetProperty(name string, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ptr := mallocMpvDataPointer(format, data)
	defer freeMpvDataPointer(format, ptr)
	return newError(C.mpv_set_property(m.handle, cname, C.mpv_format(format), ptr))
}

// SetPropertyString sets a property value as string.
// C: int mpv_set_property_string(mpv_handle *ctx, const char *name, const char *data);
func (m *Mpv) SetPropertyString(name, data string) error {
	cname := C.CString(name)
	cdata := C.CString(data)
	defer C.free(unsafe.Pointer(cname))
	defer C.free(unsafe.Pointer(cdata))
	return newError(C.mpv_set_property_string(m.handle, cname, cdata))
}

// SetPropertyAsync queues an asynchronous property write.
// C: int mpv_set_property_async(mpv_handle *ctx, uint64_t reply_userdata, const char *name, mpv_format format, void *data);
func (m *Mpv) SetPropertyAsync(name string, replyUserdata uint64, format Format, data interface{}) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	ptr := mallocMpvDataPointer(format, data)
	defer freeMpvDataPointer(format, ptr)
	return newError(C.mpv_set_property_async(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format), ptr))
}

// GetProperty gets a property value in the requested format.
// C: int mpv_get_property(mpv_handle *ctx, const char *name, mpv_format format, void *data);
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
			return newNode(&cval), nil
		}
	case FORMAT_BYTE_ARRAY:
		{
			var cval C.mpv_byte_array
			err := newError(C.mpv_get_property(m.handle, cname, C.mpv_format(format), unsafe.Pointer(&cval)))
			if err != nil {
				return nil, err
			}
			if cval.data != nil {
				defer C.mpv_free(cval.data)
			}
			if cval.data == nil || cval.size == 0 {
				return ByteArray{}, nil
			}
			return ByteArray(C.GoBytes(cval.data, C.int(cval.size))), nil
		}
	default:
		panic("No such Format")
	}
}

// GetPropertyString gets a property value as raw string.
// C: char *mpv_get_property_string(mpv_handle *ctx, const char *name);
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

// GetPropertyOsdString gets a property value as OSD-formatted string.
// C: char *mpv_get_property_osd_string(mpv_handle *ctx, const char *name);
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

// GetPropertyAsync queues an asynchronous property read.
// C: int mpv_get_property_async(mpv_handle *ctx, uint64_t reply_userdata, const char *name, mpv_format format);
func (m *Mpv) GetPropertyAsync(name string, replyUserdata uint64, format Format) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return newError(C.mpv_get_property_async(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format)))
}

// DelProperty deletes the named property.
// C: int mpv_del_property(mpv_handle *ctx, const char *name);
func (m *Mpv) DelProperty(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return newError(C.mpv_del_property(m.handle, cname))
}

// ObserveProperty subscribes to property change events.
// C: int mpv_observe_property(mpv_handle *ctx, uint64_t reply_userdata, const char *name, mpv_format format);
func (m *Mpv) ObserveProperty(replyUserdata uint64, name string, format Format) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return newError(C.mpv_observe_property(m.handle, C.uint64_t(replyUserdata), cname, C.mpv_format(format)))
}

// UnObserveProperty unsubscribes previously observed properties by userdata.
// C: int mpv_unobserve_property(mpv_handle *ctx, uint64_t registered_reply_userdata);
func (m *Mpv) UnObserveProperty(registeredReplyUserdata uint64) error {
	return newError(C.mpv_unobserve_property(m.handle, C.uint64_t(registeredReplyUserdata)))
}

// RequestEvent enables or disables receiving a specific event.
// C: int mpv_request_event(mpv_handle *ctx, mpv_event_id event, int enable);
func (m *Mpv) RequestEvent(event EventId, enable bool) error {
	return newError(C.mpv_request_event(m.handle, C.mpv_event_id(event), boolToCInt(enable)))
}

// RequestLogMessages configures log message event verbosity.
// C: int mpv_request_log_messages(mpv_handle *ctx, const char *min_level);
func (m *Mpv) RequestLogMessages(minLevel LogLevel) error {
	clevel := C.CString(LOG_LEVEL_STRING[minLevel])
	defer C.free(unsafe.Pointer(clevel))
	return newError(C.mpv_request_log_messages(m.handle, clevel))
}

// WaitAsyncRequests blocks until async requests are completed.
// C: void mpv_wait_async_requests(mpv_handle *ctx);
func (m *Mpv) WaitAsyncRequests() {
	C.mpv_wait_async_requests(m.handle)
}

// Wakeup interrupts a blocking WaitEvent call.
// C: void mpv_wakeup(mpv_handle *ctx);
func (m *Mpv) Wakeup() {
	C.mpv_wakeup(m.handle)
}

// WaitEvent waits for the next event or timeout.
// C: mpv_event *mpv_wait_event(mpv_handle *ctx, double timeout);
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
		cEvent:        cevent,
	}

	return &event
}

// Command runs a command synchronously with string arguments.
// C: int mpv_command(mpv_handle *ctx, const char **args);
func (m *Mpv) Command(commands []string) error {
	if len(commands) == 0 {
		return ERROR_INVALID_PARAMETER
	}
	carr := make([]*C.char, len(commands), len(commands)+1)
	for i, s := range commands {
		cStr := C.CString(s)
		carr[i] = cStr
		defer C.free(unsafe.Pointer(cStr))
	}
	return newError(C.mpv_command(m.handle, &carr[0]))
}

// CommandAsync queues a command and returns immediately.
// C: int mpv_command_async(mpv_handle *ctx, uint64_t reply_userdata, const char **args);
func (m *Mpv) CommandAsync(replyUserdata uint64, commands []string) error {
	if len(commands) == 0 {
		return ERROR_INVALID_PARAMETER
	}
	carr := make([]*C.char, len(commands), len(commands)+1)
	for i, s := range commands {
		cStr := C.CString(s)
		carr[i] = cStr
		defer C.free(unsafe.Pointer(cStr))
	}
	return newError(C.mpv_command_async(m.handle, C.uint64_t(replyUserdata), &carr[0]))
}

// CommandRet runs a command synchronously and returns structured result data.
// C: int mpv_command_ret(mpv_handle *ctx, const char **args, mpv_node *result);
func (m *Mpv) CommandRet(commands []string) (Node, error) {
	if len(commands) == 0 {
		return Node{}, ERROR_INVALID_PARAMETER
	}
	carr := make([]*C.char, len(commands), len(commands)+1)
	for i, s := range commands {
		cStr := C.CString(s)
		carr[i] = cStr
		defer C.free(unsafe.Pointer(cStr))
	}
	var result C.mpv_node
	cErr := C.mpv_command_ret(m.handle, &carr[0], &result)
	if err := newError(cErr); err != nil {
		return Node{}, err
	}
	defer C.mpv_free_node_contents(&result)
	return newNode(&result), nil
}

// CommandNode runs a structured command synchronously.
// C: int mpv_command_node(mpv_handle *ctx, mpv_node *args, mpv_node *result);
func (m *Mpv) CommandNode(command Node) (Node, error) {
	cnode := command.CNode()
	defer freeMpvDataPointer(FORMAT_NODE, unsafe.Pointer(cnode))
	var result C.mpv_node
	if err := newError(C.mpv_command_node(m.handle, cnode, &result)); err != nil {
		return Node{}, err
	}
	defer C.mpv_free_node_contents(&result)
	return newNode(&result), nil
}

// CommandNodeAsync queues a structured command and returns immediately.
// C: int mpv_command_node_async(mpv_handle *ctx, uint64_t reply_userdata, mpv_node *args);
func (m *Mpv) CommandNodeAsync(replyUserdata uint64, command Node) error {
	cnode := command.CNode()
	defer freeMpvDataPointer(FORMAT_NODE, unsafe.Pointer(cnode))
	return newError(C.mpv_command_node_async(m.handle, C.uint64_t(replyUserdata), cnode))
}

// CommandString runs a command string using mpv parser semantics.
// C: int mpv_command_string(mpv_handle *ctx, const char *args);
func (m *Mpv) CommandString(command string) error {
	ccommand := C.CString(command)
	defer C.free(unsafe.Pointer(ccommand))
	return newError(C.mpv_command_string(m.handle, ccommand))
}

// AbortAsyncCommand requests cancellation for async commands by userdata.
// C: void mpv_abort_async_command(mpv_handle *ctx, uint64_t reply_userdata);
func (m *Mpv) AbortAsyncCommand(replyUserdata uint64) {
	C.mpv_abort_async_command(m.handle, C.uint64_t(replyUserdata))
}

// HookAdd registers a hook handler.
// C: int mpv_hook_add(mpv_handle *ctx, uint64_t reply_userdata, const char *name, int priority);
func (m *Mpv) HookAdd(replyUserdata uint64, name string, priority int) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return newError(C.mpv_hook_add(m.handle, C.uint64_t(replyUserdata), cname, C.int(priority)))
}

// HookContinue continues a pending hook event.
// C: int mpv_hook_continue(mpv_handle *ctx, uint64_t id);
func (m *Mpv) HookContinue(id uint64) error {
	return newError(C.mpv_hook_continue(m.handle, C.uint64_t(id)))
}
