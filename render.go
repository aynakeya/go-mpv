//go:build cgo
// +build cgo

package mpv

/*
#include <mpv/client.h>
#include <mpv/render.h>
#include <mpv/render_gl.h>
#include <stdlib.h>
#include <stdint.h>

extern void goMpvRenderUpdateCallback(void *ctx);
extern void *goMpvOpenGLGetProcAddress(void *ctx, char *name);

static void set_render_update_callback(mpv_render_context *ctx, void *cb_ctx) {
	mpv_render_context_set_update_callback(ctx, goMpvRenderUpdateCallback, cb_ctx);
}

static void clear_render_update_callback(mpv_render_context *ctx) {
	mpv_render_context_set_update_callback(ctx, NULL, NULL);
}

static void *opengl_get_proc_address_bridge(void *ctx, const char *name) {
	return goMpvOpenGLGetProcAddress(ctx, (char *)name);
}

static mpv_opengl_init_params *create_opengl_init_params(void *ctx) {
	mpv_opengl_init_params *params = malloc(sizeof(*params));
	if (!params) {
		return NULL;
	}
	params->get_proc_address = opengl_get_proc_address_bridge;
	params->get_proc_address_ctx = ctx;
	return params;
}
*/
import "C"

import (
	"runtime"
	"runtime/cgo"
	"unsafe"
)

type RenderParamType int

const (
	RENDER_PARAM_INVALID               RenderParamType = C.MPV_RENDER_PARAM_INVALID
	RENDER_PARAM_API_TYPE              RenderParamType = C.MPV_RENDER_PARAM_API_TYPE
	RENDER_PARAM_OPENGL_INIT_PARAMS    RenderParamType = C.MPV_RENDER_PARAM_OPENGL_INIT_PARAMS
	RENDER_PARAM_OPENGL_FBO            RenderParamType = C.MPV_RENDER_PARAM_OPENGL_FBO
	RENDER_PARAM_FLIP_Y                RenderParamType = C.MPV_RENDER_PARAM_FLIP_Y
	RENDER_PARAM_DEPTH                 RenderParamType = C.MPV_RENDER_PARAM_DEPTH
	RENDER_PARAM_ICC_PROFILE           RenderParamType = C.MPV_RENDER_PARAM_ICC_PROFILE
	RENDER_PARAM_AMBIENT_LIGHT         RenderParamType = C.MPV_RENDER_PARAM_AMBIENT_LIGHT
	RENDER_PARAM_X11_DISPLAY           RenderParamType = C.MPV_RENDER_PARAM_X11_DISPLAY
	RENDER_PARAM_WL_DISPLAY            RenderParamType = C.MPV_RENDER_PARAM_WL_DISPLAY
	RENDER_PARAM_ADVANCED_CONTROL      RenderParamType = C.MPV_RENDER_PARAM_ADVANCED_CONTROL
	RENDER_PARAM_NEXT_FRAME_INFO       RenderParamType = C.MPV_RENDER_PARAM_NEXT_FRAME_INFO
	RENDER_PARAM_BLOCK_FOR_TARGET_TIME RenderParamType = C.MPV_RENDER_PARAM_BLOCK_FOR_TARGET_TIME
	RENDER_PARAM_SKIP_RENDERING        RenderParamType = C.MPV_RENDER_PARAM_SKIP_RENDERING
	RENDER_PARAM_DRM_DISPLAY           RenderParamType = C.MPV_RENDER_PARAM_DRM_DISPLAY
	RENDER_PARAM_DRM_DRAW_SURFACE_SIZE RenderParamType = C.MPV_RENDER_PARAM_DRM_DRAW_SURFACE_SIZE
	RENDER_PARAM_DRM_DISPLAY_V2        RenderParamType = C.MPV_RENDER_PARAM_DRM_DISPLAY_V2
	RENDER_PARAM_SW_SIZE               RenderParamType = C.MPV_RENDER_PARAM_SW_SIZE
	RENDER_PARAM_SW_FORMAT             RenderParamType = C.MPV_RENDER_PARAM_SW_FORMAT
	RENDER_PARAM_SW_STRIDE             RenderParamType = C.MPV_RENDER_PARAM_SW_STRIDE
	RENDER_PARAM_SW_POINTER            RenderParamType = C.MPV_RENDER_PARAM_SW_POINTER
)

type RenderFrameInfoFlag uint64

const (
	RENDER_FRAME_INFO_PRESENT     RenderFrameInfoFlag = C.MPV_RENDER_FRAME_INFO_PRESENT
	RENDER_FRAME_INFO_REDRAW      RenderFrameInfoFlag = C.MPV_RENDER_FRAME_INFO_REDRAW
	RENDER_FRAME_INFO_REPEAT      RenderFrameInfoFlag = C.MPV_RENDER_FRAME_INFO_REPEAT
	RENDER_FRAME_INFO_BLOCK_VSYNC RenderFrameInfoFlag = C.MPV_RENDER_FRAME_INFO_BLOCK_VSYNC
)

type RenderUpdateFlag uint64

const (
	RENDER_UPDATE_FRAME RenderUpdateFlag = C.MPV_RENDER_UPDATE_FRAME
)

const (
	RENDER_API_TYPE_OPENGL = C.MPV_RENDER_API_TYPE_OPENGL
	RENDER_API_TYPE_SW     = C.MPV_RENDER_API_TYPE_SW
)

type RenderParam struct {
	Type RenderParamType
	Data unsafe.Pointer
	keep interface{}
}

type RenderFrameInfo struct {
	Flags      uint64
	TargetTime int64
}

type OpenGLFBO struct {
	FBO            int32
	W              int32
	H              int32
	InternalFormat int32
}

type OpenGLGetProcAddress func(name string) unsafe.Pointer

type OpenGLInitParams struct {
	params *C.mpv_opengl_init_params
	handle cgo.Handle
}

type RenderContext struct {
	ctx          *C.mpv_render_context
	updateHandle cgo.Handle
}

func NewRenderParam(paramType RenderParamType, data unsafe.Pointer) RenderParam {
	return RenderParam{Type: paramType, Data: data}
}

func RenderParamAPIType(api string) RenderParam {
	capi := C.CString(api)
	return RenderParam{
		Type: RENDER_PARAM_API_TYPE,
		Data: unsafe.Pointer(capi),
		keep: cStringResource{ptr: unsafe.Pointer(capi)},
	}
}

func RenderParamOpenGLInitParams(params *OpenGLInitParams) RenderParam {
	if params == nil {
		return RenderParam{Type: RENDER_PARAM_OPENGL_INIT_PARAMS}
	}
	return RenderParam{
		Type: RENDER_PARAM_OPENGL_INIT_PARAMS,
		Data: unsafe.Pointer(params.params),
		keep: params,
	}
}

func RenderParamOpenGLFBO(fbo *OpenGLFBO) RenderParam {
	if fbo == nil {
		return RenderParam{Type: RENDER_PARAM_OPENGL_FBO}
	}
	return RenderParam{
		Type: RENDER_PARAM_OPENGL_FBO,
		Data: unsafe.Pointer(fbo),
		keep: fbo,
	}
}

func RenderParamInt(paramType RenderParamType, value *int32) RenderParam {
	return RenderParam{
		Type: paramType,
		Data: unsafe.Pointer(value),
		keep: value,
	}
}

func RenderParamSWSize(size *[2]int32) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_SW_SIZE,
		Data: unsafe.Pointer(size),
		keep: size,
	}
}

func RenderParamSWFormat(format string) RenderParam {
	cformat := C.CString(format)
	return RenderParam{
		Type: RENDER_PARAM_SW_FORMAT,
		Data: unsafe.Pointer(cformat),
		keep: cStringResource{ptr: unsafe.Pointer(cformat)},
	}
}

func RenderParamSWStride(stride *uintptr) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_SW_STRIDE,
		Data: unsafe.Pointer(stride),
		keep: stride,
	}
}

func RenderParamSWPointer(ptr unsafe.Pointer) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_SW_POINTER,
		Data: ptr,
	}
}

func RenderParamNextFrameInfo(info *RenderFrameInfo) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_NEXT_FRAME_INFO,
		Data: unsafe.Pointer(info),
		keep: info,
	}
}

func NewOpenGLInitParams(getProc OpenGLGetProcAddress) *OpenGLInitParams {
	if getProc == nil {
		return nil
	}
	h := cgo.NewHandle(getProc)
	params := C.create_opengl_init_params(unsafe.Pointer(uintptr(h)))
	if params == nil {
		h.Delete()
		return nil
	}
	return &OpenGLInitParams{params: params, handle: h}
}

func (p *OpenGLInitParams) Free() {
	if p == nil || p.params == nil {
		return
	}
	C.free(unsafe.Pointer(p.params))
	p.params = nil
	p.handle.Delete()
}

// CreateRenderContext initializes a render context for this mpv core.
// C: int mpv_render_context_create(mpv_render_context **res, mpv_handle *mpv, mpv_render_param *params);
func (m *Mpv) CreateRenderContext(params []RenderParam) (*RenderContext, error) {
	cparams := newCRenderParams(params)
	defer cparams.free()

	var ctx *C.mpv_render_context
	err := newError(C.mpv_render_context_create(&ctx, m.handle, cparams.ptr))
	runtime.KeepAlive(params)
	if err != nil {
		return nil, err
	}
	return &RenderContext{ctx: ctx}, nil
}

// SetParameter attempts to change a render context parameter.
// C: int mpv_render_context_set_parameter(mpv_render_context *ctx, mpv_render_param param);
func (r *RenderContext) SetParameter(param RenderParam) error {
	cparam := C.mpv_render_param{
		_type: uint32(param.Type),
		data:  param.Data,
	}
	err := newError(C.mpv_render_context_set_parameter(r.ctx, cparam))
	runtime.KeepAlive(param)
	return err
}

// GetInfo retrieves information from a render context.
// C: int mpv_render_context_get_info(mpv_render_context *ctx, mpv_render_param param);
func (r *RenderContext) GetInfo(param RenderParam) error {
	cparam := C.mpv_render_param{
		_type: uint32(param.Type),
		data:  param.Data,
	}
	err := newError(C.mpv_render_context_get_info(r.ctx, cparam))
	runtime.KeepAlive(param)
	return err
}

// SetUpdateCallback configures the render update callback.
// C: void mpv_render_context_set_update_callback(mpv_render_context *ctx, mpv_render_update_fn callback, void *callback_ctx);
func (r *RenderContext) SetUpdateCallback(callback func()) {
	if r.updateHandle != 0 {
		C.clear_render_update_callback(r.ctx)
		r.updateHandle.Delete()
		r.updateHandle = 0
	}
	if callback == nil {
		return
	}
	r.updateHandle = cgo.NewHandle(callback)
	C.set_render_update_callback(r.ctx, unsafe.Pointer(uintptr(r.updateHandle)))
}

// Update returns render update flags.
// C: uint64_t mpv_render_context_update(mpv_render_context *ctx);
func (r *RenderContext) Update() uint64 {
	return uint64(C.mpv_render_context_update(r.ctx))
}

// Render renders a video frame to the target described by params.
// C: int mpv_render_context_render(mpv_render_context *ctx, mpv_render_param *params);
func (r *RenderContext) Render(params []RenderParam) error {
	cparams := newCRenderParams(params)
	defer cparams.free()
	err := newError(C.mpv_render_context_render(r.ctx, cparams.ptr))
	runtime.KeepAlive(params)
	return err
}

// ReportSwap reports that a frame was presented.
// C: void mpv_render_context_report_swap(mpv_render_context *ctx);
func (r *RenderContext) ReportSwap() {
	C.mpv_render_context_report_swap(r.ctx)
}

// Free destroys the render context.
// C: void mpv_render_context_free(mpv_render_context *ctx);
func (r *RenderContext) Free() {
	if r == nil || r.ctx == nil {
		return
	}
	r.SetUpdateCallback(nil)
	C.mpv_render_context_free(r.ctx)
	r.ctx = nil
}

type cStringResource struct {
	ptr unsafe.Pointer
}

func (r cStringResource) free() {
	if r.ptr != nil {
		C.free(r.ptr)
	}
}

type cRenderParams struct {
	ptr    *C.mpv_render_param
	count  int
	params []RenderParam
}

func newCRenderParams(params []RenderParam) cRenderParams {
	count := len(params) + 1
	size := C.size_t(count) * C.size_t(unsafe.Sizeof(C.mpv_render_param{}))
	ptr := (*C.mpv_render_param)(C.calloc(1, size))
	step := unsafe.Sizeof(C.mpv_render_param{})
	base := uintptr(unsafe.Pointer(ptr))
	for i, p := range params {
		cp := (*C.mpv_render_param)(unsafe.Pointer(base + uintptr(i)*step))
		cp._type = uint32(p.Type)
		cp.data = p.Data
	}
	return cRenderParams{ptr: ptr, count: count, params: params}
}

func (p cRenderParams) free() {
	for _, param := range p.params {
		if res, ok := param.keep.(cStringResource); ok {
			res.free()
		}
	}
	C.free(unsafe.Pointer(p.ptr))
}

//export goMpvRenderUpdateCallback
func goMpvRenderUpdateCallback(ctx unsafe.Pointer) {
	if ctx == nil {
		return
	}
	handle := cgo.Handle(uintptr(ctx))
	callback, ok := handle.Value().(func())
	if ok && callback != nil {
		callback()
	}
}

//export goMpvOpenGLGetProcAddress
func goMpvOpenGLGetProcAddress(ctx unsafe.Pointer, name *C.char) unsafe.Pointer {
	if ctx == nil {
		return nil
	}
	handle := cgo.Handle(uintptr(ctx))
	callback, ok := handle.Value().(OpenGLGetProcAddress)
	if !ok || callback == nil {
		return nil
	}
	return callback(C.GoString(name))
}
