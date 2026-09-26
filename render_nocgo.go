//go:build !cgo
// +build !cgo

package mpv

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	puregoRenderUpdateCallbackPtr         = purego.NewCallback(puregoRenderUpdateCallback)
	puregoOpenGLGetProcAddressCallbackPtr = purego.NewCallback(puregoOpenGLGetProcAddress)
)

type RenderParamType int

const (
	RENDER_PARAM_INVALID               RenderParamType = 0
	RENDER_PARAM_API_TYPE              RenderParamType = 1
	RENDER_PARAM_OPENGL_INIT_PARAMS    RenderParamType = 2
	RENDER_PARAM_OPENGL_FBO            RenderParamType = 3
	RENDER_PARAM_FLIP_Y                RenderParamType = 4
	RENDER_PARAM_DEPTH                 RenderParamType = 5
	RENDER_PARAM_ICC_PROFILE           RenderParamType = 6
	RENDER_PARAM_AMBIENT_LIGHT         RenderParamType = 7
	RENDER_PARAM_X11_DISPLAY           RenderParamType = 8
	RENDER_PARAM_WL_DISPLAY            RenderParamType = 9
	RENDER_PARAM_ADVANCED_CONTROL      RenderParamType = 10
	RENDER_PARAM_NEXT_FRAME_INFO       RenderParamType = 11
	RENDER_PARAM_BLOCK_FOR_TARGET_TIME RenderParamType = 12
	RENDER_PARAM_SKIP_RENDERING        RenderParamType = 13
	RENDER_PARAM_DRM_DISPLAY           RenderParamType = 14
	RENDER_PARAM_DRM_DRAW_SURFACE_SIZE RenderParamType = 15
	RENDER_PARAM_DRM_DISPLAY_V2        RenderParamType = 16
	RENDER_PARAM_SW_SIZE               RenderParamType = 17
	RENDER_PARAM_SW_FORMAT             RenderParamType = 18
	RENDER_PARAM_SW_STRIDE             RenderParamType = 19
	RENDER_PARAM_SW_POINTER            RenderParamType = 20
)

type RenderFrameInfoFlag uint64

const (
	RENDER_FRAME_INFO_PRESENT     RenderFrameInfoFlag = 1 << 0
	RENDER_FRAME_INFO_REDRAW      RenderFrameInfoFlag = 1 << 1
	RENDER_FRAME_INFO_REPEAT      RenderFrameInfoFlag = 1 << 2
	RENDER_FRAME_INFO_BLOCK_VSYNC RenderFrameInfoFlag = 1 << 3
)

type RenderUpdateFlag uint64

const (
	RENDER_UPDATE_FRAME RenderUpdateFlag = 1 << 0
)

const (
	RENDER_API_TYPE_OPENGL = "opengl"
	RENDER_API_TYPE_SW     = "sw"
)

type OpenGLInitParams struct {
	params mpvOpenGLInitParamsRaw
	handle uintptr
	ctx    *uintptr
}

type RenderContext struct {
	ctx          unsafe.Pointer
	updateHandle uintptr
	updateCtx    *uintptr
}

type mpvRenderParamRaw struct {
	Type int32
	Data unsafe.Pointer
}

type mpvOpenGLInitParamsRaw struct {
	GetProcAddress    uintptr
	GetProcAddressCtx unsafe.Pointer
}

func RenderParamOpenGLInitParams(params *OpenGLInitParams) RenderParam {
	if params == nil {
		return RenderParam{Type: RENDER_PARAM_OPENGL_INIT_PARAMS}
	}
	return RenderParam{
		Type: RENDER_PARAM_OPENGL_INIT_PARAMS,
		Data: unsafe.Pointer(&params.params),
		keep: params,
	}
}

func NewOpenGLInitParams(getProc OpenGLGetProcAddress) *OpenGLInitParams {
	if getProc == nil {
		return nil
	}
	handle := newPuregoRenderCallbackHandle(getProc)
	params := &OpenGLInitParams{handle: handle, ctx: new(uintptr)}
	*params.ctx = handle
	params.params.GetProcAddress = puregoOpenGLGetProcAddressCallbackPtr
	params.params.GetProcAddressCtx = unsafe.Pointer(params.ctx)
	return params
}

func (p *OpenGLInitParams) Free() {
	if p == nil || p.handle == 0 {
		return
	}
	deletePuregoRenderCallbackHandle(p.handle)
	p.handle = 0
	p.ctx = nil
	p.params = mpvOpenGLInitParamsRaw{}
}

func (m *Mpv) CreateRenderContext(params []RenderParam) (*RenderContext, error) {
	if m.handle == nil {
		return nil, ERROR_INVALID_PARAMETER
	}
	raw, keep := newPuregoRenderParams(params)
	var ctx unsafe.Pointer
	err := newError(backend.renderContextCreate(&ctx, m.handle, &raw[0]))
	runtime.KeepAlive(params)
	runtime.KeepAlive(keep)
	if err != nil {
		return nil, err
	}
	return &RenderContext{ctx: ctx}, nil
}

func (r *RenderContext) SetParameter(param RenderParam) error {
	if r == nil || r.ctx == nil {
		return ERROR_INVALID_PARAMETER
	}
	data, keep := puregoRenderParamData(param)
	err := newError(backend.renderContextSetParameter(r.ctx, uintptr(uint32(param.Type)), data))
	runtime.KeepAlive(param)
	runtime.KeepAlive(keep)
	return err
}

func (r *RenderContext) GetInfo(param RenderParam) error {
	if r == nil || r.ctx == nil {
		return ERROR_INVALID_PARAMETER
	}
	data, keep := puregoRenderParamData(param)
	err := newError(backend.renderContextGetInfo(r.ctx, uintptr(uint32(param.Type)), data))
	if err == nil && keep.finish != nil {
		keep.finish()
	}
	runtime.KeepAlive(param)
	runtime.KeepAlive(keep)
	return err
}

func (r *RenderContext) SetUpdateCallback(callback func()) {
	if r == nil || r.ctx == nil {
		return
	}
	if callback == nil {
		backend.renderContextSetUpdateCallback(r.ctx, 0, nil)
		deletePuregoRenderCallbackHandle(r.updateHandle)
		return
	}
	if r.updateHandle == 0 {
		r.updateHandle = newPuregoRenderCallbackHandle(callback)
		r.updateCtx = new(uintptr)
		*r.updateCtx = r.updateHandle
	} else {
		setPuregoRenderCallbackHandle(r.updateHandle, callback)
	}
	backend.renderContextSetUpdateCallback(r.ctx, puregoRenderUpdateCallbackPtr, unsafe.Pointer(r.updateCtx))
}

func (r *RenderContext) Update() uint64 {
	if r == nil || r.ctx == nil {
		return 0
	}
	return backend.renderContextUpdate(r.ctx)
}

func (r *RenderContext) Render(params []RenderParam) error {
	if r == nil || r.ctx == nil {
		return ERROR_INVALID_PARAMETER
	}
	raw, keep := newPuregoRenderParams(params)
	err := newError(backend.renderContextRender(r.ctx, &raw[0]))
	if err == nil {
		finishPuregoRenderParams(keep)
	}
	runtime.KeepAlive(params)
	runtime.KeepAlive(keep)
	return err
}

func (r *RenderContext) ReportSwap() {
	if r == nil || r.ctx == nil {
		return
	}
	backend.renderContextReportSwap(r.ctx)
}

func (r *RenderContext) Free() {
	if r == nil || r.ctx == nil {
		return
	}
	r.SetUpdateCallback(nil)
	backend.renderContextFree(r.ctx)
	r.ctx = nil
	deletePuregoRenderCallbackHandle(r.updateHandle)
	r.updateHandle = 0
	r.updateCtx = nil
}

type puregoRenderParamKeep struct {
	value  interface{}
	finish func()
}

func newPuregoRenderParams(params []RenderParam) ([]mpvRenderParamRaw, []puregoRenderParamKeep) {
	raw := make([]mpvRenderParamRaw, len(params)+1)
	keep := make([]puregoRenderParamKeep, 0, len(params))
	for i, param := range params {
		data, k := puregoRenderParamData(param)
		raw[i] = mpvRenderParamRaw{Type: int32(param.Type), Data: data}
		if k.value != nil || k.finish != nil {
			keep = append(keep, k)
		}
	}
	return raw, keep
}

func finishPuregoRenderParams(keep []puregoRenderParamKeep) {
	for _, k := range keep {
		if k.finish != nil {
			k.finish()
		}
	}
}

func puregoRenderParamData(param RenderParam) (unsafe.Pointer, puregoRenderParamKeep) {
	switch param.Type {
	case RENDER_PARAM_API_TYPE, RENDER_PARAM_SW_FORMAT:
		if s, ok := param.keep.(renderStringParam); ok {
			b := cString(string(s))
			return unsafe.Pointer(&b[0]), puregoRenderParamKeep{value: b}
		}
	case RENDER_PARAM_NEXT_FRAME_INFO:
		if info, ok := param.keep.(*RenderFrameInfo); ok && info != nil {
			// GetInfo must not overwrite the caller's value on failure.
			raw := *info
			return unsafe.Pointer(&raw), puregoRenderParamKeep{
				value: &raw,
				finish: func() {
					*info = raw
				},
			}
		}
	}
	// Typed inputs already have the C ABI layout. Callers keep param alive
	// across the FFI call, so these values need no intermediate allocation.
	return param.Data, puregoRenderParamKeep{}
}

var (
	puregoRenderCallbackMu     sync.Mutex
	puregoRenderCallbackNextID uintptr
	puregoRenderCallbacks      = map[uintptr]interface{}{}
)

func newPuregoRenderCallbackHandle(value interface{}) uintptr {
	id := atomic.AddUintptr(&puregoRenderCallbackNextID, 1)
	setPuregoRenderCallbackHandle(id, value)
	return id
}

func setPuregoRenderCallbackHandle(id uintptr, value interface{}) {
	if id == 0 {
		return
	}
	puregoRenderCallbackMu.Lock()
	puregoRenderCallbacks[id] = value
	puregoRenderCallbackMu.Unlock()
}

func deletePuregoRenderCallbackHandle(id uintptr) {
	if id == 0 {
		return
	}
	puregoRenderCallbackMu.Lock()
	delete(puregoRenderCallbacks, id)
	puregoRenderCallbackMu.Unlock()
}

func puregoRenderCallbackValue(id uintptr) interface{} {
	puregoRenderCallbackMu.Lock()
	value := puregoRenderCallbacks[id]
	puregoRenderCallbackMu.Unlock()
	return value
}

func puregoRenderUpdateCallback(ctx unsafe.Pointer) {
	if ctx == nil {
		return
	}
	callback, ok := puregoRenderCallbackValue(*(*uintptr)(ctx)).(func())
	if ok && callback != nil {
		callback()
	}
}

func puregoOpenGLGetProcAddress(ctx unsafe.Pointer, name *byte) unsafe.Pointer {
	if ctx == nil {
		return nil
	}
	callback, ok := puregoRenderCallbackValue(*(*uintptr)(ctx)).(OpenGLGetProcAddress)
	if !ok || callback == nil {
		return nil
	}
	return callback(goString(name))
}
