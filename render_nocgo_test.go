//go:build !cgo
// +build !cgo

package mpv

import (
	"errors"
	"testing"
	"unsafe"
)

func TestNoCGORenderParamConstructors(t *testing.T) {
	raw := int32(1)
	if p := NewRenderParam(RENDER_PARAM_DEPTH, unsafe.Pointer(&raw)); p.Type != RENDER_PARAM_DEPTH || p.Data != unsafe.Pointer(&raw) {
		t.Fatalf("unexpected raw render param: %+v", p)
	}

	if p := RenderParamAPIType(RENDER_API_TYPE_SW); p.Type != RENDER_PARAM_API_TYPE {
		t.Fatalf("unexpected API type param: %+v", p)
	}

	fbo := &OpenGLFBO{FBO: 1, W: 2, H: 3, InternalFormat: 4}
	if p := RenderParamOpenGLFBO(fbo); p.Type != RENDER_PARAM_OPENGL_FBO || p.Data != unsafe.Pointer(fbo) {
		t.Fatalf("unexpected OpenGL FBO param: %+v", p)
	}

	drm := &OpenGLDRMParams{FD: -1, RenderFD: -1}
	if p := RenderParamDRMDisplay(drm); p.Type != RENDER_PARAM_DRM_DISPLAY || p.Data != unsafe.Pointer(drm) {
		t.Fatalf("unexpected DRM display param: %+v", p)
	}

	size := &OpenGLDRMDrawSurfaceSize{Width: 640, Height: 360}
	if p := RenderParamDRMDrawSurfaceSize(size); p.Type != RENDER_PARAM_DRM_DRAW_SURFACE_SIZE || p.Data != unsafe.Pointer(size) {
		t.Fatalf("unexpected DRM draw surface size param: %+v", p)
	}

	drmV2 := &OpenGLDRMParamsV2{FD: -1, RenderFD: -1}
	if p := RenderParamDRMDisplayV2(drmV2); p.Type != RENDER_PARAM_DRM_DISPLAY_V2 || p.Data != unsafe.Pointer(drmV2) {
		t.Fatalf("unexpected DRM display v2 param: %+v", p)
	}

	swSize := &[2]int32{16, 16}
	if p := RenderParamSWSize(swSize); p.Type != RENDER_PARAM_SW_SIZE || p.Data != unsafe.Pointer(swSize) {
		t.Fatalf("unexpected SW size param: %+v", p)
	}
	if p := RenderParamSWFormat("rgb0"); p.Type != RENDER_PARAM_SW_FORMAT {
		t.Fatalf("unexpected SW format param: %+v", p)
	}
	stride := uintptr(64)
	if p := RenderParamSWStride(&stride); p.Type != RENDER_PARAM_SW_STRIDE || p.Data != unsafe.Pointer(&stride) {
		t.Fatalf("unexpected SW stride param: %+v", p)
	}
	buffer := make([]byte, 16)
	if p := RenderParamSWPointer(unsafe.Pointer(&buffer[0])); p.Type != RENDER_PARAM_SW_POINTER || p.Data != unsafe.Pointer(&buffer[0]) {
		t.Fatalf("unexpected SW pointer param: %+v", p)
	}

	info := &RenderFrameInfo{}
	if p := RenderParamNextFrameInfo(info); p.Type != RENDER_PARAM_NEXT_FRAME_INFO || p.Data != unsafe.Pointer(info) {
		t.Fatalf("unexpected next frame info param: %+v", p)
	}
}

func TestNoCGORenderStubs(t *testing.T) {
	var m Mpv
	if rc, err := m.CreateRenderContext([]RenderParam{RenderParamAPIType(RENDER_API_TYPE_SW)}); !errors.Is(err, ErrCGODisabled) || rc != nil {
		t.Fatalf("CreateRenderContext = (%v, %v), want nil ErrCGODisabled", rc, err)
	}

	rc := &RenderContext{}
	if err := rc.SetParameter(RenderParamInt(RENDER_PARAM_DEPTH, new(int32))); !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("SetParameter err = %v, want ErrCGODisabled", err)
	}
	if err := rc.GetInfo(RenderParamNextFrameInfo(&RenderFrameInfo{})); !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("GetInfo err = %v, want ErrCGODisabled", err)
	}
	if err := rc.Render(nil); !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("Render err = %v, want ErrCGODisabled", err)
	}
	if flags := rc.Update(); flags != 0 {
		t.Fatalf("Update flags = %d, want 0", flags)
	}
	rc.SetUpdateCallback(func() {
		t.Fatal("nocgo render update callback should not be invoked")
	})
	rc.ReportSwap()
	rc.Free()

	if params := NewOpenGLInitParams(nil); params != nil {
		t.Fatalf("NewOpenGLInitParams(nil) = %v, want nil", params)
	}
	params := NewOpenGLInitParams(func(name string) unsafe.Pointer { return nil })
	if params == nil {
		t.Fatal("NewOpenGLInitParams returned nil for non-nil callback")
	}
	params.Free()
	params.Free()
}
