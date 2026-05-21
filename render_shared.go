package mpv

import "unsafe"

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

type RenderParam struct {
	Type RenderParamType
	Data unsafe.Pointer
	// keep records the original Go value for typed helpers. The cgo backend
	// rebuilds C-owned parameter data from it for each call, so reused params
	// do not point at freed C memory or store Go pointers in C memory.
	keep interface{}
}

type renderStringParam string

type renderPointerParam struct {
	ptr unsafe.Pointer
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
