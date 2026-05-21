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

type mpvRenderFrameInfoRaw struct {
	Flags      uint64
	TargetTime int64
}

type mpvOpenGLFboRaw struct {
	FBO            int32
	W              int32
	H              int32
	InternalFormat int32
}

type mpvOpenGLDRMParamsRaw struct {
	FD               int32
	CRTCID           int32
	ConnectorID      int32
	AtomicRequestPtr unsafe.Pointer
	RenderFD         int32
}

type mpvOpenGLDRMDrawSurfaceSizeRaw struct {
	Width  int32
	Height int32
}

type mpvOpenGLDRMParamsV2Raw struct {
	FD               int32
	CRTCID           int32
	ConnectorID      int32
	AtomicRequestPtr unsafe.Pointer
	RenderFD         int32
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
	if m == nil || m.handle == nil {
		return nil, ErrCGODisabled
	}
	if err := ensureBackend(); err != nil {
		return nil, err
	}
	raw, keep, err := newPuregoRenderParams(params)
	if err != nil {
		return nil, err
	}
	var ctx unsafe.Pointer
	err = newError(backend.renderContextCreate(&ctx, m.handle, &raw[0]))
	runtime.KeepAlive(params)
	runtime.KeepAlive(keep)
	if err != nil {
		return nil, err
	}
	return &RenderContext{ctx: ctx}, nil
}

func (r *RenderContext) SetParameter(param RenderParam) error {
	if r == nil || r.ctx == nil {
		return ErrCGODisabled
	}
	if err := ensureBackend(); err != nil {
		return err
	}
	raw, keep, err := newPuregoRenderParams([]RenderParam{param})
	if err != nil {
		return err
	}
	err = newError(backend.renderContextSetParameter(r.ctx, uintptr(uint32(raw[0].Type)), raw[0].Data))
	runtime.KeepAlive(param)
	runtime.KeepAlive(keep)
	return err
}

func (r *RenderContext) GetInfo(param RenderParam) error {
	if r == nil || r.ctx == nil {
		return ErrCGODisabled
	}
	if err := ensureBackend(); err != nil {
		return err
	}
	raw, keep, err := newPuregoRenderParams([]RenderParam{param})
	if err != nil {
		return err
	}
	err = newError(backend.renderContextGetInfo(r.ctx, uintptr(uint32(raw[0].Type)), raw[0].Data))
	if err == nil {
		finishPuregoRenderParams(keep)
	}
	runtime.KeepAlive(param)
	runtime.KeepAlive(keep)
	return err
}

func (r *RenderContext) SetUpdateCallback(callback func()) {
	if r == nil || r.ctx == nil || ensureBackend() != nil {
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
	if r == nil || r.ctx == nil || ensureBackend() != nil {
		return 0
	}
	return backend.renderContextUpdate(r.ctx)
}

func (r *RenderContext) Render(params []RenderParam) error {
	if r == nil || r.ctx == nil {
		return ErrCGODisabled
	}
	if err := ensureBackend(); err != nil {
		return err
	}
	raw, keep, err := newPuregoRenderParams(params)
	if err != nil {
		return err
	}
	err = newError(backend.renderContextRender(r.ctx, &raw[0]))
	if err == nil {
		finishPuregoRenderParams(keep)
	}
	runtime.KeepAlive(params)
	runtime.KeepAlive(keep)
	return err
}

func (r *RenderContext) ReportSwap() {
	if r == nil || r.ctx == nil || ensureBackend() != nil {
		return
	}
	backend.renderContextReportSwap(r.ctx)
}

func (r *RenderContext) Free() {
	if r == nil || r.ctx == nil || ensureBackend() != nil {
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

func newPuregoRenderParams(params []RenderParam) ([]mpvRenderParamRaw, []puregoRenderParamKeep, error) {
	raw := make([]mpvRenderParamRaw, len(params)+1)
	keep := make([]puregoRenderParamKeep, 0, len(params))
	for i, param := range params {
		data, k, err := puregoRenderParamData(param)
		if err != nil {
			return nil, nil, err
		}
		raw[i] = mpvRenderParamRaw{Type: int32(param.Type), Data: data}
		if k.value != nil || k.finish != nil {
			keep = append(keep, k)
		}
	}
	return raw, keep, nil
}

func finishPuregoRenderParams(keep []puregoRenderParamKeep) {
	for _, k := range keep {
		if k.finish != nil {
			k.finish()
		}
	}
}

func puregoRenderParamData(param RenderParam) (unsafe.Pointer, puregoRenderParamKeep, error) {
	switch param.Type {
	case RENDER_PARAM_API_TYPE, RENDER_PARAM_SW_FORMAT:
		if s, ok := param.keep.(renderStringParam); ok {
			b := cString(string(s))
			return unsafe.Pointer(&b[0]), puregoRenderParamKeep{value: b}, nil
		}
	case RENDER_PARAM_OPENGL_FBO:
		if fbo, ok := param.keep.(*OpenGLFBO); ok && fbo != nil {
			raw := &mpvOpenGLFboRaw{
				FBO:            fbo.FBO,
				W:              fbo.W,
				H:              fbo.H,
				InternalFormat: fbo.InternalFormat,
			}
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_DRM_DISPLAY:
		if drm, ok := param.keep.(*OpenGLDRMParams); ok && drm != nil {
			raw := &mpvOpenGLDRMParamsRaw{
				FD:               drm.FD,
				CRTCID:           drm.CRTCID,
				ConnectorID:      drm.ConnectorID,
				AtomicRequestPtr: drm.AtomicRequestPtr,
				RenderFD:         drm.RenderFD,
			}
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_DRM_DRAW_SURFACE_SIZE:
		if size, ok := param.keep.(*OpenGLDRMDrawSurfaceSize); ok && size != nil {
			raw := &mpvOpenGLDRMDrawSurfaceSizeRaw{Width: size.Width, Height: size.Height}
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_DRM_DISPLAY_V2:
		if drm, ok := param.keep.(*OpenGLDRMParamsV2); ok && drm != nil {
			raw := &mpvOpenGLDRMParamsV2Raw{
				FD:               drm.FD,
				CRTCID:           drm.CRTCID,
				ConnectorID:      drm.ConnectorID,
				AtomicRequestPtr: drm.AtomicRequestPtr,
				RenderFD:         drm.RenderFD,
			}
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_FLIP_Y, RENDER_PARAM_DEPTH, RENDER_PARAM_AMBIENT_LIGHT,
		RENDER_PARAM_ADVANCED_CONTROL, RENDER_PARAM_BLOCK_FOR_TARGET_TIME,
		RENDER_PARAM_SKIP_RENDERING:
		if v, ok := param.keep.(*int32); ok && v != nil {
			raw := new(int32)
			*raw = *v
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_SW_SIZE:
		if size, ok := param.keep.(*[2]int32); ok && size != nil {
			raw := &[2]int32{size[0], size[1]}
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_SW_STRIDE:
		if stride, ok := param.keep.(*uintptr); ok && stride != nil {
			raw := new(uintptr)
			*raw = *stride
			return unsafe.Pointer(raw), puregoRenderParamKeep{value: raw}, nil
		}
	case RENDER_PARAM_SW_POINTER:
		if ptr, ok := param.keep.(renderPointerParam); ok {
			return ptr.ptr, puregoRenderParamKeep{}, nil
		}
	case RENDER_PARAM_NEXT_FRAME_INFO:
		if info, ok := param.keep.(*RenderFrameInfo); ok && info != nil {
			raw := &mpvRenderFrameInfoRaw{Flags: info.Flags, TargetTime: info.TargetTime}
			return unsafe.Pointer(raw), puregoRenderParamKeep{
				value: raw,
				finish: func() {
					info.Flags = raw.Flags
					info.TargetTime = raw.TargetTime
				},
			}, nil
		}
	case RENDER_PARAM_OPENGL_INIT_PARAMS:
		if params, ok := param.keep.(*OpenGLInitParams); ok && params != nil {
			return unsafe.Pointer(&params.params), puregoRenderParamKeep{value: params}, nil
		}
	}
	return param.Data, puregoRenderParamKeep{}, nil
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
