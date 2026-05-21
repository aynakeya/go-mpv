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

static void set_render_param(mpv_render_param *params, int index, int type, void *data) {
	params[index].type = (mpv_render_param_type)type;
	params[index].data = data;
}
*/
import "C"

import (
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
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
	// keep records the original Go value for typed helpers. The cgo backend
	// rebuilds C-owned parameter data from it for each call, so reused params
	// do not point at freed C memory or store Go pointers in C memory.
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

// OpenGLDRMParams describes the deprecated DRM display parameters.
type OpenGLDRMParams struct {
	FD               int32
	CRTCID           int32
	ConnectorID      int32
	AtomicRequestPtr unsafe.Pointer
	RenderFD         int32
}

// OpenGLDRMDrawSurfaceSize describes the DRM draw surface size.
type OpenGLDRMDrawSurfaceSize struct {
	Width  int32
	Height int32
}

// OpenGLDRMParamsV2 describes the current DRM display parameters.
type OpenGLDRMParamsV2 struct {
	FD               int32
	CRTCID           int32
	ConnectorID      int32
	AtomicRequestPtr unsafe.Pointer
	RenderFD         int32
}

type OpenGLGetProcAddress func(name string) unsafe.Pointer

type OpenGLInitParams struct {
	params *C.mpv_opengl_init_params
	handle uintptr
	ctx    unsafe.Pointer
}

type RenderContext struct {
	ctx          *C.mpv_render_context
	updateHandle uintptr
	updateCtx    unsafe.Pointer
}

// NewRenderParam creates a raw render parameter. Prefer the typed helpers when
// possible; raw data is passed through as-is and must already be valid for cgo.
func NewRenderParam(paramType RenderParamType, data unsafe.Pointer) RenderParam {
	return RenderParam{Type: paramType, Data: data}
}

// RenderParamAPIType stores the Go string and allocates a temporary C string
// per call. A RenderParam can be reused safely across CreateRenderContext calls.
func RenderParamAPIType(api string) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_API_TYPE,
		keep: renderStringParam(api),
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

// RenderParamDRMDisplay creates the deprecated DRM display render parameter.
func RenderParamDRMDisplay(params *OpenGLDRMParams) RenderParam {
	if params == nil {
		return RenderParam{Type: RENDER_PARAM_DRM_DISPLAY}
	}
	return RenderParam{
		Type: RENDER_PARAM_DRM_DISPLAY,
		Data: unsafe.Pointer(params),
		keep: params,
	}
}

// RenderParamDRMDrawSurfaceSize creates the DRM draw surface size parameter.
func RenderParamDRMDrawSurfaceSize(size *OpenGLDRMDrawSurfaceSize) RenderParam {
	if size == nil {
		return RenderParam{Type: RENDER_PARAM_DRM_DRAW_SURFACE_SIZE}
	}
	return RenderParam{
		Type: RENDER_PARAM_DRM_DRAW_SURFACE_SIZE,
		Data: unsafe.Pointer(size),
		keep: size,
	}
}

// RenderParamDRMDisplayV2 creates the current DRM display render parameter.
func RenderParamDRMDisplayV2(params *OpenGLDRMParamsV2) RenderParam {
	if params == nil {
		return RenderParam{Type: RENDER_PARAM_DRM_DISPLAY_V2}
	}
	return RenderParam{
		Type: RENDER_PARAM_DRM_DISPLAY_V2,
		Data: unsafe.Pointer(params),
		keep: params,
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

// RenderParamSWFormat stores the format string in Go and converts it to a
// short-lived C string for each render call, so params can be reused safely.
func RenderParamSWFormat(format string) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_SW_FORMAT,
		keep: renderStringParam(format),
	}
}

func RenderParamSWStride(stride *uintptr) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_SW_STRIDE,
		Data: unsafe.Pointer(stride),
		keep: stride,
	}
}

// RenderParamSWPointer identifies the caller's output buffer. The cgo backend
// renders into temporary C memory and copies back on success, because libmpv
// writes through this pointer and C must not retain or receive Go memory here.
func RenderParamSWPointer(ptr unsafe.Pointer) RenderParam {
	return RenderParam{
		Type: RENDER_PARAM_SW_POINTER,
		Data: ptr,
		keep: renderPointerParam{ptr: ptr},
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
	h, cbctx, err := newRenderCallbackContext(getProc)
	if err != nil {
		return nil
	}
	params := C.create_opengl_init_params(cbctx)
	if params == nil {
		freeRenderCallbackContext(h, cbctx)
		return nil
	}
	return &OpenGLInitParams{params: params, handle: h, ctx: cbctx}
}

func (p *OpenGLInitParams) Free() {
	if p == nil || p.params == nil {
		return
	}
	C.free(unsafe.Pointer(p.params))
	p.params = nil
	freeRenderCallbackContext(p.handle, p.ctx)
	p.handle = 0
	p.ctx = nil
}

// CreateRenderContext initializes a render context for this mpv core.
// C: int mpv_render_context_create(mpv_render_context **res, mpv_handle *mpv, mpv_render_param *params);
func (m *Mpv) CreateRenderContext(params []RenderParam) (*RenderContext, error) {
	cparams, err := newCRenderParams(params)
	if err != nil {
		return nil, err
	}
	defer cparams.free()

	var ctx *C.mpv_render_context
	err = newError(C.mpv_render_context_create(&ctx, m.handle, cparams.ptr))
	runtime.KeepAlive(params)
	if err != nil {
		return nil, err
	}
	return &RenderContext{ctx: ctx}, nil
}

// SetParameter attempts to change a render context parameter.
// C: int mpv_render_context_set_parameter(mpv_render_context *ctx, mpv_render_param param);
func (r *RenderContext) SetParameter(param RenderParam) error {
	cparams, err := newCRenderParams([]RenderParam{param})
	if err != nil {
		return err
	}
	defer cparams.free()

	err = newError(C.mpv_render_context_set_parameter(r.ctx, *cparams.ptr))
	runtime.KeepAlive(param)
	return err
}

// GetInfo retrieves information from a render context.
// C: int mpv_render_context_get_info(mpv_render_context *ctx, mpv_render_param param);
func (r *RenderContext) GetInfo(param RenderParam) error {
	cparams, err := newCRenderParams([]RenderParam{param})
	if err != nil {
		return err
	}
	defer cparams.free()

	err = newError(C.mpv_render_context_get_info(r.ctx, *cparams.ptr))
	if err == nil {
		cparams.finish()
	}
	runtime.KeepAlive(param)
	return err
}

// SetUpdateCallback configures the render update callback.
// C: void mpv_render_context_set_update_callback(mpv_render_context *ctx, mpv_render_update_fn callback, void *callback_ctx);
func (r *RenderContext) SetUpdateCallback(callback func()) {
	if callback == nil {
		C.clear_render_update_callback(r.ctx)
		// The C callback context remains allocated until RenderContext.Free.
		// libmpv may have an in-flight callback with the old ctx; deleting only
		// the Go function avoids calling stale callbacks without risking a
		// use-after-free on the C context pointer.
		deleteRenderCallbackHandle(r.updateHandle)
		return
	}
	if r.updateHandle == 0 {
		handle, cbctx, err := newRenderCallbackContext(callback)
		if err != nil {
			return
		}
		r.updateHandle = handle
		r.updateCtx = cbctx
	} else {
		setRenderCallbackHandle(r.updateHandle, callback)
	}
	C.set_render_update_callback(r.ctx, r.updateCtx)
}

// Update returns render update flags.
// C: uint64_t mpv_render_context_update(mpv_render_context *ctx);
func (r *RenderContext) Update() uint64 {
	return uint64(C.mpv_render_context_update(r.ctx))
}

// Render renders a video frame to the target described by params.
// C: int mpv_render_context_render(mpv_render_context *ctx, mpv_render_param *params);
func (r *RenderContext) Render(params []RenderParam) error {
	cparams, err := newCRenderParams(params)
	if err != nil {
		return err
	}
	defer cparams.free()
	err = newError(C.mpv_render_context_render(r.ctx, cparams.ptr))
	if err == nil {
		cparams.finish()
	}
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
	freeRenderCallbackContext(r.updateHandle, r.updateCtx)
	r.updateHandle = 0
	r.updateCtx = nil
}

type renderStringParam string

type renderPointerParam struct {
	ptr unsafe.Pointer
}

// cRenderParams owns the temporary C memory used for a render call. libmpv's
// parameter array itself lives in C memory, and typed helper values are copied
// into C allocations to satisfy the cgo pointer passing rules.
type cRenderParams struct {
	ptr       *C.mpv_render_param
	count     int
	cleanups  []func()
	finishers []func()
}

func newCRenderParams(params []RenderParam) (cRenderParams, error) {
	count := len(params) + 1
	size := C.size_t(count) * C.size_t(unsafe.Sizeof(C.mpv_render_param{}))
	ptr := (*C.mpv_render_param)(C.calloc(1, size))
	if ptr == nil {
		return cRenderParams{}, ERROR_NOMEM
	}
	cparams := cRenderParams{ptr: ptr, count: count}
	swTargetBytes, err := renderSoftwareTargetBytes(params)
	if err != nil {
		cparams.free()
		return cRenderParams{}, err
	}
	for i, p := range params {
		data, cleanup, finish, err := newCRenderParamData(p, swTargetBytes)
		if err != nil {
			cparams.free()
			return cRenderParams{}, err
		}
		C.set_render_param(ptr, C.int(i), C.int(p.Type), data)
		if cleanup != nil {
			cparams.cleanups = append(cparams.cleanups, cleanup)
		}
		if finish != nil {
			cparams.finishers = append(cparams.finishers, finish)
		}
	}
	return cparams, nil
}

func (p cRenderParams) finish() {
	for _, finish := range p.finishers {
		finish()
	}
}

func (p cRenderParams) free() {
	for i := len(p.cleanups) - 1; i >= 0; i-- {
		p.cleanups[i]()
	}
	C.free(unsafe.Pointer(p.ptr))
}

func newCRenderParamData(param RenderParam, swTargetBytes int) (unsafe.Pointer, func(), func(), error) {
	switch param.Type {
	case RENDER_PARAM_API_TYPE, RENDER_PARAM_SW_FORMAT:
		if s, ok := param.keep.(renderStringParam); ok {
			return newCStringData(string(s))
		}
	case RENDER_PARAM_OPENGL_FBO:
		if fbo, ok := param.keep.(*OpenGLFBO); ok && fbo != nil {
			return newCOpenGLFBO(fbo)
		}
	case RENDER_PARAM_DRM_DISPLAY:
		if drm, ok := param.keep.(*OpenGLDRMParams); ok && drm != nil {
			return newCOpenGLDRMParams(drm)
		}
	case RENDER_PARAM_DRM_DRAW_SURFACE_SIZE:
		if size, ok := param.keep.(*OpenGLDRMDrawSurfaceSize); ok && size != nil {
			return newCOpenGLDRMDrawSurfaceSize(size)
		}
	case RENDER_PARAM_DRM_DISPLAY_V2:
		if drm, ok := param.keep.(*OpenGLDRMParamsV2); ok && drm != nil {
			return newCOpenGLDRMParamsV2(drm)
		}
	case RENDER_PARAM_FLIP_Y, RENDER_PARAM_DEPTH, RENDER_PARAM_AMBIENT_LIGHT,
		RENDER_PARAM_ADVANCED_CONTROL, RENDER_PARAM_BLOCK_FOR_TARGET_TIME,
		RENDER_PARAM_SKIP_RENDERING:
		if v, ok := param.keep.(*int32); ok && v != nil {
			return newCIntData(*v)
		}
	case RENDER_PARAM_SW_SIZE:
		if size, ok := param.keep.(*[2]int32); ok && size != nil {
			return newCIntArray2Data(size)
		}
	case RENDER_PARAM_SW_STRIDE:
		if stride, ok := param.keep.(*uintptr); ok && stride != nil {
			return newCSizeTData(*stride)
		}
	case RENDER_PARAM_SW_POINTER:
		if ptr, ok := param.keep.(renderPointerParam); ok && ptr.ptr != nil {
			return newCSoftwarePointerData(ptr.ptr, swTargetBytes)
		}
	case RENDER_PARAM_NEXT_FRAME_INFO:
		if info, ok := param.keep.(*RenderFrameInfo); ok && info != nil {
			return newCRenderFrameInfoData(info)
		}
	}
	return param.Data, nil, nil, nil
}

func newCStringData(s string) (unsafe.Pointer, func(), func(), error) {
	cstr := C.CString(s)
	if cstr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	return unsafe.Pointer(cstr), func() { C.free(unsafe.Pointer(cstr)) }, nil, nil
}

func newCIntData(v int32) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.int(0))))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	*(*C.int)(ptr) = C.int(v)
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCSizeTData(v uintptr) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.size_t(0))))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	*(*C.size_t)(ptr) = C.size_t(v)
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCIntArray2Data(v *[2]int32) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(2) * C.size_t(unsafe.Sizeof(C.int(0))))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	arr := (*[2]C.int)(ptr)
	arr[0] = C.int(v[0])
	arr[1] = C.int(v[1])
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCOpenGLFBO(fbo *OpenGLFBO) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.mpv_opengl_fbo{})))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	cfbo := (*C.mpv_opengl_fbo)(ptr)
	cfbo.fbo = C.int(fbo.FBO)
	cfbo.w = C.int(fbo.W)
	cfbo.h = C.int(fbo.H)
	cfbo.internal_format = C.int(fbo.InternalFormat)
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCOpenGLDRMParams(drm *OpenGLDRMParams) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.mpv_opengl_drm_params{})))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	cdrm := (*C.mpv_opengl_drm_params)(ptr)
	cdrm.fd = C.int(drm.FD)
	cdrm.crtc_id = C.int(drm.CRTCID)
	cdrm.connector_id = C.int(drm.ConnectorID)
	cdrm.atomic_request_ptr = (**C.struct__drmModeAtomicReq)(drm.AtomicRequestPtr)
	cdrm.render_fd = C.int(drm.RenderFD)
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCOpenGLDRMDrawSurfaceSize(size *OpenGLDRMDrawSurfaceSize) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.mpv_opengl_drm_draw_surface_size{})))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	csize := (*C.mpv_opengl_drm_draw_surface_size)(ptr)
	csize.width = C.int(size.Width)
	csize.height = C.int(size.Height)
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCOpenGLDRMParamsV2(drm *OpenGLDRMParamsV2) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.mpv_opengl_drm_params_v2{})))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	cdrm := (*C.mpv_opengl_drm_params_v2)(ptr)
	cdrm.fd = C.int(drm.FD)
	cdrm.crtc_id = C.int(drm.CRTCID)
	cdrm.connector_id = C.int(drm.ConnectorID)
	cdrm.atomic_request_ptr = (**C.struct__drmModeAtomicReq)(drm.AtomicRequestPtr)
	cdrm.render_fd = C.int(drm.RenderFD)
	return ptr, func() { C.free(ptr) }, nil, nil
}

func newCRenderFrameInfoData(info *RenderFrameInfo) (unsafe.Pointer, func(), func(), error) {
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.mpv_render_frame_info{})))
	if ptr == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	cinfo := (*C.mpv_render_frame_info)(ptr)
	cinfo.flags = C.uint64_t(info.Flags)
	cinfo.target_time = C.int64_t(info.TargetTime)
	cleanup := func() {
		C.free(ptr)
	}
	finish := func() {
		info.Flags = uint64(cinfo.flags)
		info.TargetTime = int64(cinfo.target_time)
	}
	return ptr, cleanup, finish, nil
}

func newCSoftwarePointerData(ptr unsafe.Pointer, byteLen int) (unsafe.Pointer, func(), func(), error) {
	if byteLen <= 0 {
		return nil, nil, nil, ERROR_INVALID_PARAMETER
	}
	cbuf := C.malloc(C.size_t(byteLen))
	if cbuf == nil {
		return nil, nil, nil, ERROR_NOMEM
	}
	cleanup := func() {
		C.free(cbuf)
	}
	finish := func() {
		// The software renderer writes into C memory first. Copying back only
		// after a successful render avoids both cgo pointer violations and
		// accidental corruption of the caller's buffer on render errors.
		copy(unsafeByteSlice(ptr, byteLen), unsafeByteSlice(cbuf, byteLen))
	}
	return cbuf, cleanup, finish, nil
}

// renderSoftwareTargetBytes validates enough SW render parameters to allocate
// the temporary C target buffer. libmpv requires SW_SIZE and SW_STRIDE when
// SW_POINTER is present, so rejecting incomplete helper params early avoids
// guessing the size of the caller's Go memory.
func renderSoftwareTargetBytes(params []RenderParam) (int, error) {
	var (
		height     int32
		stride     uintptr
		hasHeight  bool
		hasStride  bool
		hasPointer bool
	)
	for _, param := range params {
		switch param.Type {
		case RENDER_PARAM_SW_SIZE:
			if size, ok := param.keep.(*[2]int32); ok && size != nil {
				height = size[1]
				hasHeight = true
			}
		case RENDER_PARAM_SW_STRIDE:
			if value, ok := param.keep.(*uintptr); ok && value != nil {
				stride = *value
				hasStride = true
			}
		case RENDER_PARAM_SW_POINTER:
			if _, ok := param.keep.(renderPointerParam); ok {
				hasPointer = true
			}
		}
	}
	if !hasPointer {
		return 0, nil
	}
	if !hasHeight || !hasStride || height <= 0 {
		return 0, ERROR_INVALID_PARAMETER
	}
	byteLen := uint64(stride) * uint64(height)
	if byteLen > uint64(^uint(0)>>1) {
		return 0, ERROR_INVALID_PARAMETER
	}
	return int(byteLen), nil
}

func unsafeByteSlice(ptr unsafe.Pointer, n int) []byte {
	var b []byte
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	h.Data = uintptr(ptr)
	h.Len = n
	h.Cap = n
	return b
}

var (
	renderCallbackMu     sync.Mutex
	renderCallbackNextID uintptr
	renderCallbacks      = map[uintptr]interface{}{}
)

// newRenderCallbackContext stores the callback id in C memory instead of
// casting an integer directly to unsafe.Pointer. The race/checkptr runtime
// rejects integer-as-pointer values, and C memory is also safe for libmpv to
// hold as an opaque callback context.
func newRenderCallbackContext(value interface{}) (uintptr, unsafe.Pointer, error) {
	id := newRenderCallbackHandle(value)
	ptr := C.malloc(C.size_t(unsafe.Sizeof(C.uintptr_t(0))))
	if ptr == nil {
		deleteRenderCallbackHandle(id)
		return 0, nil, ERROR_NOMEM
	}
	*(*C.uintptr_t)(ptr) = C.uintptr_t(id)
	return id, ptr, nil
}

func freeRenderCallbackContext(id uintptr, ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
	deleteRenderCallbackHandle(id)
}

func newRenderCallbackHandle(value interface{}) uintptr {
	id := atomic.AddUintptr(&renderCallbackNextID, 1)
	setRenderCallbackHandle(id, value)
	return id
}

func setRenderCallbackHandle(id uintptr, value interface{}) {
	if id == 0 {
		return
	}
	renderCallbackMu.Lock()
	renderCallbacks[id] = value
	renderCallbackMu.Unlock()
}

func deleteRenderCallbackHandle(id uintptr) {
	if id == 0 {
		return
	}
	renderCallbackMu.Lock()
	delete(renderCallbacks, id)
	renderCallbackMu.Unlock()
}

func renderCallbackValue(id uintptr) interface{} {
	renderCallbackMu.Lock()
	value := renderCallbacks[id]
	renderCallbackMu.Unlock()
	return value
}

//export goMpvRenderUpdateCallback
func goMpvRenderUpdateCallback(ctx unsafe.Pointer) {
	if ctx == nil {
		return
	}
	id := uintptr(*(*C.uintptr_t)(ctx))
	callback, ok := renderCallbackValue(id).(func())
	if ok && callback != nil {
		callback()
	}
}

//export goMpvOpenGLGetProcAddress
func goMpvOpenGLGetProcAddress(ctx unsafe.Pointer, name *C.char) unsafe.Pointer {
	if ctx == nil {
		return nil
	}
	id := uintptr(*(*C.uintptr_t)(ctx))
	callback, ok := renderCallbackValue(id).(OpenGLGetProcAddress)
	if !ok || callback == nil {
		return nil
	}
	return callback(C.GoString(name))
}
