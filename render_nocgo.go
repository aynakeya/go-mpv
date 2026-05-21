//go:build !cgo
// +build !cgo

package mpv

import "unsafe"

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

type RenderParam struct {
	Type RenderParamType
	Data unsafe.Pointer
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

type OpenGLInitParams struct{}

type RenderContext struct{}

func NewRenderParam(paramType RenderParamType, data unsafe.Pointer) RenderParam {
	return RenderParam{Type: paramType, Data: data}
}

func RenderParamAPIType(api string) RenderParam {
	return RenderParam{Type: RENDER_PARAM_API_TYPE}
}

func RenderParamOpenGLInitParams(params *OpenGLInitParams) RenderParam {
	return RenderParam{Type: RENDER_PARAM_OPENGL_INIT_PARAMS, Data: unsafe.Pointer(params)}
}

func RenderParamOpenGLFBO(fbo *OpenGLFBO) RenderParam {
	return RenderParam{Type: RENDER_PARAM_OPENGL_FBO, Data: unsafe.Pointer(fbo)}
}

// RenderParamDRMDisplay creates the deprecated DRM display render parameter.
func RenderParamDRMDisplay(params *OpenGLDRMParams) RenderParam {
	return RenderParam{Type: RENDER_PARAM_DRM_DISPLAY, Data: unsafe.Pointer(params)}
}

// RenderParamDRMDrawSurfaceSize creates the DRM draw surface size parameter.
func RenderParamDRMDrawSurfaceSize(size *OpenGLDRMDrawSurfaceSize) RenderParam {
	return RenderParam{Type: RENDER_PARAM_DRM_DRAW_SURFACE_SIZE, Data: unsafe.Pointer(size)}
}

// RenderParamDRMDisplayV2 creates the current DRM display render parameter.
func RenderParamDRMDisplayV2(params *OpenGLDRMParamsV2) RenderParam {
	return RenderParam{Type: RENDER_PARAM_DRM_DISPLAY_V2, Data: unsafe.Pointer(params)}
}

func RenderParamInt(paramType RenderParamType, value *int32) RenderParam {
	return RenderParam{Type: paramType, Data: unsafe.Pointer(value)}
}

func RenderParamSWSize(size *[2]int32) RenderParam {
	return RenderParam{Type: RENDER_PARAM_SW_SIZE, Data: unsafe.Pointer(size)}
}

func RenderParamSWFormat(format string) RenderParam {
	return RenderParam{Type: RENDER_PARAM_SW_FORMAT}
}

func RenderParamSWStride(stride *uintptr) RenderParam {
	return RenderParam{Type: RENDER_PARAM_SW_STRIDE, Data: unsafe.Pointer(stride)}
}

func RenderParamSWPointer(ptr unsafe.Pointer) RenderParam {
	return RenderParam{Type: RENDER_PARAM_SW_POINTER, Data: ptr}
}

func RenderParamNextFrameInfo(info *RenderFrameInfo) RenderParam {
	return RenderParam{Type: RENDER_PARAM_NEXT_FRAME_INFO, Data: unsafe.Pointer(info)}
}

func NewOpenGLInitParams(getProc OpenGLGetProcAddress) *OpenGLInitParams {
	if getProc == nil {
		return nil
	}
	return &OpenGLInitParams{}
}

func (p *OpenGLInitParams) Free() {}

func (m *Mpv) CreateRenderContext(params []RenderParam) (*RenderContext, error) {
	return nil, ErrCGODisabled
}

func (r *RenderContext) SetParameter(param RenderParam) error {
	return ErrCGODisabled
}

func (r *RenderContext) GetInfo(param RenderParam) error {
	return ErrCGODisabled
}

func (r *RenderContext) SetUpdateCallback(callback func()) {}

func (r *RenderContext) Update() uint64 {
	return 0
}

func (r *RenderContext) Render(params []RenderParam) error {
	return ErrCGODisabled
}

func (r *RenderContext) ReportSwap() {}

func (r *RenderContext) Free() {}
