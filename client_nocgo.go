//go:build !cgo
// +build !cgo

package mpv

import (
	"math"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

type mpvNodeRaw struct {
	U      uint64
	Format int32
	_      int32
}

func (n *mpvNodeRaw) setPointer(p unsafe.Pointer) {
	n.U = uint64(uintptr(p))
}

func (n *mpvNodeRaw) pointer() unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&n.U))
}

func (n *mpvNodeRaw) setInt32(v int32) {
	n.U = uint64(uint32(v))
}

func (n *mpvNodeRaw) int32() int32 {
	return int32(uint32(n.U))
}

func (n *mpvNodeRaw) setInt64(v int64) {
	n.U = uint64(v)
}

func (n *mpvNodeRaw) int64() int64 {
	return int64(n.U)
}

func (n *mpvNodeRaw) setFloat64(v float64) {
	n.U = math.Float64bits(v)
}

func (n *mpvNodeRaw) float64() float64 {
	return math.Float64frombits(n.U)
}

type mpvNodeListRaw struct {
	Num    int32
	Values *mpvNodeRaw
	Keys   **byte
}

type mpvByteArrayRaw struct {
	Data unsafe.Pointer
	Size uintptr
}

type mpvEventRaw struct {
	EventID       int32
	Error         int32
	ReplyUserData uint64
	Data          unsafe.Pointer
}

type mpvEventPropertyRaw struct {
	Name   *byte
	Format int32
	Data   unsafe.Pointer
}

type mpvEventLogMessageRaw struct {
	Prefix   *byte
	Level    *byte
	Text     *byte
	LogLevel int32
}

type mpvEventHookRaw struct {
	Name *byte
	ID   uint64
}

type mpvEventCommandRaw struct {
	Result mpvNodeRaw
}

type puregoBackend struct {
	once sync.Once
	err  error
	lib  uintptr

	clientAPIVersion   func() uint64
	create             func() unsafe.Pointer
	clientName         func(unsafe.Pointer) *byte
	clientID           func(unsafe.Pointer) int64
	initialize         func(unsafe.Pointer) int32
	destroy            func(unsafe.Pointer)
	terminateDestroy   func(unsafe.Pointer)
	createClient       func(unsafe.Pointer, *byte) unsafe.Pointer
	createWeakClient   func(unsafe.Pointer, *byte) unsafe.Pointer
	loadConfigFile     func(unsafe.Pointer, *byte) int32
	getTimeUS          func(unsafe.Pointer) int64
	getTimeNS          func(unsafe.Pointer) int64
	setOption          func(unsafe.Pointer, *byte, int32, unsafe.Pointer) int32
	setOptionString    func(unsafe.Pointer, *byte, *byte) int32
	setProperty        func(unsafe.Pointer, *byte, int32, unsafe.Pointer) int32
	setPropertyString  func(unsafe.Pointer, *byte, *byte) int32
	setPropertyAsync   func(unsafe.Pointer, uint64, *byte, int32, unsafe.Pointer) int32
	getProperty        func(unsafe.Pointer, *byte, int32, unsafe.Pointer) int32
	getPropertyString  func(unsafe.Pointer, *byte) *byte
	getPropertyOSD     func(unsafe.Pointer, *byte) *byte
	getPropertyAsync   func(unsafe.Pointer, uint64, *byte, int32) int32
	delProperty        func(unsafe.Pointer, *byte) int32
	observeProperty    func(unsafe.Pointer, uint64, *byte, int32) int32
	unobserveProperty  func(unsafe.Pointer, uint64) int32
	requestEvent       func(unsafe.Pointer, int32, int32) int32
	requestLogMessages func(unsafe.Pointer, *byte) int32
	waitAsyncRequests  func(unsafe.Pointer)
	wakeup             func(unsafe.Pointer)
	waitEvent          func(unsafe.Pointer, float64) *mpvEventRaw
	command            func(unsafe.Pointer, **byte) int32
	commandAsync       func(unsafe.Pointer, uint64, **byte) int32
	commandRet         func(unsafe.Pointer, **byte, *mpvNodeRaw) int32
	commandNode        func(unsafe.Pointer, *mpvNodeRaw, *mpvNodeRaw) int32
	commandNodeAsync   func(unsafe.Pointer, uint64, *mpvNodeRaw) int32
	commandString      func(unsafe.Pointer, *byte) int32
	abortAsyncCommand  func(unsafe.Pointer, uint64)
	hookAdd            func(unsafe.Pointer, uint64, *byte, int32) int32
	hookContinue       func(unsafe.Pointer, uint64) int32
	errorString        func(int32) *byte
	eventName          func(int32) *byte
	eventToNode        func(*mpvNodeRaw, *mpvEventRaw) int32
	freeNodeContents   func(*mpvNodeRaw)
	free               func(unsafe.Pointer)

	renderContextCreate func(*unsafe.Pointer, unsafe.Pointer, *mpvRenderParamRaw) int32
	// mpv_render_context_set_parameter/get_info take mpv_render_param by value.
	// On the supported purego ABI this is passed as the type word plus data
	// word, instead of a pointer to the struct used by create/render.
	renderContextSetParameter      func(unsafe.Pointer, uintptr, unsafe.Pointer) int32
	renderContextGetInfo           func(unsafe.Pointer, uintptr, unsafe.Pointer) int32
	renderContextSetUpdateCallback func(unsafe.Pointer, uintptr, unsafe.Pointer)
	renderContextUpdate            func(unsafe.Pointer) uint64
	renderContextRender            func(unsafe.Pointer, *mpvRenderParamRaw) int32
	renderContextReportSwap        func(unsafe.Pointer)
	renderContextFree              func(unsafe.Pointer)
}

var backend puregoBackend

func libMPVCandidates(goos string) []string {
	switch goos {
	case "windows":
		return []string{"mpv-2.dll", "libmpv-2.dll", "mpv.dll", "libmpv.dll"}
	case "darwin":
		return []string{"libmpv.2.dylib", "libmpv.dylib"}
	case "linux":
		return []string{"libmpv.so.2", "libmpv.so"}
	case "freebsd":
		return []string{"libmpv.so.2", "libmpv.so"}
	case "openbsd":
		return []string{"libmpv.so.2", "libmpv.so"}
	case "netbsd":
		return []string{"libmpv.so.2", "libmpv.so"}
	default:
		return []string{"libmpv.so.2", "libmpv.so", "libmpv.dylib", "mpv-2.dll"}
	}
}

func (b *puregoBackend) load() {
	names := libMPVCandidates(runtime.GOOS)
	for _, name := range names {
		lib, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			b.lib = lib
			break
		}
	}
	if b.lib == 0 {
		b.err = ErrCGODisabled
		return
	}
	b.bindRequired()
}

func (b *puregoBackend) bindRequired() {
	register := func(fptr interface{}, name string) {
		if b.err != nil {
			return
		}
		sym, err := purego.Dlsym(b.lib, name)
		if err != nil {
			b.err = err
			return
		}
		purego.RegisterFunc(fptr, sym)
	}

	register(&b.clientAPIVersion, "mpv_client_api_version")
	register(&b.create, "mpv_create")
	register(&b.clientName, "mpv_client_name")
	register(&b.clientID, "mpv_client_id")
	register(&b.initialize, "mpv_initialize")
	register(&b.destroy, "mpv_destroy")
	register(&b.terminateDestroy, "mpv_terminate_destroy")
	register(&b.createClient, "mpv_create_client")
	register(&b.createWeakClient, "mpv_create_weak_client")
	register(&b.loadConfigFile, "mpv_load_config_file")
	register(&b.getTimeUS, "mpv_get_time_us")
	register(&b.getTimeNS, "mpv_get_time_ns")
	register(&b.setOption, "mpv_set_option")
	register(&b.setOptionString, "mpv_set_option_string")
	register(&b.setProperty, "mpv_set_property")
	register(&b.setPropertyString, "mpv_set_property_string")
	register(&b.setPropertyAsync, "mpv_set_property_async")
	register(&b.getProperty, "mpv_get_property")
	register(&b.getPropertyString, "mpv_get_property_string")
	register(&b.getPropertyOSD, "mpv_get_property_osd_string")
	register(&b.getPropertyAsync, "mpv_get_property_async")
	register(&b.delProperty, "mpv_del_property")
	register(&b.observeProperty, "mpv_observe_property")
	register(&b.unobserveProperty, "mpv_unobserve_property")
	register(&b.requestEvent, "mpv_request_event")
	register(&b.requestLogMessages, "mpv_request_log_messages")
	register(&b.waitAsyncRequests, "mpv_wait_async_requests")
	register(&b.wakeup, "mpv_wakeup")
	register(&b.waitEvent, "mpv_wait_event")
	register(&b.command, "mpv_command")
	register(&b.commandAsync, "mpv_command_async")
	register(&b.commandRet, "mpv_command_ret")
	register(&b.commandNode, "mpv_command_node")
	register(&b.commandNodeAsync, "mpv_command_node_async")
	register(&b.commandString, "mpv_command_string")
	register(&b.abortAsyncCommand, "mpv_abort_async_command")
	register(&b.hookAdd, "mpv_hook_add")
	register(&b.hookContinue, "mpv_hook_continue")
	register(&b.errorString, "mpv_error_string")
	register(&b.eventName, "mpv_event_name")
	register(&b.eventToNode, "mpv_event_to_node")
	register(&b.freeNodeContents, "mpv_free_node_contents")
	register(&b.free, "mpv_free")

	register(&b.renderContextCreate, "mpv_render_context_create")
	register(&b.renderContextSetParameter, "mpv_render_context_set_parameter")
	register(&b.renderContextGetInfo, "mpv_render_context_get_info")
	register(&b.renderContextSetUpdateCallback, "mpv_render_context_set_update_callback")
	register(&b.renderContextUpdate, "mpv_render_context_update")
	register(&b.renderContextRender, "mpv_render_context_render")
	register(&b.renderContextReportSwap, "mpv_render_context_report_swap")
	register(&b.renderContextFree, "mpv_render_context_free")
}

func ensureBackend() error {
	backend.once.Do(backend.load)
	return backend.err
}

func cString(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

func goString(p *byte) string {
	if p == nil {
		return ""
	}
	buf := make([]byte, 0, 32)
	i := uintptr(0)
	for {
		b := *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + i))
		if b == 0 {
			break
		}
		buf = append(buf, b)
		i++
	}
	return string(buf)
}

func copyBytes(p unsafe.Pointer, n uintptr) []byte {
	if p == nil || n == 0 {
		return nil
	}
	out := make([]byte, int(n))
	for i := uintptr(0); i < n; i++ {
		out[i] = *(*byte)(unsafe.Pointer(uintptr(p) + i))
	}
	return out
}

func boolToInt32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func marshalCmdArgs(commands []string) ([]*byte, [][]byte) {
	args := make([]*byte, len(commands)+1)
	refs := make([][]byte, len(commands))
	for i, s := range commands {
		refs[i] = cString(s)
		args[i] = &refs[i][0]
	}
	args[len(commands)] = nil
	return args, refs
}

type nodeArena struct {
	keep []interface{}
}

func (a *nodeArena) hold(v interface{}) {
	a.keep = append(a.keep, v)
}

func marshalNode(n Node, a *nodeArena) (*mpvNodeRaw, error) {
	raw := &mpvNodeRaw{Format: int32(n.Format)}
	a.hold(raw)

	switch n.Format {
	case FORMAT_NONE:
		return raw, nil
	case FORMAT_STRING, FORMAT_OSD_STRING:
		b := cString(n.Value.(string))
		a.hold(b)
		raw.setPointer(unsafe.Pointer(&b[0]))
		return raw, nil
	case FORMAT_FLAG:
		raw.setInt32(boolToInt32(n.Value.(bool)))
		return raw, nil
	case FORMAT_INT64:
		raw.setInt64(n.Value.(int64))
		return raw, nil
	case FORMAT_DOUBLE:
		raw.setFloat64(n.Value.(float64))
		return raw, nil
	case FORMAT_BYTE_ARRAY:
		ba := []byte(n.Value.(ByteArray))
		dup := make([]byte, len(ba))
		copy(dup, ba)
		a.hold(dup)
		vr := &mpvByteArrayRaw{Size: uintptr(len(dup))}
		if len(dup) > 0 {
			vr.Data = unsafe.Pointer(&dup[0])
		}
		a.hold(vr)
		raw.setPointer(unsafe.Pointer(vr))
		return raw, nil
	case FORMAT_NODE_ARRAY:
		nodes := n.Value.([]Node)
		values := make([]mpvNodeRaw, len(nodes))
		a.hold(values)
		for i := range nodes {
			ch, err := marshalNode(nodes[i], a)
			if err != nil {
				return nil, err
			}
			values[i] = *ch
		}
		list := &mpvNodeListRaw{Num: int32(len(values))}
		if len(values) > 0 {
			list.Values = &values[0]
		}
		a.hold(list)
		raw.setPointer(unsafe.Pointer(list))
		return raw, nil
	case FORMAT_NODE_MAP:
		m := n.Value.(map[string]Node)
		values := make([]mpvNodeRaw, len(m))
		keys := make([]*byte, len(m))
		a.hold(values)
		a.hold(keys)
		i := 0
		for k, v := range m {
			kb := cString(k)
			a.hold(kb)
			keys[i] = &kb[0]
			ch, err := marshalNode(v, a)
			if err != nil {
				return nil, err
			}
			values[i] = *ch
			i++
		}
		list := &mpvNodeListRaw{Num: int32(len(values))}
		if len(values) > 0 {
			list.Values = &values[0]
			list.Keys = (**byte)(unsafe.Pointer(&keys[0]))
		}
		a.hold(list)
		raw.setPointer(unsafe.Pointer(list))
		return raw, nil
	default:
		return nil, ERROR_NOT_IMPLEMENTED
	}
}

func nodeFromRaw(raw *mpvNodeRaw) Node {
	if raw == nil {
		return Node{Format: FORMAT_NONE}
	}
	format := Format(raw.Format)
	switch format {
	case FORMAT_NONE:
		return Node{Format: FORMAT_NONE}
	case FORMAT_STRING, FORMAT_OSD_STRING:
		return Node{Format: format, Value: goString((*byte)(raw.pointer()))}
	case FORMAT_FLAG:
		return Node{Format: FORMAT_FLAG, Value: raw.int32() == 1}
	case FORMAT_INT64:
		return Node{Format: FORMAT_INT64, Value: raw.int64()}
	case FORMAT_DOUBLE:
		return Node{Format: FORMAT_DOUBLE, Value: raw.float64()}
	case FORMAT_BYTE_ARRAY:
		ba := (*mpvByteArrayRaw)(raw.pointer())
		if ba == nil {
			return Node{Format: FORMAT_BYTE_ARRAY, Value: ByteArray{}}
		}
		return Node{Format: FORMAT_BYTE_ARRAY, Value: ByteArray(copyBytes(ba.Data, ba.Size))}
	case FORMAT_NODE_ARRAY:
		list := (*mpvNodeListRaw)(raw.pointer())
		if list == nil || list.Num <= 0 || list.Values == nil {
			return Node{Format: FORMAT_NODE_ARRAY, Value: []Node{}}
		}
		out := make([]Node, int(list.Num))
		values := unsafe.Slice(list.Values, int(list.Num))
		for i := range values {
			out[i] = nodeFromRaw(&values[i])
		}
		return Node{Format: FORMAT_NODE_ARRAY, Value: out}
	case FORMAT_NODE_MAP:
		list := (*mpvNodeListRaw)(raw.pointer())
		if list == nil || list.Num <= 0 || list.Values == nil {
			return Node{Format: FORMAT_NODE_MAP, Value: map[string]Node{}}
		}
		out := make(map[string]Node, int(list.Num))
		values := unsafe.Slice(list.Values, int(list.Num))
		keys := unsafe.Slice(list.Keys, int(list.Num))
		for i := range values {
			out[goString(keys[i])] = nodeFromRaw(&values[i])
		}
		return Node{Format: FORMAT_NODE_MAP, Value: out}
	default:
		return Node{Format: FORMAT_NONE}
	}
}

// ClientApiVersion returns the libmpv client API version.
func ClientApiVersion() uint32 {
	if err := ensureBackend(); err != nil {
		return 0
	}
	return uint32(backend.clientAPIVersion())
}

// Mpv wraps a libmpv handle in the nocgo backend.
type Mpv struct {
	handle unsafe.Pointer
}

// Create creates a new mpv client handle.
func Create() *Mpv {
	if err := ensureBackend(); err != nil {
		return nil
	}
	h := backend.create()
	if h == nil {
		return nil
	}
	return &Mpv{handle: h}
}

// ClientName returns the unique client name.
func (m *Mpv) ClientName() string {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ""
	}
	return goString(backend.clientName(m.handle))
}

// ClientId returns the unique client ID.
func (m *Mpv) ClientId() int64 {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return 0
	}
	return backend.clientID(m.handle)
}

// Initialize initializes an uninitialized mpv handle.
func (m *Mpv) Initialize() error {
	if m == nil || m.handle == nil {
		return ErrCGODisabled
	}
	if err := ensureBackend(); err != nil {
		return err
	}
	return newError(backend.initialize(m.handle))
}

// Destroy destroys this mpv handle.
func (m *Mpv) Destroy() {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return
	}
	backend.destroy(m.handle)
	m.handle = nil
}

// TerminateDestroy terminates playback and destroys the handle.
func (m *Mpv) TerminateDestroy() {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return
	}
	backend.terminateDestroy(m.handle)
	m.handle = nil
}

// CreateClient creates a new client handle connected to the same core.
func (m *Mpv) CreateClient(name string) *Mpv {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return nil
	}
	cname := cString(name)
	h := backend.createClient(m.handle, &cname[0])
	runtime.KeepAlive(cname)
	if h == nil {
		return nil
	}
	return &Mpv{handle: h}
}

// CreateWeakClient creates a new weak client handle.
func (m *Mpv) CreateWeakClient(name string) *Mpv {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return nil
	}
	cname := cString(name)
	h := backend.createWeakClient(m.handle, &cname[0])
	runtime.KeepAlive(cname)
	if h == nil {
		return nil
	}
	return &Mpv{handle: h}
}

// LoadConfigFile loads options from a config file.
func (m *Mpv) LoadConfigFile(filename string) int {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return int(ERROR_NOT_IMPLEMENTED)
	}
	cname := cString(filename)
	ret := backend.loadConfigFile(m.handle, &cname[0])
	runtime.KeepAlive(cname)
	return int(ret)
}

// GetTimeUS returns libmpv monotonic time in microseconds.
func (m *Mpv) GetTimeUS() int64 {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return 0
	}
	return backend.getTimeUS(m.handle)
}

// GetTimeNS returns libmpv monotonic time in nanoseconds.
func (m *Mpv) GetTimeNS() int64 {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return 0
	}
	return backend.getTimeNS(m.handle)
}

func marshalBasicPropertyValue(format Format, data interface{}) (unsafe.Pointer, [][]byte, error) {
	var keep [][]byte
	switch format {
	case FORMAT_NONE:
		return nil, keep, nil
	case FORMAT_STRING, FORMAT_OSD_STRING:
		cs := cString(data.(string))
		csp := &cs[0]
		keep = append(keep, cs)
		return unsafe.Pointer(&csp), keep, nil
	case FORMAT_FLAG:
		v := boolToInt32(data.(bool))
		return unsafe.Pointer(&v), keep, nil
	case FORMAT_INT64:
		v, ok := data.(int64)
		if !ok {
			v = int64(data.(int))
		}
		return unsafe.Pointer(&v), keep, nil
	case FORMAT_DOUBLE:
		v := data.(float64)
		return unsafe.Pointer(&v), keep, nil
	default:
		return nil, keep, ERROR_NOT_IMPLEMENTED
	}
}

// SetOption sets an option value with an explicit format.
func (m *Mpv) SetOption(name string, format Format, data interface{}) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ptr, keep, err := marshalBasicPropertyValue(format, data)
	if err != nil {
		return err
	}
	ret := backend.setOption(m.handle, &cname[0], int32(format), ptr)
	runtime.KeepAlive(cname)
	runtime.KeepAlive(keep)
	return newError(ret)
}

// SetOptionString sets an option value as string.
func (m *Mpv) SetOptionString(name, data string) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	cdata := cString(data)
	ret := backend.setOptionString(m.handle, &cname[0], &cdata[0])
	runtime.KeepAlive(cname)
	runtime.KeepAlive(cdata)
	return newError(ret)
}

// SetProperty sets a property value with an explicit format.
func (m *Mpv) SetProperty(name string, format Format, data interface{}) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ptr, keep, err := marshalBasicPropertyValue(format, data)
	if err != nil {
		return err
	}
	ret := backend.setProperty(m.handle, &cname[0], int32(format), ptr)
	runtime.KeepAlive(cname)
	runtime.KeepAlive(keep)
	return newError(ret)
}

// SetPropertyString sets a property value as string.
func (m *Mpv) SetPropertyString(name, data string) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	cdata := cString(data)
	ret := backend.setPropertyString(m.handle, &cname[0], &cdata[0])
	runtime.KeepAlive(cname)
	runtime.KeepAlive(cdata)
	return newError(ret)
}

// SetPropertyAsync queues an asynchronous property write.
func (m *Mpv) SetPropertyAsync(name string, replyUserdata uint64, format Format, data interface{}) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ptr, keep, err := marshalBasicPropertyValue(format, data)
	if err != nil {
		return err
	}
	ret := backend.setPropertyAsync(m.handle, replyUserdata, &cname[0], int32(format), ptr)
	runtime.KeepAlive(cname)
	runtime.KeepAlive(keep)
	return newError(ret)
}

// GetProperty gets a property value in the requested format.
func (m *Mpv) GetProperty(name string, format Format) (interface{}, error) {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return nil, ErrCGODisabled
	}
	cname := cString(name)
	switch format {
	case FORMAT_NONE:
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), nil))
		runtime.KeepAlive(cname)
		return nil, err
	case FORMAT_STRING, FORMAT_OSD_STRING:
		var cval *byte
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), unsafe.Pointer(&cval)))
		runtime.KeepAlive(cname)
		if err != nil {
			return nil, err
		}
		defer backend.free(unsafe.Pointer(cval))
		return goString(cval), nil
	case FORMAT_INT64:
		var v int64
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), unsafe.Pointer(&v)))
		runtime.KeepAlive(cname)
		return v, err
	case FORMAT_DOUBLE:
		var v float64
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), unsafe.Pointer(&v)))
		runtime.KeepAlive(cname)
		return v, err
	case FORMAT_FLAG:
		var v int32
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), unsafe.Pointer(&v)))
		runtime.KeepAlive(cname)
		return v == 1, err
	case FORMAT_NODE:
		var v mpvNodeRaw
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), unsafe.Pointer(&v)))
		runtime.KeepAlive(cname)
		if err != nil {
			return nil, err
		}
		defer backend.freeNodeContents(&v)
		n := nodeFromRaw(&v)
		return n, nil
	case FORMAT_BYTE_ARRAY:
		var v mpvByteArrayRaw
		err := newError(backend.getProperty(m.handle, &cname[0], int32(format), unsafe.Pointer(&v)))
		runtime.KeepAlive(cname)
		if err != nil {
			return nil, err
		}
		defer backend.free(v.Data)
		return ByteArray(copyBytes(v.Data, v.Size)), nil
	default:
		runtime.KeepAlive(cname)
		return nil, ERROR_NOT_IMPLEMENTED
	}
}

// GetPropertyString gets a property value as raw string.
func (m *Mpv) GetPropertyString(name string) string {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ""
	}
	cname := cString(name)
	p := backend.getPropertyString(m.handle, &cname[0])
	runtime.KeepAlive(cname)
	if p == nil {
		return ""
	}
	defer backend.free(unsafe.Pointer(p))
	return goString(p)
}

// GetPropertyOsdString gets a property value as OSD-formatted string.
func (m *Mpv) GetPropertyOsdString(name string) string {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ""
	}
	cname := cString(name)
	p := backend.getPropertyOSD(m.handle, &cname[0])
	runtime.KeepAlive(cname)
	if p == nil {
		return ""
	}
	defer backend.free(unsafe.Pointer(p))
	return goString(p)
}

// GetPropertyAsync queues an asynchronous property read.
func (m *Mpv) GetPropertyAsync(name string, replyUserdata uint64, format Format) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ret := backend.getPropertyAsync(m.handle, replyUserdata, &cname[0], int32(format))
	runtime.KeepAlive(cname)
	return newError(ret)
}

// DelProperty deletes the named property.
func (m *Mpv) DelProperty(name string) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ret := backend.delProperty(m.handle, &cname[0])
	runtime.KeepAlive(cname)
	return newError(ret)
}

// ObserveProperty subscribes to property change events.
func (m *Mpv) ObserveProperty(replyUserdata uint64, name string, format Format) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ret := backend.observeProperty(m.handle, replyUserdata, &cname[0], int32(format))
	runtime.KeepAlive(cname)
	return newError(ret)
}

// UnObserveProperty unsubscribes previously observed properties by userdata.
func (m *Mpv) UnObserveProperty(registeredReplyUserdata uint64) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	return newError(backend.unobserveProperty(m.handle, registeredReplyUserdata))
}

// RequestEvent enables or disables receiving a specific event.
func (m *Mpv) RequestEvent(event EventId, enable bool) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	return newError(backend.requestEvent(m.handle, int32(event), boolToInt32(enable)))
}

// RequestLogMessages configures log message event verbosity.
func (m *Mpv) RequestLogMessages(minLevel LogLevel) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	clevel := cString(LOG_LEVEL_STRING[minLevel])
	ret := backend.requestLogMessages(m.handle, &clevel[0])
	runtime.KeepAlive(clevel)
	return newError(ret)
}

// WaitAsyncRequests blocks until async requests are completed.
func (m *Mpv) WaitAsyncRequests() {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return
	}
	backend.waitAsyncRequests(m.handle)
}

// Wakeup interrupts a blocking WaitEvent call.
func (m *Mpv) Wakeup() {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return
	}
	backend.wakeup(m.handle)
}

// WaitEvent waits for the next event or timeout.
func (m *Mpv) WaitEvent(timeout float64) *Event {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return &Event{EventId: EVENT_NONE}
	}
	ev := backend.waitEvent(m.handle, timeout)
	if ev == nil {
		return &Event{EventId: EVENT_NONE}
	}
	return &Event{
		EventId:       EventId(ev.EventID),
		Error:         newError(ev.Error),
		ReplyUserData: ev.ReplyUserData,
		Data:          ev.Data,
		cEvent:        ev,
	}
}

// Command runs a command synchronously with string arguments.
func (m *Mpv) Command(commands []string) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	if len(commands) == 0 {
		return ERROR_INVALID_PARAMETER
	}
	args, refs := marshalCmdArgs(commands)
	ret := backend.command(m.handle, (**byte)(unsafe.Pointer(&args[0])))
	runtime.KeepAlive(args)
	runtime.KeepAlive(refs)
	return newError(ret)
}

// CommandAsync queues a command and returns immediately.
func (m *Mpv) CommandAsync(replyUserdata uint64, commands []string) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	if len(commands) == 0 {
		return ERROR_INVALID_PARAMETER
	}
	args, refs := marshalCmdArgs(commands)
	ret := backend.commandAsync(m.handle, replyUserdata, (**byte)(unsafe.Pointer(&args[0])))
	runtime.KeepAlive(args)
	runtime.KeepAlive(refs)
	return newError(ret)
}

// CommandRet runs a command synchronously and returns structured result data.
func (m *Mpv) CommandRet(commands []string) (Node, error) {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return Node{}, ErrCGODisabled
	}
	if len(commands) == 0 {
		return Node{}, ERROR_INVALID_PARAMETER
	}
	args, refs := marshalCmdArgs(commands)
	var result mpvNodeRaw
	ret := backend.commandRet(m.handle, (**byte)(unsafe.Pointer(&args[0])), &result)
	runtime.KeepAlive(args)
	runtime.KeepAlive(refs)
	if err := newError(ret); err != nil {
		return Node{}, err
	}
	defer backend.freeNodeContents(&result)
	return nodeFromRaw(&result), nil
}

// CommandNode runs a structured command synchronously.
func (m *Mpv) CommandNode(command Node) (Node, error) {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return Node{}, ErrCGODisabled
	}
	arena := &nodeArena{}
	arg, err := marshalNode(command, arena)
	if err != nil {
		return Node{}, err
	}
	var result mpvNodeRaw
	ret := backend.commandNode(m.handle, arg, &result)
	runtime.KeepAlive(arena)
	if err := newError(ret); err != nil {
		return Node{}, err
	}
	defer backend.freeNodeContents(&result)
	return nodeFromRaw(&result), nil
}

// CommandNodeAsync queues a structured command and returns immediately.
func (m *Mpv) CommandNodeAsync(replyUserdata uint64, command Node) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	arena := &nodeArena{}
	arg, err := marshalNode(command, arena)
	if err != nil {
		return err
	}
	ret := backend.commandNodeAsync(m.handle, replyUserdata, arg)
	runtime.KeepAlive(arena)
	return newError(ret)
}

// CommandString runs a command string using mpv parser semantics.
func (m *Mpv) CommandString(command string) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	ccmd := cString(command)
	ret := backend.commandString(m.handle, &ccmd[0])
	runtime.KeepAlive(ccmd)
	return newError(ret)
}

// AbortAsyncCommand requests cancellation for async commands by userdata.
func (m *Mpv) AbortAsyncCommand(replyUserdata uint64) {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return
	}
	backend.abortAsyncCommand(m.handle, replyUserdata)
}

// HookAdd registers a hook handler.
func (m *Mpv) HookAdd(replyUserdata uint64, name string, priority int) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	cname := cString(name)
	ret := backend.hookAdd(m.handle, replyUserdata, &cname[0], int32(priority))
	runtime.KeepAlive(cname)
	return newError(ret)
}

// HookContinue continues a pending hook event.
func (m *Mpv) HookContinue(id uint64) error {
	if m == nil || m.handle == nil || ensureBackend() != nil {
		return ErrCGODisabled
	}
	return newError(backend.hookContinue(m.handle, id))
}
